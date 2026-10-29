import { DragSession, type DragPoint } from '@/common/listener/dragSession'
import { useAppDialog } from '@/feature/layout/store/appDialog.store'
import { useVirtualKeyboard } from '@/common/hook/virtualKeyboard.hook'
import type { CellSpan } from '@/feature/widget/util/tile.type'
import { dropBundle, dropOrder, gridWidth, type DropFrame } from '@/feature/widget/util/tileDrop.util'
import { TILE_KINDS } from '@/feature/widget/util/tileKind.util'
import { viewIcon } from '@/feature/worker/util/fileKind.util'
import { stripGeometry } from '@/feature/widget/util/tileSplit.util'
import { inject, onBeforeUnmount, provide, ref, watch } from 'vue'
import { useBoardTrees } from './boardTrees.store'
import { TILE_GAP, type provideTileGrids } from './tileGrids.store'
import { useTileTree, type Tile } from './tileTree.store'

const TILE_DRAG_STORE_KEY = Symbol('TileDragStore')

const HOLD_MS = 500         // 터치 길게 누르기 — 목록 칸 체크(itemPress)와 같은 시간
const EDGE = 48             // 전체보기: 화면 끝 이만큼 안이면 자동 스크롤(px). 평소 = 끝 띠(grids.edgeTrack)로 칼럼씩
const SPEED = 18            // 끝에 붙었을 때 한 프레임 스크롤(px)
const OV_CELL_GAP = 14      // 전체보기 Grid 사이 — TileStrip `.overview .board-cells` gap과 같게

export type DragChip = { icon: string; color: string; name: string }

// 끌기를 시작할 자리 — 마우스 = 머리 줄 빈 곳 / 터치 = 머리 줄 + 몸통 `data-pager-press` 안 칸 밖(옛 길게 누르기 자리, 제스처-1(펼치기 패널))
export const dragFrom = (t: Element, pointerType: string) => {
    if (t.closest('button, a, input, textarea, select, [role="button"], .v-btn')) return false
    if (t.closest('.TileFrame__head')) return true
    if (pointerType === 'mouse') return false
    return !!t.closest('[data-pager-press]') && !t.closest('.DirGridCell, [data-pager-press-off]')
}

// 끄는 것 — hash = 출발 보드, bundle = 다른 보드로 갈 때 같이 가는 타일(id + 연 타일, 출발 순서)
type Source = { hash: string; id: string; bundle: Array<string>; tiles: Record<string, Tile>; rootId: string }
type Board = { order: ReadonlyArray<string>; tiles: Readonly<Record<string, Tile>>; rootId: string }

// id와 그 타일이 연 타일 전부(연 곳 사슬), order 순 — 서버 tileDoc.subtree와 같음
const openedBy = (b: Board, id: string) => {
    const inn = new Set([id])
    for (let grew = true; grew;) {
        grew = false
        Object.values(b.tiles).forEach(t => {
            if (t.parent && inn.has(t.parent) && !inn.has(t.id)) { inn.add(t.id); grew = true }
        })
    }
    return b.order.filter(x => inn.has(x))
}

// 보드 타일 끌어서 순서 바꾸기 — 끌기-2(타일 끌기)·보드-1(보드 간 이동 API), REF-tile-drag.md
// 끄는 동안 = 미리보기 순서(grids.preview)로 다시 배치 + 끄는 타일 자리 = 빈 자리 + 손가락 옆 작은 표시(안 1).
// 평소 = 지금 보드 안에서만. 전체보기 = 모든 보드 줄에서 끌고, 다른 보드 줄에 놓으면 묶음째 그 보드로(가(같은 보드 안끼리 통일))
export const provideTileDrag = (grids: ReturnType<typeof provideTileGrids>) => {
    const { tiles, order: treeOrder, rootId, hash, place, moveTile } = useTileTree()
    const boardTrees = useBoardTrees()
    const kb = useVirtualKeyboard()
    const { openDialog } = useAppDialog()
    const session = new DragSession({ holdMs: HOLD_MS })

    const pressId = ref<string | null>(null)     // 터치로 누르는 중(시작 대기)
    const dragId = ref<string | null>(null)
    const chip = ref<(DragChip & DragPoint) | null>(null)
    // 같이 가는 타일(끄는 타일 제외) — 다른 보드 줄 위일 때만 채운다(보드Q6(묶음 표시 때)). 작은 표시에 붙일 아이콘·이름은 시작 때 잰다
    const bundleIds = ref<Array<string>>([])
    const bundleChips = ref<Array<DragChip>>([])
    let wait: { src: Source; chip: DragChip; ov: boolean } | null = null
    let src: Source | null = null
    let ov = false
    let bases: Record<string, Board> = {}     // 시작 때 보드별 원래 순서(전체보기)
    let target = ''                           // 지금 미리보기가 들어간 보드
    let last: DragPoint = { x: 0, y: 0 }
    let raf = 0

    const boardOf = (h: string): Board | undefined => {
        if (h === hash.value) return { order: treeOrder.value, tiles: tiles.value, rootId: rootId.value }
        const tree = boardTrees.list.value.find(b => b.board.hash === h)?.tree
        return tree && { order: tree.order, tiles: tree.tiles, rootId: tree.rootId }
    }

    const begin = (e: PointerEvent, h: string, tileId: string, info: DragChip, inOv: boolean) => {
        const b = boardOf(h)
        if (!b || !b.tiles[tileId] || tileId === b.rootId) return
        const bundle = openedBy(b, tileId)
        wait = { src: { hash: h, id: tileId, bundle, tiles: Object.fromEntries(bundle.map(x => [x, b.tiles[x]!])), rootId: b.rootId }, chip: info, ov: inOv }
        if (!session.down(e)) wait = null     // down 안에서 hold가 바로 나온다 — 먼저 채운다
    }

    // 평소 모드 — 지금 보드 머리 줄(TileFrame)
    const down = (e: PointerEvent, tileId: string, info: DragChip) => {
        if (grids.overview.value || kb.open) return
        begin(e, hash.value, tileId, info, false)
    }
    // 전체보기 — 타일 아무 데서나(TileStrip). h = 그 타일이 있는 보드 줄
    const downOv = (e: PointerEvent, h: string, tileId: string, info: DragChip) => {
        if (!grids.overview.value) return
        begin(e, h, tileId, info, true)
    }

    // 평소 = strip 스크롤 내용 좌표
    const normalFrame = (): DropFrame | null => {
        const area = grids.area.value
        if (!area) return null
        return {
            area, cap: grids.cap.value, gap: TILE_GAP, scale: 1,
            boxes: cols => stripGeometry(cols, area.w, grids.cap.value, TILE_GAP).grids.map(g => ({ x: g.x, y: TILE_GAP })),
        }
    }
    // 전체보기 = 화면 좌표. 첫 Grid 상자 위치를 재고 나머지는 폭으로 잇는다(Grid 칸 폭 = 상자 폭, 캡션은 그 안에서 줄임).
    // 보드가 하나면 줄이 가운데 정렬이라 Grid 수가 바뀌면 시작점도 옮겨 간다
    const ovFrame = (row: Element): DropFrame | null => {
        const area = grids.area.value, strip = grids.stripEl.value
        const boxes = row.querySelectorAll<HTMLElement>('.board-cells > .cell > .box')
        if (!area || !strip || !boxes.length) return null
        const s = grids.ovScale.value
        const first = boxes[0]!.getBoundingClientRect(), lastBox = boxes[boxes.length - 1]!.getBoundingClientRect()
        const curW = lastBox.right - first.left
        const centered = !strip.classList.contains('multi') && strip.scrollWidth <= strip.clientWidth + 1
        return {
            area, cap: grids.cap.value, gap: TILE_GAP, scale: s,
            boxes: (cols: Array<CellSpan>) => {
                const ws = cols.map(c => gridWidth(c, area, TILE_GAP) * s)
                const total = ws.reduce((a, w) => a + w, 0) + OV_CELL_GAP * (ws.length - 1)
                let x = first.left + (centered ? (curW - total) / 2 : 0)
                return ws.map(w => { const o = { x, y: first.top }; x += w + OV_CELL_GAP; return o })
            },
        }
    }

    const judge = () => {
        if (!src || !dragId.value) return
        if (!ov) {
            const el = grids.stripEl.value, f = normalFrame(), cur = grids.preview.value?.orders[src.hash]
            if (!el || !f || !cur) return
            const r = el.getBoundingClientRect()
            const p = { x: last.x - r.left + el.scrollLeft, y: last.y - r.top + el.scrollTop }
            const next = dropOrder(cur, tiles.value, src.id, src.rootId, p, f)
            if (next !== cur) grids.preview.value = { orders: { [src.hash]: next }, tiles: {} }
            return
        }
        const row = document.elementFromPoint(last.x, last.y)?.closest<HTMLElement>('.board[data-hash]')
        const h = row?.dataset.hash, b = h ? bases[h] : undefined, f = row ? ovFrame(row) : null
        if (!h || !b || !f) return      // 줄 사이 = 그대로
        const prev = grids.preview.value?.orders ?? {}
        if (h === src.hash) {
            const cur = target === h ? prev[h] ?? b.order : b.order
            const next = dropOrder(cur, b.tiles, src.id, src.rootId, last, f)
            if (target !== h || next !== cur) grids.preview.value = { orders: { [h]: next }, tiles: {} }
        } else {
            const cur = target === h ? prev[h] ?? b.order : b.order
            const next = dropBundle(b.order, cur, { ...b.tiles, ...src.tiles }, src.bundle, b.rootId, last, f)
            const out = bases[src.hash]!.order.filter(x => !src!.bundle.includes(x))
            if (target !== h || next !== cur) grids.preview.value = { orders: { [src.hash]: out, [h]: next }, tiles: src.tiles }
        }
        target = h
        const along = h !== src.hash ? src.bundle.filter(x => x !== src!.id) : []
        if (along.join() !== bundleIds.value.join()) bundleIds.value = along
    }

    let lastScroll = 0
    const autoScroll = () => {
        const el = grids.stripEl.value
        if (!el || !dragId.value) return
        if (!ov) {      // 끝 띠에 머물면 칼럼 하나씩(부드러운 스크롤) — 움직이는 동안 매 프레임 다시 판정
            grids.edgeTrack(last)
            if (el.scrollLeft !== lastScroll) {
                lastScroll = el.scrollLeft
                judge()
            }
            raf = requestAnimationFrame(autoScroll)
            return
        }
        const r = el.getBoundingClientRect()
        const step = (v: number, lo: number, hi: number) => {
            const d = v < lo + EDGE ? v - (lo + EDGE) : v > hi - EDGE ? v - (hi - EDGE) : 0
            return Math.sign(d) * Math.min(SPEED, Math.abs(d) / EDGE * SPEED)
        }
        const dx = step(last.x, r.left, r.right), dy = step(last.y, r.top, r.bottom)
        if (dx || dy) {
            const bx = el.scrollLeft, by = el.scrollTop
            el.scrollLeft += dx
            el.scrollTop += dy
            if (el.scrollLeft !== bx || el.scrollTop !== by) judge()
        }
        raf = requestAnimationFrame(autoScroll)
    }

    const end = (keepPreview = false) => {
        cancelAnimationFrame(raf)
        wait = null
        src = null
        bases = {}
        target = ''
        pressId.value = null
        dragId.value = null
        chip.value = null
        bundleIds.value = []
        bundleChips.value = []
        if (!keepPreview) grids.preview.value = null
        document.documentElement.classList.remove('tile-dragging')
        grids.edgeEnd()
    }

    session.on('hold', () => pressId.value = wait?.src.id ?? null)
    session.on('start', p => {
        if (!wait) return session.cancel()
        src = wait.src
        ov = wait.ov
        pressId.value = null
        dragId.value = src.id
        chip.value = { ...wait.chip, ...p }
        last = p
        bases = {}
        if (ov) {
            const own = boardOf(hash.value)
            if (own) bases[hash.value] = own
            boardTrees.list.value.forEach(({ board, tree }) => { if (tree) bases[board.hash] = { order: tree.order, tiles: tree.tiles, rootId: tree.rootId } })
        } else {
            bases[src.hash] = boardOf(src.hash)!
        }
        target = src.hash
        // 이름 = 화면의 타일 라벨 글자(최소화 터미널은 화면에 없음)
        bundleChips.value = !ov ? [] : src.bundle.filter(x => x !== src!.id).map(x => {
            const t = src!.tiles[x]!
            const name = document.querySelector(`.Tile[data-tile-id="${CSS.escape(x)}"] .ov-name .name`)?.textContent?.trim()
            return { icon: t.type === 'viewer' ? viewIcon(t.path) : TILE_KINDS[t.type].icon, color: TILE_KINDS[t.type].color, name: name || '최소화한 터미널' }
        })
        grids.preview.value = { orders: { [src.hash]: [...bases[src.hash]!.order] }, tiles: {} }
        document.documentElement.classList.add('tile-dragging')
        if (!ov) grids.edgeStart()
        lastScroll = grids.stripEl.value?.scrollLeft ?? 0
        raf = requestAnimationFrame(autoScroll)
    })
    session.on('move', p => {
        last = p
        if (chip.value) chip.value = { ...chip.value, ...p }
        judge()
    })
    session.on('drop', () => {
        const s = src, final = grids.preview.value?.orders[target], to = target
        if (!s || !final) return end()
        const base = bases[to]!
        const changed = final.join() !== base.order.join()
        // 놓을 자리 = 미리보기에서 묶음(같은 보드면 그 타일) 바로 앞
        const firstIdx = final.findIndex(x => (to === s.hash ? [s.id] : s.bundle).includes(x))
        const after = final[firstIdx - 1] ?? ''
        if (!changed) return end()
        if (!ov) {
            place(s.id, after)
            end()
            grids.setActive(s.id)
            return
        }
        if (to === s.hash && s.hash === hash.value) {
            place(s.id, after)
            return end()
        }
        // 서버가 바꾸는 동안 미리보기를 둔다 — 응답·push가 오면 실제 트리로
        end(true)
        moveTile(s.hash, s.id, to, after)
            .catch(err => openDialog({ type: 'error', content: err?.message || '타일을 옮기지 못했습니다' }))
            .finally(() => grids.preview.value = null)
    })
    session.on('cancel', () => end())

    // 키보드 열림·전체보기 들어가기/나가기·끄는 타일이 사라짐(보드 바꾸기·다른 탭에서 닫기) = 취소
    watch([() => kb.open, grids.overview, () => !!dragId.value && !ov && !tiles.value[dragId.value]], ([k, o, gone], [, oldO]) => {
        if (k || o !== oldO || gone) session.cancel()
    })
    watch(grids.stripEl, (el, _, onCleanup) => {
        if (el) onCleanup(session.blockTouch(el))
    }, { immediate: true })
    onBeforeUnmount(() => session.dispose())

    const ctx = { pressId, dragId, chip, bundleIds, bundleChips, down, downOv }
    provide(TILE_DRAG_STORE_KEY, ctx)
    return ctx
}

export const useTileDrag = () => {
    const context = inject<ReturnType<typeof provideTileDrag>>(TILE_DRAG_STORE_KEY)
    if (!context) throw new Error('TileDragStore is not provided')
    return context
}
