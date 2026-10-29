import { BaseAxios, throwCatch, throwThen } from '@/common/api/api.util'
import { readTabId, TAB_ID_HEADER } from '@/common/util/tabId.util'
import type { TileSize } from '@/feature/widget/util/tile.type'
import type { Tile, ViewerTile } from '../store/tileTree.store'

// 백엔드: apps/core/cmd/supervisor/router/tile.go (group "/tiles") — ⑪(타일 트리 서버 저장), REF-node-ui-save.md.
// 보드(작업 공간)마다 트리 하나: hash 빈 값 = 기본 보드(`/tiles`), 그 외 `/tiles/:hash` — WS1(트리 주소)

export type TileTree = {
    rootId: string
    order: string[]     // 화면 순서, 루트가 맨 앞 — REF-tile-drag.md
    tiles: Record<string, Tile>
}

export type TileTreeResponse = {
    hash: string        // 그 트리의 보드 hash(기본 보드도 실제 hash) — TILES:UPDATE를 자기 보드 것만 받는 데 씀
    tree: TileTree
    version: number
    pins?: TilePin[]    // GET만 — 개편-1(고정 지도)
}

// 같은 폴더 노드의 타일이 내 보드를 통틀어 2개 이상일 때 그 노드의 고정 타일 — R4(고정 표시 남발). GET /tiles·TILES:PINS
export type TilePin = {
    nodeId: number
    hash: string
    tileId: string
}

// 409(버전 충돌)는 에러가 아니라 결과 — 서버 트리 위에 자기 동작을 다시 적용해야 하므로 ⑪-2(충돌 처리)
export type PutTilesResult =
    | { ok: true; version: number }
    | ({ ok: false } & TileTreeResponse)

// keepalive 요청 본문 합계 상한(브라우저 64KB). 넘으면 일반 요청으로 보낸다
const KEEPALIVE_MAX = 64 * 1024

export const tilesPath = (hash: string) => hash ? `/tiles/${hash}` : '/tiles'

// GET /tiles[/:hash] — 보드의 타일 트리(기본 보드가 없으면 서버가 루트만 있는 트리를 만듦). 서버가 고아 안전망을 먼저 돈다(WS3(고아 안전망 위치))
export const getTiles = (hash: string) => BaseAxios.get(
    tilesPath(hash)
).then(throwThen<TileTreeResponse>)
.catch(throwCatch)

// keepalive = 페이지가 사라져도 전송을 끝낸다. axios(XHR)는 지원하지 않아 fetch를 직접 쓴다
const putTilesKeepalive = (hash: string, body: string): Promise<PutTilesResult> => fetch(
    `${BaseAxios.defaults.baseURL}${tilesPath(hash)}`,
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

// POST /tiles/bind — "여기로 타일 고정"(같은 노드의 다른 고정은 서버가 풂, 해제 없음 — 개편-2(여기로 고정)). 결과는 TILES:UPDATE·TILES:PINS로 옴
export const bindTile = (hash: string, tileId: string) => BaseAxios.post(
    '/tiles/bind',
    { hash, tileId },
).then(throwThen<Record<string, never>>)
.catch(throwCatch)

// PUT /tiles[/:hash] — 읽은 version일 때만 저장. 200 {version} / 409 {hash, tree, version}.
// 저장본에 없던 터미널 타일은 넣을 수 없다(400) — 터미널 타일은 서버만 넣음(WS5(터미널 타일 추가 주체))
export const putTiles = (hash: string, tree: TileTree, version: number, keepalive = false): Promise<PutTilesResult> => {
    const body = JSON.stringify({ tree, version })
    if (keepalive && new Blob([body]).size <= KEEPALIVE_MAX) return putTilesKeepalive(hash, body)
    return BaseAxios.put(tilesPath(hash), body, { validateStatus: s => s === 200 || s === 409 })
        .then(res => res.status === 409
            ? { ok: false as const, ...(res.data as TileTreeResponse) }
            : { ok: true as const, version: (res.data as { version: number }).version })
        .catch(throwCatch)
}

// GET /tiles/viewers — 터미널 탭의 보기 줄. 보기 타일은 process가 아니라 내 보드 트리에서 찾음(REF-view-tile.md)
// 범위는 터미널 탭과 같음: folderId 0 = 전부 / 폴더 = 연 폴더가 그 하위 / mainKey = 그 장비
export type ViewerResponse = { hash: string; tileId: string; nodeId: number; instance: string; path: string }

export const listViewers = (folderId: number, mainKey?: string) => BaseAxios.get(
    '/tiles/viewers',
    { params: { folderId, mainKey: mainKey || undefined } }
).then(throwThen<ViewerResponse[]>)
.catch(throwCatch)

// POST /tiles/viewers/video — 영상은 같은 파일이 계정 전체에서 타일 하나라 서버가 넣는다(MEMORY "타일 추가 주체", videoTile.go)
// tile 있음 = 새로 넣음(hash 보드) / 없음 = 이미 있던 타일(hash 보드의 tileId)로 이동
export type VideoTileRequest = { folderId: number; instance: string; path: string; workspace: string; parentTileId: string; size: TileSize }
export type VideoTileResponse = { hash: string; tileId: string; tile?: ViewerTile; tileVersion?: number }

export const openVideoTile = (body: VideoTileRequest) => BaseAxios.post(
    '/tiles/viewers/video',
    body,
).then(throwThen<VideoTileResponse>)
.catch(throwCatch)

// S1(이어보기): 영상 재생 위치 — 키 = 파일, 계정마다. 그 파일의 보기 타일이 없으면 서버가 저장하지 않음(닫힌 뒤 늦게 온 저장)
const VIDEO_POS = '/tiles/viewers/video/position'

export const getVideoPosition = (instance: string, path: string) => BaseAxios.get(
    VIDEO_POS,
    { params: { instance, path } }
).then(throwThen<{ pos: number | null }>)
.catch(throwCatch)

// keepalive = 창을 닫는 중에도 전송을 끝낸다(axios는 지원하지 않아 fetch)
export const saveVideoPosition = (instance: string, path: string, pos: number, keepalive = false): Promise<unknown> => {
    const body = { instance, path, pos }
    if (!keepalive) return BaseAxios.put(VIDEO_POS, body).catch(throwCatch)
    return fetch(`${BaseAxios.defaults.baseURL}${VIDEO_POS}`, {
        method: 'PUT',
        keepalive: true,
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
    })
}
