import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'ScriptPage.jsx'), 'utf8');

test('private comic submit creates one pasted task and starts comic-only automation', () => {
  assert.match(source, /createPrivateComicTask\(/);
  assert.match(source, /startBatchAutomation\(batchId, \{/);
  assert.match(source, /runMode: 'video_no_submit'/);
  assert.match(source, /productionMode: 'comic'/);
  assert.match(source, /autoPublish: false/);
  assert.match(source, /settingsPatch/);
  assert.match(source, /uploadComicGlobalReference\(snapshot\.batchId, privateComicPendingReference\)/);
  assert.match(source, /mergePrivateComicReference\(initialBatchSettings, uploadedReference\)/);
});

test('repeated comic submit reuses the current task instead of creating duplicate batches', () => {
  assert.match(source, /shouldReusePrivateComicTask\(privateComicTask, privateComicDraft\.sourceText\)/);
  assert.match(source, /if \(reuseExisting\) \{/);
  assert.match(source, /已沿用当前漫画任务/);
});

test('script more-menu export writes a real TXT file', () => {
  assert.match(source, /function exportOutput\(\)/);
  assert.match(source, /new Blob\(\[\`\$\{title\}\\n\\n\$\{content\}\`\]/);
  assert.match(source, /anchor\.download = filename/);
  assert.match(source, /if \(key === 'export'\) exportOutput\(\);/);
});

test('comic page polling recognises backend status and generationState', () => {
  assert.match(source, /const activeStates = \['generating', 'processing', 'running', 'queued'\];/);
  assert.match(source, /page\?\.generationState/);
});

test('private comic keeps task identity and polls the existing runtime summary', () => {
  assert.match(source, /privateComicTaskSnapshot\(created, privateComicDraft\.sourceText\)/);
  assert.match(source, /privateComicTaskKey/);
  assert.match(source, /localStorage\.setItem\(privateComicScopedKey\(privateComicTaskKey\)/);
  assert.match(source, /getBatchRuntimeSummary\(batchId, \{ suppressGlobalError: true \}\)/);
  assert.match(source, /setInterval\(poll, 5000\)/);
});
