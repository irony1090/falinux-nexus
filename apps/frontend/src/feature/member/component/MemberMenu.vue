<script setup lang="ts">
import logoUrl from '@/assets/logo.svg';
import { PERM_ADMIN } from '@/feature/user/api/user.api';
import { useAuthStore } from '@/feature/user/store/auth.store';
import { goLastBoard } from '@/router/lastBoard';
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { VIcon, VList, VListItem, VMenu } from 'vuetify/components';
import { useMemberBoard } from '../store/memberBoard.store';

// 푸터 왼쪽 로고 메뉴(보드 BoardMenu와 같은 모양) — 보드로·내 장비·관리자, 위로 열림. compact(폰) = 로고만 — 화-13(전체 헤더)
// 관리 페이지끼리는 replace — 뒤로 가기 한 번이면 보드로(주소-1(진입만 push))
defineProps<{ compact?: boolean }>();
const { page } = useMemberBoard();
const { auth } = useAuthStore();
const router = useRouter();
const isAdmin = computed(() => !!auth.value?.perms?.includes(PERM_ADMIN));
const title = computed(() => page === 'admin' ? '관리자' : '내 장비');
</script>

<template>
<v-menu location="top start" offset="10">
    <template #activator="{ props: act }">
        <button v-bind="act" class="MemberMenu" :class="{ compact }" :title="title" :aria-label="`메뉴 - ${title}`">
            <img class="logo" :src="logoUrl" alt="" />
            <span v-if="!compact" class="name">{{ title }}</span>
            <v-icon size="16" icon="mdi-chevron-up" />
        </button>
    </template>
    <v-list density="compact" min-width="200">
        <v-list-item title="보드로" prepend-icon="mdi-view-dashboard-outline" @click="goLastBoard(router)" />
        <v-list-item title="내 장비" prepend-icon="mdi-chip" to="/devices" replace :active="page === 'mine'" color="primary" />
        <v-list-item v-if="isAdmin" title="관리자" prepend-icon="mdi-shield-account-outline" to="/admin" replace :active="page === 'admin'" color="primary" />
    </v-list>
</v-menu>
</template>

<style scoped lang="scss">
.MemberMenu {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    padding: 0 12px 0 8px;
    border: none;
    border-radius: 0;       // 알약(FootPill edge)이 모양을 자른다
    background: none;
    color: inherit;
    font: inherit;
    font-weight: vars.$weight-lg;
    cursor: pointer;
    &:hover { background: rgba(var(--v-theme-on-surface), 0.06); }
    &.compact { gap: 4px; }
}
.logo {
    flex: none;
    width: 24px;
    height: 24px;
    border-radius: 7px;
}
.name { white-space: nowrap; }
</style>
