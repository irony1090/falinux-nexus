<template>
<provide-app-layout>
    <app-body>
      <router-view />
    </app-body>

    <app-dialog />
</provide-app-layout>
</template>

<script lang="ts" setup>
import ProvideAppLayout from '@/feature/layout/component/provideAppLayout.vue';
import AppBody from '@/feature/layout/component/AppBody.vue';
import AppDialog from '@/feature/layout/component/AppDialog.vue';
import { provideAppNav } from '@/feature/layout/store/appNav.store';
import { provideAppWindow } from '@/feature/layout/store/appWindown.store';
import { provideAppDialog } from '@/feature/layout/store/appDialog.store';
import { provideAuthStore } from '@/feature/user/store/auth.store';
import { provideTestSocket } from '@/common/websocket/websocket.hook';
import { bindTabId } from '@/common/util/tabId.util';
import { bindDeviceChanges } from '@/feature/worker/hook/deviceChange.hook';
import { watch } from 'vue';

provideAppWindow()
provideAppNav()

const { auth } = provideAuthStore();

provideAppDialog();

const socket = provideTestSocket();
bindTabId(socket);
bindDeviceChanges(socket);
// 로그인 중이면 어느 페이지든 소켓 연결 — 내 장비·관리자 페이지도 장비 상태 push를 받게(REF-device-push.md)
watch(auth, val => val ? socket.connect() : socket.disconnect(), { immediate: true });

</script>

<style>
/* 문서 당김·튕김 막기(iOS는 가상 키보드가 뜰 때 문서를 밀어 올림) — REF-mobile-keyboard.md 키보드-4 */
html, body { overscroll-behavior: none; }
</style>
