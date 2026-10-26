# REF — 브라우저 API를 `/api` 아래로 (2026-10-26)

> → 이력 `history/api-prefix.md` · → `CURRENT.md` · 브랜치 `refactor/api-prefix`

## 왜
- 빌드본은 supervisor가 웹(`WEB_DIR`)과 API를 같은 출처로 서빙 -> 프론트 페이지 `/devices`에서 새로고침하면 `GET /devices`(내 장비 목록 API)가 받아 JSON이 보임(집 서버에서 발견, test3도 같음). 격리 시험은 vite 개발 서버라 안 드러남
- 사용자 결정: 웹 서버(nginx) 설정에 기대지 말고 supervisor 코드에서 API 전부 `/api`로 시작

## 결정
| 번호(이름) | 내용 |
|---|---|
| api-1(서버 묶기) | `supervisorRouter.go` `APIPrefix = "/api"`, `api := e.Group(APIPrefix)` + `mount*`(12개) 인자 `*echo.Echo` -> `*echo.Group`. 브라우저 소켓 `/api/subscribe` |
| api-2(웹 서빙 분리) | `main.go` Static `Skipper` = `/api`·`/api/*`·`WORKER_PATH` -> 없는 API는 JSON 404(index.html 아님) |
| api-3(프론트 기준 주소) | `api.util.ts` `getBaseURL()` = supervisor 주소(`VITE_API_URL`, 끝 `/` 제거, 없으면 같은 출처) + `/api`. `VITE_API_URL` 뜻은 그대로(supervisor 주소) |
| api-4(프론트 직접 URL) | **함정: `new URL('/x', base)`는 base의 `/api`를 버림** -> 문자열로 이어 붙임: `websocket.hook.ts` subscribe · `tile.api.ts` keepalive PUT. `transfer.api.ts` `downloadUrl`은 원래 이어 붙이기 |
| api-5(확인) | 아래 |
| 그대로 | worker 소켓 `/worker`(`WORKER_PATH`, 배포된 worker 손 안 댐) · 세션 쿠키 `Path=/` · nginx(test2) |
| 지움 | `GET /test`(인증 없이 process 목록을 서버 로그에 찍던 확인용) + `ProcessManager.Test()` |

- **규칙(불변)**: 새 브라우저 API는 `mount*`에 받은 그룹에만 단다(`e`에 직접 달지 말 것). 프론트에서 API 주소를 만들 땐 `getBaseURL()`/`BaseAxios.defaults.baseURL` 뒤에 문자열로 이어 붙인다
- 배포: supervisor + 웹만 바꾸면 됨(worker 그대로). 옛 웹 + 새 supervisor(또는 반대)는 안 맞음 -> 둘을 같이

## 확인 (api-5, 2026-10-26)
- 격리: 5051 + DB `nexus_api_tmp`(`nexus_member_tmp` 템플릿) + `VITE_API_URL=` 빌드 웹 + worker(장비 등록 키). scratchpad `apitest.mjs` **14/14**: 가입·로그인·장비 등록 -> worker `/worker` 접속 -> `/devices` 새로고침 = 화면 -> `/admin`(관리자 아님) = 화면(보드로) -> 소켓 `/api/subscribe` -> `GET /api/tiles` -> 실행 200 -> 출력이 소켓으로 옴 -> keepalive PUT 200 -> `/api/transfers/<없음>/data` = JSON 404
- curl: `/`·`/devices`·`/admin`·`/login`·`/signup`·`/w/abc`·`/test` = index.html 200 / `/api/devices`·`/api/subscribe` = 401 JSON / `/api`·`/api/nope` = 404 JSON
- 함정: 격리 worker의 `TMPDIR`이 scratchpad면 보관 프로세스 소켓 경로가 너무 김(126 > 85) -> worker가 안 뜸. 짧은 경로(`/tmp/nxa`)로
