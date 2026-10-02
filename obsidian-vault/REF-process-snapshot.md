# REF — 화면복원 스냅샷 (ring buffer)

> `REF-process-reconnect.md` "같은 세션 재접속=화면 그대로 복원" 항목의 구체화. 작업 이력 → `history/process-snapshot.md` / 현재 진행 → `CURRENT.md`
> 프론트 끼울 자리 = `processTerm.store.ts` `restore` 한 곳(새로고침·다른 탭 push·409 재적용 전부) → `REF-node-ui-sync.md` "4-c"
> **상태(2026-10-02): 결정 S1~S6 + C(ring + alt screen만 다시 그리기) 확정. 스냅샷-a~d 완료(서버 13/13·브라우저 10/10·사용자 PC 확인), 커밋 `3de3064`(서버)·`997d6cf`(프론트)** → 구현·확인 `REF-process-snapshot-impl.md` / 이력 `history/process-snapshot.md`(설계)·`history/process-snapshot-impl.md`(구현)
> **반드시 구현할 것(사용자 2026-10-01) — 순서 = 4 O(계정 동기화) 화면 바로 다음(4.5)**: 새로고침·다른 탭에서 터미널 화면이 비는 문제는 이것으로만 해결된다.

## 문제

새로고침·다른 탭에서 터미널 타일을 등록할 때(`restore`), 그 전까지의 터미널 화면(스크롤백)은 아무 데도 남아 있지 않아 복원이 안 된다. 지금은 `reconnected — earlier output not shown` 안내 줄만 쓴다. worker→supervisor→browser로 흐르는 출력을 어딘가 버퍼링해 뒀다가 복원 때 다시 써야 한다.

## 결정 — supervisor-side ring buffer, worker-side는 아님

- **왜 worker 저장이 아닌가**: worker는 "휘발"로 못박힌 설계(MEMORY 확정 결정 "process 영속성"). worker는 PTY 원시 바이트를 릴레이할 뿐 화면을 해석하지 않으므로, 스냅샷을 떠도 supervisor가 이미 보는 것과 같은 데이터.
- **왜 supervisor인가**: `bind.Relay.pumpOutput`(`internal/supervisor/bind/relay.go`)이 모든 출력 바이트가 지나가는 유일한 지점. 여기 append만 추가하면 새 protocol 메시지도 round-trip도 없다(코드에 TODO 마킹돼 있음).
- **ring(고정 크기)**: process당 상한 — 오래된 바이트는 밀려남. "전체 이력"이 아니라 "최근 스크롤백"이 목적.

## 결정표 S1~S6 (2026-10-02 사용자 전부 확정)

| 결정 | 내용 | 상태 | 대안·근거 |
|---|---|---|---|
| S1(이음매 처리) | `DataEvent`에 누적 offset(그 묶음이 끝나는 위치) 추가. 스냅샷은 `{data, offset}`. 프론트가 보관함 `held`에서 offset 이후분만 이어 쓴다. 빈 구간(보관함 첫 묶음 시작 > 스냅샷 offset)일 때만 안내 줄 | **확정** | 서버에서 conn별 일시 큐잉(`bind.CatchUp`) = Hub 개조 필요라 기각. 4-c에서 생긴 항상 켜진 `held`가 이 일을 대신함 |
| S2(전달 경로) | `GET /processes/:uid/snapshot` 신설, `restore`에서 실행 중인 uid마다 호출 | **확정** | lookup 응답 동봉 = 256개 × 256KB로 응답 과대 |
| S3(ring 위치·수명) | `ProcessEntry` 밖 router의 uid → ring 맵(`sizeOwners`와 같은 이유: Rebind(worker 재접속) 때 entry를 새로 만듦). 삭제 = `cleanupProcessTopic`(process 완전 종료). worker 끊김(Detach)은 유지 | **확정** | entry 안에 두면 Rebind마다 옮겨야 함 |
| S4(끝난 process) | **끝나면 버림**(새로고침하면 종료 줄만) | **확정** | 끝난 뒤 일정 시간 유지 = 메모리 수명 규칙 하나 추가 |
| S5(ring 상한) | process당 256KB | **확정** | 아래 스케일 표 |
| S6(화면 종류 판정) | 출력에서 alt screen 진입·해제 시퀀스를 보고 "지금 alt screen인가" bool을 ring 옆에 유지. 판정 위치 = `pumpOutput` ring 쓰기 자리 | **확정** | 프로그램 이름으로는 판정 불가(스크립트가 무엇을 실행할지 서버는 모름) |

## C(ring + alt screen만 다시 그리기) — 부분 갱신 프로그램 대응 (2026-10-02 확정)

**정정**: 이전 판에 "htop 같은 curses류는 최신 한두 프레임만 있어도 복원엔 충분"이라고 적었으나 **틀렸다.** ncurses는 처음 그릴 때만 전체를 보내고, 이후엔 **바뀐 칸만** 커서를 옮겨 가며 쓴다.

| 상황 | ring replay 결과 | 원인 |
|---|---|---|
| ring 안 넘침(출력 총량 < 상한) | 정상(처음 바이트부터 재실행) | 단 xterm 크기 = PTY 크기(`Record.Cols/Rows`)여야 함. 다르면 커서 위치 지정이 어긋남 |
| ring 넘침(htop은 금방) | 최근 바뀐 칸만 보이고 헤더·F키 줄 등 고정 부분은 빈칸 | 전체 그리기 바이트가 밀려남 |
| 위와 같음 | alt screen 진입·색·스크롤 영역 설정이 빠진 채 그려짐 | 모드 설정 시퀀스도 앞부분에 있었음 |

검토한 해결 방향:
| 방향 | 내용 | 판단 |
|---|---|---|
| A(강제 다시 그리기) | PTY 크기를 한 칸 바꿨다 되돌려 SIGWINCH -> ncurses 전체 다시 그리기 | 가볍지만 같은 크기 설정은 커널이 신호를 안 보냄, 다른 탭도 한 번 흔들림, 일반 로그엔 무효 |
| B(서버 쪽 터미널 상태 유지) | supervisor가 Go vt 파서로 화면 상태를 유지, 스냅샷 = 화면을 그리는 시퀀스 생성(tmux·mosh 방식) | 정확하지만 process마다 파서 비용 + 크기 변경 추적. A의 흔들림이 거슬리면 이쪽으로 |
| **C(ring + A 조합)** | 일반 출력 = ring replay / alt screen 중(S6로 판정) = A | **채택**. 4.6 EDIT(worker `vi` 편집)도 같이 해결 |

### S6(화면 종류 판정) 상세
- worker가 `TERM=xterm-256color` 강제(`cmd/worker/router/process.go:74`) -> htop·vi·less·top·nano는 시작 때 `ESC[?1049h`, 끝날 때 `ESC[?1049l`을 반드시 출력. 옛 형태 `?47h/l`·`?1047h/l`도 같이 본다.

| 마지막으로 본 시퀀스 | 판정 | 스냅샷 때 |
|---|---|---|
| `?1049h`(`?47h`·`?1047h`) | alt screen 중 | ring replay 생략 -> `ESC[?1049h`만 먼저 쓰고 강제 다시 그리기 |
| `?1049l` 또는 본 적 없음 | 일반 출력 | ring replay만 |

- **묶음 경계**: 시퀀스가 두 묶음에 걸쳐 올 수 있으므로 앞 묶음 끝 최대 7바이트를 들고 있다가 이어서 검사.
- **alt screen일 때 replay를 빼는 이유**: 다시 그리기가 화면 전체를 덮으므로 불필요하고, 잘린 바이트를 먼저 그리면 깨진 화면이 잠깐 보인다. htop 종료 후 돌아오는 일반 화면은 ring에 남은 만큼만 복원.
- **놓치는 경우(감수)**: alt screen 없이 커서 이동으로 전체를 그리는 프로그램(`clear` 후 직접 그리는 스크립트). 일반 출력으로 판정 -> ring 넘치면 깨질 수 있음.

## 흐름
```
server  : pumpOutput -> ring.Write(data) -> off += len, alt flag update (S6)
                     -> publish DATA {uid, data, off}
          GET /processes/snapshot/:processId
                     -> lock ring -> {data, off, alt, redraw, cols, rows} -> unlock
                     -> redraw ? (layout cols-1 -> wait output -> cols) : nothing
browser : restore -> lookup -> per running uid: snapshot
                  -> xterm resize to cols x rows
                  -> not truncated ? write data
                     : alt ? write ESC[?1049h (+ data if !redraw) : notice + data
                  -> held chunks: drop end <= off, cut chunk crossing off, write rest
                  -> held first start > off ? notice (gap)
```
- 다시 그리기 출력은 스냅샷 offset 뒤에 생기므로 live(`held`) 쪽으로 들어온다 — 이음매 규칙(S1) 하나로 같이 처리.
- 다시 그리기를 서버가 하는 이유: 프론트에서 하면 크기 우선권(소유 탭만 resize)에 걸림.
- ring 쓰기와 publish가 한 고루틴에서 순서대로 일어나므로 offset은 단조증가.

## 스케일 검토 — 문제없음

버퍼 수명 = process 수명(S3) -> 총 메모리 = "동시에 살아 있는 process 수 × 상한".

| 동시 process 수 | 64KB/개 | 256KB/개 | 1MB/개 |
|---|---|---|---|
| 100 | 6.4MB | 25.6MB | 100MB |
| 1,000 | 64MB | 256MB | 1GB |
| 5,000 | 320MB | 1.28GB | 5GB |

ring은 시간이 지나도 크기가 안 늘어나므로 htop을 24시간 켜도 메모리 동일(화면 품질 문제는 위 C 절). 이 규모면 supervisor 단일 프로세스 구조(WS 연결·고루틴·PG 풀)가 먼저 병목 — ring만 과설계할 실익 없음.
- **자잘한 흠집(후순위)**: ring 앞이 잘리면 ANSI 시퀀스 중간에서 시작해 첫 화면이 잠깐 깨질 수 있음(일반 출력 쪽).

## worker-side 이전 옵션 — 보류 (2026-07-16 검토)

- worker 끊김(PENDING) 중 복원 불가 -> 사용자 판단 "worker가 연결 안 됐으면 못 가져오는 게 맞다"로 장애물 아님.
- 진짜 장애물이었던 이음매 문제(유실·중복)는 당시 Hub 개조(`bind.CatchUp`)가 필요하다고 봤으나, **S1(offset + 프론트 `held`)로 풀리면 worker-side에도 같은 방식이 적용 가능**. 필요한 것: worker `RingBuffer`(가칭 `internal/ring`, `Write`/`Snapshot() (data, offset)`) + `procEntry.ring` + `MsgSnapshot`(REQ `{uid}` -> RES `{data, offset}`) + `DataEvent.Offset`.
- 지금은 supervisor-side로 진행.

## 구현·확인 → `REF-process-snapshot-impl.md`
작업 단위 스냅샷-a(서버 ring)·b(서버 API)·c(프론트)·d(확인), 파일별 구현, 확인에서 잡은 버그, 확인 항목·결과.
