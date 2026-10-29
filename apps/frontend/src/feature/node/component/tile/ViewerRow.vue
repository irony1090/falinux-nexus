<script setup lang="ts">
import { viewIcon } from '@/feature/worker/util/fileKind.util';
import { parentPathOf } from '@/feature/worker/util/workerDir.util';
import { computed, type PropType } from 'vue';
import { useRouter } from 'vue-router';
import { VIcon } from 'vuetify/components';
import type { ViewerResponse } from '../../api/tile.api';
import { useTerminalKill } from '../../hook/terminalKill.hook';
import { boardLabel, focusPath, useBoards } from '../../store/boards.store';
import { useTileGrids } from '../../store/tileGrids.store';
import { useTileTree } from '../../store/tileTree.store';
import { fileLabel, shortPath } from '../../util/tileLabel.util';

// 터미널 탭의 보기 타일 한 행 — 이동(이 보드 = 보기 / 다른 보드 = 보드로 이동) + 닫기. 모양은 RunRow와 같은 줄(파트(보기 타일))
const props = defineProps({
    viewer: {
        type: Object as PropType<ViewerResponse>,
        required: true,
    },
})

const router = useRouter();
const { tiles, hash: boardHash } = useTileTree();
const { posOf, activate } = useTileGrids();
const { list: boards } = useBoards();
const { closeTile } = useTerminalKill();

const local = computed(() => {
    const v = props.viewer;
    const t = v.hash === boardHash.value ? tiles.value[v.tileId] : undefined;
    return t?.type === 'viewer' ? t : undefined;
})
const board = computed(() => !local.value ? boards.value.find(b => b.hash === props.viewer.hash) : undefined);
const dirText = computed(() => shortPath(parentPathOf(props.viewer.path)));

const onOpen = () => {
    if (local.value) activate(local.value.id);
    else if (board.value) router.push(focusPath(board.value, props.viewer.tileId));
}
</script>

<template>
<div class="ViewerRow">
    <span class="ct-grp ct-grp--l ct-grp--sm">
        <span v-if="local" class="ct-seg pos">{{ posOf.get(local.id) }}</span>
        <span v-else class="ct-seg pos none">-</span>
    </span>
    <span class="where">
        <v-icon size="13" :icon="viewIcon(viewer.path)" color="success" />
        <span class="path" :title="viewer.path"> {{ dirText }} / </span>{{ fileLabel(viewer.path) }}
        <span class="by"> · </span>
        <template v-if="local">이 보드</template>
        <template v-else-if="board">{{ boardLabel(board) }}</template>
        <template v-else>위치 없음</template>
    </span>
    <span class="ct-grp ct-grp--r ct-grp--sm">
        <button v-if="local" type="button" class="ct-seg" @click="onOpen">보기</button>
        <button v-else-if="board" type="button" class="ct-seg" @click="onOpen">보드로 이동</button>
        <button v-if="local" type="button" class="ct-seg ct-seg--icon" title="닫기" @click="closeTile(local.id)"><v-icon size="15" icon="mdi-close" /></button>
    </span>
</div>
</template>

<style scoped lang="scss">
.ViewerRow {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto;
    align-items: center;
    gap: vars.$spacing-md;
    min-height: 32px;
    padding: vars.$spacing-xs 0;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    @include vars.font('body-small');
}
.pos {
    min-width: 26px;
    font-weight: vars.$weight-lg;
    &:not(.none) {
        background: rgb(var(--v-theme-on-surface));
        color: rgb(var(--v-theme-surface));
    }
    &.none { opacity: 0.6; }
}
.where {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.path, .by { opacity: 0.6; }
</style>
