import type { TileType } from './tile.type';

// 타일 종류 명세 — 여러 종류가 같이 쓰는 로직은 종류를 비교하지 말고 능력으로 묻는다. 새 종류 = TILE_KINDS 줄 하나
// 서버 tileKind.go와 같은 표를 유지한다 (REF-tile-kinds.md)
export interface TileKind {
	canRoot: boolean;       // 능력-1(루트 가능): 보드 루트·새 보드로 분리
	canParent: boolean;     // 능력-2(자식 가능): 다른 타일을 연다(연 곳 = parent가 될 수 있음)
	holdsProc: boolean;     // 능력-3(process 담음): uid = process uid, 최소화·단축키 끔·종료
	icon: string;
	color: string;          // Vuetify 테마 색 이름
}

export const TILE_KINDS: Record<TileType, TileKind> = {
	folder: { canRoot: true, canParent: true, holdsProc: false, icon: 'mdi-folder', color: 'warning' },
	terminal: { canRoot: false, canParent: false, holdsProc: true, icon: 'mdi-console', color: 'primary' },
	device: { canRoot: true, canParent: true, holdsProc: false, icon: 'mdi-folder-network', color: 'info' },   // 장비 타일 — worker 디렉토리(폴더-2(장비 타일))
	viewer: { canRoot: false, canParent: false, holdsProc: false, icon: 'mdi-image-outline', color: 'success' },   // 보기 타일 — 장비 파일 하나(파트(보기 타일))
};

type Typed = { type: TileType };

export const canRoot = (t: Typed) => TILE_KINDS[t.type].canRoot;
export const canParent = (t: Typed) => TILE_KINDS[t.type].canParent;
export const holdsProc = (t: Typed) => TILE_KINDS[t.type].holdsProc;
