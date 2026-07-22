<template>
  <v-app>
    <app-head />
    <app-body>
      <router-view />
    </app-body>

    <app-dialog />
  </v-app>
</template>

<script lang="ts" setup>
import AppHead from '@/feature/layout/component/AppHead.vue';
import AppBody from '@/feature/layout/component/AppBody.vue';
import AppDialog from '@/feature/layout/component/AppDialog.vue';
import { provideAppHead } from '@/feature/layout/store/appHead.store';
import { provideAppNav } from '@/feature/layout/store/appNav.store';
import { provideAppWindow } from '@/feature/layout/store/appWindown.store';
import { provideAppDialog } from '@/feature/layout/store/appDialog.store';
import { provideAuthStore } from '@/feature/user/store/auth.store';
import { provideTestSocket } from '@/common/websocket/websocket.hook';

provideAppWindow()
provideAppHead()
provideAppNav()

provideAuthStore();

provideAppDialog();

provideTestSocket();

</script>
