<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import type { PagerMode, PagerPull } from '../hook/pagerDrag.hook';
import TilePeek from './TilePeek.vue';

// 알약 끌기 중 뜨는 것 = 안내(칼럼 이동) + 타일 패널(당긴 채) — 보드 GridPager·관리 페이지 MemberDots 공용, REF-board-pager.md
// 알약은 앱 바(overflow 잘림) 안이라 body에 고정 위치로, 끌기를 시작할 때 막대 자리를 잰다
const props = defineProps<{ bar?: HTMLElement; mode: PagerMode; pullDir: 'up' | 'down'; pull: PagerPull }>();

const anchor = ref({ bottom: 0, top: 0, center: 0 });
watch(() => props.mode, m => {
    const r = props.bar?.getBoundingClientRect();
    if (m !== 'idle' && r) anchor.value = { bottom: document.documentElement.clientHeight - r.top + 8, top: r.bottom + 8, center: r.left + r.width / 2 };
});
// 안내는 막대 위쪽(bottom 기준) — 푸터-4(위로 열기). 타일 패널은 당긴 쪽
const floatStyle = computed(() => ({ bottom: `${anchor.value.bottom}px`, left: `${anchor.value.center}px` }));
const peekStyle = computed(() => props.pullDir === 'down' ? { top: `${anchor.value.top}px`, left: `${anchor.value.center}px` } : floatStyle.value);
const hintText = computed(() => props.pull === 'both' ? '위·아래로 당기면 타일 선택' : '위로 당기면 타일 선택');
</script>

<template>
<teleport to="body">
    <div class="PagerFloat__hint text-body-small" :class="{ show: mode === 'col' }" :style="floatStyle" aria-hidden="true">{{ hintText }}</div>
    <tile-peek :open="mode === 'tile'" :below="pullDir === 'down'" :style="peekStyle" />
</teleport>
</template>

<style lang="scss">
// teleport로 body에 붙으므로 scoped 밖
.PagerFloat__hint {
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
