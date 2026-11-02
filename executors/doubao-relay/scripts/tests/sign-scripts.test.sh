#!/usr/bin/env bash
# sign-macos.sh / notarize-macos.sh 的测试:只用 DRY_RUN=1(只打印要执行的命令,不碰钥匙串、
# 不调 codesign / notarytool),断言参数校验与命令形状。
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
REPO="$(cd "$HERE/../.." && pwd)"
SIGN="$REPO/scripts/sign-macos.sh"
NOTARIZE="$REPO/scripts/notarize-macos.sh"

fails=0
pass() { printf '  ok   %s\n' "$1"; }
fail() { printf '  FAIL %s\n' "$1"; fails=$((fails + 1)); }
assert_contains() {
  case "$2" in *"$3"*) pass "$1" ;; *) fail "$1: output does not contain [$3]"$'\n'"$2" ;; esac
}
assert_eq() {
  if [ "$2" = "$3" ]; then pass "$1"; else fail "$1: expected [$2] got [$3]"; fi
}
assert_not_contains() {
  case "$2" in *"$3"*) fail "$1: output unexpectedly contains [$3]" ;; *) pass "$1" ;; esac
}
expect_fail() { # name code-expected cmd...
  local name="$1" want="$2"
  shift 2
  set +e
  "$@" >/dev/null 2>&1
  local code=$?
  set -e
  if [ "$code" -eq "$want" ]; then pass "$name (exit $code)"; else fail "$name: expected exit $want got $code"; fi
}

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

# 伪造一个 tauri bundle 目录
BUNDLE="$TMP/bundle"
mkdir -p "$BUNDLE/macos/doubao-relay.app/Contents/MacOS" "$BUNDLE/dmg"
: >"$BUNDLE/macos/doubao-relay.app/Contents/MacOS/doubao-relay"
: >"$BUNDLE/dmg/doubao-relay_0.1.0_universal.dmg"

echo "sign-macos.sh"
expect_fail "no args" 2 env -u APPLE_SIGNING_IDENTITY bash "$SIGN"
expect_fail "missing identity" 2 env -u APPLE_SIGNING_IDENTITY DRY_RUN=1 bash "$SIGN" "$BUNDLE"
expect_fail "bundle dir without .app" 2 env APPLE_SIGNING_IDENTITY="Developer ID Application: X (T)" DRY_RUN=1 bash "$SIGN" "$TMP/nope"

out="$(env -u APPLE_CERTIFICATE APPLE_SIGNING_IDENTITY="Developer ID Application: Test (TEAM123)" DRY_RUN=1 bash "$SIGN" "$BUNDLE")"
assert_contains "codesign app with hardened runtime" "$out" "codesign --force --deep --timestamp --options runtime --sign Developer\\ ID\\ Application:\\ Test\\ \\(TEAM123\\)"
assert_contains "codesign targets the .app" "$out" "doubao-relay.app"
assert_contains "verifies signature" "$out" "codesign --verify --deep --strict"
assert_contains "rebuilds dmg with hdiutil" "$out" "hdiutil create -volname doubao-relay"
assert_contains "dmg keeps tauri file name" "$out" "doubao-relay_0.1.0_universal.dmg"
assert_contains "dmg stages signed app copy" "$out" "ditto"
assert_contains "dmg has Applications link" "$out" "ln -s /Applications"
assert_contains "signs the dmg" "$out" "codesign --force --timestamp --sign"
assert_not_contains "no keychain import without certificate" "$out" "security import"

out="$(APPLE_SIGNING_IDENTITY="Developer ID Application: Test (TEAM123)" APPLE_CERTIFICATE="ZmFrZQ==" APPLE_CERTIFICATE_PASSWORD="cert-secret-123" DRY_RUN=1 bash "$SIGN" "$BUNDLE")"
assert_contains "imports certificate into temp keychain" "$out" "security import"
assert_contains "creates keychain" "$out" "security create-keychain"
assert_not_contains "never prints the certificate password" "$out" "cert-secret-123"
# 随机生成的钥匙串密码与证书密码:每个 -p / -P / -k 的值都必须打码成 \*\*\*
assert_contains "create-keychain password masked" "$out" 'create-keychain -p \*\*\* '
assert_contains "unlock-keychain password masked" "$out" 'unlock-keychain -p \*\*\* '
assert_contains "import password masked" "$out" ' -P \*\*\* '
assert_contains "partition-list password masked" "$out" ' -s -k \*\*\* '
unmasked="$(printf '%s\n' "$out" | perl -ne 'next unless /^\+ security /; while (/ (?:-p|-P|-s -k) (\S+)/g) { print "$1\n" unless $1 eq q{\*\*\*} }')"
assert_eq "no unmasked keychain/certificate password" "" "$unmasked"
assert_not_contains "never prints the certificate payload" "$out" "ZmFrZQ=="

echo "notarize-macos.sh"
DMG="$BUNDLE/dmg/doubao-relay_0.1.0_universal.dmg"
expect_fail "no args" 2 bash "$NOTARIZE"
expect_fail "missing credentials" 2 env -u APPLE_ID -u APPLE_PASSWORD -u APPLE_TEAM_ID -u APPLE_API_KEY -u APPLE_API_ISSUER -u APPLE_API_KEY_PATH DRY_RUN=1 bash "$NOTARIZE" "$DMG"
expect_fail "missing dmg" 2 env APPLE_ID=a@b.c APPLE_PASSWORD=secret APPLE_TEAM_ID=TEAM123 DRY_RUN=1 bash "$NOTARIZE" "$TMP/none.dmg"

out="$(env -u APPLE_API_KEY APPLE_ID=a@b.c APPLE_PASSWORD=secret-app-pw APPLE_TEAM_ID=TEAM123 DRY_RUN=1 bash "$NOTARIZE" "$DMG")"
assert_contains "notarytool submit --wait" "$out" "xcrun notarytool submit"
assert_contains "waits for result" "$out" "--wait"
assert_contains "apple id auth" "$out" "--apple-id a@b.c --team-id TEAM123"
assert_not_contains "password masked" "$out" "secret-app-pw"
assert_contains "staples ticket" "$out" "xcrun stapler staple"
assert_contains "validates staple" "$out" "xcrun stapler validate"

out="$(env -u APPLE_ID -u APPLE_PASSWORD APPLE_API_KEY=KEY1 APPLE_API_ISSUER=ISS1 APPLE_API_KEY_PATH=/tmp/AuthKey_KEY1.p8 DRY_RUN=1 bash "$NOTARIZE" "$DMG")"
assert_contains "api key auth" "$out" "--key /tmp/AuthKey_KEY1.p8 --key-id KEY1 --issuer ISS1"

# 非 DRY_RUN:用假的 xcrun 模拟 notarytool 被拒(退出非 0)——脚本要打印 notarytool log 并退出 1,
# 而不是在 set -e 下提前退出。
FAKEBIN="$TMP/fakebin"
mkdir -p "$FAKEBIN"
cat >"$FAKEBIN/xcrun" <<'EOF2'
#!/usr/bin/env bash
echo "xcrun $*" >>"$XCRUN_LOG"
case "$1 $2" in
  "notarytool submit")
    echo "Conducting pre-submission checks..."
    printf '{"id":"sub-123","status":"%s"}\n' "$FAKE_STATUS"
    [ "$FAKE_STATUS" = "Accepted" ] || exit 1
    ;;
  "notarytool log") echo "FAKE-NOTARY-LOG $3" ;;
esac
exit 0
EOF2
chmod +x "$FAKEBIN/xcrun"

XCRUN_LOG="$TMP/xcrun-rejected.log"
set +e
out="$(env -u APPLE_API_KEY PATH="$FAKEBIN:$PATH" XCRUN_LOG="$XCRUN_LOG" FAKE_STATUS=Invalid APPLE_ID=a@b.c APPLE_PASSWORD=secret-app-pw APPLE_TEAM_ID=TEAM123 bash "$NOTARIZE" "$DMG" 2>&1)"
code=$?
set -e
assert_eq "rejected notarization exits 1" "1" "$code"
assert_contains "rejected notarization reports status" "$out" "status=Invalid"
assert_contains "rejected notarization prints notarytool log" "$out" "FAKE-NOTARY-LOG sub-123"
assert_not_contains "rejected notarization does not staple" "$(cat "$XCRUN_LOG")" "stapler"
assert_not_contains "rejected notarization masks password" "$out" "secret-app-pw"

XCRUN_LOG="$TMP/xcrun-accepted.log"
out="$(env -u APPLE_API_KEY PATH="$FAKEBIN:$PATH" XCRUN_LOG="$XCRUN_LOG" FAKE_STATUS=Accepted APPLE_ID=a@b.c APPLE_PASSWORD=secret-app-pw APPLE_TEAM_ID=TEAM123 bash "$NOTARIZE" "$DMG" 2>&1)"
assert_contains "accepted notarization staples" "$(cat "$XCRUN_LOG")" "xcrun stapler staple"

if [ "$fails" -ne 0 ]; then
  echo "sign scripts: $fails failure(s)"
  exit 1
fi
echo "sign scripts: all passed"
