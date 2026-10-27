<script setup lang="ts">
import { computed } from 'vue';
import { VCard, VCardText, VCardTitle, VDialog, VIcon } from 'vuetify/components';
import { useProcessPresence } from '../../api/presence.api';
import { levelTitle, PERM_OWNER } from '../../util/shareView.util';
import { agoText, lastSeenText, parseUserAgent } from '../../util/presenceView.util';

// 터미널 접속 상태 — 계정 > 세션(브라우저) > 탭. 열려 있는 동안 5초마다 다시 조회(push 없음, SH6(presence 범위·전달))
const props = defineProps({
    uid: {
        type: String,
        required: true,
    },
    name: {
        type: String,
        default: () => '',
    },
})
const open = defineModel<boolean>({ type: Boolean, default: () => false });

const query = useProcessPresence(computed(() => props.uid), open);
const accounts = computed(() => query.data.value ?? []);

const permText = (perm: number) => perm === PERM_OWNER ? '소유자' : levelTitle(perm) ?? '';
const tabText = (t: { connected: boolean, viewing: boolean, sizeOwner: boolean, since: number }) => [
    t.connected ? (t.viewing ? '보는 중' : '연결됨') : '끊김(60초 보관 중)',
    t.sizeOwner ? '크기 소유' : '',
    `${agoText(t.since)} 연결`,
].filter(Boolean).join(' · ');
</script>

<template>
<v-dialog v-model="open" max-width="460" min-width="296">
    <v-card :loading="query.isFetching.value" class="ct-card">
        <v-card-title>{{ name }} 접속 상태</v-card-title>
        <v-card-text>
            <p v-if="query.isPending.value" class="note">불러오는 중…</p>
            <p v-else-if="query.isError.value" class="note err">조회에 실패했습니다</p>
            <div v-for="a in accounts" :key="a.userId" class="acc">
                <div class="head">
                    <i class="dot" :class="{ on: a.online }" />
                    <b>{{ a.nickname }}</b>
                    <span class="perm">{{ permText(a.permission) }}</span>
                    <v-spacer />
                    <small>{{ lastSeenText(a.online, a.lastSeen) }}</small>
                </div>
                <div v-for="s in a.sessions" :key="s.key" class="sess">
                    <div class="ua"><v-icon size="14" icon="mdi-web" />{{ parseUserAgent(s.userAgent) }}</div>
                    <div v-for="(t, i) in s.tabs" :key="i" class="tab" :class="{ off: !t.connected }">{{ tabText(t) }}</div>
                </div>
            </div>
        </v-card-text>
        <div class="ct-actions">
            <span class="ct-grp ct-grp--r">
                <button type="button" class="ct-seg ct-seg--primary" @click="open = false">닫기</button>
            </span>
        </div>
    </v-card>
</v-dialog>
</template>

<style scoped lang="scss">
.note { @include vars.font('body-medium'); opacity: 0.7; &.err { color: rgb(var(--v-theme-error)); opacity: 1; } }
.acc + .acc { margin-top: vars.$spacing-md; padding-top: vars.$spacing-md; border-top: 1px solid rgba(var(--v-border-color), var(--v-border-opacity)); }
.head { display: flex; align-items: center; gap: vars.$spacing-sm; small { @include vars.font('body-small'); opacity: 0.7; } }
.perm { @include vars.font('body-small'); padding: 0 6px; border-radius: 8px; color: rgb(var(--v-theme-share)); background: rgba(var(--v-theme-share), 0.12); }
.dot { width: 8px; height: 8px; border-radius: 50%; background: rgba(var(--v-theme-on-surface), 0.3); &.on { background: rgb(var(--v-theme-share)); } }
.sess { margin: 4px 0 0 16px; @include vars.font('body-medium'); }
.ua { display: flex; align-items: center; gap: 4px; }
.tab { margin-left: 18px; @include vars.font('body-small'); opacity: 0.8; &.off { opacity: 0.5; } }
</style>
