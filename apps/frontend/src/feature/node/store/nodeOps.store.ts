import type { CatchErrorResponse } from '@/common/api/api.util'
import { useAppDialog } from '@/feature/layout/store/appDialog.store'
import type { ProcessType } from '@/feature/process/api/process.api'
import { useProcessTerm } from '@/feature/process/store/processTerm.store'
import { listWorkers, useWorkerRoster, type WorkerInstanceResponse, type WorkerResponse } from '@/feature/worker/api/worker.api'
import { computed, inject, provide, ref } from 'vue'
import { useNodeActions, type NodeKind, type NodeResponse } from '../api/node.api'
import { escHtml } from '../hook/nodePaste.hook'
import { useNodeRemove } from '../hook/nodeRemove.hook'
import { PERM_EDIT } from '../util/shareView.util'
import type { provideTileGrids } from './tileGrids.store'
import { useNodeClipboard } from './nodeClipboard.store'
import { useNodeSelection } from './nodeSelection.store'
import { useTileTree } from './tileTree.store'

// 노드 동작(실행·편집·이름·장비·삭제)과 그 다이얼로그 — 폴더 타일 행 메뉴와 선택 막대(타일 밖)가 같이 부른다.
// 다이얼로그는 NodeOpsHost가 한 번만 그린다. REF-node-select.md 선택-1(동작 분리)

const NODE_OPS_STORE_KEY = Symbol('NodeOpsStore')

export type NodeAct = 'open' | 'exec' | 'edit' | 'vi' | 'copy' | 'cut' | 'rename' | 'device' | 'remove' | 'select'
export type NodeActItem = { act: NodeAct; icon: string; title: string; disabled?: boolean; danger?: boolean }

// 행 메뉴·선택 막대(1개)가 같이 쓰는 목록 — SM6(막대 동작). 바꾸는 동작은 편집 단계 미만이면 비활성(서버도 막음)
export const nodeActItems = (node: NodeResponse, viDisabled: boolean): NodeActItem[] => {
    const ro = (node.permission ?? 0) < PERM_EDIT
    const script = node.kind === 'SCRIPT'
    const list: Array<NodeActItem | false> = [
        script && { act: 'edit', icon: 'mdi-file-document-edit-outline', title: '내용 편집' },
        script && { act: 'vi', icon: 'mdi-console-line', title: 'vi로 편집', disabled: viDisabled },
        { act: 'copy', icon: 'mdi-content-copy', title: '복사' },
        { act: 'cut', icon: 'mdi-content-cut', title: '잘라내기', disabled: ro },
        { act: 'rename', icon: 'mdi-pencil-outline', title: '이름 변경', disabled: ro },
        !script && { act: 'device', icon: 'mdi-lan-connect', title: '장비 지정', disabled: ro },
        { act: 'remove', icon: 'mdi-trash-can-outline', title: '삭제', disabled: ro, danger: true },
    ]
    return list.filter(v => !!v)
}

type NameTarget = { kind: NodeKind; parentId: number | null } | { node: NodeResponse }

export const provideNodeOps = (grids: ReturnType<typeof provideTileGrids>) => {
    const { tiles, hash: boardHash, openFolder, openDevice, applyServerTile, flush } = useTileTree()
    const { newSize, activate } = grids
    const { open: appDialogOpen, openDialog } = useAppDialog()
    const { exec: execTerm, execFile: execTermFile, adopt: adoptTerm } = useProcessTerm()
    const { devices } = useWorkerRoster()
    const { create, patch } = useNodeActions()
    const { removeNode } = useNodeRemove()
    const clipboard = useNodeClipboard()
    const sel = useNodeSelection()

    // 요청 중인 타일(진행 막대 위치). 막대에서 부른 동작은 그 노드를 고른 타일
    const busyTile = ref<string | null>(null)
    const saving = computed(() => busyTile.value !== null)

    const run = <T>(tileId: string, job: Promise<T>, done?: (res: T) => void) => {
        busyTile.value = tileId
        return job.then(res => done?.(res))
        .catch(err => openDialog({ type: 'error', content: err?.message || '요청에 실패했습니다' }))
        .finally(() => busyTile.value = null)
    }

    // 새 타일 크기 = 누른 화면의 최소 단위
    const open = (tileId: string, nodeId: number) => activate(openFolder(tileId, nodeId, newSize.value))

    // ---- 실행·vi 편집: 접속 인스턴스 0개 = 안내 / 1개 = 바로 / 여러 개 = 고르기 (REF-node-ui-terminal.md L, E1(진입 위치)) ----
    const typeLabel = (type: ProcessType) => type === 'EDIT' ? 'vi로 편집' : '실행'
    const pick = ref({
        open: false,
        tileId: '',
        node: undefined as NodeResponse | undefined,
        type: 'EXEC' as ProcessType,
        dir: false,                 // 장비 디렉토리 열기(폴더-3(여는 곳)) — 아니면 실행
        workers: [] as WorkerResponse[],
        offline: [] as WorkerInstanceResponse[],
    })

    // 터미널 타일은 서버가 exec 때 트리에 넣는다 — ⑪-4(타일 없는 실행 중 process). 저장 대기분 먼저(execFile과 같은 이유)
    const start = (tileId: string, node: NodeResponse, instanceKey: string, type: ProcessType) => run(tileId, flush().then(() => execTerm(node.id, instanceKey, { workspace: boardHash.value, parentTileId: tileId, size: newSize.value }, type)), res => {
        pick.value.open = false
        if (res.tile && res.tileVersion) activate(applyServerTile(res.tile, res.tileVersion))
    })

    const exec = (tileId: string, node: NodeResponse, type: ProcessType = 'EXEC') => run(tileId, listWorkers(node.id), workers => {
        if (!workers.length) {
            openDialog({ type: 'warning', title: `'${node.name}' ${typeLabel(type)}`, content: `이 스크립트를 ${typeLabel(type)}할 장비가 접속해 있지 않습니다.` })
            return
        }
        if (workers.length === 1) {
            start(tileId, node, workers[0]!.instanceKey, type)
            return
        }
        pick.value = { open: true, tileId, node, type, dir: false, workers, offline: offlineOf(workers) }
    })

    // 끊긴 것 = 명부(최대 10초 전) 중 방금 받은 접속 목록에 없는 것
    const offlineOf = (workers: WorkerResponse[]) => {
        const live = new Set(workers.map(w => w.instanceKey))
        return (devices.value.get(workers[0]!.mainKey)?.instances ?? []).filter(i => !live.has(i.instanceKey))
    }

    // ---- 장비 디렉토리 열기 — 폴더-3(여는 곳). 인스턴스 고르기는 실행과 같은 규칙, 연 폴더의 자식으로 ----
    const startDir = (tileId: string, node: NodeResponse, instanceKey: string) => {
        pick.value.open = false
        activate(openDevice(tileId, node.id, instanceKey, newSize.value))
    }

    const openDir = (tileId: string, node: NodeResponse) => run(tileId, listWorkers(node.id), workers => {
        if (!workers.length) {
            openDialog({ type: 'warning', title: `'${node.name}' 장비 디렉토리`, content: '이 폴더의 장비가 접속해 있지 않습니다.' })
            return
        }
        if (workers.length === 1) {
            startDir(tileId, node, workers[0]!.instanceKey)
            return
        }
        pick.value = { open: true, tileId, node, type: 'EXEC', dir: true, workers, offline: offlineOf(workers) }
    })

    // 장비 타일 행 "실행"(폴더-4(실행), Q23(누르는 방식)) — 인스턴스는 장비 타일 것. 터미널 타일은 장비 타일의 자식
    // 저장 대기분을 먼저 보낸다 — 막 열거나 옮긴 부모 타일이 서버 트리에 없으면 서버가 루트에 붙인다(2026-10-23 확인)
    const execFile = (tileId: string, folderId: number, instance: string, cmd: string, cwd: string) =>
        run(tileId, flush().then(() => execTermFile(instance, folderId, cmd, cwd, { workspace: boardHash.value, parentTileId: tileId, size: newSize.value })), res => {
            if (res.tile && res.tileVersion) activate(applyServerTile(res.tile, res.tileVersion))
        })

    // 미등록 process 가져오기(가져오기-4(화면)) — 서버가 장비 타일 아래 터미널 타일을 넣는다. 실패는 호출부(미등록 창)가 보여 준다
    const adopt = (tileId: string, folderId: number, instance: string, uid: string) =>
        flush().then(() => adoptTerm(instance, folderId, uid, { workspace: boardHash.value, parentTileId: tileId, size: newSize.value }))
        .then(res => {
            if (res.tile && res.tileVersion) activate(applyServerTile(res.tile, res.tileVersion))
        })

    // ---- 이름(새로 만들기 / 이름 변경) — 닫히는 동안 제목이 비지 않도록 열림 여부와 대상을 따로 둔다 ----
    const name = ref({ open: false, tileId: '', target: { kind: 'FOLDER', parentId: null } as NameTarget })
    const nameTitle = computed(() => {
        const t = name.value.target
        if ('node' in t) return '이름 변경'
        return t.kind === 'FOLDER' ? '새 폴더' : '새 스크립트'
    })
    const nameInitial = computed(() => 'node' in name.value.target ? name.value.target.node.name : '')

    const editName = (tileId: string, target: NameTarget) => name.value = { open: true, tileId, target }
    const submitName = (value: string) => {
        const { tileId, target: t } = name.value
        if ('node' in t) {
            run(tileId, patch(t.node.id, { name: value }), () => name.value.open = false)
            return
        }
        run(tileId, create({ kind: t.kind, name: value, parentId: t.parentId }), node => {
            name.value.open = false
            if (node.kind === 'SCRIPT') edit(node)   // 빈 스크립트는 바로 내용 편집으로 — 결정 I(새 스크립트 직후 동작)
        })
    }

    const editState = ref({ open: false, node: undefined as NodeResponse | undefined })
    const edit = (node: NodeResponse) => editState.value = { open: true, node }

    const device = ref({ open: false, tileId: '', node: undefined as NodeResponse | undefined })
    const editDevice = (tileId: string, node: NodeResponse) => device.value = { open: true, tileId, node }
    const submitDevice = (deviceKey: string | null) => {
        const { tileId, node } = device.value
        if (node) run(tileId, patch(node.id, { deviceKey }), () => device.value.open = false)
    }

    // ---- 삭제: 1개 = 기존 확인 창 / 여러 개 = 확인 창 + 차례로 요청 + 부분 실패 알림(SM6(막대 동작)) ----
    // 앞서 지운 폴더 안에 있던 것은 404 — 이미 지워졌으므로 실패로 치지 않는다
    const removeMany = async (tileId: string, nodes: NodeResponse[]) => {
        const fails: string[] = []
        busyTile.value = tileId
        for (const n of nodes) {
            try {
                await removeNode(n)
            } catch (err) {
                const e = err as CatchErrorResponse
                if (e?.status !== 404) fails.push(`${escHtml(n.name)}: ${escHtml(e?.message || '요청에 실패했습니다')}`)
            }
        }
        busyTile.value = null
        if (fails.length) openDialog({ type: 'warning', title: `${nodes.length}개 중 ${fails.length}개 실패`, content: fails })
    }

    // 이름은 title로만 넘긴다(1개) / 여러 개는 escape — AppDialog의 content는 v-html로 그려진다
    const remove = (tileId: string, nodes: NodeResponse[]) => {
        const one = nodes.length === 1 ? nodes[0]! : null
        const folders = nodes.some(n => n.kind === 'FOLDER')
        openDialog({
            type: 'warning',
            title: one ? `'${one.name}' 삭제` : `${nodes.length}개 삭제`,
            content: [
                ...(one ? [] : [nodes.map(n => escHtml(n.name)).join(', ')]),
                ...(folders
                    ? ['폴더와 그 안의 모든 항목이 삭제됩니다.', '이 폴더나 하위 폴더를 보고 있는 타일은 닫힙니다.']
                    : ['스크립트가 삭제됩니다.']),
            ],
            buttons: [
                { view: '취소', type: 'info', bindButton: 'Escape', action: () => appDialogOpen.value = false },
                {
                    // Enter를 묶지 않는다 — 묶지 않은 Enter·Space는 AppDialog가 닫기(취소)로 처리
                    view: '삭제', type: 'error',
                    action: () => {
                        appDialogOpen.value = false
                        if (one) run(tileId, removeNode(one))
                        else removeMany(tileId, nodes)
                    },
                },
            ],
        })
    }

    // 동작 하나를 노드에 — 1개만 받는 동작은 첫 노드(막대는 1개일 때만 그 동작을 보인다, SM6(막대 동작))
    const act = (a: NodeAct, tileId: string, nodes: NodeResponse[]) => {
        const n = nodes[0]
        if (!n) return
        switch (a) {
            case 'open': return open(tileId, n.id)
            case 'exec': return exec(tileId, n)
            case 'vi': return exec(tileId, n, 'EDIT')
            case 'edit': return edit(n)
            case 'copy': return clipboard.copy(nodes)
            case 'cut': return clipboard.cut(nodes)
            case 'rename': return editName(tileId, { node: n })
            case 'device': return editDevice(tileId, n)
            case 'remove': return remove(tileId, nodes)
            case 'select': {
                const t = tiles.value[tileId]
                return sel.toggle(n, tileId, t?.type === 'folder' ? t.nodeId : null)
            }
        }
    }

    const ctx = {
        act,
        busyTile,
        saving,
        run,
        open,
        exec,
        typeLabel,
        pick,
        start,
        startDir,
        openDir,
        execFile,
        adopt,
        name,
        nameTitle,
        nameInitial,
        editName,
        submitName,
        editState,
        edit,
        device,
        editDevice,
        submitDevice,
        remove,
    }

    provide(NODE_OPS_STORE_KEY, ctx)

    return ctx
}

export const useNodeOps = () => {
    const context = inject<ReturnType<typeof provideNodeOps>>(NODE_OPS_STORE_KEY)
    if (!context) throw new Error('NodeOpsStore is not provided')
    return context
}
