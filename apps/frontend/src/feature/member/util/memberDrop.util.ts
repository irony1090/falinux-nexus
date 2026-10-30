import { pack, type MemberTile } from './memberGrid.util'

// 관리 페이지 타일 끌기 — 놓을 자리 계산. 흔들림 방지 ①②③은 보드 `tileDrop.util.dropOrder`와 같은 규칙(REF-tile-drag.md)인데,
// 배치가 `pack`(N×M, 열 채우기)이라 보드 것(strip 2행 전제)을 못 쓴다. REF-member-drag.md

type Rect = { x: number; y: number; w: number; h: number }
type Point = { x: number; y: number }
// grids = 끌기 시작 때 잰 Grid 상자(스크롤 내용 좌표), w·h = 안쪽 여백 뺀 크기
export type MemberFrame = { gap: number; cols: number; rows: number; grids: Array<Rect> }

const inside = (r: Rect, p: Point) => p.x >= r.x && p.x < r.x + r.w && p.y >= r.y && p.y < r.y + r.h
export const sameOrder = (a: readonly MemberTile[], b: readonly MemberTile[]) =>
    a.length === b.length && a.every((t, i) => t.key === b[i]!.key)

const tileOf = (o: readonly MemberTile[], key: string) => o.find(t => t.key === key)!
const without = (o: readonly MemberTile[], key: string) => o.filter(t => t.key !== key)

// X를 before 바로 앞(없으면 맨 뒤) / after 바로 뒤(없으면 맨 뒤)
const moveBefore = (o: readonly MemberTile[], key: string, before: string) => {
    const tile = tileOf(o, key), rest = without(o, key), i = rest.findIndex(t => t.key === before)
    if (i < 0) return [...rest, tile]
    return [...rest.slice(0, i), tile, ...rest.slice(i)]
}
const moveAfter = (o: readonly MemberTile[], key: string, after: string) => {
    const tile = tileOf(o, key), rest = without(o, key), i = rest.findIndex(t => t.key === after)
    if (i < 0) return [...rest, tile]
    return [...rest.slice(0, i + 1), tile, ...rest.slice(i + 1)]
}

// 끌기 중 미리보기 순서로 배치했을 때의 타일 자리 — 판정은 이 계산 위치로(애니메이션 중 화면 위치 아님)
const layOut = (tiles: readonly MemberTile[], f: MemberFrame) => {
    const plans = pack(tiles, f.cols, f.rows)
    const first = f.grids[0]
    const stepX = f.grids.length > 1 ? f.grids[1]!.x - first!.x : 0
    // Grid가 새로 생기면(맨 뒤로 옮길 때) 잰 적이 없다 — 같은 간격으로 이어 붙인다
    const boxOf = (gi: number): Rect | undefined => {
        const g = f.grids[gi]
        if (g) return g
        if (!first || !stepX) return undefined
        const last = f.grids[f.grids.length - 1]!
        return { ...last, x: last.x + stepX * (gi - f.grids.length + 1) }
    }
    const placed: Array<{ key: string; rect: Rect }> = []
    const grids = plans.flatMap((g, gi) => {
        const box = boxOf(gi)
        if (!box) return []
        const colW = (box.w - f.gap * (f.cols - 1)) / f.cols
        const rowH = (box.h - f.gap * (f.rows - 1)) / f.rows
        for (const it of g.items) placed.push({
            key: it.tile.key,
            rect: {
                x: box.x + it.c * (colW + f.gap),
                y: box.y + it.r * (rowH + f.gap),
                w: it.cw * colW + (it.cw - 1) * f.gap,
                h: it.ch * rowH + (it.ch - 1) * f.gap,
            },
        })
        return [{ box, items: g.items.map(it => ({ key: it.tile.key, c: it.c, r: it.r, cw: it.cw, ch: it.ch })) }]
    })
    return { placed, grids }
}

// 빈 칸 (c, r)이 채우는 순서에서 몇 번째 칸인지 — `pack`과 같은 열 우선(위에서 아래로, 다음 열)
const cellIndex = (c: number, r: number, rows: number) => c * rows + r

// 판정은 배치 계산 위치(`pack`)로. 커서가 끄는 타일(빈 자리) 위면 그대로(②)
// 타일 위 = 그 타일의 **위 절반이면 앞, 아래 절반이면 뒤**(크기가 섞여도 규칙 하나 — 장비 타일(세로 2칸) 아래쪽에 놓아도 됨).
// 빈 칸 = **그 칸 자리**(채우는 순서에서 그 칸보다 앞인 마지막 타일 뒤) — 반 칸 타일 아래 빈 곳에도 들어간다
// Grid 오른쪽 너머 = 맨 뒤(왼쪽 너머 = 폰 목록 칸, 그대로). 루트 타일이 없다(관리 페이지)
export const memberDropOrder = (
    tiles: readonly MemberTile[], key: string, p: Point, f: MemberFrame,
): readonly MemberTile[] => {
    const now = layOut(tiles, f)
    const over = now.placed.find(t => inside(t.rect, p))
    if (over) {
        if (over.key === key) return tiles
        const top = p.y < over.rect.y + over.rect.h / 2
        return top ? moveBefore(tiles, key, over.key) : moveAfter(tiles, key, over.key)
    }
    const grid = now.grids.find(g => inside(g.box, p))
    if (grid) {
        const box = grid.box
        const colW = (box.w - f.gap * (f.cols - 1)) / f.cols
        const rowH = (box.h - f.gap * (f.rows - 1)) / f.rows
        const at = cellIndex(Math.floor((p.x - box.x) / (colW + f.gap)), Math.floor((p.y - box.y) / (rowH + f.gap)), f.rows)
        // 그 칸보다 앞(채우는 순서)인 마지막 타일 뒤로 — 없으면 이 Grid 첫 타일 앞
        const before = grid.items.filter(it => cellIndex(it.c, it.r, f.rows) < at).pop()
        if (before) return before.key === key ? tiles : moveAfter(tiles, key, before.key)
        const first = grid.items[0]
        return first && first.key !== key ? moveBefore(tiles, key, first.key) : tiles
    }
    const end = now.grids[now.grids.length - 1]?.box
    if (end && p.x >= end.x + end.w) return moveAfter(tiles, key, '')
    return tiles
}
