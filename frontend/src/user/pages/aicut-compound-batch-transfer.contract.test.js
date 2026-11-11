import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const page = readFileSync(new URL('./AICutPage.jsx', import.meta.url), 'utf8');
const workspace = readFileSync(new URL('./aicut/CompoundTimelineWorkspace.jsx', import.meta.url), 'utf8');

test('复合工作区复用全局多选状态并以独立载荷批量迁移', () => {
  assert.match(page, /selectedClipKeys=\{selectedClipKeys\}/);
  assert.match(page, /onToggleSelection=/);
  assert.match(workspace, /application\/x-aicut-compound-members/);
  assert.match(workspace, /compoundMembersTransferCommand/);
  assert.match(workspace, /已选择.+个成员/);
  assert.match(workspace, /批量迁移/);
  assert.match(workspace, /自动匹配.*轨/);
  assert.match(workspace, /轨道映射预览/);
  assert.match(workspace, /executeAICutCommand/);
  assert.match(workspace, /__new__/);
  assert.match(workspace, /<Modal/);
});
