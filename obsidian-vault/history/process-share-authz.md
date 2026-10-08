# history/process-share-authz — process 공유: 공유-b(인가 교체)

> 재사용 지식 → `REF-process-share-authz.md` / 공유 전체 이력 → `history/process-share.md` / 현재 진행 → `CURRENT.md`

## 2026-10-06 — 공유-b(인가 교체) 구조안 제시, 코드 없음
- 소유자 검사 자리 전수 조사(grep `OwnerUserID`·`requireOwnedEntry`·`subscribeUser`·`handoffSizeOwner`·`GetNode` 등) → 엔드포인트별 요구 단계 표 + 작업 단위 인가-1(판정 헬퍼)~인가-8(확인)
- 결정안 SB1(권한 부족 응답)~SB6(이동 사이클 검사) — 사용자 확인 대기
- 발견: 모든 브라우저 소켓이 `NODE:0`을 구독 → 다른 계정의 루트 노드 이벤트(`content` 포함)가 전 계정에 전송됨. 프론트가 `NODE:*`를 아직 안 들어서 화면엔 안 보임 → SB5(`NODE:0` 누설 수정)로 제안
- 발견: 크기 우선권 `sizeOwnerOnConnect`가 연결한 탭에 빈 우선권을 주는데, 공유 후엔 보기 단계 탭이 우선권을 가져갈 수 있음 → SB3(크기 우선권 후보)에 연결 때 필터 포함

## 2026-10-06(2) — SB1~SB6 확정, 코드 없음
- 사용자: SB1(권한 부족 응답)~SB6(이동 사이클 검사) 제안대로 확정 → 공유-b 구조 합의 완료. 코드는 다음 세션(인가-1(판정 헬퍼)부터)
- 사용자 요청으로 `NODE:*` 사용처 검사: 프론트 수신 등록 6종에 없음, dispatch가 미등록 type을 버림, 서버는 구독 1곳·발행 3핸들러뿐 → 쓰는 곳 없음 확인(전송만 일어남)
- SB5(`NODE:0` 누설 수정) 개정: 계정별 토픽 대신 `NODE:0` 구독 줄 삭제(쓰는 곳이 없으므로). 토픽 설계는 열린 질문 4(NODE 동적 구독)로 미룸. 코드는 인가-7에서

## 2026-10-07 — 공유-b 코드 작성 · 확인(커밋 전)
- 인가-1(판정 헬퍼)~인가-7(루트 토픽) 코드 작성(`router/perm.go` 신설, processApi·screen·processDto·subscribe·input·sizeOwner·process·workerApi·node·nodeDto, `process/manager.go`, 쿼리 3파일 + `sqlc generate`). 상세 → `REF-process-share-authz.md` "구현 메모"
- 구조표에 없던 판단 1건: 같은 부모 안 순서 변경(`parentId` = 지금 부모)은 이동 검사를 생략(받는 쪽이 아래 공유 행 때문에 순서도 못 바꾸는 것 방지)
- sqlc 함정 발견: 재귀 CTE를 EXISTS 안·한정 이름으로 쓰면 `table alias does not exist` → COUNT 형태
- 인가-8(확인): 임시 DB `nexus_authz_tmp` + supervisor 5051 + worker(`authz-dev`)로 격리 실행, 53개 항목 전부 통과(표 → REF "확인 결과"). 확인 뒤 프로세스 종료·DB 삭제. 개발 DB `nexus`·5050 무접촉
