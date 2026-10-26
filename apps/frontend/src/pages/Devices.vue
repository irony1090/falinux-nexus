<script setup lang="ts">
import MemberPage from '@/feature/member/component/MemberPage.vue';
import { useMemberUrl } from '@/feature/member/hook/memberUrl.hook';
import { provideMemberBoard } from '@/feature/member/store/memberBoard.store';

// 내 장비 페이지 — 내 장비 목록 + 상세 타일 — 회원-3(워커 등록·키 발급), REF-member-ui.md
useMemberUrl(provideMemberBoard('mine'));
</script>

<template>
<member-page />
</template>
