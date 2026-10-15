import test from 'node:test';
import assert from 'node:assert/strict';
import { readNovelFetchProcessSettings, writeNovelFetchProcessSettings } from './novel-fetch-process-settings.js';

test('readNovelFetchProcessSettings_withSavedV78Options_restoresAllProcessingSelections', () => {
  const storage = new Map([
    ['novel-fetch-v78-process-settings', JSON.stringify({
      platform: 15,
      maxTxt: 4000,
      textModel: '[anti-2]gemini-3.8-flash',
      selectedVersions: ['original', 'ai1'],
      rewriteMethods: { ai1: 'auto' },
      sensitiveAiProcessEnabled: true
    })]
  ]);

  const settings = readNovelFetchProcessSettings(storage);

  assert.equal(settings.platform, 15);
  assert.equal(settings.maxTxt, 4000);
  assert.deepEqual(settings.selectedVersions, ['original', 'ai1']);
  assert.equal(settings.sensitiveAiProcessEnabled, true);
});

test('writeNovelFetchProcessSettings_withV78Options_persistsValuesForNextPageVisit', () => {
  const storage = new Map();

  writeNovelFetchProcessSettings({
    platform: 2,
    maxTxt: 8000,
    textModel: 'gemini-3.5-flash',
    selectedVersions: ['original'],
    rewriteMethods: { ai1: 'high_imitation' },
    sensitiveAiProcessEnabled: false
  }, storage);

  assert.deepEqual(JSON.parse(storage.get('novel-fetch-v78-process-settings')), {
    platform: 2,
    maxTxt: 8000,
    textModel: 'gemini-3.5-flash',
    selectedVersions: ['original'],
    rewriteMethods: { ai1: 'high_imitation' },
    sensitiveAiProcessEnabled: false
  });
});
