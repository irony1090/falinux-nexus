# REF — 터미널 한글 입력 중복 (IME)

> 2026-10-20 신설. 이력 → `history/term-ime.md` / 확인 스크립트 → `REF-test-scripts-1020.md` "2026-10-20 한글 입력 중복" / 현재 진행 → `CURRENT.md`

**상태: 마감(2026-10-20) → `INDEX-term-ime.md`. 가) 프론트에서 ① 막기 + 헤드리스 재생 3/3 + 사용자 실입력 확인. xterm은 정식 6.0.0 유지.**

## 증상
웹 터미널에서 한글을 칠 때 간헐적으로 같은 글자가 두 번 들어감(`아니니네`, `앞앞에`). 처음엔 "빠르게 칠 때", 다음엔 "백스페이스 섞을 때"로 보였으나 둘 다 원인 아님.

## 원인 (SEND 로그로 확정)
리눅스 IME(Chrome 155, X11)는 받침이 다음 글자로 넘어갈 때(`닌` -> `니` + `ㄴ`) 확정을 이 순서로 한다:

| 순서 | 이벤트 | xterm 6.0.0 동작 |
|---|---|---|
| 1 | `compositionupdate ""` -> `input deleteContentBackward` -> `compositionend ""` | `_finalizeComposition(true)`가 조합 시작 위치를 복사하고 `setTimeout(0)` 예약 |
| 2 | `input insertText "니"`(comp=false) | `_inputEvent`: 리눅스 IME는 글자 키 keydown이 없어 `_keyDownSeen` false -> 조건 통과 -> **① 전송** |
| 3 | `compositionstart`(다음 글자) | 새 조합 시작 위치 기록 |
| 4 | 1의 타이머 | 입력창을 옛 시작 ~ 새 시작으로 잘라 **② 전송** -> 중복 |

- 보통 확정(`compositionend "글자"`)은 ②만 보내서 정상. 그래서 간헐적
- 6.1.0-beta.315에도 같은 경로(베타의 `_textareaChangeTimer`·`_compositionSuffix` 수정은 다른 경로) -> 베타 효과 없음 확인 후 6.0.0으로 복귀
- 오진 기록: ① 빠른 입력 타이머 누적(베타 수정 대상) ② 조합 중 Backspace keyCode 8 -> 둘 다 로그로 기각

## 해결 = 가) 프론트에서 ① 막기
- `apps/frontend/src/feature/process/util/xtermIme.util.ts` `blockImeRecommit(term)`: `term.element`(상자)에 capture 리스너 — 입력창의 `compositionend`가 빈 값이면 표시, 바로 다음 `input`이 `insertText`면 `stopPropagation`(xterm의 입력창 리스너까지 안 감). `compositionstart`·`keydown`·다음 `input`에서 표시 해제
- 입력창 값은 브라우저 기본 동작으로 바뀌므로(`stopPropagation`은 기본 동작을 안 막음) ② 타이머가 그 글자를 한 번 보낸다
- 호출 위치 = `processTerm.store.ts` `attach`의 첫 `term.open` 직후(입력창이 그때 생김)
- 안 나(onData 중복 버림)는 `ㅋㅋ`처럼 실제 반복과 구분 못 해 기각 / 안 다(업스트림 제보)는 미정

## 확인 (2026-10-20, 헤드리스 재생)
| 항목 | 결과 |
|---|---|
| 수정 없는 터미널에 로그 순서 재생 | `["a","아","니","니","네"," "]` — 실제 증상 재현 |
| 수정한 터미널 | `["a","아","니","네"," "]` — 중복 없음, 영문·조합 밖 입력 정상 |
| 다시 넣은 뒤 새 조합 없이 끝남 | `["니"]` — 글자 빠짐 없음 |

## 곁가지 (미룸)
- worker PTY에 `IUTF8` 없음 -> 줄 단위 입력(`read`·`cat`)에서 한글 백스페이스가 1바이트만 지움(`ime-test` 화면 `어라 앞앞에` vs 받음 `어라 라 앞앞에`). raw 모드(Claude Code·readline)는 무관 -> **2026-10-22 해결**(`REF-fix-board-iutf8.md`)
