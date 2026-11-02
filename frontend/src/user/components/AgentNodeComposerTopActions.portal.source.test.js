import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';

const source = fs.readFileSync(path.resolve('src/user/components/AgentNodeComposerTopActions.jsx'), 'utf8');

test('更多功能菜单挂到页面顶层，避免被画布节点裁掉', () => {
  assert.match(source, /getPopupContainer=\{\(\) => document\.body\}/);
  assert.match(source, /zIndex=\{1400\}/);
});
