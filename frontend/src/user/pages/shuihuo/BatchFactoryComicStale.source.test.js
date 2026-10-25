import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const source = fs.readFileSync(new URL('./BatchFactoryNovelList.jsx', import.meta.url), 'utf8');

test('shows a non-destructive warning when a comic page generation snapshot is stale', () => {
  assert.match(source, /配置已变化，建议重新生成/);
  assert.match(source, /configStale/);
});
