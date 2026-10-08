# history/process-share-reconcile — 공유-d 실시간-1~3(재계산·연결) 이력

> 재사용 지식 → `REF-process-share-reconcile.md` / 공유-d 전체 이력 → `history/process-share-realtime.md` / 현재 진행 → `CURRENT.md`

## 2026-10-07(12) — 실시간-1 구조안 (코드 없음)
- 사용자 "진행" → 지금 코드 확인(구독·크기 우선권·판정 쿼리·입력 인가·노드 삭제와 process) → 구조안 + 정할 것 4개(SHARE:CHANGE 대상·경합·영향 계산 위치·확인 범위) 제시, 사용자 확인 대기

## 2026-10-07(13) — 실시간-1 코드 + 쿼리 확인
- 사용자 "진행해줘" → 정할 것 4개 제안대로 확정. `shares.sql` 쿼리 3개 + sqlc generate / `shareAccess.go`(새) / `accessMu` / `subscribeLiveProcesses` 잠금 / `SHARE:CHANGE` 메시지
- `go build`·`vet` 통과. 임시 DB 픽스처로 쿼리 3개 기대값 확인 → 임시 DB 삭제

## 2026-10-07(14) — 실시간-2 코드 + 격리 실행 확인
- 사용자 "이어서 진행해줘" → `share.go` PUT·DELETE 연결 + `shareAccess.go`(`attachSharedTile`·`afterShareChange`·`liveUnder`) + `perm.go`(`mustNodePerm`·`mustProcessPerm`)
- 구현 중 발견: 스크립트 노드도 공유 가능(SH1(공유 단위)) → 스크립트 공유는 타일 없음으로 정함(SD8(받는 쪽 루트 목록)로 루트 목록에)
- 격리 실행 31/31 + 크기 우선권 4/4 통과 → 임시 DB 삭제·격리 프로세스 종료

## 2026-10-07(15) — SH1(공유 단위) 개정: 폴더만 공유
- 사용자 질문 "공유를 디렉토리 단위로만?" → 추천(받는 쪽 화면 자리·상속 장비 표시·SD8/SD9 단순화, 잃는 것 = 공유 폴더 안 스크립트별 차단·단계 조정) → 사용자 "폴더 단위로 확정하고 진행"
- `share.go` `putNodeShare` 스크립트 = 400, 스크립트 타일 분기 삭제. 개발 DB 공유 행 0개(정리 불필요). 격리 실행 재확인 31/31·4/4 → 임시 DB 삭제

## 2026-10-08(16) — 실시간-3 코드 + 격리 실행 확인
- 구조안 승인 → `node.go`(`patchNode`·`deleteNode`에 `reconcileAccess` 연결, 헬퍼 `sameParent`·`collectAround`·`mergeUsers`). build/vet 통과, 격리 실행 15/15(이동 안/밖·순서만 변경·삭제), 임시 DB 삭제(커밋 전)

## 2026-10-08(17) — 실시간-4(exec 타일) 코드 + 격리 실행 확인
- 구조안 승인 → `ListNodeAncestors` 쿼리 + `attachViewerTerminal`(`tile.go`) + `attachViewerTiles`(`processApi.go`). build/vet 통과, 격리 실행 14/14 + 실시간-3 회귀 15/15, 임시 DB 삭제(커밋 전)

## 2026-10-08(18) — 실시간-5(노드 응답) 코드 + 격리 실행 확인
- 구조안 승인 → `ListSharedRoots` 쿼리 + `listChildren` 루트에 공유 루트 합침 + `getNode` `parentVisible` + DTO 필드 2개. build/vet 통과, 격리 실행 12/12, 임시 DB 삭제(커밋 전)
