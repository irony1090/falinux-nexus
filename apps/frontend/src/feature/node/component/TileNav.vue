<script setup lang="ts">
import { computed } from 'vue';
import { VBtn, VPagination } from 'vuetify/components';
import { OV_PERCENT, useTileGrids } from '../store/tileGrids.store';

// 헤더 내비: Grid 탭(VPagination) + 칼럼 위치. prev/next 화살표 = 칼럼 단위 이동 — G(화살표 = 칼럼 이동)
const {
    grids, geometry, view, goGrid, goColumn,
    overview, ovPercent, zoomOverview,
} = useTileGrids();

// 전체보기 중: 탭·칼럼 대신 요약 + 축소 비율 조절 — ⑫(전체보기 세부)
const procCount = computed(() => grids.value.reduce((n, g) => n + g.items.filter(it => it.tile.type === 'terminal').length, 0));

const colTotal = computed(() => geometry.value.columns.length);
const colText = computed(() => {
    const { from, to } = view.value;
    return from === to ? `${from + 1}` : `${from + 1}-${to + 1}`;
})

// 현재 탭은 스크롤에서만 파생(:model-value), 클릭은 스크롤만 요청 — v-model 양방향 되먹임 방지
const onPage = (page: number) => goGrid(page - 1);
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
    <v-pagination class="tabs"
        :model-value="view.grid + 1"
        :length="grids.length"
        size="small"
        density="compact"
        active-color="primary"
        @update:model-value="onPage"
    >
        <template #prev>
            <v-btn icon="mdi-chevron-left" size="small" density="compact" variant="text" title="이전 칼럼"
                :disabled="view.col <= 0"
                @click="goColumn(-1)"
            />
        </template>
        <template #next>
            <v-btn icon="mdi-chevron-right" size="small" density="compact" variant="text" title="다음 칼럼"
                :disabled="view.to >= colTotal - 1"
                @click="goColumn(1)"
            />
        </template>
    </v-pagination>
    <span class="col text-body-small text-medium-emphasis">칼럼 {{ colText }}/{{ colTotal }}</span>
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
// VPagination은 자기 폭을 재서 버튼 수를 정한다 — min-width: 0 없으면 내용 폭으로 굳어 안 줄어듦
.tabs {
    flex: 1;
    min-width: 0;
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
