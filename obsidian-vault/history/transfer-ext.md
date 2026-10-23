# history — 파일 전송 확장

> → `REF-transfer-ext.md` / 이전 계획 → `history/worker-dir.md` / → `CURRENT.md`

## 2026-10-23 착수·조율 (코드 없음)
- 세션 시작 = 파트(파일 전송 확장) 착수. 순서·SFTP·브랜치 조율 -> Q20(진행 순서) 폴더-5(업로드)+폴더-8(전송 상태) 먼저 / Q21(SFTP) 작업이 전혀 없을 때 / Q22(브랜치) `feat/transfer-ext` 생성(master `3b59c61`)
- 큰 계획 전송-1(절대 경로 수신)~전송-6(확인) 제안 -> Q23(같은 이름 파일) 묻기 · Q24(전송 상태 위치) 헤더+타일 둘 다 · Q25(진행률 구간) 한 줄 단계 표시 · Q26(폴더째 올리기) 나중에
- 세부 -> Q27(타일 막대 범위) 그 타일에서 시작한 것 · Q28(여러 파일 겹침) 한 창에 모아서 · Q29(전송 목록 보관) 메모리만
- 조사: worker 수신 `resolveDest`가 `ProcessRoot` 밖 거절 / 업로드 REST·multipart 없음 / 장비 디렉토리 인가 `requireDeviceDir`(`workerDir.go`) / 계정 토픽 `userTopic`(`subscribe.go`)

## 2026-10-23 전송-1(절대 경로 수신) 구현
- 사용자 "진행해줘" -> 전송-1 구현. 처음엔 Abs = DestPath 전체 절대 경로였으나, Linux supervisor가 Windows 폴더+이름을 못 합쳐서 **폴더(DestPath)+이름(Name) 따로, worker가 합침**으로 바꿈
- `-overlay` 단위 테스트 통과(→ `REF-transfer-ext.md` "전송-1 구현"). 커밋 안 함(파트 단위)
- 전송-2(업로드 REST) 세부 제안 -> 겹침 확인 API(계획에 없던 것)·nginx 본문 크기 확인 대기

## 2026-10-23 전송-2(업로드 REST)·전송-3(전송 목록) 서버
- nginx 본문 1MB 문제 -> 사용자가 웹소켓 업로드를 물음 -> 비교 뒤 Claude가 처음 추천(웹소켓)을 정정, **Q30(조각 업로드) = HTTP 조각 512KB**로 결정. 같은 이름 확인 API 추가 승인
- 서버 구현(API 7개·push 2개) + 격리 확인 26/26. 격리 환경(supervisor 5051·`nexus_xferext_tmp`·worker `xe-dev`)은 화면 확인용으로 띄워 둠

## 2026-10-23 전송-4(화면: 올리기)·전송-5(화면: 전송 상태)
- 사용자 "그 모양으로 진행" -> 앱에 헤더가 없다는 것을 확인하고 Q31(전송 아이콘 자리) = 푸터 오른쪽 종 옆으로 다시 물어 정함
- 구현(→ `REF-transfer-ext.md` "전송-4·5 구현") + 헤드리스 18/18. 처음 실패 2건은 테스트 쪽(127.0.0.1 쿠키, Vue 클래스 반영 타이밍). 스크린샷 보고 "올리는 중26%" 띄어쓰기 고침

## 2026-10-23 파트 마감
- 사용자 "마감하자" -> 범위 = 업로드·전송 상태까지, 폴더-7(내려받기)·폴더-9(worker 간 전송)은 다음 파트 후보(BACKLOG). 격리 프로세스(5051·3001·xe-dev) 종료, `nexus_xferext_tmp` DROP, 빈 보관 소켓 폴더 `xe-dev` 삭제
- `INDEX-transfer-ext.md`로 접음, 브랜치 `feat/transfer-ext`에 커밋. master 머지는 사용자 브라우저 확인 뒤
