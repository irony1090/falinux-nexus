# history — 집 서버 직통 접속 (iptime + Caddy)

> → `REF-home-direct.md` / 그 전 경위(터널 속도·대안 비교·회사 준비) `history/home-server.md` 10-28 절 / `CURRENT.md`

## 2026-10-28 밤 — 직통 전환 (집, 사용자)
- iptime: 포트포워딩 `80`·`https` -> `192.168.0.2`(서버 PC 확인), 기존 `front` 5173~5174 비활성화(인터넷 노출 지적)
- 6(시험 실행): 처음 `sc query caddy`가 아무 출력 없음 = PowerShell `sc` 별칭(Set-Content) -> `sc.exe start caddy` -> `direct` 인증서 tls-alpn-01로 발급. 접속 OK·trace 안 열림·`start= auto`. 타자 반응 = 터널보다 훨씬 좋음(사용자)
- 체크리스트의 "80 막히면 DNS-01" 줄을 고침(443 방식도 있으므로 80만 막혀도 발급)
- 8(주소 전환): 사용자 질문 "direct 남은 걸 수정해도 되나" -> `direct`는 되돌아갈 길로 남기고 `nexus` 레코드 편집 권장. public hostname 삭제 -> DNS `nexus` CNAME 회색 -> Caddyfile에 `nexus` 추가(http-01 발급) -> 접속 OK(쓰던 브라우저만 trace 열림 = 옛 연결) -> cloudflared Stopped·Disabled
- 사용자 요청: 8 명령을 user vault 별도 파일로 -> `user/home-server/주소-전환-명령.md`
- REF-home-server가 10k자에 닿아 직통 부분을 `REF-home-direct.md`로 분할
- 남은 것: 보드 `imx8mp-f24` online 확인 · 며칠 뒤 `cloudflared service uninstall` + Zero Trust 터널 삭제
- 같은 밤: 1(내부 IP 고정) `192.168.0.2` 등록(사용자) · 재부팅 영향 질문 -> 자동 시작 전부 되어 있음, 열린 터미널만 사라짐
