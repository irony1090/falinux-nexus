# HISTORY — node 카탈로그 타일 UI 실제 연동

> 더미 → node/process API 연동 작업 이력. 설계·결정 → `REF-node-ui-link.md` / 현재 진행 → `CURRENT.md`. 연동 1차 1번(`GET /workers`)은 백엔드 작업이라 `history/node-label.md`.

---

### 2026-10-01 - 스크립트 편집 창 저장 동작 통일 결정
- 사용자: 저장 버튼(저장 후 닫기)과 Ctrl+S(저장하고 계속 편집)를 **"저장하고 계속 편집"으로** 맞춤. 코드 미반영(`ScriptEditDialog.vue` 저장 버튼에서 닫기 제거만 하면 됨).

### 2026-09-30(4) - 연동 1차 4(스크립트 편집): 브라우저 편집 창
- 구조안 제시 → 사용자: G(편집기 위치) 추천대로 / H(편집기 종류)는 "나중에 vi가 붙는다는 거지?" 확인(브라우저 편집기와 EDIT는 별도 경로, EDIT는 6 + input 배선 뒤 — 백엔드 EDIT 처리 코드는 supervisor·worker에 이미 있음) 후 추천대로 / I(새 스크립트 직후 동작)는 뜻을 물은 뒤 추천대로.
- **작성**: `dialog/ScriptEditDialog.vue` 신규 / `NodeRowMenu.vue`("내용 편집", 스크립트만) / `FolderTileBody.vue`(창 연결, 새 스크립트 뒤 자동 열기).
- **구조안에 없던 세부**: 저장 버튼 = 저장 후 닫기, Ctrl+S = 저장하고 계속 편집.
- **검증**: vue-tsc + 헤드리스 PC·폰 시나리오 통과(→ `REF-node-ui-link.md`). 처음엔 PC 창 높이가 226px로 나와 `VDialog` `height`로 수정. 임시 노드는 삭제(사용자가 만든 `test`/`ttt1`은 그대로 둠).

### 2026-09-30(3) - 폴더 안에서 장비 지정·접속 상태 표시 + 헤드리스 확인
- 사용자: 3(node 관리 UI) 동작 확인. 추가 요청 "장비 지정은 해당 디렉토리 아이템 내부에서도 가능해야 하고, 현재 접속 상태도 알 수 있어야".
- **작성**: `FolderTileBody.vue` 브레드크럼 줄에 지금 보는 폴더의 장비 표시 버튼(자기 키 또는 상속 키 + 접속 점) → 누르면 `NodeDeviceDialog`. 목록 행 장비 표시는 키가 길면 말줄임되도록 구조 수정. 장비 창 제목을 `장비 지정: 이름`으로(폰에서 앞부분이 잘리지 않게).
- **로그인 계정**: 사용자가 "계정 정보는 `Login.vue`에 있고 직접 살펴봐도 된다"고 허용 → 로그인 화면 기본값으로 헤드리스 로그인해 확인(이전엔 비밀번호를 매번 물었음).
- **검증**: vue-tsc + 헤드리스 시나리오(→ `REF-node-ui-link.md` "확인") 통과, 임시 노드는 시나리오 안에서 삭제.

### 2026-09-30(2) - 연동 1차 3(node 관리 UI): 생성·이름 변경·장비 지정·삭제
- 2(폴더 타일 목록)는 사용자가 브라우저에서 확인("확인했어"). 이어서 3 구조안 제시 → 사용자 결정: D(조작 메뉴 형태) = 추천대로 `⋮` 메뉴 / E(삭제된 노드를 보던 타일) = "하위 폴더는 전부 닫아버려"(추천은 오류 표시 유지였음) / F(장비 후보 범위) = "직접 입력도 있어야 돼, worker보다 먼저 추가해서 상태를 확인"(추천은 직접 입력 없음이었음).
- **작성**: `NodeRowMenu.vue`·`NodeNameDialog.vue`·`NodeDeviceDialog.vue`·`hook/nodeRemove.hook.ts` 신규 / `node.api.ts`(`useNodeActions`, `fetchPath`) / `worker.api.ts`(`useOnlineWorkers`, 10초 주기) / `FolderTileBody.vue`(`+` 메뉴, 행 `⋮`, 장비 표시, 창 3종 조립).
- **구조안과 달라진 것**: `hook/nodeRemove.hook.ts` 추가(구조안 파일 목록엔 없었음 — 삭제가 API와 타일 트리를 함께 다뤄서). 장비 선택은 콤보박스 대신 입력칸 + 접속 중 장비 칩.
- **검증**: vue-tsc 통과, dev 서버 컴파일 정상. 브라우저 동작은 미확인(로그인 세션 없음).

### 2026-09-30 - 연동 1차 2(폴더 타일 목록): 더미 목록·경로·이름을 node API로
- 구조안 제시(더미 3종 교체, 새 스토어 없음) → 사용자 결정: A(경로 조회 방식) = 백엔드 그대로 두고 `getNode` 반복 호출 / C(실행 버튼 범위) = 추천대로 더미 유지 / B(더미 시드 처리)는 "그게 뭐냐"는 질문 → 설명 후 답 대기.
- **작성**: `node.api.ts`(`Q_KEY.PATH`, `fetchNodePath`, `useNodePath`, `invalidateAll`에 경로 무효화) / `FolderTileBody.vue`(`useListChildren` + `useNodePath`, 새로고침 버튼, 진행 표시줄, 오류·다시 시도) / `TileFrame.vue`(`useGetNode` 이름) / `TerminalTileBody.vue`(가짜 출력 제거) / `tileDummy.ts`(가짜 노드 삭제). 이어서 B(더미 시드 처리) = 제거로 결정(사용자 "자동 열기가 지금 필요해??") → `index.vue` 호출 + `tileDummy.ts`의 `seedTileTree` 삭제.
- **계획과 달라진 것**: 로딩 표시를 `Skeleton` 위젯 대신 `VProgressLinear`로 — `SkeletonGroup`의 clipPath id 고정(`CLIP_PATH`)으로 다중 인스턴스 충돌.
- **검증**: vue-tsc 통과, dev 서버(3000)가 바뀐 컴포넌트를 정상 컴파일. 브라우저는 로그인 세션이 없어 사용자가 직접 확인("확인했어").
- 같은 날 선행: 테스트 스크립트 id 2~4를 폴더 id 1로 이동 + `device_key` 지정(→ `history/node-label.md`). 실제 목록 = 루트에 폴더 `HTOP_TEST_SH_MODI` 1개, 그 안에 스크립트 3개.
