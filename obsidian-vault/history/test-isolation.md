# history — 격리 실행

> ↔ `REF-test-isolation.md` / 2026-11-11 `history/test-env.md`에서 옮김 / → `CURRENT.md`

## 2026-10-08 공유-e 격리 확인 함정
- 격리 프론트 설정은 `vite`를 import하지 않고 기존 설정 객체를 펼쳐 쓴다 / 계정 동기화로 소유자 보드에 타일이 이미 열려 있어 "새 타일"을 다시 누르면 선택자가 2개로 잡힘 / 헤드리스 계정은 `POST /users` + `/users/session` + `storageState`. → `REF-test-env.md`

## 2026-10-07 — 격리 실행 방법 정리
- 공유-b 인가-8(확인) 때 개발 DB·5050을 건드리지 않으려고 임시 DB + 5051 + 별도 worker로 띄움. `IsDev` 판정과 `env` 덮어쓰기, `TMPDIR` 우회, Python 소켓 클라이언트 → `REF-test-isolation.md` "격리 실행"

## 2026-10-07(2) — 개발 DB 사고 + 격리 확인 요령 추가
- 확인 명령 끝에 넣은 `go run ./cmd/supervisor --help`가 서버를 실제로 띄워 개발 DB `nexus`에 00007·00008 적용(데이터 보존, 그대로 둠) → `REF-test-env.md` 함정 줄 추가
- worker 포함 격리 실행, `websockets.sync` 소켓 클라이언트, `go build -overlay`로 쿼리만 확인하는 법 → `REF-test-isolation.md` "격리 실행"

## 2026-10-07(작업공간-e) — 격리 프론트 3001 방법 + 함정 2개
- 소켓 주소가 코드에 박혀 있어 scratchpad `vite.iso.mts`(transform 플러그인)로 5051을 보게 함 → `REF-test-isolation.md` "격리 실행"
- 함정: `PATCH /nodes` 본문 `{valid, value}` / `pgrep -f` 패턴이 자기 셸 명령줄에 걸려 셸이 죽음(exit 144)
