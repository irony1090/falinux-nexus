<script setup lang="ts">
import { ref, watch } from 'vue';
import { VBtn, VCard, VCardActions, VCardText, VCardTitle, VDialog, VSpacer } from 'vuetify/components';
import { killUnregistered, listUnregistered, type UnregisteredProcess } from '../api/worker.api';

// 미등록 process — 주소-5(화면), REF-supervisor-move.md. 이 supervisor DB에 기록이 없는데 장비에서 살아 있는 process(다른 supervisor에서 띄운 것 등).
// 자동으로 끄지 않고 여기서 보고 끈다(주소-Q12). 목록은 열 때·끈 뒤에 다시 불러온다
const props = defineProps<{ instance: string; folderId: number }>();
const open = defineModel<boolean>({ type: Boolean, default: () => false });

// 끈 뒤 worker가 끝남을 알릴 때까지 잠깐 기다렸다 다시 읽는다
const RELOAD_AFTER_KILL_MS = 800;

const list = ref<UnregisteredProcess[]>();
const loading = ref(false);
const busy = ref<string>();
const message = ref<string>();

const load = () => {
    loading.value = true;
    return listUnregistered(props.instance, props.folderId)
    .then(r => list.value = r)
    .catch(err => message.value = err?.message || '목록을 못 읽었습니다')
    .finally(() => loading.value = false);
}

watch(open, val => {
    if (!val) return;
    list.value = undefined;
    message.value = undefined;
    load();
})

const kill = (p: UnregisteredProcess, force: boolean) => {
    busy.value = p.uid;
    message.value = undefined;
    killUnregistered(props.instance, props.folderId, p.uid, force)
    .then(() => new Promise(r => setTimeout(r, RELOAD_AFTER_KILL_MS)))
    .then(load)
    .catch(err => message.value = err?.message || '끄기에 실패했습니다')
    .finally(() => busy.value = undefined);
}

const startedText = (s: number) => s ? new Date(s * 1000).toLocaleString() : '시작 시각 모름';
// 둘째 줄 = 이름이 있으면 실제 명령도, 그다음 폴더·pid·시작 시각
const subText = (p: UnregisteredProcess) => [p.label ? p.cmd : '', p.cwd, `pid ${p.pid || '-'}`, startedText(p.startedAt)].filter(Boolean).join(' · ');
</script>

<template>
<v-dialog v-model="open" max-width="520" min-width="296">
    <v-card :loading="loading">
        <v-card-title>미등록 process</v-card-title>
        <v-card-text>
            <p class="note mono">{{ instance }}</p>
            <p class="note">이 supervisor에 기록이 없는 process입니다. 다른 supervisor에서 띄웠다면 그쪽으로 돌아가면 다시 이어집니다</p>
            <p v-if="list && !list.length" class="note">미등록 process가 없습니다</p>
            <div v-for="p in list" :key="p.uid" class="row">
                <div class="info">
                    <div class="cmd mono">
                        <span class="tag">{{ p.kind === 'EDIT' ? '편집' : '실행' }}</span>
                        {{ p.label || p.cmd || '(명령 모름)' }}
                    </div>
                    <div class="sub mono">{{ subText(p) }}</div>
                </div>
                <v-btn size="x-small" variant="tonal" color="error" :loading="busy === p.uid" :disabled="!!busy" @click="kill(p, false)">끄기</v-btn>
                <v-btn size="x-small" variant="text" color="error" :disabled="!!busy" title="응답 없을 때 강제 종료" @click="kill(p, true)">강제</v-btn>
            </div>
            <p v-if="message" class="note err">{{ message }}</p>
        </v-card-text>
        <v-card-actions>
            <v-btn variant="text" :disabled="loading" @click="load">새로 고침</v-btn>
            <v-spacer />
            <v-btn variant="text" @click="open = false">닫기</v-btn>
        </v-card-actions>
    </v-card>
</v-dialog>
</template>

<style scoped lang="scss">
.note { margin: 0 0 vars.$spacing-sm; @include vars.font('body-medium'); opacity: 0.7; &.err { color: rgb(var(--v-theme-error)); opacity: 1; } }
.mono { font-family: monospace; }
.row { display: flex; align-items: center; gap: vars.$spacing-sm; padding: vars.$spacing-sm 0; border-top: 1px solid rgba(var(--v-border-color), var(--v-border-opacity)); }
.info { flex: 1; min-width: 0; }
.cmd { @include vars.font('body-medium'); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sub { @include vars.font('body-small'); opacity: 0.7; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.tag { @include vars.font('label-small'); padding: 0 6px; margin-right: 4px; border-radius: 8px; background: rgba(var(--v-theme-on-surface), 0.08); }
</style>
