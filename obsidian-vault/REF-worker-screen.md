# REF — worker 화면 송출 (아이디어)

> 2026-10-15 신설. **아이디어 단계(사용자 "갑자기 떠오른 생각") — 계획·코드 없음, 파트 후보.** 이력 → `history/worker-screen.md` / 현재 진행 → `CURRENT.md`

## 요구 (2026-10-15 사용자)
- worker 장비의 **데스크톱 화면(GUI)** 을 브라우저에서 보기 + **조작(마우스·키보드)까지**
- 대상 장비 형태(모니터 있는 데스크톱 / 모니터 없는 서버·임베디드)는 미정

## 방식 비교
| 방식 | 동작 | 판단 |
|---|---|---|
| 가 스크린샷 주기 전송 | worker가 1~5fps JPEG 캡처 -> 지금 경로(worker -> supervisor -> 브라우저) | 보기 전용이라 조작 요구에 안 맞음 |
| **나 VNC 터널 + noVNC(추천)** | worker 장비 VNC 서버(x11vnc 등)로 가는 TCP를 worker 연결로 터널링 -> 브라우저 noVNC | 조작 포함, 변경 영역만 전송. 필요: VNC 서버 설치·설정 + TCP 중계(터널) 기능 신설 |
| (나 보충) | 브라우저는 TCP 직접 연결 불가 -> noVNC가 RFB를 JS로 처리하고 WebSocket으로 받음. 보통 중간에 `websockify`(WebSocket <-> TCP) — Nexus는 browser -> supervisor -> worker -> `localhost:5900`으로 바이트를 해석 없이 넘기는 터널이 그 역할 | |
| 다 영상 인코딩(H.264) / WebRTC | ffmpeg 인코딩 스트림 또는 브라우저-worker 직접 연결 | 가장 부드럽지만 가장 복잡, WebRTC는 NAT 넘을 때 TURN 필요할 수 있음 |

## 스트리밍(다)으로 조작까지 할 때 (2026-10-15 사용자 "스트리밍 방식으론 안되나?")
- 스트리밍은 **화면 방향만** — 조작은 **입력 채널을 따로**: 브라우저 마우스·키 이벤트 -> 기존 소켓 또는 WebRTC DataChannel -> worker가 OS에 주입(Linux XTest/uinput, Windows `SendInput`). VNC는 둘이 한 프로토콜에 묶여 있음

| 비교 | 나 VNC 터널 + noVNC | 다 스트리밍(H.264) + 입력 채널 |
|---|---|---|
| 화면 전송 | 바뀐 영역만 | 영상 압축 스트림(30~60fps) |
| 맞는 화면 | 정적 데스크톱(설정·터미널·문서) | 움직임 많은 화면(영상·3D·애니메이션) |
| 지연 | 보통 | WebRTC 낮음(대략 0.1초 이하) / WebSocket + WebCodecs·MSE는 더 김 |
| worker 부담 | 낮음 | 인코딩 CPU(GPU 있으면 낮음) — 저사양 임베디드엔 부담 |
| 구현 | VNC 서버 + TCP 터널 | 캡처 + 인코더(ffmpeg/GPU) + 전송(WebRTC = Go Pion, NAT 넘으면 TURN) + 입력 주입 |

- 판단: 일반 데스크톱 관리 = 나(간단·충분) / 움직임·반응 속도가 중요 = 다(WebRTC). 어느 쪽이든 Hub와 다른 경로

## 어느 방식이든 걸리는 것
- **캡처 권한**: worker를 서비스로 띄우면 사용자 데스크톱 세션에 접근 못 함. Linux Wayland = 사용자 동의(PipeWire portal), X11 = 비교적 쉬움, 모니터 없는 서버 = 송출할 화면 없음. Windows 서비스(세션 0)도 사용자 화면 캡처 불가
- **터미널 출력과 같은 길**: Hub 출력 큐 상한 없음(`REF-realtime.md` "발견") -> 큰 데이터를 같은 소켓으로 보내면 터미널 출력이 밀림 -> 별도 소켓 또는 프레임 버리기 규칙
- **권한**: 공유 단계와 맞춤 — 보기(10) = 보기만, 조작(20) 이상 = 입력 허용
- Windows worker(`REF-windows-worker.md`)와 겹침: Windows VNC 서버(TightVNC 등) 또는 DXGI 캡처

## 정리된 구조 (2026-10-15 대화)
```
browser(noVNC) --WS(tunnel)--> supervisor --WS(tunnel)--> worker --TCP--> localhost:<port> (VNC server)
                               auth once, then copy bytes   copy bytes
```
| 구성 | 하는 일 | 바이트 해석 |
|---|---|---|
| 브라우저(noVNC) | RFB 처리 — 화면 그리기, 마우스·키를 RFB 메시지로 | 예(유일) |
| supervisor | ① 터널 열 때 **인가 1회**(장비 공유 단계 — process 출력의 "인가 = DB 1회 / 전달 = 메모리"와 같은 원칙) ② 해당 worker로 전달 | 아니요 |
| worker | 받은 바이트를 `localhost:<port>`로, 읽은 바이트를 supervisor로 | 아니요 |
| VNC 서버(x11vnc 등) | **worker와 별개 프로그램** — 캡처·입력 주입 | 예 |
- **터널마다 별도 WebSocket**(브라우저-supervisor, supervisor-worker 모두) — 기존 소켓에 얹으면 터미널 출력이 밀림

## 생명주기 (worker가 VNC 서버를 자식 프로세스로)
| 시점 | worker | x11vnc 예 |
|---|---|---|
| 첫 열기 요청 | 서버 없으면 실행, 빈 포트 + 루프백만 | `-localhost -rfbport <port> -display :0 -shared -nopw` |
| 다른 브라우저도 열기 | 같은 서버에 터널만 추가 | `-shared` |
| 마지막 브라우저 닫음 | 유예 뒤 종료(새로고침 대비, 탭 보관 60초와 같은 생각) | 종료 신호 |
| worker 죽음 | 고아 안 남게 같이 정리 | 프로세스 그룹 종료 또는 `-once` |
- VNC 비밀번호 없음(`-nopw`): 루프백 전용 + 사람 인가는 supervisor 터널 열 때
- 후보: 새 실행 종류(예: `SCREEN`)로 process 모델(상태 전이·끊김/재접속·kill 정책)을 재사용 — 출력이 터미널 대신 터널
- 바로 안 되는 경우: x11vnc = X11만(Wayland = `wayvnc`·GNOME 원격 데스크톱, 동의 창) / worker가 그 데스크톱 세션 사용자로 실행돼야 함(`DISPLAY`·`XAUTHORITY`, 서비스로 띄우면 없음) / Windows VNC 서버는 보통 서비스 상주 — 요청마다 띄우기 확인 필요

## 결정 (2026-10-15 사용자)
| 번호(이름) | 결정 |
|---|---|
| SC1(VNC 서버 준비) | **장비에 설치 전제**(가). 동봉(나)·worker 내장(다)은 장비가 많아져 설치가 번거로우면 재검토 |
| SC2(못 쓰는 장비) | worker가 **화면 기능 가능 여부를 보고** → 불가면 화면 공유 버튼 **비활성화**(+ 이유 표시) |

SC2(못 쓰는 장비) 세부(제안, 시작할 때 확정):
- 판정 = 설치 여부만이 아니라 **실제로 쓸 수 있는지**: ① VNC 서버 실행 파일 있음 ② 화면 세션 접근 가능(X11 `DISPLAY` + 권한, Wayland면 지원 서버 있음) → 이유를 나눠 보고(`설치 안 됨` / `화면 세션 없음`)
- 시점 = 등록 때 + 재접속 때(설치·로그인 상태가 바뀔 수 있음). 주기 재검사는 필요해지면
- 단위 = **worker 인스턴스**(`메인키#서브키`) — 같은 메인키의 인스턴스끼리 다를 수 있어 버튼 상태는 고른 인스턴스 기준(장비 고르기 흐름 재사용). 보고 자리는 Windows worker WW5(OS 표시)의 등록 메시지 확장과 같음

## 시작할 때 정할 것
- 대상 장비 형태 / VNC 서버를 worker가 띄울지(설치 전제) 사용자가 미리 띄워 둘지 / 터널을 기존 worker 소켓에 얹을지 별도 연결로 할지 / 화면 타일을 노드 트리 어디에 둘지(장비 단위)
