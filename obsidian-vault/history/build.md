# history — 빌드본 분리

> → `REF-build.md` / → `CURRENT.md`

## 2026-10-22 구상·조율
- 사용자 요청: build 폴더에 supervisor·worker(+sqlite)·웹 빌드, supervisor가 웹 직접 서빙(`WEB_DIR`), env 파일명 옵션
- B1(포트·도메인) 8013 / B2(DB 분리) `nexus_build` / B3(데이터 복사) 복사해서 시작 / B4(반복 배포) `build/{이름}` / B5(웹 경로 변수) `WEB_DIR` 확정
- B6(env·sqlite 출처): sqlite·PG 복사는 사용자가 수동, env는 직전 빌드에서 복사(첫 빌드는 개발 env로 생성) / B7(실행 방법): 사용자가 직접 실행
- 빌드-1(env 옵션)·빌드-2(웹 서빙)·빌드-4(빌드 스크립트) 구현, 빌드-3(같은 출처 API)는 코드가 이미 됨. 임시 DB + 5051로 확인 끝. 8013 사용 중(다른 프로젝트 vite) 발견. 확인용 `build/_check`은 삭제 권한 거부로 남음 -> 사용자가 지울 것
- 사용자가 env·sqlite를 `build/` 공통으로 옮김. Claude가 `nexus_build` 생성·덤프(건수 일치). 스크립트·SQLITE 경로 수정 제안
- 사용자가 `worker.env` SQLITE를 `../`로 고치고 `build/_check`으로 빌드본(8013) 재실행 확인 -> 8011·8012 끄고 세션을 빌드본 위에서 다시 띄우기로. 남은 것 = 스크립트 env 단계 제거·커밋
- 남아 있던 확인용 worker(이전 세션 scratchpad `wchk.env`, 5051로 연결)를 사용자 요청으로 종료. 빌드본 supervisor·worker(`build/_check`, 8013)만 남음
- env 내뱉기 조율(tag 로더 안 -> embed 내뱉기로 사용자 결정, 개발 `-env dev.env` 필수, `cmd/*/env` git 제외) 후 Claude가 작성: main.go 2개 `initEnv`·env/dev.env·gitignore·README·build.sh 2단계·worker SQLITE 검사 버그. scratchpad에서 빌드본·개발 경우 6가지 확인(`TMPDIR` 함정 발견)

## 2026-10-22(2) 마감
- 사용자가 `build/const`로 다시 빌드·재실행 -> Claude가 8013 응답·SPA 경로·worker 연결·PostgreSQL 연결 확인
- 사용자가 개발 env 파일 이름을 `dev.env`로 바꿈 -> Claude가 main.go 2개·README 2곳·`apps/core/.gitignore` 주석·vault를 `dev.env`로 맞춤, `go build ./...` OK
- 파트(빌드본 분리) 마감: INDEX-build로 접고 코드 + vault 한 커밋

## 2026-10-22(3) `CLAUDE*` 변수 상속 발견
- 사용자: test3 터미널에서 연 Claude에 "transcript saving is off - inherited CLAUDE_CODE_CHILD_SESSION marker"
- Claude가 조상 프로세스 추적: claude -> bash -> `worker -hold` -> test3 worker 4178326(`-env ../worker.env`). worker·supervisor 환경에 이전 세션 `bb6187d5`의 `CLAUDE*` 변수 -> 재배포-3(새 것 띄우기)을 Claude Bash 도구로 띄운 탓
- 규칙·해결 방법 → `REF-build.md` "Claude가 띄운 빌드본", 코드 수정은 BACKLOG

## 2026-10-23 test3 운영 (CURRENT에서 내림)
- 오전: 사용자가 test3를 자기 터미널에서 다시 켬(`CLAUDE*` 변수 0개) -> 계획한 PC 재부팅 -> test1(8011)은 다시 안 켬(test1.falinux.com 502), `build/iutf`는 사용자가 지움
- `build/worker-dir`를 옛 태그(`IUTF-261022`)로 빌드한 것을 사용자가 짚음 -> `build.sh`에 태그 출력·경고 추가(→ `REF-build.md` "빌드 태그") -> 사용자가 태그 `WORKERDIR-261023`으로 다시 빌드·재실행(supervisor 281076·worker 282039). 보관 프로세스(`build/keep/worker -hold`)는 남고 새 worker가 다시 붙음, 첫 시작 때 `nexus_build`에 00012 적용

## 2026-10-28 — 보드 worker 2대 mochi로
- 사용자 요청: `fa24-PID`(40.211)·`imx8mp-f24`(80.119) worker 최신화(supervisor = test3, 버전 = mochi). 둘 다 `WORKER-261026-TAKE_PROCESS` -> master `28053b7` arm64(`WORKER-MOCHI-261027`, sha256 `a86e7e41…`). 10:29·10:30 재시작 -> test3 등록 성공, 보관 process 다시 붙음(713·16119 유지). 옛 파일 `worker.old-20261028-1029*`/`-1030*`(+ 원래 있던 `worker.prev`) 남김 → REF-build "보드 worker 교체"

## 2026-11-12 test3 worker 교체 (옛 빌드 폴더 삭제)
- 증상(사용자): test3 worker로 process 실행 -> `worker가 실행을 거부했습니다: fork/exec /home/jh-bae/irony/nexus/build/261111-column-flow/worker: no such file or directory`
- 원인: 사용자가 옛 빌드라 생각해 `build/261111-column-flow/`를 지움(17:14~17:22 무렵). 그런데 test3 worker(36473, 10:04 사용자 터미널에서 시작)는 그 폴더 바이너리로 돌고 있었음(`/proc/36473/exe` = `(deleted)`). supervisor만 15:11 `261112-account`로 교체되고 worker는 그대로였음. `router/hold.go` `spawn()`이 `os.Executable()` 경로로 `worker -hold`를 띄우므로 새 process만 실패, 떠 있던 보관 프로세스는 메모리에 있어 정상
- 사용자 걱정: 이 Claude 세션이 그 worker 위에서 돎 -> 프로세스 트리 확인: Claude는 보관 프로세스 807976(자기 세션, `setsid`)의 PTY 안 -> worker SIGTERM이 전달 안 됨, worker 종료 = `os.Exit`만(K4(기본 종료 동작))
- 조치(Claude, 23:13): `kill -TERM 36473` -> `build/`에서 `env -u CLAUDE* setsid nohup ./261112-account/worker -env ../worker.env >> 261112-account/worker.log` -> pid 514088, `CLAUDE*` 0개, `[hold] 다시 붙음` 2개·등록 성공. 옛 보관 프로세스 2개(37855·807976)는 부모가 `systemd --user`로 바뀌고 지워진 옛 바이너리로 계속 돎. 작업 폴더는 빌드 폴더 -> `build/`로 바뀜(env·sqlite는 같은 파일 읽음 확인)
- 사용자 확인 "이제 된다"
- 재발 방지: `REF-build.md` "정리할 때"에 이미 있던 규칙이 빌드·배포 때 보는 표에 없어서 놓침 -> `INDEX-deploy.md` 표에 "옛 빌드 폴더 정리" 줄(지우기 전 `ls -l /proc/<worker pid>/exe` 확인)
