<script setup lang="ts">
import { TILE_KINDS } from '@/feature/widget/util/tileKind.util';
import { computed } from 'vue';
import { VBtn, VChip, VIcon, VListItem, VListSubheader } from 'vuetify/components';
import { useGetNode } from '../../api/node.api';
import { bindTile } from '../../api/tile.api';
import { useTerminalKill } from '../../hook/terminalKill.hook';
import { useTileBoard } from '../../hook/tileBoard.hook';
import { boardLabel } from '../../store/boards.store';
import { useRemoteActive } from '../../store/remoteActive.store';
import { useTileGrids } from '../../store/tileGrids.store';
import { useTileTree } from '../../store/tileTree.store';
import { nodeLabel } from '../../util/tileLabel.util';
import { permInfo } from '../../util/shareView.util';
import NodeNameDialog from '../dialog/NodeNameDialog.vue';
import FolderTileBody from './FolderTileBody.vue';
import RootTabs from './RootTabs.vue';
import RootTerminalList from './RootTerminalList.vue';
import TileFrame from './TileFrame.vue';

// 폴더 타일 — 계정 폴더 노드 목록. 루트면 탭 `[파일] [터미널 N · 최소화 n]`
const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const { tiles, hash: boardHash, pinByNode, navigate } = useTileTree();
const { rootTab, overview } = useTileGrids();
const { closeTile } = useTerminalKill();

const tile = computed(() => {
    const t = tiles.value[props.tileId];
    return t?.type === 'folder' ? t : undefined;
})
const isRoot = computed(() => !tile.value?.parent);
const openerTile = computed(() => tile.value?.parent ? tiles.value[tile.value.parent] : undefined);

// useGetNode는 id가 NaN이면 조회하지 않는다 (nodeId null = 루트 목록)
const { data: node, error: nodeError } = useGetNode(computed(() => tile.value?.nodeId ?? NaN));
const { data: openerNode } = useGetNode(computed(() => openerTile.value?.type === 'folder' ? openerTile.value.nodeId ?? NaN : NaN));

const name = computed(() => tile.value ? nodeLabel(tile.value.nodeId, node.value?.name) : '');
const opener = computed(() => {
    const parent = openerTile.value;
    return parent?.type === 'folder' ? nodeLabel(parent.nodeId, openerNode.value?.name) : undefined;
})

// 공유-e-1(받는 쪽 표시): 공유받은 폴더 = 노드 응답에 소유자 닉네임
const shareOwner = computed(() => node.value?.ownerNickname);
const shareInfo = computed(() => shareOwner.value ? permInfo(node.value?.permission) : undefined);
const shareRoot = computed(() => !!shareOwner.value && node.value?.parentVisible === false); // 공유 시작 지점
const icon = computed(() => shareOwner.value ? 'mdi-folder-account' : TILE_KINDS.folder.icon);

// 볼 수 없음 = 노드 404 — 실시간-6b
const unavailable = computed(() => !!tile.value && tile.value.nodeId !== null && (nodeError.value as { status?: number } | null)?.status === 404);

// ---- 보드: 폴더 타일(루트 제외)을 하위째 새 보드로 분리(어느 보드에서든) / 다른 보드로 이동 — W1(분리 범위)·W3(이동 범위) ----
const board = useTileBoard(() => props.tileId);
const { otherBoards, busy: boardBusy, detachOpen } = board;
const boardMenu = computed(() => !!tile.value?.parent && !unavailable.value);

// 타일 고정 — 같은 폴더 타일이 내 보드 통틀어 2개 이상일 때만 핀·메뉴(R4(고정 표시 남발)), 해제 없음(R5(해제의 의미)). 루트 목록(nodeId null)은 대상 아님
const pin = computed(() => tile.value && tile.value.nodeId !== null ? pinByNode.value.get(tile.value.nodeId) : undefined);
const pinned = computed(() => !!pin.value && pin.value.hash === boardHash.value && pin.value.tileId === tile.value?.id);
const canPin = computed(() => !!pin.value && !pinned.value && !unavailable.value);
const onPin = () => board.run(bindTile(boardHash.value, props.tileId));

// 다른 계정 active — 폴더 = 보여 주는 nodeId(루트 제외). RA1(맞추는 기준)
const { byNodeId } = useRemoteActive();
const remoteUsers = computed(() => {
    const t = tile.value;
    if (!t || unavailable.value || t.nodeId === null) return [];
    return byNodeId.value.get(t.nodeId) ?? [];
})
</script>

<template>
<tile-frame v-if="tile" :tile-id="tile.id"
    :name="name"
    :icon="icon"
    :color="TILE_KINDS.folder.color"
    :shared="!!shareOwner"
    :share-info="shareInfo"
    :unavailable="unavailable"
    :remote-users="remoteUsers"
    :menu="boardMenu || canPin"
    :menu-busy="boardBusy"
    @close="closeTile(tile.id)"
>
    <template #title>
        <span class="name">{{ name }}</span>
        <v-icon v-if="pinned" size="14" icon="mdi-pin" color="warning" title="타일 고정 — 다른 계정이 실행한 터미널을 이 타일이 받음" />
        <v-chip v-if="shareRoot" size="x-small" variant="flat" color="share" prepend-icon="mdi-account" :title="`${shareOwner}님이 공유한 폴더`">{{ shareOwner }}</v-chip>
        <span v-if="opener" class="opener text-warning">← {{ opener }}</span>
    </template>

    <template #menu>
        <v-list-item v-if="canPin" title="여기로 타일 고정" prepend-icon="mdi-pin-outline" @click="onPin" />
        <template v-if="boardMenu">
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
    </template>

    <template #tabs>
        <root-tabs v-if="isRoot" />
    </template>

    <template v-if="isRoot" #unavailable>
        <span class="sub">루트 타일은 닫을 수 없어 기본 목록으로 돌아갑니다.</span>
        <v-btn size="x-small" variant="tonal" color="primary" @click="navigate(tile.id, null)">기본으로 돌아가기</v-btn>
    </template>

    <folder-tile-body v-show="!isRoot || rootTab === 'files'" :tile-id="tile.id" :inert="overview" />
    <root-terminal-list v-if="isRoot && rootTab === 'terminals'" :inert="overview" />

    <template #extra>
        <node-name-dialog v-if="boardMenu" v-model="detachOpen" title="새 보드" :initial="node?.name ?? ''" :loading="boardBusy" @submit="board.detach" />
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
.TileFrame__head > .opener {
    flex: 0 1000 auto;  // 이름보다 먼저 줄어든다
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    @include vars.font('body-small');
}
</style>
