<script setup lang="ts">
import { useResizeCallback } from '@/feature/common/store/resizeGroup.store';
import { computed, ref, watch } from 'vue';
import { usePagerDrag, type PagerPull } from '../hook/pagerDrag.hook';
import { useTileGrids } from '../store/tileGrids.store';
import FootPill from './FootPill.vue';
import TilePeek from './TilePeek.vue';

// 도트 한 칸 폭 / 인디케이터 좌우 여백 / 숫자 자리(간격 포함)
const SP = 22;
const PAD = 24;
const COUNT_W = 72;

// 푸터 인디케이터: Grid마다 도트 + 지금 위치(반 칸 = 두 도트 사이 작은 도트). 드래그 = 칼럼 이동, 당기면 타일 이동 — REF-board-pager.md
// pull = 타일 이동으로 당기는 쪽(푸터 = 위, 커서 팝업 = 위·아래)
const props = withDefaults(defineProps<{ pull?: PagerPull }>(), { pull: 'up' });
const { grids, pos, range, goGrid, goColumn } = useTileGrids();
const { mode, pullDir, down, move, up } = usePagerDrag(props.pull);
defineExpose({ mode });

const wrapEl = ref<HTMLElement>();
const barEl = ref<HTMLElement>();
const width = ref(0);
useResizeCallback(wrapEl, () => width.value = wrapEl.value?.clientWidth ?? 0);
// 알약 양 끝 슬롯(왼쪽 = 크기 버튼, 오른쪽 = 전체보기 버튼) 폭 — 도트 자리에서 뺀다
const startEl = ref<HTMLElement>();
const startW = ref(0);
useResizeCallback(startEl, () => startW.value = startEl.value?.offsetWidth ?? 0);
const endEl = ref<HTMLElement>();
const endW = ref(0);
useResizeCallback(endEl, () => endW.value = endEl.value?.offsetWidth ?? 0);

// 보이는 도트 수 — 폭이 모자라면 숫자부터 빼고(도트 5개 기준), 그래도 넘치면 현재 위치 주변만
const fit = computed(() => {
    const n = grids.value.length;
    const w = width.value - PAD - startW.value - endW.value;
    const count = w - COUNT_W >= Math.min(n, 5) * SP;
    const k = Math.max(3, Math.floor((w - (count ? COUNT_W : 0)) / SP));
    return { count, k: Math.min(n, k) };
})
// 보이는 도트의 시작(실수) — 스크롤을 따라 미끄러진다
const ws = computed(() => {
    const n = grids.value.length, { k } = fit.value;
    return n > k ? Math.max(0, Math.min(n - k, pos.value - (k - 1) / 2)) : 0;
})
// 더 있는 쪽 끝 도트는 작게
const dotScale = (i: number) => {
    const n = grids.value.length, { k } = fit.value, rel = i - ws.value;
    let s = 1;
    if (ws.value > 0.001) s = Math.min(s, 0.35 + 0.65 * Math.max(0, rel) / 1.5);
    if (ws.value < n - k - 0.001) s = Math.min(s, 0.35 + 0.65 * Math.max(0, k - 1 - rel) / 1.5);
    return s;
}

// 정수 위치 = 긴 막대, 반 칸으로 갈수록 작은 도트로 줄어든다
const pillStyle = computed(() => {
    const half = Math.abs(pos.value - Math.round(pos.value));
    const w = Math.max(5, 18 - 26 * half), h = 6 - 2 * half;
    return { width: `${w}px`, height: `${h}px`, marginTop: `${-h / 2}px`, left: `${pos.value * SP + SP / 2 - w / 2}px` };
})

const countText = computed(() => {
    const { lo, hi } = range.value;
    return `${lo === hi ? lo + 1 : `${lo + 1}-${hi + 1}`} / ${grids.value.length}`;
})

// 인디케이터는 앱 바(overflow 잘림) 안이라 안내·타일 패널은 body에 고정 위치로 — 드래그를 시작할 때 자리를 잰다
// 안내는 막대 위쪽(bottom 기준) — 푸터-4(위로 열기). 타일 패널은 당긴 쪽
const anchor = ref({ bottom: 0, top: 0, center: 0 });
watch(mode, m => {
    const r = barEl.value?.getBoundingClientRect();
    if (m !== 'idle' && r) anchor.value = { bottom: document.documentElement.clientHeight - r.top + 8, top: r.bottom + 8, center: r.left + r.width / 2 };
})
const floatStyle = computed(() => ({ bottom: `${anchor.value.bottom}px`, left: `${anchor.value.center}px` }));
const peekStyle = computed(() => pullDir.value === 'down' ? { top: `${anchor.value.top}px`, left: `${anchor.value.center}px` } : floatStyle.value);
const hintText = computed(() => props.pull === 'both' ? '위·아래로 당기면 타일 선택' : '위로 당기면 타일 선택');

// 칼럼 화살표 버튼 대신 — 인디케이터 안에 포커스가 있을 때 ←/→
const onKey = (e: KeyboardEvent) => {
    if (e.key !== 'ArrowLeft' && e.key !== 'ArrowRight') return;
    e.preventDefault();
    goColumn(e.key === 'ArrowLeft' ? -1 : 1);
}
</script>

<template>
<div ref="wrapEl" class="GridPager">
    <foot-pill class="shell">
        <span v-if="$slots.start" ref="startEl" class="end"><slot name="start" /></span>
        <nav ref="barEl" class="bar" :class="{ held: mode !== 'idle' }" aria-label="보드 Grid 위치"
            @pointerdown="down"
            @pointermove="move"
            @pointerup="up"
            @pointercancel="up"
            @keydown="onKey"
        >
            <div class="track" :class="{ dim: mode === 'tile' }" :style="{ width: `${fit.k * SP}px` }">
                <div class="row" :style="{ transform: `translateX(${-ws * SP}px)` }">
                    <div class="pill" :style="pillStyle" aria-hidden="true" />
                    <button v-for="(_, i) in grids" :key="i" type="button" class="dot-btn"
                        :style="{ width: `${SP}px` }"
                        :aria-label="`Grid ${i + 1}로 이동`"
                        :aria-current="i >= range.lo && i <= range.hi"
                        @click="goGrid(i)"
                    >
                        <span class="dot" :style="{ transform: `scale(${dotScale(i)})` }" />
                    </button>
                </div>
            </div>
            <span v-if="fit.count" class="count text-body-small text-medium-emphasis">{{ countText }}</span>
        </nav>
        <span ref="endEl" class="end"><slot name="end" /></span>
    </foot-pill>
    <teleport to="body">
        <div class="GridPager__hint text-body-small" :class="{ show: mode === 'col' }" :style="floatStyle" aria-hidden="true">{{ hintText }}</div>
        <tile-peek :open="mode === 'tile'" :below="pullDir === 'down'" :style="peekStyle" />
    </teleport>
</div>
</template>

<style scoped lang="scss">
.GridPager {
    flex: 1;
    min-width: 0;
    display: flex;
    justify-content: center;
}
// 양 끝 버튼이 알약 끝까지 채우게
.shell {
    padding: 0;
    overflow: hidden;
}
.end {
    align-self: stretch;
    display: flex;
}
// 손가락 드래그를 브라우저 스크롤·선택이 가로채지 않게
.bar {
    flex: none;
    display: flex;
    align-items: center;
    gap: 8px;
    height: 100%;
    padding: 0 12px;
    touch-action: none;
    user-select: none;
    -webkit-user-select: none;
    cursor: grab;
    &.held { cursor: grabbing; }
}
.track {
    position: relative;
    height: 100%;
    overflow: hidden;
    transition: opacity 0.2s;
    &.dim { opacity: 0.25; }
}
.row {
    position: relative;
    display: flex;
    height: 100%;
}
.pill {
    position: absolute;
    z-index: 1;
    top: 50%;
    border-radius: 3px;
    background: rgb(var(--v-theme-primary));
    pointer-events: none;
}
.dot-btn {
    flex: none;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0;
    border: 0;
    background: transparent;
    cursor: pointer;
}
.dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: rgba(var(--v-theme-on-surface), 0.3);
}
.count {
    flex: none;
    min-width: 64px;
    text-align: right;
    font-family: monospace;
    white-space: nowrap;
}
</style>

<style lang="scss">
// teleport로 body에 붙으므로 scoped 밖
.GridPager__hint {
    position: fixed;
    z-index: 2100;
    padding: 4px 10px;
    border-radius: 12px;
    background: rgb(var(--v-theme-on-surface));
    color: rgb(var(--v-theme-surface));
    white-space: nowrap;
    pointer-events: none;
    opacity: 0;
    transform: translate(-50%, 4px);
    transition: opacity 0.2s, transform 0.2s;
    &.show {
        opacity: 1;
        transform: translate(-50%, 0);
    }
}
</style>
