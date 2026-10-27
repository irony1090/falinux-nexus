<script setup lang="ts">
import { useAuthStore } from '@/feature/user/store/auth.store';
import { useQueryClient } from '@tanstack/vue-query';
import { computed, ref } from 'vue';
import CtAutocomplete from '@/feature/widget/component/CtAutocomplete.vue';
import CtTextField from '@/feature/widget/component/CtTextField.vue';
import { adminCreateDevice, createMyDevice, invalidateMember, useAdminUsers, useMyDevices } from '../api/device.api';
import { useMemberBoard } from '../store/memberBoard.store';
import { deviceKey, type MemberTile } from '../util/memberGrid.util';
import { API_KEY_PREFIX, MAIN_KEY_PATTERN } from '../util/member.util';

// 장비 등록 타일 — 이름·(관리자: 주인)·붙여넣을 키. 등록하면 같은 자리에 장비 타일(키·env 펼침) — 회원-3·회원-Q17(대신 등록)
const props = defineProps<{ tile: MemberTile }>();
const { admin, replaceTile } = useMemberBoard();
const { auth } = useAuthStore();
const queryClient = useQueryClient();

const ownersQuery = admin ? useAdminUsers({ size: 0 }) : undefined;
const owners = computed(() => (ownersQuery?.data.value?.items ?? []).map(o => ({ title: o.nickname, value: o.id })));
const myQuery = admin ? undefined : useMyDevices({ size: 1 });
const myCount = computed(() => myQuery?.data.value?.total ?? 0);

const name = ref('');
const apiKey = ref('');
const owner = ref<number>(auth.value?.id ?? 0);
const busy = ref(false);
const message = ref<string>();

const nameOk = computed(() => !name.value || MAIN_KEY_PATTERN.test(name.value));
const keyOk = computed(() => !apiKey.value.trim() || apiKey.value.trim().startsWith(API_KEY_PREFIX));
const ready = computed(() => !!name.value && nameOk.value && keyOk.value && (!admin || !!owner.value));
const note = computed(() => admin ? '상한·권한과 상관없이 등록합니다' : `내 장비 ${myCount.value} / ${auth.value?.workerLimit ?? 0}`);

const submit = () => {
    if (!ready.value || busy.value) return;
    busy.value = true;
    message.value = undefined;
    const k = apiKey.value.trim();
    const job = admin ? adminCreateDevice(owner.value, name.value, k) : createMyDevice(name.value, k);
    job.then(async r => {
        await invalidateMember(queryClient);
        replaceTile(props.tile.key, deviceKey(r.mainKey), r.apiKey);
    })
    .catch(err => message.value = err?.message || '등록에 실패했습니다')
    .finally(() => busy.value = false);
}
</script>

<template>
<form class="mb-body" @submit.prevent="submit">
    <div class="mb-fields top">
        <label for="nf-name" class="mb-lbl">장비 이름</label>
        <CtTextField id="nf-name" v-model="name" size="sm" placeholder="예: line3-pc" autocomplete="off" :disabled="busy" persistent-hint
            hint="영문·숫자·. _ - 로 64자까지 (worker env의 NAME과 같게)" :error-messages="nameOk ? undefined : '쓸 수 없는 글자가 있습니다. 영문·숫자·. _ - 만 됩니다'" />
        <template v-if="admin">
            <label for="nf-owner" class="mb-lbl">주인</label>
            <CtAutocomplete id="nf-owner" v-model="owner" size="sm" :items="owners" :disabled="busy" />
        </template>
        <label for="nf-key" class="mb-lbl">API 키</label>
        <CtTextField id="nf-key" v-model="apiKey" size="sm" class="mb-mono" placeholder="비우면 새로 발급" autocomplete="off" :disabled="busy" persistent-hint
            :hint="`다른 supervisor에서 받은 키(${API_KEY_PREFIX}…)를 그대로 쓰려면 붙여 넣습니다`" :error-messages="keyOk ? undefined : `${API_KEY_PREFIX}로 시작하는 키만 됩니다`" />
        <p v-if="message" class="mb-err full">{{ message }}</p>
    </div>
    <ol class="steps">
        <li>등록하면 이 자리에 장비 타일이 열리고 API 키와 env 두 줄이 나옵니다</li>
        <li>그 두 줄을 worker의 env에 넣고 실행합니다</li>
        <li>접속하면 인스턴스 목록에 나타나고 라벨을 달 수 있습니다</li>
    </ol>
    <button type="submit" hidden />
</form>
<div class="mb-foot">
    <span class="mb-note">{{ note }}</span>
    <div class="mb-end">
        <button type="button" class="mb-btn mb-btn--primary" :disabled="!ready || busy" @click="submit">등록</button>
    </div>
</div>
</template>

<style scoped lang="scss">
.full { grid-column: 1 / -1; }
.steps {
    margin: 0;
    padding: 12px 16px 14px 34px;
    border-top: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    display: flex;
    flex-direction: column;
    gap: 6px;
    @include vars.font('body-small');
    color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity));
}
</style>
