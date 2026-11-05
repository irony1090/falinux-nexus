<script setup lang="ts">
import { useAppDialog } from '@/feature/layout/store/appDialog.store';
import { PERM_ADMIN, PERM_WORKER_MOVE } from '@/feature/user/api/user.api';
import { useAuthStore } from '@/feature/user/store/auth.store';
import SupervisorMoveDialog from '@/feature/worker/component/SupervisorMoveDialog.vue';
import logoUrl from '@/assets/logo.svg';
import { computed, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { VBtn, VDivider, VIcon, VList, VListItem, VListSubheader, VMenu } from 'vuetify/components';
import type { BoardItem } from '../api/board.api';
import { useBoardClose } from '../hook/boardClose.hook';
import { boardLabel, boardPath, useBoards } from '../store/boards.store';
import MySharesDialog from './dialog/MySharesDialog.vue';
import NodeNameDialog from './dialog/NodeNameDialog.vue';

// 푸터 왼쪽: 로고 + 지금 보드 이름 + 보드 전환·이름 변경·삭제, 위로 열림. compact(폰) = 로고만 — REF-board-chrome.md
defineProps<{ compact?: boolean }>();
const route = useRoute();
const router = useRouter();
const { list, byHash, defaultBoard, rename } = useBoards();
const { closing, confirmClose } = useBoardClose();
const { auth } = useAuthStore();
const { openDialog } = useAppDialog();

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
// 전체 worker 주소 변경 — worker_move 권한 계정만 보임(주소-Q20(전체 명령 권한))
const moveAllOpen = ref(false);
const canMoveAll = computed(() => !!auth.value?.perms?.includes(PERM_WORKER_MOVE));
const isAdmin = computed(() => !!auth.value?.perms?.includes(PERM_ADMIN));
const onSubmitName = (name: string) => {
    if (!current.value) return;
    run(rename(current.value.hash, name), () => nameOpen.value = false);
}

// 보드 삭제 = 보드 닫기 확인(기본 보드로 보내기 / 모두 종료하고 닫기) — REF-board-close.md
const onRemove = () => confirmClose();
</script>

<template>
<v-menu location="top start" offset="10">
    <template #activator="{ props: act }">
        <button v-bind="act" class="BoardMenu" :class="{ compact }" :disabled="busy || closing" :title="`보드: ${title}`" :aria-label="`보드 메뉴 - ${title}`">
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
        <v-list-item title="내 장비" prepend-icon="mdi-chip" to="/devices" />
        <v-list-item v-if="isAdmin" title="관리자" prepend-icon="mdi-shield-account-outline" to="/admin" />
        <v-list-item v-if="isAdmin" title="Updater" prepend-icon="mdi-package-up" to="/updater" />
        <v-list-item v-if="canMoveAll" title="전체 worker 주소 변경" prepend-icon="mdi-server-network" @click="moveAllOpen = true" />
        <v-divider />
        <v-list-item :subtitle="`계정 ${auth?.nickname ?? '-'}`" disabled />
    </v-list>
</v-menu>
<my-shares-dialog v-model="mineOpen" />
<supervisor-move-dialog v-if="canMoveAll" v-model="moveAllOpen" />
<node-name-dialog v-model="nameOpen" title="보드 이름 변경" :initial="current?.name ?? ''" :loading="busy" @submit="onSubmitName" />
</template>

<style scoped lang="scss">
.BoardMenu {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    max-width: 240px;
    padding: 0 12px 0 8px;
    border: none;
    border-radius: 0;       // 알약(FootPill edge)이 모양을 자른다
    background: none;
    color: inherit;
    font: inherit;
    font-weight: vars.$weight-lg;
    cursor: pointer;
    &:hover { background: rgba(var(--v-theme-on-surface), 0.06); }
    &.compact { gap: 4px; padding: 0 12px 0 8px; }
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
