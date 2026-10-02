# HISTORY — node 타일 트리 서버 저장 구현 (⑪ 3-a~3-b')

> 2026-10-02 `history/node-ui-save.md`에서 분할. 요약·결정 → `REF-node-ui-save-impl.md` / 설계·3-c 구조안 → `REF-node-ui-save.md` / 현재 진행 → `CURRENT.md`.

---

## 2026-10-02(3) — 3-b'(exec 연동) 실패 대처 결정 + 작성 + HTTP 확인 통과
- 3-b' 구체안(실행 후 잠금, 응답 필드 덧붙이기) 제시 → 사용자 "저장이나 커밋 실패 시 대처는?" → 요청 트랜잭션은 응답 뒤 커밋이라 실패를 못 잡는 문제 발견 → ① 별도 트랜잭션 ② 실패 시 kill + 500 ③ 3-d 안전망 제안.
- 답변에서 "⑪-4"를 이름 없이 써서 "11-4가 뭐야?" 질문 받음(번호 이름 누락 3번째) → 설명 + 자동 메모리 보강.
- 사용자 "그렇게 진행" → `appendTerminalTile`(별도 트랜잭션) + `execProcess` 연결(보상 kill, 5초 분리 컨텍스트). build/vet 통과. 재기동 후 HTTP 6항목 — 첫 실행은 테스트 스크립트가 해시된 identification으로 행을 찾아 6번(보상)이 헛돌아 FAIL, 남은 process kill 후 rootId로 찾게 고쳐 전부 통과. 상세 → `REF-node-ui-save-impl.md` "3-b'(exec 연동) 구현".

## 2026-10-02(2) — 3-b(API) 작성 + HTTP 확인 통과
- 사용자 "순서 문제는 괜찮겠어?" → JSONB 키 순서 영향 확인(화면 순서는 `kids` 배열만, `tilesUnder` 닫는 순서도 결과 동일) → REF에 기록. 이어서 "3-b 진행".
- `router/tile.go`(GET/PUT + validate + marshal) + mount + go.mod uuid direct. build/vet + 검사 임시 테스트 12케이스 통과. 사용자 재기동 후 HTTP 5항목 + 첫 GET 동시 8개 경합(루트 1개) 통과. 상세 → `REF-node-ui-save-impl.md` "3-b(API) 구현".

## 2026-10-02 — 3-a(DB) 작성
- 사용자 "다음 작업 확인 진행" → 3-a 구체안(테이블·쿼리 4개) 제시 → "진행" 승인.
- `00006_tile_trees.sql` + `query/tileTrees.sql` + `sqlc generate`, build/vet 통과. PG에서 트랜잭션 롤백으로 쿼리 동작 확인. 상세 → `REF-node-ui-save-impl.md` "3-a(DB) 구현".
- 커밋 `2aa0920`(3-a·3-b·3-b' 코드, 2026-10-02).
