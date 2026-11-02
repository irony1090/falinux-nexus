<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useMemberBoard } from '@/feature/member/store/memberBoard.store';
import type { MemberTile } from '@/feature/member/util/memberGrid.util';
import { MAIN_KEY_PATTERN } from '@/feature/member/util/member.util';
import CtTextField from '@/feature/widget/component/CtTextField.vue';
import { copyText } from '@/feature/member/util/member.util';
import {
    createBundle, deleteInstance, deleteUpdDevice, patchInstance, patchUpdDevice, reissueUpdKey, showUpdKey, updaterConfigText, useUpdDevice,
    UPD_KEY_PREFIX, type Updater,
} from '../api/updater.api';
import { freshUpdKeys, useUpdConfirm, useUpdRun } from '../hook/updActions.hook';
import { buildTileKey, shortSha, timeText } from '../util/updGrid.util';
import { useUpdPrompt } from './UpdPrompt.vue';
import MbNote from '@/feature/member/component/MbNote.vue';

// 장비 타일(메인키) — 접속 키·updater.json · 인스턴스(별칭·메모·설치 경로) · 빌드 목록(빌드마다 열기 / Run·Log) · 새 빌드
// REF-update-deploy.md "화면" 순서 1~2 · REF-update-p2.md 단위-2(키·권한 화면)·단위-7(인스턴스 화면)
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

// ---- 키(정-1(접속 인증)) — 다시 보기 가능, 재발급하면 옛 키는 다음 접속부터 거절 ----
const key = ref<string | undefined>(freshUpdKeys.value[props.mainKey]);
if (key.value) { const { [props.mainKey]: _, ...rest } = freshUpdKeys.value; freshUpdKeys.value = rest; }
const copied = ref<'key' | 'json'>();
const toggleKey = () => key.value ? key.value = undefined : run(() => showUpdKey(props.mainKey), r => key.value = r.apiKey);
const copy = (what: 'key' | 'json') => run(async () => {
    const k = key.value ?? (await showUpdKey(props.mainKey)).apiKey;
    await copyText(what === 'key' ? k : updaterConfigText(props.mainKey, k));
    copied.value = what;
    setTimeout(() => copied.value = undefined, 1500);
});
const reissue = async () => {
    if (!(await confirm(`${props.mainKey} 키 ${d.value?.hasKey ? '재발급' : '발급'}`, ['새 키를 발급합니다.', '지금 붙어 있는 updater는 그대로지만, 다시 접속할 때는 updater.json에 새 키가 있어야 합니다.'], '발급'))) return;
    await run(() => reissueUpdKey(props.mainKey), r => key.value = r.apiKey);
}

// ---- 인스턴스 ----
const ABS_PATH = /^(\/|[A-Za-z]:[\\/])/;
const instSub = (i: Updater) => [i.subKey, i.os && `${i.os}/${i.arch}`, i.installDir ? `경로 ${i.installDir}` : '경로 없음', i.online && i.version, i.running && `실행 중 #${i.running}`].filter(Boolean).join(' · ');
const editInstance = async (i: Updater) => {
    const v = await prompt({
        title: `${props.mainKey}#${i.subKey} 설정`, ok: '저장',
        note: i.online ? (i.running ? '단계가 실행 중이라 설치 경로는 바꿀 수 없습니다' : '설치 경로를 바꾸면 updater가 설정 파일(updater.json)을 고쳐 씁니다. 되돌리기는 백업 때 경로로 복원합니다') : 'updater가 접속해 있지 않아 설치 경로는 바꿀 수 없습니다',
        fields: [
            { key: 'label', label: '별칭', value: i.label, optional: true, placeholder: '예: 3번 라인 지그' },
            { key: 'memo', label: '메모', value: i.memo, optional: true },
            ...(i.online && !i.running ? [{ key: 'targetDir', label: '설치 경로(TARGET_DIR)', value: i.installDir, optional: true, placeholder: '예: /opt/nexus-worker', pattern: ABS_PATH, patternMsg: '절대 경로로 적습니다' }] : []),
        ],
    });
    if (!v) return;
    const body: { label: string; memo: string; targetDir?: string } = { label: v.label ?? '', memo: v.memo ?? '' };
    if (v.targetDir !== undefined && v.targetDir !== i.installDir) body.targetDir = v.targetDir;
    await run(() => patchInstance(props.mainKey, i.subKey, body));
}
const removeInstance = async (i: Updater) => {
    if (!(await confirm(`${i.label || i.subKey} 지우기`, ['명부에서 지웁니다. 같은 updater가 다시 접속하면 새로 생깁니다.'], '지우기', 'error'))) return;
    await run(() => deleteInstance(props.mainKey, i.subKey));
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
    if (!(await confirm(`${props.mainKey} 삭제`, ['장비 루트와 접속 키를 지웁니다. 빌드가 남아 있으면 지울 수 없습니다.', '붙어 있던 updater는 끊기고 다시 접속하지 못합니다.'], '삭제', 'error'))) return;
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
            <span class="val">{{ timeText(d.createdAt) }}{{ d.keyIssuedAt ? ` · 키 발급 ${timeText(d.keyIssuedAt)}` : '' }}</span>
            <span class="mb-lbl">접속 키</span>
            <div v-if="d.hasKey" class="keyrow edit">
                <!-- 보기 = 고칠 수 없는 한 줄 칸(커서로 끝까지 이동·선택) / 숨김 = 가림 글자 — 회원 장비 타일 API 키와 같은 모양 -->
                <CtTextField v-if="key" :model-value="key" size="sm" wall="r" class="upd-mono code" hide-details inputmode="none" spellcheck="false" autocomplete="off" aria-readonly="true" aria-label="접속 키" @beforeinput.prevent @drop.prevent />
                <CtTextField v-else :model-value="`${UPD_KEY_PREFIX}${'•'.repeat(18)}`" size="sm" wall="r" class="upd-mono code hidden" hide-details readonly tabindex="-1" aria-label="접속 키(가림)" />
                <div class="ct-grp ct-grp--sm">
                    <button type="button" class="ct-seg" :disabled="busy" @click="toggleKey">{{ key ? '숨기기' : '보기' }}</button>
                    <button type="button" class="ct-seg" :disabled="busy" @click="copy('key')">{{ copied === 'key' ? '복사됨' : '복사' }}</button>
                    <button type="button" class="ct-seg ct-seg--accent" :disabled="busy" @click="reissue">재발급</button>
                </div>
            </div>
            <div v-else class="upd-set">
                <span class="ct-grp ct-grp--sm ct-grp--r"><button type="button" class="ct-seg ct-seg--primary" :disabled="busy" @click="reissue">키 발급</button></span>
            </div>
            <div v-if="key" class="full conf">
                <pre class="upd-mono">{{ updaterConfigText(mainKey, key) }}</pre>
                <button type="button" class="confcopy" @click="copy('json')">{{ copied === 'json' ? '복사됨' : 'updater.json 복사' }}</button>
            </div>
            <p v-if="message" class="mb-err full">{{ message }}</p>
        </div>
        <h3 class="mb-sec">인스턴스 {{ d.instances.length }} <span class="dim">· OS와 CPU는 updater가 알린 값입니다</span></h3>
        <div v-for="i in d.instances" :key="i.subKey" class="upd-row">
            <span class="mb-dot" :class="i.online ? 'on' : 'off'" />
            <span class="txt">
                <span class="main">{{ i.label || i.subKey }}<span v-if="i.memo" class="memo"> · {{ i.memo }}</span></span>
                <span class="sub" :title="instSub(i)">{{ instSub(i) }}</span>
            </span>
            <div class="ct-grp ct-grp--sm">
                <button type="button" class="ct-seg ct-seg--accent" :disabled="busy" @click="editInstance(i)">설정</button>
                <button v-if="!i.online" type="button" class="ct-seg ct-seg--error" :disabled="busy" @click="removeInstance(i)">지우기</button>
            </div>
        </div>
        <p v-if="!d.instances.length" class="upd-hint">updater.json(supervisor·메인키·키·설치 경로)을 넣고 updater를 실행하면 여기에 인스턴스가 나타납니다</p>
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
.memo { font-weight: vars.$weight-sm; opacity: 0.75; }
// 키 칸 + [보기|복사|재발급] = 한 덩어리, 오른쪽 벽에 붙음(회원 장비 타일 .keyrow와 같음)
.keyrow {
    min-width: 0;
    display: flex;
    align-items: center;
    .code { flex: 1; min-width: 0; &.hidden { opacity: 0.6; } }
    .ct-grp { height: 32px; border-radius: 0; }
}
.conf {
    position: relative;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.16);
    border-radius: 10px;
    background: rgba(var(--v-theme-on-surface), 0.035);
    overflow: hidden;
    pre { margin: 0; padding: 12px 14px; line-height: 1.6; white-space: pre; overflow-x: auto; @include vars.font('body-small'); }
}
// 상자 오른쪽 위 모서리에 붙는 끝 블록(찰떡) — 곡선은 상자가 자른다
.confcopy {
    position: absolute;
    top: 0;
    right: 0;
    height: 26px;
    padding: 0 12px 0 14px;
    border: 0;
    border-radius: 0 0 0 999px;
    background: rgb(var(--v-theme-primary));
    color: rgb(var(--v-theme-on-primary));
    font: inherit;
    @include vars.font('label-medium');
    font-weight: vars.$weight-md;
    cursor: pointer;
}
.build .main { display: flex; align-items: center; gap: 8px; }
.build .mb-tag { height: 18px; padding: 0 7px; @include vars.font('label-small'); }
</style>
