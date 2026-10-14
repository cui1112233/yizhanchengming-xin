const assert = require('node:assert/strict');
const { spawnSync } = require('node:child_process');
const test = require('node:test');
const path = require('node:path');

const root = path.resolve(__dirname, '..');
const envFile = '.env.v88-review';
const composeFile = 'docker-compose.v88-review.yml';
const fullOverride = 'docker-compose.v88-review.full.yml';

function composeConfig(...profiles) {
  const args = ['compose', '--env-file', envFile, '-f', composeFile];
  if (profiles.includes('full')) args.push('-f', fullOverride);
  for (const profile of profiles) args.push('--profile', profile);
  args.push('config', '--format', 'json');
  const result = spawnSync('docker', args, { cwd: root, encoding: 'utf8' });
  assert.equal(result.status, 0, 'Compose config must succeed without exposing command output');
  return JSON.parse(result.stdout);
}

test('review base profile excludes Worker and does not depend on it', () => {
  const config = composeConfig();
  assert.equal(config.services['novel-fetch-121-worker'], undefined);
  assert.equal(config.services.platform, undefined, 'Node platform must not be active in the default review profile');
});

test('review full profile includes Worker and platform waits for its health', () => {
  const config = composeConfig('full', 'node-legacy');
  assert.ok(config.services['novel-fetch-121-worker']);
  assert.deepEqual(config.services.platform.depends_on, {
    backend: { condition: 'service_healthy', required: true },
    'novel-fetch-121-worker': { condition: 'service_healthy', required: true },
  });
});
