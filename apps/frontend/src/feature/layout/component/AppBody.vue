<template>
    <v-main ref="vRef">
        <slot />
    </v-main>
</template>
<script lang="ts" setup>
import { VMain } from 'vuetify/components'
import { ref, watch } from 'vue';
import { provideStickyBoxRoot } from '@/feature/widget/store/stickyBox.store';
import { useAppWindow } from '../store/appWindown.store';

const { size: wSize, client } = useAppWindow();
const vRef = ref()

const { rootClient, maxClient } = provideStickyBoxRoot();

watch(vRef, val => {
    client.value = val?.$el
})

watch(() => wSize.value?.padding, (pd) => {
    const t = pd?.[0] ?? 0;
    const r = pd?.[1] ?? 0;
    const b = pd?.[2] ?? 0;
    const l = pd?.[3] ?? 0;
    rootClient.value = [
        t,      // TOP 
        r,      // RIGHT
        b,      // BOTTOM
        l,      // LEFT
    ]
    maxClient.value = [
        t,      // TOP 
        r,      // RIGHT
        b,      // BOTTOM
        l,      // LEFT
    ]
})
</script>
<style lang="scss" scoped>
</style>