# history — 파트(영상 Range)

> 2026-10-29 신설(`history/view-tile.md`에서 분리) ↔ `REF-video-view.md` / → `CURRENT.md`

## 2026-10-29 파트(영상 Range) 착수 전 조율
- 보드 `fa24-PID` worker 교체(REF-build) 뒤 Q1(한 타일의 뜻)~Q4(같은 계정 여러 창) 답: Q1 = b(같은 파일당 한 타일) · Q2 = 이미지와 같은 규칙 · Q3 = 다시 접속 때 보던 자리 유지 · Q4 = 재생 위치 동기화 -> 방식 질문 → REF-video-view "결정"
- Q1 재확인(계정 전체 한 타일, 다시 열면 이동 + 포커스) · Q4 = S1(이어보기)로 결정(S2(실시간 따라가기) 접음) -> 층별 계획 제안, 확인 대기
- P1 = 타일 닫으면 위치 지움 · P2 자세히 요청 -> "트리 쓰는 5곳" 표현 정정(보기 타일 추가는 PUT 하나), A/B/C 비교 제시
- 트리 저장 경로 설명(① 프론트 동작 PUT / ② 서버가 넣는 타일 / ③ 서버가 바꾸는 동작), C면 영상 열기가 ①->② · 사용자 질문 "모든 타일을 서버가 관리하면?" -> 장단점 답, 영상 파트는 C만 추천(확인 대기)
- (가) 확정("네 의견에 따르자") -> MEMORY·REF-decisions "타일 추가 주체" -> 1단계 계획 수정본 -> 사용자 "진행해"

## 2026-10-29 1단계(영상 재생) 구현·격리 확인
- 브랜치 `feat/video-view`. 서버 = `serveVideo`(ServeContent + readSeeker) · `openVideoTile` · `checkVideoTiles` / 프론트 = `VideoViewBody` · `viewFile` 영상 분기 · `applyServerTile` 넓힘 · 아이콘
- 빌드·`go test`(router, `TestCheckVideoTiles` 추가)·`vue-tsc` 통과 → 격리 확인 curl·API 15/15·화면 21/21 → REF "1단계 격리 확인"

## 2026-10-29 2단계(S1(이어보기)) 구현·격리 확인
- 사용자 "확인 이어서 진행해줘" = 1단계 확인 끝 + 2단계 진행으로 해석(밝힘)
- `video_positions`(00015) · 위치 API · PUT 닫기 때 삭제 · sweep 안전망 · `VideoViewBody` 저장·복원 → API 11/11, 화면 확인 중 새로고침 복원 버그(명부 늦게 도착 -> `<video>` 다시 만듦) 찾아 고침 → REF "2단계 격리 확인"

## 2026-10-29 test3 임시 빌드
- 사용자 "임시 빌드 표기만 해서 빌드, 8013 실행은 내가" -> 태그 `VIDEO_VIEW_TMP-261029`(env 첫 줄 2개) + `scripts/build.sh video_view_tmp` + 폴더에 `TEMP-BUILD.txt`(브랜치·기준 커밋·미커밋 표시)
- 사용자 test3 확인: "생각보다 느리지 않아 괜찮다" -> nginx `proxy_buffering` 손대지 않음

## 2026-10-29 파트 마감
- 사용자 "마감. vault 정리 후 커밋 머지, 정식 빌드는 내가 직접, 세션 재부팅 준비" -> 격리 환경 끔(DB `nexus_vv_tmp` 남김) -> vault 정리 -> `feat/video-view` 커밋 + master 머지
