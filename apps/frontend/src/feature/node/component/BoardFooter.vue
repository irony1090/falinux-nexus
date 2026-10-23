<script setup lang="ts">
import { useVirtualKeyboard } from '@/common/hook/virtualKeyboard.hook';
import { useAppActions } from '@/feature/layout/hook/appActions.hook';
import { useResizeCallback } from '@/feature/common/store/resizeGroup.store';
import { computed, ref } from 'vue';
import { VBadge, VBtn, VIcon, VMenu } from 'vuetify/components';
import { useTileGrids } from '../store/tileGrids.store';
import { useTileNotices } from '../store/tileNotice.store';
import BoardMenu from './BoardMenu.vue';
import FootPill from './FootPill.vue';
import TileNav from './TileNav.vue';
import TransferButton from '@/feature/worker/component/TransferButton.vue';
import { useTransfers } from '@/feature/worker/store/transfers.store';

// 푸터 높이·알약 높이 — 폰(가로 1칸)은 손가락 크기
const SIZE = { pc: { h: 44, pill: 34 }, phone: { h: 56, pill: 44 } } as const;
// 좌우 여백·구역 사이 최소 간격
const PAD = 8;
const GAP = 8;

// 보드 화면 맨 아래 한 줄(옛 헤더 대신): 보드 메뉴 | 인디케이터 + 전체보기 | 알림·전체 화면·로그아웃 — REF-board-chrome.md 배치 F(확정안)
const { cap } = useTileGrids();
const kb = useVirtualKeyboard();
const { unread, drawerOpen } = useTileNotices();
const { running } = useTransfers();
const { auth, toggleAuth, canFullscreen, fullscreen, toggleFullscreen } = useAppActions();

const phone = computed(() => cap.value.x === 1);
const size = computed(() => phone.value ? SIZE.phone : SIZE.pc);
const unreadText = computed(() => unread.value > 99 ? '99+' : unread.value);

// 인디케이터를 화면 가운데에 두려고 가운데 폭 = 전체 - 양쪽 중 넓은 쪽 x2. 인디케이터는 이 폭으로 도트 수를 정한다
const rootEl = ref<HTMLElement>();
const leftEl = ref<HTMLElement>();
const rightEl = ref<HTMLElement>();
const w = ref({ root: 0, left: 0, right: 0 });
const sync = () => w.value = {
    root: rootEl.value?.clientWidth ?? 0,
    left: leftEl.value?.offsetWidth ?? 0,
    right: rightEl.value?.offsetWidth ?? 0,
};
useResizeCallback(rootEl, sync);
useResizeCallback(leftEl, sync);
useResizeCallback(rightEl, sync);
const midW = computed(() => Math.max(0, w.value.root - 2 * PAD - 2 * (Math.max(w.value.left, w.value.right) + GAP)));

const style = computed(() => ({ height: `${size.value.h}px`, '--pill-h': `${size.value.pill}px`, padding: `0 ${PAD}px` }));
</script>

<template>
<!-- 키보드가 열린 동안 숨김: 높이는 그대로 둔다(빼면 Grid 영역이 바뀌어 타일이 옮겨 감 — REF-mobile-viewport.md 4) -->
<footer ref="rootEl" class="BoardFooter" :class="{ hidden: kb.open }" :style="style">
    <div ref="leftEl" class="side left">
        <foot-pill>
            <board-menu :compact="phone" />
        </foot-pill>
    </div>
    <div class="mid" :style="{ width: `${midW}px` }">
        <tile-nav />
    </div>
    <div ref="rightEl" class="side right">
        <foot-pill v-if="!phone">
            <transfer-button />
            <v-btn size="small" density="comfortable" variant="text" icon title="위치 알림" @click="drawerOpen = !drawerOpen">
                <v-badge v-if="unread" :content="unreadText" color="error" floating>
                    <v-icon icon="mdi-bell-outline" />
                </v-badge>
                <v-icon v-else icon="mdi-bell-outline" />
            </v-btn>
            <v-btn v-if="canFullscreen" size="small" density="comfortable" variant="text"
                :icon="fullscreen ? 'mdi-fullscreen-exit' : 'mdi-fullscreen'"
                :title="fullscreen ? '전체 화면 끝내기' : '전체 화면'"
                @click="toggleFullscreen"
            />
            <v-btn size="small" density="comfortable" variant="text" :icon="auth ? 'mdi-logout' : 'mdi-login'" :title="auth ? '로그아웃' : '로그인'" @click="toggleAuth" />
        </foot-pill>
        <!-- 폰: 하나로 합쳐 위로 펼침, 안 읽은 알림은 빨간 점 -->
        <foot-pill v-else>
            <v-menu location="top center" offset="10">
                <template #activator="{ props: act, isActive }">
                    <v-btn v-bind="act" density="comfortable" variant="text" icon :active="isActive" title="더 보기" aria-label="더 보기 - 전송, 알림, 전체 화면, 로그아웃">
                        <v-icon icon="mdi-dots-horizontal" />
                        <span v-if="(unread || running) && !isActive" class="dot" />
                    </v-btn>
                </template>
                <foot-pill class="stack">
                    <transfer-button dense />
                    <v-btn density="comfortable" variant="text" icon title="위치 알림" @click="drawerOpen = true">
                        <v-badge v-if="unread" :content="unreadText" color="error" floating>
                            <v-icon icon="mdi-bell-outline" />
                        </v-badge>
                        <v-icon v-else icon="mdi-bell-outline" />
                    </v-btn>
                    <v-btn v-if="canFullscreen" density="comfortable" variant="text"
                        :icon="fullscreen ? 'mdi-fullscreen-exit' : 'mdi-fullscreen'"
                        :title="fullscreen ? '전체 화면 끝내기' : '전체 화면'"
                        @click="toggleFullscreen"
                    />
                    <v-btn density="comfortable" variant="text" :icon="auth ? 'mdi-logout' : 'mdi-login'" :title="auth ? '로그아웃' : '로그인'" @click="toggleAuth" />
                </foot-pill>
            </v-menu>
        </foot-pill>
    </div>
</footer>
</template>

<style scoped lang="scss">
.BoardFooter {
    flex: none;
    position: relative;
    z-index: 3;
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    align-items: center;
    &.hidden { visibility: hidden; }
}
.side { display: flex; }
.left { justify-self: start; }
.right { justify-self: end; }
.mid {
    min-width: 0;
    display: flex;
}
.dot {
    position: absolute;
    top: 8px;
    right: 8px;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: rgb(var(--v-theme-error));
}
// 위로 펼친 세로 알약
.stack {
    flex-direction: column;
    height: auto;
    padding: 2px 0;
    gap: 2px;
}
</style>
