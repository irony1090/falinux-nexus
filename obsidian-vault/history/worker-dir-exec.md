# history — worker 폴더 타일: 폴더-4(실행)

> → `REF-worker-dir-exec.md` / 앞 이력 → `history/worker-dir-list.md` / → `CURRENT.md`

- 2026-10-22(7) 폴더-4(실행) 세부 설계 제시 + Q23(누르는 방식)·Q24(실행 대상) — 확인 대기
- 2026-10-23 Q23 실행 버튼·Q24 실행 가능한 파일만 확정. 사용자 요청: 폴더-4 전에 Windows VM에서 폴더-1·2 확인
- 2026-10-23 Windows VM에서 폴더-1·2 확인(13/13·9/9, junction 2건 고침 → `history/worker-dir-list.md`) 뒤 사용자 "진행" -> 폴더-4 구현. 마이그레이션 `00012` · worker `prepareDirect`(형식 변환 없음) · `POST /workers/:m/:s/exec` · 터미널 nodeId 선택값 · 행 "실행" 버튼. 헤드리스에서 "막 연 장비 타일에서 실행하면 루트에 붙음"을 찾아 `flush()`로 고침, 긴 경로가 이름을 가려 `shortPath`. 격리 REST 31/31·헤드리스 13/13·Windows 11/11, 미커밋
- 2026-10-23 사용자 요청으로 스크립트 실행(`nodeOps.start`)에도 `flush()` — 새로 연 폴더 타일에서 바로 실행하면 루트에 붙던 버그, 고친 뒤 8/8·고치기 전 0/8
