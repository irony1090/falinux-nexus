<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useMemberBoard } from '@/feature/member/store/memberBoard.store';
import type { MemberTile } from '@/feature/member/util/memberGrid.util';
import { MAIN_KEY_PATTERN } from '@/feature/member/util/member.util';
import CtTextField from '@/feature/widget/component/CtTextField.vue';
import { createBundle, deleteBundle, patchBundle, useBundle, type Ref } from '../api/updater.api';
import { useUpdConfirm, useUpdRun } from '../hook/updActions.hook';
import { buildTileKey, commonTileKey, filesTileKey, runTileKey, shortSha, sizeText, stepsTileKey, timeText } from '../util/updGrid.util';
import UpdDeployDialog from './UpdDeployDialog.vue';
import UpdZipMenu from './UpdZipMenu.vue';
import { useUpdPrompt } from './UpdPrompt.vue';
import MbNote from '@/feature/member/component/MbNote.vue';

// 빌드 타일 — 상세 + 열기 줄 [Files | steps.json | Run · Log] + 삭제·새 버전·전체 받기·배포. REF-update-deploy.md "화면" 순서 2~4
// Script 버튼은 없음(안 A, 11-02) — 스크립트는 Files 목록에서 연다
const props = defineProps<{ tile: MemberTile; bundle: Ref }>();
const { open, close, dropMissing } = useMemberBoard();
const prompt = useUpdPrompt();
const confirm = useUpdConfirm();
const { busy, message, run } = useUpdRun();

const q = useBundle(props.bundle);
const b = computed(() => q.data.value);
watch(() => (q.error.value as { status?: number } | null)?.status, st => { if (st === 404) dropMissing(props.tile.key); }, { immediate: true });

const counts = computed(() => {
    const files = b.value?.files ?? [];
    const common = (b.value?.uses ?? []).reduce((n, u) => n + u.files.length, 0);
    return { common, main: files.filter(f => f.scope === 'main').length, sub: files.filter(f => f.scope === 'sub').length };
});
const size = computed(() => (b.value?.files ?? []).reduce((n, f) => n + f.size, 0) + (b.value?.uses ?? []).reduce((n, u) => n + u.files.reduce((m, f) => m + f.size, 0), 0));

const note = ref<string>();
const saveNote = () => {
    const v = (note.value ?? '').trim();
    if (note.value === undefined || v === (b.value?.note ?? '')) { note.value = undefined; return; }
    run(() => patchBundle(props.bundle, v), () => note.value = undefined);
}

const openHere = (key: string) => open(key, props.tile.key);
const deploying = ref(false);   // 배포 대상 팝업

const newVersion = async () => {
    const v = await prompt({
        title: `${props.bundle.owner}/${props.bundle.version} 복사해 새 버전`, ok: '만들기',
        note: '파일·steps.json·uses를 그대로 복사합니다. 잠긴 빌드도 복사할 수 있습니다',
        fields: [{ key: 'version', label: '새 버전', placeholder: '예: 261104_a', pattern: MAIN_KEY_PATTERN, patternMsg: '영문·숫자·. _ - 로 64자까지' }],
    });
    if (!v) return;
    await run(() => createBundle('build', props.bundle.owner, v.version!, '', { owner: props.bundle.owner, version: props.bundle.version }),
        () => openHere(buildTileKey(props.bundle.owner, v.version!)));
}

const remove = async () => {
    const lockedMsg = b.value?.locked ? '잠긴 빌드입니다. ' : '';
    if (!(await confirm(`${props.bundle.owner}/${props.bundle.version} 삭제`, [
        `${lockedMsg}서버의 파일이 지워져 다시 배포할 수 없습니다.`,
        '배포 이력(이름·sha)은 남고, 장비에 깔린 설치본·백업은 그대로입니다.',
    ], '삭제', 'error'))) return;
    await run(() => deleteBundle(props.bundle), () => close(props.tile.key));
}
</script>

<template>
<div class="mb-body" data-pager-press>
    <p v-if="q.isError.value" class="mb-empty">빌드를 찾을 수 없습니다(지워졌을 수 있습니다)</p>
    <div v-else-if="b" class="upd-fields">
        <span class="mb-lbl">장비</span>
        <span class="val">{{ b.owner }}</span>
        <span class="mb-lbl">올린 때</span>
        <span class="val">{{ timeText(b.createdAt) }} · {{ b.createdByName || '—' }}</span>
        <span class="mb-lbl">sha</span>
        <span class="val mono" :title="b.sha">{{ shortSha(b.sha) }}</span>
        <span class="mb-lbl">상태</span>
        <span class="val"><template v-if="b.updatedAt !== b.createdAt">고침 {{ timeText(b.updatedAt) }} </template><span class="mb-tag" :class="b.locked ? 'mb-tag--primary' : 'dim'">{{ b.locked ? '잠김' : '수정 가능' }}</span></span>
        <span class="mb-lbl">설명</span>
        <CtTextField class="edit" :model-value="note ?? b.note" size="sm" wall="r" hide-details maxlength="500" placeholder="설명 달기(잠긴 뒤에도 고칠 수 있음)" aria-label="설명" :disabled="busy"
            @update:model-value="(v: string) => note = v"
            @keydown.enter="($event.target as HTMLInputElement).blur()"
            @update:focused="(f: boolean) => !f && saveNote()"
        />
        <span class="mb-lbl">uses</span>
        <span v-if="b.uses.length" class="upd-set">
            <span class="ct-grp ct-grp--sm ct-grp--wrap">
                <button v-for="u in b.uses" :key="u.name" type="button" class="ct-seg ct-seg--accent ct-seg--mono" :title="`${u.name}@${u.version} 공용 묶음 열기`"
                    @click="openHere(commonTileKey(u.name, u.version))">{{ u.name }}@{{ u.version }}</button>
            </span>
        </span>
        <span v-else class="val dim">없음</span>
        <span class="mb-lbl">파일</span>
        <span class="val">공용 {{ counts.common }} · 메인키 {{ counts.main }} · 서브키 {{ counts.sub }} · {{ sizeText(size) }}</span>
        <span class="mb-lbl">열기</span>
        <span class="upd-set">
            <span class="ct-grp ct-grp--sm ct-grp--wrap">
                <button type="button" class="ct-seg ct-seg--accent" @click="openHere(filesTileKey(b.owner, b.version))">Files</button>
                <button type="button" class="ct-seg ct-seg--accent" @click="openHere(stepsTileKey(b.owner, b.version))">steps.json</button>
                <button type="button" class="ct-seg ct-seg--accent" :disabled="!b.lastDeployId" :title="b.lastDeployId ? `마지막 배포 #${b.lastDeployId}` : '배포한 적이 없습니다'"
                    @click="openHere(runTileKey(b.lastDeployId))">Run · Log</button>
            </span>
        </span>
        <p v-if="message" class="mb-err full">{{ message }}</p>
    </div>
</div>
<div v-if="b" class="mb-foot">
    <mb-note :text="b.locked ? '잠겨 있어 파일을 고칠 수 없습니다. 복사해 새 버전을 만듭니다' : '잠그기 전에는 수정할 수 있습니다'" />
    <div class="mb-end">
        <button type="button" class="mb-btn mb-btn--danger" :disabled="busy" @click="remove">삭제</button>
        <button type="button" class="mb-btn" :disabled="busy" @click="newVersion">새 버전</button>
        <upd-zip-menu :bundle="bundle" :detail="b" />
        <button type="button" class="mb-btn mb-btn--primary" :disabled="busy" @click="deploying = true">배포</button>
    </div>
</div>
<upd-deploy-dialog v-if="bundle.kind === 'build'" v-model="deploying" :bundle="bundle" />
</template>

<style scoped lang="scss">
.dim { opacity: 0.6; }
.val .mb-tag { vertical-align: middle; }
</style>
