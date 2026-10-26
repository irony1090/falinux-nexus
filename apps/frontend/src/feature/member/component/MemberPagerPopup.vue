<script setup lang="ts">
import { usePagerOpen, type PagerPoint } from '@/feature/widget/hook/pagerOpen.hook';
import { useDigitShortcut } from '@/feature/widget/hook/digitShortcut.hook';
import { computed, nextTick, ref, watch } from 'vue';
import { VBtn, VIcon } from 'vuetify/components';
import { useMemberBoard } from '../store/memberBoard.store';
import MemberDots from './MemberDots.vue';
import MemberSizeCard from './MemberSizeCard.vue';

// 화면 가장자리 여백
const EDGE = 16;

// 관리 페이지 컨트롤 패널 — 보드 PagerPopup과 같은 여는 길(Ctrl+Alt 누르는 동안 / 길게 누르기, usePagerOpen).
// 알약 = [크기 | Grid 도트](전체보기 없음), 도트 끌기·당기기 = 보드와 같음. 크기 = active 타일, 카드는 버튼 가운데(Q56)·카드 뜨면 알약 숨김(Q55)
// Ctrl+Alt+숫자 = 위치 번호로 active 이동(useDigitShortcut)
const { activeKey, tiles, activate } = useMemberBoard();

const dotsRef = ref<InstanceType<typeof MemberDots>>();
const boxEl = ref<HTMLElement>();
const cardEl = ref<HTMLElement>();
const sizeOpen = ref(false);

const at = ref({ x: 0, y: 0 });
const clamp = (v: number, lo: number, hi: number) => Math.max(lo, Math.min(Math.max(lo, hi), v));
const place = async (p: PagerPoint | null) => {
    sizeOpen.value = false;
    const { clientWidth: vw, clientHeight: vh } = document.documentElement;
    const c = p ?? { x: vw / 2, y: vh / 2 };
    at.value = c;
    await nextTick();
    const w = boxEl.value?.offsetWidth ?? 0, h = boxEl.value?.offsetHeight ?? 0;
    at.value = { x: clamp(c.x, EDGE + w / 2, vw - EDGE - w / 2), y: clamp(c.y, EDGE + h / 2, vh - EDGE - h / 2) };
}

// 길게 누를 자리 = 타일 머리 줄(목록 포함) + 타일 밖 빈 곳 — Q57-1(길게 누를 자리)과 같은 규칙
const pressable = (t: Element) => !!t.closest('.mb-head') || (!!t.closest('.MemberPage .body') && !t.closest('.mb-tile'));
const { held, touchOpen, close } = usePagerOpen({
    off: () => false,
    keyOff: () => false,
    pressable,
    inside: t => t instanceof Node && (!!boxEl.value?.contains(t) || !!cardEl.value?.contains(t)),
    onOpen: p => void place(p),
});
useDigitShortcut({
    off: () => false,
    count: () => tiles.value.length,
    go: pos => {
        const t = tiles.value[pos - 1];
        if (t) activate(t.key);
    },
});

const opened = computed(() => held.value || touchOpen.value);
const dragging = computed(() => (dotsRef.value?.mode ?? 'idle') !== 'idle');
const shown = computed(() => (opened.value || dragging.value) && !sizeOpen.value);
const cardShown = computed(() => opened.value && sizeOpen.value && !!activeKey.value);
watch(opened, o => { if (!o) sizeOpen.value = false; });

const cardAt = ref({ x: 0, y: 0 });
const openSize = async (e: MouseEvent) => {
    const r = (e.currentTarget as HTMLElement).getBoundingClientRect();
    cardAt.value = { x: r.left + r.width / 2, y: r.top + r.height / 2 };
    sizeOpen.value = true;
    await nextTick();
    const { clientWidth: vw, clientHeight: vh } = document.documentElement;
    const w = cardEl.value?.offsetWidth ?? 0, h = cardEl.value?.offsetHeight ?? 0;
    cardAt.value = { x: clamp(cardAt.value.x, EDGE + w / 2, vw - EDGE - w / 2), y: clamp(cardAt.value.y, EDGE + h / 2, vh - EDGE - h / 2) };
}
// 터치로 연 것 = 도트 누르기·끌기가 끝나면 닫힘. 키로 연 것은 키를 뗄 때까지(보드 PagerPopup과 같음)
const onPillClick = () => { if (touchOpen.value && !sizeOpen.value) close(); }
watch(dragging, d => { if (!d && touchOpen.value) touchOpen.value = false; });
</script>

<template>
<teleport to="body">
    <!-- mousedown 막기 = 입력칸 포커스를 뺏지 않게 -->
    <div ref="boxEl" class="MemberPagerPopup" :class="{ shown }"
        :style="{ left: `${at.x}px`, top: `${at.y}px` }"
        aria-hidden="true"
        @mousedown.prevent
        @click="onPillClick"
    >
        <member-dots ref="dotsRef" pull="both">
            <template #start>
                <v-btn class="size" variant="text" rounded="0" height="100%" icon
                    :disabled="!activeKey"
                    title="active 타일 크기"
                    aria-label="active 타일 크기"
                    @click.stop="openSize"
                >
                    <v-icon icon="mdi-resize" size="18" />
                </v-btn>
            </template>
        </member-dots>
    </div>
    <div v-if="cardShown" ref="cardEl" class="MemberPagerPopup__card"
        :style="{ left: `${cardAt.x}px`, top: `${cardAt.y}px` }"
        @mousedown.prevent
    >
        <member-size-card :tile-key="activeKey!" />
    </div>
</teleport>
</template>

<style scoped lang="scss">
.MemberPagerPopup {
    position: fixed;
    z-index: 2050;
    display: flex;
    height: 38px;
    --pill-h: 38px;
    transform: translate(-50%, -50%);
    visibility: hidden;
    pointer-events: none;
    &.shown { visibility: visible; pointer-events: auto; }
}
.size { border-right: 1px solid rgba(var(--v-border-color), var(--v-border-opacity)); }
.MemberPagerPopup__card {
    position: fixed;
    z-index: 2060;
    transform: translate(-50%, -50%);
}
</style>
