<script setup lang="ts">
import type { TileSpan } from '@/feature/widget/util/tile.type';
import { effectiveSize } from '@/feature/widget/util/tileSplit.util';
import { useAppDialog } from '@/feature/layout/store/appDialog.store';
import type { ProcessStatus } from '@/feature/process/api/process.api';
import { isRunning, useProcessTerm } from '@/feature/process/store/processTerm.store';
import { computed, ref } from 'vue';
import { VBtn, VCard, VChip, VIcon, VMenu } from 'vuetify/components';
import { useGetNode } from '../../api/node.api';
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
const { cap, posOf, instanceOf, reveal, overview } = useTileGrids();
const { procs, restored, kill, dispose } = useProcessTerm();
const { openDialog } = useAppDialog();

const tile = computed(() => tiles.value[props.tileId]);

const openerTile = computed(() => tile.value?.parent ? tiles.value[tile.value.parent] : undefined);

// useGetNode는 id가 NaN이면 조회하지 않는다 (nodeId null = 루트 목록)
const { data: node } = useGetNode(computed(() => tile.value?.nodeId ?? NaN));
const { data: openerNode } = useGetNode(computed(() => openerTile.value?.nodeId ?? NaN));

// 이름을 받기 전에는 node id로 표시
const label = (nodeId: number | null, loaded?: string) => nodeId === null ? '기본' : loaded ?? `node ${nodeId}`;

const name = computed(() => {
    const t = tile.value;
    if (!t) return '';
    const no = t.type === 'terminal' ? instanceOf.value.get(t.id) : undefined;
    return label(t.nodeId, node.value?.name) + (no ? ` (${no})` : '');
})
const opener = computed(() => {
    const parent = openerTile.value;
    return parent?.type === 'folder' ? label(parent.nodeId, openerNode.value?.name) : undefined;
})

const proc = computed(() => tile.value?.type === 'terminal' ? procs.value[tile.value.uid] : undefined);
const running = computed(() => isRunning(proc.value?.status));

// PENDING = worker 연결이 끊겨 기다리는 중(프로세스는 살아 있음)
const PILL: Record<ProcessStatus, { text: string; color?: string }> = {
    PENDING: { text: 'WAIT', color: 'warning' },
    PROCESS: { text: 'RUN', color: 'success' },
    COMPLETED: { text: 'DONE' },
    FAILED: { text: 'FAIL', color: 'error' },
};
const pill = computed(() => proc.value ? PILL[proc.value.status] : { text: '-' });
// 실행 중 터미널은 닫기 비활성(kill 먼저) — 결정 A(닫기와 kill). 복원 전엔 상태를 몰라 막는다 — 3-d②(복원 전 닫기)
const unknown = computed(() => tile.value?.type === 'terminal' && !proc.value && !restored.value);
const closable = computed(() => !!tile.value?.parent && !running.value && !unknown.value);

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
    reveal(tile.value.id);
}
const killing = ref(false);
const onKill = () => {
    if (tile.value?.type !== 'terminal') return;
    killing.value = true;
    kill(tile.value.uid)
    .catch(err => openDialog({ type: 'error', content: err?.message || '종료하지 못했습니다' }))
    .finally(() => killing.value = false);
}
const onClose = () => {
    const t = tile.value;
    if (!t) return;
    close(t.id);
    if (t.type === 'terminal') dispose(t.uid);
}
</script>

<template>
<div v-if="tile" class="TileFrame" :class="[`TileFrame--${tile.type}`, { 'TileFrame--overview': overview }]">
    <!-- 전체보기 라벨: 축소 배율을 거꾸로 곱해 읽을 수 있는 크기 유지 -->
    <div v-if="overview" class="ov-name">
        <span class="pos">{{ posOf.get(tile.id) }}</span>
        <v-icon size="14"
            :icon="tile.type === 'folder' ? 'mdi-folder' : 'mdi-console'"
            :color="tile.type === 'folder' ? 'warning' : 'primary'"
        />
        <span class="name">{{ name }}</span>
        <v-chip v-if="tile.type === 'terminal'" size="x-small" variant="tonal" class="pill" :color="pill.color">{{ pill.text }}</v-chip>
    </div>
    <div class="TileFrame__head" :inert="overview">
        <span class="pos">{{ posOf.get(tile.id) }}</span>
        <v-icon size="14"
            :icon="tile.type === 'folder' ? 'mdi-folder' : 'mdi-console'"
            :color="tile.type === 'folder' ? 'warning' : 'primary'"
        />
        <span class="name">{{ name }}</span>
        <span v-if="opener" class="opener text-warning">← {{ opener }}</span>
        <span class="spacer" />

        <template v-if="tile.type === 'terminal'">
            <v-chip size="x-small" variant="tonal" class="pill" :color="pill.color" :title="proc?.exitCode != null ? `exit ${proc.exitCode}` : undefined">{{ pill.text }}</v-chip>
            <v-btn v-if="running" size="x-small" variant="text" icon="mdi-stop" title="kill" :loading="killing" @click="onKill" />
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
            @click="onClose"
        />
    </div>

    <folder-tile-body v-if="tile.type === 'folder'" :tile-id="tile.id" :inert="overview" />
    <terminal-tile-body v-else :tile-id="tile.id" :inert="overview" />
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
    position: relative;
}
// 전체보기: 안쪽 조작 막음(inert + 클릭은 Tile로), 폴더는 흐리게
.TileFrame--overview {
    > :not(.ov-name) { pointer-events: none; }
    &.TileFrame--folder { opacity: 0.72; }
}
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
    pointer-events: none;
    > * { flex: none; }
    .name { flex: 1; }
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
