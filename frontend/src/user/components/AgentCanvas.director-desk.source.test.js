import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const canvas = fs.readFileSync(path.join(dirname, 'AgentCanvas.jsx'), 'utf8');
const desk = fs.readFileSync(path.join(dirname, 'AgentDirectorDeskNode.jsx'), 'utf8');

test('画布通过全屏导演台适配器打开 3D 工作区', () => {
  assert.match(canvas, /DirectorDeskApp/);
  assert.match(canvas, /createDirectorDeskHostAdapter/);
  assert.match(canvas, /directorDeskNodeID/);
  assert.match(canvas, /onOpen={onOpenDirectorDesk}/);
  assert.doesNotMatch(canvas, /AgentDirectorDeskWorkspace/);
});

test('画布在上传成功后调用导演台截图插入函数', () => {
  assert.match(canvas, /insertDirectorCaptureState/);
  assert.match(canvas, /commitDirectorCapture/);
  assert.match(canvas, /uploadCapture/);
  assert.match(canvas, /uploadAgentAsset/);
});

test('画布节点将全屏工作区明确标识为 3D 导演台', () => {
  assert.match(desk, /3D 导演台/);
  assert.match(desk, /打开 3D 导演台/);
});
