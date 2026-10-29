import { useAppDialog } from '@/feature/layout/store/appDialog.store'
import type { CatchErrorResponse } from '@/common/api/api.util'
import { useQueryClient } from '@tanstack/vue-query'
import { copyNode, patchNode, useNodeQueryClient, type NodeResponse } from '../api/node.api'
import type { ClipMode, ClipNode } from '../store/nodeClipboard.store'
import { PERM_EDIT, PERM_OWNER } from '../util/shareView.util'

// 붙여넣기(클립보드) · 드롭(드래그)이 같이 쓰는 판정·요청 — REF-node-select.md 선택-5(붙여넣기 공용화)

export const escHtml = (v: string) => v.replace(/[&<>"']/g, c => `&#${c.charCodeAt(0)};`)   // AppDialog content는 v-html

// 목적지: id(null = 루트), node(루트면 없음), path = 루트부터 목적지까지
export type PasteDest = { id: number | null; node?: NodeResponse; path: Array<{ id: number }> }

// CP3(붙여넣기 제약): 막는 이유(없으면 '') — 서버 규칙(checkMove·copyNode)을 미리 비춰 볼 뿐, 최종 판정은 서버
export const pasteBlock = (mode: ClipMode, nodes: ReadonlyArray<ClipNode>, to: PasteDest) => {
    const dest = to.node
    if (to.id !== null && !dest) return '폴더 정보를 불러오는 중입니다'
    if (dest && (dest.permission ?? 0) < PERM_EDIT) return '이 폴더에 넣을 권한이 없습니다'
    if (mode === 'copy') return ''
    if (nodes.every(n => n.parentId === to.id)) return '이미 이 폴더에 있습니다'
    if (nodes.some(n => to.path.some(p => p.id === n.id))) return '잘라낸 폴더를 자기 자신이나 그 아래로 옮길 수 없습니다'
    if (!dest) return nodes.some(n => n.permission < PERM_OWNER) ? '내 노드만 기본으로 옮길 수 있습니다' : ''
    return nodes.some(n => n.ownerUserId !== dest.ownerUserId) ? '소유자가 다른 폴더로는 옮길 수 없습니다(복사는 가능)' : ''
}

// 노드마다 차례로 요청(일괄 API 없음, 담은 순서대로 끝에 붙음). 성공한 건 그대로 두고 실패만 알림 — CP-부분 실패. 반환 = 성공한 id
export const useNodePaste = () => {
    const { invalidateAllForce } = useNodeQueryClient(useQueryClient())
    const { openDialog } = useAppDialog()

    return async (mode: ClipMode, nodes: ReadonlyArray<ClipNode>, destId: number | null) => {
        const done: number[] = []
        const fails: string[] = []
        for (const n of nodes) {
            try {
                if (mode === 'copy') await copyNode(n.id, destId)
                else await patchNode(n.id, { parentId: destId })
                done.push(n.id)
            } catch (err) {
                fails.push(`${escHtml(n.name)}: ${escHtml((err as CatchErrorResponse)?.message || '요청에 실패했습니다')}`)
            }
        }
        invalidateAllForce()
        if (fails.length) openDialog({ type: 'warning', title: `${nodes.length}개 중 ${fails.length}개 실패`, content: fails })
        return done
    }
}
