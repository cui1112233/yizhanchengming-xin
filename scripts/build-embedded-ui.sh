#!/usr/bin/env bash
set -euo pipefail

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
frontend_dir="$repo_root/前台"
embed_dir="$repo_root/api/internal/webui/dist"

npm --prefix "$frontend_dir" run build
mkdir -p "$embed_dir"
find "$embed_dir" -mindepth 1 -maxdepth 1 ! -name placeholder.txt -exec rm -rf {} +
cp -R "$frontend_dir/dist/." "$embed_dir/"

git_sha=$(git -C "$repo_root" rev-parse HEAD)
mkdir -p "$repo_root/api/.staging-bin"
(cd "$repo_root/api" && go build -trimpath -ldflags "-X main.BuildGitSHA=$git_sha" -o .staging-bin/ycm-server ./cmd/server)
