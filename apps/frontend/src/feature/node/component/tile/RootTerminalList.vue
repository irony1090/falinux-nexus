<script setup lang="ts">
import { useTileGrids } from '../../store/tileGrids.store';
import TerminalListRow from './TerminalListRow.vue';

const { terminals } = useTileGrids();
</script>

<template>
<div class="RootTerminalList">
    <terminal-list-row v-for="t in terminals" :key="t.id" :tile-id="t.id" />
    <p v-if="!terminals.length" class="empty">이 보드에 터미널이 없습니다</p>
</div>
</template>

<style scoped lang="scss">
.RootTerminalList {
    flex: 1;
    min-height: 0;
    overflow: auto;
    font-size: 12px;
}
.empty { margin: vars.$spacing-md; opacity: 0.6; }
</style>
