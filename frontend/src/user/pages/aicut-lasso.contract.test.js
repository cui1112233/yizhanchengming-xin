import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const page = readFileSync(new URL('./AICutPage.jsx', import.meta.url), 'utf8');
const css = readFileSync(new URL('./aicut-page.css', import.meta.url), 'utf8');

test('时间线框选从轨道空白开始并写入统一多选状态', () => {
  assert.match(page, /beginTimelineLasso/);
  assert.match(page, /updateTimelineLasso/);
  assert.match(page, /finishTimelineLasso/);
  assert.match(page, /selectLassoClipKeys/);
  assert.match(page, /setSelectedClipKeys\(nextKeys\)/);
  assert.match(page, /data-lasso-keys/);
  assert.match(page, /data-lasso-locked/);
  assert.match(page, /aicut-timeline-lasso/);
  assert.match(css, /\.aicut-timeline-lasso/);
});
