<script setup lang="ts">
import { computed } from 'vue';
import { VBtn, VIcon } from 'vuetify/components';
import { holdsProc } from '@/feature/widget/util/tileKind.util';
import { OV_PERCENT, useTileGrids } from '../store/tileGrids.store';
import FootPill from './FootPill.vue';
import GridPager from './GridPager.vue';
import PagerPopup from './PagerPopup.vue';

// 푸터 가운데: 인디케이터(GridPager) + 알약 오른쪽 끝 전체보기 버튼(옛 OverviewButton) — REF-board-chrome.md 배치 F(확정안)
const { cap, grids, overview, ovPercent, zoomOverview, enterOverview, leaveOverview } = useTileGrids();

// 가로 1칸 화면 = 아이콘만
const phone = computed(() => cap.value.x === 1);

// 전체보기 중: 같은 자리 알약에 요약 + 축소 비율 조절 + 돌아가기 — ⑫(전체보기 세부)
const procCount = computed(() => grids.value.reduce((n, g) => n + g.items.filter(it => holdsProc(it.tile)).length, 0));
</script>

<template>
<div v-if="grids.length && overview" class="TileNav">
    <foot-pill class="ov">
        <span class="info text-body-small text-medium-emphasis">
            <template v-if="!phone">프로세스 {{ procCount }} · </template>Grid {{ grids.length }}
        </span>
        <v-btn icon="mdi-minus" size="small" density="comfortable" variant="text" title="더 작게"
            :disabled="ovPercent <= OV_PERCENT.min"
            @click="zoomOverview(-1)"
        />
        <span class="col text-body-small">{{ ovPercent }}%</span>
        <v-btn icon="mdi-plus" size="small" density="comfortable" variant="text" title="더 크게"
            :disabled="ovPercent >= OV_PERCENT.max"
            @click="zoomOverview(1)"
        />
        <v-btn class="end back" variant="tonal" rounded="pill" height="100%"
            :icon="phone"
            title="돌아가기"
            aria-label="돌아가기"
            @click="leaveOverview()"
        >
            <v-icon icon="mdi-arrow-u-left-top" size="18" />
            <span v-if="!phone" class="ml-1">돌아가기</span>
        </v-btn>
    </foot-pill>
</div>
<div v-else-if="grids.length" class="TileNav">
    <grid-pager>
        <template #end>
            <v-btn class="end" color="primary" variant="flat" rounded="pill" height="100%"
                :icon="phone"
                title="전체보기"
                aria-label="전체보기"
                @click="enterOverview"
            >
                <v-icon icon="mdi-view-grid" size="18" />
                <span v-if="!phone" class="ml-1">전체</span>
            </v-btn>
        </template>
    </grid-pager>
    <pager-popup />
</div>
</template>

<style scoped lang="scss">
.TileNav {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    justify-content: center;
}
.ov {
    padding: 0 0 0 14px;
    overflow: hidden;
    gap: 2px;
}
.info {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    margin-right: 4px;
}
.col {
    flex: none;
    min-width: 36px;
    text-align: center;
    font-family: monospace;
    font-weight: vars.$weight-lg;
    white-space: nowrap;
}
// 알약 오른쪽 끝을 테두리까지 덮는 버튼(버튼 자체가 알약)
.end {
    flex: none;
    margin: -1px -1px -1px 0;
    font-weight: vars.$weight-lg;
    &.v-btn--icon { width: 46px; }
    &.back { margin-left: 4px; }
}
</style>
