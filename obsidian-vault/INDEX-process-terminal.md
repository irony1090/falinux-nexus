# INDEX — process 터미널 기능 (마감)

> 2026-10-06 색인 2단화로 MEMORY·HISTORY에서 접은 파트(사용자 확인). **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
배선(→ `INDEX-process-wiring.md`) 위에 얹은 터미널 사용 기능 3개, 2026-10-01~10-06: 터미널 입력(소켓 `PROCESS:INPUT`, I1~I5) / SNAPSHOT(supervisor ring buffer + 누적 offset + alt screen 다시 그리기, S1~S6) / EDIT(worker `vi` 편집, 스크립트 행 메뉴 "vi로 편집", E1~E4). 동기화·탭 id·공유(`REF-process-sync*.md`)는 5 공유(P·R·T)가 이어 쓰므로 이 파트 밖.

**남은 것(CURRENT 미해결로 옮겨 둠)**: PENDING 중 입력 버림 확인(I3) / 같은 터미널을 연 탭마다 질의 시퀀스에 응답(스냅샷-d) / E3(저장 결과 표시) 첫 범위 제외.

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- 설계/재사용 지식: `REF-process-input.md` `REF-process-snapshot.md` `REF-process-snapshot-impl.md` `REF-process-exec-edit.md`
- 작업 이력: `history/process-input.md` `history/process-snapshot.md` `history/process-snapshot-impl.md` `history/process-exec-edit.md`

## 모듈 현황표 행 (MEMORY 원문, 접기 직전)
| 모듈 | 상태 | 문서 |
|------|------|------|
| process 실행 (execute/pty/manager/bind) | 배선(exec/kill/resize/재접속/구독) 마감 → `INDEX-process-wiring.md` + input(키입력) 완료(2026-10-02, `2db8998`) + SNAPSHOT(화면복원 ring) 완료(2026-10-02, 사용자 PC 확인·커밋) | `REF-process.md`(+`-exec-edit`/`-input`/`-snapshot`(+`-impl`)/`-sync`/`-sync-impl`) |

## 이 파트에만 해당하는 불변 결정
- 없음 — "frontend 끊김≠종료 / 같은 세션 재접속=보던 화면 복원" 등은 다른 파트에서도 지켜야 해서 MEMORY에 남김

## HISTORY 색인 (원문)
### `history/process-input.md` — process 터미널 입력 (2026-10-01 신설)
- 2026-10-02(3) I5(Ctrl+V 붙여넣기) = 붙여넣기 확정·반영 + 입력-c 브라우저 확인 7항목 통과
- 2026-10-02(2) 입력-b(프론트) 코드 작성(`processTerm.store.ts`, I4(Ctrl+C 복사) 포함), type-check 통과·브라우저 확인 전 + I5(Ctrl+V 붙여넣기) 발견
- 2026-10-02 입력-a(서버) 코드 작성(`router/input.go`, `PROCESS:INPUT`) + 서버 단독 확인 4항목 통과
- 2026-10-01 결정 I1~I4 확정(소켓 / 같은 계정 누구나 / PENDING 중 버림 / Ctrl+C = 선택 있으면 복사) + 순서를 ⑪(타일 트리 서버 저장) 앞으로, 코드 없음

### `history/process-snapshot.md` — 화면복원 스냅샷 (ring buffer)
- 2026-10-02 구조안 S1~S6 제시 + htop 부분 갱신 문제로 7월 기록 정정 + C(ring + alt screen만 다시 그리기)·S1~S6 전부 확정 + 작업 단위 스냅샷-a~d 제시
- 구현 이력(2026-10-02(2)~) → `history/process-snapshot-impl.md`

### `history/process-snapshot-impl.md` — 화면복원 스냅샷 구현 스냅샷-a~d (2026-10-02, process-snapshot에서 분할)
- 2026-10-02(6) 사용자 PC 확인(htop·다른 탭 약간 반짝·끝난 process) + 커밋 `3de3064`(서버)·`997d6cf`(프론트)·`9a17dbb`(vault)
- 2026-10-02(5) 스냅샷-c(프론트) 작성 + 헤드리스 10/10 — 복원 중 질의 응답 입력 막기 추가
- 2026-10-02(2) 스냅샷-a(서버 ring) 작성 — `internal/ring` + `bind.Screen`(alt 판정) + relay·`DataEvent.Off`·uid 맵, go test 통과, 미커밋
- 2026-10-02(3) 스냅샷-b(서버 API) 작성 — `GET /processes/snapshot/:processId` + `redraw`, 소유자 확인 = DB(PENDING 대응), 미커밋
- 2026-10-02(4) 스냅샷-b 서버 단독 확인 13/13(2회) — 0x0 크기 다시 그리기·두 `Layout` 사이 경합 버그 2개 수정
- 2026-07-16 ring buffer 설계 논의 착수(코드 없음, 순수 설계): supervisor-side 채택 + 스케일 검토 + worker-side 이전 시 필요한 protocol(RingBuffer/offset/MsgSnapshot) + snapshot↔live 이음매 race 발견(Hub 구조상 conn별 차등 라우팅 불가, `bind.CatchUp` 미완성)

### `history/process-exec-edit.md` — 실행 타입 EXEC vs EDIT (worker vi 편집, 2026-10-02 신설)
- 2026-10-06 E1~E4 확정 + 편집-a(프론트 진입) 작성·편집-b(확인) 사용자 PC 통과, 커밋 `97fb5c4`
- 2026-10-02 4.6 EDIT 프론트 연결 구조안 제시(E1~E4 + 편집-a·b), 노드 캐시 갱신 불필요로 정정, 코드 없음 — 사용자 확인 대기
