import { cellIndex, pack, type MemberTile } from './memberGrid.util'

// 관리 페이지 타일 끌기 — 놓을 자리 계산. 흔들림 방지 ①②③은 보드 `tileDrop.util.dropOrder`와 같은 규칙(REF-tile-drag.md)인데,
// 배치가 `pack`(행 수가 1~2로 바뀜)이라 보드 것(2행 전제)을 못 쓴다. REF-member-drag.md · REF-column-flow.md 칼럼-4(빈 칸에 놓기)

type Rect = { x: number; y: number; w: number; h: number }
type Point = { x: number; y: number }
// 끌기 시작 때 잰 칸 격자(스크롤 내용 좌표): x·y = 판 안쪽 왼쪽 위, colW·rowH = 칸 크기, gapX·gapY = 칸 사이(폰은 칼럼 사이 2gap).
// 판 폭은 쓴 칼럼 수만큼이라 순서가 바뀔 때마다 pack 결과(cols)로 다시 낸다
export type MemberFrame = { gapX: number; gapY: number; cols: number; rows: number; x: number; y: number; colW: number; rowH: number }

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
    const plan = pack(tiles, f.cols, f.rows)
    const { colW, rowH, gapX, gapY } = f
    const box = { x: f.x, y: f.y, w: plan.cols * colW + (plan.cols - 1) * gapX, h: f.rows * rowH + (f.rows - 1) * gapY }
    const placed = plan.items.map(it => ({
        key: it.tile.key,
        rect: {
            x: box.x + it.c * (colW + gapX),
            y: box.y + it.r * (rowH + gapY),
            w: it.cw * colW + (it.cw - 1) * gapX,
            h: it.ch * rowH + (it.ch - 1) * gapY,
        },
    }))
    return { placed, box, items: plan.items }
}

// 판정은 배치 계산 위치(`pack`)로. 커서가 끄는 타일(빈 자리) 위면 그대로(②)
// 타일 위 = 그 타일의 **위 절반이면 앞, 아래 절반이면 뒤**(크기가 섞여도 규칙 하나 — 장비 타일(세로 2칸) 아래쪽에 놓아도 됨).
// 빈 칸 = **그 칸 자리**(채우는 순서에서 그 칸보다 앞인 마지막 타일 뒤) — 반 칸 타일 아래 빈 곳에도 들어간다
// 판 오른쪽 너머 = 맨 뒤(왼쪽 너머 = 폰 목록 칸, 그대로). 루트 타일이 없다(관리 페이지)
const dropOnce = (
    tiles: readonly MemberTile[], key: string, p: Point, f: MemberFrame,
): readonly MemberTile[] => {
    const now = layOut(tiles, f)
    const over = now.placed.find(t => inside(t.rect, p))
    if (over) {
        if (over.key === key) return tiles
        const top = p.y < over.rect.y + over.rect.h / 2
        return top ? moveBefore(tiles, key, over.key) : moveAfter(tiles, key, over.key)
    }
    const box = now.box
    if (inside(box, p)) {
        const at = cellIndex(Math.floor((p.x - box.x) / (f.colW + f.gapX)), Math.floor((p.y - box.y) / (f.rowH + f.gapY)), f.rows)
        // 그 칸보다 앞(채우는 순서)인 마지막 타일 뒤로 — 없으면 맨 앞
        const before = now.items.filter(it => cellIndex(it.c, it.r, f.rows) < at).pop()
        if (before) return before.tile.key === key ? tiles : moveAfter(tiles, key, before.tile.key)
        const first = now.items[0]
        return first && first.tile.key !== key ? moveBefore(tiles, key, first.tile.key) : tiles
    }
    if (p.x >= box.x + box.w) return moveAfter(tiles, key, '')
    return tiles
}

const orderKey = (o: readonly MemberTile[]) => o.map(t => t.key).join()

// 놓을 자리 = 같은 좌표에서 더 안 바뀌는 순서. 커서가 그대로인데 순서가 바뀌면 포인터가 움직일 때마다 화면이 뒤집힌다.
// 넓은 타일은 옮긴 결과로 줄이 바뀌어 왕복이 생긴다 -> 왕복 안의 순서 중 끄는 타일이 가장 앞인 것으로 고정(어디서 들어와도 같은 답).
export const memberDropOrder = (
    tiles: readonly MemberTile[], key: string, p: Point, f: MemberFrame,
): readonly MemberTile[] => {
    const seen = [tiles]
    for (;;) {
        const cur = seen[seen.length - 1]!, next = dropOnce(cur, key, p, f)
        if (sameOrder(next, cur)) return cur
        const at = seen.findIndex(o => sameOrder(o, next))
        if (at < 0) {
            seen.push(next)
            continue
        }
        const rank = (o: readonly MemberTile[]) => o.findIndex(t => t.key === key)
        return seen.slice(at).reduce((best, o) => {
            const d = rank(o) - rank(best)
            return d < 0 || (d === 0 && orderKey(o) < orderKey(best)) ? o : best
        })
    }
}
