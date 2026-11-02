<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useMemberBoard } from '@/feature/member/store/memberBoard.store';
import type { MemberTile } from '@/feature/member/util/memberGrid.util';
import { deleteFile, fileUrl, modeText, openDownload, useBundle, type Ref, type UpdFile } from '../api/updater.api';
import { useUpdConfirm, useUpdMode, useUpdRun, useUpdUpload } from '../hook/updActions.hook';
import { commonTileKey, sizeText, stepsTileKey, textTileKey } from '../util/updGrid.util';
import MbNote from '@/feature/member/component/MbNote.vue';

// Files 타일 — 범위 탭 [공용 | 메인키 | 서브키]. 메인키 = 목록·업로드·삭제·새 텍스트 파일, 텍스트 파일을 누르면 Script 타일(steps.json = steps.json 타일), 바이너리는 타일 없음.
// 공용 = 읽기 전용(여러 빌드가 같이 씀, 고치려면 공용 묶음을 연다). 서브키 = 자리만(P1) — REF-update-deploy.md "P1 착수 계획" 단위-6(편집 타일)
type Scope = 'common' | 'main' | 'sub';
const props = defineProps<{ tile: MemberTile; bundle: Ref }>();
const { open, dropMissing, tiles } = useMemberBoard();
const confirm = useUpdConfirm();
const ctx = useUpdRun();
const { busy, message, progress, run } = ctx;
const { upload, newText } = useUpdUpload(ctx);

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
const onPick = async (e: Event, folder: boolean) => {
    const input = e.target as HTMLInputElement;
    if (input.files?.length) await upload(props.bundle, input.files, folder, b.value?.files ?? []);
    input.value = '';
}
const { isExec, toggleExec, editMode } = useUpdMode(props.bundle, run);

const remove = async (f: UpdFile) => {
    if (!(await confirm(`${f.path} 삭제`, f.path === 'steps.json' ? ['steps.json을 지우면 uses도 비워집니다.'] : ['이 빌드에서 파일을 지웁니다.'], '삭제', 'error'))) return;
    await run(() => deleteFile(props.bundle, f.path));
}
</script>

<template>
<div class="mb-body" data-pager-press>
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
            <div v-for="f in mainFiles" :key="f.path" class="upd-row" :class="{ click: f.text, open: openKeys.has(keyOf(f, bundle)) }" @click="openFile(f)">
                <v-icon :icon="f.text ? 'mdi-file-document-outline' : 'mdi-file-cog-outline'" size="18" class="ic" />
                <span class="txt">
                    <span class="main upd-mono">{{ f.path }}</span>
                    <span class="sub">{{ sizeText(f.size) }}{{ f.text ? '' : ' · 바이너리' }}</span>
                </span>
                <div class="ct-grp ct-grp--sm" @click.stop>
                    <button type="button" class="ct-seg" :class="{ 'ct-seg--accent': isExec(f) }" :aria-pressed="isExec(f)" :title="isExec(f) ? '실행 끄기' : '실행 켜기'" :disabled="busy || locked" @click="toggleExec(f)"><v-icon icon="mdi-play-box-outline" size="15" /></button>
                    <button type="button" class="ct-seg ct-seg--mono" title="권한(8진수)" :disabled="busy || locked" @click="editMode(f)">{{ modeText(f.mode) }}</button>
                    <button type="button" class="ct-seg" title="받기" @click="openDownload(fileUrl(bundle, f.path))"><v-icon icon="mdi-download" size="15" /></button>
                    <button type="button" class="ct-seg ct-seg--error" title="삭제" :disabled="busy || locked" @click="remove(f)"><v-icon icon="mdi-delete-outline" size="15" /></button>
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
</div>
<div v-if="b" class="mb-foot">
    <mb-note :text="progress ?? (locked ? '잠겨 있어 파일을 고칠 수 없습니다' : '깔리는 순서: 공용, 메인키, 서브키')" />
    <div v-if="scope === 'main'" class="mb-end">
        <button type="button" class="mb-btn" :disabled="busy || locked" @click="newText(bundle, p => open(textTileKey(bundle, p), tile.key))">텍스트 파일 추가</button>
        <button type="button" class="mb-btn" :disabled="busy || locked" @click="pickDir?.click()">폴더 업로드</button>
        <button type="button" class="mb-btn mb-btn--primary" :disabled="busy || locked" @click="pick?.click()">업로드</button>
    </div>
    <input ref="pick" type="file" multiple hidden @change="onPick($event, false)" />
    <input ref="pickDir" type="file" webkitdirectory hidden @change="onPick($event, true)" />
</div>
</template>

<style scoped lang="scss">
// 범위 탭 = 오른쪽 벽에 붙는 묶음
.scopes { padding: 8px 0 0; }
.msg { margin: 0 16px 8px; }
.ic { flex: none; opacity: 0.7; }
.inner { padding-left: 34px; min-height: 34px; }
</style>
