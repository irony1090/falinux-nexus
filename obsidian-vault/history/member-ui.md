# history — 회원 관리: 화면 개편

→ `REF-member-ui.md` · 화면 구현 이력 `history/member-ui-impl.md` · 구현 이력 `history/member-impl.md` · `CURRENT.md`

## 2026-10-26
- 격리 환경 다시 띄움(5051·3001·`jig-a`, 새 서브키 `xIxvhr4S`에 라벨) + 사용자 직접 실행용 worker env `cmd/worker/test.env`(-> `REF-member-impl.md`). 사용자: "시스템 동작은 잘 됨, UI는 많이 고쳐야 함" — 찰떡 유지 · 타일 방식 + 반응형 · 입력칸·버튼·정보 모양이 따로 놂 · 라벨 관리를 못 찾음(인스턴스 줄 끝 작은 버튼)
- 조율: 안 A(보드 타일 종류) 말고 B(별도 페이지 안 타일 격자) · 범위에 로그인·가입 포함, 아티팩트로 전체 확인 · 목록 = 페이지 번호 + 검색·필터 · 서버 쪽 페이지·검색 포함
- 시안 1판(Design 캔버스 https://claude.ai/artifact/XeSQZMThQjbs5CxGMage9r): PC [목록 | 상세 | 관련] 3타일, 폰 목록 -> 상세, 위 헤더
- 사용자 첨삭 -> 시안 2판: 목록 세로 전체 · 상세는 Grid에 쌓임 + 가로 스크롤 · 회원 -> 장비 = 안 가(회원 타일 안 장비 줄) · 자식 = 부모 바로 다음(보드 `READ` 순서 확인: 왼쪽 위 -> 왼쪽 아래 -> 오른쪽 위 -> 오른쪽 아래) · 크기 버튼 + 종류별 처음 크기 · 경로 · 위 헤더 대신 보드 같은 아래 푸터 · 한 페이지 줄 수 선택
- 시안 1·2판의 화면 이동이 안 됨: 캔버스 런타임에서 `ref` 콜백이 안 붙음 -> `data-scroll` + `querySelector` + `setState` 콜백으로 고침
- 질문 4개(목록 고정 여부·이동·경로·타일 수) 답 -> 보드 `capacity`(폭 728·높이 560 넘으면 2칸, 최대 2×2) 확인 -> 사용자: 관리 페이지에만 N×M, 상세 최대 3×2, 목록 폭 = 한 칸, 가로 스크롤 웬만하면 없게 -> 안 나(상세만 3×2, 목록 별도 열) 확정, 시안에 `관리자 · 큰 화면 1920` 추가
- 시안 올리는 중 캔버스 페이지가 `Main.dc.html`·`canvas.json`을 다시 저장(편집기 형식 차이뿐) -> 서버 판 위에 합쳐 올림
- 사용자 "디자인 좋아 맘에 들어" -> URL 저장 조율: 타일 구성·목록 종류·필터 = push, 페이지·줄 수·검색·보던 위치 = replace, 들어올 때 마지막 구성(`localStorage`) -> 코드 적용은 다음 세션(사용자)

## 2026-10-30 — 열기 규칙을 보드와 같게 (파트(내 장비·관리자 타일 이동) 1단계)
- 사용자: 이번 작업 = 페이지 간 이동이 아니라 **그 화면 안 순서 바꾸기** · 순서 저장은 **URL로 충분**(서버 저장 안 함) · "열기 규칙도 메인 보드와 똑같이 우선 바꿀까"
- 지금 차이: 자식 = 부모 뒤 *기존 자식들 뒤*(최신이 뒤) / 연 곳 없음 = **맨 끝**. 보드 `openTile`(`tileDoc.go:236`)은 연 곳 바로 뒤(최신이 앞) / 연 곳 없음 = 루트 바로 뒤(= 맨 앞)
- 확정(사용자 "맨 앞으로, 보드랑 똑같이") -> 화-9(자식 자리) 개정
- 코드: `memberGrid.util.ts` `insertTile`에서 `isDescendant` 제거, `at = pi >= 0 ? pi + 1 : 0` 2줄로. `memberBoard.store`의 `open`·호출처(목록 3곳·회원 타일 장비 줄)는 그대로. `vue-tsc` 통과
- 영향 없음: URL `t` 형식(옛 URL 유효) · `pack`·`anchorKey`·`view`(순서만 따름) · 백엔드·마이그레이션. 따라 바뀜: 위치 번호(Ctrl+Alt)·닫은 뒤 active 이웃 = 배열 인덱스 기준이라 함께 움직임(보드 번호-A(화면 순서)와 같음)
- 남긴 것: 경로 재열림(`openParent` -> `insertBefore` "그 자식 바로 앞")은 옛 동작 그대로. 자식이 여럿이면 첫째가 아닌 자식에서 재열릴 때 부모가 자식 묶음 중간에 낄 수 있음 — 끌기 작업 때 같이 볼지 사용자와 정할 것

## 2026-11-03 화-10(타일 크기) 개정 — 처음 높이 1칸
- 사용자: 내 장비·관리자·Updater 페이지에서 목록을 눌러 처음 연 상세 타일은 높이 1. 크기 조절 뒤 URL에 반영된 크기는 URL을 따름(`decodeTiles` 그대로)
- 결정: 회원 페이지 = 전부 1칸(장비 2 -> 1) / updater = Run·Log(`run`)·Script(`text`)만 2칸(로그·편집기), 장비·Files·steps.json·공용 묶음 2 -> 1(steps.json도 1 = 사용자)
- 코드: `memberGrid.util.ts` `MEMBER_KEYS.size` · `updGrid.util.ts` `UPD_KEYS.size` · 확인 `scripts/check/member-drag.mjs` 기댓값 `ddev-1*1x2` -> `*1x1`. `vue-tsc`·`run.sh` 전체 통과. 브랜치 `feat/updater-p3`(커밋 안 함)

## 2026-11-03 화-7(채우는 순서) 개정 — Grid 사이 빈 열 없애기
- 신고: Updater에서 Script(1x2)를 2x2로 넓히자 Files 오른쪽이 빈 열, Script는 다음 Grid로 넘어가 떨어짐(URL·화면 `user/배치에러.png`). 처음엔 스크롤 문제로 오해(수정-1(도착 확인) 제안) -> 사용자 "스크롤 문제가 아냐, 원래라면 붙어야" -> 원인 = Grid가 늘 3열 폭을 차지
- 결정: 배치-1(붙는 방식) = 보드처럼 Grid 폭 = 쓴 열 수 / 배치-2(페이지 점) = 점 하나 = 3열 그대로(사용자: "지금처럼 3열이 하나의 점")
- 코드: 배치-가 `memberGrid.util.ts` `pack`이 Grid마다 `cols`(쓴 열)·`at`(시작 열) / 배치-나 `memberBoard.store.ts` `colCount`·`dotCount`·`goDot`·`colRange`·`anchorKey`·`peekGrids` / 배치-다 `MemberPage.vue` Grid 폭·열 템플릿·스냅을 Grid별로(폰은 CSS 폭 그대로) / 배치-라 `memberDrop.util.ts` `MemberFrame` = 칸 격자(x·y·colW·rowH)로 바꾸고 Grid 상자는 pack 결과로 계산, `memberDrag.store.ts` 잼, `member-drop.check.ts` 맞춤
- 확인: vue-tsc · 끌기 판정 전수 114만 회 왕복 0 · `run.sh` 전체 통과 · 사용자 URL headless = Files 옆에 Script, 점 2개. 폰 화면은 따로 안 봄(1열이라 쓴 열 = 1)
