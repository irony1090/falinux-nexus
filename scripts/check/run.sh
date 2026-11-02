#!/usr/bin/env bash
# 끌기 판정 전수 + 브라우저 확인 + 업데이트 app(API·화면). 임시 DB·5051에서만 돌고 끝나면 지운다. 개발·빌드본(8011~8013, nexus*)은 건드리지 않는다
# 필요: docker postgres15, 헤드리스 chromium, playwright-core(PW_DIR). 실행: scripts/check/run.sh
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
root="$(cd "$here/../.." && pwd)"
export PW_DIR="${PW_DIR:-$(ls -d "$HOME"/.npm/_npx/*/node_modules/playwright-core 2>/dev/null | head -1 | xargs -r dirname)}"
export CHROME="${CHROME:-$HOME/.cache/ms-playwright/chromium-1234/chrome-linux64/chrome}"
export CHECK_DB=nexus_check_tmp
export BASE=http://127.0.0.1:5051
work="$(mktemp -d)"
psql() { docker exec postgres15 psql -U irony -d "$1" -v ON_ERROR_STOP=1 -qAtc "$2"; }

sup=""
cleanup() {
    [[ -n "$sup" ]] && kill "$sup" 2>/dev/null && wait "$sup" 2>/dev/null || true
    psql workspace "DROP DATABASE IF EXISTS $CHECK_DB" || true
    rm -rf "$work"
}
trap cleanup EXIT

(cd "$root/apps/frontend" && npx esbuild "$here/member-drop.check.ts" --bundle --platform=node --format=cjs --outfile="$work/drop.cjs" --log-level=error)
node "$work/drop.cjs"

if ss -ltn | grep -q ':5051\b'; then echo "5051 사용 중" >&2; exit 1; fi
(cd "$root/apps/frontend" && VITE_API_URL= npx vite build --outDir "$work/web" --emptyOutDir --logLevel error)
(cd "$root/apps/core" && go build -o "$work/supervisor" ./cmd/supervisor)
sed -E -e 's/^PORT=.*/PORT=5051/' -e 's/^IP=.*/IP=127.0.0.1/' -e "s/^DB_NAME=.*/DB_NAME=$CHECK_DB/" "$root/apps/core/cmd/supervisor/dev.env" > "$work/sup.env"
printf 'WEB_DIR="%s/web"\nVIEW_CACHE_DIR=""\nUPDATE_STORE_DIR="%s/store"\n' "$work" "$work" >> "$work/sup.env"
psql workspace "DROP DATABASE IF EXISTS $CHECK_DB"
psql workspace "CREATE DATABASE $CHECK_DB OWNER irony"
"$work/supervisor" -env "$work/sup.env" > "$work/sup.log" 2>&1 &
sup=$!
for _ in $(seq 1 40); do curl -s -o /dev/null "$BASE/" && break; sleep 0.25; done
node "$here/member-drag.mjs" || { tail -5 "$work/sup.log"; exit 1; }
python3 "$here/updates-api.py" || { tail -5 "$work/sup.log"; exit 1; }
node "$here/updater-ui.mjs" || { tail -5 "$work/sup.log"; exit 1; }
