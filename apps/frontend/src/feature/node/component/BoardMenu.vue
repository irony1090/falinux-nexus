<script setup lang="ts">
import { useAppDialog } from '@/feature/layout/store/appDialog.store';
import { useAuthStore } from '@/feature/user/store/auth.store';
import logoUrl from '@/assets/logo.svg';
import { computed, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { VBtn, VDivider, VIcon, VList, VListItem, VListSubheader, VMenu } from 'vuetify/components';
import type { BoardItem } from '../api/board.api';
import { boardLabel, boardPath, useBoards } from '../store/boards.store';
import MySharesDialog from './dialog/MySharesDialog.vue';
import NodeNameDialog from './dialog/NodeNameDialog.vue';

// 푸터 왼쪽: 로고 + 지금 보드 이름 + 보드 전환·이름 변경·삭제, 위로 열림. compact(폰) = 로고만 — REF-board-chrome.md
defineProps<{ compact?: boolean }>();
const route = useRoute();
const router = useRouter();
const { list, byHash, defaultBoard, rename, remove } = useBoards();
const { auth } = useAuthStore();
const { open: appDialogOpen, openDialog } = useAppDialog();

const hash = computed(() => typeof route.params.hash === 'string' ? route.params.hash : '');
const current = computed(() => hash.value ? byHash(hash.value) : defaultBoard.value);
const title = computed(() => current.value ? boardLabel(current.value) : hash.value ? '보드' : '기본 보드');

const busy = ref(false);
const run = (job: Promise<unknown>, done?: () => void) => {
    busy.value = true;
    job.then(() => done?.())
    .catch(err => openDialog({ type: 'error', content: err?.message || '요청에 실패했습니다' }))
    .finally(() => busy.value = false);
}

const go = (b: BoardItem) => router.push(boardPath(b));

const nameOpen = ref(false);
const mineOpen = ref(false);
const onSubmitName = (name: string) => {
    if (!current.value) return;
    run(rename(current.value.hash, name), () => nameOpen.value = false);
}

// 이름은 title로만 넘긴다 — AppDialog의 content는 v-html로 그려진다
const onRemove = () => {
    const b = current.value;
    if (!b || b.isDefault) return;
    openDialog({
        type: 'warning',
        title: `'${boardLabel(b)}' 삭제`,
        content: ['보드 안의 폴더·터미널 타일은 기본 보드 끝으로 옮겨집니다.', '실행 중인 터미널은 종료되지 않습니다.'],
        buttons: [
            { view: '취소', type: 'info', bindButton: 'Escape', action: () => appDialogOpen.value = false },
            {
                view: '삭제', type: 'error',
                action: () => {
                    appDialogOpen.value = false;
                    run(remove(b.hash), () => router.replace('/'));
                },
            },
        ],
    })
}
</script>

<template>
<v-menu location="top start" offset="10">
    <template #activator="{ props: act }">
        <button v-bind="act" class="BoardMenu" :class="{ compact }" :disabled="busy" :title="`보드: ${title}`" :aria-label="`보드 메뉴 - ${title}`">
            <img class="logo" :src="logoUrl" alt="" />
            <span v-if="!compact" class="name">{{ title }}</span>
            <v-icon size="16" icon="mdi-chevron-up" />
        </button>
    </template>
    <v-list density="compact" min-width="200">
        <v-list-subheader>보드</v-list-subheader>
        <v-list-item v-for="b in list" :key="b.hash"
            :title="boardLabel(b)"
            :prepend-icon="b.isDefault ? 'mdi-home-outline' : 'mdi-view-dashboard-outline'"
            :active="b.hash === current?.hash"
            color="primary"
            @click="go(b)"
        />
        <v-divider />
        <v-list-item title="이름 변경" prepend-icon="mdi-pencil-outline" :disabled="!current" @click="nameOpen = true" />
        <v-list-item v-if="current && !current.isDefault" title="보드 삭제" prepend-icon="mdi-delete-outline" base-color="error" @click="onRemove" />
        <v-divider />
        <v-list-item title="내가 공유한 것" prepend-icon="mdi-account-multiple-outline" @click="mineOpen = true" />
        <v-divider />
        <v-list-item :subtitle="`계정 ${auth?.nickname ?? '-'}`" disabled />
    </v-list>
</v-menu>
<my-shares-dialog v-model="mineOpen" />
<node-name-dialog v-model="nameOpen" title="보드 이름 변경" :initial="current?.name ?? ''" :loading="busy" @submit="onSubmitName" />
</template>

<style scoped lang="scss">
.BoardMenu {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    max-width: 240px;
    height: calc(var(--pill-h, 34px) - 6px);
    padding: 0 8px 0 2px;
    border: none;
    border-radius: 999px;
    background: none;
    color: inherit;
    font: inherit;
    font-weight: vars.$weight-lg;
    cursor: pointer;
    &:hover { background: rgba(var(--v-theme-on-surface), 0.06); }
    &.compact { gap: 4px; padding: 0 6px 0 4px; }
}
.logo {
    flex: none;
    width: 24px;
    height: 24px;
    border-radius: 7px;
}
.name {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
</style>
