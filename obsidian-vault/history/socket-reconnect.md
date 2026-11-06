# history — 브라우저 소켓 재연결 · 끊김 표시 (2026-11-06 신설)

> → `REF-socket-reconnect.md` · `CURRENT.md`

## 2026-11-06
- 사용자 질문 "모바일에서 브라우저 껐다 켜면 소켓이 끊기고 재연결 안 하는 것 같은데 맞아?" -> 코드 조사: `maxRetries: 2` 포기 + 화면 복귀 이벤트 없음 + 다시 붙어도 끊긴 사이 출력 버려짐(TUI 옛 화면). 처음엔 BACKLOG에 한 줄로 적음
- 사용자 증상 = 터미널 출력 멈춤, TUI 화면 안 바뀜 / 기대 = 돌아오면 저절로 붙고 놓친 화면도 보임
- 안 = 재연결-1(포기 없음)·재연결-2(돌아오면 바로)·재연결-3(죽은 연결 확인)·화면-1(다시 받기) + 끊김 표시 여부. 사용자 "재연결-2는 무슨 말?" -> 화면 복귀·네트워크 복구 순간에 간격을 기다리지 않고 바로 연결이라고 설명
- 사용자: 재연결-3 넣고 끊김 표시도 넣어 진행 -> 브랜치 `fix/socket-reconnect`. 서버 PING 핸들러 + hook `wake`·`probe`·`drop`·지수 간격 + `socketWake.util.ts` + `SocketLostPill.vue` + `processTerm.store.ts` `resync`. 옛 onclose 버그(연결 중 닫힘이면 재시도 안 함)도 같이 없어짐
- 확인 14/14(프록시 down·freeze, 첫 실행 1건 FAIL = 내 탭 id로 실행해 타일 push를 그 탭이 무시 -> 새로고침으로 고침) -> `REF-socket-reconnect.md` "확인"
- 20:58 사용자 요청으로 test3(8013) 교체: `build/261106-socket-reconnect`(supervisor+web, 태그 `SUPERVISOR-SOCKET-RECONNECT-261106`, worker 그대로) — `swap.sh`, 되돌림 대상 `261105-board-close`
- 사용자 모바일 확인 "모바일에서 잘 동작해" -> 마감, `fix/socket-reconnect` 커밋·master `--no-ff` 머지(push = 사용자)
