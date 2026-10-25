<script setup lang="ts">
// 푸터 알약 하나(보드 메뉴·인디케이터·앱 버튼) — 높이는 BoardFooter의 --pill-h
defineProps({
    edge: {             // 푸터 왼쪽·오른쪽 벽에 닿는 알약 — 찰떡(REF-design-chaltteok.md)
        type: String as () => 'left' | 'right',
        default: undefined,
    },
})
</script>

<template>
<div class="FootPill elevation-2" :class="edge && `edge-${edge}`">
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
// 푸터 상자는 안 보이므로 벽 쪽 모양(위 직각·아래 12px)을 알약이 직접 그린다. 안쪽 버튼은 칸으로 꽉 채운다
.edge-left,
.edge-right {
    align-items: stretch;
    padding: 0;
    overflow: hidden;
}
.edge-left { border-radius: 0 999px 999px 12px; }
.edge-right {
    border-radius: 999px 0 12px 999px;
    > :deep(.v-btn) { width: 38px; height: auto; border-radius: 0; }
    > :deep(.v-btn:first-child) { width: 44px; padding-left: 6px; }
    > :deep(.v-btn:first-child:last-child) { width: 52px; }
    > :deep(.v-btn + .v-btn) { box-shadow: inset 1px 0 0 rgba(var(--v-border-color), var(--v-border-opacity)); }
}
</style>
