<script setup lang="ts">
import { computed } from 'vue';
import { VIcon } from 'vuetify/components';
import { isNavTile, useTileTree } from '../../store/tileTree.store';

// 타일 안 뒤로·앞으로 — 뒤로-1(타일 안 이동 기록). 단축키 Ctrl+Shift+←/→ = tileNav.hook.ts
const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const { tiles, goHistory } = useTileTree();
const tile = computed(() => {
    const t = tiles.value[props.tileId];
    return isNavTile(t) ? t : undefined;
})
</script>

<template>
<span v-if="tile" class="TileNavButtons">
    <button :disabled="!tile.back?.length" title="뒤로 (Ctrl+Shift+←)" @click="goHistory(tileId, -1)">
        <v-icon size="14" icon="mdi-arrow-left" />
    </button>
    <button :disabled="!tile.fwd?.length" title="앞으로 (Ctrl+Shift+→)" @click="goHistory(tileId, 1)">
        <v-icon size="14" icon="mdi-arrow-right" />
    </button>
</span>
</template>

<style scoped lang="scss">
.TileNavButtons {
    display: inline-flex;
    align-items: center;
    > button {
        display: inline-flex;
        padding: 2px;
        border: 0;
        background: none;
        color: inherit;
        cursor: pointer;
        opacity: 0.75;
        &:disabled { cursor: default; opacity: 0.25; }
    }
}
</style>
