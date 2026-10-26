<script setup lang="ts">
import FootPill from '@/feature/node/component/FootPill.vue';
import { computed } from 'vue';

// 푸터 액션 자리 — 1개 = 그 버튼 / 2개 이상 = 편집 아이콘 하나로 접고 위로 목록. 쓸 수 없는 것도 비활성으로 보인다(REF-device-move.md 이동-Q4(여는 곳))
export type FootAction = {
    key: string
    icon: string
    title: string
    disabled?: boolean
    hint?: string       // 비활성일 때 이유(마우스를 올리면)
    run: () => void
}
const props = defineProps<{ actions: FootAction[] }>();

const single = computed(() => props.actions.length === 1 ? props.actions[0] : undefined);
const tip = (a: FootAction) => a.disabled && a.hint ? `${a.title} - ${a.hint}` : a.title;
</script>

<template>
<foot-pill v-if="actions.length" class="FootActions">
    <!-- 비활성 버튼은 마우스 이벤트를 안 받아 title이 안 뜬다 -> 감싼 칸에 단다 -->
    <span v-if="single" class="cell" :title="tip(single)">
        <v-btn size="small" density="comfortable" variant="text" :icon="single.icon" :aria-label="tip(single)" :disabled="single.disabled" @click="single.run" />
    </span>
    <v-menu v-else location="top center" offset="10">
        <template #activator="{ props: act, isActive }">
            <v-btn v-bind="act" size="small" density="comfortable" variant="text" icon="mdi-pencil-outline" :active="isActive" title="작업" aria-label="작업 목록" />
        </template>
        <v-list density="compact" class="list elevation-2">
            <v-list-item v-for="a in actions" :key="a.key" :prepend-icon="a.icon" :title="a.title" :subtitle="a.disabled ? a.hint : undefined" :disabled="a.disabled" @click="a.run" />
        </v-list>
    </v-menu>
</foot-pill>
</template>

<style scoped lang="scss">
.cell { display: inline-flex; }
.list { border-radius: 12px; }
</style>
