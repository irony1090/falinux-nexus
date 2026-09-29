import type { Area } from '@/feature/widget/util/tile.type'
import { flatten, pack } from '@/feature/widget/util/tileOrder.util'
import { capacity, effectiveSize, minSize, stripGeometry, stripView } from '@/feature/widget/util/tileSplit.util'
import { computed, inject, provide, ref, shallowRef } from 'vue'
import { useTileTree } from './tileTree.store'

const TILE_GRIDS_STORE_KEY = Symbol('TileGridsStore')

// strip padding = Grid 간격 = 타일 간격 (1칼럼 Grid 폭 계산이 이 셋이 같다고 가정)
export const TILE_GAP = 8

// 트리 + 측정한 Grid 영역 -> Grid 목록 + 가로 스크롤 위치 -> 헤더 내비 — REF-node-ui-impl.md "④ Grid 파생"
// 측정·스크롤 기록은 TileStrip, 여기는 등록된 strip 요소에 scrollTo만 한다
export const provideTileGrids = () => {
    const { tiles, rootId } = useTileTree()

    const area = ref<Area | null>(null)

    const cap = computed(() => capacity(area.value ?? { w: 0, h: 0 }))
    const order = computed(() => flatten(tiles.value, rootId))
    // 측정 전엔 그리지 않는다 — 기본 cap을 가정해 그리면 측정 직후 배치가 튄다
    const grids = computed(() => area.value
        ? pack(order.value.map(o => o.tile), t => effectiveSize(t.size, cap.value))
        : [])
    const newSize = computed(() => minSize(cap.value))

    // 타일 순번(1부터)
    const posOf = computed(() => new Map(order.value.map((o, i) => [o.tile.id, i + 1])))

    // 같은 노드를 여러 번 실행했으면 순서대로 몇 번째인지, 한 번뿐이면 없음
    const instanceOf = computed(() => {
        const byNode = new Map<number, Array<string>>()
        order.value.forEach(({ tile }) => {
            if (tile.type !== 'terminal') return
            byNode.set(tile.nodeId, [...(byNode.get(tile.nodeId) ?? []), tile.id])
        })
        const rst = new Map<string, number>()
        byNode.forEach(ids => {
            if (ids.length > 1) ids.forEach((id, i) => rst.set(id, i + 1))
        })
        return rst
    })

    // 가로 스크롤 — TileStrip이 기록, TileNav가 읽음
    const stripEl = shallowRef<HTMLElement>()
    const scrollX = ref(0)
    const viewW = ref(0)

    const geometry = computed(() => stripGeometry(grids.value.map(g => g.cols), area.value?.w ?? 0, cap.value, TILE_GAP))
    const view = computed(() => stripView(geometry.value, scrollX.value, viewW.value, TILE_GAP))

    // x = 칼럼/Grid 시작. 스냅 위치와 맞추려고 scroll-padding(gap)만큼 뺀다
    const scrollTo = (x: number) => stripEl.value?.scrollTo({ left: x - TILE_GAP, behavior: 'smooth' })

    const goGrid = (index: number) => {
        const s = geometry.value.grids[index]
        if (s) scrollTo(s.x)
    }
    const goColumn = (delta: -1 | 1) => {
        const c = geometry.value.columns[view.value.col + delta]
        if (c) scrollTo(c.x)
    }

    const ctx = {
        area,
        cap,
        order,
        grids,
        newSize,
        posOf,
        instanceOf,
        stripEl,
        scrollX,
        viewW,
        geometry,
        view,
        goGrid,
        goColumn,
    }

    provide(TILE_GRIDS_STORE_KEY, ctx)

    return ctx
}

export const useTileGrids = () => {
    const context = inject<ReturnType<typeof provideTileGrids>>(TILE_GRIDS_STORE_KEY)
    if (!context) throw new Error('TileGridsStore is not provided')
    return context
}
