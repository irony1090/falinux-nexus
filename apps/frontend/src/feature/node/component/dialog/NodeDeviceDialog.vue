<script setup lang="ts">
import { useOnlineWorkers } from '@/feature/worker/api/worker.api';
import { computed, ref, watch } from 'vue';
import { VBtn, VCard, VCardActions, VCardText, VCardTitle, VChip, VDialog, VForm, VTextField } from 'vuetify/components';

// 폴더의 장비(main_key) 지정 — 직접 입력 + 접속 중인 장비에서 고르기. worker 접속 전에도 미리 지정할 수 있다
const props = defineProps({
    title: {
        type: String,
        required: true,
    },
    initial: {
        type: String,
        default: () => '',
    },
    loading: {
        type: Boolean,
        default: () => false,
    },
})
const emit = defineEmits<{ submit: [deviceKey: string | null] }>();
const open = defineModel<boolean>({ type: Boolean, default: () => false });

const { counts } = useOnlineWorkers();

const key = ref<string | null>('');     // clearable이 null을 넣는다
const trimmed = computed(() => (key.value ?? '').trim());
const candidates = computed(() => [...counts.value.keys()].sort());

const status = computed(() => {
    if (!trimmed.value) return '지정 안 함 — 상위 폴더의 장비를 따릅니다';
    const n = counts.value.get(trimmed.value);
    return n ? `접속 중인 인스턴스 ${n}개` : '접속 중인 worker 없음 — 이 키로 worker가 접속하면 연결됩니다';
})

// 열릴 때마다 현재 값으로 되돌린다
watch(open, val => {
    if (val) key.value = props.initial;
})

const onSubmit = () => emit('submit', trimmed.value || null);
</script>

<template>
<v-dialog v-model="open" max-width="396" min-width="296" :persistent="loading">
    <v-card :loading="loading" :disabled="loading">
        <v-card-title>{{ title }}</v-card-title>
        <v-form @submit.prevent="onSubmit">
            <v-card-text>
                <v-text-field v-model="key"
                    label="장비 키 (main_key)"
                    autofocus
                    clearable
                    persistent-hint
                    :hint="status"
                />
                <div class="cands">
                    <span class="cands__label">접속 중인 장비</span>
                    <v-chip v-for="c in candidates" :key="c"
                        size="small"
                        :color="c === trimmed ? 'primary' : undefined"
                        @click="key = c"
                    >{{ c }} ({{ counts.get(c) }})</v-chip>
                    <span v-if="!candidates.length" class="cands__none">없음</span>
                </div>
            </v-card-text>
            <v-card-actions>
                <v-btn variant="text" @click="open = false">취소</v-btn>
                <v-btn type="submit" variant="flat" color="primary">확인</v-btn>
            </v-card-actions>
        </v-form>
    </v-card>
</v-dialog>
</template>

<style scoped lang="scss">
.cands {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: vars.$spacing-sm;
    margin-top: vars.$spacing-lg;
    font-size: 12px;
}
.cands__label,
.cands__none { opacity: 0.7; }
</style>
