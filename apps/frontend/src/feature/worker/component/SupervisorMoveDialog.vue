<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { VBtn, VBtnToggle, VCard, VCardActions, VCardText, VCardTitle, VDialog, VForm, VSelect, VSpacer, VTextField } from 'vuetify/components';
import { listWorkerAddrs, moveAllWorkers, moveWorker, type MoveAllResult } from '../api/worker.api';

// supervisor 주소 변경 — 주소-5(화면), REF-supervisor-move.md. instance가 있으면 그 장비 하나(장비 타일 메뉴), 없으면 접속 중 전부(헤더 메뉴, worker_move 권한)
// worker는 새 주소를 1순위로 저장하고, 못 붙으면 다음 순위(옛 주소)로 돌아간다
const props = defineProps<{ instance?: string; folderId?: number }>();
const open = defineModel<boolean>({ type: Boolean, default: () => false });

const single = computed(() => !!props.instance && props.folderId !== undefined);

const scheme = ref<'ws' | 'wss'>('ws');
const host = ref('');
const now = ref(false);
const loading = ref(false);
const message = ref<string>();
const addrs = ref<string[]>();
const results = ref<MoveAllResult[]>();
const done = ref<string>();

const load = () => {
    if (!single.value) return;
    listWorkerAddrs(props.instance!, props.folderId!)
    .then(r => addrs.value = r.addrs)
    .catch(err => message.value = err?.message || '주소 목록을 못 읽었습니다');
}

watch(open, val => {
    if (!val) return;
    host.value = '';
    now.value = false;
    message.value = undefined;
    addrs.value = undefined;
    results.value = undefined;
    done.value = undefined;
    load();
})

const onSubmit = () => {
    const h = host.value.trim();
    if (!h) {
        message.value = '주소(호스트:포트) 입력이 필수입니다';
        return;
    }
    const url = `${scheme.value}://${h}`;
    loading.value = true;
    message.value = undefined;
    done.value = undefined;
    const job = single.value
        ? moveWorker(props.instance!, props.folderId!, url, now.value).then(r => {
            addrs.value = r.addrs;
            done.value = now.value ? '옮겼습니다. 이 supervisor에서는 이 장비가 끊김으로 보입니다' : '저장했습니다. 지금 supervisor가 끊기면 새 주소로 붙습니다';
        })
        : moveAllWorkers(url, now.value).then(r => {
            results.value = r;
            const failed = r.filter(x => x.error).length;
            done.value = `${r.length - failed}대 ${now.value ? '옮김' : '저장'}` + (failed ? ` · ${failed}대 실패` : '');
        });
    job.catch(err => message.value = err?.message || '주소 변경에 실패했습니다')
    .finally(() => loading.value = false);
}
</script>

<template>
<v-dialog v-model="open" max-width="460" min-width="296" :persistent="loading">
    <v-card :loading="loading" :disabled="loading">
        <v-card-title>{{ single ? 'supervisor 주소 변경' : '전체 worker 주소 변경' }}</v-card-title>
        <v-form @submit.prevent="onSubmit">
            <v-card-text class="body">
                <p v-if="single" class="note mono">{{ instance }}</p>
                <p v-else class="note">지금 접속 중인 worker 전부에 보냅니다</p>
                <div class="addr">
                    <v-select v-model="scheme" :items="['ws', 'wss']" density="compact" hide-details class="scheme" />
                    <v-text-field v-model="host"
                        label="호스트:포트"
                        placeholder="example.com:8012"
                        density="compact"
                        autofocus
                        :error-messages="message"
                        @update:model-value="message = undefined"
                    />
                </div>
                <v-btn-toggle v-model="now" mandatory density="compact" variant="outlined" color="primary" divided>
                    <v-btn :value="false" size="small">저장만</v-btn>
                    <v-btn :value="true" size="small">바로 옮기기</v-btn>
                </v-btn-toggle>
                <p class="note">{{ now ? '지금 연결을 끊고 새 주소로 붙습니다. 못 붙으면 옛 주소로 돌아옵니다' : '지금 연결은 그대로 두고, 이 supervisor가 끊기면 새 주소부터 시도합니다' }}</p>
                <p v-if="done" class="note ok">{{ done }}</p>
                <template v-if="addrs">
                    <p class="label">저장된 주소(순위)</p>
                    <ol class="list mono">
                        <li v-for="a in addrs" :key="a">{{ a }}</li>
                    </ol>
                </template>
                <template v-if="results?.length">
                    <p class="label">결과</p>
                    <div v-for="r in results" :key="r.instanceKey" class="res mono">
                        <span class="key">{{ r.instanceKey }}</span>
                        <span v-if="r.error" class="err">{{ r.error }}</span>
                        <span v-else class="ok">{{ r.addrs[0] }}</span>
                    </div>
                </template>
            </v-card-text>
            <v-card-actions>
                <v-spacer />
                <v-btn variant="text" @click="open = false">닫기</v-btn>
                <v-btn type="submit" variant="flat" color="primary">{{ now ? '옮기기' : '저장' }}</v-btn>
            </v-card-actions>
        </v-form>
    </v-card>
</v-dialog>
</template>

<style scoped lang="scss">
.body { display: flex; flex-direction: column; gap: vars.$spacing-md; }
.addr { display: flex; gap: vars.$spacing-sm; align-items: flex-start; }
.scheme { flex: 0 0 88px; }
.note { margin: 0; @include vars.font('body-medium'); opacity: 0.7; &.ok { color: rgb(var(--v-theme-success)); opacity: 1; } }
.label { margin: 0; @include vars.font('label-medium'); opacity: 0.7; }
.list { margin: 0; padding-left: 20px; @include vars.font('body-medium'); }
.mono { font-family: monospace; }
.res { display: flex; gap: vars.$spacing-md; @include vars.font('body-small'); min-width: 0; }
.key { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.err { color: rgb(var(--v-theme-error)); }
.ok { color: rgb(var(--v-theme-success)); }
</style>
