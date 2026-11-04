import assert from 'node:assert/strict';
import test from 'node:test';
import { videoProviderForModel } from './videoProviderBinding.js';

test('uses the provider required by known video models and preserves the configured fallback otherwise', () => {
  assert.equal(videoProviderForModel('minimax-h3-video', 'personal_api'), 'autodl_comfyui');
  assert.equal(videoProviderForModel('custom-video-model', 'doubao_local_executor'), 'doubao_local_executor');
});

test('binds Seedance 2.0 official to YFAI instead of the legacy FanKe group', async () => {
  const { videoProviderForModel } = await import('./videoProviderBinding.js');
  assert.equal(videoProviderForModel('seedance-2-0-official', 'personal_api'), 'yfai');
});

test('binds FanKe concrete model selections and legacy bare entries to the FanKe engine', () => {
  assert.equal(videoProviderForModel('fanke-video-api::ft-video-v1-77e8ee7a636f15dac27b2ce6d6fcd746', 'personal_api'), 'yfai_seedance');
  assert.equal(videoProviderForModel('fanke-video-api', 'yfai'), 'yfai_seedance');
});
