# INDEX — PgUp·PgDn 버튼 (마감)

> 2026-10-19 파트(PgUp·PgDn 버튼) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
active 터미널 타일에서 vim·less 쪽 이동용 버튼. 동작-가(키 전송): PgUp `\x1b[5~` / PgDn `\x1b[6~`를 PTY로(특수 키 줄 Ctrl·Alt가 걸려 있으면 `\x1b[5;n~`). `TerminalTileBody.vue` 한 파일. 타입 검사 + 사용자 브라우저 확인

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- `REF-term-page-keys.md`(터미널 타일 PgUp·PgDn 버튼 — active일 때 키 전송) / `history/term-page-keys.md`

## HISTORY 색인 행 (원문)
### `history/term-page-keys.md` — 터미널 타일 PgUp·PgDn 버튼 (2026-10-17 신설)
- 2026-10-17 동작-가(키 전송) + 본문 오버레이, PG1(버튼 위치) 오른쪽 아래·PG2(키보드 열림) 숨김 확정(코드 없음)
- 2026-10-17(2) 페이지-1(버튼 표시)·페이지-2(키 전송) 구현, `vue-tsc` 통과
- 2026-10-19 페이지-3(확인) 사용자 확인 -> 마감·커밋

## 이 파트에만 해당하는 결정
- PG1(버튼 위치) = 본문 오른쪽 아래 / PG2(키보드 열림) = 숨김(그동안은 특수 키 줄 `Ctrl` + `F`/`B`) / 전체보기 중에도 숨김(본문 `inert`)
