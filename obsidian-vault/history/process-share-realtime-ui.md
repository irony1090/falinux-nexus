# history/process-share-realtime-ui — 공유-d 실시간-0(터미널 UI) 이력

> 재사용 지식 → `REF-process-share-realtime-ui.md` / 공유-d 전체 이력 → `history/process-share-realtime.md` / 현재 진행 → `CURRENT.md`

## 2026-10-07(10) — SD 전부 확정 + 실시간-0 구현 계획 제시 (코드 없음)
- 사용자 "추천대로 확정" → SD1(출력 구독 범위)·SD2(실행 때 터미널 타일을 넣을 계정)·SD4(`SHARE:CHANGE` 토픽)~SD9(경로 줄 멈춤) 제안 그대로 확정
- `REF-process-share-realtime.md`가 11k자 → SD3(실행 중 터미널 치우기) 전문과 실시간-0을 이 파일 쌍으로 분할
- 지금 코드 확인 → 구현 계획(파일별) + 확인할 것 4개 제시, 사용자 확인 대기
- 사용자 "제안대로 진행" → 구현 계획 + 확인할 것 4개(기본 탭 전환 시점·종료하고 닫기 제한 시간 10초·크기 우선권 그대로·남이 실행한 것 안내 줄은 실시간-4로) 확정, 코드 착수

## 2026-10-07(11) — 실시간-0 코드 + 확인
- 서버 `tile.go`(`Minimized`) + 프론트 `tile.type`·`tileTree.store`(`setMinimized`)·`tileGrids.store`(`order`/`flat`/`terminals`/`rootTab`)·`processTerm.store`(`STATUS_PILL`)·`terminalKill.hook`(새)·`TileFrame`·`RootTerminalList`/`TerminalListRow`(새)
- `go build`·`vet`·`vue-tsc`·`vite build` 통과 → 격리 실행(임시 DB `nexus_rt0_tmp` + 5051 + 프론트 3001) 헤드리스 24/24 통과, 스크린샷으로 머리 줄·루트 탭 확인. 임시 DB 삭제·격리 프로세스 종료
- 사용자 화면 확인 "현재 완성된 부분까지는 문제 없어 보임" → 실시간-0 사용자 확인 완료. 질문 "스크립트 행 실행 표시는 아직?" → 실시간-6a(스크립트 행)로 남음(모든 보드 위치 = 서버 응답 필요, 실시간-4(exec 타일) 뒤) 답함

## 2026-10-08(19) — 실시간-6a(스크립트 행) 코드 + 격리 실행 확인
- 구조안 승인 → `ListVisibleLiveProcessesInFolder` + `GET /nodes/:id/runs` + 프론트(`useScriptRuns`·`ScriptRunList`·`confirmKillUid`·행 칩). build/vet·vue-tsc 통과, REST 11/11 + 헤드리스 화면 확인(커밋 전). 칩 스타일 우선순위 수정

## 2026-10-08(20) — 실시간-6b(공유 반영) 코드 + 헤드리스 확인
- 구조안 승인 → `throwCatch` status + `useShareChange`·`terminalRestore` recheck·`gone`·`TileFrame` 배너 + 경로 멈춤(`parentVisible`)·루트 목록 소유자 칩. vue-tsc 통과, 헤드리스 2계정(공유 풀기·다시 걸기·폴더 삭제·경로·소유자 칩) 통과(커밋 전)

## 2026-10-08(22) 루트 타일이 공유 폴더 안에서 막힘 수정
- 발견(사용자, test1 계정): 루트 타일을 공유받은 폴더 안으로 열었다가(`navigate`, 트리에 `nodeId` 저장) 공유가 풀리면 "볼 수 없음" 배너만 뜨고, 루트 타일엔 닫기가 없어 빠져나올 수 없음. 실시간-6b 확인은 "새 타일로 연 경우"만 봐서 놓침
- 수정(안 A): `TileFrame.vue` 배너에 루트 폴더 타일일 때만 "기본으로 돌아가기"(`navigate(tileId, null)`). 서버·데이터 변경 없음. type-check 통과 + **사용자 화면 확인 통과**(test1에서 버튼으로 돌아옴, 커밋 전). 안 B(서버가 `reconcileAccess`에서 루트 `nodeId` 되돌림)는 안 함

