<script setup lang="ts">
import { computed, toRefs, type CSSProperties } from 'vue';

// Grid 하나: 타일 영역(cols × rows CSS grid) + 하단 정보 줄. 자식 Tile이 cell 좌표로 자리를 잡는다
// overlay 슬롯 = grid 칸을 차지하지 않는 절대배치용(스냅 앵커 등)
const props = defineProps({
    cols: {
        type: Number,
        default: () => 2,
    },
    rows: {
        type: Number,
        default: () => 2,
    },
    gap: {
        type: Number,
        default: () => 8
    },
    footHeight: {
        type: Number,
        default: () => 20
    },
})

const { cols, rows, gap, footHeight } = toRefs(props);

const tileStyle = computed<CSSProperties>(() => {
    return {
        '--cols': cols.value,
        '--rows': rows.value,
        '--gap': `${gap.value}px`,
        '--foot': `${footHeight.value}px`,
    }
})
</script>

<template>
<div class="TileLayout"
    :style="tileStyle"
>
    <div class="TileLayout__body">
        <slot />
    </div>
    <div v-if="$slots.foot" class="TileLayout__foot">
        <slot name="foot" />
    </div>
    <slot name="overlay" />
</div>
</template>
<style scoped lang="scss">
.TileLayout {
    position: relative;
    display: flex;
    flex-direction: column;
    gap: var(--gap);
    min-width: 0;
    min-height: 0;
}
.TileLayout__body {
    flex: 1;
    min-height: 0;
    display: grid;
    grid-template-columns: repeat(var(--cols, 2), minmax(0, 1fr));
    grid-template-rows: repeat(var(--rows, 2), minmax(0, 1fr));
    gap: var(--gap);
}
.TileLayout__foot {
    flex: 0 0 var(--foot);
    display: flex;
    align-items: center;
    gap: vars.$spacing-md;
    min-width: 0;
    overflow: hidden;
    white-space: nowrap;
    font-size: 11px;
    color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity));
}
</style>
