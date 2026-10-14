#!/usr/bin/env bash
# Usage: bash runtime-release-manifest.sh <exact-git-sha> <release-dir> [previous-git-sha]
# The release identity is derived from the staged Go executable. This file
# contains identities only, never .env.
set -Eeuo pipefail

python3 - "${1:-}" "${2:-}" "${3:-}" <<'PY'
import datetime
import hashlib
import json
import os
from pathlib import Path
import re
import sys
import tempfile

sha, release_dir, previous_sha = sys.argv[1:]
temporary = None
try:
    if not re.fullmatch(r"[0-9a-f]{40}", sha):
        raise ValueError("requested Git identity must be an exact 40-character SHA")
    if previous_sha and not re.fullmatch(r"[0-9a-f]{40}", previous_sha):
        raise ValueError("previous Git identity must be an exact 40-character SHA")
    if not release_dir:
        raise ValueError("release directory is required")
    release = Path(release_dir)
    binary = release / "go" / "qiantie"
    if not binary.is_file() or not os.access(binary, os.X_OK):
        raise ValueError("staged Go executable is missing")
    digest = hashlib.sha256()
    with binary.open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    manifest = {
        "gitSha": sha,
        "goBinarySha256": digest.hexdigest(),
        "activatedAtUtc": datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
        "previousReleaseGitSha": previous_sha or None,
    }
    with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", dir=release,
                                     prefix=".runtime-release.", delete=False) as stream:
        temporary = stream.name
        json.dump(manifest, stream, indent=2)
        stream.write("\n")
        stream.flush()
        os.fsync(stream.fileno())
    os.replace(temporary, release / "runtime-release.json")
    temporary = None
except (OSError, ValueError) as error:
    print(f"STATUS=FAILED runtime release identity: {error}", file=sys.stderr)
    sys.exit(9)
finally:
    if temporary is not None:
        os.unlink(temporary)
PY
