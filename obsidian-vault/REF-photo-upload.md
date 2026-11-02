# REF — 폰 사진 올리기 (실기기 확인용 업로드 도구)

> 2026-11-02 `REF-test-env.md`에서 분할(크기 10k 초과). 이력 → `history/photo-upload.md`(옛 날짜 엔트리는 `history/test-env.md`·`history/mobile-keyboard.md`·`history/board-keyboard.md`에 그대로) / 테스트 환경 → `REF-test-env.md`

## 내용 (2026-10-14, 실기기 확인용)
- `apps/frontend/dev-upload.plugin.mts`(vite 플러그인): `https://test1.falinux.com/__upload?t=<토큰>` 페이지에서 고른 사진을 user vault `obsidian-vault/user/photo_<날짜_시각>.jpg`로 저장. 900KB 넘으면 폰 브라우저에서 JPEG로 줄임(프록시 본문 한도 회피). 토큰이 없거나 틀리면 404(test1은 외부에 열려 있음)
- 올라온 사진은 사용자가 주제별 하위 폴더로 옮김(예: `user/vrKeyboard/` = 모바일 가상 키보드). user vault 전체는 git 제외(`.gitignore`)
- 평소엔 꺼 둠: `vite.config.mts`의 `// import { devUpload }`·`// devUpload(...)` 두 줄 주석을 풀면 켜짐(설정 변경 = vite 자동 재시작). 토큰 = `.env.development.local` `DEV_UPLOAD_TOKEN`
- (2026-10-20 임시) 보드 화면 오른쪽 가장자리 📎 버튼(`DevProbe.vue`, 개발 서버에서만)으로 페이지를 떠나지 않고 바로 올림. 토큰은 첫 사용 때 입력해 폰 localStorage(`dev.uploadToken`)에 — test1 JS에 토큰을 넣지 않으려고. 확인 끝나면 삭제(→ `REF-board-keyboard.md` "전체 화면 키보드 가림")
- 사진이 많으면 읽기 전에 ImageMagick `montage`로 묶거나 PIL로 필요한 부분만 잘라 보기(이미지 1장 = 컨텍스트 많이 씀)
