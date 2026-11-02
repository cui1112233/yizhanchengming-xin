import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';

const css = fs.readFileSync(new URL('./shuihuo-production.css', import.meta.url), 'utf8');

test('批量工厂小说列表在工作台剩余空间内可以纵向滚动', () => {
  assert.match(css, /\.batch-factory-novel-list-view\s*\{[^}]*min-height:\s*0[^}]*overflow-y:\s*auto/s);
});
