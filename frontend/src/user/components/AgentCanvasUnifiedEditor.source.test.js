import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const canvasSource = fs.readFileSync(path.resolve(dirname, 'AgentCanvas.jsx'), 'utf8');
const composerSource = fs.readFileSync(path.resolve(dirname, 'AgentNodeComposer.jsx'), 'utf8');
const parameterSource = fs.readFileSync(path.resolve(dirname, 'AgentNodeParameterPopover.jsx'), 'utf8');
const settingsSource = fs.readFileSync(path.resolve(dirname, 'AgentNodeSettingsPanel.jsx'), 'utf8');
const styles = fs.readFileSync(path.resolve(dirname, '../../shared/styles/agent-workspace.css'), 'utf8');

test('all editable node kinds use one single-click editor gate', () => {
  assert.match(canvasSource, /const EDITABLE_NODE_TYPES = new Set\(\['text', 'image', 'video', 'audio'\]\)/);
  assert.match(canvasSource, /EDITABLE_NODE_TYPES\.has\(normalizeCanvasNodeType\(node\.data\?\.type\)\)[\s\S]*?openNodeComposer\(node\.id\)/);
});

test('all editable nodes share inverse-scaled floating toolbar and composer shells', () => {
  assert.match(canvasSource, /agent-node-floating-toolbar/);
  assert.match(canvasSource, /agent-node-floating-composer/);
  assert.match(styles, /agent-node-floating-toolbar[^}]*scale\(var\(--agent-canvas-inverse-zoom\)\)/);
  assert.match(styles, /agent-node-floating-composer[^}]*scale\(var\(--agent-canvas-inverse-zoom\)\)/);
  assert.doesNotMatch(styles, /agent-canvas-node--(?:text|image|video|audio)\.is-editor-open\s*\{[^}]*width:/);
});

test('quick tools retain the real operations for text image video and audio', () => {
  for (const token of [
    "operation: 'generate'", "operation: 'rewrite'", "operation: 'continue'",
    "operation: 'rotate'", "operation: 'mirror'", "operation: 'grid_split'",
    "operation: 'trim'", "operation: 'crop'", "operation: 'separate_audio'", "operation: 'extract_frames'",
    "operation: 'custom_split'", "operation: 'smart_split'", "operation: 'speed'"
  ]) assert.match(canvasSource, new RegExp(token));
  assert.match(settingsSource, /结构要求/);
  assert.match(parameterSource, /生成类型/);
  assert.match(parameterSource, /文字长度/);
  assert.match(parameterSource, /创作风格/);
});

test('the unified composer retains real model parameters settings state and submission', () => {
  assert.match(composerSource, /AgentModelPicker/);
  assert.match(composerSource, /AgentNodeParameterPopover/);
  assert.match(composerSource, /AgentNodeSettingsPanel/);
  assert.match(composerSource, /agent-node-composer__state/);
  assert.match(composerSource, /submitNodeGeneration/);
  for (const fake of ['积分', '角色库', '特效库', '运镜库']) assert.doesNotMatch(composerSource, new RegExp(fake));
});

test('editor surfaces use theme variables instead of fixed dark backgrounds', () => {
  for (const variable of ['--agent-editor-surface', '--agent-editor-toolbar', '--agent-editor-action', '--agent-editor-border', '--agent-editor-text', '--agent-editor-muted']) {
    assert.match(styles, new RegExp(variable));
  }
  assert.doesNotMatch(styles, /agent-(?:node-composer|video-quick-tools|audio-quick-tools)[^}]*background:\s*#202020/);
  assert.doesNotMatch(styles, /agent-audio-waveform__bars[^}]*background:\s*#242424/);
});

test('node surfaces stay draggable while controls opt out of canvas gestures', () => {
  assert.match(canvasSource, /agent-canvas-node-drag-surface/);
  assert.match(canvasSource, /agent-node-floating-toolbar nodrag nopan/);
  assert.match(canvasSource, /agent-node-floating-composer nodrag nopan/);
  assert.doesNotMatch(canvasSource, /dragHandle:\s*['"]\.agent-canvas-node-drag-handle/);
});
