import { BaseAxios, throwCatch, throwThen } from '@/common/api/api.util'
import type { QueryOptions } from '@/common/api/query.util'
import type { Replace } from '@/common/util/index.type'
import type { QueryClient } from '@tanstack/query-core'
import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { computed, type Ref } from 'vue'

// 백엔드: apps/core/cmd/supervisor/router/node.go (group "/nodes") + nodeDto.go
// nodeResponse{id, parentId, kind, name, ord, position, deviceKey, content, createdAt(unix sec), updatedAt(*unix sec)}

export type NodeKind = 'FOLDER' | 'SCRIPT'

export type NodePoint = {
    x: number
    y: number
}

export type NodeResponse = {
    id: number
    parentId: number | null
    kind: NodeKind
    name: string
    ord: number
    position: NodePoint | null // 미배치=null
    deviceKey: string | null   // FOLDER 전용
    content: string | null     // SCRIPT 전용
    createdAt: Date            // unix sec (.Unix())
    updatedAt: Date | null     // nullable
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
        cur = node.parentId
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

    return {
        invalidateAllForce,
        invalidateAll,
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
        queryFn: () => listChildren(parentId?.value)
    })
}

export const useGetNode = (id: Ref<number>, { enabled }: QueryOptions = {}) => {
    const isReady = computed(() => !enabled ? !isNaN(id.value) : enabled.value)

    return useQuery({
        enabled: isReady,
        queryKey: Q_KEY.DETAIL(id),
        queryFn: () => getNode(id.value)
    })
}

// 브레드크럼용 경로(루트 -> id). id가 null(루트 목록)이면 조회하지 않는다
export const useNodePath = (id: Ref<number | null>) => {
    const queryClient = useQueryClient()

    return useQuery({
        enabled: computed(() => id.value !== null),
        queryKey: Q_KEY.PATH(id),
        queryFn: () => id.value === null ? [] : fetchNodePath(queryClient, id.value)
    })
}
