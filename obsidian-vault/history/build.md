# history — 빌드본 분리

> → `REF-build.md` / → `CURRENT.md`

## 2026-10-22 구상·조율
- 사용자 요청: build 폴더에 supervisor·worker(+sqlite)·웹 빌드, supervisor가 웹 직접 서빙(`WEB_DIR`), env 파일명 옵션
- B1(포트·도메인) 8013 / B2(DB 분리) `nexus_build` / B3(데이터 복사) 복사해서 시작 / B4(반복 배포) `build/{이름}` / B5(웹 경로 변수) `WEB_DIR` 확정
- B6(env·sqlite 출처): sqlite·PG 복사는 사용자가 수동, env는 직전 빌드에서 복사(첫 빌드는 개발 env로 생성) / B7(실행 방법): 사용자가 직접 실행
- 빌드-1(env 옵션)·빌드-2(웹 서빙)·빌드-4(빌드 스크립트) 구현, 빌드-3(같은 출처 API)는 코드가 이미 됨. 임시 DB + 5051로 확인 끝. 8013 사용 중(다른 프로젝트 vite) 발견. 확인용 `build/_check`은 삭제 권한 거부로 남음 -> 사용자가 지울 것
- 사용자가 env·sqlite를 `build/` 공통으로 옮김. Claude가 `nexus_build` 생성·덤프(건수 일치). 스크립트·SQLITE 경로 수정 제안
- 사용자가 `worker.env` SQLITE를 `../`로 고치고 `build/_check`으로 빌드본(8013) 재실행 확인 -> 8011·8012 끄고 세션을 빌드본 위에서 다시 띄우기로. 남은 것 = 스크립트 env 단계 제거·커밋
- 남아 있던 확인용 worker(이전 세션 scratchpad `wchk.env`, 5051로 연결)를 사용자 요청으로 종료. 빌드본 supervisor·worker(`build/_check`, 8013)만 남음
- env 내뱉기 조율(tag 로더 안 -> embed 내뱉기로 사용자 결정, 개발 `-env dev.env` 필수, `cmd/*/env` git 제외) 후 Claude가 작성: main.go 2개 `initEnv`·env/dev.env·gitignore·README·build.sh 2단계·worker SQLITE 검사 버그. scratchpad에서 빌드본·개발 경우 6가지 확인(`TMPDIR` 함정 발견)

## 2026-10-22(2) 마감
- 사용자가 `build/const`로 다시 빌드·재실행 -> Claude가 8013 응답·SPA 경로·worker 연결·PostgreSQL 연결 확인
- 사용자가 개발 env 파일 이름을 `dev.env`로 바꿈 -> Claude가 main.go 2개·README 2곳·`apps/core/.gitignore` 주석·vault를 `dev.env`로 맞춤, `go build ./...` OK
- 파트(빌드본 분리) 마감: INDEX-build로 접고 코드 + vault 한 커밋
