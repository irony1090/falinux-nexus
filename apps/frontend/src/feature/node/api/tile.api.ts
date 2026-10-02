import { BaseAxios, throwCatch, throwThen } from '@/common/api/api.util'
import { readTabId, TAB_ID_HEADER } from '@/common/util/tabId.util'
import type { Tile } from '../store/tileTree.store'

// 백엔드: apps/core/cmd/supervisor/router/tile.go (group "/tiles") — ⑪(타일 트리 서버 저장), REF-node-ui-save.md

export type TileTree = {
    rootId: string
    tiles: Record<string, Tile>
}

export type TileTreeResponse = {
    tree: TileTree
    version: number
}

// 409(버전 충돌)는 에러가 아니라 결과 — 서버 트리 위에 자기 동작을 다시 적용해야 하므로 ⑪-2(충돌 처리)
export type PutTilesResult =
    | { ok: true; version: number }
    | { ok: false; tree: TileTree; version: number }

// keepalive 요청 본문 합계 상한(브라우저 64KB). 넘으면 일반 요청으로 보낸다
const KEEPALIVE_MAX = 64 * 1024

// GET /tiles — 계정의 타일 트리(없으면 서버가 루트만 있는 트리를 만들어 돌려줌)
export const getTiles = () => BaseAxios.get(
    '/tiles'
).then(throwThen<TileTreeResponse>)
.catch(throwCatch)

// keepalive = 페이지가 사라져도 전송을 끝낸다. axios(XHR)는 지원하지 않아 fetch를 직접 쓴다
const putTilesKeepalive = (body: string): Promise<PutTilesResult> => fetch(
    new URL('/tiles', BaseAxios.defaults.baseURL),
    {
        method: 'PUT',
        keepalive: true,
        credentials: 'include',
        headers: { 'Content-Type': 'application/json', [TAB_ID_HEADER]: readTabId() },   // push에 실려 돌아올 때 이 탭이 자기 것으로 거르게(4②)
        body,
    }
).then(async res => {
    if (res.status === 409) return { ok: false, ...await res.json() }
    if (!res.ok) throw { message: `타일 저장 실패 (${res.status})` }
    return { ok: true, version: (await res.json()).version }
})

// PUT /tiles — 읽은 version일 때만 저장. 200 {version} / 409 {tree, version}
export const putTiles = (tree: TileTree, version: number, keepalive = false): Promise<PutTilesResult> => {
    const body = JSON.stringify({ tree, version })
    if (keepalive && new Blob([body]).size <= KEEPALIVE_MAX) return putTilesKeepalive(body)
    return BaseAxios.put('/tiles', body, { validateStatus: s => s === 200 || s === 409 })
        .then(res => res.status === 409
            ? { ok: false as const, ...(res.data as TileTreeResponse) }
            : { ok: true as const, version: (res.data as { version: number }).version })
        .catch(throwCatch)
}
