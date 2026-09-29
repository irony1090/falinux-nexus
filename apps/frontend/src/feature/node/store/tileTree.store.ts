import type { TileNode, TileSize } from '@/feature/widget/util/tile.type'
import { inject, provide, readonly, ref } from 'vue'

const TILE_TREE_STORE_KEY = Symbol('TileTreeStore')

type TileBase = TileNode & { parent: string | null }    // parent = 이 타일을 연 폴더 타일 (노드 트리의 부모 아님)
export type FolderTile = TileBase & { type: 'folder'; nodeId: number | null }  // null = 루트 목록
export type TerminalTile = TileBase & { type: 'terminal'; nodeId: number; uid: string }
export type Tile = FolderTile | TerminalTile

// effectiveSize가 1칸 축의 0.5를 1.0으로 보므로 어느 화면에서든 최소 단위로 보인다
const ROOT_SIZE: TileSize = { w: 0.5, h: 0.5 }

// 여는 관계 트리(계정 단위 저장본)만 관리한다. 화면 크기(size 계산)와 API 호출은 호출부 책임.
// 참조(nodeId·uid)만 담은 순수 JSON이라 그대로 서버 저장본(⑪)이 된다 — REF-node-ui-impl.md
export const provideTileTree = () => {
    const root: FolderTile = { id: crypto.randomUUID(), type: 'folder', parent: null, kids: [], size: ROOT_SIZE, nodeId: null }
    const tiles_ = ref<Record<string, Tile>>({ [root.id]: root })
    const tiles = readonly(tiles_)
    const rootId = root.id

    const folderOf = (id: string) => {
        const tile = tiles_.value[id]
        if (tile?.type !== 'folder') throw new Error(`[TILE_TREE] not a folder tile: ${id}`)
        return tile
    }

    // 부모 kids 끝에 붙인다 (= 타일 순서 규칙의 "연 순서")
    const attach = (parentId: string, tile: Tile) => {
        const parent = folderOf(parentId)
        tiles_.value[tile.id] = tile
        parent.kids = [...parent.kids, tile.id]
        return tile.id
    }

    const openFolder = (parentId: string, nodeId: number | null, size: TileSize) =>
        attach(parentId, { id: crypto.randomUUID(), type: 'folder', parent: parentId, kids: [], size, nodeId })

    // 폴더 이름 클릭 = 타일 안에서 이동 (새 타일 아님)
    const navigate = (tileId: string, nodeId: number | null) => {
        folderOf(tileId).nodeId = nodeId
    }

    const addTerminal = (parentId: string, nodeId: number, uid: string, size: TileSize) =>
        attach(parentId, { id: crypto.randomUUID(), type: 'terminal', parent: parentId, kids: [], size, nodeId, uid })

    // 폴더를 닫으면 그 폴더가 연 타일들이 부모 소속이 된다 (⑩ 타일 닫기 규칙):
    // 하위 폴더는 닫힌 자리로, 터미널은 adopted로 표시해 kids 끝에 붙인다(= 부모가 직접 실행한 프로세스 뒤).
    // 실행 중 터미널의 닫기 금지(kill 먼저)는 UI가 막는다 — 스토어는 process 상태를 모른다
    const close = (tileId: string) => {
        const tile = tiles_.value[tileId]
        if (!tile?.parent) return   // 루트는 닫을 수 없음
        const parent = folderOf(tile.parent)
        const kids = tile.kids.map(k => tiles_.value[k]).filter(k => !!k)
        const folders = kids.filter(k => k.type === 'folder')
        const terminals = kids.filter(k => k.type === 'terminal')
        // 닫힌 폴더 안에서 보이던 순서(직접 실행 -> 넘겨받음) 그대로 넘긴다
        const lifted = [...terminals.filter(k => !k.adopted), ...terminals.filter(k => k.adopted)]

        kids.forEach(k => { k.parent = parent.id })
        lifted.forEach(k => { k.adopted = true })
        parent.kids = [
            ...parent.kids.flatMap(k => k === tileId ? folders.map(f => f.id) : [k]),
            ...lifted.map(k => k.id),
        ]
        delete tiles_.value[tileId]
    }

    const resize = (tileId: string, size: TileSize) => {
        const tile = tiles_.value[tileId]
        if (tile) tile.size = { ...size }
    }

    const ctx = {
        tiles,
        rootId,
        openFolder,
        navigate,
        addTerminal,
        close,
        resize,
    }

    provide(TILE_TREE_STORE_KEY, ctx)

    return ctx
}

export const useTileTree = () => {
    const context = inject<ReturnType<typeof provideTileTree>>(TILE_TREE_STORE_KEY)
    if (!context) throw new Error('TileTreeStore is not provided')
    return context
}
