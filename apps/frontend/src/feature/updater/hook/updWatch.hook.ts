import { useTestSocket } from '@/common/websocket/websocket.hook'
import { useQueryClient } from '@tanstack/vue-query'
import { onUnmounted, ref, watch } from 'vue'
import { invalidateUpdates } from '../api/updater.api'

// updater 페이지 실시간 알림(REF-update-p3.md 단위-5(실시간 알림)) — 페이지가 열려 있는 동안 UPD:WATCH로 구독(관리자만 받음).
// UPD:CHANGE = 바뀜 신호(다시 조회), UPD:OUTPUT = 단계 출력 조각(Run·Log 타일이 runId로 받음)

type ChangeEvent = { mainKey: string; runId?: number }
type OutputEvent = { runId: number; data: string; off: number }

const BURST_MS = 250
const base64ToBytes = (b64: string) => Uint8Array.from(atob(b64), c => c.charCodeAt(0))
const outputs = new Map<number, Set<(b: Uint8Array, off: number) => void>>()
// 로그 전체가 다시 올라온 실행(재접속 뒤 재전송) — Run·Log 타일이 보고 있으면 다시 불러온다
export const updLogReload = ref<Record<number, number>>({})

// off = 이 조각까지 누적 바이트(로그 조회의 logEnd와 이음매 맞춤)
export const onUpdOutput = (runId: number, fn: (b: Uint8Array, off: number) => void) => {
    let set = outputs.get(runId)
    if (!set) outputs.set(runId, set = new Set())
    set.add(fn)
    return () => {
        set!.delete(fn)
        if (!set!.size) outputs.delete(runId)
    }
}

// 페이지(Updater.vue)에서 한 번
export const bindUpdWatch = () => {
    const qc = useQueryClient()
    const { emit, on, status } = useTestSocket()
    let timer: ReturnType<typeof setTimeout> | undefined
    const refetch = () => {
        clearTimeout(timer)
        timer = setTimeout(() => void invalidateUpdates(qc), BURST_MS)
    }
    on<ChangeEvent>('UPD:CHANGE', ev => {
        if (ev.runId) updLogReload.value = { ...updLogReload.value, [ev.runId]: (updLogReload.value[ev.runId] ?? 0) + 1 }
        refetch()
    })
    on<OutputEvent>('UPD:OUTPUT', ev => {
        const set = outputs.get(ev.runId)
        if (!set) return
        const bytes = base64ToBytes(ev.data)
        set.forEach(fn => fn(bytes, ev.off))
    })
    // 연결될 때마다 구독(서버는 탭 소켓이 끊기면 구독을 버림). 다시 연결이면 놓친 알림 대신 한 번 다시 조회
    let seen = false
    watch(status, s => {
        if (s !== 'CONNECTED') return
        emit('UPD:WATCH', { on: true })
        if (seen) refetch()
        seen = true
    }, { immediate: true })
    onUnmounted(() => {
        clearTimeout(timer)
        if (status.value === 'CONNECTED') emit('UPD:WATCH', { on: false })
    })
}
