import { tabReady } from '@/common/util/tabId.util'
import { listProcesses, lookupProcesses } from '@/feature/process/api/process.api'
import { isRunning, type provideProcessTerm } from '@/feature/process/store/processTerm.store'
import { watch } from 'vue'
import type { provideTileTree } from '../store/tileTree.store'

// 트리의 터미널 타일과 processTerm 등록을 맞춘다 — ⑪ 3-d(터미널 복원) + 4-c(터미널 등록 일원화).
// 새로고침 복원·다른 탭 push·409 재적용으로 트리에 처음 보이는 터미널이 전부 이 길로 등록된다(REF-node-ui-sync.md).
// 순서 의존:
// 1) 탭 id를 받은 뒤 — 받기 전 요청은 sizeOwner가 전부 false
// 2) 트리를 받은 뒤 process 조회 — 반대면 그 사이 다른 탭이 exec한 터미널은 트리엔 있고 등록은 안 돼 빈칸이 된다.
//    이 순서면 그 process는 "타일 없는 process"로 잡히고, adoptOrphans가 uid로 검사해 409 재적용 때 건너뛴다
// 두 스토어를 provide하는 곳(pages/index.vue)에서 부르므로 inject 대신 인자로 받는다 — provide한 컴포넌트 자신은 inject 불가
export const useTerminalRestore = (
    { ready, tiles, adoptOrphans }: ReturnType<typeof provideTileTree>,
    { procs, restored, restore, prune }: ReturnType<typeof provideProcessTerm>,
) => {
    let started = false             // 새로고침 복원 — 트리 load마다 1번
    let failed = false              // 복원 실패 — 소켓 재연결(tabReady)·트리 재로드 때만 다시(바로 재시도하면 서버가 죽었을 때 끝없이 돈다)
    let syncing = false             // 조회는 한 번에 하나, 끝나면 그 사이 바뀐 것을 다시 본다
    const missing = new Set<string>()   // 조회해도 없던 uid(남의 것·삭제) — 다시 조회하지 않음

    const treeUids = () => new Set(Object.values(tiles.value).flatMap(t => t.type === 'terminal' ? [t.uid] : []))

    const reload = () => {
        started = true
        syncing = true
        const uids = [...treeUids()]
        restore(async () => {
            const [found, live] = await Promise.all([uids.length ? lookupProcesses(uids) : [], listProcesses()])
            const byUid = new Map([...found, ...live].map(p => [p.uid, p]))
            return [...byUid.values()]
        }, { reload: true })
        .then(list => {
            const found = new Set(list.map(p => p.uid))
            uids.forEach(u => { if (!found.has(u)) missing.add(u) })
            adoptOrphans(list.filter(p => isRunning(p.status)))
        })
        .catch(err => {
            console.log('[TERMINAL_RESTORE]', err)
            started = false
            failed = true
        })
        .finally(() => {
            syncing = false
            next()
        })
    }

    // 새로고침 복원 뒤: 트리에 새로 보이는 미등록 터미널 등록 + 빠진 끝난 터미널 정리(4④)
    const sync = () => {
        if (!restored.value || syncing) return
        const uids = treeUids()
        prune(uids)
        const unknown = [...uids].filter(u => !procs.value[u] && !missing.has(u))
        if (!unknown.length) return
        syncing = true
        restore(() => lookupProcesses(unknown), { reload: false })
            .then(list => {
                const found = new Set(list.map(p => p.uid))
                unknown.forEach(u => { if (!found.has(u)) missing.add(u) })
            })
            .catch(err => console.log('[TERMINAL_RESTORE] sync', err))
            .finally(() => {
                syncing = false
                next()
            })
    }

    // 조회 중에 트리를 다시 불러왔으면 끝난 뒤 새로고침 복원부터
    const next = () => {
        if (syncing || !ready.value) return
        if (!started) {
            if (tabReady.value && !failed) reload()
        } else sync()
    }

    watch([ready, tabReady], ([r]) => {
        failed = false
        if (!r) {
            started = false
            missing.clear()
        }
        next()
    }, { immediate: true })

    watch(() => [...treeUids()].sort().join(), next)
}
