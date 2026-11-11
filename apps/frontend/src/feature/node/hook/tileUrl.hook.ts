import { onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter, type LocationQuery } from 'vue-router'
import type { provideTileGrids } from '../store/tileGrids.store'
import { isProcTile, useTileTree } from '../store/tileTree.store'

// 스크롤이 멈춘 뒤에 쓴다
const WRITE_DELAY = 300

const str = (v: LocationQuery[string]) => typeof v === 'string' && v ? v : undefined

// ?focus=<tileId> = active 타일(새로고침·접속 때 이 타일로 이동) / ?tab=terminals = 루트 터미널 탭(한 번 쓰고 지움).
// 2026-11-11 ?view(보던 칼럼) 폐지 — 새로고침도 focus로 이동. REF-tile-url.md
export const useTileUrl = (ctx: ReturnType<typeof provideTileGrids>) => {
    const { tiles, setMinimized } = useTileTree()
    const { activeId, plan, overview, rootTab, restore } = ctx
    const route = useRoute()
    const router = useRouter()

    // 이 훅이 마지막으로 쓴 값 — query가 이것과 같으면 내 쓰기가 돌아온 것
    let written: { focus?: string } = {}
    // 들어온 값을 적용하기 전엔 쓰지 않는다(덮어쓰면 복원할 값이 사라짐)
    let pending = false
    // 첫 적용(새로고침·접속) = 애니메이션 없이 바로, 그 뒤(다른 화면이 보낸 이동) = 부드럽게
    let first = true
    let timer: ReturnType<typeof setTimeout> | undefined

    const write = () => {
        clearTimeout(timer)
        timer = setTimeout(() => {
            if (pending || overview.value) return
            const focus = activeId.value ?? undefined
            written = { focus }
            if (str(route.query.focus) === focus && !route.query.view && !route.query.tab) return
            const query: LocationQuery = { ...route.query }
            delete query.focus
            delete query.view
            delete query.tab
            if (focus) query.focus = focus
            void router.replace({ query })
        }, WRITE_DELAY)
    }

    watch(() => [str(route.query.focus), str(route.query.tab), (plan.value?.items.length ?? 0) > 0] as const, async ([focus, tab, measured]) => {
        if (focus === written.focus && !tab) return
        pending = true
        if (!measured) return       // 판 측정 전 — 스크롤할 자리가 없다
        if (tab === 'terminals') rootTab.value = 'terminals'
        const t = focus ? tiles.value[focus] : undefined
        if (isProcTile(t) && t.minimized) setMinimized(t.id, false)
        // 없는 타일(그 사이 닫힘)은 건너뛰고 아래 write가 지금 상태로 덮는다 — U4(없는 타일)
        if (t) await restore(t.id, first ? 'instant' : 'smooth')
        first = false
        pending = false
        write()
    }, { immediate: true, flush: 'post' })

    watch(activeId, write)
    onBeforeUnmount(() => clearTimeout(timer))
}
