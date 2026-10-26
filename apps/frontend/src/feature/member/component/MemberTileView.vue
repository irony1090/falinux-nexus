<script setup lang="ts">
import { useResizeCallback } from '@/feature/common/store/resizeGroup.store';
import { computed, ref } from 'vue';
import { VIcon, VMenu } from 'vuetify/components';
import { useMemberUser } from '../api/device.api';
import { useMemberBoard } from '../store/memberBoard.store';
import { kindOf, mainKeyOf, userIdOf, type Placed } from '../util/memberGrid.util';
import MemberDeviceTile from './MemberDeviceTile.vue';
import MemberNewTile from './MemberNewTile.vue';
import MemberSizeCard from './MemberSizeCard.vue';
import MemberUserTile from './MemberUserTile.vue';

// 상세 타일 공통 틀 — 머리 줄 = 위치 번호 + 경로 제목(화-11) + [크기 | 닫기] 끝 블록(보드 TileFrame과 같음, 화-10). 자식 타일은 옅은 파랑 테두리(화-9)
const props = defineProps<{ placed: Placed }>();
const { admin, posOf, activeKey, setActive, close, openParent, pickTab } = useMemberBoard();

const tile = computed(() => props.placed.tile);
const kind = computed(() => kindOf(tile.value.key));
const parentId = computed(() => tile.value.parent ? userIdOf(tile.value.parent) : undefined);

// 경로 이름 — 회원 닉네임은 그 회원 조회(같은 키라 회원 타일과 캐시를 같이 쓴다)
const userId = computed(() => kind.value === 'user' ? userIdOf(tile.value.key) : parentId.value);
const userQuery = admin && userId.value !== undefined ? useMemberUser(userId.value) : undefined;
const nick = computed(() => userQuery?.data.value?.nickname ?? (userId.value !== undefined ? `#${userId.value}` : ''));

type Crumb = { label: string; go?: () => void };
const crumbs = computed<Crumb[]>(() => {
    const t = tile.value;
    const root: Crumb = kind.value === 'user' || t.parent
        ? { label: '회원', go: () => pickTab('users') }
        : { label: admin ? '장비' : '내 장비', go: admin ? () => pickTab('devices') : undefined };
    const out = [root];
    if (t.parent) out.push({ label: nick.value, go: () => openParent(t) });
    out.push({ label: kind.value === 'user' ? nick.value : kind.value === 'device' ? mainKeyOf(t.key) : admin ? '대신 등록' : '장비 등록' });
    return out;
});

const active = computed(() => activeKey.value === tile.value.key);

// 크기 팝업은 크기가 바뀌면 버튼 자리를 따라간다(보드 relocateSizeMenu와 같음)
const frameEl = ref<HTMLElement>();
const sizeMenu = ref<InstanceType<typeof VMenu>>();
const relocateSizeMenu = () => requestAnimationFrame(() => sizeMenu.value?.updateLocation?.(new Event('resize')));
useResizeCallback(frameEl, relocateSizeMenu);
</script>

<template>
<section ref="frameEl" class="MemberTileView mb-tile" :class="{ child: !!tile.parent, active }" :data-tile-key="tile.key" tabindex="-1"
    :aria-label="kind === 'user' ? '회원 타일' : kind === 'device' ? '장비 타일' : '장비 등록 타일'"
    @pointerdown.capture="setActive(tile.key)"
>
    <div class="mb-head">
        <span class="pos">{{ posOf.get(tile.key) }}</span>
        <nav class="crumbs" aria-label="경로">
            <template v-for="(c, i) in crumbs" :key="i">
                <span v-if="i > 0" class="sep" aria-hidden="true">›</span>
                <span v-if="i === crumbs.length - 1" class="cur">{{ c.label }}</span>
                <button v-else type="button" class="link" :disabled="!c.go" @click="c.go?.()">{{ c.label }}</button>
            </template>
        </nav>
        <span class="tile-btn-end close" :class="{ 'tile-btn-end--primary': active }">
            <v-menu ref="sizeMenu" location="bottom end" :close-on-content-click="false">
                <template #activator="{ props: act }">
                    <button v-bind="act" type="button" class="tile-btn" :title="`크기 ${placed.cw}×${placed.ch}칸`" aria-label="크기">
                        <v-icon size="16" icon="mdi-resize" />
                    </button>
                </template>
                <member-size-card :tile-key="tile.key" @changed="relocateSizeMenu" />
            </v-menu>
            <button type="button" class="tile-btn" title="닫기" aria-label="타일 닫기" @click="close(tile.key)"><v-icon icon="mdi-close" size="16" /></button>
        </span>
    </div>
    <member-user-tile v-if="kind === 'user'" :tile="tile" />
    <member-device-tile v-else-if="kind === 'device'" :tile="tile" />
    <member-new-tile v-else :tile="tile" />
</section>
</template>

<style scoped lang="scss">
.MemberTileView {
    &.child { border-color: rgba(var(--v-theme-primary), 0.35); }
    // active — 보드 TileFrame과 같은 바깥 선
    &.active { outline: 2px solid rgb(var(--v-theme-primary)); outline-offset: 1px; }
    // 알약 타일 이동 뒤 포커스(키 입력 대상)만 받는다 — 포커스 선은 active 선으로 대신
    &:focus:not(.active) { outline: none; }
}
.crumbs {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 2px;
    padding: 0 6px 0 6px;
    @include vars.font('body-medium');
    white-space: nowrap;
    overflow: hidden;
}
.sep { flex: none; padding: 0 2px; opacity: 0.45; }
.link {
    flex: none;
    height: 26px;
    padding: 0 8px;
    border: 0;
    border-radius: 999px;
    background: rgba(var(--v-theme-on-surface), 0.06);
    color: rgba(var(--v-theme-on-surface), 0.8);
    font: inherit;
    @include vars.font('label-medium');
    font-weight: vars.$weight-md;
    cursor: pointer;
    &:hover:not(:disabled) { background: rgba(var(--v-theme-on-surface), 0.12); }
    &:disabled { cursor: default; }
}
.cur {
    min-width: 0;
    padding: 0 4px;
    font-weight: vars.$weight-lg;
    overflow: hidden;
    text-overflow: ellipsis;
}
// 위치 번호 — 보드 TileFrame .pos와 같은 모양
.pos {
    flex: none;
    align-self: center;
    min-width: 16px;
    height: 16px;
    margin-left: 10px;
    padding: 0 4px;
    border-radius: 4px;
    display: inline-grid;
    place-items: center;
    @include vars.font('label-small');
    font-weight: vars.$weight-lg;
    background: rgb(var(--v-theme-on-surface));
    color: rgb(var(--v-theme-surface));
}
// 닫기 = 오른쪽 벽 끝 블록(찰떡) — 오른쪽 면 직각
.close { border-radius: 999px 0 0 999px; }
</style>
