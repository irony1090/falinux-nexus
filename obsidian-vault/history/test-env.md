# history/test-env — 테스트 환경 · 확인용 클라이언트

> 재사용 지식 → `REF-test-env.md` / 현재 진행 → `CURRENT.md`

## 2026-10-14 — 함정 3줄 추가 + 날짜별 확인 스크립트 절 분할
- socat fork 자식·pkill 자기 셸·`/users/presence` 범위 함정 → `REF-test-isolation.md` "격리 실행" / 날짜별 절 → `REF-test-scripts.md`(`history/test-scripts.md`)

## 2026-10-08(2) 공유-f·터미널 빈틈 확인 스크립트 요령
- Python `Client`(REST + 소켓) 재사용, 헤드리스 탭 id, "위치 없음" 만드는 법, 격리 끄기 기준 → `REF-test-scripts.md` "2026-10-08 확인 스크립트 요령"(2026-10-14 분할)

## 2026-10-06 — CURRENT에서 분리
- vault 정리(색인 2단화·CURRENT 슬림화) 때 CURRENT "다음 세션 시작 시"의 환경·함정·DB 테스트 데이터 줄을 원문 그대로 옮김. 함정 목록의 출처는 2026-10-01~10-02 입력·⑪(타일 트리 서버 저장)·4 O(계정 동기화)·4.5 SNAPSHOT 확인 작업

## 2026-10-15 — worker를 꺼도 스크립트가 고아로 남음
- 재부팅 준비 중 보드 미리보기 중간안 확인(`pw/mid.mjs`)의 `tick-A`·`tick-B` 4개가 worker 종료 뒤에도 돌고 있음 발견(부모 = systemd user) → 정리. 격리 정리 절차에 추가(`REF-test-isolation.md` "끌 때 함정 — worker 자식"). 제품 쪽 문제 여부는 CURRENT 미해결로

- 2026-11-05 보드 닫기 화면 확인용으로 개발 8011·8012 + 계정 `devtest` + worker `devtest-pc`를 띄움. `.env`의 `VITE_API_URL`이 주석이라 실행 환경변수로 넘김 -> `REF-test-env.md` "환경"

> 2026-11-11 격리 실행 엔트리(10-07·10-07(2)·10-07(작업공간-e)·10-08 공유-e) → `history/test-isolation.md`
