<script setup lang="ts">
import { useAppDialog } from '@/feature/layout/store/appDialog.store';
import { computed, ref, watch, type PropType } from 'vue';
import { useDisplay } from 'vuetify';
import { VCard, VCardTitle, VDialog } from 'vuetify/components';
import CtCode, { ctCodeLangOf } from '@/feature/widget/component/CtCode.vue';
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
const editor = ref<InstanceType<typeof CtCode>>();
// 불러오는 동안은 편집 불가라 포커스가 안 들어간다 — 다 불러온 뒤 넣는다
watch(loading, (v, was) => { if (was && !v && open.value) requestAnimationFrame(() => editor.value?.focus()); });

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

// 확장자 없는 스크립트 = 셸(shebang). .bat·.cmd·.txt만 강조 없이
const lang = computed(() => {
    const name = props.node?.name ?? '';
    const l = ctCodeLangOf(name);
    return l === 'text' && !/\.(bat|cmd|txt)$/i.test(name) ? 'sh' : l;
});
</script>

<template>
<v-dialog :model-value="open" max-width="760" :height="xs ? undefined : 640" :fullscreen="xs" @update:model-value="onClose">
    <v-card class="ScriptEditDialog ct-card"
        :loading="loading || saving"
        @keydown.ctrl.s.prevent="save"
        @keydown.meta.s.prevent="save"
    >
        <v-card-title class="title">
            <span class="name">{{ node?.name }}</span>
            <span v-if="dirty" class="dirty text-warning">수정됨</span>
        </v-card-title>
        <CtCode v-if="open" ref="editor" v-model="text"
            class="editor"
            :lang="lang"
            placeholder="#!/bin/bash"
            :disabled="loading"
        />
        <div class="ct-actions">
            <span class="hint">Ctrl+S 저장</span>
            <span class="ct-grp ct-grp--r">
                <button type="button" class="ct-seg" :disabled="saving" @click="onClose">닫기</button>
                <button type="button" class="ct-seg ct-seg--primary" :disabled="loading || saving || !dirty" @click="save">저장</button>
            </span>
        </div>
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
    .dirty { flex: none; @include vars.font('body-medium'); }
}
// 입력 영역이 카드의 남는 높이를 전부 차지한다
.editor {
    flex: 1;
    min-height: 0;
    margin: 0 vars.$spacing-xl;
}
.hint {
    align-self: center;
    margin-right: auto;
    padding-left: vars.$spacing-md;
    @include vars.font('body-small');
    opacity: 0.6;
}
</style>
