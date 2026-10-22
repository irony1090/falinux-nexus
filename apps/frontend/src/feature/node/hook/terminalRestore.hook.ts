import { tabReady } from '@/common/util/tabId.util'
import { lookupProcesses } from '@/feature/process/api/process.api'
import { type provideProcessTerm } from '@/feature/process/store/processTerm.store'
import { useTestSocket } from '@/common/websocket/websocket.hook'
import { onBeforeUnmount, watch, type Ref } from 'vue'
import { isProcTile, type provideTileTree } from '../store/tileTree.store'

// 트리의 터미널 타일과 processTerm 등록을 맞춘다 — ⑪ 3-d(터미널 복원) + 4-c(터미널 등록 일원화).
// 새로고침 복원·다른 탭 push·409 재적용으로 트리에 처음 보이는 터미널이 전부 이 길로 등록된다(REF-node-ui-sync.md).
// 순서 의존:
// 1) 탭 id를 받은 뒤 — 받기 전 요청은 sizeOwner가 전부 false
// 2) 트리를 받은 뒤 process 조회 — 트리에 있는 uid만 조회한다. 그 뒤 push로 트리에 생긴 터미널은 sync가 등록.
//    살아 있는데 타일이 없는 process는 서버가 GET /tiles 때 기본 보드에 붙인다(WS3(고아 안전망 위치)) — 프론트 안전망 없음
// 두 스토어를 provide하는 곳(pages/index.vue)에서 부르므로 inject 대신 인자로 받는다 — provide한 컴포넌트 자신은 inject 불가
// others = 전체보기에 그리는 다른 보드의 터미널 uid(미리보기-2(터미널 등록)) — 기준에 넣지 않으면 prune이 그 보드의 끝난 터미널 xterm을 지운다.
// 전체보기를 닫으면 others 때문에 새로 등록한 xterm만 지운다(중간안, REF-board-preview.md) — 보드 전환으로 있던 것은 남김
export const useTerminalRestore = (
    { ready, tiles }: ReturnType<typeof provideTileTree>,
    { procs, restored, restore, prune, dispose, markGone, clearGone, refreshAccount, pruneGone }: ReturnType<typeof provideProcessTerm>,
    others?: Readonly<Ref<ReadonlySet<string>>>,
) => {
    const { on } = useTestSocket()
    let started = false             // 새로고침 복원 — 트리 load마다 1번
    let failed = false              // 복원 실패 — 소켓 재연결(tabReady)·트리 재로드 때만 다시(바로 재시도하면 서버가 죽었을 때 끝없이 돈다)
    let syncing = false             // 조회는 한 번에 하나, 끝나면 그 사이 바뀐 것을 다시 본다
    const missing = new Set<string>()   // 조회해도 없던 uid(남의 것·삭제) — 다시 조회하지 않음
    const viaOthers = new Set<string>() // others 때문에 이 훅이 새로 등록한 uid — release가 지울 후보

    const tileUids = () => new Set(Object.values(tiles.value).flatMap(t => isProcTile(t) ? [t.uid] : []))
    const treeUids = () => new Set([...tileUids(), ...(others?.value ?? [])])

    // 조회 직전에 고른다 — 아직 등록 안 됐고 지금 보드엔 없이 others에만 있는 uid
    const othersOnly = (uids: Iterable<string>) => {
        const mine = tileUids()
        return [...uids].filter(u => !procs.value[u] && !mine.has(u) && !!others?.value.has(u))
    }
    const remember = (picked: string[]) => picked.forEach(u => { if (procs.value[u]) viaOthers.add(u) })

    // 지금 보드로 들어온 uid는 그 보드 소속이 되어 후보에서 빼고, others에서도 빠진 uid는 xterm을 지운다.
    // ready일 때만 — 다른 보드 타일을 눌러 이동하면 새 트리를 받기 전엔 tiles가 비어 이동할 터미널까지 지우게 된다
    const release = () => {
        if (!ready.value || !viaOthers.size) return
        const mine = tileUids()
        viaOthers.forEach(u => {
            if (mine.has(u)) viaOthers.delete(u)
            else if (!others?.value.has(u)) {
                dispose(u)
                viaOthers.delete(u)
            }
        })
    }

    const reload = () => {
        started = true
        syncing = true
        const uids = [...treeUids()]
        const picked = othersOnly(uids)
        restore(() => uids.length ? lookupProcesses(uids) : Promise.resolve([]), { reload: true })
        .then(list => {
            remember(picked)
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
        const picked = othersOnly(unknown)
        restore(() => lookupProcesses(unknown), { reload: false })
            .then(list => {
                remember(picked)
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
        release()
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
