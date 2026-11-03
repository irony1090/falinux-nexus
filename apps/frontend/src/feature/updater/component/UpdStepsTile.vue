<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { useMemberBoard } from '@/feature/member/store/memberBoard.store';
import type { MemberTile } from '@/feature/member/util/memberGrid.util';
import CtCode, { type CtCodeIssue } from '@/feature/widget/component/CtCode.vue';
import { checkSteps, invalidateUpdates, saveText, useBundle, useUpdDevice, type Ref, type SaveError } from '../api/updater.api';
import { useUpdConfirm, useUpdRun } from '../hook/updActions.hook';
import { useUpdLiveText } from '../hook/updText.hook';
import { osSetOf, stepsExt, stepsTemplate } from '../util/updSteps.util';
import MbNote from '@/feature/member/component/MbNote.vue';

// steps.json 타일 — P1 = 원문 편집 + 검증(줄 단위 오류, 오류가 있으면 서버가 저장 거절). 시각 편집은 다음 단계 — REF-update-deploy.md "steps.json"·단위-6(편집 타일)
const props = defineProps<{ tile: MemberTile; bundle: Ref }>();
const { dropMissing, guardClose } = useMemberBoard();
const confirm = useUpdConfirm();
const { qc, busy, message, run } = useUpdRun();

const q = useBundle(props.bundle);
const b = computed(() => q.data.value);
watch(() => (q.error.value as { status?: number } | null)?.status, st => { if (st === 404) dropMissing(props.tile.key); }, { immediate: true });
const sha = computed(() => b.value?.files.find(f => f.scope === 'main' && f.path === 'steps.json')?.sha);
const exists = computed(() => !!sha.value);
const locked = computed(() => !!b.value?.locked);

// 틀 확장자 = 같은 메인키 설치본 os. windows만 = .ps1, 섞임·없음·공용 묶음 = .sh
const isBuild = props.bundle.kind === 'build';
const dq = useUpdDevice(props.bundle.owner, isBuild);
const osSet = computed(() => osSetOf(dq.data.value?.instances));
const mixedOs = computed(() => osSet.value.has('windows') && osSet.value.size > 1);
const ext = computed(() => stepsExt(osSet.value));
const osReady = computed(() => !isBuild || !dq.isPending.value);

const { original, text, dirty, stale, load, init, saved } = useUpdLiveText(props.bundle, 'steps.json', sha, m => message.value = m, () => { issues.value = []; checked.value = false; });
const issues = ref<CtCodeIssue[]>([]);
const checked = ref(false);
const editor = ref<InstanceType<typeof CtCode>>();

// 처음 한 번 불러온다(없으면 틀) — 다시 조회돼도 고치던 글은 그대로. 틀은 설치본 os를 받은 뒤 채워야 확장자가 맞다
watch([b, osReady], ([d, ready]) => {
    if (!d || original.value !== undefined) return;
    if (!exists.value) { if (ready) init(stepsTemplate(ext.value)); return; }
    void load(sha.value);
}, { immediate: true });
// 다른 곳에서 바뀐 뒤 고치던 것을 버리고 다시 읽기
const reload = async () => {
    if (await confirm('다시 불러오기', '고치던 내용을 버리고 저장된 내용을 불러옵니다.', '버리고 불러오기', 'error')) await load(sha.value);
}
watch(text, () => { checked.value = false; });

const errors = computed(() => issues.value.filter(i => i.level === 'error').length);
const warns = computed(() => issues.value.length - errors.value);

const check = () => run(() => checkSteps(props.bundle, text.value), r => { issues.value = r.issues; checked.value = true; });
const save = async () => {
    if (busy.value || locked.value) return;
    busy.value = true;
    message.value = undefined;
    try {
        const r = await saveText(props.bundle, 'steps.json', text.value);
        issues.value = r.issues;
        saved(r.file?.sha);
        checked.value = true;
        await invalidateUpdates(qc);
    } catch (e) {
        const err = e as SaveError;
        if (err.issues) { issues.value = err.issues; checked.value = true; }
        message.value = err.message || '저장하지 못했습니다';
    } finally {
        busy.value = false;
    }
}
onBeforeUnmount(guardClose(props.tile.key, () => !dirty.value || confirm('저장하지 않은 steps.json', '저장하지 않고 닫으면 고친 내용이 사라집니다.', '저장하지 않고 닫기', 'error')));
</script>

<template>
<div class="body" data-pager-press @keydown.ctrl.s.prevent="save" @keydown.meta.s.prevent="save">
    <p v-if="q.isError.value" class="mb-empty">빌드를 찾을 수 없습니다(지워졌을 수 있습니다)</p>
    <template v-else-if="original !== undefined">
        <p class="upd-hint top">{{ exists ? '원문 편집 · 시각 편집은 다음 단계' : '아직 steps.json이 없습니다. 틀을 고쳐 저장하면 만들어집니다' }}<span v-if="!exists && mixedOs" class="dirty"> · Linux·Windows 설치본이 섞여 있음(.sh 틀)</span><span v-if="dirty" class="dirty"> · 수정됨</span></p>
        <CtCode ref="editor" v-model="text" class="code" lang="json" :issues="issues" :readonly="locked" />
        <ul v-if="issues.length" class="issues">
            <li v-for="(i, n) in issues" :key="n" :class="i.level">
                <button type="button" @click="editor?.goto(i.line, i.col)"><b>{{ i.line }}행</b> {{ i.msg }}</button>
            </li>
        </ul>
        <p v-if="message" class="mb-err msg">{{ message }}</p>
    </template>
</div>
<div v-if="original !== undefined" class="mb-foot">
    <mb-note :text="locked ? '잠겨 있어 고칠 수 없습니다' : stale ? '다른 곳에서 steps.json이 바뀌었습니다' : checked ? `오류 ${errors}건 · 경고 ${warns}건` : 'Ctrl+S 저장 · 오류가 있으면 저장되지 않습니다'" />
    <div class="mb-end">
        <button v-if="stale" type="button" class="mb-btn" :disabled="busy" @click="reload">다시 불러오기</button>
        <button type="button" class="mb-btn" :disabled="busy" @click="check">검증</button>
        <button type="button" class="mb-btn mb-btn--primary" :disabled="busy || locked || (!dirty && exists)" @click="save">저장</button>
    </div>
</div>
</template>

<style scoped lang="scss">
.body { flex: 1; min-height: 0; display: flex; flex-direction: column; }
.top { flex: none; padding: 8px 16px; }
.dirty { color: rgb(var(--v-theme-warning)); }
.code { flex: 1; min-height: 120px; margin: 0 10px; }
.issues {
    flex: none;
    max-height: 30%;
    overflow-y: auto;
    margin: 8px 10px 0;
    padding: 0;
    list-style: none;
    li > button {
        width: 100%;
        padding: 4px 10px;
        border: 0;
        border-radius: 8px;
        background: none;
        color: inherit;
        font: inherit;
        text-align: left;
        @include vars.font('body-small');
        cursor: pointer;
        &:hover { background: rgba(var(--v-theme-on-surface), 0.05); }
    }
    .error b { color: rgb(var(--v-theme-error)); }
    .warn b { color: rgb(var(--v-theme-warning)); }
}
.msg { flex: none; margin: 8px 10px 0; }
.body > :last-child { margin-bottom: 10px; }
</style>
