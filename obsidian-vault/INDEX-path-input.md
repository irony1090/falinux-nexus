# INDEX — 경로 직접 입력 (마감)

> 2026-10-23 파트(경로 직접 입력) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
장비 타일·폴더 타일 경로 줄에 **경로를 직접 치기 + 단계별 자동완성**. 브랜치 `feat/path-input`(master `b17f77e`에서).
- 경로-1(후보 요청): worker `DIR_SUGGEST`(폴더만, 앞글자 대소문자 무시, 같은 대소문자 먼저, 숨은 폴더는 `.`을 칠 때만, 상한 200, Windows `C:` = `C:\`)
- 경로-2(후보 REST): `GET /workers/:m/:s/dir/suggest`(옛 worker = 502 -> 화면은 후보 없이 이동)
- 경로-6(폴더 후보 REST): `POST /nodes/suggest {baseId, names, prefix, max}` 하나로 후보·Enter(같은 이름 = 첫 번째)
- 경로-3(화면: 입력칸)·경로-4(화면: 자동완성): `PathInput.vue` — 빈 곳 클릭·Ctrl+L, ↑↓·Enter 고르기, Tab 공통 앞부분, Esc 되돌림
- 경로-7(화면: 폴더 타일 입력): `기본/…`, 고른 단계·지금 경로 단계를 id로 기억, 이름 속 `/`는 `\/`
- 경로-5(확인): 서버 33/33 + 헤드리스 31/31(5번 연속) + Windows VM 서버 14/14·화면 12/12 + **사용자 브라우저(Linux·Windows) 확인 OK**(2026-10-23)
- 커밋: 브랜치 `feat/path-input`에 파트 마감 커밋 1개 -> master `--no-ff` 머지

## REF / history 목록
- `REF-path-input.md`(결정 Q39~Q44 · 큰 계획 경로-1~7 · 구현 · 알려진 한계) / `history/path-input.md`
- 확인 스크립트 명세 → `REF-test-scripts-1023.md`(이 파트에서 새로 엶) / VM → `REF-windows-vm.md`(이 파트 중 `REF-windows-worker.md`에서 분할)

## MEMORY 인덱스 줄 · HISTORY 색인 행 (원문)
- MEMORY: `REF-path-input.md`(파트(경로 직접 입력), 진행 중)
- HISTORY: | `history/path-input.md` | 경로 직접 입력(진행 중, 2026-10-23 신설) | 10-23 착수 Q39~Q44, 경로-1~7 구현·확인 |

## 이 파트에만 해당하는 결정 · 남은 것
- Q39(입력 들어가는 법) 빈 곳 클릭 + Ctrl+L · Q40(자동완성 대상) 폴더만 · Q41(자동완성 조작) 아래 목록 + Tab · Q42(자동완성 원천) worker가 거름 · Q43(같은 이름 폴더) 목록 순서대로 첫 번째 · Q44(이름 속 /) 자동완성으로만
- 알려진 한계(파일 경로는 Enter에서 "폴더가 아닙니다"·장비 후보 50개·`/` 이름은 목록으로만·내 루트가 공유 루트보다 먼저) → `REF-path-input.md`
- 못 본 것: 외부 접속(nginx 뒤) · 모바일(입력칸 폭·가상 키보드)
- 사용자 빌드 `build/directory_path`(test3)는 Windows `C:` 고치기 전 커밋 상태(Linux 무관). VM `worker-test3.exe`는 고친 것
