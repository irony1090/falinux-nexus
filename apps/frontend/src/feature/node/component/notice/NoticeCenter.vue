<script setup lang="ts">
import { watch } from 'vue';
import { VBtn, VNavigationDrawer } from 'vuetify/components';
import { useNoticeGo, useTileNoticeFeed } from '../../hook/tileNotice.hook';
import { useTileNotices, type NoticeItem } from '../../store/tileNotice.store';
import NoticeToasts from './NoticeToasts.vue';

// 위치 알림 받기 + 내역 드로어 + 토스트 — 개편-8(알림 UI). 종(안 읽은 수)은 BoardFooter
useTileNoticeFeed();
const { history, drawerOpen: open, markRead, clear } = useTileNotices();
const go = useNoticeGo();

watch(open, v => { if (v) markRead() });

const timeText = (at: number) => new Date(at).toLocaleTimeString('ko-KR', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
const onItem = (it: NoticeItem) => {
    open.value = false;
    go(it);
}
</script>

<template>
<v-navigation-drawer v-model="open" location="end" temporary width="320">
    <div class="head">
        <span>위치 알림</span>
        <v-btn size="x-small" variant="text" :disabled="!history.length" @click="clear">비우기</v-btn>
    </div>
    <p v-if="!history.length" class="empty">알림이 없습니다. 새로고침하면 비워집니다.</p>
    <button v-for="it in history" :key="it.id" class="item" @click="onItem(it)">
        <span class="time">{{ timeText(it.at) }}</span>
        <span class="text">{{ it.text }}</span>
    </button>
</v-navigation-drawer>
<notice-toasts @open-history="open = true" />
</template>

<style scoped lang="scss">
.head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: vars.$spacing-sm vars.$spacing-md;
    font-weight: vars.$weight-lg;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}
.empty { margin: vars.$spacing-md; opacity: 0.6; @include vars.font('body-medium'); }
.item {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    gap: vars.$spacing-sm;
    width: 100%;
    text-align: left;
    padding: vars.$spacing-sm vars.$spacing-md;
    @include vars.font('body-medium');
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    &:hover { background: rgba(var(--v-theme-on-surface), 0.04); }
}
.time { opacity: 0.6; font-family: monospace; }
.text { min-width: 0; }
</style>
