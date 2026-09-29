<script setup lang="ts">
import type { GridSlot } from '@/feature/widget/util/tile.type';
import { computed, toRefs, type CSSProperties, type PropType } from 'vue';

// TileLayout 안의 한 자리: cell 좌표(0부터)로 grid 위치만 정한다. empty면 빈 칸 표시
const props = defineProps({
    cell: {
        type: Object as PropType<GridSlot>,
        required: true,
    },
    empty: {
        type: Boolean,
        default: false,
    },
})
const { cell } = toRefs(props);

const style = computed<CSSProperties>(() => {
    const { c, r, cw, ch } = cell.value;
    return {
        gridColumn: `${c + 1} / span ${cw}`,
        gridRow: `${r + 1} / span ${ch}`,
    }
})

</script>
<template>
<div class="Tile"
    :class="{ 'Tile--empty': empty }"
    :style="style"
>
    <slot>
        <template v-if="empty">빈 칸</template>
    </slot>
</div>
</template>
<style scoped lang="scss">
.Tile {
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
}
.Tile--empty {
    align-items: center;
    justify-content: center;
    border: 1.5px dashed rgba(var(--v-border-color), 0.3);
    border-radius: 6px;
    font-size: 11px;
    color: rgba(var(--v-theme-on-surface), var(--v-disabled-opacity));
}
</style>
