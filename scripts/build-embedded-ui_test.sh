#!/usr/bin/env bash
set -euo pipefail

source_script=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)/build-embedded-ui.sh
test_root=$(mktemp -d)
trap 'rm -rf "$test_root"' EXIT

fake_repo="$test_root/repo"
mkdir -p "$fake_repo/scripts" "$fake_repo/前台/dist" "$fake_repo/后台/dist" \
  "$fake_repo/api/internal/webui/dist/unknown" "$fake_repo/api/.staging-bin" "$test_root/bin"
cp "$source_script" "$fake_repo/scripts/build-embedded-ui.sh"
printf 'user-build' > "$fake_repo/前台/dist/index.html"
printf 'admin-build' > "$fake_repo/后台/dist/index.html"
printf 'preserve-me' > "$fake_repo/api/internal/webui/dist/unknown/owned.txt"
printf 'placeholder' > "$fake_repo/api/internal/webui/dist/placeholder.txt"

cat > "$test_root/bin/npm" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >> "$FAKE_COMMAND_LOG"
EOF
cat > "$test_root/bin/go" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >> "$FAKE_COMMAND_LOG"
output=''
while (($#)); do
  if [[ "$1" == "-o" ]]; then
    output=$2
    break
  fi
  shift
done
[[ -n "$output" ]]
: > "$output"
chmod +x "$output"
EOF
cat > "$test_root/bin/git" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
args=" $* "
if [[ "$args" == *" rev-parse HEAD "* ]]; then
  printf '%s\n' '0123456789abcdef'
elif [[ "$args" == *" status --porcelain "* ]]; then
  printf '%s' "${FAKE_GIT_STATUS:-}"
else
  exit 2
fi
EOF
chmod +x "$test_root/bin/npm" "$test_root/bin/go" "$test_root/bin/git"

export PATH="$test_root/bin:$PATH"
export FAKE_COMMAND_LOG="$test_root/commands.log"
export FAKE_GIT_STATUS=' M tracked-file'

"$fake_repo/scripts/build-embedded-ui.sh"

[[ -f "$fake_repo/api/internal/webui/dist/unknown/owned.txt" ]] || {
  printf 'unknown embed content was deleted\n' >&2
  exit 1
}
[[ $(<"$fake_repo/api/internal/webui/dist/user/index.html") == 'user-build' ]]
[[ $(<"$fake_repo/api/internal/webui/dist/admin/index.html") == 'admin-build' ]]
[[ -x "$fake_repo/api/.staging-bin/ycm-server" ]]
grep -F -- '-X main.BuildGitSHA=0123456789abcdef-dirty' "$FAKE_COMMAND_LOG" >/dev/null

: > "$FAKE_COMMAND_LOG"
export FAKE_GIT_STATUS=''
"$fake_repo/scripts/build-embedded-ui.sh"
grep -F -- '-X main.BuildGitSHA=0123456789abcdef -o' "$FAKE_COMMAND_LOG" >/dev/null
