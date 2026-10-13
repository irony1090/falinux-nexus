<script setup lang="ts">
import { tabReady } from '@/common/util/tabId.util';
import type { ProcessStatus } from '@/feature/process/api/process.api';
import { isRunning } from '@/feature/process/store/processTerm.store';
import { useWorkerRoster } from '@/feature/worker/api/worker.api';
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { VBtn, VIcon, VList, VListItem, VMenu, VProgressLinear } from 'vuetify/components';
import { useListChildren, useNodePath, type NodeResponse } from '../../api/node.api';
import { useNodeShares } from '../../api/share.api';
import { dragging, pasteBlock, useNodePaste, type PasteDest } from '../../hook/nodePaste.hook';
import { useScriptRuns } from '../../hook/scriptRuns.hook';
import { toClip, useNodeClipboard, type ClipMode, type ClipNode } from '../../store/nodeClipboard.store';
import { nodeActItems, useNodeOps, type NodeAct, type NodeActItem } from '../../store/nodeOps.store';
import { useNodeSelection } from '../../store/nodeSelection.store';
import { useTileTree } from '../../store/tileTree.store';
import { agoText } from '../../util/presenceView.util';
import { PERM_OWNER } from '../../util/shareView.util';
import ShareDialog from '../dialog/ShareDialog.vue';
import NodeRowMenu from './NodeRowMenu.vue';
import ScriptRunList from './ScriptRunList.vue';

const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const { tiles, navigate } = useTileTree();
const ops = useNodeOps();
const { saving, busyTile } = ops;

const nodeId = computed(() => {
    const t = tiles.value[props.tileId];
    return t?.type === 'folder' ? t.nodeId : null;
})

const listQuery = useListChildren(computed(() => nodeId.value ?? undefined));
const pathQuery = useNodePath(nodeId);
const { devices, refetch: refetchWorkers } = useWorkerRoster();

const path = computed(() => pathQuery.data.value ?? []);
const children = computed(() => listQuery.data.value ?? []);
const busy = computed(() => listQuery.isFetching.value || pathQuery.isFetching.value);
const error = computed(() => listQuery.error.value ?? pathQuery.error.value);

// 지금 보는 폴더와 실제로 적용되는 장비(자기 것, 없으면 가장 가까운 상위 폴더 것)
const current = computed(() => path.value[path.value.length - 1]);
const deviceOwner = computed(() => [...path.value].reverse().find(p => !!p.deviceKey));
const deviceInherited = computed(() => !!deviceOwner.value && deviceOwner.value.id !== current.value?.id);

// 공유 걸기 — O1(경로 줄 공유 버튼). 내 폴더(소유자 단계)일 때만. 서버도 소유자만 허용(SH4(공유 대상 찾기))
const canShare = computed(() => !!current.value && current.value.kind === 'FOLDER' && current.value.permission === PERM_OWNER);
const shareOpen = ref(false);
const shareCount = computed(() => sharesQuery.data.value?.length ?? 0);   // O1 숫자 = 이 폴더에 적용되는 사람 수(이어진 사람 포함)
const sharesQuery = useNodeShares(computed(() => current.value?.id), canShare);

// 장비 칩: 접속 수 / 명부 수 + title에 인스턴스 목록 — RS5(표시 위치) ③
const isOnline = (key: string | null | undefined) => !!key && (devices.value.get(key)?.online ?? 0) > 0;
const deviceCount = (key: string | null | undefined) => {
    const d = key ? devices.value.get(key) : undefined;
    return d ? `${d.online}/${d.instances.length}` : '';
}
const deviceTitle = (key: string | null | undefined, from?: string) => {
    if (!key) return '지정된 장비 없음';
    const d = devices.value.get(key);
    const head = [key, from && `상위 폴더 '${from}'에서 상속`].filter(v => !!v).join(' — ');
    if (!d) return `${head}\n접속한 적 없음`;
    const lines = d.instances.map(i => `${i.online ? '●' : '○'} #${i.subKey}  ${i.online ? '접속 중' : agoText(i.lastSeen)}`);
    return [`${head}\n접속 ${d.online} · 끊김 ${d.instances.length - d.online}`, ...lines].join('\n');
}

// 스크립트 행 "실행 중 N · 최소화 n" — 눌러서 펼침(실시간-6a)
const { byNode: runsByNode } = useScriptRuns(nodeId);
const runsOf = (id: number) => runsByNode.value.get(id) ?? [];
const runsText = (id: number) => {
    const runs = runsOf(id);
    const live = runs.filter(r => isRunning(r.status as ProcessStatus)).length;
    const min = runs.filter(r => r.minimized).length;
    return [live && `실행 중 ${live}`, runs.length > live && `끝남 ${runs.length - live}`, min && `최소화 ${min}`].filter(v => !!v).join(' · ');
}
const runsOpen = ref<Record<number, boolean>>({});

const onRefresh = () => {
    listQuery.refetch();
    refetchWorkers();
    if (nodeId.value !== null) pathQuery.refetch();
}

// ---- 선택 모드 — REF-node-select.md ----
const sel = useNodeSelection();
const isSel = (n: NodeResponse) => sel.ids.value.has(n.id);
const pick = (n: NodeResponse) => sel.toggle(n, props.tileId, nodeId.value);

// SM7(사라진 선택): 목록이 바뀌면 이 타일에서 고른 것을 맞춘다
watch(() => listQuery.data.value, list => {
    if (list) sel.syncTile(props.tileId, nodeId.value, list);
})

// 행 메뉴(⋮·우클릭) 항목 = "선택" + 행 동작 — SM1(켜는 방법)
const menuItems = (n: NodeResponse): NodeActItem[] => [
    { act: 'select', icon: isSel(n) ? 'mdi-checkbox-blank-outline' : 'mdi-checkbox-marked-outline', title: isSel(n) ? '선택 해제' : '선택' },
    ...nodeActItems(n, saving.value || !tabReady.value),
];
const onAct = (act: NodeAct, n: NodeResponse) => act === 'select' ? pick(n) : ops.act(act, props.tileId, [n]);

const LONG_PRESS_MS = 500;
const PRESS_MOVE_MAX = 10;     // px — 넘게 움직이면 스크롤로 보고 취소
let press: { timer: ReturnType<typeof setTimeout>; x: number; y: number } | null = null;
let skipClick = false;         // 길게 누르기 뒤따르는 click 무시
let lastPointer = 'mouse';
const dragOk = ref(false);     // 마우스로 누를 때만 draggable — SD3(터치 드래그 제외), 길게 누르기와 겹치지 않게

const cancelPress = () => {
    if (press) clearTimeout(press.timer);
    press = null;
}
onBeforeUnmount(cancelPress);

const onPointerDown = (e: PointerEvent, n: NodeResponse) => {
    lastPointer = e.pointerType;
    dragOk.value = e.pointerType === 'mouse';
    skipClick = false;
    cancelPress();
    if (e.pointerType === 'mouse') return;
    press = {
        x: e.clientX,
        y: e.clientY,
        timer: setTimeout(() => {
            press = null;
            skipClick = true;
            pick(n);
        }, LONG_PRESS_MS),
    };
}
const onPointerMove = (e: PointerEvent) => {
    if (press && Math.hypot(e.clientX - press.x, e.clientY - press.y) > PRESS_MOVE_MAX) cancelPress();
}

// 행 누르기: 모드 중 = 토글, 평소 Ctrl(⌘)+클릭 = 모드 시작(SM1). 행 안의 버튼은 각자 처리
const onRowClick = (e: MouseEvent, n: NodeResponse) => {
    if (skipClick) {
        skipClick = false;
        return;
    }
    if ((e.target as HTMLElement).closest('.v-btn, .runs')) return;
    if (sel.on.value || e.ctrlKey || e.metaKey) pick(n);
}
// 폴더 이름: 평소엔 이 타일 안에서 열기, 선택 중·Ctrl이면 행 누르기로 넘긴다
const onNameClick = (e: MouseEvent, n: NodeResponse) => {
    if (!sel.on.value && !e.ctrlKey && !e.metaKey) navigate(props.tileId, n.id);
}

// 우클릭 = 행 메뉴를 포인터 위치에. 터치 길게 누르기가 일으키는 contextmenu는 선택으로 처리했으므로 막기만
const ctxMenu = ref<{ node: NodeResponse; at: [number, number] }>();
const ctxOpen = ref(false);
const onContextMenu = (e: MouseEvent, n: NodeResponse) => {
    e.preventDefault();
    if (lastPointer !== 'mouse') return;
    ctxMenu.value = { node: n, at: [e.clientX, e.clientY] };
    ctxOpen.value = true;
}

// ---- 클립보드 붙여넣기 · 드롭 — REF-node-clipboard.md ----
const { clip, cutIds, removeIds, clear: clearClip } = useNodeClipboard();
const paste = useNodePaste();
const pasting = ref(false);
const tileBusy = computed(() => busy.value || pasting.value || busyTile.value === props.tileId);
const clipLabel = computed(() => clip.value.mode === 'cut' ? '잘라냄' : '복사됨');

const here = computed<PasteDest>(() => ({ id: nodeId.value, node: current.value, path: path.value }));
const pasteReason = computed(() => pasteBlock(clip.value.mode, clip.value.nodes, here.value));
const pasteTitle = computed(() => {
    const names = clip.value.nodes.map(n => n.name).join(', ');
    return `${clipLabel.value} ${clip.value.nodes.length}개: ${names}\n${pasteReason.value || '눌러서 이 폴더에 붙여넣기'}`;
})
// 선택 모드 중엔 머리 줄 버튼에 글자로 — 막대에서 담은 뒤 목적지를 고르는 흐름
const pasteText = computed(() => sel.on.value ? (clip.value.mode === 'cut' ? '여기로 이동' : '여기에 붙여넣기') : '');

const doPaste = async (mode: ClipMode, nodes: ReadonlyArray<ClipNode>, destId: number | null, fromClip: boolean) => {
    pasting.value = true;
    try {
        const done = await paste(mode, nodes, destId);
        if (fromClip && mode === 'cut') removeIds(done);    // CP4(붙여넣은 뒤): 잘라내기만 비움, 실패한 건 남겨 다시 시도
        if (mode === 'cut' && done.length && sel.on.value) sel.exit();   // SM5(끝낸 뒤): 이동 뒤 모드 종료
    } finally {
        pasting.value = false;
    }
}
const onPaste = () => doPaste(clip.value.mode, clip.value.nodes, nodeId.value, true);

// 드래그: 선택 중인 행을 끌면 고른 전체, 아니면 그 행. 클립보드는 건드리지 않는다 — SD4(드롭 = 이동, Ctrl = 복사)·SD6(HTML5 DnD)
const dropAt = ref<number | 'tile' | null>(null);   // 놓을 자리 표시: 폴더 행 id / 타일 본문(지금 폴더)
const rootEl = ref<HTMLElement>();

const onDragStart = (e: DragEvent, n: NodeResponse) => {
    const nodes = sel.on.value && isSel(n) ? sel.selected.value.map(it => it.node) : [n];
    dragging.value = nodes.map(toClip);
    e.dataTransfer?.setData('application/x-nexus-nodes', nodes.map(v => v.id).join(','));
    if (e.dataTransfer) e.dataTransfer.effectAllowed = 'copyMove';
}
const onDragEnd = () => dragging.value = null;
watch(dragging, v => {
    if (!v) dropAt.value = null;
})

// SD7(드롭 위치): 폴더 행 위 = 그 폴더(끄는 노드 자신은 제외), 그 밖 = 지금 폴더
const dropTarget = (e: DragEvent, nodes: ReadonlyArray<ClipNode>): { key: number | 'tile'; dest: PasteDest } => {
    const row = (e.target as HTMLElement).closest<HTMLElement>('[data-folder-id]');
    const n = row ? children.value.find(c => c.id === Number(row.dataset.folderId)) : undefined;
    if (n && !nodes.some(d => d.id === n.id)) return { key: n.id, dest: { id: n.id, node: n, path: [...path.value, n] } };
    return { key: 'tile', dest: here.value };
}
const dropMode = (e: DragEvent): ClipMode => e.ctrlKey ? 'copy' : 'cut';

// 막힌 자리면 preventDefault를 안 한다 = 브라우저가 놓기 금지 커서
const onDragOver = (e: DragEvent) => {
    const nodes = dragging.value;
    if (!nodes || error.value) return;
    const t = dropTarget(e, nodes);
    const mode = dropMode(e);
    if (pasteBlock(mode, nodes, t.dest)) {
        dropAt.value = null;
        return;
    }
    e.preventDefault();
    if (e.dataTransfer) e.dataTransfer.dropEffect = mode === 'copy' ? 'copy' : 'move';
    dropAt.value = t.key;
}
const onDragLeave = (e: DragEvent) => {
    if (!rootEl.value?.contains(e.relatedTarget as Node | null)) dropAt.value = null;
}
const onDrop = (e: DragEvent) => {
    const nodes = dragging.value;
    dropAt.value = null;
    if (!nodes) return;
    const t = dropTarget(e, nodes);
    const mode = dropMode(e);
    if (pasteBlock(mode, nodes, t.dest)) return;
    e.preventDefault();
    dragging.value = null;
    void doPaste(mode, nodes, t.dest.id, false);
}
</script>

<template>
<div ref="rootEl" class="FolderTileBody"
    :class="{ drop: dropAt === 'tile' }"
    @dragover="onDragOver"
    @dragleave="onDragLeave"
    @drop="onDrop"
>
    <div class="crumbs">
        <button v-if="nodeId !== null" @click="navigate(tileId, null)">기본</button>
        <span v-else class="cur">기본</span>
        <template v-for="(p, i) in path" :key="p.id">
            <span>/</span>
            <span v-if="i === path.length - 1" class="cur">{{ p.name }}</span>
            <button v-else @click="navigate(tileId, p.id)">{{ p.name }}</button>
        </template>
        <span v-if="nodeId !== null && !path.length">/ …</span>
        <span class="tools">
            <button v-if="current" class="dev"
                :class="{ on: isOnline(deviceOwner?.deviceKey) }"
                :title="`${deviceTitle(deviceOwner?.deviceKey, deviceInherited ? deviceOwner?.name : undefined)}\n(눌러서 장비 지정·명부 관리)`"
                :disabled="saving"
                @click="ops.editDevice(tileId, current)"
            >
                <i /><span class="key">{{ deviceOwner?.deviceKey ?? '장비 없음' }}</span><span v-if="deviceCount(deviceOwner?.deviceKey)" class="cnt">{{ deviceCount(deviceOwner?.deviceKey) }}</span><span v-if="deviceInherited">(상속)</span>
            </button>
            <button v-if="canShare" class="tool share" title="이 폴더 공유" @click="shareOpen = true">
                <v-icon size="14" icon="mdi-account-multiple-plus" /><span v-if="shareCount" class="share-count">{{ shareCount }}</span>
            </button>
            <v-menu location="bottom end">
                <template #activator="{ props: act }">
                    <button v-bind="act" class="tool" title="새로 만들기" :disabled="saving">
                        <v-icon size="14" icon="mdi-plus" />
                    </button>
                </template>
                <v-list density="compact">
                    <v-list-item prepend-icon="mdi-folder-plus-outline" title="새 폴더" @click="ops.editName(tileId, { kind: 'FOLDER', parentId: nodeId })" />
                    <v-list-item prepend-icon="mdi-file-plus-outline" title="새 스크립트" @click="ops.editName(tileId, { kind: 'SCRIPT', parentId: nodeId })" />
                </v-list>
            </v-menu>
            <span v-if="clip.nodes.length" class="paste" :class="{ wide: !!pasteText }" :title="pasteTitle">
                <button class="tool" :disabled="saving || pasting || !!pasteReason" @click="onPaste">
                    <v-icon size="14" :icon="clip.mode === 'cut' ? 'mdi-folder-move-outline' : 'mdi-content-paste'" /><span v-if="pasteText">{{ pasteText }}</span><span class="paste-count">{{ clip.nodes.length }}</span>
                </button>
                <button class="tool" title="클립보드 비우기" @click="clearClip"><v-icon size="12" icon="mdi-close" /></button>
            </span>
            <button class="tool" title="새로고침" :disabled="busy" @click="onRefresh">
                <v-icon size="14" icon="mdi-refresh" />
            </button>
        </span>
    </div>
    <div class="bar">
        <v-progress-linear v-if="tileBusy" indeterminate height="2" color="primary" />
    </div>
    <div v-if="error" class="state">
        <span>{{ error.message || '목록을 불러오지 못했습니다' }}</span>
        <v-btn size="x-small" variant="tonal" :disabled="busy" @click="onRefresh">다시 시도</v-btn>
    </div>
    <template v-else>
        <template v-for="(n, i) in children" :key="n.id">
        <div class="row"
            :class="{ cut: cutIds.has(n.id), sel: isSel(n), drop: dropAt === n.id, picking: sel.on.value }"
            :data-folder-id="n.kind === 'FOLDER' ? n.id : undefined"
            :draggable="dragOk"
            @click="onRowClick($event, n)"
            @pointerdown="onPointerDown($event, n)"
            @pointermove="onPointerMove"
            @pointerup="cancelPress"
            @pointercancel="cancelPress"
            @contextmenu="onContextMenu($event, n)"
            @dragstart="onDragStart($event, n)"
            @dragend="onDragEnd"
        >
            <v-icon v-if="sel.on.value" size="14" class="check" :icon="isSel(n) ? 'mdi-checkbox-marked' : 'mdi-checkbox-blank-outline'" />
            <span v-else class="ord">{{ i + 1 }}</span>
            <v-icon size="14"
                :icon="n.kind === 'FOLDER' ? (n.ownerNickname ? 'mdi-folder-account' : 'mdi-folder') : 'mdi-console'"
                :class="n.kind === 'FOLDER' ? 'folder' : 'script'"
            />
            <button v-if="n.kind === 'FOLDER'" class="nm" :title="sel.on.value ? '' : '이 타일 안에서 열기'" @click="onNameClick($event, n)">{{ n.name }}</button>
            <span v-else class="nm">{{ n.name }}</span>
            <span class="tail">
                <span v-if="n.ownerNickname" class="by" :title="`${n.ownerNickname}님이 공유한 폴더`">← {{ n.ownerNickname }}</span>
                <span v-if="n.shareCount" class="shared" :title="`${n.shareCount}명에게 공유 중`"><v-icon size="11" icon="mdi-account-multiple" />{{ n.shareCount }}</span>
                <button v-if="n.kind === 'SCRIPT' && runsOf(n.id).length" class="runs" :class="{ on: runsOpen[n.id] }" title="이 스크립트의 터미널(실행 중 + 내 타일에 남은 끝난 것)" @click="runsOpen[n.id] = !runsOpen[n.id]">{{ runsText(n.id) }}</button>
                <span v-if="n.deviceKey" class="dev" :class="{ on: isOnline(n.deviceKey) }" :title="deviceTitle(n.deviceKey)">
                    <i /><span class="key">{{ n.deviceKey }}</span><span v-if="deviceCount(n.deviceKey)" class="cnt">{{ deviceCount(n.deviceKey) }}</span>
                </span>
                <!-- 선택 모드 중엔 숨김 — 행 누르기 = 선택과 겹치지 않게 -->
                <template v-if="!sel.on.value">
                    <v-btn v-if="n.kind === 'FOLDER'" size="x-small" variant="tonal" color="warning" @click="ops.open(tileId, n.id)">새 타일</v-btn>
                    <v-btn v-else size="x-small" variant="tonal" color="primary" :disabled="saving || !tabReady" @click="ops.exec(tileId, n)">실행</v-btn>
                    <node-row-menu :items="menuItems(n)" @act="a => onAct(a, n)" />
                </template>
            </span>
        </div>
        <script-run-list v-if="n.kind === 'SCRIPT' && runsOpen[n.id] && runsOf(n.id).length" :runs="runsOf(n.id)" :tile-id="tileId" />
        </template>
        <p v-if="!listQuery.isPending.value && !children.length" class="empty">비어 있음</p>
    </template>

    <node-row-menu v-if="ctxMenu" v-model="ctxOpen" :at="ctxMenu.at" :items="menuItems(ctxMenu.node)" @act="a => ctxMenu && onAct(a, ctxMenu.node)" />
    <share-dialog v-if="canShare && current" v-model="shareOpen" :node-id="current.id" :node-name="current.name" :tile-id="tileId" />
</div>
</template>

<style scoped lang="scss">
.FolderTileBody {
    flex: 1;
    min-height: 0;
    overflow: auto;
    font-size: 12px;
    button {
        padding: 0;
        border: 0;
        background: none;
        color: inherit;
        font: inherit;
        cursor: pointer;
    }
}
.crumbs {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 2px;
    padding: 5px vars.$spacing-md;
    font-size: 11px;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    button { color: rgb(var(--v-theme-primary)); }
    button.dev { color: inherit; }
    .cur { font-weight: vars.$weight-md; }
    .tools {
        margin-left: auto;
        display: flex;
        align-items: center;
        gap: vars.$spacing-md;
    }
    .tool {
        color: inherit;
        opacity: 0.7;
        &:disabled { cursor: default; opacity: 0.3; }
        &.share { display: inline-flex; align-items: center; gap: 3px; color: rgb(var(--v-theme-share)); opacity: 1; }
        .share-count { font-size: 11px; }
    }
    .paste {
        display: inline-flex;
        align-items: center;
        gap: 2px;
        .tool:first-child { display: inline-flex; align-items: center; gap: 2px; opacity: 1; &:disabled { opacity: 0.3; } }
        &.wide .tool:first-child { padding: 0 6px; border-radius: 8px; color: rgb(var(--v-theme-primary)); background: rgba(var(--v-theme-primary), 0.12); }
        .paste-count { font-size: 11px; font-family: monospace; }
    }
}
.bar { height: 2px; }
.state {
    display: flex;
    align-items: center;
    gap: vars.$spacing-md;
    margin: vars.$spacing-md;
    color: rgb(var(--v-theme-error));
}
.row {
    display: grid;
    grid-template-columns: 20px 16px minmax(0, 1fr) auto;
    align-items: center;
    gap: 6px;
    min-height: 30px;
    padding: vars.$spacing-sm vars.$spacing-xs vars.$spacing-sm vars.$spacing-md;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    user-select: none;
    -webkit-touch-callout: none;    // 길게 누르기 = 선택(SM1) — iOS 기본 메뉴 끔
    &.cut { opacity: 0.45; }    // CP2(표시 구분): 잘라낸 것만 흐리게
    &.picking { cursor: pointer; }
    &.sel { background: rgba(var(--v-theme-primary), 0.12); }
    &.drop { outline: 1px dashed rgb(var(--v-theme-primary)); outline-offset: -1px; background: rgba(var(--v-theme-primary), 0.08); }
}
.FolderTileBody.drop { box-shadow: inset 0 0 0 1px rgb(var(--v-theme-primary)); }
.check { color: rgb(var(--v-theme-primary)); }
.ord { font-family: monospace; font-size: 10px; opacity: 0.6; }
.folder { color: rgb(var(--v-theme-warning)); }
.script { color: rgb(var(--v-theme-primary)); }
.nm {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-align: left;
}
button.nm:hover { color: rgb(var(--v-theme-primary)); }
// 공유받은 루트 폴더의 소유자
.by {
    font-size: 10.5px;
    white-space: nowrap;
    color: rgb(var(--v-theme-warning));
}
.shared {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    padding: 0 6px;
    border-radius: 8px;
    font-size: 10.5px;
    white-space: nowrap;
    color: rgb(var(--v-theme-share));
    background: rgba(var(--v-theme-share), 0.12);
}
button.runs {
    padding: 0 6px;
    border-radius: 8px;
    font-size: 10.5px;
    white-space: nowrap;
    color: rgb(var(--v-theme-success));
    background: rgba(var(--v-theme-success), 0.12);
    &.on { background: rgba(var(--v-theme-success), 0.25); }
}
.tail {
    display: flex;
    align-items: center;
    gap: vars.$spacing-sm;
}
// 지정된 장비 + 접속 여부(점 색)
.dev {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    max-width: 140px;
    white-space: nowrap;
    font-size: 10.5px;
    opacity: 0.75;
    .key {
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
    }
    i {
        flex: none;
        width: 6px;
        height: 6px;
        border-radius: 50%;
        background: rgba(var(--v-theme-on-surface), var(--v-disabled-opacity));
    }
    .cnt { flex: none; font-family: monospace; }
    &.on i { background: rgb(var(--v-theme-success)); }
    &.on .cnt { color: rgb(var(--v-theme-success)); }
}
.empty { margin: vars.$spacing-md; opacity: 0.6; }
</style>
