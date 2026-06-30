<template>
<v-app-bar ref="vElRef" height="auto" elevation="2" >
    <template #prepend>
        <v-avatar v-ripple image="@/assets/logo.png" size="28" @click="moveIndex"/>
    </template>
    <v-app-bar-title class="text-primary">{{ '흠~' }}</v-app-bar-title>
    <template #append>
        <v-app-bar-nav-icon @click="toggleNaiv"/>
    </template>
    <!-- <slot /> -->
</v-app-bar>
</template>
<script lang="ts" setup>
import { VAppBar, VAppBarNavIcon, VAvatar } from 'vuetify/components'
import { useAppHead } from '../store/appHead.store';
import { useAppNav } from '../store/appNav.store';
import { useRouter } from 'vue-router';

const router = useRouter();
// @ts-ignore
const { vElRef } = useAppHead()
const { open } = useAppNav();

const toggleNaiv = () => open.value = !open.value

const moveIndex = () => router.push('/')


// const title = computed(() => titles.value[0])
// const subtitle = computed(() => {
//     if (titles.value && titles.value.length > 1) {
//         console.log(titles.value)
//         return titles.value[titles.value.length - 1]
//     } else {
//         return undefined;
//     }
// })


</script>

<style scoped lang="scss">

.v-avatar {
    cursor: pointer;
}

.v-app-bar-title {
    font-weight: vars.$weight-lg;
    margin-inline-start: vars.$spacing-md;
}
</style>