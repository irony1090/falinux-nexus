# REF — 화면복원 스냅샷 구현·확인 (스냅샷-a~d)

> 설계·결정 S1~S6 → `REF-process-snapshot.md` / 이력 → `history/process-snapshot-impl.md` / 현재 진행 → `CURRENT.md`
> 2026-10-02 `REF-process-snapshot.md`에서 분할(설계 vs 구현).

## 작업 단위 스냅샷-a~d (2026-10-02)

| 단위 | 내용 | 파일 |
|---|---|---|
| 스냅샷-a(서버 ring) | 재사용 primitive `Ring`(고정 크기 + 누적 offset, `Write`/`Snapshot() (data, off)`, 락 보유) + alt screen 판정기(S6, 묶음 경계 7바이트 이월). router에 uid -> 화면 버퍼(ring + 판정기) 맵, `startRelay`가 넘겨주고 `cleanupProcessTopic`에서 삭제. `pumpOutput`: 버퍼에 쓰기 -> `DataEvent{..., Off}` 발행. `DataEvent.Off`는 `omitempty`(worker<->supervisor 방향은 안 씀) | `internal/ring/`(신규), `bind/relay.go`, `protocol/messages.go`, `router/process.go` |
| 스냅샷-b(서버 API) | `GET /processes/snapshot/:processId`(기존 `kill/:processId` 형태에 맞춤) -> 소유자 확인 -> `{data, off, alt, redraw, cols, rows}`. 버퍼 없음(끝남·supervisor 재시작) = 404. alt면 응답 뒤 다시 그리기 = `Inter.Layout(cols-1)` -> `Layout(cols)`, **DB(`UpdateProcessLayout`)·크기 우선권은 건드리지 않음** | `router/processApi.go` |
| 스냅샷-c(프론트) | `DataEvent.off` 수신, `held` 묶음에 off 같이 보관. `restore`: 실행 중 uid는 **스냅샷을 받은 뒤에** handle 생성(그 사이 출력은 계속 `held`로) -> xterm을 cols x rows로 -> alt면 `ESC[?1049h`, 아니면 data -> `held`에서 off 이후만 -> 빈 구간이면 안내 줄. 실패·404 = 지금 동작(안내 줄) | `processTerm.store.ts`, `process.api.ts` |
| 스냅샷-d(확인) | 아래 확인 항목 | 헤드리스 + 사용자 PC |

- 순서 = a -> b(서버 단독 확인: raw 소켓 + HTTP) -> c -> d. supervisor 재기동 필요(a·b).
- **안내 줄 규칙(제안)**: ring이 넘쳐 앞이 잘렸고(off > data 길이) alt가 아니면 `earlier output truncated`. 이음매 빈 구간은 `earlier output not shown`(지금 문구 재사용).
- 알려진 한계: supervisor 재시작 뒤엔 ring이 없어 PENDING·실행 중 process도 404 -> 안내 줄(메모리 휘발, 감수).

### 스냅샷-a(서버 ring) 구현 (2026-10-02, 미커밋)
| 파일 | 내용 |
|---|---|
| `internal/ring/ring.go`(+test) | `Ring`: 고정 크기, `Write(p) int64`(쓴 뒤 누적 off) / `Snapshot() ([]byte, int64)`(복사본). 동시성 안전 아님(호출자가 잠금). cap보다 큰 쓰기 = 끝 cap 바이트만 |
| `bind/screen.go`(+test) | `Screen` = ring + alt bool + 직전 끝 7바이트, 락 하나로 `Write`/`Snapshot() (data, off, alt)`. 판정 = 경계(tail + p 앞 7바이트) 먼저, p 안 마지막 매치가 우선. 파라미터 합친 형태(`?1049;25h`)는 안 봄 |
| `bind/relay.go` | `NewRelay(uid, inter, screen, publish)` — `pumpOutput`이 `screen.Write` -> `DataEvent{..., Off}` 발행. screen nil이면 off 0 |
| `protocol/messages.go` | `DataEvent.Off`(`off,omitempty`) |
| `router/screen.go` | `screens` 맵(`open`/`get`/`drop`), `startRelay`가 `open`, `cleanupProcessTopic`이 `drop` |

- build/vet/`go test`(ring 6케이스·alt 10케이스) 통과
- Rebind 직후 옛 relay가 남은 출력을 드레인하는 동안 새 relay도 쓰면 발행 순서가 off 순서와 어긋날 수 있음 — 묶음마다 자기 범위(off - len ~ off)를 알므로 프론트 거르기는 정확, 실시간 표시 순서는 지금과 같음(감수)

### 스냅샷-b(서버 API) 구현 (2026-10-02, 미커밋)
- `router/screen.go` `snapshotProcess`: 소유자 확인 = **DB**(`ListProcessesByUids`) — PENDING(worker 끊김)은 memory entry가 없어 `requireOwnedEntry`면 404가 됨. 버퍼 없음 = 404
- `redraw` 플래그(구조안에 없던 것 추가): alt인데 worker 끊김이면 다시 그리기를 못 함 -> `redraw=false`면 프론트가 data를 그대로 씀(best effort). 프론트 규칙 = alt면 `ESC[?1049h` 먼저, `redraw`가 아니면 data
- `redraw(entry)`: 고루틴, `Layout(cols-1)` -> `Layout(cols)`(cols <= 1이면 +1). DB·크기 우선권 무접촉. 한계: 그 사이 소유 탭 resize가 끼면 마지막 `Layout(cols)`가 덮을 수 있음(드묾, 감수)
- **확인에서 잡은 것 2개**: ① resize 전(`Record.Cols/Rows` = 0)에 다시 그리기 -> PTY가 0x0이 됨 -> 크기 모르면 `redraw=false` ② 두 `Layout` 사이 경합: TUI가 신호를 처리할 때 이미 원래 크기면 "안 바뀜"으로 보고 안 그림(resize API 직접 2회 호출 시험: 간격 0/50/200ms = 성공/실패/성공, 고정 대기로 못 막음) -> **바꾼 직후 offset 기준으로 출력이 나올 때까지(최대 500ms) 기다린 뒤 되돌림**. 대가 = 화면 두 번 그려짐(cols-1, cols). 기준을 스냅샷 off로 잡으면 그 사이 평소 갱신 출력으로 기다림이 바로 끝나므로 `Layout(cols-1)` 뒤 off
- 서버 단독 확인(scratchpad `snap/server.mjs`, 테스트 전용 계정 `snap-<시각>` + 폴더 `deviceKey=irony-MAC-ADDress1`): 1a off 연속 / 1b 끝 스냅샷 = 받은 출력 끝 256KB(400KB 출력) / 1c 넘침 표시 / 1d 출력 도중 스냅샷 이음매 / 1e resize 전 0x0 / 2z resize 전 redraw 안 함 / 2y resize / 2a alt·redraw / 2b 다시 그리기 F키 줄(5회) / 2c DB 크기 그대로 / 3a 다른 계정 404 / 3b 끝난 process 404 / 3c 없는 uid 404

### 스냅샷-c(프론트) 구현 (2026-10-02, 미커밋)
| 파일 | 내용 |
|---|---|
| `process.api.ts` | `ProcessSnapshot` 타입 + `getProcessSnapshot(uid)` |
| `processTerm.store.ts` | `DataEvent.off` -> 보관함 묶음 `{bytes, off}`. `writeHeld(uid, term, chunks, from)`: from(스냅샷 off) 이하 버림·걸친 묶음은 뒷부분만·첫 묶음이 from보다 뒤에서 시작하면 그 자리에 `earlier output not shown`. `register(proc, from?)`(exec는 from 없음 = 전부). `writeSnapshot`: xterm을 cols x rows로 맞춘 뒤 -> **안 잘렸으면(off = data 길이) data 그대로** / 잘렸고 alt면 `ESC[?1049h`(+ `redraw` 아니면 data) / 잘렸고 일반이면 `earlier output truncated` + data. `restore`: 실행 중 uid 스냅샷을 `Promise.all`로 받은 **뒤에** handle 생성(받는 사이 이 탭에서 등록됐으면 건너뜀). 스냅샷 실패면 예전 안내 줄 동작 |

- **구조안과 달라진 것 2개**:
  - ① alt여도 안 잘렸으면 data를 그대로 씀. 구조안대로 alt면 무조건 건너뛰면, alt screen 이전 일반 화면(htop 전 `echo` 등)이 사라져 htop 종료 뒤 빈 화면이 됨
  - ② **복원 중 입력 막기(`replaying`)**. 스냅샷 속 옛 질의(`ESC[6n` 커서 위치, `ESC[>c` 단말 종류 — vim이 냄)에 xterm이 다시 답하고, 그 응답이 `onData` -> `sendInput`으로 실행 중인 process에 입력으로 들어감. 복원 쓰기 전에 `replaying.add`, 마지막에 `term.write('', cb)` 콜백(앞선 쓰기가 다 해석된 뒤 불림)에서 해제. 검증 = 막기를 빼면 F1이 `after=2`로 실패
- type-check 통과

### 스냅샷-d(확인) — 헤드리스 (2026-10-02, 10/10 통과)
scratchpad `snap/browser.mjs`(재부팅으로 사라짐). 시나리오마다 계정 `snapb-<이름>-<시각>` 새로 만들고, node raw 소켓 탭이 exec + resize(크기 소유) -> 브라우저 탭(1600x1200)은 push로 받음 -> 새로고침.

| 항목 | 내용 | 결과 |
|---|---|---|
| A1 | push로 받은 터미널(다른 탭 실행) 번호 연속(`n=$i` 0.02초 간격) | 통과 |
| A2·A3 | 새로고침 뒤 번호 연속 + 안내 줄 없음 | 통과 |
| B1·B2 | htop(안 잘림) 새로고침 = F키 줄·헤더 / `q` 뒤 `before-htop`·`after-htop` 둘 다 | 통과 |
| C1·C2 | `seq 1 50000` 뒤 htop(잘림 -> alt + 다시 그리기) 새로고침 = F키 줄·헤더 / `q` 뒤 `after-htop` | 통과 |
| D1 | vi 새로고침 전후 화면 같음 | 통과 |
| F1 | `printf '\033[6n'` 뒤 `cat -v` — 새로고침 뒤 응답(`^[[..R`)이 늘지 않음(1 -> 1) | 통과 |
| E | 콘솔 에러 없음 | 통과 |

- 테스트 함정: ① 세션 쿠키 값에 `=`가 들어갈 수 있음 -> `split('=')` 말고 첫 `=`에서 자를 것(안 그러면 401, 우연히 일부만 통과) ② vim swap 파일이 남으면 E325 화면 -> 실행마다 다른 파일 이름 ③ vim 시작 화면의 파일 이름 메시지는 곧 상태 줄로 바뀜(판정 기준으로 쓰지 말 것)
- **사용자 PC 확인(2026-10-02, 통과)**: htop 새로고침 복원 / 다시 그리기 때 다른 탭 = "약간 반짝하는 정도"(감수) / 끝난 process 새로고침 = 종료 줄만(S4). ring 넘침 안내 줄은 스크롤백 1000줄보다 위라 보통은 안 보임
- **발견(기존 문제, 이번 범위 밖)**: 같은 터미널을 연 탭이 여러 개면 질의 시퀀스에 **탭마다** xterm이 응답 -> process가 응답을 N번 받음. 크기 소유 탭만 응답하게 하는 방안 등 → `CURRENT.md` 미해결
