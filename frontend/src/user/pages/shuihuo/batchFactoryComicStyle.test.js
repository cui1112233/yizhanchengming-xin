import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

import * as comicStyleModule from './batchFactoryComicStyle.js';
import {
  comicStyleCoverFlowStyle,
  comicStyleIndexFromDrag,
  comicStylePreviewReference,
  comicStyleRelativeOffset,
  comicStyleSelection,
  comicStyleWheelDirection,
  normalizeComicReference,
  wrapComicStyleIndex
} from './batchFactoryComicStyle.js';

test('cover flow wraps at both ends and uses the shortest cyclic offset', () => {
  assert.equal(wrapComicStyleIndex(-1, 5), 4);
  assert.equal(wrapComicStyleIndex(5, 5), 0);
  assert.equal(comicStyleRelativeOffset(4, 0, 5), -1);
  assert.equal(comicStyleRelativeOffset(0, 4, 5), 1);
});

test('cover flow drag changes one style only after crossing the threshold', () => {
  assert.equal(comicStyleIndexFromDrag(2, -50, 5), 3);
  assert.equal(comicStyleIndexFromDrag(2, 50, 5), 1);
  assert.equal(comicStyleIndexFromDrag(2, 20, 5), 2);
});

test('cover flow ignores vertical page scrolling and accepts horizontal wheel gestures', () => {
  assert.equal(comicStyleWheelDirection(0, 120, false), 0);
  assert.equal(comicStyleWheelDirection(80, 20, false), 1);
  assert.equal(comicStyleWheelDirection(-80, 20, false), -1);
  assert.equal(comicStyleWheelDirection(0, 120, true), 1);
});

test('cover flow gives the centered card priority and mirrors both sides', () => {
  assert.deepEqual(comicStyleCoverFlowStyle(0), {
    '--comic-style-offset': 0,
    '--comic-style-distance': 0,
    '--comic-style-brightness': 1,
    '--comic-style-rotation': '0deg',
    '--comic-style-scale': 1.08,
    '--comic-style-translate-x': '0px',
    '--comic-style-translate-z': '90px',
    opacity: 1,
    pointerEvents: 'auto',
    zIndex: 20
  });
  assert.equal(comicStyleCoverFlowStyle(-2)['--comic-style-rotation'], '42deg');
  assert.equal(comicStyleCoverFlowStyle(-2)['--comic-style-translate-x'], '-264px');
  assert.equal(comicStyleCoverFlowStyle(2)['--comic-style-rotation'], '-42deg');
  assert.equal(comicStyleCoverFlowStyle(2)['--comic-style-brightness'], 0.78);
  assert.equal(comicStyleCoverFlowStyle(2)['--comic-style-scale'], 0.9);
});

test('cover flow hides distant cards so they cannot block the visible controls', () => {
  assert.equal(comicStyleCoverFlowStyle(5).pointerEvents, 'none');
  assert.equal(comicStyleCoverFlowStyle(5).opacity, 0);
});

test('cover flow downloads only the previews visible around the active card', () => {
  assert.equal(typeof comicStyleModule.comicStyleShouldLoadPreview, 'function');
  assert.equal(comicStyleModule.comicStyleShouldLoadPreview(0, 0, 20), true);
  assert.equal(comicStyleModule.comicStyleShouldLoadPreview(3, 0, 20), true);
  assert.equal(comicStyleModule.comicStyleShouldLoadPreview(4, 0, 20), false);
  assert.equal(comicStyleModule.comicStyleShouldLoadPreview(19, 0, 20), true);
});

test('comic style selection stores metadata without treating its preview as an AI reference', () => {
  const selected = comicStyleSelection({
    id: 'batch-comic-style-01',
    name: '顶级韩漫风格',
    slot: 'batch.comic-style',
    version: 2,
    previewUrl: '/api/presets/comic-style/batch-comic-style-01/preview'
  });

  assert.deepEqual(selected, {
    presetId: 'batch-comic-style-01',
    presetName: '顶级韩漫风格',
    presetSlot: 'batch.comic-style',
    presetVersion: 2
  });
  assert.equal(Object.hasOwn(selected, 'previewUrl'), false);
});

test('style preview becomes the enabled default reference for character assets and pages', () => {
  assert.deepEqual(comicStylePreviewReference({
    id: 'batch-comic-style-01',
    name: '顶级韩漫风格',
    slot: 'batch.comic-style',
    version: 2,
    referenceUrl: '/api/local-executor/v1/batch-comic-style-previews/batch-comic-style-01',
    previewUrl: '/api/presets/comic-style/batch-comic-style-01/preview'
  }), {
    assetId: 'style:batch-comic-style-01',
    url: '/api/local-executor/v1/batch-comic-style-previews/batch-comic-style-01',
    previewUrl: '/api/presets/comic-style/batch-comic-style-01/preview',
    enabled: true,
    applyToAssets: true,
    applyToPages: true,
    source: 'style-library',
    style: {
      presetId: 'batch-comic-style-01',
      presetName: '顶级韩漫风格',
      presetSlot: 'batch.comic-style',
      presetVersion: 2
    }
  });
});

test('style selection preserves a manual upload and replaces a legacy scene reference', () => {
  assert.equal(typeof comicStyleModule.comicReferenceForSelectedStyle, 'function');
  const preset = { id: 'batch-comic-style-02', name: '都市彩漫', slot: 'batch.comic-style', version: 1, previewUrl: '/styles/02' };
  const manual = normalizeComicReference({ assetId: 'manual-1', url: '/manual/1', source: 'manual-upload' });
  assert.deepEqual(comicStyleModule.comicReferenceForSelectedStyle(manual, preset), manual);
  const legacyScene = normalizeComicReference({ assetId: 'scene:campus', url: '/scene/campus', source: 'scene-library' });
  assert.equal(comicStyleModule.comicReferenceForSelectedStyle(legacyScene, preset).assetId, 'style:batch-comic-style-02');
});

test('existing comic settings restore a missing style reference without replacing a manual upload', () => {
  assert.equal(typeof comicStyleModule.comicReferenceNeedsStyleDefault, 'function');
  const preset = { id: 'batch-comic-style-02', name: '都市彩漫', previewUrl: '/styles/02' };
  assert.equal(comicStyleModule.comicReferenceNeedsStyleDefault(null, preset), true);
  assert.equal(comicStyleModule.comicReferenceNeedsStyleDefault({ assetId: 'scene:old', source: 'scene-library' }, preset), true);
  assert.equal(comicStyleModule.comicReferenceNeedsStyleDefault({ assetId: 'style:batch-comic-style-02', source: 'style-library' }, preset), false);
  assert.equal(comicStyleModule.comicReferenceNeedsStyleDefault({ assetId: 'manual-1', source: 'manual-upload' }, preset), false);
});

test('picker removes the scene library and explains the style reference scopes', () => {
  const pickerSource = fs.readFileSync(new URL('./BatchFactoryComicStylePicker.jsx', import.meta.url), 'utf8');
  assert.match(pickerSource, /comicReferenceForSelectedStyle/);
  assert.doesNotMatch(pickerSource, /场景参考图库/);
  assert.match(pickerSource, /画风库图片会默认作为参考图/);
  assert.match(pickerSource, /用于资产图（仅人物）/);
});

test('uploaded global reference keeps independent asset and page scope switches', () => {
  assert.deepEqual(normalizeComicReference({ assetId: 'lea_reference', url: '/executor/ref', previewUrl: '/preview/ref', enabled: true, applyToAssets: false }), {
    assetId: 'lea_reference',
    url: '/executor/ref',
    previewUrl: '/preview/ref',
    enabled: true,
    applyToAssets: false,
    applyToPages: true
  });
});

test('comic style cover flow keeps every preview at its natural 3:4 height', () => {
  const stylesheet = fs.readFileSync(new URL('../shuihuo-production.css', import.meta.url), 'utf8');
  assert.match(stylesheet, /\.batch-factory-comic-style-image\s*\{[^}]*aspect-ratio:\s*3\s*\/\s*4/s);
});

test('comic style names stay readable on the white preview cards', () => {
  const stylesheet = fs.readFileSync(new URL('../shuihuo-production.css', import.meta.url), 'utf8');
  assert.match(stylesheet, /\.batch-factory-comic-style-copy b\s*\{[^}]*color:\s*var\(--text-primary\)[^}]*font-size:\s*13px/s);
});

test('authenticated preview loader accepts built-in comic scene previews', () => {
  const source = fs.readFileSync(new URL('../../../shared/api/batchFactoryV11.js', import.meta.url), 'utf8');
  assert.match(source, /COMIC_REFERENCE_SCENE_PREVIEW_PATTERN/);
  assert.match(source, /comic-reference-scene/);
});

test('authenticated preview loader accepts administrator-uploaded comic material previews', () => {
  const source = fs.readFileSync(new URL('../../../shared/api/batchFactoryV11.js', import.meta.url), 'utf8');
  assert.match(source, /COMIC_MATERIAL_PREVIEW_PATTERN/);
  assert.match(source, /comic-material/);
  assert.match(source, /COMIC_MATERIAL_PREVIEW_PATTERN\.test\(path\)/);
});
