import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'shuihuo', 'BatchFactoryBookSettingsModal.jsx'), 'utf8');

test('漫画约束预设从 script 目录读取，能够把已保存 ID 匹配回中文名称', () => {
  assert.match(
    source,
    /function comicConstraintOptions\(category\)[\s\S]{0,300}?catalog\.script\.filter\(item => item\.kind === 'addon' && item\.constraintCategory === category\)/,
    '漫画约束不应从 batch 目录读取，漫画约束预设属于 script 模块'
  );
});
