import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const source = fs.readFileSync(path.resolve('src/user/pages/shuihuo/BatchFactoryNovelList.jsx'), 'utf8');

test('comic actions expose their actual scope in the button labels', () => {
  assert.match(source, /生成整本漫画/);
  assert.match(source, /生成本页漫画/);
  assert.match(source, /整理当前主图为合集/);
});

test('comic extraction explains that complete results are reused', () => {
  assert.match(source, /已有完整漫画资产、分页和提示词时直接复用/);
  assert.doesNotMatch(source, /重新分析整本漫画/);
});
