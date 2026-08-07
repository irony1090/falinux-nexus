<script setup lang="ts">
import { provideResizeGroupStore } from '@/feature/common/store/resizeGroup.store';
import { watch } from 'vue';

const {
    flag
} = provideResizeGroupStore();

const model = defineModel<boolean>({
    type: Boolean,
    default: () => false,
})

watch(model, val => {
    flag.value = val;
}, { immediate: true })

watch(flag , val => {
    if (model.value !== val) model.value = val;
})

</script>
<template>
<slot />
</template>