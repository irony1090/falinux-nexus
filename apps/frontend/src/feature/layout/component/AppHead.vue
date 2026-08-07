<template>
<v-app-bar ref="vElRef" height="auto" elevation="2" >
    <template #prepend>
        <v-avatar v-ripple image="@/assets/logo.png" size="28" @click="moveIndex"/>
    </template>
    <v-app-bar-title class="text-primary">{{ auth?.nickname ?? '-' }}</v-app-bar-title>
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
import { useAuthStore } from '@/feature/user/store/auth.store';
import { computed } from 'vue';
import { useResizeCallback } from '@/feature/common/store/resizeGroup.store';
import { computeResizeSize } from '@/common/hook/vue.hook';

const router = useRouter();
// @ts-ignore
const { vElRef, size } = useAppHead()

const elRef = computed(() => vElRef.value?.$el as HTMLElement|undefined)
useResizeCallback(elRef, () => {
    size.value = computeResizeSize(elRef.value || null)
})
const { open } = useAppNav();
const { auth, logout } = useAuthStore();

const toggleNaiv = () => {
    // open.value = !open.value
    if (!auth.value) router.push('/login')
    else logout()
}

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