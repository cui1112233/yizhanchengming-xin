import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const stylesheet = readFileSync(path.join(here, '..', 'shuihuo-production.css'), 'utf8');
const novelListSource = readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');

function cssBlock(start, end) {
  const from = stylesheet.indexOf(start);
  const to = stylesheet.indexOf(end);
  assert.notEqual(from, -1, `missing ${start}`);
  assert.notEqual(to, -1, `missing ${end}`);
  return stylesheet.slice(from, to);
}

test('workbench owns one six-column layout and one horizontal scroll surface', () => {
  const layout = cssBlock('BF_WORKBENCH_LAYOUT_CONTRACT_BEGIN', 'BF_WORKBENCH_LAYOUT_CONTRACT_END');

  assert.match(layout, /--bf-workbench-columns:\s*56px\s+minmax\(300px,\s*1\.04fr\)\s+minmax\(176px,\s*\.60fr\)\s+minmax\(330px,\s*1\.16fr\)\s+minmax\(300px,\s*1\.02fr\)\s+minmax\(220px,\s*\.72fr\)/);
  assert.match(layout, /--bf-workbench-min-width:\s*1382px/);
  assert.match(layout, /\.batch-factory-workbench\s*\{[^}]*overflow:\s*hidden/s);
  assert.match(layout, /\.batch-factory-workbench-table\s*\{[^}]*overflow-x:\s*auto[^}]*overflow-y:\s*auto/s);
  assert.match(layout, /\.batch-factory-workbench-table\s+\.shuihuo-workbench-head,[\s\S]*?\.batch-factory-book-row\s*\{[^}]*grid-template-columns:\s*var\(--bf-workbench-columns\)[^}]*min-width:\s*var\(--bf-workbench-min-width\)/s);
});

test('workbench header wraps its controls instead of creating a parent horizontal scroll width', () => {
  const layout = cssBlock('BF_WORKBENCH_LAYOUT_CONTRACT_BEGIN', 'BF_WORKBENCH_LAYOUT_CONTRACT_END');

  assert.match(layout, /\.batch-factory-workbench\s*>\s*\.shuihuo-workbench-header\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*1fr\)\s+auto[^}]*grid-template-rows:\s*auto/s);
  assert.match(layout, /\.batch-factory-workbench\s+\.shuihuo-workbench-toolbar\s*\{[^}]*grid-column:\s*1\s*\/\s*-1[^}]*overflow-x:\s*auto/s);
});

test('all production modes share one compact Batch Factory table header rule', () => {
  assert.match(stylesheet, /\.batch-factory-workbench-table \.shuihuo-workbench-head\s*\{[^}]*min-height:\s*32px/s);
  assert.doesNotMatch(stylesheet, /(?:text|comic|video)[^,{]*\.shuihuo-workbench-head\s*\{[^}]*min-height/s);
});

test('selection column header keeps only the select-all checkbox', () => {
  assert.doesNotMatch(novelListSource, /aria-label="全选小说"[\s\S]{0,300}<span>序号<\/span>/);
});

test('book title and status badges share one compact baseline', () => {
  assert.match(stylesheet, /\.batch-factory-book-title-row\s*\{[^}]*justify-content:\s*flex-start[^}]*\}/);
  assert.match(stylesheet, /\.batch-factory-book-title-row\s*>\s*strong\s*\{[^}]*flex:\s*1\s+1\s+auto[^}]*\}/);
  assert.match(stylesheet, /\.batch-factory-book-title-row\s*>\s*(?:small|span)\s*\{[^}]*flex:\s*none[^}]*\}/);
});
