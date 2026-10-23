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

test('image node preview is a React Flow non-drag control so its editor click is delivered', () => {
  assert.match(canvasSource, /className="agent-canvas-node-image-select nodrag nopan" aria-label="展开图片编辑"/);
});

test('opening a media editor prevents the canvas selection handler from immediately closing it', () => {
  const opener = canvasSource.match(/const openEditor = event => \{([\s\S]*?)\n  \};/);
  assert.ok(opener, 'media editor opener should exist');
  assert.match(opener[1], /event\.preventDefault\(\);/);
  assert.match(opener[1], /stopCanvasEvent\(event\);/);
  assert.match(canvasSource, /event\.target\?\.closest\?\.\('\.agent-node-composer, \.agent-canvas-node-image-select/);
});

test('canvas loads all configured model kinds and calls real text and image APIs', () => {
  assert.match(canvasSource, /Promise\.all\(\['text', 'image', 'video', 'audio'\]\.map\(kind => listAvailableModels\(kind\)\)\)/);
  assert.match(canvasSource, /createAgentTextGeneration\(/);
  assert.match(canvasSource, /createAgentImageGeneration\(/);
  assert.doesNotMatch(composerSource, /暂未接入生成执行器/);
});

test('canvas close lifecycle clears the composer and its material panel together', () => {
  assert.match(canvasSource, /function closeNodeComposer\(event\) \{\s*\/\/ React Flow can report nested controls as pane clicks\. Only its blank pane closes the editor\.\s*if \(event && !event\.target\?\.classList\?\.contains\('react-flow__pane'\)\) return;/);
  assert.match(canvasSource, /setActiveEditorNodeID\(''\)/);
  assert.match(canvasSource, /setAssetPanel\(null\)/);
  assert.match(canvasSource, /onPaneClick=\{closeNodeComposer\}/);
  assert.match(canvasSource, /event\.key === 'Escape'[\s\S]*closeNodeComposer\(\)/);
});

test('the image material button is protected from the canvas node-click cleanup', () => {
  assert.match(composerSource, /className="agent-node-composer__source" onPointerDown=\{stop\} onClick=\{event => \{ event\.preventDefault\(\); stop\(event\); onPickImageAsset\?\.\(node\.id, event\); \}\}/);
  assert.match(canvasSource, /const isMultiSelection = isCanvasMultiSelectionEvent\(event\);/);
  assert.match(canvasSource, /event\.defaultPrevented \|\| \(!isMultiSelection && event\.target\?\.closest\?\.\('\.agent-node-composer, \.agent-canvas-node-image-select, \.agent-canvas-node-video-placeholder, \.agent-canvas-node-audio-placeholder, \.agent-canvas-node-text-preview'\)\)/);
  assert.match(canvasSource, /const panel = \{ assetKind: 'image', mode, nodeID, position: \{ x: 0, y: 0 \}, anchor: imageAssetPanelAnchor\(event\) \};\s*\/\/ React Flow completes its node click after this nested control; open after that cleanup\.\s*window\.requestAnimationFrame\(\(\) => setAssetPanel\(panel\)\);/);
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

test('image parameter popover uses a compact visual grid instead of narrow selects', () => {
  assert.match(parameterSource, /agent-node-parameter-popover__ratio-grid/);
  assert.match(parameterSource, /agent-node-parameter-popover__ratio-shape/);
  assert.match(parameterSource, /agent-node-parameter-popover__option-grid/);
  assert.doesNotMatch(parameterSource, /labels\.image\.resolution[\s\S]{0,180}<Select/);
  assert.doesNotMatch(parameterSource, /labels\.image\.count[\s\S]{0,180}<Select/);
  assert.match(styles, /agent-node-parameter-popover__ratio-grid/);
  assert.match(styles, /agent-node-parameter-popover__ratio-shape/);
});

test('one settings entry shows only the current node kind channel', () => {
  assert.match(settingsSource, /image: '图片生成通道'/);
  assert.match(settingsSource, /video: '视频引擎'/);
  assert.match(settingsSource, /text: '文本通道'/);
  assert.match(settingsSource, /audio: '音频通道'/);
  assert.match(settingsSource, /aria-label="设置"/);
});

test('settings panel keeps select menus inside its popover interaction boundary', () => {
  assert.match(settingsSource, /getPopupContainer=\{trigger => trigger\.parentElement\}/);
});

test('settings panel stops click bubbling before it reaches the canvas node', () => {
  assert.match(settingsSource, /aria-label=\{`\$\{channelLabels\[kind\]\}设置`\} onMouseDown=\{stop\} onClick=\{stop\}/);
});
