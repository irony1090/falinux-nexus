<script setup lang="ts">
import ProcessTerminal from '@/feature/process/component/ProcessTerminal.vue';
import { pillOf, useProcessTerm } from '@/feature/process/store/processTerm.store';
import { computed, type PropType } from 'vue';
import { VChip, VIcon } from 'vuetify/components';
import { useGetNode } from '../../api/node.api';
import type { Tile } from '../../store/tileTree.store';

// 전체보기의 다른 보드 타일 — 읽기 전용(조작·입력 없음, 누르면 그 보드로). TileFrame은 지금 보드 스토어를 inject하므로 따로 둔다 — 미리보기-4(그리기)
const props = defineProps({
    tile: {
        type: Object as PropType<Tile>,
        required: true,
    },
    pos: {
        type: Number,
        default: undefined,
    },
})

const { procs } = useProcessTerm();
const { data: node } = useGetNode(computed(() => props.tile.nodeId ?? NaN));

const name = computed(() => props.tile.nodeId === null ? '기본' : node.value?.name ?? `node ${props.tile.nodeId}`);
const proc = computed(() => props.tile.type === 'terminal' ? procs.value[props.tile.uid] : undefined);
const pill = computed(() => pillOf(proc.value?.status, proc.value?.terminating));
</script>

<template>
<div class="OverviewTile" :class="`OverviewTile--${tile.type}`">
    <div class="ov-name">
        <span v-if="pos" class="pos">{{ pos }}</span>
        <v-icon size="14"
            :icon="tile.type === 'folder' ? 'mdi-folder' : 'mdi-console'"
            :color="tile.type === 'folder' ? 'warning' : 'primary'"
        />
        <span class="name">{{ name }}</span>
        <v-chip v-if="tile.type === 'terminal'" size="x-small" variant="tonal" class="pill" :color="pill.color">{{ pill.text }}</v-chip>
    </div>
    <div v-if="tile.type === 'terminal'" class="body bg-terminal" inert>
        <process-terminal :uid="tile.uid" />
    </div>
</div>
</template>

<style scoped lang="scss">
.OverviewTile {
    flex: 1;
    min-width: 0;
    min-height: 0;
    position: relative;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    border-radius: 6px;
    background: rgb(var(--v-theme-surface));
    pointer-events: none;
}
// 전체보기의 폴더는 흐리게(REF-node-ui-overview.md "전체보기 모드")
.OverviewTile--folder { opacity: 0.7; }
// 축소 배율을 거꾸로 곱해 읽을 수 있는 크기 유지 — TileFrame .ov-name과 같은 모양
.ov-name {
    position: absolute;
    left: 0;
    top: 0;
    z-index: 3;
    width: calc(100% * var(--s));
    transform: scale(calc(1 / var(--s)));
    transform-origin: 0 0;
    display: flex;
    align-items: center;
    gap: vars.$spacing-sm;
    padding: 3px 6px;
    font-size: 11.5px;
    background: rgb(var(--v-theme-surface));
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    > * { flex: none; }
    .name {
        flex: 1;
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
}
.pos {
    min-width: 16px;
    height: 16px;
    padding: 0 4px;
    border-radius: 4px;
    display: inline-grid;
    place-items: center;
    font-size: 10px;
    font-weight: vars.$weight-lg;
    background: rgb(var(--v-theme-on-surface));
    color: rgb(var(--v-theme-surface));
}
// 머리 줄(28px) 자리를 비우고 터미널 — TerminalTileBody와 같은 안쪽 여백
.body {
    flex: 1;
    min-height: 0;
    margin-top: 28px;
    padding: 4px 0 0 vars.$spacing-sm;
}
</style>
