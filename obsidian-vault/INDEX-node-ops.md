# INDEX — 노드 조작 (마감)

> 2026-10-14 색인 2단화로 MEMORY·HISTORY에서 접음(사용자 "이대로 진행"). **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
2026-06-26~10-13. Node 모듈(서버)과 그 위의 노드 조작 UI를 묶음.
| 파트 | 마감 | 한 줄 | REF |
|---|---|---|---|
| Node 모듈 · worker 명부(roster) · 서버 정리 | 2026-10-12 / 10-13 | 스키마·CRUD·PatchNode·`GET /workers` + `worker_instances` 명부(`GET/DELETE /workers/instances`) + 정리-1(부모 종류 검증 `requireFolder`). Label 모듈은 보류 | `REF-node-label.md` |
| 폴더 목록 동기화 | 2026-10-12 | node 변경을 볼 수 있는 접속 계정의 `TILES:<userId>`로, 발행 시점 DB 판정 | `REF-node-list-sync.md` |
| 노드 클립보드 | 2026-10-13 | 복사·잘라내기·붙여넣기 + 선택 모드·우클릭 메뉴·드래그 | `REF-node-clipboard.md` · `REF-node-select.md` |

**남은 것(CURRENT·미해결 이슈에 있음)**: RS6(위조 검증) 보류 / Label 모듈 보류(다시 볼 조건 → `REF-node-label.md` "Label 모듈")

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- `REF-node-label.md` / `history/node-label.md` — 모듈 REF라 MEMORY 모듈 현황표 "문서" 칸에는 그대로 남김
- `REF-node-list-sync.md`(폴더 목록 동기화 — node 변경을 볼 수 있는 접속 계정 토픽으로, 마감) / `history/node-list-sync.md`
- `REF-node-clipboard.md`(노드 복사·잘라내기·붙여넣기 UI) / `history/node-clipboard.md`
- `REF-node-select.md`(노드 선택 모드·우클릭 메뉴·드래그) / `history/node-select.md`

## MEMORY 모듈 현황표 행 (접기 전 원문)
| 모듈 | 상태 | 문서 |
|------|------|------|
| Node 카탈로그 | DB+CRUD+핸들러+PatchNode(`9c9d22e`) + `GET /workers`·exec 대상 장비 검증(`ba5b857`). UI 연동(연동 1차·타일 터미널·⑪(타일 트리 서버 저장)·4 O(계정 동기화) 화면) 마감 → `INDEX-node-ui-link.md`. **roster(명부) 마감**(2026-10-12, `GET/DELETE /workers/instances` + 칩·고르기·장비 지정 다이얼로그, `REF-node-label.md` "roster 계획") · label 보류 / **복사·잘라내기·붙여넣기 + 선택 모드·우클릭·드래그 마감(2026-10-13)** → `REF-node-clipboard.md`·`REF-node-select.md` | `REF-node-label.md`(+`REF-node-ui.md` UI 컨셉 / `REF-node-ui-overview.md` 전체보기·가상 키보드) |
| 실시간 push (socket) | 전송 토대·3모드 e2e 커밋(3a8e92e). **ping/pong(`transport.KeepAlive`, 10초/25초)로 응답 없는 소켓의 Hub 막힘 해소**(`e395433`, 2026-10-01). node Kind 어휘 확정 + process 동적구독 REST 완료 + **node CRUD 발행처 배선 완료**(`AfterCommit` 훅, 이동=2토픽, 2026-07-16, 빌드/vet 통과·e2e 미검증). **node 변경 전달 = 계정 토픽 + 발행 시점 판정, 마감**(2026-10-12, → `REF-node-list-sync.md`) | `REF-realtime.md` |

## HISTORY 색인 행 (원문)
### `history/node-label.md` — Node/Label 모듈
- 2026-10-13(2) 정리-1(부모 종류 검증) `requireFolder` 구현·격리 확인 → 파트(서버 정리) 마감·커밋
- 2026-10-13 핸들러 책임 재확인 — owner·사이클 검증은 이미 있음, 빈틈 = 부모 종류 검증(다음 파트(서버 정리))
- 2026-10-12(4) 사용자 확인 → 파트(worker 명부(roster)) 마감·커밋
- 2026-10-12(3) 명부-2(조회·삭제 API)~명부-4(확인) 구현, 격리 API 10/10 + 헤드리스 11/11(커밋 전, 사용자 확인 대기)
- 2026-10-12(2) 명부-1(기록) 구현·격리 확인(커밋 전)
- 2026-10-12 roster 계획(RS1(기록 시점)~RS6(위조 검증))·UI 스토리 확정, 다음 = 명부-1
- 2026-10-12 Label 모듈 보류 결정(공유는 5 공유로 해결·모아 보기 = 보드·검색이 먼저)
- 2026-09-30 (후속) 테스트 스크립트 id 2~4를 폴더 id 1로 이동 + `device_key` 지정(DB 직접 UPDATE)
- 2026-09-30 `GET /workers`(접속 인스턴스, `?nodeId=` 상속 해석) + `execProcess` 대상 장비 검증 — 연동 1차 1번
- 2026-06-29 Node 모듈 구현 (스키마·쿼리·핸들러·PatchNode·internal/patch, build/vet 통과)
- 2026-06-26 Node/Label 모듈 설계 (frontend 카탈로그)
### `history/node-list-sync.md` — 폴더 목록 동기화 (2026-10-12 신설)
- 2026-10-12 문제 제기(공유 폴더 목록이 바로 안 갱신됨) → NB(계정 토픽으로 보내기) 선택 → 구조안 제시
- 2026-10-12(2) 구조안 승인 → 동기화-1(서버 발행 대상)·동기화-2(프론트 수신) 구현(커밋 전)
- 2026-10-12(3) 동기화-3(확인) 격리 API·소켓 26/26 + 화면 11/11 → 사용자 확인 → 파트 마감·커밋
### `history/node-clipboard.md` — 노드 복사·잘라내기·붙여넣기 UI (2026-10-13 신설)
- 2026-10-13(3) 선택 모드 논의 → 시안 4장 → 승인(SM1~SM6·드래그 확정, 코드 없음)
- 2026-10-13(2) CP5(여러 개 담기)·CP6(표시 위치) 확정 + 클립보드-1~5 구현·격리 26/26(커밋 전)
- 2026-10-13 논의 시작, CP1(범위)~CP4(붙여넣은 뒤) + 부분 실패·계정 전환 확정(코드 없음)
### `history/node-select.md` — 노드 선택 모드·우클릭 메뉴·드래그 (2026-10-13 node-clipboard에서 분리)
- 2026-10-13(5) 사용자 확인 → 파트(노드 클립보드) 마감·커밋
- 2026-10-13(4) 선택-1(동작 분리)~선택-6(드래그) 구현 + queryFn 버그 수정, 격리 헤드리스 44/44(커밋 전, 사용자 확인 대기)

## 이 파트에만 해당하는 결정
- 생성·이동은 부모가 폴더여야 함(`requireFolder` 400, DB 제약 없음 — 생성·이동 경로가 핸들러뿐)
- 이동 = `checkMove`(목적지 편집 이상·같은 소유자·`IsDescendant` 사이클 금지) / 복사 = 폴더만, 원본을 다 읽은 뒤 넣음
- ※ device_key = main_key, roster ≠ 라우팅 authority, node 변경 = 계정 토픽은 다른 파트에서도 지켜야 해서 MEMORY 불변 결정에 남김
