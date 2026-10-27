# history — 파트(폴더에서 터미널)
> → `REF-dir-shell.md` · → `CURRENT.md`

## 2026-10-27
- 계기: 브라우저 업로드한 파일의 실행 권한이 사라짐(브라우저 `File`에 권한 비트 없음 → supervisor `upload.go` 0644 고정). 고칠 안(가(자동 판정)·나(체크박스)·다(chmod 버튼))을 냈으나 사용자가 "장비 타일에서 그 위치로 셸을 바로 켜기"로 방향 전환, 권한은 셸에서 직접 `chmod`
- 결정: 장비 주인만 · 버튼 = 경로 줄 새로고침 오른쪽 · Windows = powershell · 코드 = Claude 작성
- 구현: 치환 토큰 `{WORKER_SHELL}`(대화에서 말한 `shell: true` 대신 — API·DB 변경 없음). 브랜치 `feat/dir-shell`, 미커밋
- 확인: 격리 헤드리스 Linux OK(→ REF "확인") → 사용자가 격리 3001 + VM `win-dev`로 직접 확인, 끝 → 전부 정리·VM 끔
- 셸 바꾸기 = worker env에 `SHELL=`(REF 결정 표) 추가 → 파트 마감·`feat/dir-shell` 커밋 → master `--no-ff` 머지. BACKLOG "업로드 파일 실행 권한" 삭제
