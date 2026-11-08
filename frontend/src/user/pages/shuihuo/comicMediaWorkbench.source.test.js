import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const jsx = readFileSync(new URL('./BatchFactoryNovelList.jsx', import.meta.url), 'utf8');
const css = readFileSync(new URL('../shuihuo-production.css', import.meta.url), 'utf8');
const panel = jsx.slice(jsx.indexOf('function ComicMediaPanel'), jsx.indexOf('function MediaVersionPanel'));
const workbenchCss = css.slice(
  css.indexOf('.batch-factory-media-library-modal.is-comic'),
  css.indexOf('.batch-factory-comic-style-picker')
);

test('comic media workbench keeps the primary comic alone on the left', () => {
  assert.match(panel, /batch-factory-comic-media-workbench/);
  assert.match(panel, /batch-factory-comic-media-preview[\s\S]*AuthenticatedComicPageImage/);
  const preview = panel.slice(
    panel.indexOf('batch-factory-comic-media-preview'),
    panel.indexOf('batch-factory-comic-media-tools')
  );
  assert.doesNotMatch(preview, /上传参考图|漫画合集|ComicSourceContent/);
});

test('comic media preview constrains the full vertical comic inside the left canvas', () => {
  assert.match(css, /\.batch-factory-comic-media-preview\{[^}]*position:relative/);
  assert.match(css, /\.batch-factory-comic-media-preview>\.ant-image\{[^}]*position:absolute[^}]*inset:0/);
  assert.match(css, /\.batch-factory-comic-media-preview>\.ant-image img[^}]*object-fit:contain/);
});

test('comic media workbench follows the active light or dark theme', () => {
  assert.match(workbenchCss, /\.batch-factory-media-library-modal\.is-comic \.ant-modal-content\{[^}]*background:var\(--sh-panel\)/);
  assert.match(workbenchCss, /\.batch-factory-comic-media-tools\{[^}]*background:var\(--sh-panel\)/);
  assert.match(workbenchCss, /\.batch-factory-comic-media-header\{[^}]*border-bottom:1px solid var\(--sh-line\)[^}]*background:var\(--sh-panel-strong\)/);
  assert.doesNotMatch(workbenchCss, /#10161e|#111820|#0d131a|#202a35|#edf2f7|#8d99a8/);
});

test('comic media controls live in the fixed right workbench', () => {
  assert.match(panel, /batch-factory-comic-media-tools[\s\S]*生成漫画合集[\s\S]*漫画页[\s\S]*漫画合集/);
  assert.match(panel, /const pageTab =[\s\S]*候选图[\s\S]*上传参考图[\s\S]*ComicSourceContent/);
  assert.match(panel, /batch-factory-comic-media-scroll[\s\S]*activeView === 'pages' \? pageTab : collectionTab/);
  assert.match(panel, /batch-factory-comic-media-footer[\s\S]*重新生成本页/);
  assert.match(css, /\.batch-factory-comic-media-tools\{[^}]*grid-template-rows:[^}]*minmax\(0,1fr\)[^}]*\}/);
  assert.match(css, /\.batch-factory-comic-media-scroll\{[^}]*overflow-y:auto/);
});

test('current-page regeneration reuses the existing comic page generator', () => {
  assert.match(panel, /onGeneratePage/);
  assert.match(panel, /onGeneratePage\?\.\(selectedVideoId/);
  assert.match(jsx, /<ComicMediaPanel[\s\S]*onGeneratePage=\{[\s\S]*generateBookComicPage/);
});

test('comic media workbench exposes AI source resegmentation', () => {
  assert.match(panel, /segmentComicSource/);
  assert.match(panel, /重新分析分段/);
});

test('comic media header exposes real primary-image completion', () => {
  assert.match(panel, /readyComicPageCount/);
  assert.match(panel, /完成度 \$\{readyComicPageCount\}\/\$\{videos\.length\} 页/);
  assert.match(panel, /comicCompletionPercent/);
  assert.match(css, /batch-factory-comic-media-completion-track/);
});

test('comic collection range follows asynchronously loaded comic pages', () => {
  assert.match(panel, /collectionRangeTouched/);
  assert.match(panel, /!collectionRangeTouched && collectionEnd === 1 && videos\.length > 1/);
});

test('comic collection card downloads its ordered primary images as a ZIP', () => {
  assert.match(jsx, /downloadComicCollectionImageZip/);
  assert.match(panel, /下载图片/);
  assert.match(panel, /downloadCollectionImages\(collection\)/);
});

test('comic collection grid requests the server thumbnail before the original', () => {
  assert.match(panel, /collection\.items[\s\S]*thumbnailUrl=\{item\.thumbnailUrl\}[\s\S]*lazy/);
});
