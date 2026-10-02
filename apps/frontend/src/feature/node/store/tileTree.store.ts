import type { TileNode, TileSize } from '@/feature/widget/util/tile.type'
import { useAppDialog } from '@/feature/layout/store/appDialog.store'
import { useTestSocket } from '@/common/websocket/websocket.hook'
import { readTabId } from '@/common/util/tabId.util'
import { inject, onBeforeUnmount, provide, readonly, ref, watch } from 'vue'
import { getTiles, putTiles, type TileTree, type TileTreeResponse } from '../api/tile.api'

const TILE_TREE_STORE_KEY = Symbol('TileTreeStore')

type TileBase = TileNode & { parent: string | null }    // parent = 이 타일을 연 폴더 타일 (노드 트리의 부모 아님)
export type FolderTile = TileBase & { type: 'folder'; nodeId: number | null }  // null = 루트 목록
export type TerminalTile = TileBase & { type: 'terminal'; nodeId: number; uid: string }
export type Tile = FolderTile | TerminalTile

// 트리를 바꾸는 동작. 409·다른 탭 push 때 서버 트리 위에 다시 실행되므로 검사(throw)를 전부 먼저 하고 변경은 뒤에 한다 —
// 변경 도중 throw하면 반쯤 바뀐 트리가 남는다. 대상 타일이 없으면 throw = 그 동작은 버림.
// 멱등이어야 한다: 저장 중 push가 오면 이미 서버에 반영된 동작이 다시 적용된다(REF-node-ui-sync.md "세부 규칙")
type Op = (tree: TileTree) => void

const SAVE_DELAY = 300      // 연속 변경을 묶는 간격
const RETRY_DELAY = 3000    // 네트워크·서버 오류 뒤 재시도 간격
const FAIL_MAX = 3          // 넘으면 다음 변경 때 다시 보냄
const CONFLICT_MAX = 3      // 연속 409가 넘으면 서버 트리로 다시 불러옴

// tiles_는 반응형 proxy라 structuredClone 불가
const clone = <T>(v: T): T => JSON.parse(JSON.stringify(v))

const folderIn = (tiles: Record<string, Tile>, id: string) => {
    const tile = tiles[id]
    if (tile?.type !== 'folder') throw new Error(`[TILE_TREE] not a folder tile: ${id}`)
    return tile
}

const tileIn = (tiles: Record<string, Tile>, id: string) => {
    const tile = tiles[id]
    if (!tile) throw new Error(`[TILE_TREE] no tile: ${id}`)
    return tile
}

// 부모 kids 끝에 붙인다 (= 타일 순서 규칙의 "연 순서"). 재적용마다 새 객체 — 반응형 트리에 넣은 객체가 이후 변경돼도 동작은 그대로
const attachOp = (parentId: string, tile: Tile): Op => ({ tiles }) => {
    if (tiles[tile.id]) throw new Error(`[TILE_TREE] already attached: ${tile.id}`)
    const parent = folderIn(tiles, parentId)
    tiles[tile.id] = clone(tile)
    parent.kids = [...parent.kids, tile.id]
}

// 폴더 이름 클릭 = 타일 안에서 이동 (새 타일 아님)
const navigateOp = (tileId: string, nodeId: number | null): Op => ({ tiles }) => {
    folderIn(tiles, tileId).nodeId = nodeId
}

// 폴더를 닫으면 그 폴더가 연 타일들이 부모 소속이 된다 (⑩ 타일 닫기 규칙):
// 하위 폴더는 닫힌 자리로, 터미널은 adopted로 표시해 kids 끝에 붙인다(= 부모가 직접 실행한 프로세스 뒤).
// 실행 중 터미널의 닫기 금지(kill 먼저)는 UI가 막는다 — 스토어는 process 상태를 모른다
const closeOp = (tileId: string): Op => ({ tiles }) => {
    const tile = tileIn(tiles, tileId)
    if (!tile.parent) return   // 루트는 닫을 수 없음
    const parent = folderIn(tiles, tile.parent)
    const kids = tile.kids.map(k => tiles[k]).filter(k => !!k)
    const folders = kids.filter(k => k.type === 'folder')
    const terminals = kids.filter(k => k.type === 'terminal')
    // 닫힌 폴더 안에서 보이던 순서(직접 실행 -> 넘겨받음) 그대로 넘긴다
    const lifted = [...terminals.filter(k => !k.adopted), ...terminals.filter(k => k.adopted)]

    kids.forEach(k => { k.parent = parent.id })
    lifted.forEach(k => { k.adopted = true })
    parent.kids = [
        ...parent.kids.flatMap(k => k === tileId ? folders.map(f => f.id) : [k]),
        ...lifted.map(k => k.id),
    ]
    delete tiles[tileId]
}

const resizeOp = (tileId: string, size: TileSize): Op => ({ tiles }) => {
    tileIn(tiles, tileId).size = { ...size }
}

// exec 응답의 타일 넣기 — 이미 있으면 그대로(재적용 때 서버 트리에 있음), 부모가 닫혔으면 루트 (서버 appendTerminalTile과 같은 규칙)
const upsertOp = (tile: TerminalTile): Op => tree => {
    if (tree.tiles[tile.id]) return
    const parentId = tile.parent && tree.tiles[tile.parent]?.type === 'folder' ? tile.parent : tree.rootId
    attachOp(parentId, { ...tile, parent: parentId })(tree)
}

// 안전망 타일 크기 — 어느 화면에서든 최소 단위(effectiveSize가 1칸 축의 0.5를 1.0으로 봄), 서버 exec 기본값과 같음
const ORPHAN_SIZE: TileSize = { w: 0.5, h: 0.5 }

export type Orphan = { uid: string; nodeId: number; tileId: string }

// 살아 있는데 타일이 없는 process를 루트에 adopted 터미널로 붙인다(3-b' 결정 ③ 안전망).
// 타일 id가 아니라 uid로 검사 — 재적용 때 다른 탭이 그 사이 exec로 넣은 같은 uid 타일이 서버 트리에 있으면 건너뛴다
const adoptOp = (orphans: Array<Orphan>): Op => ({ rootId, tiles }) => {
    const root = folderIn(tiles, rootId)
    const has = new Set(Object.values(tiles).flatMap(t => t.type === 'terminal' ? [t.uid] : []))
    const add = orphans.filter(o => !has.has(o.uid))
    add.forEach(o => {
        tiles[o.tileId] = { id: o.tileId, type: 'terminal', parent: rootId, kids: [], size: { ...ORPHAN_SIZE }, nodeId: o.nodeId, uid: o.uid, adopted: true }
    })
    root.kids = [...root.kids, ...add.map(o => o.tileId)]
}

// 여는 관계 트리(계정 단위 서버 저장본, ⑪)를 관리한다. 화면 크기(size 계산)와 process API 호출은 호출부 책임.
// 화면 = 서버가 확정한 트리 + 아직 확정 안 된 내 동작(pending) — REF-node-ui-save.md "3-c(스토어) 구조안"
export const provideTileTree = () => {
    const { openDialog } = useAppDialog()
    const { on, status: socketStatus } = useTestSocket()

    const tiles_ = ref<Record<string, Tile>>({})
    const rootId_ = ref('')
    const ready_ = ref(false)               // 서버 트리를 받기 전엔 타일 UI를 그리지 않는다
    const loadError_ = ref<string | null>(null)

    let version = 0                         // 서버가 마지막으로 확정한 version
    let pending: Array<Op> = []             // 확정 안 된 내 동작, 순서대로
    let inFlight = false                    // PUT은 한 번에 하나
    let failures = 0
    let conflicts = 0
    let timer: ReturnType<typeof setTimeout> | undefined
    let gen = 0                             // load/reset마다 증가 — 그 전에 보낸 요청의 응답은 버린다
    let early: TileTreeResponse | null = null    // 불러오는 중 온 push — 받은 트리보다 새것이면 덮어씀

    const tree = (): TileTree => ({ rootId: rootId_.value, tiles: tiles_.value })

    const setTree = (t: TileTree) => {
        rootId_.value = t.rootId
        tiles_.value = t.tiles
    }

    const schedule = (delay: number) => {
        clearTimeout(timer)
        timer = setTimeout(() => void save(), delay)
    }

    // 409·다른 탭 push: 서버 트리 위에 내 동작을 순서대로 다시 적용 — ⑪-2(충돌 처리)
    const rebase = (server: TileTreeResponse) => {
        version = server.version
        let draft = server.tree
        pending = pending.filter(op => {
            const next = clone(draft)
            try {
                op(next)
            } catch {
                return false
            }
            draft = next
            return true
        })
        setTree(draft)
    }

    // 보낸 동작을 객체로 기억했다가 200이면 그것만 확정한다 — 개수로 지우면 요청 중 push 재적용에서 동작이 빠졌을 때 어긋남
    const save = async (keepalive = false) => {
        clearTimeout(timer)
        if (inFlight || !pending.length || !ready_.value) return
        inFlight = true
        const my = gen
        const sent = new Set(pending)
        const res = await putTiles(clone(tree()), version, keepalive).catch(err => {
            console.log('[TILE_TREE] save', err)
            return null
        })
        if (my !== gen) return
        inFlight = false

        if (!res) {
            if (++failures <= FAIL_MAX) schedule(RETRY_DELAY)
            return
        }
        failures = 0
        if (res.ok) {
            version = Math.max(version, res.version)    // 요청 중 더 새 push를 이미 받았으면 되돌리지 않음
            pending = pending.filter(op => !sent.has(op))
            conflicts = 0
        } else if (++conflicts > CONFLICT_MAX) {
            pending = []
            conflicts = 0
            version = res.version
            setTree(res.tree)
            openDialog({ type: 'warning', content: '다른 탭의 변경과 계속 겹쳐 타일 배치를 서버 저장본으로 다시 불러왔습니다.' })
        } else if (res.version > version) {
            rebase(res)
        }
        if (!pending.length) return
        if (document.visibilityState === 'hidden') void save(true)
        else schedule(res.ok ? SAVE_DELAY : 0)
    }

    // 화면에 바로 적용하고 저장 예약. 화면이 서버 트리로 바뀐 직후의 낡은 동작이면 버린다
    const run = (op: Op) => {
        try {
            op(tree())
        } catch (err) {
            console.log('[TILE_TREE] op', err)
            return
        }
        pending.push(op)
        schedule(SAVE_DELAY)
    }

    const clearSaving = () => {
        clearTimeout(timer)
        pending = []
        inFlight = false
        failures = 0
        conflicts = 0
    }

    const reset = () => {
        gen++
        clearSaving()
        version = 0
        setTree({ rootId: '', tiles: {} })
        ready_.value = false
        loadError_.value = null
    }

    const load = async () => {
        reset()
        early = null
        const my = gen
        try {
            const res = await getTiles()
            if (my !== gen) return
            const stashed = early as TileTreeResponse | null    // await 사이 push 핸들러가 채움(TS는 위의 null 대입만 봄)
            const latest = stashed && stashed.version > res.version ? stashed : res
            early = null
            version = latest.version
            setTree(latest.tree)
            ready_.value = true
        } catch (err) {
            if (my !== gen) return
            loadError_.value = (err as { message?: string })?.message || '타일 배치를 불러오지 못했습니다'
        }
    }

    // 다른 탭의 저장 — 4 O(계정 동기화) 화면. 내 탭 것은 내 응답이 처리하므로 거르고(4②), 옛 version도 거른다
    const offs = [on<TileTreeResponse & { tabId: string }>('TILES:UPDATE', ev => {
        if (ev.tabId && ev.tabId === readTabId()) return
        if (!ready_.value) {
            if (!early || ev.version > early.version) early = { tree: ev.tree, version: ev.version }
            return
        }
        if (ev.version > version) rebase(ev)
    })]

    // 끊긴 동안 놓친 push는 다시 받아서 맞춘다
    watch(socketStatus, s => {
        if (s !== 'CONNECTED' || !ready_.value) return
        const my = gen
        getTiles()
            .then(res => { if (my === gen && res.version > version) rebase(res) })
            .catch(err => console.log('[TILE_TREE] resync', err))
    })

    // 닫기 직전 저장: 숨겨지는 순간 묶음을 기다리지 않고 keepalive로 보낸다. pagehide는 모바일에서 OS가 탭을 죽이면 오지 않는다
    const onVisibility = () => {
        if (document.visibilityState === 'hidden') void save(true)
    }
    document.addEventListener('visibilitychange', onVisibility)
    onBeforeUnmount(() => {
        document.removeEventListener('visibilitychange', onVisibility)
        offs.forEach(off => off())
        clearTimeout(timer)
    })

    // 새 타일 id는 동작을 만들 때 정한다 — 재적용해도 같은 id라 화면·reveal이 흔들리지 않는다
    const openFolder = (parentId: string, nodeId: number | null, size: TileSize) => {
        const tile: FolderTile = { id: crypto.randomUUID(), type: 'folder', parent: parentId, kids: [], size, nodeId }
        run(attachOp(parentId, tile))
        return tile.id
    }

    const navigate = (tileId: string, nodeId: number | null) => run(navigateOp(tileId, nodeId))

    const close = (tileId: string) => run(closeOp(tileId))

    const resize = (tileId: string, size: TileSize) => run(resizeOp(tileId, size))

    // exec 응답의 터미널 타일(서버가 이미 트리에 넣음, ⑪-4). 서버 version이 바로 다음이고 저장할 게 없으면 PUT 없이 반영,
    // 아니면 동작으로 쌓는다 — 다음 PUT이 409를 받고 서버 트리(타일 있음) 위에 재적용되며 이 동작은 건너뛴다
    const applyServerTile = (tile: TerminalTile, tileVersion: number) => {
        if (tileVersion <= version) return tile.id      // 이미 받은 서버 트리에 반영돼 있음
        if (tileVersion === version + 1 && !pending.length && !inFlight) {
            upsertOp(tile)(tree())
            version = tileVersion
        } else {
            run(upsertOp(tile))
        }
        return tile.id
    }

    // procs = 살아 있는 process. 타일이 없는 것만 붙이고 저장
    const adoptOrphans = (procs: Array<{ uid: string; nodeId: number | null }>) => {
        const has = new Set(Object.values(tiles_.value).flatMap(t => t.type === 'terminal' ? [t.uid] : []))
        const orphans = procs.flatMap(p => p.nodeId !== null && !has.has(p.uid)
            ? [{ uid: p.uid, nodeId: p.nodeId, tileId: crypto.randomUUID() }]
            : [])
        if (orphans.length) run(adoptOp(orphans))
    }

    const ctx = {
        tiles: readonly(tiles_),
        rootId: readonly(rootId_),
        ready: readonly(ready_),
        loadError: readonly(loadError_),
        load,
        reset,
        openFolder,
        navigate,
        close,
        resize,
        applyServerTile,
        adoptOrphans,
    }

    provide(TILE_TREE_STORE_KEY, ctx)

    return ctx
}

export const useTileTree = () => {
    const context = inject<ReturnType<typeof provideTileTree>>(TILE_TREE_STORE_KEY)
    if (!context) throw new Error('TileTreeStore is not provided')
    return context
}
