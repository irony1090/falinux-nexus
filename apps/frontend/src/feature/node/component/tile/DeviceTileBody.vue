<script setup lang="ts">
import type { CatchErrorResponse } from '@/common/api/api.util';
import { tabReady } from '@/common/util/tabId.util';
import type { DirEntry } from '@/feature/worker/api/worker.api';
import { suggestWorkerDir, useWorkerRoster } from '@/feature/worker/api/worker.api';
import { useWorkerDir } from '@/feature/worker/hook/workerDir.hook';
import TransferRow from '@/feature/worker/component/TransferRow.vue';
import { useTransfers, type UploadSource } from '@/feature/worker/store/transfers.store';
import { useAppDialog } from '@/feature/layout/store/appDialog.store';
import { entryIcon, isDirLike, isFileLike, linkDest, parentPathOf, sizeText, timeText } from '@/feature/worker/util/workerDir.util';
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { VBtn, VIcon, VList, VListItem, VMenu, VProgressLinear } from 'vuetify/components';
import { useItemPress } from '../../hook/itemPress.hook';
import { usePathEditKey } from '../../hook/pathEdit.hook';
import { useNodeOps } from '../../store/nodeOps.store';
import { useTileTree } from '../../store/tileTree.store';
import DirGridCell from './DirGridCell.vue';
import NodeRowMenu, { type RowMenuItem } from './NodeRowMenu.vue';
import PathInput, { type PathCand } from './PathInput.vue';

// 장비 타일 본문 — 경로 줄 + 목록(스크롤 끝에서 다음 페이지). 폴더-2(장비 타일), REF-worker-dir-list.md
const props = defineProps({
    tileId: {
        type: String,
        required: true,
    },
})

const LOAD_MORE_PX = 300    // 바닥까지 이만큼 남으면 다음 페이지
const DRAG_MIME = 'application/x-nexus-device-files'   // 장비 타일 항목 끌기(OS 파일 끌어 놓기와 구분) — 전송-12

const { tiles, navigateDevice } = useTileTree();
const tile = computed(() => {
    const t = tiles.value[props.tileId];
    return t?.type === 'device' ? t : undefined;
})

const instance = computed(() => tile.value?.instance);
const dir = useWorkerDir(instance, computed(() => tile.value?.nodeId), computed(() => tile.value?.path ?? ''));
const { entries, cur, crumbs, parent, total, loading, error } = dir;

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

// 경로 직접 입력 — 경로-3(화면: 입력칸)·4(화면: 자동완성), REF-path-input.md. 경로 줄 빈 곳 클릭(Q39) 또는 Ctrl+L
const SUGGEST_MAX = 50
const editing = ref(false);
const openEdit = () => { if (tile.value) editing.value = true; };
usePathEditKey(() => props.tileId, openEdit);
const onCrumbsClick = (e: MouseEvent) => {
    if (!editing.value && !(e.target as Element).closest('button, input, .tools')) openEdit();
}
// Windows = 드라이브 경로이거나 드라이브 목록(""). `\`는 Windows에서만 구분자(Linux에선 이름 글자)
const here = computed(() => cur.value ?? tile.value?.path ?? '');    // 읽기 실패면 요청한 경로
const win = computed(() => cur.value === '' || /^[A-Za-z]:[\\/]/.test(here.value));
const sepEnd = (p: string) => (win.value ? /[\\/]$/ : /\/$/).test(p) ? p : p + (win.value ? '\\' : '/');
const splitPath = (t: string) => {
    const i = win.value ? Math.max(t.lastIndexOf('/'), t.lastIndexOf('\\')) : t.lastIndexOf('/');
    return i < 0 ? { dir: '', prefix: t } : { dir: t.slice(0, i + 1), prefix: t.slice(i + 1) };
}
const suggestPath = async (t: string) => {
    const d = tile.value;
    if (!d) return { items: [] };
    const { dir: at, prefix } = splitPath(t);
    const r = await suggestWorkerDir(d.instance, d.nodeId, { dir: at, prefix, max: SUGGEST_MAX });
    return { items: r.items.map<PathCand>(i => ({ key: i.path, label: i.name, fill: sepEnd(i.path) })), more: r.more };
}
// Enter: worker가 폴더인지 보고 정리한 경로로. 옛 worker(DIR_SUGGEST 모름 = 502)면 확인 없이 이동
const submitPath = async (t: string) => {
    const d = tile.value;
    if (!d) return;
    const target = t.trim() || (win.value ? '' : '/');
    try {
        go((await suggestWorkerDir(d.instance, d.nodeId, { dir: target })).dir);
    } catch (e) {
        const err = e as CatchErrorResponse;
        if (err.status !== 502) return err.message || '이동할 수 없습니다';
        go(target);
    }
    return undefined;
}

// 폴더-4(실행): 실행 가능한 파일(Q24(실행 대상), 판정 = worker)만 "실행"(Q23(누르는 방식)). cwd = 그 파일이 든 폴더
const { saving, busyTile, execFile } = useNodeOps();
const runFile = (e: DirEntry) => {
    const t = tile.value;
    if (t) void execFile(props.tileId, t.nodeId, t.instance, e.path, parentPathOf(e.path));
}

const bodyEl = ref<HTMLElement>();
const onScroll = () => {
    const el = bodyEl.value;
    if (el && el.scrollTop + el.clientHeight >= el.scrollHeight - LOAD_MORE_PX) void dir.loadMore();
}
// 다른 폴더로 가면 맨 위부터, 선택·고름은 비움
watch(() => tile.value?.path, () => {
    bodyEl.value?.scrollTo({ top: 0 });
    selected.value = new Set();
    focused.value = null;
});

// 항목 선택 — 폴더 타일과 같은 방식(Ctrl+클릭·길게 누르기, 하나 이상 = 선택 모드). REF-dir-grid.md "누르기 규칙"
const selected = ref(new Set<string>());
const selEntries = computed(() => entries.value.filter(e => selected.value.has(e.path)));
const toggle = (e: DirEntry) => {
    const next = new Set(selected.value);
    if (!next.delete(e.path)) next.add(e.path);
    selected.value = next;
}
const clearSel = () => { selected.value = new Set(); };

// 칸 누르기 — 더블 = 디렉토리(링크 폴더는 링크 경로 그대로)는 들어가기, 실행 파일은 실행
const canRun = computed(() => !saving.value && tabReady.value && online.value);
const press = useItemPress<DirEntry, string>({
    key: e => e.path,
    selecting: () => selected.value.size > 0,
    check: toggle,
    open: e => {
        if (isDirLike(e)) go(e.path);
        else if (e.exec && canRun.value) runFile(e);
    },
});
const { focused, dragOk } = press;
const onBodyClick = (e: MouseEvent) => {
    if (!(e.target as HTMLElement).closest('.DirGridCell')) focused.value = null;
}

// 칸 메뉴(⋮·우클릭) — 맨 위 정보 카드(판 D′: 크기·수정 시각·링크 원본은 여기서만)
type EntryAct = 'select' | 'goto';
const menuItems = (e: DirEntry): RowMenuItem<EntryAct>[] => [
    { act: 'select', icon: selected.value.has(e.path) ? 'mdi-checkbox-blank-outline' : 'mdi-checkbox-marked-outline', title: selected.value.has(e.path) ? '선택 해제' : '선택' },
    ...(linkDest(e) ? [{ act: 'goto' as const, icon: 'mdi-arrow-top-right', title: '원본 위치로' }] : []),
];
const onAct = (act: EntryAct, e: DirEntry) => act === 'select' ? toggle(e) : go(linkDest(e)!);
const kindText = (e: DirEntry) => {
    if (e.kind === 'dir') return '디렉토리';
    if (e.kind === 'link') return !e.targetPath ? '끊긴 링크' : e.targetDir ? '링크 → 디렉토리' : '링크 → 파일';
    return e.exec ? '실행 파일' : '파일';
}
const infoRows = (e: DirEntry): [string, string][] => ([
    ['종류', kindText(e)],
    e.kind === 'link' && ['원본', e.targetPath ?? `${e.linkTarget ?? ''} (없음)`],
    e.kind === 'file' && ['크기', sizeText(e.size)],
    !!e.modTime && ['수정', timeText(e.modTime)],
] as const).filter((r): r is [string, string] => !!r);
const ctxMenu = ref<{ entry: DirEntry; at: [number, number] }>();
const ctxOpen = ref(false);
const onContextMenu = (ev: MouseEvent, e: DirEntry) => {
    ev.preventDefault();
    if (press.lastPointer() !== 'mouse') return;   // 터치 길게 누르기 = 선택으로 처리
    ctxMenu.value = { entry: e, at: [ev.clientX, ev.clientY] };
    ctxOpen.value = true;
}

// 폴더-5(업로드): 지금 폴더로. Windows 드라이브 목록("")·읽기 실패·끊김이면 못 올림. REF-transfer-ext.md
const { upload, download, copyTo, spots, setSpot, dragging, byTile, lastDone } = useTransfers();
const { openDialog } = useAppDialog();
const canUpload = computed(() => online.value && !!cur.value && !error.value);
const mine = computed(() => byTile(props.tileId).filter(t => t.state === 'running' || t.state === 'failed'));  // 끝난 것은 푸터 목록에만
// 이 타일이 지금 보는 폴더 = 올리기·내려받기·worker 간 전송의 출발·도착
const spot = computed<UploadSource | undefined>(() => {
    const t = tile.value, d = cur.value;
    return t && d && canUpload.value ? { tileId: props.tileId, instance: t.instance, folderId: t.nodeId, dir: d } : undefined;
});
watch(spot, sp => setSpot(props.tileId, sp), { immediate: true });
onBeforeUnmount(() => setSpot(props.tileId, undefined));

const send = (files: File[]) => {
    if (spot.value) void upload(spot.value, files);
}

// 전송-11(화면: 선택·메뉴): 내려받기(Q33(내려받기 범위)) · 다른 장비로 보내기(받을 곳 = 보드에 열린 다른 장비 타일, Q34)
const onDownload = () => {
    if (!spot.value) return;
    void download(spot.value, selEntries.value);
    clearSel();
}
const selFiles = computed(() => selEntries.value.filter(isFileLike));
const targets = computed(() => Object.values(spots.value).filter(sp => sp.tileId !== props.tileId));
const onSendTo = (dst: UploadSource) => {
    if (!spot.value) return;
    void copyTo(spot.value, selEntries.value, dst);
    clearSel();
}

// 전송-12(화면: 끌어 놓기): 파일 줄을 끌어 다른 장비 타일에 놓기. 선택한 줄을 끌면 선택한 파일 전부
const onRowDragStart = (ev: DragEvent, e: DirEntry) => {
    if (!spot.value || !ev.dataTransfer) return;
    const list = selected.value.has(e.path) ? selFiles.value : [e];
    dragging.value = { src: spot.value, entries: list };
    ev.dataTransfer.setData(DRAG_MIME, String(list.length));
    ev.dataTransfer.effectAllowed = 'copy';
}
const onRowDragEnd = () => { dragging.value = null; };
const isDeviceDrag = (e: DragEvent) => !!e.dataTransfer?.types.includes(DRAG_MIME);
// 같은 장비의 같은 폴더(자기 자신 포함)에는 놓지 않음
const canTakeDrag = () => {
    const d = dragging.value, sp = spot.value;
    return !!d && !!sp && !(d.src.instance === sp.instance && d.src.dir === sp.dir);
}

const fileEl = ref<HTMLInputElement>();
const onPick = () => {
    const el = fileEl.value;
    if (!el?.files) return;
    send(Array.from(el.files));
    el.value = '';     // 같은 파일을 다시 고를 수 있게
}

// OS 파일 끌어 놓기 — 파일만(Q26(폴더째 올리기) = 나중에)
const dropping = ref(false);
const hasFiles = (e: DragEvent) => !!e.dataTransfer?.types.includes('Files');
const onDragOver = (e: DragEvent) => {
    if (isDeviceDrag(e) ? !canTakeDrag() : (!hasFiles(e) || !canUpload.value)) return;
    e.preventDefault();
    if (e.dataTransfer) e.dataTransfer.dropEffect = 'copy';
    dropping.value = true;
}
const onDragLeave = (e: DragEvent) => {
    if (!bodyEl.value?.contains(e.relatedTarget as Node | null)) dropping.value = false;
}
const onDrop = (e: DragEvent) => {
    if (isDeviceDrag(e)) {
        e.preventDefault();
        dropping.value = false;
        const d = dragging.value;
        if (d && spot.value && canTakeDrag()) void copyTo(d.src, d.entries, spot.value);
        dragging.value = null;
        return;
    }
    if (!hasFiles(e) || !canUpload.value) return;
    e.preventDefault();
    dropping.value = false;
    const files: File[] = [], dirs: string[] = [];
    for (const item of Array.from(e.dataTransfer?.items ?? [])) {
        const f = item.getAsFile();
        if (!f) continue;
        if (item.webkitGetAsEntry?.()?.isDirectory) dirs.push(f.name);
        else files.push(f);
    }
    if (dirs.length) openDialog({ type: 'warning', content: `폴더는 아직 올릴 수 없습니다(파일만): ${dirs.join(', ')}` });
    send(files);
}

// 이 폴더로 올린 전송이 끝나면(어느 탭에서든) 목록을 다시 읽는다
watch(lastDone, d => {
    if (d && d.instanceKey === instance.value && d.dir === cur.value) void dir.reload();
});
</script>

<template>
<div ref="bodyEl" data-pager-press class="DeviceTileBody" :class="{ dropping }" @click="onBodyClick" @scroll.passive="onScroll" @dragover="onDragOver" @dragleave="onDragLeave" @drop="onDrop">
    <div class="crumbs" :class="{ editing }" :data-pager-press-off="editing || undefined" @click="onCrumbsClick">
        <path-input v-if="editing" :value="here" :suggest="suggestPath" :submit="submitPath" @close="editing = false" />
        <!-- 입력 중엔 도구를 숨겨 입력칸이 줄 전체를 쓴다(좁은 타일에서 자동완성 목록 폭 확보) -->
        <template v-else>
            <!-- [경로·새로고침] / [도구] — 경로 자리가 최소 폭보다 좁아지면 도구가 둘째 줄(오른쪽 끝)로 -->
            <span class="lead">
                <span class="path">
                    <template v-for="(c, i) in crumbs" :key="c.path">
                        <span v-if="i > 0 && sepAfter(crumbs[i - 1]!.name)" class="sep">/</span>
                        <span v-if="i === crumbs.length - 1" class="cur">
                            <v-icon v-if="!c.name" size="13" icon="mdi-harddisk" />{{ c.name || '드라이브' }}
                        </span>
                        <button v-else @click="go(c.path)">
                            <v-icon v-if="!c.name" size="13" icon="mdi-harddisk" />{{ c.name || '드라이브' }}
                        </button>
                    </template>
                </span>
                <button class="tile-btn refresh" title="새로고침" :disabled="loading" @click="dir.reload()">
                    <v-icon size="15" icon="mdi-refresh" />
                </button>
            </span>
            <span class="tools tile-btn-row-end">
                <template v-if="selEntries.length">
                    <span class="sel">{{ selEntries.length }}개</span>
                    <button class="tile-btn" title="내려받기(여러 개·폴더는 zip)" :disabled="!canUpload" @click="onDownload">
                        <v-icon size="16" icon="mdi-download" />
                    </button>
                    <v-menu location="bottom end">
                        <template #activator="{ props: act }">
                            <button v-bind="act" class="tile-btn" title="다른 장비로 보내기(파일만)" :disabled="!canUpload || !selFiles.length">
                                <v-icon size="16" icon="mdi-content-copy" />
                            </button>
                        </template>
                        <v-list density="compact" class="targets">
                            <v-list-item v-for="sp in targets" :key="sp.tileId" :title="sp.instance" :subtitle="sp.dir" @click="onSendTo(sp)" />
                            <v-list-item v-if="!targets.length" disabled title="보드에 받을 장비 타일이 없습니다" />
                        </v-list>
                    </v-menu>
                    <button class="tile-btn" title="선택 해제" @click="clearSel">
                        <v-icon size="16" icon="mdi-close" />
                    </button>
                    <span class="tile-btn-sep" />
                </template>
                <span class="dev" :class="{ on: online }" :title="online ? '접속 중' : '접속 끊김'"><i /></span>
                <span class="tile-btn-end tile-btn-end--primary">
                    <button class="tile-btn tile-btn--text" title="이 폴더에 올리기" :disabled="!canUpload" @click="fileEl?.click()">
                        <v-icon size="15" icon="mdi-upload" />올리기
                    </button>
                </span>
            </span>
        </template>
        <input ref="fileEl" type="file" multiple hidden @change="onPick">
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
        <div v-if="entries.length" class="dir-grid">
            <dir-grid-cell v-for="e in entries" :key="e.path"
                :draggable="dragOk && isFileLike(e) && canUpload"
                :icon="entryIcon(e)"
                :tone="isDirLike(e) ? 'warning' : e.kind === 'link' && !e.targetPath ? 'error' : 'on-surface'"
                :dim="!isDirLike(e) && !e.exec"
                :name="e.name"
                :title="e.path"
                :checking="selEntries.length > 0"
                :checked="selected.has(e.path)"
                :focused="focused === e.path"
                @click="press.onClick($event, e)"
                @pointerdown="press.onPointerDown($event, e)"
                @pointermove="press.onPointerMove"
                @pointerup="press.cancelPress"
                @pointercancel="press.cancelPress"
                @contextmenu="onContextMenu($event, e)"
                @dragstart="onRowDragStart($event, e)"
                @dragend="onRowDragEnd"
            >
                <template #action>
                    <button v-if="e.exec" class="dg-btn dg-btn--l" style="--c: var(--v-theme-primary)" title="새 터미널 타일에서 실행" :disabled="!canRun" @click="runFile(e)">
                        <v-icon size="11" icon="mdi-play" />실행
                    </button>
                </template>
                <template #menu>
                    <node-row-menu :items="menuItems(e)" @act="a => onAct(a, e)">
                        <template #head><div class="dg-info"><span class="t">{{ e.name }}</span><template v-for="[k, v] in infoRows(e)" :key="k"><span class="k">{{ k }}</span><span class="v mono">{{ v }}</span></template></div></template>
                    </node-row-menu>
                </template>
                <template #left>
                    <span v-if="e.kind === 'link' && e.targetPath" class="dg-lbl" style="--c: var(--v-theme-info)"><v-icon size="10" icon="mdi-arrow-top-right" />링크</span>
                    <span v-else-if="e.kind === 'link'" class="dg-lbl" style="--c: var(--v-theme-error)">! 끊김</span>
                </template>
            </dir-grid-cell>
        </div>
        <p v-if="!loading && !entries.length" class="empty">비어 있음</p>
        <p v-if="total > entries.length" class="more">
            {{ total }}개 중 {{ entries.length }}개
            <v-btn size="x-small" variant="text" :disabled="loading" @click="dir.loadMore()">더 보기</v-btn>
        </p>
    </template>
    <node-row-menu v-if="ctxMenu" v-model="ctxOpen" :at="ctxMenu.at" :items="menuItems(ctxMenu.entry)" @act="a => ctxMenu && onAct(a, ctxMenu.entry)">
        <template #head><div class="dg-info"><span class="t">{{ ctxMenu.entry.name }}</span><template v-for="[k, v] in infoRows(ctxMenu.entry)" :key="k"><span class="k">{{ k }}</span><span class="v mono">{{ v }}</span></template></div></template>
    </node-row-menu>
    <!-- 이 타일에서 시작한 전송 — Q27(타일 막대 범위) -->
    <div v-if="mine.length" class="xfer">
        <transfer-row v-for="t in mine" :key="t.id" :item="t" />
    </div>
</div>
</template>

<style scoped lang="scss">
.DeviceTileBody {
    flex: 1;
    min-height: 0;
    overflow: auto;
    @include vars.font('body-medium');
    button:where(:not(.v-btn, .tile-btn, .dg-btn, .dg-lbl)) {    // scoped 규칙은 자식 v-btn 루트에도 걸린다 — 여백·글자를 지우지 않게(:where = 우선순위 그대로)
        padding: 0;
        border: 0;
        background: none;
        color: inherit;
        font: inherit;
        cursor: pointer;
    }
    // 터치 길게 누르기 = 타일 컨트롤 패널(길게-4) — 글자 선택·기본 메뉴 대신. 입력칸은 되살림(iOS는 막으면 입력이 안 됨)
    @media (pointer: coarse) {
        user-select: none;
        -webkit-user-select: none;
        -webkit-touch-callout: none;
        input { user-select: text; -webkit-user-select: text; }
    }
}
// 경로 줄 = 28px 한 줄, 좁은 타일은 두 줄. 경로는 앞쪽부터 줄어든다
.crumbs {
    position: sticky;
    top: 0;
    z-index: 3;     // 목록 칸 버튼·라벨(dir-grid z 1~2)보다 위 — 아니면 자동완성 목록을 뚫고 보인다
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0 2px;
    min-height: 28px;
    padding: 0 0 0 vars.$spacing-md;
    @include vars.font('body-small');
    background: rgb(var(--v-theme-surface));
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    .lead {
        flex: 1 1 160px;
        min-width: 160px;
        height: 28px;
        display: flex;
        align-items: center;
    }
    .path {
        flex: 0 1 auto;
        min-width: 0;
        display: flex;
        align-items: center;
        gap: 2px;
        overflow: hidden;
        white-space: nowrap;
        > * { flex: 0 1000 auto; min-width: 2.2em; overflow: hidden; }   // 앞쪽 경로는 글자 하나 + …까지만 줄어든다
        > .sep { flex: none; min-width: 0; opacity: 0.5; }
        > .cur { flex-shrink: 1; min-width: 0; }
        button { color: rgb(var(--v-theme-primary)); }
    }
    .path > button, .cur { display: inline-flex; align-items: center; gap: 2px; }
    .cur { font-weight: vars.$weight-md; }
    &:not(.editing) { cursor: text; }
    .refresh { min-width: 38px; margin-left: 2px; }
    .tools {
        flex: none;
        height: 28px;
        cursor: default;
        margin-left: auto;
        display: flex;
        align-items: stretch;
        > .tile-btn { min-width: 42px; }
    }
    .sel { align-self: center; margin-right: 2px; color: rgb(var(--v-theme-primary)); font-weight: vars.$weight-md; }
    .dev { align-self: center; padding: 0 vars.$spacing-sm; }
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
.DeviceTileBody.dropping { outline: 2px dashed rgb(var(--v-theme-primary)); outline-offset: -2px; }
.xfer {
    position: sticky;
    bottom: 0;
    padding: 2px vars.$spacing-md;
    background: rgb(var(--v-theme-surface));
    border-top: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}
.more { display: flex; align-items: center; gap: vars.$spacing-sm; }
</style>
