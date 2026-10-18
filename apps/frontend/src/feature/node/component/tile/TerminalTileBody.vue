<script setup lang="ts">
import { useVirtualKeyboard } from '@/common/hook/virtualKeyboard.hook';
import ProcessTerminal from '@/feature/process/component/ProcessTerminal.vue';
import { takeCsiMod } from '@/feature/process/store/keyModifier.store';
import { useProcessTerm } from '@/feature/process/store/processTerm.store';
import { computed } from 'vue';
import { VBtn, VChip, VFadeTransition } from 'vuetify/components';
import { useTileGrids } from '../../store/tileGrids.store';
import { useTileTree } from '../../store/tileTree.store';
import { isReadOnly } from '../../util/shareView.util';

const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const { tiles } = useTileTree();
const { takenBy, procs, sendText } = useProcessTerm();
const { activeId, overview } = useTileGrids();
const kb = useVirtualKeyboard();

const tile = computed(() => {
    const t = tiles.value[props.tileId];
    return t?.type === 'terminal' ? t : undefined;
})

// active 터미널의 PgUp·PgDn 버튼 — PG1(버튼 위치)·PG2(키보드 열림 시 숨김) → REF-term-page-keys.md
const showPage = computed(() => {
    const proc = tile.value && procs.value[tile.value.uid];
    return activeId.value === props.tileId && !overview.value && !kb.open
        && proc?.status === 'PROCESS' && !isReadOnly(proc.permission);
})

const sendPage = (code: 5 | 6) => {
    if (!tile.value) return;
    const n = takeCsiMod();
    sendText(tile.value.uid, n > 1 ? `\x1b[${code};${n}~` : `\x1b[${code}~`);
}
</script>

<template>
<div class="TerminalTileBody bg-terminal">
    <process-terminal v-if="tile" :uid="tile.uid" />
    <v-fade-transition>
        <v-chip v-if="tile && takenBy[tile.uid]" class="taken" size="small" variant="flat" color="warning" prepend-icon="mdi-arrow-expand-all">
            {{ takenBy[tile.uid] }}님이 크기를 가져갔습니다
        </v-chip>
    </v-fade-transition>
    <div v-if="showPage" class="page">
        <!-- pointerdown 기본 동작을 막아 xterm 포커스를 지킨다 -->
        <v-btn size="small" variant="tonal" icon="mdi-chevron-double-up" title="Page Up" @pointerdown.prevent @click="sendPage(5)" />
        <v-btn size="small" variant="tonal" icon="mdi-chevron-double-down" title="Page Down" @pointerdown.prevent @click="sendPage(6)" />
    </div>
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
.page {
    position: absolute;
    right: 12px;
    bottom: 8px;
    z-index: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
    opacity: 0.6;
    &:hover { opacity: 1; }
}
</style>
