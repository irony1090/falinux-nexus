# INDEX — process 살려 두기 (마감)

> 2026-10-22 파트(process 살려 두기) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
worker·supervisor를 다시 띄워도 터미널·화면이 유지되게 함. master에 바로 한 커밋.
- 안 가(process마다 보관 프로세스): worker가 자기 실행 파일을 `worker -hold <소켓>`(`setsid`)으로 띄우고, 보관 프로세스가 PTY master + 화면 ring을 쥠. worker는 소켓으로 붙음
- K1(worker 목록)~K7(끝난 뒤 기다림) 결정, 살리기-1(보관 프로세스)~7(supervisor 주기 삭제) 구현
- 확인: 격리 40/40 -> test1(8011, `build/keep`, `nexus_keep`) 시험 배포·사용자 확인 -> test3(8013, `nexus_build`) 재배포-1(대상 찾기)~5(확인)·사용자 "동작 확인"
- 부작용 발견: 재배포-3(새 것 띄우기)을 Claude Bash 도구로 해서 `CLAUDE*` 변수 상속 -> `REF-build.md` "Claude가 띄운 빌드본"

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- `REF-process-keep.md`(process 살려 두기) / `history/process-keep.md`

## HISTORY 색인 행 (원문)
| `history/process-keep.md` | process 살려 두기(2026-10-22 신설, 마감) | 10-22 조율 K1~K7, (2) 살리기-1~7 구현·격리 40/40, (3) test1 시험 배포, (4) test3 재배포·사용자 확인, 마감·커밋 |

## 이 파트에만 해당하는 결정 · 남은 것
- 다른 파트에서도 지킬 결정(process 영속성·K4(기본 종료 동작))은 MEMORY에 남김
- 반영 때 supervisor·worker·웹을 같이 바꿀 것(옛 worker는 off를 안 보냄) → `REF-process-keep.md` "미확인 · 주의"
- 미확인: K7(끝난 뒤 기다림) 10분 경과 경로 · K6(Windows) 실행 → `REF-process-keep.md` "미확인 · 주의"
- 마감 뒤 정리(2026-10-23 사용자와 같이): test1 내릴지 · `build/iutf` 삭제 · 시험 계정 `smoke…` 정리 → `CURRENT.md`
- BACKLOG: worker "전부 종료" 수단 · 같은 메인키 보관 소켓 폴더 공유 · PTY 환경에서 `CLAUDE*` 빼기
