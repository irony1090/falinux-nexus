<script setup lang="ts">
import { VIcon } from 'vuetify/components';
import { useNoticeGo } from '../../hook/tileNotice.hook';
import { useTileNotices, type NoticeToast } from '../../store/tileNotice.store';

// 하단 구석 토스트 — N1(표시 수단). 하나면 누르면 그 위치로, 묶음이면 알림 내역을 연다(N5(알림 폭주))
const emit = defineEmits<{ (e: 'openHistory'): void }>();
const { toasts, dismiss } = useTileNotices();
const go = useNoticeGo();

const onClick = (t: NoticeToast) => {
    dismiss(t.id);
    if (t.items.length === 1) go(t.items[0]!);
    else emit('openHistory');
}
</script>

<template>
<div class="NoticeToasts">
    <div v-for="t in toasts" :key="t.id" class="toast" role="status" @click="onClick(t)">
        <v-icon size="16" icon="mdi-map-marker-radius-outline" color="primary" />
        <div class="body">
            <template v-if="t.items.length === 1">{{ t.items[0]!.text }}</template>
            <template v-else>
                <div class="title">위치 알림 {{ t.items.length }}개</div>
                <div v-for="it in t.items.slice(0, 3)" :key="it.id" class="line">{{ it.text }}</div>
            </template>
        </div>
        <button class="x" title="닫기" @click.stop="dismiss(t.id)"><v-icon size="14" icon="mdi-close" /></button>
    </div>
</div>
</template>

<style scoped lang="scss">
.NoticeToasts {
    position: fixed;
    right: vars.$spacing-md;
    bottom: calc(56px + vars.$spacing-md);  // 보드 푸터(폰 56) 위
    z-index: 2000;
    display: flex;
    flex-direction: column;
    gap: vars.$spacing-xs;
    max-width: min(360px, calc(100vw - 32px));
    pointer-events: none;
}
.toast {
    pointer-events: auto;
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto;
    align-items: start;
    gap: vars.$spacing-xs;
    padding: vars.$spacing-sm vars.$spacing-md;
    border-radius: 8px;
    background: rgb(var(--v-theme-surface));
    border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.18);
    @include vars.font('body-medium');
    cursor: pointer;
}
.body { min-width: 0; }
.title { font-weight: vars.$weight-lg; }
.line {
    opacity: 0.75;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.x { opacity: 0.6; }
</style>
