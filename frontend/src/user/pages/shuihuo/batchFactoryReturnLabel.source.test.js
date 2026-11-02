import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const source = readFileSync(new URL('./BatchFactoryNovelList.jsx', import.meta.url), 'utf8');

test('names the batch catalogue return control accurately', () => {
  assert.match(source, /title="返回批量作品列表" aria-label="返回批量作品列表"/);
});
