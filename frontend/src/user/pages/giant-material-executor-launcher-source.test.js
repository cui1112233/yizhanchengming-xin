import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const launcher = fs.readFileSync(path.resolve(here, '../../../../executors/giant-material-executor/portable/start-giant-material-executor.cmd'), 'utf8');

test('Windows launcher keeps startup failures observable', () => {
  assert.match(launcher, /executor\.log/i);
  assert.match(launcher, /Test-NetConnection/i);
  assert.match(launcher, /17861/);
  assert.match(launcher, /pause/i);
  assert.match(launcher, /GIANT_MATERIAL_EXECUTOR_PAIRING_CODE/);
  assert.match(launcher, /请输入配对码/);
});
