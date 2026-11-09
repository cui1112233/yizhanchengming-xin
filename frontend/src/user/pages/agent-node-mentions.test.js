import test from 'node:test';
import assert from 'node:assert/strict';
import { findAgentMentionCandidates, findNodeMentionCandidates, insertAgentMention, insertNodeMention } from './agent-node-mentions.js';

test('findNodeMentionCandidates matches node titles after a trailing at sign', () => {
  const nodes = [
    { id: 'script', type: 'text', title: '完整剧本', summary: '第一集' },
    { id: 'cover', type: 'image', title: '封面图', summary: '主视觉' }
  ];
  assert.deepEqual(findNodeMentionCandidates(nodes, '请参考 @剧'), [nodes[0]]);
});

test('findNodeMentionCandidates only opens for a trailing mention token', () => {
  assert.deepEqual(findNodeMentionCandidates([{ id: 'script', title: '完整剧本' }], '请参考 @完整剧本 后续'), []);
});

test('insertNodeMention replaces the active mention token and keeps the rest of the prompt', () => {
  assert.equal(insertNodeMention('请参考 @封', { title: '封面图' }), '请参考 @封面图 ');
});

test('unified mentions return stable node and asset ids without copying transient selection state', () => {
  const nodes = [{ id: 'node-script-1', type: 'text', title: '女主剧本' }];
  const assets = [{ id: 'asset-image-1', kind: 'image', name: '女主定妆照' }];
  assert.deepEqual(findAgentMentionCandidates({ nodes, assets }, '参考 @女'), [
    { mentionSource: 'node', referenceId: 'node-script-1', kind: 'text', title: '女主剧本' },
    { mentionSource: 'asset', referenceId: 'asset-image-1', kind: 'image', title: '女主定妆照' }
  ]);
  assert.equal(insertAgentMention('参考 @女', { title: '女主定妆照' }), '参考 @女主定妆照 ');
});

test('clearing selected ids produces no stale mention context', async () => {
  const { buildSelectedNodeContext } = await import('./agent-canvas-state.js');
  const { buildStoryboardAssetContext } = await import('./agent-storyboard-model.js');
  assert.deepEqual(buildSelectedNodeContext({ nodes: [{ id: 'node-1', title: '节点' }] }, []), []);
  assert.deepEqual(buildStoryboardAssetContext([{ id: 'asset-1', kind: 'image', name: '素材' }], []), []);
});
