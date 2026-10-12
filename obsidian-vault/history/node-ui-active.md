# history/node-ui-active — active 타일 표시 · 위치 번호 단축키 이력

> 2026-10-12 신설. 재사용 지식 → `REF-node-ui-active.md` / 확정 표시 방식 → `REF-node-ui-overview.md` "active 타일 표시" / 현재 → `CURRENT.md`

## 2026-10-12(6) 사용자 확인 → 파트(active 타일·원격 배지) 마감
- 사용자가 5050 재시작 후 브라우저로 확인 "확인했어" → 코드·vault 커밋(파트 단위)

## 2026-10-12(5) RA0(범위) = 폴더까지 → 원격 active 구조 확정·구현 (커밋 전)
- 사용자 "UI를 볼 수 있을까" → 캔버스 시안 https://claude.ai/artifact/Nraz8RdacfeXkVJWGsW9nu (Main·Detail)
- 사용자 "최종적으로는 폴더 타일까지" → RA3(받는 범위·경로)를 "접속 중 다른 계정마다 판정 일괄 → `TILES:<userId>`"로 바꾸고 RA7(처음 받기 = `GET /tabs/active`) 추가 → 구조 제시 → "진행해줘"
- 구현: 서버 `active.go`(신규)·`tabs.go`·`subscribe.go`·`protocol`(`TAB:ACTIVE`·`ACTIVE:UPDATE`)·쿼리 `ListNodePermissionMatrix` / 프론트 `active.api.ts`·`remoteActive.store.ts`·`ActiveBadges.vue`(신규)·`tileGrids.store`(보내기)·`TileFrame`·`index.vue`·`vuetify.ts`(`acct1`~`acct6`)
- 확인(격리: 임시 DB `nexus_active_tmp` + 5051 + worker `active-dev` + 프론트 3001, 끝나고 끄고 DB 삭제): `pw/remote.mjs` 11/11 — 터미널 배지 / 볼 수 없는 계정 없음 / 폴더로 옮김 / 루트 = 없음 / 새 탭 처음 받기 / 같은 계정 다른 탭에 내 배지 없음 / 공유 취소 → 사라짐 / 다시 공유 → 돌아옴 / 탭 닫힘 → 사라짐
- 테스트 함정: 공유 폴더를 받는 쪽이 "새 타일"로 열면 그 계정 active가 그 폴더가 됨 → 소유자 쪽에도 반대 방향 배지가 생김(정상). "배지 없음"을 단정할 땐 누구 배지인지까지 볼 것

## 2026-10-12(4) 로컬 active 사용자 동작 확인 완료 → RA 결정으로
- 사용자 "동작 확인. 다음 작업으로" → 활성-1(상태·누르기)~활성-5(확인) 끝. 다음 = RA1(맞추는 기준)~RA6(순서)

## 2026-10-12(3) AT4(단축키 조합) = `Ctrl+Alt+숫자`로 변경
- 사용자 "동작은 잘 하는데 ctrl+alt+숫자로 바꾸자" → `tileShortcut.hook.ts` 조건 `ctrlKey && altKey`, Ctrl·Alt 어느 쪽을 떼도 두 자리 대기 끝. 헤드리스 다시 14/14(`Alt+2`만은 통과·active 안 바뀜 추가)

## 2026-10-12(2) AT 확정 + 활성-1(상태·누르기)~활성-5(확인) 구현 (커밋 전)
- 사용자 "진행해줘, 간단하게 테스트까지" → AT1(상태 위치)~AT6(이동할 때 같이 할 일) 제안대로 확정. 구현 내용 → `REF-node-ui-active.md` "구현"
- 설계 변경 1개: AT4(단축키 조합)의 "xterm 키 핸들러에서 같은 함수"는 불필요 — window capture에서 `stopPropagation`하면 xterm textarea에 키가 안 감
- 확인: 사용자 5050·3000 위에서 새 계정 `active-<시각>` 2개(폴더 11개씩, 타일 12개) + 헤드리스(`pw/active.mjs`, 이번 세션 scratchpad, playwright-core는 12185a69 세션 `pw/node_modules`를 심볼릭 링크) **13/13 통과**: 처음 흐림 없음 / 연 타일 active·reveal / 누르기·흐림 0.7·outline / `Alt+9` 포커스 / `Alt+1,2`=12 / Alt 떼면 바로 / 1초 대기 / 없는 19 → 9 / keydown `preventDefault` / 전체보기 강조·단축키로 나가기 / 전체보기 누르기 / active 닫으면 해제 / `Alt+A` 통과
- 테스트 함정: playwright `press('Alt+2')`는 Alt keydown도 보내므로 keydown 기록에서 `e.key === 'Alt'`를 빼고 셀 것 / 먼 Grid 스크롤은 0.7초 대기로 모자람(1초 이상)

## 2026-10-12(1) 구조안 제시 (코드 없음)
- 사용자 "active 타일 표시 및 키보드를 이용한 빠른 이동을 먼저" → 새 파트로 착수. 지금 코드 조사(`posOf`·`reveal`·`TileStrip` keydown·xterm 키 핸들러, ⑮ 특수 키 줄은 아직 코드 없음) → AT1(상태 위치)~AT6(이동할 때 같이 할 일) + 활성-1~5 제시
- 이어서 사용자 제안 "다른 계정의 active 위치를 한 글자 + 색으로" → RA1(맞추는 기준)~RA6(순서) 제안 추가(타일 트리가 계정마다 따로라 터미널 = uid, 폴더 = nodeId로 맞춤)
