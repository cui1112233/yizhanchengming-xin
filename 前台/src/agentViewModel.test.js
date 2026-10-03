import assert from 'node:assert/strict';
import test from 'node:test';

import {
  buildNavigationTarget,
  groupAgentTasks,
  normalizeMediaAssetIds,
  taskStatusMeta,
  toolCardMeta,
} from './agentViewModel.js';

test('groups tasks into operational buckets', () => {
  const grouped = groupAgentTasks([
    { id: '1', status: 'in_progress' },
    { id: '2', status: 'waiting' },
    { id: '3', status: 'needs_decision' },
    { id: '4', status: 'completed' },
    { id: '5', status: 'failed' },
  ]);
  assert.equal(grouped.active.length, 2);
  assert.equal(grouped.needsDecision.length, 1);
  assert.equal(grouped.completed.length, 1);
  assert.equal(grouped.failed.length, 1);
});

test('maps task statuses to user-facing labels', () => {
  assert.equal(taskStatusMeta('in_progress').label, '进行中');
  assert.equal(taskStatusMeta('waiting').label, '等待中');
  assert.equal(taskStatusMeta('needs_decision').label, '待我决定');
  assert.equal(taskStatusMeta('completed').label, '已完成');
  assert.equal(taskStatusMeta('failed').label, '失败');
});

test('builds safe internal navigation targets only', () => {
  assert.equal(buildNavigationTarget({ tool_name: 'ui.navigate', arguments: { path: '/novel-fetch' } }), '/novel-fetch');
  assert.equal(buildNavigationTarget({ tool_name: 'ui.navigate', arguments: { path: '/api-config', section: 'image-models' } }), '/api-config?section=image-models');
  assert.equal(buildNavigationTarget({ tool_name: 'ui.navigate', arguments: { path: 'https://evil.example' } }), '');
  assert.equal(buildNavigationTarget({ tool_name: 'unknown', arguments: { path: '/novel-fetch' } }), '');
});

test('creates product-language metadata for navigation tool cards', () => {
  const meta = toolCardMeta({ tool_name: 'ui.navigate', status: 'proposed', arguments: { path: '/batch-factory' } });
  assert.equal(meta.title, '打开批量工厂');
  assert.equal(meta.actionLabel, '打开');
});

test('normalizes opaque media asset ids and removes unsafe values', () => {
  assert.deepEqual(normalizeMediaAssetIds(['asset_1', 'asset-2', 'asset_1', '/tmp/a.png', 'https://x/a.png']), ['asset_1', 'asset-2']);
});
