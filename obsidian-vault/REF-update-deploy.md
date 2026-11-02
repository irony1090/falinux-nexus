# REF — 업데이트 app 설계 (현재 설계)

> 2026-10-30 신설 → **2026-11-02 전면 개정**(방향 전환: updater = worker와 독립된 범용 앱). 결정 경위·시안 기록 → `REF-update-redesign.md` / 규칙(잠금·삭제·되돌리기·단계 개별 실행·버전) → `REF-update-rules.md` / 점검 원문 → `REF-review-1030.md` / 이력 → `history/update-deploy.md` / 지금 배포 방식 → `REF-home-deploy.md`(update.ps1)·`REF-build.md` "보드 worker 교체" / 현재 진행 → `CURRENT.md`
> 시안 = Claude 아티팩트 https://claude.ai/artifact/XGMJqjg7rHdAHUhRMXHbZa (파일 `user/updater-화면-claude.html`) · DeepSeek판 `user/updater-화면-deepseek-v2.html`. 옛 `user/배포-화면*.html`은 옛 구조(참고만)

**상태(2026-11-02): 설계 합의 끝(설계-8·설계-10 제외), 시안 확정. P1(보관·화면) 코드 끝(구현 → `REF-update-p1.md`), 사용자 확인 대기.**

## 무엇인가
**updater = 각 PC에 상주하는 범용 원격 작업 앱.** worker와 독립. supervisor에 접속해 있고, 관리자가 웹에서 올린 파일을 PC에 깔고 `steps.json` 순서대로 단계를 실행한다. 미리 만든 파일을 올려 쓰며 컴파일은 하지 않는다. worker/supervisor 교체는 번들 내용의 한 사례일 뿐. **관리자만 접근**(장비 주인에게도 공개 안 함). 장비 루트(키)가 있어야 접속된다(P2 정-1).

| 지금 막히는 곳 | 내용 |
|---|---|
| supervisor 자가 교체 | 실행 중 exe는 덮어쓸 수 없고 자기를 재시작할 수도 없음 → 지금은 `Start-ScheduledTask`로 우회 |
| worker는 Job | 터미널에서 띄운 프로세스는 터미널이 끝나면 같이 죽음(JS-1) |
| 장비가 늘어남 | 집 서버 + 보드 2대 + 회사 PC 여러 대를 한 번에 관리할 수단이 없음 |

## 확정 요약
| 항목 | 결정 |
|---|---|
| 키 | `메인키#서브키`(인스턴스). 메인키 = updater 설정, 서브키 = supervisor 발급 + updater `state.json` 저장. worker와 독립(별도 `updaters` 명부) |
| 접속·인증 | 메인키별 키(`nxu_`, 다시 보기 가능) — 관리자가 장비 루트를 만들면 발급(2026-11-02 정-1(접속 인증), `REF-update-p2.md`) |
| 같은 메인키 | 같은 파일·`steps.json`을 받음. 성공 판정 = 그 메인키 인스턴스 중 한 대라도 |
| 순서 | Linux 먼저(보드 시험) → Windows |
| 별칭 | 서브키(인스턴스)에 별칭·메모를 적을 수 있음(supervisor DB `updaters` 행) |

## 릴리스 보관 구조
장비(메인키)가 루트, 안에 빌드(버전). **OS-CPU 분류는 폐기**(OS/CPU는 화면 표시용 칩).
```
common/<이름>/<버전>/        공용 묶음
<메인키>/<버전>/             빌드: 메인키 파일 + steps.json + 스크립트
<메인키>/<버전>/@<서브키>/   서브키 파일(한 인스턴스에만 줄 파일, P1에서는 자리만)
```
- 합치는 순서 `공용 -> 메인키 -> 서브키`(뒤가 앞을 덮어씀). **지우지는 않음**
- 공용은 빌드의 `steps.json` `uses`(예 `["nexus-worker@261103_a"]`)로 명시 의존. 빌드의 파일 타일에서는 읽기 전용
- 장비 루트는 관리자가 화면에서 만들고 그때 접속 키가 발급된다(첫 접속 자동 생성은 2026-11-02 정-1(접속 인증)에서 폐기)
- 내용 = supervisor 서버 파일시스템 `UPDATE_STORE_DIR`, 메타 = DB. 스크립트·`steps.json`은 빌드 폴더 안의 평범한 파일. 편집 이력 없음(덮어쓰기)
- 빌드 파일 한 번에 받기(zip): 기본 = 범위별로 그대로(`common/<이름>@<버전>/` · `main/` · `sub/<서브키>/`), 선택 = 합친 결과. 서버가 만들면서 바로 내려보냄. "한 번에 올리기"는 없음(파일 단위 업로드)

## steps.json
```json
{ "uses": ["nexus-worker@261103_a", "linux-tools@v2"],
  "deploy":   [ { "id": "stop",   "script": "00-stop.sh" },
                { "id": "backup", "builtin": "backup" },
                { "id": "copy",   "builtin": "copy" },
                { "id": "start",  "script": "03-start.sh" },
                { "id": "verify", "script": "04-verify.sh", "timeout": "15m" },
                { "id": "prune",  "script": "90-prune.sh", "manual": true } ],
  "rollback": [ { "id": "stop",    "script": "00-stop.sh" },
                { "id": "restore", "builtin": "restore" },
                { "id": "start",   "script": "03-start.sh" } ] }
```
- 단계 종류 = `script`(updater가 프로세스 실행, 종료 코드로 판정) / `builtin`(updater 프로그램 안 함수, 셸 안 거침 → OS 무관)
- `manual: true` = 배포 흐름에 안 끼고 `실행` 버튼으로만. 지우기(`prune`)도 이런 단계
- 단계에 넘기는 환경변수 = `TARGET_DIR`·`VERSION`·`STEP_ID`, 복원을 스크립트로 쓰면 `BACKUP_DIR`·`NEW_FILES`
- 복원도 단계 — 내장 `restore`를 쓰거나 `script`로 대체. 되돌리기 버튼 = `rollback` 목록 실행, 자동 실행은 기본 꺼짐(`"autoRollback": true`로 켬)

| builtin | 하는 일 (배포 시작 때 빌드 파일이 `staging/<배포 id>/`에 받아져 있음) |
|---|---|
| `backup` | 덮어쓸 파일 중 `TARGET_DIR`에 이미 있는 것 → `backups/<배포 id>/files/`, 없는 것 → `new-files.txt`, 배포 id·버전·시각 → `manifest.json`. 위치 = `TARGET_DIR` 밖 updater 전용 폴더, 최근 3회 보존 |
| `copy` | `uses` 먼저, 그다음 빌드 파일을 `TARGET_DIR`에 덮어씀. 임시 이름에 쓴 뒤 이름 변경, 잠긴 파일은 기존 것을 `*.old-<시각>`으로 바꾼 뒤 씀. 지우지 않음 |
| `restore` | 백업 `files/`를 되돌리고 `new-files.txt`의 파일만 삭제 |
- `backup` 없이 `copy`만 개별 실행하면 복원 불가 → 단계 로그에 경고(순서는 막지 않음)
- 실패·timeout: 스크립트 프로세스를 자식까지 종료(Linux 프로세스 그룹, Windows 프로세스 트리) → 배포 실패 표시. 스크립트가 `trap`으로 자기를 되돌리는 규약은 폐기

## DB (supervisor PostgreSQL, 컬럼 상세는 P1 착수 때 마이그레이션 번호와 함께)
| 테이블 | 담는 것 |
|---|---|
| `updaters` | 메인키·서브키(PK)·별칭·메모·OS/CPU·설치 경로·설치본 이름(updater 보고값)·마지막 접속. online은 메모리 파생(관측용 명부) |
| `update_bundles` | 빌드·공용 묶음 한 테이블: 종류·소유(메인키 또는 묶음 이름)·버전·폴더 sha·올린 사람·시각. (종류, 소유, 버전) 유일 |
| `update_files` | 파일 메타: 범위(메인키/서브키)·서브키·경로·크기·파일 sha |
| `update_uses` | 빌드가 쓰는 공용 묶음. `steps.json` 저장 때 다시 만듦, 공용 삭제 불가 판정에 씀 |
| `update_locks` | 잠금 기록(누가·언제·잠글 때 sha). 잠그기 때 빌드와 쓰는 공용 묶음마다 한 행, 해제 없음 |
| `update_deployments` | 배포 이력(인스턴스·시작/끝·결과·번들 이름·sha 사본). 번들이 삭제돼도 이력이 남게 연결을 끊고 사본을 둠 |
| `update_step_runs` | 단계 실행 기록(배포 중 또는 개별 실행·목록 deploy/rollback·단계 id·종류·결과) |
- 잠그기 가능 판정 = **DB 함수 한 곳**. 단계 로그 = DB에 요약(마지막 몇 줄), 전체는 파일

## 화면 (관리자 전용 타일 페이지, `MemberPage` 타일 틀·찰떡 규칙)
```
updater page (admin only)
+-- list: tabs [device | common packs]       장비 > 인스턴스, 공용 묶음
+-- device tile (mainKey)                     인스턴스·빌드 목록(빌드마다 열기 / Run·Log)
     +-- build tile                           상세 + 열기 줄 + 삭제·전체 받기·배포
          +-- Files / steps.json              형제 타일 (Script = Files 목록에서 텍스트 파일을 눌러 엶)
          +-- Run · Log                       단계 실행·로그·되돌리기·잠그기
```
| 순서 | 조작 | 열리는 것 |
|---|---|---|
| 1 | 목록에서 장비 선택 | 장비 타일 |
| 2 | 장비 타일의 빌드 `열기` | 빌드 타일(연 곳 바로 뒤) |
| 3 | 빌드 타일의 `Files` · `steps.json` 열기 → Files 목록에서 텍스트 파일 누름 | 편집 타일 / Script 타일 |
| 4 | 빌드 타일의 `배포` → 대상 인스턴스 고르기(같은 메인키 기본 전체, 오프라인은 표시만) | Run · Log 타일 |
| 5 | Run · Log의 단계 `실행` · `되돌리기` · `잠그기` | 같은 타일 |
- 장비 타일의 빌드 행마다 `Run · Log`를 바로 열 수 있음. 타일 규칙(열기 순서·`tileKinds`/`TILE_KINDS` 두 표·키의 `#` 처리·`Ct*` 입력칸)은 기존 것을 따름
- 시안 찰떡 규칙: → `REF-design-chaltteok.md` "시안에서 정리한 규칙(11-02)"

## 단계 계획
| 단계 | 내용 | 확인 |
|---|---|---|
| P1(보관·화면) | 마이그레이션·`UPDATE_STORE_DIR` 보관·타일 페이지(장비 루트 수동 생성·빌드·공용·Files·steps.json·Script) · **파일·`steps.json`·스크립트 원문 텍스트 편집 + 검증(줄 단위 오류)** · 전체 받기 · 삭제. `steps.json` 시각 편집 UI는 그다음 | 올린 폴더가 그대로 보관·받아지는가 |
| P2(updater Linux) | systemd 상주, 접속·서브키 발급·`state.json`, staging 수신, `script`·`builtin` 실행, 로그 중계, 설치본 보고 | 보드 한 대에서 단계가 돌고 화면에 인스턴스가 뜨는가 |
| P3(배포) | 배포 팝업·Run·Log·개별 실행·잠그기(DB 함수)·되돌리기, 설계-8(supervisor 되돌리기)·설계-10(자기 교체) 결정 | 일부러 실패시켜 되돌림 확인 |
| P4(Windows) | 예약 작업, PowerShell 단계, supervisor 교체 | 집 서버에서 supervisor·worker 교체 |
- 잠그기는 성공한 배포 기록이 있어야 하므로 실제 쓰는 시점은 P3(P1에서는 테이블만)

## P1 착수 계획 (2026-11-02 확정)
브랜치 `feat/updater-p1`, 커밋은 파트 마감 때 한 번. 코드 = Claude 작성.
| 단위 | 내용 |
|---|---|
| 단위-1(DB) | `00016_updates.sql` 7개 테이블 + sqlc `query/updates.sql`. 잠그기 판정 DB 함수는 P3 |
| 단위-2(보관소) | env `UPDATE_STORE_DIR`(기본 `./update-store`, `VIEW_CACHE_DIR`처럼 DB 표식 파일로 다른 DB 것이면 끔). 임시 이름에 쓴 뒤 이름 변경, 파일 sha256, 폴더 sha = 정렬된 경로+sha 목록. 패키지 `internal/supervisor/updstore` |
| 단위-3(API) | `/api/admin/updates/...`(`mountUpdates`, 관리자만): 장비 루트 생성·삭제 · 빌드·공용 CRUD + **빌드 복사로 새 버전** · 파일 목록·조각 업로드(512KB)·텍스트 저장·삭제 · `steps.json` 저장 때 검증(줄 단위 오류) + `update_uses` 재생성 · zip 받기(기본 범위별, `?merged` 합친 결과). 잠긴 빌드 쓰기 거부 |
| 단위-4(페이지 틀) | `/updater`(`meta.admin`) + 헤더 메뉴. **`MemberPage` 틀을 공용으로 떼어 냄** — 컨트롤 패널·끌기·크기·머리 줄·도트·URL 저장을 두 페이지가 같이 씀. 페이지 전용(경로 제목·칩 이름·키 접두어·목록 탭)은 페이지별 종류 표로 |
| 단위-5(목록·장비·빌드 타일) | 목록 탭 [장비 \| 공용 묶음], 인스턴스 칸은 P1에서 빈 상태. 빌드 타일 열기 줄 = `[Files \| steps.json \| Run · Log]`(**`Script` 버튼 뺌**, 안 A) + 삭제·전체 받기 |
| 단위-6(편집 타일) | Files = 목록·업로드·삭제·새 텍스트 파일, 텍스트 파일 누르면 Script 타일, `steps.json` 누르면 steps.json 타일, 바이너리는 타일 없음. 공용은 읽기 전용. `.sh` LF 저장 |
| 단위-7(CtCode) | 편집기 = CodeMirror 6 공용 컴포넌트 `CtCode`(`Ct*` 규칙, 줄 번호·줄 오류 표시). **메인 보드 `ScriptEditDialog`의 `CtTextarea`도 같이 교체** |
| 단위-8(확인) | 임시 DB + 5051 헤드리스. 회원 페이지 회귀(틀 분리) + 스크립트 편집 다이얼로그 포함 |
- P1에 안 넣음: Run·Log 동작 · 배포 · 인스턴스 별칭 화면(P2) · 서브키 파일(자리만) · `steps.json` UI 편집(P1 다음, Run·Log 단계 행 컴포넌트를 다시 씀)

## 미정 (착수 때 확정)
- **P2 전**: 2026-11-02 밤 6개 확정 → `REF-update-p2.md` "P2 전 확정"
- **P3 전**: 설계-8(supervisor 되돌리기, DB 스키마가 바뀐 릴리스) · 설계-10(자기 교체: 로그 경로 끊김·updater 자신의 교체)
- 열린 이름: 서브키 파일 탭 구현은 P1 이후

## 함정 (지난 배포에서 실제로 물린 것)
- 실행 중 exe → `*.old-<시각>`으로 물러나게 한 뒤 복사(내장 `copy`가 처리)
- `-hold`가 옛 exe를 쥠 → 남은 보관 프로세스를 세어 결과에 표시. `*.old-*`는 안 지우므로 충돌 없음
- Job(JS-1) → updater는 서비스/스케줄 작업으로 떠 있어야 함
- 자동 되돌림 vs 입력 대기 → `autoRollback` 기본 꺼짐, 대화형 단계는 시계를 멈춤(콘솔 결정과 함께)
- 서버·DB 버전 짝 → 내장 `backup`은 파일만 백업. DB는 별도(`pg_dump` 단계) — 설계-8(supervisor 되돌리기)에서 정함

## 폐기된 것 (옛 설계, 읽다 헷갈리지 않게)
| 옛 | 폐기 이유 |
|---|---|
| 릴리스 `os-arch` 폴더·자동 매칭 | 장비(메인키)가 루트. OS/CPU는 표시용 |
| 내 장비 "배포 정보 타일"(주인도 봄) | 관리자 전용 |
| 실행 중 빌드와 설치본 비교(`RegisterRequest.Build`·`//go:embed env`) | updater가 설치본 이름만 관리 |
| 단계 스크립트 `trap` 자기 되돌림, `02-copy`의 `--delete`·`/MIR` | 내장 `backup`·`copy`·`restore`로 대체, 지우지 않음 |
| arch 단위 잠금·`배포 완료` | 빌드 단위·`잠그기`, 판정은 sha |
| `update.sh`/`update.ps1` 한 번 실행 | `steps.json`이 순서를 가짐 |
