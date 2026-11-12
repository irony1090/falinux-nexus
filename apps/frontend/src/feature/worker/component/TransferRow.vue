<script setup lang="ts">
import { computed, type PropType } from 'vue';
import { VIcon, VProgressLinear } from 'vuetify/components';
import type { TransferInfo } from '../api/transfer.api';
import { useTransfers } from '../store/transfers.store';
import { KIND_ICON, transferColor, transferLabel, transferPercent } from '../util/transfer.util';
import { sizeText } from '../util/workerDir.util';

// 전송 한 줄 — 푸터 목록(받을 곳까지)·장비 타일 막대(Q24(전송 상태 위치)) 공용
const props = defineProps({
    item: {
        type: Object as PropType<TransferInfo>,
        required: true,
    },
    showDest: Boolean,
})

// 상태 칸 색 — 완료는 글자색만(바탕 기본), 중단됨은 줄 전체를 흐리게(.aborted)
const ST_CLS: Partial<Record<TransferInfo['state'], string>> = { running: 'ct-seg--accent', done: 'ct-seg--ok', failed: 'ct-seg--error' };

const { abort } = useTransfers();
const t = computed(() => props.item);
const running = computed(() => t.value.state === 'running');
const pct = computed(() => transferPercent(t.value));
// 크기를 모르면(zip) 옮긴 양만
const sizeLabel = computed(() => t.value.size < 0 ? sizeText(t.value.done) : sizeText(t.value.size));
const stCls = computed(() => ST_CLS[t.value.state] ?? '');
const dest = computed(() => t.value.fromKey ? `${t.value.fromKey} → ${t.value.instanceKey} : ${t.value.dir}` : `${t.value.instanceKey} : ${t.value.dir}`);
</script>

<template>
<div class="TransferRow" :class="t.state">
    <div class="line">
        <v-icon size="13" :icon="KIND_ICON[t.kind]" class="kind" />
        <span class="nm" :title="t.fromPath ?? `${t.dir} / ${t.name}`">{{ t.name }}</span>
        <span class="size">{{ sizeLabel }}</span>
        <span class="ct-grp ct-grp--r ct-grp--sm">
            <span class="ct-seg" :class="stCls" :title="t.reason">{{ transferLabel(t) }}{{ running && pct !== undefined ? ` ${pct}%` : '' }}</span>
            <button v-if="running" class="ct-seg ct-seg--accent ct-seg--icon" title="중단" @click.stop="abort(t.id)"><v-icon size="13" icon="mdi-close" /></button>
        </span>
    </div>
    <div v-if="showDest" class="dest" :title="dest">{{ dest }}</div>
    <v-progress-linear v-if="running" :model-value="pct" :indeterminate="pct === undefined" height="2" :color="transferColor(t)" />
</div>
</template>

<style scoped lang="scss">
.TransferRow {
    padding: 3px 0;
    @include vars.font('body-small');
}
.line {
    display: grid;
    grid-template-columns: 14px minmax(0, 1fr) auto auto;   // 상태 묶음 = 부모 오른쪽 벽에 붙음(부모 오른쪽 패딩 0)
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
.size { font-family: monospace; @include vars.font('label-small'); opacity: 0.6; }
.dest {
    padding: 0 vars.$spacing-md 0 20px;
    @include vars.font('label-small');
    opacity: 0.55;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.aborted { opacity: 0.6; }
</style>
