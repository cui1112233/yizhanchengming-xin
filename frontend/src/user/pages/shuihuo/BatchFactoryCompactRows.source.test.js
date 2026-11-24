import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const css = fs.readFileSync(new URL('../shuihuo-production.css', import.meta.url), 'utf8');

test('batch factory row controls share one compact workbench-scoped contract', () => {
  assert.match(css, /BF_COMPACT_ROW_CONTROLS_START[\s\S]*BF_COMPACT_ROW_CONTROLS_END/);
  const block = css.match(/BF_COMPACT_ROW_CONTROLS_START([\s\S]*?)BF_COMPACT_ROW_CONTROLS_END/)?.[1] || '';
  assert.match(block, /\.batch-factory-workbench-table \.shuihuo-workbench-head\s*\{[^}]*height:\s*40px/);
  assert.match(block, /\.shuihuo-workbench-head\s*>\s*div:first-child\s*\{[^}]*font-size:\s*8px/);
  assert.match(block, /\.batch-factory-book-content-action\.ant-btn\s*\{[^}]*height:\s*26px/);
  assert.match(block, /\.batch-factory-book-config-region\s*\{[^}]*padding:\s*4px 6px/);
  assert.match(block, /\.batch-factory-book-config-regions\s*\{[^}]*grid-template-rows:\s*repeat\(3,\s*40px\)/);
  assert.match(block, /\.batch-factory-book-title-row \.batch-factory-production-source\s*\{[^}]*align-self:\s*center/);
  assert.match(block, /\.batch-factory-text-page-cell\s*>\s*\.batch-factory-cell-shell/);
  assert.match(block, /\.batch-factory-text-rendered-cell\s*\{[^}]*display:\s*grid\s*!important;[^}]*grid-template-rows:\s*minmax\(0,\s*1fr\)/);
  assert.doesNotMatch(block, /\.shuihuo-workbench-toolbar|\.batch-factory-text-page-content\s+p|\.batch-factory-text-page-image\s+img/);
});
