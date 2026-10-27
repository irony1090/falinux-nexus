<script setup lang="ts">
import { LazyListener } from '@/common/listener/lazy.listener';
import { isNil } from '@/common/util/index.util';
import AuthTile from '@/feature/member/component/AuthTile.vue';
import CtTextField from '@/feature/widget/component/CtTextField.vue';
import { useAuthStore } from '@/feature/user/store/auth.store';
import Cookies from 'js-cookie';
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { VIcon } from 'vuetify/components';

const REMEMBER_ID_COOKIE_KEY = 'remember_id';

// 로그인 — 가운데 타일 하나, 오류 = 빨간 알약 한 줄 — 화-16(로그인·가입). 성공 뒤 이동은 라우터(authGuard)
const { loading, login, offline } = useAuthStore();
const router = useRouter();
const isLazyLoading = ref(false);
const lazyListener = new LazyListener(false, { time: 750, startLazy: v => !v });
lazyListener.onValue = v => isLazyLoading.value = v;
watch(loading, val => lazyListener.setValue(val), { immediate: true });

// 개발용 기본값은 `.env.development.local`(git 제외)에만 — 로그인-4(화면 정리)
const id = ref(Cookies.get(REMEMBER_ID_COOKIE_KEY) ?? import.meta.env.VITE_DEV_LOGIN_ID ?? '');
const pw = ref(import.meta.env.VITE_DEV_LOGIN_PW ?? '');
const rememberId = ref(!isNil(Cookies.get(REMEMBER_ID_COOKIE_KEY)));

const error = ref<string>();
watch([id, pw], () => error.value = undefined);
const shownError = computed(() => error.value ?? (offline.value ? '서버에 연결할 수 없습니다' : undefined));

const submit = () => {
    if (!id.value || !pw.value) {
        error.value = '아이디와 비밀번호를 입력하세요';
        return;
    }
    login(id.value, pw.value)
    .then(() => {
        if (rememberId.value) Cookies.set(REMEMBER_ID_COOKIE_KEY, id.value, { expires: 7 });
        else Cookies.remove(REMEMBER_ID_COOKIE_KEY);
    })
    .catch(err => error.value = err?.message || '로그인에 실패했습니다');
}
</script>

<template>
<auth-tile title="NEXUS-supervisor" submit-label="로그인" :busy="isLazyLoading" @submit="submit">
    <label for="au-id" class="mb-lbl">아이디</label>
    <CtTextField id="au-id" v-model="id" size="sm" name="identification" autocomplete="username" :disabled="isLazyLoading" />
    <label for="au-pw" class="mb-lbl">비밀번호</label>
    <CtTextField id="au-pw" v-model="pw" size="sm" type="password" name="password" autocomplete="current-password" :disabled="isLazyLoading" />
    <span />
    <button type="button" class="remember" :class="{ on: rememberId }" :aria-pressed="rememberId" @click="rememberId = !rememberId">
        <v-icon icon="mdi-check" size="14" class="chk" />아이디 저장
    </button>
    <p v-if="shownError" role="alert" class="mb-err full">{{ shownError }}</p>
    <template #note>처음이면 <button type="button" class="link" @click="router.push('/signup')">회원가입</button></template>
</auth-tile>
</template>

<style scoped lang="scss">
.full { grid-column: 1 / -1; }
// 아이디 저장 = 토글 알약
.remember {
    justify-self: start;
    height: 28px;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 0 12px 0 10px;
    border: 0;
    border-radius: 999px;
    background: rgba(var(--v-theme-on-surface), 0.07);
    color: rgb(var(--v-theme-on-surface));
    font: inherit;
    @include vars.font('label-medium');
    font-weight: vars.$weight-md;
    cursor: pointer;
    .chk { opacity: 0.25; }
    &.on { background: rgba(var(--v-theme-primary), 0.14); color: rgb(var(--v-theme-primary)); .chk { opacity: 1; } }
}
.link {
    padding: 0;
    border: 0;
    background: none;
    color: rgb(var(--v-theme-primary));
    font: inherit;
    font-weight: vars.$weight-lg;
    text-decoration: underline;
    cursor: pointer;
}
</style>
