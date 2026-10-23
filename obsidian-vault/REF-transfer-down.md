# REF — 내려받기·worker 간 전송 (폴더-7·폴더-9)

> 2026-10-23 신설·마감. 이력 → `history/transfer-down.md` / 원래 표 → `REF-worker-dir.md` 폴더-7(내려받기)·폴더-9(worker 간 전송) / 직전 파트(업로드·전송 상태) → `REF-transfer-ext.md`(`INDEX-transfer-ext`) / → `CURRENT.md`

**상태: 2026-10-23 마감(서버 격리 42/42 + 헤드리스 화면 26/26 + 사용자 브라우저 확인 OK) → `INDEX-transfer-down.md`.** 단계 번호는 직전 파트의 전송-1~6에 이어 붙인다. 브랜치 `feat/transfer-down`(`feat/transfer-ext` 위)

## 착수 결정 (2026-10-23 사용자 답)
| 번호(이름) | 결정 |
|---|---|
| Q32(브랜치 시작점) | **`feat/transfer-ext` 위 새 브랜치**(직전 파트는 브라우저 확인 뒤 순서대로 머지) |
| Q33(내려받기 범위) | **셋 다** — 파일 하나는 그대로, 여러 개·폴더는 zip 하나. zip은 worker가 만들며 흘려보냄 |
| Q34(worker 간 전송 시작 방법) | **타일 사이 끌어 놓기 + 메뉴**("다른 장비로 보내기", 보드에 열린 장비 타일 중에서 받을 곳 고르기) |
| Q35(복사·이동) | **복사만** — 보낸 쪽 파일은 그대로(삭제 명령 없음) |
| Q36(worker 간 폴더) | **파일만**(여러 개까지). 받는 쪽 폴더 만들기·겹침 규칙은 Q26(폴더째 올리기)와 같이 나중(BACKLOG) |
| Q37(항목 선택 방식) | **체크박스** — 줄에 마우스를 올리면 보이고, 하나라도 고르면 전부 보임. 폰도 같은 방식. 항목 클릭 동작은 그대로 |
| Q38(내려받기 진행 표시) | **전송 목록에도 올림** — 다른 탭·기기에서도 보이고 거기서 중단(계정 동기화 원칙) |

## 큰 계획 (2026-10-23 제안, Q36~Q38 반영)
| 단계 | 내용 | 손댈 곳 |
|---|---|---|
| 전송-7(worker 송신) | supervisor가 worker에서 **끌어오는** 방식: `FILE_OPEN`(경로 -> 읽기 id·크기·hash) / `FILE_READ`(id, offset -> 조각) / `FILE_CLOSE`. 다음 조각은 supervisor가 요청할 때만 읽으므로 받는 쪽(브라우저·worker B)이 느리면 같이 늦어짐 | `protocol`, worker `router/transfer.go` |
| 전송-8(내려받기: 파일 하나) | `GET /workers/:m/:s/download?folderId&path` -> 응답 본문으로 흘려보냄(Content-Length·파일 이름). 전송 목록 Kind = download | supervisor 새 `download.go` |
| 전송-9(내려받기: zip) | 여러 개·폴더 = worker가 zip을 만들며 보냄(크기 모름 -> 진행률은 보낸 양만). 경로 목록이 길어 `POST`로 등록 -> `GET /downloads/:id`(브라우저 기본 다운로드는 GET) | worker zip 스트림, `download.go` |
| 전송-10(worker 간 중계) | A에서 `FILE_READ` -> B로 `FileChunk`(Abs·Exclusive). 기존 `send`의 읽는 쪽만 "worker A"로 일반화. 같은 이름 확인 = B의 `DIR_EXISTS` + Q28(여러 파일 겹침) 창 | `transfer.go` `sendJob` 읽기 원천, 새 REST |
| 전송-11(화면: 선택·메뉴) | 장비 타일 항목 선택(지금 없음) + 메뉴 "내려받기"·"다른 장비로 보내기" | `DeviceTileBody.vue`, `transfers.store.ts` |
| 전송-12(화면: 끌어 놓기) | 장비 타일 항목을 다른 장비 타일에 끌어 놓기. 지금 있는 OS 파일 끌어 놓기와 구분(전용 MIME) | `DeviceTileBody.vue` |
| 전송-13(확인) | 격리 + 헤드리스(큰 파일·zip·중단·끊김·같은 이름) | scratchpad |

## 구현 (2026-10-23)
| 단계 | 위치 | 내용 |
|---|---|---|
| 전송-7(worker 송신) | `protocol` `FILE_OPEN`·`FILE_READ`·`FILE_CLOSE`, `DirErrIsDir` | `FileOpenRequest{ReadID, Path, Names}` — Names 없음 = 파일 하나, 있음 = Path 폴더 안 Names를 zip. 응답 Size -1 = 모름 |
| | worker `router/transfer_read.go` | 파일 하나 = `ReadAt`(offset 자유) / zip = 고루틴 `writeZip` -> `io.Pipe`(순서대로, 바로 앞 조각은 다시 줌 = 응답이 끊긴 재요청). 읽을 수 없는 하위 항목·폴더 링크·장치 파일은 건너뜀, 선택한 항목이 폴더 링크면 따라감, Deflate. FILE_READ 사이 60초 없으면 닫음 |
| | supervisor `router/workerRead.go` `workerReader` | `transfer.Reader` 구현. 호출마다 `waitWorker`로 지금 연결, 조각 실패면 3번까지 다시 — 파일 하나는 다시 열기(worker 재접속이면 읽기 목록이 빔, 크기가 바뀌면 실패), zip은 같은 offset만. `Hash()` = 빈 값(받는 worker가 hash 검사 건너뜀, 크기 검사는 함) |
| 전송-8·9(내려받기) | supervisor `router/download.go` | `POST /workers/:m/:s/downloads` {folderId, dir, names, path?, tileId} -> 인가 + worker에서 열기 -> 항목(kind/phase download). `GET /transfers/:id/data` = 한 번만, `Content-Disposition`(RFC 2231 한글), 파일 하나만 Content-Length. 실패·중단이면 **Hijack으로 연결을 끊음**(PanicMiddleware가 panic을 삼키고, 정상 종료하면 길이 없는 zip이 잘린 채 완료로 보임). GET이 1분 안 오면 aborted |
| 전송-10(worker 간 중계) | supervisor `router/copy.go` | `POST /transfers/copies` {from{folderId,mainKey,subKey,path}, to{…,dir}, overwrite, tileId} -> 두 폴더 인가 -> A에서 열기 -> 항목(kind copy, phase send, fromKey·fromPath, tileId = 받는 타일) -> `copySlots`(동시 3) -> `send`(Abs·Exclusive). 결과 분류 `endSend`를 업로드와 같이 씀 |
| 전송-11(화면: 선택·메뉴) | `DeviceTileBody.vue` | 줄 앞 체크박스(Q37), 선택 시 도구줄 "n개 · 내려받기 · 다른 장비로 보내기(v-menu = 다른 장비 타일 목록) · 해제". 폴더 이동 = 선택 비움, 작업 뒤 비움 |
| | `transfers.store.ts` | `download`(숨은 iframe으로 받음 — a 링크는 오류 응답에 앱 화면이 넘어감, 끝나고 30초 뒤 iframe 제거) · `copyTo`(파일만, 같은 이름 확인 `resolveNames`를 업로드와 같이 씀) · `spots`(장비 타일마다 지금 폴더 등록 = 받을 곳 목록) · `dragging` |
| 전송-12(화면: 끌어 놓기) | `DeviceTileBody.vue` | 파일 줄 `draggable`, MIME `application/x-nexus-device-files`(OS 파일 끌어 놓기와 구분), 선택한 줄을 끌면 선택한 파일 전부, 같은 장비의 같은 폴더에는 안 놓임 |
| 표시 | `TransferRow.vue`·`transfer.util.ts` | 종류 아이콘(올리기·내려받기·복사), "내려받는 중", 크기 모르면 옮긴 양 + 움직이는 막대, 받을 곳 = `보내는 장비 → 받는 장비 : 폴더` |

## 알려진 한계 (2026-10-23)
- zip에서 읽을 수 없는 파일은 알림 없이 빠짐 · zip 진행률 없음(옮긴 양만) · 내려받기 이어받기 없음
- worker 간 전송은 hash 검사 없음(크기만) · 폴더 전송 없음(Q36, BACKLOG "업로드 한계 보강")
- 받을 곳 목록 = 보드에 열린 장비 타일만(다른 보드·닫힌 장비는 안 보임) · 폰은 끌어 놓기 없음(메뉴로)
- 못 본 것: 실제 브라우저 drag(헤드리스는 이벤트 합성) · Windows worker zip·복사 · 외부 접속(nginx 뒤) 긴 내려받기
