import { BaseAxios, throwCatch, throwThen } from '@/common/api/api.util'
import { useQuery } from '@tanstack/vue-query'
import { computed, type Ref } from 'vue'

// 백엔드: apps/core/cmd/supervisor/router/presence.go — REST 조회만(push 없음). SH6(presence 범위·전달)

export type PresenceTab = {
    connected: boolean      // false = 끊긴 뒤 60초 보관 중
    viewing: boolean        // 이 터미널이 있는 작업 공간을 띄운 탭
    sizeOwner: boolean      // PTY 크기 우선권
    since: number           // unix sec
}
export type PresenceSession = {
    key: string             // 세션 구분용 해시
    userAgent: string
    tabs: PresenceTab[]
}
export type PresenceAccount = {
    userId: number
    nickname: string
    permission: number
    online: boolean
    lastSeen?: number       // unix sec, 접속 중이 아닐 때만
    sessions: PresenceSession[]
}

export const getProcessPresence = (uid: string) => BaseAxios.get(
    `/processes/${uid}/presence`
).then(throwThen<PresenceAccount[]>)
.catch(throwCatch)

export type UserPresence = { userId: number, online: boolean, lastSeen?: number }

export const getUsersPresence = (ids: number[]) => BaseAxios.get(
    '/users/presence',
    { params: { ids: ids.join(',') } }
).then(throwThen<UserPresence[]>)
.catch(throwCatch)

const REFRESH_MS = 5000

// 패널이 열려 있는 동안만 조회하고 주기적으로 다시 조회한다
export const useProcessPresence = (uid: Ref<string | undefined>, enabled: Ref<boolean>) => useQuery({
    enabled: computed(() => enabled.value && !!uid.value),
    queryKey: computed(() => ['PRESENCE', 'PROCESS', uid.value]),
    queryFn: () => getProcessPresence(uid.value!),
    refetchInterval: REFRESH_MS,
})

export const useUsersPresence = (ids: Ref<number[]>, enabled: Ref<boolean>) => useQuery({
    enabled: computed(() => enabled.value && ids.value.length > 0),
    queryKey: computed(() => ['PRESENCE', 'USERS', ...ids.value]),
    queryFn: () => getUsersPresence(ids.value),
    refetchInterval: REFRESH_MS,
})
