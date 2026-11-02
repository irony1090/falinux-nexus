<script setup lang="ts">
import { useResizeCallback } from '@/feature/common/store/resizeGroup.store';
import { computed, ref, watch } from 'vue';
import { VIcon, VMenu } from 'vuetify/components';
import { useMemberBoard } from '../store/memberBoard.store';
import { memberDragFrom, useMemberDrag, type MemberChip } from '../store/memberDrag.store';
import type { Placed } from '../util/memberGrid.util';
import MemberSizeCard from './MemberSizeCard.vue';

// 상세 타일 공통 틀 — 머리 줄 = 위치 번호 + 경로 제목(화-11) + [크기 | 닫기] 끝 블록(보드 TileFrame과 같음, 화-10). 자식 타일은 옅은 파랑 테두리(화-9)
// 회원 페이지(MemberTileView)·updater 페이지(UpdTileView)가 경로·끌기 표시·몸통만 바꿔 같이 쓴다 — REF-update-deploy.md 단위-4(페이지 틀)
export type MbCrumb = { label: string; go?: () => void };
const props = defineProps<{ placed: Placed; crumbs: MbCrumb[]; chip: MemberChip; label: string }>();
const { posOf, activeKey, setActive, requestClose } = useMemberBoard();

const tile = computed(() => props.placed.tile);
const active = computed(() => activeKey.value === tile.value.key);

// 끌기(순서 바꾸기) — 머리 줄·몸통 빈 곳을 잡으면 시작. 끄는 동안 이 타일 자리 = 빈 자리
const { dragKey, pressKey, down } = useMemberDrag();
const slot = computed(() => dragKey.value === tile.value.key);
const pressing = computed(() => pressKey.value === tile.value.key);
const onDown = (e: PointerEvent) => {
    if (!memberDragFrom(e.target as Element, e.pointerType)) return;
    down(e, tile.value.key, props.chip);
};

// 크기 팝업은 크기가 바뀌면 버튼 자리를 따라간다(보드 relocateSizeMenu와 같음)
const frameEl = ref<HTMLElement>();
const sizeMenu = ref<InstanceType<typeof VMenu>>();
const sizeOpen = ref(false);
const relocateSizeMenu = () => requestAnimationFrame(() => sizeMenu.value?.updateLocation?.(new Event('resize')));
useResizeCallback(frameEl, relocateSizeMenu);
// 끌기 시작하면 열려 있던 팝업은 닫는다 — 열린 채로 끌면 팝업이 버튼을 따라가며 화면이 흔들린다
watch(dragKey, k => { if (k) sizeOpen.value = false; });
</script>

<template>
<section ref="frameEl" class="MbTileFrame mb-tile" :class="{ child: !!tile.parent, active, slot, pressing }" :data-tile-key="tile.key" tabindex="-1"
    :aria-label="label"
    @pointerdown.capture="setActive(tile.key)" @pointerdown="onDown"
>
    <div class="mb-head">
        <span class="pos">{{ posOf.get(tile.key) }}</span>
        <nav class="crumbs" aria-label="경로">
            <template v-for="(c, i) in crumbs" :key="i">
                <span v-if="i > 0" class="sep" aria-hidden="true">›</span>
                <span v-if="i === crumbs.length - 1" class="cur">{{ c.label }}</span>
                <button v-else type="button" class="link" :disabled="!c.go" @click="c.go?.()">{{ c.label }}</button>
            </template>
        </nav>
        <span class="tile-btn-end close" :class="{ 'tile-btn-end--primary': active }">
            <v-menu ref="sizeMenu" v-model="sizeOpen" location="bottom end" :close-on-content-click="false">
                <template #activator="{ props: act }">
                    <button v-bind="act" type="button" class="tile-btn" :title="`크기 ${placed.cw}×${placed.ch}칸`" aria-label="크기">
                        <v-icon size="16" icon="mdi-resize" />
                    </button>
                </template>
                <member-size-card :tile-key="tile.key" @changed="relocateSizeMenu" />
            </v-menu>
            <button type="button" class="tile-btn" title="닫기" aria-label="타일 닫기" @click="requestClose(tile.key)"><v-icon icon="mdi-close" size="16" /></button>
        </span>
    </div>
    <slot />
</section>
</template>

<style scoped lang="scss">
.MbTileFrame {
    &.child { border-color: rgba(var(--v-theme-primary), 0.35); }
    // active — 보드 TileFrame과 같은 바깥 선
    &.active { outline: 2px solid rgb(var(--v-theme-primary)); outline-offset: 1px; }
    // 끌기 시작 대기(터치로 누르는 중)
    &.pressing { outline: 2px solid rgba(var(--v-theme-primary), 0.45); outline-offset: 1px; }
    // 끄는 타일의 원래 자리 = 빈 자리. 바탕은 불투명(반투명이면 GPU 잔상, REF-tile-drag-impl.md 함정)
    &.slot {
        background: linear-gradient(rgba(var(--v-theme-primary), 0.08), rgba(var(--v-theme-primary), 0.08)), rgb(var(--v-theme-surface));
        > * { visibility: hidden; }
    }
    // 알약 타일 이동 뒤 포커스(키 입력 대상)만 받는다 — 포커스 선은 active 선으로 대신
    &:focus:not(.active) { outline: none; }
}
.crumbs {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 2px;
    padding: 0 6px 0 6px;
    @include vars.font('body-medium');
    white-space: nowrap;
    overflow: hidden;
}
.sep { flex: none; padding: 0 2px; opacity: 0.45; }
.link {
    flex: none;
    height: 26px;
    padding: 0 8px;
    border: 0;
    border-radius: 999px;
    background: rgba(var(--v-theme-on-surface), 0.06);
    color: rgba(var(--v-theme-on-surface), 0.8);
    font: inherit;
    @include vars.font('label-medium');
    font-weight: vars.$weight-md;
    cursor: pointer;
    &:hover:not(:disabled) { background: rgba(var(--v-theme-on-surface), 0.12); }
    &:disabled { cursor: default; }
}
.cur {
    min-width: 0;
    padding: 0 4px;
    font-weight: vars.$weight-lg;
    overflow: hidden;
    text-overflow: ellipsis;
}
// 위치 번호 — 보드 TileFrame .pos와 같은 모양
.pos {
    flex: none;
    align-self: center;
    min-width: 16px;
    height: 16px;
    margin-left: 10px;
    padding: 0 4px;
    border-radius: 4px;
    display: inline-grid;
    place-items: center;
    @include vars.font('label-small');
    font-weight: vars.$weight-lg;
    background: rgb(var(--v-theme-on-surface));
    color: rgb(var(--v-theme-surface));
}
// 닫기 = 오른쪽 벽 끝 블록(찰떡) — 오른쪽 면 직각
.close { border-radius: 999px 0 0 999px; }
</style>
