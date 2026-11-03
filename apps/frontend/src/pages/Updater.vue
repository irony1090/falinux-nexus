<script setup lang="ts">
import MemberPage from '@/feature/member/component/MemberPage.vue';
import { useMemberUrl } from '@/feature/member/hook/memberUrl.hook';
import { provideMemberBoard } from '@/feature/member/store/memberBoard.store';
import UpdList from '@/feature/updater/component/UpdList.vue';
import UpdPrompt, { provideUpdPrompt } from '@/feature/updater/component/UpdPrompt.vue';
import UpdTileView from '@/feature/updater/component/UpdTileView.vue';
import { bindUpdWatch } from '@/feature/updater/hook/updWatch.hook';
import { UPD_KEYS } from '@/feature/updater/util/updGrid.util';

// 업데이트 app 관리 페이지(관리자 전용) — 회원 페이지와 같은 타일 틀(MemberPage)에 updater 목록·타일. REF-update-deploy.md "화면"·"P1 착수 계획"
useMemberUrl(provideMemberBoard('updater', UPD_KEYS));
const { state } = provideUpdPrompt();
bindUpdWatch();
</script>

<template>
<member-page>
    <template #list><upd-list /></template>
    <template #tile="{ placed, cell }"><upd-tile-view :placed="placed" :style="cell" /></template>
</member-page>
<upd-prompt :state="state" @done="state = null" />
</template>
