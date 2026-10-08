# REF — 공유-f 접속 상태(presence, O4)

> 2026-10-08 `REF-process-share-display.md`에서 분할. 결정 SH1~SH10 -> `REF-process-share.md` / 이력 -> `history/process-share-presence.md` / 표시 전반 -> `REF-process-share-display.md` / 현재 -> `CURRENT.md`

**상태(2026-10-08): f-1~f-5 완료(격리 확인 통과, 사용자 직접 확인 병행), 커밋 전.**

## 공유-f(presence) 구현 메모 (2026-10-08, 커밋 전)
결정: O4의 "마지막 접속" = **서버 메모리만**(탭 등록부, 재시작하면 비어 있음, DB 컬럼 없음 — 사용자 2026-10-08).
| 단위 | 파일 | 변경 |
|---|---|---|
| 공유-f-1(탭 정보 확장) | `tabs.go`·`subscribe.go` | `browserTab.userAgent`(연결 때 헤더, 다시 붙으면 새 값) · `lastSeen`(계정의 연결된 탭이 0이 된 시각) · `tabStatusesOfUser`·`userPresence` |
| 공유-f-2(process presence) | `presence.go` 신설 | `GET /processes/:uid/presence` — 보기 이상이면 호출 가능, 볼 수 있는 계정 전부, 계정 > 세션(sid 해시 8자) > 탭(connected·viewing·sizeOwner·since). **탭 id·sid 원문은 응답에 안 넣음** |
| 공유-f-3(계정 접속) | `presence.go`·`shares.sql` | `GET /users/presence?ids=` — 내가 공유를 건 계정만(`ListShareTargets`), 그 밖은 조용히 빠짐 |
| 공유-f-4(프론트) | `presence.api.ts`·`presenceView.util.ts`·`PresenceDialog.vue`(신설)·`TileFrame.vue`·`ShareDialog.vue` | 터미널 머리 줄 `mdi-account-multiple-outline` 버튼 -> 접속 상태 창(열린 동안 5초 재조회) / 공유 목록 창(O2) 행에 접속 점 + "접속 중 / 마지막 접속 N분 전 / 접속 기록 없음" / UA 파싱 유틸 |
- 빌드: go build/vet · vue-tsc 통과. 확인할 것(공유-f-5): 계정 둘·브라우저 둘 — 접속/끊김 반영, "보는 중"(다른 보드를 보면 연결됨), 크기 소유 표시, 끊고 60초 안에 "끊김(보관 중)", 마지막 접속 시각, 공유 안 건 계정 id는 빠짐, 모바일 390px

## 공유-f-5 확인 결과 (2026-10-08, 격리 실행: 임시 DB + 5051 + 격리 프론트 3001 + worker)
- API 25개 통과: 공유 전 B 조회 404 · 소유자 먼저·접속 중 먼저 정렬 · 세션 2개(UA 원문) · 탭 id·sid 원문 비노출 · exec한 탭 `sizeOwner` · 기본 보드 "보는 중" / 다른 보드(`TAB:VIEW`) 보면 "연결됨" · 받은 사람도 전체 조회 · 무관 계정 404 · O4 공유 안 건 쪽·무관 id는 빈 결과/빠짐·중복 id 1번 · 끊으면 오프라인 + `lastSeen` + 탭 보관 중 · 같은 탭 id로 재연결하면 온라인·탭 1개
- 화면(헤드리스): 접속 상태 창(소유자 "보는 중 · 크기 소유" / 받은 사람 "연결됨") · 공유 목록 창 접속 점("접속 기록 없음" -> "접속 중") · 390px 가로 넘침 없음 + 행 칩(3b) 표시
- 참고: 폴더 공유를 process 실행 뒤에 걸면 받는 쪽엔 폴더 타일만 생기고 터미널 타일은 안 생겨 "보는 중"이 아닌 "연결됨"으로 나옴(의도된 동작, 터미널 타일은 실행 시점의 볼 수 있는 계정에만 — SD2). 60초 보관 만료 후 탭이 사라지는 것과 `lastSeen` 값 자체의 시각 정확도는 안 봄(`tabGrace` 대기)

