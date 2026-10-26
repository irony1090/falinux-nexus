# history — 브라우저 API `/api` 접두어

> → `REF-api-prefix.md` / `CURRENT.md`

## 2026-10-26
- 집 서버에서 사용자: 내 장비 상한(3) 차서 장비 추가 버튼 비활성 -> 새로고침하니 JSON 화면. 원인 = 프론트 `/devices` ↔ API `GET /devices` 충돌(상한은 설계대로, 회원-Q6·Q17). 안 A(프론트 경로 바꾸기) 추천 -> 사용자 "앞으로도 생길 것" -> B 방향, 정확히는 supervisor 코드에서 전부 `/api`
- 조율: `/worker` 그대로 · `/subscribe` -> `/api/subscribe` · `/test`(+`Test()`) 지움 · 코드는 Claude · 브랜치 `refactor/api-prefix`
- 구현 api-1~4 -> 확인 api-5 14/14(REF). 커밋 전
- 같은 자리에서 다음 기능 요청: 내 장비 판의 열린 장비 타일 supervisor 주소 일괄 변경 -> `REF-device-move.md`(api-prefix 다음 순서)
- 이동-Q1~Q6 확정(열린 장비 타일 · instance 셀렉트 · 관리자도 · 푸터 액션 자리 = 계정 알약 왼쪽, 비활성으로 보임, 2개 이상이면 편집 아이콘으로 접기) -> "진행해줘" -> 서버 `deviceMove.go`·프론트 창·`FootActions` 구현 -> 격리 23/24(실패 1 = 3개 상한 함정, BACKLOG) + 바로 옮기기 복귀 OK. 같은 브랜치, 커밋 전
- 사용자: `build/`에서 take_process 관련만 남기고 정리 요청 -> 목록 확인 중
