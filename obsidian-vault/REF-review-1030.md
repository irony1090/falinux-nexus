# REF — 10-30 작업 점검 결과 (2026-11-02)

> 이력 → `history/review-1030.md` / 대상 설계 → `REF-update-deploy.md`·`REF-update-rules.md` / 대상 코드 → `REF-member-drag.md`·`REF-external-access.md` "검색 엔진 색인 금지" / 현재 진행 → `CURRENT.md`

**범위**: 금요일(10-30) 커밋 `77a9426`(내 장비·관리자 타일 이동) · `7ac7492`(검색 색인 금지) · vault·시안(파트(업데이트 app) 설계). 10-29 이전 커밋은 안 봄.
**상태(2026-11-02)**: 코드-1(빌드 막힘)~코드-5(왼쪽 끝 띠 자리)·코드-7(robots.txt 상충) 수정 끝, `fix/member-drag-review` master 머지, `scripts/check/run.sh` 통과, 사용자 실기기 확인 끝. 코드-6(색인 금지 미배포)은 새 빌드·배포 대기(사용자). 설계 항목은 코드 없는 기획 단계이고 다음 세션에 결정.

## 코드 — 번호-이름

| 번호(이름) | 심각도 | 곳 | 내용 | 근거 |
|---|---|---|---|---|
| 코드-1(빌드 막힘) | 높음 | `MemberTileView.vue:46`, `TileWorkspace.vue:41` | `npm run type-check` 실패(rc=2). `scripts/build.sh`는 `set -e`로 같은 명령을 부르므로 **master에서 빌드본을 못 만든다** | 실행 확인. build.sh 자체는 안 돌림 |
| 코드-2(보드 끝 띠 사라짐) | 높음 | `TileWorkspace.vue:41` `:edge="grids.edge"` | 템플릿은 객체 안의 ref를 풀지 않아 ShallowRef가 그대로 prop으로 간다. `edge?.prev`가 항상 undefined라 **보드 끝 띠가 안 그려진다**. 페이지 넘김 판정(`bands.track`)은 그대로 돈다. 고침 = `const { edge } = grids` 또는 `grids.edge.value` | 타입 오류 + 컴포넌트 재현(값 = BAND / ref = none). 브라우저에서는 안 돌림 |
| 코드-3(대신 등록 이름) | 낮음 | `MemberTileView.vue:46` `admin.value` | `admin`은 boolean이라 `.value`가 undefined. 관리자가 등록 타일을 끌 때 칩이 "대신 등록" 대신 "장비 등록"으로 나온다 | 타입 오류, 코드 읽기 |
| 코드-4(넓은 타일 위 왕복) | 중간 | `memberDrop.util.ts` `memberDropOrder` | 가로 2칸 타일이 있을 때 같은 좌표에서 순서가 번갈아 바뀐다. 예: 2×2 Grid, `[2x1, 1x1, ...]`, 커서가 넓은 타일 위 절반의 오른쪽 칸 -> 앞으로 옮김 -> 넓은 타일이 아래 줄로 내려감 -> 커서 자리가 빈 칸 -> 뒤로 옮김 -> 반복. REF-member-drag의 "흔들림이 없음"과 다름 | 함수 전수 시뮬레이션 1,146,880회 중 2,334회 왕복. 세로 크기만 다른 경우(가로 1칸)는 0회. 브라우저에서는 안 돌림 |
| 코드-5(왼쪽 끝 띠 자리) | 낮음 | `MemberPage.vue` `.body` | PC에서 띠 기준은 목록까지 포함한 `.body`, 판정 기준은 strip. 왼쪽 띠가 목록 위에 그려지고 실제 판정은 목록 전체 + strip 왼쪽 48px | 코드 읽기만 |
| 코드-6(색인 금지 미배포) | 중간 | 전 빌드본 | 8013에 `X-Robots-Tag` 없음, `/robots.txt`가 `text/html`(SPA 응답). 집 서버 묶음(10-29 18:24)도 커밋 이전 빌드. **공개 도메인에 아직 안 들어갔다** | `curl`·`ls`·`unzip -l` |
| 코드-7(robots.txt 상충) | 중간 | `public/robots.txt` `Disallow: /` | 크롤러가 페이지를 못 읽으면 `noindex`도 못 본다. 외부 링크로 알려진 URL은 주소만 색인될 수 있어 "두 겹"이 서로 방해한다. 색인 제거가 목적이면 `Disallow` 제거 | Google 문서 내용(기억, 이번에 다시 조회 안 함) |

고친 내용: `TileWorkspace`는 `const { edge } = grids` · `admin.value` -> `admin` · `memberDropOrder`는 한 번 판정(`dropOnce`)을 더 안 바뀔 때까지 반복하고, 왕복이면 끄는 타일이 가장 앞인 순서로 고정(결과: 왕복 2,334회 -> 0회) · 왼쪽 띠 = `--edge-left`(그림만 맞춤, 커서가 목록 위일 때 넘어가는 판정은 그대로) · `robots.txt` = `Allow: /`

통과: Go `build`·`vet`·`test`, `e.Pre` 순서(헤더 -> 정적 서빙), 열기 규칙 `insertTile`, `DragSession` 재사용, 가로 1칸 타일 끼리의 끌기 판정.

## 설계(파트(업데이트 app)) — 번호-이름

| 번호(이름) | 심각도 | 충돌 | 근거 | 고칠 방향(제안) |
|---|---|---|---|---|
| 설계-1(지우고 깔기) | 높음 | `rsync --delete`·`robocopy /MIR`를 `TARGET_DIR`에 쓰면 env(`KEY_SECRET`·`API_KEY`)·worker sqlite(서브키)·`start.ps1`·`update\nexus-before.dump`까지 지워진다. `-hold`가 쥔 `worker.exe.old-*`는 안 지워져 robocopy가 실패로 끝날 수 있다 | `REF-home-deploy.md` "서버 PC 정리 기준"에서 `C:\Nexus`에 전부 같이 있음. 시안 설치 경로도 `C:\Nexus` | 앱 파일만 릴리스 폴더(`releases/<버전>`)에 깔고 `current`를 교체, 또는 지울 대상을 manifest로 한정 |
| 설계-2(서브키 불일치) | 높음 | updater가 "worker와 같은 방식"으로 서브키를 받으면 worker와 **다른 값**이 된다. 내 장비의 worker 인스턴스 줄(`#a41c`)에서 `updaters`를 찾을 수 없고 `useMemberDevice` 합치기도 안 된다 | `cmd/worker/router/register.go`: 서브키 = supervisor 발급, 자기 sqlite에 저장 | updater가 어떤 worker와 짝인지 키를 별도로 둘지 결정 필요 |
| 설계-3(잠금 계산) | 높음 | "성공 행으로 계산, 상태 컬럼 없음" + "테스트 배포는 안 잠금" + "`배포 완료`를 눌러 잠금"은 함께 성립하지 않는다. 테스트 성공 행만으로 잠기거나, 버튼 누른 사실을 둘 곳이 없다 | `REF-update-rules.md` "잠금" | 폴더별 잠금 기록(누가·언제·sha) 저장 |
| 설계-4(잠금 전 수정) | 높음 | 테스트 배포 뒤에도 수정 가능이라 "그 내용을 가진 장비가 없으므로 위험 0"이 거짓. 옛 내용으로 성공한 기록으로 새 내용이 잠길 수 있고, 안 잠긴 이전 릴리스로 되돌리면 그때 내용이 아닐 수 있다 | 같은 절, 시안 ②의 "배포 전에는 모두 수정" | 성공 행에 폴더 sha를 남겨 지금 sha와 같을 때만 성공으로 인정 |
| 설계-5(역할 없음) | 높음 | 릴리스 폴더가 os-arch로만 갈려 worker만 쓰는 장비에도 `supervisor.exe`와 그 단계가 간다. 시안의 `irony-server#a41d`(`D:\Nexus-test`)가 같은 `windows-amd64`를 받는다. "전체 일괄"에는 supervisor 먼저/나중 순서도 없다 | `REF-update-deploy.md` "릴리스 폴더"·시안 ① | 인스턴스 역할(`supervisor`·`worker`)과 일괄 순서 추가 |
| 설계-6(버전 이름 체계) | 중간 | 설치본 = 릴리스 이름(`261029_drag`), 실행 중 = 빌드 태그(`WORKER-DRAG-261029`)라 **값이 항상 달라** "재시작 필요"가 항상 뜬다. 실행 env 첫 줄 비교 경고도 env를 유지하는 한 업그레이드마다 뜬다 | `REF-update-rules.md` "버전 표시"·`cmd/worker/env` 첫 줄 | 릴리스에 worker 태그를 기록해 태그끼리 비교 |
| 설계-7(trap 신호) | 중간 | `#!/bin/sh` 예시의 `EXIT` trap은 dash·busybox에서 SIGTERM에 안 돈다. `timeout`으로 끊긴 `02-copy`가 복원 없이 끝난다. "`set -e` 밖의 실패를 잡는다"는 설명도 틀림(그런 실패는 스크립트를 끝내지 않음) | 직접 실행: dash·busybox = trap 안 돎, bash = 돎 | `trap 'exit 143' TERM`·`INT` 추가 또는 `#!/bin/bash` 규정 |
| 설계-8(supervisor 되돌리기) | 중간 | 재배포는 바이너리만 되돌리고 DB는 그대로. `pg_dump`는 뜨지만 복원 단계가 없다. 복원하면 그 뒤 데이터가 사라지고, 타일 순서 변환(`MigrateTileOrder`)은 단방향 | `REF-update-rules.md` 조건 표 | 스키마가 바뀐 릴리스는 자동 되돌리기를 막고 복원 절차를 수동으로 명시 |
| 설계-9(개별 실행과 설치본) | 중간 | `02-copy`만 개별 실행하면 디스크는 바뀌는데 `deployments`에 행이 없어 설치본 표시가 어긋난다 | "단계 개별 실행" 표 | 설치본을 기록이 아니라 updater가 디스크에서 읽어 보고 |
| 설계-10(자기 교체) | 중간 | supervisor를 교체하는 순간 updater의 로그·결과 전달 경로가 끊긴다. updater 자신의 교체는 어디에도 없다. 진행 중 행이 남을 수 있다 | `REF-update-deploy.md` 전체 | updater가 결과를 쌓아 재접속 때 보고 + 자기 교체 절 추가 |
| 설계-11(배포 정보 타일 열기) | 낮음 | 열 때 `parent` 규칙이 안 적혀 있고(적으면 장비 타일 바로 뒤), `KEY_RE`·`kindOf`·`KINDS`·`MemberPageKind`가 새 타일 종류와 `#`을 모른다. 콘솔(xterm)에 `data-pager-press-off` 필요 | `memberGrid.util.ts`·`memberDrag.store.ts` | P1(보관·화면) 착수 때 반영 |
| 설계-12(시안 표기) | 낮음 | 시안 로그 `/api/health`는 없는 API(지금 확인 = `/api/nope` 404). `STEP=start`는 REF의 `STEP_ID`와 다름. 시안 ①은 `update.ps1`, ②는 `steps.json`. 잘못 올린 릴리스 삭제 규칙("지울 수 없습니다")과 수정 규칙이 안 맞음. 시안 ①은 `261029_drag`의 `linux-arm64`가 잠김인데 그 버전을 설치한 linux 장비가 없고 팝업은 같은 폴더를 "잠기지 않음"으로 표시 | 시안 HTML | 시안 수정 때 같이 |

CURRENT의 확인 지점 답:
- 되돌리기 vs 잠금: 잠긴 릴리스도 대상(잠긴 뒤 배포 가능)이라 충돌 없음. 안 잠긴 쪽이 문제 → 설계-4(잠금 전 수정)
- 인스턴스 키: 같은 뜻 아님 → 설계-2(서브키 불일치)
- 단계 개별 실행: 잠금 제외는 일관됨. 설치본 표시가 문제 → 설계-9(개별 실행과 설치본)
- 같은 값 두 번: 실행 중 버전이 인스턴스 줄·타일 두 곳이나 출처가 하나라 문제 없음

## vault 상태
- `REF-member-drag.md`의 "vue-tsc는 통과"는 사실과 달랐음(이번에 고침). 왜 통과로 적혔는지는 스크립트·로그가 없어 확인 불가
- 끌기 확인 스크립트 `mdrag/`는 세션 scratchpad에 있어 사라짐 -> **`scripts/check/`로 저장소에 다시 씀**(`run.sh` = 임시 DB `nexus_check_tmp` + 5051, 끝나면 지움). 사용자 지적 3건·열기 규칙·끝 띠(관리·보드)·폰 터치(헤드리스)·색인 금지 포함. 실기기 폰·PC = 사용자 확인 끝(2026-11-02)
- 커밋 안 됐던 10-30 설계 문서는 11-02 커밋에 포함. 시안 `user/*`는 git 제외(저장소에 백업 없음)
- 파일 크기는 전부 분할 기준(10k자) 안
