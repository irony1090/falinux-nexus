# REF — 터미널 타일 PgUp·PgDn 버튼

> → 이력 `history/term-page-keys.md` / → 색인 `INDEX-term-page-keys.md`(마감 2026-10-19) / → 현재 `CURRENT.md` / 관련 → `REF-mobile-keyboard.md`(특수 키 줄) · `REF-node-ui-active.md`(active 타일)
> 2026-10-17 신설. 목적: 터미널 타일이 active일 때 vim·less 같은 전체 화면 프로그램에서 쪽 단위로 이동. **상태(2026-10-17): 페이지-1(버튼 표시)·페이지-2(키 전송) 코드 작성, 타입 검사 통과. 페이지-3(확인) = 2026-10-19 사용자 브라우저 확인 완료· 마감·커밋.** 구현 메모: 전체보기(`overview`) 중에도 숨김(본문이 `inert`), 전송은 `click`(pointerdown은 막기만)

## 동작 = 키 전송 (동작-가)
- 버튼을 누르면 PgUp `\x1b[5~` / PgDn `\x1b[6~`를 PTY로 보낸다(xterm 스크롤 동작-나 아님).
- 일반 셸에선 효과가 거의 없음 — 의도된 범위(vim·less·htop용).
- 수식 키: 특수 키 줄에서 Ctrl/Alt가 눌려 있으면 `takeCsiMod()` n > 1 -> `\x1b[5;n~` / `\x1b[6;n~`.

```
TerminalTileBody.vue (overlay: PgUp / PgDn)
  show = activeId == tileId && status == PROCESS && !readOnly && !(kb.open)
  pointerdown.prevent  -> keep xterm focus
        |
        v
  takeCsiMod()  ->  n > 1 ? "\x1b[5;n~" : "\x1b[5~"   (PgDn = 6)
        |
        v
  sendText(uid, seq)  -> existing input path -> PTY
```

## 결정 (2026-10-17 사용자 확정, 제안 그대로)
| 결정 | 확정 |
|---|---|
| PG1(버튼 위치) | 타일 본문 **오른쪽 아래** 반투명 버튼 2개(오른쪽 위는 "크기를 가져갔습니다" 칩 자리) |
| PG2(키보드 열림) | 모바일 가상 키보드가 열린 동안 **숨김**(타일이 위로 이동해 아래가 특수 키 줄에 가려질 수 있음). 그 동안은 특수 키 줄 `Ctrl` + `F`/`B` |

## 작업 단위
| 단계 | 파일 | 할 일 |
|---|---|---|
| 페이지-1(버튼 표시) | `TerminalTileBody.vue` | 오버레이 + 표시 조건(`useTileGrids().activeId`, 실행 상태 `PROCESS`, `isReadOnly`, `useVirtualKeyboard().open`) |
| 페이지-2(키 전송) | 같은 파일 | `TerminalKeyBar.vue` `down()`과 같은 패턴으로 `sendText` + `takeCsiMod` |
| 페이지-3(확인) | 브라우저 | vim·less 쪽 이동, 누른 뒤 xterm 포커스 유지, 보기 전용 공유에서 안 보임 |
