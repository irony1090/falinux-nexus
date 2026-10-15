# history — Windows worker

> → `REF-windows-worker.md` / 현재 진행 → `CURRENT.md`

## 2026-10-15
- 사용자 질문 "worker 에디터 기본 지정이 정상인가, Windows는?" → 코드 점검: Linux는 동작(2026-10-06 vi 확인), `$VISUAL` 미사용·`EDITOR` 인자 분리 안 함 발견(`REF-process-exec-edit.md`) / Windows는 worker 빌드 자체가 실패
- Windows CLI 에디터 조사: 64비트 Windows엔 기본 없음, Microsoft Edit(`edit`) 2025년 공개(기본 포함 여부는 2025년 중반 자료까지만 확인)
- 사용자 "후보로 올려두자" → CURRENT 후보에 추가 → "window 지원 작업 플랜" 요청 → 패키지별 Windows 빌드로 막히는 곳이 `internal/execute/pty` 하나뿐임 확인, `x/sys/windows`에 ConPTY API 있음 확인 → 계획 윈도-0(빌드 분리)~윈도-6(확인) + 결정 WW1(실행 규칙)~WW5(OS 표시) 작성(코드 없음)
- 사용자 "VM으로 할 수 있어?" → KVM·자원 확인, `dockurr/windows` 추천(REF "VM 준비") / 남은 디스크 364GB 확인 → 사용자 "vault 기록까지만" → **보류**(코드·VM 없음)
