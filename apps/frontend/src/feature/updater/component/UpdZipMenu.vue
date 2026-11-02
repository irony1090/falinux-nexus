<script setup lang="ts">
import { computed, ref } from 'vue';
import { VCard, VMenu } from 'vuetify/components';
import { openDownload, zipUrl, type BundleDetail, type Ref } from '../api/updater.api';

// 전체 받기 — 기본 = 범위별 그대로, 선택 = 합친 결과(updater가 받는 최종 모양). 서버가 만들면서 바로 보낸다 — REF-update-deploy.md "릴리스 보관 구조"
const props = defineProps<{ bundle: Ref; detail?: BundleDetail }>();
const menu = ref(false);

const groups = computed(() => {
    const d = props.detail;
    if (!d) return [];
    return [
        ...d.uses.map(u => ({ name: `common/${u.name}@${u.version}/`, n: u.files.length })),
        { name: 'main/', n: d.files.filter(f => f.scope === 'main').length },
        { name: 'sub/', n: d.files.filter(f => f.scope === 'sub').length },
    ];
});
const merged = computed(() => {
    const d = props.detail;
    if (!d) return 0;
    return new Set([...d.uses.flatMap(u => u.files.map(f => f.path)), ...d.files.filter(f => f.scope === 'main').map(f => f.path)]).size;
});
const get = (m: boolean) => {
    menu.value = false;
    openDownload(zipUrl(props.bundle, m));
};
</script>

<template>
<button v-if="bundle.kind === 'common'" type="button" class="mb-btn" @click="get(false)">전체 받기</button>
<v-menu v-else v-model="menu" location="top end" offset="6" :close-on-content-click="false">
    <template #activator="{ props: act }">
        <button v-bind="act" type="button" class="mb-btn">전체 받기</button>
    </template>
    <v-card class="UpdZipMenu ct-card">
        <p class="head">{{ bundle.owner }} / {{ bundle.version }} — zip 하나로 받습니다</p>
        <button type="button" class="opt" @click="get(false)">
            <span class="t">범위별로 그대로</span>
            <span v-for="g in groups" :key="g.name" class="ln"><code>{{ g.name }}</code><span>{{ g.n ? `${g.n}개` : '비어 있음' }}</span></span>
            <span class="d">공용·메인키·서브키가 나뉜 채로 남습니다</span>
        </button>
        <button type="button" class="opt" @click="get(true)">
            <span class="t">합친 결과 <span class="n">{{ merged }}개</span></span>
            <span class="d">공용을 먼저 깔고 메인키로 덮어쓴 최종 모양입니다. 같은 이름은 뒤의 것이 이깁니다</span>
        </button>
    </v-card>
</v-menu>
</template>

<style scoped lang="scss">
.UpdZipMenu { width: 300px; padding: 6px; display: flex; flex-direction: column; gap: 4px; }
.head { margin: 4px 8px 2px; @include vars.font('label-medium'); color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity)); }
.opt {
    display: flex;
    flex-direction: column;
    gap: 3px;
    padding: 10px 12px;
    border: 0;
    border-radius: 10px;
    background: rgba(var(--v-theme-on-surface), 0.04);
    color: inherit;
    font: inherit;
    text-align: left;
    cursor: pointer;
    &:hover { background: rgba(var(--v-theme-primary), 0.1); }
    .t { @include vars.font('label-large'); font-weight: vars.$weight-lg; }
    .n { @include vars.font('label-medium'); opacity: 0.7; }
    .ln { display: flex; justify-content: space-between; gap: 8px; @include vars.font('body-small'); code { font-size: 11px; } }
    .d { @include vars.font('body-small'); color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity)); }
}
</style>
