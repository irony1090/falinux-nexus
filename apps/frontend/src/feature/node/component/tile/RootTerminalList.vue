<script setup lang="ts">
import { computed } from 'vue';
import { useRootRuns } from '../../hook/scriptRuns.hook';
import { useTileTree } from '../../store/tileTree.store';
import RunRow from './RunRow.vue';

// 터미널 탭(개편-6(목록 UI)) — 폴더 루트 = 경로 하위 / 장비 루트 = 그 장비 전부(useRootRuns)
const { tiles, rootId } = useTileTree();
const list = useRootRuns();
const isDevice = computed(() => tiles.value[rootId.value]?.type === 'device');
// 최소화 먼저(어디로 갔는지 찾는 곳), 그다음 서버 순서(실행 순)
const runs = computed(() => [...list.value.filter(r => r.minimized), ...list.value.filter(r => !r.minimized)]);
</script>

<template>
<div class="RootTerminalList">
    <run-row v-for="r in runs" :key="r.uid" :run="r" :target-tile-id="rootId" show-name />
    <p v-if="!runs.length" class="empty">{{ isDevice ? '이 장비에 터미널이 없습니다' : '이 경로에 터미널이 없습니다' }}</p>
</div>
</template>

<style scoped lang="scss">
.RootTerminalList {
    flex: 1;
    min-height: 0;
    overflow: auto;
    @include vars.font('body-medium');
}
.empty { margin: vars.$spacing-md; opacity: 0.6; }
</style>
