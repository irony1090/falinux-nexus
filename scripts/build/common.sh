# 빌드 스크립트 공용(source 해서 씀). 인자 = <폴더명> [--os linux|windows] [--arch amd64|arm64] -> REF-build-scripts.md
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

die() { echo "$*" >&2; exit 1; }

# 결과: name · out(build/<폴더명>) · goos · goarch(빈 값 = 이 PC) · ext(Windows = .exe)
build_parse() {
    name="" goos="" goarch=""
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --os)   [[ $# -ge 2 ]] || die "--os 값이 없음"; goos="$2"; shift 2 ;;
            --arch) [[ $# -ge 2 ]] || die "--arch 값이 없음"; goarch="$2"; shift 2 ;;
            -*)     die "모르는 옵션: $1" ;;
            *)      [[ -z "$name" ]] || die "폴더명은 하나만: $name / $1"; name="$1"; shift ;;
        esac
    done
    [[ -n "$name" ]] || die "${usage:-사용법: $0 <폴더명> [--os linux|windows] [--arch amd64|arm64]}"
    [[ "$name" != */* && "$name" != .* ]] || die "폴더명에 / 나 . 시작 금지: $name"
    [[ -z "$goos" || "$goos" == linux || "$goos" == windows ]] || die "--os = linux | windows: $goos"
    [[ -z "$goarch" || "$goarch" == amd64 || "$goarch" == arm64 ]] || die "--arch = amd64 | arm64: $goarch"
    out="$root/build/$name"
    ext=""
    [[ "$goos" == windows ]] && ext=".exe"
    return 0
}

# 빌드 구분 태그 = embed되는 env 첫 줄(# SUPERVISOR-<작업명>-<날짜>)
tag() { head -n1 "$root/apps/core/cmd/$1/env" | sed 's/^# *//'; }

refuse_exists() { [[ ! -e "$1" ]] || die "이미 있음(다른 폴더명이나 대상으로): $1"; }

# 실패 정리: 결과물을 지우고, 이번에 비게 된 build/<폴더명>도 지움
fail_clean() { rm -rf "$1"; rmdir "$out" 2>/dev/null || true; die "$2"; }

# 태그를 안 바꾸고 빌드하면 옛 빌드와 구분이 안 됨 -> 경고만. $1 = 실행 파일 이름, $2 = 태그
warn_same_tag() {
    local f
    for f in "$root"/build/*/"$1" "$root"/build/*/"$1".exe "$root"/build/*/"$1"/*/"$1" "$root"/build/*/"$1"/*/"$1".exe; do
        if [[ -f "$f" ]] && grep -aq -e "# $2\$" "$f"; then
            echo "경고: 같은 태그로 빌드한 것이 있음: $f" >&2
        fi
    done
    return 0
}

# $1 = cmd 이름, $2 = 결과 파일, $3·$4 = GOOS·GOARCH(빈 값 = 이 PC). 교차 빌드는 CGO 끔(보드 worker와 같은 방식)
go_build() {
    local envs=()
    if [[ -n "$3$4" ]]; then
        envs=(CGO_ENABLED=0)
        [[ -n "$3" ]] && envs+=("GOOS=$3")
        [[ -n "$4" ]] && envs+=("GOARCH=$4")
    fi
    env "${envs[@]}" go -C "$root/apps/core" build -o "$2" "./cmd/$1"
}

# supervisor·worker 공용: build/<폴더명>/<이름>[.exe]. 실패하면 그 파일만 지움
build_go_target() {
    local bin="$out/$1$ext" t
    t="$(tag "$1")"
    refuse_exists "$bin"
    warn_same_tag "$1" "$t"
    mkdir -p "$out"
    echo "[$1] go build ${goos:-이 PC}/${goarch:-이 PC} · 태그 $t"
    if ! go_build "$1" "$bin" "$goos" "$goarch"; then
        fail_clean "$bin" "[$1] 실패 -> $bin 지움"
    fi
    echo "[$1] 완료: $bin"
}
