<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { useMemberBoard } from '@/feature/member/store/memberBoard.store';
import type { MemberTile } from '@/feature/member/util/memberGrid.util';
import CtCode, { ctCodeLangOf } from '@/feature/widget/component/CtCode.vue';
import { invalidateUpdates, saveText, useBundle, type Ref } from '../api/updater.api';
import { useUpdConfirm, useUpdRun } from '../hook/updActions.hook';
import { useUpdLiveText } from '../hook/updText.hook';
import MbNote from '@/feature/member/component/MbNote.vue';

// Script 타일 — 빌드·공용 묶음의 텍스트 파일 하나 편집(Files 목록에서 연다, 안 A). `.sh`는 서버가 LF로 저장 — REF-update-deploy.md 단위-6(편집 타일)
const props = defineProps<{ tile: MemberTile; bundle: Ref; path: string }>();
const { close, guardClose } = useMemberBoard();
const confirm = useUpdConfirm();
const { qc, busy, message } = useUpdRun();

const q = useBundle(props.bundle);
const b = computed(() => q.data.value);
const meta = computed(() => b.value?.files.find(f => f.scope === 'main' && f.path === props.path));
const locked = computed(() => !!b.value?.locked);
const lang = computed(() => ctCodeLangOf(props.path));

const sha = computed(() => meta.value?.sha);
const { original, text, dirty, stale, load, saved } = useUpdLiveText(props.bundle, props.path, sha, m => message.value = m);
const missing = ref(false);
watch(b, d => {
    if (!d || original.value !== undefined) return;
    if (!meta.value) { missing.value = true; return; }
    void load(sha.value);
}, { immediate: true });
// 다른 곳에서 바뀐 뒤 고치던 것을 버리고 다시 읽기
const reload = async () => {
    if (await confirm('다시 불러오기', '고치던 내용을 버리고 저장된 내용을 불러옵니다.', '버리고 불러오기', 'error')) await load(sha.value);
}

const lines = computed(() => text.value ? text.value.split('\n').length - (text.value.endsWith('\n') ? 1 : 0) : 0);
const eol = computed(() => text.value.includes('\r\n') ? 'CRLF' : 'LF');

const save = async () => {
    if (busy.value || locked.value || !dirty.value) return;
    busy.value = true;
    message.value = undefined;
    try {
        const r = await saveText(props.bundle, props.path, text.value);
        // .sh는 서버가 LF로 바꿔 저장하므로 화면도 같게
        if (props.path.toLowerCase().endsWith('.sh')) text.value = text.value.replace(/\r\n/g, '\n');
        saved(r.file?.sha);
        await invalidateUpdates(qc);
    } catch (e: any) {
        message.value = e?.message || '저장하지 못했습니다';
    } finally {
        busy.value = false;
    }
}
onBeforeUnmount(guardClose(props.tile.key, () => !dirty.value || confirm('저장하지 않은 파일', `${props.path} — 저장하지 않고 닫으면 고친 내용이 사라집니다.`, '저장하지 않고 닫기', 'error')));
</script>

<template>
<div class="body" data-pager-press @keydown.ctrl.s.prevent="save" @keydown.meta.s.prevent="save">
    <p v-if="missing" class="mb-empty">파일이 없습니다(지워졌거나 이름이 바뀌었습니다)</p>
    <template v-else-if="original !== undefined">
        <p class="upd-hint top"><span class="upd-mono">{{ eol }} · {{ lines }}줄</span><template v-if="bundle.kind === 'common'"> · 공용 묶음 — 이 묶음을 쓰는 모든 빌드에 적용됩니다</template><span v-if="dirty" class="dirty"> · 수정됨</span></p>
        <CtCode v-model="text" class="code" :lang="lang" :readonly="locked" />
        <p v-if="message" class="mb-err msg">{{ message }}</p>
    </template>
</div>
<div class="mb-foot">
    <mb-note :text="locked ? '잠겨 있어 고칠 수 없습니다' : stale ? '다른 곳에서 이 파일이 바뀌었습니다' : '저장하면 sha가 바뀝니다 · 잠그기 전에는 언제든 수정할 수 있습니다'" />
    <div class="mb-end">
        <button v-if="missing" type="button" class="mb-btn" @click="close(tile.key)">닫기</button>
        <button v-else-if="stale" type="button" class="mb-btn" :disabled="busy" @click="reload">다시 불러오기</button>
        <button v-if="!missing" type="button" class="mb-btn mb-btn--primary" :disabled="busy || locked || !dirty" @click="save">저장</button>
    </div>
</div>
</template>

<style scoped lang="scss">
.body { flex: 1; min-height: 0; display: flex; flex-direction: column; padding-bottom: 10px; }
.top { flex: none; padding: 8px 16px; }
.dirty { color: rgb(var(--v-theme-warning)); }
.code { flex: 1; min-height: 120px; margin: 0 10px; }
.msg { flex: none; margin: 8px 10px 0; }
</style>
