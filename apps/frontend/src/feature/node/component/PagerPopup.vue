<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue';
import { usePagerOpen, type PagerPoint } from '@/feature/widget/hook/pagerOpen.hook';
import { useTileGrids } from '../store/tileGrids.store';
import GridPager from './GridPager.vue';
import TileSizeCard from './tile/TileSizeCard.vue';

// 팝업 상자 폭(도트 자리) / 화면 가장자리 여백
const MAX_W = 560;
const EDGE = 16;

// 커서 위치에 인디케이터 하나 더 — 이동-2(커서 위치 페이지 표시) + 팝업-3(팝업 버튼)·팝업-6(길게 누르기로 띄우기), REF-pager-tools.md
// 여는 길(키·길게 누르기)은 usePagerOpen. 키는 폰·단축키 끈 타일에선 안 띄움, 전체보기 중엔 둘 다 안 띄움
const { overview, activeId, appShortcutsOff, enterOverview } = useTileGrids();

const pagerRef = ref<InstanceType<typeof GridPager>>();
const boxEl = ref<HTMLElement>();
const cardEl = ref<HTMLElement>();
const sizeOpen = ref(false);

// 알약 가운데 = 커서(터치면 누른 자리), 화면 밖으로 나가면 안쪽으로. 상자 폭이 고정이라 알약 크기는 위치와 무관해 바로 잰다
const at = ref({ x: 0, y: 0 });
const boxW = ref(MAX_W);
const clamp = (v: number, lo: number, hi: number) => Math.max(lo, Math.min(Math.max(lo, hi), v));
const place = (p: PagerPoint | null) => {
    sizeOpen.value = false;
    const { clientWidth: vw, clientHeight: vh } = document.documentElement;
    boxW.value = Math.min(MAX_W, vw - EDGE * 2);
    const c = p ?? { x: vw / 2, y: vh / 2 };
    const pill = boxEl.value?.querySelector<HTMLElement>('.FootPill');
    const w = pill?.offsetWidth ?? 0, h = pill?.offsetHeight ?? 0;
    at.value = { x: clamp(c.x, EDGE + w / 2, vw - EDGE - w / 2), y: clamp(c.y, EDGE + h / 2, vh - EDGE - h / 2) };
}

// 길게 누를 자리 = 타일 머리 줄 + 타일 밖 빈 곳(Q57-1) + 몸통의 `data-pager-press` 안 칸 밖(길게-2, REF-press-pager.md)
const pressable = (t: Element) => {
    if (t.closest('.TileFrame__head')) return true;
    const own = t.closest('[data-pager-press]');
    if (own) return !t.closest('.DirGridCell, [data-pager-press-off]');
    return !!t.closest('.TileStrip') && !t.closest('.TileFrame');
}
const { held, touchOpen, close } = usePagerOpen({
    off: () => overview.value,
    keyOff: () => appShortcutsOff.value,
    pressable,
    inside: t => t instanceof Node && (!!boxEl.value?.contains(t) || !!cardEl.value?.contains(t)),
    onOpen: place,
});

const opened = computed(() => (held.value || touchOpen.value) && !overview.value);
const dragging = computed(() => (pagerRef.value?.mode ?? 'idle') !== 'idle');
// 크기 카드가 뜨면 알약은 숨김 — Q55(카드 뜨면 팝업)
const shown = computed(() => (opened.value || dragging.value) && !overview.value && !sizeOpen.value);
const cardShown = computed(() => opened.value && sizeOpen.value);
watch(opened, o => { if (!o) sizeOpen.value = false; });

// 크기 카드 = 크기 버튼의 가로·세로 가운데 — Q56(카드 위치). 크기를 바꿔도 자리 고정(머리 줄 팝업처럼 따라가지 않음)
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
const goOverview = () => {
    close();
    enterOverview();
}

// 터치로 연 것 = 도트 누르기·드래그가 끝나면 닫힘. 키로 연 것은 키를 뗄 때까지 둔다(기존 동작)
const onPillClick = () => { if (touchOpen.value && !sizeOpen.value) touchOpen.value = false; }
watch(dragging, d => { if (!d && touchOpen.value) touchOpen.value = false; });
watch(overview, o => { if (o) close(); });
</script>

<template>
<teleport to="body">
    <!-- 늘 마운트(폭 측정·드래그 상태 유지), 보이기만 바꾼다. mousedown 막기 = xterm 포커스를 뺏지 않게 -->
    <div ref="boxEl" class="PagerPopup" :class="{ shown }"
        :style="{ left: `${at.x}px`, top: `${at.y}px`, width: `${boxW}px` }"
        aria-hidden="true"
        @mousedown.prevent
        @click="onPillClick"
    >
        <grid-pager ref="pagerRef" pull="both">
            <!-- Q54(버튼 모양·자리): 크기 = 맨 왼쪽 / 전체보기 = 푸터와 같은 모양, 아이콘만 -->
            <template #start>
                <v-btn class="size" variant="text" rounded="0" height="100%" icon
                    :disabled="!activeId"
                    title="active 타일 크기"
                    aria-label="active 타일 크기"
                    @click.stop="openSize"
                >
                    <v-icon icon="mdi-resize" size="18" />
                </v-btn>
            </template>
            <template #end>
                <v-btn color="primary" variant="flat" rounded="pill" height="100%" icon
                    title="전체보기"
                    aria-label="전체보기"
                    @click.stop="goOverview"
                >
                    <v-icon icon="mdi-view-grid" size="18" />
                </v-btn>
            </template>
        </grid-pager>
    </div>
    <div v-if="cardShown && activeId" ref="cardEl" class="PagerPopup__card"
        :style="{ left: `${cardAt.x}px`, top: `${cardAt.y}px` }"
        @mousedown.prevent
    >
        <tile-size-card :tile-id="activeId" />
    </div>
</teleport>
</template>

<style scoped lang="scss">
.PagerPopup {
    position: fixed;
    z-index: 2050;
    display: flex;
    height: 38px;
    --pill-h: 38px;
    transform: translate(-50%, -50%);
    visibility: hidden;
    pointer-events: none;
    &.shown { visibility: visible; }
    // 상자는 클릭을 통과시키고 알약만 받는다
    &.shown :deep(.FootPill) { pointer-events: auto; }
}
.size { border-right: 1px solid rgba(var(--v-border-color), var(--v-border-opacity)); }
.PagerPopup__card {
    position: fixed;
    z-index: 2060;
    transform: translate(-50%, -50%);
}
</style>
