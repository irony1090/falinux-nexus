<script setup lang="ts">
import { computed } from 'vue';
import { VBtn } from 'vuetify/components';
import { OV_PERCENT, useTileGrids } from '../store/tileGrids.store';
import GridPager from './GridPager.vue';

// 헤더 내비: 표시줄(GridPager) — REF-board-pager.md
const { grids, overview, ovPercent, zoomOverview } = useTileGrids();

// 전체보기 중: 표시줄 대신 요약 + 축소 비율 조절 — ⑫(전체보기 세부)
const procCount = computed(() => grids.value.reduce((n, g) => n + g.items.filter(it => it.tile.type === 'terminal').length, 0));
</script>

<template>
<div v-if="grids.length && overview" class="TileNav">
    <span class="info text-body-small text-medium-emphasis">프로세스 {{ procCount }} · Grid {{ grids.length }} · 타일을 누르면 이동</span>
    <v-btn icon="mdi-minus" size="small" density="compact" variant="text" title="더 작게"
        :disabled="ovPercent <= OV_PERCENT.min"
        @click="zoomOverview(-1)"
    />
    <span class="col text-body-small">{{ ovPercent }}%</span>
    <v-btn icon="mdi-plus" size="small" density="compact" variant="text" title="더 크게"
        :disabled="ovPercent >= OV_PERCENT.max"
        @click="zoomOverview(1)"
    />
</div>
<div v-else-if="grids.length" class="TileNav">
    <grid-pager />
</div>
</template>

<style scoped lang="scss">
.TileNav {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: vars.$spacing-sm;
}
.info {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-align: right;
}
.col {
    flex: none;
    font-family: monospace;
    white-space: nowrap;
}
</style>
