# history/test-scripts — 확인 스크립트 재작성 명세

> 재사용 지식 → `REF-test-scripts.md` / 환경 → `REF-test-env.md`·`history/test-env.md` / 현재 진행 → `CURRENT.md`

## 2026-10-14 — `REF-test-env.md`에서 분할 + 파트(확인·실측) 스크립트 추가
- 함정 3줄(socat fork 자식·pkill 자기 셸·`/users/presence` 범위)을 넣으면 `REF-test-env.md`가 10k자를 넘어 날짜별 확인 스크립트 절(2026-10-08 요령 ~ 2026-10-13)을 통째로 옮김

## 2026-10-19 — CURRENT에서 내린 옛 세션 메모(원문)
- **2026-10-15(2) 세션**: 격리(5051·`nexus_orphan_tmp`) 종료·임시 DB 삭제. scratchpad `chk/client.py`·`chk/orphan.py`·`orphan/`·`old/`는 재부팅 시 사라짐 → 명세 `REF-test-scripts-1014.md` "2026-10-15 확인 스크립트 (고아 스크립트)". worker는 사용자가 새 빌드로 재실행함
- **2026-10-15 세션 재시작(Claude만, PC 재부팅 아님)**: 사용자가 띄운 supervisor(8012)·worker·프론트(8011)·postgres15는 켜진 채. 격리 테스트 프로세스·고아 스크립트 정리 끝, 임시 DB 없음. 이 세션 scratchpad(`pw/mid.mjs` 등)는 새 세션에서 경로가 달라 안 보임 — 명세 → `REF-test-scripts-1014.md` "2026-10-15". PC 재부팅 뒤라면 `docker start postgres15` -> supervisor -> worker -> `npm run dev`(→ MEMORY "로컬 개발 환경")
- 재부팅으로 사라질 것(2026-10-19): scratchpad `win/vssh`·`iso6/win_test.py`(VM 확인)·`touch/`(터치 헤드리스)·`seed/seed.py`(노드 본문은 DB에 있음) — 명세 → `REF-test-scripts-1014.md` "2026-10-19" 두 절. 격리 5051·`nexus_win_tmp`는 정리함
- 재부팅으로 사라질 것(2026-10-15): scratchpad `pw/mid.mjs`·`iso/`·`vite.iso.mts` — 명세 → `REF-test-scripts-1014.md` "2026-10-15". 격리 종료·`nexus_bp_tmp` 삭제함
- 재부팅으로 사라질 것(2026-10-14 보드 미리보기 세션): scratchpad `pw/bp.mjs`·`pw/bp2.mjs`·`pw/node_modules`·`vite.iso.mts`·`iso/`(바이너리·env)·`bp/`(시안 캔버스 원본, 캔버스 자체는 claude.ai에 남음) — 명세 → `REF-test-scripts-1014.md` "2026-10-14 보드 미리보기". 격리 프로세스 종료·임시 DB `nexus_bp_tmp` 삭제함. 테스트 계정은 임시 DB에만 있었음(개발 DB 무변경)
- 이전 재부팅들(2026-10-12~13)로 사라진 스크립트 명세는 `REF-test-scripts.md`·각 파트 REF 확인 절에
