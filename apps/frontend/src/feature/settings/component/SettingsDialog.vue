<script setup lang="ts">
import type { TileBarAt } from '@/feature/widget/util/tileBar.util';
import AccountPanel from '@/feature/user/component/AccountPanel.vue';
import { ref, watch } from 'vue';
import { accountSettings, browserSettings, setAccountSetting, setBrowserSetting, settingsDialog as open } from '../store/settings.store';

// 설정 창 — 묶음 두 개: 계정 설정(모든 기기) / 이 브라우저에서만(Q24(창 모양)). 고르면 바로 적용, 버튼은 [닫기]만(Q23(설정 창 적용)) — REF-settings-footer.md
const BAR_AT: { value: TileBarAt; text: string }[] = [
    { value: 'top', text: '위' },
    { value: 'bottom', text: '아래' },
];
// undefined = 계정 설정 따름(브라우저 값 지움)
const BAR_AT_HERE: { value: TileBarAt | undefined; text: string }[] = [{ value: undefined, text: '계정 따름' }, ...BAR_AT];

// 탭 = 설정 | 계정(REF-account.md 세션-5(설정 화면)). 열 때마다 설정 탭부터
const TABS = [{ key: 'settings', text: '설정' }, { key: 'account', text: '계정' }] as const;
const tab = ref<(typeof TABS)[number]['key']>('settings');
watch(open, v => { if (v) tab.value = 'settings'; });
</script>

<template>
<v-dialog v-model="open" max-width="440" min-width="296" scrollable>
    <v-card class="ct-card">
        <div class="top">
            <v-card-title class="ttl">{{ tab === 'settings' ? '설정' : '계정' }}</v-card-title>
            <!-- 창 오른쪽 벽에 붙는 탭 묶음 — 찰떡(REF-design-chaltteok.md) -->
            <span class="ct-grp ct-grp--r pick" role="tablist" aria-label="설정 탭">
                <button v-for="t in TABS" :key="t.key" type="button" role="tab" class="ct-seg" :class="{ 'ct-seg--primary': tab === t.key }"
                    :aria-selected="tab === t.key" @click="tab = t.key">{{ t.text }}</button>
            </span>
        </div>
        <div class="body">
        <account-panel v-if="tab === 'account'" />
        <template v-else>
        <p class="sec">계정 설정<span class="sub">모든 기기에 적용</span></p>
        <div class="row">
            <div class="lb">
                <span class="name">메뉴 줄 위치</span>
                <span class="hint">경로 줄 같은 타일 메뉴 줄을 붙일 곳<template v-if="browserSettings.tileBarAt"> · 이 브라우저는 '이 브라우저에서만' 값을 씀</template></span>
            </div>
            <!-- 창 오른쪽 벽에 붙는 묶음 — 찰떡(REF-design-chaltteok.md) -->
            <span class="ct-grp ct-grp--r pick" role="group" aria-label="메뉴 줄 위치(계정)">
                <button v-for="o in BAR_AT" :key="o.value" type="button" class="ct-seg" :class="{ 'ct-seg--primary': accountSettings.tileBarAt === o.value }"
                    :aria-pressed="accountSettings.tileBarAt === o.value" @click="setAccountSetting('tileBarAt', o.value)">{{ o.text }}</button>
            </span>
        </div>
        <p class="sec">이 브라우저에서만<span class="sub">계정 설정보다 우선 · 로그아웃하면 지워짐</span></p>
        <div class="row">
            <div class="lb">
                <span class="name">메뉴 줄 위치</span>
            </div>
            <span class="ct-grp ct-grp--r pick" role="group" aria-label="메뉴 줄 위치(이 브라우저)">
                <button v-for="o in BAR_AT_HERE" :key="o.text" type="button" class="ct-seg" :class="{ 'ct-seg--primary': browserSettings.tileBarAt === o.value }"
                    :aria-pressed="browserSettings.tileBarAt === o.value" @click="setBrowserSetting('tileBarAt', o.value)">{{ o.text }}</button>
            </span>
        </div>
        </template>
        </div>
        <div class="ct-actions">
            <span class="ct-grp ct-grp--r">
                <button type="button" class="ct-seg" @click="open = false">닫기</button>
            </span>
        </div>
    </v-card>
</v-dialog>
</template>

<style scoped lang="scss">
.top { display: flex; align-items: center; }
.ttl { flex: 1; min-width: 0; }
.body { flex: 1 1 auto; min-height: 0; overflow-y: auto; }
// 왼쪽 여백 = v-card-title과 같은 16px, 오른쪽 여백 없음(고르기 묶음이 벽에 붙게)
.sec {
    margin: vars.$spacing-md 0 vars.$spacing-sm;
    padding: 0 vars.$spacing-xl;
    @include vars.font('label-medium');
    font-weight: vars.$weight-lg;
    opacity: 0.6;
    > .sub { margin-left: vars.$spacing-sm; font-weight: vars.$weight-md; }
}
.row {
    display: flex;
    align-items: center;
    gap: vars.$spacing-xl;
    padding: vars.$spacing-sm 0 vars.$spacing-md vars.$spacing-xl;
}
.lb {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
}
.name { @include vars.font('body-medium'); font-weight: vars.$weight-md; }
.hint { @include vars.font('body-small'); opacity: 0.7; }
.pick { height: 32px; }
</style>
