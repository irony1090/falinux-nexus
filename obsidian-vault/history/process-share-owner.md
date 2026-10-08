# history — 공유-e-3 소유자 쪽 목록

> -> `REF-process-share-owner.md` / 현재 -> `CURRENT.md` (2026-10-08 `history/process-share-display.md`에서 분할)

## 2026-10-08(6) 공유-e-3 구조안·시안 확정 + 공유 표시 색 청록
- 공유-e-3a/3b/3c 구조안 제시, 사용자 1~5 모두 제안대로 → 눌러 보는 시안 작성. 사용자: 포인트 색은 "공유 흐름 시안"의 청록이 마음에 듦, **공유 관련 표시만** 청록으로
- 코드: `plugins/vuetify.ts`에 `share`·`on-share` 등록, `TileFrame.vue`·`FolderTileBody.vue`·`ShareDialog.vue`의 `info` 사용 교체. `vue-tsc` 통과 + 격리 화면 확인. 시안 두 개 같은 링크에서 청록으로 갱신
- presence 질문 정리: 접속 중 유저 상태는 공유-f(SH6) — REST 조회, 폴더 목록 O4는 계정 단위 "지금 접속 중 / 마지막 접속". 지금은 미구현

## 2026-10-08(7) 공유-e-3a(공유 목록 창) 코드·격리 확인
- `share.api.ts`(목록·해제·훅) + `ShareDialog.vue`(목록 단계) + `FolderTileBody.vue`(O1 숫자). `vue-tsc` 통과 + 격리 실행 11개 통과(서버 변경 없음). 구현 메모 → `REF-process-share-display.md`

## 2026-10-08(8) 공유-e-3b(행 공유 개수) 코드
- 서버: `CountDirectShares` 쿼리 + 폴더 목록 응답 `shareCount`(내 소유 폴더만, 직접 건 공유만, 차단 제외). 프론트: 행 칩(`share` 색) + 공유 변경 시 목록 재조회. go build/vet·vue-tsc 통과, 격리 실행 확인은 아직(커밋 전)

## 2026-10-08(9) 공유-e-3c(내가 공유한 것) 코드
- 서버 `GET /shares/mine`(쿼리 2개) + 프론트 `MySharesDialog`·보드 메뉴 항목. 해제는 기존 DELETE 재사용. go build/vet·vue-tsc 통과, 확인은 3b와 함께 한꺼번에(커밋 전)
