#!/usr/bin/env bash
set -Eeuo pipefail

echo 'Retired: V88 public releases use deploy/v88-direct/build-direct-release.ps1 and send-direct-release.ps1.' >&2
echo 'Use scripts/deploy-v88-public-unified.sh with paired immutable Node and Go images.' >&2
exit 64
