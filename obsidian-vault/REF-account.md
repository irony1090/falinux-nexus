# REF: 계정 — 로그인 기기·비밀번호 변경·평문 아이디

> 이력 → `history/account.md` / 로그인 흐름(프론트) → `REF-frontend-auth.md` / 회원 관리 → `REF-member.md` / 현재 진행 → `CURRENT.md`

## 현재 구조 (2026-11-12 조사)
| 위치 | 내용 |
|---|---|
| `internal/manager/session/sessionManager.go` | gorilla `CookieStore` — 서명된 쿠키(`sid`)에 user를 담음, 30일. **서버에 세션 목록 없음** |
| `router/user.go` `signIn` | 로그인마다 `_nonce`(랜덤 16) 저장 → sid가 로그인마다 다름 |
| `router/user.go` `requireSession` | 쿠키 + 메모리 `blocked`(정지·삭제 회원)만 대조 |
| `users.identification` | 아이디의 **sha256 hex**(`toHash`, 되돌릴 수 없음). `nickname` = 가입 때 아이디 원문(회원-Q12(닉네임)), 닉네임 변경 기능 없음, UNIQUE 아님 |

## 확정 (2026-11-12, 사용자)
- 범위 = **브라우저 로그인만**. worker(API 키)·process는 건드리지 않음
- 방식 = **세션 표**(로그인 1번 = DB 행 1개, id = nonce). 처음 낸 "세대 번호" 안은 기기 목록이 안 돼서 버림
- "모든 기기에서 로그아웃" = **지금 기기 포함 전부**
- 기기 목록 + **기기마다 [로그아웃]**
- 배포하면 **전원 한 번 재로그인**(옛 쿠키엔 행이 없음 — "목록에 없으면 무효" 규칙과 공존 불가)
- 관리자가 정지·삭제하면 그 회원 세션 행도 삭제
- **비밀번호 변경** = 현재 비밀번호 확인 → bcrypt 저장 → **지금 기기만 남기고 나머지 자동 로그아웃**. 규칙은 가입 화면과 같게(길이 제한 없음, 72바이트 초과 = 400)
- 설정 다이얼로그 = **탭 `설정 | 계정`**. 계정 탭 = 비밀번호 변경 + 로그인한 기기

## 작업 단위 (안, 브랜치 `feat/account`)
| 단위 | 내용 |
|---|---|
| 세션-1(세션 표) | 마이그레이션 `00020` `user_sessions`(id=nonce, user_id, user_agent, created_at, last_seen_at) + 시작 때 메모리 로드 + `signIn` 추가·`signOut` 삭제·`requireSession` 대조 + 30일 지난 행 주기 정리. `last_seen_at`은 메모리에서 5분에 한 번 기록 |
| 세션-2(기기 API) | `GET /api/users/sessions` · `DELETE /api/users/sessions/:key`(한 기기) · `DELETE /api/users/sessions`(전부 + 내 쿠키 만료). 지운 세션 소켓 즉시 끊음. 화면 구분값 = `sidKey`(해시 앞 8자, sid 원문 안 내보냄) |
| 세션-3(정지 연동) | 정지·삭제 때 행 삭제(`cutOff` 옆) |
| 세션-4(비밀번호 API) | `PUT /api/users/password` `{current, next}` |
| 세션-5(설정 화면) | `SettingsDialog` 탭 `설정 | 계정` |
| 세션-6(확인) | 임시 DB + 5051 헤드리스, 두 브라우저 |

## 평문 아이디 + 가입 때 아이디·닉네임 분리 (2026-11-12 사용자 추가 요청 → 한 파트로 묶음, 브랜치 `feat/account`)
- 기존 계정 이전: sha256은 되돌릴 수 없다. `encode(sha256(convert_to(nickname,'UTF8')),'hex') = identification`인 행은 `identification = nickname`으로 바꿀 수 있지만, **닉네임 ≠ 아이디인 계정도 있다**(사용자 지적, 11-12 조회: 개발 DB `nexus` 10개 중 4개·`nexus_build` 9개 중 4개 불일치 = id 3 `아이로니`·65~67 `테스트1~3`, 집 서버는 미조회) → **관리자가 다른 계정을 수정할 수 있게 한다**(사용자 11-12 "지금은 실제 서비스 문제가 안 되니까"). 로그인 때 변환 안은 안 씀
- 확정:
  - **표시 = 닉네임**(공유 목록·presence 등 일반 화면). **초대(공유 대상 찾기) = 로그인 아이디 정확히 일치**(지금도 `ShareDialog` `onFind` -> `GET /users/lookup` 아이디 일치 -> 확인 단계에서 닉네임 표시, 그대로 유지)
  - **관리자 회원 목록 = 아이디·닉네임 둘 다 표시 + 검색**(지금은 닉네임만, 검색 없음)
  - **대소문자 구분 없이 중복 금지**(아이디·닉네임 둘 다)
  - **중복 확인 = 입력을 멈추면 자동**(제출 때 서버가 다시 확인)
  - **계정 탭에 닉네임 변경 추가**(같은 중복 확인 API)

## 관리자 계정 수정 (확정 2026-11-12, 사용자)
- 관리자가 다른 회원의 **아이디·닉네임 수정 + 비밀번호 재설정**(`MemberUserTile` 버튼). 중복 확인은 가입과 같은 API
- 관리자가 비밀번호를 재설정하면 그 회원 기기 **전부 로그아웃**, 아이디·닉네임만 바꾸면 유지
- 이전 마이그레이션: 후보(닉네임 → `irony`·`test1~3`) 중 해시가 맞는 것으로 평문 지정, 아무것도 안 맞으면 **아이디 NULL = 로그인 불가**, 관리자 목록에 "아이디 없음"
  - 11-12 확인(읽기만): 개발·test3 id 3 `아이로니`(관리자) = `irony`, 65~67 `테스트1~3` = `test1~3` 해시 일치 → 두 DB 모두 NULL 0개 예상
  - 후보는 해시가 맞을 때만 적용돼 다른 DB에서도 해가 없다. 사용자 요청 = test3 관리자 `아이로니` → `irony`
  - **집 서버**: 사용자 예상 = 닉네임 = 아이디. 배포 전에 같은 대조 쿼리로 NULL이 될 계정을 확인 — **관리자가 NULL이면 아무도 고칠 수 없으니 반드시 먼저 확인**

## 구현 결과 (2026-11-12, 브랜치 `feat/account`, 커밋 전)
| 위치 | 내용 |
|---|---|
| 마이그레이션 `00020_login_id.sql` | 후보(닉네임 → `irony`·`test1~3`) 해시 대조로 평문, 안 맞으면 `''`(NULL 대신 — sqlc 타입을 string으로 두려고). `uq_users_identification`(lower, `<> ''`만)·`uq_users_nickname`(lower) |
| 마이그레이션 `00021_user_sessions.sql` | `user_sessions`(id = nonce, user_id CASCADE, user_agent, created_at, last_seen_at) |
| `router/loginSessions.go` | 메모리 `loginSessions`(시작 때 로드, `valid`가 활동 표시) · `runLoginSweep` 5분(활동 기록 + 30일 지난 행 삭제) · `endLogins` = 메모리 제거 + 그 세션 소켓에 `SESSION:END` 후 끊기 · `loginKey` = sha256(nonce) 앞 16자 |
| `router/account.go` | `GET /users/available`(로그인 없이, 모양 문제·중복을 200 `{available,message}`로) · `PUT /users/password`·`/users/nickname` · `GET/DELETE /users/sessions`·`DELETE /users/sessions/:key` · `dbWriteErr`(23505 → 400) · `setPassword`/`dropLogins`/`setNickname`/`setLoginID`(관리자와 공용) |
| `router/user.go` | 가입 = 아이디·닉네임 따로 · 로그인 = 평문 대소문자 무시 + 세션 행 추가(**메모리 등록은 커밋 전** — 응답이 커밋 전에 나가서) · `requireSession`이 `logins.valid` 대조 · 로그아웃 = 행 삭제 + 같은 브라우저 다른 탭도 끊김 |
| `router/admin.go` | `PATCH /admin/users/:id/account` · `PUT /admin/users/:id/password`(자기 자신이면 지금 기기 남김) · 목록에 아이디 + 검색 = 아이디·닉네임 · 정지·삭제 때 `dropLogins` |
| `tabs.go`·`subscribe.go` | 탭에 `login` 필드(세션 끊기·접속 중 표시) · 소켓 탭 닉네임 = DB 값(쿠키 값은 로그인 때 것) |
| 프론트 | `feature/user/api/account.api.ts` · `hook/available.hook.ts`(400ms 뒤 자동 확인, 옛 응답 버림) · `component/AccountPanel.vue`(계정 탭) · `SettingsDialog` 탭 `설정 | 계정`(오른쪽 벽 `.ct-grp`) · `Signup.vue` 닉네임 칸 · `auth.store` `refresh`/`ended`/`bindSessionEnd` · `index.vue` 계정 바뀜 기준 = `auth.id` · 관리자 `MemberAccountDialog.vue` + `MemberUserTile` 아이디 줄·[아이디·닉네임·비밀번호] · `MemberList` 아이디 표시·검색 문구 |

### 확인 (2026-11-12, 임시 DB = test3 DB 사본 + 5051, 끝나고 지움)
- 마이그레이션: test3 사본 9명 전원 평문(id 3 `irony`, 65~67 `test1~3`), 아이디 없음 0
- API 44/44(가입·중복·대소문자·기기 목록·SESSION:END·비밀번호·닉네임·초대 찾기·모든 기기·관리자 수정·재설정·정지) · 화면 15/15 + 8/8(가입 중복 표시·계정 탭·기기 로그아웃·다른 브라우저 로그인 화면으로·관리자 검색·아이디 없음 표시·수정 창·폰 390 넘침 없음) · 재시작 뒤 세션 유지
- 안 해 본 것: 5분 주기 `sweepLogins`(활동 기록·30일 정리)는 실제로 돌려 보지 않음

### 알려진 한계
- 닉네임을 바꿔도 같은 계정의 다른 탭·기기 화면은 새로고침(세션 재확인) 전까지 옛 닉네임. 원격 active 배지는 소켓을 다시 붙일 때 갱신
- 마지막 활동 시각은 5분 단위

### 배포 때
- 마이그레이션 `00020`·`00021` → **배포하면 전원 재로그인**
- **집 서버**: 배포 전에 `build/home-update/261112-account-check.sql`로 아이디가 빌 계정 확인 — 관리자가 비면 아무도 못 고침(MEMORY "배포 전 필수"와 같이 묻기)
- rollback = 파일 복원만으로는 옛 빌드 로그인 불가(아이디 평문화) → DB도 덤프로 복원

## 로그인 자동완성 끄기 (2026-11-13, 브랜치 `fix/login-autocomplete`)
- 목적 = **공용 PC**에서 다른 사람에게 계정 정보가 보이지 않게(사용자). 범위 = 로그인·가입·비밀번호 변경 칸(`AccountPanel` 닉네임 칸·관리자 계정 창은 범위 밖)
- 고친 것 = 칸마다 `autocomplete="off"` + `AccountPanel` 숨은 아이디 칸(비밀번호 관리자용) 삭제. 앱의 "아이디 저장"(쿠키 `remember_id`)은 그대로

| Edge가 띄우는 것 | 나오는 곳 | A `off`만(지금) | B A + 무작위 `name` + 누르기 전 `readonly` | C B + `type="text"` 점 가림 |
|---|---|---|---|---|
| ① 전에 입력한 값 목록 | 양식 입력 기록 | 막힘 | 막힘 | 막힘 |
| ② 저장 계정 목록·미리 채우기 | 비밀번호 관리자 | 안 막힘(Chromium은 비밀번호 칸이면 `off` 무시) | 대부분 막힘 | 막힘 |
| ③ 로그인 뒤 "저장할까요?" 창 | 비밀번호 관리자 | 안 막힘 | 가끔 뜸 | 막힘 |

- 사용자 선택 = **A만**(11-13). ②③이 문제되면 B, 그래도 ③이면 C(부작용 = 한글 입력 상태면 비밀번호가 한글로, 폰 키보드 추천 단어) → `BACKLOG.md` "로그인 자동완성 B·C"
- 공용 PC는 ③이 핵심: 저장이 안 되면 ②에 보일 것도 없음. 이미 저장된 계정은 코드로 못 지움 -> 그 PC `edge://settings/passwords`에서 지움
