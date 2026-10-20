# history — 터미널 한글 입력 중복 (IME)

> → `REF-term-ime.md` / 현재 진행 → `CURRENT.md`

## 2026-10-20
- 사용자: 웹에서 claude 세션에 한글을 빠르게 치면 같은 글자가 두 번 -> 앱 입력 경로(`onData` 하나) 확인, 후보 가) xterm / 나) Claude Code(Ink) -> 확인 노드 `ime-test`(irony `test`(14) 아래 221, REST 등록) -> bash가 받은 바이트에 중복 -> 가) 확정
- xterm 코드 읽고 `_handleAnyTextareaChanges` 타이머 누적 가설 -> 6.1.0-beta.315에 해당 수정 있음 -> 사용자 "베타로 올리고 정식 나오면 교체" -> 올림(8011 재시작 필요했음) -> 증상 남음, 사용자 관찰 "백스페이스 때"
- 콘솔 기록 스크립트로 이벤트 순서 수집: 1차는 SEND 줄 누락(앱 프레임이 문자열인데 바이너리로 해독), 콘솔 붙여넣기 관련 오답 안내도 함 -> 2차에 SEND로 확정: 리눅스 IME의 "지우고 일반 입력으로 다시 넣기" 확정에서 `_inputEvent` ① + 조합 마무리 타이머 ② (`니` 29294/29298, `앞` 39161/39165)
- 사용자 "가) 진행 + 6.0.0으로 되돌려" -> `package.json`·lock git 원복 + `npm install`, `xtermIme.util.ts` `blockImeRecommit` + `attach`에서 호출, type-check 통과, 헤드리스 재생 3/3
- 사용자 "vault 정리 하고 커밋" -> 확인 전 상태로 커밋 `26c5760`

## 2026-10-20(2)
- 8011 재시작(6.0.0 서빙 확인) -> 사용자 실입력 확인 -> "마감" -> `INDEX-term-ime.md`로 접음 + vault 커밋
