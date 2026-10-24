# INDEX — 타일 머리 줄 버튼 (마감)

> 2026-10-24 파트(타일 머리 줄 버튼) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
타일 머리 줄·경로 줄 버튼을 높이를 꽉 채운 알약 버튼으로 통일. 브랜치 `feat/tile-head`(master `bc7caeb`에서).
- 시안: 캔버스 https://claude.ai/artifact/496HgePkvczPvhc3FPtoJG — 안 A·B·C 중 C(끝 블록 강조) -> 조정 판 `C2b`(안 C′ — 구분② 경계선 색)로 확정
- 단계-1(공통 버튼 모양): 전역 `styles/tile-buttons.scss` — `.tile-btn`(알약) · `--text` · `--tonal`(`--c`) · `.tile-btn-sep` · `.tile-btn-end`(+`--primary`) · `.tile-btn-row-end/-start`(타일 끝에 닿는 면만 직각)
- 단계-2(머리 줄 틀): `TileFrame.vue` 버튼 자리 묶음, 창 묶음 = 끝 블록(active = primary), active 경계선 = primary 40%
- 단계-3(종류별 머리 버튼): 뒤로·앞으로 / 크기 가져오기·접속 상태 / RUN 초록 알약
- 단계-4·5(폴더·장비 경로 줄): [경로·새로고침] + [도구·끝 블록("새로"·"올리기")], 경로 자리 최소 160px보다 좁으면 두 줄, 입력 중 도구 숨김
- 관찰 ①: 머리 줄 보조 글자(출처·인스턴스·터미널 경로·실행자)가 이름보다 먼저 줄어듦
- 확인: 격리 화면(PC 1280·900, 폰 390) + **사용자 test1(8011 개발 서버 -> test3 API)에서 확인 OK**(2026-10-24). 배포 = test3 `build/ctrl_panel/web`만 교체(백엔드 변경 없음)
- 커밋: 브랜치 `feat/tile-head`에 파트 마감 커밋 1개 -> master `--no-ff` 머지

## REF / history 목록
- `REF-tile-head.md`(바라는 것 · 머리-Q1~Q4 · 시안 · 고른 안·조정-1~4 · 큰 계획 단계-1~6 · 구현 · 확인 · 함정) / `history/tile-head.md`
- 확인 스크립트 명세 → `REF-test-scripts-1023.md` "머리-확인"

## MEMORY 인덱스 줄 · HISTORY 색인 행 (원문)
- MEMORY: 타일 머리 줄 버튼 `REF-tile-head.md`(진행 중 파트)
- HISTORY: | `history/tile-head.md` | 타일 머리 줄 버튼(2026-10-24 조율 시작, 진행 중) | 10-24 구현(단계-1(공통 버튼 모양)~5(장비 경로 줄))·격리 확인, 사용자 확인 대기 |

## 이 파트에만 해당하는 결정 · 남은 것
- 머리-Q3(높이) 28px 유지 · 조정-2(끝 블록 직각) = 버튼은 알약, 타일 왼쪽·오른쪽 끝에 닿는 면만 직각(윗변 제외) · 조정-3(위아래 구분) = 기존 경계선을 옅은 파랑 · 조정-4(새로고침 거리) = 경로 옆
- 함정: 두 body의 scoped 버튼 리셋이 전역 `.tile-btn`을 덮음 -> `:not(.v-btn, .tile-btn)` / `.path > *` 최소 폭이 `/`에도 걸림 -> `.sep` 0
