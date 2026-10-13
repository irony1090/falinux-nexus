import { useTestSocket } from '@/common/websocket/websocket.hook'
import { computed, inject, onBeforeUnmount, provide, readonly, ref, watch } from 'vue'
import { deleteBoard, listBoards, renameBoard, type BoardItem } from '../api/board.api'

const BOARDS_STORE_KEY = Symbol('BoardsStore')

// 화면에 보이는 보드 이름 — 이름이 비어 있으면 기본/이름 없음
export const boardLabel = (b?: BoardItem) => !b ? '' : b.name || (b.isDefault ? '기본 보드' : '이름 없는 보드')

export const boardPath = (b: BoardItem) => b.isDefault ? '/' : `/w/${b.hash}`

// 보드로 가서 그 타일을 active로 — 목록 "보드로 이동"·위치 알림(N3(누르면 이동)). TileWorkspace가 처리 후 query를 지운다
export const focusPath = (b: BoardItem, tileId: string) => ({ path: boardPath(b), query: { focus: tileId } })

// 계정의 보드 목록. 만들기·이름·삭제는 WORKSPACES:UPDATE로 모든 탭(내 탭 포함)에 온다 — REF-workspace-impl.md
export const provideBoards = () => {
    const { on, status } = useTestSocket()

    const list_ = ref<BoardItem[]>([])
    const loaded_ = ref(false)
    const removedHere = new Set<string>()   // 이 탭이 지운 보드 — "다른 곳에서 삭제됨" 안내를 띄우지 않으려고
    let gen = 0

    const load = async () => {
        const my = ++gen
        try {
            const list = await listBoards()
            if (my !== gen) return
            list_.value = list
            loaded_.value = true
        } catch (err) {
            console.log('[BOARDS] load', err)
        }
    }

    const reset = () => {
        gen++
        list_.value = []
        loaded_.value = false
    }

    const off = on<{ workspaces: BoardItem[] }>('WORKSPACES:UPDATE', ev => {
        gen++
        list_.value = ev.workspaces
        loaded_.value = true
    })
    // 끊긴 동안 놓친 변경
    watch(status, s => {
        if (s === 'CONNECTED' && loaded_.value) void load()
    })
    onBeforeUnmount(off)

    const byHash = (hash: string) => list_.value.find(b => b.hash === hash)
    const defaultBoard = computed(() => list_.value.find(b => b.isDefault))

    // 응답으로 바로 반영(같은 내용이 이벤트로도 온다)
    const rename = async (hash: string, name: string) => {
        const res = await renameBoard(hash, name)
        const ws = res.workspace
        if (ws) list_.value = list_.value.map(b => b.hash === ws.hash ? ws : b)
    }

    // 막 만든 보드를 이벤트 전에 목록에 넣는다(헤더 이름이 바로 보이게)
    const add = (b: BoardItem) => {
        if (!byHash(b.hash)) list_.value = [...list_.value, b]
    }

    const remove = async (hash: string) => {
        removedHere.add(hash)
        await deleteBoard(hash)
        list_.value = list_.value.filter(b => b.hash !== hash)
    }

    const ctx = {
        list: readonly(list_),
        loaded: readonly(loaded_),
        defaultBoard,
        byHash,
        removedHere: (hash: string) => removedHere.has(hash),
        load,
        reset,
        add,
        rename,
        remove,
    }
    provide(BOARDS_STORE_KEY, ctx)
    return ctx
}

export const useBoards = () => {
    const context = inject<ReturnType<typeof provideBoards>>(BOARDS_STORE_KEY)
    if (!context) throw new Error('BoardsStore is not provided')
    return context
}
