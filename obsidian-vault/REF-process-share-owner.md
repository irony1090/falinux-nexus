# REF — 공유-e-3 소유자 쪽 목록(O1~O3·O5)

**상태(2026-10-08): 3a(공유 목록 창)·3b(행 공유 개수)·3c(내가 공유한 것) 완료, 사용자 직접 확인 통과(커밋 전).**

> 2026-10-08 `REF-process-share-display.md`에서 분할. 결정 SH1~SH10 -> `REF-process-share.md` / 이력 -> `history/process-share-owner.md` / 표시 전반 -> `REF-process-share-display.md` / 현재 -> `CURRENT.md`

## 공유-e-3(소유자 쪽 목록) 구조안 (2026-10-08 확정, 코드 없음)
시안: https://claude.ai/artifact/Cw9WxYvSjnpyY8bmcSStsg ("공유 목록 창 시안", O2·O3·O5 직접 눌러 보기)
```
O1 [공유 N] --click--> ShareDialog step list (신설)    GET /nodes/:id/shares
   |- 직접 건 행   : 단계 선택(PUT) / 해제(DELETE, 상위 공유가 있으면 "이어짐" 행으로)
   |- 이어진 행    : 점선 + "여기서 바꾸기"(차단 포함 단계 선택 -> PUT)
   |- 차단(0) 행   : 빨간 표시, 해제 = 상위 공유로 복귀
   +- [사람 추가] --> step id -> confirm -> done (공유-e-2)
```
| 결정 | 내용 |
|---|---|
| 1 | O1을 누르면 목록이 먼저(시안의 팝오버는 모바일 때문에 창으로) |
| 2 | 차단(0)은 이어진 행의 "여기서 바꾸기"에서만, 새 사람 추가엔 없음 |
| 3 | O3 = 폴더 목록 응답(소유자 REST)에 `shareCount` 추가(직접 건 공유만, 스크립트 행 없음) |
| 4 | O5 = 노드 + process 단독 공유 함께, 해제만(단계 변경은 O2) |
| 5 | 3a 먼저 → 확인 → 3b → 3c 순서(3b·3c는 서버 변경 있음) |
- O4(접속 표시)는 공유-f(presence) — 목록 행의 접속 점 자리만 남겨 둠

## 공유-e-3a 구현 메모·확인 결과 (2026-10-08, 커밋 전, 서버 변경 없음)
| 파일 | 변경 |
|---|---|
| `share.api.ts` | `listNodeShares`·`deleteNodeShare` + `useNodeShares`(키 `['SHARE','NODE',id]`)·`invalidateNodeShares` |
| `ShareDialog.vue` | `list` 단계 신설(첫 화면) — 직접 건 행(단계 `v-select` + 해제) / 이어진 행(점선 + "여기서 바꾸기" → 차단 포함 선택 + 적용) / 차단 행(빨간 표시). 사람 추가는 기존 흐름, 완료 뒤 "목록으로" |
| `FolderTileBody.vue` | O1 버튼에 숫자(이 폴더에 적용되는 사람 수) |
| `shareView.util.ts` | `PERM_BLOCK` |
- 정한 것: 직접 건 행의 단계 선택에는 차단(0)이 **이미 차단인 행에서만** 나온다(풀어 주려고). 새로 막는 건 "해제 → 이어진 행의 여기서 바꾸기"로. 서버는 어디서든 0을 받음
- 확인(격리 실행, 11개 통과): O1 숫자(이어진 것 포함) / 이어진 행 표시 / 여기서 바꾸기 → 차단 / 해제 → 상위 공유로 복귀 + 안내 문구 / 사람 추가 → 목록에 / 단계 변경 / 해제 / O1 숫자 갱신 + 모바일 390px

## 공유-e-3b 구현 메모 (2026-10-08, 커밋 전)
| 파일 | 변경 |
|---|---|
| `query/shares.sql` | `CountDirectShares(node_ids[])` — 직접 건 행만, 차단(0) 제외, `GROUP BY node_id` |
| `nodeDto.go` | `shareCount`(omitempty) + `fillShareCounts`(내 소유 폴더만 채움, 쿼리 1회) |
| `node.go` | `listChildren` 루트·자식 두 갈래 모두 `fillShareCounts` 호출 |
| `node.api.ts` / `share.api.ts` / `FolderTileBody.vue` | 타입 `shareCount` / `invalidateNodeShares`가 `['NODE','LIST']`도 무효화 / 행 칩(`share` 색, 0이면 숨김) |
- 정한 것: 칩은 직접 건 공유만(O1 숫자는 이어진 것 포함이라 기준이 다름), 차단 행 제외

## 공유-e-3c 구현 메모 (2026-10-08, 커밋 전)
| 파일 | 변경 |
|---|---|
| `query/shares.sql` | `ListMyNodeShares`(내 노드의 직접 공유, 차단 포함) · `ListMyProcessShares`(내 살아 있는 process 단독 공유, 이름 = 스크립트 이름 → 없으면 cmd) |
| `router/share.go` | `GET /shares/mine` → `myShare[]`(`kind` NODE/PROCESS, `nodeId`/`processUid`, `name`, `userId`, `nickname`, `permission`). 해제는 기존 DELETE 두 개 재사용 |
| `share.api.ts` | `listMyShares`·`deleteProcessShare`·`useMyShares`·`invalidateMyShares`, `invalidateNodeShares`가 `['SHARE','MINE']`도 무효화 |
| `dialog/MySharesDialog.vue` (신설) | 대상별 묶음 + 사람 행(단계·해제). 단계 변경 없음 |
| `BoardMenu.vue` | 구분선 아래 "내가 공유한 것" 항목 |
- 확인할 것(공유-e 직접 확인 때): 노드·process 공유가 한 창에 묶여 나옴 / 해제 후 사라짐 + 받는 쪽 타일 사라짐 / 끝난 process는 목록에서 빠짐 / 모바일 390px
