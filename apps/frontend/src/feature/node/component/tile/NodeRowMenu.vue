<script setup lang="ts">
import type { PropType } from 'vue';
import { VBtn, VList, VListItem, VMenu } from 'vuetify/components';
import type { NodeKind } from '../../api/node.api';

// 폴더 타일 목록 한 행의 조작 메뉴 — 동작은 부모가 처리
defineProps({
    kind: {
        type: String as PropType<NodeKind>,
        required: true,
    },
})
const emit = defineEmits<{ edit: []; rename: []; device: []; remove: [] }>();
</script>

<template>
<v-menu location="bottom end">
    <template #activator="{ props: act }">
        <v-btn v-bind="act" size="x-small" variant="text" icon="mdi-dots-vertical" title="메뉴" />
    </template>
    <v-list density="compact">
        <v-list-item v-if="kind === 'SCRIPT'" prepend-icon="mdi-file-document-edit-outline" title="내용 편집" @click="emit('edit')" />
        <v-list-item prepend-icon="mdi-pencil-outline" title="이름 변경" @click="emit('rename')" />
        <v-list-item v-if="kind === 'FOLDER'" prepend-icon="mdi-lan-connect" title="장비 지정" @click="emit('device')" />
        <v-list-item prepend-icon="mdi-trash-can-outline" title="삭제" base-color="error" @click="emit('remove')" />
    </v-list>
</v-menu>
</template>
