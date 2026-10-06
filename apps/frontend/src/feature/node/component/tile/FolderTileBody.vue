<script setup lang="ts">
import { tabReady } from '@/common/util/tabId.util';
import { useAppDialog } from '@/feature/layout/store/appDialog.store';
import type { ProcessType } from '@/feature/process/api/process.api';
import { useProcessTerm } from '@/feature/process/store/processTerm.store';
import { listWorkers, useOnlineWorkers, type WorkerResponse } from '@/feature/worker/api/worker.api';
import WorkerPickDialog from '@/feature/worker/component/WorkerPickDialog.vue';
import { computed, ref } from 'vue';
import { VBtn, VIcon, VList, VListItem, VMenu, VProgressLinear } from 'vuetify/components';
import { useListChildren, useNodeActions, useNodePath, type NodeKind, type NodeResponse } from '../../api/node.api';
import { useNodeRemove } from '../../hook/nodeRemove.hook';
import { useTileGrids } from '../../store/tileGrids.store';
import { useTileTree } from '../../store/tileTree.store';
import NodeDeviceDialog from '../dialog/NodeDeviceDialog.vue';
import NodeNameDialog from '../dialog/NodeNameDialog.vue';
import ScriptEditDialog from '../dialog/ScriptEditDialog.vue';
import NodeRowMenu from './NodeRowMenu.vue';

const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const { tiles, navigate, openFolder, applyServerTile } = useTileTree();
const { newSize, reveal } = useTileGrids();
const { open: appDialogOpen, openDialog } = useAppDialog();

const nodeId = computed(() => {
    const t = tiles.value[props.tileId];
    return t?.type === 'folder' ? t.nodeId : null;
})

const listQuery = useListChildren(computed(() => nodeId.value ?? undefined));
const pathQuery = useNodePath(nodeId);
const { counts: online, refetch: refetchWorkers } = useOnlineWorkers();

const path = computed(() => pathQuery.data.value ?? []);
const children = computed(() => listQuery.data.value ?? []);
const busy = computed(() => listQuery.isFetching.value || pathQuery.isFetching.value);
const error = computed(() => listQuery.error.value ?? pathQuery.error.value);

// 지금 보는 폴더와 실제로 적용되는 장비(자기 것, 없으면 가장 가까운 상위 폴더 것)
const current = computed(() => path.value[path.value.length - 1]);
const deviceOwner = computed(() => [...path.value].reverse().find(p => !!p.deviceKey));
const deviceInherited = computed(() => !!deviceOwner.value && deviceOwner.value.id !== current.value?.id);

const deviceTitle = (key: string | null | undefined, from?: string) => {
    if (!key) return '지정된 장비 없음';
    const n = online.value.get(key);
    return [key, from && `상위 폴더 '${from}'에서 상속`, n ? `접속 중 ${n}개` : '접속 안 됨'].filter(v => !!v).join(' — ');
}

const onRefresh = () => {
    listQuery.refetch();
    refetchWorkers();
    if (nodeId.value !== null) pathQuery.refetch();
}

// 새 타일 크기 = 누른 화면의 최소 단위
const onOpen = (id: number) => reveal(openFolder(props.tileId, id, newSize.value));

const saving = ref(false);
const run = <T>(job: Promise<T>, done?: (res: T) => void) => {
    saving.value = true;
    job.then(res => done?.(res))
    .catch(err => openDialog({ type: 'error', content: err?.message || '요청에 실패했습니다' }))
    .finally(() => saving.value = false);
}

// ---- 실행·vi 편집: 접속 인스턴스 0개 = 안내 / 1개 = 바로 / 여러 개 = 고르기 (REF-node-ui-terminal.md L, E1(진입 위치)) ----
const { exec } = useProcessTerm();
const typeLabel = (type: ProcessType) => type === 'EDIT' ? 'vi로 편집' : '실행';

const pickOpen = ref(false);
const pickNode = ref<NodeResponse>();
const pickType = ref<ProcessType>('EXEC');
const pickWorkers = ref<WorkerResponse[]>([]);

// 터미널 타일은 서버가 exec 때 트리에 넣는다 — ⑪-4(타일 없는 실행 중 process)
const start = (node: NodeResponse, instanceKey: string, type: ProcessType) => run(exec(node.id, instanceKey, { parentTileId: props.tileId, size: newSize.value }, type), res => {
    pickOpen.value = false;
    if (res.tile && res.tileVersion) reveal(applyServerTile(res.tile, res.tileVersion));
});

const onExec = (node: NodeResponse, type: ProcessType = 'EXEC') => run(listWorkers(node.id), workers => {
    if (!workers.length) {
        openDialog({ type: 'warning', title: `'${node.name}' ${typeLabel(type)}`, content: `이 스크립트를 ${typeLabel(type)}할 장비가 접속해 있지 않습니다.` });
        return;
    }
    if (workers.length === 1) {
        start(node, workers[0]!.instanceKey, type);
        return;
    }
    pickNode.value = node;
    pickType.value = type;
    pickWorkers.value = workers;
    pickOpen.value = true;
});

// ---- 노드 관리 (생성 / 이름 변경 / 장비 지정 / 삭제) ----
const { create, patch } = useNodeActions();
const { removeNode } = useNodeRemove();

// 닫히는 동안 제목이 비지 않도록 열림 여부와 대상을 따로 둔다
type NameTarget = { kind: NodeKind } | { node: NodeResponse };
const nameOpen = ref(false);
const nameTarget = ref<NameTarget>({ kind: 'FOLDER' });
const nameTitle = computed(() => {
    const t = nameTarget.value;
    if ('node' in t) return '이름 변경';
    return t.kind === 'FOLDER' ? '새 폴더' : '새 스크립트';
})
const nameInitial = computed(() => 'node' in nameTarget.value ? nameTarget.value.node.name : '');

const onName = (target: NameTarget) => {
    nameTarget.value = target;
    nameOpen.value = true;
}
const onSubmitName = (name: string) => {
    const t = nameTarget.value;
    if ('node' in t) {
        run(patch(t.node.id, { name }), () => nameOpen.value = false);
        return;
    }
    run(create({ kind: t.kind, name, parentId: nodeId.value }), node => {
        nameOpen.value = false;
        if (node.kind === 'SCRIPT') onEdit(node);   // 빈 스크립트는 바로 내용 편집으로 — 결정 I(새 스크립트 직후 동작)
    });
}

const editOpen = ref(false);
const editTarget = ref<NodeResponse>();
const onEdit = (node: NodeResponse) => {
    editTarget.value = node;
    editOpen.value = true;
}

const deviceOpen = ref(false);
const deviceTarget = ref<NodeResponse>();
const onDevice = (node: NodeResponse) => {
    deviceTarget.value = node;
    deviceOpen.value = true;
}
const onSubmitDevice = (deviceKey: string | null) => {
    if (!deviceTarget.value) return;
    run(patch(deviceTarget.value.id, { deviceKey }), () => deviceOpen.value = false);
}

// 이름은 title로만 넘긴다 — AppDialog의 content는 v-html로 그려진다
const onRemove = (node: NodeResponse) => {
    openDialog({
        type: 'warning',
        title: `'${node.name}' 삭제`,
        content: node.kind === 'FOLDER'
            ? ['폴더와 그 안의 모든 항목이 삭제됩니다.', '이 폴더나 하위 폴더를 보고 있는 타일은 닫힙니다.']
            : '스크립트가 삭제됩니다.',
        buttons: [
            { view: '취소', type: 'info', bindButton: 'Escape', action: () => appDialogOpen.value = false },
            {
                // Enter를 묶지 않는다 — 묶지 않은 Enter·Space는 AppDialog가 닫기(취소)로 처리
                view: '삭제', type: 'error',
                action: () => {
                    appDialogOpen.value = false;
                    run(removeNode(node));
                },
            },
        ],
    })
}
</script>

<template>
<div class="FolderTileBody">
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
                :class="{ on: !!deviceOwner?.deviceKey && online.has(deviceOwner.deviceKey) }"
                :title="`${deviceTitle(deviceOwner?.deviceKey, deviceInherited ? deviceOwner?.name : undefined)} (눌러서 이 폴더의 장비 지정)`"
                :disabled="saving"
                @click="onDevice(current)"
            >
                <i /><span class="key">{{ deviceOwner?.deviceKey ?? '장비 없음' }}</span><span v-if="deviceInherited">(상속)</span>
            </button>
            <v-menu location="bottom end">
                <template #activator="{ props: act }">
                    <button v-bind="act" class="tool" title="새로 만들기" :disabled="saving">
                        <v-icon size="14" icon="mdi-plus" />
                    </button>
                </template>
                <v-list density="compact">
                    <v-list-item prepend-icon="mdi-folder-plus-outline" title="새 폴더" @click="onName({ kind: 'FOLDER' })" />
                    <v-list-item prepend-icon="mdi-file-plus-outline" title="새 스크립트" @click="onName({ kind: 'SCRIPT' })" />
                </v-list>
            </v-menu>
            <button class="tool" title="새로고침" :disabled="busy" @click="onRefresh">
                <v-icon size="14" icon="mdi-refresh" />
            </button>
        </span>
    </div>
    <div class="bar">
        <v-progress-linear v-if="busy || saving" indeterminate height="2" color="primary" />
    </div>
    <div v-if="error" class="state">
        <span>{{ error.message || '목록을 불러오지 못했습니다' }}</span>
        <v-btn size="x-small" variant="tonal" :disabled="busy" @click="onRefresh">다시 시도</v-btn>
    </div>
    <template v-else>
        <div v-for="(n, i) in children" :key="n.id" class="row">
            <span class="ord">{{ i + 1 }}</span>
            <v-icon size="14"
                :icon="n.kind === 'FOLDER' ? 'mdi-folder' : 'mdi-console'"
                :class="n.kind === 'FOLDER' ? 'folder' : 'script'"
            />
            <button v-if="n.kind === 'FOLDER'" class="nm" title="이 타일 안에서 열기" @click="navigate(tileId, n.id)">{{ n.name }}</button>
            <span v-else class="nm">{{ n.name }}</span>
            <span class="tail">
                <span v-if="n.deviceKey" class="dev" :class="{ on: online.has(n.deviceKey) }" :title="deviceTitle(n.deviceKey)">
                    <i /><span class="key">{{ n.deviceKey }}</span>
                </span>
                <v-btn v-if="n.kind === 'FOLDER'" size="x-small" variant="tonal" color="warning" @click="onOpen(n.id)">새 타일</v-btn>
                <v-btn v-else size="x-small" variant="tonal" color="primary" :disabled="saving || !tabReady" @click="onExec(n)">실행</v-btn>
                <node-row-menu :kind="n.kind" :vi-disabled="saving || !tabReady" @edit="onEdit(n)" @vi="onExec(n, 'EDIT')" @rename="onName({ node: n })" @device="onDevice(n)" @remove="onRemove(n)" />
            </span>
        </div>
        <p v-if="!listQuery.isPending.value && !children.length" class="empty">비어 있음</p>
    </template>

    <node-name-dialog v-model="nameOpen" :title="nameTitle" :initial="nameInitial" :loading="saving" @submit="onSubmitName" />
    <script-edit-dialog v-model="editOpen" :node="editTarget" />
    <worker-pick-dialog v-model="pickOpen"
        :title="`${typeLabel(pickType)}할 장비: ${pickNode?.name ?? ''}`"
        :workers="pickWorkers"
        :loading="saving"
        @submit="key => pickNode && start(pickNode, key, pickType)"
    />
    <node-device-dialog v-model="deviceOpen"
        :title="`장비 지정: ${deviceTarget?.name ?? ''}`"
        :initial="deviceTarget?.deviceKey ?? ''"
        :loading="saving"
        @submit="onSubmitDevice"
    />
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
}
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
    &.on i { background: rgb(var(--v-theme-success)); }
}
.empty { margin: vars.$spacing-md; opacity: 0.6; }
</style>
