import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'AssetsView.jsx'), 'utf8');

test('asset editor gives non-voice assets image model and aspect-ratio overrides that can inherit project defaults', () => {
  assert.match(source, /编辑资产[\s\S]*图片模型/);
  assert.match(source, /图片模型<Select allowClear value=\{editing\?\.imageModelId \?\? undefined\}/);
  assert.match(source, /placeholder="使用项目默认" options=\{imageModels\.map\(model => \(\{ value: model\.id, label: model\.name \}\)\)\}/);
  assert.match(source, /图片画幅<Select allowClear value=\{editing\?\.imageAspectRatio \?\? undefined\}/);
  assert.match(source, /\{ value: '16:9', label: '16:9' \}, \{ value: '9:16', label: '9:16' \}, \{ value: '1:1', label: '1:1' \}/);
  assert.match(source, /imageModelId: imageModelId \?\? null/);
  assert.match(source, /imageAspectRatio: imageAspectRatio \?\? null/);
});

test('asset editor hides image overrides for voice assets and rejects unavailable image models before saving', () => {
  assert.match(source, /editing\?\.category !== 'voice' \? <>[\s\S]*图片模型[\s\S]*图片画幅[\s\S]*<\/> : null/);
  assert.match(source, /const imageModels = useMemo\(\(\) => models\.filter\(model => model\.kind === 'image' && model\.enabled\), \[models\]\)/);
  assert.match(source, /editing\.imageModelId && !imageModels\.some\(model => model\.id === editing\.imageModelId\)/);
  assert.match(source, /当前图片模型不可用，请使用项目默认或选择已启用模型/);
});

test('asset generation accepts enabled per-asset overrides without a project default and rejects unresolved selections', () => {
  assert.match(source, /function hasUsableAssetImageModel\(asset, fallbackModelId, imageModels\)/);
  assert.match(source, /const selectedImageAssets = selectedAssetIDs[\s\S]*asset\.category !== 'voice'/);
  assert.match(source, /selectedImageAssets\.every\(asset => hasUsableAssetImageModel\(asset, generationConfig\.imageModelId, imageModels\)\)/);
  assert.doesNotMatch(source, /generationConfig\.imageModelId === undefined \|\| generationConfig\.imageModelId === null/);
});
