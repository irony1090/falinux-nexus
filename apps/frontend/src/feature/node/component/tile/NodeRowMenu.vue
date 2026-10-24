<script setup lang="ts" generic="A extends string">
import type { PropType } from 'vue';
import { VIcon, VList, VListItem, VMenu } from 'vuetify/components';

export type RowMenuItem<A extends string = string> = { act: A; icon: string; title: string; disabled?: boolean; danger?: boolean }

// 목록 칸의 조작 메뉴 — 동작은 부모가 처리. ⋮ = 칸 오른쪽 위 버튼, at이 있으면 그 좌표에 뜬다(우클릭, SM1(켜는 방법))
// head 슬롯 = 맨 위 정보 카드(판 D′, hover 대신)
defineProps({
    items: {
        type: Array as PropType<RowMenuItem<A>[]>,
        required: true,
    },
    at: Array as unknown as PropType<[number, number]>,
})
const open = defineModel<boolean>({ default: false });
const emit = defineEmits<{ act: [act: A] }>();
</script>

<template>
<v-menu v-model="open" :location="at ? 'bottom start' : 'bottom end'" :target="at">
    <template v-if="!at" #activator="{ props: act }">
        <button v-bind="act" class="dg-btn dg-btn--r" :class="{ on: open }" title="메뉴"><v-icon size="14" icon="mdi-dots-vertical" /></button>
    </template>
    <v-list density="compact">
        <slot name="head" />
        <v-list-item v-for="it in items" :key="it.act"
            :prepend-icon="it.icon"
            :title="it.title"
            :disabled="it.disabled"
            :base-color="it.danger ? 'error' : undefined"
            @click="emit('act', it.act)"
        />
    </v-list>
</v-menu>
</template>

<style scoped lang="scss">
.dg-btn.on { background: rgba(var(--c), 0.3); }
</style>
