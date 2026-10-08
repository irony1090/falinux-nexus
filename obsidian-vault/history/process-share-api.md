# history/process-share-api — process 공유: 공유-c(공유 API)

> 재사용 지식 → `REF-process-share-api.md` / 공유 전체 이력 → `history/process-share.md` / 현재 진행 → `CURRENT.md`

## 2026-10-07 — 공유-c(공유 API) 구조안 제시, 코드 없음
- 범위: SH4(공유 대상 찾기) 조회 + 노드·process 공유 걸기(upsert)·풀기·목록 + SH10(노드 복사). 구독 반영·`SHARE:CHANGE`는 공유-d
- 결정안 SC1(출처 계산 위치)~SC7(복사 제외 개수) — 사용자 확인 대기. SC1은 아직 개발 DB에 적용 전인 `00007_shares.sql`을 고치는 안(가까운 공유 계산을 `nearest_node_share` 한 곳으로)
- 같은 날 사용자: 커밋은 파트가 끝날 때 한 번(작업 단위마다 안 함)

## 2026-10-07(2) — SC1~SC7 확정 + 공유-c 코드 작성 · 확인(커밋 전)
- 사용자 "진행해줘" → SC1(출처 계산 위치)~SC7(복사 제외 개수) 제안대로 확정
- 작성: `00007_shares.sql`(`nearest_node_share` + `node_permission` 재작성), `query/shares.sql`(공유 행 CRUD·목록 8개), `user.go`(`/users/lookup`), `router/share.go`·`router/nodeCopy.go`(새 파일), `supervisorRouter.go`·`node.go`(라우트)
- 계획과 달라진 것: SC1 함수가 `RETURNS TABLE`이 아니라 출처 노드 id를 돌려주는 스칼라(sqlc 1.31.1이 TABLE 컬럼을 못 읽음) / 복사 목적지가 스크립트면 400 추가
- 확인: 격리 실행 42개 + 공유-b 회귀 50개 통과. 확인 뒤 격리 프로세스(실행 위치로 골라서)·임시 DB 정리
