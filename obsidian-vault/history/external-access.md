# history/external-access — 외부 접속(모바일 확인) · HTTPS 프록시

> 요약·설정 → `REF-external-access.md` / 현재 진행 → `CURRENT.md`

## 2026-10-13 — test1~3.falinux.com(8011~8013)으로 외부 접속 설정 + HTTPS 대응 (커밋 전)
- 사용자 "모바일 확인용, 8011~8013 외부 접근 가능(test1~3), env 수정 + https 처리 필요하면 처리"
- 실측: 포트마다 표시용 임시 서버 → 127.0.0.1 바인드는 502, 0.0.0.0이면 test1=8011·test2=8012·test3=8013 / wss 에코 왕복(3초 쉬고도) 통과 → 프록시 = 다른 장비 nginx, TLS 종료 후 평문 전달
- 변경: supervisor `IP` 바인드(`main.go`) + env 3개 + 소켓 주소를 API 주소에서(https -> wss) + vite 서버 설정을 env로(`DEV_PORT`·`DEV_HOST`·`HMR_CLIENT_PORT`, 설정을 함수형 `defineConfig(({mode}) => ...)`으로)
- 확인(임시 DB `nexus_https_tmp` + 임시 supervisor 8012, 프론트는 사용자 vite(이미 8011로 재시작됨), scratchpad `pw/https.mjs`, 390px 모바일·iPhone UA): API 가입 / test1 로그인 화면 -> 보드 / 쿠키 `sid@test2.falinux.com` Lax / 앱 소켓 `wss://test2.../subscribe` / HMR `wss://test1...`(443) / 다른 세션 폴더 생성이 소켓으로 목록 반영 → 6/7, 나머지 1 = 로그인 전 세션 확인 401 콘솔 로그(정상). 확인 뒤 임시 supervisor 종료·DB 삭제
- 사용자 할 일: supervisor 재실행(8012), worker 재실행(8012로 붙음). vite는 이미 8011

## 2026-10-13(2) — test1 접속 시 30초마다 새로고침 → HMR keepalive로 수정
- 사용자 "test1으로 접속해서 그런가 주기적으로 새로고침" → 헤드리스로 150초 관찰(`pw/reload.mjs`: framenavigated·websocket·console): 30초마다 HMR 소켓 close → "[vite] server connection lost. Polling for restart..." → 재연결 → 새로고침
- 원인 = nginx 유휴 끊김 30초 + vite 클라이언트의 끊김 = 재시작 판단. 수정 = `hmrKeepAlive` 플러그인(20초마다 `server.ws.send('nexus:keepalive')`, 프록시 설정일 때만) → 100초 관찰 무변화

## 2026-10-13(3) — 마감·커밋
- 전체 화면 버튼(→ `history/frontend.md` 2026-10-13(2))과 함께 파트(노드 클립보드)와 나눠 커밋. supervisor·worker `env`(8012)도 지금 쓰는 설정이라 커밋에 포함, 프론트 `.env`는 git 제외
