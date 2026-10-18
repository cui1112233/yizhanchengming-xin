import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryAiReasoningModal.jsx'), 'utf8');

test('uses the dedicated system-preset slots for comic settings', () => {
  assert.match(source, /batch\.comic-assets/);
  assert.match(source, /batch\.comic-page/);
  assert.match(source, /comic_quality/);
  assert.match(source, /comic_restriction/);
  assert.match(source, /comic_negative/);
  assert.match(source, /comicAssets\s*=\s*useMemo/);
  assert.match(source, /comicPrompt\s*=\s*useMemo/);
  assert.doesNotMatch(source, /key:'comic',label:'漫画设置'/);
});

test('keeps comic assets and constraints when the comic page preset changes', () => {
  assert.match(source, /const selectComicPrompt=/);
  assert.match(source, /\.\.\.current\.comic,\.\.\.selection,enabled:true/);
});
