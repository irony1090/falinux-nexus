# history/process-snapshot — 화면복원 스냅샷 (ring buffer)

> 설계·재사용 지식 → `REF-process-snapshot.md` / 현재 진행 → `CURRENT.md`
> 재접속/세션 복원 본체 → `history/process-reconnect.md`

## 2026-07-16 — ring buffer 설계 논의 착수 (코드 변경 없음, 순수 설계 대화)

**출발점**: 새로고침하면 socket 재연결+REST 목록 조회는 이미 되지만, 그 전까지의 터미널 화면(스크롤백) 자체를 복원할 방법이 없다는 지적에서 시작.

**대화 흐름**
1. 사용자 제안(worker가 frontend에서 보던 내용을 파일로 저장) → 검토 후 **supervisor-side ring buffer**로 방향 수정 제시. 이유: worker는 휘발성 설계 원칙(MEMORY.md) + worker는 애초에 "화면"을 모름(원시 바이트만 릴레이) + supervisor는 이미 모든 출력이 지나가는 지점(`bind.Relay.pumpOutput`)을 갖고 있어 공짜로 버퍼링 가능.
2. **스케일 우려** 제기("사용자 많아지면 수백 개는 우습지 않나") → 메모리 계산표로 반박(process당 캡 × 동시 개수로 예측 가능한 상한, 수천 개도 수백MB~1GB 수준이라 문제없음. 더 큰 병목은 supervisor 단일 프로세스 구조 자체).
3. **htop처럼 고빈도로 화면을 다시 그리는 프로그램이면 터지지 않나** 우려 → ring buffer는 "시간이 지나도 크기가 안 늘어나는" 구조라 무관함을 설명. 오히려 TUI는 최신 프레임만 있어도 복원엔 충분해 특성이 잘 맞음. ANSI escape sequence 경계가 잘릴 수 있는 자잘한 렌더링 흠집은 별개 후순위 이슈로 분리.
4. **"나중에 worker로 부담을 옮길 수 있나"** 질문 → 가능하나 트레이드오프 지적: worker 끊김(PENDING) 중엔 supervisor가 worker에 물어볼 수 없어 복원 실패. 사용자가 **"worker 미연결 시 못 가져오는 게 맞다"고 명시적으로 판단** → 이 반론은 무효화됨.
5. **필요한 protocol 메시지 개수** 질문에 처음엔 "`MsgSnapshot` 1개면 충분"으로 답함.
6. 이어진 질문("snapshot 가져오는 사이에 process IO가 있으면?")에서 **진짜 문제 발견**: snapshot과 live 스트림 사이 이음매에서 유실 또는 중복 재생 race가 생김(구독 순서를 어느 쪽으로 잡아도 한쪽이 깨짐). supervisor-side는 버퍼 읽기·구독 시작이 같은 프로세스 안에서 원자적이라 이 문제 자체가 없는데, worker-side는 네트워크 왕복이 끼어서 못 피함 — **"메시지 1개면 된다"는 답을 정정**.
7. **"그 문제를 감당하고 진행한다면 객체 설계가 있나"** 질문에 스케치: `RingBuffer`(worker, offset 포함) + `procEntry.ring` 필드 + `DataEvent.Offset` 필드 + `MsgSnapshot`(offset 포함 응답)까지는 구체화됐으나, **`subscribe.Hub`가 구독자별 차등 라우팅(이 conn만 잠깐 큐잉)을 할 수 있는 구조가 아니라서 그 부분(가칭 `bind.CatchUp`)은 미완성으로 남김**.
8. 결론: 지금은 supervisor-side ring buffer로 진행. worker-side 이전은 이 이음매 문제(RingBuffer offset + dedup + Hub 개조)를 풀기 전까진 보류 — "커질 것 같은 주제"라 사용자 요청으로 별도 문서(`REF-process-snapshot.md`)로 분리 기록.

상세 설계는 `REF-process-snapshot.md` 참조.
