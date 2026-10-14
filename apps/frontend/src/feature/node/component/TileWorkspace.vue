<script setup lang="ts">
import { useVirtualKeyboard } from '@/common/hook/virtualKeyboard.hook';
import TerminalKeyBar, { KEY_BAR_HEIGHT } from '@/feature/process/component/TerminalKeyBar.vue';
import { useProcessTerm } from '@/feature/process/store/processTerm.store';
import { computed, nextTick, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { provideNodeOps } from '../store/nodeOps.store';
import { provideTileGrids } from '../store/tileGrids.store';
import { useTileTree } from '../store/tileTree.store';
import NodeOpsHost from './NodeOpsHost.vue';
import NodeSelectionBar from './NodeSelectionBar.vue';
import OverviewButton from './OverviewButton.vue';
import TileNav from './TileNav.vue';
import TileStrip from './TileStrip.vue';

// strip과 헤더 내비가 같은 Grid 상태를 쓰도록 공통 부모에서 provide. 내비는 Teleport로 AppHead 자리에 — F(헤더 전달 방식)
const grids = provideTileGrids();
const { activate, rootTab } = grids;
provideNodeOps(grids);     // 같은 컴포넌트에선 inject가 안 되므로 grids를 넘긴다
const { tiles, setMinimized } = useTileTree();
const route = useRoute();
const router = useRouter();

// ?focus=<tileId> = 그 타일을 active로(목록 "보드로 이동"·위치 알림, N3(누르면 이동)) / ?tab=terminals = 루트 터미널 탭(M4(위치 없음)).
// 타일이 트리에 들어올 때까지 기다렸다가 처리하고 query를 지운다
watch(() => {
    const focus = route.query.focus;
    return [focus, route.query.tab, typeof focus === 'string' ? tiles.value[focus] : undefined] as const;
}, ([focus, tab, tile]) => {
    if (typeof focus === 'string') {
        if (!tile) return;
        if (tile.type === 'terminal' && tile.minimized) setMinimized(tile.id, false);
        void activate(tile.id);
    } else if (tab === 'terminals') {
        rootTab.value = 'terminals';
    } else {
        return;
    }
    void router.replace({ query: {} });
}, { immediate: true })

// 키보드-2(가려짐 이동): 포커스된 터미널이 가상 키보드·키 줄에 가려지면 타일 영역 전체를 올린다(배치·PTY 크기 불변)
// 고침-나(커서 기준 이동): 타일이 화면보다 커도 커서 줄은 키 줄 위에 오게, 이때 타일 머리가 헤더 밑으로 들어가도 된다 — REF-mobile-keyboard.md
const kb = useVirtualKeyboard();
const { handleOf } = useProcessTerm();
const root = ref<HTMLElement>();
const shift = ref(0);
let shiftUid: string | null = null;     // 같은 터미널에서 키보드가 열려 있는 동안은 이전 이동을 기준으로 삼는다(htop처럼 커서가 뛰어도 화면이 덜 흔들리게)

// 커서 줄의 위·아래(화면 좌표). 스크롤로 커서가 화면 밖이면 undefined
const cursorRect = (uid: string) => {
    const term = handleOf(uid)?.term;
    const screen = term?.element?.querySelector('.xterm-screen')?.getBoundingClientRect();
    if (!term || !screen || term.rows <= 0) return undefined;
    const buf = term.buffer.active;
    const row = buf.baseY + buf.cursorY - buf.viewportY;
    if (row < 0 || row >= term.rows) return undefined;
    const h = screen.height / term.rows;
    return { top: screen.top + row * h, bottom: screen.top + (row + 1) * h };
}

const measure = () => {
    const uid = kb.open ? kb.termUid : null;
    const tile = uid ? root.value?.querySelector(`[data-term-uid="${CSS.escape(uid)}"]`)?.closest('.TileFrame') : undefined;
    if (!uid || !tile || !root.value) {
        shift.value = 0;
        shiftUid = null;
        return;
    }
    // 올리기 전 좌표로 계산한다. transition 중이면 화면 좌표에 중간값이 섞이므로 계산된 transform에서 지금 이동량을 읽는다
    const t = getComputedStyle(root.value).transform;
    const live = t && t !== 'none' ? -new DOMMatrix(t).m42 : 0;
    const r = tile.getBoundingClientRect();
    const top = r.top + live, bottom = r.bottom + live;
    const limit = kb.bottom - KEY_BAR_HEIGHT;
    const minTop = Math.max(kb.top, root.value.getBoundingClientRect().top + live);   // 헤더 아래·보이는 영역 위 끝
    let s = Math.max(uid === shiftUid ? shift.value : 0, Math.min(bottom - limit, top - minTop));
    const c = cursorRect(uid);
    if (c) {
        s = Math.max(s, c.bottom + live - limit);
        s = Math.min(s, c.top + live - minTop);
    }
    shift.value = Math.max(0, Math.min(s, bottom - limit));
    shiftUid = uid;
}
watch(() => [kb.open, kb.top, kb.bottom, kb.termUid], () => nextTick(measure));

// 라-2(높이 고정): resizes-content라 키보드가 뜨면 화면이 줄어든다. 그대로 두면 Grid 수용량이 바뀌어 타일이 옮겨 가고
// 크기 소유 탭은 PTY 크기까지 바뀌므로, 열린 동안은 키보드 전 높이로 고정하고 가려지는 부분은 위 이동으로 처리한다
const bodyStyle = computed(() => ({
    height: kb.open ? `calc(${kb.baseH}px - var(--v-layout-top, 0px))` : undefined,
    transform: shift.value ? `translateY(${-shift.value}px)` : undefined,
}));

// 키보드가 열린 동안 커서 이동·스크롤·크기 변경을 따라간다(그리기 중 커서가 뛰는 것은 묶어서 한 번)
let measureTimer: ReturnType<typeof setTimeout> | undefined;
const measureSoon = () => {
    clearTimeout(measureTimer);
    measureTimer = setTimeout(measure, 100);
}
watch(() => kb.open ? kb.termUid : null, (uid, _, onCleanup) => {
    const term = uid ? handleOf(uid)?.term : undefined;
    if (!term) return;
    const subs = [term.onCursorMove(measureSoon), term.onScroll(measureSoon), term.onResize(measureSoon)];
    onCleanup(() => {
        subs.forEach(d => d.dispose());
        clearTimeout(measureTimer);
    });
});
</script>

<template>
<div class="TileWorkspace">
<div ref="root" class="TileWorkspace__body" :style="bodyStyle">
    <node-selection-bar />
    <tile-strip />
    <overview-button />
    <node-ops-host />
</div>
</div>
<!-- 올라가는 영역 밖에 둔다(transform 안이면 fixed 기준이 바뀜) -->
<terminal-key-bar />
<teleport defer to="#app-head-nav">
    <tile-nav />
</teleport>
</template>

<style scoped lang="scss">
// 높이 = 화면 - 헤더. 키보드가 열려 화면이 줄면 body는 고정 높이로 넘치므로 잘라 낸다(clip = 스크립트·포커스로도 스크롤 안 됨)
.TileWorkspace {
    height: calc(100dvh - var(--v-layout-top, 0px));
    overflow: clip;
}
// 고정 버튼 기준 — strip이 스크롤돼도 제자리. 선택 막대가 뜨면 strip이 그만큼 줄어든다
.TileWorkspace__body {
    position: relative;
    height: 100%;
    display: flex;
    flex-direction: column;
    transition: transform 0.15s;
}
</style>
