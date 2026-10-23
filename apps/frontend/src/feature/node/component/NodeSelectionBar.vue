<script setup lang="ts">
import { tabReady } from '@/common/util/tabId.util';
import { computed, onBeforeUnmount, onMounted, watch } from 'vue';
import { useDisplay } from 'vuetify';
import { VBtn, VIcon, VList, VListItem, VMenu } from 'vuetify/components';
import { useNodeClipboard } from '../store/nodeClipboard.store';
import { nodeActItems, useNodeOps, type NodeAct, type NodeActItem } from '../store/nodeOps.store';
import { useNodeSelection } from '../store/nodeSelection.store';
import { useTileTree } from '../store/tileTree.store';
import { PERM_EDIT } from '../util/shareView.util';

// 선택 모드 막대(타일 위) — 선택 수 · 동작 · 클립보드 내용 · 끝내기. REF-node-select.md SM6(막대 동작)
const sel = useNodeSelection();
const ops = useNodeOps();
const { clip, clear: clearClip } = useNodeClipboard();
const { tiles, hash } = useTileTree();
const { xs } = useDisplay();

// SM7(사라진 선택): 타일이 닫히거나 다른 폴더로 가면 빼고, 보드를 바꾸거나 보드 화면을 떠나면 끝낸다
watch(tiles, t => sel.pruneTiles(t), { deep: true });
watch(hash, () => sel.exit());
onBeforeUnmount(() => sel.exit());

// Esc = 끝내기. 다이얼로그·메뉴가 열려 있거나 터미널·입력란에서 누른 Esc는 그쪽 것
const onKeyDown = (e: KeyboardEvent) => {
    if (e.key !== 'Escape' || !sel.on.value || e.defaultPrevented) return;
    const t = e.target as HTMLElement | null;
    if (t?.closest('.xterm, input, textarea, [contenteditable="true"]') || document.querySelector('.v-overlay--active')) return;
    sel.exit();
}
onMounted(() => window.addEventListener('keydown', onKeyDown));
onBeforeUnmount(() => window.removeEventListener('keydown', onKeyDown));

const nodes = computed(() => sel.selected.value.map(it => it.node));
const one = computed(() => sel.selected.value.length === 1 ? sel.selected.value[0]! : null);
// 여러 개 = 첫 노드를 고른 타일에 진행 막대
const tileId = computed(() => sel.selected.value[0]?.tileId ?? '');

// 1개 = 행 동작 전부(+ 새 타일·실행) / 여러 개 = 복사·잘라내기·삭제(편집 미만이 섞이면 잘라내기·삭제 비활성)
const items = computed<NodeActItem[]>(() => {
    const viDisabled = ops.saving.value || !tabReady.value;
    if (one.value) {
        const n = one.value.node;
        const head: NodeActItem = n.kind === 'FOLDER'
            ? { act: 'open', icon: 'mdi-dock-window', title: '새 타일' }
            : { act: 'exec', icon: 'mdi-play', title: '실행', disabled: viDisabled };
        return [head, ...nodeActItems(n, viDisabled)];
    }
    if (!nodes.value.length) return [];
    const ro = nodes.value.some(n => (n.permission ?? 0) < PERM_EDIT);
    return [
        { act: 'copy', icon: 'mdi-content-copy', title: '복사' },
        { act: 'cut', icon: 'mdi-content-cut', title: '잘라내기', disabled: ro },
        { act: 'remove', icon: 'mdi-trash-can-outline', title: '삭제', disabled: ro, danger: true },
    ];
})
// 모바일 = 복사·잘라내기·삭제만 아이콘으로, 나머지는 더보기
const MAIN: NodeAct[] = ['copy', 'cut', 'remove'];
const shown = computed(() => xs.value ? items.value.filter(it => MAIN.includes(it.act)) : items.value);
const more = computed(() => xs.value ? items.value.filter(it => !MAIN.includes(it.act)) : []);

const onAct = (a: NodeAct) => ops.act(a, one.value?.tileId ?? tileId.value, nodes.value);

const clipText = computed(() => {
    const { mode, nodes: list } = clip.value;
    return `${mode === 'cut' ? '잘라냄' : '복사됨'} ${list.length}개: ${list.map(n => n.name).join(', ')}`;
})
</script>

<template>
<div v-if="sel.on.value" class="NodeSelectionBar">
    <span class="count">{{ nodes.length }}개 선택</span>
    <span class="acts">
        <v-btn v-for="it in shown" :key="it.act"
            size="small"
            variant="text"
            :color="it.danger ? 'error' : undefined"
            :disabled="it.disabled || ops.saving.value"
            :icon="xs"
            :prepend-icon="xs ? undefined : it.icon"
            :title="it.title"
            @click="onAct(it.act)"
        >
            <v-icon v-if="xs" :icon="it.icon" size="18" />
            <template v-else>{{ it.title }}</template>
        </v-btn>
        <v-menu v-if="more.length" location="bottom end">
            <template #activator="{ props: act }">
                <v-btn v-bind="act" size="small" variant="text" icon="mdi-dots-horizontal" title="더보기" />
            </template>
            <v-list density="compact">
                <v-list-item v-for="it in more" :key="it.act"
                    :prepend-icon="it.icon"
                    :title="it.title"
                    :disabled="it.disabled || ops.saving.value"
                    @click="onAct(it.act)"
                />
            </v-list>
        </v-menu>
        <span v-if="!items.length" class="hint">행을 눌러 고르세요</span>
    </span>
    <span v-if="clip.nodes.length" class="clip" :title="clipText">
        <v-icon size="14" :icon="clip.mode === 'cut' ? 'mdi-content-cut' : 'mdi-content-copy'" />
        <span class="clip-text">{{ clipText }}</span>
        <button title="클립보드 비우기" @click="clearClip"><v-icon size="12" icon="mdi-close" /></button>
    </span>
    <v-btn size="small" variant="tonal" title="선택 끝내기(Esc)" :icon="xs" @click="sel.exit()">
        <v-icon v-if="xs" icon="mdi-check" size="18" />
        <template v-else>끝내기</template>
    </v-btn>
</div>
</template>

<style scoped lang="scss">
.NodeSelectionBar {
    flex: none;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: vars.$spacing-sm vars.$spacing-md;
    padding: vars.$spacing-xs vars.$spacing-md;
    @include vars.font('body-medium');
    background: rgba(var(--v-theme-primary), 0.08);
    border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}
.count {
    font-weight: vars.$weight-lg;
    color: rgb(var(--v-theme-primary));
    white-space: nowrap;
}
.acts {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
}
.hint { opacity: 0.6; }
.clip {
    flex: 1 1 0;
    min-width: 0;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 4px;
    opacity: 0.8;
    .clip-text {
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
    button {
        padding: 0;
        border: 0;
        background: none;
        color: inherit;
        cursor: pointer;
    }
}
.clip + .v-btn, .acts + .v-btn { margin-left: auto; }
</style>
