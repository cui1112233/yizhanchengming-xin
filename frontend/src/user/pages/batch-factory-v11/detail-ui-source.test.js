import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const read = name => fs.readFileSync(path.join(here, name), 'utf8');

test('current book UI keeps progressive detail and direct prompt editing', () => {
  const source = read('BatchFactoryV11Workbench.jsx');
  assert.match(source, /Collapse/);
  for (const label of ['人物 Prompt', '场景 Prompt', '画面提示词', '本次提交预览']) {
    assert.match(source, new RegExp(label));
  }
});

test('asset prompt saves are scoped to the selected novel', () => {
  const workbench = read('BatchFactoryV11Workbench.jsx');
  const page = read('BatchFactoryV11UiPage.jsx');
  assert.match(workbench, /onSaveAssetPrompts\?\.\(selectedBook, type, items, drafts\)/);
  assert.match(page, /scope: `\$\{batch\.id\}:\$\{book\.id\}`/);
});

test('workbench follows the five-part novel production flow', () => {
  const source = read('BatchFactoryV11Workbench.jsx');
  const labels = ['原文与 Hook', '管理资产', '生成约束', '画面提示词', 'VIDEO 卡片'];
  let previous = -1;
  for (const label of labels) {
    const current = source.indexOf(label);
    assert.ok(current > previous, `missing or out-of-order module: ${label}`);
    previous = current;
  }
  assert.doesNotMatch(source, /label="道具 Prompt"/);
  assert.match(source, /defaultActiveKey=\{\['source', 'assets', 'constraints', 'visual-prompts', 'video-cards'\]\}/);
});

test('book constraints persist a 10 or 15 second Director split target', () => {
  const scoped = read('BatchFactoryV11ScopedSettings.jsx');
  assert.match(scoped, /目标切分时长/);
  assert.match(scoped, /maxVideoDuration/);
  assert.match(scoped, /value: 10, label: '10 秒'/);
  assert.match(scoped, /value: 15, label: '15 秒'/);
});

test('book list exposes server status as a color indicator instead of a status tag', () => {
  const source = read('BatchFactoryV11Workbench.jsx');
  const bookList = source.slice(source.indexOf("id: 'book-list'"), source.indexOf("id: 'book-workbench'"));
  assert.match(bookList, /className="bf11-book-status"/);
  assert.match(bookList, /aria-label=\{`状态：\$\{book\.status/);
  assert.doesNotMatch(bookList, /<Tag color=\{statusTone\[book\.status\]\}>\{book\.status/);
});

test('batch tools contain production progress and approved merge timing UI', () => {
  const source = read('BatchFactoryV11Workbench.jsx');
  for (const label of ['视频生成进度', '跟随音频时长', 'TTS 只测时', '合并待合并', '上传待上传']) {
    assert.match(source, new RegExp(label));
  }
  assert.match(source, /1\.7/);
});

test('V11 exposes a real TTS measurement action instead of only showing a duration field', () => {
  const page = read('BatchFactoryV11UiPage.jsx');
  const workbench = read('BatchFactoryV11Workbench.jsx');
  assert.match(page, /textToSpeech/);
  assert.match(page, /measureH3Audio/);
  assert.match(page, /onMeasureAudio/);
  assert.match(workbench, /读取 TTS 配音时长/);
});

test('showcase data represents a full 100-book batch', async () => {
  const module = await import('./showcaseData.js');
  assert.equal(module.SHOWCASE_BATCH.count, 100);
  assert.equal(module.SHOWCASE_BOOKS.length, 100);
});
