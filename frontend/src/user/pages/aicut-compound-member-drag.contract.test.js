import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const workspace = readFileSync(new URL('./aicut/CompoundTimelineWorkspace.jsx', import.meta.url), 'utf8');

test('复合成员有独立拖动手柄且目标轨道提交统一跨轨命令', () => {
  assert.match(workspace, /application\/x-aicut-compound-member/);
  assert.match(workspace, /拖动.+到其他兼容轨道/);
  assert.match(workspace, /compoundMemberTrackMoveCommand/);
  assert.match(workspace, /onCommand\(compoundMemberTrackMoveCommand/);
});
