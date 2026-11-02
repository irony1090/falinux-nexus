import { useAppDialog } from '@/feature/layout/store/appDialog.store'
import { useQueryClient } from '@tanstack/vue-query'
import { ref } from 'vue'
import { invalidateUpdates, saveText, uploadFile, type Ref, type UpdFile } from '../api/updater.api'
import { useUpdPrompt } from '../component/UpdPrompt.vue'

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

// 폴더 올리기 = 고른 폴더가 번들 맨 위(첫 폴더 이름을 뗀다)
const relPathOf = (f: File, folder: boolean) => {
    const rel = folder && f.webkitRelativePath ? f.webkitRelativePath : f.name
    return folder ? rel.split('/').slice(1).join('/') || f.name : rel
}

export const useUpdUpload = (ctx: ReturnType<typeof useUpdRun>) => {
    const confirm = useUpdConfirm()
    const prompt = useUpdPrompt()
    const { busy, message, progress, run } = ctx

    // 있는 파일 목록(have)으로 겹치는 이름을 먼저 세어 한 번에 묻는다
    const upload = async (r: Ref, list: FileList | File[], folder: boolean, have: UpdFile[]) => {
        const files = Array.from(list)
        if (!files.length) return
        const paths = files.map(f => relPathOf(f, folder))
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
