import { useAppDialog } from '@/feature/layout/store/appDialog.store'
import { isRunning, useProcessTerm } from '@/feature/process/store/processTerm.store'
import { ref, watch } from 'vue'
import { useTileTree } from '../store/tileTree.store'

const CLOSE_WAIT = 10_000   // 종료하고 닫기: 이 안에 끝나지 않으면 타일을 남긴다

// kill 중인 uid — 머리 줄과 루트 터미널 탭이 같은 상태를 보도록 모듈에 둔다
const killing = ref<Record<string, true>>({})

export type KillChoice = 'kill' | 'killClose'

// 터미널 kill은 어디서든 확인창 뒤에 한다 + 닫기 — SD3(실행 중 터미널 치우기), REF-process-share-realtime-ui.md
export const useTerminalKill = () => {
    const { tiles, close } = useTileTree()
    const { procs, kill, dispose } = useProcessTerm()
    const { open: dialogOpen, openDialog } = useAppDialog()

    const closeTile = (tileId: string) => {
        const t = tiles.value[tileId]
        if (!t) return
        close(t.id)
        if (t.type === 'terminal') dispose(t.uid)
    }

    // kill은 신호만 보내고 종료는 STATUS로 비동기로 온다
    const waitEnded = (uid: string) => new Promise<boolean>(resolve => {
        if (!isRunning(procs.value[uid]?.status)) return resolve(true)
        const done = (ok: boolean) => {
            clearTimeout(timer)
            stop()
            resolve(ok)
        }
        const timer = setTimeout(() => done(false), CLOSE_WAIT)
        const stop = watch(() => procs.value[uid]?.status, s => { if (!isRunning(s)) done(true) })
    })

    const runKill = async (tileId: string | undefined, uid: string, choice: KillChoice) => {
        killing.value = { ...killing.value, [uid]: true }
        try {
            await kill(uid)
            if (choice !== 'killClose' || !tileId) return
            if (await waitEnded(uid)) closeTile(tileId)
            else openDialog({ type: 'warning', content: '프로세스가 아직 종료되지 않아 타일을 남겼습니다.' })
        } catch (err) {
            openDialog({ type: 'error', content: (err as { message?: string })?.message || '종료하지 못했습니다' })
        } finally {
            const { [uid]: _, ...rest } = killing.value
            killing.value = rest
        }
    }

    // choices = 확인창에 보일 버튼(취소 제외). 이름은 title로만 넘긴다 — AppDialog의 content는 v-html
    // tileId가 없으면(다른 보드·위치 없음) 타일을 닫을 수 없어 종료만
    const openKillDialog = (tileId: string | undefined, uid: string, name: string, choices: ReadonlyArray<KillChoice>) => {
        const view: Record<KillChoice, string> = { kill: '종료', killClose: '종료하고 닫기' }
        openDialog({
            type: 'warning',
            title: `'${name}' 종료`,
            content: choices.includes('killClose')
                ? ['실행 중인 프로세스를 종료합니다.', '종료하고 닫기: 종료가 확인되면 타일도 닫습니다.']
                : '실행 중인 프로세스를 종료합니다. 타일은 DONE으로 남습니다.',
            buttons: [
                { view: '취소', type: 'info', bindButton: 'Escape', action: () => dialogOpen.value = false },
                // Enter를 묶지 않는다 — 묶지 않은 Enter·Space는 AppDialog가 닫기(취소)로 처리
                ...choices.map(c => ({
                    view: view[c], type: 'error' as const,
                    action: () => {
                        dialogOpen.value = false
                        void runKill(tileId, uid, c)
                    },
                })),
            ],
        })
    }

    const confirmKill = (tileId: string, name: string, choices: ReadonlyArray<KillChoice> = ['kill', 'killClose']) => {
        const t = tiles.value[tileId]
        if (t?.type !== 'terminal') return
        openKillDialog(tileId, t.uid, name, choices)
    }

    // 현재 보드 트리에 없는 터미널(스크립트 행 "실행 중" 목록) — 종료만
    const confirmKillUid = (uid: string, name: string) => openKillDialog(undefined, uid, name, ['kill'])

    return { killing, closeTile, confirmKill, confirmKillUid }
}
