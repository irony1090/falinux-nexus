<script setup lang="ts">
import ProcessTerminal from '@/feature/process/component/ProcessTerminal.vue';
import { computed } from 'vue';
import { useTileTree } from '../../store/tileTree.store';

const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const { tiles } = useTileTree();

const tile = computed(() => {
    const t = tiles.value[props.tileId];
    return t?.type === 'terminal' ? t : undefined;
})
</script>

<template>
<div class="TerminalTileBody bg-terminal">
    <process-terminal v-if="tile" :uid="tile.uid" />
</div>
</template>

<style scoped lang="scss">
.TerminalTileBody {
    flex: 1;
    min-height: 0;
    padding: 4px 0 0 vars.$spacing-sm;
}
</style>
