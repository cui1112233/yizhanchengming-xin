import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';

import * as workbench from './asset-reference-workbench.js';

test('资产参考把正常输入转换成一战晟铭图片生成请求', () => {
  assert.equal(typeof workbench.buildAssetReferenceRequest, 'function');
  const request = workbench.buildAssetReferenceRequest({
    taskId: 'agent-1', nodeId: 'asset-reference-image', nodeVersion: 1,
    modelId: 'image-model', prompt: '雨夜街道', resolution: '1K', ratio: '16:9',
    count: 4, referenceAssets: [{ id: 'asset-1', kind: 'image' }], operationId: 'run-1'
  });
  assert.deepEqual(request, {
    nodeId: 'asset-reference-image', nodeVersion: 1, modelId: 'image-model', mode: 'text_to_image',
    prompt: '雨夜街道', parameters: { resolution: '1K', ratio: '16:9', quality: 'standard', count: 4 },
    advanced: {}, referenceAssetIds: ['asset-1'], idempotencyKey: 'asset-reference:agent-1:run-1'
  });
});

test('资产参考允许纯参考图生成，但拒绝提示词和参考图同时为空', () => {
  assert.equal(typeof workbench.validateAssetReferenceInput, 'function');
  assert.deepEqual(workbench.validateAssetReferenceInput({ prompt: '', referenceAssets: [{ id: 'asset-1' }], modelId: 'm' }), { ok: true, message: '' });
  assert.deepEqual(workbench.validateAssetReferenceInput({ prompt: '  ', referenceAssets: [], modelId: 'm' }), { ok: false, message: '请输入提示词或添加参考图片' });
  assert.deepEqual(workbench.validateAssetReferenceInput({ prompt: '画一只猫', referenceAssets: [], modelId: '' }), { ok: false, message: '请选择图片模型' });
});

test('资产参考约束数量、清晰度和画幅，不把洛水私有模型名带入项目', () => {
  assert.equal(typeof workbench.normalizeAssetReferenceSettings, 'function');
  assert.deepEqual(workbench.normalizeAssetReferenceSettings({ count: 99, resolution: '8K', ratio: '21:9' }), { count: 20, resolution: '1K', ratio: '21:9' });
  assert.deepEqual(workbench.normalizeAssetReferenceSettings({ count: 0, resolution: '2K', ratio: '3:4' }), { count: 1, resolution: '2K', ratio: '3:4' });
});

test('资产参考按原版链路把多张任务拆成独立单张请求', () => {
  assert.equal(typeof workbench.buildAssetReferenceRequests, 'function');
  const requests = workbench.buildAssetReferenceRequests({
    taskId: 'agent-1', nodeId: 'asset-reference-image', nodeVersion: 1,
    modelId: 'image-model', prompt: '雨夜街道', resolution: '1K', ratio: '21:9',
    count: 5, operationId: 'run-1'
  });
  assert.equal(requests.length, 5);
  assert.ok(requests.every(request => request.parameters.count === 1));
  assert.deepEqual(requests.map(request => request.idempotencyKey), [
    'asset-reference:agent-1:run-1:1',
    'asset-reference:agent-1:run-1:2',
    'asset-reference:agent-1:run-1:3',
    'asset-reference:agent-1:run-1:4',
    'asset-reference:agent-1:run-1:5'
  ]);
});

test('资产参考识别重复提交、可取消状态和失败提示', () => {
  assert.equal(typeof workbench.assetReferenceRunState, 'function');
  assert.deepEqual(workbench.assetReferenceRunState({ running: true, cancelled: false }), { canSubmit: false, canCancel: true, message: '正在生成，请勿重复提交' });
  assert.deepEqual(workbench.assetReferenceRunState({ running: false, cancelled: true }), { canSubmit: true, canCancel: false, message: '已取消本次操作' });
  assert.equal(workbench.assetReferenceErrorMessage(new Error('provider failed')), 'provider failed');
  assert.equal(workbench.assetReferenceErrorMessage(null), '图片生成失败，请稍后重试');
});

test('资产参考只展示图片结果并保留可搜索元数据', () => {
  assert.equal(typeof workbench.normalizeAssetReferenceResults, 'function');
  assert.deepEqual(workbench.normalizeAssetReferenceResults([
    { id: 'a', kind: 'image', url: '/a', name: '图A' },
    { id: 'b', kind: 'video', url: '/b' }
  ], { prompt: '雨夜', modelId: 'm', ratio: '16:9', resolution: '1K' }), [{
    id: 'a', kind: 'image', url: '/a', name: '图A', prompt: '雨夜', modelId: 'm', ratio: '16:9', resolution: '1K'
  }]);
});

test('资产参考工作台接入项目图片模型、素材上传和真实图片生成接口', () => {
  const source = fs.readFileSync(new URL('../components/AssetReferenceWorkbench.jsx', import.meta.url), 'utf8');
  assert.match(source, /listAvailableModels\('image'\)/);
  assert.match(source, /createAgentTask\(\{ projectType: 'comic' \}\)/);
  assert.match(source, /saveAgentCanvas/);
  assert.match(source, /uploadAgentAsset/);
  assert.match(source, /createAgentImageGeneration/);
  assert.match(source, /getAgentGeneration/);
  assert.match(source, /getAgentAssetBlob/);
  assert.match(source, /请输入提示词或添加参考图片/);
  assert.match(source, /模型/);
  assert.match(source, /分辨率/);
  assert.match(source, /图片尺寸/);
  assert.match(source, /数量/);
  assert.match(source, /参考图片/);
  assert.match(source, /AI生图/);
  assert.match(source, /生成结果/);
  assert.match(source, /取消/);
});

test('资产参考工作台按参考图提供完整页面骨架和批量工具', () => {
  const source = fs.readFileSync(new URL('../components/AssetReferenceWorkbench.jsx', import.meta.url), 'utf8');
  const styles = fs.readFileSync(new URL('../components/asset-reference-workbench.css', import.meta.url), 'utf8');
  assert.match(source, /asset-reference-page-head/);
  assert.match(source, /asset-reference-result-canvas/);
  assert.match(source, /已授权/);
  assert.match(source, /'取消全选' : '全选'/);
  assert.match(source, />下载</);
  assert.match(source, />批量从列表移除</);
  assert.match(source, />预设</);
  assert.match(source, />参考库</);
  assert.match(source, /输入 @ 可引用参考库卡片/);
  assert.match(source, /aria-label="关闭资产参考"/);
  assert.match(styles, /grid-template-columns:minmax\(640px,1fr\) minmax\(390px,420px\)/);
  assert.match(styles, /\.asset-reference-result-canvas\{[^}]*min-height:540px/);
  assert.match(styles, /\.asset-reference-actions\{[^}]*margin-top:auto/);
  assert.match(styles, /\.asset-reference-controls>label,\.asset-reference-controls legend,\.asset-reference-upload-head>strong\{font-size:14px;line-height:20px/);
  assert.match(styles, /@media\(max-width:1250px\)[\s\S]*?\.asset-reference-layout\{grid-template-columns:minmax\(0,1fr\) 360px/);
  assert.match(styles, /@media\(max-width:850px\)/);
  assert.doesNotMatch(styles, /\.asset-reference-controls\{order:-1/);
});
