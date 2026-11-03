# history — 업데이트 app P4(Windows)

→ `REF-update-p4.md` · 앞 단계 `history/update-p3.md` · `CURRENT.md`

## 2026-11-03 — 착수 조사 · 계획 초안
- P3 머지(`cb75903`) 뒤 사용자 "P4 작업 진행하자" -> 코드 착수 전 조사: `exec_other.go`(빈 자리)·`runPipe`(파일 바로 실행)·`copyOne`(이름 변경만, `*.old-` 처리 없음)·ConPTY Job `KILL_ON_JOB_CLOSE`·설치/빌드 절차 없음
- 착수 계획 단위-1(스크립트 실행)~단위-7(교체 빌드) 초안 + 조-1(붙을 supervisor)~조-3(설치 자리·계정) → `REF-update-p4.md`. 사용자 답 대기
- 사용자 답: 조-1 = 가(집 서버 자신, 첫 배포는 Claude가 만든 `update.ps1`로 수동) · 조-2 = VM 먼저 · 조-3 = `C:\Nexus\updater` 제안 -> 보완-1(자기 폴더 보호)·보완-2(폴더 권한) 제안

## 2026-11-03 낮 — 구현 · VM 확인
- 사용자 "그 순서 맞아, 진행해줘" -> 브랜치 `feat/updater-p4`. 단위-1~5 구현(구현 지도 `REF-update-p4.md`)
- `run.sh` 첫 실행에서 updater-deploy-ui 3개 FAIL(ask 건너뜀) -> 원인 = 보완-1이 updater 폴더 전체를 막아 설치 경로(`work/target`)가 막힘 -> 차이-1(보호 범위)로 좁힘 -> 193/193
- VM: `docker start nexus-win` -> ssh 래퍼 `win/vssh` 다시 만듦 -> 33/33 첫 실행 통과. 추가 확인에서 차이-2(Start-Process 자식이 updater 재시작에도 남음) 발견, (b) 배포 중 supervisor 재시작 통과
- 정리: 격리 supervisor 끔 · `nexus_updwin_tmp` 삭제 · VM 작업 멈춤 · `docker stop nexus-win`


## 2026-11-03 오후 — 집 서버 설치 · 마감 · 머지
- 사용자 집 서버(회사에서 Nexus 터미널로): `update.ps1`(12:35) -> `setup.ps1`(12:42) -> `C:\Nexus\updater`·예약 작업 `NexusUpdater` 동작 "잘 동작했어" → `history/home-deploy-updater.md`
- 사용자 "p4 작업도 완료가 된거지? 그러면 머지까지" -> 코드 단위는 다 끝남, 단위-7(교체 빌드)은 코드 없는 보관소 빌드라 후속으로 뗌(`REF-update-p4.md` "단위-7(교체 빌드) 후속", 정-1~정-3 사용자 답 대기) -> `go test ./...`·`GOOS=windows go vet` 통과 -> `feat/updater-p4` 커밋 + master 머지(push = 사용자)
