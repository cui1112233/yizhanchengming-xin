import assert from 'node:assert/strict';
import test from 'node:test';
import {
  modelOptionsForProvider,
  providerForCatalogModel,
  providerOptionsForModels,
  reconcileProviderModel
} from './providerModelBinding.js';
import * as providerBinding from './providerModelBinding.js';

const catalog = [
  { id: 'personal-image', kind: 'image', displayName: '个人中心图片', provider: 'personal_api' },
  { id: 'comfy-image', kind: 'image', displayName: 'Comfy 图片', provider: 'autodl_comfyui' },
  { id: 'fan-ke-video', kind: 'video', displayName: '梵客视频', provider: 'yfai_seedance' },
  { id: 'comfy-video', kind: 'video', displayName: 'Comfy 视频', provider: 'autodl_comfyui' },
  { id: 'personal-video', kind: 'video', displayName: '个人中心视频', provider: 'personal_api' }
];

test('provider selection writes the field names consumed by image and video settings', () => {
  assert.equal(typeof providerBinding.providerModelSettingsPatch, 'function');
  assert.deepEqual(providerBinding.providerModelSettingsPatch({
    models: catalog,
    kind: 'image',
    provider: 'doubao_local_executor',
    modelId: ''
  }), {
    imageProvider: 'doubao_local_executor',
    imageModelId: 'Seedream 4.5'
  });
  assert.deepEqual(providerBinding.providerModelSettingsPatch({
    models: catalog,
    kind: 'video',
    provider: 'doubao_local_executor',
    modelId: ''
  }), {
    videoProvider: 'doubao_local_executor',
    videoModelId: 'Seedance 2.0 Fast'
  });
});

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

test('video engines add YFAI without dropping or renaming existing groups', () => {
  const options = providerOptionsForModels(catalog, 'video');
  const values = options.map(item => item.value);
  for (const value of ['personal_api', 'doubao_local_executor', 'autodl_comfyui', 'yfai_seedance']) {
    assert.ok(values.includes(value), `missing existing engine: ${value}`);
  }
  assert.ok(values.includes('yfai'), 'YFAI engine group is missing');
  assert.equal(values.length, 5);
  assert.equal(options.find(item => item.value === 'yfai').label, 'YFAI');
  assert.equal(options.find(item => item.value === 'yfai_seedance').label, '梵客视频');
  // 图片通道不受影响，仍然是原来的三组。
  assert.deepEqual(providerOptionsForModels(catalog, 'image').map(item => item.value), ['personal_api', 'doubao_local_executor', 'autodl_comfyui']);
});

test('YFAI official Seedance is grouped separately while FanKe keeps its own models', () => {
  const official = { id: 'seedance-2-0-official', kind: 'video', displayName: 'Seedance 2.0 官方' };
  const yfaiOptions = modelOptionsForProvider([...catalog, official], 'video', 'yfai');
  assert.deepEqual(yfaiOptions.map(item => item.value), ['seedance-2-0-official']);
  // 梵客分组只保留它自己的型号，官方的 Seedance 不在里面。
  const fanKeOptions = modelOptionsForProvider([...catalog, official], 'video', 'yfai_seedance');
  assert.deepEqual(fanKeOptions.map(item => item.value), ['fan-ke-video']);
  assert.equal(providerForCatalogModel(official, 'personal_api'), 'yfai');
});

test('legacy yfai_seedance + official Seedance still lands in YFAI', () => {
  const official = { id: 'seedance-2-0-official', kind: 'video', displayName: 'Seedance 2.0 官方' };
  const legacy = reconcileProviderModel({ models: [...catalog, official], kind: 'video', provider: 'yfai_seedance', modelId: 'seedance-2-0-official' });
  assert.deepEqual(legacy, { provider: 'yfai', modelId: 'seedance-2-0-official' });
  // 旧配置停留在梵客分组且选的是梵客型号时，保持原组。
  const fanKe = reconcileProviderModel({ models: catalog, kind: 'video', provider: 'yfai_seedance', modelId: 'fan-ke-video' });
  assert.deepEqual(fanKe, { provider: 'yfai_seedance', modelId: 'fan-ke-video' });
});

test('text and audio channels do not offer image or video-only executors', () => {
  assert.deepEqual(providerOptionsForModels(catalog, 'text').map(item => item.value), ['personal_api']);
  assert.deepEqual(modelOptionsForProvider(catalog, 'text', 'doubao_local_executor'), []);
});

test('OpenAI compatible catalog models belong to the personal API channel', () => {
  const textModel = { id: 'custom-text', kind: 'text', provider: 'openai_compatible' };
  assert.equal(providerForCatalogModel(textModel, 'personal_api'), 'personal_api');
  assert.deepEqual(modelOptionsForProvider([textModel], 'text', 'personal_api').map(item => item.value), ['custom-text']);
});

test('FanKe video expands the six configured models even without a catalog record', () => {
  const options = modelOptionsForProvider(catalog, 'video', 'yfai_seedance', '', { expandFanKeModels: true });
  const fanKeOptions = options.filter(item => item.value.startsWith('fanke-video-api::'));
  assert.equal(fanKeOptions.length, 6, 'the six FanKe models must always be listed');
  assert.ok(fanKeOptions.every(item => !item.label.includes('fanke-video-api')), 'labels must stay human-readable');
  // 目录里的自定义梵客记录仍会追加，且不会与六个内置型号重复。
  assert.ok(options.some(item => item.value === 'fan-ke-video'));
  assert.equal(options.filter(item => item.value === 'fanke-video-api').length, 0);
});

test('a configured FanKe catalog record merges with the expanded list without duplicates', () => {
  const row = { id: 'fanke-video-api', kind: 'video', displayName: '梵客视频 API', modelId: 'ft-video-v1-20205b9fbcec3a1094671e1c08ab7f93' };
  const models = [...catalog, row];
  const options = modelOptionsForProvider(models, 'video', 'yfai_seedance', '', { expandFanKeModels: true });
  assert.equal(options.filter(item => item.value.startsWith('fanke-video-api::')).length, 6);
  assert.equal(options.filter(item => item.value === 'fanke-video-api').length, 0);
  assert.equal(providerForCatalogModel(row, 'personal_api'), 'yfai_seedance');
});

test('FanKe legacy values normalize to concrete model selections', () => {
  const defaultSelection = 'fanke-video-api::ft-video-v1-77e8ee7a636f15dac27b2ce6d6fcd746';
  assert.deepEqual(
    reconcileProviderModel({ models: catalog, kind: 'video', provider: 'yfai_seedance', modelId: 'fanke-video-api', expandFanKeModels: true }),
    { provider: 'yfai_seedance', modelId: defaultSelection }
  );
  assert.deepEqual(
    reconcileProviderModel({ models: catalog, kind: 'video', provider: 'yfai_seedance', modelId: 'ft-video-v1-dd97afc97a7c3cba8e78a6e429933597', expandFanKeModels: true }),
    { provider: 'yfai_seedance', modelId: 'fanke-video-api::ft-video-v1-dd97afc97a7c3cba8e78a6e429933597' }
  );
  const row = { id: 'fanke-video-api', kind: 'video', displayName: '梵客视频 API', modelId: 'ft-video-v1-451adae35b0c4a3d275c2c46394abc98' };
  assert.deepEqual(
    reconcileProviderModel({ models: [...catalog, row], kind: 'video', provider: 'yfai_seedance', modelId: 'fanke-video-api', expandFanKeModels: true }),
    { provider: 'yfai_seedance', modelId: 'fanke-video-api::ft-video-v1-451adae35b0c4a3d275c2c46394abc98' }
  );
});

test('YFAI official Seedance stays as the only YFAI video model', () => {
  const official = { id: 'seedance-2-0-official', kind: 'video', displayName: 'Seedance 2.0 官方' };
  const options = modelOptionsForProvider([...catalog, official], 'video', 'yfai', '', { expandFanKeModels: true });
  assert.deepEqual(options.map(item => item.value), ['seedance-2-0-official']);
  assert.equal(providerForCatalogModel(official, 'personal_api'), 'yfai');
});

test('FanKe prefixed selections are recognized as the FanKe provider', () => {
  const selection = { id: 'fanke-video-api::ft-video-v1-77e8ee7a636f15dac27b2ce6d6fcd746', kind: 'video' };
  assert.equal(providerForCatalogModel(selection, 'personal_api'), 'yfai_seedance');
});

test('selecting the FanKe engine saves a concrete model selection', () => {
  assert.deepEqual(
    providerBinding.providerModelSettingsPatch({
      models: catalog, kind: 'video', provider: 'yfai_seedance', modelId: '', expandFanKeModels: true
    }),
    { videoProvider: 'yfai_seedance', videoModelId: 'fanke-video-api::ft-video-v1-77e8ee7a636f15dac27b2ce6d6fcd746' }
  );
});
