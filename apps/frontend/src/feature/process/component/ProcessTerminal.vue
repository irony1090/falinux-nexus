<script setup lang="ts">
import '@xterm/xterm/css/xterm.css';
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useProcessTerm } from '../store/processTerm.store';

// 크기 소유 탭 = 이 영역에 맞춰 PTY 크기를 보냄 / 아니면 소유자의 rows·cols로 고정하고 넘치면 스크롤 (REF-node-ui-terminal.md M)
const props = defineProps({
    uid: {
        type: String,
        required: true,
    },
})

const { procs, owners, handleOf, attach, detach, syncSize } = useProcessTerm();

const el = ref<HTMLDivElement>();
const owner = computed(() => !!owners.value[props.uid]);
const proc = computed(() => procs.value[props.uid]);

const layout = () => {
    const h = handleOf(props.uid);
    if (!h || !el.value) return;
    if (owner.value) {
        const dim = h.fit.proposeDimensions();
        if (!dim || !(dim.cols > 0) || !(dim.rows > 0)) return;     // 크기 0(마운트 직후·숨김)이면 NaN
        if (dim.cols !== h.term.cols || dim.rows !== h.term.rows) h.term.resize(dim.cols, dim.rows);
        syncSize(props.uid, dim.cols, dim.rows);
        return;
    }
    const p = proc.value;
    if (p && p.cols > 0 && p.rows > 0 && (p.cols !== h.term.cols || p.rows !== h.term.rows)) h.term.resize(p.cols, p.rows);
}

// 타일 크기 변경·Grid 이동이 몰아서 오므로 묶어서 한 번
let timer: ReturnType<typeof setTimeout> | undefined;
const schedule = () => {
    clearTimeout(timer);
    timer = setTimeout(layout, 150);
}

let ro: ResizeObserver | undefined;
onMounted(() => {
    if (!el.value) return;
    attach(props.uid, el.value);
    ro = new ResizeObserver(schedule);
    ro.observe(el.value);
    schedule();
})
onBeforeUnmount(() => {
    clearTimeout(timer);
    ro?.disconnect();
    if (el.value) detach(props.uid, el.value);
})

watch(owner, schedule);
watch(() => [proc.value?.cols, proc.value?.rows], schedule);
</script>

<template>
<div ref="el" class="ProcessTerminal" :class="{ 'ProcessTerminal--fixed': !owner }" />
</template>

<style scoped lang="scss">
.ProcessTerminal {
    width: 100%;
    height: 100%;
    overflow: hidden;
}
// 소유자 크기로 고정된 화면은 타일보다 클 수 있다
.ProcessTerminal--fixed { overflow: auto; }
</style>
