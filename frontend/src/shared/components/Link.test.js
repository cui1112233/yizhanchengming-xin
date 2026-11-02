import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';

const linkPath = new URL('./Link.jsx', import.meta.url);

test('Link has a one-time navigation intent hook without changing click navigation', async () => {
  const source = await fs.readFile(linkPath, 'utf8');
  assert.match(source, /onIntent/);
  assert.match(source, /onMouseEnter=\{triggerIntent\}/);
  assert.match(source, /onFocus=\{triggerIntent\}/);
  assert.match(source, /window\.history\.pushState/);
});
