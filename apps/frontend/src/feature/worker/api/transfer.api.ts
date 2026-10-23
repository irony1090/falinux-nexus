import { BaseAxios, throwCatch, throwThen } from '@/common/api/api.util'
import type { DirEntry } from './worker.api'

// 백엔드: apps/core/cmd/supervisor/router/upload.go·transferList.go — 파트(파일 전송 확장), REF-transfer-ext.md

export type TransferPhase = 'upload' | 'send'      // 브라우저 -> supervisor / supervisor -> worker — Q25(진행률 구간)
export type TransferState = 'running' | 'done' | 'failed' | 'aborted'

export type TransferInfo = {
    id: string
    kind: 'upload'
    tileId?: string         // 시작한 장비 타일 — Q27(타일 막대 범위)
    instanceKey: string
    dir: string
    name: string
    size: number
    done: number            // 지금 단계에서 옮긴 바이트
    phase: TransferPhase
    state: TransferState
    code?: string           // exists | is_dir | no_parent | bad_path ...
    reason?: string
    startedAt: number       // unix ms
}

// 앞단 nginx client_max_body_size 기본 1MB 아래 — Q30(조각 업로드)
export const UPLOAD_CHUNK = 512 * 1024

const workerPath = (instance: string) => {
    const i = instance.lastIndexOf('#')
    return `/workers/${encodeURIComponent(instance.slice(0, i))}/${encodeURIComponent(instance.slice(i + 1))}`
}

// POST /workers/:m/:s/dir/exists — 이미 있는 이름만(Q28(여러 파일 겹침))
export const dirExists = (instance: string, folderId: number, path: string, names: string[]) => BaseAxios.post(
    `${workerPath(instance)}/dir/exists`,
    { folderId, path, names }
).then(throwThen<DirEntry[]>)
.catch(throwCatch)

export type UploadStart = { folderId: number; path: string; name: string; size: number; overwrite: boolean; tileId: string }

export const startUpload = (instance: string, body: UploadStart) => BaseAxios.post(
    `${workerPath(instance)}/uploads`,
    body
).then(throwThen<TransferInfo>)
.catch(throwCatch)

export const putUploadChunk = (id: string, offset: number, data: Blob, signal?: AbortSignal) => BaseAxios.put(
    `/transfers/${encodeURIComponent(id)}/data`,
    data,
    { params: { offset }, signal, headers: { 'Content-Type': 'application/octet-stream' } }
).then(throwThen<{ received: number }>)
.catch(throwCatch)

export const finishUpload = (id: string) => BaseAxios.post(
    `/transfers/${encodeURIComponent(id)}/finish`
).then(throwThen<TransferInfo>)
.catch(throwCatch)

export const abortTransfer = (id: string) => BaseAxios.delete(
    `/transfers/${encodeURIComponent(id)}`
).then(() => undefined)
.catch(throwCatch)

export const listTransfers = () => BaseAxios.get(
    '/transfers'
).then(throwThen<TransferInfo[]>)
.catch(throwCatch)
