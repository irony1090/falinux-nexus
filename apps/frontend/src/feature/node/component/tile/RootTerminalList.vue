<script setup lang="ts">
import { computed } from 'vue';
import { useScriptRuns } from '../../hook/scriptRuns.hook';
import { useTileTree } from '../../store/tileTree.store';
import RunRow from './RunRow.vue';

// 터미널 탭 = 루트 타일 경로와 그 하위의 process(개편-6(목록 UI)). 루트 경로가 바뀌면 다시 조회 — 최소화한 타일도 경로를 옮겨 찾는다
const { tiles, rootId } = useTileTree();
const rootNodeId = computed(() => {
    const t = tiles.value[rootId.value];
    return t?.type === 'folder' ? t.nodeId : null;
})
const { list } = useScriptRuns(rootNodeId, true);
// 최소화 먼저(어디로 갔는지 찾는 곳), 그다음 서버 순서(실행 순)
const runs = computed(() => [...list.value.filter(r => r.minimized), ...list.value.filter(r => !r.minimized)]);
</script>

<template>
<div class="RootTerminalList">
    <run-row v-for="r in runs" :key="r.uid" :run="r" :target-tile-id="rootId" show-name />
    <p v-if="!runs.length" class="empty">이 경로에 터미널이 없습니다</p>
</div>
</template>

<style scoped lang="scss">
.RootTerminalList {
    flex: 1;
    min-height: 0;
    overflow: auto;
    font-size: 12px;
}
.empty { margin: vars.$spacing-md; opacity: 0.6; }
</style>
