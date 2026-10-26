# REF — 열린 장비 타일 supervisor 주소 일괄 변경 (2026-10-26, master 머지 `c5f9749` · 마감 → `INDEX-member.md`)

> → 이력 `history/api-prefix.md`(같은 세션에서 시작, 착수하면 `history/device-move.md` 분리) · → `CURRENT.md` · 바탕 = `REF-supervisor-move.md` 주소-5(화면)

## 요구 (사용자 2026-10-26)
| 번호(이름) | 결정 |
|---|---|
| 이동-Q1(대상) | 내 장비·관리자 판에 **열린 장비 타일**(목록 행·필터 결과 아님) |
| 이동-Q2(instance 여럿) | 한 장비에 instance가 여럿이면 **셀렉트로 골라서** |
| 이동-Q3(표기) | 고르려면 **instance별로** 나와야 함(장비 이름 + 라벨/서브키) |
| 이동-Q4(여는 곳) | 사용자: 목록 타일 [장비 등록] 옆 설정 버튼에 액션 모음 생각 -> 모바일은 목록이 안 보임 -> 푸터? / 확정 = **푸터 오른쪽 계정 알약 바로 왼쪽의 액션 자리**(아래 "푸터 액션 자리") |
| 이동-Q5(관리자) | 관리자 판에도 필요(남의 장비 포함) |
| 이동-Q6(순서) | `/api` 정리(`REF-api-prefix.md`) 먼저, 새 API는 처음부터 `/api` 아래 |
| 단일/복수 | 고른 instance 1개 = 기존처럼 저장된 주소 순위 표시 / 2개 이상 = 바뀔 이름 목록 |

## 푸터 액션 자리 (사용자 2026-10-26)
- 액션 = 판 상태로 쓸 수 있는지가 바뀌는 버튼. 쓸 수 없을 때도 **비활성으로 보임**(액션이 있다는 티)
- 자리는 늘 버튼 하나 크기(푸터 공간 부족, 모바일 포함): 액션 **개수** 1 = 그 아이콘 버튼 / 2 이상 = 편집 아이콘(`mdi-pencil-outline`) 하나 -> 위로 목록(항목별 활성·비활성)
- 공용 = `widget/component/FootActions.vue`(`FootAction {key, icon, title, disabled, hint, run}`), 보드 푸터도 그대로 쓸 수 있음

## 구현 (`refactor/api-prefix` 브랜치에 같이)
| 자리 | 내용 |
|---|---|
| 서버 `router/deviceMove.go` | `POST /api/devices/move {targets, url, now}`(주인 아님 = 404) · `POST /api/admin/devices/move` · `GET /api/[admin/]devices/:mainKey/instances/:subKey/addrs`(접속 안 됨 = 409). 대상 1~100, instanceKey 형식 400, 주소 검사 = `normalizeMoveURL`(worker와 같은 함수). 동시 보내기 = `workerMove.go` `moveConns`(move-all과 공용으로 뽑음), 접속 안 된 대상 = 그 줄 `error` |
| 프론트 | `device.api.ts` `moveDevices`·`instanceAddrs`·`useMemberDeviceList`(타일과 같은 queryKey, `useQueries`) · `hook/openDevices.hook.ts`(열린 장비 + 접속 중 있음) · `component/DeviceMoveDialog.vue`(장비별 instance 체크, 기본 = 접속 중 전부, 1개 = 순위, 여럿 = 바뀔 이름, 결과 태그) · `MemberFooter.vue` 액션 1개 |

## 확인 (2026-10-26, 격리 5051 + `nexus_api_tmp`)
- `movetest.mjs` 23/24 + `nowtest.mjs`: API(저장 -> 순위 1위 · 남의 장비 404 · 접속 안 된 대상 = 줄 error · 형식·주소 400 · 순위 409/404 · 관리자 아님 403 · 관리자 남의 장비 200) / 화면(장비 타일 없음 = 비활성 -> 3개 열면 활성 -> 창 = 접속 중 3개 체크·기록 없는 장비 표시 -> 여럿 = 이름 -> 1개 = 순위 -> 2개 저장 = 결과 태그·"2개 저장" -> 폰 폭 버튼 보임) / 바로 옮기기 -> 없는 주소 -> 3초 뒤 원래 주소로 복귀
- 실패 1개 = 시험 탓: 같은 instance에 주소를 4번 넣어 목록 3개 상한(주소-1)에서 원래 주소가 밀려 지워짐 -> 돌아올 곳이 없음. **실사용 함정**: "저장만"을 여러 번 하면 지금 붙어 있는 주소도 밀려날 수 있음 -> BACKLOG
