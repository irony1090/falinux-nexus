<script setup lang="ts">
import type { RunResponse } from '../../api/node.api';
import RunRow from './RunRow.vue';

// 노드의 상태 리스트 — 스크립트 행 펼침(실시간-6a). 현재 타일 = 그 폴더 타일(R7(가져올 위치))
defineProps({
    runs: {
        type: Array as () => RunResponse[],
        required: true,
    },
    tileId: {
        type: String,
        required: true,
    },
})
</script>

<template>
<div class="ScriptRunList">
    <run-row v-for="r in runs" :key="r.uid" :run="r" :target-tile-id="tileId" />
</div>
</template>

<style scoped lang="scss">
.ScriptRunList {
    background: rgba(var(--v-theme-on-surface), 0.04);
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}
</style>
