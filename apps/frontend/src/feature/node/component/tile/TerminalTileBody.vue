<script setup lang="ts">
import { computed } from 'vue';
import { dummyNode, dummyStatus } from '../../dev/tileDummy';
import { useTileTree } from '../../store/tileTree.store';

// 더미 출력 — ② ProcessDialog(xterm) 타일 임베드 전까지의 자리
const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const { tiles } = useTileTree();

const tile = computed(() => {
    const t = tiles.value[props.tileId];
    return t?.type === 'terminal' ? t : undefined;
})
const lines = computed(() => tile.value ? dummyNode(tile.value.nodeId)?.lines ?? [] : []);
const done = computed(() => !!tile.value && dummyStatus(tile.value.uid) === 'DONE');
</script>

<template>
<pre class="TerminalTileBody bg-terminal"><span class="dim">{{ tile?.uid }}</span>
<template v-for="(l, i) in lines" :key="i">{{ l }}
</template><span v-if="done" class="dim">[process exited]</span><span v-else class="cursor" /></pre>
</template>

<style scoped lang="scss">
.TerminalTileBody {
    flex: 1;
    min-height: 0;
    margin: 0;
    overflow: auto;
    padding: 6px vars.$spacing-md;
    font-family: monospace;
    font-size: 11px;
    line-height: 1.5;
    white-space: pre;
}
.dim { color: rgba(var(--v-theme-on-terminal), var(--v-medium-emphasis-opacity)); }
.cursor {
    display: inline-block;
    width: 6px;
    height: 12px;
    vertical-align: -2px;
    background: rgb(var(--v-theme-on-terminal));
}
</style>
