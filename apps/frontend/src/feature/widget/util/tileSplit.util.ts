import type { Area, Capacity, TileSize } from './tile.type';

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

// 칼럼 폭(px): 가로 2칸 = 화면 반, 1칸 = 화면 전체
export const colWidth = (areaW: number, cap: Capacity, gap: number) => cap.x === 2 ? (areaW - gap) / 2 : areaW;
// 칼럼 cols개를 이은 판 폭(px)
export const flowWidth = (cols: number, areaW: number, cap: Capacity, gap: number) =>
	Math.max(0, cols * colWidth(areaW, cap, gap) + (cols - 1) * gap);

// 가로 구간(px). 좌표 = strip 스크롤 내용 기준(padding 포함)
export interface StripSpan { x: number; w: number }
// 칼럼 = 스냅 한 칸. 판이 이어져 있어 폭·간격이 모두 같다
export interface StripGeometry { columns: Array<StripSpan> }

// 칼럼 나열 기하. TileStrip CSS와 같은 가정: padding = 칼럼 간격 = gap
export const stripGeometry = (cols: number, areaW: number, cap: Capacity, gap: number): StripGeometry => {
	const w = colWidth(areaW, cap, gap);
	return { columns: Array.from({ length: cols }, (_, c) => ({ x: gap + c * (w + gap), w })) };
};

// 모두 0부터: col = 스냅 기준 왼쪽 칼럼, from~to = 절반 이상 보이는 칼럼
export interface StripView { col: number; from: number; to: number }

const overlap = (s: StripSpan, from: number, to: number) => Math.max(0, Math.min(s.x + s.w, to) - Math.max(s.x, from));

// scrollX = scrollLeft, viewW = strip clientWidth, gap = scroll-padding(스냅 위치 = 칼럼 x - gap)
export const stripView = (geo: StripGeometry, scrollX: number, viewW: number, gap: number): StripView => {
	const end = scrollX + viewW;
	let col = 0;
	geo.columns.forEach((c, i) => {
		if (Math.abs(c.x - gap - scrollX) < Math.abs(geo.columns[col]!.x - gap - scrollX)) col = i;
	});
	const shown = geo.columns.flatMap((c, i) => overlap(c, scrollX, end) >= c.w / 2 ? [i] : []);
	return { col, from: shown[0] ?? col, to: shown[shown.length - 1] ?? col };
};

// 스크롤 위치 -> 칼럼 단위 실수 위치. 칼럼 시작 사이는 선형, 양 끝 밖은 끝값
export const stripPos = (geo: StripGeometry, scrollX: number, gap: number): number => {
	const cols = geo.columns, t = scrollX + gap;
	if (!cols.length) return 0;
	for (let i = 0; i < cols.length - 1; i++) {
		const a = cols[i]!, b = cols[i + 1]!;
		if (t < b.x) return i + Math.max(0, t - a.x) / (b.x - a.x);
	}
	return cols.length - 1;
};

// 칼럼 번호(실수) -> scrollLeft. 인디케이터 드래그 중 칼럼 사이 위치
export const colScroll = (geo: StripGeometry, col: number, gap: number): number => {
	const cols = geo.columns;
	const i = Math.max(0, Math.min(cols.length - 1, Math.floor(col)));
	const a = cols[i], b = cols[i + 1];
	if (!a) return 0;
	return (b ? a.x + (col - i) * (b.x - a.x) : a.x) - gap;
};

// 타일 칼럼 [l, r]이 화면 밖이면 막 다 보이는 만큼만 옮긴 왼쪽 칼럼 — 오른쪽 밖 = 오른쪽 끝에, 왼쪽 밖 = 왼쪽 끝에.
// 화면 폭 이상(1.0 타일)이면 그 타일 시작. span = 화면에 들어가는 칼럼 수(cap.x) — REF-board-pager.md "타일 위치 규칙"
export const revealCol = (l: number, r: number, view: number, span: number, max: number): number => {
	let v = view;
	if (r - l + 1 >= span || l < view) v = l;
	else if (r > view + span - 1) v = r - span + 1;
	return Math.max(0, Math.min(max, v));
};
