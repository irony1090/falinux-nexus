<script setup lang="ts">
import { computed } from 'vue';
import { dummyStatus } from '../../dev/tileDummy';
import { useTileTree } from '../../store/tileTree.store';

// 자리 표시 — ② ProcessDialog(xterm) 타일 임베드 전까지 출력 없음
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
const done = computed(() => !!tile.value && dummyStatus(tile.value.uid) === 'DONE');
</script>

<template>
<pre class="TerminalTileBody bg-terminal"><span class="dim">{{ tile?.uid }}</span>
<span class="dim">(출력은 xterm 임베드 후 표시)</span>
<span v-if="done" class="dim">[process exited]</span><span v-else class="cursor" /></pre>
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
