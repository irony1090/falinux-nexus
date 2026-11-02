<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useMemberBoard } from '@/feature/member/store/memberBoard.store';
import type { MemberTile } from '@/feature/member/util/memberGrid.util';
import { MAIN_KEY_PATTERN } from '@/feature/member/util/member.util';
import CtTextField from '@/feature/widget/component/CtTextField.vue';
import { createBundle, deleteBundle, deleteFile, fileUrl, openDownload, patchBundle, useBundle, type Ref, type UpdFile } from '../api/updater.api';
import { useUpdConfirm, useUpdRun, useUpdUpload } from '../hook/updActions.hook';
import { buildTileKey, commonTileKey, shortSha, sizeText, textTileKey, timeText } from '../util/updGrid.util';
import UpdZipMenu from './UpdZipMenu.vue';
import { useUpdPrompt } from './UpdPrompt.vue';
import MbNote from '@/feature/member/component/MbNote.vue';

// 공용 묶음 타일 — 상세 + 파일 목록(업로드·텍스트) + 쓰는 빌드. 쓰는 빌드가 있으면 삭제 불가, 잠기면 파일 수정 불가 — REF-update-rules.md "삭제"·"잠금"
const props = defineProps<{ tile: MemberTile; bundle: Ref }>();
const { open, close, dropMissing, tiles } = useMemberBoard();
const prompt = useUpdPrompt();
const confirm = useUpdConfirm();
const ctx = useUpdRun();
const { busy, message, progress, run } = ctx;
const { upload, newText } = useUpdUpload(ctx);

const q = useBundle(props.bundle);
const b = computed(() => q.data.value);
watch(() => (q.error.value as { status?: number } | null)?.status, st => { if (st === 404) dropMissing(props.tile.key); }, { immediate: true });
const locked = computed(() => !!b.value?.locked);
const openKeys = computed(() => new Set(tiles.value.map(t => t.key)));

const note = ref<string>();
const saveNote = () => {
    const v = (note.value ?? '').trim();
    if (note.value === undefined || v === (b.value?.note ?? '')) { note.value = undefined; return; }
    run(() => patchBundle(props.bundle, v), () => note.value = undefined);
}

const openFile = (f: UpdFile) => { if (f.text) open(textTileKey(props.bundle, f.path), props.tile.key); };
const pick = ref<HTMLInputElement>();
const onPick = async (e: Event) => {
    const input = e.target as HTMLInputElement;
    if (input.files?.length) await upload(props.bundle, input.files, false, b.value?.files ?? []);
    input.value = '';
}
const removeFile = async (f: UpdFile) => {
    if (!(await confirm(`${f.path} 삭제`, ['이 묶음을 쓰는 모든 빌드에서 이 파일이 빠집니다.'], '삭제', 'error'))) return;
    await run(() => deleteFile(props.bundle, f.path));
}
const newVersion = async () => {
    const v = await prompt({
        title: `${props.bundle.owner} 새 버전`, ok: '만들기', note: `${props.bundle.version}의 파일을 그대로 복사합니다. 빌드의 uses를 새 버전으로 바꿔야 쓰입니다`,
        fields: [{ key: 'version', label: '새 버전', placeholder: '예: 261104_a', pattern: MAIN_KEY_PATTERN, patternMsg: '영문·숫자·. _ - 로 64자까지' }],
    });
    if (!v) return;
    await run(() => createBundle('common', props.bundle.owner, v.version!, '', { owner: props.bundle.owner, version: props.bundle.version }),
        () => open(commonTileKey(props.bundle.owner, v.version!), props.tile.key));
}
const remove = async () => {
    if (!(await confirm(`${props.bundle.owner}@${props.bundle.version} 삭제`, ['서버의 파일이 지워집니다.'], '삭제', 'error'))) return;
    await run(() => deleteBundle(props.bundle), () => close(props.tile.key));
}
</script>

<template>
<div class="mb-body" data-pager-press>
    <p v-if="q.isError.value" class="mb-empty">공용 묶음을 찾을 수 없습니다(지워졌을 수 있습니다)</p>
    <template v-else-if="b">
        <div class="upd-fields">
            <span class="mb-lbl">올린 때</span>
            <span class="val">{{ timeText(b.createdAt) }}</span>
            <span class="mb-lbl">sha</span>
            <span class="val mono" :title="b.sha">{{ shortSha(b.sha) }}</span>
            <span class="mb-lbl">상태</span>
            <span class="val"><span class="mb-tag" :class="locked ? 'mb-tag--primary' : 'dim'">{{ locked ? '잠김' : '수정 가능' }}</span></span>
            <span class="mb-lbl">설명</span>
            <CtTextField class="edit" :model-value="note ?? b.note" size="sm" wall="r" hide-details maxlength="500" placeholder="설명 달기" aria-label="설명" :disabled="busy"
                @update:model-value="(v: string) => note = v"
                @keydown.enter="($event.target as HTMLInputElement).blur()"
                @update:focused="(f: boolean) => !f && saveNote()"
            />
            <p v-if="locked" class="upd-hint full flat">쓰는 빌드가 잠겨서 이 묶음도 같이 잠겼습니다. 파일과 이름은 고칠 수 없습니다.</p>
            <p v-if="message" class="mb-err full">{{ message }}</p>
        </div>
        <h3 class="mb-sec">파일 {{ b.files.length }} · {{ sizeText(b.files.reduce((n, f) => n + f.size, 0)) }}</h3>
        <div v-for="f in b.files" :key="f.path" class="upd-row" :class="{ click: f.text, open: openKeys.has(textTileKey(bundle, f.path)) }" @click="openFile(f)">
            <span class="txt">
                <span class="main upd-mono">{{ f.path }}</span>
                <span class="sub">{{ sizeText(f.size) }}{{ f.text ? '' : ' · 바이너리' }}</span>
            </span>
            <div class="ct-grp ct-grp--sm" @click.stop>
                <button type="button" class="ct-seg" title="받기" @click="openDownload(fileUrl(bundle, f.path))"><v-icon icon="mdi-download" size="15" /></button>
                <button type="button" class="ct-seg ct-seg--error" title="삭제" :disabled="busy || locked" @click="removeFile(f)"><v-icon icon="mdi-delete-outline" size="15" /></button>
            </div>
        </div>
        <p v-if="!b.files.length" class="mb-empty">파일이 없습니다</p>
        <h3 class="mb-sec">사용 중인 빌드 {{ b.users.length }}</h3>
        <div v-for="u in b.users" :key="`${u.mainKey}/${u.version}`" class="upd-row click" @click="open(buildTileKey(u.mainKey, u.version), tile.key)">
            <span class="txt"><span class="main upd-mono">{{ u.mainKey }}/{{ u.version }}</span></span>
            <span class="mb-tag wall" :class="u.locked ? 'mb-tag--primary' : 'dim'">{{ u.locked ? '잠김' : '수정 가능' }}</span>
        </div>
        <p v-if="!b.users.length" class="upd-hint">쓰는 빌드가 없습니다(지울 수 있습니다)</p>
    </template>
</div>
<div v-if="b" class="mb-foot">
    <mb-note :text="progress ?? (b.users.length ? `쓰는 빌드 ${b.users.length}개 · 삭제 불가 · 수정하면 이 빌드들의 내용이 바뀝니다` : '쓰는 빌드 없음')" />
    <div class="mb-end">
        <button type="button" class="mb-btn mb-btn--danger" :disabled="busy || b.users.length > 0" @click="remove">삭제</button>
        <button type="button" class="mb-btn" :disabled="busy" @click="newVersion">새 버전</button>
        <upd-zip-menu :bundle="bundle" />
        <template v-if="!locked">
            <button type="button" class="mb-btn" :disabled="busy" @click="newText(bundle, p => open(textTileKey(bundle, p), tile.key))">텍스트 파일 추가</button>
            <button type="button" class="mb-btn mb-btn--primary" :disabled="busy" @click="pick?.click()">업로드</button>
        </template>
    </div>
    <input ref="pick" type="file" multiple hidden @change="onPick" />
</div>
</template>

<style scoped lang="scss">
.flat { padding: 4px 0; }
</style>
