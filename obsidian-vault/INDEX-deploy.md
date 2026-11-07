# INDEX — 빌드·배포 (상황 색인)

> **상황 색인**: 마감 파트가 아니라 **빌드·배포할 때만 여는 문서 묶음**(2026-11-07 사용자, 전역 CLAUDE.md "색인 2단화" 상황 색인 규칙). 세션 시작 시 읽지 않는다. REF/history 파일은 옮기지 않았다. → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`
> 새 빌드·배포 문서가 생기면 여기 표에 한 줄 추가(MEMORY·HISTORY에 따로 줄을 만들지 않음)

## 상황별로 열 문서
| 상황 | 열 문서 | 거기서 볼 것 |
|---|---|---|
| 빌드 만들기(어느 대상이든) | `REF-build-scripts.md` | `scripts/build.sh` 사용법·대상별 출력·**빌드 전에 `cmd/*/env` 첫 줄 태그 바꾸기**(env는 git 제외) |
| test3(8013) supervisor 교체 | `REF-build.md` · 직전 빌드 폴더의 `swap.sh` | 빌드본 구조·두 빌드 나란히·`CLAUDE*` 변수 상속 함정·웹만 교체. 교체 기록은 `CURRENT.md` 운영 상태 |
| 집 서버 배포 묶음(updater로 올림) | `REF-home-deploy-updater.md` | 최근 묶음 절 하나를 따라 만듦: Windows 빌드 + worker 항상 포함 + `steps.json`·`nexus-steps/` 승계, 배포 뒤 공개 해시 확인(nexus·direct) |
| 보드(i.MX, arm64) worker 교체 | `REF-build.md` "보드 worker 교체" | IP가 자주 바뀜 -> 사용자에게 먼저 물음 |
| 집 서버 접속·인증서·PC 구성 | `REF-home-direct.md`(지금, iptime + Caddy) · `REF-home-server.md`(PC 구성·재부팅·옛 Cloudflare 터널) | |
| 옛 수동 배포(update.ps1, 11-03 전) | `REF-home-deploy.md` | 참고용 |
| updater 자체 설계·규칙 | `INDEX-update-app.md` | 업데이트 app 파트(마감) |

## REF / history 목록
- `REF-build-scripts.md` / `history/build-scripts.md` · `REF-build.md` / `history/build.md`(빌드본 분리 파트 = `INDEX-build.md`)
- `REF-home-deploy-updater.md` / `history/home-deploy-updater.md` · `REF-home-deploy.md` / `history/home-deploy.md`
- `REF-home-direct.md` / `history/home-direct.md` · `REF-home-server.md` / `history/home-server.md`

## MEMORY 줄 · HISTORY 행 (원문)
- MEMORY: - **빌드본 = 사용용**: `build/<이름>`(`scripts/build.sh`, 대상별 `scripts/build/*.sh` → `REF-build-scripts.md`) = **8013(test3) · DB `nexus_build`**. 사용자 Claude 세션이 그 위에서 돎 — **Claude는 8013을 끄지 않는다**(재기동 = 사용자). 개발 실행 = `go run ./cmd/supervisor -env dev.env` 필수 (→ `REF-build.md`)
- MEMORY: - **배포·업데이트**: 업데이트 app(마감) → `INDEX-update-app.md`(설계·규칙·P1~P4·Files 타일·시험 실행·갈아치우기·10-30 점검) · 집 서버 `REF-home-deploy-updater.md`(updater 설치·묶음 배포, 지금 방식) · `REF-home-deploy.md`(옛 수동 update.ps1) · `REF-build.md`(보드 worker 교체)
- HISTORY(빌드본 분리 아래): - 2026-11-04 `history/build-scripts.md` 신설(↔ `REF-build-scripts.md`) — 빌드 스크립트 대상별로 나눔 · 집 서버 빌드로 실사용 · 커밋·master 머지
- HISTORY: | `history/home-server.md` | 집 Windows 서버 PC 구성 + Cloudflare Tunnel(2026-10-25 신설) | 10-28 원인 = 엣지 LAX -> 직통 전환 계획·회사 준비 |
- HISTORY: | `history/home-direct.md` | 집 서버 직통 접속 iptime + Caddy(2026-10-28 home-server에서 분할) ↔ `REF-home-direct.md` | 10-28 밤 직통 전환 끝, cloudflared 정지 |
- HISTORY: | `history/home-deploy.md` | 집 서버 배포·업데이트(2026-10-28 home-server에서 분할) | 10-29 저녁 drag 묶음 배포 완료(worker 항상 포함 규칙) |
- HISTORY: | `history/home-deploy-updater.md` | 집 서버 배포: updater 이후(2026-11-03 신설) ↔ `REF-home-deploy-updater.md` | 11-04 최신 배포 #2 · 11-05 보드 닫기 묶음(11-06 사용자 배포) · 11-06 소켓 재연결 묶음(업로드 대기) |
