import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const workspace = readFileSync(new URL('./aicut/CompoundTimelineWorkspace.jsx', import.meta.url), 'utf8');

test('拖动复合成员时显示其他复合组的兼容轨道投放区', () => {
  assert.match(workspace, /compoundTransferTargets/);
  assert.match(workspace, /compoundMemberTransferCommand/);
  assert.match(workspace, /拖到其他复合片段/);
  assert.match(workspace, /aicut-compound-transfer-dock/);
  assert.match(workspace, /onCommand\(compoundMemberTransferCommand/);
});
