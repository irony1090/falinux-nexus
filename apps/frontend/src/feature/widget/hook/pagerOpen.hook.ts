import { onBeforeUnmount, onMounted, ref } from 'vue'

// 터치 길게 누르기 — Q57-4(누르는 시간): 행 선택과 같은 시간, 이만큼 움직이면 취소
const LONG_PRESS_MS = 500
const PRESS_MOVE_MAX = 10

export type PagerPoint = { x: number; y: number }
export type PagerOpenOptions = {
    off: () => boolean                          // 아예 안 띄움(전체보기 중)
    keyOff: () => boolean                       // 키로 안 띄움(앱 단축키 끈 타일)
    pressable: (t: Element) => boolean          // 길게 누를 자리(버튼·입력은 이미 뺌)
    inside: (t: EventTarget | null) => boolean  // 팝업 안 — 바깥 탭에서 뺀다
    onOpen: (at: PagerPoint | null) => void     // 열기 직전 자리 잡기(키 = 마지막 커서, 터치 = 누른 자리)
}

// 커서 팝업(컨트롤 패널)을 여는 길 둘 — 보드 PagerPopup·관리 페이지 MemberPagerPopup이 같이 씀, REF-pager-tools.md 팝업-6
// 키 = Ctrl+Alt 누르는 동안(떼면 닫힘, Q53) / 터치 = 길게 누르기(바깥 탭·동작으로 닫힘, Q57-2). 키는 PC(pointer: fine)에서만
export const usePagerOpen = (opt: PagerOpenOptions) => {
    const held = ref(false)
    const touchOpen = ref(false)
    const close = () => {
        held.value = false
        touchOpen.value = false
    }

    // 마지막 마우스 위치 — 창 밖으로 나가면 모름
    let last: PagerPoint | null = null
    const onPointerMove = (e: PointerEvent) => {
        if (e.pointerType === 'mouse') last = { x: e.clientX, y: e.clientY }
    }
    const onMouseOut = (e: MouseEvent) => {
        if (!e.relatedTarget) last = null
    }

    const onKeyDown = (e: KeyboardEvent) => {
        if (opt.keyOff() || held.value || !e.ctrlKey || !e.altKey || e.shiftKey || e.metaKey) return
        if ((e.key !== 'Control' && e.key !== 'Alt') || !matchMedia('(pointer: fine)').matches) return
        touchOpen.value = false
        opt.onOpen(last)
        held.value = true
    }
    const onKeyUp = (e: KeyboardEvent) => {
        if (e.key === 'Control' || e.key === 'Alt') held.value = false
    }
    // 창 전환(Ctrl+Alt+화살표 등)으로 keyup을 못 받는 경우
    const onBlur = () => held.value = false

    let press: { timer: ReturnType<typeof setTimeout>; id: number; x: number; y: number } | null = null
    const cancelPress = () => {
        if (press) clearTimeout(press.timer)
        press = null
    }
    const pressable = (t: EventTarget | null) => {
        if (!(t instanceof Element) || opt.off()) return false
        if (t.closest('button, a, input, textarea, select, [role="button"], .v-btn')) return false
        return opt.pressable(t)
    }
    // 길게 누르기가 브라우저 기본 메뉴를 띄우지 않게(누르는 중·뜬 직후) + 손 뗄 때의 click이 막 뜬 알약을 누르지 않게
    let eatMenu = false
    let eatClick = false
    const onContextMenu = (e: Event) => {
        if (!press && !eatMenu) return
        eatMenu = false
        e.preventDefault()
    }
    const onClickCapture = (e: Event) => {
        if (!eatClick) return
        eatClick = false
        e.preventDefault()
        e.stopPropagation()
    }
    const onPointerDown = (e: PointerEvent) => {
        eatClick = false
        if (e.pointerType === 'mouse') return
        // 터치로 연 것 = 바깥 탭에 닫힘(그 탭은 원래 동작도 함) — Q57-2(닫는 법)
        if (touchOpen.value && !opt.inside(e.target)) close()
        cancelPress()
        if (!e.isPrimary || !pressable(e.target)) return
        press = {
            id: e.pointerId, x: e.clientX, y: e.clientY,
            timer: setTimeout(() => {
                const p = press
                press = null
                if (!p || opt.off()) return
                held.value = false
                opt.onOpen({ x: p.x, y: p.y })
                touchOpen.value = true
                eatMenu = true
                eatClick = true
                window.getSelection()?.removeAllRanges()
            }, LONG_PRESS_MS),
        }
    }
    const onPressMove = (e: PointerEvent) => {
        if (press && e.pointerId === press.id && Math.hypot(e.clientX - press.x, e.clientY - press.y) > PRESS_MOVE_MAX) cancelPress()
    }
    const onPressEnd = (e: PointerEvent) => {
        if (press && e.pointerId === press.id) cancelPress()
        if (eatMenu || eatClick) setTimeout(() => eatMenu = eatClick = false, 400)
    }

    onMounted(() => {
        window.addEventListener('keydown', onKeyDown, true)
        window.addEventListener('keyup', onKeyUp, true)
        window.addEventListener('blur', onBlur)
        window.addEventListener('pointermove', onPointerMove, { capture: true, passive: true })
        window.addEventListener('mouseout', onMouseOut)
        window.addEventListener('pointerdown', onPointerDown, true)
        window.addEventListener('pointermove', onPressMove, { capture: true, passive: true })
        window.addEventListener('pointerup', onPressEnd, true)
        window.addEventListener('pointercancel', onPressEnd, true)
        window.addEventListener('contextmenu', onContextMenu, true)
        window.addEventListener('click', onClickCapture, true)
    })
    onBeforeUnmount(() => {
        cancelPress()
        window.removeEventListener('keydown', onKeyDown, true)
        window.removeEventListener('keyup', onKeyUp, true)
        window.removeEventListener('blur', onBlur)
        window.removeEventListener('pointermove', onPointerMove, { capture: true })
        window.removeEventListener('mouseout', onMouseOut)
        window.removeEventListener('pointerdown', onPointerDown, true)
        window.removeEventListener('pointermove', onPressMove, { capture: true })
        window.removeEventListener('pointerup', onPressEnd, true)
        window.removeEventListener('pointercancel', onPressEnd, true)
        window.removeEventListener('contextmenu', onContextMenu, true)
        window.removeEventListener('click', onClickCapture, true)
    })

    return { held, touchOpen, close }
}
