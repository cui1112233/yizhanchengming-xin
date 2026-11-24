import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const css = fs.readFileSync(new URL('../shuihuo-production.css', import.meta.url), 'utf8');

test('batch factory compact override only shrinks the three book content buttons', () => {
  assert.match(css, /BF_COMPACT_ROW_CONTROLS_START[\s\S]*BF_COMPACT_ROW_CONTROLS_END/);
  const block = css.match(/BF_COMPACT_ROW_CONTROLS_START([\s\S]*?)BF_COMPACT_ROW_CONTROLS_END/)?.[1] || '';
  assert.match(block, /\.batch-factory-book-content-action\.ant-btn\s*\{[^}]*height:\s*26px/);
  assert.doesNotMatch(
    block,
    /\.shuihuo-workbench-head|\.batch-factory-book-title-row|\.batch-factory-book-content-actions|\.batch-factory-book-config|\.batch-factory-action-group|\.batch-factory-text-page|\.batch-factory-text-rendered-cell/,
  );
});
