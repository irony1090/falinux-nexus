<script setup lang="ts">
import { useAppDialog } from '@/feature/layout/store/appDialog.store';
import { PERM_ADMIN, PERM_WORKER_REGISTER } from '@/feature/user/api/user.api';
import { useAuthStore } from '@/feature/user/store/auth.store';
import { useQueryClient } from '@tanstack/vue-query';
import { computed, ref, watch } from 'vue';
import {
    adminDeleteUser, adminPatchUser, adminSuspendUser, invalidateMember, useAdminDevices, useMemberUser,
    type AdminUserPatch,
} from '../api/device.api';
import { useMemberBoard } from '../store/memberBoard.store';
import { deviceKey, userIdOf, type MemberTile } from '../util/memberGrid.util';
import { dayText, deviceState, deviceStateText } from '../util/member.util';

// 회원 타일(관리자) — 상태·상한·권한 + 그 회원 장비 줄(누르면 장비 타일, 화-8(회원 -> 장비)) + 정지·삭제 — 회원-5(관리자 화면)
const props = defineProps<{ tile: MemberTile }>();
const id = userIdOf(props.tile.key);
const { open, close, dropMissing } = useMemberBoard();
const { auth } = useAuthStore();
const { open: dialogOpen, openDialog } = useAppDialog();
const queryClient = useQueryClient();

const userQuery = useMemberUser(id);
const devicesQuery = useAdminDevices({ owner: id, size: 0 });
const u = computed(() => userQuery.data.value);
const devices = computed(() => devicesQuery.data.value?.items ?? []);
const me = computed(() => id === auth.value?.id);
const isAdmin = computed(() => !!u.value?.perms.includes(PERM_ADMIN));
const canReg = computed(() => !!u.value?.perms.includes(PERM_WORKER_REGISTER));
const state = computed(() => u.value?.suspendedAt ? { text: '정지', cls: 'mb-tag--error' } : isAdmin.value ? { text: '관리자', cls: 'mb-tag--primary' } : { text: '회원', cls: '' });
const note = computed(() => me.value ? '자기 자신은 정지·삭제·관리자 해제를 못 합니다'
    : u.value?.suspendedAt ? '정지 중: 로그인·worker 접속 막힘' : '정지하면 로그인·worker 접속이 바로 끊깁니다');

// 지워진 대상(404) = 타일을 닫는다 — 주소-5(없는 대상)
watch(() => (userQuery.error.value as { status?: number } | null)?.status, st => { if (st === 404) dropMissing(props.tile.key); }, { immediate: true });

const busy = ref(false);
const message = ref<string>();
const run = (job: () => Promise<unknown>, done?: () => void) => {
    busy.value = true;
    message.value = undefined;
    job().then(async () => { done?.(); await invalidateMember(queryClient); })
    .catch(err => message.value = err?.message || '요청에 실패했습니다')
    .finally(() => busy.value = false);
}
const patch = (p: AdminUserPatch) => run(() => adminPatchUser(id, p));
const stepLimit = (d: number) => {
    const n = (u.value?.workerLimit ?? 0) + d;
    if (n >= 0 && n <= 1000) patch({ workerLimit: n });
}

const suspend = () => {
    if (!u.value) return;
    const on = !u.value.suspendedAt;
    openDialog({
        type: 'warning',
        title: `${u.value.nickname} ${on ? '정지' : '정지 해제'}`,
        content: on ? ['로그인과 이 회원 장비의 worker 접속을 막습니다. 데이터는 남습니다.', '지금 열린 화면과 붙어 있는 worker 연결도 끊깁니다.'] : ['다시 로그인하고 worker가 접속할 수 있게 합니다.'],
        buttons: [
            { view: '취소', type: 'info', bindButton: 'Escape', action: () => dialogOpen.value = false },
            { view: on ? '정지' : '해제', type: 'warning', action: () => { dialogOpen.value = false; run(() => adminSuspendUser(id, on)); } },
        ],
    });
}
const remove = () => {
    if (!u.value) return;
    openDialog({
        type: 'error',
        title: `${u.value.nickname} 삭제`,
        content: ['회원과 그 회원의 폴더·스크립트·실행 기록·장비·공유를 모두 지웁니다. 되돌릴 수 없습니다.', '실행 중인 process가 있으면 지울 수 없습니다.'],
        buttons: [
            { view: '취소', type: 'info', bindButton: 'Escape', action: () => dialogOpen.value = false },
            { view: '삭제', type: 'error', action: () => { dialogOpen.value = false; run(() => adminDeleteUser(id), () => close(props.tile.key)); } },
        ],
    });
}
</script>

<template>
<div class="mb-body">
    <p v-if="userQuery.isError.value" class="mb-empty">회원을 찾을 수 없습니다(지워졌을 수 있습니다)</p>
    <template v-else-if="u">
        <div class="mb-fields">
            <span class="mb-lbl">상태</span>
            <span class="mb-tag" :class="state.cls">{{ state.text }}</span>
            <span class="mb-lbl">가입</span>
            <span class="val">{{ dayText(u.createdAt) }}</span>
            <span class="mb-lbl">장비 상한</span>
            <div class="mb-grp" role="group" aria-label="장비 상한">
                <button type="button" class="mb-btn step" :disabled="busy || u.workerLimit <= 0" aria-label="상한 줄이기" @click="stepLimit(-1)">−</button>
                <span class="mb-btn count">{{ u.deviceCount }} / {{ u.workerLimit }}</span>
                <button type="button" class="mb-btn step" :disabled="busy" aria-label="상한 늘리기" @click="stepLimit(1)">+</button>
            </div>
            <span class="mb-lbl">장비 등록</span>
            <div class="mb-grp" role="group" aria-label="장비 등록 권한">
                <button type="button" class="mb-btn" :class="{ 'mb-btn--on': canReg }" :aria-pressed="canReg" :disabled="busy" @click="!canReg && patch({ canRegister: true })">허용</button>
                <button type="button" class="mb-btn" :class="{ 'mb-btn--on': !canReg }" :aria-pressed="!canReg" :disabled="busy" @click="canReg && patch({ canRegister: false })">막음</button>
            </div>
            <span class="mb-lbl">관리자</span>
            <div class="mb-grp" role="group" aria-label="관리자 권한">
                <button type="button" class="mb-btn" :class="{ 'mb-btn--on': isAdmin }" :aria-pressed="isAdmin" :disabled="busy || me" @click="!isAdmin && patch({ admin: true })">예</button>
                <button type="button" class="mb-btn" :class="{ 'mb-btn--on': !isAdmin }" :aria-pressed="!isAdmin" :disabled="busy || me" @click="isAdmin && patch({ admin: false })">아니오</button>
            </div>
            <p v-if="message" class="mb-err full">{{ message }}</p>
        </div>
        <h3 class="mb-sec">장비 {{ devices.length }}</h3>
        <button v-for="d in devices" :key="d.mainKey" type="button" class="mb-row" @click="open(deviceKey(d.mainKey), tile.key)">
            <span class="mb-dot" :class="deviceState(d.instances)" />
            <span class="txt">
                <span class="main">{{ d.mainKey }}</span>
                <span class="sub">{{ d.instances.map(i => i.label).filter(Boolean).join(', ') || `인스턴스 ${d.instances.length}` }}</span>
            </span>
            <span class="mb-tag wall" :class="{ on: 'mb-tag--success', off: '', none: 'dim' }[deviceState(d.instances)]">{{ deviceStateText(d.instances) }}</span>
        </button>
        <p v-if="devicesQuery.isSuccess.value && !devices.length" class="mb-empty">등록한 장비가 없습니다</p>
    </template>
</div>
<div v-if="u" class="mb-foot">
    <span class="mb-note">{{ note }}</span>
    <div class="mb-end">
        <button type="button" class="mb-btn mb-btn--error" :disabled="busy || me" @click="suspend">{{ u.suspendedAt ? '정지 해제' : '정지' }}</button>
        <button type="button" class="mb-btn mb-btn--danger" :disabled="busy || me" @click="remove">삭제</button>
    </div>
</div>
</template>

<style scoped lang="scss">
.val { @include vars.font('body-medium'); }
.mb-grp { justify-self: start; }
.step { min-width: 40px; @include vars.font('title-small'); }
.count { min-width: 64px; cursor: default; background: rgba(var(--v-theme-on-surface), 0.12) !important; }
.full { grid-column: 1 / -1; }
</style>
