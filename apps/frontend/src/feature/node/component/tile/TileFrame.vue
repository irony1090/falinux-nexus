<script setup lang="ts">
import { computed, onBeforeUnmount, ref, toRef, watchEffect, type PropType } from 'vue';
import { useResizeCallback } from '@/feature/common/store/resizeGroup.store';
import { VIcon, VList, VMenu, VProgressCircular } from 'vuetify/components';
import { useTileFloat } from '../../hook/tileFloat.hook';
import { useTileSize } from '../../hook/tileSize.hook';
import { clearTileHeadActs, setTileHeadActs, type TileHeadActs } from '../../store/tileHeadActs.store';
import { useTileGrids } from '../../store/tileGrids.store';
import { useTileTree } from '../../store/tileTree.store';
import type { ActiveUser } from '../../store/remoteActive.store';
import ActiveBadges from './ActiveBadges.vue';
import TileSizeCard from './TileSizeCard.vue';

// 타일 공통 틀 — 머리 줄(위치 번호·아이콘·이름 자리·공유 단계·active 배지·메뉴·[최소화 | 크기 | ×])·볼 수 없음 표시·전체보기 라벨.
// 종류별 내용은 종류 컴포넌트(TILE_COMPONENTS)가 slot과 props로 채운다 — 정리-5(TileFrame 나누기), REF-tile-kinds.md
const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
    name: {
        type: String,
        default: '',
    },
    icon: {
        type: String,
        required: true,
    },
    color: {
        type: String,
        required: true,
    },
    shared: Boolean,                // 공유받은 것 — 공유 색 테두리·머리 줄
    shareInfo: {
        type: Object as PropType<{ icon: string; text: string }>,
        default: undefined,
    },
    unavailable: Boolean,           // 볼 수 없는 타일(공유 취소·삭제·범위 밖 이동) — 실시간-6b. 닫기만 가능
    remoteUsers: {
        type: Array as PropType<ActiveUser[]>,
        default: () => [],
    },
    menu: Boolean,                  // ⋮ 메뉴를 그림(항목 = #menu)
    menuBusy: Boolean,
    minimizable: Boolean,
    closeTitle: {
        type: String,
        default: '닫기',
    },
    closeDisabled: Boolean,
    floatUid: {                     // 맞춤-1(타일 띄우기) 대상 터미널 uid
        type: String,
        default: undefined,
    },
})

const emit = defineEmits<{
    close: []
    minimize: []
}>()

const { tiles } = useTileTree();
const { posOf, overview, activeId, setActive } = useTileGrids();
const { shownText, sizeTitle } = useTileSize(toRef(props, 'tileId'));

const tile = computed(() => tiles.value[props.tileId]);
const isRoot = computed(() => !tile.value?.parent);

// 크기를 바꾸면 버튼이 옮겨지는데 v-menu는 이를 감지 못함 → 프레임 크기 변화·크기 변경 직후에 위치 재계산
const frameEl = ref<HTMLElement>();
const sizeMenu = ref<InstanceType<typeof VMenu>>();
const relocateSizeMenu = () => requestAnimationFrame(() => sizeMenu.value?.updateLocation?.(new Event('resize')));
useResizeCallback(frameEl, relocateSizeMenu);

// 맞춤-1(타일 띄우기) — 키보드가 열린 동안 포커스된 터미널 타일만 띄운다
const floatStyle = useTileFloat(frameEl, computed(() => props.floatUid), overview);

// 컨트롤 패널이 쓸 창 버튼 — 머리 줄 [최소화 | 크기 | ×]와 같은 조건·동작(파트(패널 창 버튼))
let headActs: TileHeadActs | undefined;
watchEffect(() => setTileHeadActs(props.tileId, headActs = {
    minimize: props.minimizable && !props.unavailable ? () => emit('minimize') : undefined,
    sizable: !props.unavailable,
    close: tile.value && !isRoot.value ? () => emit('close') : undefined,
    closeTitle: props.closeTitle,
    closeDisabled: props.closeDisabled,
}));
onBeforeUnmount(() => clearTileHeadActs(props.tileId, headActs));

// active 표시 — AT3(표시). active가 있을 때만 나머지를 흐린다
const activeClass = computed(() => activeId.value ? (activeId.value === props.tileId ? 'TileFrame--active' : 'TileFrame--dim') : undefined);
</script>

<template>
<div v-if="tile" ref="frameEl" class="TileFrame" tabindex="-1"
    :data-tile-id="tile.id"
    :class="[`TileFrame--${tile.type}`, activeClass, { 'TileFrame--overview': overview, 'TileFrame--shared': shared }]"
    :style="floatStyle"
    @pointerdown.capture="setActive(tile.id)"
>
    <!-- 전체보기 라벨: 축소 배율을 거꾸로 곱해 읽을 수 있는 크기 유지 -->
    <div v-if="overview" class="ov-name">
        <span class="pos">{{ posOf.get(tile.id) }}</span>
        <v-icon size="14" :icon="icon" :color="color" />
        <span class="name">{{ name }}</span>
        <active-badges :users="remoteUsers" />
        <slot name="ov-extra" />
    </div>
    <div class="TileFrame__head" :inert="overview">
        <span class="pos">{{ posOf.get(tile.id) }}</span>
        <v-icon size="14" :icon="icon" :color="color" />
        <slot name="title" />
        <span class="spacer" />
        <span v-if="shareInfo" class="perm" :title="`공유 단계: ${shareInfo.text}`">
            <v-icon size="14" :icon="shareInfo.icon" />{{ shareInfo.text }}
        </span>
        <active-badges :users="remoteUsers" />
        <!-- 버튼 = 공통 모양(styles/tile-buttons.scss). 오른쪽 끝 창 묶음만 타일 쪽 면이 직각 -->
        <span class="TileFrame__btns tile-btn-row-end">
            <slot name="controls" />
            <span v-if="$slots.controls && menu" class="tile-btn-sep" />
            <v-menu v-if="menu" location="bottom end">
                <template #activator="{ props: act }">
                    <button v-bind="act" class="tile-btn" title="메뉴" :disabled="menuBusy">
                        <v-progress-circular v-if="menuBusy" indeterminate size="14" width="2" />
                        <v-icon v-else size="16" icon="mdi-dots-vertical" />
                    </button>
                </template>
                <v-list density="compact" min-width="180">
                    <slot name="menu" />
                </v-list>
            </v-menu>

            <!-- 창 버튼 묶음 [최소화 | 크기 | ×] — SD3(실행 중 터미널 치우기). × 동작(kill 확인 등)은 종류 컴포넌트가 정한다 -->
            <span class="tile-btn-end" :class="{ 'tile-btn-end--primary': activeId === tileId }">
                <button v-if="minimizable && !unavailable" class="tile-btn" title="최소화" @click="emit('minimize')">
                    <v-icon size="16" icon="mdi-window-minimize" />
                </button>
                <v-menu v-if="!unavailable" ref="sizeMenu" location="bottom end" :close-on-content-click="false">
                    <template #activator="{ props: act }">
                        <button v-bind="act" class="tile-btn" :class="{ 'text-warning': shownText }" :title="sizeTitle">
                            <v-icon size="16" icon="mdi-resize" />
                        </button>
                    </template>
                    <tile-size-card :tile-id="tile.id" @changed="relocateSizeMenu" />
                </v-menu>
                <button v-if="!isRoot" class="tile-btn" :disabled="closeDisabled" :title="closeTitle" @click="emit('close')">
                    <v-icon size="16" icon="mdi-close" />
                </button>
            </span>
        </span>
    </div>

    <slot name="tabs" />
    <div v-if="unavailable" class="revoked text-warning">
        <v-icon size="18" icon="mdi-link-off" />
        <span>공유가 취소되었거나 삭제되어 볼 수 없습니다.</span>
        <slot name="unavailable"><span class="sub">닫기만 할 수 있습니다.</span></slot>
    </div>
    <slot v-else />
    <slot name="extra" />
</div>
</template>

<style scoped lang="scss">
.TileFrame {
    flex: 1;
    min-width: 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    border-radius: 6px;
    background: rgb(var(--v-theme-surface));
    position: relative;
}
.revoked {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
    padding: vars.$spacing-md;
    text-align: center;
    @include vars.font('body-medium');
    :deep(.sub) { @include vars.font('body-small'); opacity: 0.7; }
}
// 전체보기: 안쪽 조작 막음(inert + 클릭은 Tile로)
.TileFrame--overview {
    > :not(.ov-name) { pointer-events: none; }
}
// active — 공유 테두리(border)와 겹쳐도 둘 다 보이게 outline. 포커스용 기본 outline은 끈다
.TileFrame:focus { outline: none; }
.TileFrame.TileFrame--active {
    outline: 2px solid rgb(var(--v-theme-primary));
    outline-offset: 1px;
    > .TileFrame__head { border-bottom-color: rgba(var(--v-theme-primary), 0.4); }   // 끝 블록(진한 파랑)과 구분되는 옅은 파랑
    // 전체보기 상자(overflow hidden)에 잘리지 않게 안쪽으로, 축소 배율을 거꾸로 곱해 화면에서 2px
    &.TileFrame--overview {
        outline-width: calc(2px / var(--s));
        outline-offset: calc(-2px / var(--s));
    }
}
.TileFrame--dim {
    opacity: 0.7;
    &:hover { opacity: 0.85; }
}
.ov-name {
    position: absolute;
    left: 0;
    top: 0;
    z-index: 3;
    width: calc(100% * var(--s));
    transform: scale(calc(1 / var(--s)));
    transform-origin: 0 0;
    display: flex;
    align-items: center;
    gap: vars.$spacing-sm;
    padding: 3px 6px;
    @include vars.font('body-small');
    background: rgb(var(--v-theme-surface));
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    pointer-events: none;
    > :deep(*) { flex: none; }
    .name { flex: 1; }
}
// slot 내용도 기본은 안 줄어듦 — 줄어들 것(이름·경로)은 종류 컴포넌트가 `.TileFrame__head > .x`로 덮는다
.TileFrame__head {
    flex: 0 0 28px;
    display: flex;
    align-items: center;
    gap: vars.$spacing-sm;
    padding: 0 0 0 vars.$spacing-md;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    @include vars.font('body-medium');
    min-width: 0;
    > :deep(*) { flex: none; }
    // 터치 길게 누르기 = Ctrl+Alt 팝업(팝업-6) — 글자 선택·기본 메뉴 대신
    @media (pointer: coarse) {
        user-select: none;
        -webkit-user-select: none;
        -webkit-touch-callout: none;
    }
}
.pos {
    min-width: 16px;
    height: 16px;
    padding: 0 4px;
    border-radius: 4px;
    display: inline-grid;
    place-items: center;
    @include vars.font('label-small');
    font-weight: vars.$weight-lg;
    background: rgb(var(--v-theme-on-surface));
    color: rgb(var(--v-theme-surface));
}
.name {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-weight: vars.$weight-md;
}
.spacer { flex: 1; }
.perm {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    @include vars.font('body-small');
    color: rgb(var(--v-theme-share));
}
// ① 공유 색 — 왼쪽 띠 + 머리 줄 바탕 (테마 색 이름 share)
.TileFrame--shared {
    border-color: rgb(var(--v-theme-share));
    box-shadow: inset 3px 0 0 rgb(var(--v-theme-share));
    > .TileFrame__head { background: rgba(var(--v-theme-share), 0.12); }
}
// 버튼 자리 — 높이를 꽉 채우고 버튼 사이 간격 없음. kill은 창 묶음에서 떼어 RUN▾에(오클릭 방지)
.TileFrame__btns {
    align-self: stretch;
    display: flex;
    align-items: stretch;
}
</style>
