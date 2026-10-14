<script setup lang="ts">
import ProcessTerminal from '@/feature/process/component/ProcessTerminal.vue';
import { useProcessTerm } from '@/feature/process/store/processTerm.store';
import { computed } from 'vue';
import { VChip, VFadeTransition } from 'vuetify/components';
import { useTileTree } from '../../store/tileTree.store';

const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const { tiles } = useTileTree();
const { takenBy } = useProcessTerm();

const tile = computed(() => {
    const t = tiles.value[props.tileId];
    return t?.type === 'terminal' ? t : undefined;
})
</script>

<template>
<div class="TerminalTileBody bg-terminal">
    <process-terminal v-if="tile" :uid="tile.uid" />
    <v-fade-transition>
        <v-chip v-if="tile && takenBy[tile.uid]" class="taken" size="small" variant="flat" color="warning" prepend-icon="mdi-arrow-expand-all">
            {{ takenBy[tile.uid] }}님이 크기를 가져갔습니다
        </v-chip>
    </v-fade-transition>
</div>
</template>

<style scoped lang="scss">
.TerminalTileBody {
    position: relative;
    flex: 1;
    min-height: 0;
    padding: 4px 0 0 vars.$spacing-sm;
}
.taken {
    position: absolute;
    top: 8px;
    right: 8px;
    pointer-events: none;
}
</style>
