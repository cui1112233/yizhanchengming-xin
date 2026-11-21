import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const app = fs.readFileSync(new URL('./App.jsx', import.meta.url), 'utf8');
const preload = fs.readFileSync(new URL('./workbenchPreload.js', import.meta.url), 'utf8');

test('application shell loads Novel Fetch through its feature-module entry', () => {
  assert.match(app, /import\('\.\.\/modules\/novel-fetch'\)/);
  assert.doesNotMatch(app, /import\('\.\/pages\/NovelFetchPage'\)/);
  assert.match(preload, /import\('\.\.\/modules\/novel-fetch'\)/);
});

test('Novel Fetch module entry owns the temporary legacy-page adapter', () => {
  const moduleEntry = fs.readFileSync(new URL('../modules/novel-fetch/index.jsx', import.meta.url), 'utf8');
  assert.match(moduleEntry, /NovelFetchModule/);
  const moduleSource = fs.readFileSync(new URL('../modules/novel-fetch/NovelFetchModule.jsx', import.meta.url), 'utf8');
  assert.match(moduleSource, /pages\/NovelFetchPage/);
});
