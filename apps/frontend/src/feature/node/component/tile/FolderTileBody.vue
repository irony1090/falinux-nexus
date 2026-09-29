<script setup lang="ts">
import { computed } from 'vue';
import { VBtn, VIcon } from 'vuetify/components';
import { dummyChildren, dummyExec, dummyPath } from '../../dev/tileDummy';
import { useTileGrids } from '../../store/tileGrids.store';
import { useTileTree } from '../../store/tileTree.store';

const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const { tiles, navigate, openFolder, addTerminal } = useTileTree();
const { newSize } = useTileGrids();

const nodeId = computed(() => {
    const t = tiles.value[props.tileId];
    return t?.type === 'folder' ? t.nodeId : null;
})
const path = computed(() => dummyPath(nodeId.value));
const children = computed(() => dummyChildren(nodeId.value));

// 새 타일 크기 = 누른 화면의 최소 단위
const onOpen = (id: number) => openFolder(props.tileId, id, newSize.value);
const onExec = (id: number) => addTerminal(props.tileId, id, dummyExec(), newSize.value);
</script>

<template>
<div class="FolderTileBody">
    <div class="crumbs">
        <button v-if="path.length" @click="navigate(tileId, null)">기본</button>
        <span v-else class="cur">기본</span>
        <template v-for="(p, i) in path" :key="p.id">
            <span>/</span>
            <span v-if="i === path.length - 1" class="cur">{{ p.name }}</span>
            <button v-else @click="navigate(tileId, p.id)">{{ p.name }}</button>
        </template>
    </div>
    <div v-for="(n, i) in children" :key="n.id" class="row">
        <span class="ord">{{ i + 1 }}</span>
        <template v-if="n.kind === 'FOLDER'">
            <v-icon size="14" icon="mdi-folder" class="folder" />
            <button class="nm" title="이 타일 안에서 열기" @click="navigate(tileId, n.id)">{{ n.name }}</button>
            <v-btn size="x-small" variant="tonal" color="warning" @click="onOpen(n.id)">새 타일</v-btn>
        </template>
        <template v-else>
            <v-icon size="14" icon="mdi-console" class="script" />
            <span class="nm">{{ n.name }}</span>
            <v-btn size="x-small" variant="tonal" color="primary" @click="onExec(n.id)">실행</v-btn>
        </template>
    </div>
    <p v-if="!children.length" class="empty">비어 있음</p>
</div>
</template>

<style scoped lang="scss">
.FolderTileBody {
    flex: 1;
    min-height: 0;
    overflow: auto;
    font-size: 12px;
    button {
        padding: 0;
        border: 0;
        background: none;
        color: inherit;
        font: inherit;
        cursor: pointer;
    }
}
.crumbs {
    display: flex;
    flex-wrap: wrap;
    gap: 2px;
    padding: 5px vars.$spacing-md;
    font-size: 11px;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    button { color: rgb(var(--v-theme-primary)); }
    .cur { font-weight: vars.$weight-md; }
}
.row {
    display: grid;
    grid-template-columns: 20px 16px minmax(0, 1fr) auto;
    align-items: center;
    gap: 6px;
    min-height: 30px;
    padding: vars.$spacing-sm 6px vars.$spacing-sm vars.$spacing-md;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}
.ord { font-family: monospace; font-size: 10px; opacity: 0.6; }
.folder { color: rgb(var(--v-theme-warning)); }
.script { color: rgb(var(--v-theme-primary)); }
.nm {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-align: left;
}
button.nm:hover { color: rgb(var(--v-theme-primary)); }
.empty { margin: vars.$spacing-md; opacity: 0.6; }
</style>
