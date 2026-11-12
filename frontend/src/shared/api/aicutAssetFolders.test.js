import test from 'node:test';
import assert from 'node:assert/strict';
import { createAssetFolder, deleteAssetFolder, listAssetFolders, moveAssetToFolder, renameAssetFolder } from './aicutAssetFolders.js';

test('folders and moves persist through project API, with empty folder meaning root', async () => {
  const calls = [];
  const request = async (path, options = {}) => { calls.push([path, options]); return { folder: { id: 'f' }, folders: [{ id: 'f' }], ok: true }; };
  assert.deepEqual(await createAssetFolder('p', '片头', request), { id: 'f' });
  assert.deepEqual(await listAssetFolders('p', request), [{ id: 'f' }]);
  await renameAssetFolder('p', 'f', '改名', request);
  await moveAssetToFolder('p', 'a', 'f', request);
  await moveAssetToFolder('p', 'a', '', request);
  await deleteAssetFolder('p', 'f', request);
  assert.equal(calls[0][0], '/api/ai-cut/projects/p/asset-folders');
  assert.equal(JSON.parse(calls[0][1].body).name, '片头');
  assert.equal(calls[3][0], '/api/ai-cut/projects/p/assets/a/folder');
  assert.deepEqual(JSON.parse(calls[4][1].body), { folderId: '' });
  assert.equal(calls[5][1].method, 'DELETE');
  await assert.rejects(() => listAssetFolders('../p', request), /无效/);
});
