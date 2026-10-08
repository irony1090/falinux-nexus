<script setup lang="ts">
import { useQueryClient } from '@tanstack/vue-query';
import { computed, ref } from 'vue';
import { VBtn, VCard, VCardActions, VCardText, VCardTitle, VDialog, VIcon, VSpacer } from 'vuetify/components';
import { deleteNodeShare, deleteProcessShare, invalidateMyShares, invalidateNodeShares, useMyShares, type MyShare } from '../../api/share.api';
import { levelTitle, PERM_BLOCK } from '../../util/shareView.util';

// 내가 공유한 것(O5) — 노드 + process 단독 공유를 대상별로 묶어 보여 주고 해제만. 단계 변경은 폴더 공유 창(O2)에서
const open = defineModel<boolean>({ type: Boolean, default: () => false });

const queryClient = useQueryClient();
const query = useMyShares(open);
const loading = ref(false);
const message = ref<string>();

type Group = { key: string, kind: MyShare['kind'], name: string, rows: MyShare[] };
const groups = computed<Group[]>(() => {
    const map = new Map<string, Group>();
    for (const s of query.data.value ?? []) {
        const key = s.kind === 'NODE' ? `N${s.nodeId}` : `P${s.processUid}`;
        let g = map.get(key);
        if (!g) map.set(key, g = { key, kind: s.kind, name: s.name, rows: [] });
        g.rows.push(s);
    }
    return [...map.values()];
});

const levelText = (perm: number) => perm === PERM_BLOCK ? '차단' : levelTitle(perm) ?? String(perm);

const release = (s: MyShare) => {
    loading.value = true;
    message.value = undefined;
    const job = s.kind === 'NODE' ? deleteNodeShare(s.nodeId!, s.userId) : deleteProcessShare(s.processUid!, s.userId);
    job.then(async () => { await (s.kind === 'NODE' ? invalidateNodeShares(queryClient, s.nodeId!) : invalidateMyShares(queryClient)); })
    .catch(err => message.value = err?.message || '해제에 실패했습니다')
    .finally(() => loading.value = false);
}
</script>

<template>
<v-dialog v-model="open" max-width="460" min-width="296" :persistent="loading">
    <v-card :loading="loading || query.isFetching.value" :disabled="loading">
        <v-card-title>내가 공유한 것</v-card-title>
        <v-card-text>
            <p v-if="query.isPending.value" class="note">불러오는 중…</p>
            <p v-else-if="!groups.length" class="note">공유한 것이 없습니다</p>
            <div v-for="g in groups" :key="g.key" class="group">
                <div class="head">
                    <v-icon size="16" :icon="g.kind === 'NODE' ? 'mdi-folder' : 'mdi-console'" color="share" />
                    <span class="nm">{{ g.name }}</span>
                    <span v-if="g.kind === 'PROCESS'" class="tag">실행 중 터미널</span>
                </div>
                <div v-for="s in g.rows" :key="s.userId" class="row">
                    <span class="who">{{ s.nickname }}</span>
                    <span class="lv" :class="{ block: s.permission === PERM_BLOCK }">{{ levelText(s.permission) }}</span>
                    <v-btn size="x-small" variant="text" color="error" @click="release(s)">해제</v-btn>
                </div>
            </div>
            <p v-if="message" class="note err">{{ message }}</p>
        </v-card-text>
        <v-card-actions>
            <v-spacer />
            <v-btn variant="text" @click="open = false">닫기</v-btn>
        </v-card-actions>
    </v-card>
</v-dialog>
</template>

<style scoped lang="scss">
.note { font-size: 13px; opacity: 0.7; &.err { color: rgb(var(--v-theme-error)); opacity: 1; } }
.group + .group { margin-top: vars.$spacing-md; }
.head { display: flex; align-items: center; gap: 6px; font-weight: vars.$weight-lg; min-width: 0; }
.nm { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.tag { font-size: 10.5px; padding: 0 6px; border-radius: 8px; color: rgb(var(--v-theme-share)); background: rgba(var(--v-theme-share), 0.12); white-space: nowrap; }
.row { display: flex; align-items: center; gap: vars.$spacing-sm; padding: 2px 0 2px 22px; font-size: 13px; }
.who { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.lv { color: rgb(var(--v-theme-share)); &.block { color: rgb(var(--v-theme-error)); } }
</style>
