import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const read = relative => fs.readFileSync(path.join(here, relative), 'utf8');

test('final UI page mounts the workbench and every approved settings surface', () => {
  const source = read('BatchFactoryV11UiPage.jsx');
  for (const marker of [
    'BatchFactoryV11Workbench',
    'ProductionSettingsDrawer',
    'BookSettingsModal',
    'VideoSettingsDrawer'
  ]) assert.match(source, new RegExp(marker));
  assert.match(source, /onRunHook={runHook}/);
  assert.match(source, /onApproveHook={approveHook}/);
  assert.match(source, /onRunDirector={runDirector}/);
  assert.match(source, /PublishSettingsDrawer/);
  assert.match(source, /onOpenPublishSettings/);
});

test('V11 maps personal-center constraint prompts into the workbench adapter', () => {
  const source = read('BatchFactoryV11UiPage.jsx');
  assert.match(source, /listPersonalConstraintPrompts/);
  assert.match(source, /savePersonalConstraintPrompt/);
  assert.match(source, /listPersonalConstraintPrompts: batchFactoryV11\.listPersonalConstraintPrompts/);
  assert.doesNotMatch(source, /shared\/api\/generation/);
  assert.match(source, /category,\s*name:/);
});
