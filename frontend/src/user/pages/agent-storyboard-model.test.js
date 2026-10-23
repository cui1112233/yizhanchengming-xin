import test from 'node:test';
import assert from 'node:assert/strict';
import {
  buildStoryboardAssetContext,
  buildStoryboardLibrary,
  currentStoryboardShots,
  findStoryboardSelection,
  patchStoryboardNodeConfig,
  storyboardItemMetaLabel,
  storyboardSelectionKey
} from './agent-storyboard-model.js';
import { scriptSourceMarker } from './agent-script-shots.js';

const nodes = [
  { id: 'text-1', type: 'text', title: '第一场对白', summary: '雨夜里两人争执', status: 'done' },
  { id: 'audio-1', type: 'audio', title: '旁白', summary: '低沉男声', status: 'draft' },
  { id: 'image-1', type: 'image', title: '雨夜车站', status: 'done', assetRef: { id: 'asset-image-1', kind: 'image', name: '车站.png', url: '/image/1' } },
  { id: 'video-1', type: 'video', title: '列车进站', status: 'running' }
];

const assets = [
  { id: 'asset-audio-1', kind: 'audio', name: '环境声.mp3', url: '/audio/1', byteSize: 512 },
  { id: 'asset-image-1', kind: 'image', name: '车站.png', url: '/image/1', byteSize: 1024 },
  { id: 'asset-image-2', kind: 'image', name: '人物.png', url: '/image/2', byteSize: 2048 },
  { id: 'asset-video-1', kind: 'video', name: '片段.mp4', url: '/video/1', byteSize: 4096 }
];

test('故事板按文字音频图片视频分类并把素材关联回来源节点', () => {
  const library = buildStoryboardLibrary(nodes, assets);
  assert.deepEqual(library.map(section => [section.kind, section.items.length]), [
    ['shot', 0],
    ['text', 1],
    ['audio', 2],
    ['image', 2],
    ['video', 2]
  ]);
  const linked = library.find(section => section.kind === 'image').items.find(item => item.id === 'asset-image-1');
  assert.equal(linked.nodeId, 'image-1');
  assert.equal(linked.title, '雨夜车站');
  assert.equal(linked.source, 'asset');
});

test('上传的音频素材即使没有画布节点也会进入音频分区', () => {
  const audioItems = buildStoryboardLibrary(nodes, assets).find(section => section.kind === 'audio').items;
  assert.deepEqual(audioItems.map(item => item.id), ['audio-1', 'asset-audio-1']);
  assert.equal(audioItems[1].source, 'asset');
  assert.equal(audioItems[1].url, '/audio/1');
});

test('没有入库素材的媒体节点仍在故事板中显示且不会和已入库素材重复', () => {
  const library = buildStoryboardLibrary(nodes, assets);
  const videoItems = library.find(section => section.kind === 'video').items;
  assert.deepEqual(videoItems.map(item => item.id), ['video-1', 'asset-video-1']);
  assert.equal(videoItems[0].source, 'node');
});

test('故事板选择使用稳定键并能在素材更新后恢复当前项', () => {
  const library = buildStoryboardLibrary(nodes, assets);
  const key = storyboardSelectionKey({ source: 'asset', id: 'asset-image-2' });
  assert.equal(key, 'asset:asset-image-2');
  assert.equal(findStoryboardSelection(library, key)?.id, 'asset-image-2');
  assert.equal(findStoryboardSelection(library, 'missing'), null);
  assert.equal(findStoryboardSelection(library, ''), null);
});

test('故事板编辑节点配置会更新原节点且不会创建重复节点', () => {
  const original = nodes.map(node => ({ ...node }));
  const updated = patchStoryboardNodeConfig(original, 'image-1', {
    parameters: { ratio: '16:9', count: 2 },
    prompt: '雨夜车站远景'
  });

  assert.equal(updated.length, original.length);
  assert.equal(updated.find(node => node.id === 'image-1').config.prompt, '雨夜车站远景');
  assert.equal(updated.find(node => node.id === 'image-1').config.parameters.ratio, '16:9');
  assert.equal(updated.find(node => node.id === 'image-1').config.parameters.count, 2);
  assert.equal(original.find(node => node.id === 'image-1').config, undefined);
});

test('素材卡只展示节点真实配置或素材真实文件信息', () => {
  assert.equal(storyboardItemMetaLabel({
    kind: 'image',
    nodeId: 'image-1',
    config: { modelId: 'seedream-5', parameters: { ratio: '16:9', quality: 'high', resolution: '4K', count: 2 } }
  }), 'seedream-5 · 16:9 · 高清画质 · 4K · 2张');
  assert.equal(storyboardItemMetaLabel({
    kind: 'image',
    asset: { contentType: 'image/png', byteSize: 1024 }
  }), 'PNG · 1 KB');
  assert.equal(storyboardItemMetaLabel({ kind: 'image', asset: {} }), '');
});

test('加入对话的素材会去重且只暴露安全的公开信息', () => {
  const imageAsset = assets.find(asset => asset.id === 'asset-image-1');
  const videoAsset = assets.find(asset => asset.id === 'asset-video-1');
  const context = buildStoryboardAssetContext([
    { ...imageAsset, contentType: 'image/png', objectKey: 'private/key.png', secret: 'no' },
    videoAsset
  ], ['asset-image-1', 'asset-image-1', 'asset-video-1', 'missing']);

  assert.deepEqual(context, [
    {
      id: 'asset-image-1',
      assetId: 'asset-image-1',
      kind: 'image',
      name: '车站.png',
      contentType: 'image/png',
      byteSize: 1024
    },
    {
      id: 'asset-video-1',
      assetId: 'asset-video-1',
      kind: 'video',
      name: '片段.mp4',
      byteSize: 4096
    }
  ]);
  assert.equal(Object.hasOwn(context[0], 'objectKey'), false);
  assert.equal(Object.hasOwn(context[0], 'secret'), false);
});

test('故事板镜头分区只投影仍与当前剧本节点一致的镜头记录', () => {
  const scriptNode = {
    id: 'script-1',
    type: 'script',
    config: {
      sourceText: '雨夜开场',
      splitState: { state: 'completed', sourceMarker: scriptSourceMarker('雨夜开场') }
    }
  };
  const shots = [
    { id: 'shot-2', scriptNodeId: 'script-1', sequence: 2, scene: '站台', dialogue: '快上车' },
    { id: 'shot-1', scriptNodeId: 'script-1', sequence: 1, scene: '候车室', narration: '暴雨敲打玻璃' }
  ];
  const library = buildStoryboardLibrary([...nodes, scriptNode], assets, shots);
  const shotSection = library.find(section => section.kind === 'shot');
  assert.deepEqual(shotSection.items.map(item => item.id), ['shot-1', 'shot-2']);
  assert.equal(shotSection.items[0].source, 'shot');
  assert.equal(shotSection.items[0].nodeId, 'script-1');

  const stale = currentStoryboardShots([{
    ...scriptNode,
    config: { ...scriptNode.config, sourceText: '改过的剧本' }
  }], shots);
  assert.deepEqual(stale, []);
});
