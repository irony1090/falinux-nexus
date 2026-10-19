<script setup lang="ts">
import { agoText } from '@/feature/node/util/presenceView.util';
import type { PropType } from 'vue';
import { VCard, VCardSubtitle, VCardText, VCardTitle, VDialog, VDivider, VList, VListItem, VListSubheader } from 'vuetify/components';
import type { WorkerInstanceResponse, WorkerResponse } from '../api/worker.api';
import { osBadge, osTitle } from '../util/workerOs.util';

// 실행 대상 인스턴스가 여러 개일 때 하나 고르기 (1개면 이 창 없이 바로 실행). 끊긴 인스턴스는 흐리게만 — RS5(표시 위치) ①
defineProps({
    title: {
        type: String,
        required: true,
    },
    workers: {
        type: Array as PropType<WorkerResponse[]>,
        default: () => [],
    },
    offline: {
        type: Array as PropType<WorkerInstanceResponse[]>,
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
        <v-card-subtitle v-if="workers.length">
            장비 {{ workers[0]!.mainKey }} · 접속 {{ workers.length }} / 명부 {{ workers.length + offline.length }}
        </v-card-subtitle>
        <v-card-text class="pa-0">
            <v-list density="compact">
                <v-list-item v-for="w in workers" :key="w.instanceKey"
                    prepend-icon="mdi-server"
                    :title="w.mainKey"
                    @click="emit('submit', w.instanceKey)"
                >
                    <template #subtitle><span class="os" :title="osTitle(w.os)">{{ osBadge(w.os) }}</span> {{ w.subKey }}</template>
                    <template #append><small class="text-success">접속 중</small></template>
                </v-list-item>
                <template v-if="offline.length">
                    <v-divider class="my-1" />
                    <v-list-subheader>끊긴 인스턴스</v-list-subheader>
                    <v-list-item v-for="w in offline" :key="w.instanceKey"
                        prepend-icon="mdi-server-off"
                        :title="w.mainKey"
                        disabled
                    >
                        <template #subtitle><span class="os" :title="osTitle(w.os)">{{ osBadge(w.os) }}</span> {{ w.subKey }}</template>
                        <template #append><small>마지막 접속 {{ agoText(w.lastSeen) }}</small></template>
                    </v-list-item>
                </template>
            </v-list>
        </v-card-text>
    </v-card>
</v-dialog>
</template>

<style scoped lang="scss">
.os { font-family: monospace; font-weight: 600; }
</style>
