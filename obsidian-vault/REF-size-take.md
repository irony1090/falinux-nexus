# REF — 크기 우선권 즉시 가져오기 (파트(크기 가져오기), 2026-10-13~, **마감 2026-10-14**)

> 이력 → `history/size-take.md` / 크기 우선권 원래 설계 → `REF-process-sync.md` Q(크기 우선권 단위)·`REF-process-share-authz.md` SB3(크기 우선권 후보) / 질의 응답 → `REF-term-query.md` / 현재 진행 → `CURRENT.md`

## 계기·현재 동작 (2026-10-13 코드 확인)
- 사용자 요청: 같은 계정끼리는 화면 크기 우선권을 즉시 가져오는 버튼
- 지금 넘김 = 소유 탭이 끊긴 뒤 `tabGrace`(60초) 지남 / 로그아웃 / 그 터미널이 없는 보드로 감(`resyncSizeOwner`)뿐 → 넘길 때 옛 소유 탭은 거의 없음 → **옛 탭에 "잃음"을 알리는 메시지가 없다**(`PROCESS:SIZE_OWNER`는 얻은 탭에만, 프론트는 `owners_[uid] = true`만 함)
- 그래서 버튼으로 가져오면 옛 탭이 살아 있는 채로 잃음 → 알림 없으면 옛 탭이 계속 소유자로 여겨 resize 403 + xterm 질의에 두 탭이 응답(파트(질의 응답 중복) 회귀) → 크기-2(잃음 알림) 필수
- 프론트는 소유자가 되면 `ProcessTerminal`의 `watch(owner, schedule)`가 자기 타일 크기로 PTY를 바로 맞춤(추가 코드 불필요). 잃으면 `--fixed`(소유자 rows·cols 고정 + 스크롤)

## 구조안 (2026-10-13 제시 → K1·K2 확정, 코드 없음)
| 단계 | 파일 | 내용 |
|---|---|---|
| 크기-0(여유 시간) | `tabs.go`·`sizeOwner.go` | K3(여유 시간) (나): 크기 넘김 대기 상수 15초 신설, `sizeOwnerOnDisconnect` 타이머를 그 상수로. `tabGrace`(60초) 옆에 두고 "크기 넘김 <= 탭 id 보관" 주석 |
| 크기-1(가져오기 API) | `processApi.go`·`sizeOwner.go` | `POST /processes/:id/size-owner`(`X-Tab-Id`). 요청 탭이 후보(`sizeOwnerCandidates`: 조작(20) 이상 + 그 터미널이 있는 보드를 보는 연결 탭)이고, 현재 소유가 **같은 계정 탭**·비어 있음·끊김 대기 중이면 교체(타이머 정지). 다른 계정 연결 탭이 소유 중이면 403 (→ K4(소유 계정 우선)로 개정, 아래 "구현") |
| 크기-2(잃음 알림) | `protocol/messages.go`·`processTerm.store.ts` | `SizeOwnerEvent`에 잃음 표시(필드 없으면 = 얻음, 기존 호환) → 옛 탭 `owners_[uid] = false` |
| 크기-3(버튼) | 터미널 타일 머리 줄(`TileFrame`/`TerminalTileBody`) | 이 탭이 소유자가 아닌 실행 중 터미널에만 `mdi-arrow-expand-all`, 툴팁 "이 화면 크기로 가져오기" |
| 크기-4(확인) | 격리 헤드리스 | 같은 계정 탭 2개 가져오기·되가져오기 / 옛 탭 고정·질의 응답 멈춤 / 다른 계정 소유 = 안내 |

## 결정표
| 번호(이름) | 내용 | 상태 |
|---|---|---|
| K1(다른 계정 소유) | 버튼은 보이고, 누르면 "다른 계정이 크기를 쓰는 중" 안내(서버 403). 숨기려면 소유 탭의 계정을 계속 내려보내야 해서 안 함. **K4(소유 계정 우선)로 개정** — 스크립트 소유 계정에는 이 막힘이 없음 | 확정(2026-10-13 사용자) |
| K2(버튼 모양) | 머리 줄 아이콘 `mdi-arrow-expand-all` + 툴팁 "이 화면 크기로 가져오기" | 확정(2026-10-13 사용자) |
| K4(소유 계정 우선) | K1 개정: 같은 계정끼리 + **스크립트 소유 계정(`owner_user_id`, SH9) = 다른 계정에게서도 가져옴**. 실행한 계정(`started_by`) 기준 아님. 공유받은 계정끼리는 서로 못 가져옴, 공유받은 계정의 버튼·안내는 K1 그대로. 근거: 조작(20) 이상은 이미 입력 가능 → 크기 변경은 권한상 위험 없음, 위험은 남의 화면을 바꾸는 사용 경험뿐 + 한 방향이라 서로 가져가기 없음 | 확정·구현(2026-10-13) |
| K5(빼앗긴 탭 안내) | 다른 계정(소유 계정)이 가져갔을 때 잃은 탭에 안내? (ㄱ) 없음 / (ㄴ) 잠깐 뜨는 알림 "A님이 크기를 가져갔습니다"(잃음 이벤트에 닉네임) — 권장 / (ㄷ) 머리 줄 툴팁에 현재 소유 계정(계속 내려보내야 함) → **(ㄴ) 확정**. 표시 = 위치 알림 토스트가 아니라 **그 터미널 타일 오른쪽 위 칩 4초**(위치 알림은 제목·누르면 이동이 타일 위치용이라 안 맞음) | 확정·구현(2026-10-13 사용자) |
| K3(여유 시간) | 사용자: 60초 -> 10~15초로 줄이고 싶음. **쟁점**: `tabGrace` 하나가 ① 크기 우선권 넘김 대기 ② 끊긴 탭 id 보관(새로고침·재연결 때 같은 탭 id 재사용, presence "보관 중", REST `belongsTo`)에 같이 쓰임 → (가) 둘 다 15초 / (나) 크기 넘김만 별도 상수 15초, 탭 id 보관은 60초 유지 → **(나)로 확정**: 크기 넘김 대기 15초(새 상수) / 탭 id 보관 `tabGrace` 60초 유지, 순서 조건 주석 | 확정(2026-10-13 사용자) |

- K3 (나)의 유지보수 검토(2026-10-13 사용자 질문): 문제없음, 단 **크기 넘김 대기 <= 탭 id 보관**을 지켜야 함 — 반대면 탭 id는 폐기됐는데 우선권은 그 탭에 묶여 대기 끝까지 아무도 resize 못 함. 두 상수를 `tabs.go`에 나란히 두고 순서 조건을 여러 줄 주석으로(선택: 시작 때 순서 검사 한 줄). 넘김 타이머(`sizeOwnerOnDisconnect`)는 이미 탭 id 보관 타이머와 별개라 구조 변경 없음

## 구현 (2026-10-13, 2026-10-14 마감·커밋)
| 단계 | 실제 구현 |
|---|---|
| 크기-0(여유 시간) | `tabs.go` const 묶음: `tabGrace` 60초(탭 id 보관) + `sizeHandoffGrace` 15초(넘김 대기), 위에 "sizeHandoffGrace <= tabGrace" 여러 줄 주석. `sizeOwnerOnDisconnect` 타이머만 새 상수로. 시작 때 순서 검사는 안 넣음 |
| 크기-1(가져오기 API) | 경로는 기존 `/resize/:processId`·`/kill/:processId` 모양에 맞춰 **`POST /processes/size-owner/:processId`**(구조안의 `/:id/size-owner`에서 바꿈). `requireProcess`(조작 이상 + memory에 있음 → PENDING은 404) → `takeSizeOwner`: 후보 아님 403 "이 화면에 없는 터미널입니다" / 다른 계정의 **연결된** 탭이 소유 403 "다른 계정이 이 터미널의 크기를 쓰는 중입니다"(스크립트 소유 계정은 예외 — K4(소유 계정 우선)) / 그 밖(같은 계정·비어 있음·끊김 대기)은 교체 + 타이머 정지. 응답 `{}`(소유 표시는 이벤트로) |
| 크기-2(잃음 알림) | `SizeOwnerEvent.Lost`(`lost,omitempty`). 보내는 곳 = `takeSizeOwner` + **`resyncSizeOwner`에서 연결된 소유자를 다른 후보로 바꿀 때도**(구조안 밖 추가 — 같은 빈틈: 다른 보드로 간 탭이 `owners_` true로 남던 것). 프론트 `owners_[uid] = !ev.lost` |
| 크기-3(버튼) | `TileFrame.vue` 머리 줄, 접속 상태 버튼 왼쪽. 조건 = 터미널 · 상태 `PROCESS`(PENDING 제외 — 서버가 404) · 보기 전용 아님 · 이 탭이 소유자 아님. 실패 = `openDialog` error로 서버 메시지. 스토어 `takeSize(uid)`는 API만 부름(응답으로 true를 세우지 않음 — 늦게 다른 탭이 가져갔을 수 있어서) |

## 크기-4(확인) 결과 (2026-10-13, 격리 헤드리스 12/12)
- 격리: 임시 DB `nexus_size_tmp` + supervisor 5051 + worker `NAME=size-dev` + 프론트 3001(`vite.iso.mts` = `base({mode:'development',command:'serve'})` 펼치고 `server.port` 3001·`hmr: undefined`, `VITE_API_URL=http://localhost:5051`). 확인 뒤 프로세스 종료·DB 삭제
- 스크립트 `pw/size.mjs`(재부팅 시 사라짐, 명세): 계정 A(같은 context 페이지 A1 1400×900·A2 1000×700·A3) + 계정 C(폴더 조작(20) 공유). 시험 스크립트(bash) = `stty -icanon -echo` 후 한 글자 읽어 `q` → `ESC[6n` 출력·1초 뒤 남은 입력의 `R` 수를 `Q<n>=<수>`로 / `s` → `S=$(stty size)`. A1 탭 id로 API exec → 세 페이지 새로고침. 상태 = `.ProcessTerminal--fixed` 수 + `button[title="이 화면 크기로 가져오기"]` 수
- 결과: ① 초기 A1 소유·A2/C 고정+버튼 ② A2 질의 응답 1 ③ A2 가져옴 → A1 고정+버튼 ③b PTY 111×21 → 78×15(`stty size` 일치) ④ 잃은 A1에서 질의 응답 1 ⑤ A1 되가져옴 + 질의 1 ⑥ C 누름 → 안내 창, 계속 고정 ⑦ A1 닫은 직후 A2가 바로 가져옴 ⑧ 16초 뒤에도 A2 유지 ⑨ A2 닫음 → A3가 **15.2초** 뒤 받음
- 대조 실행(잃음 알림 뺀 상태)은 안 함 — ③(A1 고정)이 잃음 이벤트에 직접 걸려 있어 대신함

## K4(소유 계정 우선)·K5(빼앗긴 탭 안내) 구현 (2026-10-13, 2026-10-14 마감·커밋)
- 서버: `takeSizeOwner(uid, tabID, userID, processOwner)` — 다른 계정의 연결된 탭이 소유 중이면 `processOwner`(= `processPerm == PermOwner`, 스크립트 소유 계정)일 때만 교체, 아니면 403. 교체 시 잃음 이벤트 `By` = 가져간 탭의 닉네임(`tabs.tabOf`, 다른 계정일 때만). `userOfTab` → `tabOf`(탭 전체 반환)로 바꿈. 자동 넘김(`resyncSizeOwner`)의 잃음은 `By` 비움
- 프론트: 스토어 `takenBy_[uid]`(4초 뒤 지움, 다시 얻으면 즉시 지움, `reset`에서 비움) → `TerminalTileBody.vue` 오른쪽 위 `v-chip color="warning"` "<닉네임>님이 크기를 가져갔습니다"(`pointer-events: none`, fade)
- 확인(격리 헤드리스, 같은 `pw/size.mjs` 확장) **19/19**: 앞의 12개 + ③c 같은 계정이 가져가면 칩 없음 / ⑩ A 탭이 모두 끊김 → 15.2초 뒤 C가 받음 / ⑪a 새 A 탭 고정+버튼 / ⑪b A가 C에게서 가져옴 / ⑪c C에 칩 "sizeA…님이 크기를 가져갔습니다" / ⑪d 4초 뒤 사라짐 / ⑫ C는 A에게서 못 가져옴(안내)
- 함정: 안내 창(`openDialog`)을 안 닫으면 scrim이 다음 클릭을 가림 → `.v-dialog.v-overlay--active button:has-text("닫기")`
