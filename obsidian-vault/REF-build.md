# REF — 빌드본 분리 (개발 실행과 사용 실행 나누기)

> → `history/build.md` / → `CURRENT.md` / 외부 접속 포트 → `REF-external-access.md`

## 목적
사용자가 Claude 세션 등을 Nexus(개발 실행) 위에서 쓰고 있어, 개발 중 supervisor·worker 재시작이 사용 중인 세션을 끊는다 -> 사용용 빌드본을 따로 띄운다. 파트(worker 폴더 타일)보다 먼저 함 (2026-10-22)

## 확정 (2026-10-22)
| 번호(이름) | 결정 |
|---|---|
| B1(포트·도메인) | 빌드본 = **8013(test3.falinux.com)**, 화면 + API 한 포트. 개발은 8011·8012 그대로 |
| B2(DB 분리) | 빌드본 = **별도 PG DB `nexus_build`** (같은 DB면 서로 상대 process를 PENDING으로 만들고, 개발 마이그레이션이 바로 적용됨) |
| B3(데이터 복사) | `nexus_build`는 지금 `nexus` 내용을 **복사해서 시작**. worker sqlite도 복사(신원 연속) |
| B4(반복 배포) | 빌드 스크립트, 출력 = **`build/{빌드명 또는 버전명}`**(실행할 때 이름 지정) |
| B5(웹 경로 변수) | supervisor env에 **`WEB_DIR`** 추가 -> 빌드된 웹 파일을 supervisor가 직접 서빙 |
| env 파일명 | 실행 파일에 `-env <파일명>` 옵션. 같은 폴더에 `supervisor.env`·`worker.env` |

## 메모
- 지금 env 로딩: `go:embed env`(기본값) + 비개발 실행이면 실행 파일 옆 `env`로 덮어씀(`cmd/*/main.go` `initEnv`)
- 프론트 `.env`의 `VITE_API_URL=test2`가 빌드에도 들어감 -> 빌드는 같은 출처(빈 값이면 `location.origin`)
- SPA 경로(`/`, `/w/:hash`, `/login`)와 API 경로(`nodes processes shares subscribe tabs test tiles users workers workspaces`, worker 소켓 `/worker`)는 안 겹침 -> 없는 경로 GET은 `index.html`로

## 구현 (2026-10-22, 마감 커밋)
| 단계 | 결과 | 파일 |
|---|---|---|
| 빌드-1(env 옵션) | `-env <파일>`(상대 = 실행 파일 폴더). 없으면 옛 동작(옆 `env`). **지정한 파일이 없으면 panic**(-> E3(없으면 내뱉고 종료)로 대체) — `LoadEnv`는 두 번째 로드부터 파일 없음을 넘기므로, 안 막으면 embed된 개발 env(8012·`nexus`)로 뜸. 개발 실행(`go run`)은 지금처럼 embed만 | `cmd/supervisor/main.go`, `cmd/worker/main.go` |
| 빌드-2(웹 서빙) | `WEB_DIR` 있으면 `e.Pre(StaticWithConfig{HTML5})` — 라우팅·tx 앞. 파일 없으면 API로, 404면 `index.html` | `cmd/supervisor/constants/env.go`, `cmd/supervisor/main.go` |
| 빌드-3(같은 출처 API) | 코드 수정 없음 — `getBaseURL()`이 이미 빈 값이면 `location` 사용. 스크립트가 `VITE_API_URL=`로 빌드 | `scripts/build.sh` |
| 빌드-4(빌드 스크립트) | `scripts/build.sh <이름>` -> `build/<이름>/{supervisor, worker, web/, supervisor.env, worker.env}`. 이미 있으면 거부, 실패하면 폴더 지움. env = 직전 빌드(가장 최근 폴더)에서 복사 / 첫 빌드는 개발 env에서 `PORT=8013`·`DB_NAME=nexus_build`·`WEB_DIR="./web"`·`WS_HOST="localhost:8013"`만 바꿔 생성. sqlite는 안 건드림 | `scripts/build.sh`, `.gitignore`(`build/`) |

- 함정: `npm run build`(run-p `{@}`)는 `--outDir`를 위치 인자로 넘겨 vite가 root로 읽음 -> 스크립트는 `npm run type-check` + `npx vite build --outDir` 따로
- 확인(2026-10-22): 임시 DB `nexus_chk_build` + 5051로 빌드본 실행 -> `/`·`/w/abc123`·`/login` = index.html, `/favicon.svg` 정적, `/nodes/0/runs` = API 401, `-env nope.env` = panic, worker `-env`로 접속·등록 성공, 번들에 test2 없음. 임시 DB·프로세스 정리함

## 사용자가 직접 할 것
- B1(포트·도메인) 충돌(2026-10-22): 8013을 다른 프로젝트(`irony/community` vite)가 쓰고 있었음 -> 사용자가 비움, 8013 그대로
- ~~DB(빌드-5, 1회)~~ -> 2026-10-22 Claude가 실행: `nexus_build` 생성 + `nexus` 덤프 복사(7 users·79 nodes·180 processes·10 boards·goose v11 일치). 명령: `docker exec postgres15 psql -U irony -d postgres -c "CREATE DATABASE nexus_build OWNER irony;"` -> `docker exec postgres15 sh -c "pg_dump -U irony nexus | psql -U irony -d nexus_build"`
- sqlite: `sqlite3` 명령 없음. WAL 모드라 개발 worker를 잠깐 끄고 `apps/core/cmd/worker/worker-sqlite.db`를 복사(켠 채면 `-wal`·`-shm`도 같이)
- 실행: `cd build/<이름> && ./supervisor -env supervisor.env` / `./worker -env worker.env`

## 공통 env·sqlite (2026-10-22 사용자 결정)
- 사용자가 `supervisor.env`·`worker.env`·`worker-sqlite.db`(+`-wal`·`-shm`)를 **`build/` 바로 아래**로 옮김 = 모든 빌드가 공통으로 씀
- `worker.env` `SQLITE="../worker-sqlite.db"` — 사용자가 고침(상대 = 실행 파일 폴더, `./`면 빌드 폴더에 빈 DB가 생겨 신원이 바뀜)
- 실행 = `cd build/<이름> && ./supervisor -env ../supervisor.env` / `./worker -env ../worker.env`
- **2026-10-22 사용자 재실행 확인**: 지금 도는 빌드본 = **`build/_check`**(확인용으로 만든 폴더를 그대로 씀, 8013). 사용자가 8011·8012를 끄고 Claude 세션을 빌드본 위에서 다시 띄움
- ~~남은 것: `scripts/build.sh` 3단계(env 복사·생성) 제거~~ -> E5(스크립트)로 해소

## env 내뱉기 계획 (2026-10-22 구현, 마감 커밋)
embed된 `cmd/*/env` = **배포 값**으로 단정. 빌드 전에 이 파일을 확인하는 것이 절차

| 번호(이름) | 결정 |
|---|---|
| E1(embed = 배포 값) | `cmd/*/env`(embed) = 배포 값(8013·`nexus_build`·`WEB_DIR`). tag·reflection 로더는 안 씀(사용자 결정, `LoadEnv` 그대로) |
| E2(개발 = dev.env) | 개발 값은 `cmd/*/dev.env`, `go run ./cmd/supervisor -env dev.env`(상대 = 소스 폴더). **개발 실행에서 `-env` 필수, 없으면 panic** — 빼먹으면 embed(배포 값)로 떠서 `nexus_build`에 개발 마이그레이션이 적용되고 worker가 8013에 같은 메인키로 붙음 |
| E3(없으면 내뱉고 종료) | 빌드본에서 env 경로(`-env` 또는 옆 `env`)에 파일이 없으면 embed 내용을 그 경로에 쓰고 `exit 1`. 오늘 넣은 "`-env` 파일 없음 panic"을 대체 |
| E4(git) | `cmd/*/env` = gitignore(배포 비밀번호), `cmd/*/dev.env` = git에 둠 |
| E5(스크립트) | `build.sh` 3단계(env 복사·생성) 삭제 -> "남은 것"의 스크립트 정리가 이걸로 끝남 |

- 구현: `cmd/*/main.go` `initEnv`(개발·빌드본 모두 **파일 하나만** 읽음 — embed는 내뱉기 전용, 옛 "embed 위에 파일 덮기" 폐기), `cmd/*/env` = 지금 `build/*.env` 복사, `dev.env` = 옛 `env`, `apps/core/.gitignore` `/cmd/*/env` + `git rm --cached`, README 2곳 `go run ... -env dev.env` + clone 안내(가(README 안내)), `build.sh` 2단계로(sqlite 안내 줄도 공통 `build/` 배치와 안 맞아 삭제). 별건: worker `env.go` SQLITE 검사가 `wsScheme`을 보던 복붙 버그 수정
- 확인(scratchpad, 없는 DB 이름): 빌드본 `-env` 없음 -> 파일 생성(600, embed 일치)·exit 1 / `-env` 생략 -> 옆 `env` 생성 / 생성 파일로 로드 OK / 필수 키 빠짐 -> panic / 개발 `-env` 없음 -> panic / 개발 dev.env 로드 OK
- **함정**: `IsDev()` = 실행 파일 경로에 `os.TempDir()`(`/tmp`)가 있으면 개발 -> `/tmp` 아래 둔 빌드본은 개발로 판정돼 `-env` 필수·상대 경로 = 소스 폴더. 확인 때는 `TMPDIR=/var/empty`로 우회함

## 마감 (2026-10-22)
- 사용자가 새 코드로 **`build/const`** 빌드·재실행: `cd build/const && ./supervisor -env ../supervisor.env` / `./worker -env ../worker.env` -> 8013 `/`·`/w/x` 200, worker 소켓 연결, PostgreSQL 연결 확인
- 개발 env 파일 이름: 사용자가 `env.dev` -> **`dev.env`**로 바꿈. 코드 주석·panic 메시지·README 2곳·`.gitignore` 주석도 맞춤. 개발 실행 = `go run ./cmd/supervisor -env dev.env`
- 남긴 것: 옛 확인용 `build/_check` 폴더 — 지울지 사용자 판단
