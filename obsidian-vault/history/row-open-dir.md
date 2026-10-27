# history — 파트(칸 메뉴 장비 디렉토리)
> → `REF-row-open-dir.md` · → `CURRENT.md`

## 2026-10-27
- 계기: 파트(장비 상태 push) 마감 직후 사용자 요청 — 폴더 칸 ⋮ 메뉴에 장비 디렉토리 열기, 공유받은 사람은 안 보임, 주인은 오프라인이면 꺼짐
- 조율: 장비 없는 폴더 = 숨김 · 상속 포함 · 스크립트 제외 → "진행"
- 구현: 브랜치 `feat/row-open-dir`(master `e16a519`에서), `FolderTileBody.vue`만. 격리 헤드리스 확인 OK(→ REF "확인")
- 사용자 직접 확인 → "확인" → 격리 정리 → 파트 마감·`feat/row-open-dir` 커밋 → master `--no-ff` 머지 → `INDEX-row-open-dir.md`
