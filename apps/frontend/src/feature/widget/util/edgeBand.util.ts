import { shallowRef } from 'vue'

// 끄는 동안 화면 끝 띠 — 띠 안에 머물면 칼럼 하나씩 넘긴다. 모양 = `PageEdgeBands`, 판정 = 여기(보드 칸·타일 끌기, 관리 페이지 끌기)
export const EDGE_BAND = 48       // 화면 끝 띠 폭(px)
export const EDGE_HOLD_MS = 400   // 띠에 머물러 넘기기까지 — PageEdgeBands 차오름과 같게

export type EdgeState = { side: -1 | 0 | 1; since: number; prev: boolean; next: boolean }

export const makeEdgeBands = (opt: {
    el: () => HTMLElement | undefined
    viewCol: () => number
    maxCol: () => number
    goCol: (col: number) => void
}) => {
    const edge = shallowRef<EdgeState | null>(null)
    let col = -1        // 넘긴 칼럼 — 부드러운 스크롤 중엔 viewCol이 아직 옛 값이라 따로 센다
    const start = () => {
        col = -1
        edge.value = { side: 0, since: 0, prev: opt.viewCol() > 0, next: opt.viewCol() < opt.maxCol() }
    }
    const track = (p: { x: number; y: number }) => {
        const e = edge.value, el = opt.el()
        if (!e || !el) return
        const r = el.getBoundingClientRect()
        const cur = col >= 0 ? col : opt.viewCol()
        const prev = cur > 0, next = cur < opt.maxCol()
        let side: -1 | 0 | 1 = p.y < r.top || p.y > r.bottom ? 0 : p.x < r.left + EDGE_BAND ? -1 : p.x > r.right - EDGE_BAND ? 1 : 0
        if ((side < 0 && !prev) || (side > 0 && !next)) side = 0
        const now = performance.now()
        if (side && side === e.side && now - e.since >= EDGE_HOLD_MS) {
            col = cur + side
            opt.goCol(col)
            edge.value = { side, since: now, prev: col > 0, next: col < opt.maxCol() }
        } else if (side !== e.side || prev !== e.prev || next !== e.next) {
            edge.value = { side, since: side !== e.side ? now : e.since, prev, next }
        }
    }
    const end = () => {
        edge.value = null
        col = -1
    }
    return { edge, start, track, end }
}
