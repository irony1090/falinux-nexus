import { BaseAxios, throwCatch, throwThen } from '@/common/api/api.util'
import type { Replace } from '@/common/util/index.type'

// 백엔드: apps/core/cmd/supervisor/router/user.go (group "/users")
// userResponse{identification, nickname, createdAt(unix sec), updatedAt(*unix sec)}

export type UserResponse = {
    identification: string
    nickname: string
    createdAt: Date        // unix sec (.Unix())
    updatedAt: Date | null // nullable
}
type UserResponseDto = Replace<UserResponse, {
    createdAt: number
    updatedAt: number | null
}>

export type CreateUserRequest = {
    identification: string
    nickname: string
    password: string
}

export type SignInRequest = {
    identification: string
    password: string
}

const toUserResponse = (res: UserResponseDto): UserResponse  => {
    const { createdAt, updatedAt, ...other } = res
    const rst: UserResponse = {
        ...other,
        createdAt: new Date(createdAt * 1000),
        updatedAt: updatedAt && updatedAt !== 0 ? new Date(updatedAt * 1000) : null
    };
    // rst.createdAt = new Date(res.createdAt * 1000) 
    return rst
}

// POST /users — 신규 계정 생성
export const createUser = (param: CreateUserRequest) => BaseAxios.post(
    '/users',
    param
).then(throwThen<UserResponseDto>)
.then(toUserResponse)
.catch(throwCatch)

// POST /users/session — 로그인(세션 발급)
export const signIn = (param: SignInRequest) => BaseAxios.post(
    '/users/session',
    param
).then(throwThen<UserResponseDto>)
.then(toUserResponse)
.catch(throwCatch)

// GET /users/session — 현재 세션 검증(미로그인=401)
export const checkSession = () => BaseAxios.get(
    '/users/session'
).then(throwThen<UserResponseDto>)
.then(toUserResponse)
.catch(throwCatch)

// DELETE /users/session — 로그아웃(쿠키 만료)
export const signOut = () => BaseAxios.delete(
    '/users/session'
).then(throwThen<Record<string, never>>)
.catch(throwCatch)
