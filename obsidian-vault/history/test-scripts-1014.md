# history/test-scripts-1014 — 확인 스크립트 재작성 명세 (2026-10-14 이후)

> 재사용 지식 → `REF-test-scripts-1014.md` / 그 전 → `history/test-scripts.md` / 현재 진행 → `CURRENT.md`

## 2026-10-19
- `REF-test-scripts.md`가 10k자를 넘어(터치 스크롤 `swipe.mjs` 줄 추가) 2026-10-14 절부터 끝까지를 `REF-test-scripts-1014.md`로 분할
- 2026-10-19 절: Windows worker 윈도-5(OS 표시) 격리 `iso5/os_test.py` · 윈도-6(확인) VM `win/vssh`·`iso6/win_test.py` · 터치 스크롤 `touch/run.mjs`·`swipe.mjs`
- 파일 전송 `xfer/`(overlay 확인용 REST + 지연·끊김 프록시 + `test.py` T1~T9) 절 추가
- 보드 URL 바인딩 `url/url.mjs`(격리 프론트 3001 + 5051, 20개) 절 추가
