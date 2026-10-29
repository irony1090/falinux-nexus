<script setup lang="ts">
import type { CatchErrorResponse } from '@/common/api/api.util';
import { readUserPref, writeUserPref } from '@/common/util/userPrefs.util';
import { useResizeCallback } from '@/feature/common/store/resizeGroup.store';
import { instanceOnline, probeWorkerFile, useWorkerRoster, viewFileUrl } from '@/feature/worker/api/worker.api';
import { viewKindOf } from '@/feature/worker/util/fileKind.util';
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { VIcon, VProgressCircular } from 'vuetify/components';
import { useTileTree } from '../../store/tileTree.store';

// 보기 타일 본문 — 이미지 + V4(확대 조작: 휠·핀치 확대, 끌어 이동, 더블클릭 맞춤). 파트(보기 타일), REF-file-view.md
const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const ZOOM_MIN = 0.05
const ZOOM_MAX = 20
const WHEEL_RATE = 0.0015      // deltaY 1당 확대 배율 지수
const DOUBLE_TAP_MS = 300
const TAP_MOVE_PX = 10          // 이보다 움직이면 탭이 아님
const VIEWS_KEY = 'viewer.views'
const VIEWS_MAX = 50            // 저장하는 타일 수 — 닫힌 타일 것은 오래된 순으로 밀려남

const { tiles } = useTileTree();
const tile = computed(() => {
    const t = tiles.value[props.tileId];
    return t?.type === 'viewer' ? t : undefined;
})
const kind = computed(() => tile.value ? viewKindOf(tile.value.path) : undefined);

// 장비가 끊겨도 요청 — 서버가 마지막 사본을 줌(없으면 409 = 끊김 안내). 다시 접속하면 다시 불러옴(DEVICE:CHANGE -> 명부 재조회)
const { devices } = useWorkerRoster();
const online = computed(() => !!tile.value && instanceOnline(devices.value, tile.value.instance));
const src = computed(() => tile.value ? viewFileUrl(tile.value.instance, tile.value.nodeId, tile.value.path) : undefined);

type Status = 'loading' | 'ok' | 'error';
const status = ref<Status>('loading');
const errorText = ref('');
const loadKey = ref(0);     // 바꾸면 <img>를 다시 만들어 다시 요청
const reload = () => {
    status.value = 'loading';
    loadKey.value++;
}
watch(online, (on, was) => {
    if (on && was === false) reload();
});

// <img>는 실패 원인을 알려 주지 않아 probe로 다시 물어 서버 메시지(파일 없음·권한 등)를 보인다. 끊김이면 사본이 없는 것
const onError = () => {
    status.value = 'error';
    errorText.value = '';
    const t = tile.value;
    if (!t || !online.value) return;
    probeWorkerFile(t.instance, t.nodeId, t.path)
    .then(() => errorText.value = '표시할 수 없는 이미지입니다')
    .catch((e: CatchErrorResponse) => errorText.value = e.message || '불러오지 못했습니다');
}

// ---- 확대·이동 — 화면 좌표(상자 기준 px). s = 배율, x·y = 이미지 왼쪽 위 ----
type View = { s: number; x: number; y: number };
const boxEl = ref<HTMLElement>();
const box = ref({ w: 0, h: 0 });
const nat = ref({ w: 0, h: 0 });
const measure = () => {
    const el = boxEl.value;
    if (el) box.value = { w: el.clientWidth, h: el.clientHeight };
}
useResizeCallback(boxEl, measure);

const onLoad = (e: Event) => {
    const img = e.target as HTMLImageElement;
    // 크기 없는 SVG는 natural 크기가 0 — 상자 크기로
    nat.value = { w: img.naturalWidth || box.value.w, h: img.naturalHeight || box.value.h };
    status.value = 'ok';
    measure();
}

// 맞춤 = 상자 안에 다 보이게, 작은 이미지는 키우지 않음
const fit = computed<View>(() => {
    const { w, h } = nat.value, b = box.value;
    if (!w || !h || !b.w || !b.h) return { s: 1, x: 0, y: 0 };
    const s = Math.min(1, b.w / w, b.h / h);
    return { s, x: (b.w - w * s) / 2, y: (b.h - h * s) / 2 };
});

// 저장 = 이 기기(localStorage)만, null = 맞춤(상자 크기가 바뀌면 다시 맞춤)
const loadViews = () => readUserPref<Record<string, View>>(VIEWS_KEY, {});
const view = ref<View | null>(loadViews()[props.tileId] ?? null);
let saveTimer: ReturnType<typeof setTimeout> | undefined;
const save = () => {
    clearTimeout(saveTimer);
    saveTimer = setTimeout(() => {
        const { [props.tileId]: _, ...rest } = loadViews();
        const next = view.value ? { ...rest, [props.tileId]: view.value } : rest;
        writeUserPref(VIEWS_KEY, Object.fromEntries(Object.entries(next).slice(-VIEWS_MAX)));
    }, 300);
}
onBeforeUnmount(() => clearTimeout(saveTimer));

const cur = computed(() => view.value ?? fit.value);
const setView = (v: View | null) => {
    view.value = v;
    save();
}
const toFit = () => setView(null);
const toActual = () => {
    const b = box.value, { w, h } = nat.value;
    setView({ s: 1, x: (b.w - w) / 2, y: (b.h - h) / 2 });
}
const zoomAt = (cx: number, cy: number, factor: number) => {
    const c = cur.value;
    const s = Math.min(ZOOM_MAX, Math.max(ZOOM_MIN, c.s * factor));
    const k = s / c.s;
    setView({ s, x: cx - (cx - c.x) * k, y: cy - (cy - c.y) * k });
}
const panBy = (dx: number, dy: number) => {
    const c = cur.value;
    setView({ s: c.s, x: c.x + dx, y: c.y + dy });
}

const local = (e: { clientX: number; clientY: number }) => {
    const r = boxEl.value!.getBoundingClientRect();
    return { x: e.clientX - r.left, y: e.clientY - r.top };
}

const onWheel = (e: WheelEvent) => {
    if (status.value !== 'ok') return;
    const p = local(e);
    zoomAt(p.x, p.y, Math.exp(-e.deltaY * WHEEL_RATE));
}

// 포인터 1개 = 끌어 이동, 2개 = 핀치(두 점 거리 비율로 확대 + 가운데 점 이동만큼 이동)
const pts = new Map<number, { x: number; y: number }>();
let tap = { at: 0, x: 0, y: 0, moved: false };
const onDown = (e: PointerEvent) => {
    if (status.value !== 'ok') return;
    boxEl.value?.setPointerCapture(e.pointerId);
    const p = local(e);
    pts.set(e.pointerId, p);
    if (pts.size === 1) tap = { ...tap, x: p.x, y: p.y, moved: false };
    else tap.moved = true;
}
const onMove = (e: PointerEvent) => {
    const prev = pts.get(e.pointerId);
    if (!prev) return;
    const p = local(e);
    if (Math.hypot(p.x - tap.x, p.y - tap.y) > TAP_MOVE_PX) tap.moved = true;
    if (pts.size === 1) {
        pts.set(e.pointerId, p);
        panBy(p.x - prev.x, p.y - prev.y);
        return;
    }
    const [a, b] = [...pts.values()];
    const other = a === prev ? b! : a!;
    pts.set(e.pointerId, p);
    const d0 = Math.hypot(prev.x - other.x, prev.y - other.y);
    const d1 = Math.hypot(p.x - other.x, p.y - other.y);
    const m0 = { x: (prev.x + other.x) / 2, y: (prev.y + other.y) / 2 };
    const m1 = { x: (p.x + other.x) / 2, y: (p.y + other.y) / 2 };
    if (d0 > 0) zoomAt(m0.x, m0.y, d1 / d0);
    panBy(m1.x - m0.x, m1.y - m0.y);
}
// 터치 두 번 탭 = 맞춤(마우스는 dblclick)
const onUp = (e: PointerEvent) => {
    if (!pts.delete(e.pointerId) || pts.size || e.pointerType !== 'touch' || tap.moved) return;
    const now = Date.now();
    if (now - tap.at < DOUBLE_TAP_MS) {
        toFit();
        tap.at = 0;
    } else tap.at = now;
}
const onCancel = (e: PointerEvent) => { pts.delete(e.pointerId); };

const imgStyle = computed(() => ({
    width: `${nat.value.w}px`,
    height: `${nat.value.h}px`,
    transform: `translate(${cur.value.x}px, ${cur.value.y}px) scale(${cur.value.s})`,
}));
const percent = computed(() => `${Math.round(cur.value.s * 100)}%`);
</script>

<template>
<div class="ViewerTileBody">
    <div v-if="kind !== 'image'" class="msg">
        <v-icon size="28" icon="mdi-file-question-outline" />
        <span>이 형식은 아직 볼 수 없습니다</span>
    </div>
    <template v-else>
        <!-- 몸통 안 펼치기 = 이미지 확대(컨트롤 패널 아님, 제스처-1(펼치기 패널)) -->
        <div ref="boxEl" class="box" :class="{ ready: status === 'ok' }" data-pager-pinch-off
            @wheel.prevent="onWheel"
            @pointerdown="onDown" @pointermove="onMove" @pointerup="onUp" @pointercancel="onCancel"
            @dblclick="toFit"
        >
            <img v-if="src" :key="loadKey" :src="src" alt="" draggable="false" :style="imgStyle" @load="onLoad" @error="onError">
        </div>
        <div v-if="status === 'loading'" class="msg msg--over">
            <v-progress-circular indeterminate size="24" width="2" />
        </div>
        <div v-else-if="status === 'error' && !online" class="msg msg--over">
            <v-icon size="28" icon="mdi-lan-disconnect" />
            <span>장비가 접속해 있지 않습니다</span>
            <span class="sub">다시 접속하면 불러옵니다</span>
        </div>
        <div v-else-if="status === 'error'" class="msg msg--over">
            <v-icon size="28" color="warning" icon="mdi-image-broken-variant" />
            <span>{{ errorText || '불러오지 못했습니다' }}</span>
            <span class="ct-grp ct-grp--sm">
                <button type="button" class="ct-seg" @click="reload">
                    <v-icon size="14" icon="mdi-refresh" />&nbsp;다시
                </button>
            </span>
        </div>
        <span v-if="status === 'ok' && !online" class="stale ct-grp ct-grp--sm" title="장비가 접속해 있지 않아 supervisor에 남은 사본을 보여 줍니다">
            <span class="ct-seg ct-seg--warning"><v-icon size="13" icon="mdi-lan-disconnect" />&nbsp;오프라인 · 마지막 사본</span>
        </span>
        <span v-if="status === 'ok'" class="zoom ct-grp ct-grp--sm">
            <span class="ct-seg ct-seg--txt">{{ percent }}</span>
            <button type="button" class="ct-seg" title="맞춤(더블클릭)" @click="toFit">
                <v-icon size="14" icon="mdi-fit-to-screen-outline" />
            </button>
            <button type="button" class="ct-seg" title="실제 크기" @click="toActual">1:1</button>
        </span>
    </template>
</div>
</template>

<style scoped lang="scss">
.ViewerTileBody {
    flex: 1;
    min-height: 0;
    position: relative;
    overflow: hidden;
}
.box {
    position: absolute;
    inset: 0;
    overflow: hidden;
    touch-action: none;     // 핀치·끌기를 브라우저 확대·스크롤 대신 여기서
    cursor: grab;
    &:active { cursor: grabbing; }
    > img {
        position: absolute;
        left: 0;
        top: 0;
        max-width: none;
        transform-origin: 0 0;
        user-select: none;
        visibility: hidden;     // 크기를 알기 전(맞춤 계산 전) 깜빡임 방지
    }
    &.ready > img { visibility: visible; }
}
.msg {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 16px;
    text-align: center;
    @include vars.font('body-medium');
    .sub { opacity: 0.6; @include vars.font('body-small'); }
}
.msg--over { pointer-events: none; > .ct-grp { pointer-events: auto; } }
.stale {
    position: absolute;
    left: 8px;
    top: 8px;
    pointer-events: none;
    background: rgb(var(--v-theme-surface));
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.2);
}
.zoom {
    position: absolute;
    right: 8px;
    bottom: 8px;
    background: rgb(var(--v-theme-surface));
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.2);
}
</style>
