import { BaseAxios, throwCatch, throwThen } from '@/common/api/api.util'
import { useQuery, type QueryClient } from '@tanstack/vue-query'
import { computed } from 'vue'

// 백엔드: apps/core/cmd/supervisor/router/workerApi.go (group "/workers")
// GET /workers = 접속 중만(메모리 레지스트리, 실행 대상) / /workers/instances = 명부(DB, 끊긴 것 포함) — RS2(API 모양)

export type WorkerResponse = {
    mainKey: string
    subKey: string
    instanceKey: string     // main#sub = execProcess의 authKey
    os: string              // linux | windows ... (빈 값은 서버가 linux로) — WW5(OS 표시)
}

export type WorkerInstanceResponse = WorkerResponse & {
    online: boolean
    firstSeen: number       // Unix 초
    lastSeen: number
}

// main_key 하나의 명부 요약
export type WorkerDevice = {
    mainKey: string
    online: number
    instances: WorkerInstanceResponse[]     // 접속 중 먼저, 그다음 최근 접속 순
}

// GET /workers — 전체(폴더 device_key 후보) / ?nodeId= 스크립트의 실행 대상 main_key 인스턴스만
// nodeId의 귀속 폴더(device_key)가 없으면 404
export const listWorkers = (nodeId?: number) => BaseAxios.get(
    '/workers',
    { params: { nodeId } }
).then(throwThen<WorkerResponse[]>)
.catch(throwCatch)

// GET /workers/instances — 명부 전체(?mainKey= 면 그 장비만)
export const listWorkerInstances = (mainKey?: string) => BaseAxios.get(
    '/workers/instances',
    { params: { mainKey } }
).then(throwThen<WorkerInstanceResponse[]>)
.catch(throwCatch)

// DELETE /workers/instances/:mainKey/:subKey — 끊긴 인스턴스만(접속 중이면 409) — RS4(정리)
export const deleteWorkerInstance = (mainKey: string, subKey: string) => BaseAxios.delete(
    `/workers/instances/${encodeURIComponent(mainKey)}/${encodeURIComponent(subKey)}`
).then(throwThen<{}>)
.catch(throwCatch)

const WORKER_POLL_MS = 10 * 1000
const ROSTER_KEY = ['WORKER', 'ROSTER']

export const invalidateWorkerRoster = (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: ROSTER_KEY })

// main_key별 명부 요약. 접속·끊김 알림이 없어 주기 조회한다(같은 키라 화면 전체가 요청 하나를 공유)
export const useWorkerRoster = () => {
    const query = useQuery({
        queryKey: ROSTER_KEY,
        queryFn: () => listWorkerInstances(),
        refetchInterval: WORKER_POLL_MS,
    })

    const devices = computed(() => {
        const rst = new Map<string, WorkerDevice>()
        for (const inst of query.data.value ?? []) {
            let d = rst.get(inst.mainKey)
            if (!d) rst.set(inst.mainKey, d = { mainKey: inst.mainKey, online: 0, instances: [] })
            d.instances.push(inst)
            if (inst.online) d.online++
        }
        for (const d of rst.values()) d.instances.sort((a, b) => Number(b.online) - Number(a.online) || b.lastSeen - a.lastSeen)
        return rst
    })

    return { devices, refetch: query.refetch }
}

// ---- worker 디렉토리 — 폴더-1(목록 요청), REF-worker-dir-list.md ----
export type DirKind = 'dir' | 'file' | 'link' | 'other'

export type DirEntry = {
    name: string
    path: string            // worker가 이은 전체 경로
    kind: DirKind
    size: number
    modTime: number         // Unix 초
    exec?: boolean
    linkTarget?: string     // 링크에 적힌 원문
    targetPath?: string     // 링크 원본 절대 경로(끊긴 링크면 없음)
    targetDir?: boolean
}

export type DirCrumb = { name: string; path: string }   // path "" = 최상위(Windows 드라이브 목록)

export type DirListResponse = {
    path: string
    parent: string | null   // null = 위가 없음
    crumbs: DirCrumb[]
    cursor: string
    total: number
    offset: number
    entries: DirEntry[]
}

// GET /workers/:mainKey/:subKey/dir — 첫 페이지 = path / 다음 = cursor + offset. 410 = 보관 목록 만료(처음부터 다시)
export const listWorkerDir = (instance: string, folderId: number, q: { path?: string; cursor?: string; offset?: number }) => {
    const i = instance.lastIndexOf('#')
    return BaseAxios.get(
        `/workers/${encodeURIComponent(instance.slice(0, i))}/${encodeURIComponent(instance.slice(i + 1))}/dir`,
        { params: { folderId, ...q } }
    ).then(throwThen<DirListResponse>)
    .catch(throwCatch)
}
