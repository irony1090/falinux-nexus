<script setup lang="ts">
import { cellChip } from '../hook/cellDrag.hook';
import { useTileDrag } from '../store/tileDrag.store';

// 끄는 동안 손가락·커서 옆 작은 표시(아이콘·이름) — 안 1(작은 표시 + 빈 자리), REF-tile-drag.md
// 다른 보드 줄 위면 같이 가는 타일도 아래에 붙인다(보드Q6(묶음 표시 때)) — 3개까지, 나머지는 개수
const SHOW_MAX = 3;
// 폴더 칸 끌기(칸-2(node 칸 끌기))도 같은 표시 — 여러 칸이면 "외 N개", 넣기 대상 위면 "→ 폴더 이름"
const { chip, bundleIds, bundleChips } = useTileDrag();
</script>

<template>
<teleport to="body">
    <div v-if="chip" class="TileDragChip" :style="{ left: `${chip.x}px`, top: `${chip.y}px` }" aria-hidden="true">
        <div class="row">
            <v-icon size="16" :icon="chip.icon" :color="chip.color" />
            <span class="name">{{ chip.name }}</span>
        </div>
        <template v-if="bundleIds.length">
            <div v-for="(b, i) in bundleChips.slice(0, SHOW_MAX)" :key="i" class="row along">
                <v-icon size="14" :icon="b.icon" :color="b.color" />
                <span class="name">{{ b.name }}</span>
            </div>
            <div v-if="bundleChips.length > SHOW_MAX" class="row along more">외 {{ bundleChips.length - SHOW_MAX }}개</div>
        </template>
    </div>
    <div v-else-if="cellChip" class="TileDragChip" :style="{ left: `${cellChip.x}px`, top: `${cellChip.y}px` }" aria-hidden="true">
        <div class="row">
            <v-icon size="16" :icon="cellChip.icon" :color="cellChip.color" />
            <span class="name">{{ cellChip.name }}</span>
            <span v-if="cellChip.more" class="more">외 {{ cellChip.more }}개</span>
        </div>
        <div v-if="cellChip.to" class="row to">{{ cellChip.to }}</div>
    </div>
</teleport>
</template>

<style scoped lang="scss">
.TileDragChip {
    position: fixed;
    z-index: 2100;
    max-width: 220px;
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 4px 12px 4px 8px;
    border: 1px solid rgb(var(--v-theme-primary));
    border-radius: 14px;
    background: rgb(var(--v-theme-surface));
    box-shadow: 0 6px 18px rgba(0, 0, 0, 0.18);
    @include vars.font('body-medium');
    pointer-events: none;
    // 손가락에 가리지 않게 오른쪽 위로
    transform: translate(14px, -120%);
}
.row {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
}
// 같이 가는 타일 — 끄는 타일보다 작게, 들여서
.along {
    padding-left: 8px;
    @include vars.font('body-small');
    color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity));
}
.more { flex: none; color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity)); }
.to {
    @include vars.font('body-small');
    font-weight: vars.$weight-md;
    color: rgb(var(--v-theme-primary));
}
.name {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
</style>

<style lang="scss">
// 끄는 동안 = 커서 고정·글자 선택 막음
html.tile-dragging, html.tile-dragging * {
    cursor: grabbing !important;
    user-select: none !important;
    -webkit-user-select: none !important;
}
</style>
