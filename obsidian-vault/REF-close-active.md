# REF — 닫은 뒤 active (파트(닫은 뒤 active), 2026-10-28)

> → 이력 `history/close-active.md` · → 진행 `CURRENT.md` · 계기 `REF-pager-head-btns.md`(패널에서 닫으면 active 없음) · active 원래 설계 `REF-node-ui-active.md`
> 브랜치 `feat/close-active`

## 확정 (사용자 2026-10-28 "진행")
| 항목 | 결정 |
|---|---|
| 언제 | active 타일이 **닫힐 때** 전부(패널 닫기 · 머리 줄 × · 터미널 종료 확인 뒤 닫힘). 보드 "다른 보드로 이동"도 배치에서 빠지므로 같음 |
| 어느 타일 | 위치 번호 이웃: **같은 번호(= 바로 다음 타일)**, 마지막이었으면 **앞 번호** |
| 그대로 해제 | 최소화(active 없음 유지) · 보드 바꾸기 |
| active 아닌 타일을 닫음 | active 그대로 |
| 범위 | 보드 + 관리 페이지 |
| 화면 | 새 active가 화면 밖이면 `reveal`(MEMORY "타일 이동"), 포커스는 옮기지 않음 |

## 구현
| 파일 | 내용 |
|---|---|
| `node/store/tileGrids.store.ts` | 옛 `watch(posOf)`(빠지면 null) -> `watch([posOf, hash])`: 빠지면 null, 보드가 바뀌었거나 최소화면 끝, 아니면 옛 번호 p로 새 배치에서 p -> p-1 순으로 찾아 active + `reveal` |
| `member/store/memberBoard.store.ts` | `close`: 닫기 전 인덱스 i -> 닫은 뒤 `tiles[i] ?? tiles[i-1]`을 active + `reveal`(`dropMissing = close`라 없는 대상 정리 때도 같음) |

## 확인 (격리, 2026-10-28)
- 헤드리스 `pw/near.mjs`·`pw/mini.mjs`: 가운데 닫기 -> 다음 타일 active / 마지막 닫기 -> 앞 타일 / 머리 줄 × -> 다음 / active 아닌 것 닫기 -> 그대로 / 관리 페이지 dave(2번) 닫기 -> minjun(새 2번) active / 터미널 최소화 -> active 없음. `vue-tsc` 통과
