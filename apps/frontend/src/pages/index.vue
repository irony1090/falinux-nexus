<template>
    <v-btn @click="onExec">EXEC</v-btn>
    <v-btn @click="onView">VIEW</v-btn>
    <v-btn :disabled="!process" @click="onUnsubscribe">UNSUB</v-btn>
    <v-btn :disabled="!process" @click="onSubscribe">SUB</v-btn>
    <v-btn :disabled="!process" @click="onKill">KILL</v-btn>
    <HelloWorld />
</template>

<script lang="ts" setup>
import { VBtn } from 'vuetify/components'
import { useTestSocket } from '@/common/websocket/websocket.hook';
import HelloWorld from '@/components/HelloWorld.vue'
import { useGetNode } from '@/feature/node/api/node.api';
import { useAuthStore } from '@/feature/user/store/auth.store';
import { ref, watch } from 'vue';
import { execProcess, killProcess, listSubscriptions, subscribeProcess, unsubscribeProcess } from '@/feature/process/api/process.api';
import { useProcessDialog } from '@/feature/process/store/processDialog.store';

const { openProcessDialog, process } = useProcessDialog();

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
