<script setup lang="ts">
import type { PropType } from 'vue';
import { VBtn, VList, VListItem, VMenu } from 'vuetify/components';
import type { NodeAct, NodeActItem } from '../../store/nodeOps.store';

// 폴더 타일 목록 한 행의 조작 메뉴 — 동작은 부모가 처리. at이 있으면 ⋮ 버튼 없이 그 좌표에 뜬다(우클릭, SM1(켜는 방법))
defineProps({
    items: {
        type: Array as PropType<NodeActItem[]>,
        required: true,
    },
    at: Array as unknown as PropType<[number, number]>,
})
const open = defineModel<boolean>({ default: false });
const emit = defineEmits<{ act: [act: NodeAct] }>();
</script>

<template>
<v-menu v-model="open" :location="at ? 'bottom start' : 'bottom end'" :target="at">
    <template v-if="!at" #activator="{ props: act }">
        <v-btn v-bind="act" size="x-small" variant="text" icon="mdi-dots-vertical" title="메뉴" />
    </template>
    <v-list density="compact">
        <v-list-item v-for="it in items" :key="it.act"
            :prepend-icon="it.icon"
            :title="it.title"
            :disabled="it.disabled"
            :base-color="it.danger ? 'error' : undefined"
            @click="emit('act', it.act)"
        />
    </v-list>
</v-menu>
</template>
