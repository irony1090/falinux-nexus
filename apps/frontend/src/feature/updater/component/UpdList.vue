<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { VIcon } from 'vuetify/components';
import { useMemberBoard } from '@/feature/member/store/memberBoard.store';
import { MAIN_KEY_PATTERN, pageNumbers } from '@/feature/member/util/member.util';
import CtTextField from '@/feature/widget/component/CtTextField.vue';
import { createBundle, createUpdDevice, useCommons, useUpdDevices, type CommonGroup, type UpdDevice } from '../api/updater.api';
import { useUpdRun } from '../hook/updActions.hook';
import { commonTileKey, devTileKey, timeText } from '../util/updGrid.util';
import { useUpdPrompt } from './UpdPrompt.vue';

// 검색어는 입력이 멈추고 이만큼 뒤에 보낸다(MemberList와 같음)
const Q_DELAY = 500;
const SIZES = [10, 20, 50];

// updater 목록 — [장비 | 공용 묶음](머리 줄 왼쪽 벽) · 검색 · 필터 · 페이지. 행을 누르면 타일 — REF-update-deploy.md "화면" 순서 1
const { tab, q, filter, pageNo, size, pickTab, tiles, layout, open } = useMemberBoard();
const prompt = useUpdPrompt();
const { busy, message, run } = useUpdRun();

const params = computed(() => ({ q: q.value, f: filter.value, page: pageNo.value, size: size.value }));
const commonsTab = computed(() => tab.value === 'commons');
const devQuery = useUpdDevices(() => commonsTab.value ? { size: 1 } : params.value);
const comQuery = useCommons(() => commonsTab.value ? params.value : { size: 1 });

const current = computed(() => commonsTab.value ? comQuery.data.value : devQuery.data.value);
const items = computed(() => current.value?.items ?? []);
const total = computed(() => current.value?.total ?? 0);
const pages = computed(() => Math.max(1, Math.ceil(total.value / size.value)));
watch(() => current.value?.page, p => { if (p && p !== pageNo.value) pageNo.value = p; });

const tabs = computed(() => [
    { key: 'devices' as const, label: '장비', n: commonsTab.value ? (devQuery.data.value?.total ?? 0) : (devQuery.data.value?.counts.all ?? 0) },
    { key: 'commons' as const, label: '공용 묶음', n: commonsTab.value ? (comQuery.data.value?.counts.all ?? 0) : (comQuery.data.value?.total ?? 0) },
]);
const chips = computed(() => {
    const defs = commonsTab.value ? [['all', '전체'], ['unused', '안 쓰임']] : [['all', '전체'], ['nobuild', '빌드 없음']];
    return defs.map(([key, label]) => ({ key: key!, label: label!, n: current.value?.counts[key!] ?? 0 }));
});
const pickFilter = (f: string) => { filter.value = f; pageNo.value = 1; };
const pickSize = (n: number) => { size.value = n; pageNo.value = 1; };

const qInput = ref(q.value);
let qTimer: ReturnType<typeof setTimeout> | undefined;
watch(qInput, v => {
    clearTimeout(qTimer);
    qTimer = setTimeout(() => {
        if (v.trim() === q.value) return;
        q.value = v.trim();
        pageNo.value = 1;
    }, Q_DELAY);
});
watch(q, v => { if (v !== qInput.value.trim()) qInput.value = v; });
onBeforeUnmount(() => clearTimeout(qTimer));

const range = computed(() => {
    if (!total.value) return '0건';
    const lo = (pageNo.value - 1) * size.value + 1;
    return `${lo}–${Math.min(pageNo.value * size.value, total.value)} / ${total.value}`;
});
const nums = computed(() => pageNumbers(pageNo.value, pages.value, !layout.value.phone));
const goPage = (p: number) => pageNo.value = Math.max(1, Math.min(pages.value, p));
const openKeys = computed(() => new Set(tiles.value.map(t => t.key)));

const devSub = (d: UpdDevice) => [`빌드 ${d.buildCount}`, `인스턴스 ${d.instanceCount}`, d.memo].filter(Boolean).join(' · ');
const loading = computed(() => (commonsTab.value ? comQuery.isPending.value : devQuery.isPending.value));

const NAME_RULE = { pattern: MAIN_KEY_PATTERN, patternMsg: '영문·숫자·. _ - 로 64자까지' };
// P1(보관·화면)은 updater가 없어 장비 루트를 손으로 만든다(P2부터 첫 접속 때 자동)
const addDevice = async () => {
    const v = await prompt({
        title: '장비 추가', ok: '추가', note: 'updater가 접속하면 이 메인키로 자동으로 붙습니다',
        fields: [{ key: 'mainKey', label: '메인키', placeholder: '예: fa24-PID', ...NAME_RULE }, { key: 'memo', label: '메모', placeholder: '선택', optional: true }],
    });
    if (!v) return;
    await run(() => createUpdDevice(v.mainKey!, v.memo), () => open(devTileKey(v.mainKey!)));
}
const addCommon = async () => {
    const v = await prompt({
        title: '공용 묶음 만들기', ok: '만들기', note: '여러 빌드가 steps.json의 uses로 함께 쓰는 파일 묶음입니다',
        fields: [{ key: 'name', label: '이름', placeholder: '예: nexus-worker', ...NAME_RULE }, { key: 'version', label: '버전', placeholder: '예: 261103_a', ...NAME_RULE }],
    });
    if (!v) return;
    await run(() => createBundle('common', v.name!, v.version!), () => open(commonTileKey(v.name!, v.version!)));
}
const versionSub = (g: CommonGroup, i: number) => {
    const b = g.versions[i]!;
    return [`${b.fileCount}개 파일`, b.usedBy ? `사용 중 ${b.usedBy}` : '안 쓰임', timeText(b.updatedAt)].join(' · ');
}
</script>

<template>
<section class="UpdList mb-tile" aria-label="목록">
    <div class="mb-head">
        <div class="mb-start" role="group" aria-label="목록 종류">
            <button v-for="t in tabs" :key="t.key" type="button" class="mb-btn" :class="{ 'mb-btn--on': tab === t.key }"
                :aria-pressed="tab === t.key" @click="pickTab(t.key)"
            >{{ t.label }} <span class="n">{{ t.n }}</span></button>
        </div>
        <span class="grow" />
        <div class="mb-end">
            <button type="button" class="mb-btn mb-btn--primary add" :disabled="busy" @click="commonsTab ? addCommon() : addDevice()">
                <v-icon icon="mdi-plus" size="16" />{{ commonsTab ? '공용 묶음' : '장비' }}
            </button>
        </div>
    </div>

    <div class="tools">
        <CtTextField v-model="qInput" size="sm" type="search" prepend-inner-icon="mdi-magnify" hide-details :placeholder="commonsTab ? '묶음 이름 검색' : '메인키·메모 검색'" aria-label="검색" />
        <div class="mb-grp mb-grp--sm chips" role="group" aria-label="필터">
            <button v-for="c in chips" :key="c.key" type="button" class="mb-btn" :class="{ 'mb-btn--on': filter === c.key }"
                :aria-pressed="filter === c.key" @click="pickFilter(c.key)"
            >{{ c.label }} <span class="n">{{ c.n }}</span></button>
        </div>
        <div class="per">
            <span class="range">{{ range }}</span>
            <span class="lbl">한 페이지</span>
            <div class="mb-grp mb-grp--sm sizes" role="group" aria-label="한 페이지 줄 수">
                <button v-for="n in SIZES" :key="n" type="button" class="mb-btn" :class="{ 'mb-btn--on': size === n }" :aria-pressed="size === n" @click="pickSize(n)">{{ n }}</button>
            </div>
        </div>
        <p v-if="message" class="mb-err">{{ message }}</p>
    </div>

    <div class="mb-body">
        <template v-if="commonsTab">
            <template v-for="g in (items as CommonGroup[])" :key="g.name">
                <h3 class="mb-sec grp">{{ g.name }} <span class="n">버전 {{ g.versions.length }}</span></h3>
                <button v-for="(b, i) in g.versions" :key="b.version" type="button" class="mb-row" :class="{ open: openKeys.has(commonTileKey(g.name, b.version)) }" @click="open(commonTileKey(g.name, b.version))">
                    <span class="txt">
                        <span class="main upd-mono">{{ b.version }}</span>
                        <span class="sub">{{ versionSub(g, i) }}</span>
                    </span>
                    <span class="mb-tag wall" :class="b.locked ? 'mb-tag--primary' : 'dim'">{{ b.locked ? '잠김' : '수정 가능' }}</span>
                </button>
            </template>
        </template>
        <template v-else>
            <button v-for="d in (items as UpdDevice[])" :key="d.mainKey" type="button" class="mb-row" :class="{ open: openKeys.has(devTileKey(d.mainKey)) }" @click="open(devTileKey(d.mainKey))">
                <span class="mb-dot" :class="d.instanceCount ? 'off' : ''" />
                <span class="txt">
                    <span class="main">{{ d.mainKey }}</span>
                    <span class="sub">{{ devSub(d) }}</span>
                </span>
                <span class="mb-tag wall dim">{{ d.lastBuildAt ? timeText(d.lastBuildAt) : '빌드 없음' }}</span>
            </button>
        </template>
        <p v-if="!loading && !items.length" class="mb-empty">{{ q || filter !== 'all' ? '조건에 맞는 항목이 없습니다' : commonsTab ? '공용 묶음이 없습니다' : '장비가 없습니다. 오른쪽 위 + 장비로 메인키를 추가합니다' }}</p>
    </div>

    <nav class="pager" aria-label="페이지">
        <div class="mb-end">
            <button type="button" class="mb-btn arrow" :disabled="pageNo <= 1" aria-label="이전 페이지" @click="goPage(pageNo - 1)"><v-icon icon="mdi-chevron-left" size="18" /></button>
            <template v-for="(p, i) in nums" :key="i">
                <span v-if="p === null" class="gap" aria-hidden="true">…</span>
                <button v-else type="button" class="mb-btn num" :class="{ 'mb-btn--on': p === pageNo }" :aria-current="p === pageNo ? 'page' : undefined" @click="goPage(p)">{{ p }}</button>
            </template>
            <button type="button" class="mb-btn arrow" :disabled="pageNo >= pages" aria-label="다음 페이지" @click="goPage(pageNo + 1)"><v-icon icon="mdi-chevron-right" size="18" /></button>
        </div>
    </nav>
</section>
</template>

<style scoped lang="scss">
// 모양은 MemberList와 같다(목록 칸 공용 규칙)
.UpdList { height: 100%; }
.n { @include vars.font('label-medium'); opacity: 0.8; }
.grow { flex: 1; }
.add { gap: 2px; padding: 0 14px 0 16px !important; @include vars.font('label-medium'); font-weight: vars.$weight-md; }
.tools {
    flex: none;
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 10px;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}
.chips {
    align-self: flex-start;
    border-radius: 0 999px 999px 0;
    > .mb-btn { padding: 0 12px; }
    > .mb-btn:first-child { padding-left: 12px; }
    > .mb-btn:last-child { padding-right: 14px; }
}
.per {
    display: flex;
    align-items: center;
    gap: 8px;
    @include vars.font('body-small');
    color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity));
    .range { flex: 1; min-width: 0; padding-left: 2px; }
    .sizes { height: 26px; border-radius: 999px 0 0 999px; > .mb-btn { min-width: 34px; padding: 0 8px; } > .mb-btn:first-child { padding-left: 12px; } > .mb-btn:last-child { padding-right: 10px; } }
}
.grp { display: flex; align-items: baseline; gap: 8px; color: rgb(var(--v-theme-on-surface)); }
.pager {
    flex: none;
    height: 36px;
    display: flex;
    justify-content: flex-end;
    border-top: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    .mb-end { background: rgba(var(--v-theme-on-surface), 0.07); }
    .mb-btn { background: transparent; box-shadow: none; @include vars.font('label-medium'); font-weight: vars.$weight-md; }
    .mb-btn--on { background: rgb(var(--v-theme-primary)); }
    .num { min-width: 32px; padding: 0 6px; }
    .arrow { min-width: 36px; padding: 0 4px !important; }
    .arrow:first-child { padding-left: 8px !important; }
    .gap { min-width: 20px; display: flex; align-items: center; justify-content: center; @include vars.font('label-medium'); opacity: 0.6; }
}
</style>
