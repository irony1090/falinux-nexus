<script setup lang="ts">
import { useResizeCallback } from '@/feature/common/store/resizeGroup.store';
import { isRunning, useProcessTerm } from '@/feature/process/store/processTerm.store';
import { removeUnit } from '@/common/util/index.util';
import Tile from '@/feature/widget/component/Tile.vue';
import TileLayout from '@/feature/widget/component/TileLayout.vue';
import type { Area, CellSpan } from '@/feature/widget/util/tile.type';
import { columnStops } from '@/feature/widget/util/tileSplit.util';
import { computed, onBeforeUnmount, onMounted, watch, type CSSProperties } from 'vue';
import { useRouter } from 'vue-router';
import type { BoardItem } from '../api/board.api';
import { useTileShortcut } from '../hook/tileShortcut.hook';
import { boardLabel, focusPath, useBoards } from '../store/boards.store';
import { TILE_FOOT as FOOT, TILE_GAP as GAP, useTileGrids, type BoardGrids } from '../store/tileGrids.store';
import { useTileTree } from '../store/tileTree.store';
import OverviewTile from './tile/OverviewTile.vue';
import TileFrame from './tile/TileFrame.vue';

const {
    grids, cap, area, stripEl, scrollX, viewW,
    overview, ovScale, ovDir, otherBoards, scrollbar, leaveOverview, activate,
} = useTileGrids();
useTileShortcut();
const { hash } = useTileTree();
const { byHash } = useBoards();
const { procs } = useProcessTerm();
const router = useRouter();

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

// 전체보기에 다른 보드 줄이 있음 — EV2(보드 쌓는 방향)
const multi = computed(() => overview.value && otherBoards.value.length > 0);

// 가로 한 줄 전체보기에선 세로 휠도 가로로. 보드 줄이 여럿이면 세로 휠 = 보드 사이 이동
const onWheel = (e: WheelEvent) => {
    const el = stripEl.value;
    if (!el || !overview.value || multi.value || ovDir.value !== 'row' || Math.abs(e.deltaY) <= Math.abs(e.deltaX)) return;
    e.preventDefault();
    el.scrollLeft += e.deltaY;
}

const onKey = (e: KeyboardEvent) => {
    if (e.key === 'Escape' && overview.value) leaveOverview();
}
onMounted(() => window.addEventListener('keydown', onKey));
onBeforeUnmount(() => window.removeEventListener('keydown', onKey));

// 전체보기에서 고른 타일 = active, 나가며 그 Grid로
const onTileKey = (e: KeyboardEvent, tileId: string) => {
    if (!overview.value || (e.key !== 'Enter' && e.key !== ' ')) return;
    e.preventDefault();
    activate(tileId);
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

const stripClass = computed(() => overview.value ? ['overview', ovDir.value, { multi: multi.value }] : []);

const currentName = computed(() => boardLabel(byHash(hash.value)) || '기본 보드');

// 보드 줄 요약 — 배치된 터미널 수 + 실행 중 수
const summaryOf = (b: BoardGrids) => {
    const uids = b.grids.flatMap(g => g.items.flatMap(it => it.tile.type === 'terminal' ? [it.tile.uid] : []));
    const running = uids.filter(u => isRunning(procs.value[u]?.status)).length;
    return uids.length ? `터미널 ${uids.length} · 실행 중 ${running}` : '터미널 없음';
}

// 다른 보드 타일 누름 = 그 보드로 가서 그 타일 active(기존 ?focus 처리) — 미리보기-5(이동), EV5(누르면)
const goOther = (board: BoardItem, tileId: string) => router.push(focusPath(board, tileId));
const onOtherKey = (e: KeyboardEvent, board: BoardItem, tileId: string) => {
    if (e.key !== 'Enter' && e.key !== ' ') return;
    e.preventDefault();
    void goOther(board, tileId);
}
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
<!-- .board·.board-cells는 평소 display: contents — 전체보기에서만 상자가 된다(위 재마운트 금지와 같은 이유로 두 모드 모두 둔다) -->
<div class="board" :class="{ current: multi }">
    <div v-if="multi" class="board-cap">
        <b>{{ currentName }}</b>
        <span class="now">지금 보드</span>
    </div>
    <div class="board-cells">
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
                    @click="overview && activate(it.tile.id)"
                    @keydown="onTileKey($event, it.tile.id)"
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
</div>

<!-- 다른 보드 줄 — 미리보기-4(그리기). 같은 cap·축소 비율로 배치 그대로 -->
<div v-for="b in otherBoards" :key="b.board.hash" class="board">
    <div class="board-cap">
        <b>{{ boardLabel(b.board) }}</b>
        <span>{{ b.loading ? '불러오는 중…' : b.error ?? summaryOf(b) }}</span>
    </div>
    <div class="board-cells">
        <div v-for="(g, gi) in b.grids" :key="gi" class="cell" :class="{ narrow: g.cols === 1 }">
            <div class="cap"><b>G{{ gi + 1 }}</b> · {{ g.items.length }}타일</div>
            <div class="box" :style="boxStyle(g.cols)">
                <tile-layout :cols="g.cols" :gap="GAP" :foot-height="FOOT" :style="layoutStyle(g.cols)">
                    <tile v-for="it in g.items" :key="it.tile.id" :cell="it"
                        tabindex="0"
                        role="button"
                        :aria-label="`${boardLabel(b.board)} 보드로 이동`"
                        @click="goOther(b.board, it.tile.id)"
                        @keydown="onOtherKey($event, b.board, it.tile.id)"
                    >
                        <overview-tile :tile="it.tile" :pos="b.posOf.get(it.tile.id)" />
                    </tile>
                    <tile v-for="(e, ei) in g.empty" :key="`empty-${ei}`" :cell="e" empty />
                </tile-layout>
            </div>
        </div>
    </div>
</div>
</div>
</template>

<style scoped lang="scss">
.TileStrip {
    // ⑥ Grid 높이: 남은 높이 전부(헤더를 뺀 높이는 TileWorkspace가 정함, 선택 막대가 있으면 그만큼 줄어듦)
    flex: 1;
    min-height: 0;
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

// 평소엔 상자 없음 — cell이 strip의 flex 항목
.board, .board-cells { display: contents; }

// 전체보기: 긴 축 한 줄 + 짧은 축 가운데, 자유 스크롤. strip = 보드 줄을 세로로 쌓음(보드가 하나면 지금 보드 한 줄)
.TileStrip.overview {
    flex-direction: column;
    gap: 14px;
    overflow: auto;
    scroll-snap-type: none;
    background: rgb(var(--v-theme-surface-light));
    .board {
        display: flex;
        flex: none;
        gap: 4px;
    }
    .board-cells {
        display: flex;
        gap: 14px;
        align-items: center;
    }
    // PC 가로: Grid 가로 한 줄 + 세로 가운데
    &.row {
        justify-content: safe center;
        align-items: flex-start;
    }
    // 폰 세로: 보드 안 Grid도 세로 + 가로 가운데, 보드 사이엔 이름표 — EV2(보드 쌓는 방향) 나
    &.col {
        align-items: safe center;
        overflow-x: hidden;
        .board { flex-direction: column; }
        .board-cells { flex-direction: column; }
    }
    // PC 가로 + 보드 여럿: 보드마다 가로 한 줄(이름 왼쪽), 위에서부터 — EV2(보드 쌓는 방향) 가
    &.row.multi {
        justify-content: flex-start;
        align-items: flex-start;
        .board {
            gap: 12px;
            padding: 8px 10px;
            border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
            border-radius: 8px;
            background: rgba(var(--v-theme-surface), 0.6);
        }
        .board-cap {
            width: 140px;
            flex-direction: column;
            align-items: flex-start;
            justify-content: center;
        }
    }
    // 지금 보드 강조 — EV4(지금 보드 위치)
    &.multi .board.current {
        border: 2px solid rgb(var(--v-theme-primary));
        border-radius: 8px;
        background: rgb(var(--v-theme-surface));
    }
    &.col.multi .board {
        padding: 6px;
        border-radius: 8px;
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
.board-cap {
    flex: none;
    display: flex;
    align-items: baseline;
    gap: 6px;
    font-size: 12px;
    color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity));
    b { font-size: 13px; color: rgb(var(--v-theme-on-surface)); }
    .now { color: rgb(var(--v-theme-primary)); }
}
.cap {
    font-size: 11px;
    font-family: monospace;
    color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity));
    b { color: rgb(var(--v-theme-on-surface)); }
}
</style>
