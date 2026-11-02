#!/bin/sh
# 보드에서 root로: sh install.sh <updater 실행 파일> <updater.json> — /opt/nexus-updater에 넣고 서비스 시작. REF-update-p2.md
set -eu
[ $# -eq 2 ] || { echo "사용: sh install.sh <updater> <updater.json>" >&2; exit 2; }
dir=/opt/nexus-updater
here="$(cd "$(dirname "$0")" && pwd)"
mkdir -p "$dir"
cp "$1" "$dir/updater.new" && chmod 755 "$dir/updater.new" && mv "$dir/updater.new" "$dir/updater"
[ -f "$dir/updater.json" ] || install -m 600 "$2" "$dir/updater.json"
cp "$here/nexus-updater.service" /etc/systemd/system/nexus-updater.service
systemctl daemon-reload
systemctl enable nexus-updater
systemctl restart nexus-updater
sleep 2
journalctl -u nexus-updater -n 5 --no-pager
