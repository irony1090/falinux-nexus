import { BaseAxios, throwCatch, throwThen } from '@/common/api/api.util'
import { keepPreviousData, useQueries, useQuery, type QueryClient } from '@tanstack/vue-query'
import { computed, toValue, type MaybeRefOrGetter } from 'vue'
import type { WorkerInstanceResponse } from '@/feature/worker/api/worker.api'

// 백엔드: apps/core/cmd/supervisor/router/device.go("/devices") · admin.go("/admin") — REF-member.md

export type DeviceResponse = {
    mainKey: string
    ownerId: number
    ownerNickname?: string      // 관리자 목록에서만
    createdAt: number           // Unix 초
    keyIssuedAt: number
    instances: WorkerInstanceResponse[]
}

export type DeviceKeyResponse = { mainKey: string, apiKey: string }

export type AdminUserResponse = {
    id: number
    nickname: string
    perms: string[]
    workerLimit: number
    deviceCount: number
    createdAt: number
    suspendedAt: number | null
}

// 목록 = 서버에서 검색·필터·페이지(화-4(서버 목록)). size 0 = 전부
export type ListParams = { q?: string, f?: string, page?: number, size?: number, owner?: number }
export type ListResponse<T> = {
    items: T[]
    total: number                   // 필터 뒤 개수
    page: number                    // 범위 안으로 당긴 페이지
    counts: Record<string, number>  // 검색 뒤 필터별 개수
}

export type AdminUserPatch = { workerLimit?: number, canRegister?: boolean, admin?: boolean }

const enc = encodeURIComponent
type Empty = Record<string, never>

// ---- 내 장비 ----
export const listMyDevices = (params: ListParams = {}) => BaseAxios.get('/devices', { params }).then(throwThen<ListResponse<DeviceResponse>>).catch(throwCatch)

export const getMyDevice = (mainKey: string) => BaseAxios.get(`/devices/${enc(mainKey)}`).then(throwThen<DeviceResponse>).catch(throwCatch)

// apiKey 빈 값 = 새로 발급 / 있으면 붙여넣기(다른 supervisor에서 받은 키)
export const createMyDevice = (mainKey: string, apiKey = '') => BaseAxios.post('/devices', { mainKey, apiKey })
.then(throwThen<DeviceKeyResponse>).catch(throwCatch)

export const showMyDeviceKey = (mainKey: string) => BaseAxios.get(`/devices/${enc(mainKey)}/key`)
.then(throwThen<DeviceKeyResponse>).catch(throwCatch)

export const reissueMyDeviceKey = (mainKey: string, apiKey = '') => BaseAxios.post(`/devices/${enc(mainKey)}/key`, { apiKey })
.then(throwThen<DeviceKeyResponse>).catch(throwCatch)

export const deleteMyDevice = (mainKey: string) => BaseAxios.delete(`/devices/${enc(mainKey)}`)
.then(throwThen<Empty>).catch(throwCatch)

export const labelMyInstance = (mainKey: string, subKey: string, label: string) => BaseAxios.patch(`/devices/${enc(mainKey)}/instances/${enc(subKey)}`, { label })
.then(throwThen<Empty>).catch(throwCatch)

// ---- 관리자 ----
export const adminListUsers = (params: ListParams = {}) => BaseAxios.get('/admin/users', { params }).then(throwThen<ListResponse<AdminUserResponse>>).catch(throwCatch)

export const adminGetUser = (id: number) => BaseAxios.get(`/admin/users/${id}`).then(throwThen<AdminUserResponse>).catch(throwCatch)

export const adminPatchUser = (id: number, patch: AdminUserPatch) => BaseAxios.patch(`/admin/users/${id}`, patch)
.then(throwThen<Empty>).catch(throwCatch)

export const adminSuspendUser = (id: number, suspended: boolean) => BaseAxios.post(`/admin/users/${id}/suspend`, { suspended })
.then(throwThen<Empty>).catch(throwCatch)

export const adminDeleteUser = (id: number) => BaseAxios.delete(`/admin/users/${id}`).then(throwThen<Empty>).catch(throwCatch)

export const adminListDevices = (params: ListParams = {}) => BaseAxios.get('/admin/devices', { params }).then(throwThen<ListResponse<DeviceResponse>>).catch(throwCatch)

export const adminGetDevice = (mainKey: string) => BaseAxios.get(`/admin/devices/${enc(mainKey)}`).then(throwThen<DeviceResponse>).catch(throwCatch)

// 상한·등록 권한 무시(회원-Q17)
export const adminCreateDevice = (ownerId: number, mainKey: string, apiKey = '') => BaseAxios.post('/admin/devices', { ownerId, mainKey, apiKey })
.then(throwThen<DeviceKeyResponse>).catch(throwCatch)

export const adminChangeOwner = (mainKey: string, ownerId: number) => BaseAxios.patch(`/admin/devices/${enc(mainKey)}`, { ownerId })
.then(throwThen<Empty>).catch(throwCatch)

export const adminDeleteDevice = (mainKey: string) => BaseAxios.delete(`/admin/devices/${enc(mainKey)}`)
.then(throwThen<Empty>).catch(throwCatch)

export const adminShowKey = (mainKey: string) => BaseAxios.get(`/admin/devices/${enc(mainKey)}/key`)
.then(throwThen<DeviceKeyResponse>).catch(throwCatch)

export const adminReissueKey = (mainKey: string, apiKey = '') => BaseAxios.post(`/admin/devices/${enc(mainKey)}/key`, { apiKey })
.then(throwThen<DeviceKeyResponse>).catch(throwCatch)

export const adminLabelInstance = (mainKey: string, subKey: string, label: string) => BaseAxios.patch(`/admin/devices/${enc(mainKey)}/instances/${enc(subKey)}`, { label })
.then(throwThen<Empty>).catch(throwCatch)

// ---- supervisor 주소 일괄 변경(열린 장비 타일) — deviceMove.go, REF-device-move.md ----
export type DeviceMoveResult = { instanceKey: string, addrs: string[], error?: string }
const deviceBase = (admin: boolean) => admin ? '/admin/devices' : '/devices'

// targets = instanceKey(메인키#서브키). 접속 안 된 것은 그 줄만 error
export const moveDevices = (admin: boolean, targets: string[], url: string, now: boolean) => BaseAxios.post(`${deviceBase(admin)}/move`, { targets, url, now })
.then(throwThen<DeviceMoveResult[]>).catch(throwCatch)

// worker에 저장된 주소(1순위부터) — 접속 중일 때만(아니면 409)
export const instanceAddrs = (admin: boolean, mainKey: string, subKey: string) => BaseAxios.get(`${deviceBase(admin)}/${enc(mainKey)}/instances/${enc(subKey)}/addrs`)
.then(throwThen<{ addrs: string[] }>).catch(throwCatch)

// 장비 목록·장비 하나(내 장비·관리자) 전부 — 장비 상태 알림 때
export const invalidateMemberDevices = (queryClient: QueryClient) => Promise.all(
    [['MEMBER', 'DEVICES'], ['MEMBER', 'DEVICE'], ['ADMIN', 'DEVICES'], ['ADMIN', 'DEVICE']].map(queryKey => queryClient.invalidateQueries({ queryKey })))

// ---- 조회 훅 ----
const POLL_MS = 60 * 1000   // 안전망 — 평소엔 DEVICE:CHANGE 알림으로 다시 조회(deviceChange.hook.ts)
type P = MaybeRefOrGetter<ListParams>

// 목록 — 페이지를 넘길 때 앞 결과를 둬서 줄이 깜빡이지 않게
export const useMyDevices = (params: P) => useQuery({
    queryKey: computed(() => ['MEMBER', 'DEVICES', toValue(params)]),
    queryFn: () => listMyDevices(toValue(params)),
    refetchInterval: POLL_MS,
    placeholderData: keepPreviousData,
})
export const useAdminUsers = (params: P) => useQuery({
    queryKey: computed(() => ['ADMIN', 'USERS', toValue(params)]),
    queryFn: () => adminListUsers(toValue(params)),
    placeholderData: keepPreviousData,
})
export const useAdminDevices = (params: P) => useQuery({
    queryKey: computed(() => ['ADMIN', 'DEVICES', toValue(params)]),
    queryFn: () => adminListDevices(toValue(params)),
    refetchInterval: POLL_MS,
    placeholderData: keepPreviousData,
})

// 타일 하나 — admin이면 관리자 API
export const useMemberUser = (id: number) => useQuery({
    queryKey: ['ADMIN', 'USER', id],
    queryFn: () => adminGetUser(id),
    retry: false,
})
const memberDeviceQuery = (mainKey: string, admin: boolean) => ({
    queryKey: admin ? ['ADMIN', 'DEVICE', mainKey] : ['MEMBER', 'DEVICE', mainKey],
    queryFn: () => admin ? adminGetDevice(mainKey) : getMyDevice(mainKey),
    refetchInterval: POLL_MS,
    retry: false,
})
export const useMemberDevice = (mainKey: string, admin: boolean) => useQuery(memberDeviceQuery(mainKey, admin))
// 여러 장비(열린 장비 타일) — 타일과 같은 queryKey라 캐시를 같이 쓴다. 못 읽은 장비는 빠짐
export const useMemberDeviceList = (mainKeys: MaybeRefOrGetter<string[]>, admin: boolean) => useQueries({
    queries: computed(() => toValue(mainKeys).map(k => memberDeviceQuery(k, admin))),
    combine: results => results.flatMap(r => r.data ? [r.data] : []),
})

export const invalidateMember = (queryClient: QueryClient) => Promise.all([
    queryClient.invalidateQueries({ queryKey: ['MEMBER'] }),
    queryClient.invalidateQueries({ queryKey: ['ADMIN'] }),
    queryClient.invalidateQueries({ queryKey: ['WORKER', 'ROSTER'] }),
])
