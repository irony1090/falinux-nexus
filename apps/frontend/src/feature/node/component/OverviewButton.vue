<script setup lang="ts">
import { computed } from 'vue';
import { VBtn, VIcon } from 'vuetify/components';
import { TILE_FOOT, TILE_GAP, useTileGrids } from '../store/tileGrids.store';

const HEIGHT = 22;

const { cap, grids, overview, scrollbar, enterOverview, leaveOverview } = useTileGrids();

// Grid 하단 정보 줄 오른쪽 끝에 고정(타일 안 가림), 스크롤바 두께만큼 비킴 — ⑭(전체보기 버튼 최종 위치) 현재안
const style = computed(() => ({
    right: `${TILE_GAP + scrollbar.value.w}px`,
    bottom: `${TILE_GAP + (TILE_FOOT - HEIGHT) / 2 + scrollbar.value.h}px`,
    height: `${HEIGHT}px`,
}))
const label = computed(() => overview.value ? '돌아가기' : '전체보기');
// 가로 1칸 화면 = 아이콘만
const iconOnly = computed(() => cap.value.x === 1);

const onClick = () => overview.value ? leaveOverview() : enterOverview();
</script>

<template>
<v-btn v-if="grids.length" class="OverviewButton"
    :style="style"
    :color="overview ? 'surface-variant' : 'primary'"
    variant="flat"
    rounded="pill"
    size="x-small"
    :icon="iconOnly"
    :title="label"
    :aria-label="label"
    :aria-pressed="overview"
    @click="onClick"
>
    <v-icon :icon="overview ? 'mdi-arrow-u-left-top' : 'mdi-view-grid'" size="14" />
    <span v-if="!iconOnly" class="ml-1">{{ label }}</span>
</v-btn>
</template>

<style scoped lang="scss">
.OverviewButton {
    position: absolute;
    z-index: 10;
    font-weight: vars.$weight-lg;
    &.v-btn--icon {
        width: 30px;
    }
}
</style>
