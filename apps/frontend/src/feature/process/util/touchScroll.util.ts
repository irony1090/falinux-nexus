import type { Terminal } from '@xterm/xterm'

const TAP_SLOP = 10           // px — 이보다 덜 움직이면 탭(포커스·키보드)으로 둔다
const VELOCITY_WINDOW = 100   // ms — 손 뗄 때 속도를 이 구간으로 잰다
const FRICTION = 0.95         // 관성: 16ms마다 속도에 곱한다
const MIN_SPEED = 0.05        // px/ms — 이보다 느리면 관성을 멈춘다
const TERM_GAIN = 2           // 터미널로 넘기는 양 = 손가락 이동 × 이 값(크기 고정 상자는 손가락 그대로) — 팝업-4(터치 스크롤 배율), Q48 실기기로 정함

// bindTouchScroll은 터미널 화면을 쓸어 스크롤하게 한다(REF-term-touch.md 안 나(쓸어서 스크롤)). 반환 = 해제 함수.
// 세로 이동은 scrollBox()면 el 상자 스크롤을 먼저 쓰고, 남은 양을 터미널에 넘긴다:
// 정상 화면 = scrollLines로 스크롤백 / alt 화면·마우스 모드 = 휠 이벤트(xterm이 화살표 키·휠 보고로 PC 휠과 같게 처리).
// 정상 화면에 휠 이벤트를 못 쓰는 이유: 그 경로(VS Code 스크롤 요소)는 옛 wheelDeltaY를 읽는데 스크립트로 만든 이벤트는 늘 0이다.
// scrollBox()가 false일 때 el을 건드리면 안 된다: overflow hidden이어도 scrollTop은 바뀌어 화면이 밀린다.
export const bindTouchScroll = (el: HTMLElement, getTerm: () => Terminal | undefined, scrollBox: () => boolean): (() => void) => {
    let active = false
    let dragging = false
    let startX = 0, startY = 0, lastX = 0, lastY = 0
    let samples: { t: number; y: number }[] = []
    let target: Element | null = null
    let raf = 0
    let rest = 0   // 한 줄이 안 되는 px(정상 화면 scrollLines용)

    const stopInertia = () => {
        cancelAnimationFrame(raf)
        raf = 0
    }

    const scroll = (dy: number) => {
        if (scrollBox()) {
            const before = el.scrollTop
            el.scrollTop = before + dy
            // 끝에 닿았을 때만 남은 양을 넘긴다 — scrollTop 정수 반올림 찌꺼기가 쌓여 스크롤백이 먼저 움직이지 않게
            const atEdge = el.scrollTop <= 0 || el.scrollTop >= el.scrollHeight - el.clientHeight - 1
            dy = atEdge ? dy - (el.scrollTop - before) : 0
        }
        const term = getTerm()
        if (dy === 0 || !term?.element) return
        dy *= TERM_GAIN
        if (term.buffer.active.type === 'normal' && term.modes.mouseTrackingMode === 'none') {
            const cell = (term.element.querySelector('.xterm-screen')?.clientHeight ?? 0) / term.rows
            if (!(cell > 0)) return
            rest += dy
            const lines = Math.trunc(rest / cell)
            rest -= lines * cell
            if (lines) term.scrollLines(lines)
            return
        }
        const t = target?.isConnected ? target : term.element.querySelector('.xterm-screen')
        t?.dispatchEvent(new WheelEvent('wheel', {
            deltaY: dy, deltaMode: WheelEvent.DOM_DELTA_PIXEL,
            clientX: lastX, clientY: lastY,   // 마우스 모드 앱에 보고되는 위치
            bubbles: true, cancelable: true,
        }))
    }

    const onStart = (e: TouchEvent) => {
        stopInertia()
        const p = e.touches[0]
        target = e.target instanceof Element ? e.target : null
        // 스크롤바는 자체 pointer 드래그(touch-action: none)가 맡는다
        active = e.touches.length === 1 && !!p && !target?.closest('.scrollbar')
        if (!active || !p) return
        dragging = false
        rest = 0
        startX = lastX = p.clientX
        startY = lastY = p.clientY
        samples = [{ t: e.timeStamp, y: p.clientY }]
    }

    const onMove = (e: TouchEvent) => {
        const p = e.touches[0]
        if (!active || !p || e.touches.length !== 1) {
            active = false
            return
        }
        if (!dragging) {
            const dx = p.clientX - startX, dy = p.clientY - startY
            if (Math.hypot(dx, dy) < TAP_SLOP) return
            // 가로 쓸기는 브라우저(touch-action: pan-x, 크기 고정 상자의 가로 스크롤)에 맡긴다
            if (Math.abs(dx) > Math.abs(dy)) {
                active = false
                return
            }
            dragging = true
            lastY = p.clientY   // 탭 판정 거리만큼 튀지 않게 여기서부터 잰다
        }
        if (e.cancelable) e.preventDefault()   // 쓸기 뒤 click(포커스) 막음
        const dy = lastY - p.clientY
        lastX = p.clientX
        lastY = p.clientY
        samples.push({ t: e.timeStamp, y: p.clientY })
        samples = samples.filter(s => s.t >= e.timeStamp - VELOCITY_WINDOW)
        scroll(dy)
    }

    const onEnd = (e: TouchEvent) => {
        const fling = active && dragging && e.type === 'touchend'
        active = false
        if (!fling) return
        const recent = samples.filter(s => s.t >= e.timeStamp - VELOCITY_WINDOW)
        const first = recent[0], last = recent[recent.length - 1]
        if (!first || !last || last.t <= first.t) return
        let v = (first.y - last.y) / (last.t - first.t)   // px/ms, 양수 = 아래(뒤쪽 출력)로
        if (Math.abs(v) < MIN_SPEED) return
        let prev = performance.now()
        const step = (now: number) => {
            const dt = now - prev
            prev = now
            scroll(v * dt)
            v *= FRICTION ** (dt / 16)
            raf = Math.abs(v) >= MIN_SPEED ? requestAnimationFrame(step) : 0
        }
        raf = requestAnimationFrame(step)
    }

    el.addEventListener('touchstart', onStart, { passive: true })
    el.addEventListener('touchmove', onMove, { passive: false })
    el.addEventListener('touchend', onEnd)
    el.addEventListener('touchcancel', onEnd)
    return () => {
        stopInertia()
        el.removeEventListener('touchstart', onStart)
        el.removeEventListener('touchmove', onMove)
        el.removeEventListener('touchend', onEnd)
        el.removeEventListener('touchcancel', onEnd)
    }
}
