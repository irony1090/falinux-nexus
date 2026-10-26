<script setup lang="ts">
import { useQueryClient } from '@tanstack/vue-query';
import { computed, ref, watch } from 'vue';
import { instanceAddrs, invalidateMember, moveDevices, type DeviceMoveResult, type DeviceResponse } from '../api/device.api';
import { useMemberBoard } from '../store/memberBoard.store';
import { instanceText } from '../util/member.util';

// 열린 장비 타일의 supervisor 주소 일괄 변경 — instance를 골라 보냄. 1개 = 저장된 주소 순위 / 여럿 = 바뀔 이름 (REF-device-move.md)
const props = defineProps<{ devices: DeviceResponse[] }>();
const open = defineModel<boolean>({ type: Boolean, default: () => false });
const { admin } = useMemberBoard();
const queryClient = useQueryClient();

const scheme = ref<'ws' | 'wss'>('ws');
const host = ref('');
const now = ref(false);
const busy = ref(false);
const message = ref<string>();
const done = ref<string>();
const picked = ref<string[]>([]);
const addrs = ref<string[]>();
const results = ref<Record<string, DeviceMoveResult>>({});

const rows = computed(() => props.devices.flatMap(d => d.instances.map(i => ({ ...i, text: instanceText(i.subKey, i.label) }))));
const pickedRows = computed(() => rows.value.filter(r => picked.value.includes(r.instanceKey)));
const one = computed(() => pickedRows.value.length === 1 ? pickedRows.value[0] : undefined);

watch(open, val => {
    if (!val) return;
    host.value = '';
    now.value = false;
    message.value = undefined;
    done.value = undefined;
    results.value = {};
    picked.value = rows.value.filter(r => r.online).map(r => r.instanceKey);   // 기본 = 접속 중 전부
});

const toggle = (key: string) => {
    picked.value = picked.value.includes(key) ? picked.value.filter(k => k !== key) : [...picked.value, key];
};

// 하나만 고르면 그 worker의 주소 순위를 읽는다
watch(() => open.value && one.value?.instanceKey, key => {
    addrs.value = undefined;
    if (!key || !one.value) return;
    instanceAddrs(admin, one.value.mainKey, one.value.subKey)
    .then(r => { if (one.value?.instanceKey === key) addrs.value = r.addrs; })
    .catch(err => message.value = err?.message || '주소 목록을 못 읽었습니다');
}, { immediate: true });

const submit = () => {
    const h = host.value.trim();
    if (!h) {
        message.value = '주소(호스트:포트) 입력이 필수입니다';
        return;
    }
    busy.value = true;
    message.value = undefined;
    done.value = undefined;
    moveDevices(admin, picked.value, `${scheme.value}://${h}`, now.value)
    .then(async r => {
        results.value = Object.fromEntries(r.map(x => [x.instanceKey, x]));
        const failed = r.filter(x => x.error).length;
        done.value = `${r.length - failed}개 ${now.value ? '옮김' : '저장'}` + (failed ? ` · ${failed}개 실패` : '');
        if (one.value && !results.value[one.value.instanceKey]?.error) addrs.value = results.value[one.value.instanceKey]?.addrs;
        await invalidateMember(queryClient);
    })
    .catch(err => message.value = err?.message || '주소 변경에 실패했습니다')
    .finally(() => busy.value = false);
};
</script>

<template>
<v-dialog v-model="open" max-width="480" min-width="296" :persistent="busy">
    <v-card :loading="busy" class="ct-card">
        <v-card-title>supervisor 주소 변경</v-card-title>
        <v-card-text class="body">
            <p class="note">열린 장비 타일의 instance 중 고른 것에 보냅니다. 접속 중인 것만 고를 수 있습니다</p>
            <div class="targets">
                <p v-if="!devices.length" class="mb-empty">열린 장비 타일이 없습니다</p>
                <template v-for="d in devices" :key="d.mainKey">
                    <p class="mb-sec mb-mono dev">{{ d.mainKey }}</p>
                    <p v-if="!d.instances.length" class="none">접속 기록 없음</p>
                    <label v-for="i in d.instances" :key="i.instanceKey" class="inst" :class="{ off: !i.online }">
                        <input type="checkbox" :checked="picked.includes(i.instanceKey)" :disabled="!i.online || busy" @change="toggle(i.instanceKey)" />
                        <span class="mb-dot" :class="i.online ? 'on' : 'off'" />
                        <span class="name">{{ instanceText(i.subKey, i.label) }}</span>
                        <span v-if="results[i.instanceKey]?.error" class="mb-tag mb-tag--error">{{ results[i.instanceKey].error }}</span>
                        <span v-else-if="results[i.instanceKey]" class="mb-tag mb-tag--success mb-mono">{{ results[i.instanceKey].addrs[0] }}</span>
                        <span v-else-if="!i.online" class="mb-tag dim">접속 안 됨</span>
                    </label>
                </template>
            </div>

            <form class="addr" @submit.prevent="submit">
                <label class="mb-select scheme">
                    <select v-model="scheme" class="mb-input" aria-label="방식" :disabled="busy">
                        <option value="ws">ws</option>
                        <option value="wss">wss</option>
                    </select>
                    <v-icon icon="mdi-chevron-down" size="18" class="chev" />
                </label>
                <input v-model="host" class="mb-input mb-mono" :class="{ bad: !!message }" placeholder="호스트:포트 (예: example.com:8012)" aria-label="호스트:포트"
                    autocomplete="off" autofocus :disabled="busy" @input="message = undefined" />
            </form>
            <div class="mb-grp mb-grp--sm">
                <button type="button" class="mb-btn" :class="{ 'mb-btn--on': !now }" :aria-pressed="!now" :disabled="busy" @click="now = false">저장만</button>
                <button type="button" class="mb-btn" :class="{ 'mb-btn--on': now }" :aria-pressed="now" :disabled="busy" @click="now = true">바로 옮기기</button>
            </div>
            <p class="note">{{ now ? '지금 연결을 끊고 새 주소로 붙습니다. 못 붙으면 옛 주소로 돌아옵니다' : '지금 연결은 그대로 두고, 이 supervisor가 끊기면 새 주소부터 시도합니다' }}</p>

            <template v-if="one">
                <p class="mb-lbl">저장된 주소(순위) · {{ one.text }}</p>
                <ol v-if="addrs" class="ranks mb-mono">
                    <li v-for="a in addrs" :key="a">{{ a }}</li>
                </ol>
            </template>
            <template v-else-if="pickedRows.length > 1">
                <p class="mb-lbl">바뀔 instance {{ pickedRows.length }}개</p>
                <p class="names mb-mono">{{ pickedRows.map(r => `${r.mainKey} · ${r.text}`).join('\n') }}</p>
            </template>
            <p v-if="done" class="note ok">{{ done }}</p>
            <p v-if="message" class="mb-err">{{ message }}</p>
        </v-card-text>
        <div class="ct-actions">
            <span class="ct-grp ct-grp--r">
                <button type="button" class="ct-seg" :disabled="busy" @click="open = false">닫기</button>
                <button type="button" class="ct-seg ct-seg--primary" :disabled="busy || !picked.length" @click="submit">{{ now ? '옮기기' : '저장' }}</button>
            </span>
        </div>
    </v-card>
</v-dialog>
</template>

<style scoped lang="scss">
.body { display: flex; flex-direction: column; gap: vars.$spacing-md; }
.note { margin: 0; @include vars.font('body-medium'); opacity: 0.7; &.ok { color: rgb(var(--v-theme-success)); opacity: 1; } }
.targets { max-height: 240px; overflow-y: auto; border-radius: 12px; border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity)); }
.dev { border-top: 0; }
.none { margin: 0; padding: 8px 12px; @include vars.font('body-small'); opacity: 0.7; }
.inst {
    display: flex;
    align-items: center;
    gap: 8px;
    min-height: 40px;
    padding: 4px 12px;
    cursor: pointer;
    &.off { cursor: default; opacity: 0.6; }
    > input { flex: none; margin: 0; }
    .name { flex: 1; min-width: 0; @include vars.font('body-medium'); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .mb-tag { max-width: 50%; overflow: hidden; text-overflow: ellipsis; }
}
.addr { display: flex; gap: vars.$spacing-sm; }
.scheme { flex: 0 0 92px; }
.ranks { margin: 0; padding-left: 22px; }
.names { margin: 0; white-space: pre-line; }
</style>
