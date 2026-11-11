import type { Capacity, CellSpan, FlowPlan, GridItem, GridSlot, TileSize } from './tile.type';

// 화면 순서 = 문서 order 그대로, 연 곳(parent)과 무관 (REF-tile-drag.md 구조-1(순서 배열))
export function ordered<T>(order: ReadonlyArray<string>, tiles: Readonly<Record<string, T>>): Array<T> {
	return order.flatMap(id => tiles[id] ? [tiles[id]] : []);
}

// 칸 번호 = 채우는 순서(위->아래, 다음 칼럼): c*2 + r
const at = (c: number, r: number) => c * 2 + r;

// 타일 크기 -> 칸 수. 가로 1칸 화면은 칼럼 하나 = 화면 폭이라 늘 1칼럼, 세로는 판이 늘 2행이라 1.0 = 2행
export const spanOf = (size: TileSize, cap: Capacity): { cw: CellSpan; ch: CellSpan } => ({
	cw: cap.x === 2 && size.w === 1 ? 2 : 1,
	ch: cap.y === 1 || size.h === 1 ? 2 : 1,
});

function fits(occ: Array<boolean>, c: number, r: number, cw: CellSpan, ch: CellSpan) {
	if (r + ch > 2) return false;
	for (let i = 0; i < cw; i++)
		for (let j = 0; j < ch; j++)
			if (occ[at(c + i, r + j)]) return false;
	return true;
}

// 순서대로 이어진 칼럼에 채운다 — Grid 경계 없음(칼럼-3(채우는 순서)), 앞으로만 채움(건너뛴 칸은 빈칸) = ⑧(같은 Grid 앞쪽 빈칸 채우기) 그대로
export function pack<T>(list: Array<T>, spans: (tile: T) => { cw: CellSpan; ch: CellSpan }): FlowPlan<T> {
	const occ: Array<boolean> = [];
	const items: Array<GridItem<T>> = [];
	let cursor = -1;
	for (const tile of list) {
		const { cw, ch } = spans(tile);
		let i = cursor + 1;
		while (!fits(occ, Math.floor(i / 2), i % 2, cw, ch)) i++;
		cursor = i;
		const c = Math.floor(i / 2), r = (i % 2) as 0 | 1;
		for (let a = 0; a < cw; a++)
			for (let b = 0; b < ch; b++)
				occ[at(c + a, r + b)] = true;
		items.push({ tile, c, r, cw, ch });
	}
	const cols = items.reduce((n, it) => Math.max(n, it.c + it.cw), 0);
	return { items, empty: emptySlots(occ, cols), cols };
}

// 쓴 칼럼 안의 빈 칸 — 한 칼럼이 통째로 비면 슬롯 하나
function emptySlots(occ: Array<boolean>, cols: number): Array<GridSlot> {
	const out: Array<GridSlot> = [];
	for (let c = 0; c < cols; c++) {
		const top = !occ[at(c, 0)], bottom = !occ[at(c, 1)];
		if (top && bottom) out.push({ c, r: 0, cw: 1, ch: 2 });
		else if (top) out.push({ c, r: 0, cw: 1, ch: 1 });
		else if (bottom) out.push({ c, r: 1, cw: 1, ch: 1 });
	}
	return out;
}

// 칸 (c, r) 앞에서 채우는 순서상 가장 늦게 시작한 타일 — 빈 칸에 놓기(칼럼-4)
export function lastBefore<T>(items: ReadonlyArray<GridItem<T>>, c: number, r: number, skip: (tile: T) => boolean = () => false): T | undefined {
	let best: GridItem<T> | undefined;
	for (const it of items)
		if (!skip(it.tile) && at(it.c, it.r) < at(c, r) && (!best || at(it.c, it.r) > at(best.c, best.r))) best = it;
	return best?.tile;
}
