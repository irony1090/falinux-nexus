# INDEX — URL 바인딩 (마감)

> 2026-10-19 파트(URL 바인딩) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
새로고침해도 보던 화면으로 돌아오게 보드 URL에 `?focus=<tileId>`(active 타일)·`?view=<tileId>`(보던 칼럼의 타일)를 계속 맞춤. `hook/tileUrl.hook.ts`(신규) + `tileGrids.store.ts`(`anchorId`·`restore`·`columnsOf`) + `TileWorkspace.vue`(일회용 `?focus`·`?tab` watch 교체). 격리 헤드리스 20/20 + 사용자 브라우저 확인

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- `REF-tile-url.md`(보드 URL 바인딩 — `?focus=`·`?view=`로 새로고침 복원) / `history/tile-url.md`
- 확인 스크립트 명세 → `REF-test-scripts-1014.md` "2026-10-19 보드 URL 바인딩"

## HISTORY 색인 행 (원문)
### `history/tile-url.md` — 보드 URL 바인딩 (2026-10-19 신설)
- 2026-10-19 U1(넣을 것)~U4(없는 타일) 확정·구현(`tileUrl.hook.ts`), 격리 헤드리스 20/20(커밋 전)
- 2026-10-19(2) 사용자 브라우저 확인 -> 마감·커밋

## 이 파트에만 해당하는 결정
- U1(넣을 것) = focus + view(Grid 번호는 화면 크기마다 달라 안 씀) / U2(쓰는 방식) = `router.replace` + 300ms 지연, 전체보기 중엔 안 씀 / U3(복원 때 포커스) = PC(`pointer: fine`)만 / U4(없는 타일) = 지금 상태로 덮음
- view 없는 `?focus=`(목록 "보드로 이동"·알림 `focusPath`) = 예전처럼 `activate`(폰에서도 포커스)
