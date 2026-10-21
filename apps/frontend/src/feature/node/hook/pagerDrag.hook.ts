import { onBeforeUnmount, ref } from 'vue'
import { TILE_GAP, useTileGrids } from '../store/tileGrids.store'

// 이만큼 움직여야 드래그 — 그 전엔 포인터를 잡지 않아 도트 클릭이 그대로 동작
const START = 6
// 당긴 거리가 이만큼 = 타일 이동, 다시 이 아래로 = 칼럼 이동. 푸터는 위로만(푸터-4(위로 열기)), 커서 팝업은 위·아래 — 이동-2(커서 위치 페이지 표시)
const TILE_PULL = 36
const TILE_BACK = 16
// 칼럼 하나 / 타일 하나에 드는 가로 px
const COL_PX = 36
const TILE_PX = 29
// 스냅을 다시 켜기 전 smooth 스크롤을 기다리는 최대 시간
const SETTLE_MAX = 700

export type PagerMode = 'idle' | 'col' | 'tile'
export type PagerPull = 'up' | 'both'

type Drag = { id: number; el: HTMLElement; x0: number; y0: number; held: boolean; col0: number; col: number; tile0: number }

// 인디케이터 드래그 — x축 = 칼럼 이동, 당긴 채 x축 = active 타일 이동. REF-board-pager.md
export const usePagerDrag = (pull: PagerPull = 'up') => {
    const {
        geometry, view, order, cap, activeId, maxCol, scrubbing, stripEl, overview,
        goCol, scrubTo, colRange, stepTo, focusTile,
    } = useTileGrids()
    const mode = ref<PagerMode>('idle')
    // 타일 이동에 들어간 쪽 — 타일 패널을 그쪽에 띄운다
    const pullDir = ref<'up' | 'down'>('up')
    let d: Drag | null = null
    let stopSettle = () => {}

    const clampCol = (c: number) => Math.max(0, Math.min(maxCol.value, c))
    const ids = () => order.value.map(o => o.tile.id)

    // 화면(왼쪽 칼럼 v부터 cap.x칸)에 다 들어가는가
    const inView = (tileId: string, v: number) => {
        const r = colRange(tileId)
        return !!r && r.l >= v && r.r <= v + cap.value.x - 1
    }

    // 칼럼에 맞춘 뒤 스냅을 다시 켠다. smooth 도중에 켜면 그 자리에서 다른 칼럼으로 튈 수 있어 도착을 기다린다
    // (scrollend는 직전 scrubTo 것이 늦게 올 수 있으므로 위치까지 확인, 스크롤이 없었으면 시간 초과로)
    const settle = (col: number) => {
        stopSettle()
        goCol(col)
        const el = stripEl.value
        const c = geometry.value.columns[col]
        const finish = () => {
            stopSettle()
            scrubbing.value = false
        }
        const check = () => {
            if (el && c && Math.abs(el.scrollLeft - (c.x - TILE_GAP)) <= 1) finish()
        }
        const timer = setTimeout(finish, SETTLE_MAX)
        el?.addEventListener('scrollend', check)
        stopSettle = () => {
            clearTimeout(timer)
            el?.removeEventListener('scrollend', check)
            stopSettle = () => {}
        }
    }

    const down = (e: PointerEvent) => {
        if (overview.value || (e.pointerType === 'mouse' && e.button !== 0)) return
        const col = view.value.col
        d = { id: e.pointerId, el: e.currentTarget as HTMLElement, x0: e.clientX, y0: e.clientY, held: false, col0: col, col, tile0: 0 }
    }

    const move = (e: PointerEvent) => {
        if (!d || e.pointerId !== d.id) return
        const dx = e.clientX - d.x0
        const dy = e.clientY - d.y0
        if (!d.held) {
            if (Math.abs(dx) < START && Math.abs(dy) < START) return
            d.held = true
            d.el.setPointerCapture(d.id)
            stopSettle()
            scrubbing.value = true
            mode.value = 'col'
        }

        // 당긴 거리. 타일 이동 중엔 들어간 쪽 기준
        const pulled = mode.value === 'tile' ? (pullDir.value === 'up' ? -dy : dy) : (pull === 'both' ? Math.abs(dy) : -dy)
        if (mode.value === 'col' && pulled > TILE_PULL) {
            pullDir.value = dy < 0 ? 'up' : 'down'
            // 칼럼 사이면 가까운 칼럼에 맞추고, active가 화면에 없으면 화면의 첫 타일부터
            const v = clampCol(Math.round(d.col))
            goCol(v)
            const list = ids()
            const cur = activeId.value && inView(activeId.value, v) ? activeId.value : list.find(id => inView(id, v))
            if (cur) stepTo(cur, v)
            Object.assign(d, { col: v, x0: e.clientX, tile0: cur ? list.indexOf(cur) : 0 })
            mode.value = 'tile'
            return
        }
        if (mode.value === 'tile' && pulled < TILE_BACK) {
            Object.assign(d, { col0: d.col, x0: e.clientX })
            mode.value = 'col'
            return
        }

        if (mode.value === 'col') {
            d.col = clampCol(d.col0 + dx / COL_PX)
            scrubTo(d.col)
            return
        }
        const list = ids()
        const next = list[Math.max(0, Math.min(list.length - 1, d.tile0 + Math.round(dx / TILE_PX)))]
        if (next && next !== activeId.value) d.col = stepTo(next, d.col)
    }

    const up = (e: PointerEvent) => {
        if (!d || e.pointerId !== d.id) return
        const { held, el, id, col } = d
        const wasTile = mode.value === 'tile'
        d = null
        if (!held) return
        if (el.hasPointerCapture(id)) el.releasePointerCapture(id)
        mode.value = 'idle'
        settle(clampCol(Math.round(col)))
        // 터치에선 포커스를 주지 않는다 — 가상 키보드가 올라온다(U3(복원 때 포커스)와 같은 규칙)
        if (wasTile && activeId.value && matchMedia('(pointer: fine)').matches) focusTile(activeId.value)
    }

    onBeforeUnmount(() => stopSettle())

    return { mode, pullDir, down, move, up }
}
