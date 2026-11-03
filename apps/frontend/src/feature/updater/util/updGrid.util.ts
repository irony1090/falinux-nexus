import type { GridKeys } from '@/feature/member/util/memberGrid.util'
import type { BundleKind, Ref } from '../api/updater.api'

// updater 페이지 타일 키 — 회원 페이지와 같은 틀(MemberPage), 키 규칙만 다르다. REF-update-deploy.md "화면" · 단위-4(페이지 틀)
//   u<메인키>                   장비 타일
//   b<메인키>~<버전>            빌드 타일
//   f<메인키>~<버전>            Files 타일
//   s<메인키>~<버전>            steps.json 타일
//   c<이름>~<버전>              공용 묶음 타일(파일 목록 포함)
//   t<b|c><주인>~<버전>~<경로>  Script(텍스트 파일) 타일 — 경로는 퍼센트 인코딩
//   r<배포 id>                  Run · Log 타일(배포 하나 × 인스턴스 하나, REF-update-p3.md "합의한 해석")
// `~`는 이름 글자([A-Za-z0-9._-])에 없고, URL 구분자 `*` `^` `,`는 경로 인코딩에서 빠진다

export type UpdTileKind = 'device' | 'build' | 'files' | 'steps' | 'common' | 'text' | 'run'

const N = '[A-Za-z0-9._-]{1,64}'
const RE = new RegExp(`^(u${N}|[bfsc]${N}~${N}|t[bc]${N}~${N}~[A-Za-z0-9._%!'()-]{1,1500}|r[1-9][0-9]{0,17})$`)

const KIND: Record<string, UpdTileKind> = { u: 'device', b: 'build', f: 'files', s: 'steps', c: 'common', t: 'text', r: 'run' }
export const updKindOf = (key: string): UpdTileKind => KIND[key[0]!] ?? 'device'

export const UPD_KEYS: GridKeys = {
    re: RE,
    // 처음 높이: 로그·편집기(Run·Log, Script)만 2칸, 나머지 1칸(2026-11-03, REF-member-ui.md 화-10)
    size: key => ({ w: 1, h: ['text', 'run'].includes(updKindOf(key)) ? 2 : 1 }),
}

// `~`·`*`도 인코딩 — encodeURIComponent는 둘 다 남긴다
const encPath = (p: string) => encodeURIComponent(p).replace(/[~*]/g, c => `%${c.charCodeAt(0).toString(16).toUpperCase()}`)

export const devTileKey = (mainKey: string) => `u${mainKey}`
export const buildTileKey = (mainKey: string, version: string) => `b${mainKey}~${version}`
export const filesTileKey = (mainKey: string, version: string) => `f${mainKey}~${version}`
export const stepsTileKey = (mainKey: string, version: string) => `s${mainKey}~${version}`
export const commonTileKey = (name: string, version: string) => `c${name}~${version}`
export const runTileKey = (deployId: number) => `r${deployId}`
export const textTileKey = (r: Ref, path: string) => `t${r.kind === 'build' ? 'b' : 'c'}${r.owner}~${r.version}~${encPath(path)}`

// 키 -> 대상. 장비 타일은 version 없음, Run · Log는 배포 id만(장비·빌드는 조회해서 안다)
export type UpdTarget = { kind: UpdTileKind; mainKey: string; ref?: Ref; path?: string; deployId?: number }
export const updTargetOf = (key: string): UpdTarget => {
    const kind = updKindOf(key)
    if (kind === 'device') return { kind, mainKey: key.slice(1) }
    if (kind === 'run') return { kind, mainKey: '', deployId: Number(key.slice(1)) }
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
export const durText = (sec: number) => sec < 60 ? `${sec}s` : sec < 3600 ? `${Math.floor(sec / 60)}:${String(sec % 60).padStart(2, '0')}` : `${Math.floor(sec / 3600)}h ${Math.floor(sec % 3600 / 60)}m`
