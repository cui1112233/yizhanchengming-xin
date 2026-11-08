import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const source = readFileSync(new URL('./BatchFactoryNovelList.jsx', import.meta.url), 'utf8');

test('小说列表不展示敏感词处理提示卡片', () => {
  assert.doesNotMatch(source, /batch-factory-sensitive-review/);
  assert.doesNotMatch(source, /sensitiveReviewPresentation\(book\.sourceMetadata/);
});
