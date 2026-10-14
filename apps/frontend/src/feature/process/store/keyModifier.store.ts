import { readonly, reactive } from 'vue'

// 특수 키 줄의 Ctrl·Alt — 모든 터미널이 같은 값을 보므로 모듈 하나에 둔다. VK3(Ctrl 동작) → REF-mobile-keyboard.md
// 탭 = 다음 글자 하나에만 / 누른 채 글자 입력 = 뗄 때 해제(휴대폰 Shift 방식).
// 누르는 중 가상 키보드 터치로 pointercancel이 와도 걸린 채 남으므로 다음 글자에는 적용된다(기기마다 다른 부분을 덮음)

export type Mod = 'ctrl' | 'alt'
type ModState = { armed: boolean; held: boolean; used: boolean; wasArmed: boolean }

const state = reactive<Record<Mod, ModState>>({
    ctrl: { armed: false, held: false, used: false, wasArmed: false },
    alt: { armed: false, held: false, used: false, wasArmed: false },
})

export const pressMod = (m: Mod) => {
    const s = state[m]
    Object.assign(s, { wasArmed: s.armed, armed: true, held: true, used: false })
}

// 누른 채 입력했으면 해제, 걸린 상태에서 다시 탭하면 해제, 아니면 다음 글자까지 유지
export const releaseMod = (m: Mod) => {
    const s = state[m]
    if (!s.held) return
    s.held = false
    if (s.used || s.wasArmed) s.armed = false
}

const consume = () => {
    (Object.keys(state) as Mod[]).forEach(m => {
        const s = state[m]
        if (!s.armed) return
        if (s.held) s.used = true
        else s.armed = false
    })
}

// 두벌식 자판 자리 -> 영문(한글 자판 상태에서 Ctrl+ㅊ = Ctrl+C)
const HANGUL_KEYS: Record<string, string> = Object.fromEntries(
    [...'ㅂㅈㄷㄱㅅㅛㅕㅑㅐㅔㅁㄴㅇㄹㅎㅗㅓㅏㅣㅋㅌㅊㅍㅠㅜㅡㅃㅉㄸㄲㅆㅒㅖ'].map((ch, i) => [ch, 'qwertyuiopasdfghjklzxcvbnmqwertop'[i]]),
)

const CTRL_SPECIAL: Record<string, number> = { '@': 0, ' ': 0, '[': 0x1b, '\\': 0x1c, ']': 0x1d, '^': 0x1e, '_': 0x1f, '/': 0x1f, '-': 0x1f, '?': 0x7f }

const ctrlOf = (ch: string) => {
    if (/^[a-z]$/i.test(ch)) return String.fromCharCode(ch.toLowerCase().charCodeAt(0) & 0x1f)
    return ch in CTRL_SPECIAL ? String.fromCharCode(CTRL_SPECIAL[ch]) : ch
}

export const modsActive = () => state.ctrl.armed || state.alt.armed

// 걸린 Ctrl·Alt를 글자 하나에 적용하고 소비한다. 여러 글자(붙여넣기·자동 완성)·질의 응답은 그대로 둔다
export const applyMods = (data: string): string => {
    if (!modsActive() || [...data].length !== 1) return data
    let ch = HANGUL_KEYS[data] ?? data
    if (state.ctrl.armed) ch = ctrlOf(ch)
    if (state.alt.armed) ch = '\x1b' + ch
    consume()
    return ch
}

// 방향키 등 CSI 키의 수정자 번호(xterm: 1 + Alt 2 + Ctrl 4). 없으면 1, 쓰면 소비
export const takeCsiMod = () => {
    const n = 1 + (state.alt.armed ? 2 : 0) + (state.ctrl.armed ? 4 : 0)
    if (n > 1) consume()
    return n
}

export const keyMods = readonly(state)
