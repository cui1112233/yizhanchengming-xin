import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');

test('asset popup uses durable book-asset APIs instead of draft prompts', () => {
  assert.match(source, /listBookAssets/);
  assert.match(source, /createBookAsset/);
  assert.match(source, /updateBookAsset/);
  assert.doesNotMatch(source, /kind:\s*'asset-prompt'/);
});

test('asset popup uploads real image bytes and manages durable image versions', () => {
  assert.match(source, /uploadBookAssetImage/);
  assert.match(source, /listBookAssetImages/);
  assert.match(source, /setPrimaryBookAssetImage/);
  assert.match(source, /上传图片版本/);
  assert.match(source, /切换为主图/);
  assert.doesNotMatch(source, /https:\/\/images\.example/);
});

test('SD plain-text storyboards fall back to their final prompt when structured asset references are absent', () => {
  const start = source.indexOf('function storyboardAssetDefaults(');
  const end = source.indexOf('\nfunction readStoryboardAssetSelection(', start);
  const helper = start >= 0 && end > start ? source.slice(start, end) : '';
  assert.match(helper, /storyboardAssetNamesFromPrompt/);
  assert.match(helper, /draft\.final_prompt/);
  assert.match(helper, /promptReferences/);
});

test('prompt panel renders image-backed mentions and opens the picker in edit mode', () => {
  assert.match(source, /batch-factory-prompt-rich-view/);
  assert.match(source, /batch-factory-prompt-editor-rich-wrap/);
  assert.match(source, /className="batch-factory-prompt-rich-view is-editing"[\s\S]{0,220}contentEditable/);
  assert.match(source, /renderStoryboardPromptHTML/);
  assert.match(source, /data-fallback/);
  assert.match(source, /contentEditableTextOffset/);
  assert.match(source, /batch-factory-asset-mention-menu/);
  assert.match(source, /aria-label="可编辑分镜视频提示词"/);
  assert.match(source, /assetThumbs\[String\(asset\.id\)\]/);
});
