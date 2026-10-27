<script setup lang="ts">
import { ref, watch } from 'vue';
import { VCard, VCardText, VCardTitle, VDialog, VForm, VTextField } from 'vuetify/components';

// 노드 이름 입력 — 생성·이름 변경 공용
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
const emit = defineEmits<{ submit: [name: string] }>();
const open = defineModel<boolean>({ type: Boolean, default: () => false });

const name = ref('');
const message = ref<string>();

// 열릴 때마다 초기값으로 되돌린다
watch(open, val => {
    if (!val) return;
    name.value = props.initial;
    message.value = undefined;
})

const onSubmit = () => {
    const v = name.value.trim();
    if (!v) {
        message.value = '이름 입력이 필수입니다';
        return;
    }
    emit('submit', v);
}
</script>

<template>
<v-dialog v-model="open" max-width="396" min-width="296" :persistent="loading">
    <v-card :loading="loading" :disabled="loading" class="ct-card">
        <v-card-title>{{ title }}</v-card-title>
        <v-form @submit.prevent="onSubmit">
            <v-card-text>
                <v-text-field v-model="name"
                    label="이름"
                    autofocus
                    :error-messages="message"
                    @update:model-value="message = undefined"
                />
            </v-card-text>
            <div class="ct-actions">
                <span class="ct-grp ct-grp--r">
                    <button type="button" class="ct-seg" @click="open = false">취소</button>
                    <button type="submit" class="ct-seg ct-seg--primary">확인</button>
                </span>
            </div>
        </v-form>
    </v-card>
</v-dialog>
</template>
