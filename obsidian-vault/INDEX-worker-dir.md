# INDEX — worker 폴더 타일 (마감)

> 2026-10-23 파트(worker 폴더 타일) 마감 때 MEMORY·HISTORY에서 접음. **세션 시작 시 읽지 않는다.** REF/history 파일은 옮기지 않았다(링크 보존). 전체 계획 `REF-worker-dir.md`는 다음 파트(파일 전송 확장) 계획도 담고 있어 MEMORY에 남김. → `MEMORY.md` / `HISTORY.md` / `CURRENT.md`

## 파트 요약
세 번째 타일 종류 = **장비 타일**(`device`): worker의 실제 디렉토리를 폴더 타일처럼 보여 줌. Linux·Windows worker 둘 다.
- 폴더-1(목록 요청): `GET /workers/:main/:sub/dir`, worker가 페이지 단위로 읽음(보관 목록·410), 링크·junction 표시
- 폴더-2(장비 타일)·3(여는 곳): 폴더 타일 경로 줄의 장비 디렉토리 열기 버튼 -> 장비 타일. 인가 = 연 폴더 소유자 + 적용 device_key == 메인키
- 뒤로-1(타일 안 이동 기록): 폴더·장비 타일 공통 뒤로·앞으로(타일 트리 저장, 20개)
- 폴더-4(실행): 실행 가능한 파일만 "실행" -> 새 터미널 타일, 마이그레이션 00012 `dir_node_id`, 형식 변환 없는 직접 실행
- 폴더-6(보드 이동)·Q9(루트 장비 탭): 장비 타일 새 보드로 분리·다른 보드로 이동, 루트면 `[파일] [터미널 N]`(그 장비의 볼 수 있는 process 전부, `GET /workers/:main/runs`), 연 폴더 삭제 = "볼 수 없음"
- 확인: 격리 REST·헤드리스·Windows VM 각 단계 + 사용자 브라우저 확인 OK(test3 `build/worker-dir` + VM worker `win-dev`, 2026-10-23)
- 커밋: 브랜치 `feat/worker-dir`(`refactor/tile-kinds` 위) -> 2026-10-23 master `--no-ff` 머지. 같은 커밋에 `scripts/build.sh` 태그 출력(→ `REF-build.md` "빌드 태그")

## REF / history 목록 (MEMORY 인덱스에서 옮김)
- `REF-worker-dir-list.md`(폴더-1(목록 요청)·2(장비 타일)·3(여는 곳)·뒤로-1(타일 안 이동 기록)) / `history/worker-dir-list.md`
- `REF-worker-dir-exec.md`(폴더-4(실행)) / `history/worker-dir-exec.md`
- `REF-worker-dir-board.md`(폴더-6(보드 이동)·Q9(루트 장비 탭)) / `history/worker-dir-board.md`
- (MEMORY에 남김) `REF-worker-dir.md`(전체 계획 + 파트(파일 전송 확장) 폴더-5·7·8·9) / `history/worker-dir.md`

## HISTORY 색인 행 (원문)
| `history/worker-dir-list.md` | 폴더-1(목록 요청)·2(장비 타일)·3(여는 곳)·뒤로-1(타일 안 이동 기록)(2026-10-22 분할) | 10-22 밤 폴더-1·2·3·뒤로-1 구현, 격리 18/18·헤드리스 22/22 / 10-23 Windows VM 13/13·9/9(junction 고침), 미커밋 |
| `history/worker-dir-board.md` | 폴더-6(보드 이동)·Q9(루트 장비 탭)(2026-10-23 신설) | 10-23 Q25·Q26 = 가 -> 이동-1~4 구현, 격리 REST 20/20·헤드리스 24/24 -> 사용자 브라우저 확인 OK(Linux·Windows), 미커밋 |
| `history/worker-dir-exec.md` | 폴더-4(실행)(2026-10-23 분할) | 10-23 설계·Q23(실행 버튼)·Q24(실행 가능한 파일만) 확정 -> 구현(마이그레이션 00012·형식 변환 없는 직접 실행·flush), 격리 31/31·헤드리스 13/13·Windows 11/11, 미커밋 |

## 이 파트에만 해당하는 결정 · 남은 것
- 장비 디렉토리 인가(연 폴더 소유자 + device_key == 메인키, `requireDeviceFolder`)는 다음 파트(업로드·내려받기)도 같은 규칙을 쓸 것 → `REF-worker-dir-board.md` "구현", `REF-worker-dir-list.md`
- 못 본 것: 504(worker 무응답) · 여러 인스턴스 고르기 창 · 화면에서 상위 폴더를 지울 때 루트 장비 분기 · 모바일 화면 → 각 REF "확인"
- 다음 파트(파일 전송 확장) 폴더-8(전송 상태)·7(내려받기)·5(업로드)·9(worker 간 전송) + SFTP 아이디어 → `REF-worker-dir.md`, `BACKLOG.md`
