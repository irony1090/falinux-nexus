import { BaseAxios, throwCatch, throwThen } from '@/common/api/api.util'

// 백엔드: apps/core/cmd/supervisor/router/account.go — 중복 확인·비밀번호·닉네임·로그인한 기기 → REF-account.md
type Empty = Record<string, never>

export type AvailableResponse = { available: boolean, message: string }
export type AvailableKind = 'identification' | 'nickname'

export type LoginResponse = {
    key: string
    userAgent: string
    createdAt: number   // unix sec
    lastSeenAt: number  // unix sec, 5분 단위로 기록
    current: boolean    // 이 기기
    online: boolean     // 연결된 탭이 있음
}

// GET /users/available — 로그인 없이(가입 화면). 저장 때 서버가 다시 확인한다
export const checkAvailable = (kind: AvailableKind, value: string) => BaseAxios.get(
    '/users/available', { params: { [kind]: value } },
).then(throwThen<AvailableResponse>).catch(throwCatch)

export const changePassword = (current: string, next: string) => BaseAxios.put('/users/password', { current, next })
.then(throwThen<Empty>).catch(throwCatch)

export const changeNickname = (nickname: string) => BaseAxios.put('/users/nickname', { nickname })
.then(throwThen<Empty>).catch(throwCatch)

export const listLogins = () => BaseAxios.get('/users/sessions')
.then(throwThen<{ items: LoginResponse[] }>).then(r => r.items).catch(throwCatch)

// current = 이 기기였음(쿠키도 만료됨)
export const endLogin = (key: string) => BaseAxios.delete(`/users/sessions/${encodeURIComponent(key)}`)
.then(throwThen<{ current: boolean }>).catch(throwCatch)

// 이 기기 포함 전부
export const endAllLogins = () => BaseAxios.delete('/users/sessions')
.then(throwThen<Empty>).catch(throwCatch)
