import { BaseAxios, throwCatch, throwThen } from '@/common/api/api.util'
import type { Replace } from '@/common/util/index.type'
import type { TileSize } from '@/feature/widget/util/tile.type'

// 백엔드: apps/core/cmd/supervisor/router/processApi.go (group "/processes") + processDto.go
// processResponse{uid, type, nodeId, deviceKey, cmd, args, env, cwd, rows, cols, status, pid,
//   exitCode, createdAt(unix sec), startedAt(*unix sec), finishedAt(*unix sec), updatedAt(*unix sec), sizeOwner?}

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
    sizeOwner?: boolean     // 요청 탭이 PTY 크기 소유자인지. REST 응답에만 있고 소켓 PROCESS:UPDATE엔 없음
}
export type ProcessResponseDto = Replace<ProcessResponse, {
    createdAt: number
    startedAt: number | null
    finishedAt: number | null
    updatedAt: number | null
}>

export const toProcessResponse = (res: ProcessResponseDto): ProcessResponse => {
    const { createdAt, startedAt, finishedAt, updatedAt, ...other } = res
    return {
        ...other,
        createdAt: new Date(createdAt * 1000),
        startedAt: startedAt && startedAt !== 0 ? new Date(startedAt * 1000) : null,
        finishedAt: finishedAt && finishedAt !== 0 ? new Date(finishedAt * 1000) : null,
        updatedAt: updatedAt && updatedAt !== 0 ? new Date(updatedAt * 1000) : null,
    }
}

// GET /processes — 그 계정의 살아 있는(PENDING/PROCESS) process, sizeOwner는 요청 탭 기준
export const listProcesses = () => BaseAxios.get(
    '/processes'
).then(throwThen<ProcessResponseDto[]>)
.then(res => res.map(toProcessResponse))
.catch(throwCatch)

// GET /processes/lookup?uids=a&uids=b — 요청한 uid만, 상태 무관(끝난 것 포함). 남의 것·없는 uid는 빠짐
// axios 기본 배열 직렬화는 uids[]=a라서 반복 형식(indexes: null)으로 보낸다
export const lookupProcesses = (uids: string[]) => BaseAxios.get(
    '/processes/lookup',
    { params: { uids }, paramsSerializer: { indexes: null } }
).then(throwThen<ProcessResponseDto[]>)
.then(res => res.map(toProcessResponse))
.catch(throwCatch)

export type ExecProcessRequest = {
    nodeId: number
    authKey: string     // worker 인스턴스 키(main#sub). 후보 = listWorkers(nodeId)
    type?: ProcessType   // 비우면 EXEC
    parentTileId?: string   // 터미널 타일을 붙일 폴더 타일. 없거나 닫혔으면 서버가 루트에 붙임
    size?: TileSize         // 없으면 0.5×0.5
}

// 서버가 타일 트리에 넣은 터미널 타일 (tileTree.store TerminalTile과 같은 모양) — ⑪-4(타일 없는 실행 중 process)
export type ExecTile = {
    id: string
    type: 'terminal'
    parent: string
    kids: string[]
    size: TileSize
    nodeId: number
    uid: string
}

// process가 없는 실행(폴더)이면 tile·tileVersion 없음
export type ExecProcessResponse = {
    proc: ProcessResponse
    tile?: ExecTile
    tileVersion?: number
}
type ExecProcessResponseDto = ProcessResponseDto & Omit<ExecProcessResponse, 'proc'>

// POST /processes/exec — 노드 실행(계정의 모든 탭 소켓이 구독됨, 요청 탭 = 크기 소유자). X-Tab-Id 필수
export const execProcess = (param: ExecProcessRequest) => BaseAxios.post(
    '/processes/exec',
    param
).then(throwThen<ExecProcessResponseDto>)
.then(({ tile, tileVersion, ...proc }): ExecProcessResponse => ({ proc: toProcessResponse(proc), tile, tileVersion }))
.catch(throwCatch)

// POST /processes/kill/:processId — 실행 중인 process 종료
export const killProcess = (processId: string) => BaseAxios.post(
    `/processes/kill/${processId}`
).then(throwThen<Record<string, never>>)
.catch(throwCatch)

export type ResizeProcessRequest = {
    rows: number
    cols: number
}

// POST /processes/resize/:processId — 실행 중인 process의 PTY 창 크기 변경(worker 응답 확인 후
// DB/memory 동기화까지 끝난 최신 process 객체를 반환 — 실패 시 DB/memory는 그대로). 크기 소유 탭이 아니면 403
export const resizeProcess = (processId: string, param: ResizeProcessRequest) => BaseAxios.post(
    `/processes/resize/${processId}`,
    param
).then(throwThen<ProcessResponseDto>)
.then(toProcessResponse)
.catch(throwCatch)
