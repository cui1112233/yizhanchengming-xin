import test from 'node:test';
import assert from 'node:assert/strict';
import { groupResultLibrary, lineageLabel, locateShotNode } from './agent-result-lineage.js';

const shots = [
  { id: 'shot-1', scriptNodeId: 'script-1', sequence: 3, scene: '雨夜车站' }
];

const nodes = [
  {
    id: 'video-2',
    type: 'video',
    title: '镜头 3 视频',
    assetRef: { id: 'asset-2', kind: 'video', name: '第二版.mp4', url: '/asset-2' },
    config: { source: { scriptNodeId: 'script-1', shotId: 'shot-1', operation: 'video_generation', attempt: 2 } }
  },
  { id: 'script-1', type: 'script', title: '第一场剧本' }
];

const generations = [
  {
    id: 'generation-1', nodeId: 'video-2', kind: 'video', state: 'completed', selectedResultId: 'result-1',
    settings: { source: { scriptNodeId: 'script-1', shotId: 'shot-1', operation: 'video_generation', attempt: 1 } },
    results: [{ id: 'result-1', assetId: 'asset-1', kind: 'video', createdAt: '2026-10-22T01:00:00Z' }]
  },
  {
    id: 'generation-2', nodeId: 'video-2', kind: 'video', state: 'completed', selectedResultId: 'result-2',
    settings: { source: { scriptNodeId: 'script-1', shotId: 'shot-1', operation: 'video_generation', attempt: 2 } },
    results: [{ id: 'result-2', assetId: 'asset-2', kind: 'video', createdAt: '2026-10-22T02:00:00Z' }]
  }
];

const assets = [
  { id: 'asset-1', kind: 'video', name: '第一版.mp4', url: '/asset-1' },
  { id: 'asset-2', kind: 'video', name: '第二版.mp4', url: '/asset-2' }
];

test('result library groups the active version beneath its source shot', () => {
  const groups = groupResultLibrary({ shots, nodes, generations, assets });
  assert.equal(groups.length, 1);
  assert.equal(groups[0].shot.id, 'shot-1');
  assert.equal(groups[0].items[0].active, true);
  assert.equal(groups[0].items[0].asset.id, 'asset-2');
  assert.equal(groups[0].items[0].source.shotId, 'shot-1');
  assert.equal(groups[0].items[1].active, false);
});

test('the node assetRef is the active version even when an older generation selected a result', () => {
  const [group] = groupResultLibrary({ shots, nodes, generations, assets });
  assert.deepEqual(group.items.filter(item => item.active).map(item => item.asset.id), ['asset-2']);
});

test('lineage labels show the source shot operation and attempt', () => {
  const node = {
    ...nodes[0],
    config: { source: { ...nodes[0].config.source, sequence: 3 } }
  };
  assert.equal(lineageLabel(node, generations[1]), '镜头 03 · 视频生成 · 第 2 次');
});

test('locateShotNode returns the source script node before derived media nodes', () => {
  assert.equal(locateShotNode(nodes, 'shot-1').id, 'script-1');
  assert.equal(locateShotNode(nodes, 'missing'), null);
});

test('locateShotNode falls back to the selected shot script node before media exists', () => {
  const scriptOnly = [{ id: 'script-1', type: 'script', title: '第一场剧本' }];
  assert.equal(locateShotNode(scriptOnly, 'shot-1', { scriptNodeId: 'script-1' })?.id, 'script-1');
  assert.equal(locateShotNode(scriptOnly, 'shot-1', { nodeId: 'script-1' })?.id, 'script-1');
});

test('missing node or task state stays explicitly unknown', () => {
  const missingNode = groupResultLibrary({
    shots,
    nodes: [],
    generations: [{
      id: 'generation-missing-node', nodeId: 'missing', kind: 'video',
      settings: { source: { scriptNodeId: 'script-1', shotId: 'shot-1', operation: 'video_generation', attempt: 1 } },
      results: [{ id: 'result-unknown', assetId: 'asset-1', kind: 'video' }]
    }],
    assets
  });
  assert.equal(missingNode[0].items[0].status, 'unknown');

  const missingTask = groupResultLibrary({ shots, nodes: [{
    id: 'video-unknown', type: 'video', assetRef: assets[0],
    config: { source: { scriptNodeId: 'script-1', shotId: 'shot-1', operation: 'video_generation', attempt: 1 } }
  }], generations: [], assets });
  assert.equal(missingTask[0].items[0].status, 'unknown');
});
