import type { Area, Capacity, GridItem, TileSize } from './tile.type';
import { lastBefore, pack, spanOf } from './tileOrder.util';
import { colWidth, flowWidth } from './tileSplit.util';

export type Rect = { x: number; y: number; w: number; h: number };
export type DropPoint = { x: number; y: number };

const inside = (r: Rect, p: DropPoint) => p.x >= r.x && p.x < r.x + r.w && p.y >= r.y && p.y < r.y + r.h;

// 배치 계산 위치. 판 안은 TileStrip CSS와 같은 가정(칼럼 폭 = colWidth, 행 = 2)
// origin = 판 왼쪽 위(p와 같은 좌표계, 판 폭(px)에 따라 달라질 수 있음 — 전체보기 가운데 정렬), scale = 전체보기 축소 비율(평소 1)
export type DropFrame = { area: Area; cap: Capacity; gap: number; scale: number; origin: (flowW: number) => DropPoint };

type Placed = { id: string; rect: Rect };
type Laid<T> = { tiles: Array<Placed>; flow: Rect; cell: (p: DropPoint) => { c: number; r: number }; items: Array<GridItem<T>> };

const layOut = <T extends { id: string; size: TileSize }>(list: Array<T>, f: DropFrame): Laid<T> => {
    const plan = pack(list, t => spanOf(t.size, f.cap));
    const w = flowWidth(plan.cols, f.area.w, f.cap, f.gap);
    const o = f.origin(w * f.scale);
    const colW = colWidth(f.area.w, f.cap, f.gap), rowH = (f.area.h - f.gap) / 2, s = f.scale;
    const tiles = plan.items.map(it => ({
        id: it.tile.id,
        rect: {
            x: o.x + it.c * (colW + f.gap) * s, y: o.y + it.r * (rowH + f.gap) * s,
            w: (it.cw * colW + (it.cw - 1) * f.gap) * s, h: (it.ch * rowH + (it.ch - 1) * f.gap) * s,
        },
    }));
    const cell = (p: DropPoint) => ({
        c: Math.floor((p.x - o.x) / ((colW + f.gap) * s)),
        r: Math.min(1, Math.floor((p.y - o.y) / ((rowH + f.gap) * s))),
    });
    return { tiles, flow: { x: o.x, y: o.y, w: w * s, h: f.area.h * s }, cell, items: plan.items };
};

type Laying = { id: string; size: TileSize; minimized?: true };
const visibleOf = <T extends Laying>(o: ReadonlyArray<string>, tiles: Readonly<Record<string, T>>) =>
    o.flatMap(x => tiles[x] && !tiles[x]!.minimized ? [tiles[x]!] : []);

// order(문서 전체, 최소화 포함)에서 ids를 빼고 at 앞에 그 순서대로 넣는다(at 없음 = 맨 뒤)
const insert = (order: ReadonlyArray<string>, ids: ReadonlyArray<string>, at: string | null) => {
    const rest = order.filter(x => !ids.includes(x));
    const i = at === null ? rest.length : rest.indexOf(at);
    rest.splice(i, 0, ...ids);
    return rest;
};
const insertAfter = (order: ReadonlyArray<string>, ids: ReadonlyArray<string>, y: string) => {
    const rest = order.filter(x => !ids.includes(x));
    return insert(order, ids, rest[rest.indexOf(y) + 1] ?? null);
};

// 끄는 중 미리보기 순서(같은 보드, 끄는 타일 하나) — 흔들림 방지 규칙 ①②③(REF-tile-drag.md):
// ① 판정은 배치 계산 위치로 ② 커서가 끄는 타일(빈 자리) 위면 그대로
// ③ 옮긴 결과에서 끄는 타일이 커서를 덮을 때만 옮긴다(아니면 반대쪽, 둘 다 아니면 그대로) — 크기가 섞여 있으면 덮지 않는 자리로 갔다가 되돌아오며 흔들린다
// 빈 칸·판 밖은 ③ 없이 옮긴다 — 빈 칸 = 채우는 순서상 그 칸 앞 마지막 타일 뒤(칼럼-4(빈 칸에 놓기)), 옮긴 뒤 그 타일이 다시 같은 답이라 흔들리지 않는다. root = 맨 앞 고정
export const dropOrder = <T extends Laying>(
    order: ReadonlyArray<string>, tiles: Readonly<Record<string, T>>, id: string, rootId: string, p: DropPoint, f: DropFrame,
): ReadonlyArray<string> => {
    const now = layOut(visibleOf(order, tiles), f);
    const covers = (o: Array<string>) => {
        const r = layOut(visibleOf(o, tiles), f).tiles.find(t => t.id === id)?.rect;
        return !!r && inside(r, p);
    };
    const pick = (...cands: Array<Array<string>>) => cands.find(c => c.join() !== order.join() && covers(c)) ?? order;

    const over = now.tiles.find(t => inside(t.rect, p));
    if (over) {
        if (over.id === id || over.id === rootId) return order;
        const vis = now.tiles.map(t => t.id);
        const forward = vis.indexOf(over.id) > vis.indexOf(id);
        const before = insert(order, [id], over.id), behind = insertAfter(order, [id], over.id);
        return forward ? pick(behind, before) : pick(before, behind);
    }
    // 빈 칸 = 그 칸 앞 마지막 타일 뒤, 판 오른쪽 너머 = 맨 뒤
    if (inside(now.flow, p)) {
        const { c, r } = now.cell(p);
        const before = lastBefore(now.items, c, r)?.id;
        return before && before !== id ? insertAfter(order, [id], before) : order;
    }
    if (now.flow.w && p.x >= now.flow.x + now.flow.w) return insert(order, [id], null);
    return order;
};

// 다른 보드 줄 위 미리보기 — 묶음(끄는 타일 + 그 타일이 연 타일, 출발 순서) 전체를 넣는다. 보드-1(보드 간 이동 API)
// base = 그 보드의 원래 순서(묶음 없음), cur = 지금 미리보기. 규칙 ①②는 dropOrder와 같고 ③은 묶음 단위, 묶음 타일 위 = 그대로,
// 타일 밖(빈 칸 말고 줄의 빈 곳) = 맨 뒤(보드Q3(빈 곳에 놓기))
export const dropBundle = <T extends Laying>(
    base: ReadonlyArray<string>, cur: ReadonlyArray<string>, tiles: Readonly<Record<string, T>>,
    bundle: ReadonlyArray<string>, rootId: string, p: DropPoint, f: DropFrame,
): ReadonlyArray<string> => {
    const now = layOut(visibleOf(cur, tiles), f);
    // ③을 묶음 단위로 — 묶음 어느 타일이든 커서를 덮으면(연 터미널이 폴더보다 앞일 수 있어 끄는 타일만 보면 맞는 자리가 거의 없다). 옮긴 뒤엔 "묶음 위 = 그대로"라 흔들리지 않음
    const covers = (o: Array<string>) => layOut(visibleOf(o, tiles), f).tiles.some(t => bundle.includes(t.id) && inside(t.rect, p));
    const keep = (o: Array<string>) => o.join() === cur.join() ? cur : o;

    const over = now.tiles.find(t => inside(t.rect, p));
    if (over) {
        if (bundle.includes(over.id)) return cur;
        const cands = over.id === rootId ? [insertAfter(base, bundle, over.id)] : [insert(base, bundle, over.id), insertAfter(base, bundle, over.id)];
        const hit = cands.find(covers);
        return hit ? keep(hit) : cur.some(x => bundle.includes(x)) ? cur : keep(insert(base, bundle, null));
    }
    if (inside(now.flow, p)) {
        const { c, r } = now.cell(p);
        const before = lastBefore(now.items, c, r, t => bundle.includes(t.id))?.id;
        return before ? keep(insertAfter(base, bundle, before)) : cur;
    }
    return keep(insert(base, bundle, null));
};
