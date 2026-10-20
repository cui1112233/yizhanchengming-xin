import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const read = relativePath => fs.readFileSync(path.resolve(dirname, relativePath), 'utf8');

test('工作区侧栏提供节点素材搜索和拖动改宽', () => {
  const drawer = read('../components/AgentWorkspaceDrawer.jsx');
  assert.match(drawer, /画布节点/);
  assert.match(drawer, /项目素材/);
  assert.match(drawer, /setPointerCapture/);
  assert.match(drawer, /onWidthChange/);
  assert.match(drawer, /draggable/);
});

test('页面读取项目素材并把定位与创建动作交给真实画布', () => {
  const page = read('./AgentPageV2.jsx');
  assert.match(page, /listAgentAssets\(activeTaskId\)/);
  assert.match(page, /canvasRef\.current\?\.locateNode/);
  assert.match(page, /canvasRef\.current\?\.createFromAsset/);
});
