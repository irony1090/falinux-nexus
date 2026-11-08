# REF — 설정 두 겹: 계정 + 이 브라우저 (2026-11-08 신설, 파트(설정 두 겹))

> → 이력 `history/settings-account.md` · 앞 파트 설정 저장소·설정 창 `REF-settings-footer.md` · 결정 요약 `REF-decisions-ui.md` "푸터 오른쪽·설정" · 진행 `CURRENT.md`

## 결정 (확정 2026-11-08, 사용자)
- **Q2(설정 범위) 개정**: "항목마다 범위 하나" -> **모든 항목이 두 겹**. 고르면 계정(서버)에 저장, 별도 "이 브라우저에서만" 값이 있으면 그게 우선
- 쓰이는 값 = 브라우저 값 ?? 계정 값 ?? 기본값(`SPECS.init`)

| 이름 | 결정 |
|---|---|
| Q24(창 모양) | 묶음 두 개: **계정 설정** = [위 \| 아래] / **이 브라우저에서만** = [계정 설정 따름 \| 위 \| 아래]. "계정 설정 따름" = 브라우저 저장값 지움(따로 지우기 버튼 없음) |
| Q26(다른 기기 반영) | 계정 값을 바꾸면 열려 있는 다른 기기·탭에 **바로 반영**(`USER:<id>` push) |
| Q27(적용 범위) | 앞으로 생길 설정 전부 이 구조. 첫 항목 = 메뉴 줄 위치(`tileBarAt`) |
| Q25(새로고침 깜빡임) | 계정 값 사본을 브라우저(userPrefs `settingsAccount`)에 둠 — 받기 전에도 마지막 계정 값, 로그아웃 때 같이 삭제 |

## 구현 계획

| 단계 | 할 일 | 파일(안) |
|---|---|---|
| 단계-1(서버 저장) | `user_settings(user_id, key, value jsonb, updated_at, PK(user_id,key))`. 서버는 키를 모르고 값만 보관, 모양 검사는 프론트 `SPECS.ok` | 새 마이그레이션 `00019`(착수 때 실제 다음 번호) + sqlc 쿼리 |
| 단계-2(서버 API·push) | `GET /api/settings`(전체) · `PUT /api/settings/:key`. 저장 뒤 `USER:<id>`로 `SETTINGS:UPDATE`(계정 값 전체) — `WORKSPACES:UPDATE`와 같은 방식 | 새 `cmd/supervisor/router/settings.go`(`mountSettings`) · `internal/protocol/messages.go` |
| 단계-3(프론트 저장소) | 값 = `account`/`browser` 두 겹 + 쓰이는 값 computed. 로그인·세션 확인 때 GET, push 받으면 계정 값 교체. 브라우저 겹은 지금 userPrefs 그대로(키 `settings.<항목>`) | `feature/settings/store/settings.store.ts` |
| 단계-4(설정 창) | Q24(창 모양)대로 묶음 두 개 | `SettingsDialog.vue` |
| 단계-5(확인) | 두 기기 동시 반영 · 브라우저 값 우선 · 로그아웃하면 브라우저 값만 지워짐 · 계정 값은 다음 로그인에도 남음 | — |

- 지금 저장된 `nexus.user.settings.tileBarAt`(11-07 브라우저 범위)은 그대로 "이 브라우저에서만" 값으로 읽힘 — 옮기기 없음


## 구현 결과 (2026-11-08, 브랜치 `feat/settings-account`)
| 파일 | 바뀐 것 |
|---|---|
| 새 `supervisor/db/migrations/00019_user_settings.sql` + `query/userSettings.sql`(+ sqlc) | 테이블 `user_settings` · `ListUserSettings`(key·value·updated_at) · `SaveUserSetting`(upsert) |
| `internal/protocol/messages.go` | `SETTINGS:UPDATE` + `SettingsUpdateEvent{settings, rev}` |
| 새 `cmd/supervisor/router/settings.go` + `supervisorRouter.go` | `mountSettings`: `GET /api/settings` · `PUT /api/settings/:key {value}`. 키 = 영문·숫자 64자, 값 = JSON 4KB, 계정마다 키 64개. 응답 = push와 같은 모양(계정 값 전체 + rev) |
| 새 `feature/settings/api/settings.api.ts` | `getAccountSettings` · `putAccountSetting` |
| `common/util/userPrefs.util.ts` | `removeUserPref` 추가 |
| `feature/settings/store/settings.store.ts` | 두 겹: `browser`·`account` ref, `settings`(쓰이는 값) · `accountSettings` · `browserSettings`(없음 = 계정 따름) · `setAccountSetting`(고른 값 바로 표시 후 PUT) · `setBrowserSetting(k, undefined)` = 지움 · `bindAccountSettings(socket)`(push + 다시 연결되면 GET) · 로그인 상태 watch(`immediate`) = 로그인 때 GET |
| `SettingsDialog.vue` · `App.vue` | 묶음 두 개(계정 설정 / 이 브라우저에서만), 브라우저 값이 있으면 계정 칸에 안내. `App.vue`에서 `bindAccountSettings(socket)` |

- 도착 순서 처리(설정-1(rev 비교)): HTTP 응답과 push는 길이 달라 늦게 온 옛 값이 새 값을 되돌릴 수 있음 → 서버가 `rev` = 가장 최근 `updated_at`(µs)을 같이 보내고, 프론트는 rev가 작은 것을 버림. 이 탭 저장이 진행 중이면 고른 값을 그대로 두고 최신 서버 값만 기억했다가 마지막 요청이 끝날 때 적용(실패하면 그때 서버 값으로 되돌아감)
- 설정-2(로그아웃 뒤 응답): 로그인 상태가 바뀌면 `gen`을 올려, 로그아웃 뒤 도착한 GET·PUT 응답이 이전 계정 값을 사본에 쓰지 않게 함
- 계획과 다르게 한 것: 브라우저 칸 첫 버튼 글자 = **"계정 따름"**(Q24(창 모양)의 "계정 설정 따름"을 줄임 — 폰 390px에서 세 버튼이 한 줄에 들어가게)
- 확인(격리 5051 + 임시 DB `nexus_set_tmp`, 계정 `setcheck`, 빌드한 화면을 supervisor `WEB_DIR`로 — 끝나고 DB 지움): 기기 A(탭 둘)·기기 B 같은 계정. 처음 위 · A 계정 아래 = DB 행·A 두 탭·**B 새로고침 없이** 아래 · 두 기기 사본 저장 · B 이 브라우저 위 = B만 위·계정 칸 안내 · A 계정 위->아래 연속 = B는 위 유지 · B 계정 따름 = 아래·브라우저 저장값 지워짐 · **계정 응답 3초 늦춰도 새로고침 직후 아래(사본)** · B 로그아웃 = `nexus.user.*` 전부 지워짐 · 다시 로그인 = 계정 값 아래 · A 영향 없음 · 콘솔 에러 없음 = **24/24**(첫 판 2 FAIL = 스크립트가 로그아웃 확인 창을 안 누름). 폰 390px 설정 창 = 고르기 묶음 오른쪽 벽에 붙음, 넘침 없음. API: 잘못된 키·빈 값 = 400. `go test` router 통과 · type-check 통과 · `run.sh` = 낡은 기대값 2건(updater 되돌리기) 외 통과. 스크립트 = 세션 scratchpad `iso/setup.mjs`·`iso/check.mjs`·`iso/phone.mjs`(재부팅 시 사라짐)
- **마감(2026-11-08)**: 사용자 화면 확인 "확인했어" · 커밋·master `--no-ff` 머지. 배포할 땐 마이그레이션 `00019`가 같이 감 — 추가만 하는 테이블이라 옛 빌드로 되돌려도 DB 복원 불필요
