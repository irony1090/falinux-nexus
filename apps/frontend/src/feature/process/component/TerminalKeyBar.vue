<script lang="ts">
export const KEY_BAR_HEIGHT = 40;    // 타일 영역 이동(키보드-2)도 이만큼 비운다
</script>

<script setup lang="ts">
import { useVirtualKeyboard } from '@/common/hook/virtualKeyboard.hook';
import { isReadOnly } from '@/feature/node/util/shareView.util';
import { computed } from 'vue';
import { applyMods, keyMods, pressMod, releaseMod, takeCsiMod, type Mod } from '../store/keyModifier.store';
import { useProcessTerm } from '../store/processTerm.store';

// 가상 키보드 바로 위 특수 키 한 줄 — 키보드-3(특수 키 줄), VK1(키 목록)·VK2(표시 조건) → REF-mobile-keyboard.md
// 버튼은 pointerdown 기본 동작을 막아 터미널 포커스를 지킨다(포커스가 빠지면 가상 키보드가 닫힌다)

const kb = useVirtualKeyboard();
const { procs, sendText, appCursor } = useProcessTerm();

const proc = computed(() => kb.termUid ? procs.value[kb.termUid] : undefined);
const show = computed(() => kb.open && proc.value?.status === 'PROCESS' && !isReadOnly(proc.value.permission));

type Key = { label: string; send?: string; arrow?: 'A' | 'B' | 'C' | 'D'; mod?: Mod; apply?: boolean };
const KEYS: Key[] = [
    { label: 'Esc', send: '\x1b' },
    { label: 'Tab', send: '\t' },
    { label: 'Ctrl', mod: 'ctrl' },
    { label: 'Alt', mod: 'alt' },
    { label: '←', arrow: 'D' },
    { label: '↑', arrow: 'A' },
    { label: '↓', arrow: 'B' },
    { label: '→', arrow: 'C' },
    { label: '|', send: '|', apply: true },
    { label: '~', send: '~', apply: true },
    { label: '/', send: '/', apply: true },
    { label: '-', send: '-', apply: true },
];

const arrowOf = (uid: string, dir: string) => {
    const n = takeCsiMod();
    if (n > 1) return `\x1b[1;${n}${dir}`;
    return appCursor(uid) ? `\x1bO${dir}` : `\x1b[${dir}`;
}

const down = (k: Key) => {
    const uid = kb.termUid;
    if (k.mod) return pressMod(k.mod);
    if (!uid) return;
    if (k.arrow) return sendText(uid, arrowOf(uid, k.arrow));
    if (k.send) sendText(uid, k.apply ? applyMods(k.send) : k.send);
}
const up = (k: Key) => { if (k.mod) releaseMod(k.mod); }

const style = computed(() => ({ top: `${kb.bottom - KEY_BAR_HEIGHT}px`, height: `${KEY_BAR_HEIGHT}px` }));
</script>

<template>
<div v-if="show" class="TerminalKeyBar" :style="style">
    <div v-for="k in KEYS" :key="k.label" class="key" role="button"
         :class="{ on: k.mod && keyMods[k.mod].armed }"
         @pointerdown.prevent="down(k)" @pointerup="up(k)" @pointercancel="up(k)" @contextmenu.prevent>
        {{ k.label }}
    </div>
</div>
</template>

<style scoped lang="scss">
.TerminalKeyBar {
    position: fixed;
    left: 0;
    right: 0;
    z-index: 2000;
    display: flex;
    background: rgb(var(--v-theme-surface));
    border-top: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
    touch-action: manipulation;
    user-select: none;
    -webkit-user-select: none;
}
.key {
    flex: 1 1 0;
    min-width: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 13px;
    color: rgb(var(--v-theme-on-surface));
    cursor: pointer;
    &:active { background: rgba(var(--v-theme-on-surface), 0.12); }
    &.on {
        background: rgb(var(--v-theme-primary));
        color: rgb(var(--v-theme-on-primary));
    }
}
</style>
