# REF — process 살려 두기 (worker·supervisor 재시작에도 터미널 유지)

> 2026-10-22 신설. 이력 → `history/process-keep.md` / 현재 진행 → `CURRENT.md` / 바뀌는 불변 결정 → `MEMORY.md` "process 영속성"·"worker 자체 종료 ≠ 끊김" / 관련 → `REF-process-snapshot.md`(ring, "worker-side 이전 옵션") · `REF-worker-orphan.md`(종료 정리) · `REF-process-reconnect.md`

**상태: 2026-10-22 마감·커밋**(→ `INDEX-process-keep.md`). 살리기-1(보관 프로세스)~7(supervisor 주기 삭제) 구현·격리 40/40·test1 시험 배포·test3 재배포 확인. 미확인 K7(끝난 뒤 기다림) 10분 경과·K6(Windows)

## 요청 (2026-10-22 사용자)
supervisor·worker를 다시 띄울 때 실행 중인 터미널이 사라져 곤란 -> worker가 process를 sqlite로 관리하고, worker가 꺼져도 process는 살아 있다가, worker가 다시 켜지면 sqlite와 process를 매칭.

## 지금 상태 (원인)
| 재시작 | 지금 | 원인 |
|---|---|---|
| supervisor | process는 삶(worker 재접속 루프 + PENDING) / **화면 내용(ring)은 사라짐** | ring = supervisor 메모리(`REF-process-snapshot.md` S3) |
| worker | process 죽음 | 정상 종료 = 일부러 정리(`handleSignals`, 고아 스크립트 파트) / 비정상 종료 = PTY master가 닫혀 SIGHUP |

## 사용자 답 (1차 질문)
| 질문 | 답 |
|---|---|
| supervisor 재시작 때 process가 죽나 | 안 죽음. 화면 내용이 사라지는 게 문제 -> 같이 해결 대상 |
| 살릴 범위 | worker **비정상 종료까지** -> 안 다(fd 넘겨주기) 탈락 |
| 끌 때 정리 | 하나씩 정리도 되지만 "끌 때 전부 종료" 수단이 있으면 좋음 |
| Windows | 진행 예정. Linux 먼저가 효율적이면 Linux 먼저 |

## 방법 후보
| 안 | 내용 | 비정상 종료 | 판단 |
|---|---|---|---|
| 가(process마다 보관 프로세스) | process마다 보관 프로세스가 PTY master + ring 보유, worker는 소켓으로 붙음 | 삶 | 추천 |
| 나(보관 데몬 하나) | 보관 프로세스 하나가 모든 PTY | 삶 | 데몬 재시작 = 전부 죽음 |
| 다(fd 넘겨주기) | 옛 worker가 새 worker에 fd 전달 | 죽음 | 탈락(비정상 종료 포함) |
| 라(tmux) | process = tmux 세션 | 삶 | 외부 의존, Windows 불가 |

## 구조안 = 안 가(process마다 보관 프로세스) (2026-10-22 제시)
| 무엇 | 지금 | 바꾼 뒤 |
|---|---|---|
| PTY master | worker 메모리 | 보관 프로세스(process당 1개, `setsid`로 worker와 다른 세션) |
| ring(256KB + alt screen) | supervisor 메모리 | 보관 프로세스(worker가 없어도 출력을 계속 읽어 쌓음) |
| process 목록 | worker 메모리 | worker sqlite(uid, 소켓 경로, 보관 pid, 시작 시각) |
| 종료 코드 | worker | 보관 프로세스가 갖고 있다가 worker가 가져가면 스스로 끝남 |
- worker 재시작: sqlite 목록 -> 각 소켓에 다시 붙음 -> 기존 재접속 처리("live 상태 보고"). 못 붙으면 끝난 것으로 보고
- supervisor 재시작: 화면 복원(snapshot) 요청 = supervisor -> worker -> 보관 프로세스 ring
- Windows: Linux 먼저. 소켓(Windows 10+ AF_UNIX)·ConPTY도 같은 구조 -> OS 부분만 `_windows.go`

## 사용자 답 (2차 질문)
| 질문 | 답 | 반영 |
|---|---|---|
| 끌 때 동작 | 기본 = process 살려 둠 | "전부 종료" 수단은 미정(3차 질문) |
| 보관 프로세스 실행 파일·통신 형식 | 지금부터 최대한 같은 형식 유지, 크게 바뀔 때만 바꿈 | worker 실행 파일 1개(모드로 구분) + 연결 첫 메시지에 형식 버전 번호(나중 worker가 옛 보관 프로세스를 알아보게) |
| systemd | (되물음) nohup·`&` 실행은 문제없나 | 문제없음: 보관 프로세스는 `setsid`로 터미널 세션 밖. 터미널을 닫아 worker가 SIGHUP을 받아도 기본 동작(살려 둠)으로 끝남 |
| 추가 요청 | 끝난 process를 하루·N시간 지나면 sqlite에서 주기적으로 삭제 | 3차 질문 참고 |

## worker sqlite 현황 (2026-10-22 확인)
- 지금은 `identity`(메인키·서브키) 테이블 하나뿐. process 이력은 supervisor PostgreSQL에 있음

## 3차 (2026-10-22) — 사용자: "worker 테이블 추가 불필요, supervisor에서 종료되고 오래된 것만 주기 삭제"
- worker 테이블 없이 가능한 조건: worker가 **보관 프로세스 소켓 폴더를 훑어서** 다시 붙음(소켓 파일 이름 = process uid, 붙으면 살아 있음 / 못 붙으면 소켓 파일 지우고 끝난 것으로 보고). 소켓 폴더가 곧 목록 -> sqlite 불필요. 폴더 후보 `$XDG_RUNTIME_DIR`(tmpfs, 재부팅 때 비워짐) 아래 메인키별
- supervisor 주기 삭제 = 살려 두기와 별개의 새 범위. 걸리는 점: R2-a(끝난 process) — 보드에 터미널 타일이 있는 끝난 process는 목록에 남는다(`REF-tile-binding-rework.md`). 타일이 가리키는 process를 지우면 타일이 빈 uid를 가리킴. `process_shares`는 `ON DELETE CASCADE`라 같이 지워짐

## 4차 결정 (2026-10-22 사용자)
| 이름 | 결정 |
|---|---|
| K1(worker 목록) | **worker sqlite 테이블**로(소켓 폴더 훑기 안 씀). 행 = 살아 있는 process만, 끝나면 바로 삭제. 이점: 소켓이 망가져도 pid + 시작 시각으로 보관 프로세스를 찾아 정리 가능 |
| K2(supervisor 정리) | 끝난 process는 바로 지우지 않고 **주기 삭제, 끝난 지 24시간** |
| K3(타일 같이 정리) | 지우는 process의 터미널 타일도 같이 정리(안 ㄴ). 이번 파트에 포함 |
| K4(기본 종료 동작) | worker 종료 신호 = process 살려 둠. "전부 종료" 수단은 BACKLOG |
| K5(통신 형식) | worker 실행 파일 하나를 모드로 띄움, 첫 메시지에 형식 버전 번호, 형식은 최대한 유지 |
| K6(Windows) | Linux 먼저, OS 부분만 나중에 `_windows.go` |
| K7(끝난 뒤 기다림) | 스크립트가 끝났는데 worker가 없으면 보관 프로세스는 **10분**만 기다리고 스스로 끝남. 그 뒤 worker가 켜지면 행은 있는데 소켓 없음 -> "끝남"(종료 코드 없음)으로 보고 |

## 실행 방식 (2026-10-22 사용자 질문으로 정리)
- 사용자는 worker를 지금처럼 한 번만 실행. worker가 process마다 자기 실행 파일을 `-hold <uid> ...`로 띄움(`os.Executable()`)
- 구조: worker -> `worker -hold`(보관 프로세스, `setsid`) -> 스크립트. worker가 꺼져도 보관 프로세스·스크립트는 남음
- 보관 프로세스가 끝나는 경우: 스크립트 끝남(kill 포함) + 종료 코드를 worker가 가져감 / K7(끝난 뒤 기다림) 10분 경과 / 재부팅 / `pkill -f 'worker -hold'`

## 구현 (2026-10-22, Claude 작성) — 살리기-1~7
| 단계 | 파일 | 내용 |
|---|---|---|
| 살리기-1(보관 프로세스) | `internal/execute/hold/`(`proto.go`·`server.go`·`client.go`·`spawn.go`·`spawn_unix.go`·`spawn_windows.go`) | 소켓 위 JSON 줄(`Msg`, `Version = 1`). 보관 프로세스 = `pty.ExecInteractive` + `ring.Screen`(256KB). 붙은 worker는 하나(새로 붙으면 옛 연결 끊음), 붙으면 hello(상태·pid·종료 코드) + 스냅샷 먼저. 출력 쓰기·전송·스냅샷은 한 락(offset 순서). 끝남은 남은 출력 뒤에, `DoneWait` 10분 또는 release로 끝남. 쓰기 5초 막히면 연결 끊음 |
| 살리기-2(worker 테이블) | `internal/worker/db/migrations/00002_holds.sql`, `query/holds.sql` | `holds(uid, kind, edit_path, sock, holder_pid, pid, created_at)` |
| 살리기-3(연결 교체) | `cmd/worker/router/hold.go`(새), `process.go` | `spawn`(띄움 -> 붙음 -> procs·sqlite) / `forward`(process당 고루틴 하나가 data·snap·status를 순서대로 MsgData(Off)·MsgScreen·MsgStatus로) / `finish`(종료 보고는 register를 마친 conn이 생길 때까지 기다림 -> EDIT 회수 -> 목록 삭제 -> release). `shutdown.go` 삭제 |
| 살리기-4(재시작 복구) | `hold.go` `Recover`, `main.go` | 시작할 때 sqlite 목록에 다시 붙음. 못 붙으면 cmdline으로 보관 프로세스인지 확인 후 죽이고 소켓·로그 삭제 + 목록 삭제(supervisor는 sync에 없는 uid를 FAILED 502로) |
| 살리기-5(종료 신호 동작) | `cmd/worker/main.go` | `-hold <소켓>`이면 init(env·sqlite) 건너뛰고 `hold.Main`. 종료 신호 = 로그 후 `os.Exit(0)`(SIGHUP 무시 판정 유지) |
| 살리기-6(화면 내용 이전) | `internal/ring/screen.go`(bind에서 옮김), `bind/relay.go`, `execute/agentInteractive.go`, `router/process.go`, 프론트 `processTerm.store.ts` | offset = 보관 프로세스가 매긴 값. supervisor ring은 캐시: `WriteAt`(이미 가진 것 버림, 빈 구간이면 비우고 이어 씀)·`Seed`(MsgScreen 스냅샷으로 통째로, 새로 생겼을 때만). AgentInteractive 출력 = `Chunk{Data, Off, Seed, Alt}`라 스냅샷도 출력과 같은 줄로 relay. register는 MsgSync 뒤 각 보관 프로세스에 스냅샷 요청. 프론트 `writeLive` = lastOff로 겹친 앞부분 버림·빈 구간이면 안내 줄 |
| 살리기-7(supervisor 주기 삭제) | `router/processPurge.go`, `processes.sql`·`tileTrees.sql` | 시작할 때 + 1시간마다: 끝난 지 24시간 지난 uid -> 그 터미널 타일이 있는 계정마다 트리 전부 잠가 타일 빼고 push -> 아직 타일이 남은 uid는 이번엔 남기고 나머지 `DELETE`(공유는 CASCADE) |

- 소켓 폴더 = `$TMPDIR/nexus-hold-<uid>/<메인키>`(재부팅 때 빔). unix 소켓 경로 108바이트 상한 -> 폴더가 길면 worker가 시작 때 에러(`HoldDir`)
- 보관 프로세스 stderr = 소켓 옆 `.log`, 정상 종료하면 지움

## 확인 (2026-10-22, 격리 supervisor 5051 + 임시 DB `nexus_keep_tmp` + worker `keep-dev`)
scratchpad `keep/keep_test.py` **40/40**(명세 → `REF-test-scripts-1020.md` "2026-10-22 process 살려 두기")
| 시나리오 | 확인 |
|---|---|
| basic | 보관 프로세스·sqlite 행 생김, 출력에 off, kill -> FAILED 143, 보관 프로세스·행·소켓·로그 사라짐 |
| worker(TERM·KILL 각각) | worker 꺼져도 보관 프로세스 삶, PENDING -> 다시 켜면 PROCESS, 열린 화면의 tick 번호가 꺼진 동안 것까지 빈 곳 없이 이어짐, 스냅샷도 이어짐 |
| supervisor | 재시작 뒤 PROCESS, 스냅샷 404 아님(tick 1부터), 열린 화면 겹침·빈 번호 없음 |
| done | worker 꺼진 동안 끝난 스크립트: 보관 프로세스가 기다림 -> 다시 켜면 종료 코드 7 전달·정리 |
| lost | worker 꺼진 동안 보관 프로세스를 죽임 -> FAILED 502, 행 삭제, 스크립트도 끝남. (처음엔 소켓·로그가 남음 -> `hold.Cleanup` 추가 후 다시 확인) |
| io·edit | 입력·resize(PTY 30x100)·EDIT(vi `:wq`) 회수 저장 |
| purge | finished_at을 25시간 전으로 -> 재시작 때 삭제·타일도 빠짐, 최근 것·그 타일은 남음 |

## 시험 배포 확인 (2026-10-22 밤, test1 = `build/keep` 8011, DB `nexus_keep`)
- Claude: test1 경유 로그인·소켓·스크립트 실행(출력 10줄, exit 0, 끝난 뒤 보관 프로세스 안 남음)
- 사용자 터미널 2개 띄운 상태에서 worker SIGTERM -> 보관 프로세스 2개 생존 -> 새 worker `다시 붙음` 2건 / supervisor SIGTERM -> worker 2초 뒤 재접속, 브라우저 스냅샷 성공
- **사용자 확인: 화면 이어지고 입력 됨**
- 알게 된 것: supervisor DB의 `device_key` = `<메인키>#<인스턴스키>`이고 인스턴스키는 worker sqlite마다 다름 -> sqlite만 다르면 같은 DB여도 reconcile(502 종결)은 서로 안 건드림. 복사해 온 8013 process는 test1에서 PROCESS로 남고 스냅샷만 실패(버퍼 없음)
- 다음: 8013 재배포 → 아래 "test3 재배포"

## test3 재배포 (2026-10-22 21:36, 8013 = `nexus_build`)
| 번호(이름) | 결과 |
|---|---|
| 재배포-1(대상 찾기) | `pgrep -f '(supervisor\|worker) -env'` + `readlink /proc/<pid>/cwd` -> cwd `build/iutf` 2개만 대상 |
| 재배포-2(옛 것 끄기) | 옛 worker -> supervisor SIGTERM. 옛 정책이라 test3 터미널 1개(노드 223 bash)가 같이 끝남 — 사용자 승인 뒤 |
| 재배포-3(새 것 띄우기) | `cd build/keep && setsid ./supervisor -env ../supervisor.env` -> 2초 뒤 worker 같은 방식. 같은 바이너리를 env만 바꿔 두 번 띄움(test1 = `./*.env`, test3 = `../*.env`) |
| 재배포-4(예상 부작용) | purge 173개. worker sqlite(`build/worker-sqlite.db`)에 `00002_holds`가 이때 처음 적용(이전 기록의 "적용됨"은 test1 sqlite 얘기였음) |
| 재배포-5(확인) | Node 스크립트로 https://test3 경유: index.html·가입·로그인·소켓 TAB:ID·폴더/스크립트 생성·`/workers?nodeId=`·`/processes/exec` -> DATA 10줄·STATUS exit 0·보관 프로세스 안 남음. 사용자 "동작 확인" |

- 시험 스크립트 요점(scratchpad라 사라짐): 소켓 프레임은 **바이너리** JSON(`{k,id,t,e,d}`) -> Node `WebSocket`은 `binaryType='arraybuffer'` 후 디코드. 쿠키는 `new WebSocket(url, {headers:{Cookie}})`(Node 22). exec 응답 키는 `uid`(스냅샷 경로도 uid가 아닌 process id라 404 — 확인 항목에서 뺌)
- 주의: test1·test3 worker가 **같은 메인키라 보관 소켓 폴더 `/tmp/nexus-hold-1000/irony-MAC-ADDress1/`를 같이 씀**. 다시 붙는 대상은 각자 sqlite 목록이라 지금은 영향 없음(test3 기동 뒤에도 test1 보관 프로세스 생존) → BACKLOG
- 부작용: 재배포-3을 Claude가 띄워 test3 터미널에 `CLAUDE*` 변수가 상속됨 -> 그 안의 Claude가 transcript 저장 안 함 → `REF-build.md` "Claude가 띄운 빌드본"

## 미확인 · 주의
- K7(끝난 뒤 기다림) 10분 경과로 스스로 끝나는 경로는 시간 때문에 안 돌림(release 경로만 확인)
- Windows(`spawn_windows.go`: 새 프로세스 그룹 + 콘솔 없음, `IsHolder` = false라 망가진 보관 프로세스를 정리 안 함) — 빌드만 확인, K6(Windows)
- 반영: **supervisor·worker·웹을 같이 바꿀 것**(옛 worker는 off를 안 보내 supervisor가 직접 세므로, supervisor 재시작 뒤 프론트 lastOff와 어긋남). 빌드본 worker를 바꾸는 순간 옛 worker 위 process는 옛 정책대로 정리됨 — 그 뒤 새로 띄운 것부터 살아남음
- 브라우저 화면 확인은 사용자 몫
