import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const sourcePath = path.join(path.dirname(fileURLToPath(import.meta.url)), 'AccountCenterLayout.jsx');
const cssPath = path.join(path.dirname(fileURLToPath(import.meta.url)), '..', 'styles', 'account-center-visual-rebuild.css');

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

test('keeps profile, member, and security navigation available on every center page', () => {
  const source = fs.readFileSync(sourcePath, 'utf8');

  assert.doesNotMatch(source, /const profileOpen/);
  assert.doesNotMatch(source, /profileOpen \?/);
});

test('integrated CSS does not retain the fixed overlay or second content offset', () => {
  const css = fs.readFileSync(cssPath, 'utf8');

  assert.doesNotMatch(css, /\.account-center-popover\s*\{\s*position:\s*fixed/);
  assert.doesNotMatch(css, /account-center-drawer-open .legacy-main[\s\S]*margin-left:\s*284px/);
  assert.match(css, /\.account-center-workspace\s*\{/);
});
