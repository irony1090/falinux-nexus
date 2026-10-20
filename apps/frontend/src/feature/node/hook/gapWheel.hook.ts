import { useTileGrids } from '../store/tileGrids.store'

// 이만큼(px) 모이면 한 칸 / 한 칸 넘긴 뒤 다음 칸까지 최소 간격 / 이만큼 멈추면 모은 양을 버림
const STEP = 50
const STEP_GAP_MS = 250
const RESET_MS = 200
// smooth 스크롤이 끝나기 전 다음 칸은 목표 칼럼에서 센다
const TARGET_MS = 600

// 타일이 아닌 곳(타일 사이·빈 칸·푸터)의 세로 휠 = 칼럼(반 칸) 이동. 타일 위 휠은 터미널 스크롤 그대로 — 푸터-7(빈 곳 휠)
// 가로 휠(트랙패드 가로 쓸기·Shift+휠)은 막지 않는다 — 브라우저 스크롤 + 스냅
export const useGapWheel = (grids: ReturnType<typeof useTileGrids>) => {
    const { overview, view, maxCol, goCol } = grids
    let acc = 0
    let last = 0
    let stepAt = 0
    let target = 0

    return (e: WheelEvent) => {
        if (overview.value || Math.abs(e.deltaY) <= Math.abs(e.deltaX)) return
        if ((e.target as Element | null)?.closest('.Tile:not(.Tile--empty)')) return
        e.preventDefault()
        const now = performance.now()
        if (now - last > RESET_MS) acc = 0
        last = now
        acc += e.deltaMode === WheelEvent.DOM_DELTA_LINE ? e.deltaY * 16 : e.deltaY
        if (Math.abs(acc) < STEP || now - stepAt < STEP_GAP_MS) return
        const base = now - stepAt < TARGET_MS ? target : view.value.col
        const next = Math.max(0, Math.min(maxCol.value, base + Math.sign(acc)))
        acc = 0
        stepAt = now
        target = next
        if (next !== base) goCol(next)
    }
}
