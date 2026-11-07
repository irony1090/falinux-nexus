# history — 공용 타일 틀 · 메뉴 줄 위치/숨김 · 설정 (2026-11-07 신설)

> → `REF-tile-frame.md` · `CURRENT.md`

## 2026-11-07
- 사용자: 노드·장비 디렉토리 타일 경로 줄 숨김이 다름 -> 확인(노드 = sticky 없음, 장비 = `sticky; top: 0`). 루트 탭 `[파일] [터미널 N]`은 둘 다 스크롤 밖이라 안 숨겨짐, 이번 작업에서 제외
- 요구 조율 Q1(위/아래 고정의 뜻)~Q10(루트 탭) 합의 -> `REF-tile-frame.md` "합의된 결정". 타일 = 보드 형태 페이지의 칸에 들어가는 것(로그인·가입 제외)
- 영향 조사: 틀 두 벌(`TileFrame`·`MbTileFrame`) + 직접 `.mb-tile`(목록 패널), 스크롤 요소 의존 코드(`loadMore`·`cellDrag`·끌어 놓기·`.xfer`) -> 계획 단계-1(공용 틀 리팩터링)~단계-4(오른쪽 메뉴 통일) 작성. 열린 질문 Q11(메뉴 줄 slot 자리)~Q13(목록 패널)
- Q11(메뉴 줄 slot 자리) = `TileScroll` · Q12(관리 타일 아래 줄) = 아래 고정 · Q13(목록 패널) = `TileShell`만 — 사용자 확정. 계획 확정, 착수 대기
- 단계-1(공용 틀 리팩터링) 구현(`refactor/tile-shell`): `TileShell`·`TileScroll`·`elRef` 새로, `TileFrame`·`MbTileFrame`·목록 패널 2개 -> `TileShell`, 폴더·장비 본문 + 관리·updater 상세 9개 -> `TileScroll`. `#foot` 칸은 안 만듦(이유 → REF). 확인 = type-check·build·run.sh 125 PASS(FAIL 2 = 옛 기대값, 무관)·보드 전후 비교 같음 → `REF-tile-frame.md` "단계-1 결과"
- 단계-2(메뉴 줄 위치·숨김) 구현: sticky 방식(사용자 확인) · `TileScroll` `barAt`/`barHide`/`barPin` + 임시 `tileBarAt` · 로딩 줄·테두리 이동 · `.xfer` 변수 · 자동완성 위로. 헤드리스 위/아래 0 FAIL, run.sh 통과(낡은 2건 제외) → `REF-tile-frame.md` "단계-2 결과"
- Q14(로그아웃 때 설정) = 지워짐(userPrefs 그대로) 사용자 확정. 사용자 화면 확인용 개발 화면 띄움(supervisor 8012 `go run . -env dev.env` + vite 8011 `VITE_API_URL=https://test2.falinux.com`). dev worker(`cmd/worker/dev.env`)는 API 키 없어 등록 거절 -> 재시도 중, 끄는 명령은 권한 규칙에 막힘
