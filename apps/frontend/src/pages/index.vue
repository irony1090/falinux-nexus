<template>
<tile-workspace />
</template>

<script lang="ts" setup>
import TileWorkspace from '@/feature/node/component/TileWorkspace.vue';
import { provideTileTree } from '@/feature/node/store/tileTree.store';
import { provideProcessTerm } from '@/feature/process/store/processTerm.store';
import { useTestSocket } from '@/common/websocket/websocket.hook';
import { useAuthStore } from '@/feature/user/store/auth.store';
import { watch } from 'vue';

const { connect, disconnect } = useTestSocket();
const { auth } = useAuthStore();

provideTileTree();  // 루트 폴더 타일 1개로 시작 (서버 저장본 ⑪ 연동 전까지)
provideProcessTerm();

watch(auth, val => {
    if (val) {
        connect();
    } else {
        disconnect();
    }
}, { immediate: true })

</script>
