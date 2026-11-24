import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const componentPath = new URL('./InlineTextBookPages.jsx', import.meta.url);
const novelListPath = new URL('./BatchFactoryNovelList.jsx', import.meta.url);

test('inline text workbench keeps assigned copy and rendered image on one page index', () => {
  const source = fs.readFileSync(componentPath, 'utf8');
  assert.match(source, /listTextPages\(batchId, book\.id/);
  assert.match(source, /const \[selectedIndex, setSelectedIndex\] = useState\(0\)/);
  assert.match(source, /item\.id === page\?\.activeArtifactId/);
  assert.match(source, /<AuthenticatedTextPageImage/);
  assert.match(source, />页面内容</);
  assert.match(source, />文字页</);
  assert.match(source, /page\?\.text/);
  assert.match(source, /onOpen\?\.\(page\.id\)/);
  assert.match(source, /'上一文字页'/);
  assert.match(source, /'下一文字页'/);
});

test('text rows use the shared inline component instead of a media-library placeholder button', () => {
  const source = fs.readFileSync(novelListPath, 'utf8');
  assert.match(source, /<InlineTextBookPages[\s\S]*onOpen=\{pageId => openTextBook\(book, 'pages', pageId\)\}/);
  assert.doesNotMatch(source, /productionMode === "text" \? <Button aria-label="打开文本媒体库"/);
});
