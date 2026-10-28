import type { TileNode, TileSize } from '@/feature/widget/util/tile.type'
import { canParent, holdsProc } from '@/feature/widget/util/tileKind.util'
import { useAppDialog } from '@/feature/layout/store/appDialog.store'
import { useTestSocket } from '@/common/websocket/websocket.hook'
import { readTabId } from '@/common/util/tabId.util'
import { computed, inject, onBeforeUnmount, provide, readonly, ref, watch } from 'vue'
import { getTiles, putTiles, type TilePin, type TileTree, type TileTreeResponse } from '../api/tile.api'
import { createBoard, moveBoardFolder, type BoardChangeResult } from '../api/board.api'

const TILE_TREE_STORE_KEY = Symbol('TileTreeStore')

type TileBase = TileNode & { parent: string | null }    // parent = 이 타일을 연 폴더 타일 (노드 트리의 부모 아님)
type NavBase = { back?: ReadonlyArray<string>; fwd?: ReadonlyArray<string> }   // 이동 기록(folder = 노드 id, "" = 루트 목록 / device = 경로) — 뒤로-1(타일 안 이동 기록)
export type FolderTile = TileBase & NavBase & { type: 'folder'; nodeId: number | null; uid?: string; bound?: boolean }  // null = 루트 목록, uid = 서버 발급(새 타일은 저장 뒤 생김), 공유로 받은 타일은 공유한 타일과 같음, bound = 수동 타일 고정(서버만 씀, 표시는 pins)
export type TerminalTile = TileBase & { type: 'terminal'; nodeId: number | null; uid: string; shortcutsOff?: true }   // nodeId null = 장비 파일 실행(폴더-4(실행)), shortcutsOff = active인 동안 앱 단축키 끔(이동-3(타일별 단축키 끔))
export type DeviceTile = TileBase & NavBase & { type: 'device'; nodeId: number; instance: string; path?: string }    // nodeId = 연 폴더(Q15(장비 디렉토리 id)), instance = main#sub, path 없음 = 최상위
export type ViewerTile = TileBase & { type: 'viewer'; nodeId: number; instance: string; path: string }    // 보기 타일 — 장비 타일과 같은 필드, path = 파일(파트(보기 타일), REF-file-view.md)
export type Tile = FolderTile | TerminalTile | DeviceTile | ViewerTile

// 능력별 타일 — 가드가 TILE_KINDS를 믿으므로 종류를 더하면 이 두 타입도 같이 (REF-tile-kinds.md)
export type ParentTile = FolderTile | DeviceTile     // 능력-2(자식 가능)
export type ProcTile = TerminalTile     // 능력-3(process 담음)
export const isParentTile = (t: Tile | undefined): t is ParentTile => !!t && canParent(t)
export const isProcTile = (t: Tile | undefined): t is ProcTile => !!t && holdsProc(t)
export const procUid = (t: Tile | undefined) => isProcTile(t) ? t.uid : undefined

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

const deviceIn = (tiles: Record<string, Tile>, id: string) => {
    const tile = tiles[id]
    if (tile?.type !== 'device') throw new Error(`[TILE_TREE] not a device tile: ${id}`)
    return tile
}

const parentIn = (tiles: Record<string, Tile>, id: string) => {
    const tile = tiles[id]
    if (!isParentTile(tile)) throw new Error(`[TILE_TREE] not a parent tile: ${id}`)
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
    const parent = parentIn(tiles, parentId)
    tiles[tile.id] = clone(tile)
    parent.kids = [...parent.kids, tile.id]
}

// ---- 타일 안 위치 + 이동 기록(뒤로-1) ----
const HIST_MAX = 20     // 뒤로·앞으로 합쳐서 — Q16(기록 보관 위치), 서버 tileHistMax와 같게
export type NavTile = FolderTile | DeviceTile
export const isNavTile = (t: Tile | undefined): t is NavTile => t?.type === 'folder' || t?.type === 'device'

// 위치를 문자열로 — 기록에 쌓는 값
export const locOf = (t: NavTile) => t.type === 'folder' ? (t.nodeId === null ? '' : String(t.nodeId)) : (t.path ?? '')

const setLoc = (t: NavTile, loc: string) => {
    if (t.type === 'device') {
        if (loc) t.path = loc
        else delete t.path
        return
    }
    const nodeId = loc === '' ? null : Number(loc)
    if (t.nodeId !== nodeId) delete t.bound   // 다른 폴더로 가면 서버도 고정을 푼다 — BD3(저장 위치). 내 탭은 TILES:UPDATE를 거르므로 여기서 맞춤
    t.nodeId = nodeId
}

const setHist = (t: NavTile, back: string[], fwd: string[]) => {
    if (back.length) t.back = back
    else delete t.back
    if (fwd.length) t.fwd = fwd
    else delete t.fwd
}

// 새 위치로 가면 지금 위치를 뒤로 기록에 쌓고 앞으로 기록은 비운다. 같은 위치면 아무것도 안 함(재적용 멱등)
const moveOp = (tileId: string, loc: string): Op => ({ tiles }) => {
    const t = tiles[tileId]
    if (!isNavTile(t)) throw new Error(`[TILE_TREE] not a nav tile: ${tileId}`)
    const cur = locOf(t)
    if (cur === loc) return
    setHist(t, [...(t.back ?? []), cur].slice(-HIST_MAX), [])
    setLoc(t, loc)
}

// 뒤로(-1)·앞으로(1). from = 누른 때의 위치 — 재적용 때 이미 옮겨졌으면(위치가 다름) 다시 옮기지 않는다
const histOp = (tileId: string, dir: -1 | 1, from: string): Op => ({ tiles }) => {
    const t = tiles[tileId]
    if (!isNavTile(t) || locOf(t) !== from) return
    const back = [...(t.back ?? [])]
    const fwd = [...(t.fwd ?? [])]
    const to = dir < 0 ? back.pop() : fwd.shift()
    if (to === undefined) return
    if (dir < 0) fwd.unshift(from)
    else back.push(from)
    setHist(t, back, fwd)
    setLoc(t, to)
}

// 폴더 이름 클릭 = 타일 안에서 이동 (새 타일 아님)
const navigateOp = (tileId: string, nodeId: number | null): Op => tree => {
    folderIn(tree.tiles, tileId)
    moveOp(tileId, nodeId === null ? '' : String(nodeId))(tree)
}

// 장비 타일 안 이동 — 폴더-2(장비 타일)
const navigateDeviceOp = (tileId: string, path: string): Op => tree => {
    deviceIn(tree.tiles, tileId)
    moveOp(tileId, path)(tree)
}

// 폴더를 닫으면 그 폴더가 연 타일들이 부모 소속이 된다 (⑩ 타일 닫기 규칙):
// process 타일(터미널)은 adopted로 표시해 kids 끝에 붙이고(= 부모가 직접 실행한 프로세스 뒤), 나머지(하위 폴더)는 닫힌 자리로.
// 실행 중 터미널의 닫기 금지(kill 먼저)는 UI가 막는다 — 스토어는 process 상태를 모른다
const closeOp = (tileId: string): Op => ({ tiles }) => {
    const tile = tileIn(tiles, tileId)
    if (!tile.parent) return   // 루트는 닫을 수 없음
    const parent = parentIn(tiles, tile.parent)
    const kids = tile.kids.map(k => tiles[k]).filter(k => !!k)
    const folders = kids.filter(k => !isProcTile(k))
    const terminals = kids.filter(isProcTile)
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

// process 타일만 최소화한다. 다시 열면 트리 자리·크기 그대로 배치에 돌아온다
const minimizeOp = (tileId: string, on: boolean): Op => ({ tiles }) => {
    const tile = tileIn(tiles, tileId)
    if (!isProcTile(tile)) throw new Error(`[TILE_TREE] not a process tile: ${tileId}`)
    if (on) tile.minimized = true
    else delete tile.minimized
}

const shortcutsOffOp = (tileId: string, on: boolean): Op => ({ tiles }) => {
    const tile = tileIn(tiles, tileId)
    if (!isProcTile(tile)) throw new Error(`[TILE_TREE] not a process tile: ${tileId}`)
    if (on) tile.shortcutsOff = true
    else delete tile.shortcutsOff
}

// exec 응답의 타일 넣기 — 이미 있으면 그대로(재적용 때 서버 트리에 있음), 부모가 닫혔으면 루트 (서버 appendTerminalTile과 같은 규칙)
const upsertOp = (tile: TerminalTile): Op => tree => {
    if (tree.tiles[tile.id]) return
    const parentId = tile.parent && isParentTile(tree.tiles[tile.parent]) ? tile.parent : tree.rootId
    attachOp(parentId, { ...tile, parent: parentId })(tree)
}

const wait = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

// 여는 관계 트리(보드 하나의 서버 저장본, ⑪)를 관리한다. 보드를 바꾸면 load(hash)로 다시 불러온다 — 인스턴스는 하나.
// 화면 크기(size 계산)와 process API 호출은 호출부 책임.
// 고아 안전망은 서버가 GET 때 돈다(WS3(고아 안전망 위치)) — 터미널 타일은 서버만 넣으므로 여기서 새로 만들지 않는다(WS5(터미널 타일 추가 주체))
// 화면 = 서버가 확정한 트리 + 아직 확정 안 된 내 동작(pending) — REF-node-ui-save.md "3-c(스토어) 구조안"
export const provideTileTree = () => {
    const { openDialog } = useAppDialog()
    const { on, status: socketStatus } = useTestSocket()

    const tiles_ = ref<Record<string, Tile>>({})
    const rootId_ = ref('')
    const ready_ = ref(false)               // 서버 트리를 받기 전엔 타일 UI를 그리지 않는다
    const loadError_ = ref<string | null>(null)
    const hash_ = ref('')                   // 불러온 보드의 실제 hash(기본 보드 포함) — TILES:UPDATE 거르기용
    let reqHash = ''                        // 요청한 보드(라우트), 빈 값 = 기본

    let version = 0                         // 서버가 마지막으로 확정한 version
    let pending: Array<Op> = []             // 확정 안 된 내 동작, 순서대로
    let inFlight = false                    // PUT은 한 번에 하나
    let failures = 0
    let conflicts = 0
    let timer: ReturnType<typeof setTimeout> | undefined
    let gen = 0                             // load/reset마다 증가 — 그 전에 보낸 요청의 응답은 버린다
    const early = new Map<string, TileTreeResponse>()   // 불러오는 중 온 push(보드 hash별) — 받은 트리보다 새것이면 덮어씀
    const pins_ = ref<TilePin[]>([])        // 계정 전체 고정 지도(보드와 무관, 개편-1(고정 지도)) — 보드를 바꿔도 비우지 않음
    let pinsPushed = false                  // GET 도중 TILES:PINS가 왔으면 GET의 (더 옛) 지도로 덮지 않는다

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
        const res = await putTiles(reqHash, clone(tree()), version, keepalive).catch(err => {
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
        hash_.value = ''
        setTree({ rootId: '', tiles: {} })
        ready_.value = false
        loadError_.value = null
    }

    const load = async (hash = reqHash) => {
        reset()
        reqHash = hash
        early.clear()
        pinsPushed = false
        const my = gen
        try {
            const res = await getTiles(hash)
            if (my !== gen) return
            if (!pinsPushed && res.pins) pins_.value = res.pins
            const stashed = early.get(res.hash)     // await 사이 push 핸들러가 채움
            const latest = stashed && stashed.version > res.version ? stashed : res
            early.clear()
            hash_.value = res.hash
            version = latest.version
            setTree(latest.tree)
            ready_.value = true
        } catch (err) {
            if (my !== gen) return
            loadError_.value = (err as { message?: string })?.message || '타일 배치를 불러오지 못했습니다'
        }
    }

    // 다른 탭의 저장 — 4 O(계정 동기화) 화면. 내 탭 것은 내 응답이 처리하므로 거르고(4②), 다른 보드 것(WS6(TILES 이벤트 구분))·옛 version도 거른다.
    // 서버 동작(exec 제외 분리·이동·삭제·고아 안전망)은 tabId가 비어 있어 내 탭도 받는다
    const offs = [on<TileTreeResponse & { tabId: string }>('TILES:UPDATE', ev => {
        if (ev.tabId && ev.tabId === readTabId()) return
        if (!ready_.value) {
            const prev = early.get(ev.hash)
            if (!prev || ev.version > prev.version) early.set(ev.hash, { hash: ev.hash, tree: ev.tree, version: ev.version })
            return
        }
        if (ev.hash === hash_.value && ev.version > version) rebase(ev)
    }), on<{ pins: TilePin[] }>('TILES:PINS', ev => {
        pins_.value = ev.pins
        pinsPushed = true
    })]

    // 끊긴 동안 놓친 push는 다시 받아서 맞춘다
    watch(socketStatus, s => {
        if (s !== 'CONNECTED' || !ready_.value) return
        const my = gen
        getTiles(reqHash)
            .then(res => {
                if (my !== gen) return
                if (res.pins) pins_.value = res.pins
                if (res.hash === hash_.value && res.version > version) rebase(res)
            })
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

    // 장비 타일 — 연 폴더의 자식, 최상위부터(Q6(처음 여는 위치)). 폴더-3(여는 곳)
    const openDevice = (parentId: string, nodeId: number, instance: string, size: TileSize) => {
        const tile: DeviceTile = { id: crypto.randomUUID(), type: 'device', parent: parentId, kids: [], size, nodeId, instance }
        run(attachOp(parentId, tile))
        return tile.id
    }

    // 보기 타일 — 연 장비 타일의 자식
    const openViewer = (parentId: string, nodeId: number, instance: string, path: string, size: TileSize) => {
        const tile: ViewerTile = { id: crypto.randomUUID(), type: 'viewer', parent: parentId, kids: [], size, nodeId, instance, path }
        run(attachOp(parentId, tile))
        return tile.id
    }

    const navigate = (tileId: string, nodeId: number | null) => run(navigateOp(tileId, nodeId))

    const navigateDevice = (tileId: string, path: string) => run(navigateDeviceOp(tileId, path))

    // 뒤로·앞으로 — 기록이 없으면 아무것도 안 함
    const goHistory = (tileId: string, dir: -1 | 1) => {
        const t = tiles_.value[tileId]
        if (!isNavTile(t) || !(dir < 0 ? t.back : t.fwd)?.length) return
        run(histOp(tileId, dir, locOf(t)))
    }

    const close = (tileId: string) => run(closeOp(tileId))

    const resize = (tileId: string, size: TileSize) => run(resizeOp(tileId, size))

    const setMinimized = (tileId: string, on: boolean) => run(minimizeOp(tileId, on))

    const setShortcutsOff = (tileId: string, on: boolean) => run(shortcutsOffOp(tileId, on))

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

    // 저장 대기분을 먼저 보낸다 — 서버 동작 API는 출발 보드 version을 검사하므로(WS2(변경 방식)). 저장이 실패하면 진행하지 않음
    const flush = async () => {
        const my = gen
        const until = Date.now() + 5000
        while ((pending.length || inFlight) && my === gen && Date.now() < until) {
            if (inFlight) {
                await wait(100)
                continue
            }
            const before = failures
            await save()
            if (failures > before) break
        }
        if (my !== gen) throw { message: '보드가 바뀌어 취소했습니다' }
        if (pending.length || inFlight) throw { message: '타일 배치를 저장하지 못해 진행하지 않았습니다' }
    }

    // 서버가 트리를 바꾸는 동작(분리·이동): 대기분 저장 -> 호출 -> 409면 서버 트리로 맞추고 다시. 성공하면 응답 트리를 바로 반영
    // (같은 내용이 TILES:UPDATE로도 오지만 version 비교로 한 번만 적용된다)
    const serverChange = async (call: (version: number) => Promise<BoardChangeResult>) => {
        if (!ready_.value) throw { message: '타일 배치를 아직 불러오지 못했습니다' }
        const my = gen
        for (let i = 0; i <= CONFLICT_MAX; i++) {
            await flush()
            const res = await call(version)
            if (my !== gen) throw { message: '보드가 바뀌어 취소했습니다' }
            if (res.ok) {
                if (res.hash === hash_.value && res.tree && res.version !== undefined && res.version > version) {
                    rebase({ hash: res.hash, tree: res.tree, version: res.version })
                }
                return res
            }
            if (res.version > version) rebase(res)
        }
        throw { message: '다른 탭의 변경과 계속 겹쳐 진행하지 못했습니다' }
    }

    // 폴더 타일을 하위째 새 보드로(기본 보드에서만, W1(분리 범위)) — 만든 보드를 돌려준다
    const detachToBoard = async (folderTileId: string, name: string) => {
        const res = await serverChange(v => createBoard({ from: hash_.value, folderTileId, name, version: v }))
        if (!res.ok || !res.workspace) throw { message: '보드를 만들지 못했습니다' }
        return res.workspace
    }

    // 폴더 타일을 하위째 다른 보드 루트 끝으로(W3(이동 범위))
    const moveToBoard = (folderTileId: string, to: string) => serverChange(v => moveBoardFolder(hash_.value, { folderTileId, to, version: v }))

    // 노드별 고정 타일(타일 2개 이상인 노드만) — 핀 표시·"여기로 타일 고정" 메뉴(개편-7(고정 표시 UI))
    const pinByNode = computed(() => new Map(pins_.value.map(p => [p.nodeId, p])))

    const ctx = {
        tiles: readonly(tiles_),
        pinByNode,
        rootId: readonly(rootId_),
        ready: readonly(ready_),
        loadError: readonly(loadError_),
        hash: readonly(hash_),
        load,
        reset,
        openFolder,
        openDevice,
        openViewer,
        navigate,
        navigateDevice,
        goHistory,
        close,
        resize,
        setMinimized,
        setShortcutsOff,
        applyServerTile,
        flush,
        detachToBoard,
        moveToBoard,
    }

    provide(TILE_TREE_STORE_KEY, ctx)

    return ctx
}

export const useTileTree = () => {
    const context = inject<ReturnType<typeof provideTileTree>>(TILE_TREE_STORE_KEY)
    if (!context) throw new Error('TileTreeStore is not provided')
    return context
}
