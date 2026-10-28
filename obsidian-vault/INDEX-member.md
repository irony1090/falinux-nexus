# INDEX — 회원 관리 (마감)

> 2026-10-26 밤 파트(회원 관리) 마감 때 MEMORY·HISTORY에서 접음. 같은 날 같이 배포된 `/api` 정리·열린 장비 주소 일괄 변경·폰 수정(`fix/phone-layout`)도 이 INDEX에 묶었다(사용자 확인). **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
회원 가입·관리자, 회원별 장비(메인키) 등록과 API 키, 내 장비·관리자 두 관리 페이지.
- 결정 회원-Q1~Q23(`REF-member.md`): 워커 수 = 메인키 개수 · API 키 = 메인키마다 하나 · 누구나 가입, 첫 회원 = 관리자 · 키 없으면 접속 거부 · 실행 권한 = 장비 주인 + 지정 폴더 공유받은 사람 · bcrypt 전환 · 장비 이름 전체 유일·변경 없음 · 인스턴스 라벨 · 키 다시 보기(해시 + AES-GCM, `KEY_SECRET`) · 정지·삭제 · 키 붙여넣기
- 구현: 마이그레이션 `00014_members.sql`, 관리 페이지 `MemberPage`(목록 + 상세 타일, PC = 목록 옆 Grid / 폰 = 가로 한 줄), URL 저장(push), 판 계약 `PagerBoard`(보드·관리 페이지 공용 — 알약 끌기·미니맵·컨트롤 패널·Ctrl+Alt+숫자) → `REF-member-ui.md`
- 후속(같은 파트로 묶음): supervisor-move 미등록 process 가져오기(`fa50c64`, `INDEX-supervisor-move`) · **`/api` 정리**(브라우저 API 전부 `/api` 아래, `REF-api-prefix.md`) · **열린 장비 주소 일괄 변경**(푸터 액션 `FootActions` + `DeviceMoveDialog`, `REF-device-move.md`) · **폰 수정**(첫 타일 이동·빈 페이지 도트·다 닫으면 목록, `history/member-ui.md`)
- 확인: 격리 회귀 17/17 · `/api` 14/14 · 주소 일괄 변경 23/24(실패 1 = 시험 탓) · 사용자 test3·폰 확인
- 커밋: `feat/member` -> master `4ae97a7` · `refactor/api-prefix` -> `c5f9749` · `fix/phone-layout` -> `8de3e05`
- 배포: test3(`build/api_prefix`, 웹은 `8de3e05`) · 집 서버(Nexus 터미널로 회원 관리 16:56, `update.ps1`로 `/api`+폰 수정 22:19) → `REF-home-deploy.md`

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- `REF-member.md`(결정 회원-Q1~Q23 · 배포 전 확인) / `history/member.md`
- `REF-member-impl.md`(구현·확인·격리 환경 다시 띄우기) / `history/member-impl.md`
- `REF-member-ui.md`(화면 개편 화-1~14 · URL 저장 · 판 계약) / `history/member-ui.md`
- `REF-api-prefix.md`(api-1~5) / `history/api-prefix.md`
- `REF-device-move.md`(이동-Q1~Q6 · 푸터 액션 자리) / 이력은 `history/api-prefix.md`에 같이

## MEMORY 인덱스 줄 · HISTORY 색인 행 · 모듈 현황 (원문)
- MEMORY: **진행 중 파트**: 회원 관리 `REF-member.md`·`REF-member-impl.md`·`REF-member-ui.md`(화면 개편·판 계약) — test3·집 서버 배포 끝, 마감 대기 / `/api` 정리 `REF-api-prefix.md` / 열린 장비 주소 일괄 변경 `REF-device-move.md`
- MEMORY 모듈 현황: | 프론트엔드 (apps/frontend) | 타일 UI·터미널·⑮(모바일 가상 키보드)·로그인·active 타일·보드 미리보기 전부 마감, 회원 관리 진행 중 | ...
- HISTORY:
  - | `history/api-prefix.md` | 브라우저 API `/api` 접두어(2026-10-26) ↔ `REF-api-prefix.md` | 10-26 api-1~5, 14/14, 다음 = 열린 장비 주소 일괄 변경 |
  - | `history/member.md` | 회원 관리 — 결정·조율(2026-10-26 착수, 진행 중) ↔ `REF-member.md` | 10-26 회원-Q1~Q23 |
  - | `history/member-impl.md` | 회원 관리 — 구현·확인 ↔ `REF-member-impl.md` | 10-26 오후 회귀 17/17·커밋 `1bd58e7`·`fa50c64`·사용자 test3 배포·관리자 지정 |
  - | `history/member-ui.md` | 회원 관리 — 화면 개편(2026-10-26 `member-impl`에서 분할) ↔ `REF-member-ui.md` | 10-26 밤 폰 첫 타일 이동 버그(`memberUrl.hook` 감시 원천)·빈 페이지 도트·다 닫으면 목록으로, `fix/phone-layout` master 머지 |

## 이 파트에만 해당하는 결정 · 남은 것
- (다른 파트에서도 지킬 것은 MEMORY "확정 결정"에 남김: 회원·장비 인가 / 배포 전 필수 / `/api` 규칙 / 판 계약)
- 403 = 의도: 회원 관리 배포 뒤 `devices`에 없는 옛 폴더 장비는 실행·목록 403 — 주인이 "내 장비"에 같은 이름으로 등록해야 실행됨
- 닉네임 = 아이디(회원-Q12, 임시 방편) · 공유받은 사람의 장비 디렉토리 목록은 403(폴더 소유자만, 원래부터)
- 관리 페이지 폰에서 상세를 다 닫으면 목록 칸으로(가(화면 이동만)), 목록 active 표시는 안 함
- 남은 것 → `BACKLOG.md`: 주소 목록 3개 상한이 지금 주소를 밀어냄(`REF-device-move.md` "확인") · Windows 터미널 출력 줄 빠짐
- 남긴 DB: `nexus_member_tmp`·`nexus_api_tmp`(격리 시험 데이터, 필요 없으면 DROP)
