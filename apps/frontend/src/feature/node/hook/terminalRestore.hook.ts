import { tabReady } from '@/common/util/tabId.util'
import { lookupProcesses } from '@/feature/process/api/process.api'
import { type provideProcessTerm } from '@/feature/process/store/processTerm.store'
import { useTestSocket } from '@/common/websocket/websocket.hook'
import { onBeforeUnmount, watch } from 'vue'
import type { provideTileTree } from '../store/tileTree.store'

// 트리의 터미널 타일과 processTerm 등록을 맞춘다 — ⑪ 3-d(터미널 복원) + 4-c(터미널 등록 일원화).
// 새로고침 복원·다른 탭 push·409 재적용으로 트리에 처음 보이는 터미널이 전부 이 길로 등록된다(REF-node-ui-sync.md).
// 순서 의존:
// 1) 탭 id를 받은 뒤 — 받기 전 요청은 sizeOwner가 전부 false
// 2) 트리를 받은 뒤 process 조회 — 트리에 있는 uid만 조회한다. 그 뒤 push로 트리에 생긴 터미널은 sync가 등록.
//    살아 있는데 타일이 없는 process는 서버가 GET /tiles 때 기본 보드에 붙인다(WS3(고아 안전망 위치)) — 프론트 안전망 없음
// 두 스토어를 provide하는 곳(pages/index.vue)에서 부르므로 inject 대신 인자로 받는다 — provide한 컴포넌트 자신은 inject 불가
export const useTerminalRestore = (
    { ready, tiles }: ReturnType<typeof provideTileTree>,
    { procs, restored, restore, prune, markGone, clearGone, refreshAccount, pruneGone }: ReturnType<typeof provideProcessTerm>,
) => {
    const { on } = useTestSocket()
    let started = false             // 새로고침 복원 — 트리 load마다 1번
    let failed = false              // 복원 실패 — 소켓 재연결(tabReady)·트리 재로드 때만 다시(바로 재시도하면 서버가 죽었을 때 끝없이 돈다)
    let syncing = false             // 조회는 한 번에 하나, 끝나면 그 사이 바뀐 것을 다시 본다
    const missing = new Set<string>()   // 조회해도 없던 uid(남의 것·삭제) — 다시 조회하지 않음

    const treeUids = () => new Set(Object.values(tiles.value).flatMap(t => t.type === 'terminal' ? [t.uid] : []))

    const reload = () => {
        started = true
        syncing = true
        const uids = [...treeUids()]
        restore(() => uids.length ? lookupProcesses(uids) : Promise.resolve([]), { reload: true })
        .then(list => {
            const found = new Set(list.map(p => p.uid))
            uids.forEach(u => { if (!found.has(u)) { missing.add(u); markGone(u) } })
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
                unknown.forEach(u => { if (!found.has(u)) { missing.add(u); markGone(u) } })
            })
            .catch(err => console.log('[TERMINAL_RESTORE] sync', err))
            .finally(() => {
                syncing = false
                next()
            })
    }

    // 공유 변화(SHARE:CHANGE, 실시간-6b): 트리의 터미널을 다시 조회 — 응답에 없으면 볼 수 없음(타일 비활성화), 다시 보이면 풀고 sync가 새로 등록
    let again = false
    const recheck = () => {
        if (!ready.value || !started) return    // 복원 전이면 reload가 어차피 조회
        if (syncing) {
            again = true
            return
        }
        const uids = [...treeUids()]
        pruneGone(new Set(uids))
        if (!uids.length) return
        syncing = true
        lookupProcesses(uids)
            .then(list => {
                const found = new Set(list.map(p => p.uid))
                list.forEach(refreshAccount)
                uids.forEach(u => {
                    if (found.has(u)) {
                        missing.delete(u)
                        clearGone(u)
                    } else {
                        missing.add(u)
                        markGone(u)
                    }
                })
            })
            .catch(err => console.log('[TERMINAL_RESTORE] recheck', err))
            .finally(() => {
                syncing = false
                next()
            })
    }
    const off = on('SHARE:CHANGE', recheck)
    onBeforeUnmount(off)

    // 조회 중에 트리를 다시 불러왔으면 끝난 뒤 새로고침 복원부터
    const next = () => {
        if (syncing || !ready.value) return
        if (again) {
            again = false
            recheck()
            return
        }
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
