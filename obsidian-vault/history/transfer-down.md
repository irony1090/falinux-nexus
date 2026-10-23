# history — 내려받기·worker 간 전송

> → `REF-transfer-down.md` / 이전 파트 → `history/transfer-ext.md` / → `CURRENT.md`

## 2026-10-23 착수·조율 (코드 없음)
- 사용자 "worker간 파일 전송 & 웹에서 워커의 파일 다운로드 작업을 진행하고 싶어" -> BACKLOG의 "내려받기·worker 간 전송"을 파트(내려받기·worker 간 전송)로 착수
- 조사: 전송 모듈은 supervisor -> worker만(`FILE_INIT`·`FILE_CHUNK`·`FILE_RESULT`·`FILE_ABORT`), `send`는 seek 되는 읽기 원천(`ReadFile`·`ReadBuffer`) + 전체 hash. 장비 타일에 항목 선택·항목 메뉴 없음(본문 OS 파일 끌어 놓기만)
- Q32(브랜치 시작점)~Q35(복사·이동) 답 -> `REF-transfer-down.md`
- 큰 계획 전송-7(worker 송신)~13(확인) 제안. Claude 판단: 끌어오기(pull) 방식 = 받는 쪽 속도에 맞춰짐 + 파일 하나는 offset으로 이어받기, zip은 순서대로만
- 질문(같은 날 답 -> `REF-transfer-down.md`: Q36 파일만 · Q37 체크박스 · Q38 전송 목록에도):
  - Q36(worker 간 폴더): 파일만(폴더는 Q26(폴더째 올리기)와 같이 나중) vs 폴더도(B에 폴더 만들기 명령 추가)
  - Q37(항목 선택 방식): 체크박스(마우스 올리거나 선택 중이면 보임) vs Ctrl/Shift 클릭 + 폰 길게 누르기
  - Q38(내려받기 진행 표시): 전송 목록에도 올림 vs 브라우저 다운로드 막대만

## 2026-10-23 구현·확인 (전송-7~13)
- 사용자 "진행" -> 브랜치 `feat/transfer-down`(`feat/transfer-ext` 위) -> 전송-7(worker 송신)~10(worker 간 중계) 서버, 전송-11(화면: 선택·메뉴)·12(화면: 끌어 놓기) 화면 -> 상세 `REF-transfer-down.md` "구현"
- 결정(Claude): 다운로드 주소는 POST로 등록 -> GET 한 번(id = 16자 랜덤, 세션도 확인 — 쿠키 SameSite=Lax라 브라우저 기본 다운로드에 실림) / worker 간은 hash 생략(미리 구하면 원본을 한 번 더 다 읽어야 하고 10초 호출 제한에 걸림) / 동시 중계 3
- 확인: 서버 격리 42/42(`dl_test.py`), 헤드리스 26/26 두 번(`dl.mjs`) → `REF-test-scripts-1022.md`. C5(보내는 worker kill 뒤 다시 열어 이어감) 로그로 다시 열기 경로 확인
- 헤드리스 drag 불안정 -> 이벤트 합성으로 바꿈(함정 기록). 정리 때 cmdline 매칭으로 내 셸까지 끔(함정 기록 → `REF-test-env.md`). 격리 프로세스·`nexus_xferdown_tmp` 정리함

## 2026-10-23 마감
- 사용자 브라우저 확인 "확인은 했어. 잘돼." -> 파트(내려받기·worker 간 전송) 마감, `INDEX-transfer-down.md`로 접음
- 커밋: `feat/transfer-down`에 파트 마감 커밋 -> master에 `feat/transfer-ext`(파트(파일 전송 확장))·`feat/transfer-down` 차례로 `--no-ff` 머지. push 안 함
