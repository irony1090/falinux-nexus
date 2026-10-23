<script setup lang="ts">
import type { TileSpan } from '@/feature/widget/util/tile.type';
import { toRef } from 'vue';
import { useTileSize } from '../../hook/tileSize.hook';

// 크기 카드(가로·세로 0.5/1.0) — 머리 줄 크기 팝업·Ctrl+Alt 팝업이 같이 씀, 팝업-1(크기 카드 공용화).
// 카드 자리 잡기(따라다니기 등)는 띄우는 쪽이 한다 — changed로 알림
const props = defineProps<{ tileId: string }>();
const emit = defineEmits<{ changed: [] }>();

const { tile, cap, sizeTitle, setSize } = useTileSize(toRef(props, 'tileId'));

const pick = (axis: 'w' | 'h', v: TileSpan) => {
    setSize(axis, v);
    emit('changed');
}
</script>

<template>
<v-card v-if="tile" class="TileSizeCard">
    <p class="size-now">{{ sizeTitle }}</p>
    <div v-for="axis in (['w', 'h'] as const)" :key="axis" class="size-row">
        <span>{{ axis === 'w' ? '가로' : '세로' }}</span>
        <v-btn v-for="v in ([0.5, 1] as const)" :key="v"
            size="x-small"
            :variant="tile.size[axis] === v ? 'flat' : 'outlined'"
            :color="tile.size[axis] === v ? 'primary' : undefined"
            :disabled="v === 0.5 && cap[axis === 'w' ? 'x' : 'y'] === 1"
            @click="pick(axis, v)"
        >{{ v.toFixed(1) }}</v-btn>
    </div>
    <p v-if="cap.x === 1 || cap.y === 1" class="size-hint">1칸인 축의 0.5는 이 화면에선 1.0으로 보임</p>
</v-card>
</template>

<style scoped lang="scss">
.TileSizeCard {
    display: flex;
    flex-direction: column;
    gap: vars.$spacing-sm;
    padding: vars.$spacing-md vars.$spacing-lg;
    @include vars.font('body-medium');
}
.size-row {
    display: flex;
    align-items: center;
    gap: vars.$spacing-sm;
    > span { width: 30px; opacity: 0.7; }
}
.size-now {
    margin: 0;
    font-family: monospace;
    @include vars.font('label-small');
    opacity: 0.7;
}
.size-hint {
    max-width: 180px;
    margin: 0;
    @include vars.font('label-small');
    opacity: 0.7;
}
</style>
