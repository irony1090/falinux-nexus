<script setup lang="ts">
// 타일 공용 틀 — 칸 순서(겹침 / 머리 줄 / 탭 띠 / 본문 / 덧붙임)와 바탕 모양만. 스토어·동작은 쓰는 틀(TileFrame·MbTileFrame)이 쥔다 — REF-tile-frame.md
withDefaults(defineProps<{ tag?: string }>(), { tag: 'div' });
</script>

<template>
<component :is="tag" class="TileShell">
    <slot name="overlay" />
    <slot name="head" />
    <slot name="tabs" />
    <slot />
    <slot name="extra" />
</component>
</template>

<style scoped lang="scss">
// :where = 우선순위 0 — 쓰는 틀이 테두리 색·크기 등을 덮는다
:where(.TileShell) {
    min-width: 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    position: relative;
    border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    border-radius: 6px;
    background: rgb(var(--v-theme-surface));
}
</style>
