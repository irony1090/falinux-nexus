<script setup lang="ts">
import { TILE_KINDS } from '@/feature/widget/util/tileKind.util';
import { useWorkerRoster } from '@/feature/worker/api/worker.api';
import { viewIcon, viewKindOf } from '@/feature/worker/util/fileKind.util';
import { computed } from 'vue';
import { useGetNode } from '../../api/node.api';
import { useTerminalKill } from '../../hook/terminalKill.hook';
import { useTileGrids } from '../../store/tileGrids.store';
import { useTileTree } from '../../store/tileTree.store';
import { fileLabel } from '../../util/tileLabel.util';
import TileFrame from './TileFrame.vue';
import VideoViewBody from './VideoViewBody.vue';
import ViewerTileBody from './ViewerTileBody.vue';

// 보기 타일 — 장비 파일 하나(파트(보기 타일), REF-file-view.md). 권한·연결은 연 폴더(nodeId) + 인스턴스로 장비 타일과 같음
const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const { tiles } = useTileTree();
const { overview } = useTileGrids();
const { closeTile } = useTerminalKill();

const tile = computed(() => {
    const t = tiles.value[props.tileId];
    return t?.type === 'viewer' ? t : undefined;
})

const name = computed(() => tile.value ? fileLabel(tile.value.path) : '');
const { devices: roster } = useWorkerRoster();
const instLabel = computed(() => {
    const key = tile.value?.instance ?? '';
    return roster.value.get(key.slice(0, key.lastIndexOf('#')))?.instances.find(i => i.instanceKey === key)?.label;
});

// 볼 수 없음 = 연 폴더가 지워짐(404) — 장비 타일과 같음
const { error: folderError } = useGetNode(computed(() => tile.value?.nodeId ?? NaN));
const unavailable = computed(() => (folderError.value as { status?: number } | null)?.status === 404);
</script>

<template>
<tile-frame v-if="tile" :tile-id="tile.id"
    :name="name"
    :icon="viewIcon(tile.path)"
    :color="TILE_KINDS.viewer.color"
    :unavailable="unavailable"
    @close="closeTile(tile.id)"
>
    <template #title>
        <span class="name" :title="tile.path">{{ name }}</span>
        <span class="inst" :title="`인스턴스 ${tile.instance}`">{{ instLabel || tile.instance }}</span>
    </template>

    <video-view-body v-if="viewKindOf(tile.path) === 'video'" :tile-id="tile.id" :inert="overview" />
    <viewer-tile-body v-else :tile-id="tile.id" :inert="overview" />
</tile-frame>
</template>

<style scoped lang="scss">
.TileFrame__head > .name {
    flex: 0 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-weight: vars.$weight-md;
}
.TileFrame__head > .inst {
    flex: 0 1000 auto;  // 이름보다 먼저 줄어든다
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    opacity: 0.6;
    font-family: monospace;
    @include vars.font('body-small');
}
</style>
