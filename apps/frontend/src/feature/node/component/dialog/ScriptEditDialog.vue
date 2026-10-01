<script setup lang="ts">
import { useAppDialog } from '@/feature/layout/store/appDialog.store';
import { computed, ref, watch, type PropType } from 'vue';
import { useDisplay } from 'vuetify';
import { VBtn, VCard, VCardActions, VCardTitle, VDialog, VTextarea } from 'vuetify/components';
import { getNode, useNodeActions, type NodeResponse } from '../../api/node.api';

// 스크립트 내용(nodes.content) 브라우저 편집 — REF-node-ui-link.md "4(스크립트 편집)"
const props = defineProps({
    node: {
        type: Object as PropType<NodeResponse>,
        default: () => undefined,
    },
})
const open = defineModel<boolean>({ type: Boolean, default: () => false });

const { patch } = useNodeActions();
const { open: appDialogOpen, openDialog } = useAppDialog();
const { xs } = useDisplay();

const original = ref('');
const text = ref('');
const loading = ref(false);
const saving = ref(false);
const dirty = computed(() => text.value !== original.value);

const fail = (err: { message?: string }, fallback: string) => openDialog({ type: 'error', content: err?.message || fallback });

// 열 때마다 최신 내용을 새로 조회한다(목록 캐시는 최대 1분 묵은 값)
watch(open, val => {
    const id = props.node?.id;
    if (!val || id === undefined) return;
    original.value = text.value = '';
    loading.value = true;
    getNode(id).then(res => {
        if (!open.value || props.node?.id !== id) return;   // 조회 중에 닫혔거나 대상이 바뀜
        original.value = text.value = res.content ?? '';
    }).catch(err => {
        open.value = false;
        fail(err, '내용을 불러오지 못했습니다');
    }).finally(() => loading.value = false);
})

// 저장 버튼·Ctrl+S 모두 저장하고 계속 편집(닫기는 닫기 경로로만)
const save = () => {
    const id = props.node?.id;
    if (id === undefined || loading.value || saving.value || !dirty.value) return;
    const content = text.value;
    saving.value = true;
    patch(id, { content: content === '' ? null : content }).then(() => {
        original.value = content;
    }).catch(err => fail(err, '저장하지 못했습니다'))
    .finally(() => saving.value = false);
}

// Esc·바깥 클릭·닫기 버튼 공통: 고친 내용이 있으면 확인을 거친다
const onClose = () => {
    if (saving.value) return;
    if (!dirty.value) {
        open.value = false;
        return;
    }
    openDialog({
        type: 'warning',
        title: '저장하지 않은 내용',
        content: '저장하지 않고 닫으면 고친 내용이 사라집니다.',
        buttons: [
            { view: '계속 편집', type: 'info', bindButton: 'Escape', action: () => appDialogOpen.value = false },
            {
                view: '저장하지 않고 닫기', type: 'error',
                action: () => {
                    appDialogOpen.value = false;
                    open.value = false;
                },
            },
        ],
    })
}

// Tab은 포커스 이동 대신 탭 문자를 넣는다
const onTab = (e: KeyboardEvent) => {
    const el = e.target;
    if (!(el instanceof HTMLTextAreaElement)) return;
    el.setRangeText('\t', el.selectionStart, el.selectionEnd, 'end');
    el.dispatchEvent(new Event('input', { bubbles: true }));
}
</script>

<template>
<v-dialog :model-value="open" max-width="760" :height="xs ? undefined : 640" :fullscreen="xs" @update:model-value="onClose">
    <v-card class="ScriptEditDialog"
        :loading="loading || saving"
        @keydown.ctrl.s.prevent="save"
        @keydown.meta.s.prevent="save"
    >
        <v-card-title class="title">
            <span class="name">{{ node?.name }}</span>
            <span v-if="dirty" class="dirty text-warning">수정됨</span>
        </v-card-title>
        <v-textarea v-model="text"
            class="editor"
            variant="outlined"
            hide-details
            no-resize
            autofocus
            wrap="off"
            spellcheck="false"
            placeholder="#!/bin/bash"
            :disabled="loading"
            @keydown.tab.exact.prevent="onTab"
        />
        <v-card-actions>
            <span class="hint">Ctrl+S 저장</span>
            <v-btn variant="text" :disabled="saving" @click="onClose">닫기</v-btn>
            <v-btn variant="flat" color="primary" :disabled="loading || saving || !dirty" @click="save">저장</v-btn>
        </v-card-actions>
    </v-card>
</v-dialog>
</template>

<style scoped lang="scss">
// 높이는 VDialog의 height가 정한다 — 카드는 그 안의 flex 항목이라 카드에 준 height는 무시된다
.ScriptEditDialog {
    display: flex;
    flex-direction: column;
}
.title {
    display: flex;
    align-items: baseline;
    gap: vars.$spacing-md;
    .name {
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
    }
    .dirty { flex: none; font-size: 12px; }
}
// 입력 영역이 카드의 남는 높이를 전부 차지하게 한다(Vuetify 기본은 rows 기준 높이)
.editor {
    flex: 1;
    min-height: 0;
    margin: 0 vars.$spacing-xl;
    grid-template-rows: minmax(0, 1fr);
    :deep(.v-input__control),
    :deep(.v-field),
    :deep(.v-field__field) { height: 100%; }
    :deep(textarea) {
        height: 100%;
        font-family: monospace;
        font-size: 13px;
        line-height: 1.5;
        tab-size: 4;
        white-space: pre;
        overflow: auto;
    }
}
.hint {
    margin-right: auto;
    padding-left: vars.$spacing-md;
    font-size: 11px;
    opacity: 0.6;
}
</style>
