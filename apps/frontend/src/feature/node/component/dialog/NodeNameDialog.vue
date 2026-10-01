<script setup lang="ts">
import { ref, watch } from 'vue';
import { VBtn, VCard, VCardActions, VCardText, VCardTitle, VDialog, VForm, VTextField } from 'vuetify/components';

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
    <v-card :loading="loading" :disabled="loading">
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
            <v-card-actions>
                <v-btn variant="text" @click="open = false">취소</v-btn>
                <v-btn type="submit" variant="flat" color="primary">확인</v-btn>
            </v-card-actions>
        </v-form>
    </v-card>
</v-dialog>
</template>
