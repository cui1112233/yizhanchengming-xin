import test from 'node:test';
import assert from 'node:assert/strict';

import { getMediaAsset } from './agentApi.js';

test('getMediaAsset requests the private media preview endpoint', async () => {
  let requested = '';
  const fetchImpl = async url => {
    requested = url;
    return {
      ok: true,
      status: 200,
      json: async () => ({ asset: { id: 'asset_1', media_type: 'image' }, preview_url: 'https://signed.example/a.png' }),
    };
  };
  const result = await getMediaAsset('asset_1', { fetchImpl });
  assert.equal(requested, '/api/media/assets/asset_1');
  assert.equal(result.asset.id, 'asset_1');
  assert.equal(result.preview_url, 'https://signed.example/a.png');
});

test('getMediaAsset rejects an empty id before making a request', async () => {
  await assert.rejects(() => getMediaAsset('', { fetchImpl: async () => { throw new Error('must not fetch'); } }), /media asset id is required/);
});
