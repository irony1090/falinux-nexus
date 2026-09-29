<script setup lang="ts">
import { useResizeCallback } from '@/feature/common/store/resizeGroup.store';
import { removeUnit } from '@/common/util/index.util';
import Tile from '@/feature/widget/component/Tile.vue';
import TileLayout from '@/feature/widget/component/TileLayout.vue';
import type { Area, CellSpan } from '@/feature/widget/util/tile.type';
import { columnStops } from '@/feature/widget/util/tileSplit.util';
import { TILE_GAP as GAP, useTileGrids } from '../store/tileGrids.store';
import TileFrame from './tile/TileFrame.vue';

const FOOT = 20;    // Grid 하단 정보 줄 높이

const { grids, cap, area, stripEl, scrollX, viewW } = useTileGrids();

// 칼럼 단위 스냅 앵커 위치(px). 전체 폭 Grid의 폭 = 측정한 area.w
const stopsOf = (cols: CellSpan) => columnStops(area.value?.w ?? 0, cap.value, GAP, cols);

// Grid 영역 = strip - padding - 정보 줄. 스크롤바 두께는 빼지 않는다(빼면 스크롤바 등장/소멸에 따라 칸 수가 진동)
useResizeCallback(stripEl, () => {
    const el = stripEl.value;
    if (!el) return;
    const { width, height } = el.getBoundingClientRect();
    const { paddingTop, paddingRight, paddingBottom, paddingLeft } = getComputedStyle(el);
    area.value = {
        w: width - removeUnit(paddingLeft) - removeUnit(paddingRight),
        h: height - removeUnit(paddingTop) - removeUnit(paddingBottom) - FOOT - GAP,
    } satisfies Area;
    viewW.value = el.clientWidth;
    scrollX.value = el.scrollLeft;
})

const onScroll = () => {
    if (stripEl.value) scrollX.value = stripEl.value.scrollLeft;
}
</script>

<template>
<div ref="stripEl" class="TileStrip wk-scollbar"
    :style="{ '--gap': `${GAP}px` }"
    @scroll.passive="onScroll"
>
    <tile-layout v-for="(g, gi) in grids" :key="gi"
        :class="{ narrow: g.cols === 1 }"
        :cols="g.cols"
        :gap="GAP"
        :foot-height="FOOT"
    >
        <tile v-for="it in g.items" :key="it.tile.id" :cell="it">
            <tile-frame :tile-id="it.tile.id" />
        </tile>
        <tile v-for="(e, ei) in g.empty" :key="`empty-${ei}`" :cell="e" empty />

        <template #foot>
            <span class="tag">Grid {{ gi + 1 }}</span>
            <span>{{ g.axis ?? 'single' }} · {{ g.items.length }}타일</span>
        </template>
        <template #overlay>
            <span v-for="x in stopsOf(g.cols)" :key="x" class="snap" :style="{ left: `${x}px` }" />
        </template>
    </tile-layout>
</div>
</template>

<style scoped lang="scss">
.TileStrip {
    // ⑥ Grid 높이: VAppBar(fixed)만큼 VMain이 padding으로 비키므로 그만큼 뺀다
    height: calc(100dvh - var(--v-layout-top, 0px));
    display: flex;
    gap: var(--gap);
    padding: var(--gap);
    overflow-x: auto;
    overflow-y: hidden;
    // 칼럼(Grid 가로 절반) 단위 스냅 — Grid 자체엔 snap-align 없음, 앵커(.snap)만
    scroll-snap-type: x mandatory;
    scroll-padding: var(--gap);
    > .TileLayout {
        flex: 0 0 100%;
        // 오른쪽 칼럼이 통째로 빈 Grid = 칼럼 하나 폭
        &.narrow {
            flex-basis: calc((100% - var(--gap)) / 2);
        }
    }
}
.snap {
    position: absolute;
    top: 0;
    width: 1px;
    height: 1px;
    scroll-snap-align: start;
    pointer-events: none;
}
.tag {
    font-family: monospace;
    padding: 0 5px;
    border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    border-radius: 4px;
}
</style>
