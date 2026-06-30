<template>
    <HelloWorld />
</template>

<script lang="ts" setup>
import { useTestSocket } from '@/common/websocket/websocket.hook';
import HelloWorld from '@/components/HelloWorld.vue'
import { useAuthStore } from '@/feature/user/store/auth.store';
import { watch } from 'vue';

const { auth } = useAuthStore();

const {
    connect,
    disconnect,
    call,
    emit,
    on,
    status
} = useTestSocket();

on('TTTT', d => {
    console.log('[TTTT]', d);
})

watch(auth, val => {
    console.log(val, status.value);
    if (!val){
        if (status.value === 'CONNECTED')
            disconnect();

    } else if (status.value === 'PENDING') {
        console.log('ATTEMP')
        connect();
    }
    

}, { immediate: true })

watch(status, async val => {
    console.log(val);
    if (val !== 'CONNECTED') return;
    // REQ → RES 왕복: 서버 subscribe.go 의 TEST 핸들러가 'RES' 를 돌려준다.
    const res = await call<string>('TEST', 'TEST_MESSAGE');
    console.log('TEST 응답:', res);
    emit('TEST_ON', 'TEST_ON_MESSAGE')
}, { immediate: true })

</script>
