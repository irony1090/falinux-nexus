# REF — 회원 관리: 구현·확인·다음 할 것 (파트(회원 관리))

→ 결정 `REF-member.md` · 이력 `history/member-impl.md` · 현재 `CURRENT.md`

## 작업 위치 (2026-10-26)
- **worktree `/home/jh-bae/irony/nexus-member`, 브랜치 `feat/member`** — 머지 커밋 `6f87465`(= `19cc8bd` + `feat/supervisor-move`, vault 충돌 4개는 master 쪽) 위에 **커밋 안 한 변경**이 있음(재부팅해도 디스크에 남음). 커밋은 파트 마감 때
- worktree에만 있는 git 제외 파일: `apps/core/cmd/{supervisor,worker}/env`(원래 폴더에서 복사, supervisor `env`에 `KEY_SECRET` 추가함) · `apps/frontend/node_modules` = 원래 폴더 `node_modules`로의 심볼릭 링크
- 원래 폴더(`/home/jh-bae/irony/nexus`, master)와 `feat/supervisor-move`는 안 건드림. vault는 원래 폴더에서 기록

## 구현 현황 (커밋 전)
| 번호(이름) | 상태 | 파일 |
|---|---|---|
| 회원-1(DB) | 끝 | `00014_members.sql`(users `worker_limit`·`suspended_at`, 기존 회원 `worker_register` 부여, `devices`, `worker_instances.label`) · query `devices.sql`·`users.sql`·`workers.sql` |
| 회원-2(회원가입) | 끝 | `router/user.go`(advisory lock 첫 회원 admin·bcrypt·옛 sha256 재저장·정지 403·`userResponse`에 `id`·`workerLimit`) · `pages/Signup.vue` |
| 회원-3(장비 등록·키) | 끝 | `router/device.go`(`/devices`) · `router/deviceKey.go`(키 형식·해시·AES-GCM) · `constants/env.go` `KEY_SECRET` · `pages/Devices.vue` · `feature/member/` |
| 회원-4(worker 접속 인증) | 끝 | worker `constants/env.go` `API_KEY`(로그엔 `***`) · `protocol` `RegisterRequest.apiKey` · supervisor `register.go` `checkWorkerKey`(주인 정지도 거부) |
| 이벤트 검사 | 끝 | `process.go` `output`·`screen`·`status`·`editResult`를 `(&auth)` 클로저로, `ownsProcess` |
| 회원-6(실행 권한) | 끝 | `device.go` `requireDeviceOwner` <- 폴더 만들기·`patchNode`·`resolveMainKey`·`requireDeviceFolder` / 보이는 범위 `ListVisibleWorkerInstances`·`GET /workers` 내 장비만 |
| 회원-5(관리자) | 끝 | `router/admin.go`(`/admin`, `blockedUsers`·`cutOff`) · `pages/Admin.vue` |
| 라벨 표시 | 끝 | `DeviceTile.vue` 머리 줄 · `NodeDeviceDialog.vue` · `WorkerPickDialog.vue` · `FolderTileBody.vue` 툴팁 · `GET /workers` label |
| 진입 | 끝 | 로그인 "회원가입" 링크 · 보드 메뉴 "내 장비"·"관리자"(admin만) · 라우트 `/signup`(public)·`/devices`·`/admin`(`meta.admin` 가드) |
| **UI 개편** | **시안 확정, 코드 다음 세션** | `REF-member-ui.md` |

- 장비 이름 규칙 `[A-Za-z0-9._-]{1,64}`(기존 메인키 `win-dev`·`fa24-PID`·`irony-MAC-ADDress1`·`imx8mp-f24` 모두 해당) / 키 = `nxk_` + base64url(32B)
- 삭제 순서: process -> `nodes.created_by`를 노드 주인으로(NOT NULL) -> nodes -> users(장비·공유·타일 트리 CASCADE)
- 한계: 삭제된 회원의 옛 쿠키는 supervisor 재시작 뒤 `blockedUsers`에 없음(정지만 DB에서 다시 채움) — `GET /users/session` 401이라 화면은 로그아웃, id 재사용 없음

## UI 개편 -> `REF-member-ui.md`
- 2026-10-26 시안 확정, 다음 세션에 코드 적용(UI-1(서버 목록)~UI-7(확인)). 분할 전 내용(결정·시안 2판·Grid N×M)은 전부 그 파일로 옮김

## 확인 결과 (격리, 전부 OK)
- 서버: 가입 7개(동시 첫 가입 5명 -> 관리자 1명·bcrypt 재저장) · 장비 8개(중복·형식·상한·남의 것 404·키 보기·붙여넣기) · 접속 인증 4개 + 재발급 뒤 옛 키 거부 · 이벤트 검사 3개 · 실행 권한(남의 장비 지정 403·공유받은 사람 실행 대상 보임·주인 바뀌면 옛 폴더 403) · 관리자 10개 + 정지·해제·삭제·재시작 뒤 정지 유지
- 화면(헤드리스 PC 1100·폰 390): 로그인·가입·내 장비·관리자 회원/장비 10장 + 흐름(가입 확인 칸·가입 -> 보드·일반 회원 `/admin` -> `/`·메뉴·잘못된 이름·등록 뒤 키/env 펼침·보드로)

## 격리 환경 다시 띄우기 (재부팅 뒤 — scratchpad가 사라지므로 아래대로 다시 만든다)
1. `docker start postgres15` — DB **`nexus_member_tmp`는 남아 있음**(관리자 `alice`/`pw`, 일반 `bob`·`erin*`·`fay*`, 정지 `dave`, 장비 `jig-a`(라벨 "3번 라인 PC" 인스턴스)·`line2-pc`)
2. supervisor: `cd /home/jh-bae/irony/nexus-member/apps/core && go build -o <scratch>/iso/sup/supervisor ./cmd/supervisor`, `dev.env`를 `PORT=5051`·`DB_NAME=nexus_member_tmp`·`IP=127.0.0.1`로 바꾼 사본을 `<scratch>/iso/sup/dev.env`에 두고 그 폴더에서 `env -u CLAUDECODE ./supervisor -env <절대경로>/dev.env`
3. worker(`jig-a` 접속 표시용): `go build -o <scratch>/iso/wrk/worker ./cmd/worker`, env = `WS_SCHEME="ws"`·`WS_HOST="localhost:5051"`·`SQLITE="<scratch>/iso/wrk/w.db"`·`NAME="jig-a"`·`API_KEY=<alice로 GET /devices/jig-a/key>` — sqlite가 새로 생기면 서브키가 새로 붙으니 라벨을 다시 달 것(`PATCH /devices/jig-a/instances/:subKey {"label":"3번 라인 PC"}`)
4. vite 3001: scratchpad `vite.iso.mts` = `import base from '/home/jh-bae/irony/nexus-member/apps/frontend/vite.config.mts'` 후 `(env)=>{const c=base(env); return {...c, root:'/home/jh-bae/irony/nexus-member/apps/frontend', server:{...c.server, fs:{allow:['/home/jh-bae/irony']}}}}` — **fs.allow가 없으면 링크된 node_modules의 아이콘 폰트가 403**. worktree `apps/frontend`에서 `VITE_API_URL=http://localhost:5051 DEV_PORT=3001 DEV_HOST=localhost env -u CLAUDECODE ./node_modules/.bin/vite --config <scratch>/vite.iso.mts`
5. 헤드리스: playwright = `/home/jh-bae/.npm/_npx/e41f203b7505f1fb/node_modules/playwright/index.mjs`(`import { chromium } from '<경로>'`). 로그인 = `request.post(API+'/users/session')` 후 `storageState`. 화면·API 둘 다 `localhost`(쿠키 호스트 맞춤)
- 사용자 직접 실행용 worker env(2026-10-26): worktree `apps/core/cmd/worker/test.env`(5051, sqlite `./test-worker.db`, `NAME`·`API_KEY` 빈 값 -> "내 장비" 등록 뒤 채움). 실행 = `go run ./cmd/worker -env test.env`. **git 제외 대상 아님 -> 파트 커밋 때 빼거나 지울 것**
- 끌 때: `/proc/<pid>/cwd`가 `<scratch>/iso/sup`·`iso/wrk`인 것, vite는 cwd가 worktree `apps/frontend`인 것만. `rm`은 `"${S:?}"/...` 또는 절대 경로(변수 빈 값 `rm` 안전장치에 걸림)
- 외부(test1·test2)로 보이기: supervisor가 8012에 떠야 하는데 **8012는 다른 프로젝트(`irony/community`) vite가 씀** -> 사용자가 밖에서 보겠다고 하면 방법을 따로 맞춤

## 다음 세션 시작 (2026-10-26 세션 끝, 배포 뒤)
1. 사용자가 test3(`build/take_process`)에서 본 결과를 듣는다 -> 고칠 것 있으면 worktree `feat/member`에서
2. 격리가 필요하면 위 "격리 환경 다시 띄우기"(전부 닫혀 있음, DB `nexus_member_tmp`만 남음 — scratchpad 경로는 새 세션 것으로)
3. 문제없으면 파트 마감(아래 3번)
- 이 세션 확인 스크립트(scratchpad `panel`·`url`·`auth`·`drag`·`wheel`·`regress`·`adopt1~3.mjs`, 공용 `lib.mjs`)는 세션이 끝나면 사라짐 — 무엇을 봤는지는 `REF-member-ui.md` "구현"·이 파일 "파트 마감까지 남은 것" 2번·`REF-supervisor-move.md` "가져오기 구현" 가져오기-5

## 파트 마감까지 남은 것
1. ~~UI 개편 코드 적용(`REF-member-ui.md` UI-1~UI-7)~~ — 끝(2026-10-26 사용자 브라우저 확인 "맘에 든다")
2. 마감 전 회귀 — **2026-10-26 격리 17/17**(scratchpad `regress.mjs`, 보드 탭을 열어 그 탭 id로 API): alice 폴더(`jig-a`)·스크립트·실행 대상·실행·실행 기록 / 장비 타일 목록·올리기·내려받기·장비 파일 실행 / bob에 실행 권한(20) 공유 -> 실행 대상 보임·실행 OK·`/workers`엔 남의 장비 안 보임 / **`devices`에 없는 옛 폴더(SQL로 `ghost-dev`) = 실행 403 "ghost-dev 는 이 폴더 소유자가 등록한 장비가 아닙니다"**, 실행 대상 목록도 403, 장비 타일 목록 403. 참고: 공유받은 bob의 장비 디렉토리 목록 403은 원래부터(`requireDeviceFolder` = 폴더 소유자만, Q5). **403 = 의도대로(2026-10-26 사용자 확인)** — 배포 뒤 기존 폴더는 주인이 "내 장비"에 같은 이름으로 등록해야 실행됨. 시험 폴더·`/tmp/regress-*`는 지움
3. ~~파트 커밋~~ — **`1bd58e7`**(2026-10-26, `feat/member`, 69파일. 뺀 것 = `cmd/worker/test.env`(디스크에 남김)·`apps/frontend/node_modules` 링크) -> `feat/supervisor-move` master 머지(`efbe403`) -> master를 `feat/member`에 머지(`fb2fa56`) -> 가져오기 `fa50c64` -> **사용자 test3 배포(`build/take_process`), 관리자 = `아이로니`(id 3)** -> 남은 것 = 배포 확인 뒤 파트 마감(MEMORY 불변 결정·`INDEX-member` 접기·master 머지). 배포는 `REF-member.md` "배포 전 확인" 순서대로(관리자 회원 묻기 필수)
