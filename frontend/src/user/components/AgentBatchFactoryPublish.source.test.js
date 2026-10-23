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
