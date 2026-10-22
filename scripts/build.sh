#!/usr/bin/env bash
# 빌드본 만들기: build/<이름>/{supervisor, worker, web/} -> REF-build.md
# embed되는 apps/core/cmd/*/env = 배포 값 -> 빌드 전에 확인할 것. 실행 시 -env 파일이 없으면 이 내용으로 만들고 종료
# 실행: ./supervisor -env ../supervisor.env / ./worker -env ../worker.env
set -euo pipefail

name="${1:-}"
if [[ -z "$name" ]]; then
    echo "사용법: $0 <빌드명 또는 버전명>" >&2
    exit 1
fi

root="$(cd "$(dirname "$0")/.." && pwd)"
out="$root/build/$name"
if [[ -e "$out" ]]; then
    echo "이미 있음: $out" >&2
    exit 1
fi

mkdir -p "$out"
trap 'echo "실패 -> $out 지움" >&2; rm -rf "$out"' ERR

echo "[1/2] Go 빌드"
(cd "$root/apps/core" && go build -o "$out/supervisor" ./cmd/supervisor && go build -o "$out/worker" ./cmd/worker)

echo "[2/2] 웹 빌드"
# VITE_API_URL을 비워 같은 출처로 요청한다(.env의 개발용 test2 값을 덮음)
# npm run build(run-p)는 --outDir를 넘기지 못해 타입 검사와 vite build를 따로 부른다
(cd "$root/apps/frontend" && npm run type-check && VITE_API_URL= npx vite build --outDir "$out/web" --emptyOutDir)

echo "완료: $out"
