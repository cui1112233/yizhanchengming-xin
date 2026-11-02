import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const canvas = fs.readFileSync(path.join(dirname, 'AgentCanvas.jsx'), 'utf8');
const workspaceStyles = fs.readFileSync(path.resolve(dirname, '../../shared/styles/agent-workspace.css'), 'utf8');

test('导演台以画布入口和独立预演工作区共用同一份节点配置', () => {
  const desk = fs.readFileSync(path.join(dirname, 'AgentDirectorDeskNode.jsx'), 'utf8');
  assert.match(desk, /directorDeskConfig/);
  assert.match(desk, /预演参考/);
  assert.match(desk, /onConfigChange/);
  assert.match(desk, /onPointerDown/);
  assert.match(desk, /captureDirectorDeskScreenshot/);
  assert.match(desk, /onSaveScreenshot/);
  assert.match(desk, /AgentDirectorDeskWorkspace/);
  assert.match(desk, /打开 3D 导演台/);
  assert.match(canvas, /AgentDirectorDeskNode/);
  assert.match(canvas, /DirectorDeskApp/);
  assert.match(canvas, /createDirectorDeskHostAdapter/);
  assert.match(canvas, /directorDeskNodeID/);
  assert.match(canvas, /key: 'director'/);
  assert.match(canvas, /uploadAgentAsset\(requestedTaskId, file\)/);
});

test('导演台节点可直接保存场景描述并通过项目素材库上传或选择参考图', () => {
  const desk = fs.readFileSync(path.join(dirname, 'AgentDirectorDeskNode.jsx'), 'utf8');
  assert.match(desk, /描述想要搭建的场景/);
  assert.match(desk, /scenePrompt/);
  assert.match(desk, /上传图片/);
  assert.match(desk, /保存构图/);
  assert.match(desk, /onPickReferenceAsset/);
  assert.match(desk, /onConfigChange\?\.\(node\.id, \{ directorDesk:/);
  assert.match(canvas, /onPickReferenceAsset=\{onPickReferenceAsset\}/);
  assert.match(canvas, /onConfigChange=\{onNodeConfigChange\}/);
});

test('导演台画布节点使用紧凑深色控件，而不是浏览器默认白色表单', () => {
  assert.match(workspaceStyles, /\.agent-page \.agent-director-desk-node\s*\{/);
  assert.match(workspaceStyles, /\.agent-page \.agent-director-desk-node__composer textarea\s*\{/);
  assert.match(workspaceStyles, /\.agent-page \.agent-director-desk-node__composer button\s*\{/);
  assert.match(workspaceStyles, /\.agent-page \.agent-director-desk-node__summary small\s*\{[^}]*overflow: hidden;/s);
  assert.match(workspaceStyles, /\.agent-page \.agent-director-desk-node__composer textarea\s*\{[^}]*background: var\(--agent-elevated\);/s);
});
