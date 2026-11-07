<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { tileBarAt, type TileBarAt } from '../util/tileBar.util';

// 타일 본문 스크롤 영역 — 메뉴 줄(#bar) + 목록. 루트 = 스크롤 요소(elRef로 잡는다) — REF-tile-frame.md 단계-2(메뉴 줄 위치·숨김)
// 메뉴 줄 = 스크롤 영역 안 sticky: 제 자리를 차지하므로 본문 패딩 계산이 없고, 두 줄이 돼도 맞는다(Q6(숨길 때 본문 높이))
const props = withDefaults(defineProps<{
    tag?: string;
    barAt?: TileBarAt;      // 없으면 설정값
    barHide?: boolean;      // 내리면 숨김
    barPin?: boolean;       // 숨기지 않음(경로 입력 중·끌기 중 등, 쓰는 쪽이 정함)
}>(), { tag: 'div', barAt: undefined, barHide: true });

const SHOW_PX = 24;     // 이만큼 올리면 다시 보임("약간 올리면")
const at = computed(() => props.barAt ?? tileBarAt.value);

const rootEl = ref<HTMLElement>();
const barEl = ref<HTMLElement>();
const hidden = ref(false);
const barH = ref(0);

// 메뉴 줄 안에 열린 팝업(v-menu 활성자) 또는 포커스가 있으면 숨기지 않는다
const held = () => props.barPin || !!barEl.value?.matches(':focus-within') || !!barEl.value?.querySelector('[aria-expanded="true"]');

let last = 0, acc = 0, lastH = 0;
const onScroll = () => {
    const el = rootEl.value;
    if (!el || !props.barHide) return;
    const max = el.scrollHeight - el.clientHeight;
    const st = Math.min(Math.max(el.scrollTop, 0), max);    // 끝에서 튕기는 값(관성) 잘라냄
    // 높이가 바뀌어 생긴 스크롤(모바일 키보드 열고 닫기·타일 크기)은 방향으로 보지 않는다 — REF-mobile-viewport.md
    if (el.clientHeight !== lastH) { lastH = el.clientHeight; last = st; acc = 0; return; }
    const d = st - last;
    last = st;
    if (st <= 0 || (at.value === 'bottom' && st >= max - 1)) { hidden.value = false; acc = 0; return; }   // 맨 위(아래 위치면 맨 끝도) = 항상 보임
    if (!d) return;
    acc = Math.sign(d) === Math.sign(acc) ? acc + d : d;
    if (acc > 0 && st > barH.value && !held()) hidden.value = true;     // 제 자리를 다 지난 뒤에만 숨김 — 빈 자리가 드러나지 않게
    else if (acc <= -SHOW_PX) hidden.value = false;
};

onMounted(() => { lastH = rootEl.value?.clientHeight ?? 0; last = rootEl.value?.scrollTop ?? 0; });

let ro: ResizeObserver | undefined;
watch(barEl, el => {
    ro?.disconnect();
    if (!el) return;
    ro = new ResizeObserver(() => { barH.value = el.offsetHeight; });
    ro.observe(el);
}, { flush: 'post' });
onBeforeUnmount(() => ro?.disconnect());
watch(() => props.barPin, pin => { if (pin) hidden.value = false; });
watch(at, () => { hidden.value = false; });

// 아래 위치일 때 아래에 붙는 다른 sticky(장비 타일 전송 줄)가 메뉴 줄 위에 오도록 — var(--tile-bar-bottom)
const style = computed(() => ({ '--tile-bar-bottom': at.value === 'bottom' && !hidden.value ? `${barH.value}px` : '0px' }));
</script>

<template>
<component :is="tag" ref="rootEl" class="TileScroll" :style="style" @scroll.passive="onScroll">
    <div v-if="$slots.bar && at === 'top'" ref="barEl" class="TileScroll__bar TileScroll__bar--top" :class="{ hidden }"><slot name="bar" /></div>
    <div class="TileScroll__body"><slot /></div>
    <div v-if="$slots.bar && at === 'bottom'" ref="barEl" class="TileScroll__bar TileScroll__bar--bottom" :class="{ hidden }"><slot name="bar" /></div>
</component>
</template>

<style scoped lang="scss">
:where(.TileScroll) {
    flex: 1;
    min-height: 0;
    overflow: auto;
    display: flex;
    flex-direction: column;
}
.TileScroll__body { flex: 1 0 auto; }    // 목록이 짧아도 아래 위치 메뉴 줄은 바닥에
.TileScroll__bar {
    flex: none;
    position: sticky;
    z-index: 3;     // 목록 칸 버튼·라벨(dir-grid z 1~2)보다 위 — 아니면 자동완성 목록을 뚫고 보인다
    display: flex;
    flex-direction: column;
    background: rgb(var(--v-theme-surface));
    transition: transform 0.18s ease;
}
.TileScroll__bar--top {
    top: 0;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    &.hidden { transform: translateY(-100%); }
}
// 아래 위치 = 순서를 뒤집어 본문 쪽이 위(로딩 줄 등이 목록과 맞닿음)
.TileScroll__bar--bottom {
    bottom: 0;
    flex-direction: column-reverse;
    border-top: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    &.hidden { transform: translateY(100%); }
}
</style>
