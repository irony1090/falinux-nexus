import { useTestSocket } from '@/common/websocket/websocket.hook'
import { useQueryClient } from '@tanstack/vue-query'
import { computed, onUnmounted, type Ref } from 'vue'
import { invalidateNodeRuns, useNodeRuns, type RunResponse } from '../api/node.api'

// 노드의 상태 리스트(실시간-6a)·터미널 탭(deep, 개편-6(목록 UI)). 실행·종료·종료 요청·타일 변화·공유 변화가 오면 목록을 다시 받는다
export const useScriptRuns = (folderId: Ref<number | null>, deep = false) => {
    const queryClient = useQueryClient()
    const { on } = useTestSocket()
    const query = useNodeRuns(folderId, deep)

    const invalidate = () => { void invalidateNodeRuns(queryClient) }
    const offs = ['TILES:UPDATE', 'STATUS', 'PROCESS:UPDATE', 'SHARE:CHANGE'].map(type => on(type, invalidate))
    onUnmounted(() => offs.forEach(off => off()))

    const byNode = computed(() => {
        const map = new Map<number, RunResponse[]>()
        for (const r of query.data.value ?? []) map.set(r.nodeId, [...(map.get(r.nodeId) ?? []), r])
        return map
    })

    const list = computed(() => query.data.value ?? [])

    return { byNode, list }
}
