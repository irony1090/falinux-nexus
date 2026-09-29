import type { NodeKind } from '@/feature/node/api/node.api'
import { reactive } from 'vue'
import type { provideTileTree } from '../store/tileTree.store'

// 타일 UI 미리보기용 더미 — 실제 node API·process 연동 전까지만 쓴다

export type DummyNode = {
    id: number
    parentId: number | null
    kind: NodeKind
    name: string
    lines?: Array<string>   // SCRIPT 실행 시 보여줄 출력
}

const NODES: Array<DummyNode> = [
    { id: 1, parentId: null, kind: 'FOLDER', name: 'Rack A' },
    { id: 2, parentId: null, kind: 'FOLDER', name: '사무실 2F' },
    { id: 3, parentId: null, kind: 'SCRIPT', name: 'supervisor-log', lines: ['supervisor :5050 - registry live', 'INFO registry heartbeat irony-rack-a#UhQ2 ok'] },
    { id: 4, parentId: 1, kind: 'SCRIPT', name: 'htop', lines: ['  CPU[|||||||||      23.4%]   Tasks: 61', '  Mem[||||||||||| 1.21G/3.78G]', '', '  PID USER   CPU% Command', ' 1184 irony  11.2 nexus-worker'] },
    { id: 5, parentId: 1, kind: 'FOLDER', name: 'Shelf 1' },
    { id: 6, parentId: 1, kind: 'SCRIPT', name: 'fw-flash', lines: ['esptool.py v4.7.0', 'Connecting....', 'Writing at 0x00010000... (9%)'] },
    { id: 7, parentId: 5, kind: 'SCRIPT', name: 'serial-console', lines: ['--- /dev/ttyUSB0 115200 8N1 ---', 'I (1024) app: jig ready'] },
    { id: 8, parentId: 2, kind: 'SCRIPT', name: 'build-watch', lines: ['$ npm run dev -- --host', 'VITE ready in 412 ms'] },
    { id: 9, parentId: 2, kind: 'SCRIPT', name: 'log-tail', lines: ['$ journalctl -fu nexus-worker'] },
]

const byId = new Map(NODES.map(n => [n.id, n]))

export const dummyNode = (id: number | null) => id === null ? undefined : byId.get(id)
export const dummyChildren = (parentId: number | null) => NODES.filter(n => n.parentId === parentId)

// 루트부터 id까지 경로 (브레드크럼용)
export const dummyPath = (id: number | null): Array<DummyNode> => {
    const node = dummyNode(id)
    return node ? [...dummyPath(node.parentId), node] : []
}

type DummyStatus = 'RUNNING' | 'DONE'
const status = reactive(new Map<string, DummyStatus>())

export const dummyExec = () => {
    const uid = 'p-' + Math.random().toString(36).slice(2, 6)
    status.set(uid, 'RUNNING')
    return uid
}
export const dummyKill = (uid: string) => { status.set(uid, 'DONE') }
export const dummyStatus = (uid: string): DummyStatus => status.get(uid) ?? 'DONE'

// 시안의 "순서 예시": root가 Rack A 열기 -> supervisor-log 2번 실행 -> 사무실 2F 열기, Rack A가 htop 실행
export const seedTileTree = (tree: ReturnType<typeof provideTileTree>) => {
    const size = { w: 0.5, h: 0.5 } as const
    const root = tree.rootId
    const rackA = tree.openFolder(root, 1, size)
    tree.addTerminal(root, 3, dummyExec(), size)
    tree.addTerminal(root, 3, dummyExec(), size)
    tree.openFolder(root, 2, size)
    tree.addTerminal(rackA, 4, dummyExec(), size)
}
