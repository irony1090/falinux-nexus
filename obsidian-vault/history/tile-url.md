# history — 보드 URL 바인딩

> → `REF-tile-url.md` / 현재 진행 → `CURRENT.md`

## 2026-10-19
- 사용자: "F1·F2 고치기 전에, 새로고침 대비해서 포커스·타일 위치를 URL에 바인딩할 수 있나" -> 조사: 보드는 `/w/:hash`, `TileWorkspace`에 일회용 `?focus=`·`?tab=`(처리 뒤 지움) -> U1(넣을 것)~U4(없는 타일) 제시 -> 사용자 "u1 나도 진행"
- 구현: `tileUrl.hook.ts` 신규 + `tileGrids.store.ts`(`anchorId`·`restore`·`columnsOf`) + `TileWorkspace.vue` watch 교체. `vue-tsc` 통과
- 확인: 로그인 폼 기본값이 주석 처리돼 있어 실제 계정 대신 격리 supervisor(5051, 파일 전송 확인 때 띄운 것)에 계정을 만들고 격리 프론트 3001로 헤드리스 20/20(첫 실행 D5 1개 실패 = smooth 스크롤 + 300ms 전에 읽은 타이밍, 기다리게 고쳐 통과)

## 2026-10-19(2)
- 사용자 "url 바인딩 확인" -> 파트(URL 바인딩) 마감, `INDEX-tile-url.md`로 접음, 코드 3파일 + vault 한 커밋
