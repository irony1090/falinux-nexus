# INDEX — 파일 전송 확장 (업로드·전송 상태, 마감)

> 2026-10-23 파트(파일 전송 확장) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). 폴더-7(내려받기)·폴더-9(worker 간 전송)은 이 파트에서 빼고 다음 파트 후보로(BACKLOG). → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
장비 타일이 **보고 있는 폴더에 업로드** + 계정 단위 **전송 상태**. 브랜치 `feat/transfer-ext`(master `3b59c61`에서).
- 전송-1(절대 경로 수신): worker 수신에 Abs 모드(받을 폴더 + 이름, worker가 합침)·Exclusive(이미 있으면 거절)·거절 코드 4개
- 전송-2(업로드 REST): 같은 이름 확인 `dir/exists` -> `uploads` 시작 -> 512KB 조각 `PUT /transfers/:id/data` -> `finish` -> 뒤에서 worker로 send. Q30(조각 업로드) = HTTP 조각(앞단 nginx 1MB 대비)
- 전송-3(전송 목록): supervisor 메모리 목록 + 계정 토픽 `TRANSFER:UPDATE`·`TRANSFER:GONE`, 중단 `DELETE /transfers/:id`(worker `.part`까지), 끝난 항목 5분·멈춘 올리기 10분
- 전송-4(화면: 올리기)·전송-5(화면: 전송 상태): 장비 타일 올리기 버튼·OS 파일 끌어 놓기·바닥 막대(그 타일에서 시작한 것), 푸터 오른쪽 전송 버튼(Q31), 같은 이름 한 창(Q28), 끝나면 그 폴더 목록 다시 읽기
- 확인: 단위(overlay) + 서버 격리 26/26 + 헤드리스 18/18. **사용자 브라우저 확인 전**(test3 빌드본 아님)
- 커밋: 브랜치 `feat/transfer-ext`에 파트 마감 커밋 1개. master 머지는 사용자 브라우저 확인 뒤

## REF / history 목록
- `REF-transfer-ext.md`(결정 Q20~Q31 · 큰 계획 전송-1~6 · 구현 · 알려진 한계) / `history/transfer-ext.md`
- 확인 스크립트 명세 → `REF-test-scripts-1022.md` "2026-10-23 전송-2+전송-3"·"전송-4·5"
- (MEMORY에 남김) `REF-worker-dir.md`(폴더-7·9 원래 표 = 다음 파트 계획)

## HISTORY 색인 행 (원문)
| `history/transfer-ext.md` | 파일 전송 확장(진행 중, 2026-10-23 신설) | 10-23 착수 Q20~Q26, 큰 계획 전송-1~6 |

## 이 파트에만 해당하는 결정 · 남은 것
- Q23(같은 이름 파일) 묻기 · Q24(전송 상태 위치) 푸터+타일 · Q25(진행률 구간) 한 줄 단계 표시 · Q26(폴더째 올리기) 나중에 · Q27(타일 막대 범위) 그 타일에서 시작한 것 · Q28(여러 파일 겹침) 한 창 · Q29(전송 목록 보관) 메모리만 · Q31(전송 아이콘 자리) 푸터 오른쪽 종 옆
- 알려진 한계(파일 하나씩·탭 닫으면 이어 올리기 없음·다시 읽기는 첫 페이지·받을 곳 뒤 잘림, Exclusive 재시도 오판, Linux 이름의 `\`) → `REF-transfer-ext.md`
- 못 본 것: Windows worker로 업로드(경로 합치기는 worker가 하므로 될 것으로 봄) · 외부 접속(nginx 뒤) 큰 파일 · 모바일 화면
- 다음 파트 후보: 폴더-7(내려받기)·폴더-9(worker 간 전송) — 전송 목록·push를 그대로 쓴다(Kind만 추가) → `REF-worker-dir.md`, `BACKLOG.md`
