import { BaseAxios, getBaseURL, throwCatch, throwThen } from '@/common/api/api.util'
import { keepPreviousData, useQuery, type QueryClient } from '@tanstack/vue-query'
import { computed, toValue, type MaybeRefOrGetter } from 'vue'
import type { ListParams, ListResponse } from '@/feature/member/api/device.api'
import type { CtCodeIssue } from '@/feature/widget/component/CtCode.vue'

// 백엔드: apps/core/cmd/supervisor/router/update.go · updateFiles.go · updateZip.go("/admin/updates") — REF-update-deploy.md "P1 착수 계획" 단위-3(API)

export type BundleKind = 'build' | 'common'
export type FileScope = 'main' | 'sub'

export type UpdDevice = {
    mainKey: string
    memo: string
    buildCount: number
    instanceCount: number
    lastBuildAt: number | null
    createdAt: number
    hasKey: boolean
    keyIssuedAt?: number | null
}

export type Updater = {
    subKey: string
    label: string
    memo: string
    os: string
    arch: string
    installDir: string
    installedVersion: string
    lastSeenAt: number | null
    online: boolean
    running: number | null
    version: string
}

export type Bundle = {
    kind: BundleKind
    owner: string
    version: string
    note: string
    sha: string
    locked: boolean
    usedBy: number
    fileCount: number
    totalSize: number
    createdByName: string
    createdAt: number
    updatedAt: number
}

export type UpdDeviceDetail = UpdDevice & { instances: Updater[]; builds: Bundle[] }
export type CommonGroup = { name: string; versions: Bundle[] }

export type UpdFile = { scope: FileScope; subKey: string; path: string; size: number; sha: string; text: boolean; mode: number; updatedAt: number }
export type UpdUse = { name: string; version: string; sha: string; locked: boolean; files: UpdFile[] }
export type UpdUser = { mainKey: string; version: string; locked: boolean }
export type BundleDetail = Bundle & { files: UpdFile[]; uses: UpdUse[]; users: UpdUser[] }

export type Ref = { kind: BundleKind; owner: string; version: string }

const enc = encodeURIComponent
type Empty = Record<string, never>
const base = '/admin/updates'
export const bundlePath = (r: Ref) => `${base}/bundles/${r.kind}/${enc(r.owner)}/${enc(r.version)}`
export const refName = (r: Ref) => `${r.owner}/${r.version}`

// ---- 장비 루트 ----
export const listUpdDevices = (params: ListParams = {}) => BaseAxios.get(`${base}/devices`, { params }).then(throwThen<ListResponse<UpdDevice>>).catch(throwCatch)
export const getUpdDevice = (mainKey: string) => BaseAxios.get(`${base}/devices/${enc(mainKey)}`).then(throwThen<UpdDeviceDetail>).catch(throwCatch)
export type UpdKey = { mainKey: string; apiKey: string }
// 만들 때 키 발급(응답에 원문) — REF-update-p2.md 정-1(접속 인증)
export const createUpdDevice = (mainKey: string, memo = '') => BaseAxios.post(`${base}/devices`, { mainKey, memo }).then(throwThen<UpdKey>).catch(throwCatch)
export const showUpdKey = (mainKey: string) => BaseAxios.get(`${base}/devices/${enc(mainKey)}/key`).then(throwThen<UpdKey>).catch(throwCatch)
export const reissueUpdKey = (mainKey: string) => BaseAxios.post(`${base}/devices/${enc(mainKey)}/key`).then(throwThen<UpdKey>).catch(throwCatch)
export const UPD_KEY_PREFIX = 'nxu_'

// ---- 인스턴스 ----
// targetDir = 설치 경로(정-5) — updater가 접속 중이고 실행 중이 아닐 때만, updater가 설정 파일을 고쳐 씀
export const patchInstance = (mainKey: string, subKey: string, body: { label?: string; memo?: string; targetDir?: string }) =>
    BaseAxios.patch(`${base}/devices/${enc(mainKey)}/instances/${enc(subKey)}`, body, { timeout: 15_000 }).then(throwThen<Empty>).catch(throwCatch)
export const deleteInstance = (mainKey: string, subKey: string) => BaseAxios.delete(`${base}/devices/${enc(mainKey)}/instances/${enc(subKey)}`).then(throwThen<Empty>).catch(throwCatch)

// updater.json 원문 — supervisor 주소 = 이 화면의 API 주소(/api 앞, http -> ws)
export const updaterConfigText = (mainKey: string, apiKey: string, targetDir = '') => JSON.stringify({
    supervisor: getBaseURL().replace(/\/api$/, '').replace(/^http/, 'ws'), mainKey, apiKey, targetDir,
}, null, 2)
export const patchUpdDevice = (mainKey: string, memo: string) => BaseAxios.patch(`${base}/devices/${enc(mainKey)}`, { memo }).then(throwThen<Empty>).catch(throwCatch)
export const deleteUpdDevice = (mainKey: string) => BaseAxios.delete(`${base}/devices/${enc(mainKey)}`).then(throwThen<Empty>).catch(throwCatch)
export const listCommons = (params: ListParams = {}) => BaseAxios.get(`${base}/commons`, { params }).then(throwThen<ListResponse<CommonGroup>>).catch(throwCatch)

// ---- 번들 ----
// from = 복사할 번들(같은 종류). 없으면 빈 번들
export const createBundle = (kind: BundleKind, owner: string, version: string, note = '', from?: { owner: string; version: string }) =>
    BaseAxios.post(`${base}/bundles/${kind}/${enc(owner)}`, { version, note, from }).then(throwThen<{ version: string }>).catch(throwCatch)
export const getBundle = (r: Ref) => BaseAxios.get(bundlePath(r)).then(throwThen<BundleDetail>).catch(throwCatch)
export const patchBundle = (r: Ref, note: string) => BaseAxios.patch(bundlePath(r), { note }).then(throwThen<Empty>).catch(throwCatch)
export const deleteBundle = (r: Ref) => BaseAxios.delete(bundlePath(r)).then(throwThen<Empty>).catch(throwCatch)

// ---- 파일 ----
export type IssuesResult = { issues: CtCodeIssue[]; file?: UpdFile }
export type SaveError = { message?: string; status?: number; issues?: CtCodeIssue[] }

// 422 = steps.json 검증 오류(저장 안 됨) — 줄 목록을 그대로 돌려준다(throwCatch는 본문을 버려서 따로 받는다)
const withIssues = <T,>(p: Promise<{ data: T }>) => p.then(r => r.data).catch(e => {
    const res = e?.response
    if (res?.status === 422 && res.data?.issues) throw { message: '검증 오류가 있어 저장하지 않았습니다', status: 422, issues: res.data.issues } as SaveError
    return throwCatch(e) as never
})

export const readText = (r: Ref, path: string, scope: FileScope = 'main', sub = '') =>
    BaseAxios.get(`${bundlePath(r)}/file`, { params: { path, scope, sub }, responseType: 'text', transformResponse: x => x }).then(throwThen<string>).catch(throwCatch)
export const saveText = (r: Ref, path: string, content: string, create = false) =>
    withIssues<IssuesResult>(BaseAxios.put(`${bundlePath(r)}/text`, { path, content, create }))
export const checkSteps = (r: Ref, content: string) => BaseAxios.post(`${bundlePath(r)}/check`, { content }).then(throwThen<IssuesResult>).catch(throwCatch)
export const setMode = (r: Ref, path: string, mode: number) => BaseAxios.patch(`${bundlePath(r)}/mode`, { path, mode }).then(throwThen<{ mode: number }>).catch(throwCatch)
export const modeText = (m: number) => m.toString(8).padStart(3, '0')
export const deleteFile = (r: Ref, path: string) => BaseAxios.delete(`${bundlePath(r)}/file`, { params: { path } }).then(throwThen<Empty>).catch(throwCatch)

// 받기 주소(쿠키로 인가) — a[download]로 연다
export const fileUrl = (r: Ref, path: string) => `${BaseAxios.defaults.baseURL}${bundlePath(r)}/file?path=${enc(path)}&dl=1`
export const zipUrl = (r: Ref, merged: boolean) => `${BaseAxios.defaults.baseURL}${bundlePath(r)}/zip${merged ? '?merged=1' : ''}`
export const openDownload = (url: string) => {
    const a = document.createElement('a')
    a.href = url
    a.rel = 'noopener'
    document.body.appendChild(a)
    a.click()
    a.remove()
}

// ---- 조각 업로드(512KB — 앞단 nginx 1MB 아래, transfer.api와 같은 크기) ----
export const UPD_CHUNK = 512 * 1024
export type UploadStart = { id: string } | { exists: true }
export const startUpload = (r: Ref, path: string, size: number, overwrite: boolean) =>
    BaseAxios.post(`${bundlePath(r)}/uploads`, { path, size, overwrite }).then(res => res.data as { id: string }).catch(e => {
        if (e?.response?.status === 409 && e.response.data?.exists) return { exists: true as const }
        return throwCatch(e) as never
    })
type ChunkResult = { received: number; done: boolean; file?: UpdFile; issues?: CtCodeIssue[] }
const putChunk = (id: string, offset: number, data: Blob) =>
    withIssues<ChunkResult>(BaseAxios.put(`${base}/uploads/${enc(id)}`, data, { params: { offset }, headers: { 'Content-Type': 'application/octet-stream' }, timeout: 60_000 }))
export const cancelUpload = (id: string) => BaseAxios.delete(`${base}/uploads/${enc(id)}`).catch(() => undefined)

// 파일 하나 올리기. exists = 같은 이름이 있어 확인이 필요함(overwrite로 다시 부른다)
export const uploadFile = async (r: Ref, path: string, file: File, overwrite: boolean, onProgress?: (done: number) => void): Promise<'exists' | ChunkResult> => {
    const st = await startUpload(r, path, file.size, overwrite)
    if ('exists' in st) return 'exists'
    let off = 0
    try {
        for (;;) {
            const res = await putChunk(st.id, off, file.slice(off, off + UPD_CHUNK))
            off = res.received
            onProgress?.(off)
            if (res.done) return res
        }
    } catch (e) {
        void cancelUpload(st.id)
        throw e
    }
}

// ---- 다시 조회 ----
export const invalidateUpdates = (qc: QueryClient) => qc.invalidateQueries({ queryKey: ['UPDATES'] })

// ---- 조회 훅 ----
type P = MaybeRefOrGetter<ListParams>
export const useUpdDevices = (params: P) => useQuery({
    queryKey: computed(() => ['UPDATES', 'DEVICES', toValue(params)]),
    queryFn: () => listUpdDevices(toValue(params)),
    placeholderData: keepPreviousData,
})
export const useCommons = (params: P) => useQuery({
    queryKey: computed(() => ['UPDATES', 'COMMONS', toValue(params)]),
    queryFn: () => listCommons(toValue(params)),
    placeholderData: keepPreviousData,
})
// 인스턴스 접속 상태는 push가 없어 열려 있는 동안 10초마다 다시 조회(P2, 실시간 알림은 P3 Run·Log와 같이)
export const useUpdDevice = (mainKey: string) => useQuery({
    queryKey: ['UPDATES', 'DEVICE', mainKey],
    queryFn: () => getUpdDevice(mainKey),
    retry: (n, e: any) => e?.status !== 404 && n < 2,
    refetchInterval: 10_000,
})
export const useBundle = (r: Ref) => useQuery({
    queryKey: ['UPDATES', 'BUNDLE', r.kind, r.owner, r.version],
    queryFn: () => getBundle(r),
    retry: (n, e: any) => e?.status !== 404 && n < 2,
})
