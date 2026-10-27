<script setup lang="ts">
import { VTextarea } from 'vuetify/components';

// 찰떡 여러 줄 칸 — 알약 대신 둥근 상자(여러 줄이라 반원이 안 맞음). 높이 = rows. Vuetify 속성·슬롯은 그대로 넘긴다
defineProps({
    wall: {
        type: String as () => 'l' | 'r' | 'lr',
        default: undefined,
    },
})
</script>

<template>
<v-textarea class="ct-field ct-field--area" :class="wall && `ct-field--wall-${wall}`" variant="outlined" density="compact" hide-details="auto">
    <template v-for="(_, name) in $slots" #[name]="p">
        <slot :name="name" v-bind="p ?? {}" />
    </template>
</v-textarea>
</template>
