import { onBeforeUnmount, reactive, ref } from 'vue'

const LONG_PRESS_MS = 500   // itemPress.hook.ts와 같은 값
const PRESS_MOVE_MAX = 10   // px — 넘게 움직이면 스크롤로 보고 취소

export type PickState = 'none' | 'some' | 'all'

// Files 타일 선택 — 행을 길게 누르면(마우스·터치) 선택 모드, 이후 짧게 누르기 = 토글(결정-1(선택 방식)). 선택 = 파일 경로 집합
// itemPress.hook.ts는 마우스 길게 누르기를 건너뛰고 고름·더블클릭이 섞여 있어 따로 둔다
export const useHoldSelect = () => {
    const selecting = ref(false)
    const picked = reactive(new Set<string>())
    let press: { timer: ReturnType<typeof setTimeout>; x: number; y: number } | null = null
    let skipClick = false   // 길게 누르기 뒤따르는 click 무시
    let pointer = 'mouse'

    const cancelPress = () => {
        if (press) clearTimeout(press.timer)
        press = null
    }
    onBeforeUnmount(cancelPress)

    // onHold = 길게 누른 순간 할 일. 행 안의 버튼에서 누른 것은 무시
    const onPointerDown = (e: PointerEvent, onHold: () => void) => {
        pointer = e.pointerType
        skipClick = false
        cancelPress()
        if (e.button !== 0 || (e.target as HTMLElement).closest('button')) return
        press = {
            x: e.clientX,
            y: e.clientY,
            timer: setTimeout(() => {
                press = null
                skipClick = true
                onHold()
            }, LONG_PRESS_MS),
        }
    }
    const onPointerMove = (e: PointerEvent) => {
        if (press && Math.hypot(e.clientX - press.x, e.clientY - press.y) > PRESS_MOVE_MAX) cancelPress()
    }
    // 터치 길게 누르기의 기본 메뉴만 막는다(마우스 오른쪽 클릭은 그대로)
    const onContextMenu = (e: MouseEvent) => { if (pointer !== 'mouse') e.preventDefault() }
    // click 처리 맨 앞에서 부른다 — true면 길게 누르기의 뒤따르는 click이니 아무것도 하지 않는다
    const consumeClick = () => {
        const s = skipClick
        skipClick = false
        return s
    }

    const state = (paths: string[]): PickState => {
        const n = paths.filter(p => picked.has(p)).length
        return n === 0 ? 'none' : n === paths.length ? 'all' : 'some'
    }
    // 전부 체크돼 있으면 전부 해제, 아니면 전부 체크(폴더 = 하위 전체)
    const toggle = (paths: string[]) => {
        const off = state(paths) === 'all'
        for (const p of paths) off ? picked.delete(p) : picked.add(p)
    }
    const start = (paths: string[]) => {
        selecting.value = true
        for (const p of paths) picked.add(p)
    }
    const clear = () => {
        selecting.value = false
        picked.clear()
    }
    // 목록에서 사라진 경로(지워짐·다른 탭에서 바뀜)를 뺀다
    const keepOnly = (paths: Iterable<string>) => {
        const have = new Set(paths)
        for (const p of [...picked]) if (!have.has(p)) picked.delete(p)
    }

    return { selecting, picked, onPointerDown, onPointerMove, cancelPress, onContextMenu, consumeClick, state, toggle, start, clear, keepOnly }
}
