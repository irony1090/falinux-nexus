import { tabReady } from '@/common/util/tabId.util'
import { listProcesses, lookupProcesses } from '@/feature/process/api/process.api'
import { isRunning, type provideProcessTerm } from '@/feature/process/store/processTerm.store'
import { watch } from 'vue'
import type { provideTileTree } from '../store/tileTree.store'

// 새로고침 뒤 터미널 타일 복원 — ⑪ 3-d(터미널 복원), REF-node-ui-save.md "3-d(터미널 복원) 구조안".
// 순서 의존:
// 1) 탭 id를 받은 뒤 — 받기 전 요청은 sizeOwner가 전부 false
// 2) 트리를 받은 뒤 process 조회 — 반대면 그 사이 다른 탭이 exec한 터미널은 트리엔 있고 등록은 안 돼 빈칸이 된다.
//    이 순서면 그 process는 "타일 없는 process"로 잡히고, adoptOrphans가 uid로 검사해 409 재적용 때 건너뛴다
// 두 스토어를 provide하는 곳(pages/index.vue)에서 부르므로 inject 대신 인자로 받는다 — provide한 컴포넌트 자신은 inject 불가
export const useTerminalRestore = (
    { ready, tiles, adoptOrphans }: ReturnType<typeof provideTileTree>,
    { restore }: ReturnType<typeof provideProcessTerm>,
) => {

    let started = false     // 트리 load마다 1번

    const run = () => {
        started = true
        const uids = Object.values(tiles.value).flatMap(t => t.type === 'terminal' ? [t.uid] : [])
        restore(async () => {
            const [found, live] = await Promise.all([uids.length ? lookupProcesses(uids) : [], listProcesses()])
            const byUid = new Map([...found, ...live].map(p => [p.uid, p]))
            return [...byUid.values()]
        })
        .then(list => adoptOrphans(list.filter(p => isRunning(p.status))))
        .catch(err => {
            console.log('[TERMINAL_RESTORE]', err)
            started = false     // 다음 소켓 재연결(tabReady) 때 다시 시도
        })
    }

    watch([ready, tabReady], ([r, t]) => {
        if (!r) started = false
        else if (t && !started) run()
    }, { immediate: true })
}
