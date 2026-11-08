import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'ScriptPage.jsx'), 'utf8');

test('private comic submit uses the single-book stages without batch automation presets', () => {
  assert.match(source, /ensureCurrentPrivateComicTask\(\)/);
  assert.match(source, /runBookStage\(snapshot\.batchId, snapshot\.bookId, 'assets'/);
  assert.match(source, /runBookStage\(snapshot\.batchId, snapshot\.bookId, 'comic'/);
  assert.doesNotMatch(source, /startBatchAutomation/);
  assert.doesNotMatch(source, /listAutomationPresets/);
  assert.match(source, /uploadComicGlobalReference\(snapshot\.batchId, privateComicPendingReference\)/);
  assert.match(source, /mergePrivateComicReference\(settingsForStart, uploadedReference\)/);
});

test('private comic task creation is shared by extraction and generation', () => {
  assert.match(source, /async function ensureCurrentPrivateComicTask\(\)/);
  assert.match(source, /ensurePrivateComicTask\(\{/);
  assert.match(source, /const ensured = await ensureCurrentPrivateComicTask\(\)/);
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
  assert.match(source, /const ensured = await ensurePrivateComicTask\(\{/);
  assert.match(source, /if \(ensured\.created\) setPrivateComicTask\(ensured\.task\)/);
  assert.match(source, /privateComicTaskKey/);
  assert.match(source, /localStorage\.setItem\(privateComicScopedKey\(privateComicTaskKey\)/);
  assert.match(source, /getBatchRuntimeSummary\(batchId, \{ suppressGlobalError: true \}\)/);
  assert.match(source, /setInterval\(poll, 5000\)/);
});
