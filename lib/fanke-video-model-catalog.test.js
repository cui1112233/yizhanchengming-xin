const test = require('node:test');
const assert = require('node:assert/strict');

const { normalizeModelCatalog, resolveCatalogModel, publicModel } = require('./model-catalog');

test('梵客视频 API 平台预设保存用户选择的模型 ID', () => {
  const catalog = normalizeModelCatalog([{
    id: 'fanke-video-api',
    kind: 'video',
    displayName: '梵客视频 API',
    modelId: 'ft-video-v1-77e8ee7a636f15dac27b2ce6d6fcd746',
    credential: 'fanke-test-key',
    enabled: true
  }], { modelCatalogVersion: 1 });

  const model = resolveCatalogModel(catalog, 'fanke-video-api', 'video');
  assert.equal(model.adapterKind, 'fanke_video_api');
  assert.equal(model.modelId, 'ft-video-v1-77e8ee7a636f15dac27b2ce6d6fcd746');
  assert.equal(model.baseUrl, 'https://ai.fanke2026.xyz/api/open/v1');
  assert.equal(publicModel(model).hasCredential, true);
});
