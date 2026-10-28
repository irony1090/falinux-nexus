<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue';

// 가로로 넘치는 버튼 줄 — 막대 숨김, 넘친 쪽 가장자리 흐림, 세로 휠 = 가로 스크롤(파트(경로 줄 정리))
const el = ref<HTMLElement>();
const inner = ref<HTMLElement>();
const fadeL = ref(false);
const fadeR = ref(false);

const update = () => {
    const e = el.value;
    if (!e) return;
    const max = e.scrollWidth - e.clientWidth;
    fadeL.value = max > 1 && e.scrollLeft > 1;
    fadeR.value = max > 1 && e.scrollLeft < max - 1;
}
const onWheel = (ev: WheelEvent) => {
    const e = el.value;
    if (!e || e.scrollWidth <= e.clientWidth || Math.abs(ev.deltaX) >= Math.abs(ev.deltaY)) return;
    e.scrollLeft += ev.deltaY;
    ev.preventDefault();
}

// 바깥 폭(타일 크기)과 안쪽 폭(버튼이 늘고 줆) 둘 다 본다
const ro = new ResizeObserver(update);
onMounted(() => {
    if (el.value) ro.observe(el.value);
    if (inner.value) ro.observe(inner.value);
})
onBeforeUnmount(() => ro.disconnect());
</script>

<template>
<span ref="el" class="HScroll" :class="{ 'fade-l': fadeL, 'fade-r': fadeR }" @scroll.passive="update" @wheel="onWheel">
    <span ref="inner" class="in"><slot /></span>
</span>
</template>

<style scoped lang="scss">
.HScroll {
    flex: 0 1 auto;
    min-width: 0;
    display: flex;
    align-items: stretch;
    overflow-x: auto;
    overflow-y: hidden;
    scrollbar-width: none;
    &::-webkit-scrollbar { display: none; }
    --fade: 18px;
    &.fade-l { mask-image: linear-gradient(to right, transparent, black var(--fade)); }
    &.fade-r { mask-image: linear-gradient(to left, transparent, black var(--fade)); }
    &.fade-l.fade-r { mask-image: linear-gradient(to right, transparent, black var(--fade), black calc(100% - var(--fade)), transparent); }
}
.in {
    flex: none;
    display: flex;
    align-items: center;
    gap: 4px;
}
</style>
