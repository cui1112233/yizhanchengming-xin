import test from 'node:test';
import assert from 'node:assert/strict';
import { actComicAssetMedia } from './comicScreenplayAssetActions.js';
test('asset and voice actions preserve explicit schema and cancellable project request', async () => {
  const calls = [], controller = new AbortController();
  const request = async (path, options) => { calls.push({ path, options }); return {}; };
  await actComicAssetMedia('p', { action: 'voice_design', targetId: 'voice-draft:1', prompt: '温柔', text: '试听', expectedRevision: 7, idempotencyKey: 'once' }, { signal: controller.signal }, request);
  await actComicAssetMedia('p', { action: 'voice_design_drafts', items: [{ id: 1, name: '旁白', role: '叙述', prompt: '', preview_text: '' }], expectedRevision: 8 }, {}, request);
  assert.equal(calls[0].path, '/api/comic-screenplays/projects/p/media/actions');
  assert.equal(calls[0].options.signal, controller.signal);
  assert.equal(calls[0].options.timeoutMs, 0);
  assert.equal(JSON.parse(calls[0].options.body).action, 'voice_design');
  assert.equal(JSON.parse(calls[1].options.body).items[0].preview_text, '');
  assert.equal(calls.length, 2);
  assert.throws(() => actComicAssetMedia('../p', {}, {}, request), /无效/);
});
