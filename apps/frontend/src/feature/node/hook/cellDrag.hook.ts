import { DragSession, type DragPoint } from '@/common/listener/dragSession'
import { computed, onBeforeUnmount, onMounted, ref, shallowRef, watch, type Ref } from 'vue'
import type { NodeResponse } from '../api/node.api'
import { toClip, type ClipMode, type ClipNode } from '../store/nodeClipboard.store'
import { pasteBlock, type PasteDest } from './nodePaste.hook'

const HOLD_MS = 500         // 터치 길게 누르기 — 타일 끌기·옛 칸 체크와 같은 시간
const LIFT_SLOP = 8         // 터치: 길게 누른 뒤 이만큼 움직이면 끌기, 안 움직이고 떼면 체크(칸끌기-A)
const DWELL_MS = 600        // 폴더 칸 가운데에 이만큼 머물면 넣기(칸Q1(넣기 vs 순서))
const EDGE = 32             // 몸통 위·아래 끝 이만큼 안이면 자동 스크롤(px)
const SPEED = 12
const LAND_ANIM_MS = 240   // 넣은 칸 줄어드는 시간 — FolderTileBody `dg-land`와 같게
const LAND_MS = 3000        // 넣은 칸 사라지는 표시를 목록이 다시 올 때까지 두는 최대 시간

export type CellChip = { x: number; y: number; icon: string; color: string; name: string; more: number; to: string }
export const cellIcon = (n: NodeResponse) => n.kind === 'FOLDER' ? (n.ownerNickname ? 'mdi-folder-account' : 'mdi-folder') : 'mdi-console'
export const cellTone = (n: NodeResponse) => n.kind === 'FOLDER' ? 'warning' : 'primary'

// 끄는 중 — 모든 폴더 타일이 같이 본다(다른 타일 강조). 한 번에 하나
type Into = { tileId: string; key: number | 'tile'; dest: PasteDest; name: string }
type Live = {
    tileId: string              // 출발 타일
    nodes: ReadonlyArray<ClipNode>
    base: ReadonlyArray<number> // 시작 때 출발 목록 순서
    order: ReadonlyArray<number>// 미리보기 순서
    into: Into | null           // 넣기 대상(강조)
    dwell: number | null        // 머무는 중인 폴더 칸(출발 타일)
    copy: boolean               // Ctrl — 넣기에만(칸Q4(Ctrl 복사))
}
const live = shallowRef<Live | null>(null)
const chipAt = shallowRef<Omit<CellChip, 'to'> | null>(null)
// 넣기 대상은 머무르기 타이머로도 바뀌므로 끄는 상태에서 바로 만든다
export const cellChip = computed<CellChip | null>(() => {
    const c = chipAt.value, into = live.value?.into
    return c && { ...c, to: into ? `→ ${into.name}${live.value!.copy ? '에 복사' : ''}` : '' }
})

// 다른 타일에 놓기 판정용 — 폴더 타일 몸통마다 등록
type Body = {
    tileId: string; el: () => HTMLElement | undefined; here: () => PasteDest | null
    cell: (id: number) => NodeResponse | undefined; destOf: (n: NodeResponse) => PasteDest
    land: (key: number | 'tile') => void    // 받은 쪽 젤리
}
const bodies = new Set<Body>()

type Opts = {
    tileId: () => string
    root: Ref<HTMLElement | undefined>          // 스크롤 몸통(.FolderTileBody)
    list: () => ReadonlyArray<NodeResponse>     // 서버 목록 순서
    here: () => PasteDest | null                // 이 타일 폴더(경로를 아직 모르면 null)
    destOf: (n: NodeResponse) => PasteDest      // 폴더 칸 = 넣을 곳
    canOrder: () => boolean                     // 이 목록 순서 바꾸기 권한(편집 이상)
    movable: (n: NodeResponse) => boolean       // 순서에 들어가는 칸(루트 = 내 것만, 칸Q3(루트 목록))
    group: (n: NodeResponse) => ReadonlyArray<NodeResponse>   // 끌 칸들(선택 중이면 고른 전체)
    check: (n: NodeResponse) => void            // 터치 길게 누르고 안 움직임
    order: (ids: ReadonlyArray<number>, after: number | null) => Promise<void>
    into: (mode: ClipMode, nodes: ReadonlyArray<ClipNode>, destId: number | null) => Promise<ReadonlyArray<number>>   // 반환 = 옮겨진 id
    pager: {                                    // 옆 페이지 = 화면 끝 띠(tileGrids edge*, 타일 끌기와 공용)
        el: () => HTMLElement | null | undefined
        start: () => void
        track: (p: DragPoint) => void
        end: () => void
    }
}

// 폴더 칸 끌기 — 칸-2(node 칸 끌기), REF-tile-drag.md 칸Q1~Q5. 마우스 = 바로, 터치 = 길게 누른 뒤 움직이면(안 움직이면 체크)
// 출발 타일: 칸 가장자리 = 순서 미리보기(칸이 비켜 감), 폴더 칸 가운데에 머묾 = 넣기. 다른 폴더 타일 = 그 폴더(칸 위면 그 칸)로 넣기
export const useCellDrag = (o: Opts) => {
    const session = new DragSession({ holdMs: HOLD_MS })
    const pending = shallowRef<ReadonlyArray<number> | null>(null)   // 저장 응답·새 목록 전까지 보일 순서
    const landing = ref(new Set<number>())                             // 넣어서 사라지는 칸
    // 넣는 동안 = 놓을 때 목록 그대로. 응답이 빠르면 줄어드는 칸이 끝나기 전에 빠진다(10-29 확인)
    const held = shallowRef<ReadonlyArray<NodeResponse> | null>(null)
    let settled = false     // 요청·줄어들기 둘 다 끝남
    const landed = ref<number | 'tile' | null>(null)                   // 받은 폴더 칸 젤리
    const self: Body = {
        tileId: o.tileId(), el: () => o.root.value, here: o.here,
        cell: id => o.list().find(n => n.id === id), destOf: o.destOf,
        land: key => {      // 같은 칸에 연달아 넣어도 다시 튀게 한 프레임 비운다
            landed.value = null
            requestAnimationFrame(() => landed.value = key)
        },
    }
    let wait: NodeResponse | null = null
    let lift: DragPoint | null = null       // 터치: 길게 눌러 들린 자리(아직 안 끌림)
    let last: DragPoint = { x: 0, y: 0 }
    let dwellTimer: ReturnType<typeof setTimeout> | undefined
    let raf = 0
    let stripX = 0
    let landTimer: ReturnType<typeof setTimeout> | undefined

    const mine = computed(() => live.value?.tileId === o.tileId() ? live.value : null)
    const draggingIds = computed(() => new Set(mine.value?.nodes.map(n => n.id) ?? []))
    const lifted = ref<number | null>(null)

    // 보이는 순서 = 넣는 중 > 끄는 중 미리보기 > 저장 대기 > 서버 순서
    const shown = computed(() => {
        if (held.value) return held.value
        const list = o.list(), seq = mine.value?.order ?? pending.value
        if (!seq) return list
        const byId = new Map(list.map(n => [n.id, n]))
        const out = seq.map(id => byId.get(id)).filter((n): n is NodeResponse => !!n)
        return out.length === list.length ? out : list     // 그새 목록이 바뀌었으면 서버 것
    })
    // 이 타일에서 강조할 넣기 대상(칸 id 또는 타일 전체)
    const intoKey = computed(() => live.value?.into?.tileId === o.tileId() ? live.value.into.key : null)
    const dwellId = computed(() => mine.value?.dwell ?? null)

    const release = () => {
        clearTimeout(landTimer)
        held.value = null
        landing.value = new Set()
    }
    watch(() => o.list(), list => {
        pending.value = null
        if (held.value && settled && !list.some(n => landing.value.has(n.id))) release()
    })

    const set = (patch: Partial<Live>) => {
        if (live.value) live.value = { ...live.value, ...patch }
    }
    const modeOf = (l: Live): ClipMode => l.copy ? 'copy' : 'cut'
    const ok = (l: Live, dest: PasteDest | null) => !!dest && !pasteBlock(modeOf(l), l.nodes, dest)
    const intoOf = (l: Live, tileId: string, key: number | 'tile', dest: PasteDest | null): Into | null =>
        dest && ok(l, dest) ? { tileId, key, dest, name: dest.node?.name ?? '기본' } : null

    const clearDwell = () => {
        clearTimeout(dwellTimer)
        dwellTimer = undefined
    }

    // moving을 빼고 target 앞(before)/뒤에. target 없음 = 순서에 드는 칸 맨 뒤
    const placeAt = (order: ReadonlyArray<number>, moving: ReadonlyArray<number>, target: number | null, before: boolean) => {
        const rest = order.filter(id => !moving.includes(id))
        let at: number
        if (target === null) {
            const byId = new Map(o.list().map(n => [n.id, n]))
            at = rest.reduce((a, id, i) => byId.get(id) && o.movable(byId.get(id)!) ? i + 1 : a, 0)
        } else {
            at = rest.indexOf(target) + (before ? 0 : 1)
        }
        const pick = order.filter(id => moving.includes(id))
        return [...rest.slice(0, at), ...pick, ...rest.slice(at)]
    }
    const same = (a: ReadonlyArray<number>, b: ReadonlyArray<number>) => a.length === b.length && a.every((v, i) => v === b[i])

    // 배치 위치(offset, 비켜 가는 중 transform 무시)로 칸을 찾는다 — 흔들림 방지 ①(배치 계산 위치로 판정)
    const cellAt = (grid: HTMLElement, p: DragPoint) => {
        const r = grid.getBoundingClientRect()
        const x = p.x - r.left, y = p.y - r.top
        for (const c of grid.children) {
            const el = c as HTMLElement
            if (!el.dataset.nodeId) continue
            if (x >= el.offsetLeft && x < el.offsetLeft + el.offsetWidth && y >= el.offsetTop && y < el.offsetTop + el.offsetHeight)
                return { id: Number(el.dataset.nodeId), fx: (x - el.offsetLeft) / el.offsetWidth }
        }
        const inGrid = x >= 0 && y >= 0 && x < r.width && y < r.height
        return inGrid ? 'empty' as const : null
    }

    const judge = () => {
        const l = live.value
        if (!l) return
        const hit = document.elementFromPoint(last.x, last.y)
        const body = hit ? [...bodies].find(b => b.el()?.contains(hit)) : undefined
        if (!body) {
            clearDwell()
            return set({ order: l.base, into: null, dwell: null })
        }
        // 다른 타일 = 그 폴더(폴더 칸 위면 그 칸), 순서는 원래대로 — SD7(드롭 위치)
        if (body.tileId !== l.tileId) {
            clearDwell()
            const row = hit!.closest<HTMLElement>('[data-folder-id]')
            const n = row ? body.cell(Number(row.dataset.folderId)) : undefined
            const into = (n && !l.nodes.some(d => d.id === n.id) && intoOf(l, body.tileId, n.id, body.destOf(n))) || intoOf(l, body.tileId, 'tile', body.here())
            return set({ order: l.base, into, dwell: null })
        }
        const grid = o.root.value?.querySelector<HTMLElement>('.dir-grid')
        const at = grid ? cellAt(grid, last) : null
        const moving = l.nodes.map(n => n.id)
        const orderable = o.canOrder() && moving.every(id => o.list().some(c => c.id === id && o.movable(c)))
        if (!at || (at === 'empty' && !orderable)) {
            clearDwell()
            return set({ into: null, dwell: null })
        }
        if (at === 'empty') {
            clearDwell()
            return set({ order: placeAt(l.order, moving, null, false), into: null, dwell: null })
        }
        // 빈 자리 위 = 그대로 — 흔들림 방지 ②
        if (moving.includes(at.id)) {
            clearDwell()
            return set({ into: null, dwell: null })
        }
        const n = o.list().find(c => c.id === at.id)
        if (!n) return
        const folder = n.kind === 'FOLDER' && ok(l, o.destOf(n))
        // 넣을 수 있는 폴더 = 가장자리 ¼만 순서, 가운데 = 머물면 넣기 / 그 밖 = 반씩 앞·뒤
        const zone = folder ? (at.fx < 0.25 ? 'before' : at.fx > 0.75 ? 'after' : 'center') : (at.fx < 0.5 ? 'before' : 'after')
        if (zone === 'center') {
            if (l.dwell === n.id) return
            clearDwell()
            set({ dwell: n.id, into: null })
            dwellTimer = setTimeout(() => {
                const cur = live.value
                if (!cur || cur.dwell !== n.id) return
                // 빈 자리는 원래 자리로 — 단 그러면 이 폴더 칸이 커서 밑에서 비켜 나가는 때는 그대로 둔다
                const back = cur.base.indexOf(n.id) === cur.order.indexOf(n.id)
                set({ into: intoOf(cur, l.tileId, n.id, o.destOf(n)), order: back ? cur.base : cur.order })
            }, DWELL_MS)
            return
        }
        clearDwell()
        if (!orderable || !o.movable(n)) return set({ into: null, dwell: null })
        const next = placeAt(l.order, moving, n.id, zone === 'before')
        set({ order: same(next, l.order) ? l.order : next, into: null, dwell: null })
    }

    const autoScroll = () => {
        if (!live.value) return
        const hit = document.elementFromPoint(last.x, last.y)
        const el = (hit && [...bodies].find(b => b.el()?.contains(hit))?.el()) || o.root.value
        if (el) {
            const r = el.getBoundingClientRect()
            const d = last.y < r.top + EDGE ? last.y - (r.top + EDGE) : last.y > r.bottom - EDGE ? last.y - (r.bottom - EDGE) : 0
            if (d && last.x >= r.left && last.x <= r.right) {
                const before = el.scrollTop
                el.scrollTop += Math.sign(d) * Math.min(SPEED, Math.abs(d) / EDGE * SPEED)
                if (el.scrollTop !== before) judge()
            }
        }
        o.pager.track(last)
        const x = o.pager.el()?.scrollLeft ?? 0
        if (x !== stripX) {     // 칼럼 넘기는 동안(부드러운 스크롤) 매 프레임 다시 판정
            stripX = x
            judge()
        }
        raf = requestAnimationFrame(autoScroll)
    }


    const begin = (n: NodeResponse, p: DragPoint) => {
        const nodes = o.group(n)
        const base = o.list().map(c => c.id)
        lift = null
        lifted.value = null
        live.value = { tileId: o.tileId(), nodes: nodes.map(toClip), base, order: base, into: null, dwell: null, copy: false }
        last = p
        chipAt.value = { x: p.x, y: p.y, icon: cellIcon(n), color: cellTone(n), name: n.name, more: nodes.length - 1 }
        document.documentElement.classList.add('tile-dragging')
        stripX = o.pager.el()?.scrollLeft ?? 0
        o.pager.start()
        raf = requestAnimationFrame(autoScroll)
    }

    const end = () => {
        cancelAnimationFrame(raf)
        clearDwell()
        wait = null
        lift = null
        lifted.value = null
        if (mine.value) {
            live.value = null
            chipAt.value = null
            document.documentElement.classList.remove('tile-dragging')
            o.pager.end()
        }
    }

    // Ctrl은 끄는 중 바뀔 수 있다 — 넣기 대상을 다시 판정
    const onKey = (e: KeyboardEvent | PointerEvent) => {
        const l = mine.value, copy = e.ctrlKey || e.metaKey
        if (!l || l.copy === copy) return
        set({ copy })
        judge()
    }

    session.on('start', p => {
        if (!wait) return session.cancel()
        if (lift === null && pointerType !== 'mouse') {     // 터치 = 들기만, 움직여야 끌기
            lift = p
            lifted.value = wait.id
            return
        }
        begin(wait, p)
    })
    session.on('move', p => {
        if (lift && wait) {
            if (Math.hypot(p.x - lift.x, p.y - lift.y) < LIFT_SLOP) return
            begin(wait, p)
        }
        if (!mine.value) return
        last = p
        if (chipAt.value) chipAt.value = { ...chipAt.value, x: p.x, y: p.y }
        judge()
    })
    session.on('drop', () => {
        const n = wait, l = mine.value, snap = shown.value
        if (lift && n) {     // 길게 눌렀다 안 움직이고 뗌 = 체크(칸끌기-A)
            end()
            return o.check(n)
        }
        end()
        if (!l) return
        if (l.into) return void drop(l, l.into, snap)
        if (same(l.order, l.base)) return
        const ids = l.nodes.map(c => c.id)
        const first = l.order.findIndex(id => ids.includes(id))
        pending.value = l.order
        o.order(ids, l.order[first - 1] ?? null).catch(() => pending.value = null)
    })
    session.on('cancel', end)

    // 넣기 — 옮겨 가는 칸은 줄어들며 사라지고, 받은 폴더 칸은 젤리(칸Q5(애니메이션))
    const drop = async (l: Live, into: Into, snap: ReadonlyArray<NodeResponse>) => {
        const mode = modeOf(l)
        ;[...bodies].find(b => b.tileId === into.tileId)?.land(into.key)
        if (mode === 'cut') {
            clearTimeout(landTimer)
            held.value = snap
            landing.value = new Set(l.nodes.map(c => c.id))
            settled = false
            landTimer = setTimeout(release, LAND_MS)
        }
        const [done] = await Promise.all([o.into(mode, l.nodes, into.dest.id), new Promise(r => setTimeout(r, LAND_ANIM_MS))])
        if (mode !== 'cut') return
        // 못 옮긴 칸은 되살리고, 옮긴 칸이 새 목록에서 빠지면 놓는다
        landing.value = new Set([...landing.value].filter(id => done.includes(id)))
        settled = true
        if (!o.list().some(n => landing.value.has(n.id))) release()
    }
    let pointerType = 'mouse'
    // 칸 pointerdown — 칸 안 버튼·라벨은 각자 동작
    const down = (e: PointerEvent, n: NodeResponse) => {
        if ((e.target as Element).closest('button, a, input, .v-btn')) return
        pointerType = e.pointerType
        wait = n
        if (!session.down(e)) wait = null
    }

    onMounted(() => {
        bodies.add(self)
        window.addEventListener('keydown', onKey, true)
        window.addEventListener('keyup', onKey, true)
        window.addEventListener('pointermove', onKey, true)
    })
    let unblock = () => {}
    watch(o.root, el => {
        unblock()
        unblock = el ? session.blockTouch(el) : () => {}
    }, { immediate: true })
    onBeforeUnmount(() => {
        bodies.delete(self)
        window.removeEventListener('keydown', onKey, true)
        window.removeEventListener('keyup', onKey, true)
        window.removeEventListener('pointermove', onKey, true)
        unblock()
        clearTimeout(landTimer)
        session.dispose()
        end()
    })

    return { shown, draggingIds, lifted, dwellId, intoKey, landing, landed, down, active: computed(() => !!mine.value) }
}
