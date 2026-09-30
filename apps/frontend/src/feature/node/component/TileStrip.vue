<script setup lang="ts">
import { useResizeCallback } from '@/feature/common/store/resizeGroup.store';
import { removeUnit } from '@/common/util/index.util';
import Tile from '@/feature/widget/component/Tile.vue';
import TileLayout from '@/feature/widget/component/TileLayout.vue';
import type { Area, CellSpan } from '@/feature/widget/util/tile.type';
import { columnStops } from '@/feature/widget/util/tileSplit.util';
import { computed, onBeforeUnmount, onMounted, watch, type CSSProperties } from 'vue';
import { TILE_FOOT as FOOT, TILE_GAP as GAP, useTileGrids } from '../store/tileGrids.store';
import TileFrame from './tile/TileFrame.vue';

const {
    grids, cap, area, stripEl, scrollX, viewW,
    overview, ovScale, ovDir, scrollbar, leaveOverview,
} = useTileGrids();

// 칼럼 단위 스냅 앵커 위치(px). 전체 폭 Grid의 폭 = 측정한 area.w. 전체보기에선 앵커 없음(스냅 끔)
const stopsOf = (cols: CellSpan) => overview.value ? [] : columnStops(area.value?.w ?? 0, cap.value, GAP, cols);

const syncScrollbar = () => {
    const el = stripEl.value;
    if (el) scrollbar.value = { w: el.offsetWidth - el.clientWidth, h: el.offsetHeight - el.clientHeight };
}

// Grid 영역 = strip - padding - 정보 줄. 스크롤바 두께는 빼지 않는다(빼면 스크롤바 등장/소멸에 따라 칸 수가 진동)
// 전체보기도 strip 크기·padding은 그대로 둔다 — 여기서 재는 값이 바뀌면 "현재 배치 그대로"가 깨짐
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
    syncScrollbar();
})
// Grid 수·모드·배율이 바뀌면 스크롤바가 생기거나 사라진다
watch([grids, overview, ovScale], syncScrollbar, { flush: 'post' });

const onScroll = () => {
    if (stripEl.value) scrollX.value = stripEl.value.scrollLeft;
}

// 가로 한 줄 전체보기에선 세로 휠도 가로로
const onWheel = (e: WheelEvent) => {
    const el = stripEl.value;
    if (!el || !overview.value || ovDir.value !== 'row' || Math.abs(e.deltaY) <= Math.abs(e.deltaX)) return;
    e.preventDefault();
    el.scrollLeft += e.deltaY;
}

const onKey = (e: KeyboardEvent) => {
    if (e.key === 'Escape' && overview.value) leaveOverview();
}
onMounted(() => window.addEventListener('keydown', onKey));
onBeforeUnmount(() => window.removeEventListener('keydown', onKey));

const onTileKey = (e: KeyboardEvent, gi: number) => {
    if (!overview.value || (e.key !== 'Enter' && e.key !== ' ')) return;
    e.preventDefault();
    leaveOverview(gi);
}

// 전체보기: Grid를 실제 px(타일 영역만, 정보 줄 제외)로 그린 뒤 축소, 바깥 상자는 축소 크기만 차지
const gridPx = (cols: CellSpan) => ({
    w: cols === 1 ? ((area.value?.w ?? 0) - GAP) / 2 : area.value?.w ?? 0,
    h: area.value?.h ?? 0,
})
const boxStyle = (cols: CellSpan): CSSProperties | undefined => {
    if (!overview.value) return undefined;
    const { w, h } = gridPx(cols);
    return { width: `${w * ovScale.value}px`, height: `${h * ovScale.value}px` };
}
const layoutStyle = (cols: CellSpan): CSSProperties | undefined => {
    if (!overview.value) return undefined;
    const { w, h } = gridPx(cols);
    return { width: `${w}px`, height: `${h}px`, transform: `scale(${ovScale.value})` };
}

const stripClass = computed(() => overview.value ? ['overview', ovDir.value] : []);
</script>

<template>
<!--
    재마운트 금지: 캡션(v-if)·축소 상자는 두 모드 모두 같은 자리에 두고 스타일만 바꾼다.
    감싸는 요소를 모드별로 바꾸면 전환마다 타일이 새로 생성되어, 임베드될 xterm 화면이 사라진다(화면복원 미구현).
-->
<div ref="stripEl" class="TileStrip wk-scollbar"
    :class="stripClass"
    :style="{ '--gap': `${GAP}px`, '--s': ovScale }"
    @scroll.passive="onScroll"
    @wheel="onWheel"
>
    <div v-for="(g, gi) in grids" :key="gi" class="cell" :class="{ narrow: g.cols === 1 }">
        <div v-if="overview" class="cap"><b>G{{ gi + 1 }}</b> · {{ g.items.length }}타일</div>
        <div class="box" :style="boxStyle(g.cols)">
            <tile-layout
                :cols="g.cols"
                :gap="GAP"
                :foot-height="FOOT"
                :style="layoutStyle(g.cols)"
            >
                <tile v-for="it in g.items" :key="it.tile.id" :cell="it"
                    :tabindex="overview ? 0 : undefined"
                    :role="overview ? 'button' : undefined"
                    @click="overview && leaveOverview(gi)"
                    @keydown="onTileKey($event, gi)"
                >
                    <tile-frame :tile-id="it.tile.id" />
                </tile>
                <tile v-for="(e, ei) in g.empty" :key="`empty-${ei}`" :cell="e" empty />

                <template v-if="!overview" #foot>
                    <span class="tag">Grid {{ gi + 1 }}</span>
                    <span>{{ g.axis ?? 'single' }} · {{ g.items.length }}타일</span>
                </template>
                <template #overlay>
                    <span v-for="x in stopsOf(g.cols)" :key="x" class="snap" :style="{ left: `${x}px` }" />
                </template>
            </tile-layout>
        </div>
    </div>
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
}
.cell {
    flex: 0 0 100%;
    min-width: 0;
    display: flex;
    flex-direction: column;
    // 오른쪽 칼럼이 통째로 빈 Grid = 칼럼 하나 폭
    &.narrow {
        flex-basis: calc((100% - var(--gap)) / 2);
    }
}
.box {
    flex: 1;
    min-height: 0;
    display: flex;
    > .TileLayout { flex: 1; }
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

// 전체보기: 긴 축 한 줄 + 짧은 축 가운데, 자유 스크롤
.TileStrip.overview {
    gap: 14px;
    scroll-snap-type: none;
    align-items: center;
    background: rgb(var(--v-theme-surface-light));
    &.col {
        flex-direction: column;
        overflow-x: hidden;
        overflow-y: auto;
    }
    .cell {
        flex: none;
        gap: 4px;
    }
    .box {
        flex: none;
        position: relative;
        overflow: hidden;
        border-radius: 6px;
        background: rgb(var(--v-theme-background));
        > .TileLayout {
            position: absolute;
            left: 0;
            top: 0;
            flex: none;
            transform-origin: 0 0;
        }
    }
    .Tile:not(.Tile--empty) {
        cursor: pointer;
        border-radius: 6px;
        // 축소 배율을 거꾸로 곱해 화면에서 3px로 보이게
        &:hover, &:focus-visible {
            outline: calc(3px / var(--s)) solid rgb(var(--v-theme-primary));
            outline-offset: calc(-3px / var(--s));
        }
    }
}
.cap {
    font-size: 11px;
    font-family: monospace;
    color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity));
    b { color: rgb(var(--v-theme-on-surface)); }
}
</style>
