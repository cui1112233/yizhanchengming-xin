import test from 'node:test';
import assert from 'node:assert/strict';
import { isCatalogModelAvailable, isPlatformManagedModel, mergeGroupEntries, modelIdentity, normalizeEndpoint } from './apiConfigCatalogMerge.js';

test('地址归一化忽略大小写与末尾斜杠', () => {
  assert.equal(normalizeEndpoint(' HTTPS://API.Example.com/v1/ '), 'https://api.example.com/v1');
  assert.equal(modelIdentity('https://api.example.com/v1/', ' GPT-Image-2.5 '), 'https://api.example.com/v1|gpt-image-2.5');
});

test('地址或模型名缺失时不判定为同一个模型', () => {
  assert.equal(modelIdentity('', 'gpt-image-2'), '');
  assert.equal(modelIdentity('https://api.example.com', ''), '');
});

test('与目录模型同源的渠道被合并，不重复展示', () => {
  const result = mergeGroupEntries({
    catalogModels: [{ id: 'm1', kind: 'image', displayName: 'gpt-image-2', baseUrl: 'https://grsai.ai', modelId: 'gpt-image-2', enabled: true }],
    channelEntries: [{ id: 'grsai', name: 'grsai', baseUrl: 'https://grsai.ai/', model: 'GPT-Image-2' }]
  });
  assert.equal(result.catalog.length, 1);
  assert.equal(result.channels.length, 0);
  assert.equal(result.mergedChannels.length, 1);
  assert.equal(result.mergedChannels[0].name, 'grsai');
});

test('不同地址或不同模型名不合并', () => {
  const result = mergeGroupEntries({
    catalogModels: [{ id: 'm1', kind: 'text', displayName: '演示', baseUrl: 'https://a.example.com/v1', modelId: 'demo' }],
    channelEntries: [
      { id: 'c1', name: '同模型不同地址', baseUrl: 'https://b.example.com/v1', model: 'demo' },
      { id: 'c2', name: '同地址不同模型', baseUrl: 'https://a.example.com/v1', model: 'other' }
    ]
  });
  assert.equal(result.catalog.length, 1);
  assert.equal(result.channels.length, 2);
  assert.equal(result.mergedChannels.length, 0);
});

test('目录内部重复条目只保留第一条', () => {
  const result = mergeGroupEntries({
    catalogModels: [
      { id: 'm1', baseUrl: 'https://a.example.com/v1', modelId: 'demo', enabled: true },
      { id: 'm2', baseUrl: 'https://A.example.com/v1/', modelId: 'DEMO', enabled: false }
    ]
  });
  assert.equal(result.catalog.length, 1);
  assert.equal(result.catalog[0].id, 'm1');
  assert.equal(result.catalogDuplicates, 1);
});

test('两块都为空时返回空结果', () => {
  const result = mergeGroupEntries({});
  assert.deepEqual(result, { catalog: [], channels: [], mergedChannels: [], catalogDuplicates: 0 });
});

test('服务端下发的 available 优先，页面与下拉同口径', () => {
  assert.equal(isCatalogModelAvailable({ id: 'm1', hasCredential: true, available: false }), false);
  assert.equal(isCatalogModelAvailable({ id: 'm2', hasCredential: false, available: true }), true);
});

test('没有 available 时按凭证推导：有 Key 必出现，空 Key 必消失', () => {
  assert.equal(isCatalogModelAvailable({ id: 'm1', hasCredential: true }), true);
  assert.equal(isCatalogModelAvailable({ id: 'm2', hasCredential: false, enabled: true }), false);
});

test('已下线与旧 enabled=false 条目都不展示', () => {
  assert.equal(isCatalogModelAvailable({ id: 'm1', displayName: '[off]gpt-image-2.5', hasCredential: true }), false);
  assert.equal(isCatalogModelAvailable({ id: 'm2', hasCredential: true, enabled: false }), false);
});

test('本地执行器没有 Key 也不会被删除', () => {
  assert.equal(isCatalogModelAvailable({ id: 'local-doubao-executor-video', provider: 'local_executor', hasCredential: false }), true);
  assert.equal(isCatalogModelAvailable({ id: 'local-grok-image', hasCredential: false }), true);
});

test('平台预设按面板归属过滤，新增预设 id 自动生效', () => {
  const presetIds = new Set(['yd2-mini-video', 'minimax-h3-video']);
  assert.equal(isPlatformManagedModel({ id: 'yd2-mini-video' }, presetIds), true);
  // 面板以后新增的条目不需要改组里的代码。
  assert.equal(isPlatformManagedModel({ id: 'new-video-preset' }, new Set([...presetIds, 'new-video-preset'])), true);
  assert.equal(isPlatformManagedModel({ id: 'local-doubao-executor-video', provider: 'local_executor' }, presetIds), true);
});

test('普通目录模型不会被当成平台预设藏起来', () => {
  assert.equal(isPlatformManagedModel({ id: 'custom-text-model' }, new Set(['yd2-mini-video'])), false);
});
