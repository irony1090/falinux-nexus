<script setup lang="ts">
import { tabReady } from '@/common/util/tabId.util';
import type { DirEntry } from '@/feature/worker/api/worker.api';
import { useWorkerRoster } from '@/feature/worker/api/worker.api';
import { useWorkerDir } from '@/feature/worker/hook/workerDir.hook';
import { entryIcon, isDirLike, linkDest, parentPathOf, sizeText, timeText } from '@/feature/worker/util/workerDir.util';
import { computed, ref, watch } from 'vue';
import { VBtn, VIcon, VProgressLinear } from 'vuetify/components';
import { useNodeOps } from '../../store/nodeOps.store';
import { useTileTree } from '../../store/tileTree.store';

// 장비 타일 본문 — 경로 줄 + 목록(스크롤 끝에서 다음 페이지). 폴더-2(장비 타일), REF-worker-dir-list.md
const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const LOAD_MORE_PX = 300    // 바닥까지 이만큼 남으면 다음 페이지

const { tiles, navigateDevice } = useTileTree();
const tile = computed(() => {
    const t = tiles.value[props.tileId];
    return t?.type === 'device' ? t : undefined;
})

const instance = computed(() => tile.value?.instance);
const dir = useWorkerDir(instance, computed(() => tile.value?.nodeId), computed(() => tile.value?.path ?? ''));
const { entries, crumbs, parent, total, loading, error } = dir;

// 접속 여부 = 명부 주기 조회. 다시 접속하면 다시 읽는다
const { devices } = useWorkerRoster();
const online = computed(() => {
    const key = instance.value;
    if (!key) return false;
    return !!devices.value.get(key.slice(0, key.lastIndexOf('#')))?.instances.find(i => i.instanceKey === key)?.online;
})
watch(online, (on, was) => {
    if (on && was === false && error.value) void dir.reload();
})

// 루트(`/`·`C:\`) 뒤에는 구분자를 안 붙인다(이름이 이미 구분자로 끝남)
const sepAfter = (name: string) => !/[\\/]$/.test(name);

const go = (path: string) => navigateDevice(props.tileId, path);

// 폴더-4(실행): 실행 가능한 파일(Q24(실행 대상), 판정 = worker)만 "실행"(Q23(누르는 방식)). cwd = 그 파일이 든 폴더
const { saving, busyTile, execFile } = useNodeOps();
const runFile = (e: DirEntry) => {
    const t = tile.value;
    if (t) void execFile(props.tileId, t.nodeId, t.instance, e.path, parentPathOf(e.path));
}
const onRow = (e: DirEntry) => {
    if (isDirLike(e)) go(e.path);   // 링크 폴더는 링크 경로 그대로 들어간다
}

const bodyEl = ref<HTMLElement>();
const onScroll = () => {
    const el = bodyEl.value;
    if (el && el.scrollTop + el.clientHeight >= el.scrollHeight - LOAD_MORE_PX) void dir.loadMore();
}
// 다른 폴더로 가면 맨 위부터
watch(() => tile.value?.path, () => bodyEl.value?.scrollTo({ top: 0 }));
</script>

<template>
<div ref="bodyEl" class="DeviceTileBody" @scroll.passive="onScroll">
    <div class="crumbs">
        <template v-for="(c, i) in crumbs" :key="c.path">
            <span v-if="i > 0 && sepAfter(crumbs[i - 1]!.name)" class="sep">/</span>
            <span v-if="i === crumbs.length - 1" class="cur">
                <v-icon v-if="!c.name" size="13" icon="mdi-harddisk" />{{ c.name || '드라이브' }}
            </span>
            <button v-else @click="go(c.path)">
                <v-icon v-if="!c.name" size="13" icon="mdi-harddisk" />{{ c.name || '드라이브' }}
            </button>
        </template>
        <span class="tools">
            <span class="dev" :class="{ on: online }" :title="online ? '접속 중' : '접속 끊김'"><i /></span>
            <button class="tool" title="새로고침" :disabled="loading" @click="dir.reload()">
                <v-icon size="14" icon="mdi-refresh" />
            </button>
        </span>
    </div>
    <div class="bar">
        <v-progress-linear v-if="loading || busyTile === tileId" indeterminate height="2" color="primary" />
    </div>

    <div v-if="error" class="state">
        <span>{{ error.status === 409 ? '인스턴스 접속이 끊겼습니다. 다시 접속하면 다시 읽습니다.' : (error.message || '목록을 불러오지 못했습니다') }}</span>
        <span class="acts">
            <v-btn v-if="parent !== null && error.status !== 409" size="x-small" variant="tonal" @click="go(parent)">상위로</v-btn>
            <v-btn size="x-small" variant="tonal" :disabled="loading" @click="dir.reload()">다시 시도</v-btn>
        </span>
    </div>
    <template v-else>
        <div v-for="e in entries" :key="e.path" class="row" :class="{ dir: isDirLike(e) }" @click="onRow(e)">
            <v-icon size="14" :icon="entryIcon(e)" :class="isDirLike(e) ? 'folder' : 'file'" />
            <span class="nm" :title="e.path">{{ e.name }}</span>
            <span class="tail">
                <span v-if="e.kind === 'link'" class="link" :title="`링크: ${e.linkTarget ?? ''}${e.targetPath ? `\n원본: ${e.targetPath}` : '\n원본 없음(끊긴 링크)'}`">→ {{ e.targetPath ?? e.linkTarget }}</span>
                <button v-if="linkDest(e)" class="to" title="원본 위치로" @click.stop="go(linkDest(e)!)">
                    <v-icon size="13" icon="mdi-arrow-top-right" />
                </button>
                <v-btn v-if="e.exec" class="run" size="x-small" variant="tonal" color="primary" :disabled="saving || !tabReady || !online" title="새 터미널 타일에서 실행" @click.stop="runFile(e)">실행</v-btn>
                <span v-if="e.kind === 'file'" class="size">{{ sizeText(e.size) }}</span>
                <span class="time">{{ timeText(e.modTime) }}</span>
            </span>
        </div>
        <p v-if="!loading && !entries.length" class="empty">비어 있음</p>
        <p v-if="total > entries.length" class="more">
            {{ total }}개 중 {{ entries.length }}개
            <v-btn size="x-small" variant="text" :disabled="loading" @click="dir.loadMore()">더 보기</v-btn>
        </p>
    </template>
</div>
</template>

<style scoped lang="scss">
.DeviceTileBody {
    flex: 1;
    min-height: 0;
    overflow: auto;
    font-size: 12px;
    button {
        padding: 0;
        border: 0;
        background: none;
        color: inherit;
        font: inherit;
        cursor: pointer;
    }
}
.crumbs {
    position: sticky;
    top: 0;
    z-index: 1;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 2px;
    padding: 5px vars.$spacing-md;
    font-size: 11px;
    background: rgb(var(--v-theme-surface));
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    button { display: inline-flex; align-items: center; gap: 2px; color: rgb(var(--v-theme-primary)); }
    .cur { display: inline-flex; align-items: center; gap: 2px; font-weight: vars.$weight-md; }
    .sep { opacity: 0.5; }
    .tools {
        margin-left: auto;
        display: flex;
        align-items: center;
        gap: vars.$spacing-md;
    }
    .tool {
        opacity: 0.7;
        &:disabled { cursor: default; opacity: 0.3; }
    }
}
.bar { height: 2px; }
.state {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: vars.$spacing-md;
    margin: vars.$spacing-md;
    color: rgb(var(--v-theme-error));
    .acts { display: inline-flex; gap: vars.$spacing-sm; }
}
.row {
    display: grid;
    grid-template-columns: 16px minmax(0, 1fr) auto;
    align-items: center;
    gap: 6px;
    min-height: 26px;
    padding: 2px vars.$spacing-md;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    &.dir { cursor: pointer; }
    &.dir:hover .nm { color: rgb(var(--v-theme-primary)); }
}
.folder { color: rgb(var(--v-theme-warning)); }
.file { opacity: 0.7; }
.nm {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.tail {
    display: flex;
    align-items: center;
    gap: vars.$spacing-sm;
    font-size: 10.5px;
    white-space: nowrap;
    .link {
        max-width: 160px;
        overflow: hidden;
        text-overflow: ellipsis;
        color: rgb(var(--v-theme-info));
    }
    .to { display: inline-flex; color: rgb(var(--v-theme-info)); }
    .size { min-width: 44px; text-align: right; font-family: monospace; opacity: 0.75; }
    .time { font-family: monospace; opacity: 0.6; }
}
// 접속 여부 점 — 폴더 타일 장비 칩과 같은 색
.dev {
    display: inline-flex;
    i {
        width: 6px;
        height: 6px;
        border-radius: 50%;
        background: rgba(var(--v-theme-on-surface), var(--v-disabled-opacity));
    }
    &.on i { background: rgb(var(--v-theme-success)); }
}
.empty, .more { margin: vars.$spacing-md; opacity: 0.6; }
.more { display: flex; align-items: center; gap: vars.$spacing-sm; }
</style>
