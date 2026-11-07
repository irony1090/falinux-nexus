<script setup lang="ts">
import TileScroll from '@/feature/widget/component/TileScroll.vue';
import { computed, ref, watch } from 'vue';
import { useMemberBoard } from '@/feature/member/store/memberBoard.store';
import type { MemberTile } from '@/feature/member/util/memberGrid.util';
import MbNote from '@/feature/member/component/MbNote.vue';
import CtSelect from '@/feature/widget/component/CtSelect.vue';
import {
    cancelRun, startStepRun, useBundle, useTrial, useUpdDevice,
    type Ref, type StepDef, type StepList, type Updater, type UpdRun,
} from '../api/updater.api';
import { useUpdRun } from '../hook/updActions.hook';
import { useUpdLog, useUpdNow } from '../hook/updLog.hook';
import { shortSha, timeText } from '../util/updGrid.util';
import { RUN_ICON as ICON, RUN_TEXT as TEXT, runText as runTextAt, runTone as tone } from '../util/updRun.util';

// 시험 타일 — 배포 전에 steps.json 단계를 하나씩 장비에서 실행(REF-update-trial.md 결정-1~9).
// 기록 = 배포 밖 단독 실행(deployment_id NULL) -> 잠금 판정에 안 들어감. 실행 = Run · Log와 같은 startStepRun
const props = defineProps<{ tile: MemberTile; bundle: Ref }>();
const { dropMissing } = useMemberBoard();
const { busy, message, run } = useUpdRun();
const now = useUpdNow();
const runText = (r?: UpdRun, def?: StepDef) => runTextAt(now.value, r, def);

const bq = useBundle(props.bundle);
const b = computed(() => bq.data.value);
watch(() => (bq.error.value as { status?: number } | null)?.status, st => { if (st === 404) dropMissing(props.tile.key); }, { immediate: true });

// ---- 대상 인스턴스 ----
const dev = useUpdDevice(props.bundle.owner);
const instances = computed(() => dev.data.value?.instances ?? []);
const free = (i: Updater) => i.online && !i.running && !i.deploying;
const state = (i: Updater) => !i.online ? '오프라인' : i.deploying ? `배포 중 #${i.deploying}` : i.running ? `단계 실행 중 #${i.running}` : '대기';
const sub = ref('');
// 맞춤-1(자동 선택): 고를 수 있는(온라인·대기) 인스턴스가 하나뿐이면 그것
watch(instances, list => {
    if (sub.value) return;
    const ok = list.filter(free);
    if (ok.length === 1) sub.value = ok[0]!.subKey;
}, { immediate: true });
// 고른 것은 바빠져도(내 시험 실행 중) 목록에 그대로 보여야 한다
const items = computed(() => instances.value.map(i => ({
    title: `${i.label ? `${i.label} (${i.subKey})` : i.subKey} · ${state(i)}`, value: i.subKey, props: { disabled: !free(i) && i.subKey !== sub.value },
})));
const inst = computed(() => instances.value.find(i => i.subKey === sub.value));

// ---- 단계 ----
const tq = useTrial(props.bundle, sub);
const tr = computed(() => tq.data.value);
const list = ref<StepList>('deploy');
const defs = computed(() => (list.value === 'deploy' ? tr.value?.deploy : tr.value?.rollback) ?? []);
const runs = computed(() => tr.value?.runs ?? []);   // 최근 순
const lastOf = (id: string) => runs.value.find(r => r.list === list.value && r.stepId === id);
const rowSub = (def: StepDef) => def.kind === 'script' ? `script ${def.name} · ${runText(lastOf(def.id), def)}` : `builtin ${def.name} · 배포에서만`;
const mine = computed(() => runs.value.find(r => r.result === 'running'));
const canRun = computed(() => !!inst.value && free(inst.value) && !busy.value && !tr.value?.stepsIssue);

// ---- 로그 ----
const defOf = (r?: UpdRun) => r && tr.value ? (r.list === 'deploy' ? tr.value.deploy : tr.value.rollback).find(s => s.id === r.stepId) : undefined;
const { box, selected, shown, pick, follow, reset, size } = useUpdLog({
    runs: () => tr.value?.runs,
    interactive: r => !!defOf(r)?.interactive,
});
const shownDef = computed(() => defOf(shown.value));
watch(sub, () => reset());

// ---- 조작 ----
const runStep = (def: StepDef) => run(
    () => startStepRun(props.bundle.owner, sub.value, { version: props.bundle.version, list: list.value, stepId: def.id, ...size() }),
    r => follow(r.runId),
);
const cancel = () => mine.value && run(() => cancelRun(mine.value!.id));

// 맞춤-3(안내 문구): 막힌 이유가 있으면 그것, 없으면 기본 안내
const note = computed(() => {
    if (message.value) return message.value;
    if (tr.value?.stepsIssue) return tr.value.stepsIssue;
    if (dev.data.value && !instances.value.length) return '이 장비에 붙은 updater가 없습니다';
    const i = inst.value;
    if (!i) return '대상 인스턴스를 고르세요';
    if (!i.online) return 'updater가 접속해 있지 않습니다';
    if (i.deploying) return `배포 중입니다 #${i.deploying}`;
    if (i.running && i.running !== mine.value?.id) return `다른 단계가 실행 중입니다 #${i.running}`;
    return '배포가 아닙니다 · 잠금 판정에 안 들어갑니다';
});
</script>

<template>
<tile-scroll class="mb-body" data-pager-press>
    <p v-if="bq.isError.value" class="mb-empty">빌드를 찾을 수 없습니다</p>
    <template v-else>
        <div class="upd-fields">
            <span class="mb-lbl">대상</span>
            <CtSelect v-model="sub" class="edit" size="sm" wall="r" :items="items" hide-details placeholder="인스턴스 고르기" aria-label="대상 인스턴스"
                :disabled="busy || !instances.length" no-data-text="이 장비에 붙은 updater가 없습니다" />
            <span class="mb-lbl">빌드</span>
            <span class="val">{{ bundle.owner }} / {{ bundle.version }}<template v-if="b"> · sha <span class="upd-mono">{{ shortSha(b.sha) }}</span>{{ b.locked ? ' · 잠김(시험 가능)' : '' }}</template></span>
        </div>
        <div class="upd-set upd-lists">
            <span class="ct-grp ct-grp--sm ct-grp--wrap">
                <button type="button" class="ct-seg" :class="list === 'deploy' ? 'ct-seg--primary' : 'ct-seg--accent'" :aria-pressed="list === 'deploy'" @click="list = 'deploy'">deploy</button>
                <button type="button" class="ct-seg" :class="list === 'rollback' ? 'ct-seg--primary' : 'ct-seg--accent'" :aria-pressed="list === 'rollback'" @click="list = 'rollback'">rollback</button>
            </span>
        </div>
        <div v-for="def in defs" :key="def.id" class="upd-row" :class="{ click: !!lastOf(def.id), open: lastOf(def.id)?.id === selected, off: def.kind !== 'script' }"
            @click="pick(lastOf(def.id))">
            <span class="upd-ic" :class="tone(lastOf(def.id)?.result)">{{ ICON[lastOf(def.id)?.result ?? 'none'] }}</span>
            <span class="txt">
                <span class="main upd-mono">{{ def.id }}</span>
                <span class="sub" :title="rowSub(def)">{{ rowSub(def) }}</span>
            </span>
            <div v-if="def.kind === 'script'" class="ct-grp ct-grp--sm">
                <button type="button" class="ct-seg ct-seg--accent" :disabled="!canRun" title="이 단계만 실행(배포 아님)" @click.stop="runStep(def)">실행</button>
            </div>
        </div>
        <p v-if="tr && !tr.stepsIssue && !defs.length" class="upd-hint">{{ list === 'rollback' ? 'rollback 목록이 없습니다' : '단계가 없습니다' }}</p>

        <h3 class="mb-sec">로그 <span class="dim">{{ shown ? `· ${shown.list} ${shown.stepId} · ${TEXT[shown.result]}` : '' }}{{ shown?.result === 'running' && shownDef?.interactive ? ' · 터미널을 누르고 입력' : '' }}</span></h3>
        <div ref="box" class="upd-log" data-pager-press-off />

        <template v-if="runs.length">
            <h3 class="mb-sec">시험 기록 {{ runs.length }} <span class="dim">· {{ sub }}</span></h3>
            <div v-for="r in runs" :key="r.id" class="upd-row click" :class="{ open: r.id === selected }" @click="pick(r)">
                <span class="upd-ic" :class="tone(r.result)">{{ ICON[r.result] }}</span>
                <span class="txt">
                    <span class="main upd-mono">{{ r.list }} {{ r.stepId }}</span>
                    <span class="sub">{{ timeText(r.startedAt) }} · {{ runText(r, defOf(r)) }}</span>
                </span>
            </div>
        </template>
    </template>
</tile-scroll>
<div v-if="!bq.isError.value" class="mb-foot">
    <mb-note :text="note" />
    <div class="mb-end">
        <button v-if="mine" type="button" class="mb-btn mb-btn--danger" :disabled="busy" @click="cancel">취소</button>
    </div>
</div>
</template>

<style scoped lang="scss">
.dim { font-weight: vars.$weight-md; opacity: 0.75; }
.off .txt { opacity: 0.55; }
</style>
