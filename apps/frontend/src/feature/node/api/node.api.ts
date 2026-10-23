import { BaseAxios, throwCatch, throwThen, type CatchErrorResponse } from '@/common/api/api.util'
import type { QueryOptions } from '@/common/api/query.util'
import type { Replace } from '@/common/util/index.type'
import type { QueryClient } from '@tanstack/query-core'
import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { computed, ref, type Ref } from 'vue'

// 백엔드: apps/core/cmd/supervisor/router/node.go (group "/nodes") + nodeDto.go
// nodeResponse{id, parentId, ownerUserId, kind, name, ord, position, deviceKey, content, createdAt(unix sec), updatedAt(*unix sec)}

export type NodeKind = 'FOLDER' | 'SCRIPT'

export type NodePoint = {
    x: number
    y: number
}

export type NodeResponse = {
    id: number
    parentId: number | null
    ownerUserId: number        // 같은 소유자 판정(잘라내기 붙여넣기) — CP3(붙여넣기 제약)
    kind: NodeKind
    name: string
    ord: number
    position: NodePoint | null // 미배치=null
    deviceKey: string | null   // FOLDER 전용
    content: string | null     // SCRIPT 전용
    createdAt: Date            // unix sec (.Unix())
    updatedAt: Date | null     // nullable
    permission?: number        // 요청 계정의 단계(REST 응답만)
    ownerNickname?: string     // 공유받은 루트 목록 항목만 — SD8(받는 쪽 루트 목록)
    shareCount?: number        // 목록 응답만 — 내 폴더에 직접 건 공유 수(차단 제외, 0이면 없음) — O3
    parentVisible?: boolean    // GET /nodes/:id만 — 부모를 볼 수 있나, false면 경로가 여기서 멈춤(SD9(경로 줄 멈춤))
}
type NodeResponseDto = Replace<NodeResponse, {
    createdAt: number
    updatedAt: number | null
}>

export type CreateNodeRequest = {
    kind: NodeKind
    name: string
    parentId?: number | null  // 생략/null=루트
    ord?: number | null       // 생략/null=형제 끝에 append(핸들러 계산)
    position?: NodePoint | null // 생략/null=캔버스 미배치
    deviceKey?: string | null   // FOLDER 전용
    content?: string | null     // SCRIPT 전용
}

// PatchField는 백엔드 internal/patch.Field[T]의 {valid, value} 3-state 와이어 형식이다.
// 호출부가 이 형태를 직접 다루면 관리가 힘드므로(PatchNodeRequest는 평범한 값만 받음),
// patchNode 내부에서만 변환한다: undefined=valid:false(건드리지 않음) / null·값=valid:true.
type PatchField<T> = { valid: true; value: T | null } | { valid: false }

const toPatchField = <T>(value: T | null | undefined): PatchField<T> =>
    value === undefined ? { valid: false } : { valid: true, value }

export type PatchNodeRequest = Partial<{
    name: string | null
    ord: number | null
    parentId: number | null
    position: NodePoint | null
    deviceKey: string | null
    content: string | null
}>

const toNodeResponse = (res: NodeResponseDto): NodeResponse => {
    const { createdAt, updatedAt, ...other } = res
    const rst: NodeResponse = {
        ...other,
        createdAt: new Date(createdAt * 1000),
        updatedAt: updatedAt && updatedAt !== 0 ? new Date(updatedAt * 1000) : null
    };
    return rst
}

// POST /nodes — FOLDER/SCRIPT 노드 생성
export const createNode = (param: CreateNodeRequest) => BaseAxios.post(
    '/nodes',
    param
).then(throwThen<NodeResponseDto>)
.then(toNodeResponse)
.catch(throwCatch)

// GET /nodes?parentId= — 자식 목록(parentId 없으면 루트)
export const listChildren = (parentId?: number) => BaseAxios.get(
    '/nodes',
    { params: { parentId } }
).then(throwThen<NodeResponseDto[]>)
.then(res => res.map(toNodeResponse))
.catch(throwCatch)

// 폴더 자식 스크립트에서 실행 중인 process + 내 보드에서의 위치 — 실시간-6a(스크립트 행 "실행 중 N")
export type RunResponse = {
    uid: string
    nodeId: number | null  // null = 장비 파일 실행(폴더-4(실행))
    name: string           // 스크립트 이름 — 단독 공유 process는 노드를 못 봐도 이걸로. 장비 실행 = 파일 이름
    cwd?: string           // 장비 실행만 — 경로 표시
    status: string
    permission: number     // 요청 계정의 process 단계
    startedBy: string      // 실행한 계정 닉네임
    startedByMe?: boolean  // 내가 실행 — "나"로 표시
    terminating?: boolean  // 종료 요청됨 — ProcessResponse와 같은 뜻
    forced?: boolean
    hash?: string          // 내 보드 중 이 터미널 타일이 있는 곳(없으면 위치 없음)
    tileId?: string
    minimized?: boolean
}

// GET /nodes/:id/runs — id 0 = 루트 목록. 살아 있는 것 + 내 보드에 타일이 있는 끝난 것(R2-a(끝난 process)).
// deep = 하위 전체(터미널 탭, 개편-3(목록 API)) — id 0이면 볼 수 있는 process 전부(단독 공유 포함)
export const listRuns = (folderId: number, deep = false) => BaseAxios.get(
    `/nodes/${folderId}/runs`,
    { params: deep ? { deep: 1 } : undefined },
).then(throwThen<RunResponse[]>)
.catch(throwCatch)

// GET /workers/:mainKey/runs?folderId= — 루트 장비 타일 터미널 탭(이동-3(장비 목록 API)): 그 장비의 볼 수 있는 process 전부. 연 폴더 소유자만
export const listDeviceRuns = (mainKey: string, folderId: number) => BaseAxios.get(
    `/workers/${encodeURIComponent(mainKey)}/runs`,
    { params: { folderId } },
).then(throwThen<RunResponse[]>)
.catch(throwCatch)

// GET /nodes/:id — 노드 1건(없음/타인 소유=404)
export const getNode = (id: number) => BaseAxios.get(
    `/nodes/${id}`
).then(throwThen<NodeResponseDto>)
.then(toNodeResponse)
.catch(throwCatch)

// PATCH /nodes/:id — 보낸 필드만 부분수정(undefined=미변경, null=NULL로, 값=그 값으로)
export const patchNode = (id: number, param: PatchNodeRequest) => BaseAxios.patch(
    `/nodes/${id}`,
    {
        name: toPatchField(param.name),
        ord: toPatchField(param.ord),
        parentId: toPatchField(param.parentId),
        position: toPatchField(param.position),
        deviceKey: toPatchField(param.deviceKey),
        content: toPatchField(param.content),
    }
).then(throwThen<NodeResponseDto>)
.then(toNodeResponse)
.catch(throwCatch)

// DELETE /nodes/:id — 삭제(FOLDER면 하위 subtree 동반삭제)
export const deleteNode = (id: number) => BaseAxios.delete(
    `/nodes/${id}`
).then(throwThen<Record<string, never>>)
.catch(throwCatch)

// POST /nodes/:id/copy — 노드(폴더면 하위 전체)를 parentId 아래에 복사(null = 내 루트), SH10(노드 복사)
export const copyNode = (id: number, parentId: number | null) => BaseAxios.post(
    `/nodes/${id}/copy`,
    { parentId }
).then(throwThen<{ nodes: NodeResponseDto[] }>)
.then(res => res.nodes.map(toNodeResponse))
.catch(throwCatch)

const Q_KEY = {
    LIST: (parentId?: Ref<number | undefined> | number) =>
        parentId !== undefined
        ? ['NODE', 'LIST', parentId]
        : ['NODE', 'LIST'],
    DETAIL: (id: Ref<number> | number) => ['NODE', 'DETAIL', id],
    PATH: (id?: Ref<number | null> | number) =>
        id !== undefined
        ? ['NODE', 'PATH', id]
        : ['NODE', 'PATH'],
}

const RUNS_KEY = ['NODE', 'RUNS'] as const

// queryFn은 ref가 아니라 queryKey에서 값을 읽는다 — 같은 키의 쿼리는 마지막으로 조회한 observer의 queryFn을 기억하므로,
// ref를 읽으면 그 observer(타일)가 다른 폴더로 옮겨 간 뒤 무효화 재조회가 엉뚱한 폴더를 가져와 이 키에 넣는다

const PATH_DEPTH_MAX = 64   // parentId 사이클 방어

// 루트부터 id까지 경로. 경로 API가 없어 getNode를 parentId 따라 반복 호출한다(조상은 DETAIL 캐시 공유)
const fetchNodePath = async (queryClient: QueryClient, id: number) => {
    const path: Array<NodeResponse> = []
    let cur: number | null = id
    while (cur !== null && path.length < PATH_DEPTH_MAX) {
        const nodeId: number = cur
        const node: NodeResponse = await queryClient.fetchQuery({
            queryKey: Q_KEY.DETAIL(nodeId),
            queryFn: () => getNode(nodeId),
        })
        path.unshift(node)
        cur = node.parentVisible === false ? null : node.parentId
    }
    return path
}

export const useNodeQueryClient = (queryClient: QueryClient) => {

    const invalidateAllForce = () => {
        queryClient.invalidateQueries({
            queryKey: Q_KEY.LIST().slice(0, 1),
            exact: false,
        })
    }

    // 이동(parentId 변경)이면 old parentId로 한 번 더 호출해 양쪽 목록을 모두 무효화한다.
    const invalidateAll = (id: Ref<number> | number, parentId?: Ref<number | undefined> | number) => {
        queryClient.invalidateQueries({
            queryKey: Q_KEY.LIST(parentId),
            exact: !!parentId,
        })

        queryClient.invalidateQueries({
            queryKey: Q_KEY.DETAIL(id),
            exact: true,
        })

        // 조상의 이름·위치가 바뀌면 자손 경로도 달라지므로 경로는 전부 무효화
        queryClient.invalidateQueries({
            queryKey: Q_KEY.PATH(),
            exact: false,
        })
    }

    // 다른 곳에서 바뀐 노드 — 부모 목록 + 그 노드를 담은 목록 전부(옛 부모·루트의 공유 루트 자리) — NS4(프론트 반영)
    const invalidateChanged = (node: { id: number, parentId: number | null }) => {
        const parentKey = node.parentId ?? undefined
        queryClient.invalidateQueries({
            queryKey: Q_KEY.LIST().slice(0, 2),
            predicate: q => q.queryKey[2] === parentKey
                || (q.state.data as NodeResponse[] | undefined)?.some(n => n.id === node.id) === true,
        })
        queryClient.invalidateQueries({ queryKey: Q_KEY.DETAIL(node.id), exact: true })
        queryClient.invalidateQueries({ queryKey: Q_KEY.PATH(), exact: false })
    }

    return {
        invalidateAllForce,
        invalidateAll,
        invalidateChanged,
        fetchPath: (id: number) => fetchNodePath(queryClient, id),
    }
}

// 생성·수정 뒤 관련 조회(목록·상세·경로)를 무효화한다. 삭제는 타일 정리가 끼어서 nodeRemove.hook.ts
export const useNodeActions = () => {
    const { invalidateAll } = useNodeQueryClient(useQueryClient())

    const done = (node: NodeResponse) => {
        invalidateAll(node.id, node.parentId ?? undefined)
        return node
    }

    return {
        create: (param: CreateNodeRequest) => createNode(param).then(done),
        patch: (id: number, param: PatchNodeRequest) => patchNode(id, param).then(done),
    }
}

export const useListChildren = (parentId?: Ref<number | undefined>, { enabled }: QueryOptions = {}) => {
    return useQuery({
        enabled,
        queryKey: Q_KEY.LIST(parentId),
        queryFn: ({ queryKey }) => listChildren(queryKey[2] as number | undefined)
    })
}

export const useGetNode = (id: Ref<number>, { enabled }: QueryOptions = {}) => {
    const isReady = computed(() => !enabled ? !isNaN(id.value) : enabled.value)

    return useQuery({
        enabled: isReady,
        queryKey: Q_KEY.DETAIL(id),
        queryFn: ({ queryKey }) => getNode(queryKey[2] as number),
        // 404 = 볼 수 없음(공유 취소·삭제) — 재시도하지 않고 바로 알린다(실시간-6b)
        retry: (count, err) => (err as CatchErrorResponse)?.status !== 404 && count < 3,
    })
}

// 브레드크럼용 경로(루트 -> id). id가 null(루트 목록)이면 조회하지 않는다
export const useNodePath = (id: Ref<number | null>) => {
    const queryClient = useQueryClient()

    return useQuery({
        enabled: computed(() => id.value !== null),
        queryKey: Q_KEY.PATH(id),
        queryFn: ({ queryKey }) => queryKey[2] === null ? [] : fetchNodePath(queryClient, queryKey[2] as number)
    })
}

// 폴더(null = 루트 목록)의 자식 스크립트별 실행 중 목록. 갱신은 useScriptRuns가 소켓 이벤트로 무효화
export const useNodeRuns = (folderId: Ref<number | null>, deep = false, enabled: Ref<boolean> = ref(true)) => useQuery({
    enabled,
    queryKey: [...RUNS_KEY, folderId, deep],
    queryFn: ({ queryKey }) => listRuns((queryKey[2] as number | null) ?? 0, deep),
})

// 장비(main_key)의 process — 무효화는 노드 목록과 같은 키(invalidateNodeRuns)
export const useDeviceRuns = (mainKey: Ref<string>, folderId: Ref<number>, enabled: Ref<boolean>) => useQuery({
    enabled,
    queryKey: [...RUNS_KEY, 'DEVICE', mainKey, folderId],
    queryFn: ({ queryKey }) => listDeviceRuns(queryKey[3] as string, queryKey[4] as number),
})

export const invalidateNodeRuns = (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: RUNS_KEY })
