import test from 'node:test';
import assert from 'node:assert/strict';
import {
  normalizeNovelAsset,
  normalizeNovelPrompt,
  normalizeNovelProject,
  normalizeNovelPanelError
} from './novelPanelContracts.js';

test('normalizes a project while preserving its data payload', () => {
  const result = normalizeNovelProject({
    id: 'p-1',
    name: '测试项目',
    data: { sourceText: '第一行' },
    updated_at: '2026-10-15T10:00:00Z'
  });

  assert.deepEqual(result, {
    id: 'p-1',
    name: '测试项目',
    data: { sourceText: '第一行' },
    createdAt: '',
    updatedAt: '2026-10-15T10:00:00Z'
  });
});

test('normalizes image and metadata aliases for an asset', () => {
  const result = normalizeNovelAsset({
    type: 'character',
    id: 'char-1',
    display_name: '主角',
    main_url: '/api/reference-assets/file/character/char-1/main',
    revision: 2
  });

  assert.deepEqual(result, {
    assetType: 'character',
    assetId: 'char-1',
    name: '主角',
    imageUrl: '/api/reference-assets/file/character/char-1/main',
    revision: 2
  });
});

test('normalizes prompt text and version aliases', () => {
  assert.deepEqual(normalizeNovelPrompt({ prompt_text: '画面提示词', prompt_version: 'v2' }), {
    text: '画面提示词',
    version: 'v2'
  });
});

test('preserves retryability when normalizing an API error', () => {
  assert.deepEqual(normalizeNovelPanelError({ status: 503, error: '服务暂时不可用', retryable: true }), {
    code: 503,
    message: '服务暂时不可用',
    retryable: true
  });
});
