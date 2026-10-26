<script setup lang="ts">
import { useVirtualKeyboard } from '@/common/hook/virtualKeyboard.hook';
import { useAppActions } from '@/feature/layout/hook/appActions.hook';
import FootPill from '@/feature/node/component/FootPill.vue';
import FootActions, { type FootAction } from '@/feature/widget/component/FootActions.vue';
import { computed, ref } from 'vue';
import { useOpenDevices } from '../hook/openDevices.hook';
import { useMemberBoard } from '../store/memberBoard.store';
import DeviceMoveDialog from './DeviceMoveDialog.vue';
import MemberDots from './MemberDots.vue';
import MemberMenu from './MemberMenu.vue';

// 높이 — 보드 푸터(BoardFooter SIZE)와 같게
const SIZE = { pc: { h: 44, pill: 34 }, phone: { h: 56, pill: 44 } } as const;

// 관리 페이지 아래 줄 = 보드 푸터 모양: 로고 메뉴 | Grid 도트 | 계정·로그아웃 — 화-13(전체 헤더)
const { layout } = useMemberBoard();
const { auth, toggleAuth } = useAppActions();
const kb = useVirtualKeyboard();

const phone = computed(() => layout.value.phone);
const size = computed(() => phone.value ? SIZE.phone : SIZE.pc);
const style = computed(() => ({ height: `${size.value.h}px`, '--pill-h': `${size.value.pill}px` }));

// 액션 = 판 상태로 쓸 수 있는지가 바뀌는 버튼(계정 알약 왼쪽) — REF-device-move.md
const { devices, anyOnline } = useOpenDevices();
const moveOpen = ref(false);
const actions = computed<FootAction[]>(() => [
    { key: 'move', icon: 'mdi-server-network', title: '열린 장비 supervisor 주소 변경', disabled: !anyOnline.value, hint: '열린 장비 타일에 접속 중인 instance가 없습니다', run: () => moveOpen.value = true },
]);
</script>

<template>
<!-- 키보드가 열린 동안 숨김: 높이는 그대로(빼면 Grid 영역이 바뀌어 타일이 옮겨 감 — REF-mobile-viewport.md 4) -->
<footer class="MemberFooter" :class="{ hidden: kb.open }" :style="style">
    <div class="side left">
        <foot-pill edge="left">
            <member-menu :compact="phone" />
        </foot-pill>
    </div>
    <member-dots />
    <div class="side right">
        <foot-actions :actions="actions" />
        <foot-pill edge="right" class="acct">
            <span v-if="!phone" class="nick"><span class="dot" />{{ auth?.nickname }}</span>
            <v-btn size="small" density="comfortable" variant="text" icon="mdi-logout" title="로그아웃" @click="toggleAuth" />
        </foot-pill>
    </div>
    <device-move-dialog v-model="moveOpen" :devices="devices" />
</footer>
</template>

<style scoped lang="scss">
.MemberFooter {
    flex: none;
    position: relative;
    z-index: 3;
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    align-items: center;
    padding: 0 8px;
    &.hidden { visibility: hidden; }
}
.side { display: flex; min-width: 0; }
.left { justify-self: start; }
.right { justify-self: end; align-items: center; gap: 8px; }
.nick {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 0 10px 0 14px;
    @include vars.font('label-large');
    white-space: nowrap;
    border-right: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}
.dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: rgb(var(--v-theme-primary));
}
// 닉네임 칸이 첫 칸이면 로그아웃 버튼은 FootPill의 첫 버튼 모양(넓게)을 쓰지 않는다
.acct:has(.nick) > :deep(.v-btn) { width: 44px; padding-left: 0; }
</style>
