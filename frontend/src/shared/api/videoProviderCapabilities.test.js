import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import contract from '../../user/pages/batch-factory-v11/videoProviderCapabilities.json' with { type: 'json' };

const FRONTEND_ROOT = path.resolve(import.meta.dirname, '..', '..', '..');

test('videoProviderCapabilities.json is the single provider list for the UI', () => {
  const ids = contract.providers.map(item => item.id);
  for (const id of ['personal_api', 'doubao_local_executor', 'autodl_comfyui', 'yfai_seedance']) {
    assert.ok(ids.includes(id), `missing provider: ${id}`);
  }
  assert.equal(new Set(ids).size, ids.length, 'provider ids must be unique');
  // 后端已移除 vidu_image_to_video；契约文件不能再把它带回来。
  assert.ok(!ids.includes('vidu_image_to_video'));
});

test('settings drawer and admin catalog read the same provider contract', () => {
  const drawer = fs.readFileSync(path.join(FRONTEND_ROOT, 'src/user/pages/batch-factory-v11/BatchFactoryV11SettingsDrawers.jsx'), 'utf8');
  const catalog = fs.readFileSync(path.join(FRONTEND_ROOT, 'src/admin/pages/ShuihuoModelCatalogPage.jsx'), 'utf8');
  assert.match(drawer, /videoProviderCapabilities\.json/);
  assert.match(catalog, /videoProviderCapabilities\.json/);
});
