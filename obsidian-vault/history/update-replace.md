# history/update-replace — Files 타일 단위-다(갈아치우기)

→ REF-update-replace.md · → history/update-files-tile.md · → CURRENT

## 2026-11-03 밤 — 단위-다(갈아치우기) 구현
- 확인-1(토글 깊이) 남은 제안 확정(사용자 "제안대로"): 부모·자식 둘 다 있으면 경고 + 배포 메시지엔 부모만, 토글로 부모를 켜면 자식 항목 자동 삭제
- 다-1~6 구현(상세 = `REF-update-files-tile.md` "단위-다" 구현 결과). 시험: Go `updstore`·`cmd/updater` 통과(처음 1건 = 시험 기대 줄 번호 오기), `updater-deploy.py` 7절 실제 updater로 통과, `updater-files-ui.mjs` 토글 확인 통과, 전체 run.sh exit 0
- 결정(구현 중): `*.old-*`(물린 옛 파일)는 갈아치우기에서도 남김 · 이번에 쓸 파일은 지우지 않고 덮어씀(결과 같음) · 장비에 아직 없는 폴더는 건너뜀 · 푸터 안내는 좁은 타일에서 잘려 `갈아치우기 N곳`으로 줄임

## 2026-11-03 밤 — 화면 확인 · 열린 편집 타일 맞추기
- test2 supervisor 다시 띄움(사용자 요청). 사용자 화면 확인: 토글 켜기·자식 흐림·합치기 끄기 OK, steps.json 타일은 새로고침해야 반영 — 원인 = 편집 타일이 처음 한 번만 읽는 구조(고치던 글 보호)
- 기대 모습 확인 뒤(고치던 중이면 덮지 않고 안내 + 다시 불러오기, Script 타일도) `useUpdLiveText`로 두 타일 공용화. `updater-files-ui.mjs` +5, 전체 run.sh exit 0
