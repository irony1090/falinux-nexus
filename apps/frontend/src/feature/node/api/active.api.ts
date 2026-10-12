import { BaseAxios, throwCatch, throwThen } from '@/common/api/api.util'

// 백엔드: apps/core/cmd/supervisor/router/active.go — 다른 계정 탭의 active 타일(원격 active 배지)

// ACTIVE:UPDATE 이벤트와 같은 모양. uid·nodeId 둘 다 없음 = 그 탭의 active 지움
export type RemoteActive = {
    key: string             // 탭 구분용 해시
    userId: number
    nickname: string
    uid?: string            // 터미널
    nodeId?: number         // 폴더
}

// 다른 계정 탭의 active 중 내가 볼 수 있는 것
export const listActive = () => BaseAxios.get(
    '/tabs/active'
).then(throwThen<RemoteActive[]>)
.catch(throwCatch)
