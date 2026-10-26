<script setup lang="ts">
import { useNodeOps } from '@/feature/node/store/nodeOps.store';
import { ref, watch } from 'vue';
import { VCard, VCardText, VCardTitle, VDialog } from 'vuetify/components';
import { killUnregistered, listUnregistered, type UnregisteredProcess } from '../api/worker.api';

// 미등록 process — 주소-5(화면), REF-supervisor-move.md. 이 supervisor DB에 기록이 없는데 장비에서 살아 있는 process(다른 supervisor에서 띄운 것 등).
// 자동으로 끄지 않고 여기서 보고 끄거나, 실행(EXEC)은 가져와 내 터미널 타일로(가져오기-4(화면)). 목록은 열 때·끈 뒤에 다시 불러온다
const props = defineProps<{ tileId: string; instance: string; folderId: number }>();
const open = defineModel<boolean>({ type: Boolean, default: () => false });

// 끈 뒤 worker가 끝남을 알릴 때까지 잠깐 기다렸다 다시 읽는다
const RELOAD_AFTER_KILL_MS = 800;

const { adopt: adoptOp } = useNodeOps();
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

// 가져오면 장비 타일 아래 터미널 타일이 뜨므로 창은 닫는다
const adopt = (p: UnregisteredProcess) => {
    busy.value = p.uid;
    message.value = undefined;
    adoptOp(props.tileId, props.folderId, props.instance, p.uid)
    .then(() => open.value = false)
    .catch(err => {
        message.value = err?.message || '가져오기에 실패했습니다';
        return load();
    })
    .finally(() => busy.value = undefined);
}

const startedText = (s: number) => s ? new Date(s * 1000).toLocaleString() : '시작 시각 모름';
// 둘째 줄 = 이름이 있으면 실제 명령도, 그다음 폴더·pid·시작 시각
const subText = (p: UnregisteredProcess) => [p.label ? p.cmd : '', p.cwd, `pid ${p.pid || '-'}`, startedText(p.startedAt)].filter(Boolean).join(' · ');
</script>

<template>
<v-dialog v-model="open" max-width="560" min-width="296">
    <v-card :loading="loading" class="ct-card">
        <v-card-title>미등록 process</v-card-title>
        <v-card-text>
            <p class="note mono">{{ instance }}</p>
            <p class="note">이 supervisor에 기록이 없는 process입니다. 가져오면 내 터미널 타일이 되고, 다른 supervisor로 돌아가면 그쪽 기록으로 이어집니다</p>
            <p v-if="list && !list.length" class="note">미등록 process가 없습니다</p>
            <div v-for="p in list" :key="p.uid" class="row">
                <div class="info">
                    <div class="cmd mono">
                        <span class="tag">{{ p.kind === 'EDIT' ? '편집' : '실행' }}</span>
                        {{ p.label || p.cmd || '(명령 모름)' }}
                    </div>
                    <div class="sub mono">{{ subText(p) }}</div>
                </div>
                <!-- 찰떡: 행 끝 묶음(오른쪽 = 창 안쪽 벽 쪽 곧게). 편집은 결과를 저장할 곳이 없어 가져오기 없음 -->
                <span class="ct-grp ct-grp--r">
                    <button v-if="p.kind !== 'EDIT'" type="button" class="ct-seg ct-seg--accent" :disabled="!!busy" title="내 터미널 타일로 가져오기" @click="adopt(p)">가져오기</button>
                    <button type="button" class="ct-seg ct-seg--error" :disabled="!!busy" @click="kill(p, false)">끄기</button>
                    <button type="button" class="ct-seg ct-seg--txt" :disabled="!!busy" title="응답 없을 때 강제 종료" @click="kill(p, true)">강제</button>
                </span>
            </div>
            <p v-if="message" class="note err">{{ message }}</p>
        </v-card-text>
        <div class="ct-actions">
            <span class="ct-grp ct-grp--r">
                <button type="button" class="ct-seg" :disabled="loading" @click="load">새로 고침</button>
                <button type="button" class="ct-seg ct-seg--primary" @click="open = false">닫기</button>
            </span>
        </div>
    </v-card>
</v-dialog>
</template>

<style scoped lang="scss">
.note { margin: 0 0 vars.$spacing-sm; @include vars.font('body-medium'); opacity: 0.7; &.err { color: rgb(var(--v-theme-error)); opacity: 1; } }
.mono { font-family: monospace; }
.row { display: flex; align-items: center; gap: vars.$spacing-sm; min-height: 44px; padding: vars.$spacing-xs 0; border-top: 1px solid rgba(var(--v-border-color), var(--v-border-opacity)); }
.info { flex: 1; min-width: 0; }
.cmd { @include vars.font('body-medium'); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.sub { @include vars.font('body-small'); opacity: 0.7; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.tag { @include vars.font('label-small'); padding: 0 6px; margin-right: 4px; border-radius: 999px; background: rgba(var(--v-theme-on-surface), 0.08); }
</style>
