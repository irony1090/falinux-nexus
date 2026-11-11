import { useVirtualKeyboard } from '@/common/hook/virtualKeyboard.hook'
import { DragSession, type DragPoint } from '@/common/listener/dragSession'
import { makeEdgeBands } from '@/feature/widget/util/edgeBand.util'
import { inject, onBeforeUnmount, provide, ref, watch } from 'vue'
import { memberDropOrder, sameOrder, type MemberFrame } from '../util/memberDrop.util'
import { GAP, type MemberTileKind } from '../util/memberGrid.util'
import type { provideMemberBoard } from './memberBoard.store'

const MEMBER_DRAG_STORE_KEY = Symbol('MemberDragStore')

const HOLD_MS = 500     // 터치 길게 누르기 — 보드 끌기와 같게

// 끌기를 시작할 자리 — 마우스 = 머리 줄 빈 곳 / 터치 = 머리 줄 + 몸통 `data-pager-press`(입력칸·버튼·장비 줄 제외). 보드 dragFrom과 같음
export const memberDragFrom = (t: Element, pointerType: string) => {
    if (t.closest('button, a, input, textarea, select, [role="button"], .v-btn')) return false
    if (t.closest('.mb-head')) return true
    if (pointerType === 'mouse') return false
    return !!t.closest('[data-pager-press]') && !t.closest('[data-pager-press-off]')
}

export type MemberChip = { icon: string; color: string; name: string }
const KINDS: Record<MemberTileKind, { icon: string; color: string }> = {
    user: { icon: 'mdi-account', color: 'primary' },
    device: { icon: 'mdi-folder-network', color: 'info' },
    new: { icon: 'mdi-plus', color: 'success' },
}
export const memberChip = (kind: MemberTileKind, name: string): MemberChip => ({ ...KINDS[kind], name })

// 관리 페이지(내 장비·관리자) 타일 끌어서 순서 바꾸기 — 같은 판 안 = 그 타일 하나(보드와 같음), 순서는 URL `t`에만 저장.
// 끄는 동안 = 미리보기 순서(`preview`)로 다시 배치 + 끄는 타일 자리 = 빈 자리 + 작은 표시(안 1). REF-member-drag.md
export const provideMemberDrag = (board: ReturnType<typeof provideMemberBoard>) => {
    const { tiles, preview, scrubbing, scrollEl, scrollX, layout, maxCol, step } = board
    const kb = useVirtualKeyboard()
    const session = new DragSession({ holdMs: HOLD_MS })

    const pressKey = ref<string | null>(null)     // 터치로 누르는 중(시작 대기)
    const dragKey = ref<string | null>(null)
    const chip = ref<(MemberChip & DragPoint) | null>(null)
    let wait: { key: string; chip: MemberChip } | null = null
    let base: MemberFrame | null = null
    let last: DragPoint = { x: 0, y: 0 }
    let raf = 0
    let lastScroll = 0

    const viewCol = () => step.value ? Math.round(scrollX.value / step.value) : 0
    const bands = makeEdgeBands({ el: () => scrollEl.value, viewCol, maxCol: () => maxCol.value, goCol: c => board.goCol(c) })

    // 끌기 시작 때 한 번 잰다 — 칸 크기는 끄는 동안 안 바뀐다(끝 띠 스크롤로 위치만 움직임)
    const frame = () => {
        const el = scrollEl.value
        if (!el) return null
        const r = el.getBoundingClientRect()
        const first = el.querySelector<HTMLElement>('.mb-grid')
        const k = board.plan.value?.cols
        if (!first || !k) return null
        const { cols, rows } = layout.value
        const cs = getComputedStyle(first), b = first.getBoundingClientRect()
        const w = b.width - parseFloat(cs.paddingLeft) - parseFloat(cs.paddingRight)
        const h = b.height - parseFloat(cs.paddingTop) - parseFloat(cs.paddingBottom)
        const gapX = parseFloat(cs.columnGap) || GAP, gapY = parseFloat(cs.rowGap) || GAP
        return {
            gapX, gapY, cols, rows,
            x: b.left - r.left + el.scrollLeft + parseFloat(cs.paddingLeft),
            y: b.top - r.top + parseFloat(cs.paddingTop),
            colW: (w - gapX * (k - 1)) / k,
            rowH: (h - gapY * (rows - 1)) / rows,
        }
    }

    const judge = () => {
        const key = dragKey.value, cur = preview.value, el = scrollEl.value
        if (!key || !cur || !base || !el) return
        const r = el.getBoundingClientRect()
        const p = { x: last.x - r.left + el.scrollLeft, y: last.y - r.top }
        const next = memberDropOrder(cur, key, p, base)
        if (!sameOrder(next, cur)) preview.value = [...next]
    }

    // 끝 띠에 머물면 칼럼 하나씩 넘어간다 — 스크롤이 움직인 만큼 자리를 다시 판정
    const autoScroll = () => {
        const el = scrollEl.value
        if (!el || !dragKey.value) return
        bands.track(last)
        if (el.scrollLeft !== lastScroll) {
            lastScroll = el.scrollLeft
            judge()
        }
        raf = requestAnimationFrame(autoScroll)
    }

    const end = () => {
        cancelAnimationFrame(raf)
        wait = null
        base = null
        pressKey.value = null
        dragKey.value = null
        chip.value = null
        preview.value = null
        scrubbing.value = false
        bands.end()
        document.documentElement.classList.remove('tile-dragging')
    }

    const down = (e: PointerEvent, key: string, info: MemberChip) => {
        if (kb.open) return
        wait = { key, chip: info }
        if (!session.down(e)) wait = null
    }

    session.on('hold', () => pressKey.value = wait?.key ?? null)
    session.on('start', p => {
        if (!wait) return session.cancel()
        base = frame()
        if (!base) return session.cancel()
        pressKey.value = null
        dragKey.value = wait.key
        chip.value = { ...wait.chip, ...p }
        last = p
        preview.value = [...tiles.value]
        scrubbing.value = true
        document.documentElement.classList.add('tile-dragging')
        bands.start()
        lastScroll = scrollEl.value?.scrollLeft ?? 0
        raf = requestAnimationFrame(autoScroll)
    })
    session.on('move', p => {
        last = p
        if (chip.value) chip.value = { ...chip.value, ...p }
        judge()
    })
    session.on('drop', () => {
        const key = dragKey.value, final = preview.value
        if (!key || !final) return end()
        const changed = !sameOrder(final, tiles.value)
        tiles.value = [...final]
        end()
        if (changed) board.activate(key)
    })
    session.on('cancel', () => end())

    // 키보드 열림·끄는 타일이 사라짐(다른 탭에서 닫기) = 취소
    watch([() => kb.open, () => !!dragKey.value && !tiles.value.some(t => t.key === dragKey.value)], ([k, gone]) => {
        if (k || gone) session.cancel()
    })
    watch(scrollEl, (el, _, onCleanup) => {
        if (el) onCleanup(session.blockTouch(el))
    }, { immediate: true })
    onBeforeUnmount(() => session.dispose())

    const ctx = { pressKey, dragKey, chip, edge: bands.edge, down }
    provide(MEMBER_DRAG_STORE_KEY, ctx)
    return ctx
}

export const useMemberDrag = () => {
    const ctx = inject<ReturnType<typeof provideMemberDrag>>(MEMBER_DRAG_STORE_KEY)
    if (!ctx) throw new Error('MemberDragStore is not provided')
    return ctx
}
