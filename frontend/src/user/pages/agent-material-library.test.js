import assert from 'node:assert/strict';
import test from 'node:test';
import { materialLibraryRows } from './agent-material-library.js';

test('material library rows keep media kind, tags, and source provenance', () => {
  assert.deepEqual(materialLibraryRows([{ id: 'item-1', folderId: 'folder-1', sourceProjectId: 'project-a', sourceAssetId: 'asset-a', kind: 'image', name: '女主.png', version: 3, tags: ['女主', '古装'] }]), [{ id: 'item-1', folderId: 'folder-1', sourceProjectId: 'project-a', sourceAssetId: 'asset-a', kind: 'image', name: '女主.png', version: 3, tags: ['女主', '古装'], sourceLabel: '项目 project-a' }]);
});
