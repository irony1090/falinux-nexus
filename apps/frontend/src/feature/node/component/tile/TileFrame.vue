<script setup lang="ts">
import type { TileSpan } from '@/feature/widget/util/tile.type';
import { effectiveSize } from '@/feature/widget/util/tileSplit.util';
import { useAppDialog } from '@/feature/layout/store/appDialog.store';
import { isRunning, pillOf, useProcessTerm } from '@/feature/process/store/processTerm.store';
import { computed, ref } from 'vue';
import { useResizeCallback } from '@/feature/common/store/resizeGroup.store';
import { useRouter } from 'vue-router';
import { VBtn, VCard, VChip, VIcon, VList, VListItem, VListSubheader, VMenu } from 'vuetify/components';
import { useGetNode, useNodePath, useNodeRuns } from '../../api/node.api';
import { bindTile } from '../../api/tile.api';
import { useTerminalKill } from '../../hook/terminalKill.hook';
import { useTileFloat } from '../../hook/tileFloat.hook';
import { boardLabel, boardPath, useBoards } from '../../store/boards.store';
import { useTileGrids } from '../../store/tileGrids.store';
import { useTileTree } from '../../store/tileTree.store';
import { isReadOnly, permInfo } from '../../util/shareView.util';
import NodeNameDialog from '../dialog/NodeNameDialog.vue';
import PresenceDialog from '../dialog/PresenceDialog.vue';
import { useRemoteActive } from '../../store/remoteActive.store';
import ActiveBadges from './ActiveBadges.vue';
import FolderTileBody from './FolderTileBody.vue';
import RootTerminalList from './RootTerminalList.vue';
import TerminalTileBody from './TerminalTileBody.vue';

const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const { tiles, hash: boardHash, pinByNode, resize, setMinimized, setShortcutsOff, detachToBoard, moveToBoard, navigate } = useTileTree();
const { list: boards, add: addBoard } = useBoards();
const router = useRouter();
const { cap, posOf, instanceOf, reveal, overview, rootTab, activeId, setActive } = useTileGrids();
const { procs, owners, restored, gone, takeSize } = useProcessTerm();
const { killing, closeTile, confirmKill } = useTerminalKill();
const { openDialog } = useAppDialog();

const tile = computed(() => tiles.value[props.tileId]);

const openerTile = computed(() => tile.value?.parent ? tiles.value[tile.value.parent] : undefined);

// useGetNode는 id가 NaN이면 조회하지 않는다 (nodeId null = 루트 목록)
const { data: node, error: nodeError } = useGetNode(computed(() => tile.value?.nodeId ?? NaN));
const { data: openerNode } = useGetNode(computed(() => openerTile.value?.nodeId ?? NaN));

// 이름을 받기 전에는 node id로 표시
const label = (nodeId: number | null, loaded?: string) => nodeId === null ? '기본' : loaded ?? `node ${nodeId}`;

const name = computed(() => {
    const t = tile.value;
    if (!t) return '';
    const no = t.type === 'terminal' ? instanceOf.value.get(t.id) : undefined;
    return label(t.nodeId, node.value?.name) + (no ? ` (${no})` : '');
})
const opener = computed(() => {
    const parent = openerTile.value;
    return parent?.type === 'folder' ? label(parent.nodeId, openerNode.value?.name) : undefined;
})

// ---- 공유-e-1(받는 쪽 표시): 공유받은 것 = 소유자 닉네임이 있는 노드·process. 폴더 = 노드 응답 / 터미널 = process 응답 ----
const proc = computed(() => tile.value?.type === 'terminal' ? procs.value[tile.value.uid] : undefined);
const shareOwner = computed(() => tile.value?.type === 'terminal' ? proc.value?.ownerNickname : node.value?.ownerNickname);
const sharePerm = computed(() => tile.value?.type === 'terminal' ? proc.value?.permission : node.value?.permission);
const shareInfo = computed(() => shareOwner.value ? permInfo(sharePerm.value) : undefined);
const readOnly = computed(() => tile.value?.type === 'terminal' && isReadOnly(sharePerm.value));
const shareRoot = computed(() => tile.value?.type === 'folder' && !!shareOwner.value && node.value?.parentVisible === false); // 공유 시작 지점
const startedBy = computed(() => proc.value?.startedByMe ? '나' : proc.value?.startedBy);

// 터미널 머리 줄 `[실행한 타일 위치 번호] 실행된 경로 / 이름` — 번호 = 소유(타일 트리), 경로 = 스크립트가 있는 폴더(노드 트리). REF-workspace.md "터미널 머리 줄"
const ownerPos = computed(() => tile.value?.type === 'terminal' && tile.value.parent ? posOf.value.get(tile.value.parent) : undefined);
const { data: scriptPath } = useNodePath(computed(() => tile.value?.type === 'terminal' ? tile.value.nodeId : null));
const pathText = computed(() => [shareOwner.value, ...(scriptPath.value ?? []).slice(0, -1).map(p => p.name)].filter(Boolean).join(' / '));

// 볼 수 없는 타일(공유 취소·삭제·범위 밖 이동) — 실시간-6b. 폴더 = 노드 404 / 터미널 = 서버가 process를 안 돌려줌. 닫기만 가능
const unavailable = computed(() => {
    const t = tile.value;
    if (!t) return false;
    if (t.type === 'terminal') return !!gone.value[t.uid];
    return t.nodeId !== null && (nodeError.value as { status?: number } | null)?.status === 404;
})

const running = computed(() => isRunning(proc.value?.status));
const terminating = computed(() => running.value && !!proc.value?.terminating);
const presenceOpen = ref(false);   // 접속 상태 패널(공유-f)

// 크기-3(버튼): 이 탭이 크기 소유자가 아닌 실행 중 터미널만. PENDING은 서버 memory에 없어 가져올 수 없다
const canTakeSize = computed(() => tile.value?.type === 'terminal' && proc.value?.status === 'PROCESS' && !readOnly.value && !owners.value[tile.value.uid]);
const taking = ref(false);
const onTakeSize = () => {
    const t = tile.value;
    if (t?.type !== 'terminal') return;
    taking.value = true;
    takeSize(t.uid)
    .catch(err => openDialog({ type: 'error', content: err?.message || '크기를 가져오지 못했습니다' }))
    .finally(() => taking.value = false);
}

const pill = computed(() => pillOf(proc.value?.status, proc.value?.terminating));
// 복원 전엔 상태를 몰라 X를 막는다 — 3-d②(복원 전 닫기)
const unknown = computed(() => tile.value?.type === 'terminal' && !proc.value && !restored.value);

const fmt = (v: TileSpan) => v === 1 ? '1' : '.5';
const sizeText = computed(() => tile.value ? `${fmt(tile.value.size.w)}×${fmt(tile.value.size.h)}` : '');
// 이 화면에서 실제로 보이는 크기가 다를 때만
const shownText = computed(() => {
    if (!tile.value) return '';
    const s = effectiveSize(tile.value.size, cap.value);
    const text = `${fmt(s.w)}×${fmt(s.h)}`;
    return text === sizeText.value ? '' : text;
})

// 크기를 바꾸면 버튼이 옮겨지는데 v-menu는 이를 감지 못함 → 프레임 크기 변화·크기 변경 직후에 위치 재계산
const frameEl = ref<HTMLElement>();
const sizeMenu = ref<InstanceType<typeof VMenu>>();
const relocateSizeMenu = () => requestAnimationFrame(() => sizeMenu.value?.updateLocation?.(new Event('resize')));
useResizeCallback(frameEl, relocateSizeMenu);

// 맞춤-1(타일 띄우기) — 키보드가 열린 동안 포커스된 터미널 타일만 띄운다
const termUid = computed(() => tile.value?.type === 'terminal' ? tile.value.uid : undefined);
const floatStyle = useTileFloat(frameEl, termUid, overview);

const setSize = (axis: 'w' | 'h', v: TileSpan) => {
    if (!tile.value) return;
    resize(tile.value.id, { ...tile.value.size, [axis]: v });
    reveal(tile.value.id);
    relocateSizeMenu();
}
const sizeTitle = computed(() => `크기 ${sizeText.value}` + (shownText.value ? ` (이 화면에서는 ${shownText.value})` : ''));

// ---- 머리 줄 A안 `RUN▾ [최소화 | 크기 | ×]` — SD3(실행 중 터미널 치우기). X는 실행 중이면 kill 확인창(종료 요청 뒤엔 강제 종료·닫기) ----
const isRoot = computed(() => !tile.value?.parent);
// 탭 숫자 = 터미널 탭 목록(루트 경로 하위, RootTerminalList와 같은 조회 — 개편-6(목록 UI)). 갱신은 폴더 타일의 useScriptRuns가 무효화
const { data: rootRuns } = useNodeRuns(computed(() => tile.value?.type === 'folder' ? tile.value.nodeId : null), true, isRoot);
const runCount = computed(() => rootRuns.value?.length ?? 0);
const minCount = computed(() => rootRuns.value?.filter(r => r.minimized).length ?? 0);
const setTab = (v: 'files' | 'terminals') => rootTab.value = v;

// 최소화한 타일이 어디로 갔는지 보이게 루트 타일을 터미널 탭으로
const onMinimize = () => {
    if (tile.value?.type !== 'terminal') return;
    setMinimized(tile.value.id, true);
    setTab('terminals');
}
// ---- 보드: 폴더 타일(루트 제외)을 하위째 새 보드로 분리(어느 보드에서든) / 다른 보드로 이동 — W1(분리 범위)·W3(이동 범위) ----
const otherBoards = computed(() => boards.value.filter(b => b.hash !== boardHash.value));
const boardMenu = computed(() => tile.value?.type === 'folder' && !!tile.value.parent && !unavailable.value);
const boardBusy = ref(false);
const boardRun = (job: Promise<unknown>) => {
    boardBusy.value = true;
    job.catch(err => openDialog({ type: 'error', content: err?.message || '요청에 실패했습니다' }))
    .finally(() => boardBusy.value = false);
}

const detachOpen = ref(false);
const onDetach = (name: string) => {
    const t = tile.value;
    if (!t) return;
    boardRun(detachToBoard(t.id, name).then(b => {
        detachOpen.value = false;
        addBoard(b);
        router.push(boardPath(b));
    }));
}
const onMove = (to: string) => {
    const t = tile.value;
    if (t) boardRun(moveToBoard(t.id, to));
}

// 타일 고정 — 같은 폴더 타일이 내 보드 통틀어 2개 이상일 때만 핀·메뉴(R4(고정 표시 남발)), 해제 없음(R5(해제의 의미)). 루트 목록(nodeId null)은 대상 아님
const pin = computed(() => tile.value?.type === 'folder' && tile.value.nodeId !== null ? pinByNode.value.get(tile.value.nodeId) : undefined);
const pinned = computed(() => !!pin.value && pin.value.hash === boardHash.value && pin.value.tileId === tile.value?.id);
const canPin = computed(() => !!pin.value && !pinned.value && !unavailable.value);
const onPin = () => {
    const t = tile.value;
    if (t) boardRun(bindTile(boardHash.value, t.id));
}

// 터미널 타일 메뉴 — 이 타일이 active인 동안 앱 단축키 끄기(이동-3(타일별 단축키 끔))
const terminalMenu = computed(() => tile.value?.type === 'terminal' && !unavailable.value);
const shortcutsOff = computed(() => tile.value?.type === 'terminal' && !!tile.value.shortcutsOff);
const toggleShortcuts = () => {
    if (tile.value?.type === 'terminal') setShortcutsOff(tile.value.id, !shortcutsOff.value);
}

// 다른 계정 active — 터미널 = uid / 폴더 = 보여 주는 nodeId(루트 제외). RA1(맞추는 기준)
const { byUid, byNodeId } = useRemoteActive();
const remoteUsers = computed(() => {
    const t = tile.value;
    if (!t || unavailable.value) return [];
    if (t.type === 'terminal') return byUid.value.get(t.uid) ?? [];
    return t.nodeId === null ? [] : byNodeId.value.get(t.nodeId) ?? [];
})

// active 표시 — AT3(표시). active가 있을 때만 나머지를 흐린다
const activeClass = computed(() => activeId.value ? (activeId.value === props.tileId ? 'TileFrame--active' : 'TileFrame--dim') : undefined);

const onX = () => {
    const t = tile.value;
    if (!t) return;
    // 보기 계정은 종료할 수 없고 "모두 닫음"에도 안 세므로 그냥 닫는다(REF-process-kill.md)
    if (t.type === 'terminal' && running.value && !unavailable.value && !readOnly.value) confirmKill(t.id, name.value);
    else closeTile(t.id);
}
</script>

<template>
<div v-if="tile" ref="frameEl" class="TileFrame" tabindex="-1"
    :data-tile-id="tile.id"
    :class="[`TileFrame--${tile.type}`, activeClass, { 'TileFrame--overview': overview, 'TileFrame--shared': !!shareOwner }]"
    :style="floatStyle"
    @pointerdown.capture="setActive(tile.id)"
>
    <!-- 전체보기 라벨: 축소 배율을 거꾸로 곱해 읽을 수 있는 크기 유지 -->
    <div v-if="overview" class="ov-name">
        <span class="pos">{{ posOf.get(tile.id) }}</span>
        <v-icon size="14"
            :icon="tile.type === 'folder' ? (shareOwner ? 'mdi-folder-account' : 'mdi-folder') : 'mdi-console'"
            :color="tile.type === 'folder' ? 'warning' : 'primary'"
        />
        <span class="name">{{ name }}</span>
        <active-badges :users="remoteUsers" />
        <v-chip v-if="tile.type === 'terminal'" size="x-small" variant="tonal" class="pill" :color="pill.color">{{ pill.text }}</v-chip>
    </div>
    <div class="TileFrame__head" :inert="overview">
        <span class="pos">{{ posOf.get(tile.id) }}</span>
        <v-icon size="14"
            :icon="tile.type === 'folder' ? (shareOwner ? 'mdi-folder-account' : 'mdi-folder') : 'mdi-console'"
            :color="tile.type === 'folder' ? 'warning' : 'primary'"
        />
        <template v-if="tile.type === 'terminal'">
            <span v-if="ownerPos" class="owner" title="이 터미널을 실행한 타일">[{{ ownerPos }}]</span>
            <span v-if="pathText" class="path" :title="pathText">{{ pathText }} /</span>
            <span class="name">{{ name }}</span>
            <span v-if="startedBy" class="by" :title="proc?.startedByMe ? '내가 실행' : `${startedBy}님이 실행`">실행: {{ startedBy }}</span>
            <v-icon v-if="shortcutsOff" class="keys-off" size="14" icon="mdi-keyboard-off-outline" title="단축키 꺼짐 — 이 타일이 active인 동안 Ctrl+Alt 단축키가 터미널로 갑니다" />
        </template>
        <template v-else>
            <span class="name">{{ name }}</span>
            <v-icon v-if="pinned" size="14" icon="mdi-pin" color="warning" title="타일 고정 — 다른 계정이 실행한 터미널을 이 타일이 받음" />
            <v-chip v-if="shareRoot" size="x-small" variant="flat" color="share" prepend-icon="mdi-account" :title="`${shareOwner}님이 공유한 폴더`">{{ shareOwner }}</v-chip>
            <span v-if="opener" class="opener text-warning">← {{ opener }}</span>
        </template>
        <span class="spacer" />
        <span v-if="shareInfo" class="perm" :title="`공유 단계: ${shareInfo.text}`">
            <v-icon size="14" :icon="shareInfo.icon" />{{ shareInfo.text }}
        </span>
        <active-badges :users="remoteUsers" />

        <template v-if="tile.type === 'terminal' && !unavailable">
            <v-btn v-if="canTakeSize" size="x-small" variant="text" icon="mdi-arrow-expand-all" title="이 화면 크기로 가져오기" :loading="taking" @click="onTakeSize" />
            <v-btn size="x-small" variant="text" icon="mdi-account-multiple-outline" title="접속 상태" @click="presenceOpen = true" />
            <v-menu v-if="running && !readOnly" location="bottom end">
                <template #activator="{ props: act }">
                    <v-chip v-bind="act" size="x-small" variant="tonal" class="pill" :color="pill.color" append-icon="mdi-menu-down" :disabled="!!killing[tile.uid]">{{ pill.text }}</v-chip>
                </template>
                <v-list v-if="terminating" density="compact">
                    <v-list-item prepend-icon="mdi-stop-circle" title="강제 종료…" @click="confirmKill(tile.id, name, ['force'])" />
                </v-list>
                <v-list v-else density="compact">
                    <v-list-item prepend-icon="mdi-stop" title="종료…" @click="confirmKill(tile.id, name, ['kill'])" />
                    <v-list-item prepend-icon="mdi-close-box-outline" title="종료하고 닫기…" @click="confirmKill(tile.id, name, ['killClose'])" />
                </v-list>
            </v-menu>
            <v-chip v-else size="x-small" variant="tonal" class="pill" :color="pill.color" :title="proc?.exitCode != null ? `exit ${proc.exitCode}` : undefined">{{ pill.text }}</v-chip>
        </template>

        <v-menu v-if="boardMenu || canPin || terminalMenu" location="bottom end">
            <template #activator="{ props: act }">
                <v-btn v-bind="act" size="x-small" variant="text" icon="mdi-dots-vertical" title="메뉴" :loading="boardBusy" />
            </template>
            <v-list density="compact" min-width="180">
                <v-list-item v-if="terminalMenu"
                    :title="shortcutsOff ? '단축키 켜기' : '단축키 끄기'"
                    subtitle="이 타일이 active일 때 Ctrl+Alt 단축키"
                    :prepend-icon="shortcutsOff ? 'mdi-keyboard-outline' : 'mdi-keyboard-off-outline'"
                    @click="toggleShortcuts"
                />
                <v-list-item v-if="canPin" title="여기로 타일 고정" prepend-icon="mdi-pin-outline" @click="onPin" />
                <template v-if="boardMenu">
                    <v-list-item title="새 보드로 분리" prepend-icon="mdi-view-dashboard-outline" @click="detachOpen = true" />
                    <template v-if="otherBoards.length">
                        <v-list-subheader>다른 보드로 이동</v-list-subheader>
                        <v-list-item v-for="b in otherBoards" :key="b.hash"
                            :title="boardLabel(b)"
                            :prepend-icon="b.isDefault ? 'mdi-home-outline' : 'mdi-arrow-right'"
                            @click="onMove(b.hash)"
                        />
                    </template>
                </template>
            </v-list>
        </v-menu>

        <span class="win">
            <button v-if="tile.type === 'terminal' && !unavailable" title="최소화" @click="onMinimize">
                <v-icon size="14" icon="mdi-window-minimize" />
            </button>
            <v-menu v-if="!unavailable" ref="sizeMenu" location="bottom end" :close-on-content-click="false">
                <template #activator="{ props: act }">
                    <button v-bind="act" :class="{ 'text-warning': shownText }" :title="sizeTitle">
                        <v-icon size="14" icon="mdi-resize" />
                    </button>
                </template>
                <v-card class="size-pop">
                    <p class="size-now">{{ sizeTitle }}</p>
                    <div v-for="axis in (['w', 'h'] as const)" :key="axis" class="size-row">
                        <span>{{ axis === 'w' ? '가로' : '세로' }}</span>
                        <v-btn v-for="v in ([0.5, 1] as const)" :key="v"
                            size="x-small"
                            :variant="tile.size[axis] === v ? 'flat' : 'outlined'"
                            :color="tile.size[axis] === v ? 'primary' : undefined"
                            :disabled="v === 0.5 && cap[axis === 'w' ? 'x' : 'y'] === 1"
                            @click="setSize(axis, v)"
                        >{{ v.toFixed(1) }}</v-btn>
                    </div>
                    <p v-if="cap.x === 1 || cap.y === 1" class="size-hint">1칸인 축의 0.5는 이 화면에선 1.0으로 보임</p>
                </v-card>
            </v-menu>
            <button v-if="!isRoot"
                :disabled="unknown"
                :title="tile.type === 'terminal' && running && !readOnly ? (terminating ? '강제 종료·닫기…' : '종료…') : '닫기'"
                @click="onX"
            >
                <v-icon size="14" icon="mdi-close" />
            </button>
        </span>
    </div>

    <!-- 루트 타일 탭 `[파일] [터미널 N · 최소화 n]` -->
    <div v-if="isRoot" class="root-tabs" :inert="overview">
        <button :class="{ on: rootTab === 'files' }" @click="setTab('files')">파일</button>
        <button :class="{ on: rootTab === 'terminals' }" @click="setTab('terminals')">
            터미널 {{ runCount }}<template v-if="minCount"> · 최소화 {{ minCount }}</template>
        </button>
    </div>
    <span v-if="readOnly && !unavailable" class="ro">보기 전용 · 입력 안 됨</span>
    <div v-if="unavailable" class="revoked text-warning">
        <v-icon size="18" icon="mdi-link-off" />
        <span>공유가 취소되었거나 삭제되어 볼 수 없습니다.</span>
        <template v-if="isRoot && tile.type === 'folder'">
            <span class="sub">루트 타일은 닫을 수 없어 기본 목록으로 돌아갑니다.</span>
            <v-btn size="x-small" variant="tonal" color="primary" @click="navigate(tile.id, null)">기본으로 돌아가기</v-btn>
        </template>
        <span v-else class="sub">닫기만 할 수 있습니다.</span>
    </div>
    <template v-else-if="tile.type === 'folder'">
        <folder-tile-body v-show="!isRoot || rootTab === 'files'" :tile-id="tile.id" :inert="overview" />
        <root-terminal-list v-if="isRoot && rootTab === 'terminals'" :inert="overview" />
    </template>
    <terminal-tile-body v-else :tile-id="tile.id" :inert="overview" />
    <presence-dialog v-if="tile.type === 'terminal' && !unavailable" v-model="presenceOpen" :uid="tile.uid" :name="name" />
    <node-name-dialog v-if="boardMenu" v-model="detachOpen" title="새 보드" :initial="node?.name ?? ''" :loading="boardBusy" @submit="onDetach" />
</div>
</template>

<style scoped lang="scss">
.TileFrame {
    flex: 1;
    min-width: 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    border-radius: 6px;
    background: rgb(var(--v-theme-surface));
    position: relative;
}
.revoked {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
    padding: vars.$spacing-md;
    text-align: center;
    font-size: 12px;
    .sub { font-size: 11px; opacity: 0.7; }
}
// 전체보기: 안쪽 조작 막음(inert + 클릭은 Tile로)
.TileFrame--overview {
    > :not(.ov-name) { pointer-events: none; }
}
// active — 공유 테두리(border)와 겹쳐도 둘 다 보이게 outline. 포커스용 기본 outline은 끈다
.TileFrame:focus { outline: none; }
.TileFrame.TileFrame--active {
    outline: 2px solid rgb(var(--v-theme-primary));
    outline-offset: 1px;
    > .TileFrame__head { border-bottom-color: rgb(var(--v-theme-primary)); }
    // 전체보기 상자(overflow hidden)에 잘리지 않게 안쪽으로, 축소 배율을 거꾸로 곱해 화면에서 2px
    &.TileFrame--overview {
        outline-width: calc(2px / var(--s));
        outline-offset: calc(-2px / var(--s));
    }
}
.TileFrame--dim {
    opacity: 0.7;
    &:hover { opacity: 0.85; }
}
.ov-name {
    position: absolute;
    left: 0;
    top: 0;
    z-index: 3;
    width: calc(100% * var(--s));
    transform: scale(calc(1 / var(--s)));
    transform-origin: 0 0;
    display: flex;
    align-items: center;
    gap: vars.$spacing-sm;
    padding: 3px 6px;
    font-size: 11.5px;
    background: rgb(var(--v-theme-surface));
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    pointer-events: none;
    > * { flex: none; }
    .name { flex: 1; }
}
.TileFrame__head {
    flex: 0 0 28px;
    display: flex;
    align-items: center;
    gap: vars.$spacing-sm;
    padding: 0 vars.$spacing-xs 0 vars.$spacing-md;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    font-size: 12px;
    min-width: 0;
    > * { flex: none; }
}
.pos {
    min-width: 16px;
    height: 16px;
    padding: 0 4px;
    border-radius: 4px;
    display: inline-grid;
    place-items: center;
    font-size: 10px;
    font-weight: vars.$weight-lg;
    background: rgb(var(--v-theme-on-surface));
    color: rgb(var(--v-theme-surface));
}
.name {
    flex: 0 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-weight: vars.$weight-md;
}
.opener {
    flex: 0 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 11px;
}
.owner {
    font-family: monospace;
    font-size: 11px;
    opacity: 0.75;
}
.path {
    flex: 0 3 auto;     // 좁으면 경로부터 줄어든다
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 11px;
    opacity: 0.7;
}
.spacer { flex: 1; }
.keys-off { flex: none; }
.by {
    flex: 0 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 11px;
    opacity: 0.7;
}
.perm {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    font-size: 11px;
    color: rgb(var(--v-theme-share));
}
// ① 공유 색 — 왼쪽 띠 + 머리 줄 바탕 (테마 색 이름 share)
.TileFrame--shared {
    border-color: rgb(var(--v-theme-share));
    box-shadow: inset 3px 0 0 rgb(var(--v-theme-share));
    > .TileFrame__head { background: rgba(var(--v-theme-share), 0.12); }
}
.ro {
    position: absolute;
    right: 8px;
    bottom: 8px;
    z-index: 2;
    padding: 0 6px;
    font-size: 11px;
    border: 1px solid rgba(var(--v-theme-on-terminal), 0.35);
    border-radius: 4px;
    background: rgba(var(--v-theme-terminal), 0.7);
    color: rgb(var(--v-theme-on-terminal));
    pointer-events: none;
}
.pill { font-family: monospace; }
// 창 버튼 묶음 [최소화 | 크기 | ×] — kill은 여기서 떼어 RUN▾ 칩에(오클릭 방지)
.win {
    display: inline-flex;
    height: 20px;
    border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    border-radius: 4px;
    overflow: hidden;
    > button {
        width: 22px;
        display: grid;
        place-items: center;
        padding: 0;
        border: 0;
        background: none;
        color: inherit;
        cursor: pointer;
        & + button { border-left: 1px solid rgba(var(--v-border-color), var(--v-border-opacity)); }
        &:hover:not(:disabled) { background: rgba(var(--v-theme-on-surface), 0.08); }
        &:disabled { cursor: default; opacity: 0.3; }
    }
}
.root-tabs {
    flex: none;
    display: flex;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    font-size: 11.5px;
    > button {
        padding: 4px vars.$spacing-md;
        border: 0;
        border-bottom: 2px solid transparent;
        background: none;
        color: inherit;
        opacity: 0.65;
        cursor: pointer;
        &.on {
            opacity: 1;
            border-bottom-color: rgb(var(--v-theme-primary));
            font-weight: vars.$weight-md;
        }
    }
}
.size-pop {
    display: flex;
    flex-direction: column;
    gap: vars.$spacing-sm;
    padding: vars.$spacing-md vars.$spacing-lg;
    font-size: 12px;
}
.size-row {
    display: flex;
    align-items: center;
    gap: vars.$spacing-sm;
    > span { width: 30px; opacity: 0.7; }
}
.size-now {
    margin: 0;
    font-family: monospace;
    font-size: 10.5px;
    opacity: 0.7;
}
.size-hint {
    max-width: 180px;
    margin: 0;
    font-size: 10.5px;
    opacity: 0.7;
}
</style>
