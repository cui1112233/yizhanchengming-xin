import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const canvas = fs.readFileSync(path.join(here, 'AgentCanvas.jsx'), 'utf8');
const styles = fs.readFileSync(path.resolve(here, '../../shared/styles/agent-workspace.css'), 'utf8');

test('画布节点主体可直接拖动，只有真实交互控件阻止拖动', () => {
  assert.doesNotMatch(canvas, /dragHandle:\s*'\.agent-canvas-node-drag-handle'/);
  assert.match(canvas, /className=\{`agent-canvas-node agent-canvas-node--\$\{nodeType\}/);
  assert.match(canvas, /className="nodrag nopan agent-canvas-node-card-header-action"[\s\S]*?onClick=\{openEditor\}/);
  assert.match(canvas, /className="agent-canvas-node-image-select agent-canvas-node-drag-surface"/);
  assert.match(canvas, /AuthenticatedAgentImage[\s\S]*?draggable=\{false\}/);
  assert.match(canvas, /AuthenticatedAgentVideo[\s\S]*?className="agent-canvas-node-media agent-canvas-node-drag-surface"[\s\S]*?draggable=\{false\}/);
  assert.match(canvas, /className="agent-canvas-node-audio-preview agent-canvas-node-drag-surface"/);
  assert.match(canvas, /AuthenticatedAgentAudio[\s\S]*?className="nodrag nopan"/);
  assert.match(canvas, /className="agent-canvas-node-text-preview agent-canvas-node-drag-surface"/);
  assert.doesNotMatch(canvas, /agent-canvas-node-(?:image-select|video-placeholder|audio-placeholder|text-preview)[^"\n]*nodrag/);
  assert.doesNotMatch(canvas, /closest\?\.\('\.agent-node-composer,[^']*agent-canvas-node-(?:image-select|video-placeholder|audio-placeholder|text-preview)/);
  assert.match(styles, /agent-canvas-node-drag-handle[^}]*cursor:\s*grab/);
  assert.match(styles, /agent-canvas-node-drag-surface[^}]*cursor:\s*grab/);
});

test('画布连线把手始终可见，不能依赖悬停才能开始连线', () => {
  assert.match(styles, /agent-canvas-node-quick-add[^}]*opacity:\s*1/);
  assert.match(styles, /react-flow__handle\.agent-canvas-node-quick-add[^}]*pointer-events:\s*auto/);
});

test('分组外框不拦截画布操作，但收起和定位按钮仍可点击', () => {
  assert.match(styles, /agent-canvas-group[^}]*pointer-events:\s*none/);
  assert.match(styles, /agent-canvas-group\s+\.ant-btn[^}]*pointer-events:\s*auto/);
});

test('节点和生成框使用稳定的 Liblib 对标内容层级', () => {
  assert.match(canvas, /AuthenticatedAgentAudio[\s\S]*waveform/);
  assert.match(canvas, /agent-canvas-node-audio-preview/);
  assert.match(styles, /agent-canvas-node-image-placeholder,[\s\S]*agent-canvas-node-video-placeholder/);
  assert.match(styles, /agent-audio-waveform__bars/);
  assert.match(styles, /agent-node-composer__prompt[^}]*min-height:\s*78px/);
  assert.match(styles, /agent-node-composer__bottom-bar[^}]*overflow-x:\s*auto/);
});

test('选中的音频节点使用统一快捷处理条', () => {
  assert.match(canvas, /AgentAudioQuickTools/);
  assert.match(canvas, /operation: 'trim'/);
  assert.match(canvas, /operation: 'custom_split'/);
  assert.match(canvas, /operation: 'smart_split'/);
  assert.match(canvas, /operation: 'speed'/);
  assert.match(canvas, /downloadTitle="下载音频"/);
  assert.match(styles, /agent-node-quick-tools/);
});
