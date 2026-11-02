import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8')
  + '\n' + fs.readFileSync(path.join(here, 'batchFactoryNovelListHelpers.js'), 'utf8');

test('asset popup uses durable book-asset APIs instead of draft prompts', () => {
  assert.match(source, /listBookAssets/);
  assert.match(source, /createBookAsset/);
  assert.match(source, /updateBookAsset/);
  assert.doesNotMatch(source, /kind:\s*'asset-prompt'/);
	assert.match(source, /source: productionMode === 'comic' \? 'comic' : 'manual'/);
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
  const end = source.indexOf('function readStoryboardAssetSelection(', start);
  const helper = start >= 0 && end > start ? source.slice(start, end) : '';
  assert.match(helper, /storyboardAssetNamesFromPrompt/);
  assert.match(helper, /draft\.final_prompt/);
  assert.match(helper, /promptReferences/);
});

test('asset generation stops waiting when the background task reports a real failure', () => {
  assert.match(source, /getBookAssetImageGenerationTask/);
  assert.match(source, /task\.status === 'failed'/);
  assert.match(source, /task\.error \|\| '资产图片生成失败'/);
  assert.match(source, /waitForBookAssetImages\(batchId, book\.id, assetIds, result\?\.taskId \|\| ''\)/);
});

test('asset generation keeps its terminal result visible in the asset panel', () => {
  const start = source.indexOf('async function generateImages(');
  const end = source.indexOf('\n  async function generateEditedAssetImage()', start);
  const handler = start >= 0 && end > start ? source.slice(start, end) : '';
  assert.match(handler, /setImageGenerationResult\(\{ type: 'warning', message: summary \}\)/);
  assert.match(handler, /setImageGenerationResult\(\{ type: 'success', message: summary \}\)/);
  assert.match(handler, /const failure = error\?\.message \|\| '图片生成失败'/);
  assert.match(handler, /setImageGenerationResult\(\{ type: 'error', message: failure \}\)/);
  assert.match(source, /imageGenerationResult \? <Alert type=\{imageGenerationResult\.type\} showIcon closable onClose=\{\(\) => setImageGenerationResult\(null\)\} message="资产图片生成结果" description=\{imageGenerationResult\.message\} \/> : null/);
});

test('asset editor persists its single name as the character formal name', () => {
  assert.match(source, /名称<Input/);
  assert.match(source, /人物事实/);
  assert.match(source, /人物外形/);
  assert.match(source, /人物关系/);
  assert.match(source, /formalName: editing\.name/);
  assert.doesNotMatch(source, /正式人物名单<Input/);
  assert.match(source, /facts: editing\.facts/);
  assert.match(source, /appearance: editing\.appearance/);
  assert.match(source, /relationships: editing\.relationships/);
});

test('asset editor uses one name as the character formal name and groups the remaining character details', () => {
  const css = fs.readFileSync(path.join(here, '..', 'shuihuo-production.css'), 'utf8');
  assert.match(source, /formalName: editing\.name/);
  assert.match(source, /className="batch-factory-asset-editor-character-fields"/);
  assert.doesNotMatch(source, /正式人物名单<Input/);
  assert.match(css, /\.batch-factory-asset-editor-character-fields \{ display: grid; grid-template-columns: repeat\(2, minmax\(0, 1fr\)\); gap: 10px; \}/);
});

test('batch asset editor keeps its image-model picker bound to the book image channel', () => {
  assert.match(source, /import \{ modelOptionsForProvider \} from '\.\/providerModelBinding';/);
  assert.match(source, /listAvailableModels\('image'\)/);
  assert.match(source, /modelOptionsForProvider\(imageModels, 'image', imageProvider\)/);
  assert.match(source, /batch-factory-asset-editor-config-row/);
  assert.match(source, /资产分类<\/span><Select[\s\S]{0,900}图片模型<\/span><Select[\s\S]{0,900}图片画幅<\/span><Select/);
  assert.match(source, /图片模型<\/span><Select allowClear value=\{editing\?\.imageModelId \|\| imageModelId \|\| undefined\}/);
  assert.match(source, /图片画幅<\/span><Select allowClear value=\{editing\?\.imageAspectRatio \|\| imageAspectRatio \|\| undefined\}/);
  assert.match(source, /imageModelId: editing\.imageModelId \|\| ''/);
  assert.match(source, /imageAspectRatio: editing\.imageAspectRatio \|\| ''/);
  assert.doesNotMatch(source, /if \(!imageModelId\) \{ message\.warning\('请先在引擎配置选择图片模型。'\); return false; \}/);
  assert.doesNotMatch(source, /disabled=\{panelBusy \|\| !imageModelId\}/);
});

test('asset generation settings use a compact hierarchy instead of three plain form boxes', () => {
  const css = fs.readFileSync(path.join(here, '..', 'shuihuo-production.css'), 'utf8');
  assert.match(source, /className="batch-factory-asset-editor-config-row"[\s\S]{0,420}className="batch-factory-asset-editor-config-field is-model"/);
  assert.match(css, /\.batch-factory-asset-editor-config-row \{ display: grid; grid-template-columns: 108px minmax\(0, 1fr\) 120px; align-items: end; gap: 8px; padding: 10px; border: 1px solid/);
  assert.match(css, /\.batch-factory-asset-editor-config-field > span \{ color: #8fa8c4; font-size: 11px; \}/);
});

test('asset settings use a 16:9 default while keeping explicit per-asset overrides', () => {
  assert.match(source, /const imageAspectRatio = '16:9'/);
  assert.match(source, /资产画幅未单独选择时固定为 16:9/);
  assert.match(source, /placeholder="默认 16:9"/);
  assert.match(source, /imageAspectRatio: imageAspectRatio \|\| '16:9'/);
  assert.match(source, /图片模型<\/span><Select allowClear value=\{editing\?\.imageModelId \|\| imageModelId \|\| undefined\}/);
  assert.match(source, /图片画幅<\/span><Select allowClear value=\{editing\?\.imageAspectRatio \|\| imageAspectRatio \|\| undefined\}/);
  assert.match(source, /imageModelId: editing\.imageModelId \|\| ''/);
  assert.match(source, /imageAspectRatio: editing\.imageAspectRatio \|\| ''/);
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
  const end = source.indexOf('function contentEditablePromptText(root)', start);
  const helper = start >= 0 && end > start ? source.slice(start, end) : '';
  assert.match(helper, /if \(!root\)/);
  assert.match(helper, /if \(!selection\?\.rangeCount \|\| !root\.contains\(selection\.anchorNode\)\)/);
});
