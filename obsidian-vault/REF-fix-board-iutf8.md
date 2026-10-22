# REF — 새 보드·IUTF8 (작은 수정 두 개)

**상태: 마감(2026-10-22) → `INDEX-fix-board-iutf8.md`.**

> → `history/fix-board-iutf8.md` / → `CURRENT.md` / 보드 설계 → `REF-workspace.md` W1(분리 범위) / 한글 입력 → `REF-term-ime.md`

## 목적
BACKLOG "같이 진행"의 두 항목을 짧은 파트 하나로 (2026-10-22 사용자: worker 폴더 타일보다 먼저)

## 수정-1(새 보드) — 어느 보드에서든 "새 보드로 분리"
- 원인: 버그가 아니라 W1(분리 범위) 때 정한 "**`/`(기본 보드)에서만 만듦**". 서버 `createWorkspace`가 기본 트리만 잠금(`GetTileTreeForUpdate`), 프론트 메뉴 `v-if="isDefaultBoard"`. 리팩터링과 무관(`3b80310`부터)
- 결정(2026-10-22 사용자): **어느 보드에서든 분리 가능**
- 서버 `cmd/supervisor/router/workspace.go`: `workspaceCreateRequest.From`(출발 hash). 비면 옛 동작(기본, `ensureTileTree` + `GetTileTreeForUpdate`) / 있으면 `GetTileTreeByHashForUpdate`, 없거나 남의 것 = 404. 트리 하나만 잠그므로 잠금 순서 규칙 대상 아님. 루트 폴더 타일 거부(400)는 `detachFolder` 그대로
- 프론트: `board.api.ts` `createBoard`에 `from`, `tileTree.store.ts` `detachToBoard`가 `hash_.value`(불러온 보드의 실제 hash) 전달, `TileFrame.vue` `isDefaultBoard`·"옮길 보드가 없습니다" 줄 삭제(분리 항목이 늘 있어 메뉴가 빌 일 없음)
- 응답 `hash` = 출발 보드 -> 프론트 `serverChange`가 지금 보는 트리를 그대로 갱신

## 수정-2(IUTF8) — PTY 줄 편집의 한글 백스페이스
- 원인: `internal/execute/pty/execInteractive_linux.go`가 PTY termios를 안 건드려 `IUTF8` 꺼짐 -> canonical 모드(`read`·`cat`·`head`)에서 백스페이스가 1바이트만 지움
- 수정: `openPty`에서 slave를 연 직후 `setUTF8` — `TCGETS` -> `Iflag |= IUTF8` -> `TCSETS`(파일의 기존 `syscall.Syscall` ioctl 방식). 실패하면 master·slave 닫고 에러. Windows(ConPTY)는 해당 없음(`GOOS=windows` 빌드 확인)
- 덤 관찰: IUTF8이 꺼진 채 bash `read`에 깨진 바이트가 들어가면 bash가 나머지 바이트를 기다리며 **멈춤**(Python `pty`로도 같음 = bash 동작) — 이번 수정으로 이 경우도 안 생김

## 확인 (2026-10-22)
- 수정-2(IUTF8): scratchpad `iutf8_test.go`를 `go test -overlay`로 `internal/execute/pty`에 끼워 실행(저장소에 파일 안 둠). `bash -c "stty -a | grep iutf8; head -n1 | od -An -tx1"`에 `가나` + DEL + 줄바꿈 -> 수정 후 `iutf8` · `ea b0 80 0a`(= 가) / 비교용 `stty -iutf8` 먼저 -> `ea b0 80 eb 82 0a`(나의 1바이트만 지워짐)
- 수정-1(새 보드): 격리 supervisor(5051, 임시 DB `nexus_newboard_tmp`, `-env iso.env` + `TMPDIR`) + scratchpad `newboard_test.py` **16/16**: from 비움 = 기본에서 분리 / 기본 아닌 보드에서 분리 200·응답 hash = 출발·출발 트리에서 빠짐·기본 그대로 / 루트 타일 400 / 없는 from 404 / 옛 version 409 / 다른 계정 보드 404 / 목록 3개. 끝나고 프로세스(cwd 확인)·DB 삭제
- 브라우저 확인(빌드본 `build/iutf`, 2026-10-22 17:25~): 수정-1(새 보드) **사용자 확인 OK**(기본 아닌 보드에서 분리 -> 옮겨짐) / 수정-2(IUTF8) 이 세션 PTY `stty`에 `iutf8` 켜짐
  - 수정-2 브라우저 1차 시도 `cat` + `가나` + 백스페이스 -> `가ㄴ`: 조합 중인 `나`에서 백스페이스를 **IME가 먹어 `ㅏ`만 지움**(PTY에 안 감) = 정상 IME 동작, 이 수정의 확인이 아님. 확인하려면 조합을 먼저 확정(한/영 키 등)한 뒤 백스페이스를 눌러야 PTY로 감
  - 수정-2 2차 시도(한/영으로 확정 후 백스페이스): `cat` 출력 줄 `가` -> **사용자 확인 OK**. 입력 줄은 `가  `(나의 2칸이 공백으로 지워짐)

## 반영 범위
| 이름 | 다시 빌드 | 비고 |
|---|---|---|
| 수정-1(새 보드) | supervisor + 웹 | 웹만이면 옛 서버가 `from`을 무시하고 기본 보드에서 타일을 찾아 404 |
| 수정-2(IUTF8) | worker 재실행 | 새로 띄운 process부터. **빌드본 worker 교체 = 그 위 process 전부 종료**(사용자 Claude 세션 포함) -> 세션 끝에 하거나 개발 실행에서 확인 |
