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

test('opening the video editor does not resize the video node', () => {
  assert.doesNotMatch(styles, /agent-canvas-node--video\.is-editor-open\s*\{[^}]*width:/);
  assert.match(canvasSource, /'--agent-canvas-inverse-zoom':\s*1\s*\/\s*Math\.max\(canvasZoom,\s*0\.25\)/);
});

test('video tools and composer are inverse-scaled floating surfaces', () => {
  assert.match(canvasSource, /agent-video-floating-toolbar/);
  assert.match(canvasSource, /agent-video-floating-composer/);
  assert.match(styles, /agent-video-floating-toolbar[^}]*position:\s*absolute[^}]*scale\(var\(--agent-canvas-inverse-zoom\)\)/);
  assert.match(styles, /agent-video-floating-composer[^}]*position:\s*absolute[^}]*scale\(var\(--agent-canvas-inverse-zoom\)\)/);
});
