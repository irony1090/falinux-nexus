<script setup lang="ts">
import { tabReady } from '@/common/util/tabId.util';
import type { ProcessStatus } from '@/feature/process/api/process.api';
import { isRunning } from '@/feature/process/store/processTerm.store';
import { useWorkerRoster } from '@/feature/worker/api/worker.api';
import { osName } from '@/feature/worker/util/workerOs.util';
import { computed, ref, watch } from 'vue';
import { VBtn, VIcon, VList, VListItem, VMenu, VProgressLinear } from 'vuetify/components';
import { suggestNodes, useListChildren, useNodePath, type NodeResponse } from '../../api/node.api';
import { useNodeShares } from '../../api/share.api';
import { useItemPress } from '../../hook/itemPress.hook';
import { dragging, pasteBlock, useNodePaste, type PasteDest } from '../../hook/nodePaste.hook';
import { usePathEditKey } from '../../hook/pathEdit.hook';
import { useScriptRuns } from '../../hook/scriptRuns.hook';
import { toClip, useNodeClipboard, type ClipMode, type ClipNode } from '../../store/nodeClipboard.store';
import { nodeActItems, useNodeOps, type NodeAct } from '../../store/nodeOps.store';
import { useNodeSelection } from '../../store/nodeSelection.store';
import { useTileTree } from '../../store/tileTree.store';
import { agoText } from '../../util/presenceView.util';
import { PERM_EDIT, PERM_OWNER } from '../../util/shareView.util';
import ShareDialog from '../dialog/ShareDialog.vue';
import DirGridCell from './DirGridCell.vue';
import NodeRowMenu, { type RowMenuItem } from './NodeRowMenu.vue';
import PathInput, { type PathCand } from './PathInput.vue';
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

// 경로 직접 입력 — 경로-7(화면: 폴더 타일 입력), REF-path-input.md. 시작 = 기본 목록(앞의 `기본/`은 있어도 없어도)
// 이름 속 `/`·`\`는 `\/`·`\\`로 적는다(Q44(이름 속 /) = 목록에서 골라야 들어감). 같은 이름은 첫 번째(Q43(같은 이름 폴더))
const SUGGEST_MAX = 50
const ROOT_NAME = '기본'
const editing = ref(false);
const openEdit = () => { editing.value = true; };
usePathEditKey(() => props.tileId, openEdit);
const onCrumbsClick = (e: MouseEvent) => {
    if (!editing.value && !(e.target as Element).closest('button, input, .tools')) openEdit();
}
const escName = (n: string) => n.replace(/\\/g, '\\\\').replace(/\//g, '\\/');
// 고른 단계 = 그 글자(끝 `/` 없이) -> 폴더 id. 그 앞을 고치면 글자가 안 맞아 저절로 안 쓰임
let picks = new Map<string, number>();
const pathText = computed(() => [ROOT_NAME, ...path.value.map(p => escName(p.name))].join('/'));
watch(editing, on => {
    if (!on) return;
    picks = new Map();
    path.value.forEach((p, i) => picks.set([ROOT_NAME, ...path.value.slice(0, i + 1).map(q => escName(q.name))].join('/'), p.id));
});
// 글자 -> 단계(`\x` = 글자 x). 마지막 칸 = 아직 끝나지 않은 앞글자, head = 그 앞까지의 글자
const splitNames = (t: string) => {
    const names = [''];
    let head = 0;
    for (let i = 0; i < t.length; i++) {
        const ch = t[i]!;
        if (ch === '\\' && i + 1 < t.length) names[names.length - 1] += t[++i];
        else if (ch === '/') { names.push(''); head = i + 1; }
        else names[names.length - 1] += ch;
    }
    return { names, head };
}
// 가장 긴 고른 단계에서 출발(exact = 그 단계 글자 그대로도). 없으면 기본 목록에서(맨 앞 빈 칸·`기본` 칸은 뺀다)
const resolveBase = (t: string, exact: boolean) => {
    let key = '', baseId = 0;
    for (const [k, id] of picks) {
        if (k.length > key.length && ((exact && t === k) || t.startsWith(k + '/'))) { key = k; baseId = id; }
    }
    if (baseId) return { baseId, names: t === key ? [] : splitNames(t.slice(key.length + 1)).names };
    const { names } = splitNames(t);
    if (names[0] === '' && names.length > 1) names.shift();
    if (names[0] === ROOT_NAME && names.length > 1) names.shift();
    return { baseId, names };
}
const suggestPath = async (t: string) => {
    const { baseId, names } = resolveBase(t, false);
    const prefix = names.pop() ?? '';
    const r = await suggestNodes({ baseId, names, prefix, max: SUGGEST_MAX });
    const head = t.slice(0, splitNames(t).head);
    const seen = new Map<string, number>();
    const items = r.items.map<PathCand & { id: number }>(i => {
        const k = (seen.get(i.name) ?? 0) + 1;
        seen.set(i.name, k);
        return { key: String(i.id), id: i.id, label: i.name, fill: head + escName(i.name) + '/', hint: k > 1 ? `같은 이름 ${k}번째` : undefined };
    });
    return { items, more: r.more };
}
const onPick = (c: PathCand) => {
    const id = (c as PathCand & { id?: number }).id;
    if (id !== undefined) picks.set(c.fill.slice(0, -1), id);
}
const submitPath = async (t: string) => {
    const { baseId, names } = resolveBase(t.trim(), true);
    if (names[names.length - 1] === '') names.pop();     // 끝 `/`
    if (!baseId && (!names.length || (names.length === 1 && names[0] === ROOT_NAME))) {
        navigate(props.tileId, null);
        return undefined;
    }
    try {
        navigate(props.tileId, (await suggestNodes({ baseId, names })).id);
    } catch (e) {
        return (e as { message?: string }).message || '이동할 수 없습니다';
    }
    return undefined;
}

// 장비 디렉토리 열기 — 폴더-3(여는 곳). 내 폴더 + 적용 장비(상속 포함)가 있을 때만(Q5·Q12)
const canOpenDir = computed(() => !!current.value && current.value.kind === 'FOLDER' && current.value.permission === PERM_OWNER && !!deviceOwner.value);

// 공유 걸기 — O1(경로 줄 공유 버튼). 내 폴더(소유자 단계)일 때만. 서버도 소유자만 허용(SH4(공유 대상 찾기))
const canShare = computed(() => !!current.value && current.value.kind === 'FOLDER' && current.value.permission === PERM_OWNER);
const shareOpen = ref(false);
const shareFor = ref<NodeResponse>();     // 공유 창 대상 = 지금 폴더(경로 줄) 또는 칸의 폴더(⋮·공유 라벨)
const openShare = (n: NodeResponse) => {
    shareFor.value = n;
    shareOpen.value = true;
}
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
    const lines = d.instances.map(i => `${i.online ? '●' : '○'} #${i.subKey} (${osName(i.os)})  ${i.online ? '접속 중' : agoText(i.lastSeen)}`);
    return [`${head}\n접속 ${d.online} · 끊김 ${d.instances.length - d.online}`, ...lines].join('\n');
}

// 스크립트 칸 실행 목록 표식(숫자 = 터미널 수) — 눌러서 칸 아래 팝업(실시간-6a, 아이콘-Q4(실행 목록))
const { byNode: runsByNode } = useScriptRuns(nodeId);
const runsOf = (id: number) => runsByNode.value.get(id) ?? [];
const runsText = (id: number) => {
    const runs = runsOf(id);
    const live = runs.filter(r => isRunning(r.status as ProcessStatus)).length;
    const min = runs.filter(r => r.minimized).length;
    return [live && `실행 중 ${live}`, runs.length > live && `끝남 ${runs.length - live}`, min && `최소화 ${min}`].filter(v => !!v).join(' · ');
}
const runsLive = (id: number) => runsOf(id).filter(r => isRunning(r.status as ProcessStatus)).length;

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

// 칸 메뉴(⋮·우클릭) 항목 = "선택" + 행 동작 + 공유(내 폴더만, 삭제 앞) — SM1(켜는 방법), REF-dir-grid.md "추가 요청"
type RowAct = NodeAct | 'share';
const ownFolder = (n: NodeResponse) => n.kind === 'FOLDER' && n.permission === PERM_OWNER;
const canEditDevice = (n: NodeResponse) => n.kind === 'FOLDER' && (n.permission ?? 0) >= PERM_EDIT;
const menuItems = (n: NodeResponse): RowMenuItem<RowAct>[] => {
    const items: RowMenuItem<RowAct>[] = [
        { act: 'select', icon: isSel(n) ? 'mdi-checkbox-blank-outline' : 'mdi-checkbox-marked-outline', title: isSel(n) ? '선택 해제' : '선택' },
        ...nodeActItems(n, saving.value || !tabReady.value),
    ];
    if (ownFolder(n)) items.splice(items.length - 1, 0, { act: 'share', icon: 'mdi-account-multiple-plus', title: '공유' });
    return items;
}
// ⋮ 메뉴 맨 위 정보 카드 — 판 D′(hover 대신). 상태 라벨의 뜻을 글자로
const infoRows = (n: NodeResponse): [string, string][] => {
    const d = n.deviceKey ? devices.value.get(n.deviceKey) : undefined;
    return ([
        ['종류', n.kind === 'SCRIPT' ? '스크립트' : n.ownerNickname ? '공유받은 폴더' : '폴더'],
        n.ownerNickname && ['소유자', n.ownerNickname],
        n.deviceKey && ['장비', `${n.deviceKey} · ${d ? `접속 ${d.online} / ${d.instances.length}` : '접속한 적 없음'}`],
        n.shareCount && ['공유', `${n.shareCount}명에게 공유 중`],
        n.kind === 'SCRIPT' && runsOf(n.id).length && ['터미널', runsText(n.id)],
    ] as const).filter((r): r is [string, string] => !!r);
}
const onAct = (act: RowAct, n: NodeResponse) => {
    if (act === 'select') pick(n);
    else if (act === 'share') openShare(n);
    else ops.act(act, props.tileId, [n]);
}
// 장비·공유 라벨 누르기 = 지정 창. 선택 모드·Ctrl이면 칸 누르기(체크)로 넘긴다
const onLabel = (e: MouseEvent, run: () => void) => {
    if (!sel.on.value && !e.ctrlKey && !e.metaKey) run();
}

// 칸 누르기 — REF-dir-grid.md "누르기 규칙". 더블 = 폴더는 이 타일 안에서 열기, 스크립트는 실행
const press = useItemPress<NodeResponse, number>({
    key: n => n.id,
    selecting: () => sel.on.value,
    check: pick,
    open: n => {
        if (n.kind === 'FOLDER') navigate(props.tileId, n.id);
        else if (!saving.value && tabReady.value) void ops.exec(props.tileId, n);
    },
});
const { focused, dragOk } = press;
watch(nodeId, () => focused.value = null);
const onBodyClick = (e: MouseEvent) => {
    if (!(e.target as HTMLElement).closest('.DirGridCell')) focused.value = null;
}

// 우클릭 = 행 메뉴를 포인터 위치에. 터치 길게 누르기가 일으키는 contextmenu는 선택으로 처리했으므로 막기만
const ctxMenu = ref<{ node: NodeResponse; at: [number, number] }>();
const ctxOpen = ref(false);
const onContextMenu = (e: MouseEvent, n: NodeResponse) => {
    e.preventDefault();
    if (press.lastPointer() !== 'mouse') return;
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
    @click="onBodyClick"
    @dragover="onDragOver"
    @dragleave="onDragLeave"
    @drop="onDrop"
>
    <div class="crumbs" :class="{ editing }" @click="onCrumbsClick">
        <path-input v-if="editing" :value="pathText" :suggest="suggestPath" :submit="submitPath" @pick="onPick" @close="editing = false" />
        <!-- 입력 중엔 도구를 숨겨 입력칸이 줄 전체를 쓴다(좁은 타일에서 자동완성 목록 폭 확보) -->
        <template v-else>
            <!-- [경로·새로고침] / [도구] — 경로 자리가 최소 폭보다 좁아지면 도구가 둘째 줄(오른쪽 끝)로 -->
            <span class="lead">
                <span class="path">
                    <button v-if="nodeId !== null" @click="navigate(tileId, null)">기본</button>
                    <span v-else class="cur">기본</span>
                    <template v-for="(p, i) in path" :key="p.id">
                        <span class="sep">/</span>
                        <span v-if="i === path.length - 1" class="cur">{{ p.name }}</span>
                        <button v-else @click="navigate(tileId, p.id)">{{ p.name }}</button>
                    </template>
                    <span v-if="nodeId !== null && !path.length">/ …</span>
                </span>
                <button class="tile-btn refresh" title="새로고침" :disabled="busy" @click="onRefresh">
                    <v-icon size="15" icon="mdi-refresh" />
                </button>
            </span>
            <span class="tools tile-btn-row-end">
                <button v-if="current" class="tile-btn dev"
                    :class="{ on: isOnline(deviceOwner?.deviceKey) }"
                    :title="`${deviceTitle(deviceOwner?.deviceKey, deviceInherited ? deviceOwner?.name : undefined)}\n(눌러서 장비 지정·명부 관리)`"
                    :disabled="saving"
                    @click="ops.editDevice(tileId, current)"
                >
                    <i /><span class="key">{{ deviceOwner?.deviceKey ?? '장비 없음' }}</span><span v-if="deviceCount(deviceOwner?.deviceKey)" class="cnt">{{ deviceCount(deviceOwner?.deviceKey) }}</span><span v-if="deviceInherited">(상속)</span>
                </button>
                <span v-if="current && (canOpenDir || canShare)" class="tile-btn-sep" />
                <button v-if="canOpenDir" class="tile-btn" :class="{ 'text-success': isOnline(deviceOwner?.deviceKey) }"
                    :disabled="saving || !isOnline(deviceOwner?.deviceKey)"
                    :title="isOnline(deviceOwner?.deviceKey) ? '장비 디렉토리 열기' : '접속 중인 인스턴스 없음'"
                    @click="current && ops.openDir(tileId, current)"
                >
                    <v-icon size="16" icon="mdi-folder-network" />
                </button>
                <button v-if="canShare" class="tile-btn share" title="이 폴더 공유" @click="current && openShare(current)">
                    <v-icon size="16" icon="mdi-account-multiple-plus" /><span v-if="shareCount" class="share-count">{{ shareCount }}</span>
                </button>
                <template v-if="clip.nodes.length">
                    <span v-if="current || canOpenDir || canShare" class="tile-btn-sep" />
                    <button class="tile-btn paste" :class="{ 'tile-btn--text tile-btn--tonal': !!pasteText }" :title="pasteTitle"
                        :disabled="saving || pasting || !!pasteReason" @click="onPaste"
                    >
                        <v-icon size="16" :icon="clip.mode === 'cut' ? 'mdi-folder-move-outline' : 'mdi-content-paste'" /><span v-if="pasteText">{{ pasteText }}</span><span class="paste-count">{{ clip.nodes.length }}</span>
                    </button>
                    <button class="tile-btn" title="클립보드 비우기" @click="clearClip"><v-icon size="14" icon="mdi-close" /></button>
                </template>
                <span class="tile-btn-end tile-btn-end--primary">
                    <v-menu location="bottom end">
                        <template #activator="{ props: act }">
                            <button v-bind="act" class="tile-btn tile-btn--text" title="새로 만들기" :disabled="saving">
                                <v-icon size="15" icon="mdi-plus" />새로
                            </button>
                        </template>
                        <v-list density="compact">
                            <v-list-item prepend-icon="mdi-folder-plus-outline" title="새 폴더" @click="ops.editName(tileId, { kind: 'FOLDER', parentId: nodeId })" />
                            <v-list-item prepend-icon="mdi-file-plus-outline" title="새 스크립트" @click="ops.editName(tileId, { kind: 'SCRIPT', parentId: nodeId })" />
                        </v-list>
                    </v-menu>
                </span>
            </span>
        </template>
    </div>
    <div class="bar">
        <v-progress-linear v-if="tileBusy" indeterminate height="2" color="primary" />
    </div>
    <div v-if="error" class="state">
        <span>{{ error.message || '목록을 불러오지 못했습니다' }}</span>
        <v-btn size="x-small" variant="tonal" :disabled="busy" @click="onRefresh">다시 시도</v-btn>
    </div>
    <template v-else>
        <div v-if="children.length" class="dir-grid">
            <dir-grid-cell v-for="n in children" :key="n.id"
                :class="{ cut: cutIds.has(n.id), drop: dropAt === n.id }"
                :data-folder-id="n.kind === 'FOLDER' ? n.id : undefined"
                :draggable="dragOk"
                :icon="n.kind === 'FOLDER' ? (n.ownerNickname ? 'mdi-folder-account' : 'mdi-folder') : 'mdi-console'"
                :tone="n.kind === 'FOLDER' ? 'warning' : 'primary'"
                :name="n.name"
                :checking="sel.on.value"
                :checked="isSel(n)"
                :focused="focused === n.id"
                @click="press.onClick($event, n)"
                @pointerdown="press.onPointerDown($event, n)"
                @pointermove="press.onPointerMove"
                @pointerup="press.cancelPress"
                @pointercancel="press.cancelPress"
                @contextmenu="onContextMenu($event, n)"
                @dragstart="onDragStart($event, n)"
                @dragend="onDragEnd"
            >
                <template #action>
                    <button v-if="n.kind === 'FOLDER'" class="dg-btn dg-btn--l" style="--c: var(--v-theme-warning)" title="새 타일에서 열기" @click="ops.open(tileId, n.id)">새 타일</button>
                    <button v-else class="dg-btn dg-btn--l" style="--c: var(--v-theme-primary)" title="새 터미널 타일에서 실행" :disabled="saving || !tabReady" @click="ops.exec(tileId, n)">
                        <v-icon size="11" icon="mdi-play" />실행
                    </button>
                </template>
                <template #menu>
                    <node-row-menu :items="menuItems(n)" @act="a => onAct(a, n)">
                        <template #head><div class="dg-info"><span class="t">{{ n.name }}</span><template v-for="[k, v] in infoRows(n)" :key="k"><span class="k">{{ k }}</span><span class="v">{{ v }}</span></template></div></template>
                    </node-row-menu>
                </template>
                <template #right>
                    <component :is="ownFolder(n) ? 'button' : 'span'" v-if="n.shareCount" class="dg-lbl dg-lbl--solid" style="--c: var(--v-theme-share); --on-c: var(--v-theme-on-share)"
                        :title="`${n.shareCount}명에게 공유 중${ownFolder(n) ? ' — 눌러서 공유 관리' : ''}`"
                        @click="ownFolder(n) && onLabel($event, () => openShare(n))"
                    >
                        <v-icon size="10" icon="mdi-account-multiple" />{{ n.shareCount }}
                    </component>
                    <v-menu v-if="n.kind === 'SCRIPT' && runsOf(n.id).length" location="bottom end" :offset="72" :close-on-content-click="false" :disabled="sel.on.value">
                        <template #activator="{ props: act }">
                            <button v-bind="act" class="dg-lbl" :class="{ 'dg-lbl--solid': runsLive(n.id) > 0 }"
                                :style="runsLive(n.id) > 0 ? '--c: var(--v-theme-success); --on-c: var(--v-theme-on-success)' : ''"
                                :title="`이 스크립트의 터미널 — ${runsText(n.id)}`"
                            >
                                <v-icon size="10" icon="mdi-play" />{{ runsOf(n.id).length }}
                            </button>
                        </template>
                        <div class="dg-pop"><script-run-list :runs="runsOf(n.id)" :tile-id="tileId" /></div>
                    </v-menu>
                </template>
                <template #left>
                    <component :is="canEditDevice(n) ? 'button' : 'span'" v-if="n.deviceKey" class="dg-lbl" :style="isOnline(n.deviceKey) ? '--c: var(--v-theme-success)' : ''"
                        :title="`${deviceTitle(n.deviceKey)}${canEditDevice(n) ? '\n(눌러서 장비 지정)' : ''}`"
                        :disabled="canEditDevice(n) && saving"
                        @click="canEditDevice(n) && onLabel($event, () => ops.editDevice(tileId, n))"
                    >
                        <i class="dot" /><span class="k">{{ n.deviceKey }}</span>{{ devices.get(n.deviceKey)?.online || '' }}
                    </component>
                    <span v-if="n.ownerNickname" class="dg-lbl" style="--c: var(--v-theme-warning)" :title="`${n.ownerNickname}님이 공유한 폴더`">
                        <span class="k">← {{ n.ownerNickname }}</span>
                    </span>
                </template>
            </dir-grid-cell>
        </div>
        <p v-if="!listQuery.isPending.value && !children.length" class="empty">비어 있음</p>
    </template>

    <node-row-menu v-if="ctxMenu" v-model="ctxOpen" :at="ctxMenu.at" :items="menuItems(ctxMenu.node)" @act="a => ctxMenu && onAct(a, ctxMenu.node)">
        <template #head><div class="dg-info"><span class="t">{{ ctxMenu.node.name }}</span><template v-for="[k, v] in infoRows(ctxMenu.node)" :key="k"><span class="k">{{ k }}</span><span class="v">{{ v }}</span></template></div></template>
    </node-row-menu>
    <share-dialog v-if="shareFor" v-model="shareOpen" :node-id="shareFor.id" :node-name="shareFor.name" :tile-id="tileId" />
</div>
</template>

<style scoped lang="scss">
.FolderTileBody {
    flex: 1;
    min-height: 0;
    overflow: auto;
    @include vars.font('body-medium');
    button:where(:not(.v-btn, .tile-btn, .dg-btn, .dg-lbl)) {    // scoped 규칙은 자식 v-btn 루트에도 걸린다 — 여백·글자를 지우지 않게(:where = 우선순위 그대로)
        padding: 0;
        border: 0;
        background: none;
        color: inherit;
        font: inherit;
        cursor: pointer;
    }
}
// 경로 줄 = 28px 한 줄, 좁은 타일은 두 줄. 경로는 앞쪽부터 줄어든다
.crumbs {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0 2px;
    min-height: 28px;
    padding: 0 0 0 vars.$spacing-md;
    @include vars.font('body-small');
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    .lead {
        flex: 1 1 160px;
        min-width: 160px;
        height: 28px;
        display: flex;
        align-items: center;
    }
    .path {
        flex: 0 1 auto;
        min-width: 0;
        display: flex;
        align-items: center;
        gap: 2px;
        overflow: hidden;
        white-space: nowrap;
        > * { flex: 0 1000 auto; min-width: 2.2em; overflow: hidden; text-overflow: ellipsis; }   // 앞쪽 경로는 글자 하나 + …까지만 줄어든다
        > .sep { flex: none; min-width: 0; }
        > .cur { flex-shrink: 1; min-width: 0; }
        button { color: rgb(var(--v-theme-primary)); }
    }
    .cur { font-weight: vars.$weight-md; }
    &:not(.editing) { cursor: text; }
    .refresh { min-width: 38px; margin-left: 2px; }
    .tools {
        flex: none;
        height: 28px;
        cursor: default;
        margin-left: auto;
        display: flex;
        align-items: stretch;
        > .tile-btn:not(.tile-btn--text, .dev) { min-width: 42px; }
    }
    .dev { padding: 0 10px; opacity: 1; }
    .share { color: rgb(var(--v-theme-share)); .share-count { @include vars.font('body-small'); } }
    .paste-count { @include vars.font('body-small'); font-family: monospace; }
    .paste.tile-btn--tonal { --c: var(--v-theme-primary); }
}
.bar { height: 2px; }
.state {
    display: flex;
    align-items: center;
    gap: vars.$spacing-md;
    margin: vars.$spacing-md;
    color: rgb(var(--v-theme-error));
}
.DirGridCell {
    &.cut { opacity: 0.45; }    // CP2(표시 구분): 잘라낸 것만 흐리게
    &.drop :deep(.box) { outline: 1px dashed rgb(var(--v-theme-primary)); outline-offset: -1px; }
}
.FolderTileBody.drop { box-shadow: inset 0 0 0 1px rgb(var(--v-theme-primary)); }
// 지정된 장비 + 접속 여부(점 색)
.dev {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    max-width: 140px;
    white-space: nowrap;
    @include vars.font('label-small');
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
