<script setup lang="ts">
import { computed } from 'vue';
import { useMemberUser } from '../api/device.api';
import { useMemberBoard } from '../store/memberBoard.store';
import { memberChip } from '../store/memberDrag.store';
import { kindOf, mainKeyOf, userIdOf, type Placed } from '../util/memberGrid.util';
import MbTileFrame, { type MbCrumb } from './MbTileFrame.vue';
import MemberDeviceTile from './MemberDeviceTile.vue';
import MemberNewTile from './MemberNewTile.vue';
import MemberUserTile from './MemberUserTile.vue';

// 회원 페이지 상세 타일 — 경로·끌기 표시·몸통(회원/장비/등록). 틀 = MbTileFrame
const props = defineProps<{ placed: Placed }>();
const { admin, openParent, pickTab } = useMemberBoard();

const tile = computed(() => props.placed.tile);
const kind = computed(() => kindOf(tile.value.key));
const parentId = computed(() => tile.value.parent ? userIdOf(tile.value.parent) : undefined);

// 경로 이름 — 회원 닉네임은 그 회원 조회(같은 키라 회원 타일과 캐시를 같이 쓴다)
const userId = computed(() => kind.value === 'user' ? userIdOf(tile.value.key) : parentId.value);
const userQuery = admin && userId.value !== undefined ? useMemberUser(userId.value) : undefined;
const nick = computed(() => userQuery?.data.value?.nickname ?? (userId.value !== undefined ? `#${userId.value}` : ''));

const name = computed(() => kind.value === 'user' ? nick.value : kind.value === 'device' ? mainKeyOf(tile.value.key) : admin ? '대신 등록' : '장비 등록');
const crumbs = computed<MbCrumb[]>(() => {
    const t = tile.value;
    const root: MbCrumb = kind.value === 'user' || t.parent
        ? { label: '회원', go: () => pickTab('users') }
        : { label: admin ? '장비' : '내 장비', go: admin ? () => pickTab('devices') : undefined };
    const out = [root];
    if (t.parent) out.push({ label: nick.value, go: () => openParent(t) });
    out.push({ label: name.value });
    return out;
});
const chip = computed(() => memberChip(kind.value, name.value));
const label = computed(() => kind.value === 'user' ? '회원 타일' : kind.value === 'device' ? '장비 타일' : '장비 등록 타일');
</script>

<template>
<mb-tile-frame class="MemberTileView" :placed="placed" :crumbs="crumbs" :chip="chip" :label="label">
    <member-user-tile v-if="kind === 'user'" :tile="tile" />
    <member-device-tile v-else-if="kind === 'device'" :tile="tile" />
    <member-new-tile v-else :tile="tile" />
</mb-tile-frame>
</template>
