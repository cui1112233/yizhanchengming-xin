import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');

test('comic prompt panel offers a separate all-pages regeneration action', () => {
  assert.match(source, /重新生成全部漫画提示词/);
  assert.match(source, /onRegenerateAllVisual/);
  assert.match(source, /onRegenerateAllVisual=\{\(\) => runBookStageAction\(promptBook, 'comic', 'force'\)\}/);
});

test('comic prompt regeneration polls the current book so completed pages appear incrementally', () => {
  assert.match(source, /pollComicPromptBook/);
  assert.match(source, /setInterval\(pollComicPromptBook/);
});
