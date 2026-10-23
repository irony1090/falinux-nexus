# REF — Windows VM (`nexus-win`)

> 2026-10-23 `REF-windows-worker.md`에서 분할(10k자 기준, "VM 실행"·"VM 준비" 절을 통째로). 이력 → `history/windows-vm.md` / Windows worker 계획·결정 → `REF-windows-worker.md`(`INDEX-windows-worker`) / 확인 스크립트 → `REF-test-scripts-1014.md` "윈도-6"·`REF-test-scripts-1023.md` "Windows(VM)" / → `CURRENT.md`

## VM 실행 (2026-10-19 사용자 승인 → 띄움)
| 항목 | 값 |
|---|---|
| 컨테이너 | `nexus-win`(`dockurr/windows`, v6.06), `--restart no`(장비 재부팅 때 자동 시작 안 함 → `docker start nexus-win`) |
| Windows | `VERSION=11`(Pro) · `LANGUAGE=Korean` · `REGION=ko-KR`(CP949·PS 5.1 BOM 문제를 실제로 보려고 한국어) · `KEYBOARD=en-US` / 계정 기본값 `Docker` / `admin` |
| 자원 | `RAM_SIZE=6G` · `CPU_CORES=4` · `DISK_SIZE=64G` |
| 폴더 | `/home/jh-bae/vm/nexus-win/` 아래 `storage`(VM 디스크, 저장소 밖) · `shared`(-> Windows 바탕화면 Shared·`Z:`, 교차 빌드 `worker.exe` 둠) · `oem/install.bat`(설치 끝에 OpenSSH 서버 켬, 로그 `C:\OEM\ssh-setup.log`) |
| 포트(전부 127.0.0.1) | 8006 웹 화면 · 3389 RDP · 2222 -> VM 22(SSH) |
| VM -> 이 장비 | docker0 `172.17.0.1`(supervisor 8012는 0.0.0.0, 격리 supervisor는 IP 바인드 주의) |
- 마감 후(2026-10-19): VM은 남겨 둠. worker `win-dev`는 사용자 supervisor 8012에 연결(env `WS_HOST=172.17.0.1:8012`·`SQLITE=./worker-8012.db`, 옛 `worker.db` = 격리 5051용 서브키). 안 쓸 때 `docker stop nexus-win`(RAM 6G). **2026-10-23~ 규칙: 시험할 때만 켜고 작업 중엔 끈다**(MEMORY 로컬 개발 환경)
- 2026-10-23 test3(8013)에 붙임: 교차 빌드 `shared/worker-test3.exe`(태그 `WORKERDIR-261023`) + `shared/start-test3.ps1`(env `C:\nexus\env-test3` = `WS_HOST=172.17.0.1:8013`·`SQLITE=./worker-8012.db`·`NAME=win-dev`, 로그 `C:\nexus\worker-test3.log`). 서브키 `6RFUUw2P` 그대로 이어짐(`nexus_build` = `nexus` 복사본). 실행 = vssh로 `powershell -File \\host.lan\Data\start-test3.ps1`. 사용자 확인 뒤 VM 끔(2026-10-23)
- 2026-10-23 파트(경로 직접 입력): VM 켜서 격리 확인(`worker-path.exe`·`start-path.ps1`, NAME `dir-win`) 뒤 `win-dev`를 test3에 `worker-test3.exe`(태그 `DIRECTORYPATH-261023`)로 다시 붙임 → `REF-test-scripts-1023.md` "Windows(VM)"
- 지우기: `docker rm -f nexus-win` + `/home/jh-bae/vm/nexus-win` 삭제(이미지는 `docker rmi dockurr/windows`)

## VM 준비 (2026-10-15 조사)
| 확인 | 결과 |
|---|---|
| 가상화 | CPU 플래그 있음, `/dev/kvm` = `jh-bae` 읽기·쓰기(ACL) |
| 자원 | 메모리 31GB(가용 16GB) · 18코어 · 디스크 여유 364GB(937GB 중), docker 사용 약 26GB(회수 가능 약 7.5GB) |
| 설치된 도구 | qemu·virsh·virt-manager·VirtualBox·vagrant 없음 / docker 있음 |

| 방법 | 내용 |
|---|---|
| **가 `dockurr/windows`(추천)** | docker 안 QEMU/KVM. ISO 자동 다운로드·무인 설치, 화면 = `localhost:8006`(웹)·RDP. sudo 불필요. 설치 후 스크립트로 OpenSSH 켜면 Claude가 ssh로 worker 실행·확인 가능. 공유 폴더로 교차 빌드한 `worker.exe` 전달 |
| 나 qemu/virt-manager | `sudo apt install` + ISO 수동 설치 |

- 비용: 다운로드 약 6GB + VM 디스크 기본 64GB(실사용만큼 커짐), 메모리 4~8GB, 설치 30분 이상
- 접속: VM worker -> 이 장비 supervisor(8012는 0.0.0.0 바인드) 또는 격리 supervisor(임시 DB + 5051)
- 주의: 정품 인증 없는 Windows(테스트용, 워터마크 등) 또는 Enterprise Evaluation 90일. `dockurr/windows` 세부 옵션(버전 이름·호스트 주소)은 기억 기준 → 띄울 때 문서 확인
- 띄울 때 사용자 승인 필요(이미지 pull·다운로드·디스크 사용)

