import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const canvasSource = fs.readFileSync(path.resolve(dirname, '../components/AgentCanvas.jsx'), 'utf8');
const composerSource = fs.readFileSync(path.resolve(dirname, '../components/AgentNodeComposer.jsx'), 'utf8');
const parameterSource = fs.readFileSync(path.resolve(dirname, '../components/AgentNodeParameterPopover.jsx'), 'utf8');
const settingsSource = fs.readFileSync(path.resolve(dirname, '../components/AgentNodeSettingsPanel.jsx'), 'utf8');
const styles = fs.readFileSync(path.resolve(dirname, '../../shared/styles/agent-workspace.css'), 'utf8');

test('canvas renders one adaptive composer for text image video and audio nodes', () => {
  assert.match(canvasSource, /import AgentNodeComposer from '\.\/AgentNodeComposer'/);
  assert.match(canvasSource, /<AgentNodeComposer/);
  assert.match(composerSource, /\['text', 'image', 'video', 'audio'\]\.includes\(kind\)/);
  assert.match(composerSource, /<AgentModelPicker kind=\{kind\}/);
  assert.match(composerSource, /reconcileNodeModel\(config, models, kind\)/);
});

test('canvas loads all configured model kinds and calls real text and image APIs', () => {
  assert.match(canvasSource, /Promise\.all\(\['text', 'image', 'video', 'audio'\]\.map\(kind => listAvailableModels\(kind\)\)\)/);
  assert.match(canvasSource, /createAgentTextGeneration\(/);
  assert.match(canvasSource, /createAgentImageGeneration\(/);
  assert.doesNotMatch(composerSource, /暂未接入生成执行器/);
});

test('canvas close lifecycle clears the composer and its material panel together', () => {
  assert.match(canvasSource, /function closeNodeComposer\(\)/);
  assert.match(canvasSource, /setActiveEditorNodeID\(''\)/);
  assert.match(canvasSource, /setAssetPanel\(null\)/);
  assert.match(canvasSource, /onPaneClick=\{closeNodeComposer\}/);
  assert.match(canvasSource, /event\.key === 'Escape'[\s\S]*closeNodeComposer\(\)/);
});

test('adaptive composer is wide rather than tall and keeps controls on one line', () => {
  assert.doesNotMatch(styles, /agent-canvas-generation-panel[^}]*width:\s*800px/);
  assert.match(styles, /agent-node-composer[^}]*width:\s*min\(680px,/);
  assert.match(styles, /agent-node-composer__toolbar[^}]*flex-wrap:\s*nowrap/);
  assert.match(styles, /agent-node-composer__model[^}]*min-width:\s*220px/);
});

test('image composer shows the complete compact settings summary and opens all four settings', () => {
  assert.match(composerSource, /<AgentNodeParameterPopover/);
  assert.match(composerSource, /<AgentNodeSettingsPanel/);
  assert.equal((composerSource.match(/<AgentNodeSettingsPanel/g) || []).length, 1);
  assert.match(styles, /agent-node-composer__parameters/);
  assert.match(parameterSource, /formatNodeParameterSummary/);
  assert.match(parameterSource, /比例/);
  assert.match(parameterSource, /画质/);
  assert.match(parameterSource, /清晰度/);
  assert.match(parameterSource, /数量/);
});

test('one settings entry shows only the current node kind channel', () => {
  assert.match(settingsSource, /image: '图片生成通道'/);
  assert.match(settingsSource, /video: '视频引擎'/);
  assert.match(settingsSource, /text: '文本通道'/);
  assert.match(settingsSource, /audio: '音频通道'/);
  assert.match(settingsSource, /aria-label="设置"/);
});
