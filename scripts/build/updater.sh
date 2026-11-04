#!/usr/bin/env bash
# updater 빌드: build/<폴더명>/updater/{linux-arm64, windows-amd64}/ + 설치 스크립트 — REF-update-p4.md
# --os·--arch와 상관없이 둘 다 만든다(보드 = linux-arm64, 집 서버 = windows-amd64)
set -euo pipefail
source "$(dirname "$0")/common.sh"
build_parse "$@"

dir="$out/updater"
t="$(tag updater)"
refuse_exists "$dir"
warn_same_tag updater "$t"
[[ -z "$goos$goarch" ]] || echo "참고: updater는 --os·--arch를 무시하고 linux-arm64·windows-amd64 둘 다 만듦" >&2
mkdir -p "$dir/linux-arm64" "$dir/windows-amd64"
echo "[updater] go build linux/arm64 · windows/amd64 · 태그 $t"
if ! { go_build updater "$dir/linux-arm64/updater" linux arm64 \
    && go_build updater "$dir/windows-amd64/updater.exe" windows amd64 \
    && cp "$root/scripts/updater/install.sh" "$root/scripts/updater/nexus-updater.service" "$dir/linux-arm64/" \
    && cp "$root/scripts/updater/install.ps1" "$dir/windows-amd64/"; }; then
    fail_clean "$dir" "[updater] 실패 -> $dir 지움"
fi
echo "[updater] 완료: $dir"
find "$dir" -type f | sort
