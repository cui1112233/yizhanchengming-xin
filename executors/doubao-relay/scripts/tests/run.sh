#!/usr/bin/env bash
# 运行 scripts/ 下发布脚本的全部测试(CI 的 scripts job 与本地共用)。
set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
status=0
for t in "$HERE"/*.test.sh; do
  bash "$t" || status=1
done
exit "$status"
