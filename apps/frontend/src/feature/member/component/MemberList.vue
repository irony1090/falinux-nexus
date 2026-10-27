<script setup lang="ts">
import { PERM_ADMIN, PERM_WORKER_REGISTER } from '@/feature/user/api/user.api';
import { useAuthStore } from '@/feature/user/store/auth.store';
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { VIcon } from 'vuetify/components';
import CtTextField from '@/feature/widget/component/CtTextField.vue';
import { useAdminDevices, useAdminUsers, useMyDevices, type AdminUserResponse, type DeviceResponse } from '../api/device.api';
import { useMemberBoard } from '../store/memberBoard.store';
import { deviceKey, userKey } from '../util/memberGrid.util';
import { dayText, deviceState, deviceStateText, pageNumbers } from '../util/member.util';

// 검색어는 입력이 멈추고 이만큼 뒤에 보낸다(주소-2(replace))
const Q_DELAY = 500;
const SIZES = [10, 20, 50];

// 목록 타일 — 검색·필터 알약·줄 수·페이지 번호. 관리자 = 회원/장비 전환(머리 줄 왼쪽 벽) — REF-member-ui.md 화-3(목록)·화-4(서버 목록)
const { admin, tab, q, filter, pageNo, size, pickTab, tiles, layout, open } = useMemberBoard();
const { auth } = useAuthStore();

const params = computed(() => ({ q: q.value, f: filter.value, page: pageNo.value, size: size.value }));
const usersTab = computed(() => admin && tab.value === 'users');
const usersQuery = admin ? useAdminUsers(() => usersTab.value ? params.value : { size: 1 }) : undefined;
const devicesQuery = admin
    ? useAdminDevices(() => usersTab.value ? { size: 1 } : params.value)
    : useMyDevices(params);
// 내 장비 개수(검색과 무관) — 머리 줄 "N / 상한"
const myAllQuery = admin ? undefined : useMyDevices({ size: 1 });

const users = computed(() => usersQuery?.data.value);
const devices = computed(() => devicesQuery.data.value);
const current = computed(() => usersTab.value ? users.value : devices.value);
const items = computed(() => current.value?.items ?? []);
const total = computed(() => current.value?.total ?? 0);
const pages = computed(() => Math.max(1, Math.ceil(total.value / size.value)));
// 서버가 범위 안으로 당긴 페이지를 따른다(지우고 나서 끝 페이지가 사라진 경우)
watch(() => current.value?.page, p => { if (p && p !== pageNo.value) pageNo.value = p; });

const tabs = computed(() => [
    { key: 'users' as const, label: '회원', n: usersTab.value ? (users.value?.counts.all ?? 0) : (users.value?.total ?? 0) },
    { key: 'devices' as const, label: '장비', n: usersTab.value ? (devices.value?.total ?? 0) : (devices.value?.counts.all ?? 0) },
]);
const myCount = computed(() => myAllQuery?.data.value?.total ?? 0);
const myLimit = computed(() => auth.value?.workerLimit ?? 0);
const canRegister = computed(() => !!auth.value?.perms?.includes(PERM_WORKER_REGISTER));
const addOff = computed(() => !admin && (!canRegister.value || myCount.value >= myLimit.value));
const addTitle = computed(() => admin ? '상한·권한과 상관없이 등록'
    : !canRegister.value ? '장비 등록 권한이 없습니다' : myCount.value >= myLimit.value ? `${myLimit.value}개까지 등록할 수 있습니다` : '장비 이름 등록');

const chips = computed(() => {
    const defs = usersTab.value
        ? [['all', '전체'], ['admin', '관리자'], ['suspended', '정지'], ['noreg', '등록 막힘']]
        : [['all', '전체'], ['on', '접속 중'], ['off', '꺼짐'], ['none', '기록 없음']];
    return defs.map(([key, label]) => ({ key: key!, label: label!, n: current.value?.counts[key!] ?? 0 }));
});
const pickFilter = (f: string) => {
    filter.value = f;
    pageNo.value = 1;
}
const pickSize = (n: number) => {
    size.value = n;
    pageNo.value = 1;
}

// 검색 입력 — 칸은 바로, 서버 요청은 멈춘 뒤
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
const userTags = (u: AdminUserResponse) => [
    u.id === auth.value?.id && { text: '나', cls: '' },
    u.perms.includes(PERM_ADMIN) && { text: '관리자', cls: 'mb-tag--primary' },
    u.suspendedAt && { text: '정지', cls: 'mb-tag--error' },
    !u.perms.includes(PERM_WORKER_REGISTER) && { text: '등록 막힘', cls: 'dim' },
].filter(t => !!t);
const deviceSub = (d: DeviceResponse) => {
    const labels = d.instances.map(i => i.label).filter(Boolean);
    return [admin && `주인 ${d.ownerNickname ?? d.ownerId}`, labels.length ? labels.join(', ') : `인스턴스 ${d.instances.length}`].filter(Boolean).join(' · ');
}
const stateCls = (d: DeviceResponse) => ({ on: 'mb-tag--success', off: '', none: 'dim' })[deviceState(d.instances)];
const loading = computed(() => (usersTab.value ? usersQuery?.isPending.value : devicesQuery.isPending.value) ?? false);
</script>

<template>
<section class="MemberList mb-tile" aria-label="목록">
    <div class="mb-head">
        <div v-if="admin" class="mb-start" role="group" aria-label="목록 종류">
            <button v-for="t in tabs" :key="t.key" type="button" class="mb-btn" :class="{ 'mb-btn--on': tab === t.key }"
                :aria-pressed="tab === t.key" @click="pickTab(t.key)"
            >{{ t.label }} <span class="n">{{ t.n }}</span></button>
        </div>
        <h2 v-else class="title">내 장비 <span class="n">{{ myCount }} / {{ myLimit }}</span></h2>
        <span class="grow" />
        <div v-if="!usersTab" class="mb-end">
            <button type="button" class="mb-btn mb-btn--primary add" :disabled="addOff" :title="addTitle" @click="open('new')">
                <v-icon icon="mdi-plus" size="16" />{{ admin ? '대신 등록' : '장비 등록' }}
            </button>
        </div>
    </div>

    <div class="tools">
        <CtTextField v-model="qInput" size="sm" type="search" prepend-inner-icon="mdi-magnify" hide-details :placeholder="usersTab ? '닉네임 검색' : admin ? '이름·주인·라벨 검색' : '이름·라벨 검색'" aria-label="검색" />
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
    </div>

    <div class="mb-body">
        <template v-if="usersTab">
            <button v-for="u in (items as AdminUserResponse[])" :key="u.id" type="button" class="mb-row" :class="{ open: openKeys.has(userKey(u.id)) }" @click="open(userKey(u.id))">
                <span class="txt">
                    <span class="main">{{ u.nickname }}</span>
                    <span class="sub">장비 {{ u.deviceCount }} / {{ u.workerLimit }} · 가입 {{ dayText(u.createdAt) }}</span>
                </span>
                <span class="tags">
                    <span v-for="t in userTags(u)" :key="t.text" class="mb-tag wall" :class="t.cls">{{ t.text }}</span>
                </span>
            </button>
        </template>
        <template v-else>
            <button v-for="d in (items as DeviceResponse[])" :key="d.mainKey" type="button" class="mb-row" :class="{ open: openKeys.has(deviceKey(d.mainKey)) }" @click="open(deviceKey(d.mainKey))">
                <span class="mb-dot" :class="deviceState(d.instances)" />
                <span class="txt">
                    <span class="main">{{ d.mainKey }}</span>
                    <span class="sub">{{ deviceSub(d) }}</span>
                </span>
                <span class="mb-tag wall" :class="stateCls(d)">{{ deviceStateText(d.instances) }}</span>
            </button>
        </template>
        <p v-if="!loading && !items.length" class="mb-empty">{{ q || filter !== 'all' ? '조건에 맞는 항목이 없습니다' : usersTab ? '회원이 없습니다' : '등록된 장비가 없습니다' }}</p>
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
.MemberList { height: 100%; }
.n { @include vars.font('label-medium'); opacity: 0.8; }
.title {
    margin: 0;
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 12px;
    @include vars.font('title-small');
    font-weight: vars.$weight-md;
    .n { color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity)); opacity: 1; }
}
.grow { flex: 1; }
.add { gap: 2px; padding: 0 14px 0 16px !important; @include vars.font('label-medium'); font-weight: vars.$weight-md; }
.tools {
    flex: none;
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 10px;   // 좌우 = 위 여백
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}
// 필터 묶음 = 왼쪽 끝 곧게 / 한 페이지 묶음 = 오른쪽 끝 곧게(줄 좌우 여백 10px, REF-ui-chaltteok.md 시안 A′)
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
