const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const {
  UpdateManager,
  UpdateSecurityError,
  validateUpdateBaseUrl,
  parseUpdateManifest,
  compareVersions
} = require('../src/electron/update-manager');

test('updater rejects HTTP and accepts HTTPS update origins', () => {
  assert.throws(
    () => validateUpdateBaseUrl('http://115.190.156.223:3000/downloads/local-executor/updates'),
    error => error instanceof UpdateSecurityError && error.code === 'UPDATE_HTTPS_REQUIRED'
  );
  assert.equal(
    validateUpdateBaseUrl('https://updates.example.test/downloads/local-executor/updates/'),
    'https://updates.example.test/downloads/local-executor/updates'
  );
});

test('manifest validation is fail-closed and rejects traversal', () => {
  const valid = parseUpdateManifest({
    schemaVersion: 1,
    channel: 'beta',
    version: '1.0.3',
    platform: 'win32',
    arch: 'x64',
    file: 'yizhan-local-executor-v88-1.0.3-win-x64.exe',
    sha256: 'a'.repeat(64),
    size: 12345,
    publishedAt: '2026-09-03T00:00:00.000Z'
  }, { channel: 'beta', platform: 'win32', arch: 'x64' });
  assert.equal(valid.version, '1.0.3');

  assert.throws(() => parseUpdateManifest({ ...valid, file: '../evil.exe' }, {
    channel: 'beta', platform: 'win32', arch: 'x64'
  }), /invalid update installer filename/i);
  assert.throws(() => parseUpdateManifest({ ...valid, channel: 'stable' }, {
    channel: 'beta', platform: 'win32', arch: 'x64'
  }), /channel/i);
  assert.throws(() => parseUpdateManifest({ ...valid, sha256: 'xyz' }, {
    channel: 'beta', platform: 'win32', arch: 'x64'
  }), /sha256/i);
});

test('semantic version comparison never treats an older version as newer', () => {
  assert.equal(compareVersions('1.0.3', '1.0.2') > 0, true);
  assert.equal(compareVersions('1.0.2', '1.0.2'), 0);
  assert.equal(compareVersions('1.0.1', '1.0.2') < 0, true);
  assert.equal(compareVersions('1.10.0', '1.9.9') > 0, true);
});

test('active VIDEO task defers verified installer launch', async () => {
  let spawned = 0;
  const manager = new UpdateManager({
    currentVersion: '1.0.2',
    platform: 'win32',
    arch: 'x64',
    updateBaseUrl: 'https://updates.example.test/downloads/local-executor/updates',
    userDataDir: 'C:/executor-data',
    currentTask: () => ({ id: 'job-1' }),
    spawnInstaller: async () => { spawned += 1; },
    quitApp: () => {},
    preferencesStore: fakePreferences('beta')
  });
  manager.markDownloadedForTest({
    version: '1.0.3',
    filePath: path.join('C:/executor-data', 'updates', '1.0.3.exe'),
    sha256: 'b'.repeat(64)
  });

  const result = await manager.installDownloaded();
  assert.equal(result.deferred, true);
  assert.equal(spawned, 0);
  assert.equal(manager.getState().status, 'install_deferred');
});

test('idle executor launches NSIS in update mode and forces the updated app to restart', async () => {
  const calls = [];
  const manager = new UpdateManager({
    currentVersion: '1.0.2',
    platform: 'win32',
    arch: 'x64',
    updateBaseUrl: 'https://updates.example.test/downloads/local-executor/updates',
    userDataDir: 'C:/executor-data',
    currentTask: () => null,
    spawnInstaller: async (filePath, args) => calls.push([filePath, args]),
    quitApp: () => calls.push(['quit']),
    preferencesStore: fakePreferences('beta')
  });
  manager.markDownloadedForTest({
    version: '1.0.3',
    filePath: 'C:/executor-data/updates/1.0.3.exe',
    sha256: 'c'.repeat(64)
  });

  const result = await manager.installDownloaded();
  assert.equal(result.installing, true);
  assert.deepEqual(calls[0], ['C:/executor-data/updates/1.0.3.exe', ['--updated', '/S', '--force-run']]);
  assert.deepEqual(calls[1], ['quit']);
  assert.equal(manager.getState().status, 'installing');
});

test('hash mismatch deletes the downloaded candidate and never makes it installable', async t => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'yizhan-updater-hash-'));
  t.after(() => fs.rmSync(root, { recursive: true, force: true }));
  const file = 'yizhan-local-executor-v88-1.0.3-win-x64.exe';
  const manager = new UpdateManager({
    currentVersion: '1.0.2',
    platform: 'win32',
    arch: 'x64',
    updateBaseUrl: 'https://updates.example.test/downloads/local-executor/updates',
    userDataDir: root,
    currentTask: () => null,
    preferencesStore: fakePreferences('beta'),
    fetchJson: async () => ({
      schemaVersion: 1,
      channel: 'beta',
      version: '1.0.3',
      platform: 'win32',
      arch: 'x64',
      file,
      sha256: 'e'.repeat(64),
      size: 4,
      publishedAt: '2026-09-03T00:00:00.000Z'
    }),
    downloadFile: async (_url, filePath) => {
      fs.mkdirSync(path.dirname(filePath), { recursive: true });
      fs.writeFileSync(filePath, 'test');
      return { size: 4 };
    },
    hashFile: async () => 'f'.repeat(64)
  });

  await manager.checkForUpdates();
  const state = await manager.downloadAvailable();
  assert.equal(state.status, 'error');
  assert.equal(state.errorCode, 'UPDATE_HASH_MISMATCH');
  assert.equal(fs.existsSync(path.join(root, 'updates', file)), false);
  await assert.rejects(() => manager.installDownloaded(), /no verified update installer/i);
});

test('successful update download cleans older updater installers but preserves unrelated files', async t => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'yizhan-updater-cleanup-'));
  t.after(() => fs.rmSync(root, { recursive: true, force: true }));
  const updateDir = path.join(root, 'updates');
  fs.mkdirSync(updateDir, { recursive: true });
  const oldFile = 'yizhan-local-executor-v88-1.0.2-win-x64.exe';
  const newFile = 'yizhan-local-executor-v88-1.0.3-win-x64.exe';
  fs.writeFileSync(path.join(updateDir, oldFile), 'old');
  fs.writeFileSync(path.join(updateDir, 'keep-me.txt'), 'keep');

  const manager = new UpdateManager({
    currentVersion: '1.0.2',
    platform: 'win32',
    arch: 'x64',
    updateBaseUrl: 'https://updates.example.test/downloads/local-executor/updates',
    userDataDir: root,
    currentTask: () => null,
    preferencesStore: fakePreferences('beta'),
    fetchJson: async () => ({
      schemaVersion: 1,
      channel: 'beta',
      version: '1.0.3',
      platform: 'win32',
      arch: 'x64',
      file: newFile,
      sha256: 'a'.repeat(64),
      size: 4,
      publishedAt: '2026-09-03T00:00:00.000Z'
    }),
    downloadFile: async (_url, filePath) => {
      fs.writeFileSync(filePath, 'next');
      return { size: 4 };
    },
    hashFile: async () => 'a'.repeat(64)
  });

  await manager.checkForUpdates();
  const state = await manager.downloadAvailable();
  assert.equal(state.status, 'downloaded');
  assert.equal(fs.existsSync(path.join(updateDir, oldFile)), false);
  assert.equal(fs.existsSync(path.join(updateDir, newFile)), true);
  assert.equal(fs.existsSync(path.join(updateDir, 'keep-me.txt')), true);
});

test('channel selection persists and changes the feed path', async () => {
  const store = fakePreferences('beta');
  const manager = new UpdateManager({
    currentVersion: '1.0.2',
    platform: 'win32',
    arch: 'x64',
    updateBaseUrl: 'https://updates.example.test/downloads/local-executor/updates',
    userDataDir: '/tmp/executor',
    currentTask: () => null,
    preferencesStore: store,
    fetchJson: async url => {
      assert.equal(url, 'https://updates.example.test/downloads/local-executor/updates/stable/manifest.json');
      return {
        schemaVersion: 1,
        channel: 'stable',
        version: '1.0.2',
        platform: 'win32',
        arch: 'x64',
        file: 'yizhan-local-executor-v88-1.0.2-win-x64.exe',
        sha256: 'd'.repeat(64),
        size: 100,
        publishedAt: '2026-09-03T00:00:00.000Z'
      };
    }
  });
  await manager.setChannel('stable');
  assert.equal(store.load().channel, 'stable');
  const state = await manager.checkForUpdates();
  assert.equal(state.status, 'up_to_date');
});

function fakePreferences(initial) {
  let channel = initial;
  return {
    load() { return { channel }; },
    save(next) { channel = next.channel; return { channel }; }
  };
}
