<script setup lang="ts">
import { useResizeCallback } from '@/feature/common/store/resizeGroup.store';
import { useGapWheel } from '@/feature/widget/hook/gapWheel.hook';
import { computed, ref, watch } from 'vue';
import { useMemberBoard } from '../store/memberBoard.store';
import { GAP } from '../util/memberGrid.util';
import MemberFooter from './MemberFooter.vue';
import MemberList from './MemberList.vue';
import MemberPagerPopup from './MemberPagerPopup.vue';
import MemberTileView from './MemberTileView.vue';

// 내 장비·관리자 페이지 틀 — 목록 열 + 상세 Grid strip + 보드와 같은 아래 푸터. REF-member-ui.md 화-1(방향)·화-6(Grid 크기)·화-13(전체 헤더)·화-14(폰)
const { area, layout, grids, scrollEl, scrollX, viewW, scrubbing, pager } = useMemberBoard();
// 타일 밖(타일 사이·빈 칸·푸터) 세로 휠 = 칼럼 한 칸 — 보드 푸터-7(빈 곳 휠)과 같음. 목록·상세 타일 위는 그 안 스크롤
const onWheel = useGapWheel(pager, t => !!t.closest('.mb-tile'));

// 푸터 위 전체 영역을 재서 열·행을 정한다(화-6)
const bodyEl = ref<HTMLElement>();
useResizeCallback(bodyEl, () => area.value = { w: bodyEl.value?.clientWidth ?? 0, h: bodyEl.value?.clientHeight ?? 0 });

// PC = strip만 스크롤, 폰 = 목록 + Grid 한 줄 스크롤
const outerEl = ref<HTMLElement>();
const stripEl = ref<HTMLElement>();
const phone = computed(() => layout.value.phone);
const measure = () => {
    const el = phone.value ? outerEl.value : stripEl.value;
    scrollEl.value = el;
    scrollX.value = el?.scrollLeft ?? 0;
    viewW.value = el ? el.clientWidth - parseFloat(getComputedStyle(el).paddingLeft) : 0;   // 칼럼 폭 계산 = 안쪽 여백(.strip 4px) 뺀 폭
}
useResizeCallback(outerEl, measure);
useResizeCallback(stripEl, measure);
// 첫 측정 전엔 폭 0 = 폰으로 계산되므로, PC·폰이 바뀌면 스크롤 요소를 다시 고른다
watch(phone, measure, { flush: 'post' });
const onScroll = (e: Event) => {
    if (e.currentTarget === scrollEl.value) scrollX.value = (e.currentTarget as HTMLElement).scrollLeft;
}

const gridStyle = computed(() => ({
    gridTemplateColumns: `repeat(${layout.value.cols}, minmax(0, 1fr))`,
    gridTemplateRows: `repeat(${layout.value.rows}, minmax(0, 1fr))`,
}));
const listStyle = computed(() => phone.value ? undefined : { width: `${layout.value.cellW}px` });
// 칼럼마다 스냅 자리(타일이 없는 칼럼에도 멈추게) — 보드처럼 칼럼 단위 정지
const snaps = computed(() => Array.from({ length: layout.value.cols }, (_, c) => c));
</script>

<template>
<div class="MemberPage" :class="{ phone, scrubbing }" :style="{ '--gap': `${GAP}px` }" @wheel="onWheel">
    <div ref="bodyEl" class="body">
        <div ref="outerEl" class="outer" @scroll.passive="onScroll">
            <div class="list" :style="listStyle">
                <member-list />
            </div>
            <div ref="stripEl" class="strip" @scroll.passive="onScroll">
                <div v-if="!grids.length" class="grid">
                    <div class="empty">목록에서 누르면 여기에 타일이 쌓입니다</div>
                </div>
                <div v-for="(g, gi) in grids" :key="gi" class="grid" :style="gridStyle">
                    <span v-for="c in snaps" :key="`s${c}`" class="snap" :style="{ gridColumn: c + 1 }" aria-hidden="true" />
                    <member-tile-view v-for="it in g.items" :key="it.tile.key" :placed="it"
                        :style="{ gridColumn: `${it.c + 1} / span ${it.cw}`, gridRow: `${it.r + 1} / span ${it.ch}` }"
                    />
                </div>
            </div>
        </div>
    </div>
    <member-footer />
    <member-pager-popup />
</div>
</template>

<style scoped lang="scss">
.MemberPage {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    background: rgb(var(--v-theme-background));
}
.body {
    flex: 1;
    min-height: 0;
    display: flex;
}
.outer {
    flex: 1;
    min-width: 0;
    display: flex;
    gap: var(--gap);
    padding: var(--gap) 0 0 var(--gap);
    overflow: hidden;
}
.list {
    flex: none;
    min-height: 0;
}
.strip {
    flex: 1;
    min-width: 0;
    display: flex;
    overflow-x: auto;
    overflow-y: hidden;
    scroll-snap-type: x mandatory;
    scrollbar-width: none;
    &::-webkit-scrollbar { display: none; }
    // active 바깥 선(타일 밖 3px)이 잘리지 않게 — 위·왼쪽은 간격 안으로 4px 넓히고(타일 자리 그대로), 아래는 4px 비움
    margin: -4px 0 0 -4px;
    padding: 4px 0 4px 4px;
    scroll-padding-left: 4px;
}
.grid {
    flex: none;
    width: 100%;
    height: 100%;
    box-sizing: border-box;
    display: grid;
    gap: var(--gap);
    padding-right: var(--gap);
}
.snap {
    grid-row: 1;
    align-self: start;
    height: 0;
    scroll-snap-align: start;
    pointer-events: none;
}
.empty {
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 1px dashed rgba(var(--v-theme-on-surface), 0.25);
    border-radius: 6px;
    @include vars.font('body-medium');
    color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity));
}
// 알약 끌기 중엔 스냅 끔 — 켜 두면 scrollLeft를 줄 때마다 칼럼으로 튄다(보드 TileStrip과 같음)
.scrubbing .strip, .scrubbing.phone .outer { scroll-snap-type: none; }
// 길게 누르기(컨트롤 패널) 자리에서 글자 선택·기본 메뉴가 뜨지 않게 — 타일 안 내용은 그대로
@media (pointer: coarse) {
    .grid, .outer { user-select: none; -webkit-touch-callout: none; }
    :deep(.mb-tile) { user-select: text; -webkit-touch-callout: default; }
    :deep(.mb-head) { user-select: none; -webkit-touch-callout: none; }
}
// 화-14(폰): 목록이 가로 줄 첫 칸, 넘기면 상세 Grid가 이어진다. 한 칸 = 화면 폭(100cqw)
.phone {
    .outer {
        gap: 0;
        padding: 0;
        overflow-x: auto;
        container-type: inline-size;
        scroll-snap-type: x mandatory;
        scrollbar-width: none;
        &::-webkit-scrollbar { display: none; }
    }
    .list {
        width: 100cqw;
        box-sizing: border-box;
        padding: var(--gap);
        scroll-snap-align: start;
    }
    .strip {
        flex: none;
        overflow: visible;
        scroll-snap-type: none;
        margin: 0;
        padding: 0;   // 폰 = Grid 여백(gap)이 선 자리
    }
    .grid {
        width: 100cqw;
        padding: var(--gap);
        scroll-snap-align: start;
    }
    .snap { display: none; }
}
</style>
