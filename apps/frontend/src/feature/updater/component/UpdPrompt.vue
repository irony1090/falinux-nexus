<script lang="ts">
import { inject, provide, ref, type InjectionKey, type Ref } from 'vue';

// 이름 입력 창(새 장비·새 빌드·새 버전·새 텍스트 파일) — 페이지에 하나, 타일이 prompt()로 연다
export type PromptField = { key: string; label: string; placeholder?: string; value?: string; optional?: boolean; hint?: string; pattern?: RegExp; patternMsg?: string; items?: { title: string; value: string }[] }
export type PromptSpec = { title: string; note?: string; ok: string; fields: PromptField[] }
type PromptState = { spec: PromptSpec; resolve: (v: Record<string, string> | null) => void }

const KEY: InjectionKey<{ state: Ref<PromptState | null>; prompt: (s: PromptSpec) => Promise<Record<string, string> | null> }> = Symbol('UpdPrompt');

export const provideUpdPrompt = () => {
    const state = ref<PromptState | null>(null);
    const prompt = (spec: PromptSpec) => new Promise<Record<string, string> | null>(resolve => {
        state.value?.resolve(null);
        state.value = { spec, resolve };
    });
    provide(KEY, { state, prompt });
    return { state, prompt };
}
export const useUpdPrompt = () => {
    const ctx = inject(KEY);
    if (!ctx) throw new Error('UpdPrompt is not provided');
    return ctx.prompt;
}
</script>

<script setup lang="ts">
import { computed, watch } from 'vue';
import { VCard, VCardTitle, VDialog } from 'vuetify/components';
import CtSelect from '@/feature/widget/component/CtSelect.vue';
import CtTextField from '@/feature/widget/component/CtTextField.vue';

const props = defineProps<{ state: PromptState | null }>();
const emit = defineEmits<{ done: [] }>();

const values = ref<Record<string, string>>({});
watch(() => props.state, s => {
    values.value = Object.fromEntries((s?.spec.fields ?? []).map(f => [f.key, f.value ?? '']));
}, { immediate: true });

const bad = (f: PromptField) => {
    const v = (values.value[f.key] ?? '').trim();
    return !!v && !!f.pattern && !f.pattern.test(v);
}
const ready = computed(() => !!props.state && props.state.spec.fields.every(f => f.items || ((f.optional || (values.value[f.key] ?? '').trim()) && !bad(f))));

const finish = (ok: boolean) => {
    const s = props.state;
    if (!s) return;
    if (ok && !ready.value) return;
    s.resolve(ok ? Object.fromEntries(Object.entries(values.value).map(([k, v]) => [k, v.trim()])) : null);
    emit('done');
}
</script>

<template>
<v-dialog :model-value="!!state" max-width="440" @update:model-value="v => !v && finish(false)">
    <v-card v-if="state" class="UpdPrompt ct-card">
        <v-card-title>{{ state.spec.title }}</v-card-title>
        <form class="fields" @submit.prevent="finish(true)">
            <template v-for="(f, i) in state.spec.fields" :key="f.key">
                <CtSelect v-if="f.items" v-model="values[f.key]" size="sm" :label="f.label" :items="f.items" hide-details="auto" :hint="f.hint" persistent-hint />
                <CtTextField v-else v-model="values[f.key]" size="sm" :label="f.label" :placeholder="f.placeholder" :autofocus="i === 0" autocomplete="off" spellcheck="false"
                    :hint="f.hint" persistent-hint :error-messages="bad(f) ? (f.patternMsg ?? '쓸 수 없는 글자가 있습니다') : undefined" />
            </template>
            <p v-if="state.spec.note" class="note">{{ state.spec.note }}</p>
            <button type="submit" hidden />
        </form>
        <div class="ct-actions">
            <span class="ct-grp ct-grp--r">
                <button type="button" class="ct-seg" @click="finish(false)">취소</button>
                <button type="button" class="ct-seg ct-seg--primary" :disabled="!ready" @click="finish(true)">{{ state.spec.ok }}</button>
            </span>
        </div>
    </v-card>
</v-dialog>
</template>

<style scoped lang="scss">
.fields {
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 6px vars.$spacing-xl 4px;
}
.note {
    margin: 0;
    @include vars.font('body-small');
    color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity));
}
</style>
