// 타일 크기 = Grid 대비 비율 (0.5 = 반 칸, 1 = 한 줄 전체)
export type TileSpan = 0.5 | 1;
export type TileSize = { w: TileSpan; h: TileSpan };

export type TileType = 'folder' | 'terminal';

// 순서·채우기 유틸이 요구하는 최소 필드 (여는 관계 트리의 노드)
export interface TileNode {
	id: string;
	type: TileType;
	kids: ReadonlyArray<string>;    // 이 폴더 타일이 연 타일들, 연 순서대로 (갱신은 배열 교체로)
	size: TileSize;
	adopted?: true;         // 닫힌 폴더에서 넘겨받은 터미널 — 직접 실행한 프로세스 뒤에 온다
}

export type Area = { w: number; h: number };

// 축별 칸 수: x = 가로, y = 세로
export type Capacity = { x: 1 | 2; y: 1 | 2 };

export type Cell = 0 | 1;
export type CellSpan = 1 | 2;

// Grid 안 위치: c = 칼럼, r = 행, cw/ch = 차지하는 칸 수
export interface GridSlot {
	c: Cell;
	r: Cell;
	cw: CellSpan;
	ch: CellSpan;
}

export interface GridItem<T> extends GridSlot {
	tile: T;
}

export interface GridPlan<T> {
	items: Array<GridItem<T>>;
	empty: Array<GridSlot>;             // 빈 칸 (통째로 빈 Band는 슬롯 하나)
	cols: CellSpan;                     // 오른쪽 칼럼이 통째로 비면 1
	axis: 'row' | 'column' | null;      // Band 방향, 표시용 (null = 1×1 타일 하나)
}
