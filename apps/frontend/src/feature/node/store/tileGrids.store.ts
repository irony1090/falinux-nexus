import type { Area, Capacity, GridPlan } from '@/feature/widget/util/tile.type'
import { flatten, pack } from '@/feature/widget/util/tileOrder.util'
import { capacity, colScroll, effectiveSize, minSize, revealCol, stripGeometry, stripPos, stripView } from '@/feature/widget/util/tileSplit.util'
import { readUserPref, writeUserPref } from '@/common/util/userPrefs.util'
import { providePagerBoard } from '@/feature/widget/store/pagerBoard.store'
import { useProcessTerm } from '@/feature/process/store/processTerm.store'
import { useTestSocket } from '@/common/websocket/websocket.hook'
import { tabReady } from '@/common/util/tabId.util'
import { computed, inject, nextTick, onBeforeUnmount, provide, ref, shallowRef, watch } from 'vue'
import type { BoardItem } from '../api/board.api'
import { useBoardTrees } from './boardTrees.store'
import { isProcTile, useTileTree, type Tile } from './tileTree.store'

const TILE_GRIDS_STORE_KEY = Symbol('TileGridsStore')

// strip padding = Grid 간격 = 타일 간격 (1칼럼 Grid 폭 계산이 이 셋이 같다고 가정)
export const TILE_GAP = 8
// 전체보기 축소 비율(%) — ⑫(전체보기 세부) 사용자 조절
export const OV_PERCENT = { init: 30, min: 15, max: 60, step: 5 } as const
const OV_PERCENT_KEY = 'tile.overviewPercent'

// 저장값이 숫자가 아니거나 범위·단위를 벗어나면 기본값
const readOvPercent = () => {
    const v = readUserPref<unknown>(OV_PERCENT_KEY, OV_PERCENT.init)
    const { init, min, max, step } = OV_PERCENT
    return typeof v === 'number' && v >= min && v <= max && (v - min) % step === 0 ? v : init
}

// 배치 = 최소화 제외(최소화한 터미널은 위치 번호가 없다) -> 같은 Grid 규칙으로 채움. 지금 보드·전체보기의 다른 보드가 같이 씀(미리보기-3(Grid 파생))
const layoutOf = (tiles: Readonly<Record<string, Tile>>, rootId: string, cap: Capacity) => {
    const order = flatten(tiles, rootId).filter(o => !o.tile.minimized)
    return { order, grids: pack(order.map(o => o.tile), t => effectiveSize(t.size, cap)) }
}

export type BoardGrids = { board: BoardItem; grids: Array<GridPlan<Tile>>; posOf: Map<string, number>; loading: boolean; error?: string }

// 트리 + 측정한 Grid 영역 -> Grid 목록 + 가로 스크롤 위치 -> 헤더 내비 — REF-node-ui-impl.md "④ Grid 파생"
// 측정·스크롤 기록은 TileStrip, 여기는 등록된 strip 요소에 scrollTo만 한다
export const provideTileGrids = () => {
    const { tiles, rootId, ready, hash } = useTileTree()
    const boardTrees = useBoardTrees()
    const { focus: focusTerm } = useProcessTerm()
    const { emit } = useTestSocket()

    const area = ref<Area | null>(null)

    const cap = computed(() => capacity(area.value ?? { w: 0, h: 0 }))
    const flat = computed(() => flatten(tiles.value, rootId.value))
    const layout = computed(() => layoutOf(tiles.value, rootId.value, cap.value))
    const order = computed(() => layout.value.order)
    // 측정 전엔 그리지 않는다 — 기본 cap을 가정해 그리면 측정 직후 배치가 튄다
    const grids = computed(() => area.value ? layout.value.grids : [])
    const newSize = computed(() => minSize(cap.value))

    // 타일 순번(1부터)
    const posOf = computed(() => new Map(order.value.map((o, i) => [o.tile.id, i + 1])))

    // 같은 노드를 여러 번 실행했으면 순서대로 몇 번째인지, 한 번뿐이면 없음. 최소화 포함 — 최소화로 번호가 바뀌지 않게
    const instanceOf = computed(() => {
        const byNode = new Map<number, Array<string>>()
        flat.value.forEach(({ tile }) => {
            if (!isProcTile(tile) || tile.nodeId === null) return   // 장비 파일 실행은 번호 없음
            byNode.set(tile.nodeId, [...(byNode.get(tile.nodeId) ?? []), tile.id])
        })
        const rst = new Map<string, number>()
        byNode.forEach(ids => {
            if (ids.length > 1) ids.forEach((id, i) => rst.set(id, i + 1))
        })
        return rst
    })

    // 루트 타일 터미널 탭 목록 — 최소화 먼저, 그다음 배치 순서
    const terminals = computed(() => {
        const all = flat.value.flatMap(({ tile }) => isProcTile(tile) ? [tile] : [])
        return [...all.filter(t => t.minimized), ...all.filter(t => !t.minimized)]
    })

    // 루트 타일 탭. 보드를 불러올 때 최소화가 있으면 터미널 탭 / 이 탭에서 최소화하면 터미널 탭으로(TileFrame) — REF-process-share-realtime-ui.md
    const rootTab = ref<'files' | 'terminals'>('files')
    watch(ready, r => {
        if (r) rootTab.value = terminals.value.some(t => t.minimized) ? 'terminals' : 'files'
    }, { immediate: true })

    // 가로 스크롤 — TileStrip이 기록, TileNav가 읽음
    const stripEl = shallowRef<HTMLElement>()
    const scrollX = ref(0)
    const viewW = ref(0)

    const geometry = computed(() => stripGeometry(grids.value.map(g => g.cols), area.value?.w ?? 0, cap.value, TILE_GAP))
    const view = computed(() => stripView(geometry.value, scrollX.value, viewW.value, TILE_GAP))

    // x = 칼럼/Grid 시작. 스냅 위치와 맞추려고 scroll-padding(gap)만큼 뺀다
    const scrollTo = (x: number, behavior: ScrollBehavior = 'smooth') => stripEl.value?.scrollTo({ left: x - TILE_GAP, behavior })

    const goGrid = (index: number) => {
        const s = geometry.value.grids[index]
        if (s) scrollTo(s.x)
    }
    const goColumn = (delta: -1 | 1) => {
        const c = geometry.value.columns[view.value.col + delta]
        if (c) scrollTo(c.x)
    }

    // ---- 인디케이터(GridPager) — REF-board-pager.md ----
    // 도트 위치 = Grid 단위 실수(2.5 = Grid 2 오른쪽 칼럼부터)
    const pos = computed(() => stripPos(geometry.value, scrollX.value, TILE_GAP))
    // 화면에 절반 이상 보이는 칼럼의 Grid 범위 — 인디케이터 숫자·Grid 정보 토스트(푸터-6)
    const range = computed(() => {
        const cols = geometry.value.columns
        const lo = cols[view.value.from]?.grid ?? 0
        return { lo, hi: cols[view.value.to]?.grid ?? lo }
    })
    // 화면 왼쪽 끝이 될 수 있는 마지막 칼럼 — 그 뒤 칼럼은 스크롤 끝에 막힌다
    const maxCol = computed(() => {
        const { grids: gs, columns } = geometry.value
        const last = gs[gs.length - 1]
        if (!last) return 0
        const maxScroll = last.x + last.w + TILE_GAP - viewW.value
        return columns.reduce((k, c, i) => c.x - TILE_GAP <= maxScroll + 1 ? i : k, 0)
    })
    // 인디케이터 드래그 중엔 스냅을 끈다(TileStrip) — 켜 두면 scrollLeft를 줄 때마다 칼럼으로 튄다
    const scrubbing = ref(false)
    const goCol = (col: number, behavior: ScrollBehavior = 'smooth') => {
        const c = geometry.value.columns[col]
        if (c) scrollTo(c.x, behavior)
    }
    // col = 칼럼 번호(실수), 손가락을 바로 따라가게 애니메이션 없이
    const scrubTo = (col: number) => {
        if (stripEl.value) stripEl.value.scrollLeft = colScroll(geometry.value, col, TILE_GAP)
    }

    // 타일이 화면 밖이면 인디케이터와 같은 "타일 위치 규칙"(revealCol)으로 이동, 보이면 그대로 — 이동-1(타일 위치 규칙 통일).
    // nextTick 필수: 트리 변경 직후엔 늘어난 Grid가 아직 DOM에 없어 scrollTo가 옛 스크롤 폭에서 잘린다.
    // 스크롤 위치도 scrollX(스크롤 이벤트로 늦게 갱신)가 아니라 요소에서 직접 읽는다 — 폭이 줄며 잘린 직후일 수 있음
    const reveal = async (tileId: string) => {
        await nextTick()
        if (overview.value) return
        const el = stripEl.value
        const range = colRange(tileId)
        if (!el || !range) return
        const cur = stripView(geometry.value, el.scrollLeft, el.clientWidth, TILE_GAP).col
        const v = revealCol(range.l, range.r, cur, cap.value.x, maxCol.value)
        if (v !== cur) goCol(v)
    }

    // 타일이 차지하는 칼럼 — gi = Grid 번호
    const columnsOf = (tileId: string) => {
        const gi = grids.value.findIndex(g => g.items.some(it => it.tile.id === tileId))
        const it = grids.value[gi]?.items.find(it => it.tile.id === tileId)
        if (!it) return undefined
        const cols = geometry.value.columns.filter(c => c.grid === gi)
        return { gi, first: cols[Math.min(it.c, cols.length - 1)]!, last: cols[Math.min(it.c + it.cw - 1, cols.length - 1)]! }
    }

    // 타일이 차지하는 칼럼 번호 [l, r]
    const colRange = (tileId: string) => {
        const span = columnsOf(tileId)
        if (!span) return undefined
        const cols = geometry.value.columns
        return { l: cols.indexOf(span.first), r: cols.indexOf(span.last) }
    }

    // 보던 자리 = 스냅 기준 왼쪽 칼럼을 차지하는 타일(위쪽 먼저). URL ?view= — REF-tile-url.md U1(넣을 것)
    const anchorId = computed(() => {
        if (overview.value) return undefined
        const cols = geometry.value.columns
        const col = cols[view.value.col]
        if (!col) return undefined
        const local = view.value.col - cols.findIndex(c => c.grid === col.grid)
        const items = grids.value[col.grid]?.items.filter(it => it.c <= local && local < it.c + it.cw) ?? []
        return [...items].sort((a, b) => a.r - b.r)[0]?.tile.id
    })

    // 전체보기 — 현재 배치를 축소해 한 번에 보기 (REF-node-ui-overview.md "전체보기 모드")
    const overview = ref(false)
    const ovPercent = ref<number>(readOvPercent())
    watch(ovPercent, v => writeUserPref(OV_PERCENT_KEY, v))
    const ovScale = computed(() => ovPercent.value / 100)
    // 긴 축을 따라 Grid 한 줄
    const ovDir = computed<'row' | 'col'>(() => (area.value?.w ?? 0) >= (area.value?.h ?? 0) ? 'row' : 'col')
    let normScroll = 0

    // 전체보기의 다른 보드 줄 — 같은 cap으로 채워 "배치 그대로"(EV3(다른 보드 크기) = 같은 축소 비율)
    const otherBoards = computed<BoardGrids[]>(() => !overview.value || !area.value ? [] : boardTrees.list.value.map(({ board, tree, error }) => {
        const l = tree ? layoutOf(tree.tiles, tree.rootId, cap.value) : undefined
        return {
            board,
            grids: l?.grids ?? [],
            posOf: new Map((l?.order ?? []).map((o, i) => [o.tile.id, i + 1])),
            loading: !tree && !error,
            error,
        }
    }))

    const zoomOverview = (delta: -1 | 1) => {
        const { min, max, step } = OV_PERCENT
        ovPercent.value = Math.min(max, Math.max(min, ovPercent.value + delta * step))
    }

    // 들어가기 전 스크롤을 기억하고, 보던 Grid가 전체보기에서도 보이게 둔다
    const enterOverview = async () => {
        const el = stripEl.value
        if (!el || overview.value) return
        normScroll = el.scrollLeft
        const gi = view.value.grid
        overview.value = true
        boardTrees.open()
        await nextTick()
        // 첫 Grid 기준 — 앞에 보드 이름 칸이 있어도 첫 Grid면 스크롤 0
        const cells = el.querySelectorAll<HTMLElement>('.cell')
        const cell = cells[gi], first = cells[0]
        el.scrollLeft = cell && first && ovDir.value === 'row' ? cell.offsetLeft - first.offsetLeft : 0
        el.scrollTop = cell && first && ovDir.value === 'col' ? cell.offsetTop - first.offsetTop : 0
    }

    // tileId 없음(Esc·돌아가기) = 원래 위치 / tileId = 원래 위치 기준으로 그 타일이 보이게(reveal)
    const leaveOverview = async (tileId?: string) => {
        const el = stripEl.value
        if (!el || !overview.value) return
        overview.value = false
        boardTrees.close()
        await nextTick()
        el.scrollTop = 0
        el.scrollLeft = normScroll
        // 위치가 그대로면 scroll 이벤트가 안 와서 전체보기 중 값이 남는다
        scrollX.value = el.scrollLeft
        if (tileId) await reveal(tileId)
    }

    // ---- active 타일 = 키 입력 대상. 탭 메모리만, 보드 화면마다 1개 — REF-node-ui-active.md AT1(상태 위치) ----
    const activeId = ref<string | null>(null)
    // 닫힘·최소화·다른 보드로 이동 = 배치에서 빠짐 -> 해제(다시 나타나도 되살리지 않음)
    watch(posOf, m => {
        if (activeId.value && !m.has(activeId.value)) activeId.value = null
    })
    watch(hash, () => activeId.value = null)

    // 앱 단축키를 끄는가 — 모든 앱 단축키는 처리 맨 앞에서 이것만 본다(새 단축키도). 이동-3(타일별 단축키 끔)
    const appShortcutsOff = computed(() => {
        const t = activeId.value ? tiles.value[activeId.value] : undefined
        return isProcTile(t) && !!t.shortcutsOff
    })

    // 타일 안 누르기 — 포커스는 브라우저 기본 동작에 맡긴다
    const setActive = (tileId: string) => {
        if (posOf.value.has(tileId)) activeId.value = tileId
    }

    // 터미널 = xterm / 그 밖(폴더·볼 수 없는 터미널) = 타일 요소(tabindex -1)
    const focusTile = (tileId: string) => {
        const t = tiles.value[tileId]
        if (isProcTile(t) && focusTerm(t.uid)) return
        stripEl.value?.querySelector<HTMLElement>(`.TileFrame[data-tile-id="${CSS.escape(tileId)}"]`)?.focus({ preventScroll: true })
    }

    // 단축키·이 탭에서 연 타일 — active + 화면 밖이면 스크롤 + 포커스. 전체보기 중이면 나가며 그 타일로 (AT2(active가 되는 때)·AT6(이동할 때 같이 할 일))
    const activate = async (tileId: string) => {
        await nextTick()    // 막 연 타일은 아직 배치(posOf)에 없을 수 있다
        if (!posOf.value.has(tileId)) return
        activeId.value = tileId
        if (overview.value) await leaveOverview(tileId)
        else await reveal(tileId)
        focusTile(tileId)
    }

    // URL 복원(새로고침) — view 타일 칼럼으로 바로 스크롤(없으면 active가 보이게) + active.
    // 포커스는 PC에서만: 폰에서 새로고침마다 가상 키보드가 뜨지 않게 — U3(복원 때 포커스)
    const restore = async (focusId?: string, viewId?: string) => {
        await nextTick()
        const fid = focusId && posOf.value.has(focusId) ? focusId : null
        if (fid) activeId.value = fid
        const span = viewId ? columnsOf(viewId) : undefined
        if (span) scrollTo(span.first.x, 'instant')
        else if (fid) await reveal(fid)
        if (fid && matchMedia('(pointer: fine)').matches) focusTile(fid)
    }

    // 다른 계정에 알릴 active 대상 — 터미널 = uid / 폴더 = 보여 주는 nodeId(루트 = 계정마다 다른 목록이라 없음). REF-node-ui-active.md "원격 active 구조안"
    const activeTarget = computed(() => {
        const t = activeId.value ? tiles.value[activeId.value] : undefined
        if (isProcTile(t)) return { uid: t.uid }
        if (t?.type === 'folder' && t.nodeId !== null) return { nodeId: t.nodeId }
        return {}
    })
    const targetKey = computed(() => JSON.stringify(activeTarget.value))
    // 바뀔 때만. 서버는 탭이 끊기면 지우므로 다시 붙으면 처음부터
    let sent = '{}'
    watch([targetKey, tabReady], ([k, r]) => {
        if (!r) {
            sent = '{}'
            return
        }
        if (k === sent) return
        emit('TAB:ACTIVE', activeTarget.value)
        sent = k
    }, { immediate: true })
    // 보드를 바꾸면 이 화면이 내려간다
    onBeforeUnmount(() => {
        if (sent !== '{}') emit('TAB:ACTIVE', {})
        boardTrees.close()
    })

    // 알약 끌기·미니맵·빈 곳 휠 계약 — 좌표 = strip 스크롤 내용(padding 포함), 스냅 위치 = 칼럼 x - gap
    const pager = providePagerBoard({
        off: overview,
        span: computed(() => cap.value.x),
        maxCol,
        viewCol: computed(() => view.value.col),
        order: computed(() => order.value.map(o => o.tile.id)),
        activeId,
        scrollEl: stripEl,
        scrubbing,
        colLeft: col => { const c = geometry.value.columns[col]; return c ? c.x - TILE_GAP : undefined },
        goCol,
        scrubTo,
        colRange,
        setActive,
        focusTile,
        peekGrids: computed(() => grids.value.map((g, i) => {
            const s = geometry.value.grids[i] ?? { x: 0, w: 0 }
            return { x: s.x, w: s.w, cols: g.cols, rows: 2, half: g.cols === 1, items: g.items.map(it => ({ id: it.tile.id, c: it.c, r: it.r, cw: it.cw, ch: it.ch })) }
        })),
        peekView: computed(() => ({ l: scrollX.value + TILE_GAP, r: scrollX.value + viewW.value - TILE_GAP })),
        aspect: computed(() => area.value?.h ? area.value.w / area.value.h : 0),
    })

    const ctx = {
        area,
        cap,
        order,
        grids,
        newSize,
        posOf,
        instanceOf,
        terminals,
        rootTab,
        stripEl,
        scrollX,
        viewW,
        geometry,
        view,
        goGrid,
        goColumn,
        pos,
        range,
        maxCol,
        scrubbing,
        goCol,
        scrubTo,
        colRange,
        reveal,
        overview,
        ovPercent,
        ovScale,
        ovDir,
        otherBoards,
        zoomOverview,
        enterOverview,
        leaveOverview,
        activeId,
        appShortcutsOff,
        setActive,
        focusTile,
        activate,
        anchorId,
        restore,
        pager,
    }

    provide(TILE_GRIDS_STORE_KEY, ctx)
    return ctx

}

export const useTileGrids = () => {
    const context = inject<ReturnType<typeof provideTileGrids>>(TILE_GRIDS_STORE_KEY)
    if (!context) throw new Error('TileGridsStore is not provided')
    return context
}
