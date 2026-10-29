import { onBeforeUnmount, onMounted, ref } from 'vue'

// 두 손가락 펼치기 — 처음 거리보다 이만큼 벌어지면 연다
const PINCH_MIN = 40

export type PagerPoint = { x: number; y: number }
export type PagerOpenOptions = {
    off: () => boolean                          // 아예 안 띄움(전체보기 중)
    keyOff: () => boolean                       // 키로 안 띄움(앱 단축키 끈 타일)
    pinchable: (t: Element) => boolean          // 펼치기를 받을 자리(두 손가락 모두)
    inside: (t: EventTarget | null) => boolean  // 팝업 안 — 바깥 탭에서 뺀다
    onOpen: (at: PagerPoint | null) => void     // 열기 직전 자리 잡기(키 = 마지막 커서, 터치 = 두 손가락 가운데)
}

// 커서 팝업(컨트롤 패널)을 여는 길 둘 — 보드 PagerPopup·관리 페이지 MemberPagerPopup이 같이 씀, REF-pager-tools.md 팝업-6
// 키 = Ctrl+Alt 누르는 동안(떼면 닫힘, Q53) / 터치 = 두 손가락 펼치기(바깥 탭·동작으로 닫힘) — 제스처-1(펼치기 패널), 길게 누르기는 타일 끌기로 넘김(REF-tile-drag.md)
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

    // 터치로 연 것 = 바깥 탭에 닫힘(그 탭은 원래 동작도 함) — Q57-2(닫는 법)
    const onPointerDown = (e: PointerEvent) => {
        if (e.pointerType !== 'mouse' && touchOpen.value && !opt.inside(e.target)) close()
    }

    // 펼치기는 touch 이벤트로 잰다 — 브라우저가 두 손가락 이동을 스크롤로 가져가면 pointer 이벤트는 pointercancel로 끊긴다
    let pinch: { d0: number } | null = null
    const span = (t: TouchList) => Math.hypot(t[0]!.clientX - t[1]!.clientX, t[0]!.clientY - t[1]!.clientY)
    const onTouchStart = (e: TouchEvent) => {
        pinch = null
        if (e.touches.length !== 2 || opt.off()) return
        const ok = [...e.touches].every(t => t.target instanceof Element && opt.pinchable(t.target))
        if (ok) pinch = { d0: span(e.touches) }
    }
    const onTouchMove = (e: TouchEvent) => {
        if (!pinch || e.touches.length !== 2 || span(e.touches) - pinch.d0 < PINCH_MIN) return
        pinch = null
        if (opt.off()) return
        const [a, b] = [e.touches[0]!, e.touches[1]!]
        held.value = false
        opt.onOpen({ x: (a.clientX + b.clientX) / 2, y: (a.clientY + b.clientY) / 2 })
        touchOpen.value = true
    }
    const onTouchEnd = (e: TouchEvent) => {
        if (e.touches.length < 2) pinch = null
    }

    onMounted(() => {
        window.addEventListener('keydown', onKeyDown, true)
        window.addEventListener('keyup', onKeyUp, true)
        window.addEventListener('blur', onBlur)
        window.addEventListener('pointermove', onPointerMove, { capture: true, passive: true })
        window.addEventListener('mouseout', onMouseOut)
        window.addEventListener('pointerdown', onPointerDown, true)
        window.addEventListener('touchstart', onTouchStart, { capture: true, passive: true })
        window.addEventListener('touchmove', onTouchMove, { capture: true, passive: true })
        window.addEventListener('touchend', onTouchEnd, { capture: true, passive: true })
        window.addEventListener('touchcancel', onTouchEnd, { capture: true, passive: true })
    })
    onBeforeUnmount(() => {
        window.removeEventListener('keydown', onKeyDown, true)
        window.removeEventListener('keyup', onKeyUp, true)
        window.removeEventListener('blur', onBlur)
        window.removeEventListener('pointermove', onPointerMove, { capture: true })
        window.removeEventListener('mouseout', onMouseOut)
        window.removeEventListener('pointerdown', onPointerDown, true)
        window.removeEventListener('touchstart', onTouchStart, { capture: true })
        window.removeEventListener('touchmove', onTouchMove, { capture: true })
        window.removeEventListener('touchend', onTouchEnd, { capture: true })
        window.removeEventListener('touchcancel', onTouchEnd, { capture: true })
    })

    return { held, touchOpen, close }
}
