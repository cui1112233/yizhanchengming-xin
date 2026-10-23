import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const canvas = fs.readFileSync(path.join(dirname, 'AgentCanvas.jsx'), 'utf8');

test('导演台是画布中的预演节点，并使用同一份节点配置保存', () => {
  const desk = fs.readFileSync(path.join(dirname, 'AgentDirectorDeskNode.jsx'), 'utf8');
  assert.match(desk, /directorDeskConfig/);
  assert.match(desk, /预演参考/);
  assert.match(desk, /onConfigChange/);
  assert.match(desk, /onPointerDown/);
  assert.match(desk, /captureDirectorDeskScreenshot/);
  assert.match(desk, /onSaveScreenshot/);
  assert.match(canvas, /AgentDirectorDeskNode/);
  assert.match(canvas, /key: 'director'/);
  assert.match(canvas, /uploadAgentAsset\(requestedTaskId, file\)/);
});
