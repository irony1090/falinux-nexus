<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch, type PropType } from 'vue';

// 경로 직접 입력 + 자동완성 — 경로-3(화면: 입력칸)·경로-4(화면: 자동완성), REF-path-input.md. 장비 타일·폴더 타일이 같이 쓴다
export type PathCand = {
    key: string
    label: string
    fill: string        // 고르면 입력칸에 들어갈 전체 글자(구분자로 끝남 = 바로 다음 단계 후보)
    hint?: string
}

const props = defineProps({
    value: { type: String, required: true },
    // 후보 — 실패·옛 worker면 빈 목록
    suggest: { type: Function as PropType<(text: string) => Promise<{ items: PathCand[]; more?: boolean }>>, required: true },
    // Enter — 오류 문구를 돌려주면 입력칸에 남고, 없으면 닫는다(이동은 부른 쪽이)
    submit: { type: Function as PropType<(text: string) => Promise<string | undefined>>, required: true },
})
const emit = defineEmits<{ close: []; pick: [c: PathCand] }>();

const SUGGEST_WAIT_MS = 120

const inputEl = ref<HTMLInputElement>();
const listEl = ref<HTMLElement>();
const text = ref(props.value);
const items = ref<PathCand[]>([]);
const more = ref(false);
const hi = ref(-1);         // ↑↓로 고른 후보(-1 = 없음 -> Enter는 이동)
const err = ref('');
const busy = ref(false);

let seq = 0;
let timer: ReturnType<typeof setTimeout> | undefined;
const load = async () => {
    const my = ++seq;
    const t = text.value;
    try {
        const r = await props.suggest(t);
        if (my !== seq) return;     // 그사이 더 쳤음
        items.value = r.items;
        more.value = !!r.more;
    } catch {
        if (my !== seq) return;
        items.value = [];
        more.value = false;
    }
    hi.value = -1;
}
const loadSoon = () => {
    clearTimeout(timer);
    timer = setTimeout(() => void load(), SUGGEST_WAIT_MS);
}

// 글자가 바뀌면 옛 목록·강조를 바로 버린다 — 새 후보가 오기 전 Enter가 옛 후보를 다시 고르지 않게
const drop = () => {
    items.value = [];
    more.value = false;
    hi.value = -1;
}

const onInput = () => {
    err.value = '';
    hi.value = -1;
    loadSoon();
}

const choose = (c: PathCand) => {
    text.value = c.fill;
    err.value = '';
    drop();
    emit('pick', c);
    clearTimeout(timer);
    void load();
    void nextTick(() => {
        const el = inputEl.value;
        if (el) el.selectionStart = el.selectionEnd = el.value.length;
    });
}

// Tab: 후보 하나면 그것, 여럿이면 공통 앞부분까지(셸과 같음)
const complete = () => {
    const list = items.value;
    if (!list.length) return;
    if (list.length === 1) return choose(list[0]!);
    let common = list[0]!.fill;
    for (const c of list) {
        let i = 0;
        while (i < common.length && i < c.fill.length && common[i] === c.fill[i]) i++;
        common = common.slice(0, i);
    }
    if (common.length > text.value.length) {
        text.value = common;
        hi.value = -1;
        loadSoon();
    }
}

const doSubmit = async () => {
    if (busy.value) return;
    busy.value = true;
    try {
        const e = await props.submit(text.value);
        if (e) err.value = e;
        else emit('close');
    } finally {
        busy.value = false;
    }
}

const move = (d: number) => {
    const n = items.value.length;
    if (!n) return;
    hi.value = hi.value < 0 ? (d > 0 ? 0 : n - 1) : (hi.value + d + n) % n;
    void nextTick(() => listEl.value?.children[hi.value]?.scrollIntoView({ block: 'nearest' }));
}

const onKey = (e: KeyboardEvent) => {
    if (e.isComposing) return;
    switch (e.key) {
    case 'ArrowDown': e.preventDefault(); move(1); break;
    case 'ArrowUp': e.preventDefault(); move(-1); break;
    case 'Tab': e.preventDefault(); complete(); break;
    case 'Escape': e.preventDefault(); e.stopPropagation(); emit('close'); break;
    case 'Enter': {
        e.preventDefault();
        const c = items.value[hi.value];
        if (c) choose(c);
        else void doSubmit();
        break;
    }
    }
}

watch(() => props.value, v => { text.value = v; });
onMounted(() => {
    inputEl.value?.focus();
    inputEl.value?.select();
});
onBeforeUnmount(() => {
    clearTimeout(timer);
    seq++;
});
</script>

<template>
<span class="PathInput">
    <input ref="inputEl" v-model="text" class="box" :class="{ bad: !!err }" spellcheck="false" autocomplete="off"
        aria-label="경로" @input="onInput" @keydown="onKey" @blur="emit('close')">
    <!-- mousedown.prevent: 고르는 동안 입력칸이 blur(= 닫힘)되지 않게 -->
    <span v-if="err || items.length" class="drop" @mousedown.prevent>
        <span v-if="err" class="err">{{ err }}</span>
        <span ref="listEl" class="list">
            <button v-for="(c, i) in items" :key="c.key" type="button" class="cand" :class="{ hi: i === hi }" tabindex="-1" @click="choose(c)">
                <span class="nm">{{ c.label }}</span><span v-if="c.hint" class="hint">{{ c.hint }}</span>
            </button>
        </span>
        <span v-if="more" class="more">더 있음 — 더 쳐서 좁히기</span>
    </span>
</span>
</template>

<style scoped lang="scss">
.PathInput {
    position: relative;
    flex: 1 1 200px;
    min-width: 0;
    display: flex;
}
.box {
    flex: 1;
    min-width: 0;
    padding: 1px 4px;
    font: inherit;
    font-family: monospace;
    color: inherit;
    border: 1px solid rgb(var(--v-theme-primary));
    border-radius: 3px;
    outline: none;
    background: rgb(var(--v-theme-surface));
    &.bad { border-color: rgb(var(--v-theme-error)); }
}
.drop {
    position: absolute;
    top: calc(100% + 2px);
    left: 0;
    right: 0;
    z-index: 5;
    display: flex;
    flex-direction: column;
    max-height: 240px;
    background: rgb(var(--v-theme-surface));
    border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    border-radius: 3px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}
.err { padding: 3px 6px; color: rgb(var(--v-theme-error)); }
.list { display: flex; flex-direction: column; overflow-y: auto; }
.cand {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 3px 6px;
    border: 0;
    background: none;
    text-align: left;
    color: inherit;
    font: inherit;
    font-family: monospace;
    cursor: pointer;
    &.hi, &:hover { background: rgba(var(--v-theme-primary), 0.12); }
    .nm { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .hint { margin-left: auto; opacity: 0.6; @include vars.font('label-small'); }
}
.more { padding: 3px 6px; opacity: 0.6; @include vars.font('label-small'); }
</style>
