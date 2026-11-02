<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { useMemberBoard } from '@/feature/member/store/memberBoard.store';
import type { MemberTile } from '@/feature/member/util/memberGrid.util';
import CtCode, { type CtCodeIssue } from '@/feature/widget/component/CtCode.vue';
import { checkSteps, invalidateUpdates, readText, saveText, useBundle, type Ref, type SaveError } from '../api/updater.api';
import { useUpdConfirm, useUpdRun } from '../hook/updActions.hook';
import MbNote from '@/feature/member/component/MbNote.vue';

// steps.json 타일 — P1 = 원문 편집 + 검증(줄 단위 오류, 오류가 있으면 서버가 저장 거절). 시각 편집은 다음 단계 — REF-update-deploy.md "steps.json"·단위-6(편집 타일)
const TEMPLATE = `{
  "uses": [],
  "deploy": [
    { "id": "stop", "script": "00-stop.sh" },
    { "id": "backup", "builtin": "backup" },
    { "id": "copy", "builtin": "copy" },
    { "id": "start", "script": "03-start.sh" }
  ],
  "rollback": [
    { "id": "stop", "script": "00-stop.sh" },
    { "id": "restore", "builtin": "restore" },
    { "id": "start", "script": "03-start.sh" }
  ]
}
`;
const props = defineProps<{ tile: MemberTile; bundle: Ref }>();
const { dropMissing, guardClose } = useMemberBoard();
const confirm = useUpdConfirm();
const { qc, busy, message, run } = useUpdRun();

const q = useBundle(props.bundle);
const b = computed(() => q.data.value);
watch(() => (q.error.value as { status?: number } | null)?.status, st => { if (st === 404) dropMissing(props.tile.key); }, { immediate: true });
const exists = computed(() => !!b.value?.files.some(f => f.scope === 'main' && f.path === 'steps.json'));
const locked = computed(() => !!b.value?.locked);

const original = ref<string>();
const text = ref('');
const issues = ref<CtCodeIssue[]>([]);
const checked = ref(false);
const dirty = computed(() => original.value !== undefined && text.value !== original.value);
const editor = ref<InstanceType<typeof CtCode>>();

// 처음 한 번 불러온다(없으면 틀) — 다시 조회돼도 고치던 글은 그대로
watch(b, d => {
    if (!d || original.value !== undefined) return;
    if (!exists.value) { original.value = ''; text.value = TEMPLATE; return; }
    readText(props.bundle, 'steps.json').then(t => { original.value = t; text.value = t; }).catch(e => message.value = e?.message || '불러오지 못했습니다');
}, { immediate: true });
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
        original.value = text.value;
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
        <p class="upd-hint top">{{ exists ? '원문 편집 · 시각 편집은 다음 단계' : '아직 steps.json이 없습니다. 틀을 고쳐 저장하면 만들어집니다' }}<span v-if="dirty" class="dirty"> · 수정됨</span></p>
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
    <mb-note :text="locked ? '잠겨 있어 고칠 수 없습니다' : checked ? `오류 ${errors}건 · 경고 ${warns}건` : 'Ctrl+S 저장 · 오류가 있으면 저장되지 않습니다'" />
    <div class="mb-end">
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
