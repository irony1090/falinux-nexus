import { useAuthStore } from '@/feature/user/store/auth.store'
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAppDialog } from '../store/appDialog.store'

// 옛 헤더(AppHead)의 앱 버튼 동작 — 지금은 보드 푸터가 쓴다. REF-board-chrome.md 푸터-2(요소 옮기기)
export const useAppActions = () => {
    const router = useRouter()
    const { auth, logout } = useAuthStore()
    const { open: dialogOpen, openDialog } = useAppDialog()

    // 로그아웃은 확인 뒤 — 이동은 라우터(authGuard)가 한다 — 로그인-4(화면 정리)
    const toggleAuth = () => {
        if (!auth.value) {
            router.push('/login')
            return
        }
        openDialog({
            type: 'warning',
            title: '로그아웃',
            content: '로그아웃할까요? 실행 중인 터미널은 계속 실행됩니다.',
            buttons: [
                { view: '취소', type: 'info', bindButton: 'Escape', action: () => dialogOpen.value = false },
                { view: '로그아웃', type: 'warning', bindButton: 'Enter', action: () => { dialogOpen.value = false; logout() } },
            ],
        })
    }

    // 전체 화면 — 모바일 주소창을 숨기는 용도(문서 스크롤이 없어 브라우저가 스스로 숨기지 않음). 지원 안 하는 브라우저(iPhone 등)는 버튼 숨김
    const canFullscreen = document.fullscreenEnabled
    const fullscreen = ref(!!document.fullscreenElement)
    // 뒤로가기·Esc 등 브라우저가 풀어도 맞도록 이벤트로 읽는다
    const onFullscreenChange = () => fullscreen.value = !!document.fullscreenElement
    onMounted(() => document.addEventListener('fullscreenchange', onFullscreenChange))
    onBeforeUnmount(() => document.removeEventListener('fullscreenchange', onFullscreenChange))
    const toggleFullscreen = () => {
        const job = document.fullscreenElement ? document.exitFullscreen() : document.documentElement.requestFullscreen()
        job.catch(err => openDialog({ type: 'error', content: `전체 화면을 바꾸지 못했습니다: ${err?.message ?? err}` }))
    }

    return { auth, toggleAuth, canFullscreen, fullscreen, toggleFullscreen }
}
