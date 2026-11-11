import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const source = readFileSync(new URL('./AICutPage.jsx', import.meta.url), 'utf8');
const inspector = readFileSync(new URL('./aicut/InspectorPanel.jsx', import.meta.url), 'utf8');
const overlay = readFileSync(new URL('./aicut/TrackingRegionOverlay.jsx', import.meta.url), 'utf8');

test('区域物体消除可在播放器框选并通过统一画面命令写入时间线', () => {
  assert.match(source, /objectRemovalSelectionActive/);
  assert.match(source, /commitObjectRemovalRegion/);
  assert.match(source, /clip\.visual_processing_set/);
  assert.match(source, /label="消除区域"/);
  assert.match(inspector, /播放器框选消除区域/);
  assert.match(inspector, /启用区域消除/);
  assert.match(overlay, /label = '跟踪目标'/);
});

test('区域物体消除说明预览和导出的真实边界而不是伪装成AI模型', () => {
  assert.match(inspector, /FFmpeg 周边像素修复/);
  assert.match(inspector, /适合固定水印/);
  assert.match(inspector, /移动物体请先使用画面区域跟踪/);
});
