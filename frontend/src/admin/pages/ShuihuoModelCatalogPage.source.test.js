import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const source = readFileSync(new URL('./ShuihuoModelCatalogPage.jsx', import.meta.url), 'utf8');

test('model catalog offers safe disable and enable actions instead of only destructive removal', () => {
  assert.match(source, /setAdminModelEnabled/);
  assert.match(source, /enabled \? '停用' : '启用'/);
  assert.match(source, /确认停用这个模型/);
});
