import test from 'node:test';
import assert from 'node:assert/strict';
import {
  connectionDropChoices,
  createTypedCanvasEdge,
  syncEffectiveReferences
} from './agent-canvas-connections.js';

const imageAsset = { id: 'asset-image-1', kind: 'image', name: '角色图.png', url: '/image' };
const imageNode = { id: 'image-1', data: { type: 'image', assetRef: imageAsset, config: {} } };
const videoNode = { id: 'video-1', data: { type: 'video', config: { manualReferences: [{ id: 'manual-1', kind: 'image', name: '手工参考.png' }] } } };

test('图片节点拉到空白处只给出兼容的图片和视频选项', () => {
  assert.deepEqual(connectionDropChoices(imageNode).map(item => item.kind), ['image', 'video']);
});

test('图片连到视频时创建首帧关系并自动进入视频输入区', () => {
  const edge = createTypedCanvasEdge({ source: 'image-1', target: 'video-1' }, [imageNode, videoNode]);
  assert.equal(edge.data.relation, 'first_frame');
  const synced = syncEffectiveReferences(videoNode, [edge], [imageNode, videoNode]);
  assert.deepEqual(synced.data.config.edgeReferences.map(item => item.assetId), ['asset-image-1']);
  assert.deepEqual(synced.data.config.referenceAssets.map(item => item.id), ['manual-1', 'asset-image-1']);
});

test('断开连线只清除自动引用并保留手工素材', () => {
  const connected = syncEffectiveReferences(videoNode, [createTypedCanvasEdge({ source: 'image-1', target: 'video-1' }, [imageNode, videoNode])], [imageNode, videoNode]);
  const disconnected = syncEffectiveReferences(connected, [], [imageNode, connected]);
  assert.deepEqual(disconnected.data.config.manualReferences.map(item => item.id), ['manual-1']);
  assert.deepEqual(disconnected.data.config.edgeReferences, []);
  assert.deepEqual(disconnected.data.config.referenceAssets.map(item => item.id), ['manual-1']);
});
