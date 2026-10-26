// 관리 페이지(내 장비·관리자) 타일 격자 — REF-member-ui.md 화-5(상세 쌓기)~화-11(경로), 주소-3(형식). 보드 tileOrder.util.ts와 따로 둔다

// 최소 칸 / 간격 / 상세 최대 열·행 / 전체 열 최대(목록 1 + 상세 3)
export const CELL_MIN_W = 364
export const CELL_MIN_H = 280
export const GAP = 8
export const MAX_COLS = 3
export const MAX_ROWS = 2

// 타일 키 = 종류 + id: 회원 `m<id>` / 장비 `d<메인키>` / 등록 `new`
export type MemberTileKind = 'user' | 'device' | 'new'
export type MemberTile = {
    key: string
    w: number
    h: number
    parent: string | null
}

export const kindOf = (key: string): MemberTileKind => key === 'new' ? 'new' : key.startsWith('m') ? 'user' : 'device'
export const userKey = (id: number) => `m${id}`
export const deviceKey = (mainKey: string) => `d${mainKey}`
export const userIdOf = (key: string) => Number(key.slice(1))
export const mainKeyOf = (key: string) => key.slice(1)

// 처음 크기 = 종류별 고정(화-10): 장비 = 세로 2칸(인스턴스 포함), 나머지 1칸
export const newTile = (key: string, parent: string | null = null): MemberTile => ({ key, w: 1, h: kindOf(key) === 'device' ? 2 : 1, parent })

// 화-6(Grid 크기): area = 푸터 위 전체 영역. 전체 열 = 목록 1 + 상세, 1열이면 폰 방식(목록 = 가로 줄 첫 타일)
export type MemberLayout = { phone: boolean; cols: number; rows: number; cellW: number }
export const layoutOf = (w: number, h: number): MemberLayout => {
    const total = Math.max(1, Math.min(MAX_COLS + 1, Math.floor((w - GAP) / (CELL_MIN_W + GAP))))
    const phone = total === 1
    const rows = Math.max(1, Math.min(MAX_ROWS, Math.floor((h - GAP) / (CELL_MIN_H + GAP))))
    const cellW = phone ? w : Math.floor((w - 2 * GAP - GAP * (total - 1)) / total)
    return { phone, cols: phone ? 1 : total - 1, rows, cellW }
}

export type Placed = { tile: MemberTile; c: number; r: number; cw: number; ch: number }
export type MemberGrid = { items: Placed[] }

// 화-7(채우는 순서): 한 열을 위에서 아래로, 다음 열. Grid 안에서 앞으로만 채우고(앞 빈칸으로 안 돌아감) 안 맞으면 새 Grid
export const pack = (tiles: readonly MemberTile[], cols: number, rows: number): MemberGrid[] => {
    const read: Array<[number, number]> = []
    for (let c = 0; c < cols; c++) for (let r = 0; r < rows; r++) read.push([c, r])
    const grids: Array<MemberGrid & { occ: boolean[]; cursor: number }> = []
    for (const tile of tiles) {
        const cw = Math.min(tile.w, cols), ch = Math.min(tile.h, rows)
        const fits = (occ: boolean[], c: number, r: number) => {
            if (c + cw > cols || r + ch > rows) return false
            for (let i = 0; i < cw; i++) for (let j = 0; j < ch; j++) if (occ[(c + i) * rows + r + j]) return false
            return true
        }
        let g = grids[grids.length - 1]
        let idx = g ? read.findIndex(([c, r], i) => i > g!.cursor && fits(g!.occ, c, r)) : -1
        if (!g || idx < 0) {
            g = { occ: new Array(cols * rows).fill(false), items: [], cursor: -1 }
            grids.push(g)
            idx = read.findIndex(([c, r]) => fits(g!.occ, c, r))
        }
        g.cursor = idx
        const [c, r] = read[idx]!
        for (let i = 0; i < cw; i++) for (let j = 0; j < ch; j++) g.occ[(c + i) * rows + r + j] = true
        g.items.push({ tile, c, r, cw, ch })
    }
    return grids.map(({ items }) => ({ items }))
}

const isDescendant = (tiles: readonly MemberTile[], t: MemberTile, ancestor: string) => {
    let p = t.parent
    for (let guard = 0; p && guard < 20; guard++) {
        if (p === ancestor) return true
        p = tiles.find(x => x.key === p)?.parent ?? null
    }
    return false
}

// 화-9(자식 자리): 부모 바로 다음(부모의 앞선 자식들 뒤). 부모가 없으면 끝
export const insertTile = (tiles: readonly MemberTile[], tile: MemberTile): MemberTile[] => {
    let at = tiles.length
    const pi = tile.parent ? tiles.findIndex(x => x.key === tile.parent) : -1
    if (pi >= 0) {
        at = pi + 1
        while (at < tiles.length && isDescendant(tiles, tiles[at]!, tile.parent!)) at++
    }
    return [...tiles.slice(0, at), tile, ...tiles.slice(at)]
}

// 화-11(경로): 닫힌 부모를 다시 열 땐 그 자식 바로 앞
export const insertBefore = (tiles: readonly MemberTile[], tile: MemberTile, beforeKey: string): MemberTile[] => {
    const at = tiles.findIndex(x => x.key === beforeKey)
    if (at < 0) return [...tiles, tile]
    return [...tiles.slice(0, at), tile, ...tiles.slice(at)]
}

// 주소-3(형식): `m2*1x1,drig-14*1x2^m2,new*1x1`. 메인키 글자([A-Za-z0-9._-])와 안 겹치는 구분자만 쓴다
const KEY_RE = /^(m\d+|d[A-Za-z0-9._-]{1,64}|new)$/
export const encodeTiles = (tiles: readonly MemberTile[]) => tiles.map(t => `${t.key}*${t.w}x${t.h}${t.parent ? `^${t.parent}` : ''}`).join(',')
export const decodeTiles = (s: string | undefined): MemberTile[] => {
    if (!s) return []
    const out: MemberTile[] = []
    for (const part of s.split(',')) {
        const [body = '', parent] = part.split('^')
        const [key = '', size = ''] = body.split('*')
        if (!KEY_RE.test(key) || out.some(t => t.key === key)) continue
        const m = /^(\d)x(\d)$/.exec(size)
        const base = newTile(key, parent && KEY_RE.test(parent) ? parent : null)
        out.push(m ? { ...base, w: Math.max(1, +m[1]!), h: Math.max(1, +m[2]!) } : base)
    }
    return out
}
