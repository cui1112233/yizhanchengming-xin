import test from 'node:test';
import assert from 'node:assert/strict';
import {
  connectionDropChoices,
  createTypedCanvasEdge,
  syncEffectiveReferences,
  syncDirectorVideoLinks
} from './agent-canvas-connections.js';
import { appendReferenceAsset, removeReferenceAsset } from './agent-canvas-context-menu.js';

const imageAsset = { id: 'asset-image-1', kind: 'image', name: '角色图.png', url: '/image' };
const imageNode = { id: 'image-1', data: { type: 'image', assetRef: imageAsset, config: {} } };
const videoNode = { id: 'video-1', data: { type: 'video', config: { manualReferences: [{ id: 'manual-1', kind: 'image', name: '手工参考.png' }] } } };

test('图片节点拉到空白处只给出兼容的图片和视频选项', () => {
  assert.deepEqual(connectionDropChoices(imageNode).map(item => item.kind), ['image', 'video']);
});

test('图片可连接到导演台，并成为可保存的预演素材', () => {
  const directorNode = { id: 'director-1', data: { type: 'director', config: {} } };
  const edge = createTypedCanvasEdge({ source: 'image-1', target: 'director-1' }, [imageNode, directorNode]);
  const synced = syncEffectiveReferences(directorNode, [edge], [imageNode, directorNode]);

  assert.equal(edge.data.relation, 'reference');
  assert.deepEqual(synced.data.config.referenceAssets.map(item => item.id), ['asset-image-1']);
});

test('图片连到视频时创建首帧关系并自动进入视频输入区', () => {
  const edge = createTypedCanvasEdge({ source: 'image-1', target: 'video-1' }, [imageNode, videoNode]);
  assert.equal(edge.data.relation, 'first_frame');
  const synced = syncEffectiveReferences(videoNode, [edge], [imageNode, videoNode]);
  assert.deepEqual(synced.data.config.edgeReferences.map(item => item.assetId), ['asset-image-1']);
  assert.deepEqual(synced.data.config.referenceAssets.map(item => item.id), ['manual-1', 'asset-image-1']);
});

test('音频连到视频时只同步配音时长而不把音频当作模型参考文件', () => {
  const audioAsset = { id: 'asset-audio-1', kind: 'audio', name: '配音.mp3', url: '/audio' };
  const audioNode = {
    id: 'audio-1',
    data: { type: 'audio', status: 'confirmed', assetRef: audioAsset, config: { durationSeconds: 7.4 } }
  };
  const target = { id: 'video-audio-1', data: { type: 'video', config: { referenceAssets: [audioAsset] } } };
  const edge = createTypedCanvasEdge({ source: audioNode.id, target: target.id }, [audioNode, target]);
  const synced = syncEffectiveReferences(target, [edge], [audioNode, target]);

  assert.equal(edge.data.relation, 'audio_source');
  assert.equal(synced.data.config.connectedAudioDurationSeconds, 7.4);
  assert.deepEqual(synced.data.config.referenceAssets, []);
  assert.deepEqual(synced.data.config.edgeReferences, []);
});

test('文字连到图片时把上游正文保存为独立提示词来源，不覆盖节点自己填写的提示词', () => {
  const textNode = {
    id: 'text-1',
    data: { type: 'text', summary: '上游的雨夜追逐设定', config: {} }
  };
  const imageTarget = {
    id: 'image-target',
    data: { type: 'image', config: { prompt: '镜头推进', manualReferences: [] } }
  };
  const edge = createTypedCanvasEdge({ source: 'text-1', target: 'image-target' }, [textNode, imageTarget]);
  const synced = syncEffectiveReferences(imageTarget, [edge], [textNode, imageTarget]);

  assert.equal(edge.data.relation, 'script_source');
  assert.deepEqual(synced.data.config.edgePromptSources, [{ nodeId: 'text-1', text: '上游的雨夜追逐设定' }]);
  assert.equal(synced.data.config.prompt, '镜头推进');
});

test('断开连线只清除自动引用并保留手工素材', () => {
  const connected = syncEffectiveReferences(videoNode, [createTypedCanvasEdge({ source: 'image-1', target: 'video-1' }, [imageNode, videoNode])], [imageNode, videoNode]);
  const disconnected = syncEffectiveReferences(connected, [], [imageNode, connected]);
  assert.deepEqual(disconnected.data.config.manualReferences.map(item => item.id), ['manual-1']);
  assert.deepEqual(disconnected.data.config.edgeReferences, []);
  assert.deepEqual(disconnected.data.config.referenceAssets.map(item => item.id), ['manual-1']);
});

test('从收藏添加的手工图片经过画布同步后仍保留在图片节点中', () => {
  const imageTarget = {
    id: 'image-target',
    data: {
      type: 'image',
      config: { manualReferences: [], referenceAssets: [], edgeReferences: [], autoReferenceAssetIDs: [] }
    }
  };
  const selected = appendReferenceAsset({ node: imageTarget, asset: imageAsset });
  const synced = syncEffectiveReferences(selected, [], [selected]);

  assert.deepEqual(synced.data.config.manualReferences.map(item => item.id), ['asset-image-1']);
  assert.deepEqual(synced.data.config.referenceAssets.map(item => item.id), ['asset-image-1']);
});

test('删除手工参考图片经过画布同步后不会被重新加入', () => {
  const imageTarget = {
    id: 'image-target',
    data: {
      type: 'image',
      config: { manualReferences: [], referenceAssets: [], edgeReferences: [], autoReferenceAssetIDs: [] }
    }
  };
  const selected = appendReferenceAsset({ node: imageTarget, asset: imageAsset });
  const removed = removeReferenceAsset({ node: selected, assetID: imageAsset.id });
  const synced = syncEffectiveReferences(removed, [], [removed]);

  assert.deepEqual(synced.data.config.manualReferences, []);
  assert.deepEqual(synced.data.config.referenceAssets, []);
});

test('导演台只回填直接相连的视频支点，不覆盖无关视频', () => {
  const director = { id: 'director-1', data: { type: 'director', config: {} } };
  const linked = { id: 'video-linked', data: { type: 'video', config: { prompt: '旧' } } };
  const unrelated = { id: 'video-other', data: { type: 'video', config: { prompt: '保持' } } };
  const project = { shotPrompt: '镜头向前推进', scene: { assetId: 'scene-1' }, objects: [{ kind: 'actor', assetId: 'hero-1' }], directorCamera: { fov: 35 } };
  const next = syncDirectorVideoLinks([director, linked, unrelated], [{ source: director.id, target: linked.id }], director.id, project);
  assert.equal(next.find(node => node.id === linked.id).data.config.prompt, '镜头向前推进');
  assert.deepEqual(next.find(node => node.id === linked.id).data.config.referenceAssetIds, ['scene-1', 'hero-1']);
  assert.equal(next.find(node => node.id === unrelated.id).data.config.prompt, '保持');
});
