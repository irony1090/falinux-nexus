import { readUserPref, userPrefKey, writeUserPref } from '@/common/util/userPrefs.util'
import { authState } from '@/feature/user/store/auth.store'
import { computed, readonly, ref, watch } from 'vue'
import type { NodeKind, NodeResponse } from '../api/node.api'

// 노드 클립보드(복사·잘라내기) — 모든 폴더 타일이 같은 값을 보므로 모듈 하나에 둔다. 결정 → REF-node-clipboard.md
// 저장 = userPrefs(같은 브라우저 탭끼리 공유, 로그아웃 때 전부 삭제 — CP-계정 전환). name·kind는 표시용, 붙여넣기는 id로 서버가 판정

export type ClipMode = 'copy' | 'cut'
export type ClipNode = { id: number; parentId?: number | null; name: string; kind: NodeKind; ownerUserId: number; permission: number }   // parentId = 같은 폴더로 옮기기 판정(옛 저장값엔 없음)
type Clip = { mode: ClipMode; nodes: ClipNode[] }

const KEY = 'nodeClipboard'
const empty = (): Clip => ({ mode: 'copy', nodes: [] })

// 저장값 모양이 다르면(옛 형식·손상) 비운 것으로
const read = (): Clip => {
    const v = readUserPref<Clip | null>(KEY, null)
    return v && (v.mode === 'copy' || v.mode === 'cut') && Array.isArray(v.nodes) ? v : empty()
}

const clip = ref<Clip>(read())

const save = (next: Clip) => {
    clip.value = next
    writeUserPref(KEY, next)
}

// storage 이벤트는 다른 탭의 변경만 온다(key null = 전체 삭제)
window.addEventListener('storage', ev => {
    if (ev.key === null || ev.key === userPrefKey(KEY)) clip.value = read()
})
// 이 탭의 로그아웃은 storage 이벤트가 없어서 로그인 상태로 다시 읽는다
watch(authState, () => clip.value = read())

export const toClip = (n: NodeResponse): ClipNode => ({ id: n.id, parentId: n.parentId, name: n.name, kind: n.kind, ownerUserId: n.ownerUserId, permission: n.permission ?? 0 })

// SM3(클립보드 담기): 고른 것으로 바꾼다(옛 CP5 덧붙임 대체)
const put = (mode: ClipMode, nodes: NodeResponse[]) => save({ mode, nodes: nodes.map(toClip) })

// 붙여넣기에 성공한 것을 뺀다 — 잘라내기만(CP4(붙여넣은 뒤): 복사는 유지)
const removeIds = (ids: number[]) => {
    if (!ids.length) return
    const drop = new Set(ids)
    save({ ...clip.value, nodes: clip.value.nodes.filter(n => !drop.has(n.id)) })
}

const state = readonly(clip)
const cutIds = computed(() => new Set(clip.value.mode === 'cut' ? clip.value.nodes.map(n => n.id) : []))

export const useNodeClipboard = () => ({
    clip: state,
    cutIds,
    copy: (nodes: NodeResponse[]) => put('copy', nodes),
    cut: (nodes: NodeResponse[]) => put('cut', nodes),
    removeIds,
    clear: () => save(empty()),
})
