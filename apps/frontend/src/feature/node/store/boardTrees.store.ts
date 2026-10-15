import { useTestSocket } from '@/common/websocket/websocket.hook'
import { computed, inject, onBeforeUnmount, provide, readonly, ref, watch, type Ref } from 'vue'
import type { BoardItem } from '../api/board.api'
import { getTiles, type TileTree, type TileTreeResponse } from '../api/tile.api'
import type { provideBoards } from './boards.store'

const BOARD_TREES_STORE_KEY = Symbol('BoardTreesStore')

type Entry = { tree: TileTree; version: number }

export type OtherBoard = { board: BoardItem; tree?: TileTree; error?: string }

// 전체보기에 그릴 다른 보드들의 트리(읽기 전용) — 미리보기-1(다른 보드 트리), REF-board-preview.md.
// 열린 동안만 불러와 두고 다른 보드의 TILES:UPDATE로 맞춘다. 저장은 하지 않는다(보드의 저장은 그 보드를 보는 탭의 tileTree가 함)
// 두 스토어를 provide하는 곳(pages/index.vue)에서 부르므로 boards는 인자로 받는다
export const provideBoardTrees = (current: Readonly<Ref<string>>, boards: ReturnType<typeof provideBoards>) => {
    const { on, status } = useTestSocket()

    const open_ = ref(false)
    const trees_ = ref<Record<string, Entry>>({})
    const errors_ = ref<Record<string, string>>({})
    let gen = 0                                 // 열기·닫기마다 증가 — 그 전 요청의 응답은 버린다
    const early = new Map<string, Entry>()      // 불러오는 중 온 push — 받은 트리보다 새것이면 덮어씀

    const others = computed(() => boards.list.value.filter(b => b.hash !== current.value))

    const keep = (hash: string, e: Entry) => {
        const prev = trees_.value[hash]
        if (!prev || e.version > prev.version) trees_.value[hash] = e
    }

    const fetchOne = async (hash: string) => {
        const my = gen
        try {
            const res = await getTiles(hash)
            if (my !== gen || !open_.value) return
            const stashed = early.get(hash)
            early.delete(hash)
            delete errors_.value[hash]
            keep(hash, stashed && stashed.version > res.version ? stashed : { tree: res.tree, version: res.version })
        } catch (err) {
            if (my !== gen) return
            errors_.value[hash] = (err as { message?: string })?.message || '불러오지 못했습니다'
        }
    }

    // 목록에 새로 생긴 보드만 불러오고 사라진 보드는 버린다
    const syncList = () => {
        if (!open_.value) return
        const want = new Set(others.value.map(b => b.hash))
        Object.keys(trees_.value).forEach(h => { if (!want.has(h)) delete trees_.value[h] })
        Object.keys(errors_.value).forEach(h => { if (!want.has(h)) delete errors_.value[h] })
        want.forEach(h => { if (!trees_.value[h]) void fetchOne(h) })
    }

    const open = () => {
        if (open_.value) return
        gen++
        open_.value = true
        syncList()
    }

    const close = () => {
        if (!open_.value) return
        gen++
        open_.value = false
        trees_.value = {}
        errors_.value = {}
        early.clear()
    }

    watch(() => others.value.map(b => b.hash).join(), syncList)

    // 다른 탭·서버의 변경. 지금 보드 것은 tileTree가 받는다
    const off = on<TileTreeResponse>('TILES:UPDATE', ev => {
        if (!open_.value || ev.hash === current.value) return
        const e = { tree: ev.tree, version: ev.version }
        if (trees_.value[ev.hash]) keep(ev.hash, e)
        else if (!early.has(ev.hash) || ev.version > early.get(ev.hash)!.version) early.set(ev.hash, e)
    })
    // 끊긴 동안 놓친 push는 다시 받아서 맞춘다
    watch(status, s => {
        if (s === 'CONNECTED' && open_.value) others.value.forEach(b => void fetchOne(b.hash))
    })
    onBeforeUnmount(off)

    // 보드 목록 순서대로. 아직 못 받은 보드는 tree 없음
    const list = computed<OtherBoard[]>(() => open_.value
        ? others.value.map(board => ({ board, tree: trees_.value[board.hash]?.tree, error: errors_.value[board.hash] }))
        : [])

    // 다른 보드 트리의 터미널 uid — 터미널 등록 기준에 더한다(미리보기-2(터미널 등록))
    const uids = computed(() => new Set(Object.values(trees_.value)
        .flatMap(e => Object.values(e.tree.tiles).flatMap(t => t.type === 'terminal' ? [t.uid] : []))))

    const ctx = {
        isOpen: readonly(open_),
        list,
        uids,
        open,
        close,
    }
    provide(BOARD_TREES_STORE_KEY, ctx)
    return ctx
}

export const useBoardTrees = () => {
    const context = inject<ReturnType<typeof provideBoardTrees>>(BOARD_TREES_STORE_KEY)
    if (!context) throw new Error('BoardTreesStore is not provided')
    return context
}
