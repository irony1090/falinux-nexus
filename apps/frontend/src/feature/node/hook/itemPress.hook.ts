import { onBeforeUnmount, ref, type Ref } from 'vue';

const LONG_PRESS_MS = 500;
const PRESS_MOVE_MAX = 10;     // px — 넘게 움직이면 스크롤로 보고 취소
const DOUBLE_MS = 300;         // 같은 칸을 이 안에 두 번 = 더블클릭·더블탭

type Opts<T, K> = {
    key: (t: T) => K
    selecting: () => boolean   // 선택 모드 중이면 누르기 = 체크 토글
    check: (t: T) => void
    open: (t: T) => void
}

// 목록 칸 누르기 — 파트(목록 아이콘 보기), REF-dir-grid.md "누르기 규칙"
// 한 번 = 고름(focused) · 더블 = open · Ctrl(⌘)+클릭·길게 누르기(터치) = check. 칸 안의 버튼은 평소엔 각자 처리
// 더블탭은 click 간격으로 직접 잰다 — 폰 브라우저 dblclick은 확대와 겹친다(칸에 touch-action: manipulation)
export const useItemPress = <T, K>(o: Opts<T, K>) => {
    const focused = ref<K | null>(null) as Ref<K | null>;
    const dragOk = ref(false);     // 마우스로 누를 때만 draggable — SD3(터치 드래그 제외), 길게 누르기와 겹치지 않게
    let press: { timer: ReturnType<typeof setTimeout>; x: number; y: number } | null = null;
    let skipClick = false;         // 길게 누르기 뒤따르는 click 무시
    let pointer = 'mouse';
    let last: { key: K; at: number } | null = null;

    const cancelPress = () => {
        if (press) clearTimeout(press.timer);
        press = null;
    }
    onBeforeUnmount(cancelPress);

    const onPointerDown = (e: PointerEvent, t: T) => {
        pointer = e.pointerType;
        dragOk.value = e.pointerType === 'mouse';
        skipClick = false;
        cancelPress();
        if (e.pointerType === 'mouse') return;
        press = {
            x: e.clientX,
            y: e.clientY,
            timer: setTimeout(() => {
                press = null;
                skipClick = true;
                last = null;
                o.check(t);
            }, LONG_PRESS_MS),
        };
    }
    const onPointerMove = (e: PointerEvent) => {
        if (press && Math.hypot(e.clientX - press.x, e.clientY - press.y) > PRESS_MOVE_MAX) cancelPress();
    }

    const onClick = (e: MouseEvent, t: T) => {
        if (skipClick) {
            skipClick = false;
            return;
        }
        // 선택 모드 = 칸 안 어디든 체크(위쪽 버튼은 숨겨져 있다). Ctrl은 상태 라벨 위여도 체크
        const el = e.target as HTMLElement;
        const onBtn = !!el.closest('button, .v-btn');
        if (o.selecting() || ((!onBtn || !!el.closest('.dg-lbl')) && (e.ctrlKey || e.metaKey))) {
            last = null;
            o.check(t);
            return;
        }
        if (onBtn) return;
        const k = o.key(t), now = e.timeStamp;
        focused.value = k;
        if (last && last.key === k && now - last.at < DOUBLE_MS) {
            last = null;
            o.open(t);
            return;
        }
        last = { key: k, at: now };
    }

    return { focused, dragOk, onPointerDown, onPointerMove, cancelPress, onClick, lastPointer: () => pointer };
}
