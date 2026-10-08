# history — 공유-e(프론트 표시)

> → `REF-process-share-display.md` / 현재 → `CURRENT.md`

## 2026-10-08 공유-e-0(제안 항목 확정)
- 정리: O4(접속 표시)는 공유-f(presence) 데이터라 공유-e에서 제외. 숨기기는 복원 UI(숨긴 공유 N)·서버 작업이 같이 따라와 나중으로. O5는 서버 API 신설이 필요해 공유-e-3 마지막 단위로 분리
- 사용자 질문 정리: presence는 REST 조회로 시작(패널 열 때 조회, 필요하면 열려 있는 동안만 재조회, push는 나중) / 숨기기는 받는 쪽 화면에서만 치우는 기능이며 `hidden_at`을 비우면 복원
- 시안 "실행 위치 묶음 시안" 확인 후 ①~④·⑥ 채택, B 제외. 단계 아이콘 편집(30)=`pencil`, ⑥은 소유자 쪽에도 표시, 서버 응답 3곳 추가(process 응답 `startedBy`·`ownerNickname`, 노드 단건 `ownerNickname`) 방향 확정. 코드 없음

## 2026-10-08(2) 공유-e-1(받는 쪽 표시) 코드
- 서버: `ListUserNicknames` 신설 + process 응답 `ownerNickname`·`startedBy` + `GET /nodes/:id` `ownerNickname` (`go build`·`go vet` 통과)
- 프론트: `shareView.util.ts` 신설 + `TileFrame.vue` ①~⑥ + `FolderTileBody.vue` ④ + `PROCESS:UPDATE`가 계정 기준 값 보존 (`vue-tsc` 통과). 구현 메모 → `REF-process-share-display.md`
- 화면 확인 전, 커밋 전

## 2026-10-08(3) 공유-e-1 격리 실행 확인
- 계정 둘 + worker + 격리 프론트로 ①~⑥·보기 전용·소유자 쪽·모바일 확인 통과. 발견 1건(단계 변경 때 터미널 `permission`이 갱신 안 됨) → `refreshAccount` 추가 후 통과. 결과 표 → `REF-process-share-display.md`

## 2026-10-08(4) 공유-e-2(공유 걸기 창) 구조안·시안 확정
- 구조안 제시(범위·흐름·파일별 변경·정할 것 5개) → 직접 눌러 보는 시안 아티팩트 작성 → 사용자 "괜찮은 것 같아". 코드 없음. 내용 → `REF-process-share-display.md`

## 2026-10-08(5) 공유-e-2(공유 걸기 창) 코드·격리 확인
- `share.api.ts`·`ShareDialog.vue` 신설 + `FolderTileBody.vue` O1 버튼(소유자 폴더만) + `shareView.util.ts` 상수. `vue-tsc` 통과, 격리 실행 11개 통과(서버 변경 없음). 구현 메모 → `REF-process-share-display.md`

> 공유-e-3 이력 -> `history/process-share-owner.md` / 공유-f 이력 -> `history/process-share-presence.md`
