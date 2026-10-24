import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const canvas = fs.readFileSync(path.join(here, 'AgentCanvas.jsx'), 'utf8');
const styles = fs.readFileSync(path.resolve(here, '../../shared/styles/agent-workspace.css'), 'utf8');

test('画布节点只从明确的拖动把手开始移动，编辑入口不再占用拖动区域', () => {
  assert.match(canvas, /dragHandle:\s*'\.agent-canvas-node-drag-handle'/);
  assert.match(canvas, /className="agent-canvas-node-drag-handle" aria-label="拖动节点"/);
  assert.match(canvas, /className="agent-canvas-node-card-header-action"[\s\S]*?onClick=\{openEditor\}/);
  assert.match(styles, /agent-canvas-node-drag-handle[^}]*cursor:\s*grab/);
});

test('画布连线把手始终可见，不能依赖悬停才能开始连线', () => {
  assert.match(styles, /agent-canvas-node-quick-add[^}]*opacity:\s*1/);
  assert.match(styles, /react-flow__handle\.agent-canvas-node-quick-add[^}]*pointer-events:\s*auto/);
});

test('节点和生成框使用稳定的 Liblib 对标内容层级', () => {
  assert.match(canvas, /AuthenticatedAgentAudio[\s\S]*waveform/);
  assert.match(canvas, /agent-canvas-node-audio-preview/);
  assert.match(styles, /agent-canvas-node-image-placeholder,[\s\S]*agent-canvas-node-video-placeholder/);
  assert.match(styles, /agent-audio-waveform__bars/);
  assert.match(styles, /agent-node-composer__prompt[^}]*min-height:\s*78px/);
  assert.match(styles, /agent-node-composer__bottom-bar[^}]*overflow-x:\s*auto/);
});

test('选中的音频节点显示 Liblib 风格的快捷处理条', () => {
  assert.match(canvas, /AgentAudioQuickTools/);
  assert.match(canvas, /operation: 'trim'/);
  assert.match(canvas, /operation: 'custom_split'/);
  assert.match(canvas, /operation: 'smart_split'/);
  assert.match(canvas, /operation: 'speed'/);
  assert.match(canvas, /download title="下载音频"/);
  assert.match(styles, /agent-audio-quick-tools/);
});
