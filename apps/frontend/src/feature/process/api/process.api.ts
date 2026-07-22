import { BaseAxios, throwCatch, throwThen } from '@/common/api/api.util'
import type { Replace } from '@/common/util/index.type'

// 백엔드: apps/core/cmd/supervisor/router/processApi.go (group "/processes") + processDto.go
// processResponse{uid, type, nodeId, deviceKey, cmd, args, env, cwd, rows, cols, status, pid,
//   exitCode, createdAt(unix sec), startedAt(*unix sec), finishedAt(*unix sec), updatedAt(*unix sec)}

export type ProcessType = 'EXEC' | 'EDIT'
export type ProcessStatus = 'PENDING' | 'PROCESS' | 'COMPLETED' | 'FAILED'

export type ProcessResponse = {
    uid: string
    type: ProcessType
    nodeId: number | null
    deviceKey: string
    cmd: string
    args: string[]
    env: string[]
    cwd: string
    rows: number
    cols: number
    status: ProcessStatus
    pid: number | null
    exitCode: number | null
    createdAt: Date
    startedAt: Date | null
    finishedAt: Date | null
    updatedAt: Date | null
}
type ProcessResponseDto = Replace<ProcessResponse, {
    createdAt: number
    startedAt: number | null
    finishedAt: number | null
    updatedAt: number | null
}>

const toProcessResponse = (res: ProcessResponseDto): ProcessResponse => {
    const { createdAt, startedAt, finishedAt, updatedAt, ...other } = res
    return {
        ...other,
        createdAt: new Date(createdAt * 1000),
        startedAt: startedAt && startedAt !== 0 ? new Date(startedAt * 1000) : null,
        finishedAt: finishedAt && finishedAt !== 0 ? new Date(finishedAt * 1000) : null,
        updatedAt: updatedAt && updatedAt !== 0 ? new Date(updatedAt * 1000) : null,
    }
}

// GET /processes/subscriptions — 현재 세션(sid)이 구독 중인 process 목록
export const listSubscriptions = () => BaseAxios.get(
    '/processes/subscriptions'
).then(throwThen<ProcessResponseDto[]>)
.then(res => res.map(toProcessResponse))
.catch(throwCatch)

// POST /processes/subscribe/:processId — 현재 세션을 구독자로 등록(수동 구독)
export const subscribeProcess = (processId: string) => BaseAxios.post(
    `/processes/subscribe/${processId}`
).then(throwThen<Record<string, never>>)
.catch(throwCatch)

// DELETE /processes/subscribe/:processId — 구독 해지
export const unsubscribeProcess = (processId: string) => BaseAxios.delete(
    `/processes/subscribe/${processId}`
).then(throwThen<Record<string, never>>)
.catch(throwCatch)

export type ExecProcessRequest = {
    nodeId: number
    authKey: string     // worker 인스턴스 키(main#sub). 인스턴스 선택 UI 미착수 — 지금은 직접 지정.
    type?: ProcessType   // 비우면 EXEC
}

// POST /processes/exec — 노드 실행(요청 세션이 별도 구독 요청 없이 자동으로 구독자가 됨)
export const execProcess = (param: ExecProcessRequest) => BaseAxios.post(
    '/processes/exec',
    param
).then(throwThen<{ uid: string }>)
.catch(throwCatch)

// POST /processes/kill/:processId — 실행 중인 process 종료(요청 세션 자동 구독)
export const killProcess = (processId: string) => BaseAxios.post(
    `/processes/kill/${processId}`
).then(throwThen<Record<string, never>>)
.catch(throwCatch)
