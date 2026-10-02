<template>
<tile-workspace v-if="ready" />
<div v-else-if="loadError" class="state">
    <span>{{ loadError }}</span>
    <v-btn size="small" variant="tonal" @click="load">다시 시도</v-btn>
</div>
<div v-else-if="auth" class="state">
    <v-progress-circular indeterminate size="24" color="primary" />
</div>
</template>

<script lang="ts" setup>
import TileWorkspace from '@/feature/node/component/TileWorkspace.vue';
import { provideTileTree } from '@/feature/node/store/tileTree.store';
import { useTerminalRestore } from '@/feature/node/hook/terminalRestore.hook';
import { provideProcessTerm } from '@/feature/process/store/processTerm.store';
import { useTestSocket } from '@/common/websocket/websocket.hook';
import { useAuthStore } from '@/feature/user/store/auth.store';
import { watch } from 'vue';
import { VBtn, VProgressCircular } from 'vuetify/components';

const { connect, disconnect } = useTestSocket();
const { auth } = useAuthStore();

const tileTree = provideTileTree();  // 로그인 뒤 서버 저장본(⑪)을 불러온다
const processTerm = provideProcessTerm();
const { ready, loadError, load, reset } = tileTree;
const { reset: resetTerms } = processTerm;
useTerminalRestore(tileTree, processTerm);

watch(auth, val => {
    if (val) {
        connect();
    } else {
        disconnect();
    }
}, { immediate: true })

// 계정이 바뀔 때만 다시 불러온다 — 같은 계정의 세션 재확인(auth 객체 교체)에 미확정 변경을 버리지 않도록.
// 이전 계정의 xterm도 비운다 — 3-d④(계정 바뀜)
watch(() => auth.value?.identification, id => {
    resetTerms();
    if (id) load();
    else reset();
}, { immediate: true })

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
