<script setup lang="ts">
import { TILE_KINDS } from '@/feature/widget/util/tileKind.util';
import { computed } from 'vue';
import { VListItem, VListSubheader } from 'vuetify/components';
import { useGetNode } from '../../api/node.api';
import { useTerminalKill } from '../../hook/terminalKill.hook';
import { useTileBoard } from '../../hook/tileBoard.hook';
import { boardLabel } from '../../store/boards.store';
import { useTileGrids } from '../../store/tileGrids.store';
import { useTileTree } from '../../store/tileTree.store';
import { deviceLabel, nodeLabel } from '../../util/tileLabel.util';
import NodeNameDialog from '../dialog/NodeNameDialog.vue';
import DeviceTileBody from './DeviceTileBody.vue';
import RootTabs from './RootTabs.vue';
import RootTerminalList from './RootTerminalList.vue';
import TileFrame from './TileFrame.vue';
import TileNavButtons from './TileNavButtons.vue';

// 장비 타일 — worker 디렉토리(폴더-2(장비 타일), REF-worker-dir-list.md). 연 폴더 = nodeId. 루트면 탭 `[파일] [터미널 N · 최소화 n]`(이동-2(루트 탭))
const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const { tiles } = useTileTree();
const { rootTab, overview } = useTileGrids();
const { closeTile } = useTerminalKill();

const tile = computed(() => {
    const t = tiles.value[props.tileId];
    return t?.type === 'device' ? t : undefined;
})

const isRoot = computed(() => !tile.value?.parent);

const { data: folder, error: folderError } = useGetNode(computed(() => tile.value?.nodeId ?? NaN));
const name = computed(() => tile.value ? deviceLabel(tile.value.instance, tile.value.path) : '');
const opener = computed(() => tile.value ? nodeLabel(tile.value.nodeId, folder.value?.name) : undefined);

// 볼 수 없음 = 연 폴더가 지워짐(404). 루트는 그대로 두고 표시만 — Q26(연 폴더 삭제 때)
const unavailable = computed(() => (folderError.value as { status?: number } | null)?.status === 404);

// 보드: 루트가 아니면 하위째 새 보드로 분리 / 다른 보드로 이동 — 이동-1(메뉴), 폴더 타일과 같음
const board = useTileBoard(() => props.tileId);
const { otherBoards, busy: boardBusy, detachOpen } = board;
const boardMenu = computed(() => !!tile.value?.parent && !unavailable.value);
</script>

<template>
<tile-frame v-if="tile" :tile-id="tile.id"
    :name="name"
    :icon="TILE_KINDS.device.icon"
    :color="TILE_KINDS.device.color"
    :unavailable="unavailable"
    :menu="boardMenu"
    :menu-busy="boardBusy"
    @close="closeTile(tile.id)"
>
    <template #title>
        <span class="name">{{ name }}</span>
        <span class="inst" :title="`인스턴스 ${tile.instance}`">{{ tile.instance }}</span>
        <span v-if="opener" class="opener text-warning">← {{ opener }}</span>
    </template>

    <template #controls>
        <tile-nav-buttons v-if="!unavailable" :tile-id="tile.id" />
    </template>

    <template v-if="boardMenu" #menu>
        <v-list-item title="새 보드로 분리" prepend-icon="mdi-view-dashboard-outline" @click="detachOpen = true" />
        <template v-if="otherBoards.length">
            <v-list-subheader>다른 보드로 이동</v-list-subheader>
            <v-list-item v-for="b in otherBoards" :key="b.hash"
                :title="boardLabel(b)"
                :prepend-icon="b.isDefault ? 'mdi-home-outline' : 'mdi-arrow-right'"
                @click="board.move(b.hash)"
            />
        </template>
    </template>

    <template #tabs>
        <root-tabs v-if="isRoot && !unavailable" />
    </template>

    <template v-if="isRoot" #unavailable>
        <span class="sub">장비 타일을 연 폴더가 지워졌습니다. 보드를 삭제하면 이 타일이 기본 보드로 옮겨져 닫을 수 있습니다.</span>
    </template>

    <device-tile-body v-show="!isRoot || rootTab === 'files'" :tile-id="tile.id" :inert="overview" />
    <root-terminal-list v-if="isRoot && rootTab === 'terminals'" :inert="overview" />

    <template #extra>
        <node-name-dialog v-if="boardMenu" v-model="detachOpen" title="새 보드" :initial="name" :loading="boardBusy" @submit="board.detach" />
    </template>
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
.TileFrame__head > .inst,
.TileFrame__head > .opener {
    flex: 0 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    @include vars.font('body-small');
}
.TileFrame__head > .inst { opacity: 0.6; font-family: monospace; }
</style>
