import { BaseAxios, throwCatch, throwThen } from '@/common/api/api.util'
import { useQuery } from '@tanstack/vue-query'
import { computed } from 'vue'

// 백엔드: apps/core/cmd/supervisor/router/workerApi.go (group "/workers")
// 접속 중인 worker 인스턴스만(메모리 레지스트리) — 끊긴 인스턴스는 안 나옴

export type WorkerResponse = {
    mainKey: string
    subKey: string
    instanceKey: string     // main#sub = execProcess의 authKey
}

// GET /workers — 전체(폴더 device_key 후보) / ?nodeId= 스크립트의 실행 대상 main_key 인스턴스만
// nodeId의 귀속 폴더(device_key)가 없으면 404
export const listWorkers = (nodeId?: number) => BaseAxios.get(
    '/workers',
    { params: { nodeId } }
).then(throwThen<WorkerResponse[]>)
.catch(throwCatch)

const WORKER_POLL_MS = 10 * 1000

// main_key별 접속 중인 인스턴스 수. 접속·끊김 알림이 없어 주기 조회한다(같은 키라 화면 전체가 요청 하나를 공유)
export const useOnlineWorkers = () => {
    const query = useQuery({
        queryKey: ['WORKER', 'LIST'],
        queryFn: () => listWorkers(),
        refetchInterval: WORKER_POLL_MS,
    })

    const counts = computed(() => {
        const rst = new Map<string, number>()
        for (const { mainKey } of query.data.value ?? []) rst.set(mainKey, (rst.get(mainKey) ?? 0) + 1)
        return rst
    })

    return { counts, refetch: query.refetch }
}
