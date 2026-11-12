<script setup lang="ts">
import { useAvailable } from '@/feature/user/hook/available.hook';
import CtTextField from '@/feature/widget/component/CtTextField.vue';
import { computed, ref, watch } from 'vue';
import { adminPatchAccount, adminResetPassword, type AdminUserResponse } from '../api/device.api';

// 회원 계정 수정(관리자) — 아이디·닉네임(로그인 유지) / 비밀번호 재설정(그 회원 기기 전부 로그아웃) → REF-account.md "관리자 계정 수정"
const props = defineProps<{ user: AdminUserResponse, me: boolean }>();
const emit = defineEmits<{ saved: [] }>();
const open = defineModel<boolean>({ default: false });

const ident = ref('');
const nick = ref('');
const pw = ref('');
const pw2 = ref('');
const busy = ref(false);
const message = ref<{ text: string, err?: boolean }>();
watch(open, v => {
    if (!v) return;
    ident.value = props.user.identification;
    nick.value = props.user.nickname;
    pw.value = pw2.value = '';
    message.value = undefined;
});
watch([ident, nick, pw, pw2], () => message.value = undefined);

const same = (now: () => string) => (v: string) => v.toLowerCase() === now().toLowerCase();
const idCheck = useAvailable('identification', ident, same(() => props.user.identification));
const nickCheck = useAvailable('nickname', nick, same(() => props.user.nickname));
const idChanged = computed(() => ident.value.trim() !== props.user.identification);
const nickChanged = computed(() => nick.value.trim() !== props.user.nickname);
const checking = computed(() => idCheck.state.value === 'checking' || nickCheck.state.value === 'checking');
const accountOff = computed(() => busy.value || checking.value || (!idChanged.value && !nickChanged.value)
    || idCheck.state.value === 'bad' || nickCheck.state.value === 'bad');
const hint = (c: ReturnType<typeof useAvailable>, idle: string) =>
    c.state.value === 'bad' ? undefined : c.state.value === 'checking' ? '확인 중…' : c.state.value === 'ok' ? '쓸 수 있습니다' : idle;
const mismatch = computed(() => !!pw2.value && pw.value !== pw2.value);

const run = (job: () => Promise<unknown>, done: string) => {
    busy.value = true;
    job().then(() => { message.value = { text: done }; emit('saved'); })
    .catch(err => message.value = { text: err?.message || '요청에 실패했습니다', err: true })
    .finally(() => busy.value = false);
}
const saveAccount = () => {
    if (accountOff.value) return;
    const p: { identification?: string, nickname?: string } = {};
    if (idChanged.value) p.identification = ident.value.trim();
    if (nickChanged.value) p.nickname = nick.value.trim();
    run(() => adminPatchAccount(props.user.id, p), '아이디·닉네임을 바꿨습니다');
}
const resetPw = () => {
    if (!pw.value) { message.value = { text: '새 비밀번호를 입력하세요', err: true }; return; }
    if (pw.value !== pw2.value) { message.value = { text: '새 비밀번호 확인이 다릅니다', err: true }; return; }
    run(() => adminResetPassword(props.user.id, pw.value).then(() => { pw.value = pw2.value = ''; }),
        props.me ? '비밀번호를 바꿨습니다 · 다른 기기는 로그아웃했습니다' : '비밀번호를 바꿨습니다 · 이 회원의 모든 기기를 로그아웃했습니다');
}
</script>

<template>
<v-dialog v-model="open" max-width="440" min-width="296" :persistent="busy" scrollable>
    <v-card :loading="busy" class="ct-card">
        <v-card-title>{{ user.nickname }} 계정 수정</v-card-title>
        <v-card-text class="body">
            <form class="grp" novalidate @submit.prevent="saveAccount">
                <p class="mb-lbl">아이디</p>
                <CtTextField v-model="ident" size="sm" autocomplete="off" aria-label="아이디" :disabled="busy" placeholder="로그인 아이디(공백 없이)"
                    :error-messages="idCheck.state.value === 'bad' ? idCheck.message.value : undefined" :hint="hint(idCheck, user.identification ? '' : '아이디가 없어 로그인할 수 없습니다')" persistent-hint />
                <p class="mb-lbl">닉네임</p>
                <CtTextField v-model="nick" size="sm" autocomplete="off" aria-label="닉네임" :disabled="busy"
                    :error-messages="nickCheck.state.value === 'bad' ? nickCheck.message.value : undefined" :hint="hint(nickCheck, '')" persistent-hint />
                <div class="act"><span class="ct-grp ct-grp--r"><button type="submit" class="ct-seg ct-seg--primary" :disabled="accountOff">아이디·닉네임 저장</button></span></div>
            </form>
            <form class="grp" novalidate @submit.prevent="resetPw">
                <p class="mb-lbl">비밀번호 재설정 <span class="dim">{{ me ? '· 다른 기기는 로그아웃' : '· 이 회원의 모든 기기 로그아웃' }}</span></p>
                <CtTextField v-model="pw" size="sm" type="password" autocomplete="new-password" placeholder="새 비밀번호" aria-label="새 비밀번호" :disabled="busy" />
                <CtTextField v-model="pw2" size="sm" type="password" autocomplete="new-password" placeholder="새 비밀번호 확인" aria-label="새 비밀번호 확인" :disabled="busy"
                    :error-messages="mismatch ? '새 비밀번호가 서로 다릅니다' : undefined" />
                <div class="act"><span class="ct-grp ct-grp--r"><button type="submit" class="ct-seg ct-seg--warning" :disabled="busy">비밀번호 재설정</button></span></div>
            </form>
            <p v-if="message" class="ct-msg" :class="{ 'ct-msg--error': message.err }" role="status">{{ message.text }}</p>
        </v-card-text>
        <div class="ct-actions">
            <span class="ct-grp ct-grp--r">
                <button type="button" class="ct-seg" :disabled="busy" @click="open = false">닫기</button>
            </span>
        </div>
    </v-card>
</v-dialog>
</template>

<style scoped lang="scss">
.body { display: flex; flex-direction: column; gap: vars.$spacing-lg; padding-right: 0 !important; }
.grp { display: flex; flex-direction: column; gap: vars.$spacing-sm; padding-right: vars.$spacing-xl; }
.act { display: flex; justify-content: flex-end; height: 32px; margin-right: calc(-1 * vars.$spacing-xl); }
.dim { opacity: 0.6; font-weight: normal; }
.grp > .mb-lbl { margin: 0; }
</style>
