<script setup lang="ts">
import type { PropType } from 'vue';
import { VCard, VCardText, VCardTitle, VDialog, VList, VListItem } from 'vuetify/components';
import type { WorkerResponse } from '../api/worker.api';

// 실행 대상 인스턴스가 여러 개일 때 하나 고르기 (1개면 이 창 없이 바로 실행)
defineProps({
    title: {
        type: String,
        required: true,
    },
    workers: {
        type: Array as PropType<WorkerResponse[]>,
        default: () => [],
    },
    loading: {
        type: Boolean,
        default: () => false,
    },
})
const emit = defineEmits<{ submit: [instanceKey: string] }>();
const open = defineModel<boolean>({ type: Boolean, default: () => false });
</script>

<template>
<v-dialog v-model="open" max-width="396" min-width="296" :persistent="loading">
    <v-card :loading="loading" :disabled="loading">
        <v-card-title>{{ title }}</v-card-title>
        <v-card-text class="pa-0">
            <v-list density="compact">
                <v-list-item v-for="w in workers" :key="w.instanceKey"
                    prepend-icon="mdi-server"
                    :title="w.mainKey"
                    :subtitle="w.subKey"
                    @click="emit('submit', w.instanceKey)"
                />
            </v-list>
        </v-card-text>
    </v-card>
</v-dialog>
</template>
