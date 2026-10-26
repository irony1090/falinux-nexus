# history — 회원 관리: 구현·확인

→ `REF-member-impl.md` · 결정 이력 `history/member.md` · `CURRENT.md`

## 2026-10-26
- 출발점: worktree `/home/jh-bae/irony/nexus-member`, `feat/member` = `19cc8bd` + `feat/supervisor-move` 머지(`6f87465`, vault 충돌 4개 master 쪽)
- 회원-1(DB) `00014_members.sql`: `users.worker_limit`(3)·기존 회원 `worker_register` 부여·`devices`(이름은 `workers` 대신 `devices` — node `device_key`·`/devices` API와 맞춤)·`worker_instances.label`. 임시 DB up/down/up OK
- 회원-2(회원가입) 서버: 가입 = 아이디·비밀번호, 닉네임 = 아이디, `pg_advisory_xact_lock` -> 회원 수 0이면 `admin`. bcrypt + 옛 sha256 로그인 때 다시 저장. 확인 7개 OK(동시 첫 가입 5명 -> 관리자 1명)
- 회원-3(장비 등록·키) 서버: `/devices` GET·POST(새 발급/붙여넣기)·`/:mainKey/key` GET(보기)·POST(재발급)·DELETE(접속 중 409)·`/:mainKey/instances/:subKey` PATCH(라벨). 키 = `nxk_`+base64url(32B), 해시(sha256)+암호문(AES-GCM, env `KEY_SECRET` 필수 — 없으면 supervisor 안 뜸). 장비 이름 = `[A-Za-z0-9._-]{1,64}`(기존 메인키 4개 모두 해당). 확인 OK(중복·형식·상한·남의 것 404·라벨)
- 회원-4(worker 접속 인증): worker env `API_KEY` -> `RegisterRequest.apiKey` -> `checkWorkerKey`(상수 시간 비교). 확인 OK(맞는 키 접속 / 틀린 키·키 없음·미등록 이름 거부 / 재발급 뒤 옛 키 거부)
- 이벤트 검사(회원-Q14): `output`·`screen`·`status`·`editResult`를 `(&auth)` 클로저로 바꿔 process `device_key` == 이 연결 인스턴스 키일 때만 처리(`ownsProcess`, 메모리에 없으면 DB). `status`의 미등록 process 정리(supervisor-move)는 같은 인스턴스 것만. 확인 OK(등록 안 한 연결·다른 인스턴스 무시 / 그 인스턴스는 반영) — 임시 클라이언트 `cmd/zz_evtest`는 지움
- 회원-6(실행 권한): `requireDeviceOwner`(장비 주인 == 폴더·노드 소유자) — 폴더 만들기·`patchNode` 장비 지정 / `resolveMainKey`(실행·`?nodeId=`) / `requireDeviceFolder`(장비 디렉토리 전부). 보이는 범위: `GET /workers`(nodeId 없음) = 내 장비만, `/workers/instances` = 내 장비 + 볼 수 있는 폴더의 장비(`ListVisibleWorkerInstances`), 명부 삭제 = 내 장비만. 확인 OK(남의 장비 지정 403·공유받은 사람 실행 대상 보임·주인 바뀌면 옛 폴더 403)
- 회원-Q21(회원 삭제) = 사용자 (다): **정지 + 삭제 둘 다**. 정지 = `users.suspended_at`(00014에 추가) — 로그인 403·기존 세션 401(메모리 `blockedUsers`, 시작 때 DB에서 채움)·그 회원 장비 worker 등록 거부(`GetDeviceAuth`)·정지 순간 브라우저 소켓·worker 연결 끊음(`cutOff`). 삭제 = 살아 있는 process 있으면 409, 아니면 process -> `created_by`를 노드 주인으로(NOT NULL이라 비울 수 없음, 첫 시도 500으로 발견) -> nodes -> users(장비·공유·타일 트리 CASCADE). 자기 자신 정지·삭제·관리자 해제 금지
- 회원-5(관리자 API) `/admin`: 회원 목록·상한/등록 권한/관리자(다른 권한 유지)·정지·삭제 / 장비 목록(주인 닉네임)·대신 등록(상한 무시)·주인 바꾸기·삭제·키 보기·재발급·라벨. 확인 OK(일반 회원 403 등 10개 + 정지·해제·삭제·재시작 뒤 정지 유지)
- 한계: 삭제된 회원의 옛 쿠키는 재시작 뒤 `blockedUsers`에 없음 — `GET /users/session`이 401이라 화면은 로그아웃되고, id는 재사용 안 됨(BIGSERIAL)
- 화면 첫 판(사용자 "우선 만들고 실제 보면서 UI를 잡자"): `/signup`(아이디·비밀번호·확인, 가입 뒤 바로 로그인) · `/devices`(등록 수 n/상한·등록 폼(키 붙여넣기 접힘)·장비 카드) · `/admin`(탭 회원/장비, `meta.admin` 가드 = 관리자 아니면 `/`) · 공용 `feature/member/`(API `device.api.ts`, `DeviceCard.vue` = 키 보기·복사·env 두 줄 복사·키 바꾸기·인스턴스 라벨, `MemberPage.vue` 틀) · 로그인 화면 "회원가입" 링크 · 보드 메뉴 "내 장비"·"관리자"(admin만) · 라벨 표시(장비 타일 머리 줄·장비 지정 창·실행 인스턴스 고르기·폴더 장비 칩 툴팁, `GET /workers`에 label 추가) · `userResponse`에 `id`·`workerLimit` 추가
- 헤드리스 확인(PC 1100·폰 390): 화면 10장 + 흐름(가입 확인 칸·가입 -> 보드 이동·일반 회원 `/admin` -> `/`·메뉴 항목·잘못된 이름·등록 뒤 키/env 펼침·보드로) OK. 스크립트 scratchpad `pw/member.mjs`·`flow.mjs`·`signup.mjs`(playwright = `~/.npm/_npx/e41f203b7505f1fb/node_modules/playwright`)
- 격리 환경 함정: worktree `apps/frontend/node_modules` = 원래 폴더로의 심볼릭 링크 -> vite가 아이콘 폰트(`@fs/...`)를 403 -> scratchpad `vite.iso.mts`로 `server.fs.allow` 넓혀 띄움(코드 문제 아님)

- 재부팅 전 vault 정리: REF·history를 결정(`REF-member`)/구현(`REF-member-impl`) 두 쌍으로 분할, 사용자 요청으로 마지막 응답 원문을 `REF-member-impl.md` "다음 세션 첫 안내"에 보관

## 2026-10-26 오후 — 마감 전 회귀 · 커밋 · 배포
- 마감 전 회귀 17/17(`regress.mjs`, 보드 탭 id로 API): 실행·장비 타일(목록·올리기·내려받기·파일 실행)·공유 실행 OK. `devices`에 없는 옛 폴더 = 실행 403 -> 사용자 "의도대로"(질문: API 키 문제냐 장비 등록 문제냐 -> 장비 등록 쪽이라고 설명)
- 함정: `users.identification`은 sha256 해시라 아이디로 못 찾음(닉네임으로) · 실행 기록 API는 폴더 id
- 파트 커밋 `1bd58e7`(69파일, `test.env`·`node_modules` 링크 뺌) -> supervisor-move master 머지 -> master를 `feat/member`로 `fb2fa56`(vault = master 기준)
- 추가 요청 = 미등록 process 가져오기(→ `history/supervisor-move.md`) -> 커밋 `fa50c64`
- 배포: 사용자가 test3에 직접(`build/take_process`, goose v14). 원래 폴더에서 `git checkout feat/member`가 worktree 때문에 막힘 -> worktree 폴더에서 빌드 권장. 관리자 = 사용자 "irony 닉네임" -> 실제 닉네임 `아이로니`(로그인 아이디 `irony`, id 3) 확인 뒤 Claude가 SQL로 지정
