# REF — 파일 전송 확장 (업로드·전송 상태·내려받기·worker 간 전송)

> 2026-10-23 신설·마감(→ `INDEX-transfer-ext.md`). 이력 → `history/transfer-ext.md` / 전체 계획(폴더-5·7·8·9 원래 표) → `REF-worker-dir.md` / 전송 모듈 → `REF-transfer.md`·`REF-transfer-e2e.md`(`INDEX-infra`·`INDEX-transfer-e2e`) / 장비 디렉토리 인가 → `REF-worker-dir-list.md` / → `CURRENT.md`

**상태: 전송-1(절대 경로 수신) 구현 끝(단위 확인만, 실제 전송 확인은 전송-6). 전송-1~5 구현, 서버 격리 26/26 + 헤드리스 화면 18/18. 남은 것 = 사용자 확인 -> 파트 마감(폴더-7(내려받기)·폴더-9(worker 간 전송)은 다음 파트로 나눌지 결정).** 브랜치 `feat/transfer-ext`(master `3b59c61`에서)

## 착수 결정 (2026-10-23 사용자 답)
| 번호(이름) | 결정 |
|---|---|
| Q20(진행 순서) | **폴더-5(업로드)+폴더-8(전송 상태) 같이 -> 폴더-7(내려받기) -> 폴더-9(worker 간 전송)**. 업로드가 기존 supervisor -> worker 송신을 그대로 써서 전송 상태의 첫 사용처가 됨 |
| Q21(SFTP) | **진행할 작업이 전혀 없을 때 이야기** (BACKLOG) |
| Q22(브랜치) | `feat/transfer-ext` |
| Q23(같은 이름 파일) | **묻기** — 덮어쓸지 확인 창. worker도 덮어쓰기 표시 없으면 거절(목록이 오래된 경우 대비) |
| Q24(전송 상태 위치) | **둘 다** — 헤더 아이콘(진행 중 개수 배지)+팝업 목록 **그리고** 장비 타일 안 진행 막대 |
| Q25(진행률 구간) | **한 줄, 단계 표시** — "올리는 중 n%"(브라우저 -> supervisor) 끝나면 "장비로 보내는 중 n%"(supervisor -> worker) |
| Q26(폴더째 올리기) | **나중에** — 이번엔 파일 여러 개까지. 폴더는 폴더-7(내려받기) 폴더 zip과 같이 생각 |
| Q27(타일 막대 범위) | **그 타일에서 시작한 것** — 폴더를 옮겨도 계속 보임. 전송 항목에 시작한 타일 id를 서버가 같이 보관(새로고침·보드 이동 뒤에도 이어짐) |
| Q28(여러 파일 겹침) | **한 창에 모아서** — 겹치는 목록 + "모두 덮어쓰기 / 겹치는 것 건너뛰기 / 취소" |
| Q29(전송 목록 보관) | **메모리만** — supervisor 재시작 시 목록·진행 중 전송 사라짐, 끝난 항목은 몇 분 뒤 빠짐 |
| Q31(전송 아이콘 자리) | **푸터 오른쪽 알약, 위치 알림 종 옆** — 앱에 헤더가 없음(헤더·푸터 개편 때 앱 버튼이 푸터로). 전송이 있을 때만 보임, 배지 = 진행 중 개수, 실패만 남으면 빨간 점, 폰은 '더 보기' 안 |
| Q30(조각 업로드) | **HTTP 조각(512KB씩)**. 이유 = nginx(다른 장비) `client_max_body_size` 기본 1MB를 설정과 무관하게 피함 + 이어받기 + 기존 REST(세션·`panic(web.Err)`) 그대로. 웹소켓(업로드 전용 소켓)도 검토 — 이점이 같고 재연결·인증·프레임 형식을 새로 만들어야 해서 안 함. 같은 이름 확인 API(①)도 추가 승인 |

## 큰 계획 — 폴더-5(업로드)+폴더-8(전송 상태) (2026-10-23 제안)
| 단계 | 내용 | 손댈 곳 |
|---|---|---|
| 전송-1(절대 경로 수신) | worker 수신에 "절대 경로" 모드 추가(디스크 전체, Q2). `ProcessRoot` 밖 차단은 기존 실행 전 선배치에만 | `protocol`, worker `router/transfer.go` `resolveDest` |
| 전송-2(업로드 REST) | `POST /workers/:mainKey/:subKey/upload?folderId&path`, 인가 = `requireDeviceDir` 재사용. 본문 -> supervisor 임시 파일 -> 바로 응답 -> 뒤에서 `SendFile` | `workerApi.go`, 새 `upload.go` |
| 전송-3(전송 목록) | supervisor 메모리 목록 {id·계정·시작 타일 id(Q27)·장비·경로·크기·보낸 양·상태}. `send`에 진행 콜백 -> `USER:<id>` push(초당 몇 번 제한). 중단 = `AbortFile` | supervisor `router/transfer.go`, `subscribe.go` |
| 전송-4(화면: 올리기) | 장비 타일 메뉴 "올리기"(여러 파일 선택) + OS 파일 끌어 놓기 | `DeviceTileBody.vue`, api |
| 전송-5(화면: 전송 상태) | 헤더 아이콘+팝업 목록, 장비 타일 안 막대(Q24), 중단 버튼, 끝나면 그 경로를 보는 장비 타일 목록 새로 읽기 | 새 컴포넌트 |
| 전송-6(확인) | 격리 확인(큰 파일·중단·끊김·권한 없는 경로·같은 이름) | scratchpad 스크립트 |

## 전송-1(절대 경로 수신) 구현 (2026-10-23)
| 위치 | 변경 |
|---|---|
| `protocol` `FileInitRequest` | `Abs`: DestPath = 받을 **폴더**(절대), Name = 그 안 파일 이름. 합치기·이름 검사는 worker(Linux supervisor가 Windows 경로를 못 합치므로) / `Exclusive`: 이미 있으면 거절. 둘 다 기본 false = 기존 선배치 동작 그대로 |
| `protocol` `FileInitResponse` | `Code` = `FileErrExists`·`FileErrIsDir`·`FileErrNoParent`·`FileErrBadPath` |
| worker `router/transfer.go` | `resolveAbsDest`(절대 폴더 + 이름, 이름에 `/`·`\`·`.`·`..`·빈 값이면 거절) · `checkDest`(같은 이름 폴더 = 항상 거절, 파일 = Exclusive일 때, Abs면 폴더가 있어야 함 — `NewSaveFile`의 MkdirAll로 아무 데나 만들지 않게) |
| supervisor `router/transfer.go` | `send(job *sendJob)`·`sendOnce(.., job, ..)`로 정리, `sendJob.dir`(비어 있지 않으면 Abs)·`exclusive` / 거절 = `rejectedError{Code,Reason}`(`errors.Is(err, errRejected)` 유지, 코드는 `errors.As`) |
- 확인: `-overlay` 단위 테스트(scratchpad `xfer1/abs_test.go`) 통과 — 거절 7가지, 합치기, `checkDest` 6가지. 기존 `go test ./cmd/supervisor/router/` 통과
- 알려진 한계: ① Linux에서 이름에 `\`가 든 파일은 올릴 수 없음(Windows와 같은 규칙) ② 받기 끝(rename) 뒤 FileResult 응답만 끊겨 재시도하면 Exclusive가 "이미 있음"으로 거절(실제로는 올라감) ③ `.part`가 받을 폴더에 보임(`이름.part`), 같은 이름의 진짜 `.part` 파일은 지워질 수 있음

## 전송-2(업로드 REST) 세부 제안 (2026-10-23, 채택 — 단 본문은 Q30(조각 업로드)으로 바뀜)
| 항목 | 제안 |
|---|---|
| 겹침 확인 API(**계획에 없던 것**) | `POST /workers/:mainKey/:subKey/dir/exists?folderId` {path, names[]} -> 이미 있는 이름(폴더/파일 구분). worker REQ 새로 하나. 장비 타일 목록은 500개씩 페이지라 화면 목록만으로는 모자람 -> Q28(여러 파일 겹침) 창의 근거 |
| 업로드 요청 | 파일 하나 = 요청 하나, 본문 = 파일 바이트 그대로(multipart 아님). `POST /workers/:mainKey/:subKey/files?folderId&path&name&overwrite&tileId` |
| 서버 흐름 | 인가(`requireDeviceDir`) -> 전송 목록에 등록(단계 = 올리는 중) -> 본문을 임시 파일로(받은 양 push) -> 응답 -> 뒤에서 `send`(단계 = 장비로 보내는 중) -> 끝나면 임시 파일 삭제 |
| 진행률 출처 | 두 단계 다 **서버 push 하나로**(브라우저 XHR 진행률 안 씀) -> 다른 탭·기기도 같은 목록(계정 동기화 원칙) |
| 중단 | 올리는 중 = 브라우저가 요청 취소 -> 서버가 본문 읽기 실패로 "중단" 기록 / 보내는 중 = `AbortFile` |
| 임시 파일 | `os.TempDir()/nexus-upload/`, supervisor 시작 때 비움 |
| 외부 접속 주의 | test 도메인 앞 nginx(다른 장비)의 `client_max_body_size` 기본 1MB -> 큰 파일 413. nginx 설정 확인 필요(`0` + `proxy_request_buffering off`) |

## 전송-2(업로드 REST)·전송-3(전송 목록) 구현 (2026-10-23)
| API | 하는 일 |
|---|---|
| `POST /workers/:m/:s/dir/exists` {folderId, path, names} | worker `DIR_EXISTS` -> 이미 있는 것만 `DirEntry[]`(kind dir/file). 잘못된 이름은 빠짐 |
| `POST /workers/:m/:s/uploads` {folderId, path, name, size, overwrite, tileId} | 인가 `requireDeviceDir` + 이름 검사(`protocol.ValidFileName`, worker와 같은 규칙) -> 항목 등록(upload/running) + 임시 파일 -> `TransferInfo` |
| `PUT /transfers/:id/data?offset=` 본문 = 조각 | 1MB 넘으면 413, offset > 받은 양 409, 작으면 덮어씀(멱등) -> {received} |
| `POST /transfers/:id/finish` | 다 받았을 때만 -> 단계 send(done 0) -> 뒤에서 `send`(Abs, exclusive = !overwrite) |
| `DELETE /transfers/:id` | 중단. send 단계면 `AbortFile`로 worker `.part`도 지움 |
| `GET /transfers` | 내 목록(시작 순) |
| push `TRANSFER:UPDATE`(TransferInfo) · `TRANSFER:GONE`{id} | 계정 토픽. 진행률은 250ms마다, 단계·상태 변화는 바로 |
- 파일: `router/upload.go`(시작·조각·finish·runUpload) · `router/transferList.go`(항목·push·끝내기·목록·중단·멈춘 올리기 정리·임시 폴더) · `router/transfer.go` `sendJob`에 `id`·`ctx`(시작 전 중단도 먹게)·`onProgress` · worker `dirExists`
- 정리 규칙: 끝난 항목 5분 뒤 목록에서 빠짐(`TRANSFER:GONE`) / 조각이 10분 안 오면 중단 / 임시 폴더 = 실행마다 `$TMPDIR/nexus-upload-*`(두 빌드가 서로 안 지움), 하루 지난 옛 폴더는 시작 때 지움 / 결과 분류: 성공 done · ctx 취소 aborted · FileInit 거절 failed+code · 그 밖 failed+reason
- 확인: 격리 26/26(→ `REF-test-scripts-1022.md` "2026-10-23 전송-2+전송-3")

## 전송-4(화면: 올리기)·전송-5(화면: 전송 상태) 구현 (2026-10-23)
| 파일 | 내용 |
|---|---|
| `feature/worker/api/transfer.api.ts` | API 6개 + `UPLOAD_CHUNK` 512KB |
| `feature/worker/store/transfers.store.ts` | `provideTransfers`(`pages/index.vue`). 목록 = `TRANSFER:*` 이벤트 + 연결될 때 `GET /transfers`. 늦게 온 옛 상태는 `newer`(끝남 > 진행, send > upload, done 큰 쪽)로 버림. `upload(src, files)` = 같은 이름 확인 -> Q28 창 -> 파일 하나씩 시작·조각·finish(네트워크 오류만 조각당 3번 다시) / `abort` / `lastDone`(끝난 받을 폴더) |
| `feature/worker/component/TransferRow.vue`·`TransferButton.vue` | 한 줄(이름·단계 %·크기·중단·받을 곳) / 푸터 버튼 + 위로 펼치는 목록 |
| `DeviceTileBody.vue` | 경로 줄 도구에 올리기 버튼(+숨은 file input), OS 파일 끌어 놓기(점선, 폴더는 빼고 안내), 바닥 고정 막대 = 이 타일에서 시작한 진행 중·실패만(끝난 것은 푸터에만), `lastDone`이 지금 폴더면 목록 다시 읽기 |
| `workerDir.hook.ts` | `cur` = worker가 푼 지금 폴더(올리기 대상). Windows 드라이브 목록("")·읽기 실패·끊김이면 올리기 꺼짐 |
| `BoardFooter.vue` | Q31(전송 아이콘 자리). 폰 '더 보기' 점에 진행 중 전송도 |
- 확인: 헤드리스 18/18(→ `REF-test-scripts-1022.md` "2026-10-23 전송-4·5")
- 알려진 한계: ① 여러 파일은 한 번에 하나씩(대기 중인 파일은 목록에 안 보임) ② 탭을 닫으면 올리는 중이던 전송은 10분 뒤 중단 처리(이어서 올리기 없음) ③ 끝난 뒤 다시 읽기는 첫 페이지부터(500개 넘게 펼쳐 둔 목록은 접힘) ④ 푸터 목록의 받을 곳은 뒤가 잘림(전체는 title)

## 남은 질문
- 없음. 마감 범위 = 업로드·전송 상태까지(2026-10-23 사용자 "마감하자", Claude 권유: 내려받기는 worker -> supervisor 송신을 새로 만드는 일이라 파트를 나눔)
