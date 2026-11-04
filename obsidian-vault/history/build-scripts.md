# history — 빌드 스크립트 (2026-11-04 신설)

> → `REF-build-scripts.md` · 빌드본 분리 이력 `history/build.md`

## 2026-11-04 — 대상별 빌드 스크립트
- 사용자 요청: supervisor·worker·updater·frontend 각각 빌드할 쉘로 나누고 `build.sh`가 불러오게, 결과는 `build/{폴더명}`에
- 조율 Q1(대상 생략 시)~Q4(updater 출력 모양) → 사용자 "전부 추천대로": 생략 = supervisor·worker·frontend / 대상별 이미 있음 / `--os`·`--arch` 넣음 / updater = `build/<폴더명>/updater/{linux-arm64, windows-amd64}`
- 구현 `scripts/build.sh` + `scripts/build/{common,supervisor,worker,frontend,updater}.sh`, 옛 `scripts/updater/build.sh` 삭제. 시험 중 고침: 같은 태그 검사가 `build/*/updater` 폴더를 grep(`-f`로), 실패 뒤 빈 폴더 남음(`fail_clean`)
- `REF-build.md`가 10.2k라 "빌드 태그" 절과 함께 새 쌍으로 분리
- 첫 실사용(사용자, 집 서버 빌드): 대상 생략 = updater 빠짐을 되물음 -> 대상에 `updater`를 적으면 한 줄로 됨(--os 무시 안내 출력). 같은 태그 경고 위치 = `warn_same_tag`(build/ 아래 실행 파일 안 `# 태그` grep) 설명
- 커밋·master `--no-ff` 머지(11-04, push = 사용자). 같은 커밋에 master에서 넘어온 `.gitignore`(`/build/`)·`apps/core/.gitignore`(updater 실행 흔적)·vault 포함
