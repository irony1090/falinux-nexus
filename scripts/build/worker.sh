#!/usr/bin/env bash
# worker 빌드: build/<폴더명>/worker[.exe] — 인자 = common.sh
set -euo pipefail
source "$(dirname "$0")/common.sh"
build_parse "$@"
build_go_target worker
