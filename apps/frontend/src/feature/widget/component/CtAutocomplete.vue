<script setup lang="ts">
import { VAutocomplete } from 'vuetify/components';
import { ctFieldClass, ctFieldProps } from '../util/ctField.util';

// 찰떡 찾아 고르기 칸 — Vuetify 속성·슬롯은 그대로 넘긴다. 펼침 목록 = 앱 테마를 따름
const props = defineProps(ctFieldProps);
</script>

<template>
<v-autocomplete :class="ctFieldClass(props)" variant="outlined" density="compact" hide-details="auto" :menu-props="{ contentClass: 'ct-menu' }">
    <template v-for="(_, name) in $slots" #[name]="p">
        <slot :name="name" v-bind="p ?? {}" />
    </template>
</v-autocomplete>
</template>
