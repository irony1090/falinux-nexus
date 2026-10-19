# history — Windows worker: ConPTY 구현

> → `REF-windows-conpty.md` / 파트 계획 이력 → `history/windows-worker.md` / 현재 진행 → `CURRENT.md`

## 2026-10-19
- `REF-windows-worker.md`가 10k자를 넘어 윈도-0 결과·ConPTY 설계를 `REF-windows-conpty.md`로 분할
- 윈도-1(ConPTY 실행)·윈도-2(종료) 설계 작성(코드 전): `x/sys/windows` API 존재 확인, 시작 순서 9단계(멈춘 채 job에 넣고 재개 / ConPTY에 넘긴 파이프 끝 닫기) · 끝날 때 6단계(닫힘 표시 -> `ClosePseudoConsole` -> 출력 EOF 대기 -> 핸들 닫기 -> Completed) → 사용자 확인 대기
- 윈도-1(ConPTY 실행)·윈도-2(종료) 설계대로 구현(`execInteractive_windows.go`) + `x/sys` direct 의존. 설계 추가 = `STARTF_USESTDHANDLES`(윈도-6 확인 1순위). Windows 교차 빌드·vet 통과, 실행은 미확인(wine 없음)
- 윈도-6(확인) VM 실행: 15/18 통과 + 실패 원인 = Ctrl+C 무시 속성 상속(overlay 실험으로 확인) / `.py`는 Python 없음 / worker 강제 종료 시 스크립트 정리 확인
- 사용자 "반영해줘" → `allowCtrlCInChildren()`(`SetConsoleCtrlHandler(NULL, FALSE)`, `sync.Once`, 재발 방지 주석) 추가 → 저장소 빌드 worker.exe로 VM 재실행 17/17 통과(`.py` 제외). overlay 실험 파일·`worker_ctrlc.exe` 삭제
