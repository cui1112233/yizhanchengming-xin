import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'NovelFetchPage.jsx'), 'utf8');
const css = fs.readFileSync(path.join(here, 'novel-fetch.css'), 'utf8');

test('处理页上方两块卡片底边对齐，当前批次保持整行展示', () => {
  const mainStart = source.indexOf('className="novel-fetch-process-main"');
  const sideStart = source.indexOf('className="novel-fetch-process-side"');
  const batchStart = source.indexOf('novel-fetch-current-batch');

  assert.notEqual(mainStart, -1, '缺少处理主栏');
  assert.notEqual(sideStart, -1, '缺少处理状态栏');
  assert.notEqual(batchStart, -1, '当前批次必须保留');
  assert.ok(sideStart < batchStart, '当前批次应排在左右两块上方卡片之后');
  assert.match(css, /\.novel-fetch-workbench \.novel-fetch-process-main\s*\{[^}]*display:\s*flex[^}]*flex-direction:\s*column/s);
  assert.match(css, /\.novel-fetch-workbench \.novel-fetch-process-grid\s*\{[^}]*align-items:\s*stretch/s);
  assert.match(css, /\.novel-fetch-workbench \.novel-fetch-process-main > \.novel-fetch-public-workspace\s*\{[^}]*flex:\s*1 1 auto/s);
  assert.match(css, /\.novel-fetch-workbench \.novel-fetch-process-main > \.novel-fetch-public-workspace\s*\{[^}]*display:\s*flex[^}]*flex-direction:\s*column[^}]*padding-bottom:\s*17px/s);
  assert.match(css, /\.novel-fetch-workbench \.novel-fetch-process-main > \.novel-fetch-public-workspace > \.novel-fetch-public-actions\s*\{[^}]*margin-top:\s*auto/s);
  assert.match(css, /\.novel-fetch-workbench \.novel-fetch-public-actions\s*\{[^}]*width:\s*100%[^}]*min-width:\s*0[^}]*flex-wrap:\s*wrap/s);
  assert.match(css, /\.novel-fetch-workbench \.novel-fetch-process-grid\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*2fr\)\s+minmax\(300px,\s*1fr\)/s);
});

test('待处理清单始终保持可读的稳定高度', () => {
  assert.match(css, /\.novel-fetch-workbench \.novel-fetch-public-input,[\s\S]*?height:\s*400px\s*!important/);
  assert.match(css, /\.novel-fetch-workbench \.novel-fetch-public-input,[\s\S]*?min-height:\s*400px\s*!important/);
  assert.match(css, /\.novel-fetch-workbench \.novel-fetch-public-input\.has-content,[\s\S]*?height:\s*400px\s*!important/);
});
