# REF — supervisor 주소 실시간 변경

> → 이력 `history/supervisor-move.md` · → 진행 `CURRENT.md`
> 파트(supervisor 주소 실시간 변경), 2026-10-24 착수. 용도 = supervisor 이사·업데이트, 임시로 연결 유지
> **상태: 마감(2026-10-26) — 집 서버 B 빌드로 운영 확인, master 머지**(`40e0942` wip가 구현 전부)

## 결정 (2026-10-24 조율)

| 질문 | 결정 |
|---|---|
| 주소-Q1(바꾸는 주체) | supervisor가 worker에 명령 |
| 주소-Q4(용도) | supervisor 이사·업데이트, 임시로 연결 유지 |
| 사용자 모델(주소 목록) | 명령 = worker가 옛 주소 + 새 주소를 목록으로 저장. 끊기면 목록을 주기적으로 돌며 시도, 붙은 주소를 1순위로 올림 (Q2(저장)·Q3(실패 시) 흡수) |
| 주소-Q9(즉시 옮김 여부) | 명령에서 선택: 바로 옮기기 / 저장만 |
| 주소-Q10(새 주소 순위) | 명령 직후 새 주소 = 1순위 |
| 주소-Q11(목록 정리) · Q13(N 값) | 최대 3개, 넘치면 꼴찌 삭제 |
| 주소-Q6(새 쪽 DB) | 스키마 같음, 데이터는 다를 수 있음 |
| 주소-Q12(새 쪽이 모르는 process) | 입양(새 DB 등록)·자동 종료 둘 다 안 함. **지금처럼 유지 + 화면에서 끄는 수단** |
| 주소-Q21(부를 이름) | **미등록 process** (= 붙은 supervisor DB에 행이 없는 process. "고아" 안 씀) |
| 주소-Q18(미등록 표시 정보) | worker `holds`에 실행 명령·폴더 열 추가해 같이 보임 |
| 주소-Q19(미등록 목록 자리·권한) | 장비 타일 메뉴. 보기·종료 = 그 장비를 등록한 계정 전부(`requireDeviceFolder`: 폴더 소유자 + 상속 포함 device_key == 메인키) |
| 주소-Q7(대상 범위) | 장비 하나 + 전체 둘 다 |
| 주소-Q8·Q20(권한) | 장비 하나 = 그 장비를 등록한 계정 누구나(Q19와 같은 판정, 사용자 확인 10-24) / 전체 = `worker_move` 권한 가진 계정만 |
| 주소-Q16·Q22(권한 형태) | `users.perms TEXT[]`, 전체 옮기기 = `worker_move` |
| 주소-Q23(권한 주는 법) | DB 직접 `UPDATE` (화면 없음) |

### 기각한 안 (다시 꺼내지 않게)
- 입양(새 DB에 process 등록): 소유자 계정·실행 명령 저장·화면 자리가 다 필요해 범위 큼
- 자동 종료: 데이터 다른 supervisor로 임시 이동하면 process(사용자 Claude 세션 포함)가 전부 꺼짐. 되돌릴 수 없음

## 지금 코드 (착수 전 조사)
- worker 주소 = env `WS_SCHEME`·`WS_HOST`를 시작 때 한 번 읽어 URL 고정, 재접속 루프 내내 같은 URL (`cmd/worker/main.go`)
- 서브키 = worker sqlite `identity`(메인키당 1행). supervisor `register`는 받은 서브키를 검사 안 함 -> 다른 supervisor로 가도 서브키 그대로
- 재접속 보고 `SyncEntry` = uid·status·pid만. `reconcileReconnect`: DB에만 있음 = FAILED(502) / worker에만 있음 = 로그 남기고 무시(`process.go` "고아 process 무시")
- worker `kill`(MsgKill)은 worker `procs`에서 uid로 찾음 -> supervisor DB 행 없이도 worker 연결로 바로 보낼 수 있음
- worker `holds` = uid·kind·edit_path·sock·holder_pid·pid·created_at (명령·폴더 없음)
- 장비 명부 `worker_instances`는 register 때 `recordSeen` upsert -> 새 supervisor에도 자동 등록
- `users` = identification·password·nickname만, 역할 없음. 다음 supervisor 마이그레이션 = `00013`, worker = `00003`

## 작업 단위 (계획)

| 단위 | 내용 | 손댈 곳 |
|---|---|---|
| 주소-1(주소 목록) | worker sqlite `supervisors(url, rank)` 최대 3. 비었으면 env 값으로 채움. 재접속 루프 = 순위대로 한 번씩, 한 바퀴 다 실패하면 backoff. 붙으면(register 성공) 그 주소 1순위로 | worker 마이그레이션·query·`main.go` |
| 주소-2(옮기기 명령) | `MsgMove`(sup -> worker REQ: url, now). worker = 새 주소 1순위 저장(3개 넘으면 꼴찌 삭제), now면 지금 연결 닫음. 응답에 저장 뒤 목록. 목록만 읽는 REQ도 | protocol·worker router·supervisor REST |
| 주소-3(권한 열) | `00013` `users.perms TEXT[] DEFAULT '{}'`, 로그인·세션 응답에 perms | supervisor 마이그레이션·user.go·프론트 auth |
| 주소-4(미등록 process) | worker `holds`에 cmd·cwd 열(`00003`과 같은 마이그레이션 가능), `SyncEntry`에 kind·cmd·cwd·시작 시각. supervisor = 인스턴스별 미등록 목록(메모리), 끝남(status)·끊김에 뺌. REST 목록·종료(MsgKill 직접) | worker hold·protocol·supervisor reconcile·REST |
| 주소-5(화면) | 장비 타일 메뉴: "supervisor 주소 변경"(주소 + 바로 옮기기/저장만, 지금 목록 보기) · "미등록 process"(명령·폴더·종류·pid·시작 시각 + 끄기). 헤더 메뉴(`BoardMenu`): "전체 worker 주소 변경"(`worker_move`만 보임, 접속 중 worker 전부) | 프론트 |
| 주소-6(확인) | 격리: supervisor 2개(5051·5052, 임시 DB 2개) + worker 1. 바로 옮기기·저장만 뒤 옛 쪽 끄기·새 쪽 실패 시 되돌아감·3개 제한·재시작 뒤 1순위·미등록 목록·끄기·권한 | scratchpad 스크립트 |

순서: 주소-1(주소 목록) -> 주소-2(옮기기 명령) -> 주소-3(권한 열) -> 주소-4(미등록 process) -> 주소-5(화면) -> 주소-6(확인)

## 주의
- 확인은 격리 환경만. 8013(사용자 세션)·`build/keep` 보관 프로세스 건드리지 않음
- 미등록 process는 worker가 다시 원래 supervisor에 붙으면 거기선 등록 process -> 그대로 이어 붙음(목록은 붙은 supervisor 기준)

## 구현 (2026-10-24, 브랜치 `feat/supervisor-move`)
| 단위 | 결과 |
|---|---|
| 주소-1(주소 목록) | worker `00003_supervisors.sql`(`supervisors(url, rank)` + `settings`) · `router/addrs.go` `AddrBook`(`LoadAddrBook`·`List`·`Promote`, 저장 = `StorePool.Tx`로 지우고 다시 넣기) · `main.go` `connectRound`(순위대로 한 번씩, 붙었던 세션이 끝나면 1순위부터, 한 바퀴 다 실패하면 backoff). 접속 시도 상한 10초(`dialer.HandshakeTimeout`) |
| 주소-2(옮기기 명령) | `MsgMove`·`MsgAddrs`(protocol) · worker `router/move.go`(Now면 300ms 뒤 연결 닫음) · 주소 검사 `protocol.NormalizeSupervisorAddr`(양쪽 같은 함수) · supervisor `router/workerMove.go` REST: `GET /workers/:m/:s/addrs` · `POST /workers/:m/:s/move` · `POST /workers/move-all` |
| 주소-3(권한 열) | `00013_user_perms.sql` `users.perms TEXT[]` · `userResponse.perms` · `requirePerm`(DB에서 다시 읽음) · `permWorkerMove = "worker_move"` |
| 주소-4(미등록 process) | worker `00004_hold_cmd.sql`(holds `label`·`cmd`·`cwd`) · `ProcessSpec.Label`(supervisor `launch`가 스크립트 노드 이름, 장비 파일 실행은 빈 값) · `SyncEntry`에 kind·label·cmd·cwd·startedAt · supervisor `unregisteredSet`(reconcileReconnect에서 통째로 교체 / status 끝남·끊김에 뺌) · REST `GET /workers/:m/:s/unregistered` · `POST .../unregistered/:uid/kill`(MsgKill 직접) |
| 주소-5(화면) | `worker/component/SupervisorMoveDialog.vue`(instance 있으면 하나, 없으면 전체) · `UnregisteredDialog.vue` · 장비 타일 메뉴(루트에도) 두 항목 · 푸터 왼쪽 보드 메뉴(`BoardMenu`, 계정 표시 있는 곳)에 `worker_move` 계정만 "전체 worker 주소 변경" |

### 내가 정한 것 (사용자에게 알림)
- **env 주소 규칙**: 시작 때 env 주소가 지난번(`settings.env_seed`)과 다를 때만 그 주소를 1순위로. 같으면 옮긴 순위 유지 — 재시작마다 env 주소가 1순위로 돌아오면 옮긴 게 풀리기 때문
- 창 기본 = 저장만 / 미등록 목록은 열 때·끈 뒤(800ms) 다시 읽음, push 없음 / 끄기·강제 버튼 둘
- 이 기능 전에 띄운 process는 holds에 명령이 없어 "(명령 모름)"
- 스크립트 노드 process의 명령 = worker 임시 파일 경로 `/tmp/<노드 id>/<uid>.sh` -> 사용자 요청(10-24)으로 **스크립트 이름(label) 표시 추가**: 창 첫 줄 = 이름(없으면 명령), 둘째 줄 = 명령·폴더·pid·시작 시각

## 확인 (2026-10-24, 격리)
- 서버 시나리오 28/29 OK(T6b = 기대값 함정, 위 "스크립트 노드 명령") · 화면 8/8 OK
- 이름 표시 추가 뒤 다시: 서버 29/29(T6b = label `tick.sh` + 명령 = 임시 파일) · 화면 8/8(U2 = worker 재시작 뒤에도 `tick.sh` — 보관 목록에서 다시 읽음) → `REF-test-scripts-1023.md` "2026-10-24 주소-6"
- 확인된 흐름: 저장만이면 연결 유지 -> 옛 supervisor 끄면 새 쪽으로(서브키 그대로) · 새 쪽에서 미등록 2개 보임·다른 계정 403·끄기 · 권한 없으면 move-all 403, SQL로 주면 200 · 바로 옮기기로 A 복귀 시 남은 것 PROCESS·B에서 끈 것 FAILED · 없는 주소로 바로 옮기기 -> 옛 주소로 돌아옴 · 3개 제한 · worker 재시작 뒤 순위 유지·process 이어짐 · env 고치면 그 주소 1순위

## 사용자 확인 (2026-10-26)
- 집 서버로 옮기기 실제로 잘 됨(회사 test3 -> 집 서버). 머지 OK — 순서는 `CURRENT.md`(회원 관리 확인 뒤)
