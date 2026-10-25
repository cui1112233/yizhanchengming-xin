import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const canvas = fs.readFileSync(path.resolve(dirname, '../components/AgentCanvas.jsx'), 'utf8');
const composer = fs.readFileSync(path.resolve(dirname, '../components/AgentNodeComposer.jsx'), 'utf8');
const tools = fs.readFileSync(path.resolve(dirname, '../components/AgentImageTools.jsx'), 'utf8');
const styles = fs.readFileSync(path.resolve(dirname, '../../shared/styles/agent-workspace.css'), 'utf8');

test('image composer exposes deterministic image tools only when a current image exists', () => {
  assert.match(composer, /<AgentImageTools/);
  assert.match(composer, /sourceAsset=\{node\.data\?\.assetRef\}/);
  assert.match(tools, /裁剪/);
  assert.match(tools, /旋转 90°/);
  assert.match(tools, /水平镜像/);
  assert.match(tools, /2 x 2 拆图/);
});

test('canvas submits the tool job and displays its shared task status', () => {
  assert.match(canvas, /createAgentImageTool/);
  assert.match(canvas, /async function submitImageTool/);
  assert.match(canvas, /onImageTool=\{submitImageTool\}/);
});

test('image tools present named actions and disclose labelled crop fields only on demand', () => {
  assert.match(tools, /const \[cropOpen, setCropOpen\] = useState\(false\)/);
  assert.match(tools, /agent-image-tools__action-grid/);
  assert.match(tools, /旋转 90°/);
  assert.match(tools, /水平镜像/);
  assert.match(tools, /2 x 2 拆图/);
  assert.match(tools, /裁剪/);
  assert.match(tools, /agent-image-tools__crop-fields/);
  assert.match(tools, /起点 X/);
  assert.match(tools, /起点 Y/);
  assert.match(tools, /宽度/);
  assert.match(tools, /高度/);
  assert.match(tools, /生成裁剪结果/);
});

test('image tool controls use scoped canvas theme styles including Ant number handlers', () => {
  assert.match(styles, /\.agent-page \.agent-image-tools\s*\{/);
  assert.match(styles, /agent-image-tools[^}]*var\(--agent-editor-surface/);
  assert.match(styles, /agent-image-tools__action-grid/);
  assert.match(styles, /agent-image-tools__crop-fields/);
  assert.match(styles, /agent-image-tools \.ant-input-number-handler-wrap/);
});
