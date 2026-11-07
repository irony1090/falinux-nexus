<script setup lang="ts">
import TileScroll from '@/feature/widget/component/TileScroll.vue';
import { elRef } from '@/feature/widget/util/elRef.util';
import { computed, reactive, ref, watch } from 'vue';
import { useMemberBoard } from '@/feature/member/store/memberBoard.store';
import type { MemberTile } from '@/feature/member/util/memberGrid.util';
import { fileUrl, modeText, openDownload, getUpdDevice, saveText, setReplace, useBundle, type Ref, type SaveError, type UpdFile } from '../api/updater.api';
import { useUpdConfirm, useUpdMode, useUpdRemove, useUpdRun, useUpdUpload } from '../hook/updActions.hook';
import { useHoldSelect } from '../hook/updSelect.hook';
import { buildTree, filesUnder, parentDir, treeRows, type TreeRow } from '../util/updTree.util';
import { pickedItems, readDropped, takeDropped } from '../util/updDrop.util';
import { osSetOf, stepsExt, stepsTemplate } from '../util/updSteps.util';
import { commonTileKey, sizeText, stepsTileKey, textTileKey } from '../util/updGrid.util';
import MbNote from '@/feature/member/component/MbNote.vue';

// Files 타일 — 범위 탭 [공용 | 메인키 | 서브키]. 메인키 = 목록·업로드·삭제·새 텍스트 파일, 텍스트 파일을 누르면 Script 타일(steps.json = steps.json 타일), 바이너리는 타일 없음.
// 공용 = 읽기 전용(여러 빌드가 같이 씀, 고치려면 공용 묶음을 연다). 서브키 = 자리만(P1) — REF-update-deploy.md "P1 착수 계획" 단위-6(편집 타일)
type Scope = 'common' | 'main' | 'sub';
const props = defineProps<{ tile: MemberTile; bundle: Ref }>();
const { open, dropMissing, tiles } = useMemberBoard();
const ctx = useUpdRun();
const { busy, message, progress, run } = ctx;
const { upload, newText } = useUpdUpload(ctx);
const removeMany = useUpdRemove(ctx);

const q = useBundle(props.bundle);
const b = computed(() => q.data.value);
watch(() => (q.error.value as { status?: number } | null)?.status, st => { if (st === 404) dropMissing(props.tile.key); }, { immediate: true });

const scope = ref<Scope>('main');
const mainFiles = computed(() => (b.value?.files ?? []).filter(f => f.scope === 'main'));
const subFiles = computed(() => (b.value?.files ?? []).filter(f => f.scope === 'sub'));
const commonCount = computed(() => (b.value?.uses ?? []).reduce((n, u) => n + u.files.length, 0));
const locked = computed(() => !!b.value?.locked);
const openKeys = computed(() => new Set(tiles.value.map(t => t.key)));

const keyOf = (f: UpdFile, r: Ref) => r.kind === 'build' && f.path === 'steps.json' ? stepsTileKey(r.owner, r.version) : textTileKey(r, f.path);
const openFile = (f: UpdFile, r: Ref = props.bundle) => { if (f.text) open(keyOf(f, r), props.tile.key); };

const pick = ref<HTMLInputElement>();
const pickDir = ref<HTMLInputElement>();
const onPick = async (e: Event) => {
    const input = e.target as HTMLInputElement;
    if (input.files?.length) await upload(props.bundle, pickedItems(input.files), b.value?.files ?? []);
    input.value = '';
}
const { isExec, toggleExec, editMode } = useUpdMode(props.bundle, run);

// 폴더 트리(결정-3(폴더 트리)) — 처음엔 모두 펼침, 접힘 상태는 타일을 닫으면 잊는다
const closed = reactive(new Set<string>());
const tree = computed(() => buildTree(mainFiles.value));
const rows = computed(() => treeRows(tree.value, closed));
const pathsOf = (r: TreeRow) => r.kind === 'dir' ? filesUnder(r.dir) : [r.path];
const fold = (path: string) => closed.has(path) ? closed.delete(path) : closed.add(path);

// 선택 삭제(결정-1(선택 방식)) — 잠긴 빌드는 선택 모드에 들어가지 않는다(결정-7(잠긴 빌드))
const sel = useHoldSelect();
const selCount = computed(() => sel.picked.size);
watch(mainFiles, fs => sel.keepOnly(fs.map(f => f.path)));
watch([scope, locked], () => sel.clear());
const onRowDown = (e: PointerEvent, r: TreeRow) => { if (!locked.value) sel.onPointerDown(e, () => sel.start(pathsOf(r))); };
const onRowClick = (r: TreeRow) => {
    if (sel.consumeClick()) return;
    if (sel.selecting.value) sel.toggle(pathsOf(r));
    else if (r.kind === 'dir') fold(r.path);
    else openFile(r.file);
}
const pickAll = () => sel.toggle(mainFiles.value.map(f => f.path));
const removePicked = async () => {
    const paths = mainFiles.value.map(f => f.path).filter(p => sel.picked.has(p));
    if (await removeMany(props.bundle, paths, `${paths.length}개 파일 삭제`)) sel.clear();
}
const removeDir = (r: TreeRow & { kind: 'dir' }) => removeMany(props.bundle, filesUnder(r.dir), `${r.path}/ 삭제`);

// PC 탐색기 끌어 놓기(결정-5(끌어 놓기 규칙)) — 폴더 행 = 그 폴더, 파일 행 = 그 파일이 든 폴더, 빈 곳 = 루트. 메인키 탭·안 잠김·선택 모드 아님·올리는 중 아님일 때만
const bodyEl = ref<HTMLElement>();
const setBody = elRef(bodyEl);
const dropping = ref(false);
const dropDir = ref('');
const canDrop = computed(() => scope.value === 'main' && !locked.value && !sel.selecting.value && !busy.value);
const hasFiles = (e: DragEvent) => !!e.dataTransfer?.types.includes('Files');
const dirAt = (t: EventTarget | null) => (t instanceof Element ? t.closest<HTMLElement>('[data-drop-dir]')?.dataset.dropDir : undefined) ?? '';
const dropAt = (r: TreeRow) => r.kind === 'dir' ? r.path : parentDir(r.path);
// 못 받는 때도 기본 동작(브라우저가 파일을 열어 페이지를 떠남)은 막고 '안 됨' 커서만 보인다
const onDragOver = (e: DragEvent) => {
    if (!hasFiles(e)) return;
    e.preventDefault();
    if (e.dataTransfer) e.dataTransfer.dropEffect = canDrop.value ? 'copy' : 'none';
    if (!canDrop.value) return;
    dropping.value = true;
    dropDir.value = dirAt(e.target);
}
const onDragLeave = (e: DragEvent) => {
    if (!bodyEl.value?.contains(e.relatedTarget as Node | null)) dropping.value = false;
}
const onDrop = async (e: DragEvent) => {
    if (!hasFiles(e)) return;
    e.preventDefault();
    if (!canDrop.value) return;
    dropping.value = false;
    const dir = dirAt(e.target), taken = takeDropped(e.dataTransfer);
    await upload(props.bundle, await readDropped(taken, dir), b.value?.files ?? []);
}
watch(canDrop, v => { if (!v) dropping.value = false; });

// 갈아치우기(결정-8(갈아치우기의 뜻)·결정-15(토글 모양)) — 켜진 폴더 = steps.json replace. 부모가 켜져 있으면 자식은 켜진 채 흐림(확인-1(토글 깊이))
const confirm = useUpdConfirm();
const replace = computed(() => b.value?.replace ?? []);
const under = (p: string, d: string) => p.startsWith(`${d}/`);
const repOf = (path: string): 'own' | 'parent' | undefined =>
    replace.value.some(d => under(path, d)) ? 'parent' : replace.value.includes(path) ? 'own' : undefined;
const setReplaceOn = async (path: string, on: boolean) => {
    if (on && !(await confirm(`${path}/ 를 갈아치우기로 바꿀까요?`, [`배포 때 장비의 ${path}/ 안 기존 파일을 모두 지우고 새 파일로 채웁니다.`, '지워질 파일도 배포 전에 백업됩니다.'], '갈아치우기로 바꾸기'))) return;
    // 부모를 켜면 자식 항목은 지운다(부모가 이미 비우므로)
    const next = on ? [...replace.value.filter(d => !under(d, path)), path] : replace.value.filter(d => d !== path);
    const hasSteps = mainFiles.value.some(f => f.path === 'steps.json');
    await run(async () => {
        try {
            // steps.json이 없으면 틀부터(결정-9(저장 위치)) — 확장자는 설치본 os로
            if (!hasSteps) await saveText(props.bundle, 'steps.json', stepsTemplate(stepsExt(osSetOf((await getUpdDevice(props.bundle.owner)).instances))), true);
            await setReplace(props.bundle, next);
        } catch (e) {
            const err = e as SaveError;
            throw err.status === 422 ? { message: 'steps.json에 오류가 있어 바꾸지 못했습니다 — steps.json 타일에서 먼저 고쳐 주세요' } : err;
        }
    });
}
</script>

<template>
<tile-scroll :ref="setBody" class="mb-body" :class="{ dropping }" data-pager-press @dragover="onDragOver" @dragleave="onDragLeave" @drop="onDrop">
    <p v-if="q.isError.value" class="mb-empty">빌드를 찾을 수 없습니다(지워졌을 수 있습니다)</p>
    <template v-else-if="b">
        <div class="upd-set scopes">
            <span class="ct-grp ct-grp--sm ct-grp--r" role="group" aria-label="범위">
                <button type="button" class="ct-seg" :class="scope === 'common' ? 'ct-seg--primary' : 'ct-seg--accent'" :aria-pressed="scope === 'common'" @click="scope = 'common'">공용 {{ commonCount }}</button>
                <button type="button" class="ct-seg" :class="scope === 'main' ? 'ct-seg--primary' : 'ct-seg--accent'" :aria-pressed="scope === 'main'" @click="scope = 'main'">메인키 {{ mainFiles.length }}</button>
                <button type="button" class="ct-seg" :class="scope === 'sub' ? 'ct-seg--primary' : 'ct-seg--accent'" :aria-pressed="scope === 'sub'" @click="scope = 'sub'">서브키 {{ subFiles.length }}</button>
            </span>
        </div>
        <div class="upd-fields">
            <span class="mb-lbl">경로</span>
            <span class="val mono">{{ b.owner }}/{{ b.version }}</span>
        </div>
        <p v-if="message" class="mb-err msg">{{ message }}</p>

        <template v-if="scope === 'main'">
            <div class="tree" data-pager-press-off>
                <div v-for="r in rows" :key="r.kind + r.path" class="upd-row" :style="{ '--d': r.depth }"
                    :class="{ click: r.kind === 'dir' || r.file.text || sel.selecting.value, open: r.kind === 'file' && openKeys.has(keyOf(r.file, bundle)), sel: sel.selecting.value && sel.state(pathsOf(r)) !== 'none', drop: dropping && r.kind === 'dir' && r.path === dropDir, replace: r.kind === 'dir' && !!repOf(r.path) }" :data-drop-dir="dropAt(r)"
                    @pointerdown="onRowDown($event, r)" @pointermove="sel.onPointerMove" @pointerup="sel.cancelPress" @pointercancel="sel.cancelPress" @contextmenu="sel.onContextMenu" @click="onRowClick(r)">
                    <span v-if="sel.selecting.value" class="cb" :class="sel.state(pathsOf(r))"><v-icon :icon="sel.state(pathsOf(r)) === 'some' ? 'mdi-minus' : 'mdi-check'" size="12" /></span>
                    <template v-if="r.kind === 'dir'">
                        <v-icon icon="mdi-chevron-right" size="18" class="chev" :class="{ on: !closed.has(r.path) }" />
                        <v-icon icon="mdi-folder-outline" size="18" class="ic dir" />
                        <span class="txt">
                            <span class="main upd-mono">{{ r.name }}/</span>
                            <span class="sub">{{ r.dir.count }}개 파일{{ repOf(r.path) === 'own' ? ` · 배포 때 장비의 ${r.path}/ 를 비우고 채움` : repOf(r.path) ? ' · 상위 폴더 갈아치우기를 따름' : '' }}</span>
                        </span>
                        <span v-if="dropping && r.path === dropDir" class="tail">여기에 놓기</span>
                        <div v-else-if="!sel.selecting.value" class="ct-grp ct-grp--sm" @click.stop>
                            <button type="button" class="ct-seg" :class="repOf(r.path) ? 'ct-seg--dim' : 'ct-seg--txt'" :aria-pressed="!repOf(r.path)" title="합치기 — 배포 때 겹치는 파일만 바꾸고 나머지는 둠" :disabled="busy || locked || repOf(r.path) === 'parent'" @click="repOf(r.path) === 'own' && setReplaceOn(r.path, false)">합치기</button>
                            <button type="button" class="ct-seg" :class="repOf(r.path) ? 'ct-seg--warning' : 'ct-seg--dim'" :aria-pressed="!!repOf(r.path)" title="갈아치우기 — 배포 때 장비의 이 폴더를 비우고 채움" :disabled="busy || locked || repOf(r.path) === 'parent'" @click="!repOf(r.path) && setReplaceOn(r.path, true)">갈아치우기</button>
                            <button type="button" class="ct-seg ct-seg--error" title="폴더 삭제(안의 파일 전부)" :disabled="busy || locked" @click="removeDir(r)"><v-icon icon="mdi-delete-outline" size="15" /></button>
                        </div>
                    </template>
                    <template v-else>
                        <span class="sp" />
                        <v-icon :icon="r.file.text ? 'mdi-file-document-outline' : 'mdi-file-cog-outline'" size="18" class="ic" />
                        <span class="txt" :title="r.path">
                            <span class="main upd-mono">{{ r.name }}</span>
                            <span class="sub">{{ sizeText(r.file.size) }}{{ r.file.text ? '' : ' · 바이너리' }}</span>
                        </span>
                        <div v-if="!sel.selecting.value" class="ct-grp ct-grp--sm" @click.stop>
                            <button type="button" class="ct-seg" :class="isExec(r.file) ? 'ct-seg--accent' : 'ct-seg--dim'" :aria-pressed="isExec(r.file)" :title="isExec(r.file) ? '실행 권한 끄기(장비에 깔릴 때 실행 비트 뺌, 지금 실행하지 않음)' : '실행 권한 켜기(장비에 깔릴 때 실행 비트 줌, 지금 실행하지 않음)'" :disabled="busy || locked" @click="toggleExec(r.file)">실행 권한</button>
                            <button type="button" class="ct-seg ct-seg--mono" title="권한(8진수)" :disabled="busy || locked" @click="editMode(r.file)">{{ modeText(r.file.mode) }}</button>
                            <button type="button" class="ct-seg" title="받기" @click="openDownload(fileUrl(bundle, r.path))"><v-icon icon="mdi-download" size="15" /></button>
                            <button type="button" class="ct-seg ct-seg--error" title="삭제" :disabled="busy || locked" @click="removeMany(bundle, [r.path], `${r.path} 삭제`)"><v-icon icon="mdi-delete-outline" size="15" /></button>
                        </div>
                    </template>
                </div>
            </div>
            <p v-if="!mainFiles.length" class="mb-empty">파일이 없습니다. 아래에서 올리거나 텍스트 파일을 만듭니다</p>
        </template>

        <template v-else-if="scope === 'common'">
            <p class="upd-hint">읽기 전용입니다. 여러 빌드가 함께 쓰는 파일이라, 고치려면 공용 묶음을 엽니다.</p>
            <template v-for="u in b.uses" :key="u.name">
                <div class="upd-row">
                    <span class="txt">
                        <span class="main upd-mono">{{ u.name }}@{{ u.version }}</span>
                        <span class="sub">{{ u.files.length }}개 파일{{ u.locked ? ' · 잠김' : '' }}</span>
                    </span>
                    <div class="ct-grp ct-grp--sm"><button type="button" class="ct-seg ct-seg--accent" @click="open(commonTileKey(u.name, u.version), tile.key)">공용 묶음 열기</button></div>
                </div>
                <div v-for="f in u.files" :key="`${u.name}/${f.path}`" class="upd-row inner">
                    <span class="txt">
                        <span class="main upd-mono">{{ f.path }}</span>
                        <span class="sub">{{ u.name }} · {{ sizeText(f.size) }} · {{ modeText(f.mode) }}</span>
                    </span>
                </div>
            </template>
            <p v-if="!b.uses.length" class="mb-empty">steps.json의 uses에 공용 묶음이 없습니다</p>
        </template>

        <p v-else class="mb-empty">서브키 탭은 한 인스턴스에만 줄 파일을 두는 자리입니다.<br>이번 범위에서는 자리만 있습니다.</p>
    </template>
    <p v-if="dropping" class="drop-tip">놓으면 <span class="upd-mono">{{ dropDir ? `${dropDir}/` : '맨 위(루트)' }}</span> 안으로 올라갑니다(폴더 이름 유지)</p>
</tile-scroll>
<div v-if="b" class="mb-foot">
    <mb-note :text="progress ?? (sel.selecting.value ? `${selCount}개 선택` : locked ? '잠겨 있어 파일을 고칠 수 없습니다' : replace.length ? `갈아치우기 ${replace.length}곳` : '깔리는 순서: 공용, 메인키, 서브키')" />
    <div v-if="scope === 'main' && sel.selecting.value" class="mb-end">
        <button type="button" class="mb-btn" :disabled="busy" @click="pickAll">{{ selCount === mainFiles.length ? '전체 해제' : '전체 선택' }}</button>
        <button type="button" class="mb-btn" :disabled="busy" @click="sel.clear()">취소</button>
        <button type="button" class="mb-btn mb-btn--danger" :disabled="busy || !selCount" @click="removePicked">삭제 {{ selCount }}</button>
    </div>
    <div v-else-if="scope === 'main'" class="mb-end">
        <button type="button" class="mb-btn" :disabled="busy || locked" @click="newText(bundle, p => open(textTileKey(bundle, p), tile.key))">텍스트 파일 추가</button>
        <button type="button" class="mb-btn" :disabled="busy || locked" @click="pickDir?.click()">폴더 업로드</button>
        <button type="button" class="mb-btn mb-btn--primary" :disabled="busy || locked" @click="pick?.click()">업로드</button>
    </div>
    <input ref="pick" type="file" multiple hidden @change="onPick" />
    <input ref="pickDir" type="file" webkitdirectory hidden @change="onPick" />
</div>
</template>

<style scoped lang="scss">
// 범위 탭 = 오른쪽 벽에 붙는 묶음
.scopes { padding: 8px 0 0; }
.msg { margin: 0 16px 8px; }
.ic { flex: none; opacity: 0.7; }
.inner { padding-left: 34px; min-height: 34px; }
// 트리 행 — 깊이마다 16px 들여쓰기, 길게 누르기 중 글자 선택·터치 메뉴 막음
.tree {
    user-select: none;
    -webkit-touch-callout: none;
    > .upd-row { gap: 6px; padding-left: calc(10px + var(--d, 0) * 16px); }
    > .upd-row.replace { background: rgba(var(--v-theme-warning), 0.08); }
    > .upd-row.sel { background: rgba(var(--v-theme-primary), 0.1); }
}
// 끌어 놓기 — 몸통 점선, 놓일 폴더 행 강조, 아래쪽 안내(시안 ③, 끌기-1(안내 문구) = 끄는 항목 이름 없이)
.mb-body.dropping { outline: 2px dashed rgb(var(--v-theme-primary)); outline-offset: -2px; }
.tree > .upd-row.drop { background: rgba(var(--v-theme-primary), 0.18); box-shadow: inset 0 0 0 2px rgb(var(--v-theme-primary)); }
.tail { flex: none; padding-right: 10px; @include vars.font('body-small'); color: rgb(var(--v-theme-primary)); white-space: nowrap; }
.drop-tip {
    position: sticky;
    bottom: 8px;
    margin: 8px 12px;
    padding: 6px 10px;
    border-radius: 6px;
    background: rgb(var(--v-theme-primary));
    color: rgb(var(--v-theme-on-primary));
    @include vars.font('body-small');
    pointer-events: none;
}
.chev { flex: none; opacity: 0.6; transition: transform 0.12s; &.on { transform: rotate(90deg); } }
.sp { flex: none; width: 18px; }
.dir { color: rgb(var(--v-theme-primary)); opacity: 1; }
.cb {
    flex: none;
    display: grid;
    place-items: center;
    width: 18px;
    height: 18px;
    margin-right: 2px;
    border-radius: 50%;
    border: 1.5px solid rgba(var(--v-theme-on-surface), var(--v-disabled-opacity));
    color: transparent;
    &.some, &.all { background: rgb(var(--v-theme-primary)); border-color: rgb(var(--v-theme-primary)); color: rgb(var(--v-theme-on-primary)); }
}
</style>
