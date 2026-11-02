import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

test('material library drawer keeps save and insert separate from current project assets', async () => {
  const source = await readFile(new URL('./AgentMaterialLibraryDrawer.jsx', import.meta.url), 'utf8');
  assert.match(source, /保存当前素材/);
  assert.match(source, /放入当前画布/);
  assert.match(source, /来源项目/);
  assert.match(source, /listAgentMaterialLibraryFolders/);
});
