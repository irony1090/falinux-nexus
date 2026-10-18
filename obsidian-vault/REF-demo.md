# REF — 시연 데이터 (원격 생산 지그)

> → 이력 `history/demo.md` / → 현재 `CURRENT.md`
> 사람들에게 "원격으로 기기 동작 테스트 + 계정별 스크립트 관리·공유"를 보여주기 위한 데이터. 2026-10-16 등록, 코드 변경 없음(REST로 데이터만).

## 결정
| 항목 | 결정 |
|---|---|
| 시나리오 | 안 C(생산 지그 시나리오) — 사용자 선택. 안 A(실제 하드웨어 점검)의 실측 + 진행 막대·판정 연출. 시뮬레이션 값은 출력에 "(모의)"로 표시 |
| 강조점 | 계정별 스크립트 관리 + 공유(보기·조작·편집 단계 차이) |
| 카메라 | 실제 촬영 허용 — 사진을 터미널에 컬러 블록(`▀` + 24bit 글자색·배경색)으로 출력 |
| 등록 방법 | DB 직접 INSERT가 아니라 REST(`POST /nodes`, `PUT /nodes/:id/shares/:userId`). 이유: 공유 시 받는 쪽 보드에 폴더 타일이 붙는 처리(`attachSharedFolder`)와 검증·실시간 발행을 함께 거침 |

## 계정 · 공유
| 계정(로그인 id / 비밀번호) | users.id | 역할 | 단계 |
|---|---|---|---|
| 아이로니 (`irony` / `!Fa1289`) | 3 | 시연자, 소유자 | 40 소유자 |
| 테스트1 (`test1` / `test1234`) | 65 | 편집 시연 | 30 편집 |
| 테스트2 (`test2` / `test1234`) | 66 | 조작 시연(실행·입력 O, 수정 X) | 20 조작 |
| 테스트3 (`test3` / `test1234`) | 67 | 보기 시연(입력 X) | 10 보기 |

- `users.identification`은 sha256 해시로 저장됨 → DB에서 계정을 찾을 땐 `printf %s test1 | sha256sum`으로 비교

## 노드 구성 (owner = irony)
| id | 종류 | 이름 | 내용 |
|---|---|---|---|
| 200 | FOLDER | `시연-생산지그` | `device_key = irony-MAC-ADDress1`(이 노트북 worker). 공유는 이 폴더 하나에만 |
| 201 | SCRIPT | `00 시연 안내` | 폴더 구성 + 계정별 단계 출력 |
| 202 | FOLDER | `01 검사 시퀀스` | |
| 203 | SCRIPT | `DUT 검사 - 정상` | 전원 -> 통신 -> 버스 -> 센서 -> 카메라 -> 판정, 진행 막대, PASS 상자 |
| 204 | SCRIPT | `DUT 검사 - 불량` | 203과 같은 본문, 첫 줄 `FAIL_AT=4` → 센서 단계 FAIL(가속도 센서(모의)) 후 중단, exit 1 |
| 205 | FOLDER | `02 실시간 모니터` | |
| 206 | SCRIPT | `라인 모니터` | 1초 갱신 대시보드(`\e[H` 덮어쓰기). 실측 CPU·최고 온도·메모리·기본 경로 NIC 송수신 + 생산 카운터(모의). 끝나지 않음 → process 공유·kill 시연 |
| 207 | FOLDER | `03 대화형 점검` | |
| 208 | SCRIPT | `점검 메뉴` | 번호 입력 메뉴(온도·네트워크·USB·카메라(y/n 후 촬영)·전체) |
| 209 | FOLDER | `04 카메라` | |
| 210 | SCRIPT | `카메라 스냅샷` | `ffmpeg -f v4l2`로 11번째 프레임 1장 → rgb24 → `od` + `awk`로 블록 출력. 폭 = `tput cols`(최대 100) |

## 스크립트 작성 규칙 (재작성·수정 시)
- **본문의 원본 = DB `nodes.content`**. 작성용 파일(scratchpad `demo/*.sh`, `seed.py`)은 재부팅 시 사라짐. 꺼내기: `docker exec postgres15 psql -U irony -d nexus -At -c "select content from nodes where id=203"`
- 외부 의존 최소: bash + awk(mawk) + coreutils + `ip`·`ping` + `ffmpeg`·`v4l2-ctl`. **python 안 씀**(worker 환경의 pyenv shim을 믿지 않음)
- 한글 라벨 정렬: `export LC_ALL=C.UTF-8` + `pad()`(한글 1자 = 2칸으로 계산). `printf %-Ns`는 한글 폭을 1칸으로 세서 줄이 어긋남
- 실측 기준에 여유: 최고 온도 기준 95 °C(확인 중 `x86_pkg_temp` 82 °C까지 나옴 — 85였으면 시연 중 정상 스크립트가 FAIL 날 뻔). 모니터 경고 90 °C
- 끝나지 않는 스크립트: `sleep 1 & wait $!` + `trap cleanup INT TERM HUP`(포그라운드 `sleep`이면 TERM 때 bash가 "종료됨"을 찍고 트랩이 늦음). 커서 숨김(`\e[?25l`)은 cleanup에서 복구

## 시연 순서 (제안)
1. irony: 자기 트리 + `00 시연 안내` 실행
2. irony: `라인 모니터` 실행 → test3(보기) 화면에서 같은 터미널이 실시간으로 보이고 입력은 막힘
3. test2(조작): `DUT 검사 - 정상`·`불량` 실행 가능, 편집은 막힘
4. test1(편집): 스크립트 수정 가능
5. `점검 메뉴` → 4번 카메라 → y → 원격 카메라 사진

## 정리
- 지우기 = 폴더 200 삭제(하위·공유 함께). 다시 넣기 = `REF-demo.md` 표대로 REST 재등록(본문은 지우기 전에 DB에서 꺼내둘 것)
