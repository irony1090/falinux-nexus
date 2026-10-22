# REF — 외부 접속(모바일 확인) · HTTPS 프록시 (2026-10-13~)

> 이력 → `history/external-access.md` / 테스트 환경 일반 → `REF-test-env.md` / 현재 진행 → `CURRENT.md`

## 경로 (2026-10-13 실측)
| 도메인 | 이 장비 포트 | 쓰는 곳 |
|---|---|---|
| https://test1.falinux.com | 8011 | 프론트(vite 개발 서버) |
| https://test2.falinux.com | 8012 | supervisor(REST + 소켓 `/subscribe`) |
| https://test3.falinux.com | 8013 | **빌드본**(supervisor가 화면 + API, 2026-10-22~ → `REF-build.md`) |

- 앞단 = 다른 장비의 nginx가 TLS를 끝내고 이 장비 LAN IP(192.168.40.132 등)의 포트로 **평문 HTTP** 전달 → 이 장비 서버는 https 처리 불필요, 대신 **0.0.0.0 바인드 필수**(localhost만 열면 502)
- WebSocket 업그레이드 통과 확인(wss 에코). **nginx 유휴 끊김 = 30초**(실측) — 앱 소켓은 ping 10초(`transport.KeepAlive`)로 안 걸림
- worker는 외부를 거치지 않고 같은 장비 supervisor에 직접(`ws://localhost:8012`)

## 설정 (2026-10-13)
| 파일 | 값 | 비고 |
|---|---|---|
| `apps/core/cmd/supervisor/env` | `PORT=8012`, `IP=0.0.0.0` | `main.go`가 `IP`를 바인드 주소로(없으면 localhost). **env는 `go:embed`라 재빌드·재실행해야 반영** |
| `apps/core/cmd/worker/env` | `WS_HOST="localhost:8012"` | supervisor PORT와 맞춤 |
| `apps/frontend/.env`(git 제외) | `VITE_API_URL="https://test2.falinux.com"`, `DEV_PORT=8011`, `DEV_HOST=0.0.0.0`, `HMR_CLIENT_PORT=443` | `DEV_*`·`HMR_*`는 `vite.config.mts`가 `loadEnv`로 읽음(접두사 없음 = 클라이언트로 안 나감) |

- 되돌리기(로컬 전용) = 세 파일의 주석 처리된 옛 값(5050·localhost)으로. vite 세 줄을 지우면 3000·localhost
- 이 설정이면 **PC도 test1로 접속**: localhost:8011 화면 -> test2 API는 다른 사이트라 Lax 쿠키가 안 붙어 로그인 안 됨

## HTTPS 때문에 바꾼 코드 (2026-10-13)
- 소켓 주소: `websocket.hook.ts`에 `ws://localhost:5050` 고정이던 것을 `getBaseURL()`(= `VITE_API_URL`)에서 만듦(https -> wss). https 화면에서 `ws://`는 혼합 콘텐츠로 막히고, 폰에서 localhost는 폰 자신
- vite HMR: TLS 프록시 뒤에선 브라우저가 붙을 포트가 443 → `hmr.clientPort`. 안 하면 `wss://test1...:8011`로 붙으려다 실패
- 쿠키: test1 화면 -> test2 API는 같은 사이트(falinux.com)라 기존 `SameSite=Lax`·`Secure=false` 그대로 동작(https에서도 Secure 없는 쿠키는 저장됨). CORS는 기존 `http*://*` 패턴이 origin을 되돌려 줌
- **30초마다 새로고침(2026-10-13 발견·수정)**: vite HMR 소켓은 무통신이라 nginx가 30초에 끊고, vite 클라이언트는 끊기면 "서버 재시작"으로 보고 다시 붙자마자 `location.reload()`. → `vite.config.mts` `hmrKeepAlive` 플러그인이 프록시 뒤(`HMR_CLIENT_PORT` 있음)에서만 20초마다 빈 사용자 이벤트 `nexus:keepalive`를 보냄. 확인 = 수정 전 30·60·90초 새로고침 / 수정 뒤 100초 무변화
- 주의: `.env`·`vite.config.mts`를 바꾸면 돌고 있는 vite가 **스스로 재시작**해 새 포트로 옮겨 감(2026-10-13 3000 -> 8011)

## 확인 결과 → `history/external-access.md` 2026-10-13
