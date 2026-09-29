<script setup lang="ts">
import type { TileSpan } from '@/feature/widget/util/tile.type';
import { effectiveSize } from '@/feature/widget/util/tileSplit.util';
import { computed } from 'vue';
import { VBtn, VCard, VChip, VIcon, VMenu } from 'vuetify/components';
import { dummyKill, dummyNode, dummyStatus } from '../../dev/tileDummy';
import { useTileGrids } from '../../store/tileGrids.store';
import { useTileTree } from '../../store/tileTree.store';
import FolderTileBody from './FolderTileBody.vue';
import TerminalTileBody from './TerminalTileBody.vue';

const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const { tiles, close, resize } = useTileTree();
const { cap, posOf, instanceOf } = useTileGrids();

const tile = computed(() => tiles.value[props.tileId]);

const name = computed(() => {
    const t = tile.value;
    if (!t) return '';
    if (t.type === 'folder') return dummyNode(t.nodeId)?.name ?? '기본';
    const no = instanceOf.value.get(t.id);
    return (dummyNode(t.nodeId)?.name ?? `node ${t.nodeId}`) + (no ? ` (${no})` : '');
})
const opener = computed(() => {
    const parent = tile.value?.parent ? tiles.value[tile.value.parent] : undefined;
    return parent?.type === 'folder' ? dummyNode(parent.nodeId)?.name ?? '기본' : undefined;
})

const running = computed(() => tile.value?.type === 'terminal' && dummyStatus(tile.value.uid) === 'RUNNING');
// 실행 중 터미널은 닫기 비활성(kill 먼저) — 결정 A(닫기와 kill)
const closable = computed(() => !!tile.value?.parent && !running.value);

const fmt = (v: TileSpan) => v === 1 ? '1' : '.5';
const sizeText = computed(() => tile.value ? `${fmt(tile.value.size.w)}×${fmt(tile.value.size.h)}` : '');
const shownText = computed(() => {
    if (!tile.value) return '';
    const s = effectiveSize(tile.value.size, cap.value);
    const text = `${fmt(s.w)}×${fmt(s.h)}`;
    return text === sizeText.value ? '' : text;
})

const setSize = (axis: 'w' | 'h', v: TileSpan) => {
    if (!tile.value) return;
    resize(tile.value.id, { ...tile.value.size, [axis]: v });
}
const onKill = () => {
    if (tile.value?.type === 'terminal') dummyKill(tile.value.uid);
}
</script>

<template>
<div v-if="tile" class="TileFrame" :class="`TileFrame--${tile.type}`">
    <div class="TileFrame__head">
        <span class="pos">{{ posOf.get(tile.id) }}</span>
        <v-icon size="14"
            :icon="tile.type === 'folder' ? 'mdi-folder' : 'mdi-console'"
            :color="tile.type === 'folder' ? 'warning' : 'primary'"
        />
        <span class="name">{{ name }}</span>
        <span v-if="opener" class="opener text-warning">← {{ opener }}</span>
        <span class="spacer" />

        <template v-if="tile.type === 'terminal'">
            <v-chip size="x-small" variant="tonal" class="pill" :color="running ? 'success' : undefined">{{ running ? 'RUN' : 'DONE' }}</v-chip>
            <v-btn v-if="running" size="x-small" variant="text" icon="mdi-stop" title="kill" @click="onKill" />
        </template>

        <v-menu location="bottom end" :close-on-content-click="false">
            <template #activator="{ props: act }">
                <button v-bind="act" class="size-btn" title="크기 조절">
                    {{ sizeText }}<i v-if="shownText" class="text-warning" title="이 화면에서는 이 크기로 보임">={{ shownText }}</i>
                </button>
            </template>
            <v-card class="size-pop">
                <div v-for="axis in (['w', 'h'] as const)" :key="axis" class="size-row">
                    <span>{{ axis === 'w' ? '가로' : '세로' }}</span>
                    <v-btn v-for="v in ([0.5, 1] as const)" :key="v"
                        size="x-small"
                        :variant="tile.size[axis] === v ? 'flat' : 'outlined'"
                        :color="tile.size[axis] === v ? 'primary' : undefined"
                        :disabled="v === 0.5 && cap[axis === 'w' ? 'x' : 'y'] === 1"
                        @click="setSize(axis, v)"
                    >{{ v.toFixed(1) }}</v-btn>
                </div>
                <p v-if="cap.x === 1 || cap.y === 1" class="size-hint">1칸인 축의 0.5는 이 화면에선 1.0으로 보임</p>
            </v-card>
        </v-menu>

        <v-btn size="x-small" variant="text" icon="mdi-close"
            :disabled="!closable"
            :title="running ? '실행 중에는 닫을 수 없음 (kill 먼저)' : '닫기'"
            @click="close(tile.id)"
        />
    </div>

    <folder-tile-body v-if="tile.type === 'folder'" :tile-id="tile.id" />
    <terminal-tile-body v-else :tile-id="tile.id" />
</div>
</template>

<style scoped lang="scss">
.TileFrame {
    flex: 1;
    min-width: 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    border-radius: 6px;
    background: rgb(var(--v-theme-surface));
}
.TileFrame__head {
    flex: 0 0 28px;
    display: flex;
    align-items: center;
    gap: vars.$spacing-sm;
    padding: 0 vars.$spacing-xs 0 vars.$spacing-md;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    font-size: 12px;
    min-width: 0;
    > * { flex: none; }
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
.name {
    flex: 0 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-weight: vars.$weight-md;
}
.opener {
    flex: 0 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 11px;
}
.spacer { flex: 1; }
.pill { font-family: monospace; }
.size-btn {
    height: 20px;
    padding: 0 5px;
    border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    border-radius: 4px;
    background: none;
    color: inherit;
    font-family: monospace;
    font-size: 10px;
    cursor: pointer;
    i { font-style: normal; margin-left: 2px; }
}
.size-pop {
    display: flex;
    flex-direction: column;
    gap: vars.$spacing-sm;
    padding: vars.$spacing-md vars.$spacing-lg;
    font-size: 12px;
}
.size-row {
    display: flex;
    align-items: center;
    gap: vars.$spacing-sm;
    > span { width: 30px; opacity: 0.7; }
}
.size-hint {
    max-width: 180px;
    margin: 0;
    font-size: 10.5px;
    opacity: 0.7;
}
</style>
