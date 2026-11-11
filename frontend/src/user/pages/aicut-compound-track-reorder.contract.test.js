import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const page = readFileSync(new URL('./AICutPage.jsx', import.meta.url), 'utf8');
const workspace = readFileSync(new URL('./aicut/CompoundTimelineWorkspace.jsx', import.meta.url), 'utf8');

test('复合工作区通过统一命令提供按钮和拖动两种轨道重排入口', () => {
  assert.match(page, /type:\s*'compound\.track_reorder'/);
  assert.match(page, /onReorderTrack=/);
  assert.match(workspace, /application\/x-aicut-compound-track/);
  assert.match(workspace, /draggable/);
  assert.match(workspace, /上移轨道/);
  assert.match(workspace, /下移轨道/);
  assert.match(workspace, /onReorderTrack/);
});
