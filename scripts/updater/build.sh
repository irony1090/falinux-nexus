#!/usr/bin/env bash
# updater 빌드(단위-5): build/updater/<태그>/ 에 보드용 linux-arm64 · Windows amd64 + 설치 스크립트. 태그 = apps/core/cmd/updater/env 첫 줄 — REF-update-p4.md
set -euo pipefail
root="$(cd "$(dirname "$0")/../.." && pwd)"
tag="$(head -n1 "$root/apps/core/cmd/updater/env" | sed 's/^# *//')"
out="$root/build/updater/$tag"
if [[ -e "$out" ]]; then echo "이미 있음(태그를 바꿀 것): $out" >&2; exit 1; fi
mkdir -p "$out"
cd "$root/apps/core"
GOOS=linux GOARCH=arm64 go build -o "$out/linux-arm64/updater" ./cmd/updater
GOOS=windows GOARCH=amd64 go build -o "$out/windows-amd64/updater.exe" ./cmd/updater
cp "$root/scripts/updater/install.sh" "$root/scripts/updater/nexus-updater.service" "$out/linux-arm64/"
cp "$root/scripts/updater/install.ps1" "$out/windows-amd64/"
echo "$tag -> $out"
find "$out" -type f | sort
