import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const sourcePath = path.join(path.dirname(fileURLToPath(import.meta.url)), 'AccountCenterLayout.jsx');

test('renders an integrated personal-center shell with a return action', () => {
  const source = fs.readFileSync(sourcePath, 'utf8');

  assert.match(source, /className="account-center-workspace"/);
  assert.match(source, /aria-label="返回工作台"/);
  assert.match(source, /onClick=\{onReturn\}/);
  assert.doesNotMatch(source, /legacy-sidebar/);
  assert.doesNotMatch(source, /account-center-popover/);
});

test('keeps current DEV and manager menu gates', () => {
  const source = fs.readFileSync(sourcePath, 'utf8');

  assert.match(source, /\['dev', 'manager'\]\.includes\(account\?\.role\)/);
  assert.match(source, /account\?\.role === 'dev'/);
});
