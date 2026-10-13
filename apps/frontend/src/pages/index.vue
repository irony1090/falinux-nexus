<template>
<tile-workspace v-if="ready" />
<div v-else-if="loadError" class="state">
    <span>{{ loadError }}</span>
    <v-btn size="small" variant="tonal" @click="load(hash)">다시 시도</v-btn>
    <v-btn v-if="hash" size="small" variant="text" to="/">기본 보드로</v-btn>
</div>
<div v-else-if="auth" class="state">
    <v-progress-circular indeterminate size="24" color="primary" />
</div>
<teleport v-if="auth" defer to="#app-head-title">
    <board-menu />
</teleport>
<notice-center v-if="auth" />
</template>

<script lang="ts" setup>
import TileWorkspace from '@/feature/node/component/TileWorkspace.vue';
import BoardMenu from '@/feature/node/component/BoardMenu.vue';
import NoticeCenter from '@/feature/node/component/notice/NoticeCenter.vue';
import { useTileNotices } from '@/feature/node/store/tileNotice.store';
import { provideTileTree } from '@/feature/node/store/tileTree.store';
import { provideBoards } from '@/feature/node/store/boards.store';
import { provideRemoteActive } from '@/feature/node/store/remoteActive.store';
import { useTerminalRestore } from '@/feature/node/hook/terminalRestore.hook';
import { useShareChange } from '@/feature/node/hook/shareChange.hook';
import { useNodeChange } from '@/feature/node/hook/nodeChange.hook';
import { provideProcessTerm } from '@/feature/process/store/processTerm.store';
import { useAppDialog } from '@/feature/layout/store/appDialog.store';
import { useTestSocket } from '@/common/websocket/websocket.hook';
import { tabReady } from '@/common/util/tabId.util';
import { useAuthStore } from '@/feature/user/store/auth.store';
import { computed, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { VBtn, VProgressCircular } from 'vuetify/components';

const { connect, disconnect, emit } = useTestSocket();
const { auth } = useAuthStore();
const { openDialog } = useAppDialog();
const route = useRoute();
const router = useRouter();

// 보는 보드: `/` = 기본(빈 값), `/w/{hash}` — `/`와 `/w/:hash`가 같은 컴포넌트라 인스턴스가 재사용될 수 있어 라우트를 지켜본다
const hash = computed(() => typeof route.params.hash === 'string' ? route.params.hash : '');

const tileTree = provideTileTree();  // 로그인 뒤 서버 저장본(⑪)을 불러온다
const processTerm = provideProcessTerm();
const boards = provideBoards();
provideRemoteActive();
const { ready, loadError, load, reset } = tileTree;
const { reset: resetTerms } = processTerm;
const { clear: clearNotices } = useTileNotices();
useTerminalRestore(tileTree, processTerm);
useShareChange();
useNodeChange();

watch(auth, val => {
    if (val) {
        connect();
    } else {
        disconnect();
    }
}, { immediate: true })

// 계정이 바뀔 때만 xterm을 비운다 — 3-d④(계정 바뀜). 같은 계정의 세션 재확인(auth 객체 교체)에는 미확정 변경을 버리지 않도록 다시 불러오지 않는다
watch(() => auth.value?.identification, id => {
    resetTerms();
    clearNotices();     // 위치 알림 내역은 계정마다
    if (id) {
        load(hash.value);
        boards.load();
    } else {
        reset();
        boards.reset();
    }
}, { immediate: true })

// 보드 전환 = 트리만 다시 불러옴. 다른 보드의 xterm은 남겨 둔다(돌아오면 그대로)
watch(hash, h => {
    if (auth.value) load(h);
})

// 보는 보드를 서버에 알린다(크기 우선권 후보 — WS7(보는 작업 공간 전달)). 소켓이 다시 붙으면 TAB:ID 뒤에 다시
watch([tabReady, hash], ([r, h]) => {
    if (r) emit('TAB:VIEW', { hash: h });
}, { immediate: true })

// 보던 보드가 목록에서 사라짐(다른 곳에서 삭제) -> 기본 보드로(안의 타일은 서버가 기본 보드 끝으로 옮김, W2(삭제 처리)).
// "있다가 없어짐"만 본다 — 막 만든 보드는 목록 이벤트보다 이동이 먼저일 수 있고, 없는 hash는 트리 404로 안내된다
// 로그아웃 때 목록을 비우는 것(reset, loaded = false)은 삭제가 아니다
watch(() => [hash.value, !!boards.byHash(hash.value), boards.loaded.value] as const, ([h, has, loaded], [prevH, had]) => {
    if (!h || h !== prevH || !had || has || !loaded) return;
    const removedHere = boards.removedHere(h);
    router.replace('/');
    if (!removedHere) openDialog({ type: 'info', content: '보던 보드가 삭제되어 기본 보드로 이동했습니다. 안의 타일은 기본 보드 끝으로 옮겨졌습니다.' });
})
</script>

<style scoped lang="scss">
.state {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: vars.$spacing-md;
    padding: vars.$spacing-md;
    min-height: 120px;
}
</style>
