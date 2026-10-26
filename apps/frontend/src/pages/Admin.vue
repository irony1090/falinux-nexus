<script setup lang="ts">
import MemberPage from '@/feature/member/component/MemberPage.vue';
import { useMemberUrl } from '@/feature/member/hook/memberUrl.hook';
import { provideMemberBoard } from '@/feature/member/store/memberBoard.store';

// 관리자 페이지 — 회원·장비 목록 + 상세 타일 — 회원-5(관리자 화면), REF-member-ui.md
useMemberUrl(provideMemberBoard('admin'));
</script>

<template>
<member-page />
</template>
