import type { SocketContext } from '@/common/websocket/websocket.hook'
import { invalidateMemberDevices } from '@/feature/member/api/device.api'
import { invalidateWorkerRoster } from '@/feature/worker/api/worker.api'
import { useQueryClient } from '@tanstack/vue-query'
import { watch } from 'vue'

const BURST_MS = 300    // 주소 일괄 변경처럼 알림이 몰려 오면 한 번만 다시 조회

// 장비 상태 push(REF-device-push.md): 알림 = "바뀜" 신호만, 장비 목록 조회를 다시 부른다. 앱 루트에서 한 번
export const bindDeviceChanges = (socket: SocketContext) => {
    const queryClient = useQueryClient()
    let timer: ReturnType<typeof setTimeout> | undefined
    const refetch = () => {
        clearTimeout(timer)
        timer = setTimeout(() => {
            void invalidateWorkerRoster(queryClient)
            void invalidateMemberDevices(queryClient)
        }, BURST_MS)
    }
    socket.on('DEVICE:CHANGE', refetch)
    socket.on('SHARE:CHANGE', refetch)    // 공유가 바뀌면 보이는 장비도 바뀜
    // 끊겨 있던 동안 놓친 알림 — 다시 연결되면 한 번(첫 연결은 처음 조회가 함)
    let connected = false
    watch(socket.status, s => {
        if (s !== 'CONNECTED') return
        if (connected) refetch()
        connected = true
    })
}
