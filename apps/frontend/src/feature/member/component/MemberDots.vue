<script setup lang="ts">
import FootPill from '@/feature/node/component/FootPill.vue';
import PagerFloat from '@/feature/widget/component/PagerFloat.vue';
import { usePagerDrag, type PagerPull } from '@/feature/widget/hook/pagerDrag.hook';
import { computed, ref } from 'vue';
import { useMemberBoard } from '../store/memberBoard.store';

// 도트 한 칸 폭 — 보드 GridPager와 같게
const SP = 22;

// Grid 도트 + 지금 위치(칼럼 사이면 작은 도트). 폰은 맨 앞 도트 = 목록(진하게). 양 끝 슬롯 = 커서 팝업 버튼 — 화-13(전체 헤더)
// 끌기 = 칼럼 이동, 당긴 채 끌기 = 타일 하나씩 active(보드 GridPager와 같음, usePagerDrag). pull = 당기는 쪽(푸터 = 위, 컨트롤 패널 = 위·아래)
const props = withDefaults(defineProps<{ pull?: PagerPull }>(), { pull: 'up' });
const { pos, dotCount, lead, goDot } = useMemberBoard();
const { mode, pullDir, down, move, up } = usePagerDrag(props.pull);
defineExpose({ mode });
const barEl = ref<HTMLElement>();

const pillStyle = computed(() => {
    const half = Math.abs(pos.value - Math.round(pos.value));
    const w = Math.max(5, 18 - 26 * half), h = 6 - 2 * half;
    return { width: `${w}px`, height: `${h}px`, marginTop: `${-h / 2}px`, left: `${pos.value * SP + SP / 2 - w / 2}px` };
});
const label = (i: number) => i < lead.value ? '목록' : `Grid ${i - lead.value + 1}`;
</script>

<template>
<foot-pill class="MemberDots">
    <span v-if="$slots.start" class="end"><slot name="start" /></span>
    <nav ref="barEl" class="bar" :class="{ held: mode !== 'idle' }" aria-label="Grid 위치"
        @pointerdown="down"
        @pointermove="move"
        @pointerup="up"
        @pointercancel="up"
    >
        <div class="row" :class="{ dim: mode === 'tile' }" :style="{ width: `${dotCount * SP}px` }">
            <div class="pill" :style="pillStyle" aria-hidden="true" />
            <button v-for="i in dotCount" :key="i" type="button" class="dot-btn" :style="{ width: `${SP}px` }"
                :aria-label="`${label(i - 1)}로 이동`" @click="goDot(i - 1)"
            >
                <span class="dot" :class="{ list: i - 1 < lead }" />
            </button>
        </div>
    </nav>
    <span v-if="$slots.end" class="end"><slot name="end" /></span>
    <pager-float :bar="barEl" :mode="mode" :pull-dir="pullDir" :pull="pull" />
</foot-pill>
</template>

<style scoped lang="scss">
.MemberDots {
    justify-self: center;
    max-width: 100%;
    padding: 0;
    overflow: hidden;
}
.end {
    align-self: stretch;
    display: flex;
}
.bar {
    min-width: 0;
    height: 100%;
    padding: 0 12px;
    overflow: hidden;
    // 손가락 끌기를 브라우저 스크롤·선택이 가로채지 않게
    touch-action: none;
    user-select: none;
    -webkit-user-select: none;
    cursor: grab;
    &.held { cursor: grabbing; }
}
.row {
    position: relative;
    display: flex;
    height: 100%;
    transition: opacity 0.2s;
    &.dim { opacity: 0.25; }
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
    &.list { background: rgba(var(--v-theme-on-surface), 0.6); }
}
</style>
