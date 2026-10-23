import { BaseAxios, throwCatch, throwThen } from '@/common/api/api.util'
import type { DirEntry } from './worker.api'

// 백엔드: apps/core/cmd/supervisor/router/upload.go·transferList.go·download.go·copy.go — REF-transfer-ext.md, REF-transfer-down.md

export type TransferKind = 'upload' | 'download' | 'copy'
export type TransferPhase = 'upload' | 'send' | 'download'     // 브라우저 -> supervisor / supervisor -> worker / worker -> 브라우저 — Q25(진행률 구간)
export type TransferState = 'running' | 'done' | 'failed' | 'aborted'

export type TransferInfo = {
    id: string
    kind: TransferKind
    tileId?: string         // 시작한 장비 타일(worker 간 = 받는 타일) — Q27(타일 막대 범위)
    instanceKey: string     // 받는 장비(내려받기 = 보내는 장비)
    dir: string             // 받을 폴더(내려받기 = 내려받는 폴더)
    name: string
    size: number            // -1 = 모름(zip)
    done: number            // 지금 단계에서 옮긴 바이트
    fromKey?: string        // worker 간: 보내는 장비
    fromPath?: string       // worker 간: 보내는 파일
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

// 내려받기 = 전송-8·9 — path가 있으면 그 파일 하나, 없으면 dir 안의 names를 zip으로(Q33(내려받기 범위))
export type DownloadStart = { folderId: number; dir: string; names: string[]; path?: string; tileId: string }

export const startDownload = (instance: string, body: DownloadStart) => BaseAxios.post(
    `${workerPath(instance)}/downloads`,
    body
).then(throwThen<TransferInfo>)
.catch(throwCatch)

// 브라우저 기본 다운로드가 받는 주소(한 번만, 쿠키 세션)
export const downloadUrl = (id: string) => `${BaseAxios.defaults.baseURL}/transfers/${encodeURIComponent(id)}/data`

// worker 간 전송 = 전송-10, 파일 하나 = 요청 하나(Q36(worker 간 폴더))
type DeviceRef = { folderId: number; mainKey: string; subKey: string }
export type CopyStart = { from: DeviceRef & { path: string }; to: DeviceRef & { dir: string }; overwrite: boolean; tileId: string }

export const deviceRef = (instance: string, folderId: number): DeviceRef => {
    const i = instance.lastIndexOf('#')
    return { folderId, mainKey: instance.slice(0, i), subKey: instance.slice(i + 1) }
}

export const startCopy = (body: CopyStart) => BaseAxios.post(
    '/transfers/copies',
    body
).then(throwThen<TransferInfo>)
.catch(throwCatch)
