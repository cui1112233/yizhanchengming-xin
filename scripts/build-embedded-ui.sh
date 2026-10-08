#!/usr/bin/env bash
set -euo pipefail

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
frontend_dir="$repo_root/前台"
admin_dir="$repo_root/后台"
embed_dir="$repo_root/api/internal/webui/dist"
user_embed_dir="$embed_dir/user"
admin_embed_dir="$embed_dir/admin"

npm --prefix "$frontend_dir" run build
npm --prefix "$admin_dir" run build
mkdir -p "$embed_dir"
rm -rf "$user_embed_dir" "$admin_embed_dir"
mkdir -p "$user_embed_dir" "$admin_embed_dir"
cp -R "$frontend_dir/dist/." "$user_embed_dir/"
cp -R "$admin_dir/dist/." "$admin_embed_dir/"

git_sha=$(git -C "$repo_root" rev-parse HEAD)
if [[ -n "$(git -C "$repo_root" status --porcelain --untracked-files=normal)" ]]; then
  git_sha="${git_sha}-dirty"
fi
mkdir -p "$repo_root/api/.staging-bin"
(cd "$repo_root/api" && go build -trimpath -ldflags "-X main.BuildGitSHA=$git_sha" -o .staging-bin/ycm-server ./cmd/server)
