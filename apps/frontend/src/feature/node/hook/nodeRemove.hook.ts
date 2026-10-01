import { useQueryClient } from '@tanstack/vue-query'
import { nextTick } from 'vue'
import { deleteNode, useNodeQueryClient, type NodeResponse } from '../api/node.api'
import { useTileTree } from '../store/tileTree.store'

// 노드 삭제 + 삭제된 폴더(또는 그 하위)를 보던 폴더 타일 닫기 — REF-node-ui-link.md 결정 E(삭제된 노드를 보던 타일)
export const useNodeRemove = () => {
    const { tiles, rootId, close, navigate } = useTileTree()
    const { invalidateAll, fetchPath } = useNodeQueryClient(useQueryClient())

    // nodeId 폴더나 그 하위를 보고 있는 폴더 타일
    const tilesUnder = async (nodeId: number) => {
        const folders = Object.values(tiles.value).flatMap(t =>
            t.type === 'folder' && t.nodeId !== null ? [{ id: t.id, nodeId: t.nodeId }] : []
        )
        const paths = await Promise.all(folders.map(f => fetchPath(f.nodeId).catch(() => [])))
        return folders.filter((_, i) => paths[i]?.some(p => p.id === nodeId)).map(f => f.id)
    }

    // 순서 의존:
    // 1) 대상 타일은 삭제 전에 구한다 — 경로를 getNode로 따라가므로 삭제 후에는 404
    // 2) 무효화는 타일이 언마운트된 뒤(nextTick) — 먼저 하면 닫힐 타일이 삭제된 노드를 다시 조회한다
    const removeNode = async (node: NodeResponse) => {
        const targets = node.kind === 'FOLDER' ? await tilesUnder(node.id) : []
        await deleteNode(node.id)
        // 루트 타일은 닫을 수 없어 기본 목록으로 돌린다
        targets.forEach(id => id === rootId ? navigate(id, null) : close(id))
        await nextTick()
        invalidateAll(node.id, node.parentId ?? undefined)
    }

    return { removeNode }
}
