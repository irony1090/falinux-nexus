import type { Cell, CellSpan, GridItem, GridPlan, GridSlot, TileNode, TileSize } from './tile.type';

export type OrderedTile<T> = { tile: T; depth: number };

// 폴더 -> 직접 실행한 프로세스 -> 넘겨받은 프로세스 -> 연 하위 폴더(재귀), 묶음 안은 kids 순 (REF-node-ui-layout.md "타일 순서 규칙")
export function flatten<T extends TileNode>(tiles: Record<string, T>, rootId: string): Array<OrderedTile<T>> {
	const out: Array<OrderedTile<T>> = [];
	const walk = (id: string, depth: number) => {
		const tile = tiles[id];
		if (!tile) return;
		out.push({ tile, depth });
		const kids = tile.kids.map(k => tiles[k]).filter((k): k is T => !!k);
		const terminals = kids.filter(k => k.type === 'terminal');
		[...terminals.filter(k => !k.adopted), ...terminals.filter(k => k.adopted)]
			.forEach(k => out.push({ tile: k, depth: depth + 1 }));
		kids.filter(k => k.type === 'folder').forEach(k => walk(k.id, depth + 1));
	};
	walk(rootId, 0);
	return out;
}

// 칸 찾는 순서 [c, r] = 위->아래, 다음 칼럼 (모든 기기 공통)
const READ: ReadonlyArray<[Cell, Cell]> = [[0, 0], [0, 1], [1, 0], [1, 1]];

type Packing<T> = { occ: Array<boolean>; items: Array<GridItem<T>> };

const at = (c: number, r: number) => c * 2 + r;
const span = (v: number): CellSpan => v === 1 ? 2 : 1;

function fits(occ: Array<boolean>, c: Cell, r: Cell, cw: CellSpan, ch: CellSpan) {
	if (c + cw > 2 || r + ch > 2) return false;
	for (let i = 0; i < cw; i++)
		for (let j = 0; j < ch; j++)
			if (occ[at(c + i, r + j)]) return false;
	return true;
}

// 순서대로 2×2칸에 채운다. 안 들어가면 그 타일부터 다음 Grid (이전 Grid로 돌아가지 않음 = 순서 유지)
export function pack<T>(list: Array<T>, sizeOf: (tile: T) => TileSize): Array<GridPlan<T>> {
	const grids: Array<Packing<T>> = [];
	for (const tile of list) {
		const size = sizeOf(tile);
		const cw = span(size.w), ch = span(size.h);
		const last = grids[grids.length - 1];
		// 같은 Grid 안에선 앞쪽 빈칸도 채움 — 열린 질문 ⑧(앞쪽 빈칸 채우기)의 시안 가정
		const spot = last && READ.find(([c, r]) => fits(last.occ, c, r, cw, ch));
		const g = last && spot ? last : newPacking(grids);
		const [c, r] = spot ?? [0, 0];
		for (let i = 0; i < cw; i++)
			for (let j = 0; j < ch; j++)
				g.occ[at(c + i, r + j)] = true;
		g.items.push({ tile, c, r, cw, ch });
	}
	return grids.map(toPlan);
}

function newPacking<T>(grids: Array<Packing<T>>): Packing<T> {
	const g: Packing<T> = { occ: [false, false, false, false], items: [] };
	grids.push(g);
	return g;
}

function toPlan<T>({ occ, items }: Packing<T>): GridPlan<T> {
	const spanC = items.some(it => it.cw === 2);
	const spanR = items.some(it => it.ch === 2);
	const axis = spanC && spanR ? null : spanC ? 'row' : 'column';
	const cols: CellSpan = occ[at(1, 0)] || occ[at(1, 1)] ? 2 : 1;
	return { items, empty: emptySlots(occ, axis, cols), cols, axis };
}

// Band 단위로 빈칸을 모은다: 통째로 빈 Band는 슬롯 하나, 아니면 빈 칸마다 하나
function emptySlots(occ: Array<boolean>, axis: GridPlan<unknown>['axis'], cols: CellSpan): Array<GridSlot> {
	if (!axis) return [];
	const out: Array<GridSlot> = [];
	const bands: Array<Cell> = axis === 'column' && cols === 1 ? [0] : [0, 1];
	for (const b of bands) {
		const cells: Array<[Cell, Cell]> = axis === 'row' ? [[0, b], [1, b]] : [[b, 0], [b, 1]];
		const free = cells.filter(([c, r]) => !occ[at(c, r)]);
		if (free.length === 2)
			out.push(axis === 'row' ? { c: 0, r: b, cw: 2, ch: 1 } : { c: b, r: 0, cw: 1, ch: 2 });
		else
			free.forEach(([c, r]) => out.push({ c, r, cw: 1, ch: 1 }));
	}
	return out;
}
