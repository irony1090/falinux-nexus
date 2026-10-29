import { useTileGrids } from '../store/tileGrids.store'
import { useDigitShortcut } from '@/feature/widget/hook/digitShortcut.hook'

// Ctrl+Alt+숫자 = 위치 번호로 active 전환 — REF-node-ui-active.md AT4(단축키 조합)·AT5(10번 이상). 키 처리는 useDigitShortcut
export const useTileShortcut = () => {
    const { posOf, order, activate, appShortcutsOff } = useTileGrids()
    useDigitShortcut({
        off: () => appShortcutsOff.value,
        count: () => posOf.value.size,
        go: pos => {
            const t = order.value[pos - 1]
            if (t) activate(t.id)
        },
    })
}
