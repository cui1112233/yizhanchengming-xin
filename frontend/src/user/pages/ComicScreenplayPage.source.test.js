import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'ComicScreenplayPage.jsx'), 'utf8');

test('项目库入口没有可点击空壳', () => {
  assert.match(source, /历史<\/Button>|>历史<\/span>/);
  assert.match(source, /setHistoryOpen\(true\)/);
  assert.match(source,/listSharedComics/);
  assert.match(source,/scope==='team'/);
  assert.match(source,/canGenerate = !!access\?\.generate/);
  assert.match(source, /href="\/agent\/chat"/);
  assert.doesNotMatch(source, /<Button[^>]*icon=\{<History[^>]*>旧版<\/Button>/);
});

test('剧本正文可编辑、保存新版本并恢复历史版本', () => {
  assert.match(source, /saveComicScreenplayDocument/);
  assert.match(source, /restoreComicScreenplayDocument/);
  assert.match(source, /value=\{documentDraft\}/);
  assert.match(source, /保存新版本/);
  assert.match(source, /恢复此版本/);
});

test('生成后的分镜条目提供真实编辑闭环而不是数量展示', () => {
  assert.match(source, /createComicStoryboardShot/);
  assert.match(source, /updateComicStoryboardShot/);
  assert.match(source, /splitComicStoryboardShot/);
  assert.match(source, /mergeComicStoryboardShot/);
  assert.match(source, /deleteComicStoryboardShot/);
  assert.match(source, />在后面新增</);
  assert.match(source, />拆分</);
  assert.match(source, />与下一条合并</);
  assert.match(source, /删除这条分镜/);
  assert.match(source, /storyboard-shot-card/);
});
