import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';

const appPath = new URL('./App.jsx', import.meta.url);

test('UserApp defines safe idle preloading for the three slow workbench modules', async () => {
  const source = await fs.readFile(appPath, 'utf8');
  assert.match(source, /preloadWorkbenchPages/);
  assert.match(source, /NovelFetchPage/);
  assert.match(source, /NovelPanelPage/);
  assert.match(source, /BatchFactoryWorkbenchPage/);
  assert.match(source, /Promise\.allSettled/);
});

test('workbench pages keep their asynchronous loading boundaries', async () => {
  const [novelFetch, novelPanel, batchFactory] = await Promise.all([
    fs.readFile(new URL('./pages/NovelFetchPage.jsx', import.meta.url), 'utf8'),
    fs.readFile(new URL('./pages/NovelPanelPage.jsx', import.meta.url), 'utf8'),
    fs.readFile(new URL('./pages/BatchFactoryWorkbenchPage.jsx', import.meta.url), 'utf8')
  ]);
  assert.match(novelFetch, /src=\{`\/batch-rewrite\/v2\?theme=\$\{effectiveTheme\}`\}/);
  assert.match(novelFetch, /qiantie-theme-sync/);
  assert.match(novelPanel, /NovelPanelReactPage/);
  assert.doesNotMatch(novelPanel, /iframe|workbenchSrc/);
  assert.match(batchFactory, /listBatchSummaries\(\)/);
});

test('novel fetch canonical entry keeps the public V78 visual baseline during React parity work', async () => {
  const app = await fs.readFile(appPath, 'utf8');
  const novelFetch = await fs.readFile(new URL('./pages/NovelFetchPage.jsx', import.meta.url), 'utf8');
  assert.match(app, /const loadNovelFetchPage = \(\) => import\('\.\/pages\/NovelFetchPage'\)/);
  assert.match(novelFetch, /effectiveTheme/);
  assert.match(novelFetch, /onLoad=/);
});
