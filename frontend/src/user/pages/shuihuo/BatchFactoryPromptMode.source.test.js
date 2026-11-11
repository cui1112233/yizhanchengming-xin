import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');

test('locks book prompt cards and editors to the selected production mode', () => {
  assert.match(source, /function InlineBookPrompts\(\{[^}]*productionMode = 'video'/);
  assert.match(source, /function PromptPanel\(\{[^}]*productionMode = 'video'/);
  assert.match(source, /const promptKind = productionMode === 'comic' \? 'visual' : 'video'/);
  assert.doesNotMatch(source, /setPromptKind/);
  assert.match(source, /batch-factory-prompt-mode-label is-\$\{productionMode\}/);
  assert.match(source, /productionMode=\{productionMode\}/);
  assert.doesNotMatch(source, /batch-factory-prompt-modal-tabs/);
});

test('skips the video final-prompt request in comic mode and keeps the video default', () => {
  // 漫画模式没有导演分镜，卡片与弹窗都不应请求视频最终提示词。
  assert.match(source, /const finalPromptEnabled = productionMode !== 'comic'/);
  assert.match(source, /useCompiledVideoPrompt\(batchId, book, video, settingsRevision, \{ enabled: finalPromptEnabled \}\)/);
  assert.match(source, /useCompiledVideoPrompt\(batchId, book, selectedVideo, settingsRevision, \{ enabled: productionMode !== 'comic' \}\)/);
  // 模式未加载时仍按默认 video 处理，不会因为缺少模式数据就不请求。
  assert.match(source, /function InlineBookPrompts\(\{[^}]*productionMode = 'video'/);
  assert.match(source, /function PromptPanel\(\{[^}]*productionMode = 'video'/);
});
