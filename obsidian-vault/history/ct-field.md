# history — 파트(찰떡 입력칸)
> → `REF-ct-field.md` · → `CURRENT.md`

## 2026-10-27
- 착수: 파트(찰떡 마무리) 마감·머지(`d5b4963`) 뒤 브랜치 `feat/ct-field`. 범위 = input·select 전부 + 체크박스 + API 키 보기 칸, 뺌 = 경로 직접 입력칸·숨은 파일 선택
- 진행 중 추가 요청 3개: 방식 묶음 위치 · 방식·호스트 칸 붙이기 · 관리 타일 포커스 선 잘림(→ REF "추가 항목")
- F1(조각): CtTextField + ct-field.scss + 임시 실험 페이지. 끝 여백 = 반지름보다 안쪽(sm 20·md 24) · 32px 글자 쏠림 = v-input flex/1fr 원인(디버거 붙으면 재현 안 됨 -> 화면에 크기 찍기로 잡음) · 메시지 줄 높이·묶음 메시지 · 라벨 = 선 끊기(안 A, outlined)
- 관리 타일 포커스 선 잘림: `.strip` 4px 넓힘(위·왼쪽)·아래 4px 비움 + viewW 보정
- F2(나머지 컴포넌트): CtSelect·CtAutocomplete·CtTextarea·CtCheckbox + 공용 util. 펼침 목록 사방 8px
- F3(옮기기): 창 5개 -> 로그인·관리 페이지. 방식 묶음 오른쪽(두 창), 공유 대상 = 글자 칸, autocomplete = 장비 주인만
- F4(정리): member.scss 안 쓰는 스타일·실험 페이지 지움, 빌드 통과. 마감 확인 대기
- 마감(사용자 "괜찮아" -> "격리 환경 닫고, vault 정리부터 하고 커밋·머지"): 격리 정리(5051·3001·worker `dp-dev` kill, 보관 폴더 삭제, DB DROP) -> `INDEX-ct-field` 접음 · MEMORY/`REF-decisions` "입력칸" 불변 결정 추가 · `REF-design-chaltteok` 쓰는 곳 표에 입력칸 -> 커밋·머지
