import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');
const apiSource = fs.readFileSync(path.join(here, '../../../shared/api/batchFactoryV11.js'), 'utf8');

test('keeps AI prompt candidates separate from the editable video prompt and video generation', () => {
  assert.match(apiSource, /export function createDirectorPromptCandidate\(batchId, bookId, expectedRevision\)/);
  assert.match(apiSource, /export function getDirectorPromptCandidate\(batchId, bookId\)/);
  assert.match(apiSource, /export function acceptDirectorPromptCandidate\(batchId, bookId, videoId, candidateId, expectedRevision\)/);
  assert.match(apiSource, /export function cancelDirectorPromptCandidate\(batchId, bookId, expectedRevision\)/);
  assert.match(source, /查看 AI 原始输出/);
  assert.match(source, /aiOriginalVideoPrompt \|\| selectedVideo\?\.videoPrompt/);
  assert.match(source, /setVideoPrompt\(selectedVideo\?\.videoPrompt \|\| ''\)/);
  assert.match(source, /createDirectorPromptCandidate\(batchId, book\.id, Number\(book\.revision \|\| 0\)\)/);
  assert.match(source, /acceptDirectorPromptCandidate\(batchId, book\.id, selectedVideo\.id, candidate\.id, Number\(selectedVideo\.revision \|\| 0\)\)/);
  assert.match(source, /cancelDirectorPromptCandidate\(batchId, book\.id,/);
  assert.match(source, /const patch = promptKind === 'visual' \? \{ comicPrompt \} : \{ videoPrompt: savedVideoPrompt \};/);
  assert.match(source, /await saveVideoOverride\(batchId, book\.id, selectedVideo\.id, \{ patch, expectedRevision: Number\(selectedVideo\.revision \|\| 0\) \}\);[\s\S]*?let h3CompileWarning = '';/);
  assert.match(source, /if \(promptKind === 'video' && \(book\?\.directorRevision\?\.output\?\.h3_director \|\| book\?\.directorRevision\?\.output\?\.h3Director\) && promptEditTrace\) \{[\s\S]*?try \{[\s\S]*?await compileH3Video\(batchId, book\.id, h3PromptEditRequest\(promptEditTrace, selectedIndex, savedVideoPrompt\)\);[\s\S]*?\} catch \(error\) \{[\s\S]*?h3CompileWarning = error\?\.message/);
  assert.match(source, /message\.success\(promptKind === 'visual'[\s\S]*?if \(h3CompileWarning\) message\.warning\(`提示词已保存；\$\{h3CompileWarning\}`\);/);
  assert.doesNotMatch(source, /编辑 \? '完成编辑' : '编辑'/);
  assert.match(source, /!editing \? <Button onClick=\{toggleEditing\}>编辑<\/Button> : null/);
  assert.match(source, /catch \(_\) \{ setPromptEditTrace\(null\); setVideoPrompt\(submittedVideoPrompt\); \}/);
  assert.match(source, /className="batch-factory-prompt-primary-action"[\s\S]*?>生成视频<\/Button>/);
  assert.match(source, /'重新生成 AI'/);
  assert.match(source, />采用新结果</);
  assert.match(source, />取消</);
  assert.match(source, /onGenerateVideo\?\.\(selectedVideo\?\.id\)/);
  assert.doesNotMatch(source, /onRegenerate=\{\(\) => runBookStageAction\(promptBook, 'director', 'force'\)\}/);
});
