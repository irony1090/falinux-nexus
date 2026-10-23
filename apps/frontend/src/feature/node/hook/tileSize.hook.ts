import type { TileSpan } from '@/feature/widget/util/tile.type'
import { effectiveSize } from '@/feature/widget/util/tileSplit.util'
import { computed, type Ref } from 'vue'
import { useTileGrids } from '../store/tileGrids.store'
import { useTileTree } from '../store/tileTree.store'

// 타일 크기 표시·바꾸기 — 머리 줄 크기 버튼과 크기 카드(TileSizeCard)가 같이 씀, 팝업-1(크기 카드 공용화)
export const useTileSize = (tileId: Ref<string | null | undefined>) => {
    const { tiles, resize } = useTileTree()
    const { cap, reveal } = useTileGrids()

    const tile = computed(() => tileId.value ? tiles.value[tileId.value] : undefined)
    const fmt = (v: TileSpan) => v === 1 ? '1' : '.5'
    const sizeText = computed(() => tile.value ? `${fmt(tile.value.size.w)}×${fmt(tile.value.size.h)}` : '')
    // 이 화면에서 실제로 보이는 크기가 다를 때만
    const shownText = computed(() => {
        if (!tile.value) return ''
        const s = effectiveSize(tile.value.size, cap.value)
        const text = `${fmt(s.w)}×${fmt(s.h)}`
        return text === sizeText.value ? '' : text
    })
    const sizeTitle = computed(() => `크기 ${sizeText.value}` + (shownText.value ? ` (이 화면에서는 ${shownText.value})` : ''))

    const setSize = (axis: 'w' | 'h', v: TileSpan) => {
        if (!tile.value) return
        resize(tile.value.id, { ...tile.value.size, [axis]: v })
        reveal(tile.value.id)
    }

    return { tile, cap, shownText, sizeTitle, setSize }
}
