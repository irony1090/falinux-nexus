# history/process-snapshot-impl — 화면복원 스냅샷 구현 (스냅샷-a~d)

> 구현·확인 지식 → `REF-process-snapshot-impl.md` / 설계 이력 → `history/process-snapshot.md` / 현재 진행 → `CURRENT.md`
> 2026-10-02 `history/process-snapshot.md`에서 분할.

## 2026-10-02(2) — 스냅샷-a(서버 ring) 작성 (미커밋)
- 사용자 "시작해줘" → `internal/ring`(신규) + `bind/screen.go`(ring + alt 판정, S6) + relay 배선 + `DataEvent.Off` + `router/screen.go`(uid 맵, S3). 상세 → `REF-process-snapshot-impl.md` "스냅샷-a 구현"
- 작성 중 잡은 것: alt 판정 경계용 tail을 "tail + p 앞 7바이트"의 끝으로 갱신하던 실수 → 긴 묶음이면 p의 끝 7바이트여야 함(테스트 "split after long chunk"로 고정)
- build/vet/go test 통과. 다음 = 스냅샷-b(서버 API)

## 2026-10-02(3) — 스냅샷-b(서버 API) 작성 (미커밋)
- `GET /processes/snapshot/:processId` + `redraw`(SIGWINCH 유도). 상세 → `REF-process-snapshot-impl.md` "스냅샷-b 구현"
- 작성 중 잡은 것: ① 소유자 확인을 memory entry로 하면 PENDING이 404 → DB로 ② alt + worker 끊김이면 다시 그리기 불가 → 응답에 `redraw` 추가, false면 data를 그대로 씀
- build/vet 통과, 실행 확인은 supervisor 재기동 대기

## 2026-10-02(4) — 스냅샷-b 서버 단독 확인 (supervisor 재기동 2회)
- 1차 9/11: resize 전 0x0 크기로 다시 그리기 → PTY 0x0 버그 → 크기 모르면 `redraw=false`
- 2차 12/13: htop 다시 그리기 출력 없음 → 두 `Layout` 사이 경합 확인(`snap/probe.mjs`, 간격 0/50/200ms = 성공/실패/성공) → 바꾼 뒤 출력이 나올 때까지(최대 500ms) 기다렸다 되돌림(`Screen.Off`·`Ring.Off` 추가). 2b는 5회 반복 검사로 강화
- 3차(재기동 후): **13/13 통과, 연속 2회**(2b 다시 그리기 5/5). 테스트 계정 `snap-*`·`snap-other-*`·`snap-probe-*` DB에 남음(노드는 삭제함)

## 2026-10-02(5) — 스냅샷-c(프론트) 작성 + 스냅샷-d 헤드리스 확인 10/10 (미커밋)
- `process.api.ts` `getProcessSnapshot` + `processTerm.store.ts`(보관함 off, `writeHeld`, `writeSnapshot`, restore = 스냅샷 받은 뒤 handle 생성). 상세 → `REF-process-snapshot-impl.md` "스냅샷-c 구현"
- 작성 중 바꾼 것: ① 빈 구간 안내 줄을 보관분 뒤가 아니라 빠진 자리에 ② alt여도 안 잘렸으면 data 그대로(일반 화면 보존) ③ 복원 중 xterm 질의 응답 입력 막기(`replaying`) — vim 스냅샷의 `ESC[6n` 발견에서
- 확인 1차 5/6(쿠키 `=` 파싱 함정으로 C 401), 2차 8/9(vi 판정 기준 틀림), 3차 9/10(남은 vim swap 파일), 4차 **10/10**. 막기 제거 시 F1 실패(after=2) 확인 후 원복
- 발견(범위 밖): 여러 탭이 같은 터미널을 열면 질의 응답이 탭 수만큼 감 → CURRENT 미해결

## 2026-10-02(6) — 사용자 PC 확인 + 커밋
- 사용자: htop 확인 / 다른 탭 다시 그리기 = "약간 반짝하는 정도" / 끝난 process 새로고침 확인 → 4.5 SNAPSHOT 마감, 커밋 `3de3064`(서버)·`997d6cf`(프론트)·`9a17dbb`(vault), push 안 함
