import { BaseAxios, throwCatch, throwThen } from '@/common/api/api.util'
import type { TileTreeResponse } from './tile.api'

// 백엔드: apps/core/cmd/supervisor/router/workspace.go (group "/workspaces") — 화면 이름 "보드", 코드 이름 workspace.
// 분리·이동은 출발 보드 version이 다르면 409 {hash, tree, version}(PUT /tiles와 같은 모양) — WS2(변경 방식), REF-workspace-impl.md

export type BoardItem = {
    hash: string
    name: string
    isDefault: boolean
    ord: number
}

// 만들기·이동·삭제 = 요청한 탭이 보던 트리(만들기·이동 = 출발, 삭제 = 기본), 이름 변경 = workspace만
export type BoardChangeResponse = Partial<TileTreeResponse> & { workspace?: BoardItem }

export type BoardChangeResult =
    | ({ ok: true } & BoardChangeResponse)
    | ({ ok: false } & TileTreeResponse)

const conflictable = { validateStatus: (s: number) => s === 200 || s === 409 }

const toResult = (res: { status: number; data: unknown }): BoardChangeResult => res.status === 409
    ? { ok: false, ...(res.data as TileTreeResponse) }
    : { ok: true, ...(res.data as BoardChangeResponse) }

// GET /workspaces — 계정의 보드 목록(기본이 없으면 서버가 만듦), ord 순
export const listBoards = () => BaseAxios.get(
    '/workspaces'
).then(throwThen<BoardItem[]>)
.catch(throwCatch)

// POST /workspaces — 기본 보드의 폴더 타일을 하위째 떼어 새 보드로(W1(분리 범위)). version = 기본 보드 version
export const createBoard = (param: { from: string; folderTileId: string; name: string; version: number }) => BaseAxios.post(
    '/workspaces', param, conflictable
).then(toResult)
.catch(throwCatch)

// POST /workspaces/:hash/move — 폴더 타일을 하위째 다른 보드 루트 끝으로(W3(이동 범위)). version = 출발 보드 version
export const moveBoardFolder = (from: string, param: { folderTileId: string; to: string; version: number }) => BaseAxios.post(
    `/workspaces/${from}/move`, param, conflictable
).then(toResult)
.catch(throwCatch)

// PATCH /workspaces/:hash — 이름 변경(빈 이름 허용, 64자까지)
export const renameBoard = (hash: string, name: string) => BaseAxios.patch(
    `/workspaces/${hash}`, { name }
).then(throwThen<BoardChangeResponse>)
.catch(throwCatch)

// DELETE /workspaces/:hash — 안의 타일 전부(실행 중 터미널 포함, kill 없음)를 기본 보드 루트 끝으로 옮기고 삭제(W2(삭제 처리))
export const deleteBoard = (hash: string) => BaseAxios.delete(
    `/workspaces/${hash}`
).then(throwThen<BoardChangeResponse>)
.catch(throwCatch)
