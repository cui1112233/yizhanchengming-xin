import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const controls = fs.readFileSync(path.resolve(dirname, '../components/AgentCanvasControls.jsx'), 'utf8');
const canvas = fs.readFileSync(path.resolve(dirname, '../components/AgentCanvas.jsx'), 'utf8');

test('画布控制区提供持久显示开关、自动布局和快捷键帮助', () => {
  for (const label of ['小地图', '显示连线', '网格吸附', '自动横排', '自动竖排', '快捷键']) assert.match(controls, new RegExp(label));
  assert.match(canvas, /<MiniMap/);
  assert.match(canvas, /snapToGrid=\{workspaceView\.snapToGrid\}/);
  assert.match(canvas, /layoutCanvas/);
});
