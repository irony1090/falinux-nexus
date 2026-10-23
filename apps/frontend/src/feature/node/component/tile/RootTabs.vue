<script setup lang="ts">
import { computed } from 'vue';
import { useRootRuns } from '../../hook/scriptRuns.hook';
import { useTileGrids } from '../../store/tileGrids.store';

// 루트 타일 탭 `[파일] [터미널 N · 최소화 n]` — 폴더·장비 루트 공용. 숫자 = 터미널 탭 목록(RootTerminalList와 같은 조회)
const { rootTab, overview } = useTileGrids();
const runs = useRootRuns();
const minCount = computed(() => runs.value.filter(r => r.minimized).length);
</script>

<template>
<div class="RootTabs" :inert="overview">
    <button :class="{ on: rootTab === 'files' }" @click="rootTab = 'files'">파일</button>
    <button :class="{ on: rootTab === 'terminals' }" @click="rootTab = 'terminals'">
        터미널 {{ runs.length }}<template v-if="minCount"> · 최소화 {{ minCount }}</template>
    </button>
</div>
</template>

<style scoped lang="scss">
.RootTabs {
    flex: none;
    display: flex;
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    @include vars.font('body-small');
    > button {
        padding: 4px vars.$spacing-md;
        border: 0;
        border-bottom: 2px solid transparent;
        background: none;
        color: inherit;
        opacity: 0.65;
        cursor: pointer;
        &.on {
            opacity: 1;
            border-bottom-color: rgb(var(--v-theme-primary));
            font-weight: vars.$weight-md;
        }
    }
}
</style>
