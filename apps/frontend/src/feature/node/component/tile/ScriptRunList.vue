<script setup lang="ts">
import { STATUS_PILL } from '@/feature/process/store/processTerm.store';
import type { ProcessStatus } from '@/feature/process/api/process.api';
import { useQueryClient } from '@tanstack/vue-query';
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import { VBtn, VChip, VIcon } from 'vuetify/components';
import { useAppDialog } from '@/feature/layout/store/appDialog.store';
import { invalidateNodeRuns, type RunResponse } from '../../api/node.api';
import { openProcessTile } from '../../api/share.api';
import { useTerminalKill } from '../../hook/terminalKill.hook';
import { boardLabel, boardPath, useBoards } from '../../store/boards.store';
import { useTileGrids } from '../../store/tileGrids.store';
import { useTileTree } from '../../store/tileTree.store';

// 스크립트 행 펼침 — 이 스크립트의 실행 중 터미널(내 모든 보드). 실시간-6a, REF-process-share-realtime-ui.md
const props = defineProps({
    runs: {
        type: Array as () => RunResponse[],
        required: true,
    },
    name: {
        type: String,
        required: true,
    },
})

const PERM_OPERATE = 20  // 종료는 조작(20) 이상 — 서버 superdb.PermOperate

const router = useRouter();
const { tiles, hash: boardHash, setMinimized } = useTileTree();
const { posOf, reveal } = useTileGrids();
const { list: boards } = useBoards();
const { killing, confirmKill, confirmKillUid } = useTerminalKill();
const queryClient = useQueryClient();
const { openDialog } = useAppDialog();

// 내 어느 보드에도 타일이 없는 터미널 -> 기본 보드 끝에 타일을 만든다. 만들어지면 TILES:UPDATE로 트리가 갱신되고 행이 "보드로 이동/보기"로 바뀜
const opening = ref<Record<string, boolean>>({});
const onAdd = (row: (typeof rows.value)[number]) => {
    const uid = row.r.uid;
    opening.value[uid] = true;
    openProcessTile(uid)
    .then(() => invalidateNodeRuns(queryClient))
    .catch(err => openDialog({ type: 'error', content: err?.message || '열지 못했습니다' }))
    .finally(() => opening.value[uid] = false);
}

// 현재 보드에 있으면 로컬 타일이 기준(최소화 즉시 반영), 다른 보드·없음은 서버 응답
const rows = computed(() => props.runs.map(r => {
    const local = r.hash === boardHash.value && r.tileId ? tiles.value[r.tileId] : undefined;
    const board = r.hash && !local ? boards.value.find(b => b.hash === r.hash) : undefined;
    return { r, local, board, minimized: local ? !!local.minimized : !!r.minimized };
}))

const onOpen = (row: (typeof rows.value)[number]) => {
    if (row.local) {
        if (row.local.minimized) setMinimized(row.local.id, false);
        reveal(row.local.id);
    } else if (row.board) {
        router.push(boardPath(row.board));
    }
}
const onKill = (row: (typeof rows.value)[number]) => {
    if (row.local) confirmKill(row.local.id, props.name);
    else confirmKillUid(row.r.uid, props.name);
}
const pill = (status: string) => STATUS_PILL[status as ProcessStatus] ?? { text: status };
</script>

<template>
<div class="ScriptRunList">
    <div v-for="row in rows" :key="row.r.uid" class="run">
        <span v-if="row.local && !row.minimized" class="pos">{{ posOf.get(row.local.id) }}</span>
        <span v-else-if="row.minimized" class="min" title="최소화"><v-icon size="11" icon="mdi-window-minimize" /></span>
        <span v-else class="pos none">-</span>
        <span class="where">
            <template v-if="row.local">이 보드</template>
            <template v-else-if="row.board">{{ boardLabel(row.board) }}</template>
            <template v-else>위치 없음</template>
            <span class="by"> · {{ row.r.startedBy }}</span>
        </span>
        <v-chip size="x-small" variant="tonal" class="pill" :color="pill(row.r.status).color">{{ pill(row.r.status).text }}</v-chip>
        <span class="acts">
            <v-btn v-if="row.local" size="x-small" variant="tonal" :color="row.minimized ? 'primary' : undefined" @click="onOpen(row)">{{ row.minimized ? '열기' : '보기' }}</v-btn>
            <v-btn v-else-if="row.board" size="x-small" variant="tonal" @click="onOpen(row)">보드로 이동</v-btn>
            <v-btn v-else size="x-small" variant="tonal" color="primary" :loading="!!opening[row.r.uid]" @click="onAdd(row)">내 보드에 열기</v-btn>
            <v-btn v-if="row.r.permission >= PERM_OPERATE" size="x-small" variant="text" icon="mdi-stop" title="종료…" :loading="!!killing[row.r.uid]" @click="onKill(row)" />
        </span>
    </div>
</div>
</template>

<style scoped lang="scss">
.ScriptRunList {
    background: rgba(var(--v-theme-on-surface), 0.04);
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}
.run {
    display: grid;
    grid-template-columns: 20px minmax(0, 1fr) auto auto;
    align-items: center;
    gap: 6px;
    min-height: 28px;
    padding: vars.$spacing-xs vars.$spacing-xs vars.$spacing-xs 36px;
    font-size: 11.5px;
}
.pos, .min {
    min-width: 16px;
    height: 16px;
    padding: 0 4px;
    border-radius: 4px;
    display: inline-grid;
    place-items: center;
    font-size: 10px;
    font-weight: vars.$weight-lg;
}
.pos {
    background: rgb(var(--v-theme-on-surface));
    color: rgb(var(--v-theme-surface));
    &.none { opacity: 0.35; }
}
.min {
    border: 1px dashed rgba(var(--v-border-color), 0.6);
    opacity: 0.75;
}
.where {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.by { opacity: 0.6; }
.pill { font-family: monospace; }
.acts {
    display: flex;
    align-items: center;
    gap: vars.$spacing-xs;
}
</style>
