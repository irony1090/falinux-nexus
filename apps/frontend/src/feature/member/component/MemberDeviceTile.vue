<script setup lang="ts">
import { useAppDialog } from '@/feature/layout/store/appDialog.store';
import { lastSeenText } from '@/feature/node/util/presenceView.util';
import { deleteWorkerInstance, type WorkerInstanceResponse } from '@/feature/worker/api/worker.api';
import { useQueryClient } from '@tanstack/vue-query';
import { computed, ref, watch } from 'vue';
import { VIcon } from 'vuetify/components';
import {
    adminChangeOwner, adminDeleteDevice, adminLabelInstance, adminReissueKey, adminShowKey,
    deleteMyDevice, invalidateMember, labelMyInstance, reissueMyDeviceKey, showMyDeviceKey,
    useAdminUsers, useMemberDevice,
} from '../api/device.api';
import { useMemberBoard } from '../store/memberBoard.store';
import { mainKeyOf, type MemberTile } from '../util/memberGrid.util';
import { API_KEY_PREFIX, copyText, dayText, deviceState, deviceStateText, workerEnvLines } from '../util/member.util';

// 장비 타일 — 상태·주인(관리자)·키·env + 인스턴스 라벨 바로 고치기(화-15(라벨)) + 삭제. 내 장비·관리자 공용 — 회원-3·회원-5
const props = defineProps<{ tile: MemberTile }>();
const mainKey = mainKeyOf(props.tile.key);
const { admin, close, dropMissing, freshKeys } = useMemberBoard();
const { open: dialogOpen, openDialog } = useAppDialog();
const queryClient = useQueryClient();

const api = admin
    ? { showKey: adminShowKey, reissueKey: adminReissueKey, remove: adminDeleteDevice, label: adminLabelInstance }
    : { showKey: showMyDeviceKey, reissueKey: reissueMyDeviceKey, remove: deleteMyDevice, label: labelMyInstance };

const deviceQuery = useMemberDevice(mainKey, admin);
const d = computed(() => deviceQuery.data.value);
// 주인 후보 = 회원 전부(관리자만)
const ownersQuery = admin ? useAdminUsers({ size: 0 }) : undefined;
const owners = computed(() => ownersQuery?.data.value?.items ?? []);

const state = computed(() => d.value ? deviceState(d.value.instances) : 'none');
const stateCls = computed(() => ({ on: 'mb-tag--success', off: '', none: 'dim' })[state.value]);

// 지워진 대상(404) = 타일을 닫는다 — 주소-5(없는 대상)
watch(() => (deviceQuery.error.value as { status?: number } | null)?.status, st => { if (st === 404) dropMissing(props.tile.key); }, { immediate: true });

const busy = ref(false);
const message = ref<string>();
const run = <T,>(job: () => Promise<T>, done?: (v: T) => void) => {
    busy.value = true;
    message.value = undefined;
    return job().then(async v => { done?.(v); await invalidateMember(queryClient); })
    .catch(err => message.value = err?.message || '요청에 실패했습니다')
    .finally(() => busy.value = false);
}

// ---- 키 ----
const key = ref<string | undefined>(freshKeys.value[mainKey]);
const copied = ref<'key' | 'env'>();
const toggleKey = () => {
    if (key.value) { key.value = undefined; return; }
    busy.value = true;
    message.value = undefined;
    api.showKey(mainKey)
    .then(r => key.value = r.apiKey)
    .catch(err => message.value = err?.message || '키를 불러오지 못했습니다')
    .finally(() => busy.value = false);
}
const copy = async (what: 'key' | 'env') => {
    try {
        const k = key.value ?? (await api.showKey(mainKey)).apiKey;
        await copyText(what === 'key' ? k : workerEnvLines(mainKey, k));
        copied.value = what;
        setTimeout(() => copied.value = undefined, 1500);
    } catch (err: any) {
        message.value = err?.message || '복사하지 못했습니다';
    }
}
const changing = ref(false);
const pasted = ref('');
const pastedValid = computed(() => !pasted.value.trim() || pasted.value.trim().startsWith(API_KEY_PREFIX));
const changeKey = () => {
    const apiKey = pasted.value.trim();
    openDialog({
        type: 'warning',
        title: `${mainKey} 키 바꾸기`,
        content: [apiKey ? '붙여넣은 키로 바꿉니다.' : '새 키를 발급합니다.', '지금 붙어 있는 worker는 그대로지만, 다시 접속할 때는 새 키가 env에 있어야 합니다.'],
        buttons: [
            { view: '취소', type: 'info', bindButton: 'Escape', action: () => dialogOpen.value = false },
            {
                view: '바꾸기', type: 'warning', action: () => {
                    dialogOpen.value = false;
                    run(() => api.reissueKey(mainKey, apiKey), r => { key.value = r.apiKey; changing.value = false; pasted.value = ''; });
                },
            },
        ],
    });
}

// ---- 주인(관리자) ----
const changeOwner = (e: Event) => {
    const sel = e.target as HTMLSelectElement;
    const id = Number(sel.value);
    if (!d.value || id === d.value.ownerId) return;
    const to = owners.value.find(o => o.id === id)?.nickname ?? String(id);
    sel.value = String(d.value.ownerId);    // 확인 전엔 원래 값
    openDialog({
        type: 'warning',
        title: `${mainKey} 주인 바꾸기`,
        content: [`주인을 ${to}(으)로 바꿉니다.`, '옛 주인 폴더에 지정된 이 장비는 실행할 수 없게 됩니다.'],
        buttons: [
            { view: '취소', type: 'info', bindButton: 'Escape', action: () => dialogOpen.value = false },
            { view: '바꾸기', type: 'warning', action: () => { dialogOpen.value = false; run(() => adminChangeOwner(mainKey, id)); } },
        ],
    });
}

// ---- 인스턴스 라벨 — 칸에서 바로 고치고 Enter·포커스 이동 때 저장 ----
const drafts = ref<Record<string, string>>({});
const labelOf = (i: WorkerInstanceResponse) => drafts.value[i.subKey] ?? i.label ?? '';
const saveLabel = (i: WorkerInstanceResponse) => {
    const v = (drafts.value[i.subKey] ?? i.label ?? '').trim();
    const clear = () => { const { [i.subKey]: _, ...rest } = drafts.value; drafts.value = rest; };
    if (v === (i.label ?? '')) { clear(); return; }
    run(() => api.label(mainKey, i.subKey, v), clear);
}
const removeInstance = (i: WorkerInstanceResponse) => run(() => deleteWorkerInstance(mainKey, i.subKey));

const remove = () => openDialog({
    type: 'warning',
    title: `${mainKey} 삭제`,
    content: ['장비 이름과 키를 지웁니다. 이 이름의 worker는 더 이상 접속할 수 없습니다.', '접속 중인 인스턴스가 있으면 지울 수 없습니다.'],
    buttons: [
        { view: '취소', type: 'info', bindButton: 'Escape', action: () => dialogOpen.value = false },
        { view: '삭제', type: 'error', action: () => { dialogOpen.value = false; run(() => api.remove(mainKey), () => close(props.tile.key)); } },
    ],
});
</script>

<template>
<div class="mb-body">
    <p v-if="deviceQuery.isError.value" class="mb-empty">장비를 찾을 수 없습니다(지워졌을 수 있습니다)</p>
    <template v-else-if="d">
        <div class="mb-fields mb-fields--wall">
            <span class="mb-lbl">상태</span>
            <span class="mb-tag" :class="stateCls"><span class="mb-dot" :class="state" />{{ deviceStateText(d.instances) }}</span>
            <template v-if="admin">
                <span class="mb-lbl">주인</span>
                <label class="mb-select">
                    <select class="mb-input" :value="d.ownerId" :disabled="busy" aria-label="주인" @change="changeOwner">
                        <option v-if="!owners.some(o => o.id === d!.ownerId)" :value="d.ownerId">{{ d.ownerNickname ?? d.ownerId }}</option>
                        <option v-for="o in owners" :key="o.id" :value="o.id">{{ o.nickname }}</option>
                    </select>
                    <v-icon icon="mdi-chevron-down" size="18" class="chev" />
                </label>
            </template>
            <span class="mb-lbl">등록</span>
            <span class="val">{{ dayText(d.createdAt) }} · 키 발급 {{ dayText(d.keyIssuedAt) }}</span>
            <span class="mb-lbl">API 키</span>
            <div class="keyrow">
                <!-- 보기 = 고칠 수 없는 입력칸(한 줄, 커서로 끝까지 이동·선택) / 숨김 = 가림 글자. readonly는 Chromium에서 키보드 커서 이동이 안 돼서 입력만 막는다 -->
                <input v-if="key" class="mb-input mb-mono code" :value="key" inputmode="none" spellcheck="false" autocomplete="off" aria-readonly="true" aria-label="API 키" @beforeinput.prevent @drop.prevent />
                <code v-else class="mb-input mb-mono code hidden">{{ `${API_KEY_PREFIX}${'•'.repeat(18)}` }}</code>
                <div class="mb-grp mb-grp--sm">
                    <button type="button" class="mb-btn" :disabled="busy" @click="toggleKey">{{ key ? '숨기기' : '보기' }}</button>
                    <button type="button" class="mb-btn" :disabled="busy" @click="copy('key')">{{ copied === 'key' ? '복사됨' : '복사' }}</button>
                    <button type="button" class="mb-btn mb-btn--accent" :disabled="busy" :aria-pressed="changing" @click="changing = !changing">바꾸기</button>
                </div>
            </div>
            <div v-if="changing" class="full change">
                <input v-model="pasted" class="mb-input mb-mono" :class="{ bad: !pastedValid }" placeholder="비우면 새로 발급 · 다른 supervisor 키는 붙여넣기" autocomplete="off" :disabled="busy" />
                <div class="mb-grp mb-grp--sm">
                    <button type="button" class="mb-btn mb-btn--primary" :disabled="busy || !pastedValid" @click="changeKey">{{ pasted.trim() ? '이 키로' : '새로 발급' }}</button>
                </div>
            </div>
            <div v-if="key" class="full env">
                <pre class="mb-mono">{{ workerEnvLines(mainKey, key) }}</pre>
                <button type="button" class="envcopy" @click="copy('env')">{{ copied === 'env' ? '복사됨' : 'env 두 줄 복사' }}</button>
            </div>
            <p v-if="message" class="mb-err full">{{ message }}</p>
        </div>
        <h3 class="mb-sec">인스턴스 {{ d.instances.length }}</h3>
        <div v-for="i in d.instances" :key="i.subKey" class="inst">
            <div class="line">
                <span class="mb-dot" :class="i.online ? 'on' : 'off'" />
                <input class="mb-input" :value="labelOf(i)" maxlength="64" placeholder="라벨 달기 (예: 3번 라인 PC)" aria-label="라벨" :disabled="busy"
                    @input="drafts = { ...drafts, [i.subKey]: ($event.target as HTMLInputElement).value }"
                    @keydown.enter="($event.target as HTMLInputElement).blur()"
                    @blur="saveLabel(i)"
                />
                <button v-if="!admin && !i.online" type="button" class="mb-btn mb-btn--error rm" :disabled="busy" @click="removeInstance(i)">지우기</button>
            </div>
            <span class="meta"><code class="mb-mono">{{ i.subKey }}</code> · {{ i.os }} · {{ lastSeenText(i.online, i.lastSeen) }} · 처음 {{ dayText(i.firstSeen) }}</span>
        </div>
        <p v-if="!d.instances.length" class="mb-empty">아직 접속한 worker가 없습니다.<br>API 키의 env 두 줄을 worker에 넣고 실행하면 여기에 나타납니다.</p>
    </template>
</div>
<div v-if="d" class="mb-foot">
    <span class="mb-note">키를 바꾸면 옛 키의 worker는 다음 접속부터 거부됩니다</span>
    <div class="mb-end">
        <button type="button" class="mb-btn mb-btn--danger" :disabled="busy" @click="remove">삭제</button>
    </div>
</div>
</template>

<style scoped lang="scss">
.val { padding-left: 14px; @include vars.font('body-medium'); }
// 키 칸 + [보기|복사|바꾸기] = 한 덩어리, 오른쪽 벽에 붙음
.keyrow {
    min-width: 0;
    display: flex;
    align-items: center;
    .code { flex: 1; min-width: 0; display: flex; align-items: center; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; border-radius: 999px 0 0 999px; border-right: 0;
        &.hidden { opacity: 0.6; user-select: none; } }
    .mb-grp { height: 32px; border-radius: 0; }
}
.change {
    display: flex;
    align-items: center;
    gap: 6px;
    .mb-grp { height: 32px; }
}
.env {
    position: relative;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.16);
    border-radius: 10px;
    background: rgba(var(--v-theme-on-surface), 0.035);
    overflow: hidden;
    pre { margin: 0; padding: 12px 14px; line-height: 1.6; white-space: pre; overflow-x: auto; }
}
// 상자 오른쪽 위 모서리에 붙는 끝 블록(찰떡) — 곡선은 상자가 자른다
.envcopy {
    position: absolute;
    top: 0;
    right: 0;
    height: 26px;
    padding: 0 12px 0 14px;
    border: 0;
    border-radius: 0 0 0 999px;
    background: rgb(var(--v-theme-primary));
    color: rgb(var(--v-theme-on-primary));
    font: inherit;
    @include vars.font('label-medium');
    font-weight: vars.$weight-md;
    cursor: pointer;
}
.inst {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 10px 12px;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    .line { display: flex; align-items: center; gap: 8px; }
    .rm { height: 32px; border-radius: 999px; padding: 0 13px; @include vars.font('label-medium'); font-weight: vars.$weight-md; }
    .meta { padding-left: 16px; @include vars.font('body-small'); color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity)); code { font-size: 11px; } }
}
</style>
