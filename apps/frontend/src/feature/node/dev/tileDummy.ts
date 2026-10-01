import { reactive } from 'vue'

// process 연동 전까지 실행 상태만 흉내낸다 (노드 목록·이름은 실제 node API)

type DummyStatus = 'RUNNING' | 'DONE'
const status = reactive(new Map<string, DummyStatus>())

export const dummyExec = () => {
    const uid = 'p-' + Math.random().toString(36).slice(2, 6)
    status.set(uid, 'RUNNING')
    return uid
}
export const dummyKill = (uid: string) => { status.set(uid, 'DONE') }
export const dummyStatus = (uid: string): DummyStatus => status.get(uid) ?? 'DONE'
