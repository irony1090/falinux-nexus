import { shallowReactive } from 'vue'

// 타일 머리 줄 창 버튼(숨기기·닫기) — 컨트롤 패널이 active 타일 것을 그대로 부른다(파트(패널 창 버튼)).
// 등록 = TileFrame(종류별 동작은 종류 컴포넌트가 정하므로 머리 줄과 같은 emit을 넘김). 없는 버튼은 undefined
export type TileHeadActs = {
    minimize?: () => void
    sizable: boolean
    close?: () => void
    closeTitle: string
    closeDisabled: boolean
}

const acts = shallowReactive(new Map<string, TileHeadActs>())

export const setTileHeadActs = (tileId: string, a: TileHeadActs) => { acts.set(tileId, a); }
// 자기가 등록한 것만 지운다 — 타일이 다른 Grid로 옮겨지면 새 TileFrame이 먼저 등록하고 옛 것이 나중에 언마운트되므로,
// 무조건 지우면 새 등록이 사라져 패널에 버튼이 안 나온다(2026-10-28 실제로 겪음)
export const clearTileHeadActs = (tileId: string, mine: TileHeadActs | undefined) => {
    if (acts.get(tileId) === mine) acts.delete(tileId);
}
export const tileHeadActs = (tileId: string | null | undefined) => tileId ? acts.get(tileId) : undefined
