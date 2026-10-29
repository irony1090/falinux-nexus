<script setup lang="ts">
import type { CatchErrorResponse } from '@/common/api/api.util';
import { instanceOnline, probeWorkerFile, useWorkerRoster, viewFileUrl } from '@/feature/worker/api/worker.api';
import { useTransfers } from '@/feature/worker/store/transfers.store';
import { parentPathOf } from '@/feature/worker/util/workerDir.util';
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { VIcon, VProgressCircular } from 'vuetify/components';
import { getVideoPosition, saveVideoPosition } from '../../api/tile.api';
import { useTileTree } from '../../store/tileTree.store';
import { fileLabel } from '../../util/tileLabel.util';

// 영상 보기 본문 — 사본 없이 Range(서버 serveVideo) + S1(이어보기). 파트(영상 Range), REF-video-view.md
const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const SAVE_EVERY_MS = 10_000    // 재생 중 위치 저장 간격
const END_MARGIN = 1            // 끝에서 이만큼(초) 안의 저장 위치는 처음부터

const { tiles } = useTileTree();
const tile = computed(() => {
    const t = tiles.value[props.tileId];
    return t?.type === 'viewer' ? t : undefined;
})

const { devices } = useWorkerRoster();
const online = computed(() => !!tile.value && instanceOnline(devices.value, tile.value.instance));
const src = computed(() => tile.value ? viewFileUrl(tile.value.instance, tile.value.nodeId, tile.value.path) : undefined);

type Status = 'loading' | 'ok' | 'offline' | 'unplayable' | 'error';
const status = ref<Status>('loading');
const errorText = ref('');
const videoEl = ref<HTMLVideoElement>();
const loadKey = ref(0);     // 바꾸면 <video>를 다시 만들어 다시 요청

// Q3(오프라인): 끊기기 직전 자리 — 다시 불러온 뒤 그 자리로(재생 중이었으면 이어서)
let resume: { t: number; play: boolean } | null = null;
const keepSpot = () => {
    const el = videoEl.value;
    if (!resume && el && el.readyState > 0) resume = { t: el.currentTime, play: !el.paused };
}
const reload = () => {
    keepSpot();
    status.value = 'loading';
    loadKey.value++;
}
watch(online, (on, was) => {
    if (!on && was) {
        savePos();
        keepSpot();
        videoEl.value?.pause();
        status.value = 'offline';
    } else if (on && was === false && (status.value === 'offline' || status.value === 'error')) reload();   // 첫 명부 도착(false -> true)은 다시 접속이 아님
});

// S1(이어보기): 처음 불러올 때 서버에 저장된 자리 — 메타데이터보다 늦게 와도 그때 옮긴다
let saved: number | null = null;
let savedReady: Promise<unknown> = Promise.resolve();
onMounted(() => {
    const t = tile.value;
    if (t) savedReady = getVideoPosition(t.instance, t.path).then(r => saved = r.pos).catch(() => {});
});

const onMeta = async (e: Event) => {
    const el = e.target as HTMLVideoElement;
    status.value = 'ok';
    lastAt = Date.now();
    if (resume) {
        el.currentTime = resume.t;
        if (resume.play) void el.play().catch(() => {});
        resume = null;
        return;
    }
    await savedReady;
    if (el !== videoEl.value) return;   // 기다리는 사이 <video>를 다시 만듦 — 새 요소의 onMeta가 옮긴다
    if (saved !== null && saved < el.duration - END_MARGIN && el.currentTime === 0) el.currentTime = saved;
    saved = null;
}

// 저장 = 정지·탐색·끝·재생 중 10초마다·창 닫기·타일 사라짐. 끝까지 봤으면 0(다음엔 처음부터)
let lastPos = -1;
let lastAt = 0;
const savePos = (keepalive = false) => {
    const el = videoEl.value, t = tile.value;
    if (!el || !t || status.value !== 'ok' || el.readyState === 0) return;
    const pos = el.ended ? 0 : el.currentTime;
    lastAt = Date.now();
    if (Math.abs(pos - lastPos) < 0.5) return;
    lastPos = pos;
    void saveVideoPosition(t.instance, t.path, pos, keepalive).catch(() => {});
}
const onTime = () => {
    if (Date.now() - lastAt >= SAVE_EVERY_MS) savePos();
}
const onHide = () => {
    if (document.visibilityState === 'hidden') savePos(true);
}
const onPageHide = () => savePos(true);
onMounted(() => {
    document.addEventListener('visibilitychange', onHide);
    window.addEventListener('pagehide', onPageHide);
});
// 닫기로 사라질 때도 보냄 — 닫는 PUT보다 먼저 닿으면 그 PUT이 지우고, 늦으면 서버가 버림(P1(위치 지우는 시점))
onBeforeUnmount(() => {
    document.removeEventListener('visibilitychange', onHide);
    window.removeEventListener('pagehide', onPageHide);
    savePos(true);
});

// <video>는 HTTP 오류와 코덱 오류를 같은 코드로 알려 probe로 다시 묻는다 — 서버가 받아 주면 브라우저가 못 푸는 것(V5(재생 불가))
const onError = () => {
    const el = videoEl.value, t = tile.value;
    keepSpot();
    if (!online.value || !t) {
        status.value = 'offline';
        return;
    }
    status.value = 'error';
    errorText.value = '';
    const code = el?.error?.code;
    if (code === MediaError.MEDIA_ERR_NETWORK) {
        errorText.value = '불러오는 중 끊겼습니다';
        return;
    }
    probeWorkerFile(t.instance, t.nodeId, t.path)
    .then(() => status.value = 'unplayable')
    .catch((e: CatchErrorResponse) => errorText.value = e.message || '불러오지 못했습니다');
}

// V5(재생 불가) = 내려받기. 진행 막대는 푸터 전송 목록에(이 타일엔 막대 자리 없음)
const { download } = useTransfers();
const onDownload = () => {
    const t = tile.value;
    if (!t) return;
    void download(
        { tileId: props.tileId, instance: t.instance, folderId: t.nodeId, dir: parentPathOf(t.path) },
        [{ name: fileLabel(t.path), path: t.path, kind: 'file', size: 0, modTime: 0 }],
    );
}
</script>

<template>
<div class="VideoViewBody">
    <video v-if="src" ref="videoEl" :key="loadKey" :src="src" controls preload="metadata" playsinline
        @loadedmetadata="onMeta" @error="onError"
        @pause="savePos()" @seeked="savePos()" @ended="savePos()" @timeupdate="onTime"
    />
    <div v-if="status === 'loading'" class="msg msg--over">
        <v-progress-circular indeterminate size="24" width="2" />
    </div>
    <div v-else-if="status === 'offline'" class="msg msg--cover">
        <v-icon size="28" icon="mdi-lan-disconnect" />
        <span>장비가 접속해 있지 않습니다</span>
        <span class="sub">다시 접속하면 보던 자리부터 불러옵니다</span>
    </div>
    <div v-else-if="status === 'unplayable'" class="msg msg--cover">
        <v-icon size="28" color="warning" icon="mdi-movie-off-outline" />
        <span>이 브라우저에서 재생할 수 없는 영상입니다</span>
        <span class="sub">코덱(HEVC 등)을 지원하지 않습니다</span>
        <span class="ct-grp ct-grp--sm">
            <button type="button" class="ct-seg" :disabled="!online" @click="onDownload">
                <v-icon size="14" icon="mdi-download" />&nbsp;내려받기
            </button>
        </span>
    </div>
    <div v-else-if="status === 'error'" class="msg msg--cover">
        <v-icon size="28" color="warning" icon="mdi-movie-off-outline" />
        <span>{{ errorText || '불러오지 못했습니다' }}</span>
        <span class="ct-grp ct-grp--sm">
            <button type="button" class="ct-seg" @click="reload">
                <v-icon size="14" icon="mdi-refresh" />&nbsp;다시
            </button>
        </span>
    </div>
</div>
</template>

<style scoped lang="scss">
.VideoViewBody {
    flex: 1;
    min-height: 0;
    position: relative;
    overflow: hidden;
    background: rgb(var(--v-theme-terminal));     // 두 테마 다 어두운 색 — 영상 여백
    > video {
        position: absolute;
        inset: 0;
        width: 100%;
        height: 100%;
        object-fit: contain;
    }
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
.msg--over { pointer-events: none; color: rgb(var(--v-theme-on-terminal)); }
.msg--cover { background: rgb(var(--v-theme-surface)); }
</style>
