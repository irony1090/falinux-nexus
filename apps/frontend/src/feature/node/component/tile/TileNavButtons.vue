<script setup lang="ts">
import { computed } from 'vue';
import { VIcon } from 'vuetify/components';
import { isNavTile, useTileTree } from '../../store/tileTree.store';

// 경로 줄 맨 앞 [뒤로 | 앞으로 | 새로고침] — 뒤로-1(타일 안 이동 기록). 단축키 Ctrl+Shift+←/→ = tileNav.hook.ts
const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
    busy: Boolean,
})
const emit = defineEmits<{ refresh: [] }>();

const { tiles, goHistory } = useTileTree();
const tile = computed(() => {
    const t = tiles.value[props.tileId];
    return isNavTile(t) ? t : undefined;
})
</script>

<template>
<span class="TileNavButtons ct-grp ct-grp--l ct-grp--keep">
    <button type="button" class="ct-seg ct-seg--dim" :disabled="!tile?.back?.length" title="뒤로 (Ctrl+Shift+←)" @click="goHistory(tileId, -1)">
        <v-icon size="16" icon="mdi-arrow-left" />
    </button>
    <button type="button" class="ct-seg ct-seg--dim" :disabled="!tile?.fwd?.length" title="앞으로 (Ctrl+Shift+→)" @click="goHistory(tileId, 1)">
        <v-icon size="16" icon="mdi-arrow-right" />
    </button>
    <button type="button" class="ct-seg ct-seg--dim" title="새로고침" :disabled="busy" @click="emit('refresh')">
        <v-icon size="15" icon="mdi-refresh" />
    </button>
</span>
</template>

<style scoped lang="scss">
// 타일 왼쪽 벽에 붙는 묶음(찰떡 규칙) — 높이 = 경로 줄
.TileNavButtons {
    height: 28px;
    margin-right: 4px;
    > .ct-seg { min-width: 42px; }
    > .ct-seg:first-child, > .ct-seg:last-child { min-width: 44px; }
}
</style>
