#!/usr/bin/env bash
set -Eeuo pipefail

# This is retained only as a rollback/archive tool while the legacy retirement
# gate is open. It must never be used by a normal release: the supported
# package is deploy/v88-direct/build-direct-release.ps1 and contains Go only.
if [ "${QIANTIE_ALLOW_NODE_LEGACY_RELEASE:-}" != "1" ]; then
  echo 'Node legacy release packaging is disabled. Use the Go-only direct release builder.' >&2
  echo 'For an explicitly approved rollback archive, set QIANTIE_ALLOW_NODE_LEGACY_RELEASE=1.' >&2
  exit 78
fi

sha="${1:-${GITHUB_SHA:-}}"
out="${2:-}"

[ -n "$sha" ] || { echo 'usage: package-node-release.sh <git-sha> [output.tar.gz]' >&2; exit 2; }
[ -n "$out" ] || out="release/v88-node-${sha}.tar.gz"

mkdir -p "$(dirname "$out")"
mkdir -p data outputs
printf '%s\n' "$sha" > RELEASE-SHA

# Keep this whitelist aligned with the production Node runtime surface in Dockerfile.
required=(
  server.js
  app.js
  index.html
  package.json
  package-lock.json
  lib
  middleware
  pets
  prompts
  public
  routes
  data
  outputs
  frontend/dist
  node_modules
  RELEASE-SHA
  deploy/v88-direct/stage-node-host.sh
)

for path in "${required[@]}"; do
  [ -e "$path" ] || { echo "required release path is missing: $path" >&2; exit 3; }
done

# Explicit hard guards: these areas are never part of a Node-only payload.
# forbid backend
# forbid docs
# forbid tests
# forbid .github
# forbid services/121-browser-worker
forbidden=(backend docs tests .github services/121-browser-worker)

rm -f "$out"
tar -czf "$out" "${required[@]}"
[ -s "$out" ] || { echo 'release archive is empty' >&2; exit 4; }

members="$(mktemp)"
trap 'rm -f "$members"' EXIT
tar -tzf "$out" > "$members"
for path in "${forbidden[@]}"; do
  if grep -Eq "^${path}(/|$)" "$members"; then
    echo "forbidden path leaked into Node release: $path" >&2
    exit 5
  fi
done

PAYLOAD_BYTES="$(wc -c < "$out" | tr -d '[:space:]')"
printf 'PAYLOAD_BYTES=%s\n' "$PAYLOAD_BYTES"
du -h "$out"
