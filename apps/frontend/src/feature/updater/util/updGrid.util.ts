import type { GridKeys } from '@/feature/member/util/memberGrid.util'
import type { BundleKind, Ref } from '../api/updater.api'

// updater 페이지 타일 키 — 회원 페이지와 같은 틀(MemberPage), 키 규칙만 다르다. REF-update-deploy.md "화면" · 단위-4(페이지 틀)
//   u<메인키>                   장비 타일
//   b<메인키>~<버전>            빌드 타일
//   f<메인키>~<버전>            Files 타일
//   s<메인키>~<버전>            steps.json 타일
//   c<이름>~<버전>              공용 묶음 타일(파일 목록 포함)
//   t<b|c><주인>~<버전>~<경로>  Script(텍스트 파일) 타일 — 경로는 퍼센트 인코딩
// `~`는 이름 글자([A-Za-z0-9._-])에 없고, URL 구분자 `*` `^` `,`는 경로 인코딩에서 빠진다

export type UpdTileKind = 'device' | 'build' | 'files' | 'steps' | 'common' | 'text'

const N = '[A-Za-z0-9._-]{1,64}'
const RE = new RegExp(`^(u${N}|[bfsc]${N}~${N}|t[bc]${N}~${N}~[A-Za-z0-9._%!'()-]{1,1500})$`)

const KIND: Record<string, UpdTileKind> = { u: 'device', b: 'build', f: 'files', s: 'steps', c: 'common', t: 'text' }
export const updKindOf = (key: string): UpdTileKind => KIND[key[0]!] ?? 'device'

export const UPD_KEYS: GridKeys = {
    re: RE,
    size: key => ({ w: 1, h: ['device', 'files', 'steps', 'common', 'text'].includes(updKindOf(key)) ? 2 : 1 }),
}

// `~`·`*`도 인코딩 — encodeURIComponent는 둘 다 남긴다
const encPath = (p: string) => encodeURIComponent(p).replace(/[~*]/g, c => `%${c.charCodeAt(0).toString(16).toUpperCase()}`)

export const devTileKey = (mainKey: string) => `u${mainKey}`
export const buildTileKey = (mainKey: string, version: string) => `b${mainKey}~${version}`
export const filesTileKey = (mainKey: string, version: string) => `f${mainKey}~${version}`
export const stepsTileKey = (mainKey: string, version: string) => `s${mainKey}~${version}`
export const commonTileKey = (name: string, version: string) => `c${name}~${version}`
export const textTileKey = (r: Ref, path: string) => `t${r.kind === 'build' ? 'b' : 'c'}${r.owner}~${r.version}~${encPath(path)}`

// 키 -> 대상. 장비 타일은 version 없음
export type UpdTarget = { kind: UpdTileKind; mainKey: string; ref?: Ref; path?: string }
export const updTargetOf = (key: string): UpdTarget => {
    const kind = updKindOf(key)
    if (kind === 'device') return { kind, mainKey: key.slice(1) }
    if (kind === 'text') {
        const bk: BundleKind = key[1] === 'c' ? 'common' : 'build'
        const [owner = '', version = '', p = ''] = key.slice(2).split('~')
        let path = p
        try { path = decodeURIComponent(p) } catch { /* 깨진 키 — 그대로 */ }
        return { kind, mainKey: bk === 'build' ? owner : '', ref: { kind: bk, owner, version }, path }
    }
    const [owner = '', version = ''] = key.slice(1).split('~')
    const bk: BundleKind = kind === 'common' ? 'common' : 'build'
    return { kind, mainKey: bk === 'build' ? owner : '', ref: { kind: bk, owner, version } }
}

export const sizeText = (n: number) =>
    n < 1024 ? `${n} B` : n < 1024 * 1024 ? `${(n / 1024).toFixed(1)} KB` : n < 1024 ** 3 ? `${(n / 1024 / 1024).toFixed(1)} MB` : `${(n / 1024 ** 3).toFixed(2)} GB`
export const shortSha = (sha: string) => sha ? sha.slice(0, 8) : '—'
export const timeText = (sec: number) => new Date(sec * 1000).toLocaleString('sv-SE', { dateStyle: 'short', timeStyle: 'short' })
