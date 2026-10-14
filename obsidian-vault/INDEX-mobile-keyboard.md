# INDEX — 모바일 가상 키보드 (마감, ★중요 — 다른 프로젝트에서도 생기는 문제)

> 2026-10-14 파트(모바일 가상 키보드) 마감 때 MEMORY·HISTORY에서 접음(사용자 "별도 색인, 중요도 높음"). **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`
> **프로젝트와 무관한 지식은 반복 주제 `REF-mobile-viewport.md`가 먼저** — 모바일 입력·키보드·뷰포트 문제를 만나면 여기부터

## 파트 요약
터미널 타일에 폰 가상 키보드로 입력할 때: 특수 키 줄(Esc·Tab·Ctrl·Alt·방향키·`| ~ / -`, Ctrl·Alt 탭/누른 채, 한글 자모 -> 영문) + 가려진 타일을 커서 줄 기준으로 올리기 + `interactive-widget=resizes-content` 전환(키보드 열린 동안 타일 영역 높이 고정으로 재배치·PTY 크기 변경 막음). Galaxy Fold 8 크롬 정확, Edge 폰 크기 기본 화면(하단 막대)은 약 34 빈틈(브라우저 값 오류, 해결 불가로 기록). iPhone 확인은 나중

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- 반복 주제(프로젝트 무관): `REF-mobile-viewport.md` / `history/mobile-viewport.md`
- Nexus 구현: `REF-mobile-keyboard.md`(구조안·결정 VK1~VK4·구현표·헤드리스 결과·확인 목록) / `history/mobile-keyboard.md`(측정·실험 경위 2026-10-14(1)~(7))
- 결정 근거(A+B 방향, 위험 5가지): `REF-node-ui-overview.md` "모바일 가상 키보드"
- 확인 스크립트 명세: `REF-test-scripts.md`(`pw/kb.mjs` 22 · `kb2.mjs` 10 · `kb3.mjs` 12) / 폰 사진 올리기: `REF-test-env.md` "폰 사진 올리기"

## HISTORY 색인 행 (원문)
### `history/mobile-keyboard.md` — 모바일 가상 키보드 구현 (2026-10-14 신설)
- 2026-10-14(7) 마감·커밋(KbDebug 삭제, 업로드 도구 주석, 반복 주제 분리)
- 2026-10-14(6) 라(resizes-content 전환) 구현 + 헤드리스 12/12 + 폰 확인 OK(크롬 정확, Edge 빈틈 34)
- 2026-10-14(5) 사진 14장 + 측정(KbDebug) -> Edge 하단 막대가 원인, 안 가(가상 키보드 API)·안 라(resizes-content)·svh/dvh/lvh 실험, 웹 조사, 확인 목록 3~7 OK
- 2026-10-14(4) 고침-나(커서 기준 이동) + 헤드리스 10/10
- 2026-10-14(3) 첫 실기기 화면(Fold 8): 이상-가(키 줄 뜸)·이상-나(아래 가려짐)
- 2026-10-14(2) 키보드-0~4 구현 + 헤드리스 22/22
- 2026-10-14 착수: 코드 없음 확인 + 구조안 키보드-0~5 제시 + VK3(Ctrl 동작)·VK4(사용자 키 추가)
### `history/mobile-viewport.md` — 모바일 가상 키보드·뷰포트 반복 주제 (2026-10-14 신설)
- 2026-10-14 반복 주제로 분리

## 이 파트에만 해당하는 결정
- VK1(키 목록) `Esc Tab Ctrl Alt ← ↑ ↓ → | ~ / -` / VK2(표시 조건) 터미널 포커스 + 키보드 열림 + 조작 권한 + PROCESS 상태 / VK3(Ctrl 동작) 탭·누른 채 둘 다 + 한글 자모 변환 / VK4(사용자 키 추가) 안 함
- 키보드 열린 동안 타일 영역은 키보드 전 높이 고정(라-2) — 타일 재배치·PTY 크기 변경 금지
