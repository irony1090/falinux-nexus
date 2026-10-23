# history — worker 폴더 타일: 폴더-1(목록 요청)·뒤로-1(타일 안 이동 기록)

> → `REF-worker-dir-list.md` / 앞 이력 → `history/worker-dir.md` / → `CURRENT.md`

## 2026-10-22(3) 폴더-1(목록 요청) 세부 설계 제시
- 사용자 "2파트 진행" -> `refactor/tile-kinds`(`79b292b`)에서 브랜치 `feat/worker-dir` 생성
- 폴더-1 세부 설계(REST·인가·DIR_LIST·경로 규칙·실패 코드) + Q12(상속 장비 폴더)·Q13(큰 폴더 상한)·Q14(링크 폴더) 제시 → `REF-worker-dir.md` "폴더-1(목록 요청) 세부 설계". 코드 없음, 확인 대기
- 사용자 답: Q12 하위 폴더도 허용 / Q13 페이지로 / Q14 링크 구분 + 원본 위치 -> 페이지 방식 안 가·나·다 제시, Q15(장비 디렉토리 id)·Q14 "본다"의 뜻 되물음
- 사용자 답 2: Q15 연 곳 기준 / Q14 원본 위치로 이동 포함 / Q13 worker 페이지 관리 효율 질문 -> 비교 답함
- 사용자 답 3: Q13 안 다(worker 캐시) / 추가 요청 뒤로가기(장비·폴더 타일) -> 뒤로-1(타일 안 이동 기록) 제안, Q16~Q18 되물음
- 사용자 답 4: Q16 타일 트리 저장·개수 제한 / Q17 앞으로도 / Q18 단축키도 -> Q19(브라우저 뒤로와 겹침) 되물음
- 사용자 답 5: Q19 키 = Ctrl+Alt+←/→ -> GNOME 작업 공간 전환과 겹침 확인, 되물음
- 사용자 답 6: Q19 키 = Ctrl+Shift+←/→ (충돌 없음 확인)
- Q12~Q19 결정 정리 → `REF-worker-dir-list.md`

## 2026-10-22(4) 폴더-1(목록 요청) 구현
- 사용자 "진행해줘" -> protocol `DIR_LIST`, worker `dir.go`·`dir_other.go`·`dir_windows.go` + `workerRouter` 등록, supervisor `workerDir.go` + `mountWorkers` 한 줄
- `go build`·`go vet`(Linux + `GOOS=windows` worker)·router·transport 테스트 통과
- 격리 실행(5051·`nexus_dir_tmp`·worker `dir-dev`) REST 확인 18/18(첫 실행 1건은 기대값 오타) -> 프로세스 종료·DB 삭제. Windows·504 미확인
- 미커밋(파트 단위 커밋)
- 사용자 요청: 인가 직후 트랜잭션 닫기 -> `tx.go` `EndTx` 추가, `workerDir.go`에서 호출. 격리 재확인 18/18
- 2026-10-22(5) 폴더-2(장비 타일) 세부 설계 제시 + Q20(폴더-3 같이)·Q21(경로 동기화)·Q22(링크 원본이 파일) — 확인 대기
- Q20 같이·Q21·Q22 승인 -> 폴더-3(여는 곳) 세부 설계 제시, 진행 확인 대기

## 2026-10-22(6) 폴더-2(장비 타일)·폴더-3(여는 곳) 구현
- Q20(폴더-3 같이)·Q21(경로 동기화)·Q22(링크 원본이 파일) 승인 -> 폴더-3 설계 제시 -> "진행해줘"
- 서버 `tileDevice`·필드·검사·테스트, worker `crumbs`, 프론트 store(장비 타일·이동 기록)·컴포넌트 3개·hook 2개·util·폴더-3 버튼·인스턴스 고르기
- 격리 헤드리스 첫 실행에서 버그 발견: 목록 실패 때 직전 폴더의 경로 줄·상위가 남아 "상위로"가 한 단계 더 올라감 -> 요청 경로 기준으로 고침, 시험도 정확 일치로 바꿈 -> 22/22
- 미커밋
- 2026-10-23 Windows VM 확인: REST 13/13·화면 9/9. junction 두 문제(`other`로 분류·원본이 자기 자신) 찾아 `dir.go` 고침, Linux 회귀 18/18. VM 켜 둠
- 폴더-4(실행) 이력 → `history/worker-dir-exec.md`
