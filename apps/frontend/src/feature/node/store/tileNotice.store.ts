import { ref } from 'vue'

// 타일 위치 알림 — 개편-8(알림 UI), REF-tile-binding-rework.md "위치 알림". 탭 메모리(새로고침·재접속 시 사라짐, N1(표시 수단)).
// 내역 = 계정의 이벤트 전부(N6(내역 범위)) / 토스트 = 호출부가 고른 것만(N4(받는 탭))

export type TileSpot = { hash: string; tileId: string }

// 서버 TILES:NOTICE (protocol.TilesNoticeEvent)
export type TileNoticeEvent = {
    kind: 'BIND' | 'ATTACH' | 'NO_PLACE' | 'PIN_MOVED'
    uid?: string
    nodeId: number
    name: string
    from?: TileSpot
    to?: TileSpot
    at: number
}

// go = 누르면 갈 곳: 그 보드의 타일, 또는 지금 보드의 터미널 탭(M4(위치 없음))
export type NoticeItem = TileNoticeEvent & { id: number; text: string; go: { hash: string; tileId?: string } | { tab: 'terminals' } }

export type NoticeToast = { id: number; items: NoticeItem[]; created: number }

const HISTORY_MAX = 100
const TOAST_MS = 5000
const GROUP_MS = 800        // 이 안에 온 알림은 토스트 하나로 묶음 — N5(알림 폭주)
const TOAST_MAX = 3

const history = ref<NoticeItem[]>([])     // 새것 먼저
const unread = ref(0)
const toasts = ref<NoticeToast[]>([])
const drawerOpen = ref(false)      // 내역 드로어 — 종은 푸터, 드로어는 NoticeCenter
let seq = 0
const timers = new Map<number, ReturnType<typeof setTimeout>>()

const dismiss = (toastId: number) => {
    clearTimeout(timers.get(toastId))
    timers.delete(toastId)
    toasts.value = toasts.value.filter(t => t.id !== toastId)
}

const arm = (toastId: number) => {
    clearTimeout(timers.get(toastId))
    timers.set(toastId, setTimeout(() => dismiss(toastId), TOAST_MS))
}

const push = (item: Omit<NoticeItem, 'id'>, toast: boolean) => {
    const it = { ...item, id: ++seq }
    history.value = [it, ...history.value].slice(0, HISTORY_MAX)
    unread.value++
    if (!toast) return
    const last = toasts.value[toasts.value.length - 1]
    if (last && Date.now() - last.created < GROUP_MS) {
        last.items = [...last.items, it]
        arm(last.id)
        return
    }
    const t = { id: it.id, items: [it], created: Date.now() }
    toasts.value = [...toasts.value, t].slice(-TOAST_MAX)
    arm(t.id)
}

const markRead = () => { unread.value = 0 }

const clear = () => {
    timers.forEach(t => clearTimeout(t))
    timers.clear()
    history.value = []
    toasts.value = []
    unread.value = 0
}

export const useTileNotices = () => ({ history, unread, toasts, drawerOpen, push, dismiss, markRead, clear })
