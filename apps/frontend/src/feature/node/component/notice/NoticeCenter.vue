<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useAppHead } from '@/feature/layout/store/appHead.store';
import { VBadge, VBtn, VIcon, VNavigationDrawer } from 'vuetify/components';
import { useNoticeGo, useTileNoticeFeed } from '../../hook/tileNotice.hook';
import { useTileNotices, type NoticeItem } from '../../store/tileNotice.store';
import NoticeToasts from './NoticeToasts.vue';

// 위치 알림 받기 + 헤더 종(안 읽은 수) + 내역 드로어 + 토스트 — 개편-8(알림 UI)
useTileNoticeFeed();
const { history, unread, markRead, clear } = useTileNotices();
const go = useNoticeGo();

const open = ref(false);
// 헤더 높이가 auto라 Vuetify 레이아웃이 모름 — AppBody처럼 직접 내려 준다(안 하면 헤더가 드로어 윗부분을 덮음)
const { size: headSize } = useAppHead();
const drawerStyle = computed(() => {
    const h = headSize.value?.outer[1] ?? 0;
    return { top: `${h}px`, height: `calc(100% - ${h}px)` };
})
watch(open, v => { if (v) markRead() });

const timeText = (at: number) => new Date(at).toLocaleTimeString('ko-KR', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
const onItem = (it: NoticeItem) => {
    open.value = false;
    go(it);
}
</script>

<template>
<teleport defer to="#app-head-tools">
    <v-btn size="small" variant="text" icon title="위치 알림" @click="open = !open">
        <v-badge v-if="unread" :content="unread > 99 ? '99+' : unread" color="primary" floating>
            <v-icon icon="mdi-bell-outline" />
        </v-badge>
        <v-icon v-else icon="mdi-bell-outline" />
    </v-btn>
</teleport>
<v-navigation-drawer v-model="open" location="end" temporary width="320" :style="drawerStyle">
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
.empty { margin: vars.$spacing-md; opacity: 0.6; font-size: 12.5px; }
.item {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    gap: vars.$spacing-sm;
    width: 100%;
    text-align: left;
    padding: vars.$spacing-sm vars.$spacing-md;
    font-size: 12.5px;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    &:hover { background: rgba(var(--v-theme-on-surface), 0.04); }
}
.time { opacity: 0.6; font-family: monospace; }
.text { min-width: 0; }
</style>
