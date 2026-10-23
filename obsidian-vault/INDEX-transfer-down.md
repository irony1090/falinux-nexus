# INDEX — 내려받기·worker 간 전송 (마감)

> 2026-10-23 파트(내려받기·worker 간 전송) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
장비 타일에서 **worker 파일 내려받기**(파일 하나 그대로, 여러 개·폴더는 zip) + **worker 간 파일 복사**(메뉴·끌어 놓기). 브랜치 `feat/transfer-down`(`feat/transfer-ext` 위).
- 전송-7(worker 송신): supervisor가 worker에서 끌어옴 `FILE_OPEN`·`FILE_READ`·`FILE_CLOSE`. 파일 하나는 offset 자유(재접속이면 다시 열어 이어감), zip은 worker가 만들며 순서대로. supervisor `workerReader` = `transfer.Reader`
- 전송-8(내려받기: 파일 하나)·전송-9(내려받기: zip): POST 등록 -> 숨은 iframe GET `/transfers/:id/data` 한 번. 실패·중단은 Hijack으로 연결을 끊어 잘린 파일이 완료로 안 보이게
- 전송-10(worker 간 중계): A에서 읽어 B로 `send`(Abs·Exclusive), 동시 3, hash 생략(크기만)
- 전송-11(화면: 선택·메뉴)·전송-12(화면: 끌어 놓기): 체크박스 선택 + 도구줄(내려받기·다른 장비로 보내기·해제), 파일 줄을 다른 장비 타일로 끌어 놓기(전용 MIME)
- 전송-13(확인): 서버 격리 42/42 + 헤드리스 26/26(두 번) + **사용자 브라우저 확인 OK**(2026-10-23)
- 커밋: 브랜치 `feat/transfer-down`에 파트 마감 커밋 1개 -> master에 `feat/transfer-ext`·`feat/transfer-down` 차례로 `--no-ff` 머지

## REF / history 목록
- `REF-transfer-down.md`(결정 Q32~Q38 · 큰 계획 전송-7~13 · 구현 · 알려진 한계) / `history/transfer-down.md`
- 확인 스크립트 명세 → `REF-test-scripts-1022.md` "전송-7~10 `dl_test.py`"·"전송-11·12 `dl.mjs`" / 함정(헤드리스 drag 합성, cmdline 매칭 kill) → `REF-test-env.md`

## MEMORY 인덱스 줄 · HISTORY 색인 행 (원문)
- MEMORY: `REF-transfer-down.md`(파트(내려받기·worker 간 전송), 진행 중)
- HISTORY: | `history/transfer-down.md` | 내려받기·worker 간 전송(진행 중, 2026-10-23 신설) | 10-23 착수 Q32~Q38, 전송-7~13 구현·확인 |

## 이 파트에만 해당하는 결정 · 남은 것
- Q32(브랜치 시작점) `feat/transfer-ext` 위 · Q33(내려받기 범위) 셋 다 · Q34(worker 간 전송 시작 방법) 끌어 놓기 + 메뉴 · Q35(복사·이동) 복사만 · Q36(worker 간 폴더) 파일만 · Q37(항목 선택 방식) 체크박스 · Q38(내려받기 진행 표시) 전송 목록에도
- 알려진 한계(zip에서 읽을 수 없는 파일은 조용히 빠짐·zip 진행률 없음·내려받기 이어받기 없음·worker 간 hash 없음·받을 곳 = 보드에 열린 장비 타일만) → `REF-transfer-down.md`
- 못 본 것: Windows worker zip·복사 · 외부 접속(nginx 뒤) 긴 내려받기
- 폴더 전송(Q36)은 Q26(폴더째 올리기)와 함께 → `BACKLOG.md` "업로드 한계 보강"
