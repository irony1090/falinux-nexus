<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useTileGrids } from '../store/tileGrids.store';
import GridPager from './GridPager.vue';

// 팝업 상자 폭(도트 자리) / 화면 가장자리 여백
const MAX_W = 560;
const EDGE = 16;

// Ctrl+Alt를 누르는 동안 커서 위치에 인디케이터 하나 더 — 이동-2(커서 위치 페이지 표시).
// 떼면 닫힘(드래그 중이면 끝날 때까지) / 커서를 모르면 화면 가운데 / 폰·전체보기·단축키 끈 타일은 안 띄움
const { overview, appShortcutsOff } = useTileGrids();

const pagerRef = ref<InstanceType<typeof GridPager>>();
const boxEl = ref<HTMLElement>();
const held = ref(false);
const dragging = computed(() => (pagerRef.value?.mode ?? 'idle') !== 'idle');
const shown = computed(() => (held.value || dragging.value) && !overview.value);

// 마지막 마우스 위치 — 창 밖으로 나가면 모름
let last: { x: number; y: number } | null = null;
const onPointerMove = (e: PointerEvent) => {
    if (e.pointerType === 'mouse') last = { x: e.clientX, y: e.clientY };
}
const onMouseOut = (e: MouseEvent) => {
    if (!e.relatedTarget) last = null;
}

// 알약 가운데 = 커서, 화면 밖으로 나가면 안쪽으로. 상자 폭이 고정이라 알약 크기는 위치와 무관해 바로 잰다
const at = ref({ x: 0, y: 0 });
const boxW = ref(MAX_W);
const clamp = (v: number, lo: number, hi: number) => Math.max(lo, Math.min(Math.max(lo, hi), v));
const place = () => {
    const { clientWidth: vw, clientHeight: vh } = document.documentElement;
    boxW.value = Math.min(MAX_W, vw - EDGE * 2);
    const p = last ?? { x: vw / 2, y: vh / 2 };
    const pill = boxEl.value?.querySelector<HTMLElement>('.FootPill');
    const w = pill?.offsetWidth ?? 0, h = pill?.offsetHeight ?? 0;
    at.value = { x: clamp(p.x, EDGE + w / 2, vw - EDGE - w / 2), y: clamp(p.y, EDGE + h / 2, vh - EDGE - h / 2) };
}

const onKeyDown = (e: KeyboardEvent) => {
    if (appShortcutsOff.value || held.value || !e.ctrlKey || !e.altKey || e.shiftKey || e.metaKey) return;
    if ((e.key !== 'Control' && e.key !== 'Alt') || !matchMedia('(pointer: fine)').matches) return;
    place();
    held.value = true;
}
const onKeyUp = (e: KeyboardEvent) => {
    if (e.key === 'Control' || e.key === 'Alt') held.value = false;
}
// 창 전환(Ctrl+Alt+화살표 등)으로 keyup을 못 받는 경우
const onBlur = () => held.value = false;

onMounted(() => {
    window.addEventListener('keydown', onKeyDown, true);
    window.addEventListener('keyup', onKeyUp, true);
    window.addEventListener('blur', onBlur);
    window.addEventListener('pointermove', onPointerMove, { capture: true, passive: true });
    window.addEventListener('mouseout', onMouseOut);
})
onBeforeUnmount(() => {
    window.removeEventListener('keydown', onKeyDown, true);
    window.removeEventListener('keyup', onKeyUp, true);
    window.removeEventListener('blur', onBlur);
    window.removeEventListener('pointermove', onPointerMove, { capture: true });
    window.removeEventListener('mouseout', onMouseOut);
})
</script>

<template>
<teleport to="body">
    <!-- 늘 마운트(폭 측정·드래그 상태 유지), 보이기만 바꾼다. mousedown 막기 = xterm 포커스를 뺏지 않게 -->
    <div ref="boxEl" class="PagerPopup" :class="{ shown }"
        :style="{ left: `${at.x}px`, top: `${at.y}px`, width: `${boxW}px` }"
        aria-hidden="true"
        @mousedown.prevent
    >
        <grid-pager ref="pagerRef" pull="both" />
    </div>
</teleport>
</template>

<style scoped lang="scss">
.PagerPopup {
    position: fixed;
    z-index: 2050;
    display: flex;
    height: 38px;
    --pill-h: 38px;
    transform: translate(-50%, -50%);
    visibility: hidden;
    pointer-events: none;
    &.shown { visibility: visible; }
    // 상자는 클릭을 통과시키고 알약만 받는다
    &.shown :deep(.FootPill) { pointer-events: auto; }
}
</style>
