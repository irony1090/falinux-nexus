<script setup lang="ts">
import { defaultKeymap, history, historyKeymap, insertTab } from '@codemirror/commands';
import { json } from '@codemirror/lang-json';
import { HighlightStyle, StreamLanguage, bracketMatching, indentOnInput, syntaxHighlighting } from '@codemirror/language';
import { powerShell } from '@codemirror/legacy-modes/mode/powershell';
import { shell } from '@codemirror/legacy-modes/mode/shell';
import { lintGutter, setDiagnostics, type Diagnostic } from '@codemirror/lint';
import { highlightSelectionMatches, search, searchKeymap } from '@codemirror/search';
import { Compartment, EditorState, type Extension } from '@codemirror/state';
import { EditorView, drawSelection, highlightActiveLine, highlightActiveLineGutter, keymap, lineNumbers, placeholder as cmPlaceholder } from '@codemirror/view';
import { tags } from '@lezer/highlight';
import { onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue';

// 찰떡 코드 편집 칸 — CodeMirror 6(줄 번호·줄 단위 오류). 모양은 CtTextarea와 같은 둥근 상자 + 벽(ct-field.scss), 색 = 테마 색 이름
// 스크립트(nodes.content)·updater 텍스트 파일·steps.json이 같이 쓴다 — REF-update-deploy.md "P1 착수 계획" 단위-7(CtCode)

export type CtCodeLang = 'sh' | 'ps1' | 'json' | 'text'
export type CtCodeIssue = { line: number; col?: number; level: 'error' | 'warn'; msg: string }

const props = defineProps<{
    lang?: CtCodeLang
    readonly?: boolean
    disabled?: boolean
    placeholder?: string
    autofocus?: boolean
    issues?: CtCodeIssue[]
    wall?: 'l' | 'r' | 'lr'
}>();
const model = defineModel<string>({ default: '' });

const host = ref<HTMLElement>();
const view = shallowRef<EditorView>();
const langBox = new Compartment();
const roBox = new Compartment();

const highlight = HighlightStyle.define([
    { tag: [tags.keyword, tags.controlKeyword, tags.operatorKeyword], color: 'rgb(var(--v-theme-primary))' },
    { tag: [tags.string, tags.special(tags.string)], color: 'rgb(var(--v-theme-success))' },
    { tag: [tags.number, tags.bool, tags.null, tags.atom], color: 'rgb(var(--v-theme-warning))' },
    { tag: [tags.propertyName, tags.variableName, tags.definition(tags.variableName)], color: 'rgb(var(--v-theme-info))' },
    { tag: [tags.comment, tags.lineComment, tags.blockComment], color: 'rgba(var(--v-theme-on-surface), 0.5)', fontStyle: 'italic' },
    { tag: tags.invalid, color: 'rgb(var(--v-theme-error))' },
]);

const langExt = (l?: CtCodeLang): Extension => {
    switch (l) {
        case 'sh': return StreamLanguage.define(shell);
        case 'ps1': return StreamLanguage.define(powerShell);
        case 'json': return json();
        default: return [];
    }
}
const roExt = () => [EditorState.readOnly.of(!!(props.readonly || props.disabled)), EditorView.editable.of(!(props.readonly || props.disabled))];

// 줄·칸(1부터, 글자 단위) -> 문서 위치. 범위를 넘으면 끝으로
const posOf = (s: EditorState, line: number, col = 1) => {
    const l = s.doc.line(Math.min(Math.max(1, line), s.doc.lines));
    return Math.min(l.from + Math.max(0, col - 1), l.to);
}
const pushIssues = () => {
    const v = view.value;
    if (!v) return;
    const ds: Diagnostic[] = (props.issues ?? []).map(i => {
        const from = posOf(v.state, i.line, i.col);
        const end = v.state.doc.lineAt(from).to;
        return { from, to: Math.max(from, end), severity: i.level === 'error' ? 'error' : 'warning', message: i.msg };
    });
    v.dispatch(setDiagnostics(v.state, ds));
}

onMounted(() => {
    view.value = new EditorView({
        parent: host.value!,
        state: EditorState.create({
            doc: model.value,
            extensions: [
                lineNumbers(),
                highlightActiveLineGutter(),
                highlightActiveLine(),
                history(),
                drawSelection(),
                indentOnInput(),
                bracketMatching(),
                highlightSelectionMatches(),
                search({ top: true }),
                lintGutter(),
                syntaxHighlighting(highlight),
                EditorState.tabSize.of(4),
                // Tab = 탭 문자(포커스 이동 아님) — 옛 ScriptEditDialog와 같음
                keymap.of([{ key: 'Tab', run: insertTab }, ...defaultKeymap, ...historyKeymap, ...searchKeymap]),
                props.placeholder ? cmPlaceholder(props.placeholder) : [],
                langBox.of(langExt(props.lang)),
                roBox.of(roExt()),
                EditorView.updateListener.of(u => {
                    if (u.docChanged) model.value = u.state.doc.toString();
                }),
            ],
        }),
    });
    pushIssues();
    if (props.autofocus) view.value.focus();
});
onBeforeUnmount(() => view.value?.destroy());

// 밖에서 바뀐 값(불러오기·되돌리기)만 문서에 넣는다 — 입력으로 바뀐 값은 이미 같다
watch(model, val => {
    const v = view.value;
    if (!v || v.state.doc.toString() === val) return;
    v.dispatch({ changes: { from: 0, to: v.state.doc.length, insert: val } });
});
watch(() => props.lang, l => view.value?.dispatch({ effects: langBox.reconfigure(langExt(l)) }));
watch(() => [props.readonly, props.disabled], () => view.value?.dispatch({ effects: roBox.reconfigure(roExt()) }));
watch(() => props.issues, pushIssues, { deep: true });

// 오류 목록에서 줄을 누르면 그 줄로
const goto = (line: number, col?: number) => {
    const v = view.value;
    if (!v) return;
    const pos = posOf(v.state, line, col);
    v.dispatch({ selection: { anchor: pos }, effects: EditorView.scrollIntoView(pos, { y: 'center' }) });
    v.focus();
}
const focus = () => view.value?.focus();
defineExpose({ goto, focus });
</script>

<script lang="ts">
// 파일 이름으로 언어 고르기(확장자)
export const ctCodeLangOf = (name: string): 'sh' | 'ps1' | 'json' | 'text' => {
    const n = name.toLowerCase();
    if (n.endsWith('.sh') || n.endsWith('.bash')) return 'sh';
    if (n.endsWith('.ps1') || n.endsWith('.psm1')) return 'ps1';
    if (n.endsWith('.json')) return 'json';
    return 'text';
}
</script>

<template>
<div ref="host" class="CtCode" :class="[wall && `CtCode--wall-${wall}`, { 'CtCode--disabled': disabled }]" />
</template>

<style scoped lang="scss">
// 상자 모양 = .ct-field--area와 같은 값(16px 둥근 상자, 벽 쪽만 곧게, 테두리 1px 색만 바뀜)
.CtCode {
    --r-l: 16px;
    --r-r: 16px;
    display: flex;
    min-height: 0;
    border-radius: var(--r-l) var(--r-r) var(--r-r) var(--r-l);
    background: rgba(var(--v-theme-on-surface), 0.035);
    box-shadow: inset 0 0 0 1px rgba(var(--v-theme-on-surface), 0.18);
    overflow: hidden;
    @media (hover: hover) {
        &:hover { box-shadow: inset 0 0 0 1px rgba(var(--v-theme-on-surface), 0.32); }
    }
    &:focus-within { box-shadow: inset 0 0 0 1px rgb(var(--v-theme-primary)); }
    &--wall-l { --r-l: 0; }
    &--wall-r { --r-r: 0; }
    &--wall-lr { --r-l: 0; --r-r: 0; }
    &--disabled { opacity: 0.6; }

    :deep(.cm-editor) {
        flex: 1;
        min-width: 0;
        height: 100%;
        background: transparent;
        color: rgb(var(--v-theme-on-surface));
        @include vars.font('body-medium');
        &.cm-focused { outline: none; }
    }
    :deep(.cm-scroller) {
        font-family: monospace;
        line-height: 1.5;
        overflow: auto;
    }
    :deep(.cm-content) { padding: 8px 0; caret-color: rgb(var(--v-theme-primary)); }
    :deep(.cm-line) { padding: 0 12px 0 6px; }
    :deep(.cm-cursor) { border-left-color: rgb(var(--v-theme-primary)); }
    :deep(.cm-gutters) {
        background: transparent;
        border-right: 1px solid rgba(var(--v-theme-on-surface), 0.1);
        color: rgba(var(--v-theme-on-surface), 0.45);
    }
    :deep(.cm-activeLine) { background: rgba(var(--v-theme-on-surface), 0.04); }
    :deep(.cm-activeLineGutter) { background: rgba(var(--v-theme-on-surface), 0.06); color: rgb(var(--v-theme-on-surface)); }
    :deep(.cm-selectionBackground), :deep(.cm-focused .cm-selectionBackground) { background: rgba(var(--v-theme-primary), 0.22) !important; }
    :deep(.cm-selectionMatch) { background: rgba(var(--v-theme-primary), 0.12); }
    :deep(.cm-matchingBracket) { background: rgba(var(--v-theme-primary), 0.18); outline: none; }
    :deep(.cm-placeholder) { color: rgba(var(--v-theme-on-surface), 0.4); }
    :deep(.cm-lintRange-error) { background-image: none; text-decoration: underline wavy rgb(var(--v-theme-error)); text-underline-offset: 3px; }
    :deep(.cm-lintRange-warning) { background-image: none; text-decoration: underline wavy rgb(var(--v-theme-warning)); text-underline-offset: 3px; }
    :deep(.cm-gutter-lint) { width: 12px; }
    :deep(.cm-panels) { background: rgb(var(--v-theme-surface)); color: inherit; border-color: rgba(var(--v-theme-on-surface), 0.12); }
    :deep(.cm-tooltip) { background: rgb(var(--v-theme-surface)); border: 1px solid rgba(var(--v-theme-on-surface), 0.18); border-radius: 8px; }
    :deep(.cm-diagnostic-error) { border-left-color: rgb(var(--v-theme-error)); }
    :deep(.cm-diagnostic-warning) { border-left-color: rgb(var(--v-theme-warning)); }
}
</style>
