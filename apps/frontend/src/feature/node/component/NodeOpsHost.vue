<script setup lang="ts">
import WorkerPickDialog from '@/feature/worker/component/WorkerPickDialog.vue';
import { useNodeOps } from '../store/nodeOps.store';
import NodeDeviceDialog from './dialog/NodeDeviceDialog.vue';
import NodeNameDialog from './dialog/NodeNameDialog.vue';
import ScriptEditDialog from './dialog/ScriptEditDialog.vue';

// 노드 동작 다이얼로그 — 보드에 한 번만(폴더 타일·선택 막대가 같이 씀). 상태·동작은 nodeOps.store.ts
const { saving, typeLabel, pick, start, name, nameTitle, nameInitial, submitName, editState, device, submitDevice } = useNodeOps();
</script>

<template>
<node-name-dialog v-model="name.open" :title="nameTitle" :initial="nameInitial" :loading="saving" @submit="submitName" />
<script-edit-dialog v-model="editState.open" :node="editState.node" />
<worker-pick-dialog v-model="pick.open"
    :title="`${typeLabel(pick.type)}할 장비: ${pick.node?.name ?? ''}`"
    :workers="pick.workers"
    :offline="pick.offline"
    :loading="saving"
    @submit="key => pick.node && start(pick.tileId, pick.node, key, pick.type)"
/>
<node-device-dialog v-model="device.open"
    :title="`장비 지정: ${device.node?.name ?? ''}`"
    :initial="device.node?.deviceKey ?? ''"
    :loading="saving"
    @submit="submitDevice"
/>
</template>
