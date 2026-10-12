<script setup lang="ts">
import { computed, type PropType } from 'vue';
import type { ActiveUser } from '../../store/remoteActive.store';

// 다른 계정 active 배지 — 닉네임 첫 글자 + 계정 색(user id % 6), 3명 넘으면 "+N" (REF-node-ui-active.md RA4(표시))
const props = defineProps({
    users: {
        type: Array as PropType<ActiveUser[]>,
        required: true,
    },
})

const MAX = 3;
const shown = computed(() => props.users.slice(0, MAX));
const rest = computed(() => props.users.slice(shown.value.length));
const color = (u: ActiveUser) => `bg-acct${(u.userId % 6) + 1}`;
const initial = (u: ActiveUser) => [...u.nickname][0] ?? '?';
</script>

<template>
<span v-if="users.length" class="ActiveBadges">
    <span v-for="u in shown" :key="u.userId" class="badge" :class="color(u)" :title="`${u.nickname} · 이 타일을 보는 중`">{{ initial(u) }}</span>
    <span v-if="rest.length" class="badge more" :title="rest.map(u => u.nickname).join('\n')">+{{ rest.length }}</span>
</span>
</template>

<style scoped lang="scss">
.ActiveBadges {
    display: inline-flex;
    gap: 3px;
}
.badge {
    width: 18px;
    height: 18px;
    border-radius: 50%;
    display: inline-grid;
    place-items: center;
    font-size: 10px;
    font-weight: vars.$weight-lg;
    line-height: 1;
    box-shadow: 0 0 0 1.5px rgb(var(--v-theme-surface));
    cursor: default;
}
.more {
    font-size: 9px;
    background: rgba(var(--v-theme-on-surface), 0.12);
    color: rgb(var(--v-theme-on-surface));
}
</style>
