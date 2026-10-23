<script setup lang="ts">
import { computed, type PropType } from 'vue';
import { VIcon, VProgressLinear } from 'vuetify/components';
import type { TransferInfo } from '../api/transfer.api';
import { useTransfers } from '../store/transfers.store';
import { transferColor, transferLabel, transferPercent } from '../util/transfer.util';
import { sizeText } from '../util/workerDir.util';

// 전송 한 줄 — 푸터 목록(받을 곳까지)·장비 타일 막대(Q24(전송 상태 위치)) 공용
const props = defineProps({
    item: {
        type: Object as PropType<TransferInfo>,
        required: true,
    },
    showDest: Boolean,
})

const { abort } = useTransfers();
const t = computed(() => props.item);
const running = computed(() => t.value.state === 'running');
const pct = computed(() => transferPercent(t.value));
</script>

<template>
<div class="TransferRow" :class="t.state">
    <div class="line">
        <v-icon size="13" icon="mdi-upload" class="kind" />
        <span class="nm" :title="`${t.dir} / ${t.name}`">{{ t.name }}</span>
        <span class="st" :title="t.reason">{{ transferLabel(t) }}{{ running ? ` ${pct}%` : '' }}</span>
        <span class="size">{{ sizeText(t.size) }}</span>
        <button v-if="running" class="x" title="중단" @click.stop="abort(t.id)"><v-icon size="13" icon="mdi-close" /></button>
    </div>
    <div v-if="showDest" class="dest" :title="t.dir">{{ t.instanceKey }} : {{ t.dir }}</div>
    <v-progress-linear v-if="running" :model-value="pct" height="2" :color="transferColor(t)" />
</div>
</template>

<style scoped lang="scss">
.TransferRow {
    padding: 3px 0;
    font-size: 11.5px;
    button {
        padding: 0;
        border: 0;
        background: none;
        color: inherit;
        cursor: pointer;
    }
}
.line {
    display: grid;
    grid-template-columns: 14px minmax(0, 1fr) auto auto 14px;
    align-items: center;
    gap: 6px;
}
.kind { opacity: 0.6; }
.nm {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.st { white-space: nowrap; opacity: 0.8; }
.size { font-family: monospace; font-size: 10.5px; opacity: 0.6; }
.x { display: inline-flex; opacity: 0.7; }
.dest {
    padding-left: 20px;
    font-size: 10.5px;
    opacity: 0.55;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.done .st { color: rgb(var(--v-theme-success)); }
.failed .st { color: rgb(var(--v-theme-error)); }
.aborted { opacity: 0.6; }
</style>
