import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const root = path.dirname(fileURLToPath(import.meta.url));

function source(relativePath) {
  return fs.readFileSync(path.join(root, relativePath), 'utf8');
}

test('removes Novel Panel from the user application entry points', () => {
  for (const relativePath of ['App.jsx', '../shared/layouts/UserLayout.jsx', 'pages/HomePage.jsx']) {
    const contents = source(relativePath);
    assert.doesNotMatch(contents, /novel-panel|NovelPanel|小说面板/);
  }
});
