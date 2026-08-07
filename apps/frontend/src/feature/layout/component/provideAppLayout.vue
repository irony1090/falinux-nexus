<script setup lang="ts">
import { provideResizeGroupStore } from '@/feature/common/store/resizeGroup.store';
import { VApp } from 'vuetify/components';
const { flag } = provideResizeGroupStore();
flag.value = true;
</script>
<template>
<v-app>
    <slot />
</v-app>
</template>