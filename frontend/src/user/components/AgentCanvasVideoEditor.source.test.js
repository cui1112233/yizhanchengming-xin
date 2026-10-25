import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const canvasSource = fs.readFileSync(path.resolve(dirname, 'AgentCanvas.jsx'), 'utf8');
const composerSource = fs.readFileSync(path.resolve(dirname, 'AgentNodeComposer.jsx'), 'utf8');
const styles = fs.readFileSync(path.resolve(dirname, '../../shared/styles/agent-workspace.css'), 'utf8');

test('single-clicking a video node opens its editor', () => {
  assert.match(canvasSource, /normalizeCanvasNodeType\(node\.data\?\.type\) === 'video'[\s\S]*?openNodeComposer\(node\.id\)/);
});

test('an open video node exposes real quick tools above its preview', () => {
  assert.match(canvasSource, /function AgentVideoQuickTools/);
  assert.match(canvasSource, /operation:\s*'separate_audio'/);
  for (const label of ['截取', '裁剪', '逐帧拉片', '分离音视频', '下载视频']) {
    assert.match(canvasSource, new RegExp(label));
  }
  assert.match(canvasSource, /isEditorOpen && nodeType === 'video'/);
});

test('the quick toolbar opens the existing functional video tools panel', () => {
  assert.match(canvasSource, /videoToolsRequest=\{videoToolsRequest\}/);
  assert.match(composerSource, /videoToolsRequest = 0/);
  assert.match(composerSource, /setToolsOpen\(true\)/);
});

test('the selected video, toolbar, and lower editor share one adaptive width', () => {
  assert.match(styles, /agent-canvas-node--video\.is-editor-open[^}]*width:\s*min\(680px,/);
  assert.match(styles, /agent-video-quick-tools[^}]*min-height:\s*40px/);
  assert.match(styles, /agent-canvas-node--video\.is-editor-open \.agent-node-composer[^}]*width:\s*100%/);
});
