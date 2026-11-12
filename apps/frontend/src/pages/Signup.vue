<script setup lang="ts">
import AuthTile from '@/feature/member/component/AuthTile.vue';
import { createUser } from '@/feature/user/api/user.api';
import { useAvailable } from '@/feature/user/hook/available.hook';
import { useAuthStore } from '@/feature/user/store/auth.store';
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import CtTextField from '@/feature/widget/component/CtTextField.vue';

// 회원가입 — 로그인 아이디 · 공개 닉네임 따로, 입력이 멈추면 중복 확인. 가입하면 바로 로그인 — 아이디-2(가입 분리)·아이디-3(가입 화면), 모양 = 화-16(로그인·가입)
const { login } = useAuthStore();
const router = useRouter();

const id = ref('');
const nick = ref('');
const pw = ref('');
const pw2 = ref('');
const busy = ref(false);
const error = ref<string>();
watch([id, nick, pw, pw2], () => error.value = undefined);
const mismatch = computed(() => !!pw2.value && pw.value !== pw2.value);

const idCheck = useAvailable('identification', id);
const nickCheck = useAvailable('nickname', nick);
const errorOf = (c: ReturnType<typeof useAvailable>) => c.state.value === 'bad' ? c.message.value : undefined;
const hintOf = (c: ReturnType<typeof useAvailable>, idle: string, ok: string) =>
    c.state.value === 'checking' ? '확인 중…' : c.state.value === 'ok' ? ok : idle;

// 성공 뒤 이동은 라우터(authGuard)가 한다
const submit = () => {
    if (!id.value.trim() || !nick.value.trim() || !pw.value) { error.value = '아이디·닉네임·비밀번호를 입력하세요'; return; }
    if (idCheck.state.value === 'bad' || nickCheck.state.value === 'bad') { error.value = '아이디나 닉네임을 바꿔 주세요'; return; }
    if (pw.value !== pw2.value) { error.value = '비밀번호 확인이 다릅니다'; return; }
    busy.value = true;
    createUser({ identification: id.value.trim(), nickname: nick.value.trim(), password: pw.value })
    .then(() => login(id.value.trim(), pw.value))
    .catch(err => error.value = err?.message || '가입에 실패했습니다')
    .finally(() => busy.value = false);
}
</script>

<template>
<auth-tile title="회원가입" submit-label="가입" back :busy="busy" @submit="submit" @back="router.push('/login')">
    <label for="au-id" class="mb-lbl">아이디</label>
    <CtTextField id="au-id" v-model="id" size="sm" autocomplete="username" :disabled="busy"
        :error-messages="errorOf(idCheck)" :hint="hintOf(idCheck, '로그인할 때 씁니다(공백 없이)', '쓸 수 있는 아이디입니다')" persistent-hint />
    <label for="au-nick" class="mb-lbl">닉네임</label>
    <CtTextField id="au-nick" v-model="nick" size="sm" autocomplete="nickname" :disabled="busy"
        :error-messages="errorOf(nickCheck)" :hint="hintOf(nickCheck, '화면에 이름으로 보입니다', '쓸 수 있는 닉네임입니다')" persistent-hint />
    <label for="au-pw" class="mb-lbl">비밀번호</label>
    <CtTextField id="au-pw" v-model="pw" size="sm" type="password" autocomplete="new-password" :disabled="busy" />
    <label for="au-pw2" class="mb-lbl">비밀번호 확인</label>
    <CtTextField id="au-pw2" v-model="pw2" size="sm" type="password" autocomplete="new-password" :disabled="busy" :error-messages="mismatch ? '비밀번호가 서로 다릅니다' : undefined" />
    <p v-if="error" role="alert" class="mb-err full">{{ error }}</p>
    <template #note>가입하면 바로 로그인됩니다</template>
</auth-tile>
</template>

<style scoped lang="scss">
.full { grid-column: 1 / -1; }
</style>
