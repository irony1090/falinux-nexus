// 타일 크기 = 화면(타일 영역) 대비 비율 (0.5 = 반 칸, 1 = 화면 폭·높이 전체)
export type TileSpan = 0.5 | 1;
export type TileSize = { w: TileSpan; h: TileSpan };

export type TileType = 'folder' | 'terminal' | 'device' | 'viewer';

// 채우기 유틸이 요구하는 최소 필드. 화면 순서는 문서의 order 배열(REF-tile-drag.md)
export interface TileNode {
	id: string;
	type: TileType;
	size: TileSize;
	minimized?: true;       // 최소화한 터미널 — 트리엔 남고 배치에서만 빠진다(SD3(실행 중 터미널 치우기))
}

export type Area = { w: number; h: number };

// 축별 칸 수: x = 가로, y = 세로
export type Capacity = { x: 1 | 2; y: 1 | 2 };

export type Cell = 0 | 1;
export type CellSpan = 1 | 2;

// 판 위치: c = 칼럼(0부터, 끝없이 이어짐), r = 행, cw/ch = 차지하는 칸 수
export interface GridSlot {
	c: number;
	r: Cell;
	cw: CellSpan;
	ch: CellSpan;
}

export interface GridItem<T> extends GridSlot {
	tile: T;
}

// 보드 배치 = 2행짜리 칼럼이 이어진 판 하나(Grid 경계 없음) — REF-node-ui-projection.md 칼럼-2(Grid 개념 폐지)
export interface FlowPlan<T> {
	items: Array<GridItem<T>>;
	empty: Array<GridSlot>;             // 빈 칸 (한 칼럼이 통째로 비면 슬롯 하나)
	cols: number;                       // 쓴 칼럼 수
}
