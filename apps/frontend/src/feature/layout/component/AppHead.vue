<template>
<v-app-bar ref="vElRef" height="auto" elevation="2" >
    <template #prepend>
        <v-avatar v-ripple image="@/assets/logo.svg" size="28" rounded="lg" title="Nexus" @click="moveIndex"/>
    </template>
    <!-- 보드 페이지는 이 자리에 보드 메뉴를 Teleport로 넣고(닉네임은 그 메뉴 안), 그 외 페이지는 닉네임 -->
    <v-app-bar-title class="text-primary">
        <div id="app-head-title" />
        <template v-if="!route.meta.board">{{ auth?.nickname ?? '-' }}</template>
    </v-app-bar-title>
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
import { useAppDialog } from '../store/appDialog.store';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '@/feature/user/store/auth.store';
import { computed } from 'vue';
import { useResizeCallback } from '@/feature/common/store/resizeGroup.store';
import { computeResizeSize } from '@/common/hook/vue.hook';

const router = useRouter();
const route = useRoute();
// @ts-ignore
const { vElRef, size } = useAppHead()

const elRef = computed(() => vElRef.value?.$el as HTMLElement|undefined)
useResizeCallback(elRef, () => {
    size.value = computeResizeSize(elRef.value || null)
})
const { open } = useAppNav();
const { auth, logout } = useAuthStore();
const { open: dialogOpen, openDialog } = useAppDialog();

// 로그아웃은 확인 뒤 — 이동은 라우터(authGuard)가 한다 — 로그인-4(화면 정리)
const toggleNaiv = () => {
    // open.value = !open.value
    if (!auth.value) {
        router.push('/login');
        return;
    }
    openDialog({
        type: 'warning',
        title: '로그아웃',
        content: '로그아웃할까요? 실행 중인 터미널은 계속 실행됩니다.',
        buttons: [
            { view: '취소', type: 'info', bindButton: 'Escape', action: () => dialogOpen.value = false },
            { view: '로그아웃', type: 'warning', bindButton: 'Enter', action: () => { dialogOpen.value = false; logout(); } },
        ],
    })
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