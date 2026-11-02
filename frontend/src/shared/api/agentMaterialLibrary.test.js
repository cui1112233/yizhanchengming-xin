import assert from 'node:assert/strict';
import test from 'node:test';
import { insertAgentMaterialLibraryItem, listAgentMaterialLibraryItems, saveAgentMaterialLibraryItem } from './agent.js';

test('material library calls keep folder, source asset, and target canvas explicit', async () => {
  const calls = [];
  const request = async (path, options) => {
    calls.push({ path, options });
    return {};
  };
  await listAgentMaterialLibraryItems('folder /1', { kind: 'image', query: ' 夜景 ' }, request);
  await saveAgentMaterialLibraryItem('folder /1', { projectId: 'source project', assetId: 'asset/1', tags: ['夜景'] }, request);
  await insertAgentMaterialLibraryItem('target project', 'item/1', request);
  assert.deepEqual(calls, [
    { path: '/api/agent/material-library/folders/folder%20%2F1/items?kind=image&q=%E5%A4%9C%E6%99%AF', options: { suppressGlobalError: true } },
    { path: '/api/agent/material-library/folders/folder%20%2F1/items', options: { method: 'POST', body: '{"projectId":"source project","assetId":"asset/1","tags":["夜景"]}', suppressGlobalError: true } },
    { path: '/api/agent/tasks/target%20project/material-library/items/item%2F1/insert', options: { method: 'POST', suppressGlobalError: true } }
  ]);
});
