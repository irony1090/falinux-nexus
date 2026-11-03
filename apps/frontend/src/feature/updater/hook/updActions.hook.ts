import { useAppDialog } from '@/feature/layout/store/appDialog.store'
import { useQueryClient } from '@tanstack/vue-query'
import { ref } from 'vue'
import { deleteFile, invalidateUpdates, modeText, saveText, setMode, uploadFile, type Ref, type UpdFile } from '../api/updater.api'
import { useUpdPrompt } from '../component/UpdPrompt.vue'
import type { UpItem } from '../util/updDrop.util'

// 장비 루트를 만든 직후 받은 키 — 장비 타일이 처음 열릴 때 펼쳐 보인다(회원 페이지 freshKeys와 같은 뜻)
export const freshUpdKeys = ref<Record<string, string>>({})

// updater 타일 공용 — 요청 하나(busy·오류 한 줄) · 확인 창 · 업로드(겹치는 이름은 한 번에 확인) · 새 텍스트 파일
export const useUpdRun = () => {
    const qc = useQueryClient()
    const busy = ref(false)
    const message = ref<string>()
    const progress = ref<string>()
    const run = async <T,>(job: () => Promise<T>, done?: (v: T) => void): Promise<T | undefined> => {
        busy.value = true
        message.value = undefined
        try {
            const v = await job()
            done?.(v)
            await invalidateUpdates(qc)
            return v
        } catch (err: any) {
            message.value = err?.message || '요청에 실패했습니다'
            return undefined
        } finally {
            busy.value = false
        }
    }
    return { qc, busy, message, progress, run }
}

export const useUpdConfirm = () => {
    const { open, openDialog } = useAppDialog()
    // 확인 창 — 누르면 true
    return (title: string, content: string | string[], ok: string, type: 'warning' | 'error' = 'warning') => new Promise<boolean>(resolve => {
        openDialog({
            type, title, content,
            buttons: [
                { view: '취소', type: 'info', bindButton: 'Escape', action: () => { open.value = false; resolve(false) } },
                { view: ok, type, action: () => { open.value = false; resolve(true) } },
            ],
        })
    })
}

export const useUpdUpload = (ctx: ReturnType<typeof useUpdRun>) => {
    const confirm = useUpdConfirm()
    const prompt = useUpdPrompt()
    const { busy, message, progress, run } = ctx

    // 버튼·끌어 놓기 공용(items = 올릴 경로까지 정한 목록). 있는 파일 목록(have)으로 겹치는 이름을 먼저 세어 한 번에 묻는다
    const upload = async (r: Ref, items: UpItem[], have: UpdFile[]) => {
        if (!items.length) { message.value = '올릴 파일이 없습니다'; return }
        const files = items.map(it => it.file), paths = items.map(it => it.path)
        const exists = new Set(have.filter(f => f.scope === 'main').map(f => f.path))
        const dup = paths.filter(p => exists.has(p))
        if (dup.length && !(await confirm('같은 이름 파일', [`${dup.length}개 파일이 이미 있습니다. 덮어쓸까요?`, dup.slice(0, 5).join(', ') + (dup.length > 5 ? ' …' : '')], '덮어쓰기'))) return
        const total = files.reduce((a, f) => a + f.size, 0)
        let sent = 0
        await run(async () => {
            for (const [i, f] of files.entries()) {
                const before = sent
                const res = await uploadFile(r, paths[i]!, f, true, d => {
                    progress.value = `올리는 중 ${i + 1}/${files.length} · ${total ? Math.floor((before + d) / total * 100) : 100}%`
                })
                if (res === 'exists') throw { message: `${paths[i]} 파일이 이미 있습니다` }
                sent += f.size
            }
        })
        progress.value = undefined
    }

    // 새 텍스트 파일 — 이름만 받고 빈 내용으로 만든 뒤 Script 타일로 연다(onCreated)
    const newText = async (r: Ref, onCreated: (path: string) => void) => {
        const v = await prompt({
            title: '새 텍스트 파일', ok: '만들기',
            fields: [{ key: 'path', label: '경로', placeholder: '예: 03-start.sh · conf/app.env', hint: '폴더는 / 로 나눕니다. .sh는 LF로 저장됩니다' }],
        })
        if (!v?.path) return
        const path = v.path.replace(/\\/g, '/').replace(/^\/+/, '')
        const ok = await run(() => saveText(r, path, path.endsWith('.sh') ? '#!/bin/sh\nset -eu\n' : '', true))
        if (ok) onCreated(path)
    }
    return { busy, message, progress, upload, newText }
}

// 파일 지우기(한 개·여러 개 공용) — 확인창 한 번, 차례로 지우다 실패하면 멈추고 지운 개수·실패 경로를 보인다(결정-2(삭제 확인), 원자적 아님)
export const useUpdRemove = (ctx: ReturnType<typeof useUpdRun>) => {
    const confirm = useUpdConfirm()
    const { progress, run } = ctx
    return async (r: Ref, paths: string[], title: string): Promise<boolean> => {
        if (!paths.length) return false
        const lines = paths.length === 1 ? ['이 빌드에서 파일을 지웁니다.'] : [`${paths.length}개 파일을 이 빌드에서 지웁니다.`, paths.slice(0, 5).join(', ') + (paths.length > 5 ? ' …' : '')]
        if (paths.includes('steps.json')) lines.push('steps.json을 지우면 uses도 비워집니다.')
        if (!(await confirm(title, lines, paths.length === 1 ? '삭제' : `삭제 ${paths.length}`, 'error'))) return false
        let done = 0
        await run(async () => {
            for (const p of paths) {
                if (paths.length > 1) progress.value = `지우는 중 ${done + 1}/${paths.length}`
                try {
                    await deleteFile(r, p)
                } catch (err: any) {
                    throw { message: `${done ? `${done}개 지운 뒤 ` : ''}${p} 삭제 실패: ${err?.message || '요청에 실패했습니다'}` }
                }
                done++
            }
        })
        progress.value = undefined
        return done === paths.length
    }
}

// 파일 권한(정-2(실행 권한)) — `실행 권한` 칸 = 실행 비트(111)만 켜고 끔(단위-라(실행 권한 이름)), 8진수 = 그대로. updater가 copy 때 적용(Files·공용 묶음 타일)
export const useUpdMode = (bundle: Ref, run: ReturnType<typeof useUpdRun>['run']) => {
    const prompt = useUpdPrompt()
    const isExec = (f: UpdFile) => (f.mode & 0o100) !== 0
    const toggleExec = (f: UpdFile) => run(() => setMode(bundle, f.path, isExec(f) ? f.mode & ~0o111 : f.mode | 0o111))
    const editMode = async (f: UpdFile) => {
        const v = await prompt({
            title: `${f.path} 권한`, ok: '바꾸기',
            note: '장비에 깔릴 때 이 권한으로 적용됩니다(Windows는 무시). 권한만 바꿔도 sha가 바뀝니다',
            fields: [{ key: 'mode', label: '권한(8진수)', value: modeText(f.mode), placeholder: '예: 755', pattern: /^[0-7]{3}$/, patternMsg: '0~7 숫자 세 자리' }],
        })
        if (v?.mode) await run(() => setMode(bundle, f.path, parseInt(v.mode!, 8)))
    }
    return { isExec, toggleExec, editMode }
}
