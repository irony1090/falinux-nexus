<script setup lang="ts">
import { VIcon } from 'vuetify/components';

// 목록 격자 한 칸(판 D′) — 상자 + 위쪽 버튼 + 큰 아이콘 + 좌우 상태 라벨 + 이름. 누르기는 부모(itemPress.hook). REF-dir-grid.md
defineProps({
    icon: {
        type: String,
        required: true,
    },
    tone: {                 // 아이콘 색 = 테마 색 이름
        type: String,
        default: 'on-surface',
    },
    dim: Boolean,           // 일반 파일 아이콘은 옅게
    name: {
        type: String,
        required: true,
    },
    title: String,
    checking: Boolean,      // 선택 모드 = 왼쪽 위에 체크, 버튼 숨김
    checked: Boolean,
    focused: Boolean,
})
</script>

<template>
<div class="DirGridCell" :class="{ checked, focused }">
    <div class="box">
        <span v-if="checking" class="dg-btn dg-btn--l check" :class="{ on: checked }">
            <v-icon v-if="checked" size="13" icon="mdi-check-bold" /><i v-else />
        </span>
        <template v-else>
            <slot name="action" />
            <slot name="menu" />
        </template>
        <v-icon class="ico" :class="{ dim }" size="46" :icon="icon" :style="{ color: `rgb(var(--v-theme-${tone}))` }" />
        <span class="dg-r"><slot name="right" /></span>
        <span class="dg-l"><slot name="left" /></span>
        <span class="nm" :title="title ?? name">{{ name }}</span>
    </div>
</div>
</template>

<style scoped lang="scss">
.DirGridCell {
    --dg-box: color-mix(in srgb, rgb(var(--v-theme-on-surface)) 5%, rgb(var(--v-theme-surface)));
    position: relative;
    height: 112px;
    user-select: none;
    -webkit-touch-callout: none;    // 길게 누르기 = 체크 — iOS 기본 메뉴 끔
    touch-action: manipulation;     // 더블탭 확대 끔(더블탭 = 열기·실행)
    cursor: default;
    &.checked { --dg-box: color-mix(in srgb, rgb(var(--v-theme-primary)) 13%, rgb(var(--v-theme-surface))); }
}
.box {
    position: absolute;
    inset: 0;
    overflow: hidden;
    border-radius: 10px;
    background: var(--dg-box);
    box-shadow: inset 0 0 0 1px rgba(var(--v-border-color), var(--v-border-opacity));
    .checked & { box-shadow: inset 0 0 0 1px rgba(var(--v-theme-primary), 0.35); }
    .focused & { box-shadow: inset 0 0 0 2px rgba(var(--v-theme-on-surface), 0.55); }
    .focused.checked & { box-shadow: inset 0 0 0 2px rgb(var(--v-theme-primary)); }
}
.check {
    --c: var(--v-theme-on-surface);
    width: 30px;
    padding: 0 4px 0 0;
    i { width: 11px; height: 11px; box-sizing: border-box; border: 1.5px solid currentColor; border-radius: 3px; opacity: 0.6; }
    &.on { --c: var(--v-theme-primary); color: rgb(var(--v-theme-on-primary)); background: rgb(var(--v-theme-primary)); }
}
.ico {
    position: absolute;
    top: 21px;
    left: 50%;
    margin-left: -23px;
    &.dim { opacity: 0.6; }
}
// 오른쪽 = 아이콘 위 오른쪽 모서리 옆에서 아래로, 왼쪽 = 아이콘 아래 왼쪽 모서리에서 위로 쌓인다(이름을 침범하지 않게)
.dg-r, .dg-l {
    position: absolute;
    z-index: 1;
    display: flex;
    gap: 4px;
}
.dg-r { right: 0; top: 28px; flex-direction: column; align-items: flex-end; }
.dg-l { left: 0; bottom: 43px; flex-direction: column-reverse; align-items: flex-start; }
.nm {
    position: absolute;
    left: 6px;
    right: 6px;
    top: 74px;
    @include vars.font('body-small');
    line-height: 15px;
    text-align: center;
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow-wrap: anywhere;
}
</style>
