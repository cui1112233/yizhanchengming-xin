import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const source = readFileSync(new URL('./BatchFactoryNovelList.jsx', import.meta.url), 'utf8');
const stylesheet = readFileSync(new URL('../shuihuo-production.css', import.meta.url), 'utf8');

test('novel list replaces the production table instead of opening over it', () => {
  assert.match(source, /novelListOpen \? <section className="batch-factory-novel-list-view"/);
  assert.match(source, /aria-label="返回生产表"/);
  assert.doesNotMatch(source, /<Modal title=\{`小说列表 · \$\{books\.length\} 本`\}/);
});

test('keeps the batch factory action header visible while the page scrolls', () => {
  assert.match(stylesheet, /\.batch-factory-workbench > \.shuihuo-workbench-header\s*\{[\s\S]*position:\s*sticky;[\s\S]*top:\s*0;[\s\S]*z-index:\s*40;/);
});
