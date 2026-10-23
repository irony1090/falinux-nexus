import { onBeforeUnmount, onMounted } from 'vue'
import { useTileGrids } from '../store/tileGrids.store'

// Ctrl+L = active 타일의 경로 입력칸 열기 — 경로-3(화면: 입력칸), REF-path-input.md. 브라우저 주소창보다 먼저 받는다
export const usePathEditKey = (tileId: () => string, open: () => void) => {
    const { activeId } = useTileGrids()
    const onKeyDown = (e: KeyboardEvent) => {
        if (!e.ctrlKey || e.altKey || e.shiftKey || e.metaKey || e.code !== 'KeyL' || activeId.value !== tileId()) return
        e.preventDefault()
        e.stopPropagation()
        open()
    }
    onMounted(() => window.addEventListener('keydown', onKeyDown, true))
    onBeforeUnmount(() => window.removeEventListener('keydown', onKeyDown, true))
}
