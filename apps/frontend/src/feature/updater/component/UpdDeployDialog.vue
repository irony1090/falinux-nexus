<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { VCard, VCardTitle, VDialog } from 'vuetify/components';
import CtCheckbox from '@/feature/widget/component/CtCheckbox.vue';
import { useMemberBoard } from '@/feature/member/store/memberBoard.store';
import { deployBuild, refName, useUpdDevice, type Ref, type Updater } from '../api/updater.api';
import { useUpdRun } from '../hook/updActions.hook';
import { buildTileKey, runTileKey } from '../util/updGrid.util';

// 배포 대상 고르기 — 가운데 팝업(REF-update-p3.md "합의한 해석"). 같은 메인키 인스턴스, 기본 = 붙어 있고 한가한 것 전부.
// 배포가 시작된 인스턴스마다 Run · Log 타일을 빌드 타일 뒤에 연다
const props = defineProps<{ bundle: Ref }>();
const open = defineModel<boolean>({ required: true });
const { open: openTile } = useMemberBoard();
const { busy, message, run } = useUpdRun();

const q = useUpdDevice(props.bundle.owner);
const instances = computed(() => q.data.value?.instances ?? []);
const free = (i: Updater) => i.online && !i.running && !i.deploying;
const picked = ref<Record<string, boolean>>({});
const errors = ref<Record<string, string>>({});
watch(open, v => {
    if (!v) return;
    errors.value = {};
    message.value = undefined;
    picked.value = Object.fromEntries(instances.value.map(i => [i.subKey, free(i)]));
}, { immediate: true });
const chosen = computed(() => instances.value.filter(i => free(i) && picked.value[i.subKey]).map(i => i.subKey));

const state = (i: Updater) => !i.online ? '오프라인' : i.deploying ? `배포 중 #${i.deploying}` : i.running ? `단계 실행 중 #${i.running}` : '대기';
const sub = (i: Updater) => [i.subKey, i.os && `${i.os}/${i.arch}`, `설치본 ${i.installedVersion || '없음'}`, i.installDir || '경로 없음'].filter(Boolean).join(' · ');

const go = () => run(() => deployBuild(props.bundle, chosen.value), res => {
    const parent = buildTileKey(props.bundle.owner, props.bundle.version);
    for (const it of res.items) if (it.deployId && !it.error) openTile(runTileKey(it.deployId), parent);
    errors.value = Object.fromEntries(res.items.filter(it => it.error).map(it => [it.subKey, it.error!]));
    if (!Object.keys(errors.value).length) open.value = false;
});
</script>

<template>
<v-dialog v-model="open" max-width="480">
    <v-card class="UpdDeployDialog ct-card">
        <v-card-title>배포 · {{ refName(bundle) }}</v-card-title>
        <div class="list">
            <label v-for="i in instances" :key="i.subKey" class="upd-row" :class="{ off: !free(i) }">
                <CtCheckbox v-model="picked[i.subKey]" :disabled="!free(i) || busy" :aria-label="`${i.label || i.subKey} 고르기`" />
                <span class="txt">
                    <span class="main">{{ i.label || i.subKey }}</span>
                    <span class="sub" :title="sub(i)">{{ sub(i) }}</span>
                    <span v-if="errors[i.subKey]" class="mb-err">{{ errors[i.subKey] }}</span>
                </span>
                <span class="mb-tag" :class="free(i) ? 'mb-tag--success' : 'dim'">{{ state(i) }}</span>
            </label>
            <p v-if="!instances.length" class="upd-hint">이 장비에 붙은 updater가 없습니다</p>
        </div>
        <p class="note">같은 메인키 인스턴스는 같은 파일과 steps.json을 받습니다. 성공 판정은 한 대라도 끝까지 성공하면 됩니다.</p>
        <p v-if="message" class="mb-err note">{{ message }}</p>
        <div class="ct-actions">
            <span class="ct-grp ct-grp--r">
                <button type="button" class="ct-seg" @click="open = false">{{ Object.keys(errors).length ? '닫기' : '취소' }}</button>
                <button type="button" class="ct-seg ct-seg--primary" :disabled="busy || !chosen.length" @click="go">배포 {{ chosen.length }}대</button>
            </span>
        </div>
    </v-card>
</v-dialog>
</template>

<style scoped lang="scss">
.list {
    max-height: 50vh;
    overflow-y: auto;
    margin-top: 4px;
    border-top: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    > .upd-row { cursor: pointer; padding-left: 8px; }
    > .upd-row.off { cursor: default; opacity: 0.6; }
    .mb-err { @include vars.font('body-small'); }
}
// 상태 라벨 = 행 오른쪽 벽(닿은 끝 곧게) — 찰떡 "리스트 행의 상태 라벨"
.mb-tag { border-radius: 999px 0 0 999px; &.dim { opacity: 0.75; } }
.note {
    margin: 0;
    padding: 10px vars.$spacing-xl 0;
    @include vars.font('body-small');
    color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity));
}
</style>
