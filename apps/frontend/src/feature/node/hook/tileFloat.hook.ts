import { useVirtualKeyboard } from '@/common/hook/virtualKeyboard.hook'
import { KEY_BAR_HEIGHT } from '@/feature/process/component/TerminalKeyBar.vue'
import { computed, onBeforeUnmount, ref, watch, type CSSProperties, type Ref } from 'vue'

// 맞춤-1(타일 띄우기): 키보드가 열린 동안 포커스된 터미널 타일만 가로 자리·폭은 그대로, 높이 = 보이는 영역 위 끝~키 줄 위 끝 — REF-board-keyboard.md "맞춤 계획"
// 빠진 grid 칸(부모 .Tile)은 빈 채 남아 다른 타일이 안 움직인다. 보드가 가로로 움직이면 그 칸 위치를 다시 읽어 따라간다
// setup 안에서 동기로 불러야 한다(onBeforeUnmount 등록)
export const useTileFloat = (frameEl: Ref<HTMLElement | undefined>, termUid: Ref<string | undefined>, overview: Ref<boolean>) => {
    const kb = useVirtualKeyboard()
    const floating = computed(() => kb.open && !overview.value && !!termUid.value && kb.termUid === termUid.value)
    const slot = ref<{ left: number; width: number }>()
    const readSlot = () => {
        const r = frameEl.value?.parentElement?.getBoundingClientRect()
        slot.value = r && { left: r.left, width: r.width }
    }
    let slotRaf = 0
    const onScroll = (e: Event) => {
        if (!(e.target instanceof Node) || !frameEl.value || !e.target.contains(frameEl.value)) return   // 터미널 안 스크롤 등은 무시
        cancelAnimationFrame(slotRaf)
        slotRaf = requestAnimationFrame(readSlot)
    }
    const unlisten = () => {
        window.removeEventListener('scroll', onScroll, { capture: true })
        cancelAnimationFrame(slotRaf)
    }
    watch(floating, on => {
        unlisten()
        if (!on) return
        readSlot()
        window.addEventListener('scroll', onScroll, { capture: true, passive: true })
    })
    onBeforeUnmount(unlisten)

    return computed<CSSProperties | undefined>(() => floating.value && slot.value ? {
        position: 'fixed',
        left: `${slot.value.left}px`,
        width: `${slot.value.width}px`,
        top: `${kb.top}px`,
        height: `${kb.bottom - KEY_BAR_HEIGHT - kb.top}px`,
        zIndex: 1000,   // 다른 타일 위, 키 줄(2000) 아래
    } : undefined)
}
