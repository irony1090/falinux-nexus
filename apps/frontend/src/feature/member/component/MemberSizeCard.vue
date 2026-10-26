<script setup lang="ts">
import { computed } from 'vue';
import { VBtn, VCard } from 'vuetify/components';
import { useMemberBoard } from '../store/memberBoard.store';

// 크기 카드(폭 1~열 수 · 높이 1~행 수) — 머리 줄 크기 팝업·컨트롤 패널이 같이 씀(보드 TileSizeCard와 같은 모양), 화-10(타일 크기)
const props = defineProps<{ tileKey: string }>();
const emit = defineEmits<{ changed: [] }>();
const { tiles, layout, setSize } = useMemberBoard();

const tile = computed(() => tiles.value.find(t => t.key === props.tileKey));
const max = computed(() => ({ w: layout.value.cols, h: layout.value.rows }));
const shown = (axis: 'w' | 'h') => Math.min(tile.value?.[axis] ?? 1, max.value[axis]);
const pick = (axis: 'w' | 'h', v: number) => {
    setSize(props.tileKey, axis, v);
    emit('changed');
}
</script>

<template>
<v-card v-if="tile" class="MemberSizeCard">
    <p class="size-now">크기 {{ shown('w') }}×{{ shown('h') }}칸</p>
    <div v-for="axis in (['w', 'h'] as const)" :key="axis" class="size-row">
        <span>{{ axis === 'w' ? '폭' : '높이' }}</span>
        <v-btn v-for="v in max[axis]" :key="v"
            size="x-small"
            :variant="shown(axis) === v ? 'flat' : 'outlined'"
            :color="shown(axis) === v ? 'primary' : undefined"
            @click="pick(axis, v)"
        >{{ v }}</v-btn>
    </div>
    <p v-if="max.w === 1 && max.h === 1" class="size-hint">이 화면은 한 칸만 들어갑니다</p>
</v-card>
</template>

<style scoped lang="scss">
.MemberSizeCard {
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
