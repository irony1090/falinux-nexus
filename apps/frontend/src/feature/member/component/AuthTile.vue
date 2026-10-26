<script setup lang="ts">
import { VIcon } from 'vuetify/components';

// 로그인·가입 공통 타일 — 가운데 하나(최대 380), 머리 줄 제목(+ 가입은 왼쪽 벽 "‹ 로그인"), 아래 줄 = 왼쪽 안내 + 오른쪽 끝 블록 — 화-16(로그인·가입)
defineProps<{ title: string; submitLabel: string; busy?: boolean; back?: boolean }>();
const emit = defineEmits<{ submit: []; back: [] }>();
</script>

<template>
<div class="AuthTile">
    <form class="mb-tile card" novalidate @submit.prevent="emit('submit')">
        <div class="mb-head head">
            <div v-if="back" class="mb-start">
                <button type="button" class="mb-btn" @click="emit('back')"><v-icon icon="mdi-chevron-left" size="16" />로그인</button>
            </div>
            <h1 class="title">{{ title }}</h1>
        </div>
        <div class="mb-fields top">
            <slot />
        </div>
        <div class="mb-foot foot">
            <span class="mb-note"><slot name="note" /></span>
            <div class="mb-end">
                <button type="submit" class="mb-btn mb-btn--primary submit" :disabled="busy">{{ submitLabel }}</button>
            </div>
        </div>
    </form>
</div>
</template>

<style scoped lang="scss">
.AuthTile {
    width: 100%;
    height: 100%;
    box-sizing: border-box;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 16px;
    background: rgb(var(--v-theme-background));
}
.card { width: 100%; max-width: 380px; }
.head { height: 40px; }
.title {
    margin: 0;
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    padding: 0 14px;
    @include vars.font('title-small');
    font-weight: vars.$weight-lg;
    color: rgb(var(--v-theme-primary));
}
.foot { height: 44px; }
.submit { padding: 0 24px 0 26px !important; @include vars.font('label-large'); }
:deep(.mb-note) { white-space: normal; gap: 6px; }
</style>
