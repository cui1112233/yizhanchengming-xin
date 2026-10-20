import assert from 'node:assert/strict';
import test from 'node:test';
import {
  modelOptionsForProvider,
  providerForCatalogModel,
  providerOptionsForModels,
  reconcileProviderModel
} from './providerModelBinding.js';

const catalog = [
  { id: 'personal-image', kind: 'image', displayName: '个人中心图片', provider: 'personal_api' },
  { id: 'comfy-image', kind: 'image', displayName: 'Comfy 图片', provider: 'autodl_comfyui' },
  { id: 'fan-ke-video', kind: 'video', displayName: '梵客视频', provider: 'yfai_seedance' },
  { id: 'comfy-video', kind: 'video', displayName: 'Comfy 视频', provider: 'autodl_comfyui' },
  { id: 'personal-video', kind: 'video', displayName: '个人中心视频', provider: 'personal_api' }
];

test('only exposes models belonging to the selected provider', () => {
  assert.deepEqual(modelOptionsForProvider(catalog, 'image', 'autodl_comfyui').map(item => item.value), ['comfy-image']);
  assert.deepEqual(modelOptionsForProvider(catalog, 'video', 'yfai_seedance').map(item => item.value), ['fan-ke-video']);
  assert.deepEqual(modelOptionsForProvider(catalog, 'video', 'personal_api').map(item => item.value), ['personal-video']);
});

test('switching provider replaces a model that does not belong to the new provider', () => {
  assert.deepEqual(
    reconcileProviderModel({ models: catalog, kind: 'video', provider: 'yfai_seedance', modelId: 'comfy-video' }),
    { provider: 'yfai_seedance', modelId: 'fan-ke-video' }
  );
});

test('local Doubao uses its fixed supported model and catalog models keep their real provider', () => {
  assert.deepEqual(
    reconcileProviderModel({ models: catalog, kind: 'image', provider: 'doubao_local_executor', modelId: 'comfy-image' }),
    { provider: 'doubao_local_executor', modelId: 'Seedream 4.5' }
  );
  assert.equal(providerForCatalogModel(catalog[2], 'personal_api'), 'yfai_seedance');
  assert.equal(providerForCatalogModel({ id: 'minimax-h3-video' }, 'personal_api'), 'autodl_comfyui');
});

test('provider options include enabled catalog providers such as FanKe video', () => {
  const options = providerOptionsForModels(catalog);
  assert.ok(options.some(item => item.value === 'yfai_seedance' && item.label === '梵客视频'));
});

test('text and audio channels do not offer image or video-only executors', () => {
  assert.deepEqual(providerOptionsForModels(catalog, 'text').map(item => item.value), ['personal_api']);
  assert.deepEqual(modelOptionsForProvider(catalog, 'text', 'doubao_local_executor'), []);
});
