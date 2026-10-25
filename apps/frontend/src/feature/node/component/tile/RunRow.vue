<script setup lang="ts">
import { isRunning, pillOf } from '@/feature/process/store/processTerm.store';
import type { ProcessStatus } from '@/feature/process/api/process.api';
import { useQueryClient } from '@tanstack/vue-query';
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import { VChip, VIcon, VProgressCircular } from 'vuetify/components';
import { useAppDialog } from '@/feature/layout/store/appDialog.store';
import { invalidateNodeRuns, useNodePath, type RunResponse } from '../../api/node.api';
import { bindProcessTile } from '../../api/share.api';
import { useTerminalKill } from '../../hook/terminalKill.hook';
import { boardLabel, focusPath, useBoards } from '../../store/boards.store';
import { useTileGrids } from '../../store/tileGrids.store';
import { isProcTile, useTileTree } from '../../store/tileTree.store';
import { shortPath } from '../../util/tileLabel.util';

// 터미널 탭·노드의 상태 리스트의 한 행 — 명령 = 보기/보드로 이동 + 현재 타일에 바인딩 + 종료·닫기(REF-tile-binding-rework.md "목록 행 명령")
const props = defineProps({
    run: {
        type: Object as () => RunResponse,
        required: true,
    },
    targetTileId: {     // "현재 타일" — 노드의 상태 리스트 = 그 폴더 타일 / 터미널 탭 = 루트 타일(R7(가져올 위치))
        type: String,
        required: true,
    },
    showName: {         // 터미널 탭은 여러 스크립트가 섞여 경로·이름을 보여 줌
        type: Boolean,
        default: false,
    },
})

const PERM_OPERATE = 20  // 종료는 조작(20) 이상 — 서버 superdb.PermOperate

const router = useRouter();
const queryClient = useQueryClient();
const { openDialog } = useAppDialog();
const { tiles, hash: boardHash, setMinimized } = useTileTree();
const { posOf, activate } = useTileGrids();
const { list: boards } = useBoards();
const { killing, closeTile, confirmKill, confirmKillUid } = useTerminalKill();

// 현재 보드에 있으면 로컬 타일이 기준(최소화 즉시 반영), 다른 보드·없음은 서버 응답
const local = computed(() => {
    const r = props.run;
    const t = r.hash === boardHash.value && r.tileId ? tiles.value[r.tileId] : undefined;
    return isProcTile(t) ? t : undefined;
})
const board = computed(() => props.run.hash && !local.value ? boards.value.find(b => b.hash === props.run.hash) : undefined);
const minimized = computed(() => local.value ? !!local.value.minimized : !!props.run.minimized);
const bound = computed(() => local.value?.parent === props.targetTileId);
const running = computed(() => isRunning(props.run.status as ProcessStatus));
const pill = computed(() => pillOf(props.run.status as ProcessStatus, props.run.terminating));

// 단독 공유 process는 노드를 못 봐 경로 조회가 404 — 이름만(R3(경로 없는 process))
const { data: scriptPath } = useNodePath(computed(() => props.showName ? props.run.nodeId : null));
const pathText = computed(() => props.run.nodeId === null ? shortPath(props.run.cwd ?? '') : (scriptPath.value ?? []).slice(0, -1).map(p => p.name).join(' / '));   // 장비 실행 = 실행한 폴더(끝 두 단계)

const onOpen = () => {
    if (local.value) {
        if (local.value.minimized) setMinimized(local.value.id, false);
        activate(local.value.id);
    } else if (board.value && props.run.tileId) {
        router.push(focusPath(board.value, props.run.tileId));
    }
}

const binding = ref(false);
const onBind = () => {
    binding.value = true;
    bindProcessTile(props.run.uid, boardHash.value, props.targetTileId)
    .then(() => invalidateNodeRuns(queryClient))
    .catch(err => openDialog({ type: 'error', content: err?.message || '바인딩하지 못했습니다' }))
    .finally(() => binding.value = false);
}

const canKill = computed(() => running.value && props.run.permission >= PERM_OPERATE);
const onKill = () => {
    if (local.value) confirmKill(local.value.id, props.run.name);
    else confirmKillUid(props.run.uid, props.run.name, !!props.run.terminating);
}
</script>

<template>
<!-- 찰떡: 위치 번호는 타일 왼쪽 벽, 명령 묶음은 오른쪽 벽에 붙는다(REF-design-chaltteok.md) -->
<div class="RunRow" :class="{ named: showName }">
    <span class="ct-grp ct-grp--l ct-grp--sm">
        <span v-if="local && !minimized" class="ct-seg pos">{{ posOf.get(local.id) }}</span>
        <span v-else-if="minimized" class="ct-seg ct-seg--txt pos" title="최소화"><v-icon size="12" icon="mdi-window-minimize" /></span>
        <span v-else class="ct-seg pos none">-</span>
    </span>
    <span class="where">
        <template v-if="showName">
            <span v-if="pathText" class="path" :title="run.cwd">{{ pathText }} / </span>{{ run.name }}
            <span class="by"> · </span>
        </template>
        <template v-if="local">이 보드</template>
        <template v-else-if="board">{{ boardLabel(board) }}</template>
        <template v-else>위치 없음</template>
        <span class="by"> · {{ run.startedByMe ? '나' : run.startedBy }}</span>
    </span>
    <v-chip size="x-small" variant="tonal" class="pill" :color="pill.color">{{ pill.text }}</v-chip>
    <span class="ct-grp ct-grp--r ct-grp--sm">
        <button v-if="local" type="button" class="ct-seg" :class="{ 'ct-seg--accent': minimized }" @click="onOpen">{{ minimized ? '열기' : '보기' }}</button>
        <button v-else-if="board" type="button" class="ct-seg" @click="onOpen">보드로 이동</button>
        <button v-if="!bound" type="button" class="ct-seg ct-seg--accent" :disabled="binding" title="이 터미널 타일을 현재 타일 아래로 옮김(없으면 새로 엶)" @click="onBind">
            <v-progress-circular v-if="binding" indeterminate size="11" width="2" />현재 타일에 바인딩
        </button>
        <button v-if="canKill" type="button" class="ct-seg ct-seg--icon" :title="run.terminating ? '강제 종료…' : '종료…'" :disabled="!!killing[run.uid]" @click="onKill">
            <v-progress-circular v-if="killing[run.uid]" indeterminate size="13" width="2" />
            <v-icon v-else size="15" :icon="run.terminating ? 'mdi-stop-circle' : 'mdi-stop'" />
        </button>
        <button v-else-if="local && !running" type="button" class="ct-seg ct-seg--icon" title="닫기" @click="closeTile(local.id)"><v-icon size="15" icon="mdi-close" /></button>
    </span>
</div>
</template>

<style scoped lang="scss">
.RunRow {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto auto;
    align-items: center;
    gap: vars.$spacing-md;
    min-height: 32px;
    padding: vars.$spacing-xs 0;
    @include vars.font('body-small');
    &.named { border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity)); }
}
.pos {
    min-width: 26px;
    font-weight: vars.$weight-lg;
    &:not(.ct-seg--txt):not(.none) {
        background: rgb(var(--v-theme-on-surface));
        color: rgb(var(--v-theme-surface));
    }
    &.none { opacity: 0.6; }
}
.where {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.path, .by { opacity: 0.6; }
.pill { font-family: monospace; }
</style>
