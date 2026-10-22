# INDEX — 빌드본 분리 (마감)

> 2026-10-22 파트(빌드본 분리) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
목적 = 사용자가 Nexus 위에서 쓰는 Claude 세션이 개발 중 supervisor·worker 재시작으로 끊기지 않게, 사용용 빌드본을 따로 띄운다. master에 바로 한 커밋.
- B1(포트·도메인) 8013(test3) · B2(DB 분리) PostgreSQL `nexus_build` · B3(데이터 복사) `nexus` 덤프 + worker sqlite 복사 · B4(반복 배포) `scripts/build.sh <이름>` -> `build/<이름>` · B5(웹 경로 변수) `WEB_DIR`로 supervisor가 웹 직접 서빙
- 빌드-1(env 옵션)·빌드-2(웹 서빙)·빌드-3(같은 출처 API)·빌드-4(빌드 스크립트) 구현
- env 내뱉기: E1(embed = 배포 값)·E2(개발 = dev.env)·E3(없으면 내뱉고 종료)·E4(git)·E5(스크립트)
- 공통 env·sqlite = `build/` 바로 아래(사용자 배치), 실행 = `cd build/<이름> && ./supervisor -env ../supervisor.env`
- 마감 확인 = `build/const`(8013) 실사용

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- `REF-build.md`(빌드본 분리) / `history/build.md`

## HISTORY 색인 행 (원문)
### `history/build.md` — 빌드본 분리 (2026-10-22 신설)
- 2026-10-22 조율 B1~B7 -> 빌드-1(env 옵션)·2(웹 서빙)·4(빌드 스크립트) 구현·확인, `nexus_build` 덤프, env·sqlite `build/` 공통, 빌드본 8013 실사용 시작 / env 내뱉기 E1(embed = 배포 값)·E2(개발 = dev.env)·E3(없으면 내뱉고 종료)·E4(git)·E5(스크립트) 구현·확인
- 2026-10-22(2) `build/const` 재실행 확인, `dev.env` 이름 맞춤, 파트 마감·커밋

## 이 파트에만 해당하는 결정
- 함정(빌드 스크립트): `npm run build`는 `--outDir`를 vite root로 넘김 -> `type-check` + `npx vite build --outDir` 따로 / `IsDev()`는 실행 파일 경로에 `/tmp`가 있으면 개발로 판정 -> `/tmp` 아래 빌드본은 `-env` 상대 경로 = 소스 폴더
- 다른 파트에도 지킬 것(8013 끄지 않기, 개발 실행 `-env dev.env` 필수, embed `cmd/*/env` = 배포 값)은 MEMORY 불변 결정 "빌드본 = 사용용"에 남김
