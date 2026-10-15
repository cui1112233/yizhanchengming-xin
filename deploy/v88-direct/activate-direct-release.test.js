const assert = require('node:assert/strict');
const { spawnSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const test = require('node:test');

const here = __dirname;
const SHA = 'a'.repeat(40);
const PREVIOUS_SHA = 'c'.repeat(40);
const hasBash = spawnSync('bash', ['-n', path.join(here, 'activate-direct-release.sh')]).status === 0;

function fixture(t, { previous = true, buildSha = SHA } = {}) {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'qiantie-direct-binary-'));
  t.after(() => fs.rmSync(root, { recursive: true, force: true }));
  const direct = path.join(root, 'direct');
  const release = path.join(direct, 'releases', SHA);
  const previousRelease = path.join(direct, 'releases', PREVIOUS_SHA);
  const bin = path.join(root, 'bin');
  const service = path.join(root, 'qiantie.service');
  fs.mkdirSync(path.join(release, 'go'), { recursive: true });
  fs.writeFileSync(path.join(release, 'go', 'qiantie'), '#!/bin/sh\nexit 0\n', { mode: 0o755 });
  fs.mkdirSync(bin);
  fs.writeFileSync(service, '[Service]\nExecStart=/opt/qiantie/v88/direct/current/go/qiantie\n');
  if (previous) {
    fs.mkdirSync(path.join(previousRelease, 'go'), { recursive: true });
    fs.writeFileSync(path.join(previousRelease, 'go', 'qiantie'), '#!/bin/sh\nexit 0\n', { mode: 0o755 });
    fs.writeFileSync(path.join(previousRelease, 'runtime-release.json'), JSON.stringify({ gitSha: PREVIOUS_SHA }));
    fs.symlinkSync(previousRelease, path.join(direct, 'current'));
  }
  const systemctlLog = path.join(root, 'systemctl.log');
  const curlLog = path.join(root, 'curl.log');
  fs.writeFileSync(path.join(bin, 'systemctl'), `#!/bin/sh\necho "$*" >> "$QIANTIE_SYSTEMCTL_LOG"\nexit 0\n`, { mode: 0o755 });
  fs.writeFileSync(path.join(bin, 'curl'), `#!/bin/sh\necho "$*" >> "$QIANTIE_CURL_LOG"\ncase "$*" in *runtime-build-info*) echo '{"git_sha":"${buildSha}"}' ;; *) echo ok ;; esac\n`, { mode: 0o755 });
  const env = {
    ...process.env,
    PATH: `${bin}:${process.env.PATH}`,
    DIRECT_ROOT: direct,
    QIANTIE_SERVICE_TEMPLATE: service,
    QIANTIE_SERVICE_FILE: path.join(root, 'installed-qiantie.service'),
    QIANTIE_SYSTEMCTL_LOG: systemctlLog,
    QIANTIE_CURL_LOG: curlLog,
    SYSTEMCTL_BIN: path.join(bin, 'systemctl'),
    CURL_BIN: path.join(bin, 'curl'),
    QIANTIE_HEALTH_URL: 'http://health.test/health',
    QIANTIE_BUILD_INFO_URL: 'http://health.test/api/runtime-build-info'
  };
  return { direct, release, previousRelease, systemctlLog, curlLog, env, activate(extra = {}) {
    return spawnSync('bash', [path.join(here, 'activate-direct-release.sh'), SHA], { env: { ...env, ...extra }, encoding: 'utf8', timeout: 10000 });
  }};
}

test('activates a Go release through systemd and records the previous release', { skip: !hasBash }, (t) => {
  const f = fixture(t);
  const result = f.activate();
  assert.equal(result.status, 0, result.stdout + result.stderr);
  assert.equal(fs.readlinkSync(path.join(f.direct, 'current')), f.release);
  assert.equal(fs.readlinkSync(path.join(f.direct, 'previous')), f.previousRelease);
  assert.match(fs.readFileSync(f.systemctlLog, 'utf8'), /daemon-reload/);
  assert.match(fs.readFileSync(f.systemctlLog, 'utf8'), /restart qiantie\.service/);
  assert.equal(fs.existsSync(path.join(f.direct, 'current', 'runtime-release.json')), true);
});

test('rolls back current when runtime build identity is wrong', { skip: !hasBash }, (t) => {
  const f = fixture(t, { buildSha: 'b'.repeat(40) });
  const result = f.activate();
  assert.notEqual(result.status, 0, result.stdout + result.stderr);
  assert.equal(fs.readlinkSync(path.join(f.direct, 'current')), f.previousRelease);
  assert.match(fs.readFileSync(f.systemctlLog, 'utf8'), /restart qiantie\.service/);
});

test('first direct activation has no rollback link', { skip: !hasBash }, (t) => {
  const f = fixture(t, { previous: false });
  const result = f.activate();
  assert.equal(result.status, 0, result.stdout + result.stderr);
  assert.equal(fs.existsSync(path.join(f.direct, 'previous')), false);
});
