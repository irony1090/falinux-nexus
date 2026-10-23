# INDEX — 타일 종류 정리 (마감)

> 2026-10-23 파트(타일 종류 정리) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
타일 종류가 늘어도 분기를 덧붙이지 않게, 여러 종류가 같이 쓰는 로직을 `type` 비교 대신 능력(능력-1(루트 가능)·2(자식 가능)·3(process 담음))으로 묻게 바꿈. 동작 변경 없음.
- 정리-1(종류 명세)~7(확인): 서버 `tileKind.go` `tileKinds` / 프론트 `tileKind.util.ts` `TILE_KINDS` 표, `TileFrame` = 공통 틀 + 종류 컴포넌트(`tileComponents.ts`)
- 확인: 옛·새 프론트 헤드리스 비교 전부 같음(2026-10-22) -> 사용자 브라우저 확인 OK + 공유받은 타일·볼 수 없는 타일·모바일 키보드 타일 띄우기(2026-10-23)
- 커밋: 브랜치 `refactor/tile-kinds`(`79b292b`) -> 2026-10-23 master `--no-ff` 머지. 첫 새 종류 = 장비 타일(`device`, 파트(worker 폴더 타일))

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- `REF-tile-kinds.md`(타일 종류 정리, "새 종류를 더할 때" 절차) / `history/tile-kinds.md`

## HISTORY 색인 행 (원문)
| `history/tile-kinds.md` | 타일 종류 정리(진행 중, 2026-10-22 신설) | 10-22 구상, 파트 3개 나누기 승인 / (2) 정리-1~7 구현, 옛·새 프론트 비교 전부 같음 / (3) 브랜치 `refactor/tile-kinds` 커밋, 브라우저 확인 대기 |

## 이 파트에만 해당하는 결정 · 남은 것
- 다른 파트에서도 지킬 결정("타일 종류 = 능력으로 묻는다", 두 표를 같게)은 MEMORY에 남김
- 종류 비교가 남은 곳(폴더·터미널 전용 컴포넌트, SQL `type == "terminal"` 3곳)은 의도된 것 → `REF-tile-kinds.md` "구현"
