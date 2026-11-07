# INDEX — 업데이트 app (마감)

> 2026-11-07 밤 세션 끝에 MEMORY·HISTORY가 9.7k자에 닿아 접음(11-06 BACKLOG 제안). **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
10-30 설계 시작 -> 11-02 설계 개정(10-30 점검 결과 반영) -> 11-02 P1(보관·화면)·P2(updater Linux) -> 11-03 P3(배포)·P4(Windows)·Files 타일 개선(단위-가 폴더 트리·선택 삭제 / 시험 실행 / 단위-나 끌어 놓기 / 단위-다 갈아치우기 / 단위-라 실행 권한 이름) -> 11-04 집 서버를 updater로 배포(단위-7(교체 빌드) 후속, 그 뒤 11-05·11-06·11-07 묶음도 updater로). 전부 master 머지
- 다른 파트에서도 지킬 결정(updater 인가·updater 되돌리기·배포 전 필수)은 MEMORY "확정 결정"과 `REF-decisions.md`에 그대로 남김
- 계속 쓰는 것(접지 않음): 집 서버 배포 `REF-home-deploy-updater.md` · 빌드 스크립트 `REF-build-scripts.md`
- 남은 것: `run.sh` 되돌리기 기대값 2건 낡음(`BACKLOG.md`) · (updater 교체용 임시 업로드 폴더 삭제 = 11-07 사용자 확인 "폴더 없음", 끝)

## REF / history 목록
| REF / history | 내용 |
|---|---|
| `REF-update-deploy.md` / `history/update-deploy.md` | 현재 설계: 독립 범용 updater·장비>빌드 릴리스·steps.json·builtin·DB 7테이블·타일 페이지 |
| `REF-update-rules.md` / `history/update-rules.md` | 규칙: 개별 실행·빌드 단위 잠그기·삭제·되돌리기=rollback 단계·설치본 이름 |
| `REF-update-redesign.md` / `history/update-redesign.md` | 11-02 설계 개정 결정 경위·시안 |
| `REF-update-p1.md` ~ `-p4.md` / `history/update-p1.md` ~ `-p4.md` | P1(보관·화면) · P2(updater Linux) · P3(배포) · P4(Windows, 단위-7(교체 빌드) 후속) 구현 지도 |
| `REF-update-files-tile.md` / `history/update-files-tile.md` | Files 타일 개선 결정-1~15·구현 |
| `REF-update-replace.md` / `history/update-replace.md` | 단위-다(갈아치우기) |
| `REF-update-trial.md` / `history/update-trial.md` | 배포 전 단계별 시험 실행 |
| `REF-review-1030.md` / `history/review-1030.md` | 10-30 작업 점검(11-02 코드 머지, 설계 항목은 update-redesign으로) |

## MEMORY 인덱스 줄 · HISTORY 색인 행 · BACKLOG 줄 (원문)
- MEMORY: - **배포·업데이트**: `REF-update-deploy.md`(업데이트 app 현재 설계: 독립 범용 updater·장비>빌드 릴리스 구조·steps.json·builtin·DB 7테이블·타일 페이지·단계 계획) · `REF-update-rules.md`(규칙: 개별 실행·빌드 단위 잠그기·삭제·되돌리기=rollback 단계·설치본 이름) · `REF-update-redesign.md`(11-02 결정 경위·시안 기록) · `REF-update-p1.md`(P1 구현: 파일 지도·계획과 달라진 것·쓰기 순서·키) · `REF-update-p2.md`(P2 전 확정 6개·착수 계획) · `REF-update-p3.md`(P3 계획·구현 지도·계획과 달라진 것·동작 규칙) · `REF-update-trial.md`(11-03 배포 전 단계별 시험 실행: 시험 타일·결정-1~9·시험-1~5 계획) · `REF-update-files-tile.md`(11-03 Files 타일 개선: 폴더 트리·선택 삭제·끌어 놓기·폴더별 합치기/갈아치우기 결정-1~15·시안·구현 때 볼 곳) · `REF-update-replace.md`(11-03 단위-다(갈아치우기) 계획·확인-1~4·구현 결과 다-1~6) · `REF-home-deploy.md`(수동 update.ps1 묶음) · `REF-home-deploy-updater.md`(11-03~ updater 설치 묶음·updater로 교체) · `REF-build.md`(보드 worker 교체)
- MEMORY: - **10-30 작업 점검(11-02, 코드 머지됨)**: `REF-review-1030.md` / 업데이트 app 설계 개정(11-02, 점검 결과의 설계 항목 처리) → `REF-update-redesign.md`
- MEMORY: - **진행 중 파트**: 업데이트 app(11-02 P1(보관·화면) 마감·master 머지, P2(updater Linux) 마감·master 머지, 설계 전부 닫음, P3(배포) 11-03 마감·master 머지, P4(Windows) 11-03 마감·master 머지, Files 타일 개선 11-03 마감·master 머지(단위-가 폴더 트리·선택 삭제 / 시험 실행 / 단위-나 끌어 놓기 / 단위-다 갈아치우기 / 단위-라 실행 권한 이름) → `REF-update-trial.md` · `REF-update-files-tile.md` · `REF-update-replace.md` / 단위-7(교체 빌드) 집 서버 확인은 사용자 답 대기) → `REF-update-p4.md`(Windows 구현 · 단위-7 후속) · `REF-update-p3.md` · `REF-update-p2.md` · `REF-update-p1.md` · 설계 `REF-update-deploy.md` · 규칙 `REF-update-rules.md` · 경위 `REF-update-redesign.md` / 그 전 마감 파트 = `HISTORY.md` "마감 파트"(11-03 다섯 파트 INDEX로 접음: file-view·view-tile·video-view·tile-drag·member-drag)
- BACKLOG: - **업데이트 app 파트 INDEX 접기** · 2026-11-06 제안 · HISTORY가 다시 ~10k에 닿을 때(11-06 9.6k) · 업데이트 app(P1~P4·Files 타일 개선·시험 실행·갈아치우기)을 `INDEX-update-app.md` 하나로 접자고 함, 파트 범위 사용자 답 대기 → CLAUDE.md "색인 2단화"
- HISTORY: | `history/update-deploy.md` | 파트(업데이트 app) 설계·화면(2026-10-30 신설, 10-30 분할) ↔ `REF-update-deploy.md` | 11-02 저녁 시안 확정 · P1(보관·화면) 계획 확정, 코드 착수 전 |
- HISTORY: | `history/update-rules.md` | 파트(업데이트 app) 규칙 — 잠금·삭제·되돌리기·버전(2026-10-30 분할) ↔ `REF-update-rules.md` | 11-04 로컬 updater 시험 · 되돌리기 순서 안 A |
- HISTORY: | `history/update-redesign.md` | 파트(업데이트 app) 설계 개정 결정·시안(2026-11-02 신설) ↔ `REF-update-redesign.md` | 11-02 설계-1~7·9·11·12 합의 + 시안 보정 + 본문 개정 |
- HISTORY: | `history/update-p1.md` | 파트(업데이트 app) P1(보관·화면) 구현(2026-11-02 신설) ↔ `REF-update-p1.md` | 11-02 밤 P1 마감, master 머지(`1f5eda9`) |
- HISTORY: | `history/update-p2.md` | 파트(업데이트 app) P2(updater Linux)(2026-11-02 신설) ↔ `REF-update-p2.md` | 11-02 밤 P2 실기 확인 통과·마감, master 머지(`bd93fa4`) |
- HISTORY: | `history/update-p3.md` | 파트(업데이트 app) P3(배포)(2026-11-02 신설) ↔ `REF-update-p3.md` | 11-03 오전 사용자 확인·로그 보관 정리·master 머지 |
- HISTORY: | `history/update-p4.md` | 파트(업데이트 app) P4(Windows)(2026-11-03 신설) ↔ `REF-update-p4.md` | 11-03 낮 구현·VM 확인 39/39 · 오후 집 서버 설치 · 마감·머지 |
- HISTORY: | `history/update-files-tile.md` | 파트(업데이트 app) 후속 — Files 타일 개선(2026-11-03 신설) ↔ `REF-update-files-tile.md` | 11-03 밤 단위-가~라 구현·확인 · 마감(`3caff7b`·master 머지) |
- HISTORY: | `history/update-replace.md` | Files 타일 단위-다(갈아치우기)(2026-11-03 분리) ↔ `REF-update-replace.md` | 11-03 밤 확인-1 확정 · 다-1~6 구현·시험 |
- HISTORY: | `history/update-trial.md` | 파트(업데이트 app) 후속 — 배포 전 단계별 시험 실행(2026-11-03 신설) ↔ `REF-update-trial.md` | 11-03 시험-1~5 구현·확인(끝) |
- HISTORY: | `history/review-1030.md` | 10-30 작업 점검(2026-11-02 신설) ↔ `REF-review-1030.md` | 11-02 코드 항목 수정·머지, 설계 항목은 오후에 처리(→ update-redesign) |
