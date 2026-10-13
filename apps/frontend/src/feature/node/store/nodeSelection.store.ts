import { authState } from '@/feature/user/store/auth.store'
import { computed, readonly, ref, shallowRef, watch } from 'vue'
import type { NodeResponse } from '../api/node.api'
import type { Tile } from './tileTree.store'

// 노드 선택 모드 — 여러 타일에 걸쳐 고른다(SM2(선택 범위)). 이 탭에서만 쓰고 저장하지 않는다. 결정 → REF-node-select.md

type SelItem = { node: NodeResponse; tileId: string; folderId: number | null }   // folderId = 고를 때 그 타일이 보던 폴더(null = 루트)

const on = ref(false)
const items = shallowRef(new Map<number, SelItem>())   // 바꿀 때마다 새 Map — shallowRef 반응성

const exit = () => {
    on.value = false
    items.value = new Map()
}

// SM1(켜는 방법): 누른 노드가 선택된 채로 시작, 이미 모드면 토글
const toggle = (node: NodeResponse, tileId: string, folderId: number | null) => {
    const next = new Map(items.value)
    if (on.value && next.has(node.id)) next.delete(node.id)
    else next.set(node.id, { node, tileId, folderId })
    on.value = true
    items.value = next
}

const keep = (pred: (it: SelItem) => boolean) => {
    const next = new Map([...items.value].filter(([, it]) => pred(it)))
    if (next.size !== items.value.size) items.value = next
}

// SM7(사라진 선택): 그 타일 목록에서 빠진 노드는 빼고, 남은 건 최신 값으로
const syncTile = (tileId: string, folderId: number | null, children: NodeResponse[]) => {
    if (![...items.value.values()].some(it => it.tileId === tileId)) return
    const byId = new Map(children.map(c => [c.id, c]))
    const next = new Map<number, SelItem>()
    items.value.forEach((it, id) => {
        if (it.tileId !== tileId) return next.set(id, it)
        const fresh = byId.get(id)
        if (it.folderId === folderId && fresh) next.set(id, { ...it, node: fresh })
    })
    items.value = next
}

// SM7(사라진 선택): 타일이 닫혔거나 다른 폴더로 갔으면 그 타일에서 고른 것을 뺀다
const pruneTiles = (tiles: Record<string, Tile>) => keep(it => {
    const t = tiles[it.tileId]
    return t?.type === 'folder' && t.nodeId === it.folderId
})

watch(authState, exit)

const selected = computed(() => [...items.value.values()])
const ids = computed(() => new Set(items.value.keys()))

export const useNodeSelection = () => ({
    on: readonly(on),
    selected,
    ids,
    toggle,
    exit,
    syncTile,
    pruneTiles,
})
