import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const workspace = readFileSync(new URL('./aicut/CompoundTimelineWorkspace.jsx', import.meta.url), 'utf8');

test('子复合块有独立拖动载荷并使用同一跨组投放栏重新挂载', () => {
  assert.match(workspace, /application\/x-aicut-compound-block/);
  assert.match(workspace, /compoundBlockTransferTargets/);
  assert.match(workspace, /compoundBlockTransferCommand/);
  assert.match(workspace, /拖动.+复合块到其他复合片段/);
});
