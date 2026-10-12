import { useAppDialog } from '@/feature/layout/store/appDialog.store'
import { isRunning, useProcessTerm } from '@/feature/process/store/processTerm.store'
import { ref, watch } from 'vue'
import { useTileTree } from '../store/tileTree.store'

const CLOSE_WAIT = 10_000   // 종료하고 닫기: 이만큼 기다린 뒤 닫는다

// kill 중인 uid — 머리 줄과 루트 터미널 탭이 같은 상태를 보도록 모듈에 둔다
const killing = ref<Record<string, true>>({})

// kill = SIGTERM / force = SIGKILL / close = 종료 요청 뒤 타일만 닫기
export type KillChoice = 'kill' | 'killClose' | 'force' | 'close'

const VIEW: Record<KillChoice, string> = { kill: '종료', killClose: '종료하고 닫기', force: '강제 종료', close: '닫기' }

const NOTE: Record<KillChoice, string> = {
    kill: '실행 중인 프로세스에 종료를 요청합니다(SIGTERM — 정리할 기회를 줌).',
    killClose: '종료하고 닫기: 최대 10초 기다린 뒤 타일을 닫습니다. 그때도 끝나지 않았고 조작할 수 있는 다른 계정도 열어 두지 않았으면 강제 종료됩니다.',
    force: '강제 종료: 바로 끝냅니다(SIGKILL).',
    close: '닫기: 타일만 닫습니다. 조작할 수 있는 계정이 모두 닫으면 강제 종료됩니다.',
}

// 터미널 kill은 어디서든 확인창 뒤에 한다 + 닫기 — SD3(실행 중 터미널 치우기), 종료 요청·강제 종료 → REF-process-kill.md
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
    const waitEnded = (uid: string) => new Promise<void>(resolve => {
        if (!isRunning(procs.value[uid]?.status)) return resolve()
        const done = () => {
            clearTimeout(timer)
            stop()
            resolve()
        }
        const timer = setTimeout(done, CLOSE_WAIT)
        const stop = watch(() => procs.value[uid]?.status, s => { if (!isRunning(s)) done() })
    })

    const runKill = async (tileId: string | undefined, uid: string, choice: KillChoice) => {
        if (choice === 'close') {
            if (tileId) closeTile(tileId)
            return
        }
        killing.value = { ...killing.value, [uid]: true }
        try {
            await kill(uid, choice === 'force')
            if (choice !== 'killClose' || !tileId) return
            // 안 끝났어도 닫는다 — 닫기로 '모두 닫음' 판정이 돌아 강제 종료될 수 있음(2026-10-12 결정 "기다렸다 닫기")
            await waitEnded(uid)
            closeTile(tileId)
        } catch (err) {
            openDialog({ type: 'error', content: (err as { message?: string })?.message || '종료하지 못했습니다' })
        } finally {
            const { [uid]: _, ...rest } = killing.value
            killing.value = rest
        }
    }

    // choices = 확인창에 보일 버튼(취소 제외). 이름은 title로만 넘긴다 — AppDialog의 content는 v-html
    const openKillDialog = (tileId: string | undefined, uid: string, name: string, choices: ReadonlyArray<KillChoice>, terminating: boolean) => {
        openDialog({
            type: 'warning',
            title: terminating ? `'${name}' 종료 중` : `'${name}' 종료`,
            content: [
                ...(terminating ? ['종료를 요청했지만 아직 끝나지 않았습니다.'] : []),
                ...choices.map(c => NOTE[c]),
            ],
            buttons: [
                { view: '취소', type: 'info', bindButton: 'Escape', action: () => dialogOpen.value = false },
                // Enter를 묶지 않는다 — 묶지 않은 Enter·Space는 AppDialog가 닫기(취소)로 처리
                ...choices.map(c => ({
                    view: VIEW[c], type: c === 'close' ? 'info' as const : 'error' as const,
                    action: () => {
                        dialogOpen.value = false
                        void runKill(tileId, uid, c)
                    },
                })),
            ],
        })
    }

    // choices 생략 = 상태에 맞춘 기본: 처음 = 종료·종료하고 닫기 / 종료 요청됨 = 강제 종료·닫기(두 번째 종료 = 강제)
    const confirmKill = (tileId: string, name: string, choices?: ReadonlyArray<KillChoice>) => {
        const t = tiles.value[tileId]
        if (t?.type !== 'terminal') return
        const terminating = !!procs.value[t.uid]?.terminating
        openKillDialog(tileId, t.uid, name, choices ?? (terminating ? ['force', 'close'] : ['kill', 'killClose']), terminating)
    }

    // 현재 보드 트리에 없는 터미널(스크립트 행 "실행 중" 목록) — 타일을 닫을 수 없어 종료·강제 종료만
    const confirmKillUid = (uid: string, name: string, terminating: boolean) =>
        openKillDialog(undefined, uid, name, terminating ? ['force'] : ['kill'], terminating)

    return { killing, closeTile, confirmKill, confirmKillUid }
}
