<script setup lang="ts">
import { provideTileGrids } from '../store/tileGrids.store';
import TileNav from './TileNav.vue';
import TileStrip from './TileStrip.vue';

// strip과 헤더 내비가 같은 Grid 상태를 쓰도록 공통 부모에서 provide. 내비는 Teleport로 AppHead 자리에 — F(헤더 전달 방식)
provideTileGrids();
</script>

<template>
<tile-strip />
<teleport defer to="#app-head-nav">
    <tile-nav />
</teleport>
</template>
