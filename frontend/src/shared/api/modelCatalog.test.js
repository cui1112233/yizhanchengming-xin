import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

// Resolve the browser's extensionless client import for the native Node runner.
const source = readFileSync(new URL('./modelCatalog.js', import.meta.url), 'utf8')
  .replace(/from '\.\/client(?:\.js)?'/, `from '${new URL('./client.js', import.meta.url).href}'`);
const modelCatalogModule = await import(`data:text/javascript;base64,${Buffer.from(source).toString('base64')}`);
const { isModelKind, modelKindLabel, modelSelectOptions } = modelCatalogModule;

test('uses backend displayName for typed model choices', () => {
  assert.deepEqual(modelSelectOptions([
    { id: 'text-a', kind: 'text', displayName: '后台文本模型', name: '旧名' },
    { id: 'video-a', kind: 'video', name: '后台视频模型' }
  ], 'text'), [{ value: 'text-a', label: '后台文本模型' }]);
});

test('uses the available label fallback and skips entries without a usable ID', () => {
  assert.deepEqual(modelSelectOptions([
    { id: 'model-name', kind: 'image', name: '名称' },
    { id: 'model-id', kind: 'image', modelId: '模型标识' },
    { id: '   ', kind: 'image', displayName: '无 ID' },
    { id: 'text-a', kind: 'text', displayName: '其他类型' }
  ], 'image'), [
    { value: 'model-name', label: '名称' },
    { value: 'model-id', label: '模型标识' }
  ]);
});

test('treats audio as a supported managed model kind', () => {
  assert.equal(isModelKind('audio'), true);
  assert.equal(modelKindLabel('audio'), '音频');
});

test('treats ASR as a separate managed model kind', () => {
  assert.equal(isModelKind('asr'), true);
  assert.equal(modelKindLabel('asr'), '语音识别');
});

test('one shared catalog keeps successful model kinds and is reused by every dropdown', async () => {
  assert.equal(typeof modelCatalogModule.getAvailableModelCatalog, 'function');
  assert.equal(typeof modelCatalogModule.invalidateAvailableModelCatalog, 'function');
  modelCatalogModule.invalidateAvailableModelCatalog();
  const calls = [];
  const first = await modelCatalogModule.getAvailableModelCatalog({ loadKind: async kind => {
    calls.push(kind);
    if (kind === 'video') throw new Error('video temporarily unavailable');
    return [{ id: `${kind}-configured`, kind, enabled: true }];
  } });
  assert.deepEqual(first.models.map(model => model.id), ['text-configured', 'image-configured', 'asr-configured']);
  assert.deepEqual(first.failedKinds, ['video']);

  const second = await modelCatalogModule.getAvailableModelCatalog({ loadKind: async () => {
    throw new Error('the shared result should have been reused');
  } });
  assert.strictEqual(second, first);
  assert.deepEqual(calls, ['text', 'image', 'video', 'asr']);
});
