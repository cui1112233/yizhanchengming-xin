import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';

test('keeps the legacy workbench fallback while exposing an explicit React trial entry', async () => {
  const source = await fs.readFile(new URL('../NovelPanelPage.jsx', import.meta.url), 'utf8');
  assert.match(source, /searchParams\.get\('react'\)/);
  assert.match(source, /NovelPanelReactPage/);
  assert.match(source, /workbenchSrc/);
});
