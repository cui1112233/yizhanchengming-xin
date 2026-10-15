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
  assert.match(novelFetch, /export function LegacyNovelFetchPage/);
  assert.match(novelFetch, /知识库/);
  assert.match(novelFetch, /处理规则/);
  assert.doesNotMatch(novelFetch, /batch-rewrite\/v2|<iframe/);
  assert.match(novelPanel, /NovelPanelReactPage/);
  assert.doesNotMatch(novelPanel, /iframe|workbenchSrc/);
  assert.match(batchFactory, /listBatchSummaries\(\)/);
});

test('novel fetch uses the complete React workbench as its public entry', async () => {
  const app = await fs.readFile(appPath, 'utf8');
  const novelFetch = await fs.readFile(new URL('./pages/NovelFetchPage.jsx', import.meta.url), 'utf8');
  assert.match(app, /const loadNovelFetchPage = \(\) => import\('\.\/pages\/NovelFetchPage'\)/);
  assert.match(app, /'\/novel-fetch-react': NovelFetchPage/);
  assert.match(novelFetch, /export function NovelFetchPage/);
  assert.match(novelFetch, /return <LegacyNovelFetchPage \/>/);
  assert.doesNotMatch(novelFetch, /effectiveTheme|onLoad=/);
});

test('all novel fetch entry routes use the same migrated React source', async () => {
  const app = await fs.readFile(appPath, 'utf8');
  assert.match(app, /'\/novel-fetch': NovelFetchPage/);
  assert.match(app, /'\/novel-fetch-react': NovelFetchPage/);
  assert.match(app, /'\/novel-fetch-workshop': NovelFetchPage/);
});

test('novel fetch React source keeps the V78 processing controls', async () => {
  const source = await fs.readFile(new URL('./pages/NovelFetchPage.jsx', import.meta.url), 'utf8');
  for (const marker of ['批量原文改文系统', 'V78', '批量输入', '本次处理设置', 'AI1', 'AI5', '敏感处理', '启用敏感词 AI 修复']) {
    assert.match(source, new RegExp(marker));
  }
  assert.doesNotMatch(source, /iframe|batch-rewrite\/index\.html/);
});
