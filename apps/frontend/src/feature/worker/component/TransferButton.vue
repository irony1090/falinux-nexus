<script setup lang="ts">
import { computed } from 'vue';
import { VBadge, VBtn, VIcon, VMenu } from 'vuetify/components';
import { useTransfers } from '../store/transfers.store';
import TransferRow from './TransferRow.vue';

// 푸터 전송 버튼 — 전송이 있을 때만(끝난 것은 5분 동안 남음), 배지 = 진행 중 개수. Q24(전송 상태 위치)·Q31(전송 아이콘 자리)
defineProps({ dense: Boolean })

const { list, running, failed } = useTransfers();
const icon = computed(() => running.value ? 'mdi-swap-vertical-bold' : 'mdi-swap-vertical');
</script>

<template>
<v-menu v-if="list.length" location="top end" offset="10" :close-on-content-click="false">
    <template #activator="{ props: act, isActive }">
        <v-btn v-bind="act" :size="dense ? undefined : 'small'" density="comfortable" variant="text" icon :active="isActive" title="전송" aria-label="전송 목록">
            <v-badge v-if="running" :content="running" color="primary" floating>
                <v-icon :icon="icon" />
            </v-badge>
            <v-badge v-else-if="failed" dot color="error" floating>
                <v-icon :icon="icon" />
            </v-badge>
            <v-icon v-else :icon="icon" />
        </v-btn>
    </template>
    <div class="TransferList">
        <div class="head">전송</div>
        <transfer-row v-for="t in list" :key="t.id" :item="t" show-dest />
    </div>
</v-menu>
</template>

<style scoped lang="scss">
.TransferList {
    width: min(360px, calc(100vw - 32px));
    max-height: 50vh;
    overflow: auto;
    padding: vars.$spacing-sm vars.$spacing-md;
    background: rgb(var(--v-theme-surface));
    border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    border-radius: 8px;
}
.head {
    padding-bottom: 4px;
    @include vars.font('body-small');
    font-weight: vars.$weight-md;
    opacity: 0.7;
}
</style>
