import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'ScriptPage.jsx'), 'utf8');

test('script video generation checks the shared local executor status for the Doubao model', () => {
  assert.match(source, /useLocalExecutorStatus/);
  assert.match(source, /localExecutorStatus\.online/);
  assert.doesNotMatch(source, /apiRequest\('\/api\/shuihuo-production\/local-executors'/);
});

test('script image configuration exposes the same Doubao executor and the entity panel waits for its task result', () => {
  assert.match(source, /图片生成通道/);
  assert.match(source, /providerOptionsForModels/);
  const panel = fs.readFileSync(path.join(here, '..', 'components', 'EntityImagePanel.jsx'), 'utf8');
  assert.match(panel, /getReferenceAssetImageTask/);
  assert.match(panel, /图片正在后台生成/);
});
