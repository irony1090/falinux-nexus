# history/process-share-db — process 공유: 공유-a(DB)

> 재사용 지식 → `REF-process-share-db.md` / 공유 전체 이력 → `history/process-share.md` / 현재 진행 → `CURRENT.md`

## 2026-10-06 — 공유-a(DB) 구조안 제시, 코드 없음
- 구조안: `00007_shares.sql` = `node_shares`·`process_shares`(PK (대상, 받는 계정), 단계 SMALLINT) + `nodes.created_by`·`processes.started_by` + 판정 함수 `node_permission`·`process_permission`. 판정 쿼리 6개(`query/shares.sql`)
- 결정안 SA1(단계 저장 형식) = SMALLINT 0/10/20/30, 소유자 40은 함수 반환만 / SA2(판정 위치) = DB 함수 / SA3(숨기기 시점) = 공유-e 때 / SA4(기존 행 채우기) = 소유자로 채우고 NOT NULL — 사용자 확인 대기
- 스크래치 확인: sqlc v1.31.1이 스키마 안 `CREATE FUNCTION ... RETURNS SMALLINT`를 읽어 함수 호출 결과를 `int16`으로 생성함(스크래치 디렉터리, 프로젝트 무변경)

## 2026-10-06(2) — SA1~SA4 확정 + 공유-a 코드 작성 · 확인
- 사용자: SA1(단계 저장 형식)~SA4(기존 행 채우기) 제안대로 확정
- 작성: `migrations/00007_shares.sql`, `query/shares.sql`(판정 쿼리 6개), `nodes.sql` CreateFolder/CreateScript에 `created_by`, `processes.sql` CreateProcess에 `started_by`, `gen/const.go` Perm 상수, 호출부 2곳(`router/node.go`, `process/manager.go`), `sqlc generate`
- 확인: `go build ./...`·`go vet ./...` 통과 / 임시 DB `nexus_share_tmp`에서 판정 함수·목록 쿼리·Down 전부 기대값(표 → `REF-process-share-db.md` "확인 방법"), 확인 뒤 DB 삭제
- 동작 변화 없음(판정 쿼리는 아직 호출부 없음 — 공유-b(인가 교체)에서 연결). 커밋 전

## 2026-10-07 — 공유-a 검토 + `ListChildrenVisible` 보완
- 사용자 질문("문제가 될 요소")으로 검토: 커밋을 막을 문제 없음. 판정 함수가 "하위 노드 소유자 = 상위 노드 소유자"에 기대고 DB로 강제 안 됨(당시 `createNode`·이동에 부모 소유자 검사 없음 → 공유-b에서 해결) / 소유자 자신에게 건 공유 행이면 `ListChildrenVisible`이 자식을 숨길 수 있음 / 루트(NULL) 불가. 개발 DB `nexus` 읽기 점검: 소유자 섞인 트리 0·사이클 0
- 사용자: parent_perm 조건 한 줄 넣기로 → `ListChildrenVisible`에 `parent_perm = 40`이면 40 + 루트 불가 주석. 공유-b 확인 때 함께 확인(통과)
