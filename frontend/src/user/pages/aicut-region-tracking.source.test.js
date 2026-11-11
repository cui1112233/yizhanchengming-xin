import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const source = readFileSync(new URL('./AICutPage.jsx', import.meta.url), 'utf8');
const inspector = readFileSync(new URL('./aicut/InspectorPanel.jsx', import.meta.url), 'utf8');
const overlay = readFileSync(new URL('./aicut/TrackingRegionOverlay.jsx', import.meta.url), 'utf8');

test('AI-CUT区域跟踪走真实取帧、灰度分析和统一关键帧命令', () => {
  assert.match(source, /extractAICutTrackingFrames/);
  assert.match(source, /trackRegionSequence/);
  assert.match(source, /trackingAbortRef/);
  assert.match(source, /clip\.tracking_keyframes_apply/);
  assert.match(source, /跟踪已写入/);
});

test('AI-CUT区域跟踪明确拒绝倒放、定格和曲线变速并支持取消', () => {
  assert.match(source, /倒放片段暂不支持区域跟踪/);
  assert.match(source, /定格片段不需要区域跟踪/);
  assert.match(source, /曲线变速片段暂不支持区域跟踪/);
  assert.match(source, /controller\.abort\(\)/);
  assert.match(source, /已取消区域跟踪/);
});

test('AI-CUT区域跟踪可在播放器直接框选、拖动和缩放目标', () => {
  assert.match(source, /<TrackingRegionOverlay/);
  assert.match(source, /trackingSelectionActive/);
  assert.match(inspector, /播放器框选目标/);
  assert.match(inspector, /完成框选/);
  assert.match(overlay, /begin\('draw'/);
  assert.match(overlay, /begin\('move'/);
  assert.match(overlay, /begin\('resize'/);
  assert.match(overlay, /目标跟踪框/);
});
