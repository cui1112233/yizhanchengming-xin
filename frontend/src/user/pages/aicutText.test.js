import test from 'node:test';
import assert from 'node:assert/strict';
import { aicutCurvedTextLayout, aicutFontFaceCSS, aicutFontOptionsForAssets, aicutTextPreviewStyle, aicutTextResourcePreset, aicutTextRunStyleAt, normalizeAICutTextStyle, textAnimationFrame } from './aicutText.js';

test('文字样式会清理字体颜色和排版数值', () => {
  assert.deepEqual(
    normalizeAICutTextStyle({
      fontFamily: 'Microsoft YaHei', fontSize: 999, textColor: '#ffcc00', strokeColor: 'bad',
      strokeWidth: -1, backgroundColor: '#11223380', letterSpacing: 99, lineHeight: .2, textAlign: 'unknown',
      shadowColor: '#334455cc', shadowOffsetX: 99, shadowOffsetY: -99, shadowBlur: 99
    }),
    {
	  fontFamily: 'Microsoft YaHei', fontSize: 240, textOpacity: 1, textCurve: 0, textColor: '#ffcc00', strokeColor: '#000000',
	  strokeWidth: 0, backgroundColor: '#11223380', shadowColor: '#334455cc', shadowOffsetX: 40,
      shadowOffsetY: -40, shadowBlur: 20, letterSpacing: 40, lineHeight: .8, textAlign: 'center'
    }
  );
});

test('标题阴影和背景板生成真实浏览器样式', () => {
  const style = aicutTextPreviewStyle({ backgroundColor: '#112233cc', shadowColor: '#000000cc', shadowOffsetX: 4, shadowOffsetY: 6, shadowBlur: 8 });
  assert.equal(style.textShadow, '4px 6px 8px #000000cc');
  assert.equal(style.backgroundColor, '#112233cc');
  assert.equal(style.padding, '0.18em 0.35em');
});

test('small preview scales text, spacing, outlines and mixed text runs using the same canvas unit', () => {
  const unit = 'var(--aicut-preview-px, 1px)';
  const style = aicutTextPreviewStyle({ fontSize: 54, strokeWidth: 3, letterSpacing: 2, shadowOffsetX: 4, shadowOffsetY: 6, shadowBlur: 8 }, unit);
  assert.equal(style.fontSize, `calc(54 * ${unit})`);
  assert.equal(style.letterSpacing, `calc(2 * ${unit})`);
  assert.ok(style.WebkitTextStroke.startsWith(`calc(3 * ${unit})`));
  assert.ok(style.textShadow.startsWith(`calc(4 * ${unit}) calc(6 * ${unit}) calc(8 * ${unit})`));
  assert.equal(aicutTextRunStyleAt({ textRuns: [{ start: 0, end: 2, fontSize: 80 }] }, 1, unit).fontSize, `calc(80 * ${unit})`);
});

test('打字机和逐字高亮按片段进度逐字推进', () => {
  assert.deepEqual(textAnimationFrame('剪辑开始', 'typewriter', 500, 1000), { visibleText: '剪辑', highlightedCount: 0 });
  assert.deepEqual(textAnimationFrame('剪辑开始', 'word-highlight', 750, 1000), { visibleText: '剪辑开始', highlightedCount: 3 });
  assert.deepEqual(textAnimationFrame('剪辑开始', 'none', 10, 1000), { visibleText: '剪辑开始', highlightedCount: 0 });
});

test('用户字体保留素材引用并生成项目字体选项和 font-face', () => {
  const assets = [{ id: 'font-1', kind: 'font', name: 'title.ttf', fontFamily: '一战测试字体', url: '/font-1' }];
  assert.deepEqual(normalizeAICutTextStyle({ fontAssetId: 'font-1', fontFamily: '一战测试字体' }).fontAssetId, 'font-1');
  assert.deepEqual(aicutFontOptionsForAssets(assets).at(-1), { value: 'asset:font-1', label: '一战测试字体 · 用户字体' });
  const css = aicutFontFaceCSS(assets);
  assert.match(css, /@font-face/);
  assert.match(css, /一战测试字体/);
  assert.match(css, /\/font-1/);
});

test('曲线文字按统一弧度生成逐字位移与旋转并限制边界', () => {
  assert.equal(normalizeAICutTextStyle({ textCurve: 999 }).textCurve, 180);
  const glyphs = aicutCurvedTextLayout('曲线文字', 120, 60, 2);
  assert.equal(glyphs.length, 4);
  assert.ok(glyphs[1].translateY < 0);
  assert.ok(glyphs[0].rotate < 0);
  assert.ok(glyphs.at(-1).rotate > 0);
  assert.deepEqual(aicutCurvedTextLayout('直线', 0, 54, 0).map((item) => [item.translateY, item.rotate]), [[0, 0], [0, 0]]);
});

test('混排文字清理选区并按字符返回颜色字号粗斜体', () => {
  const style = normalizeAICutTextStyle({ textRuns: [
    { start: 1, end: 3, textColor: '#ff0000', fontSize: 80, bold: true, italic: true },
    { start: -4, end: 900, textColor: 'bad', fontSize: 999 }
  ] });
  assert.deepEqual(style.textRuns, [
    { start: 0, end: 512, textColor: '#ffffff', fontSize: 240, bold: false, italic: false },
    { start: 1, end: 3, textColor: '#ff0000', fontSize: 80, bold: true, italic: true }
  ]);
  assert.deepEqual(aicutTextRunStyleAt(style, 2), { color: '#ff0000', fontSize: '80px', fontWeight: 700, fontStyle: 'italic' });
  assert.deepEqual(aicutTextRunStyleAt(style, 4), { color: '#ffffff', fontSize: '240px', fontWeight: 400, fontStyle: 'normal' });
});

test('十二个文字模板都有稳定且可保存的真实排版预设', () => {
  const ids = ['text-basic-caption', 'text-bold-title', 'text-outline-caption', 'text-shadow-caption', 'text-panel-caption', 'text-word-highlight', 'text-news-bar', 'text-lower-third', 'text-comic', 'text-handwriting', 'text-vertical', 'text-gradient-title'];
  const presets = ids.map(aicutTextResourcePreset);
  assert.equal(presets.length, 12);
  assert.ok(presets.every((item) => item && item.fontSize >= 12));
  assert.equal(aicutTextResourcePreset('text-bold-title').fontSize, 82);
  assert.equal(aicutTextResourcePreset('text-panel-caption').backgroundColor, '#000000b8');
  assert.equal(aicutTextResourcePreset('text-news-bar').textAlign, 'left');
  assert.equal(aicutTextResourcePreset('text-handwriting').fontFamily, 'KaiTi');
  assert.equal(aicutTextResourcePreset('unknown'), null);
});
