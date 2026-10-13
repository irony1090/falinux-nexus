import { BaseAxios, throwCatch, throwThen } from '@/common/api/api.util'
import type { QueryClient } from '@tanstack/query-core'
import { useQuery } from '@tanstack/vue-query'
import { computed, type Ref } from 'vue'

// 백엔드: apps/core/cmd/supervisor/router/share.go (PUT /nodes/:id/shares/:userId) + user.go (GET /users/lookup)

export type LookupUserResponse = { id: number, nickname: string }

// 아이디(users.identification) 정확히 일치만 — 없는 아이디 404 / 자기 자신 400
export const lookupUser = (identification: string) => BaseAxios.get(
    '/users/lookup',
    { params: { identification } }
).then(throwThen<LookupUserResponse>)
.catch(throwCatch)

export type NodeShareEntry = {
    userId: number
    nickname: string
    permission: number
    direct: boolean
    sourceNodeId: number
    sourceName: string
}
export type PutNodeShareResponse = {
    entry: NodeShareEntry | null
    previous: number | null     // 이 노드에 있던 직접 단계, 없으면 null
}

// 폴더만(스크립트 400). 같은 사람에게 다시 걸면 단계 변경(upsert). tileId = 공유 버튼을 누른 폴더 타일(받는 쪽 타일이 그 uid를 받음)
export const putNodeShare = (nodeId: number, userId: number, permission: number, tileId?: string) => BaseAxios.put(
    `/nodes/${nodeId}/shares/${userId}`,
    { permission, tileId }
).then(throwThen<PutNodeShareResponse>)
.catch(throwCatch)

// 이 노드에 실제로 적용되는 공유 전부(직접 + 상위에서 이어진 것, 차단 포함, 소유자 제외) — 소유자만
export const listNodeShares = (nodeId: number) => BaseAxios.get(
    `/nodes/${nodeId}/shares`
).then(throwThen<NodeShareEntry[]>)
.catch(throwCatch)

// 직접 건 행만 지운다. 응답 entry = 지운 뒤 그 계정에 적용되는 공유(상위에서 이어지면 그 값, 없으면 null)
export const deleteNodeShare = (nodeId: number, userId: number) => BaseAxios.delete(
    `/nodes/${nodeId}/shares/${userId}`
).then(throwThen<{ entry: NodeShareEntry | null }>)
.catch(throwCatch)

const SHARES_KEY = ['SHARE', 'NODE'] as const

// O1 숫자·O2 목록 — 공유를 걸거나 풀면 invalidateNodeShares로 다시 조회
export const useNodeShares = (nodeId: Ref<number | undefined>, enabled: Ref<boolean>) => useQuery({
    enabled: computed(() => enabled.value && nodeId.value !== undefined),
    queryKey: computed(() => [...SHARES_KEY, nodeId.value]),
    queryFn: ({ queryKey }) => listNodeShares(queryKey[2] as number),   // ref 말고 키에서 — node.api.ts RUNS_KEY 주석
})

// 폴더 목록 행의 shareCount(O3)도 같이 갱신
export const invalidateNodeShares = (queryClient: QueryClient, nodeId: number) => Promise.all([
    queryClient.invalidateQueries({ queryKey: [...SHARES_KEY, nodeId] }),
    queryClient.invalidateQueries({ queryKey: ['NODE', 'LIST'] }),
    queryClient.invalidateQueries({ queryKey: ['SHARE', 'MINE'] }),
])

// ---- 내가 공유한 것(O5) — 노드 + 살아 있는 process 단독 공유, 해제만 ----
export type MyShare = {
    kind: 'NODE' | 'PROCESS'
    nodeId?: number
    processUid?: string
    name: string
    userId: number
    nickname: string
    permission: number
}

export const listMyShares = () => BaseAxios.get(
    '/shares/mine'
).then(throwThen<MyShare[]>)
.catch(throwCatch)

export const deleteProcessShare = (processUid: string, userId: number) => BaseAxios.delete(
    `/processes/${processUid}/shares/${userId}`
).then(throwThen<{ entry: NodeShareEntry | null }>)
.catch(throwCatch)

const MINE_KEY = ['SHARE', 'MINE'] as const

export const useMyShares = (enabled: Ref<boolean>) => useQuery({
    enabled,
    queryKey: MINE_KEY,
    queryFn: listMyShares,
})

export const invalidateMyShares = (queryClient: QueryClient) => queryClient.invalidateQueries({ queryKey: MINE_KEY })

// "현재 타일에 바인딩"(개편-4) — 그 터미널 타일을 어느 보드에 있든 hash 보드의 tileId 폴더 타일 아래로 옮긴다. 없으면 새로 만듦(살아 있는 것만).
// 이미 그 타일 아래면 added=false. 결과는 TILES:UPDATE·TILES:NOTICE로 옴
export const bindProcessTile = (uid: string, hash: string, tileId: string) => BaseAxios.post(
    `/processes/${uid}/tile`,
    { hash, tileId },
).then(throwThen<{ added: boolean }>)
.catch(throwCatch)

