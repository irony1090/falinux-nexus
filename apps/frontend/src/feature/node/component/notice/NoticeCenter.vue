<script setup lang="ts">
import { watch } from 'vue';
import { VNavigationDrawer } from 'vuetify/components';
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
    <div class="head tile-btn-row-end">
        <span class="ttl">위치 알림</span>
        <div class="tile-btn-end"><button class="tile-btn tile-btn--text" :disabled="!history.length" @click="clear">비우기</button></div>
    </div>
    <p v-if="!history.length" class="empty">알림이 없습니다. 새로고침하면 비워집니다.</p>
    <button v-for="it in history" :key="it.id" class="item" @click="onItem(it)">
        <span class="ct-grp ct-grp--l ct-grp--sm"><span class="ct-seg ct-seg--txt time">{{ timeText(it.at) }}</span></span>
        <span class="text">{{ it.text }}</span>
    </button>
</v-navigation-drawer>
<notice-toasts @open-history="open = true" />
</template>

<style scoped lang="scss">
// 찰떡 — 비우기 = 오른쪽 벽 끝 블록, 시각 = 왼쪽 벽 라벨(REF-design-chaltteok.md)
.head {
    display: flex;
    align-items: stretch;
    height: 40px;
    padding-left: vars.$spacing-lg;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}
.ttl { flex: 1 1 auto; align-self: center; font-weight: vars.$weight-lg; }
.empty { margin: vars.$spacing-md; opacity: 0.6; @include vars.font('body-medium'); }
.item {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    align-items: start;
    gap: vars.$spacing-md;
    width: 100%;
    text-align: left;
    padding: 6px vars.$spacing-md 6px 0;
    @include vars.font('body-medium');
    color: inherit;
    background: none;   // 브라우저 기본 버튼 바탕·테두리 지움
    border: 0;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    &:hover { background: rgba(var(--v-theme-on-surface), 0.04); }
}
.item > .ct-grp { align-self: start; }
.time { font-family: monospace; }
.text { min-width: 0; padding-top: 2px; }
</style>
