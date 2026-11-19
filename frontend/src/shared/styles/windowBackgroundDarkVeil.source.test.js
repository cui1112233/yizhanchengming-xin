import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const stylesheet = fs.readFileSync(path.join(here, 'global.css'), 'utf8');

// 深色主题的文字是浅色系：窗口背景透出的遮罩必须足够实，否则亮色背景图
// 会把面板衬成灰白，正文浅字失去对比（生产实况：批量工厂表格几乎不可读）。
// 深色分支显式给出各层浓度，避免 calc 超 100% 让 color-mix 失效。
test('dark theme keeps a solid window-background veil for text contrast', () => {
  const darkBlock = stylesheet.match(/\[data-theme='dark'\] \{[^}]*\}/);
  assert.ok(darkBlock, 'global.css 应保留 [data-theme=dark] 窗口背景块');
  const block = darkBlock[0];
  assert.match(block, /--window-bg-veil: 55%;/);
  assert.match(block, /--window-bg-panel-veil: 92%;/);
  assert.match(block, /--window-bg-main-veil: 88%;/);
});
