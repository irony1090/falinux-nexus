# history/workspace-server — 작업 공간 여러 개: 서버 코드 작성 이력 (작업공간-a~c)

> 2026-10-07 `history/workspace-impl.md`에서 분할. 재사용 지식 → `REF-workspace-server.md` / 구조·결정 이력 → `history/workspace-impl.md` / 현재 진행 → `CURRENT.md`

## 2026-10-07(3) — 작업공간-a(DB) 코드·확인 (커밋 전)
- 사용자 "시작해줘" → hash 생성 위치는 제안대로(기존 행 SQL, 새 행 Go). `00008_workspaces.sql` + `tileTrees.sql`(기본 쿼리 4개에 `is_default` 조건, 쿼리 8개 추가) + `tile.go` `newTreeHash()`
- 확인: 임시 DB에서 마이그레이션·쿼리 20개·down/up 통과 → 상세 `REF-workspace-server.md` "작업공간-a(DB) 구현 메모"
- **사고**: 확인 명령 끝에 불필요한 `go run ./cmd/supervisor --help`를 넣어 supervisor가 실제로 뜸(`--help` 무시) → `init()`이 **개발 DB `nexus`에 00007·00008 적용**(12:49). 약 2분 뒤 발견해 해당 pid 2개만 종료. 데이터 확인: 트리 49행 전부 기본·hash 채워짐·version 보존, nodes `created_by` 전부 채워짐, 공유 행 0. 5050은 비어 있었음. 되돌릴지는 사용자 판단 → 함정은 `REF-test-env.md`에 기록

## 2026-10-07(4) — 작업공간-b(API) 코드·확인 (커밋 전)
- 사용자 "다음작업 진행하자" → 개발 DB 되돌리기 질문엔 답 없음, 버전 8 그대로 둠
- `tile.go`(`/tiles/:hash`, WS5(터미널 타일 추가 주체) 검사, `publishTiles`에 hash) + `workspace.go`(목록·만들기·이동·이름·삭제) + protocol(`TilesUpdateEvent.Hash`, `WORKSPACES:UPDATE`)
- 확인: 격리 실행 41개 통과 → `REF-workspace-server.md` "작업공간-b(API) 구현 메모"
- 과도기 위험 기록: 프론트 `adoptOrphans` PUT이 400이 됨(작업공간-c·d에서 해소)

## 2026-10-07(5) — 작업공간-c(exec·안전망·탭) 코드·확인 (커밋 전)
- 사용자 "다음 작업 진행해보자" → exec `workspace` · 서버 고아 안전망(WS3(고아 안전망 위치)) · `TAB:VIEW` + 크기 우선권 후보 좁힘(WS7(보는 작업 공간 전달))
- 구현 중 정한 것(REF에 없던 것): 안전망을 `/tiles/:hash` GET에서도 돌림 / exec 타일 추가 때 전체 트리 잠금 + uid 있으면 그 타일 반환 / TAB:VIEW 전 = 기본을 봄(지금 프론트 호환) / `resyncSizeOwner` — 소유 탭이 터미널을 화면에 두지 않으면 넘김·후보 없으면 비움, 작업 공간 동작·TAB:VIEW·연결 때 호출
- 확인: worker 포함 격리 실행 20개 통과(끊김 넘김 62초 포함) → `REF-workspace-server.md` "작업공간-c"
