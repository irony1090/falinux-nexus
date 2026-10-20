<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { useTileGrids } from '../store/tileGrids.store';

// 보이는 시간 / 스크롤이 이만큼 멈춰야 표시
const SHOW_MS = 2500;
const SETTLE_MS = 150;

// Grid 정보(옛 Grid 하단 정보 줄): 보이는 Grid 범위가 바뀌면 왼쪽 위에 잠깐, 누르면 바로 사라짐 — 푸터-6(Grid 정보 토스트)
const { grids, range, scrollX, scrubbing, overview } = useTileGrids();

const visible = ref(false);
const key = computed(() => `${range.value.lo}-${range.value.hi}`);
const label = computed(() => {
    const { lo, hi } = range.value;
    return lo === hi ? `${lo + 1}` : `${lo + 1}-${hi + 1}`;
});
const infos = computed(() => grids.value.slice(range.value.lo, range.value.hi + 1).map(g => `${g.axis ?? 'single'} · ${g.items.length}타일`));

let shownKey = '';
let settleTimer: ReturnType<typeof setTimeout> | undefined;
let hideTimer: ReturnType<typeof setTimeout> | undefined;
const hide = () => {
    clearTimeout(hideTimer);
    visible.value = false;
}
const show = () => {
    if (key.value === shownKey) return;
    shownKey = key.value;
    visible.value = true;
    clearTimeout(hideTimer);
    hideTimer = setTimeout(hide, SHOW_MS);
}

// 스크롤마다 다시 재고, 멈춘 뒤 범위가 바뀌었으면 표시. 인디케이터 드래그 중엔 놓을 때까지 대기
// 전체보기에서 돌아오면 다시 보여 준다(shownKey 초기화)
watch([key, scrollX, scrubbing, overview, () => grids.value.length], () => {
    clearTimeout(settleTimer);
    if (overview.value || !grids.value.length) {
        hide();
        shownKey = '';
        return;
    }
    if (scrubbing.value) return;
    settleTimer = setTimeout(show, SETTLE_MS);
}, { immediate: true });

onBeforeUnmount(() => {
    clearTimeout(settleTimer);
    clearTimeout(hideTimer);
});
</script>

<template>
<transition name="GridToast">
    <!-- pointerdown 기본 동작을 막아 터미널 포커스를 지킨다(포커스가 빠지면 가상 키보드가 닫힌다) -->
    <button v-if="visible" type="button" class="GridToast" aria-label="Grid 정보 닫기"
        @pointerdown.prevent
        @click="hide"
    >
        <b class="tag">Grid {{ label }}</b>
        <span>{{ infos.join(' / ') }}</span>
    </button>
</transition>
</template>

<style scoped lang="scss">
.GridToast {
    position: absolute;
    top: 14px;
    left: 14px;
    z-index: 4;
    display: flex;
    align-items: center;
    gap: 8px;
    height: 30px;
    padding: 0 12px 0 6px;
    border: 0;
    border-radius: 999px;
    background: rgba(var(--v-theme-on-surface), 0.86);
    color: rgb(var(--v-theme-surface));
    font-size: 12px;
    white-space: nowrap;
    cursor: pointer;
}
.tag {
    padding: 2px 8px;
    border-radius: 999px;
    background: rgb(var(--v-theme-primary));
    color: rgb(var(--v-theme-on-primary));
    font-size: 11px;
}
.GridToast-enter-active, .GridToast-leave-active { transition: opacity 0.25s, transform 0.25s; }
.GridToast-enter-from, .GridToast-leave-to {
    opacity: 0;
    transform: translateY(-6px);
}
</style>
