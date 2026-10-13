<script setup lang="ts">
import { watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { provideNodeOps } from '../store/nodeOps.store';
import { provideTileGrids } from '../store/tileGrids.store';
import { useTileTree } from '../store/tileTree.store';
import NodeOpsHost from './NodeOpsHost.vue';
import NodeSelectionBar from './NodeSelectionBar.vue';
import OverviewButton from './OverviewButton.vue';
import TileNav from './TileNav.vue';
import TileStrip from './TileStrip.vue';

// strip과 헤더 내비가 같은 Grid 상태를 쓰도록 공통 부모에서 provide. 내비는 Teleport로 AppHead 자리에 — F(헤더 전달 방식)
const grids = provideTileGrids();
const { activate, rootTab } = grids;
provideNodeOps(grids);     // 같은 컴포넌트에선 inject가 안 되므로 grids를 넘긴다
const { tiles, setMinimized } = useTileTree();
const route = useRoute();
const router = useRouter();

// ?focus=<tileId> = 그 타일을 active로(목록 "보드로 이동"·위치 알림, N3(누르면 이동)) / ?tab=terminals = 루트 터미널 탭(M4(위치 없음)).
// 타일이 트리에 들어올 때까지 기다렸다가 처리하고 query를 지운다
watch(() => {
    const focus = route.query.focus;
    return [focus, route.query.tab, typeof focus === 'string' ? tiles.value[focus] : undefined] as const;
}, ([focus, tab, tile]) => {
    if (typeof focus === 'string') {
        if (!tile) return;
        if (tile.type === 'terminal' && tile.minimized) setMinimized(tile.id, false);
        void activate(tile.id);
    } else if (tab === 'terminals') {
        rootTab.value = 'terminals';
    } else {
        return;
    }
    void router.replace({ query: {} });
}, { immediate: true })
</script>

<template>
<div class="TileWorkspace">
    <node-selection-bar />
    <tile-strip />
    <overview-button />
    <node-ops-host />
</div>
<teleport defer to="#app-head-nav">
    <tile-nav />
</teleport>
</template>

<style scoped lang="scss">
// 고정 버튼 기준 — strip이 스크롤돼도 제자리. 높이 = 화면 - 헤더, 선택 막대가 뜨면 strip이 그만큼 줄어든다
.TileWorkspace {
    position: relative;
    height: calc(100dvh - var(--v-layout-top, 0px));
    display: flex;
    flex-direction: column;
}
</style>
