# REF — 파트(장비 상태 push): 장비 접속 현황을 소켓 알림으로
> → `history/device-push.md` · → `CURRENT.md` · 실시간 push 일반 = `REF-realtime.md`

## 결정 (2026-10-27 사용자)
| 항목 | 결정 |
|---|---|
| 계기 | 장비 접속 현황이 10초 REST 폴링이라 반영이 느려 불편 |
| 알림 내용 | **"바뀜" 신호만** `DEVICE:CHANGE {mainKey}` → 받은 화면이 기존 API로 다시 조회(공유 변경 `SHARE:CHANGE`와 같은 방식). 보이는 범위 판단은 기존 API가 맡음 |
| 받는 사람 | 접속 중 계정 중 **관리자 + 장비 주인 + 그 장비가 지정된 폴더를 볼 수 있는 계정**(perm>=10, `ListVisibleWorkerInstances`와 같은 범위). 삭제·주인 변경은 **바뀌기 전 주인**도 |
| 보내는 때 | worker 접속·끊김 · 인스턴스 라벨 변경(주인·관리자) · 장비 등록·삭제(주인·관리자) · 주인 변경(관리자) · 명부 인스턴스 지우기. 키 재발급·주소 일괄 변경·계정 정지/삭제는 worker 끊김·접속으로 따라옴 |
| 범위 | 사용자 접속 표시(presence, 5초 폴링) 빼고 전부 = 보드(장비 타일 접속 점·폴더 타일 접속 수) + 내 장비·관리자 장비 페이지 |
| 폴링 | 10초 → **60초 안전망**으로 남김 |
| 그 밖 | `SHARE:CHANGE`도 장비 목록 다시 조회(공유가 바뀌면 보이는 장비가 바뀜) · 소켓 재연결 때 1회 다시 조회 |
| 소켓 연결 위치 | **`App.vue`(로그인 중이면 어느 페이지든)** — 원래 보드 페이지(`pages/index.vue`)에서만 연결해서 `/devices`·`/admin`에 바로 들어오면 소켓이 없었음(보드에서 넘어오면 연결 유지). 사용자 승인 2026-10-27. 부작용 = 관리 페이지만 연 탭도 서버 "탭"(presence에 접속으로 보임, 보드에서 넘어온 경우와 같아짐) |

## 구현 (2026-10-27)
| 층 | 파일 | 내용 |
|---|---|---|
| protocol | `messages.go` | `MsgDeviceChange = "DEVICE:CHANGE"` · `DeviceChangeEvent{MainKey}` |
| 쿼리 | `workers.sql` `FilterDeviceWatchers` | user_ids 중 `'admin' = ANY(perms)` / 주인 / `nodes.device_key = main_key AND node_permission >= 10` |
| supervisor | `roster.go` `publishDevice(mainKey, also...)` | 후보 = `tabs.onlineUsers()` → 쿼리로 거름 → `also`(이전 주인) 합침 → `publishUsers`. **DB·레지스트리에 반영한 뒤** 부른다(받은 쪽이 바로 다시 조회하므로) |
| supervisor | `register.go` | `workers.Append` + `recordSeen` 뒤 `go publishDevice` |
| supervisor | `supervisorRouter.go` | 끊김: `workers.Remove` + `recordSeen(false)` 뒤 |
| supervisor | `device.go`·`admin.go`·`workerApi.go` | 등록·삭제·라벨·주인 변경·명부 인스턴스 지우기 = `AfterCommit`. `deleteDevice`는 `superdb.Device`를 받아 주인 id를 `also`로 |
| 프론트 | `worker/hook/deviceChange.hook.ts` `bindDeviceChanges(socket)` | `DEVICE:CHANGE`·`SHARE:CHANGE` → 300ms 묶어서 `invalidateWorkerRoster` + `invalidateMemberDevices`(`['MEMBER','DEVICES'|'DEVICE']`·`['ADMIN','DEVICES'|'DEVICE']`) / 재연결(두 번째 CONNECTED부터) 때 같은 재조회 |
| 프론트 | `App.vue` | `bindDeviceChanges(socket)` + `watch(auth)` 소켓 connect/disconnect(`pages/index.vue`에서 옮김) |
| 프론트 | `worker.api.ts`·`device.api.ts` | 폴링 60초 |

## 확인 (2026-10-27, 격리 5051 + `nexus_dp_tmp` + vite 3001)
- 받는 사람(계정 4개: a1 관리자 · o2 주인 · v3 폴더 공유받음(10) · x4 무관, 소켓 클라이언트 4개) **10/10**: 접속·라벨(주인/관리자)·끊김·명부 지우기 = a1·o2·v3 / 등록(주인) = a1·o2 / 주인 변경 o2→x4 = a1·o2·x4 / 관리자 삭제 = a1·x4 / 관리자 등록(x4 것)·x4 삭제 = a1·x4
- 보드(o2, 헤드리스): 폴더 접속 표시 켜짐·꺼짐 약 0.4초(300ms 묶음 포함), 알림당 명부 조회 1번
- 내 장비(o2)·관리자 장비 탭(a1) `/devices`·`/admin` 바로 열기: 접속·끊김 1.5초 안에 반영 — 소켓 위치 옮기기 전엔 안 됨(위 결정 표)
- 확인 안 한 것: 소켓 재연결 때 재조회(코드만)
- **사용자 직접 확인(2026-10-27)**: 같은 격리(계정 a1·o2·v3·x4, worker `dp-dev`를 `dp/on.sh`·`off.sh`로 끄고 켬, 둘 다 브라우저 재조회 1초 안) → "확인 끝" → **파트 마감**(→ `INDEX-device-push.md`). 정리 끝: 프로세스(exe 경로로 고름)·DB DROP·`/tmp/nexus-hold-1000/dp-dev`, 5051·3001 닫힘, 8013 그대로
- 함정: `setsid ... &`의 `$!` = setsid pid(worker가 아님) → 끄기는 exe 경로로 찾는다. `pgrep -f <문자열>`은 그 문자열을 담은 내 bash 명령도 잡는다
