#!/usr/bin/env bash
# 빌드본 만들기: build/<폴더명>/ 에 대상별로(scripts/build/<대상>.sh를 차례로 부름) -> REF-build-scripts.md
# 사용법: scripts/build.sh <폴더명> [supervisor|worker|frontend|updater ...] [--os linux|windows] [--arch amd64|arm64]
#   대상 생략 = supervisor worker frontend. 이미 있는 대상만 거절하므로 같은 폴더에 나중에 대상을 더 빌드할 수 있음
# embed되는 apps/core/cmd/*/env = 배포 값 -> 빌드 전에 확인할 것. 실행 시 -env 파일이 없으면 이 내용으로 만들고 종료
# 실행: ./supervisor -env ../supervisor.env / ./worker -env ../worker.env
set -euo pipefail
here="$(cd "$(dirname "$0")/build" && pwd)"
source "$here/common.sh"
usage="사용법: $0 <폴더명> [supervisor|worker|frontend|updater ...] [--os linux|windows] [--arch amd64|arm64]"

targets=() rest=()
while [[ $# -gt 0 ]]; do
    case "$1" in
        supervisor|worker|frontend|updater) [[ " ${targets[*]-} " == *" $1 "* ]] || targets+=("$1"); shift ;;
        --os|--arch) rest+=("$1" "${2:-}"); shift; [[ $# -gt 0 ]] && shift ;;
        *) rest+=("$1"); shift ;;
    esac
done
[[ ${#targets[@]} -gt 0 ]] || targets=(supervisor worker frontend)
build_parse "${rest[@]}"

has() { [[ " ${targets[*]} " == *" $1 "* ]]; }
if has supervisor && has worker; then
    sup_tag="$(tag supervisor)" wrk_tag="$(tag worker)"
    [[ "${sup_tag#SUPERVISOR-}" == "${wrk_tag#WORKER-}" ]] || echo "경고: supervisor·worker 태그가 다름: $sup_tag / $wrk_tag" >&2
fi

echo "빌드: $out · 대상 ${targets[*]}${goos:+ · os $goos}${goarch:+ · arch $goarch}"
for t in "${targets[@]}"; do
    "$here/$t.sh" "${rest[@]}"
done
echo "완료: $out (${targets[*]})"
