<script setup lang="ts">
import { provideTileGrids } from '../store/tileGrids.store';
import OverviewButton from './OverviewButton.vue';
import TileNav from './TileNav.vue';
import TileStrip from './TileStrip.vue';

// strip과 헤더 내비가 같은 Grid 상태를 쓰도록 공통 부모에서 provide. 내비는 Teleport로 AppHead 자리에 — F(헤더 전달 방식)
provideTileGrids();
</script>

<template>
<div class="TileWorkspace">
    <tile-strip />
    <overview-button />
</div>
<teleport defer to="#app-head-nav">
    <tile-nav />
</teleport>
</template>

<style scoped lang="scss">
// 고정 버튼 기준 — strip이 스크롤돼도 제자리
.TileWorkspace {
    position: relative;
}
</style>
