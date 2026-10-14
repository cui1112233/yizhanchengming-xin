const assert = require('node:assert/strict');
const { spawnSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const test = require('node:test');

const here = __dirname;
const repo = path.resolve(here, '../..');
const SHA = 'a'.repeat(40);
const PREVIOUS_SHA = 'c'.repeat(40);
// Independently checked with shasum -a 256 for the executable fixture below.
const GO_HASH = '306c6ca7407560340797866e077e053627ad409277d1b9da58106fce4cf717cb';

function fixture(t, { sha = SHA, previous = true, relative = false, legacyPrevious = false } = {}) {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'qiantie-direct-release-'));
  t.after(() => fs.rmSync(root, { recursive: true, force: true }));
  const direct = path.join(root, 'direct');
  const deploy = path.join(root, 'deploy');
  const release = path.join(direct, 'releases', sha);
  const previousRelease = path.join(direct, 'releases', PREVIOUS_SHA);
  const bin = path.join(root, 'bin');
  const log = path.join(root, 'docker.log');
  fs.mkdirSync(path.join(release, 'go'), { recursive: true });
  fs.mkdirSync(bin);
  fs.writeFileSync(path.join(release, 'go', 'qiantie'), '#!/bin/sh\nexit 0\n', { mode: 0o755 });
  fs.mkdirSync(deploy);
  fs.writeFileSync(path.join(deploy, 'docker-compose.yml'), 'services: {}\n');
  fs.writeFileSync(path.join(deploy, '.env'), `DIRECT_RELEASE_SHA=${previous ? PREVIOUS_SHA : ''}\n`);

  // Consume the uploader's declared local assets and stage those exact files.
  // This exercises the uploaded activation tree without SSH or PowerShell.
  const uploader = fs.readFileSync(path.join(here, 'send-direct-release.ps1'), 'utf8');
  for (const match of uploader.matchAll(/-Path \(Join-Path \$RepoRoot '([^']+)'\) -Destination \$directRoot -Force/g)) {
    const source = path.join(repo, ...match[1].split('\\'));
    fs.copyFileSync(source, path.join(direct, path.basename(source)));
  }
  if (previous) {
    fs.symlinkSync(relative ? `releases/${PREVIOUS_SHA}` : previousRelease, path.join(direct, 'current'));
    fs.symlinkSync(previousRelease, path.join(direct, 'previous'));
  }
  fs.writeFileSync(path.join(bin, 'docker'), `#!/bin/sh
echo "$*" >> "$QIANTIE_TEST_DOCKER_LOG"
case "$*" in
  *"config -q"*) if [ "\${QIANTIE_TEST_FAIL_CONFIG:-}" = 1 ]; then exit 1; fi ;;
  *"ps -q go-api"*) echo go-api-test ;;
  *"inspect --format"*) echo healthy ;;
esac
`, { mode: 0o755 });
  fs.writeFileSync(path.join(bin, 'curl'), `#!/bin/sh\necho '{"git_sha":"${sha}"}'\n`, { mode: 0o755 });
  const env = {
    ...process.env,
    PATH: `${bin}:${process.env.PATH}`,
    DIRECT_ROOT: direct,
    DEPLOY_DIR: deploy,
    QIANTIE_TEST_DOCKER_LOG: log
  };
  return {
    direct, deploy, release, previousRelease, log, env,
    activate(extraEnv = {}) {
      return spawnSync('bash', [path.join(direct, 'activate-direct-release.sh'), sha], {
        env: { ...env, ...extraEnv }, encoding: 'utf8', timeout: 10000
      });
    }
  };
}

function assertNoCutover(f, result) {
  assert.notEqual(result.status, 0, result.stdout + result.stderr);
  assert.equal(fs.readlinkSync(path.join(f.direct, 'current')), f.previousRelease);
  assert.equal(fs.readlinkSync(path.join(f.direct, 'previous')), f.previousRelease);
  assert.equal(fs.readFileSync(path.join(f.deploy, '.env'), 'utf8'), `DIRECT_RELEASE_SHA=${PREVIOUS_SHA}\n`);
  assert.equal(fs.existsSync(f.log), false, 'identity rejection must precede every Docker command');
}

test('uploaded activation writes exact Go identity and retains the prior release', (t) => {
  const f = fixture(t);
  const result = f.activate();
  assert.equal(result.status, 0, result.stdout + result.stderr);
  const manifest = JSON.parse(fs.readFileSync(path.join(f.release, 'runtime-release.json'), 'utf8'));
  assert.deepEqual(Object.keys(manifest).sort(), ['activatedAtUtc', 'gitSha', 'goBinarySha256', 'previousReleaseGitSha']);
  assert.equal(manifest.gitSha, SHA);
  assert.equal(manifest.goBinarySha256, GO_HASH);
  assert.equal(manifest.previousReleaseGitSha, PREVIOUS_SHA);
  assert.match(manifest.activatedAtUtc, /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/);
  assert.equal(fs.readlinkSync(path.join(f.direct, 'current')), f.release);
  assert.equal(fs.readlinkSync(path.join(f.direct, 'previous')), f.previousRelease);
  assert.deepEqual(fs.readdirSync(f.release).sort(), ['go', 'runtime-release.json']);
});

test('direct activation waits for Go health before reloading Nginx', (t) => {
  const f = fixture(t);
  const result = f.activate();
  assert.equal(result.status, 0, result.stdout + result.stderr);
  const calls = fs.readFileSync(f.log, 'utf8');
  assert.match(calls, /up -d --no-deps --force-recreate --pull never go-api[\s\S]*inspect --format[\s\S]*restart nginx/);
  assert.match(calls, /restart nginx[\s\S]*rm -sf v88-node/);
  assert.doesNotMatch(calls, /up -d[^\n]*v88-node/);
});

for (const sha of ['abcdef0', 'g'.repeat(40), 'a'.repeat(41)]) {
  test(`activation rejects non-exact Git SHA ${sha.length} characters before cutover`, (t) => {
    const f = fixture(t, { sha });
    assertNoCutover(f, f.activate());
  });
}

test('a manifest write failure cannot switch current or change rollback state', (t) => {
  const f = fixture(t);
  fs.mkdirSync(path.join(f.release, 'runtime-release.json'));
  assertNoCutover(f, f.activate());
  assert.deepEqual(fs.readdirSync(f.release).sort(), ['go', 'runtime-release.json']);
});

test('activation independently rejects a generated manifest with the wrong Git identity', (t) => {
  const f = fixture(t);
  fs.writeFileSync(path.join(f.direct, 'runtime-release-manifest.sh'), `#!/bin/sh\nprintf '%s' '{"gitSha":"${'b'.repeat(40)}"}' > "$2/runtime-release.json"\n`);
  assertNoCutover(f, f.activate());
});

for (const [name, override] of [
  ['missing Go hash', { goBinarySha256: undefined }],
  ['malformed Go hash', { goBinarySha256: 'short' }],
  ['wrong Go binary hash', { goBinarySha256: 'b'.repeat(64) }],
  ['wrong previous identity', { previousReleaseGitSha: 'b'.repeat(40) }],
  ['missing activation time', { activatedAtUtc: undefined }]
]) {
  test(`activation rejects a generated manifest with ${name} before cutover`, (t) => {
    const f = fixture(t);
    const manifest = {
      gitSha: SHA, goBinarySha256: GO_HASH,
      activatedAtUtc: '2026-10-06T00:00:00Z', previousReleaseGitSha: PREVIOUS_SHA,
      ...override
    };
    fs.writeFileSync(path.join(f.direct, 'runtime-release-manifest.sh'), `#!/bin/sh\nprintf '%s' '${JSON.stringify(manifest)}' > "$2/runtime-release.json"\n`);
    assertNoCutover(f, f.activate());
  });
}

test('a missing uploaded manifest generator fails before cutover', (t) => {
  const f = fixture(t);
  fs.unlinkSync(path.join(f.direct, 'runtime-release-manifest.sh'));
  assertNoCutover(f, f.activate());
});

test('a malformed previous runtime identity fails before cutover', (t) => {
  const f = fixture(t);
  fs.mkdirSync(path.join(f.previousRelease), { recursive: true });
  fs.writeFileSync(path.join(f.previousRelease, 'runtime-release.json'), JSON.stringify({ gitSha: 'short' }));
  assertNoCutover(f, f.activate());
});

test('activation resolves the previous identity from a relative current symlink', (t) => {
  const f = fixture(t, { relative: true });
  const result = f.activate();
  assert.equal(result.status, 0, result.stdout + result.stderr);
  const manifest = JSON.parse(fs.readFileSync(path.join(f.release, 'runtime-release.json'), 'utf8'));
  assert.equal(manifest.previousReleaseGitSha, PREVIOUS_SHA);
  assert.equal(fs.readlinkSync(path.join(f.direct, 'previous')), f.previousRelease);
});

test('first activation explicitly records no previous direct release', (t) => {
  const f = fixture(t, { previous: false });
  const result = f.activate();
  assert.equal(result.status, 0, result.stdout + result.stderr);
  const manifest = JSON.parse(fs.readFileSync(path.join(f.release, 'runtime-release.json'), 'utf8'));
  assert.equal(manifest.previousReleaseGitSha, null);
  assert.equal(fs.existsSync(path.join(f.direct, 'previous')), false);
});

test('post-cutover Compose failure still restores the previous current release', (t) => {
  const f = fixture(t);
  const result = f.activate({ QIANTIE_TEST_FAIL_CONFIG: '1' });
  assert.notEqual(result.status, 0);
  assert.equal(fs.readlinkSync(path.join(f.direct, 'current')), f.previousRelease);
  assert.equal(fs.readFileSync(path.join(f.deploy, '.env'), 'utf8'), `DIRECT_RELEASE_SHA=${PREVIOUS_SHA}\n`);
});
