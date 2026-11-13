<script setup lang="ts">
import { useAppDialog } from '@/feature/layout/store/appDialog.store';
import { agoText, parseUserAgent } from '@/feature/node/util/presenceView.util';
import CtTextField from '@/feature/widget/component/CtTextField.vue';
import { computed, onMounted, ref, watch } from 'vue';
import { VIcon } from 'vuetify/components';
import { changeNickname, changePassword, endAllLogins, endLogin, listLogins, type LoginResponse } from '../api/account.api';
import { useAvailable } from '../hook/available.hook';
import { useAuthStore } from '../store/auth.store';

// 설정 창 계정 탭 — 닉네임 · 비밀번호 · 로그인한 기기 → REF-account.md 세션-5(설정 화면)
const { auth, refresh, ended } = useAuthStore();
const { open: dialogOpen, openDialog } = useAppDialog();

// ---- 닉네임 ----
const nick = ref(auth.value?.nickname ?? '');
const nickCheck = useAvailable('nickname', nick, v => v.toLowerCase() === (auth.value?.nickname ?? '').toLowerCase());
const nickBusy = ref(false);
const nickMsg = ref<{ text: string, err?: boolean }>();
const nickChanged = computed(() => nick.value.trim() !== '' && nick.value.trim() !== auth.value?.nickname);
const nickOff = computed(() => nickBusy.value || !nickChanged.value || nickCheck.state.value === 'bad' || nickCheck.state.value === 'checking');
const nickHint = computed(() => nickMsg.value?.text
    ?? (nickCheck.state.value === 'bad' ? nickCheck.message.value : nickCheck.state.value === 'checking' ? '확인 중…' : nickCheck.state.value === 'ok' ? '쓸 수 있는 닉네임입니다' : '화면에 이름으로 보입니다'));
const nickErr = computed(() => nickMsg.value ? !!nickMsg.value.err : nickCheck.state.value === 'bad');
watch(nick, () => nickMsg.value = undefined);
const saveNick = () => {
    if (nickOff.value) return;
    nickBusy.value = true;
    changeNickname(nick.value.trim())
    .then(() => refresh())
    .then(() => nickMsg.value = { text: '바꿨습니다' })
    .catch(err => nickMsg.value = { text: err?.message || '바꾸지 못했습니다', err: true })
    .finally(() => nickBusy.value = false);
}

// ---- 비밀번호 ----
const cur = ref('');
const next = ref('');
const next2 = ref('');
const pwBusy = ref(false);
const pwMsg = ref<{ text: string, err?: boolean }>();
const mismatch = computed(() => !!next2.value && next.value !== next2.value);
watch([cur, next, next2], () => pwMsg.value = undefined);
const savePw = () => {
    if (!cur.value || !next.value) { pwMsg.value = { text: '현재 비밀번호와 새 비밀번호를 입력하세요', err: true }; return; }
    if (next.value !== next2.value) { pwMsg.value = { text: '새 비밀번호 확인이 다릅니다', err: true }; return; }
    pwBusy.value = true;
    changePassword(cur.value, next.value)
    .then(() => {
        cur.value = next.value = next2.value = '';
        pwMsg.value = { text: '바꿨습니다 · 다른 기기는 로그아웃했습니다' };
        void loadLogins();
    })
    .catch(err => pwMsg.value = { text: err?.message || '바꾸지 못했습니다', err: true })
    .finally(() => pwBusy.value = false);
}

// ---- 로그인한 기기 ----
const logins = ref<LoginResponse[]>([]);
const loginErr = ref<string>();
const loginBusy = ref(false);
// 지금 기기 먼저, 나머지는 서버 순서(최근 로그인 먼저)
const sorted = computed(() => [...logins.value].sort((a, b) => Number(b.current) - Number(a.current)));
const loadLogins = () => listLogins()
    .then(items => { logins.value = items; loginErr.value = undefined; })
    .catch(err => loginErr.value = err?.message || '기기 목록을 불러오지 못했습니다');
onMounted(loadLogins);

const loginSub = (l: LoginResponse) => [
    l.current && '지금 기기',
    l.online ? '접속 중' : `마지막 활동 ${agoText(l.lastSeenAt)}`,
    `로그인 ${agoText(l.createdAt)}`,
].filter(Boolean).join(' · ');

const logoutOne = (l: LoginResponse) => {
    loginBusy.value = true;
    endLogin(l.key)
    .then(r => r.current ? ended() : loadLogins())
    .catch(err => { loginErr.value = err?.message || '로그아웃하지 못했습니다'; void loadLogins(); })
    .finally(() => loginBusy.value = false);
}

const logoutAll = () => openDialog({
    type: 'warning',
    title: '모든 기기에서 로그아웃',
    content: ['이 기기를 포함해 로그인한 모든 기기에서 로그아웃합니다.', '실행 중인 process는 그대로 남습니다.'],
    buttons: [
        { view: '취소', type: 'info', bindButton: 'Escape', action: () => dialogOpen.value = false },
        { view: '모두 로그아웃', type: 'warning', action: () => {
            dialogOpen.value = false;
            loginBusy.value = true;
            endAllLogins()
            .then(() => ended())
            .catch(err => loginErr.value = err?.message || '로그아웃하지 못했습니다')
            .finally(() => loginBusy.value = false);
        } },
    ],
});
</script>

<template>
<div class="AccountPanel">
    <p class="sec">닉네임<span class="sub">아이디 {{ auth?.identification }}</span></p>
    <form class="row" novalidate @submit.prevent="saveNick">
        <div class="lb">
            <CtTextField v-model="nick" size="sm" autocomplete="nickname" aria-label="닉네임" :disabled="nickBusy" />
            <span class="hint" :class="{ err: nickErr }">{{ nickHint }}</span>
        </div>
        <span class="ct-grp ct-grp--r pick">
            <button type="submit" class="ct-seg ct-seg--primary" :disabled="nickOff">변경</button>
        </span>
    </form>

    <p class="sec">비밀번호<span class="sub">바꾸면 다른 기기는 로그아웃</span></p>
    <form class="row top" novalidate @submit.prevent="savePw">
        <div class="lb fields">
            <CtTextField v-model="cur" size="sm" type="password" autocomplete="off" placeholder="현재 비밀번호" aria-label="현재 비밀번호" :disabled="pwBusy" />
            <CtTextField v-model="next" size="sm" type="password" autocomplete="off" placeholder="새 비밀번호" aria-label="새 비밀번호" :disabled="pwBusy" />
            <CtTextField v-model="next2" size="sm" type="password" autocomplete="off" placeholder="새 비밀번호 확인" aria-label="새 비밀번호 확인" :disabled="pwBusy"
                :error-messages="mismatch ? '새 비밀번호가 서로 다릅니다' : undefined" />
            <span v-if="pwMsg" class="hint" :class="{ err: pwMsg.err }" role="status">{{ pwMsg.text }}</span>
        </div>
        <span class="ct-grp ct-grp--r pick">
            <button type="submit" class="ct-seg ct-seg--primary" :disabled="pwBusy">변경</button>
        </span>
    </form>

    <p class="sec">로그인한 기기<span class="sub">{{ logins.length }}대</span></p>
    <p v-if="loginErr" class="row hint err">{{ loginErr }}</p>
    <div v-for="l in sorted" :key="l.key" class="row">
        <div class="lb">
            <span class="name"><v-icon size="14" :icon="l.online ? 'mdi-circle' : 'mdi-circle-outline'" :color="l.online ? 'success' : undefined" class="dot" />{{ parseUserAgent(l.userAgent) }}</span>
            <span class="hint">{{ loginSub(l) }}</span>
        </div>
        <span v-if="!l.current" class="ct-grp ct-grp--r pick">
            <button type="button" class="ct-seg" :disabled="loginBusy" @click="logoutOne(l)">로그아웃</button>
        </span>
    </div>
    <div class="row">
        <div class="lb"><span class="hint">이 기기를 포함해 전부 로그아웃합니다</span></div>
        <span class="ct-grp ct-grp--r pick">
            <button type="button" class="ct-seg ct-seg--error" :disabled="loginBusy" @click="logoutAll">모든 기기에서 로그아웃</button>
        </span>
    </div>
</div>
</template>

<style scoped lang="scss">
// SettingsDialog와 같은 여백 — 왼쪽 16px, 오른쪽 0(버튼 묶음이 창 벽에 붙게)
.sec {
    margin: vars.$spacing-md 0 vars.$spacing-sm;
    padding: 0 vars.$spacing-xl;
    @include vars.font('label-medium');
    font-weight: vars.$weight-lg;
    opacity: 0.6;
    > .sub { margin-left: vars.$spacing-sm; font-weight: vars.$weight-md; }
}
.row {
    display: flex;
    align-items: center;
    gap: vars.$spacing-xl;
    margin: 0;
    padding: vars.$spacing-sm 0 vars.$spacing-md vars.$spacing-xl;
    &.top { align-items: flex-start; }
}
.lb {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
}
.fields { gap: vars.$spacing-sm; }
.name { @include vars.font('body-medium'); font-weight: vars.$weight-md; display: flex; align-items: center; gap: 6px; }
.dot { opacity: 0.8; }
.hint { @include vars.font('body-small'); opacity: 0.7; &.err { color: rgb(var(--v-theme-error)); opacity: 1; } }
.pick { height: 32px; flex: none; }
</style>
