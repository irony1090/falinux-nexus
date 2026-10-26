<script setup lang="ts">
import AuthTile from '@/feature/member/component/AuthTile.vue';
import { createUser } from '@/feature/user/api/user.api';
import { useAuthStore } from '@/feature/user/store/auth.store';
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

// 회원가입 — 아이디·비밀번호만, 닉네임 = 아이디(서버). 가입하면 바로 로그인 — 회원-2(회원가입)·회원-Q12(닉네임), 모양 = 화-16(로그인·가입)
const { login } = useAuthStore();
const router = useRouter();

const id = ref('');
const pw = ref('');
const pw2 = ref('');
const busy = ref(false);
const error = ref<string>();
watch([id, pw, pw2], () => error.value = undefined);
const mismatch = computed(() => !!pw2.value && pw.value !== pw2.value);

// 성공 뒤 이동은 라우터(authGuard)가 한다
const submit = () => {
    if (!id.value || !pw.value) { error.value = '아이디와 비밀번호를 입력하세요'; return; }
    if (id.value.length > 64) { error.value = '아이디는 64자까지입니다'; return; }
    if (pw.value !== pw2.value) { error.value = '비밀번호 확인이 다릅니다'; return; }
    busy.value = true;
    createUser({ identification: id.value, password: pw.value })
    .then(() => login(id.value, pw.value))
    .catch(err => error.value = err?.message || '가입에 실패했습니다')
    .finally(() => busy.value = false);
}
</script>

<template>
<auth-tile title="회원가입" submit-label="가입" back :busy="busy" @submit="submit" @back="router.push('/login')">
    <label for="au-id" class="mb-lbl">아이디</label>
    <div class="col">
        <input id="au-id" v-model="id" class="mb-input" autocomplete="username" :disabled="busy" />
        <span class="mb-hint">아이디가 화면에 이름으로 보입니다</span>
    </div>
    <label for="au-pw" class="mb-lbl">비밀번호</label>
    <input id="au-pw" v-model="pw" class="mb-input" type="password" autocomplete="new-password" :disabled="busy" />
    <label for="au-pw2" class="mb-lbl">비밀번호 확인</label>
    <div class="col">
        <input id="au-pw2" v-model="pw2" class="mb-input" :class="{ bad: mismatch }" type="password" autocomplete="new-password" :disabled="busy" />
        <span v-if="mismatch" class="mb-hint bad">비밀번호가 서로 다릅니다</span>
    </div>
    <p v-if="error" role="alert" class="mb-err full">{{ error }}</p>
    <template #note>가입하면 바로 로그인됩니다</template>
</auth-tile>
</template>

<style scoped lang="scss">
.col { min-width: 0; display: flex; flex-direction: column; gap: 4px; }
.full { grid-column: 1 / -1; }
</style>
