import type { Area, Capacity, CellSpan, TileSize } from './tile.type';

// 2분할 기준치: Grid 영역이 이 값 이상인 축만 2칸 (REF-node-ui-projection.md)
export const TILE_THRESHOLD: Readonly<Area> = { w: 728, h: 560 };

// area = Grid 영역(타일 공간). 스크롤바 두께는 빼지 않는다 — 빼면 스크롤바 등장/소멸에 따라 칸 수가 진동
export const capacity = (area: Area, th: Area = TILE_THRESHOLD): Capacity => ({
	x: area.w >= th.w ? 2 : 1,
	y: area.h >= th.h ? 2 : 1,
});

// 1칸 축에 저장된 0.5는 이 화면에서만 1.0으로 본다 (저장값은 그대로)
export const effectiveSize = (size: TileSize, cap: Capacity): TileSize => ({
	w: cap.x === 2 ? size.w : 1,
	h: cap.y === 2 ? size.h : 1,
});

// 새 타일 크기 = 이 화면의 최소 단위
export const minSize = (cap: Capacity): TileSize => ({
	w: cap.x === 2 ? 0.5 : 1,
	h: cap.y === 2 ? 0.5 : 1,
});

// Grid 안 스냅 오프셋(px): 가로 2칸 + 2칼럼 폭이면 칼럼 단위, 아니면 Grid 단위
export const columnStops = (gridWidth: number, cap: Capacity, tileGap: number, cols: CellSpan = 2): Array<number> =>
	cap.x === 2 && cols === 2 ? [0, (gridWidth + tileGap) / 2] : [0];

// 가로 구간(px). 좌표 = strip 스크롤 내용 기준(padding 포함)
export interface StripSpan { x: number; w: number }
export interface StripColumn extends StripSpan { grid: number }
export interface StripGeometry {
	grids: Array<StripSpan>;
	columns: Array<StripColumn>;    // 칼럼 = 스냅 한 칸 (가로 1칸 화면에선 Grid 하나)
}

// Grid 나열 기하. TileStrip CSS와 같은 가정: padding = Grid 간격 = gap, 전체 폭 Grid = areaW, 1칼럼 Grid = (areaW - gap) / 2
export const stripGeometry = (gridCols: ReadonlyArray<CellSpan>, areaW: number, cap: Capacity, gap: number): StripGeometry => {
	const half = (areaW - gap) / 2;
	const grids: Array<StripSpan> = [];
	const columns: Array<StripColumn> = [];
	let x = gap;
	gridCols.forEach((cols, grid) => {
		const w = cols === 1 ? half : areaW;
		const stops = columnStops(areaW, cap, gap, cols);
		grids.push({ x, w });
		stops.forEach(s => columns.push({ x: x + s, w: stops.length === 2 ? half : w, grid }));
		x += w + gap;
	});
	return { grids, columns };
};

// 모두 0부터: grid = 가장 많이 보이는 Grid(동점이면 뒤쪽), col = 스냅 기준 왼쪽 칼럼, from~to = 절반 이상 보이는 칼럼
export interface StripView { grid: number; col: number; from: number; to: number }

const overlap = (s: StripSpan, from: number, to: number) => Math.max(0, Math.min(s.x + s.w, to) - Math.max(s.x, from));

// scrollX = scrollLeft, viewW = strip clientWidth, gap = scroll-padding(스냅 위치 = 칼럼 x - gap)
export const stripView = (geo: StripGeometry, scrollX: number, viewW: number, gap: number): StripView => {
	const end = scrollX + viewW;

	// Grid가 보이는 양 = 칼럼(타일 영역)이 보이는 양의 합. Grid 구간으로 재면 칼럼 사이 gap까지 세어 동점이 깨짐
	const seen = geo.grids.map(() => 0);
	geo.columns.forEach(c => { seen[c.grid] = (seen[c.grid] ?? 0) + overlap(c, scrollX, end); });

	let grid = 0;
	let best = -1;
	seen.forEach((v, i) => {
		// 0.5px 이내는 동점 — 소수 폭 오차로 앞쪽이 이기지 않게
		if (v >= best - 0.5) { grid = i; best = Math.max(best, v); }
	});

	let col = 0;
	geo.columns.forEach((c, i) => {
		if (Math.abs(c.x - gap - scrollX) < Math.abs(geo.columns[col]!.x - gap - scrollX)) col = i;
	});

	const shown = geo.columns.flatMap((c, i) => overlap(c, scrollX, end) >= c.w / 2 ? [i] : []);
	return {
		grid,
		col,
		from: shown[0] ?? col,
		to: shown[shown.length - 1] ?? col,
	};
};
