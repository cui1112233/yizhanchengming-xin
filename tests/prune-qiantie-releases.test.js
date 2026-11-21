import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const repo = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const script = path.join(repo, 'scripts', 'prune-qiantie-releases.sh');

function shellPath(value) {
  if (process.platform !== 'win32') return value;
  const normalized = path.resolve(value);
  const drive = normalized.slice(0, 1).toLowerCase();
  return `/tmp/docker-desktop-root/run/desktop/mnt/host/${drive}/${normalized.slice(3).replaceAll('\\', '/')}`;
}

function runPruner(root, keepPrevious = 1) {
  if (process.platform === 'win32') {
    return spawnSync('wsl.exe', ['env', `QIANTIE_RELEASE_ROOT=${shellPath(root)}`, 'sh', shellPath(script), String(keepPrevious)], { encoding: 'utf8' });
  }
  return spawnSync('sh', [script, String(keepPrevious)], {
    env: { ...process.env, QIANTIE_RELEASE_ROOT: root },
    encoding: 'utf8',
  });
}

test('direct-file deployment keeps current plus newest rollback and removes stale rollback binaries', () => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'qiantie-prune-direct-'));
  const current = path.join(root, 'qiantie');
  const oldRollback = path.join(root, 'qiantie.rollback-old');
  const newRollback = path.join(root, 'qiantie.rollback-new');
  fs.writeFileSync(current, 'current');
  fs.writeFileSync(oldRollback, 'old');
  fs.writeFileSync(newRollback, 'new');
  fs.utimesSync(oldRollback, new Date('2026-11-20T00:00:00Z'), new Date('2026-11-20T00:00:00Z'));
  fs.utimesSync(newRollback, new Date('2026-11-21T00:00:00Z'), new Date('2026-11-21T00:00:00Z'));

  const result = runPruner(root, 1);

  assert.equal(result.status, 0, result.stderr || result.stdout);
  assert.equal(fs.readFileSync(current, 'utf8'), 'current');
  assert.equal(fs.existsSync(newRollback), true);
  assert.equal(fs.existsSync(oldRollback), false);
});
