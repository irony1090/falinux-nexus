# REF — 빌드 스크립트 · 빌드 태그 (2026-11-04 `REF-build.md`에서 분리)

> → 이력 `history/build-scripts.md` · 빌드본 분리 설계 `REF-build.md` · updater 설치 `REF-update-p4.md` · 진행 `CURRENT.md`

## 빌드 스크립트 (2026-11-04 대상별로 나눔, 브랜치 `feat/build-scripts`)
| 파일 | 하는 일 |
|---|---|
| `scripts/build.sh <폴더명> [대상...] [--os linux\|windows] [--arch amd64\|arm64]` | 대상 스크립트를 차례로 부름. 대상 생략 = `supervisor worker frontend`(Q1(대상 생략 시)). supervisor·worker를 같이 빌드할 때만 두 태그 뒷부분 비교 경고 |
| `scripts/build/common.sh` | source 전용: 인자 해석(`build_parse`) · 태그 · 같은 태그 경고 · `go_build`(`go -C apps/core`, 교차 빌드면 `CGO_ENABLED=0`) · 실패 정리(`fail_clean`) |
| `scripts/build/supervisor.sh`·`worker.sh` | `build/<폴더명>/<이름>[.exe]` |
| `scripts/build/frontend.sh` | `build/<폴더명>/web/`(`VITE_API_URL=` 같은 출처, 타입 검사 + vite build). `--os`·`--arch` 무시 |
| `scripts/build/updater.sh` | `build/<폴더명>/updater/{linux-arm64, windows-amd64}/` + 설치 스크립트(`scripts/updater/` 것 복사). `--os`·`--arch` 무시하고 둘 다(Q4(updater 출력 모양)) — 옛 `scripts/updater/build.sh`(→ `build/updater/<태그>/`)는 지움 |

| 규칙 | 내용 |
|---|---|
| 이미 있음 | **대상 결과물만** 보고 거절(Q2(이미 있음 규칙)). 폴더가 있어도 다른 대상은 추가 빌드 가능(예: `build.sh 261104 updater`) |
| 실패 | 그 대상 결과물만 지우고, 이번에 비게 된 `build/<폴더명>`도 지움. 다음 대상은 안 돎(`set -e`) |
| OS·아키텍처 | `--os windows` = `.exe`, `--arch arm64` = 보드용. 없으면 이 PC(CGO 기본값 그대로) (Q3(OS·아키텍처)) |
| 폴더명 | 하나만, `/`·`.` 시작 금지 |
| 각 스크립트 단독 실행 | `scripts/build/worker.sh <폴더명> --arch arm64`처럼 됨 |
| 집 서버 한 줄(11-04 실사용) | `scripts/build.sh <폴더명> supervisor worker frontend updater --os windows` — 대상 생략엔 updater가 없어 적어야 함, updater는 `--os` 무시(참고 줄 출력). 옵션은 `--os windows`(대시 둘, `windows`) |
- **함정**: 루트 `.gitignore`의 `build/`가 아무 깊이의 `build` 폴더를 무시해 `scripts/build/`가 git에 안 잡힘 -> `/build/`(루트만)로 고침(11-04)
- 확인(11-04): 인자 오류 4종 · supervisor windows(PE32+) · 같은 대상 거절 · worker arm64(aarch64) · updater 두 종 · 같은 태그 경고(updater 하위 폴더 포함, 폴더는 grep 안 함) · go 없음 실패 -> 파일·빈 폴더 지움 · 대상 생략 전체 빌드 · 같은 폴더에 updater 추가. 시험 폴더 `_bt*`는 지움

## 빌드 태그 (2026-10-23, `REF-build.md`에서 옮김)
- 사용자 관례: embed되는 `cmd/supervisor/env`·`cmd/worker/env` **첫 줄 = `# SUPERVISOR-<작업명>-<YYMMDD>` / `# WORKER-<작업명>-<YYMMDD>`** -> 내뱉은 env로 어느 빌드인지 앎. 바이너리에서 직접 = `grep -aoE '# (SUPERVISOR|WORKER)-[A-Za-z0-9_-]+' <바이너리>`
- **빌드 전에 태그를 바꿀 것**(2026-10-23 `build/worker-dir`를 옛 태그 `IUTF-261022`로 빌드함 — `build/keep`도 같은 태그)
- `build.sh`(2026-10-23): 빌드 전에 태그 출력 + 경고(supervisor·worker 태그 뒷부분이 다름 / 같은 태그로 빌드한 `build/*/supervisor`가 있음), 완료 줄에도 태그. 경고만 하고 멈추지 않음. 바이너리 안에서는 env 첫 줄 앞에 줄바꿈이 없어 `grep -x`가 안 맞음 -> 줄 끝만 고정(`"# 태그$"`)

- 11-04 나눈 뒤: 태그 검사는 대상 스크립트마다 자기 것(`warn_same_tag`), updater 태그 = `cmd/updater/env` 첫 줄
