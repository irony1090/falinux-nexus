<script setup lang="ts">
import { VTextField } from 'vuetify/components';
import { ctFieldClass, ctFieldProps } from '../util/ctField.util';

// 찰떡 입력칸 — Vuetify 속성·슬롯은 그대로 넘긴다
const props = defineProps(ctFieldProps);
</script>

<template>
<v-text-field :class="ctFieldClass(props)" variant="outlined" density="compact" hide-details="auto">
    <template v-for="(_, name) in $slots" #[name]="p">
        <slot :name="name" v-bind="p ?? {}" />
    </template>
</v-text-field>
</template>
