import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const canvasSource = fs.readFileSync(path.resolve(dirname, '../components/AgentCanvas.jsx'), 'utf8');
const workspaceStyles = fs.readFileSync(path.resolve(dirname, '../../shared/styles/agent-workspace.css'), 'utf8');

test('media nodes render one temporary inline editor controlled by preview activation', () => {
  assert.match(canvasSource, /const \[activeEditorNodeID, setActiveEditorNodeID\] = useState\(''\)/);
  assert.match(canvasSource, /isEditorOpen=\{activeEditorNodeID === props\.id\}/);
  assert.match(canvasSource, /onOpenEditor=\{setActiveEditorNodeID\}/);
  assert.match(canvasSource, /const generationPanel = isEditorOpen/);
  assert.match(canvasSource, /\{generationPanel\}/);
  assert.match(canvasSource, /setActiveEditorNodeID\(''\)/);
});

test('image preview opens editing while a named editor action opens the asset picker', () => {
  assert.match(canvasSource, /aria-label="展开图片编辑"/);
  assert.match(canvasSource, /aria-label=\{data\.assetRef\?\.url \? '更换图片素材' : '添加图片素材'\}/);
  assert.doesNotMatch(canvasSource, /aria-label="选择图片素材"[\s\S]*?onClick=\{event => \{ stopCanvasEvent\(event\); onPickImageAsset/);
});

test('only text nodes use the manual double-click editor', () => {
  assert.match(canvasSource, /if \(normalizeCanvasNodeType\(node\.data\?\.type\) !== 'text'\) return;/);
});

test('loaded and serialized nodes both use the canonical type', () => {
  assert.match(canvasSource, /type:\s*normalizeCanvasNodeType\(node\.type\)/);
  assert.match(canvasSource, /type:\s*normalizeCanvasNodeType\(node\.data\.type\)/);
  assert.ok(workspaceStyles.includes('.agent-page .agent-canvas-generation-panel'));
});

test('expanded video editor retains reachable create cancel retry and confirm actions', () => {
  for (const label of ['生成视频', '取消视频任务', '重试视频任务', '确认视频结果']) {
    assert.match(canvasSource, new RegExp(`aria-label="${label}"`));
  }
  assert.match(canvasSource, /onSubmitVideo\?\.\(id\)/);
  assert.match(canvasSource, /onCancelVideo\?\.\(id\)/);
  assert.match(canvasSource, /onRetryVideo\?\.\(id\)/);
  assert.match(canvasSource, /onConfirmVideo\?\.\(id\)/);
  assert.match(canvasSource, /videoState=\{videoNodeViewState\(/);
});

test('canvas restores project video jobs and rejects stale project responses', () => {
  assert.match(canvasSource, /listAgentVideoJobs\(requestedTaskId\)/);
  assert.match(canvasSource, /shouldAcceptVideoJobResponse\(\{[\s\S]*?requestedTaskId[\s\S]*?currentTaskId:/);
  assert.match(canvasSource, /mergeLatestVideoJobsByNode\(current,/);
});
