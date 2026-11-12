import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'PrivateComicWorkspace.jsx'), 'utf8');
const scriptPageSource = fs.readFileSync(path.join(here, 'ScriptPage.jsx'), 'utf8');
const pickerSource = fs.readFileSync(path.join(here, 'shuihuo', 'BatchFactoryComicStylePicker.jsx'), 'utf8');
const constraintsSource = fs.readFileSync(path.join(here, 'private-studio', 'PrivateStudioConstraintsModal.jsx'), 'utf8');

test('private comic workspace accepts pasted text and exposes the comic actions', () => {
  assert.match(source, /name="sourceText"/);
  assert.match(source, /name="title"/);
  assert.match(source, /刷新状态/);
  assert.match(source, /查看漫画详情/);
  assert.match(source, /不自动发布/);
});

test('private comic workspace reuses the published style library and comic prompt catalog', () => {
  assert.match(source, /BatchFactoryComicStylePicker/);
  assert.match(source, /onReferenceDataUrl=\{onPendingReferenceChange\}/);
  assert.match(source, /batch\.comic-page/);
  assert.match(source, /batch\.comic-image-derivation/);
  assert.match(source, /PRIVATE_STUDIO_CONSTRAINT_FIELDS\.comic\.map/);
  assert.match(constraintsSource, /key: 'comic_quality'/);
  assert.match(constraintsSource, /key: 'comic_restriction'/);
  assert.match(constraintsSource, /key: 'comic_negative'/);
  assert.match(source, /private-customization-reference-note/);
});

test('comic constraints are persisted to the server task with a readback check', () => {
  assert.match(source, /onPersistSettings/);
  assert.match(source, /async function saveSettings\(\)/);
  assert.match(source, /await onPersistSettings\?\.\(\)/);
  assert.match(source, /message\.error\(error\?\.message \|\| '漫画约束保存失败'\)/);
  assert.match(source, /onOk=\{\(\) => void saveSettings\(\)\}/);
  assert.match(source, /okButtonProps=\{\{ loading: savingSettings \}\}/);
  assert.match(source, /okText=\{task\?\.batchId \? '保存到任务' : '完成'\}/);
});

test('private comic workspace shows a live page estimate before generation', () => {
  assert.match(source, /import \{ comicEstimatedPageRange \} from '\.\/shuihuo\/batchFactoryContentRange\.js';/);
  assert.match(source, /const pageEstimate = comicEstimatedPageRange\(sourceLength\);/);
  assert.match(source, /已读取 \$\{sourceLength\} 字 · 预计约 \$\{pageEstimate\.min\}～\$\{pageEstimate\.max\} 张/);
});

test('private comic models have one editor and the toolbar only shows the effective values', () => {
  assert.match(source, /resolveModelLabel\(modelOptions\.text, draft\.textModelId\)/);
  assert.match(source, /resolveModelLabel\(modelOptions\.image, draft\.imageModelId\)/);
  assert.match(source, /private-studio-tool-value/);
  assert.doesNotMatch(source, /onChange=\{textModelId => patchDraft\(\{ textModelId/);
  assert.doesNotMatch(source, /onChange=\{imageModelId => patchDraft\(\{ imageModelId/);
});

test('comic configuration edits the current comic while script configuration keeps account defaults', () => {
  assert.match(scriptPageSource, /if \(customizationWorkspace === 'comic'\) \{[\s\S]{0,700}textModelId: privateComicDraft\.textModelId \|\| scriptModelSelection\.textModelId/);
  assert.match(scriptPageSource, /if \(customizationWorkspace === 'comic'\) \{[\s\S]{0,700}setPrivateComicDraft\(current => \(\{[\s\S]{0,300}textModelId: next\.textModelId/);
  assert.match(scriptPageSource, /if \(!scriptModelsLoaded \|\| !scriptDefaultsLoaded\) return;/);
  assert.match(scriptPageSource, /okText=\{customizationWorkspace === 'comic' \? '应用到当前漫画' : '保存为默认配置'\}/);
  assert.match(scriptPageSource, /customizationWorkspace !== 'comic' \? <>[\s\S]{0,1800}视频模型/);
});

test('private comic treats asset images as a required step for character consistency', () => {
  assert.match(source, /key: 'assetImages', label: '资产图', optional: false/);
  assert.match(source, /find\(step => !step\.done && !step\.optional\)/);
  assert.match(source, /生成资产图（缺 \$\{missingAssetImages\}）/);
});

test('changing the style clears a pre-submit custom reference', () => {
  assert.match(pickerSource, /onReferenceDataUrl\?\.\('\'\)/);
});

test('clearing pasted comic text asks for confirmation first', () => {
  assert.match(source, /<Popconfirm/);
  assert.match(source, /清空已粘贴的 \$\{sourceLength\} 字正文？/);
  assert.match(source, /该操作不可撤销。/);
  assert.match(source, /onConfirm=\{\(\) => patchDraft\(\{ sourceText: '' \}\)\}/);
});

test('comic page cards load the thumbnail first and fetch the original only on demand', () => {
  assert.match(source, /const sources = resolveComicImageSources\(primary\)/);
  assert.match(source, /const load = target => getComicPageImageBlob\(target\)/);
  assert.match(source, /load\(sources\.thumbnail\)/);
  assert.match(source, /async function ensureOriginal\(\)/);
  assert.match(source, /async function openViewer\(\)/);
  assert.match(source, /点击查看原图/);
  assert.match(source, /原图读取失败，请稍后重试/);
});

test('comic output cards render the first screen lazily instead of all pages at once', () => {
  assert.match(source, /nextComicOutputVisibleCount/);
  assert.match(source, /new IntersectionObserver/);
  assert.match(source, /const visiblePages = pageList\.slice\(0, visibleCount\)/);
  assert.match(source, /visiblePages\.map/);
  assert.match(source, /shot-output-cards-more/);
});

test('single-book comic assets open the shared asset editor instead of a passive preview', () => {
  assert.match(source, /import \{ PrivateStudioAssetEditor \} from '\.\/private-studio\/PrivateStudioAssetEditor';/);
  assert.match(source, /<button[\s\S]{0,160}className="private-comic-asset-name"/);
  assert.match(source, /<PrivateStudioAssetEditor/);
  assert.match(source, /onSaveAsset/);
  assert.match(source, /onUploadAssetImage/);
  assert.match(source, /onSetPrimaryAssetImage/);
  assert.match(source, /onGenerateAssetImage/);
  assert.doesNotMatch(source, /title=\{assetPreview \? `\$\{assetPreview\.title\} · \$\{assetPreview\.items\.length\}` : ''\}/);
});
