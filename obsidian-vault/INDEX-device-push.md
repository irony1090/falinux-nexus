# INDEX — 장비 상태 push (마감)

> 2026-10-27 파트(장비 상태 push) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
브랜치 `feat/device-push`(master `4cf6855`에서).
- 계기: 장비 접속 현황이 10초 REST 폴링이라 반영이 느려 불편 → 소켓 알림으로
- 구현: 접속·끊김·라벨·등록/삭제·주인 변경·명부 인스턴스 지우기 → `DEVICE:CHANGE {mainKey}` 신호(받는 사람 = 접속 중인 관리자·주인·그 장비 폴더를 볼 수 있는 계정) → 보드·내 장비·관리자 화면이 300ms 묶어 다시 조회. `SHARE:CHANGE`·소켓 재연결 때도 재조회. 폴링 60초 안전망. 소켓 connect를 `App.vue`로 옮김
- 확인: 격리 받는 사람 10/10 · 보드 0.4초 · 관리 페이지 1.5초 안 + 사용자 직접 확인(격리 3001, 계정 4개, worker 끄고 켜기) OK(2026-10-27). 배포 안 함
- 커밋: 브랜치 `feat/device-push`에 파트 마감 커밋 1개 -> master `--no-ff` 머지

## REF / history 목록
- `REF-device-push.md`(결정 · 구현 표 · 확인) / `history/device-push.md`
- 실시간 push 일반 = `REF-realtime.md`(장비 상태 이벤트 절이 이 파트를 가리킴)

## MEMORY 인덱스 줄 · HISTORY 색인 행 (원문)
- MEMORY: **진행 중 파트**: 장비 상태 push(2026-10-27, 브랜치 `feat/device-push`) → `REF-device-push.md`
- HISTORY: | `history/device-push.md` | 파트(장비 상태 push, 진행 중) ↔ `REF-device-push.md` | 10-27 구현·격리 확인(받는 사람 10/10), 사용자 확인 대기 |

## 이 파트에만 해당하는 결정 · 남은 것
- 알림 = 신호만, 보이는 범위 판단은 기존 API · 받는 사람 쿼리 `FilterDeviceWatchers`는 `ListVisibleWorkerInstances`와 같은 범위여야 함(바꾸면 같이)
- 사용자 접속 표시(presence)는 5초 폴링 그대로(범위 밖)
- 관리 페이지만 연 탭도 서버 "탭"(presence에 접속으로 보임)
- 소켓 재연결 때 재조회 = 코드만, 따로 확인 안 함
