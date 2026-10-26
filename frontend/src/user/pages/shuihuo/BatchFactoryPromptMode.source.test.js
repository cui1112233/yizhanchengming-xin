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
