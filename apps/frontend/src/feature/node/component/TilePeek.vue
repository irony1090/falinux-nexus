<script setup lang="ts">
import { computed } from 'vue';
import { TILE_GAP, useTileGrids } from '../store/tileGrids.store';

// 미니맵 높이 / 미니맵 사이 / 전체 Grid 미니맵 폭 범위 / 패널 최대 폭
const MH = 26;
const MG = 6;
const MW = { min: 12, max: 52 };
const MAX_W = 560;

// 인디케이터를 위로 당긴 동안의 타일 패널: Grid별 미니맵 + 지금 보이는 범위 틀 + active — REF-board-pager.md
defineProps<{ open: boolean }>();
const { grids, geometry, area, scrollX, viewW, activeId, order } = useTileGrids();

const fullW = computed(() => {
    const { w = 0, h = 0 } = area.value ?? {};
    return Math.max(MW.min, Math.min(MW.max, Math.round(h ? MH * w / h : MW.max)));
})

// 미니맵 칸 = Grid 2x2 기준 비율. 좁은 Grid(칼럼 하나)는 폭도 반
const minis = computed(() => {
    let x = 0;
    return grids.value.map(g => {
        const w = g.cols === 1 ? Math.round((fullW.value - MG) / 2) : fullW.value;
        const m = {
            x,
            w,
            tiles: g.items.map(it => ({
                id: it.tile.id,
                on: it.tile.id === activeId.value,
                style: {
                    left: `${it.c / g.cols * 100}%`,
                    top: `${it.r * 50}%`,
                    width: `${it.cw / g.cols * 100}%`,
                    height: `${it.ch * 50}%`,
                },
            })),
        };
        x += w + MG;
        return m;
    });
})
const total = computed(() => {
    const last = minis.value[minis.value.length - 1];
    return last ? last.x + last.w : 0;
})

// strip 좌표 -> 미니맵 좌표. Grid 안은 비례, Grid 사이 간격은 다음 미니맵까지 선형
const toMini = (sx: number) => {
    const gs = geometry.value.grids, ms = minis.value;
    for (let i = 0; i < gs.length; i++) {
        const g = gs[i]!, m = ms[i]!;
        if (sx <= g.x) return i ? m.x : 0;
        if (sx <= g.x + g.w) return m.x + (sx - g.x) / g.w * m.w;
        const ng = gs[i + 1], nm = ms[i + 1];
        if (ng && nm && sx < ng.x) return m.x + m.w + (sx - g.x - g.w) / (ng.x - g.x - g.w) * (nm.x - m.x - m.w);
    }
    return total.value;
}
const frame = computed(() => {
    const l = toMini(scrollX.value + TILE_GAP), r = toMini(scrollX.value + viewW.value - TILE_GAP);
    return { l, r };
})

// 패널보다 길면 보이는 범위 틀 주변만, 더 있는 쪽 끝은 흐리게
const clipW = computed(() => Math.min(total.value, Math.max(120, Math.min(viewW.value - 32, MAX_W))));
const off = computed(() => {
    const over = total.value - clipW.value;
    if (over <= 0) return 0;
    const { l, r } = frame.value;
    return Math.max(0, Math.min(over, (l + r) / 2 - clipW.value / 2));
})
const clipStyle = computed(() => {
    const left = off.value > 0.5, right = off.value < total.value - clipW.value - 0.5;
    const mask = `linear-gradient(to right, ${left ? 'transparent' : 'black'}, black 28px, black calc(100% - 28px), ${right ? 'transparent' : 'black'})`;
    return { width: `${clipW.value}px`, maskImage: mask, WebkitMaskImage: mask };
})

const label = computed(() => {
    const i = order.value.findIndex(o => o.tile.id === activeId.value);
    return i < 0 ? '' : `타일 ${i + 1} / ${order.value.length}`;
})
</script>

<template>
<div class="TilePeek" :class="{ open }" aria-live="polite">
    <div class="clip" :style="clipStyle">
        <div class="row" :style="{ width: `${total}px`, transform: `translateX(${-off}px)` }">
            <div v-for="(m, gi) in minis" :key="gi" class="mini" :style="{ left: `${m.x}px`, width: `${m.w}px` }">
                <span v-for="t in m.tiles" :key="t.id" class="cell" :class="{ on: t.on }" :style="t.style" />
            </div>
            <div class="frame" :style="{ left: `${frame.l - 3}px`, width: `${frame.r - frame.l + 6}px` }" />
        </div>
    </div>
    <span class="text-body-small">{{ label }}</span>
</div>
</template>

<style lang="scss">
// teleport로 body에 붙으므로 scoped 밖. 위치(bottom·left)는 GridPager가 준다
.TilePeek {
    position: fixed;
    z-index: 2100;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    padding: 10px 12px 8px;
    border-radius: 12px;
    background: rgb(var(--v-theme-surface));
    border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.18);
    pointer-events: none;
    opacity: 0;
    transform: translate(-50%, 8px) scale(0.96);
    transform-origin: 50% 100%;
    transition: opacity 0.2s, transform 0.2s;
    &.open {
        opacity: 1;
        transform: translate(-50%, 0) scale(1);
    }
    .clip {
        overflow: hidden;
        padding: 4px 0;
    }
    .row {
        position: relative;
        height: 26px;
    }
    .mini {
        position: absolute;
        top: 0;
        height: 100%;
        box-sizing: border-box;
        border-radius: 3px;
        border: 1px solid rgba(var(--v-theme-on-surface), 0.25);
    }
    .cell {
        position: absolute;
        box-sizing: border-box;
        border: 1px solid rgb(var(--v-theme-surface));
        border-radius: 2px;
        background: rgba(var(--v-theme-on-surface), 0.15);
        transition: background 0.15s;
        &.on { background: rgb(var(--v-theme-primary)); }
    }
    .frame {
        position: absolute;
        top: -4px;
        bottom: -4px;
        box-sizing: border-box;
        border: 2px solid rgb(var(--v-theme-primary));
        border-radius: 5px;
    }
}
</style>
