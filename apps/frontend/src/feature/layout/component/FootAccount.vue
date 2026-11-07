<script setup lang="ts">
import FootPill from '@/feature/node/component/FootPill.vue';
import { openSettings } from '@/feature/settings/store/settings.store';
import { computed } from 'vue';
import { VBtn, VIcon, VList, VListItem, VMenu } from 'vuetify/components';
import { useAppActions } from '../hook/appActions.hook';

// 푸터 오른쪽 메뉴(모든 페이지 공용): PC = [● 이름][보드 전용 slot][전체 화면], 폰 = [⋯] 하나 — REF-settings-footer.md "최종 모양"
// 보드 전용 버튼(전송·알림)은 기본 slot으로 받는다 — 전송(useTransfers)이 보드 페이지 provide라 여기서 부르면 관리 페이지에서 깨짐
defineProps<{
    phone: boolean;
    dot?: boolean;      // ⋯에 빨간 점(보드 = 안 읽은 알림·전송 중)
}>();

const { auth, toggleAuth, canFullscreen, fullscreen, toggleFullscreen } = useAppActions();
const name = computed(() => auth.value?.nickname ?? '');   // identification은 서버가 해시로 줌 — 화면엔 닉네임만
const fsIcon = computed(() => fullscreen.value ? 'mdi-fullscreen-exit' : 'mdi-fullscreen');
const fsTitle = computed(() => fullscreen.value ? '전체 화면 끝내기' : '전체 화면');
</script>

<template>
<foot-pill v-if="!phone" edge="right">
    <v-menu v-if="auth" location="top end" offset="10">
        <template #activator="{ props: act, isActive }">
            <button v-bind="act" type="button" class="acct" :class="{ on: isActive }" :title="name" :aria-label="`계정 메뉴 - ${name}`">
                <span class="me" />
                <span class="nm">{{ name }}</span>
                <v-icon size="16" icon="mdi-chevron-up" />
            </button>
        </template>
        <v-list density="compact" min-width="200">
            <v-list-item title="설정" prepend-icon="mdi-cog-outline" @click="openSettings" />
            <v-list-item title="로그아웃" prepend-icon="mdi-logout" @click="toggleAuth" />
        </v-list>
    </v-menu>
    <slot :phone="false" />
    <v-btn v-if="canFullscreen" size="small" density="comfortable" variant="text" :icon="fsIcon" :title="fsTitle" @click="toggleFullscreen" />
    <v-btn v-if="!auth" size="small" density="comfortable" variant="text" icon="mdi-login" title="로그인" @click="toggleAuth" />
</foot-pill>
<!-- 폰: 하나로 합쳐 위로 펼침(가로 알약) -->
<foot-pill v-else edge="right">
    <v-menu location="top end" offset="10">
        <template #activator="{ props: act, isActive }">
            <v-btn v-bind="act" density="comfortable" variant="text" icon :active="isActive" title="더 보기" aria-label="더 보기 - 계정, 전체 화면, 설정, 로그아웃">
                <v-icon icon="mdi-dots-horizontal" />
                <span v-if="dot && !isActive" class="dot" />
            </v-btn>
        </template>
        <foot-pill class="pop">
            <span v-if="auth" class="who" :title="name"><span class="me" /><span class="nm">{{ name }}</span></span>
            <slot :phone="true" />
            <v-btn v-if="canFullscreen" density="comfortable" variant="text" :icon="fsIcon" :title="fsTitle" @click="toggleFullscreen" />
            <v-btn v-if="auth" density="comfortable" variant="text" icon="mdi-cog-outline" title="설정" @click="openSettings" />
            <v-btn density="comfortable" variant="text" :icon="auth ? 'mdi-logout' : 'mdi-login'" :title="auth ? '로그아웃' : '로그인'" @click="toggleAuth" />
        </foot-pill>
    </v-menu>
</foot-pill>
</template>

<style scoped lang="scss">
$sep: inset 1px 0 0 rgba(var(--v-border-color), var(--v-border-opacity));

.acct {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    max-width: 200px;
    padding: 0 8px 0 16px;
    border: none;
    border-radius: 0;       // 알약(FootPill edge)이 모양을 자른다
    background: none;
    color: inherit;
    font: inherit;
    @include vars.font('label-large');
    font-weight: vars.$weight-md;
    cursor: pointer;
    &:hover { background: rgba(var(--v-theme-on-surface), 0.06); }
    &.on { background: rgba(var(--v-theme-primary), 0.12); color: rgb(var(--v-theme-primary)); }
    & + :deep(.v-btn) { box-shadow: $sep; }
}
.me {
    flex: none;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: rgb(var(--v-theme-success));
}
.nm {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.dot {
    position: absolute;
    top: 8px;
    right: 8px;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: rgb(var(--v-theme-error));
}
// 위로 펼친 가로 알약 — 메뉴는 푸터 밖(body)에 그려져 --pill-h를 못 받으므로 높이를 직접
.pop {
    height: 44px;
    padding: 0;
    align-items: stretch;
    overflow: hidden;
    > :deep(.v-btn) { width: 46px; height: auto; border-radius: 0; }
    > :deep(* + .v-btn) { box-shadow: $sep; }
}
.who {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    max-width: 128px;
    padding: 0 12px 0 16px;
    @include vars.font('label-large');
    font-weight: vars.$weight-md;
}
</style>
