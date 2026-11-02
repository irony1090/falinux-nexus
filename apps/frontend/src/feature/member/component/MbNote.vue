<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';

// 아래 줄 왼쪽 안내 글 — 혼자 한 줄이면 줄바꿈, 버튼 옆이면 ...로 줄이고 마우스 올리기·누르기(터치)로 전체 글 툴팁
const props = defineProps<{ text: string }>();

const note = ref<HTMLElement>();
const txt = ref<HTMLElement>();
const alone = ref(false);
const cut = ref(false);
const open = ref(false);

// 혼자 = 같은 아래 줄의 버튼 묶음(.mb-end)이 없거나 다음 줄로 내려감
const measure = () => {
    const n = note.value, e = txt.value;
    if (!n || !e) return;
    const end = n.parentElement?.querySelector<HTMLElement>(':scope > .mb-end');
    alone.value = !end || end.getBoundingClientRect().top >= n.getBoundingClientRect().bottom - 1;
    cut.value = !alone.value && e.scrollWidth > e.clientWidth;
    if (!cut.value) open.value = false;
};
let ro: ResizeObserver | undefined;
onMounted(() => {
    ro = new ResizeObserver(measure);
    if (note.value) ro.observe(note.value);
    if (txt.value) ro.observe(txt.value);
});
onBeforeUnmount(() => ro?.disconnect());
watch(() => props.text, () => nextTick(measure));

// 터치는 pointerenter/leave도 오므로 마우스만 호버로 받고, 터치는 누를 때마다 열고 닫는다
const enter = (e: PointerEvent) => { if (e.pointerType === 'mouse' && cut.value) open.value = true; };
const leave = (e: PointerEvent) => { if (e.pointerType === 'mouse') open.value = false; };
const tap = (e: PointerEvent) => { if (e.pointerType !== 'mouse' && cut.value) open.value = !open.value; };
</script>

<template>
<span ref="note" class="mb-note" :class="{ 'mb-note--wrap': alone }" @pointerenter="enter" @pointerleave="leave" @pointerup="tap">
    <span ref="txt" class="mb-note-txt">{{ text }}</span>
    <v-tooltip v-model="open" activator="parent" location="top" max-width="320"
        :open-on-hover="false" :open-on-click="false" :open-on-focus="false" :persistent="false"
    >{{ text }}</v-tooltip>
</span>
</template>
