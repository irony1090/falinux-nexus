<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useMemberBoard } from '@/feature/member/store/memberBoard.store';
import type { MemberTile } from '@/feature/member/util/memberGrid.util';
import MbNote from '@/feature/member/component/MbNote.vue';
import {
    cancelDeployment, cancelRun, getBundle, lockBuild, rollbackDeployment, startStepRun, useDeployment,
    type StepDef, type StepList, type UpdRun,
} from '../api/updater.api';
import { useUpdConfirm, useUpdRun } from '../hook/updActions.hook';
import { useUpdLog, useUpdNow } from '../hook/updLog.hook';
import { shortSha, timeText } from '../util/updGrid.util';
import { RUN_ICON as ICON, RUN_TEXT as TEXT, runText as runTextAt, runTone as tone } from '../util/updRun.util';

// Run · Log 타일 — 배포 하나 × 인스턴스 하나(REF-update-p3.md "합의한 해석"). 단계 행 `실행`(개별 실행) · 로그(입력 받는 단계는 입력) · 되돌리기 · 잠그기
// 시안 ③(배포 진행·잠그기). 단계 상태 = 이 배포의 기록, 개별 실행은 아래 따로
const props = defineProps<{ tile: MemberTile; deployId: number }>();
const { dropMissing } = useMemberBoard();
const confirm = useUpdConfirm();
const { busy, message, run } = useUpdRun();

const q = useDeployment(props.deployId);
const d = computed(() => q.data.value);
watch(() => (q.error.value as { status?: number } | null)?.status, st => { if (st === 404) dropMissing(props.tile.key); }, { immediate: true });
const ref_ = computed(() => d.value ? { kind: 'build' as const, owner: d.value.mainKey, version: d.value.version } : undefined);

// ---- 목록 ----
const list = ref<StepList>('deploy');
watch(() => d.value?.rollbackResult, (r, old) => { if (r === 'running' && old !== 'running') list.value = 'rollback'; });
type Row = { id: string; def?: StepDef; run?: UpdRun };
const rows = computed<Row[]>(() => {
    const dd = d.value;
    if (!dd) return [];
    const runs = dd.runs.filter(r => r.list === list.value);
    const defs = list.value === 'deploy' ? dd.deploy : dd.rollback;
    const last = (id: string) => runs.filter(r => r.stepId === id).at(-1);
    if (defs.length) return defs.map(def => ({ id: def.id, def, run: last(def.id) }));
    // 빌드가 지워졌거나 steps.json에 오류 — 기록만으로
    return [...new Set(runs.map(r => r.stepId))].map(id => ({ id, run: last(id) }));
});

const now = useUpdNow();
const runText = (r?: UpdRun, def?: StepDef) => runTextAt(now.value, r, def);
const rowSub = (row: Row) => [row.def ? `${row.def.kind} ${row.def.name}` : row.run?.kind, runText(row.run, row.def)].filter(Boolean).join(' · ');

// ---- 상태 ----
const RES: Record<string, string> = { running: '진행 중', ok: '성공', fail: '실패', cancel: '취소' };
const running = computed(() => d.value?.result === 'running' || d.value?.rollbackResult === 'running');
const manualRunning = computed(() => d.value?.manual.find(r => r.result === 'running'));
const canRun = computed(() => !!d.value?.build && !!d.value.online && !d.value.busy && !busy.value);
const target = computed(() => d.value ? `${d.value.mainKey}#${d.value.subKey}` : '');

// ---- 로그 ----
const { box, selected, shown, pick, follow, unpin, size } = useUpdLog({
    runs: () => d.value && [...d.value.runs, ...d.value.manual],
    interactive: r => !!defOf(r)?.interactive,
});
const defOf = (r?: UpdRun) => r && d.value ? (r.list === 'deploy' ? d.value.deploy : d.value.rollback).find(s => s.id === r.stepId) : undefined;
const shownDef = computed(() => defOf(shown.value));

// ---- 조작 ----
const runStep = (row: Row) => {
    const dd = d.value!;
    return run(() => startStepRun(dd.mainKey, dd.subKey, { version: dd.version, list: list.value, stepId: row.id, deployId: `d${dd.id}`, ...size() }), r => follow(r.runId));
};
const rollback = async () => {
    const dd = d.value!;
    const names = dd.rollback.filter(s => !s.manual).map(s => s.id).join(' -> ');
    if (!(await confirm(`되돌리기 · ${target.value}`, [
        `이 배포(#${dd.id} ${dd.version})를 취소하고 배포 직전 상태로 돌아갑니다. ${dd.version}으로 가는 것이 아닙니다.`,
        `rollback 목록을 순서대로 실행합니다: ${names}`,
        '특정 버전으로 가려면 되돌리기 대신 그 빌드를 다시 배포하세요.',
    ], '되돌리기'))) return;
    await run(() => rollbackDeployment(dd.id, size()), () => { list.value = 'rollback'; unpin(); });
};
const cancel = () => {
    const dd = d.value!;
    if (running.value) return run(() => cancelDeployment(dd.id));
    if (manualRunning.value) return run(() => cancelRun(manualRunning.value!.id));
};
const lock = async () => {
    const dd = d.value!;
    const uses = (await getBundle(ref_.value!).catch(() => undefined))?.uses ?? [];
    if (!(await confirm(`잠그기 · ${dd.mainKey}/${dd.version}`, [
        `지금 내용(sha ${shortSha(dd.buildSha)})과 같은 내용으로 성공한 배포가 있어 잠글 수 있습니다.`,
        uses.length ? `같이 잠기는 공용 묶음: ${uses.map(u => `${u.name}@${u.version}`).join(', ')}` : '같이 잠기는 공용 묶음은 없습니다.',
        '잠그면 이 빌드와 공용 묶음의 파일·이름을 고칠 수 없습니다. 배포는 계속할 수 있습니다.',
    ], '잠그기'))) return;
    await run(() => lockBuild(ref_.value!));
};

const note = computed(() => {
    const dd = d.value;
    if (!dd) return '';
    if (message.value) return message.value;
    if (!dd.build) return '빌드가 지워져 단계 실행·되돌리기를 할 수 없습니다(기록은 남음)';
    if (!dd.online) return 'updater가 접속해 있지 않습니다';
    if (running.value) return dd.result === 'running' ? '배포 진행 중 · 개별 실행은 잠금 판정에 들어가지 않습니다' : '되돌리기 진행 중';
    if (dd.rollbackBlock && dd.result !== 'running') return dd.rollbackBlock;
    if (dd.buildLocked) return '잠긴 빌드입니다. 배포는 계속할 수 있습니다';
    if (!dd.lockable) return dd.sha !== dd.buildSha ? '배포 뒤 빌드를 고쳐 지금 내용으로 성공한 기록이 없습니다' : '끝까지 성공한 배포가 있어야 잠글 수 있습니다';
    return '같은 내용으로 성공한 배포가 있어 잠글 수 있습니다';
});
</script>

<template>
<div class="mb-body" data-pager-press>
    <p v-if="q.isError.value" class="mb-empty">배포 기록을 찾을 수 없습니다</p>
    <template v-else-if="d">
        <div class="upd-fields">
            <span class="mb-lbl">배포 id</span>
            <span class="val mono">#{{ d.id }}</span>
            <span class="mb-lbl">대상</span>
            <span class="val">{{ target }}</span>
            <span class="mb-lbl">sha</span>
            <span class="val mono">{{ shortSha(d.sha) }}<template v-if="d.build && d.buildSha !== d.sha"> · 지금 {{ shortSha(d.buildSha) }}</template></span>
            <span class="mb-lbl">시작</span>
            <span class="val">{{ timeText(d.startedAt) }} · {{ d.startedByName || '—' }}</span>
            <span class="mb-lbl">결과</span>
            <span class="val">
                <span class="mb-tag" :class="`t-${d.result}`">배포 {{ RES[d.result] }}</span>
                <span v-if="d.rollbackResult" class="mb-tag" :class="`t-${d.rollbackResult}`">되돌리기 {{ RES[d.rollbackResult] }}</span>
            </span>
            <p v-if="d.msg" class="full msg">{{ d.msg }}</p>
        </div>
        <div class="upd-set upd-lists">
            <span class="ct-grp ct-grp--sm ct-grp--wrap">
                <button type="button" class="ct-seg" :class="list === 'deploy' ? 'ct-seg--primary' : 'ct-seg--accent'" :aria-pressed="list === 'deploy'" @click="list = 'deploy'">deploy</button>
                <button type="button" class="ct-seg" :class="list === 'rollback' ? 'ct-seg--primary' : 'ct-seg--accent'" :aria-pressed="list === 'rollback'" @click="list = 'rollback'">rollback{{ d.autoRollback ? ' · 자동' : '' }}</button>
            </span>
        </div>
        <div v-for="row in rows" :key="row.id" class="upd-row click" :class="{ open: row.run && row.run.id === selected }" @click="pick(row.run)">
            <span class="upd-ic" :class="tone(row.run?.result)">{{ ICON[row.run?.result ?? 'none'] }}</span>
            <span class="txt">
                <span class="main upd-mono">{{ row.id }}</span>
                <span class="sub" :title="rowSub(row)">{{ rowSub(row) }}</span>
            </span>
            <div class="ct-grp ct-grp--sm">
                <button type="button" class="ct-seg ct-seg--accent" :disabled="!canRun || !row.def" :title="row.def ? '이 단계만 실행(잠금 판정에 안 들어감)' : undefined" @click.stop="runStep(row)">실행</button>
            </div>
        </div>
        <p v-if="!rows.length" class="upd-hint">{{ list === 'rollback' ? 'rollback 목록이 없습니다' : '단계가 없습니다' }}</p>
        <p v-if="list === 'rollback' && rows.length" class="upd-hint">아래 줄의 되돌리기가 이 목록을 순서대로 실행합니다(manual 단계 빼고)</p>

        <h3 class="mb-sec">로그 <span class="dim">{{ shown ? `· ${shown.list} ${shown.stepId} · ${TEXT[shown.result]}` : '' }}{{ shown?.result === 'running' && shownDef?.interactive ? ' · 터미널을 누르고 입력' : '' }}</span></h3>
        <div ref="box" class="upd-log" data-pager-press-off />

        <template v-if="d.manual.length">
            <h3 class="mb-sec">시험·단독 실행 {{ d.manual.length }} <span class="dim">· 잠금 판정에 들어가지 않습니다</span></h3>
            <div v-for="r in d.manual" :key="r.id" class="upd-row click" :class="{ open: r.id === selected }" @click="pick(r)">
                <span class="upd-ic" :class="tone(r.result)">{{ ICON[r.result] }}</span>
                <span class="txt">
                    <span class="main upd-mono">{{ r.list }} {{ r.stepId }}</span>
                    <span class="sub">{{ timeText(r.startedAt) }} · {{ runText(r) }}</span>
                </span>
            </div>
        </template>
    </template>
</div>
<div v-if="d" class="mb-foot">
    <mb-note :text="note" />
    <div class="mb-end">
        <button v-if="running || manualRunning" type="button" class="mb-btn mb-btn--danger" :disabled="busy" @click="cancel">취소</button>
        <button type="button" class="mb-btn" :disabled="!canRun || running || !!d.rollbackBlock || !d.rollback.some(s => !s.manual)" :title="d.rollbackBlock || '이 배포를 취소하고 배포 직전 상태로'" @click="rollback">되돌리기</button>
        <button type="button" class="mb-btn mb-btn--primary" :disabled="busy || !d.lockable || d.buildLocked" @click="lock">{{ d.buildLocked ? '잠김' : '잠그기' }}</button>
    </div>
</div>
</template>

<style scoped lang="scss">
.dim { font-weight: vars.$weight-md; opacity: 0.75; }
.val .mb-tag { margin-left: 6px; vertical-align: middle; }
.t-ok { --c: var(--v-theme-success); }
.t-fail { --c: var(--v-theme-error); }
.t-running { --c: var(--v-theme-primary); }
.msg { margin: 0 16px 6px 0; @include vars.font('body-small'); color: rgb(var(--v-theme-error)); overflow-wrap: anywhere; }
</style>
