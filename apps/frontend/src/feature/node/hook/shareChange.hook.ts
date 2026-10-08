import { useTestSocket } from '@/common/websocket/websocket.hook'
import { useQueryClient } from '@tanstack/vue-query'
import { onBeforeUnmount } from 'vue'

// 공유가 걸리거나 풀리거나 노드가 이동·삭제돼 내 단계가 바뀜 — 실시간-6b(공유 반영). 내용 없는 신호라 노드 조회를 전부 다시 받는다.
// 터미널 쪽 재조회는 terminalRestore.hook.ts가 같은 신호로 한다
export const useShareChange = () => {
    const queryClient = useQueryClient()
    const { on } = useTestSocket()
    const off = on('SHARE:CHANGE', () => { void queryClient.invalidateQueries({ queryKey: ['NODE'] }) })
    onBeforeUnmount(off)
}
