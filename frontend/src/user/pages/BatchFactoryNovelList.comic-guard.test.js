import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const source = fs.readFileSync(new URL('./shuihuo/BatchFactoryNovelList.jsx', import.meta.url), 'utf8');

test('guards comic page generation against duplicate clicks for the same page', () => {
  assert.match(source, /const comicGenerationInFlightRef = useRef\(new Set\(\)\);/);
  assert.match(source, /comicGenerationInFlightRef\.current\.has\(requestKey\)/);
  assert.match(source, /comicGenerationInFlightRef\.current\.add\(requestKey\)/);
  assert.match(source, /comicGenerationInFlightRef\.current\.delete\(requestKey\)/);
});

test('refreshes comic page records after generation completes', () => {
  assert.match(source, /const \[comicPageRefreshNonce, setComicPageRefreshNonce\] = useState\(0\);/);
  assert.match(source, /setComicPageRefreshNonce\(value => value \+ 1\)/);
  assert.match(source, /refreshNonce = 0/);
  assert.match(source, /\[batchId, book, refreshNonce\]/);
});
