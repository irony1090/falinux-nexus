<script setup lang="ts">
import { VIcon } from 'vuetify/components';
import { EDGE_BAND, EDGE_HOLD_MS, useTileGrids } from '../store/tileGrids.store';

// 끄는 동안 화면 끝 띠(넘길 페이지가 있는 쪽만) — 안에 머물면 차오르고, 다 차면 칼럼 하나(0.5페이지) 넘김. 판정 = tileGrids edgeTrack
// 끌기 판정(elementFromPoint)을 가리지 않게 pointer-events 없음
const { edge } = useTileGrids();
const SIDES = [
    { side: -1, key: 'prev', icon: 'mdi-chevron-left', text: '이전 페이지' },
    { side: 1, key: 'next', icon: 'mdi-chevron-right', text: '다음 페이지' },
] as const;
</script>

<template>
<template v-for="s in SIDES" :key="s.key">
    <transition name="pe-fade">
        <div v-if="edge?.[s.key]" class="PageEdgeBand" :class="[s.key, { on: edge.side === s.side }]"
            :style="{ width: `${EDGE_BAND}px`, '--hold': `${EDGE_HOLD_MS}ms` }" aria-hidden="true"
        >
            <i v-if="edge.side === s.side" :key="edge.since" class="fill" />
            <span class="lbl">
                <v-icon size="20" :icon="s.icon" />
                <span class="t">{{ s.text }}</span>
            </span>
        </div>
    </transition>
</template>
</template>

<style scoped lang="scss">
.PageEdgeBand {
    position: absolute;
    top: 0;
    bottom: 0;
    z-index: 30;
    overflow: hidden;
    pointer-events: none;
    color: rgb(var(--v-theme-primary));
    transition: background-color 150ms ease;
    &.prev { left: 0; background: linear-gradient(to right, rgba(var(--v-theme-primary), 0.16), transparent); }
    &.next { right: 0; background: linear-gradient(to left, rgba(var(--v-theme-primary), 0.16), transparent); }
    &.on { background-color: rgba(var(--v-theme-primary), 0.12); }
}
// 머무는 동안 아래에서 위로 차오름 — 다 차면 넘김(키 = 들어간 때라 넘긴 뒤 다시 차오름)
.fill {
    position: absolute;
    inset: 0;
    background: rgba(var(--v-theme-primary), 0.2);
    transform-origin: bottom;
    animation: pe-fill var(--hold) linear forwards;
}
.lbl {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
    opacity: 0.55;
    transition: opacity 150ms ease;
    .on & { opacity: 1; }
}
.t {
    writing-mode: vertical-rl;
    @include vars.font('label-medium');
    font-weight: vars.$weight-md;
    white-space: nowrap;
}
@keyframes pe-fill { from { transform: scaleY(0); } to { transform: scaleY(1); } }
.pe-fade-enter-active, .pe-fade-leave-active { transition: opacity 180ms ease; }
.pe-fade-enter-from, .pe-fade-leave-to { opacity: 0; }
</style>
