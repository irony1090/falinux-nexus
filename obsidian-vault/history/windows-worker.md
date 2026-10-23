# history — Windows worker

> → `REF-windows-worker.md` / 현재 진행 → `CURRENT.md`

## 2026-10-15
- 사용자 질문 "worker 에디터 기본 지정이 정상인가, Windows는?" → 코드 점검: Linux는 동작(2026-10-06 vi 확인), `$VISUAL` 미사용·`EDITOR` 인자 분리 안 함 발견(`REF-process-exec-edit.md`) / Windows는 worker 빌드 자체가 실패
- Windows CLI 에디터 조사: 64비트 Windows엔 기본 없음, Microsoft Edit(`edit`) 2025년 공개(기본 포함 여부는 2025년 중반 자료까지만 확인)
- 사용자 "후보로 올려두자" → CURRENT 후보에 추가 → "window 지원 작업 플랜" 요청 → 패키지별 Windows 빌드로 막히는 곳이 `internal/execute/pty` 하나뿐임 확인, `x/sys/windows`에 ConPTY API 있음 확인 → 계획 윈도-0(빌드 분리)~윈도-6(확인) + 결정 WW1(실행 규칙)~WW5(OS 표시) 작성(코드 없음)
- 사용자 "VM으로 할 수 있어?" → KVM·자원 확인, `dockurr/windows` 추천(REF "VM 준비" → 2026-10-23부터 `REF-windows-vm.md`) / 남은 디스크 364GB 확인 → 사용자 "vault 기록까지만" → **보류**(코드·VM 없음)

## 2026-10-19
- 사용자가 다음 파트로 결정 → 결정 WW1(실행 규칙)부터 하나씩 확정 진행
- WW1(실행 규칙) = **가**(노드 이름 확장자 -> 실행기) 확정. DB 확인: 기존 스크립트 노드 이름엔 확장자 없음 -> Linux 영향 없음, Windows에서 확장자 없음·미지원 = 거절 + 안내
- WW2(종료 방식) = 추천안 확정: 요청 = ConPTY 입력 `0x03` / 강제 = `TerminateJobObject`(137) / worker 종료 = Ctrl+C -> 5초 -> 강제 / `KILL_ON_JOB_CLOSE` / Ctrl+C 종료 exit code는 그대로
- WW3(에디터 기본값) = 나 확정: PATH `edit`->`nvim`->`vim`->`nano`, `$VISUAL` 우선, 인자 분리는 공백만. 코드 확인: `resolveEditor`가 `$VISUAL`을 안 읽음(주석과 불일치), 인자 붙은 EDITOR 실패
- WW5(OS 표시) = **나** 확정(추천 가 대신): 등록 OS + 레지스트리 + `worker_instances.os`(`00011`), 빈 값 = linux. -> 결정 WW1~WW5 전부 확정, 다음 윈도-0(빌드 분리)
- 윈도-0(빌드 분리) 구현: `pty`를 공통 `execInteractive.go` + `_linux.go` + `_windows.go`(빈 구현)로 분리, `execFIFO.go` -> `execFIFO_linux.go`, `execute/doc.go` 주석 갱신. Linux 빌드·vet, Windows 교차 빌드(worker·supervisor) 통과. Linux 동작 확인 = `go test -overlay` 임시 테스트(scratchpad `w0/zz_split_test.go`, 재부팅 시 사라짐 — 내용: 한글·크기·입력·exit code / Kill 143·137)
- WW1(실행 규칙) 개정: 사용자 질문(새 확장자도 재배포 없이? / `#!`는? / Windows 기본 매체로 실행만?) 거쳐 = `.ps1` 표 + 나머지 `cmd /c`(파일 연결), `#!` 안 씀. **전제 = 스크립트(터미널) 노드 문법은 그 OS에서 그대로 실행되게 작성**. WW5 검사는 "확장자 있음"만
- 파일 손질 = 사용자 제안(worker env로 인코딩·줄 끝) 반영 확정: 기본값 + `SCRIPT_FORMAT` 확장자별 덮어쓰기(`utf8|utf8bom|cp949` × `keep|lf|crlf`). 윈도-3(실행 규칙) 구현 + overlay 테스트 통과(scratchpad `w3/`, 재부팅 시 사라짐). `x/text` direct
- 윈도-4(에디터) 구현: `$VISUAL` > `EDITOR` > OS 기본값(Windows PATH 탐색) + `prepareEdit` 인자 분리. 정정: env 파일 `EDITOR`가 OS `EDITOR`를 덮음(Overload). overlay 테스트 통과(scratchpad `w4/`)
- 윈도-5(OS 표시) 구현: 프론트 표시 = 사용자 결정 "모든 인스턴스 L/W + 툴팁". `00011_worker_os` + `workerOS` 맵 + exec 확장자 검사. 격리 확인 14/14(`iso5/os_test.py`)
- 윈도-6(확인) 착수: 사용자 승인 → VM `nexus-win`(Windows 11 한국어, 6G/4코어/64G, OEM으로 OpenSSH) 띄움, `worker.exe` 교차 빌드해 shared에
- 사용자 재시작 후 브라우저 확인: OS 표시 위치(툴팁·펼친 목록) 이대로 유지, 장비 칩 표시는 안 넣음. 사용자 supervisor에 `00011` 적용됨(명부 `os=linux`)
- 윈도-6(확인) VM 실행 → Ctrl+C 상속 수정(사용자 "반영해줘") → 17/17(상세 `history/windows-conpty.md`)
- 사용자 "VM 쪽 내가 실행해볼 수 있어?" → 접속 방법(8006 noVNC·3389 RDP, 127.0.0.1 바인드) 안내, 8006은 이미지 내장 뷰어 → 사용자 "8012에 연결해줘" → VM env를 8012 + 새 SQLite 파일로, 등록 `win-dev os=windows` 확인
- 셸 노드 안내(`shell.bat` = `@powershell -NoLogo`) · Docker로 Windows = Linux 컨테이너 안 QEMU/KVM VM(이 장비 KVM 있음, `docker start`로 재사용)
- 사용자 확인: 셸 노드 실행 → supervisor(8012) 재시작 → 다시 붙음·입력 계속(재접속 통과) → **파트(Windows worker) 마감**. 격리 supervisor(5051) 종료·`nexus_win_tmp` 삭제, VM 유지 → `INDEX-windows-worker.md`
