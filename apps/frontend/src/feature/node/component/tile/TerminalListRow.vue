<script setup lang="ts">
import { isRunning, STATUS_PILL, useProcessTerm } from '@/feature/process/store/processTerm.store';
import { computed } from 'vue';
import { VBtn, VChip, VIcon } from 'vuetify/components';
import { useGetNode, useNodePath } from '../../api/node.api';
import { useTerminalKill } from '../../hook/terminalKill.hook';
import { useTileGrids } from '../../store/tileGrids.store';
import { useTileTree } from '../../store/tileTree.store';

const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const { tiles, setMinimized } = useTileTree();
const { posOf, instanceOf, reveal } = useTileGrids();
const { procs, restored } = useProcessTerm();
const { killing, closeTile, confirmKill } = useTerminalKill();

const tile = computed(() => {
    const t = tiles.value[props.tileId];
    return t?.type === 'terminal' ? t : undefined;
})

const { data: node } = useGetNode(computed(() => tile.value?.nodeId ?? NaN));
const { data: scriptPath } = useNodePath(computed(() => tile.value?.nodeId ?? null));
const pathText = computed(() => (scriptPath.value ?? []).slice(0, -1).map(p => p.name).join(' / '));
const name = computed(() => {
    const t = tile.value;
    if (!t) return '';
    const no = instanceOf.value.get(t.id);
    return (node.value?.name ?? `node ${t.nodeId}`) + (no ? ` (${no})` : '');
})

const proc = computed(() => tile.value ? procs.value[tile.value.uid] : undefined);
const running = computed(() => isRunning(proc.value?.status));
const pill = computed(() => proc.value ? STATUS_PILL[proc.value.status] : { text: '-' });
// 복원 전엔 상태를 몰라 닫기를 막는다 — 3-d②(복원 전 닫기)
const closable = computed(() => !running.value && (!!proc.value || restored.value));

const onOpen = () => {
    const t = tile.value;
    if (!t) return;
    if (t.minimized) setMinimized(t.id, false);
    reveal(t.id);
}
</script>

<template>
<div v-if="tile" class="TerminalListRow">
    <span v-if="tile.minimized" class="min" title="최소화"><v-icon size="11" icon="mdi-window-minimize" /></span>
    <span v-else class="pos">{{ posOf.get(tile.id) }}</span>
    <v-icon size="14" icon="mdi-console" color="primary" />
    <span class="nm" :title="pathText ? `${pathText} / ${name}` : name">
        <span v-if="pathText" class="path">{{ pathText }} / </span>{{ name }}
    </span>
    <v-chip size="x-small" variant="tonal" class="pill" :color="pill.color">{{ pill.text }}</v-chip>
    <span class="acts">
        <v-btn size="x-small" variant="tonal" :color="tile.minimized ? 'primary' : undefined" @click="onOpen">{{ tile.minimized ? '열기' : '보기' }}</v-btn>
        <v-btn v-if="running" size="x-small" variant="text" icon="mdi-stop" title="종료…" :loading="!!killing[tile.uid]" @click="confirmKill(tile.id, name)" />
        <v-btn v-else size="x-small" variant="text" icon="mdi-close" title="닫기" :disabled="!closable" @click="closeTile(tile.id)" />
    </span>
</div>
</template>

<style scoped lang="scss">
.TerminalListRow {
    display: grid;
    grid-template-columns: 20px 16px minmax(0, 1fr) auto auto;
    align-items: center;
    gap: 6px;
    min-height: 30px;
    padding: vars.$spacing-sm vars.$spacing-xs vars.$spacing-sm vars.$spacing-md;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}
.pos, .min {
    min-width: 16px;
    height: 16px;
    padding: 0 4px;
    border-radius: 4px;
    display: inline-grid;
    place-items: center;
    font-size: 10px;
    font-weight: vars.$weight-lg;
}
.pos {
    background: rgb(var(--v-theme-on-surface));
    color: rgb(var(--v-theme-surface));
}
.min {
    border: 1px dashed rgba(var(--v-border-color), 0.6);
    opacity: 0.75;
}
.nm {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.path { opacity: 0.6; }
.pill { font-family: monospace; }
.acts {
    display: flex;
    align-items: center;
    gap: vars.$spacing-xs;
}
</style>
