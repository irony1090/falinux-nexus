import type { Terminal } from '@xterm/xterm'

// blockImeRecommit는 IME가 조합을 비우고 일반 입력으로 다시 넣는 확정(compositionend "" -> insertText "X")에서
// xterm이 X를 두 번 보내는 것을 막는다: ① _inputEvent(글자 keydown이 없어 조건 통과) ② 조합 마무리 타이머(입력창을 잘라 보냄).
// ①만 막는다 — 입력창 값은 브라우저 기본 동작으로 바뀌므로 ②가 한 번 보낸다. term.open 뒤에 부를 것 → REF-term-ime.md
export const blockImeRecommit = (term: Terminal) => {
    const box = term.element
    const ta = term.textarea
    if (!box || !ta) return
    let emptyEnd = false   // 직전 compositionend가 빈 값
    box.addEventListener('compositionend', e => {
        if (e.target === ta) emptyEnd = (e as CompositionEvent).data === ''
    }, true)
    box.addEventListener('compositionstart', () => { emptyEnd = false }, true)
    box.addEventListener('keydown', () => { emptyEnd = false }, true)
    // 상자의 capture 단계라 입력창에 붙은 xterm 리스너보다 먼저 돈다
    box.addEventListener('input', e => {
        const ev = e as InputEvent
        if (ev.target !== ta || !emptyEnd) return
        emptyEnd = false
        if (ev.inputType === 'insertText') ev.stopPropagation()
    }, true)
}
