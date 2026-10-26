import { onBeforeUnmount, onMounted } from 'vue'

// 두 자리 번호의 둘째 숫자를 기다리는 시간
const SECOND_DIGIT_MS = 1000

export type DigitShortcutOptions = {
    off: () => boolean              // 앱 단축키 끔(appShortcutsOff)
    count: () => number             // 번호 개수
    go: (pos: number) => void       // 1부터
}

// Ctrl+Alt+숫자 = 위치 번호로 이동 — 보드 useTileShortcut·관리 페이지가 같이 씀, REF-node-ui-active.md AT4(단축키 조합)·AT5(10번 이상).
// window capture에서 멈춘다: xterm 포커스 중에도 터미널 프로그램보다 먼저 받는다
export const useDigitShortcut = (opt: DigitShortcutOptions) => {
    let pending: number | null = null
    let timer: ReturnType<typeof setTimeout> | undefined

    const commit = () => {
        clearTimeout(timer)
        if (pending !== null) opt.go(pending)
        pending = null
    }

    // 둘째 숫자가 올 수 있을 때(d0~ 번호가 있음)만 기다린다. 없는 두 자리면 새 첫 숫자로
    const onDigit = (d: number) => {
        if (pending !== null) {
            const two = pending * 10 + d
            clearTimeout(timer)
            pending = null
            if (two <= opt.count()) return opt.go(two)
        }
        if (d === 0) return
        if (d * 10 > opt.count()) return opt.go(d)
        pending = d
        timer = setTimeout(commit, SECOND_DIGIT_MS)
    }

    const onKeyDown = (e: KeyboardEvent) => {
        if (opt.off() || !e.ctrlKey || !e.altKey || e.metaKey || e.shiftKey || !/^Digit\d$/.test(e.code)) return
        e.preventDefault()
        e.stopPropagation()
        if (!e.repeat) onDigit(Number(e.code.slice(5)))
    }
    // Ctrl·Alt 중 하나를 떼면 기다리지 않고 바로
    const onKeyUp = (e: KeyboardEvent) => {
        if (e.key === 'Alt' || e.key === 'Control') commit()
    }

    onMounted(() => {
        window.addEventListener('keydown', onKeyDown, true)
        window.addEventListener('keyup', onKeyUp, true)
    })
    onBeforeUnmount(() => {
        clearTimeout(timer)
        window.removeEventListener('keydown', onKeyDown, true)
        window.removeEventListener('keyup', onKeyUp, true)
    })
}
