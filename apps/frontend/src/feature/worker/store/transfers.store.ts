import { useTestSocket } from '@/common/websocket/websocket.hook'
import { useAppDialog } from '@/feature/layout/store/appDialog.store'
import { computed, inject, onBeforeUnmount, provide, ref, shallowRef, watch } from 'vue'
import {
    abortTransfer, dirExists, finishUpload, listTransfers, putUploadChunk, startUpload, UPLOAD_CHUNK,
    type TransferInfo,
} from '../api/transfer.api'

const TRANSFERS_STORE_KEY = Symbol('TransfersStore')

const CHUNK_TRIES = 3   // 조각 하나를 네트워크 오류로 다시 보내는 횟수

// 업로드를 시작한 곳 — 장비 타일 하나
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

    const put = (t: TransferInfo) => {
        const old = items.value[t.id]
        if (old && !newer(t, old)) return
        items.value = { ...items.value, [t.id]: t }
        if (t.state === 'done' && old?.state !== 'done') lastDone.value = { instanceKey: t.instanceKey, dir: t.dir, at: Date.now() }
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

    // Q28(여러 파일 겹침): 한 창에 모아서. 같은 이름 폴더는 덮어쓸 수 없어 늘 뺀다
    const askConflict = (files: string[], dirs: string[]) => new Promise<Choice>(resolve => {
        const close = (c: Choice) => () => { dialogOpen.value = false; resolve(c) }
        const content = [
            ...(files.length ? [`이미 있는 파일 ${files.length}개: ${files.join(', ')}`] : []),
            ...(dirs.length ? [`같은 이름의 폴더가 있어 올리지 않음: ${dirs.join(', ')}`] : []),
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
                { view: '나머지 올리기', type: 'info', bindButton: 'Enter', action: close('skip') },
            ],
        })
    })

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
        let found
        try {
            found = await dirExists(src.instance, src.folderId, src.dir, files.map(f => f.name))
        } catch (err) {
            openDialog({ type: 'error', content: `올리지 못했습니다: ${errMsg(err)}` })
            return
        }
        const dirs = new Set(found.filter(e => e.kind === 'dir').map(e => e.name))
        const same = new Set(found.filter(e => e.kind !== 'dir').map(e => e.name))
        let todo = files.filter(f => !dirs.has(f.name))
        let overwrite = false
        if (same.size || dirs.size) {
            if (!todo.length) {
                openDialog({ type: 'warning', content: `같은 이름의 폴더가 있어 올리지 않았습니다: ${[...dirs].join(', ')}` })
                return
            }
            const c = await askConflict([...same], [...dirs])
            if (c === 'cancel') return
            if (c === 'skip') todo = todo.filter(f => !same.has(f.name))
            overwrite = c === 'overwrite'
        }
        const errors: string[] = []
        for (const f of todo) {
            const e = await uploadOne(src, f, overwrite && same.has(f.name))
            if (e) errors.push(e)
        }
        if (errors.length) openDialog({ type: 'error', title: '올리기 실패', content: errors })
    }

    // 중단 — 올리는 중이면 이 탭의 남은 조각부터 멈춘다
    const abort = async (id: string) => {
        uploading.get(id)?.abort()
        try {
            await abortTransfer(id)
        } catch (err) {
            openDialog({ type: 'error', content: `중단하지 못했습니다: ${errMsg(err)}` })
        }
    }

    const ctx = { list, running, failed, lastDone, byTile, upload, abort }
    provide(TRANSFERS_STORE_KEY, ctx)
    return ctx
}

export const useTransfers = () => {
    const context = inject<ReturnType<typeof provideTransfers>>(TRANSFERS_STORE_KEY)
    if (!context) throw new Error('TransfersStore is not provided')
    return context
}
