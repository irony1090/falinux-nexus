<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useMemberBoard } from '@/feature/member/store/memberBoard.store';
import type { MemberTile } from '@/feature/member/util/memberGrid.util';
import { MAIN_KEY_PATTERN } from '@/feature/member/util/member.util';
import CtTextField from '@/feature/widget/component/CtTextField.vue';
import { createBundle, deleteUpdDevice, patchUpdDevice, useUpdDevice } from '../api/updater.api';
import { useUpdConfirm, useUpdRun } from '../hook/updActions.hook';
import { buildTileKey, shortSha, timeText } from '../util/updGrid.util';
import { useUpdPrompt } from './UpdPrompt.vue';
import MbNote from '@/feature/member/component/MbNote.vue';

// 장비 타일(메인키) — 인스턴스(updater가 접속하면 자동, P1은 빈 자리) · 빌드 목록(빌드마다 열기 / Run·Log) · 새 빌드 — REF-update-deploy.md "화면" 순서 1~2
const props = defineProps<{ tile: MemberTile; mainKey: string }>();
const { open, close, dropMissing } = useMemberBoard();
const prompt = useUpdPrompt();
const confirm = useUpdConfirm();
const { busy, message, run } = useUpdRun();

const q = useUpdDevice(props.mainKey);
const d = computed(() => q.data.value);
watch(() => (q.error.value as { status?: number } | null)?.status, st => { if (st === 404) dropMissing(props.tile.key); }, { immediate: true });

// 메모 — 칸에서 바로 고치고 Enter·포커스 이동 때 저장
const memo = ref<string>();
const saveMemo = () => {
    const v = (memo.value ?? '').trim();
    if (memo.value === undefined || v === (d.value?.memo ?? '')) { memo.value = undefined; return; }
    run(() => patchUpdDevice(props.mainKey, v), () => memo.value = undefined);
}

const openBuild = (version: string) => open(buildTileKey(props.mainKey, version), props.tile.key);

// 새 빌드 = 빈 빌드 또는 기존 빌드 복사(잠긴 빌드는 고칠 수 없으니 복사가 주로 쓰는 길)
const addBuild = async () => {
    const builds = d.value?.builds ?? [];
    const v = await prompt({
        title: `${props.mainKey} 새 빌드`, ok: '만들기',
        note: '버전 이름은 이 장비 안에서만 겹치지 않으면 됩니다',
        fields: [
            { key: 'version', label: '버전', placeholder: '예: 261103_a', pattern: MAIN_KEY_PATTERN, patternMsg: '영문·숫자·. _ - 로 64자까지' },
            { key: 'from', label: '시작 내용', value: builds[0]?.version ?? '', items: [{ title: '빈 빌드', value: '' }, ...builds.map(b => ({ title: `${b.version} 복사${b.locked ? ' (잠김)' : ''}`, value: b.version }))] },
        ],
    });
    if (!v) return;
    const from = v.from ? { owner: props.mainKey, version: v.from } : undefined;
    await run(() => createBundle('build', props.mainKey, v.version!, '', from), () => openBuild(v.version!));
}

const remove = async () => {
    if (!(await confirm(`${props.mainKey} 삭제`, ['장비 루트를 지웁니다. 빌드가 남아 있으면 지울 수 없습니다.', 'updater가 다시 접속하면 다시 생깁니다.'], '삭제', 'error'))) return;
    await run(() => deleteUpdDevice(props.mainKey), () => close(props.tile.key));
}
</script>

<template>
<div class="mb-body" data-pager-press>
    <p v-if="q.isError.value" class="mb-empty">장비를 찾을 수 없습니다(지워졌을 수 있습니다)</p>
    <template v-else-if="d">
        <div class="upd-fields">
            <span class="mb-lbl">메모</span>
            <CtTextField class="edit" :model-value="memo ?? d.memo" size="sm" wall="r" hide-details maxlength="200" placeholder="메모 달기" aria-label="메모" :disabled="busy"
                @update:model-value="(v: string) => memo = v"
                @keydown.enter="($event.target as HTMLInputElement).blur()"
                @update:focused="(f: boolean) => !f && saveMemo()"
            />
            <span class="mb-lbl">만든 때</span>
            <span class="val">{{ timeText(d.createdAt) }}</span>
            <p v-if="message" class="mb-err full">{{ message }}</p>
        </div>
        <h3 class="mb-sec">인스턴스 {{ d.instances.length }} <span class="dim">· OS와 CPU는 표시용 정보입니다</span></h3>
        <div v-for="i in d.instances" :key="i.subKey" class="upd-row">
            <span class="mb-dot" :class="i.online ? 'on' : 'off'" />
            <span class="txt">
                <span class="main">{{ i.label || i.subKey }}</span>
                <span class="sub">{{ [i.os && `${i.os}/${i.arch}`, i.installedVersion && `설치본 ${i.installedVersion}`].filter(Boolean).join(' · ') }}</span>
            </span>
        </div>
        <p v-if="!d.instances.length" class="upd-hint">인스턴스는 updater가 접속하면 자동으로 보입니다</p>
        <h3 class="mb-sec">빌드 {{ d.builds.length }} <span class="dim">· 이름은 이 장비 안에서만 고유하면 됩니다</span></h3>
        <div v-for="b in d.builds" :key="b.version" class="upd-row build">
            <span class="txt">
                <span class="main upd-mono">{{ b.version }} <span class="mb-tag" :class="b.locked ? 'mb-tag--primary' : 'dim'">{{ b.locked ? '잠김' : '수정 가능' }}</span></span>
                <span class="sub">{{ b.fileCount }}개 파일 · sha {{ shortSha(b.sha) }} · {{ timeText(b.updatedAt) }}</span>
            </span>
            <div class="ct-grp ct-grp--sm">
                <button type="button" class="ct-seg ct-seg--accent" @click="openBuild(b.version)">열기</button>
                <button type="button" class="ct-seg ct-seg--accent" disabled title="배포 단계(P3)에서 엽니다">Run · Log</button>
            </div>
        </div>
        <p v-if="!d.builds.length" class="mb-empty">빌드가 없습니다. 아래 + 새 빌드로 만듭니다</p>
    </template>
</div>
<div v-if="d" class="mb-foot">
    <mb-note text="빌드를 열면 파일·steps.json을 고칠 수 있습니다" />
    <div class="mb-end">
        <button type="button" class="mb-btn mb-btn--danger" :disabled="busy || d.builds.length > 0" :title="d.builds.length ? '빌드를 먼저 지웁니다' : undefined" @click="remove">삭제</button>
        <button type="button" class="mb-btn mb-btn--primary" :disabled="busy" @click="addBuild">+ 새 빌드</button>
    </div>
</div>
</template>

<style scoped lang="scss">
.dim { font-weight: vars.$weight-md; opacity: 0.75; }
.build .main { display: flex; align-items: center; gap: 8px; }
.build .mb-tag { height: 18px; padding: 0 7px; @include vars.font('label-small'); }
</style>
