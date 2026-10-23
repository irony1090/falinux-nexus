import { onBeforeUnmount, onMounted } from 'vue'
import { useTileGrids } from '../store/tileGrids.store'
import { isNavTile, useTileTree } from '../store/tileTree.store'

// 입력 칸에서는 Ctrl+Shift+←/→ = 단어 선택이라 가로채지 않는다
const typing = (el: EventTarget | null) =>
    el instanceof HTMLElement && (el.isContentEditable || el.tagName === 'INPUT' || el.tagName === 'TEXTAREA')

// Ctrl+Shift+←/→ = active 타일(폴더·장비) 안 뒤로·앞으로 — 뒤로-1(타일 안 이동 기록) Q19(단축키)
export const useTileNavShortcut = () => {
    const { activeId, appShortcutsOff } = useTileGrids()
    const { tiles, goHistory } = useTileTree()

    const onKeyDown = (e: KeyboardEvent) => {
        if (appShortcutsOff.value || !e.ctrlKey || !e.shiftKey || e.altKey || e.metaKey) return
        if (e.key !== 'ArrowLeft' && e.key !== 'ArrowRight') return
        const id = activeId.value
        if (!id || !isNavTile(tiles.value[id]) || typing(e.target)) return
        e.preventDefault()
        e.stopPropagation()
        goHistory(id, e.key === 'ArrowLeft' ? -1 : 1)
    }

    onMounted(() => window.addEventListener('keydown', onKeyDown, true))
    onBeforeUnmount(() => window.removeEventListener('keydown', onKeyDown, true))
}
