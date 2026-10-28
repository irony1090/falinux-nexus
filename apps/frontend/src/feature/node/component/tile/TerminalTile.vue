<script setup lang="ts">
import { useAppDialog } from '@/feature/layout/store/appDialog.store';
import { isRunning, pillOf, useProcessTerm } from '@/feature/process/store/processTerm.store';
import { TILE_KINDS } from '@/feature/widget/util/tileKind.util';
import { computed, ref } from 'vue';
import { VChip, VIcon, VList, VListItem, VMenu, VProgressCircular } from 'vuetify/components';
import { useGetNode, useNodePath } from '../../api/node.api';
import { useTerminalKill } from '../../hook/terminalKill.hook';
import { useRemoteActive } from '../../store/remoteActive.store';
import { useTileGrids } from '../../store/tileGrids.store';
import { useTileTree } from '../../store/tileTree.store';
import { fileLabel, nodeLabel, shortPath } from '../../util/tileLabel.util';
import { isReadOnly, permInfo } from '../../util/shareView.util';
import PresenceDialog from '../dialog/PresenceDialog.vue';
import TerminalTileBody from './TerminalTileBody.vue';
import TileFrame from './TileFrame.vue';

// 터미널 타일 — process 하나(능력-3(process 담음)). 머리 줄 A안 `RUN▾ [최소화 | 크기 | ×]` — SD3(실행 중 터미널 치우기)
const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const { tiles, setMinimized, setShortcutsOff } = useTileTree();
const { posOf, instanceOf, rootTab, overview } = useTileGrids();
const { procs, owners, restored, gone, takeSize } = useProcessTerm();
const { killing, closeTile, confirmKill } = useTerminalKill();
const { openDialog } = useAppDialog();

const tile = computed(() => {
    const t = tiles.value[props.tileId];
    return t?.type === 'terminal' ? t : undefined;
})

const proc = computed(() => tile.value ? procs.value[tile.value.uid] : undefined);

// nodeId null = 장비 파일 실행(폴더-4(실행)) — 이름 = 파일 이름, 경로 = 실행한 폴더
const { data: node } = useGetNode(computed(() => tile.value?.nodeId ?? NaN));
const name = computed(() => {
    const t = tile.value;
    if (!t) return '';
    if (t.nodeId === null) return proc.value ? fileLabel(proc.value.cmd, proc.value.args) : '실행 파일';
    const no = instanceOf.value.get(t.id);
    return nodeLabel(t.nodeId, node.value?.name) + (no ? ` (${no})` : '');
})

// ---- 공유-e-1(받는 쪽 표시): 공유받은 터미널 = process 응답에 소유자 닉네임 ----
const shareOwner = computed(() => proc.value?.ownerNickname);
const shareInfo = computed(() => shareOwner.value ? permInfo(proc.value?.permission) : undefined);
const readOnly = computed(() => isReadOnly(proc.value?.permission));
const startedBy = computed(() => proc.value?.startedByMe ? '나' : proc.value?.startedBy);

// 터미널 머리 줄 `[실행한 타일 위치 번호] 실행된 경로 / 이름` — 번호 = 소유(타일 트리), 경로 = 스크립트가 있는 폴더(노드 트리). REF-workspace.md "터미널 머리 줄"
const ownerPos = computed(() => tile.value?.parent ? posOf.value.get(tile.value.parent) : undefined);
const { data: scriptPath } = useNodePath(computed(() => tile.value?.nodeId ?? null));
const pathFull = computed(() => tile.value?.nodeId === null
    ? proc.value?.cwd ?? ''
    : [shareOwner.value, ...(scriptPath.value ?? []).slice(0, -1).map(p => p.name)].filter(Boolean).join(' / '));
const pathText = computed(() => tile.value?.nodeId === null ? shortPath(pathFull.value) : pathFull.value);

// 볼 수 없음 = 서버가 process를 안 돌려줌 — 실시간-6b
const unavailable = computed(() => !!tile.value && !!gone.value[tile.value.uid]);

const running = computed(() => isRunning(proc.value?.status));
const terminating = computed(() => running.value && !!proc.value?.terminating);
const presenceOpen = ref(false);   // 접속 상태 패널(공유-f)

// 크기-3(버튼): 이 탭이 크기 소유자가 아닌 실행 중 터미널만. PENDING은 서버 memory에 없어 가져올 수 없다
const canTakeSize = computed(() => !!tile.value && proc.value?.status === 'PROCESS' && !readOnly.value && !owners.value[tile.value.uid]);
const taking = ref(false);
const onTakeSize = () => {
    const t = tile.value;
    if (!t) return;
    taking.value = true;
    takeSize(t.uid)
    .catch(err => openDialog({ type: 'error', content: err?.message || '크기를 가져오지 못했습니다' }))
    .finally(() => taking.value = false);
}

const pill = computed(() => pillOf(proc.value?.status, proc.value?.terminating));
const pillStyle = computed(() => pill.value.color ? { '--c': `var(--v-theme-${pill.value.color})` } : undefined);
// 복원 전엔 상태를 몰라 X를 막는다 — 3-d②(복원 전 닫기)
const unknown = computed(() => !proc.value && !restored.value);

// 최소화한 타일이 어디로 갔는지 보이게 루트 타일을 터미널 탭으로
const onMinimize = () => {
    if (!tile.value) return;
    setMinimized(tile.value.id, true);
    rootTab.value = 'terminals';
}

// 터미널 타일 메뉴 — 이 타일이 active인 동안 앱 단축키 끄기(이동-3(타일별 단축키 끔))
const shortcutsOff = computed(() => !!tile.value?.shortcutsOff);
const toggleShortcuts = () => {
    if (tile.value) setShortcutsOff(tile.value.id, !shortcutsOff.value);
}

// 다른 계정 active — 터미널 = uid. RA1(맞추는 기준)
const { byUid } = useRemoteActive();
const remoteUsers = computed(() => {
    const t = tile.value;
    if (!t || unavailable.value) return [];
    return byUid.value.get(t.uid) ?? [];
})

// X는 실행 중이면 kill 확인창(종료 요청 뒤엔 강제 종료·닫기). 보기 계정은 종료할 수 없고 "모두 닫음"에도 안 세므로 그냥 닫는다(REF-process-kill.md)
const killOnClose = computed(() => running.value && !unavailable.value && !readOnly.value);
const closeTitle = computed(() => running.value && !readOnly.value ? (terminating.value ? '강제 종료·닫기…' : '종료…') : '닫기');
const onX = () => {
    const t = tile.value;
    if (!t) return;
    if (killOnClose.value) confirmKill(t.id, name.value);
    else closeTile(t.id);
}
</script>

<template>
<tile-frame v-if="tile" :tile-id="tile.id"
    :name="name"
    :icon="TILE_KINDS.terminal.icon"
    :color="TILE_KINDS.terminal.color"
    :shared="!!shareOwner"
    :share-info="shareInfo"
    :unavailable="unavailable"
    :remote-users="remoteUsers"
    :menu="!unavailable"
    minimizable
    :close-title="closeTitle"
    :close-disabled="unknown"
    :float-uid="tile.uid"
    @close="onX"
    @minimize="onMinimize"
>
    <template #ov-extra>
        <v-chip size="x-small" variant="tonal" class="pill" :color="pill.color">{{ pill.text }}</v-chip>
    </template>

    <template #title>
        <span v-if="ownerPos" class="owner" title="이 터미널을 실행한 타일">[{{ ownerPos }}]</span>
        <span v-if="pathText" class="path" :title="pathFull">{{ pathText }} /</span>
        <span class="name">{{ name }}</span>
        <span v-if="startedBy" class="by" :title="proc?.startedByMe ? '내가 실행' : `${startedBy}님이 실행`">실행: {{ startedBy }}</span>
        <v-icon v-if="shortcutsOff" class="keys-off" size="14" icon="mdi-keyboard-off-outline" title="단축키 꺼짐 — 이 타일이 active인 동안 Ctrl+Alt 단축키가 터미널로 갑니다" />
    </template>

    <template v-if="!unavailable" #controls>
        <button v-if="canTakeSize" class="tile-btn" title="이 화면 크기로 가져오기" :disabled="taking" @click="onTakeSize">
            <v-progress-circular v-if="taking" indeterminate size="14" width="2" />
            <v-icon v-else size="16" icon="mdi-arrow-expand-all" />
        </button>
        <button class="tile-btn" title="접속 상태" @click="presenceOpen = true">
            <v-icon size="16" icon="mdi-account-multiple-outline" />
        </button>
        <span class="tile-btn-sep" />
        <v-menu v-if="running && !readOnly" location="bottom end">
            <template #activator="{ props: act }">
                <button v-bind="act" class="tile-btn tile-btn--text tile-btn--tonal pill" :style="pillStyle" :disabled="!!killing[tile.uid]">
                    {{ pill.text }}<v-icon size="14" icon="mdi-menu-down" />
                </button>
            </template>
            <v-list v-if="terminating" density="compact">
                <v-list-item prepend-icon="mdi-stop-circle" title="강제 종료…" @click="confirmKill(tile.id, name, ['force'])" />
            </v-list>
            <v-list v-else density="compact">
                <v-list-item prepend-icon="mdi-stop" title="종료…" @click="confirmKill(tile.id, name, ['kill'])" />
                <v-list-item prepend-icon="mdi-close-box-outline" title="종료하고 닫기…" @click="confirmKill(tile.id, name, ['killClose'])" />
            </v-list>
        </v-menu>
        <span v-else class="tile-btn tile-btn--text tile-btn--tonal pill" :style="pillStyle" :title="proc?.exitCode != null ? `exit ${proc.exitCode}` : undefined">{{ pill.text }}</span>
    </template>

    <template #menu>
        <v-list-item
            :title="shortcutsOff ? '단축키 켜기' : '단축키 끄기'"
            subtitle="이 타일이 active일 때 Ctrl+Alt 단축키"
            :prepend-icon="shortcutsOff ? 'mdi-keyboard-outline' : 'mdi-keyboard-off-outline'"
            @click="toggleShortcuts"
        />
    </template>

    <span v-if="readOnly" class="ro">보기 전용 · 입력 안 됨</span>
    <terminal-tile-body :tile-id="tile.id" :inert="overview" />

    <template #extra>
        <presence-dialog v-if="!unavailable" v-model="presenceOpen" :uid="tile.uid" :name="name" />
    </template>
</tile-frame>
</template>

<style scoped lang="scss">
.TileFrame__head > .name {
    flex: 0 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-weight: vars.$weight-md;
}
.owner {
    font-family: monospace;
    @include vars.font('body-small');
    opacity: 0.75;
}
.TileFrame__head > .path {
    flex: 0 1000 auto;  // 이름보다 먼저 줄어든다
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    @include vars.font('body-small');
    opacity: 0.7;
}
.TileFrame__head > .by {
    flex: 0 1000 auto;  // 이름보다 먼저 줄어든다
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    @include vars.font('body-small');
    opacity: 0.7;
}
.ro {
    position: absolute;
    right: 8px;
    bottom: 8px;
    z-index: 2;
    padding: 0 6px;
    @include vars.font('body-small');
    border: 1px solid rgba(var(--v-theme-on-terminal), 0.35);
    border-radius: 4px;
    background: rgba(var(--v-theme-terminal), 0.7);
    color: rgb(var(--v-theme-on-terminal));
    pointer-events: none;
}
.pill { font-family: monospace; }
</style>
