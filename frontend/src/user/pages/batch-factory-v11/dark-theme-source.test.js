import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const read = name => fs.readFileSync(path.join(here, name), 'utf8');

test('BF11 dark theme owns readable tokens and paints the page gutters', () => {
  const css = read('batch-factory-v11-theme.css');
  // 色板改为引用全局语义变量，深浅两套由 global.css 提供，不再各自写死。
  for (const token of ['--surface-base', '--surface-panel', '--surface-2', '--border-default', '--text-primary', '--text-secondary', '--text-tertiary', '--theme-primary']) {
    assert.match(css, new RegExp(`var\\(${token}\\)`), `missing ${token}`);
  }
  assert.match(css, /\[data-bf-v11-ui=['"]final['"]\]/);
  assert.match(css, /background:\s*var\(--bf11-bg\)/);
});

test('BF11 dark theme explicitly fixes AntD contrast-sensitive controls', () => {
  const css = read('batch-factory-v11-theme.css');
  for (const selector of ['ant-typography-secondary', 'ant-tag', 'ant-segmented', 'ant-input', 'ant-btn']) {
    assert.match(css, new RegExp(selector));
  }
  assert.match(css, /::placeholder/);
});
