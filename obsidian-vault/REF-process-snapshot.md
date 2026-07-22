# REF — 화면복원 스냅샷 (ring buffer)

> `REF-process-reconnect.md` "같은 세션 재접속=화면 그대로 복원" 항목의 구체화. CURRENT.md "화면복원" 미착수 항목의 설계 대화.
> **코드 미착수 — 순수 설계 논의만.** 작업 이력 → `history/process-snapshot.md`.

## 문제

새로고침/재접속하면 socket이 재연결되고(`handleSubscribeWS`가 살아있는 process를 Hub에 자동 재구독) `GET /processes/subscriptions`로 UI 재구성은 되지만, **재접속 전까지의 터미널 화면(스크롤백)은 아무 데도 안 남아있어서 복원이 안 된다.** worker→supervisor→browser로 흐르는 출력을 어딘가 buffering해뒀다가 재접속 시 replay해야 한다.

## 결정 — supervisor-side ring buffer, worker-side는 아님

- **왜 worker가 파일로 저장하는 방식이 아닌가**: worker는 "휘발"로 못박힌 설계(MEMORY.md 확정 결정 "process 영속성: worker 휘발 / supervisor 영속"). 게다가 worker는 "frontend가 뭘 보고 있었는지" 자체를 모른다 — PTY 원시 바이트를 릴레이할 뿐, 화면 해석(스크롤백·커서)은 xterm.js(프론트) 몫. worker가 뭘 스냅샷 떠도 결국 원시 바이트라 supervisor가 이미 실시간으로 보는 것과 동일한 데이터.
- **왜 supervisor인가**: `bind.Relay.pumpOutput`이 이미 모든 출력 바이트가 지나가는 유일한 지점(worker→sup `MsgData` EVENT 드레인). 여기 append만 추가하면 새 protocol 메시지도, 새 round-trip도 없이 공짜로 버퍼링된다(코드에 이미 TODO 마킹돼 있음: `internal/supervisor/bind/relay.go` pumpOutput 주석 "ring buffer append / DB sink 배선 지점").
- **ring(원형/고정크기)이어야 함**: process 하나당 고정 상한(예: 128~256KB) — 오래된 바이트는 밀려나 사라짐. "그 process의 전체 이력"이 아니라 "최근 스크롤백"만 목적.

## 스케일 검토 — 문제없음

버퍼 수명이 `ProcessEntry` 수명과 같이 감(process 끝나면 entry가 즉시 GC되고 버퍼도 같이 사라짐 — `REF-process-trigger.md` "종료 후 Hub 구독 정리" 절 참조, 같은 원리). 그래서 총 메모리는 "동시에 살아있는 process 수 × process당 상한"으로 예측 가능하게 캡된다.

| 동시 process 수 | 64KB/개 | 256KB/개 | 1MB/개 |
|---|---|---|---|
| 100 | 6.4MB | 25.6MB | 100MB |
| 1,000 | 64MB | 256MB | 1GB |
| 5,000 | 320MB | 1.28GB | 5GB |

수백~수천 개 규모도 합리적 상한(수백 KB/개)이면 전혀 부담 없음. 이 정도 규모라면 ring buffer보다 **supervisor 단일 프로세스 구조 자체**(WS 연결 수, 고루틴 수, PG 커넥션풀 등)가 먼저 병목이 될 것 — ring buffer만 별도로 과설계할 실익 없음(YAGNI).

**고빈도 풀-리드로우 프로그램(htop 등)도 문제없음**: ring은 "시간이 지나도 크기가 안 늘어나는" 구조라 24시간 켜놔도 메모리 사용량 동일. 오히려 htop처럼 매초 화면 전체를 새로 그리는 프로그램은 같은 용량 안에 담기는 시간 폭이 짧아질 뿐(예: 몇 초~1분 치)인데, curses류는 "최신 한두 프레임"만 있어도 화면 복원엔 충분해서 오히려 잘 맞는 특성.
- **자잘한 흠집**: 바이트 단위로 앞을 잘라내면 ANSI escape sequence 중간이 잘릴 수 있어 replay 시작 순간 살짝 렌더링이 깨졌다가 다음 리드로우에 정상화될 수 있음 — 메모리/터짐 문제 아니고 품질 디테일, 후순위.

## worker-side 이전 옵션 — 검토했으나 지금은 보류

"나중에 문제 되면 worker로 부담을 옮길 수 있나?"는 논의에서 나온 대안. 두 단계로 평가가 바뀜.

**1차 평가 (기각 사유였다가 → 사용자가 허용 가능하다고 판단, 무효화됨)**
worker가 끊긴 동안(PENDING)엔 supervisor가 worker에 물어볼 수 없어 복원이 안 됨 — supervisor-side라면 worker 끊김 중에도 캐시된 버퍼로 복원 가능한 것과 대비됨. **사용자 판단: "worker가 연결 안 됐으면 못 가져오는 게 맞다"(정직한 UX, 억지로 stale 데이터 보여주는 것보다 낫다)** → 이 반론은 더 이상 worker-side를 막는 이유가 아님.

**2차 평가 (진짜 남은 장애물)**
worker-side로 옮기면 **snapshot(worker가 갖고 있던 과거분)과 live 스트림(그 이후 실시간분) 사이의 이음매** 문제가 생긴다:
- snapshot 먼저 뜨고 나중에 구독 → 그 사이 출력 **유실**
- 구독 먼저 하고 snapshot 나중에 뜨면 → 그 사이 출력이 snapshot과 live 양쪽에 다 포함돼 **중복 재생**

supervisor-side는 이 문제 자체가 안 생긴다 — "버퍼 읽기"와 "구독 시작"이 같은 프로세스·같은 락 안에서 원자적으로 처리 가능(네트워크 왕복이 없음). worker-side는 REQ/RES 네트워크 왕복이 끼기 때문에 이 race를 피할 수 없다.

**필요해지는 것 (worker-side로 갈 경우, 설계만 스케치됨 — 미완성)**
- `RingBuffer`(신규 재사용 primitive, 가칭 `internal/ring`): `Write(p []byte)` / `Snapshot() (data []byte, offset int64)`. **단조증가 offset**을 같이 들고 있어야 dedup이 가능.
- worker `procEntry`(`cmd/worker/router/process.go`)에 `ring *RingBuffer` 필드 — 기존 `pumpOutput` 루프 안에서 그냥 같이 Write(새 고루틴 불필요).
- 프로토콜: `MsgSnapshot`(신규, REQ sup→worker: `{uid}` → RES `{data, offset}`) + 기존 `protocol.DataEvent`에 `Offset int64` 필드 추가(각 배치가 끝나는 시점의 누적 offset을 실어 보냄).
- **핵심 미해결**: `subscribe.Hub`가 "구독시키면 전원에게 동일하게 즉시 fan-out"하는 모델이라, "이 conn 하나만 snapshot 받을 때까지 live 이벤트를 잠깐 큐잉했다가 offset 기준으로 dedup 후 합류"를 할 자리가 없음. 가칭 `bind.CatchUp`(uid+conn 단위, join 하는 동안만 사는 임시 객체 — pending 큐 + offset 비교)이 필요하다는 것까지만 나왔고, Hub의 `send` 콜백을 conn별로 일시적으로 갈아끼우는 메커니즘 자체는 **설계되지 않음**.

**결론**: worker-side 이전은 "메시지 하나 추가하면 끝"이 아니라 이 이음매 문제(RingBuffer offset + dedup + Hub 개조)까지 포함하는 진짜 엔지니어링. 지금은 supervisor-side로 진행하고, 이 문서를 이후 재논의 시작점으로 삼는다.

## 다음 (미착수)

1. `bind.Relay.pumpOutput`에 ring buffer append 배선(supervisor-side, 위 "결정" 절).
2. `handleSubscribeWS` 재접속 흐름에서 ring buffer 스냅샷을 새 conn에 SNAPSHOT으로 먼저 흘려보내고 그 다음 Hub 실시간 구독으로 전환 — 이 전환도 "버퍼 읽기 + 구독 시작"이 같은 락 안에서 원자적이어야 함(supervisor-side라 network round-trip 없이 가능하지만, 구현 시 순서는 여전히 신경 써야 함).
3. process당 ring 상한 기본값 확정(권장: 128~256KB, 위 스케일 표 참조).
