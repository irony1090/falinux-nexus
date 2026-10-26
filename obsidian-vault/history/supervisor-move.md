# history — supervisor 주소 실시간 변경

> → REF `REF-supervisor-move.md` · → 진행 `CURRENT.md`

## 2026-10-24 착수·조율
- BACKLOG "supervisor 주소 실시간 변경"을 CURRENT로 올림
- 조사: worker 주소 = env를 시작 때 한 번 읽어 고정 / 서브키는 worker sqlite, supervisor는 검사 안 함 / 재접속 때 DB에 없는 uid는 무시 / worker kill은 uid만으로 가능 / users에 역할 없음
- 조율 흐름: Q1(바꾸는 주체) supervisor 명령 -> Q4(용도) 이사·업데이트·임시 -> 사용자가 주소 목록 모델 제시(옛+새 저장, 끊기면 돌며 시도, 붙은 주소 1순위) -> Q9(즉시 옮김) 선택 · Q10 새 주소 1순위 · Q11·Q13 최대 3 · Q6 데이터 다를 수 있음
- Q12(새 쪽이 모르는 process): 입양 제안 -> 너무 큼 -> 사용자가 자동 종료 제안 -> 임시 이동 때 전부 꺼지는 위험 짚음 -> 사용자가 "유지 + 화면에서 끄기"로 정함. 이름 = 미등록 process(Q21)
- 권한: Q8 특정 계정 -> Q16 "지금은 필요 없음" -> Q20 전체 옮기기는 특정 계정 필요 -> `users.perms TEXT[]`(Q22) + DB 직접 UPDATE(Q23)
- 작업 단위 주소-1(주소 목록)~주소-6(확인) 계획 작성 → REF "작업 단위". 진행 신호 대기, 코드 손대지 않음

## 2026-10-24 구현·격리 확인
- "진행" 받고 브랜치 `feat/supervisor-move`. 주소-1(주소 목록)~주소-5(화면) 구현 → REF "구현"
- 확인 중 발견: exec는 `X-Tab-Id` 필요(확인 스크립트에서 탭 소켓 엶) · 스크립트 노드 명령은 임시 파일 경로 · Vuetify 목록 role 함정
- 격리 확인 서버 28/29(T6b 기대값만) + 화면 8/8, 격리 환경 정리(5051·5052·3001·DB 둘·`mv-dev` 보관 소켓). 사용자 확인 대기
- 사용자 "스크립트 이름 표시 넣어줘" -> `ProcessSpec.Label` + holds `label`(00004에 합침, 아직 어디에도 적용 안 됨) -> 격리 다시 서버 29/29·화면 8/8, 정리

## 2026-10-26
- 사용자 확인: 회사 Linux B 빌드(`build/supervisor_move`)를 test3에 띄운 뒤 화면으로 집 서버로 옮기기 **잘 됨** -> "맘 편히 머지해도 될 것 같다". 머지는 회원 관리 확인 뒤(`CURRENT.md` 순서). 마감 커밋 + `INDEX-supervisor-move.md`로 접기는 머지 때
