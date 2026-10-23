import test from 'node:test';
import assert from 'node:assert/strict';
import { findNodeMentionCandidates, insertNodeMention } from './agent-node-mentions.js';

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
