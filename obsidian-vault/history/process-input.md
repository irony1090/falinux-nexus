# HISTORY — process 터미널 입력 (키 입력 배선)

> 설계·결정 → `REF-process-input.md` / 현재 진행 → `CURRENT.md`.

---

## 2026-10-02(3) — I5(Ctrl+V 붙여넣기) 확정·반영 + 입력-c 브라우저 확인 7항목 통과 (미커밋)
- 사용자: dev 서버 띄움 + "추천대로" → I5 = Ctrl+V 붙여넣기. `copyOnCtrlC` → `clipboardKeys`로 바꾸고 Ctrl+V면 `false` 반환.
- 헤드리스 7항목 통과(타이핑·방향키·Ctrl+V·다른 탭 입력·전체보기 차단·선택+Ctrl+C 복사·Ctrl+C 중단). 첫 실행은 테스트 스크립트의 쿠키 도메인 실수(127.0.0.1)로 4번에서 멈춤 → 남은 process kill 후 수정·재실행. 테스트 노드 38·39 삭제. 상세 → `REF-process-input.md` "입력-c 브라우저 확인".
- 남은 확인: PENDING 중 입력 버림(socat 프록시).
- 사용자 PC 직접 사용 → "딱히 이상 없어 보임".

## 2026-10-02(2) — 입력-b(프론트) 코드 작성 (type-check 통과, 브라우저 확인 전·미커밋)
- 사용자 "입력-b 진행" 승인.
- `processTerm.store.ts`: `disableStdin` 제거 + `createHandle(uid)`에 `onData`/`onBinary` -> `sendInput`(PROCESS일 때만 emit) + `copyOnCtrlC`(I4(Ctrl+C 복사)). 상세 → `REF-process-input.md` "입력-b 구현".
- xterm 소스 확인으로 Ctrl+V = `0x16` 전송 + 붙여넣기 막힘 발견 → I5(Ctrl+V 붙여넣기) 결정 대기로 올림.

## 2026-10-02 — 입력-a(서버) 코드 작성 + 서버 단독 확인 통과 (미커밋)
- 사용자 "입력-a부터 진행" 승인.
- `protocol/messages.go`: `MsgProcessInput = "PROCESS:INPUT"`(payload = `DataEvent` 재사용).
- `router/input.go` 신규: `onProcessInput`(conn.On 등록, 연결 시점 userID 캡처) + `input`(버림 = entry 없음·`Inter` nil·소유 계정 불일치(로그 1줄)·상태 != PROCESS, 통과 시 `Inter.Write`).
- `router/subscribe.go`: `handleSubscribeWS`에서 Serve 전 `r.onProcessInput(conn, userID)`. `process.go`의 옛 입력 TODO 주석을 한 줄 포인터로 교체.
- 처음엔 supervisor가 떠 있다고 잘못 말함(재부팅 전 기록 근거) → 사용자 지적, 확인해 보니 꺼져 있었음.
- 사용자가 새 빌드로 supervisor·worker를 띄움 → 서버 단독 확인 4항목 통과(A 입력 에코 / B 계정 입력 버림 / `0x03` → exitCode 130 / 종료 uid 입력 버림). 테스트 계정 `input-test-b` 가입됨(DB에 남음), 테스트 노드 34~37은 삭제. 상세 → `REF-process-input.md` "입력-a 서버 단독 확인"

## 2026-10-01 — 입력 결정 I1~I4 확정 + 순서를 ⑪(타일 트리 서버 저장) 앞으로 (코드 없음)
- 1단계 커밋(`78bb8b2`·`e5200ce`·`78ee99c`) 뒤 사용자 "터미널 인터랙션은 언제쯤?" → 남은 구간(브라우저 -> supervisor, xterm 입력)과 결정 3개 제시, ⑪(타일 트리 서버 저장) 앞에 넣기를 추천.
- 사용자: I1(전달 경로) = 소켓 동의 / I2(입력 권한) = 같은 계정이면 누구나 / I3(PENDING 중 입력) = 전부 버림.
- 사용자 지적: 답변에서 "⑪"을 이름 없이 씀 → 번호에는 항상 이름을 붙일 것(전역 규칙 재확인).
- `REF-process-input.md` 신설(구조안 + 작업 단위 입력-a~입력-c).
- 사용자 "타일 동기화도 스냅샷 작업에서 같이 되나?" → 별개(배치 = ⑪(타일 트리 서버 저장)·4 O(계정 동기화) 화면 / 터미널 글자 = SNAPSHOT) 설명 → SNAPSHOT을 4단계 다음(4.5)으로 순서에 넣음(`REF-process-sync.md` "진행 순서").
- 사용자 "Ctrl+C도 버리나?" → PROCESS면 `0x03`으로 전달, PENDING이면 버림 설명 + 복사 충돌 제시 → I4(Ctrl+C 복사) = 선택 있으면 복사·없으면 `0x03` 확정.
