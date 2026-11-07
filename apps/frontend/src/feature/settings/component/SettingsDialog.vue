<script setup lang="ts">
import type { TileBarAt } from '@/feature/widget/util/tileBar.util';
import { setSetting, settings, settingsDialog as open } from '../store/settings.store';

// 설정 창 — 한 열: 묶음 제목 + 항목(Q17(설정 창 구성)). 고르면 바로 적용, 버튼은 [닫기]만(Q23(설정 창 적용)) — REF-settings-footer.md
const BAR_AT: { value: TileBarAt; text: string }[] = [
    { value: 'top', text: '위' },
    { value: 'bottom', text: '아래' },
];
</script>

<template>
<v-dialog v-model="open" max-width="440" min-width="296">
    <v-card class="ct-card">
        <v-card-title>설정</v-card-title>
        <p class="sec">타일</p>
        <div class="row">
            <div class="lb">
                <span class="name">메뉴 줄 위치</span>
                <span class="hint">경로 줄 같은 타일 메뉴 줄을 붙일 곳 · 이 브라우저에만 저장(로그아웃하면 처음 값)</span>
            </div>
            <!-- 창 오른쪽 벽에 붙는 묶음 — 찰떡(REF-design-chaltteok.md) -->
            <span class="ct-grp ct-grp--r pick" role="group" aria-label="메뉴 줄 위치">
                <button v-for="o in BAR_AT" :key="o.value" type="button" class="ct-seg" :class="{ 'ct-seg--primary': settings.tileBarAt === o.value }"
                    :aria-pressed="settings.tileBarAt === o.value" @click="setSetting('tileBarAt', o.value)">{{ o.text }}</button>
            </span>
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
// 왼쪽 여백 = v-card-title과 같은 16px, 오른쪽 여백 없음(고르기 묶음이 벽에 붙게)
.sec {
    margin: vars.$spacing-md 0 vars.$spacing-sm;
    padding: 0 vars.$spacing-xl;
    @include vars.font('label-medium');
    font-weight: vars.$weight-lg;
    opacity: 0.6;
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
