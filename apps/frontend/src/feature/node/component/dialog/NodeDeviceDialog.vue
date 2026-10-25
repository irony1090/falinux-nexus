<script setup lang="ts">
import { useAppDialog } from '@/feature/layout/store/appDialog.store';
import { deleteWorkerInstance, invalidateWorkerRoster, useWorkerRoster, type WorkerInstanceResponse } from '@/feature/worker/api/worker.api';
import { useQueryClient } from '@tanstack/vue-query';
import { computed, ref, watch } from 'vue';
import { VCard, VCardText, VCardTitle, VDialog, VForm, VIcon, VTextField } from 'vuetify/components';
import { osBadge, osTitle } from '@/feature/worker/util/workerOs.util';
import { agoText } from '../../util/presenceView.util';

// 폴더의 장비(main_key) 지정 — 직접 입력 + 명부에서 고르기 + 끊긴 인스턴스 정리. worker 접속 전에도 미리 지정할 수 있다 — RS5(표시 위치) ②
const props = defineProps({
    title: {
        type: String,
        required: true,
    },
    initial: {
        type: String,
        default: () => '',
    },
    loading: {
        type: Boolean,
        default: () => false,
    },
})
const emit = defineEmits<{ submit: [deviceKey: string | null] }>();
const open = defineModel<boolean>({ type: Boolean, default: () => false });

const { devices } = useWorkerRoster();
const queryClient = useQueryClient();
const { open: appDialogOpen, openDialog } = useAppDialog();

const key = ref<string | null>('');     // clearable이 null을 넣는다
const trimmed = computed(() => (key.value ?? '').trim());
const list = computed(() => [...devices.value.values()].sort((a, b) => a.mainKey.localeCompare(b.mainKey)));
const expanded = ref<Record<string, boolean>>({});

const status = computed(() => {
    if (!trimmed.value) return '지정 안 함 — 상위 폴더의 장비를 따릅니다';
    const d = devices.value.get(trimmed.value);
    if (!d) return '명부에 없는 장비 — 이 키로 worker가 접속하면 연결됩니다';
    return d.online ? `접속 중인 인스턴스 ${d.online}개 / 명부 ${d.instances.length}개` : `접속 중인 worker 없음 / 명부 ${d.instances.length}개`;
})

// 열릴 때마다 현재 값으로 되돌리고, 그 장비를 펼친다
watch(open, val => {
    if (!val) return;
    key.value = props.initial;
    expanded.value = props.initial ? { [props.initial]: true } : {};
})

const removing = ref(false);
const remove = (inst: WorkerInstanceResponse) => {
    // 키는 title로만 넘긴다 — AppDialog의 content는 v-html로 그려진다
    openDialog({
        type: 'warning',
        title: `'${inst.instanceKey}' 명부에서 삭제`,
        content: [`마지막 접속 ${agoText(inst.lastSeen)}`, '이 인스턴스가 다시 접속하면 명부에 다시 올라옵니다. 실행 중인 터미널에는 영향이 없습니다.'],
        buttons: [
            { view: '취소', type: 'info', bindButton: 'Escape', action: () => appDialogOpen.value = false },
            {
                view: '삭제', type: 'error',
                action: () => {
                    appDialogOpen.value = false;
                    removing.value = true;
                    deleteWorkerInstance(inst.mainKey, inst.subKey)
                    .catch(err => openDialog({ type: 'error', content: err?.message || '삭제에 실패했습니다' }))
                    .finally(() => {
                        removing.value = false;
                        invalidateWorkerRoster(queryClient);
                    });
                },
            },
        ],
    })
}

const onSubmit = () => emit('submit', trimmed.value || null);
</script>

<template>
<v-dialog v-model="open" max-width="440" min-width="296" :persistent="loading">
    <v-card :loading="loading || removing" :disabled="loading || removing" class="ct-card">
        <v-card-title>{{ title }}</v-card-title>
        <v-form @submit.prevent="onSubmit">
            <v-card-text>
                <v-text-field v-model="key"
                    label="장비 키 (main_key)"
                    autofocus
                    clearable
                    persistent-hint
                    :hint="status"
                />
                <div class="roster">
                    <span class="roster__label">알려진 장비 (명부)</span>
                    <div v-if="list.length" class="roster__box">
                        <template v-for="d in list" :key="d.mainKey">
                            <div class="dev" :class="{ sel: d.mainKey === trimmed }">
                                <button type="button" class="dev__pick" title="이 장비로 지정" @click="key = d.mainKey">
                                    <i class="dot" :class="{ on: d.online > 0 }" />
                                    <span class="mono">{{ d.mainKey }}</span>
                                </button>
                                <!-- 찰떡: 행 끝 묶음(오른쪽 = 명부 벽 쪽 곧게) -->
                                <span class="ct-grp ct-grp--r">
                                    <span class="ct-seg ct-seg--txt cnt" :class="{ on: d.online > 0 }">접속 {{ d.online }} / {{ d.instances.length }}</span>
                                    <button type="button" class="ct-seg ct-seg--icon" :aria-label="expanded[d.mainKey] ? '접기' : '인스턴스 펼치기'" @click="expanded[d.mainKey] = !expanded[d.mainKey]">
                                        <v-icon size="16" :icon="expanded[d.mainKey] ? 'mdi-chevron-down' : 'mdi-chevron-right'" />
                                    </button>
                                </span>
                            </div>
                            <template v-if="expanded[d.mainKey]">
                                <div v-for="inst in d.instances" :key="inst.subKey" class="inst" :class="{ off: !inst.online }">
                                    <i class="dot" :class="{ on: inst.online }" />
                                    <span class="mono inst__os" :title="osTitle(inst.os)">{{ osBadge(inst.os) }}</span>
                                    <span class="mono inst__key">#{{ inst.subKey }}</span>
                                    <span v-if="inst.online" class="inst__seen on">접속 중 · 처음 {{ new Date(inst.firstSeen * 1000).toLocaleDateString() }}</span>
                                    <template v-else>
                                        <span class="inst__seen">{{ agoText(inst.lastSeen) }}</span>
                                        <span class="ct-grp ct-grp--r">
                                            <button type="button" class="ct-seg ct-seg--error" @click="remove(inst)">삭제</button>
                                        </span>
                                    </template>
                                </div>
                            </template>
                        </template>
                    </div>
                    <span v-else class="roster__none">없음 — 접속한 worker가 아직 없습니다</span>
                </div>
            </v-card-text>
            <div class="ct-actions">
                <span class="ct-grp ct-grp--r">
                    <button type="button" class="ct-seg" @click="open = false">취소</button>
                    <button type="submit" class="ct-seg ct-seg--primary">확인</button>
                </span>
            </div>
        </v-form>
    </v-card>
</v-dialog>
</template>

<style scoped lang="scss">
.roster {
    display: flex;
    flex-direction: column;
    gap: vars.$spacing-sm;
    margin-top: vars.$spacing-lg;
    @include vars.font('body-medium');
    button:where(:not(.v-btn, .ct-seg)) {    // scoped 규칙은 자식 v-btn 루트에도 걸린다 — 여백·글자를 지우지 않게(:where = 우선순위 그대로)
        padding: 0;
        border: 0;
        background: none;
        color: inherit;
        font: inherit;
        cursor: pointer;
    }
}
.roster__label,
.roster__none { opacity: 0.7; }
.roster__box {
    max-height: 240px;
    overflow: auto;
}
.mono { font-family: monospace; }
.dot {
    flex: none;
    width: 7px;
    height: 7px;
    border-radius: 50%;
    border: 1.5px solid rgba(var(--v-theme-on-surface), var(--v-disabled-opacity));
    &.on { border-color: rgb(var(--v-theme-success)); background: rgb(var(--v-theme-success)); }
}
.dev {
    display: flex;
    align-items: center;
    gap: vars.$spacing-sm;
    min-height: 36px;
    padding-left: 6px;
    border-radius: 6px 0 0 6px;
    &.sel { background: rgba(var(--v-theme-primary), 0.08); }
}
.dev__pick {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: vars.$spacing-sm;
    text-align: left;
    span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    &:hover { color: rgb(var(--v-theme-primary)); }
}
.cnt.on { color: rgb(var(--v-theme-success)); }
.inst {
    display: flex;
    align-items: center;
    gap: vars.$spacing-sm;
    min-height: 32px;
    padding-left: 22px;
    @include vars.font('body-small');
    &.off .inst__key { opacity: 0.7; }
}
.inst__key { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; }
.inst__os { font-weight: 600; }
.inst__seen {
    opacity: 0.7;
    white-space: nowrap;
    &.on { color: rgb(var(--v-theme-success)); opacity: 1; }
}
</style>
