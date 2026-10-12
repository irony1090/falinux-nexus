# 종료 정책 확인용 스크립트

> 파트(process 종료 정책) 사용자 확인용. 설명 → `REF-process-kill.md` "종료-6(확인) 결과"
> 스크립트 노드 내용에 그대로 붙여 넣는다(EXEC = 직접 실행이라 첫 줄 shebang 필수).

## trap — SIGTERM에 안 끝남
SIGTERM을 받으면 `got-term`만 찍고 계속 돈다. 끝내려면 강제 종료(SIGKILL) → **137**.

```bash
#!/bin/bash
trap 'echo got-term' TERM
echo "trap start pid=$$"
while true; do sleep 0.2; done
```

## plain — SIGTERM에 끝남
첫 종료(SIGTERM)에 바로 끝난다 → **143**.

```bash
#!/bin/bash
echo "plain start pid=$$"
while true; do sleep 0.2; done
```

## 종료 코드
| 코드 | 뜻 |
|---|---|
| 143 | 128 + 15(SIGTERM) — 종료 요청으로 끝남 |
| 137 | 128 + 9(SIGKILL) — 강제 종료로 끝남 |
| 126 | 실행 불가(shebang 없음 등) |
