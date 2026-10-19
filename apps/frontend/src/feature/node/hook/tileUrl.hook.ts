import { onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter, type LocationQuery } from 'vue-router'
import type { provideTileGrids } from '../store/tileGrids.store'
import { useTileTree } from '../store/tileTree.store'

// 스크롤이 멈춘 뒤에 쓴다
const WRITE_DELAY = 300

const str = (v: LocationQuery[string]) => typeof v === 'string' && v ? v : undefined

// ?focus=<tileId> = active 타일 / ?view=<tileId> = 보던 칼럼의 타일 / ?tab=terminals = 루트 터미널 탭(한 번 쓰고 지움).
// 새로고침해도 보던 화면으로 돌아오게 URL을 계속 맞춘다 — REF-tile-url.md
export const useTileUrl = (ctx: ReturnType<typeof provideTileGrids>) => {
    const { tiles, setMinimized } = useTileTree()
    const { activeId, anchorId, grids, overview, rootTab, activate, restore } = ctx
    const route = useRoute()
    const router = useRouter()

    // 이 훅이 마지막으로 쓴 값 — query가 이것과 같으면 내 쓰기가 돌아온 것
    let written: { focus?: string; view?: string } = {}
    // 들어온 값을 적용하기 전엔 쓰지 않는다(덮어쓰면 복원할 값이 사라짐)
    let pending = false
    let timer: ReturnType<typeof setTimeout> | undefined

    const write = () => {
        clearTimeout(timer)
        timer = setTimeout(() => {
            if (pending || overview.value) return
            const focus = activeId.value ?? undefined
            const view = anchorId.value
            written = { focus, view }
            if (str(route.query.focus) === focus && str(route.query.view) === view && !route.query.tab) return
            const query: LocationQuery = { ...route.query }
            delete query.focus
            delete query.view
            delete query.tab
            if (focus) query.focus = focus
            if (view) query.view = view
            void router.replace({ query })
        }, WRITE_DELAY)
    }

    watch(() => [str(route.query.focus), str(route.query.view), str(route.query.tab), grids.value.length > 0] as const, async ([focus, view, tab, measured]) => {
        if (focus === written.focus && view === written.view && !tab) return
        pending = true
        if (!measured) return       // Grid 측정 전 — 스크롤할 자리가 없다
        if (tab === 'terminals') rootTab.value = 'terminals'
        const t = focus ? tiles.value[focus] : undefined
        if (t?.type === 'terminal' && t.minimized) setMinimized(t.id, false)
        // view 있음 = 이 훅이 쓴 URL(새로고침) -> 복원 / 없음 = 다른 화면이 보낸 이동(focusPath) -> 예전처럼 activate.
        // 없는 타일(그 사이 닫힘)은 건너뛰고 아래 write가 지금 상태로 덮는다 — U4(없는 타일)
        if (view) await restore(focus, view)
        else if (t) await activate(t.id)
        pending = false
        write()
    }, { immediate: true, flush: 'post' })

    watch([activeId, anchorId], write)
    onBeforeUnmount(() => clearTimeout(timer))
}
