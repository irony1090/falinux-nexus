#!/usr/bin/env bash
# supervisor 빌드: build/<폴더명>/supervisor[.exe] — 인자 = common.sh
set -euo pipefail
source "$(dirname "$0")/common.sh"
build_parse "$@"
build_go_target supervisor
