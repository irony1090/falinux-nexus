import { providePagerBoard } from '@/feature/widget/store/pagerBoard.store'
import { revealCol } from '@/feature/widget/util/tileSplit.util'
import { computed, inject, nextTick, provide, ref, shallowRef } from 'vue'
import {
    GAP, insertBefore, insertTile, layoutOf, mainKeyOf, newTile, pack, type MemberTile,
} from '../util/memberGrid.util'

const MEMBER_BOARD_KEY = Symbol('MemberBoard')

export type MemberPageKind = 'admin' | 'mine'
export type MemberTab = 'users' | 'devices'

// 관리 페이지 한 화면 — 목록 상태 + 상세 타일 + 가로 스크롤. REF-member-ui.md 화-5(상세 쌓기)~화-12(공유 Grid)
// 스크롤 단위 = 칼럼(보드 revealCol 그대로). PC = 상세 strip만 스크롤 / 폰 = 목록 + Grid가 한 줄로 스크롤
export const provideMemberBoard = (page: MemberPageKind) => {
    const admin = page === 'admin'

    // ---- 목록 ----
    const tab = ref<MemberTab>(admin ? 'users' : 'devices')
    const q = ref('')
    const filter = ref('all')
    const pageNo = ref(1)
    const size = ref(10)
    const pickTab = (t: MemberTab) => {
        if (!admin || t === tab.value) return
        tab.value = t
        q.value = ''
        filter.value = 'all'
        pageNo.value = 1
    }

    // ---- 타일 ----
    const tiles = ref<MemberTile[]>([])
    const activeKey = ref<string | null>(null)
    const area = ref({ w: 0, h: 0 })
    const layout = computed(() => layoutOf(area.value.w, area.value.h))
    const grids = computed(() => area.value.w ? pack(tiles.value, layout.value.cols, layout.value.rows) : [])
    // 폰은 목록이 맨 앞 한 칸(Grid 0 앞)
    const lead = computed(() => layout.value.phone ? 1 : 0)

    // ---- 가로 스크롤 — MemberPage가 요소·값을 기록 ----
    const scrollEl = shallowRef<HTMLElement>()
    const scrollX = ref(0)
    const viewW = ref(0)
    // 칼럼 한 칸 = 화면 폭 / 열 수(Grid 폭 = 화면 폭, 열 사이 간격 포함)
    const span = computed(() => layout.value.phone ? 1 : layout.value.cols)
    const step = computed(() => viewW.value / span.value)
    const colCount = computed(() => lead.value + grids.value.length * span.value)
    const maxCol = computed(() => Math.max(0, colCount.value - span.value))
    // 도트 위치 = 화면 단위 실수(폰은 0 = 목록)
    const pos = computed(() => viewW.value ? scrollX.value / viewW.value : 0)
    // 타일이 없어도 빈 안내 페이지가 한 칸 — 빼면 그 페이지의 위치 표시가 도트 줄 밖에 그려진다
    const dotCount = computed(() => lead.value + Math.max(1, grids.value.length))

    const clampCol = (col: number) => Math.max(0, Math.min(maxCol.value, col))
    const goCol = (col: number, behavior: ScrollBehavior = 'smooth') =>
        scrollEl.value?.scrollTo({ left: clampCol(col) * step.value, behavior })
    const goDot = (i: number) => goCol(i * span.value)

    const placeOf = (key: string) => {
        for (const [gi, g] of grids.value.entries()) {
            const it = g.items.find(x => x.tile.key === key)
            if (it) return { gi, ...it }
        }
        return undefined
    }
    // 타일 칼럼 [l, r] — 목록 칸 포함 전체 번호
    const colRange = (key: string) => {
        const p = placeOf(key)
        if (!p) return undefined
        const l = lead.value + p.gi * span.value + p.c
        return { l, r: l + p.cw - 1 }
    }

    // 타일 위치 규칙 — 화면 밖일 때만 막 다 보이는 만큼(보드 reveal과 같은 규칙). 막 바뀐 배치가 DOM에 붙은 뒤 읽는다
    const reveal = async (key: string) => {
        await nextTick()
        const el = scrollEl.value
        const range = colRange(key)
        if (!el || !range || !step.value) return
        const cur = Math.round(el.scrollLeft / step.value)
        const v = revealCol(range.l, range.r, cur, span.value, maxCol.value)
        if (v !== cur) goCol(v)
    }

    // 보던 자리 = 화면 왼쪽 칼럼을 차지하는 타일(위쪽 먼저). 폰에서 목록을 보는 중이면 없음 — 주소-4(보던 위치)
    const anchorKey = computed(() => {
        if (!step.value) return undefined
        const col = Math.round(scrollX.value / step.value) - lead.value
        if (col < 0) return undefined
        const g = grids.value[Math.floor(col / span.value)]
        const local = col % span.value
        const items = g?.items.filter(it => it.c <= local && local < it.c + it.cw) ?? []
        return [...items].sort((a, b) => a.r - b.r)[0]?.tile.key
    })
    // URL 복원 — 그 타일 칼럼으로 바로. 스크롤 폭(viewW) 측정이 Grid 측정보다 늦을 수 있어 몇 프레임 기다린다
    const restoreView = async (key: string) => {
        await nextTick()
        for (let i = 0; i < 10 && !step.value; i++) await new Promise(r => requestAnimationFrame(r))
        const range = colRange(key)
        const el = scrollEl.value
        if (!range || !el) return
        goCol(range.l, 'instant')
        scrollX.value = el.scrollLeft     // scroll 이벤트를 기다리지 않고 보던 자리(anchorKey)를 맞춘다
    }

    const setActive = (key: string) => {
        if (tiles.value.some(t => t.key === key)) activeKey.value = key
    }

    // 화-5(상세 쌓기): 이미 열렸으면 그 타일로 이동, 아니면 부모 다음(화-9(자식 자리))에 끼워 넣는다
    const open = (key: string, parent: string | null = null) => {
        if (!tiles.value.some(t => t.key === key)) tiles.value = insertTile(tiles.value, newTile(key, parent))
        activeKey.value = key
        void reveal(key)
    }
    // 부모를 닫아도 자식은 남김(화-9)
    const close = (key: string) => {
        tiles.value = tiles.value.filter(t => t.key !== key)
        if (activeKey.value === key) activeKey.value = null
        // 다 닫으면 목록으로(폰 = 첫 칸, PC는 목록이 늘 보여 그대로)
        if (!tiles.value.length) void nextTick(() => goCol(0))
    }
    // 주소-5(없는 대상): 지워진 회원·장비 타일은 닫는다(URL은 늘 replace라 기록이 쌓이지 않음)
    const dropMissing = close
    // 화-11(경로): 앞 조각 = 그 타일로, 닫혀 있으면 자식 바로 앞에 다시 엶
    const openParent = (child: MemberTile) => {
        const p = child.parent
        if (!p) return
        if (!tiles.value.some(t => t.key === p)) tiles.value = insertBefore(tiles.value, newTile(p), child.key)
        activeKey.value = p
        void reveal(p)
    }
    // 등록 타일 -> 같은 자리에 장비 타일(키 펼침)
    const freshKeys = ref<Record<string, string>>({})
    const replaceTile = (from: string, to: string, apiKey?: string) => {
        if (apiKey) freshKeys.value = { ...freshKeys.value, [mainKeyOf(to)]: apiKey }
        const exists = tiles.value.some(t => t.key === to)
        tiles.value = exists ? tiles.value.filter(t => t.key !== from) : tiles.value.map(t => t.key === from ? newTile(to, t.parent) : t)
        activeKey.value = to
        void reveal(to)
    }

    // 화-10(타일 크기): 머리 줄 크기 팝업·컨트롤 패널 크기 카드가 같이 씀. 화면이 작아 잘린 칸 수는 그대로 둔다
    const setSize = (key: string, axis: 'w' | 'h', v: number) => {
        tiles.value = tiles.value.map(t => t.key === key ? { ...t, [axis]: v } : t)
        activeKey.value = key
        void reveal(key)
    }

    // 위치 번호(1부터) = 타일 순서 — Ctrl+Alt+숫자
    const posOf = computed(() => new Map(tiles.value.map((t, i) => [t.key, i + 1])))
    const activate = (key: string) => {
        if (!posOf.value.has(key)) return
        activeKey.value = key
        void reveal(key)
    }

    // ---- 알약 끌기·미니맵(PagerBoard 계약) — 보드 GridPager와 같은 동작 ----
    // 끄는 동안 스냅 끔(MemberPage가 class로 건다)
    const scrubbing = ref(false)
    const focusTile = (key: string) =>
        scrollEl.value?.querySelector<HTMLElement>(`[data-tile-key="${CSS.escape(key)}"]`)?.focus({ preventScroll: true })
    // Grid 안쪽 여백 — PC = 오른쪽만, 폰 = 양쪽(MemberPage .grid padding)
    const padL = computed(() => layout.value.phone ? GAP : 0)
    const pager = providePagerBoard({
        off: computed(() => false),
        span,
        maxCol,
        viewCol: computed(() => step.value ? Math.round(scrollX.value / step.value) : 0),
        order: computed(() => tiles.value.map(t => t.key)),
        activeId: activeKey,
        scrollEl,
        scrubbing,
        colLeft: col => step.value ? clampCol(col) * step.value : undefined,
        goCol,
        scrubTo: col => { if (scrollEl.value) scrollEl.value.scrollLeft = clampCol(col) * step.value },
        colRange,
        setActive,
        focusTile,
        peekGrids: computed(() => grids.value.map((g, gi) => ({
            x: (lead.value + gi * span.value) * step.value + padL.value,
            w: span.value * step.value - padL.value - GAP,
            cols: layout.value.cols,
            rows: layout.value.rows,
            half: false,
            items: g.items.map(it => ({ id: it.tile.key, c: it.c, r: it.r, cw: it.cw, ch: it.ch })),
        }))),
        peekView: computed(() => ({ l: scrollX.value + padL.value, r: scrollX.value + viewW.value - GAP })),
        aspect: computed(() => area.value.h ? viewW.value / area.value.h : 0),
    })

    const ctx = {
        page, admin,
        tab, q, filter, pageNo, size, pickTab,
        tiles, activeKey, area, layout, grids, lead,
        scrollEl, scrollX, viewW, span, step, maxCol, pos, dotCount, scrubbing,
        pager, goCol, goDot, colRange, reveal, anchorKey, restoreView, setActive, open, close, dropMissing, openParent, setSize, posOf, activate, freshKeys, replaceTile,
    }
    provide(MEMBER_BOARD_KEY, ctx)
    return ctx
}

export const useMemberBoard = () => {
    const ctx = inject<ReturnType<typeof provideMemberBoard>>(MEMBER_BOARD_KEY)
    if (!ctx) throw new Error('MemberBoard is not provided')
    return ctx
}
