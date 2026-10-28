# history — 닫은 뒤 active (파트(닫은 뒤 active))

> → `REF-close-active.md` / `CURRENT.md`

## 2026-10-28
- 계기: 파트(패널 창 버튼) 재현 중 "패널에서 닫으면 active 없음 -> 패널 버튼 없음" -> 사용자 "가까운 타일을 active로" -> 해석(위치 번호 이웃, 머리 줄 ×도, 보드·관리 둘 다) 확인 "진행"
- 구현(보드 = posOf·hash 감시, 관리 = close) -> 격리 헤드리스 6/6(관리 1건은 판정식 실수, 결과는 정상) → REF "확인"
