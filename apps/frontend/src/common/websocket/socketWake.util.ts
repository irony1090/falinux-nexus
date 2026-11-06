import type { SocketContext } from './websocket.hook'

// 재연결-2(돌아오면 바로): 화면 복귀·bfcache 복원·네트워크 복구 때 소켓을 깨운다 — App.vue에서 한 번(App 수명이라 해제 안 함), REF-socket-reconnect.md
export const bindSocketWake = (socket: SocketContext) => {
    document.addEventListener('visibilitychange', () => {
        if (document.visibilityState === 'visible') socket.wake()
    })
    window.addEventListener('pageshow', () => socket.wake())
    window.addEventListener('online', () => socket.wake())
}
