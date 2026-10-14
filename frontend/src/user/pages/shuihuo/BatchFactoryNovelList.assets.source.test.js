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

test('asset editor exposes and persists the four character structure fields', () => {
  assert.match(source, /正式人物名单/);
  assert.match(source, /人物事实/);
  assert.match(source, /人物外形/);
  assert.match(source, /人物关系/);
  assert.match(source, /formalName: editing\.formalName/);
  assert.match(source, /facts: editing\.facts/);
  assert.match(source, /appearance: editing\.appearance/);
  assert.match(source, /relationships: editing\.relationships/);
});

test('prompt panel renders image-backed mentions and opens the picker in edit mode', () => {
  assert.match(source, /batch-factory-prompt-rich-view/);
  assert.match(source, /batch-factory-prompt-editor-rich-wrap/);
  assert.match(source, /className="batch-factory-prompt-rich-view is-editing"[\s\S]{0,220}contentEditable/);
  assert.match(source, /renderStoryboardPromptHTML/);
  assert.match(source, /data-fallback/);
  assert.match(source, /compactUnifiedAssetDescriptions\(withStoryboardAssetMentions\(prompt, book, video\), book\)/g);
  assert.match(source, /function withStoryboardAssetMentions\(prompt, book, video\)/);
  assert.match(source, /return stripStandaloneAssetMentionLine\(text, book\)/);
  assert.doesNotMatch(source, /return `\$\{text\}\\n资产：\$\{mentions\}`/);
  assert.match(source, /contentEditableTextOffset/);
  assert.match(source, /batch-factory-asset-mention-menu/);
  assert.match(source, /aria-label="可编辑分镜视频提示词"/);
  assert.match(source, /assetThumbs\[String\(asset\.id\)\]/);
});

test('batch card keeps the prompt and asset mentions in one content layer', () => {
  assert.match(source, /batch-factory-prompt-entry-rich-view/);
  assert.doesNotMatch(source, /<small title=\{assetPromptText\}>/);
});

test('batch card keeps rich prompt text visible after legacy span hiding rules', () => {
  const css = fs.readFileSync(path.join(here, '..', 'shuihuo-production.css'), 'utf8');
  assert.match(css, /batch-factory-prompt-entry-rich-view>span\{display:inline!important;/);
});

test('prompt display removes the legacy standalone asset row', () => {
  assert.match(source, /function stripStandaloneAssetMentionLine\(prompt, book\)/);
  assert.match(source, /const match = line\.match/);
  assert.match(source, /remainder\.replace/);
});

test('prompt editor provides cancel editing and preserves an entry snapshot', () => {
  assert.match(source, /editingSnapshot/);
  assert.match(source, /取消编辑/);
  assert.match(source, /setVideoPrompt\(editingSnapshot\.videoPrompt\)/);
});

test('prompt editor opens asset mentions from an explicit at-key trigger', () => {
  assert.match(source, /onKeyUp=\{handleEditablePromptKeyUp\}/);
  assert.match(source, /onKeyDown=\{handleEditablePromptKeyDown\}/);
  assert.match(source, /function handleEditablePromptKeyUp\(event\)/);
  assert.match(source, /function handleEditablePromptKeyDown\(event\)/);
  assert.doesNotMatch(source, /asset\?\.id && imageUrl && String\(assetName\(asset\)/);
});

test('asset picker shows a kind fallback when an asset has no image', () => {
  assert.match(source, /batch-factory-asset-mention-fallback/);
  assert.match(source, /暂无可用资产，请先在资产设置添加人物、场景或道具/);
  assert.match(source, /asset\.kind === 'character' \? '人' : asset\.kind === 'scene' \? '景' : '物'/);
});

test('asset picker is positioned below the active caret', () => {
  assert.match(source, /getBoundingClientRect/);
  assert.match(source, /style=\{\{ left: assetMentionMenu\.left, top: assetMentionMenu\.top \}\}/);
  assert.match(source, /const trailingMatch = value\.match/);
});

test('asset chips hide the trigger @ while preserving the underlying mention token', () => {
  assert.match(source, /data-asset-mention=\"\$\{escapePromptHTML\(asset\.mention\)\}\"/);
  assert.match(source, /escapePromptHTML\(assetName\(asset\)\)/);
  assert.match(source, /\{assetName\(asset\)\}/);
  assert.match(source, /function contentEditablePromptText\(root\)/);
  assert.match(source, /node\.dataset\?\.assetMention/);
  assert.match(source, /const nextValue = contentEditablePromptText\(root\)/);
  assert.match(source, /contentEditable=\{false\} className=\{`batch-factory-asset-mention/);
  assert.match(source, /contenteditable=\"false\" data-asset-mention=/);
});

test('prompt caret offset tolerates a missing editor root', () => {
  const start = source.indexOf('function contentEditableTextOffset(root)');
  const end = source.indexOf('\nfunction contentEditablePromptText(root)', start);
  const helper = start >= 0 && end > start ? source.slice(start, end) : '';
  assert.match(helper, /if \(!root\)/);
  assert.match(helper, /if \(!selection\?\.rangeCount \|\| !root\.contains\(selection\.anchorNode\)\)/);
});
