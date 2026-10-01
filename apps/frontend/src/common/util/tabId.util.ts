import { readonly, ref, watch } from 'vue'
import type { SocketContext } from '@/common/websocket/websocket.hook'

// 탭 id: 서버가 소켓 연결 직후 TAB:ID로 발급한다 — REF-process-sync.md "탭 id 수명"
// sessionStorage = 탭마다 따로 + 새로고침해도 유지. localStorage는 모든 탭이 공유하므로 쓰면 안 된다

const KEY = 'nexus.tabId'
export const TAB_ID_HEADER = 'X-Tab-Id'

let memory = ''             // 저장소를 못 쓰는 환경(시크릿 창 등) 대비
const ready_ = ref(false)   // 이번 소켓 연결에서 TAB:ID를 받았는지 — 받기 전엔 실행 버튼 비활성

export const tabReady = readonly(ready_)

export const readTabId = (): string => {
    if (memory) return memory
    try {
        return sessionStorage.getItem(KEY) ?? ''
    } catch {
        return ''
    }
}

const writeTabId = (id: string) => {
    memory = id
    try {
        sessionStorage.setItem(KEY, id)
    } catch { /* 저장 불가 */ }
}

// 소켓을 provide하는 곳(App.vue)에서 한 번 호출
export const bindTabId = (socket: SocketContext) => {
    socket.on<{ tabId: string }>('TAB:ID', ev => {
        writeTabId(ev.tabId)
        ready_.value = true
    })
    watch(socket.status, s => {
        if (s !== 'CONNECTED') ready_.value = false
    })
}
