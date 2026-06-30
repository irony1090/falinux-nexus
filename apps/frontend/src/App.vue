<template>
  <v-app>
    <app-head />
    <app-body>
      <router-view />
    </app-body>
  </v-app>
</template>

<script lang="ts" setup>
import AppHead from '@/feature/layout/component/AppHead.vue';
import AppBody from '@/feature/layout/component/AppBody.vue';
import { provideAppHead } from '@/feature/layout/store/appHead.store';
import { provideAppNav } from '@/feature/layout/store/appNav.store';
import { provideAppWindow } from '@/feature/layout/store/appWindown.store';

provideAppWindow()
provideAppHead()
provideAppNav()
</script>
