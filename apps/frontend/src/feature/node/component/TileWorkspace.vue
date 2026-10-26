<script setup lang="ts">
import { useVirtualKeyboard } from '@/common/hook/virtualKeyboard.hook';
import TerminalKeyBar from '@/feature/process/component/TerminalKeyBar.vue';
import { useGapWheel } from '@/feature/widget/hook/gapWheel.hook';
import { computed } from 'vue';
import { useTileUrl } from '../hook/tileUrl.hook';
import { provideNodeOps } from '../store/nodeOps.store';
import { provideTileGrids } from '../store/tileGrids.store';
import BoardFooter from './BoardFooter.vue';
import GridToast from './GridToast.vue';
import NodeOpsHost from './NodeOpsHost.vue';
import NodeSelectionBar from './NodeSelectionBar.vue';
import TileStrip from './TileStrip.vue';

// strip과 푸터 인디케이터가 같은 Grid 상태를 쓰도록 공통 부모에서 provide
const grids = provideTileGrids();
provideNodeOps(grids);     // 같은 컴포넌트에선 inject가 안 되므로 grids를 넘긴다
// ?focus=(목록 "보드로 이동"·위치 알림 N3(누르면 이동) + active) · ?view=(보던 칼럼) · ?tab=terminals(M4(위치 없음)) — 새로고침 복원
useTileUrl(grids);
const onWheel = useGapWheel(grids.pager, t => !!t.closest('.Tile:not(.Tile--empty)'));

// 라-2(높이 고정): resizes-content라 키보드가 뜨면 화면이 줄어든다. 그대로 두면 Grid 수용량이 바뀌어 타일이 옮겨 가므로
// 열린 동안은 키보드 전 높이로 고정한다. 포커스된 터미널 타일만 키보드 위에 맞춰 띄운다(TileFrame 맞춤-1(타일 띄우기))
const kb = useVirtualKeyboard();
const bodyStyle = computed(() => ({ height: kb.open ? `${kb.baseH}px` : undefined }));
</script>

<template>
<div class="TileWorkspace">
<div class="TileWorkspace__body" :style="bodyStyle" @wheel="onWheel">
    <node-selection-bar />
    <!-- Grid 정보 토스트의 위치 기준(strip은 가로로 스크롤되므로 그 밖) -->
    <div class="TileWorkspace__stage">
        <tile-strip />
        <grid-toast />
    </div>
    <board-footer />
    <node-ops-host />
</div>
</div>
<!-- 타일 영역 밖(키보드 열림 때 고정 높이로 잘리는 상자 밖)에 둔다 -->
<terminal-key-bar />
</template>

<style scoped lang="scss">
// 높이 = 화면(헤더 없음, 푸터는 body 안). 키보드가 열려 화면이 줄면 body는 고정 높이로 넘치므로 잘라 낸다(clip = 스크립트·포커스로도 스크롤 안 됨)
.TileWorkspace {
    height: 100dvh;
    overflow: clip;
}
// 선택 막대·푸터가 뜨면 strip이 그만큼 줄어든다
.TileWorkspace__body {
    position: relative;
    height: 100%;
    display: flex;
    flex-direction: column;
}
.TileWorkspace__stage {
    position: relative;
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
}
</style>
