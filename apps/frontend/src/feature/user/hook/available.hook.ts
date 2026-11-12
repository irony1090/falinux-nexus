import { onBeforeUnmount, ref, watch, type Ref } from 'vue'
import { checkAvailable, type AvailableKind } from '../api/account.api'

// 입력이 멈추고 이만큼 뒤에 확인(아이디-2(가입 분리))
const DELAY = 400

export type AvailableState = 'idle' | 'checking' | 'ok' | 'bad'

// 아이디·닉네임 중복 자동 확인 — same(v)가 true면(자기 지금 값 등) 확인하지 않고 idle
export const useAvailable = (kind: AvailableKind, value: Ref<string>, same: (v: string) => boolean = () => false) => {
    const state = ref<AvailableState>('idle')
    const message = ref('')
    let timer: ReturnType<typeof setTimeout> | undefined
    let seq = 0     // 늦게 온 옛 응답 버림

    watch(value, raw => {
        clearTimeout(timer)
        const v = raw.trim()
        const my = ++seq
        message.value = ''
        if (!v || same(v)) { state.value = 'idle'; return }
        state.value = 'checking'
        timer = setTimeout(() => {
            checkAvailable(kind, v)
            .then(r => {
                if (my !== seq) return
                state.value = r.available ? 'ok' : 'bad'
                message.value = r.message
            })
            .catch(() => { if (my === seq) state.value = 'idle' })   // 확인 실패 = 저장 때 서버가 다시 봄
        }, DELAY)
    })
    onBeforeUnmount(() => clearTimeout(timer))

    return { state, message }
}
