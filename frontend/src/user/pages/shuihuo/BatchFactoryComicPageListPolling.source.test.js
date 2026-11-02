import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');
const comicList = source.slice(source.indexOf('function InlineComicLibrary'), source.indexOf('const COMIC_PAGE_GENERATION_LIMIT'));

test('comic page list refresh is keyed by stable book ID, not a replaced book object', () => {
  assert.match(comicList, /listComicPages\(batchId, book\.id, \{ silent: true, suppressGlobalError: true \}\)[\s\S]*?\}, \[batchId, book\?\.id, refreshNonce\]\);/);
  assert.doesNotMatch(comicList, /\}, \[batchId, book, refreshNonce\]\);/);
});
