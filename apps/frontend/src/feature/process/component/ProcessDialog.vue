<template>
<v-dialog
    v-model="open"
    width="900"
    persistent
>
    <v-card class="process-dialog-card">
        <v-card-title class="title-bar">
            <div class="info">
                <span class="uid">{{ process?.uid }}</span>
                <v-chip size="small" :color="statusColor" variant="flat">{{ process?.status }}</v-chip>
            </div>
            <v-btn icon="mdi-close" variant="text" density="comfortable" @click="close" />
        </v-card-title>
        <v-divider />
        <v-card-text class="body">
            <div ref="termEl" class="term-el" />
        </v-card-text>
    </v-card>
</v-dialog>
</template>

<script lang="ts" setup>
import { useTestSocket } from '@/common/websocket/websocket.hook';
import { resizeProcess, type ProcessStatus } from '@/feature/process/api/process.api';
import { useProcessDialog } from '@/feature/process/store/processDialog.store';
import { FitAddon } from '@xterm/addon-fit';
import { Terminal } from '@xterm/xterm';
import '@xterm/xterm/css/xterm.css';
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue';
import { VBtn, VCard, VCardText, VCardTitle, VChip, VDialog, VDivider } from 'vuetify/components';

const { on } = useTestSocket();
const { open, process, closeProcessDialog } = useProcessDialog();

const termEl = ref<HTMLDivElement | null>(null);
let terminal: Terminal | null = null;
let fitAddon: FitAddon | null = null;

// fit() 직후 확정되는 초기 rows/cols. 지금은 서버로 보낼 곳이 없어 보관만 한다
// (input/resize 배선 미착수 — obsidian-vault/CURRENT.md 참조).
const rows = ref(0);
const cols = ref(0);

const STATUS_COLOR: Record<ProcessStatus, string> = {
    PENDING: 'warning',
    PROCESS: 'success',
    COMPLETED: 'grey',
    FAILED: 'error',
};
const statusColor = computed(() => {
    const status = process.value?.status;
    return status ? STATUS_COLOR[status] : 'grey';
});
// worker→sup DataEvent 미러(internal/protocol/messages.go). data는 Frame.Data(json) 봉투라
// []byte가 base64 문자열로 실려온다(임의 바이트 — PTY 출력이 유효 UTF-8이란 보장이 없어
// term.write에 Uint8Array로 넘긴다).
type DataEventPayload = {
    uid: string
    data: string
}

const base64ToBytes = (b64: string) => {
    const binary = atob(b64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
    return bytes;
};

on<DataEventPayload>('DATA', ev => {
    if (!terminal || ev.uid !== process.value?.uid) return;
    terminal.write(base64ToBytes(ev.data));
});

on('PROCESS:UPDATE', v => {
    console.log('[PROCESS:UPDATE] ->', v)
})

// 다이얼로그가 열린 채로 대상 process가 바뀌면(uid 변경) 이전 process의 화면이 남아있지
// 않도록 비운다. open watch(아래)는 open=false→true 전이에만 반응해 이 케이스를 못 잡는다.
watch(() => process.value?.uid, (uid, prevUid) => {
    if (uid && prevUid && uid !== prevUid) {
        terminal?.reset();
    }
});

watch([process, cols, rows], ([prc, c, r]) => {
    if (!prc || c === 0 || r === 0) return;
    resizeProcess(prc.uid, {
        cols: c,
        rows: r
    }).then(res => {
        console.log('[RESIZE] SUC' , res)
    }).catch(err => {
        console.log('[RESIZE] ERR' , err)

    })
})

const mountTerminal = () => {
    if (!termEl.value) return;
    if (!terminal) {
        terminal = new Terminal({ convertEol: true, disableStdin: true });
        fitAddon = new FitAddon();
        terminal.loadAddon(fitAddon);
        terminal.open(termEl.value);
    }
    fitAddon!.fit();
    rows.value = terminal.rows;
    cols.value = terminal.cols;
};

const disposeTerminal = () => {
    terminal?.dispose();
    terminal = null;
    fitAddon = null;
};

watch(open, async val => {
    if (val) {
        await nextTick();
        mountTerminal();
    } else {
        disposeTerminal();
    }
}, { immediate: true });

onBeforeUnmount(disposeTerminal);

const close = () => {
    closeProcessDialog();
};
</script>

<style lang="scss" scoped>
.process-dialog-card {
    display: flex;
    flex-direction: column;
    height: 640px;
}
.title-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
}
.info {
    display: flex;
    align-items: center;
    gap: vars.$spacing-sm;
}
.body {
    flex: 1;
    min-height: 0;
    padding: 0;
    overflow: hidden;
}
.term-el {
    width: 100%;
    height: 100%;
}
</style>
