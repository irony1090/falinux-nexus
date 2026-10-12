import { useTestSocket } from '@/common/websocket/websocket.hook'
import { useQueryClient } from '@tanstack/vue-query'
import { onBeforeUnmount, watch } from 'vue'
import { useNodeQueryClient } from '../api/node.api'

type NodeChangeEvent = { id: number, parentId: number | null }

// 볼 수 있는 노드가 다른 탭·계정에서 바뀜 — REF-node-list-sync.md. 서버는 그 노드를 볼 수 있는 접속 계정에만 보낸다
export const useNodeChange = () => {
    const queryClient = useQueryClient()
    const { invalidateChanged } = useNodeQueryClient(queryClient)
    const { on, status } = useTestSocket()

    const offs = (['NODE:CREATE', 'NODE:UPDATE', 'NODE:DELETE'] as const)
        .map(kind => on<NodeChangeEvent>(kind, invalidateChanged))

    // 다시 붙으면 끊긴 동안 놓친 변경을 받는다 — NS6(재접속). 첫 연결은 어차피 처음 조회라 거른다
    let connectedOnce = status.value === 'CONNECTED'
    watch(status, s => {
        if (s !== 'CONNECTED') return
        if (connectedOnce) void queryClient.invalidateQueries({ queryKey: ['NODE'] })
        connectedOnce = true
    })

    onBeforeUnmount(() => offs.forEach(off => off()))
}
