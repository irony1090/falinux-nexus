<script setup lang="ts">
import { useQueryClient } from '@tanstack/vue-query';
import { computed, ref, watch } from 'vue';
import { VAvatar, VCard, VCardText, VCardTitle, VDialog, VForm, VIcon, VList, VListItem, VMenu, VRadio, VRadioGroup } from 'vuetify/components';
import CtTextField from '@/feature/widget/component/CtTextField.vue';
import { deleteNodeShare, invalidateNodeShares, lookupUser, putNodeShare, useNodeShares, type LookupUserResponse, type NodeShareEntry } from '../../api/share.api';
import { useUsersPresence } from '../../api/presence.api';
import { lastSeenText } from '../../util/presenceView.util';
import { levelTitle, PERM_BLOCK, PERM_VIEW, SHARE_LEVELS } from '../../util/shareView.util';

// 폴더 공유 — 목록(O2) → 사람 추가(아이디 → 닉네임 확인 + 단계 → 완료). 시안: 공유-e-2(공유 걸기 창)·공유-e-3a(공유 목록 창)
const props = defineProps({
    nodeId: {
        type: Number,
        required: true,
    },
    nodeName: {
        type: String,
        default: () => '',
    },
    tileId: {
        type: String,
        default: () => '',
    },
})
const open = defineModel<boolean>({ type: Boolean, default: () => false });

type Step = 'list' | 'id' | 'confirm' | 'done';
const step = ref<Step>('list');
const identification = ref('');
const message = ref<string>();
const found = ref<LookupUserResponse>();
const level = ref(PERM_VIEW);
const previous = ref<number | null>(null);
const loading = ref(false);

const queryClient = useQueryClient();
const sharesQuery = useNodeShares(computed(() => props.nodeId), open);
const entries = computed(() => sharesQuery.data.value ?? []);
const presence = useUsersPresence(computed(() => entries.value.map(e => e.userId)), computed(() => open.value && step.value === 'list'));
const presenceOf = (userId: number) => presence.data.value?.find(p => p.userId === userId);
const note = ref<string>();         // 목록 아래 한 줄 안내
const editing = ref<number>();      // "여기서 바꾸기"를 누른 이어진 행의 userId
const pick = ref(PERM_VIEW);

// 열릴 때마다 목록부터
watch(open, val => {
    if (!val) return;
    step.value = 'list';
    identification.value = '';
    message.value = undefined;
    found.value = undefined;
    level.value = PERM_VIEW;
    previous.value = null;
    note.value = undefined;
    editing.value = undefined;
})

const refresh = () => invalidateNodeShares(queryClient, props.nodeId);
const fail = (err: { message?: string, status?: number }) => {
    message.value = err.status === 404 ? `'${identification.value.trim()}' 계정이 없습니다. 아이디를 다시 확인하세요`
        : err.status === 400 ? '자기 자신에게는 공유할 수 없습니다'
        : err.message || '요청에 실패했습니다';
}

// ---- 목록(O2) ----
const sub = (e: NodeShareEntry) => e.direct ? '여기서 공유' : `상위 '${e.sourceName ?? ''}'에서 이어짐`;
// 직접 건 행의 선택지 — 차단(0)은 이미 차단인 행에서만(풀어 주려고). 새로 막는 건 이어진 행의 "여기서 바꾸기"
const directItems = (e: NodeShareEntry) => (e.permission === PERM_BLOCK ? [{ value: PERM_BLOCK, title: '차단' }] : []).concat(SHARE_LEVELS.map(l => ({ value: l.value, title: l.title })));
const pickItems = [{ value: PERM_BLOCK, title: '차단' }, ...SHARE_LEVELS.map(l => ({ value: l.value, title: l.title }))];

const act = (job: Promise<unknown>) => {
    loading.value = true;
    message.value = undefined;
    return job.catch(err => note.value = err?.message || '요청에 실패했습니다').finally(() => loading.value = false);
}
const onChange = (e: NodeShareEntry, value: number) => act(putNodeShare(props.nodeId, e.userId, value, props.tileId).then(() => {
    note.value = `${e.nickname}님의 단계를 ${levelTitle(value) ?? '차단'}(으)로 바꿨습니다`;
    return refresh();
}))
const onRelease = (e: NodeShareEntry) => act(deleteNodeShare(props.nodeId, e.userId).then(res => {
    note.value = res.entry ? `${e.nickname}님은 상위 공유(${levelTitle(res.entry.permission) ?? '차단'})로 돌아갑니다` : `${e.nickname}님과의 공유를 풀었습니다`;
    return refresh();
}))
const onEditStart = (e: NodeShareEntry) => {
    editing.value = e.userId;
    pick.value = e.permission;
}
const onApply = (e: NodeShareEntry) => act(putNodeShare(props.nodeId, e.userId, pick.value, props.tileId).then(() => {
    editing.value = undefined;
    note.value = `${e.nickname}님은 이 폴더에서 ${levelTitle(pick.value) ?? '차단'}(으)로 바뀌었습니다`;
    return refresh();
}))

// ---- 사람 추가 ----
const onFind = () => {
    const id = identification.value.trim();
    if (!id) {
        message.value = '아이디를 입력하세요';
        return;
    }
    loading.value = true;
    lookupUser(id)
        .then(u => {
            found.value = u;
            step.value = 'confirm';
        })
        .catch(fail)
        .finally(() => loading.value = false);
}

const onShare = () => {
    if (!found.value) return;
    loading.value = true;
    putNodeShare(props.nodeId, found.value.id, level.value, props.tileId)
        .then(res => {
            previous.value = res.previous;
            step.value = 'done';
            return refresh();
        })
        .catch(err => message.value = err.message || '공유에 실패했습니다')
        .finally(() => loading.value = false);
}

const doneText = () => {
    const nick = found.value?.nickname ?? '';
    const now = levelTitle(level.value);
    const before = levelTitle(previous.value);
    return before && previous.value !== level.value ? `${nick}님의 단계를 ${before} → ${now}(으)로 바꿨습니다`
        : before ? `${nick}님은 이미 ${now} 단계로 공유 중입니다`
        : `${nick}님에게 '${props.nodeName}'을(를) ${now} 단계로 공유했습니다`;
}

const onAdd = () => {
    identification.value = '';
    message.value = undefined;
    level.value = PERM_VIEW;
    step.value = 'id';
}
</script>

<template>
<v-dialog v-model="open" max-width="460" min-width="296" :persistent="loading">
    <v-card :loading="loading" :disabled="loading" class="ct-card">
        <v-card-title>{{ nodeName }} 공유</v-card-title>

        <template v-if="step === 'list'">
            <v-card-text>
                <p class="sub">가까운 공유가 적용됩니다. 점선은 상위 폴더에서 이어진 것입니다.</p>
                <p v-if="!entries.length" class="empty">아직 공유한 사람이 없습니다.</p>
                <div v-for="e in entries" :key="e.userId" class="who" :class="{ inherited: !e.direct, blocked: e.permission === PERM_BLOCK }">
                    <v-avatar :color="e.permission === PERM_BLOCK ? 'error' : 'share'" :variant="e.direct ? 'flat' : 'outlined'" size="30">{{ e.nickname.charAt(0) }}</v-avatar>
                    <div class="id">
                        <b>{{ e.nickname }}</b><span v-if="e.permission === PERM_BLOCK" class="text-error blk">차단</span>
                        <small>{{ sub(e) }}</small>
                        <small v-if="presenceOf(e.userId)" class="seen"><i class="dot" :class="{ on: presenceOf(e.userId)!.online }" />{{ lastSeenText(presenceOf(e.userId)!.online, presenceOf(e.userId)!.lastSeen) }}</small>
                    </div>
                    <!-- 찰떡: 행 끝 묶음(오른쪽 = 목록 벽 쪽 곧게) -->
                    <span v-if="e.direct" class="ct-grp ct-grp--r">
                        <v-menu location="bottom end">
                            <template #activator="{ props: act }">
                                <button v-bind="act" type="button" class="ct-seg lv-pick" :aria-label="`${e.nickname} 단계`">{{ levelTitle(e.permission) ?? '차단' }}<v-icon size="16" icon="mdi-menu-down" /></button>
                            </template>
                            <v-list density="compact">
                                <v-list-item v-for="it in directItems(e)" :key="it.value" :title="it.title" :active="it.value === e.permission" @click="it.value !== e.permission && onChange(e, it.value)" />
                            </v-list>
                        </v-menu>
                        <button type="button" class="ct-seg ct-seg--error" @click="onRelease(e)">해제</button>
                    </span>
                    <span v-else-if="editing === e.userId" class="ct-grp ct-grp--r">
                        <v-menu location="bottom end">
                            <template #activator="{ props: act }">
                                <button v-bind="act" type="button" class="ct-seg lv-pick" aria-label="새 단계">{{ levelTitle(pick) ?? '차단' }}<v-icon size="16" icon="mdi-menu-down" /></button>
                            </template>
                            <v-list density="compact">
                                <v-list-item v-for="it in pickItems" :key="it.value" :title="it.title" :active="it.value === pick" @click="pick = it.value" />
                            </v-list>
                        </v-menu>
                        <button type="button" class="ct-seg ct-seg--share" @click="onApply(e)">적용</button>
                        <button type="button" class="ct-seg" @click="editing = undefined">취소</button>
                    </span>
                    <span v-else class="ct-grp ct-grp--r">
                        <span class="ct-seg ct-seg--txt">{{ levelTitle(e.permission) ?? '차단' }}</span>
                        <button type="button" class="ct-seg" @click="onEditStart(e)">여기서 바꾸기</button>
                    </span>
                </div>
                <p v-if="note" class="note">{{ note }}</p>
            </v-card-text>
            <div class="ct-actions">
                <span class="ct-grp ct-grp--r">
                    <button type="button" class="ct-seg ct-seg--share" @click="onAdd">사람 추가</button>
                    <button type="button" class="ct-seg ct-seg--primary" @click="open = false">닫기</button>
                </span>
            </div>
        </template>

        <v-form v-else-if="step === 'id'" @submit.prevent="onFind">
            <v-card-text>
                <CtTextField v-model="identification"
                    label="아이디"
                    hint="목록에서 고르지 않고 아이디를 정확히 입력합니다"
                    persistent-hint
                    autofocus
                    :error-messages="message"
                    @update:model-value="message = undefined"
                />
            </v-card-text>
            <div class="ct-actions">
                <span class="ct-grp ct-grp--r">
                    <button type="button" class="ct-seg" @click="step = 'list'">뒤로</button>
                    <button type="submit" class="ct-seg ct-seg--primary">찾기</button>
                </span>
            </div>
        </v-form>

        <template v-else-if="step === 'confirm'">
            <v-card-text>
                <div class="found">
                    <v-avatar color="share" size="32">{{ found?.nickname.charAt(0) }}</v-avatar>
                    <div>
                        <b>{{ found?.nickname }}</b>
                        <small>{{ identification.trim() }}</small>
                    </div>
                </div>
                <v-radio-group v-model="level" hide-details class="levels">
                    <v-radio v-for="l in SHARE_LEVELS" :key="l.value" :value="l.value" color="share">
                        <template #label>
                            <span class="lv"><b>{{ l.title }}</b><small>{{ l.desc }}</small></span>
                        </template>
                    </v-radio>
                </v-radio-group>
                <p v-if="message" class="text-error msg">{{ message }}</p>
            </v-card-text>
            <div class="ct-actions">
                <span class="ct-grp ct-grp--r">
                    <button type="button" class="ct-seg" @click="step = 'id'; message = undefined">뒤로</button>
                    <button type="button" class="ct-seg ct-seg--primary" @click="onShare">공유</button>
                </span>
            </div>
        </template>

        <template v-else>
            <v-card-text>
                <p class="done">{{ doneText() }}</p>
                <p class="sub">받는 쪽은 열려 있는 탭 전부에 바로 반영되고, 기본 보드에 폴더 타일이 생깁니다.</p>
            </v-card-text>
            <div class="ct-actions">
                <span class="ct-grp ct-grp--r">
                    <button type="button" class="ct-seg" @click="step = 'list'">목록으로</button>
                    <button type="button" class="ct-seg ct-seg--primary" @click="open = false">닫기</button>
                </span>
            </div>
        </template>
    </v-card>
</v-dialog>
</template>

<style scoped lang="scss">
.sub { @include vars.font('body-medium'); opacity: 0.7; margin-bottom: vars.$spacing-sm; }
.empty { padding: vars.$spacing-lg 0; text-align: center; opacity: 0.7; }
.who {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto;
    align-items: center;
    gap: vars.$spacing-md;
    padding: vars.$spacing-sm 0;
    &.inherited .v-avatar { border-style: dashed; border-width: 1.5px; }
    .id small { display: block; @include vars.font('body-small'); opacity: 0.7; }
    .blk { margin-left: vars.$spacing-sm; @include vars.font('body-medium'); font-weight: vars.$weight-md; }
}
.who .id small.seen { display: flex; align-items: center; gap: 4px; }
.dot { width: 7px; height: 7px; border-radius: 50%; background: rgba(var(--v-theme-on-surface), 0.3); &.on { background: rgb(var(--v-theme-share)); } }
.lv-pick { min-width: 64px; }
.note { margin-top: vars.$spacing-md; @include vars.font('body-medium'); color: rgb(var(--v-theme-share)); }
.found {
    display: flex;
    align-items: center;
    gap: vars.$spacing-md;
    margin-bottom: vars.$spacing-md;
    small { display: block; opacity: 0.7; }
}
.lv {
    display: flex;
    flex-direction: column;
    small { opacity: 0.7; }
}
.msg { @include vars.font('body-medium'); margin-top: vars.$spacing-sm; }
.done { font-weight: vars.$weight-md; }
</style>
