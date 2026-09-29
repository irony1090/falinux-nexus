<template>
<v-app-bar ref="vElRef" height="auto" elevation="2" >
    <template #prepend>
        <v-avatar v-ripple image="@/assets/logo.png" size="28" @click="moveIndex"/>
    </template>
    <v-app-bar-title class="text-primary">{{ auth?.nickname ?? '-' }}</v-app-bar-title>
    <!-- 페이지가 Teleport로 자기 내비를 넣는 자리 (예: TileWorkspace) -->
    <div id="app-head-nav" class="AppHead__nav" />
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
    flex: 0 1 auto;     // 기본 flex:1 이면 내비 자리와 폭을 나눠 가짐
    font-weight: vars.$weight-lg;
    margin-inline-start: vars.$spacing-md;
}
.AppHead__nav {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    padding-inline: vars.$spacing-md;
}
</style>