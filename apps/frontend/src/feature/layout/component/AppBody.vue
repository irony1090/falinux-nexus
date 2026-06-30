<template>
    <v-main :style="{'--v-layout-top': header}" ref="vRef">
        <slot />
    </v-main>
</template>
<script lang="ts" setup>
import { VMain } from 'vuetify/components'
import { computed, ref, watch } from 'vue';
import { useAppHead } from '../store/appHead.store';
import { provideStickyBoxRoot } from '@/feature/widget/store/stickyBox.store';
import { useAppWindow } from '../store/appWindown.store';

const { size: hSize } = useAppHead();
const { size: wSize, client } = useAppWindow();
const vRef = ref()

const headHeight = computed(() => hSize.value?.outer[1] ?? 0)
const header = computed(() => {
    return `${headHeight.value}px`;
})

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