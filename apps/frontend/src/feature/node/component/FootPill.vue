<script setup lang="ts">
// 푸터 알약 하나(보드 메뉴·인디케이터·앱 버튼) — 높이는 BoardFooter의 --pill-h
</script>

<template>
<div class="FootPill elevation-2">
    <slot />
</div>
</template>

<style scoped lang="scss">
.FootPill {
    flex: none;
    display: flex;
    align-items: center;
    height: var(--pill-h, 34px);
    padding: 0 3px;
    box-sizing: border-box;
    border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    border-radius: 999px;
    background: rgb(var(--v-theme-surface));
}
</style>
