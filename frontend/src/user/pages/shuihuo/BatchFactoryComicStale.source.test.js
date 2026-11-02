import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const source = fs.readFileSync(new URL('./BatchFactoryNovelList.jsx', import.meta.url), 'utf8');

test('shows a non-destructive warning only when the original source changed', () => {
  assert.match(source, /原文已变化，建议重新分析分页与提示词/);
  assert.match(source, /configStale/);
});
