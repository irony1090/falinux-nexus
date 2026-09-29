<template>
<tile-workspace />
</template>

<script lang="ts" setup>
import TileWorkspace from '@/feature/node/component/TileWorkspace.vue';
import { provideTileTree } from '@/feature/node/store/tileTree.store';
import { seedTileTree } from '@/feature/node/dev/tileDummy';
import { useTestSocket } from '@/common/websocket/websocket.hook';
import { useGetNode } from '@/feature/node/api/node.api';
import { useAuthStore } from '@/feature/user/store/auth.store';
import { ref, watch } from 'vue';
import { execProcess, killProcess, listSubscriptions, subscribeProcess, unsubscribeProcess } from '@/feature/process/api/process.api';
import { useProcessDialog } from '@/feature/process/store/processDialog.store';
import { useAppWindow } from '@/feature/layout/store/appWindown.store';

const { openProcessDialog, process } = useProcessDialog();

const { size } = useAppWindow();

const { connect, disconnect, status, on } = useTestSocket();
const { auth } = useAuthStore();
const nodeId = ref(3);
// const processId = ref<string>();
const { data } = useGetNode(nodeId, )

const onSubscribe = () => {
    if (!process.value) return;
    subscribeProcess(process.value.uid)
    .then(res => {
        console.log('[SUCCESS]',res);
    }).catch(err => {
        console.log('[ERR]', err);
    })
}
const onUnsubscribe = () => {
    if (!process.value) return;
    unsubscribeProcess(process.value.uid)
    .then(res => {
        console.log('[SUCCESS]',res);
    }).catch(err => {
        console.log('[ERR]', err);
    })
}

const onKill = () => {
    if (!process.value) return;
    killProcess(process.value.uid)
    .then(res => {
        console.log('[SUCCESS]',res);
    }).catch(err => {
        console.log('[ERR]', err);
    })
}

const onView = () => {
    listSubscriptions()
    .then(res => {
        console.log('[SUCCESS]',res);
    }).catch(err => {
        console.log('[ERR]', err);
    })
}

const onExec = () => {
    execProcess({
        authKey: 'irony-MAC-ADDress1#UhQ2l5hG',
        nodeId: nodeId.value,
    }).then(res => {
        openProcessDialog(res)
    }).catch(err => {
        console.log('[ERR]', err);
    })
}

const tileTree = provideTileTree();
seedTileTree(tileTree);     // 미리보기용 더미 트리 (실제 저장본 연동 전까지)


on('NODE:UPDATE', val => {
    console.log(val);
})

watch(data, val => {

    console.log('[NODE] DATA',val)
}, { immediate: true })

watch(status, val => {
    console.log('[SOCKET] STATUS - ', val)
}, { immediate: true })

watch(auth, val => {
    if (val) {
        connect();
    } else {
        disconnect();
    }
}, { immediate: true })

</script>
