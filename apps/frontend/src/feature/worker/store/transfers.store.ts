import { useTestSocket } from '@/common/websocket/websocket.hook'
import { useAppDialog } from '@/feature/layout/store/appDialog.store'
import { computed, inject, onBeforeUnmount, provide, ref, shallowRef, watch } from 'vue'
import {
    abortTransfer, deviceRef, dirExists, downloadUrl, finishUpload, listTransfers, putUploadChunk, startCopy, startDownload, startUpload, UPLOAD_CHUNK,
    type TransferInfo,
} from '../api/transfer.api'
import type { DirEntry } from '../api/worker.api'
import { isFileLike } from '../util/workerDir.util'

const TRANSFERS_STORE_KEY = Symbol('TransfersStore')

const CHUNK_TRIES = 3   // 조각 하나를 네트워크 오류로 다시 보내는 횟수

// 장비 타일 하나가 지금 보는 폴더 — 올리기·내려받기·worker 간 전송의 출발·도착
export type UploadSource = { tileId: string; instance: string; folderId: number; dir: string }

type Choice = 'overwrite' | 'skip' | 'cancel'

// 늦게 온 옛 이벤트·응답이 새 상태를 덮지 않게: 끝남 > 진행, 보내는 중 > 올리는 중, 같으면 많이 옮긴 쪽
const rank = (t: TransferInfo) => [t.state === 'running' ? 0 : 1, t.phase === 'upload' ? 0 : 1, t.done] as const
const newer = (a: TransferInfo, b: TransferInfo) => {
    const ra = rank(a), rb = rank(b)
    for (let i = 0; i < ra.length; i++) if (ra[i] !== rb[i]) return ra[i]! > rb[i]!
    return true
}

const errMsg = (err: unknown) => (err as { message?: string })?.message || '알 수 없는 오류'

// 전송 목록 — 폴더-8(전송 상태). 서버 목록(메모리, Q29(전송 목록 보관))을 TRANSFER:* 이벤트로 따라간다. REF-transfer-ext.md
export const provideTransfers = () => {
    const { on, status } = useTestSocket()
    const { open: dialogOpen, openDialog } = useAppDialog()

    const items = shallowRef<Record<string, TransferInfo>>({})
    // 끝난 전송의 받을 폴더 — 그 폴더를 보는 장비 타일이 목록을 다시 읽는다
    const lastDone = ref<{ instanceKey: string; dir: string; at: number } | null>(null)
    // 이 탭이 조각을 올리는 중인 전송 — 중단하면 남은 조각을 안 보낸다
    const uploading = new Map<string, AbortController>()

    // 내려받기를 받는 숨은 iframe — 끝나고 조금 뒤에 뺀다(완료 알림이 응답 처리보다 먼저 올 수 있음). a 링크로 받으면 서버 오류 응답에 앱 화면이 넘어감
    const frames = new Map<string, HTMLIFrameElement>()

    const put = (t: TransferInfo) => {
        const old = items.value[t.id]
        if (old && !newer(t, old)) return
        items.value = { ...items.value, [t.id]: t }
        if (t.state === 'done' && old?.state !== 'done' && t.kind !== 'download') lastDone.value = { instanceKey: t.instanceKey, dir: t.dir, at: Date.now() }
        const frame = frames.get(t.id)
        if (frame && t.state !== 'running') {
            frames.delete(t.id)
            setTimeout(() => frame.remove(), 30_000)
        }
    }

    const load = async () => {
        try {
            const list = await listTransfers()
            items.value = Object.fromEntries(list.map(t => [t.id, t]))
        } catch (err) {
            console.log('[TRANSFERS] load', err)
        }
    }

    const offs = [
        on<TransferInfo>('TRANSFER:UPDATE', put),
        on<{ id: string }>('TRANSFER:GONE', ev => {
            const { [ev.id]: _, ...rest } = items.value
            items.value = rest
        }),
    ]
    // 끊긴 동안 놓친 변경
    watch(status, s => {
        if (s === 'CONNECTED') void load()
    }, { immediate: true })
    onBeforeUnmount(() => offs.forEach(off => off()))

    const list = computed(() => Object.values(items.value).sort((a, b) => a.startedAt - b.startedAt))
    const running = computed(() => list.value.filter(t => t.state === 'running').length)
    const failed = computed(() => list.value.some(t => t.state === 'failed'))
    const byTile = (tileId: string) => list.value.filter(t => t.tileId === tileId)

    // Q28(여러 파일 겹침): 한 창에 모아서. 같은 이름 폴더는 덮어쓸 수 없어 늘 뺀다. verb = '올리' | '보내'
    const askConflict = (files: string[], dirs: string[], verb: string) => new Promise<Choice>(resolve => {
        const close = (c: Choice) => () => { dialogOpen.value = false; resolve(c) }
        const content = [
            ...(files.length ? [`이미 있는 파일 ${files.length}개: ${files.join(', ')}`] : []),
            ...(dirs.length ? [`같은 이름의 폴더가 있어 ${verb}지 않음: ${dirs.join(', ')}`] : []),
        ]
        openDialog({
            type: 'warning',
            title: '같은 이름',
            content,
            buttons: files.length ? [
                { view: '취소', type: 'info', bindButton: 'Escape', action: close('cancel') },
                { view: '겹치는 것 건너뛰기', type: 'info', action: close('skip') },
                { view: '모두 덮어쓰기', type: 'warning', bindButton: 'Enter', action: close('overwrite') },
            ] : [
                { view: '취소', type: 'info', bindButton: 'Escape', action: close('cancel') },
                { view: `나머지 ${verb}기`, type: 'info', bindButton: 'Enter', action: close('skip') },
            ],
        })
    })

    // 받을 폴더의 같은 이름 확인 -> Q28 창. 보낼 이름과 덮어쓸지를 돌려준다(취소·실패 = undefined)
    const resolveNames = async (dst: UploadSource, names: string[], verb: string) => {
        let found
        try {
            found = await dirExists(dst.instance, dst.folderId, dst.dir, names)
        } catch (err) {
            openDialog({ type: 'error', content: `${verb}지 못했습니다: ${errMsg(err)}` })
            return
        }
        const dirs = new Set(found.filter(e => e.kind === 'dir').map(e => e.name))
        const same = new Set(found.filter(e => e.kind !== 'dir').map(e => e.name))
        let todo = names.filter(n => !dirs.has(n))
        let overwrite = false
        if (same.size || dirs.size) {
            if (!todo.length) {
                openDialog({ type: 'warning', content: `같은 이름의 폴더가 있어 ${verb}지 않았습니다: ${[...dirs].join(', ')}` })
                return
            }
            const c = await askConflict([...same], [...dirs], verb)
            if (c === 'cancel') return
            if (c === 'skip') todo = todo.filter(n => !same.has(n))
            overwrite = c === 'overwrite'
        }
        return { todo: new Set(todo), overwrite: (name: string) => overwrite && same.has(name) }
    }

    // 파일 하나: 시작 -> 조각 -> finish. 실패 메시지를 돌려준다(없으면 성공 또는 중단)
    const uploadOne = async (src: UploadSource, file: File, overwrite: boolean): Promise<string | undefined> => {
        let info: TransferInfo
        try {
            info = await startUpload(src.instance, { folderId: src.folderId, path: src.dir, name: file.name, size: file.size, overwrite, tileId: src.tileId })
        } catch (err) {
            return `${file.name}: ${errMsg(err)}`
        }
        put(info)
        const ac = new AbortController()
        uploading.set(info.id, ac)
        try {
            for (let offset = 0; offset < file.size; offset += UPLOAD_CHUNK) {
                const chunk = file.slice(offset, offset + UPLOAD_CHUNK)
                for (let tryNo = 1; ; tryNo++) {
                    if (ac.signal.aborted) return
                    try {
                        await putUploadChunk(info.id, offset, chunk, ac.signal)
                        break
                    } catch (err) {
                        // 서버가 답한 오류(409 중단됨 등)·탭이 중단 = 그만 / 네트워크 오류만 다시
                        if (ac.signal.aborted) return
                        if ((err as { status?: number }).status || tryNo >= CHUNK_TRIES) throw err
                        await new Promise(r => setTimeout(r, 1000 * tryNo))
                    }
                }
            }
            put(await finishUpload(info.id))
        } catch (err) {
            // 다른 탭·기기에서 중단했으면 조각이 409로 거절됨 — 실패로 알리지 않는다
            if (ac.signal.aborted || items.value[info.id]?.state !== 'running') return
            void abortTransfer(info.id).catch(() => {})
            return `${file.name}: ${errMsg(err)}`
        } finally {
            uploading.delete(info.id)
        }
    }

    // 장비 타일에서 올리기(메뉴·끌어 놓기) — 겹침 확인 뒤 한 파일씩
    const upload = async (src: UploadSource, files: File[]) => {
        if (!files.length) return
        const r = await resolveNames(src, files.map(f => f.name), '올리')
        if (!r) return
        const errors: string[] = []
        for (const f of files.filter(f => r.todo.has(f.name))) {
            const e = await uploadOne(src, f, r.overwrite(f.name))
            if (e) errors.push(e)
        }
        if (errors.length) openDialog({ type: 'error', title: '올리기 실패', content: errors })
    }

    // 내려받기 — 파일 하나는 그대로, 여러 개·폴더는 zip(Q33(내려받기 범위)). 본문은 브라우저 기본 다운로드가 받는다
    const download = async (src: UploadSource, entries: DirEntry[]) => {
        if (!entries.length) return
        const one = entries.length === 1 && isFileLike(entries[0]!) ? entries[0] : undefined
        let info: TransferInfo
        try {
            info = await startDownload(src.instance, { folderId: src.folderId, dir: src.dir, names: entries.map(e => e.name), path: one?.path, tileId: src.tileId })
        } catch (err) {
            openDialog({ type: 'error', content: `내려받지 못했습니다: ${errMsg(err)}` })
            return
        }
        const frame = document.createElement('iframe')
        frame.hidden = true
        frame.src = downloadUrl(info.id)
        document.body.appendChild(frame)
        frames.set(info.id, frame)
        put(info)
    }

    // worker 간 전송 — 파일만(Q36(worker 간 폴더)), 겹침 확인 뒤 파일마다 요청(동시 수는 서버가 제한). 받는 타일에 막대
    const copyTo = async (src: UploadSource, entries: DirEntry[], dst: UploadSource) => {
        const files = entries.filter(isFileLike)
        const skipped = entries.filter(e => !isFileLike(e)).map(e => e.name)
        const notes = skipped.length ? [`파일이 아니라 보내지 않음: ${skipped.join(', ')}`] : []
        if (!files.length) {
            openDialog({ type: 'warning', content: notes })
            return
        }
        const r = await resolveNames(dst, files.map(e => e.name), '보내')
        if (!r) return
        const errors: string[] = []
        for (const e of files.filter(e => r.todo.has(e.name))) {
            try {
                put(await startCopy({
                    from: { ...deviceRef(src.instance, src.folderId), path: e.path },
                    to: { ...deviceRef(dst.instance, dst.folderId), dir: dst.dir },
                    overwrite: r.overwrite(e.name),
                    tileId: dst.tileId,
                }))
            } catch (err) {
                errors.push(`${e.name}: ${errMsg(err)}`)
            }
        }
        if (errors.length || notes.length) openDialog({ type: errors.length ? 'error' : 'warning', title: '보내기', content: [...notes, ...errors] })
    }

    // 보드에 열린 장비 타일이 지금 보는 폴더(올릴 수 있을 때만) — "다른 장비로 보내기"의 받을 곳 목록(Q34(worker 간 전송 시작 방법))
    const spots = shallowRef<Record<string, UploadSource>>({})
    const setSpot = (tileId: string, spot: UploadSource | undefined) => {
        const { [tileId]: _, ...rest } = spots.value
        spots.value = spot ? { ...rest, [tileId]: spot } : rest
    }

    // 장비 타일 사이 끌어 놓기 중인 항목 — dragover에서는 dataTransfer 내용을 못 읽어서 여기 둔다(전송-12)
    const dragging = shallowRef<{ src: UploadSource; entries: DirEntry[] } | null>(null)

    // 중단 — 올리는 중이면 이 탭의 남은 조각부터 멈춘다
    const abort = async (id: string) => {
        uploading.get(id)?.abort()
        try {
            await abortTransfer(id)
        } catch (err) {
            openDialog({ type: 'error', content: `중단하지 못했습니다: ${errMsg(err)}` })
        }
    }

    const ctx = { list, running, failed, lastDone, byTile, upload, download, copyTo, spots, setSpot, dragging, abort }
    provide(TRANSFERS_STORE_KEY, ctx)
    return ctx
}

export const useTransfers = () => {
    const context = inject<ReturnType<typeof provideTransfers>>(TRANSFERS_STORE_KEY)
    if (!context) throw new Error('TransfersStore is not provided')
    return context
}
