#!/usr/bin/env bash
# 웹 빌드: build/<폴더명>/web/ — 인자 = common.sh(--os·--arch는 상관없음)
set -euo pipefail
source "$(dirname "$0")/common.sh"
build_parse "$@"

dir="$out/web"
refuse_exists "$dir"
mkdir -p "$out"
echo "[frontend] 타입 검사 + vite build"
# VITE_API_URL을 비워 같은 출처로 요청한다(.env의 개발용 test2 값을 덮음)
# npm run build(run-p)는 --outDir를 넘기지 못해 타입 검사와 vite build를 따로 부른다
if ! (cd "$root/apps/frontend" && npm run type-check && VITE_API_URL= npx vite build --outDir "$dir" --emptyOutDir); then
    fail_clean "$dir" "[frontend] 실패 -> $dir 지움"
fi
echo "[frontend] 완료: $dir"
