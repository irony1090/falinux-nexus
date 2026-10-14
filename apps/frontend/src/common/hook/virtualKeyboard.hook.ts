import { onBeforeUnmount, onMounted, readonly, reactive } from 'vue'

// 가상 키보드 상태(visualViewport 기준) + 포커스된 터미널 — REF-mobile-keyboard.md 키보드-1(키보드 상태)
// 좌표는 레이아웃 뷰포트 기준(getBoundingClientRect·position: fixed와 같은 기준)
// 라-3(열림 판정): resizes-content에선 레이아웃도 같이 줄어 "레이아웃 - 보이는 높이"가 0이다.
// 그래서 키보드가 없을 때의 보이는 높이(baseH)를 기억해 두고 그보다 OPEN_MIN 넘게 줄면 열림으로 본다(iOS처럼 보이는 영역만 줄어도 같은 식)

const OPEN_MIN = 120    // 주소창 접힘·Edge 전체 화면 끝내기(110) 같은 키보드 아닌 변화는 무시
const RESET_MS = 1000   // 전체 화면 전환 직후엔 높이가 여러 번 바뀌므로 이 동안은 기준 높이를 따라간다

const state = reactive({
    open: false,
    top: 0,         // 보이는 영역 위 끝
    bottom: 0,      // 보이는 영역 아래 끝(= 키보드 위 끝)
    baseH: 0,       // 키보드가 없을 때의 보이는 높이 — 열린 동안 타일 영역 높이를 이 값으로 고정(라-2(높이 고정))
    termUid: null as string | null,
})

let lastW = -1
let resetUntil = 0

const update = () => {
    const vv = window.visualViewport
    const layoutH = document.documentElement.clientHeight
    if (!vv) {
        Object.assign(state, { open: false, top: 0, bottom: layoutH, baseH: layoutH })
        return
    }
    const visible = vv.height * vv.scale     // 핀치 확대는 scale로 되돌려 키보드만 센다
    const reset = innerWidth !== lastW || performance.now() < resetUntil     // 폴드 펴기·회전·전체 화면 전환
    lastW = innerWidth
    state.open = !reset && visible < state.baseH - OPEN_MIN
    if (!state.open) state.baseH = visible
    state.top = vv.offsetTop
    state.bottom = vv.offsetTop + vv.height
}

const onFullscreen = () => {
    resetUntil = performance.now() + RESET_MS
    update()
}

const termOf = (el: EventTarget | null) =>
    el instanceof Element ? el.closest<HTMLElement>('[data-term-uid]')?.dataset.termUid ?? null : null

const onFocusIn = (ev: FocusEvent) => { state.termUid = termOf(ev.target) }
const onFocusOut = (ev: FocusEvent) => { if (!termOf(ev.relatedTarget)) state.termUid = null }

let users = 0
const start = () => {
    if (users++ > 0) return
    update()
    window.visualViewport?.addEventListener('resize', update)
    window.visualViewport?.addEventListener('scroll', update)
    window.addEventListener('resize', update)
    document.addEventListener('fullscreenchange', onFullscreen)
    document.addEventListener('focusin', onFocusIn)
    document.addEventListener('focusout', onFocusOut)
}
const stop = () => {
    if (--users > 0) return
    window.visualViewport?.removeEventListener('resize', update)
    window.visualViewport?.removeEventListener('scroll', update)
    window.removeEventListener('resize', update)
    document.removeEventListener('fullscreenchange', onFullscreen)
    document.removeEventListener('focusin', onFocusIn)
    document.removeEventListener('focusout', onFocusOut)
}

export const useVirtualKeyboard = () => {
    onMounted(start)
    onBeforeUnmount(stop)
    return readonly(state)
}
