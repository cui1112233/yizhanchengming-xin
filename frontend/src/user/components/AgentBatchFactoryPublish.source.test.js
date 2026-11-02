import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const read = name => fs.readFileSync(path.join(here, name), 'utf8');

test('final Batch Factory node submits through the shared book publish route', () => {
  const canvas = read('AgentCanvas.jsx');
  const composer = read('AgentNodeComposer.jsx');

  assert.match(canvas, /submitBookTo121/);
  assert.match(canvas, /await submitBookTo121\(action\.batchId, action\.bookId/);
  assert.match(canvas, /await syncAgentBatchFactoryBook\(requestedTaskId, action\.bookId\)/);
  assert.match(composer, /resourceType === 'merge'/);
  assert.match(composer, /上传到视频管理系统/);
  assert.match(composer, /onPublishBatchFactory/);
});

test('final Batch Factory node lets the user save a shared organization before upload', () => {
  const canvas = read('AgentCanvas.jsx');
  const composer = read('AgentNodeComposer.jsx');

  assert.match(canvas, /get121OrganizationOptions/);
  assert.match(canvas, /saveBookOverride\(action\.batchId, action\.bookId/);
  assert.match(canvas, /publishSettings: \{ organization \}/);
  assert.match(canvas, /bookRevision/);
  assert.match(composer, /视频管理系统组织归属/);
  assert.match(composer, /onPublishOrganizationChange/);
});

test('Batch Factory script cards show their shared production state instead of an unrelated split warning', () => {
  const scriptComposer = read('AgentScriptNodeComposer.jsx');

  assert.match(scriptComposer, /批量工厂正文已同步/);
});
