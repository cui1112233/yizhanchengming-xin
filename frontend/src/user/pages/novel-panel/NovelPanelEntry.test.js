import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';

test('uses the React page as the only novel panel entry', async () => {
  const source = await fs.readFile(new URL('../NovelPanelPage.jsx', import.meta.url), 'utf8');
  assert.match(source, /NovelPanelReactPage/);
  assert.doesNotMatch(source, /iframe|workbenchSrc|legacy/);
});
