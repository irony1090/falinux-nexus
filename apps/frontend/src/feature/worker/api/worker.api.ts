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
    label?: string          // 인스턴스 라벨(회원-Q11), 명부 응답에만
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

const WORKER_POLL_MS = 60 * 1000   // 안전망 — 평소엔 DEVICE:CHANGE 알림으로 다시 조회(deviceChange.hook.ts)
const ROSTER_KEY = ['WORKER', 'ROSTER']

export const invalidateWorkerRoster = (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: ROSTER_KEY })

// 인스턴스(main#sub)가 명부에서 접속 중인지
export const instanceOnline = (devices: ReadonlyMap<string, WorkerDevice>, instance: string) =>
    !!devices.get(instance.slice(0, instance.lastIndexOf('#')))?.instances.find(i => i.instanceKey === instance)?.online

// main_key별 명부 요약. 접속·끊김은 알림으로 다시 조회 + 긴 주기 조회(같은 키라 화면 전체가 요청 하나를 공유)
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

// 경로 자동완성 — 경로-2(후보 REST), REF-path-input.md. 폴더만. max 0(생략) = dir가 폴더인지만(정리된 경로를 돌려줌)
export type DirSuggestResponse = { dir: string; items: DirCrumb[]; more?: boolean }

export const suggestWorkerDir = (instance: string, folderId: number, q: { dir: string; prefix?: string; max?: number }) => {
    const i = instance.lastIndexOf('#')
    return BaseAxios.get(
        `/workers/${encodeURIComponent(instance.slice(0, i))}/${encodeURIComponent(instance.slice(i + 1))}/dir/suggest`,
        { params: { folderId, ...q } }
    ).then(throwThen<DirSuggestResponse>)
    .catch(throwCatch)
}

// ---- supervisor 주소 옮기기·미등록 process — 주소-2(옮기기 명령)·주소-4(미등록 process), REF-supervisor-move.md ----
const instanceUrl = (instance: string, rest: string) => {
    const i = instance.lastIndexOf('#')
    return `/workers/${encodeURIComponent(instance.slice(0, i))}/${encodeURIComponent(instance.slice(i + 1))}/${rest}`
}

// ---- 보기 타일 — 파트(보기 타일), REF-file-view.md ----
// GET .../file — <img src>에 그대로(쿠키 세션). ETag가 같으면 304
export const viewFileUrl = (instance: string, folderId: number, path: string) =>
    `${BaseAxios.defaults.baseURL}${instanceUrl(instance, 'file')}?folderId=${folderId}&path=${encodeURIComponent(path)}`

// GET .../file/probe — 앞 8KB에 NUL이 있으면 바이너리(D(더블클릭 열기)). 화면 표시 실패의 원인 확인에도 씀
export const probeWorkerFile = (instance: string, folderId: number, path: string) => BaseAxios.get(
    instanceUrl(instance, 'file/probe'),
    { params: { folderId, path } }
).then(throwThen<{ binary: boolean }>)
.catch(throwCatch)

// worker에 저장된 supervisor 주소 목록(1순위부터, 최대 3)
export type AddrsResponse = { addrs: string[] }

export const listWorkerAddrs = (instance: string, folderId: number) => BaseAxios.get(
    instanceUrl(instance, 'addrs'),
    { params: { folderId } }
).then(throwThen<AddrsResponse>)
.catch(throwCatch)

// now = 바로 옮기기 / false = 저장만(지금 supervisor가 끊기면 옮김) — 주소-Q9(즉시 옮김 여부)
export const moveWorker = (instance: string, folderId: number, url: string, now: boolean) => BaseAxios.post(
    instanceUrl(instance, 'move'),
    { folderId, url, now }
).then(throwThen<AddrsResponse>)
.catch(throwCatch)

export type MoveAllResult = { instanceKey: string; addrs: string[]; error?: string }

// 접속 중인 worker 전부 — worker_move 권한만(없으면 403)
export const moveAllWorkers = (url: string, now: boolean) => BaseAxios.post(
    '/workers/move-all',
    { url, now }
).then(throwThen<MoveAllResult[]>)
.catch(throwCatch)

// 미등록 process = 이 supervisor DB에 기록이 없는데 worker에서 살아 있는 것(다른 supervisor에서 띄운 것 등)
export type UnregisteredProcess = {
    uid: string
    kind: 'EXEC' | 'EDIT'
    label: string           // 스크립트 노드 이름(장비 파일 실행·이 기능 전에 띄운 것은 빈 값)
    cmd: string             // 명령 + 인자(이 기능 전에 띄운 것은 빈 값). 스크립트 노드면 worker 임시 파일 경로
    cwd: string
    pid: number
    startedAt: number       // Unix 초, 0 = 모름
}

export const listUnregistered = (instance: string, folderId: number) => BaseAxios.get(
    instanceUrl(instance, 'unregistered'),
    { params: { folderId } }
).then(throwThen<UnregisteredProcess[]>)
.catch(throwCatch)

export const killUnregistered = (instance: string, folderId: number, uid: string, force: boolean) => BaseAxios.post(
    instanceUrl(instance, `unregistered/${encodeURIComponent(uid)}/kill`),
    { folderId, force }
).then(throwThen<{}>)
.catch(throwCatch)
