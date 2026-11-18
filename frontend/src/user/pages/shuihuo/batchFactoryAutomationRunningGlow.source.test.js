import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');
const stylesheet = fs.readFileSync(path.join(here, '../shuihuo-production.css'), 'utf8');

// 自动生产执行中要在书籍行上给出可感知的状态：整行静态光边，按阶段换色，不播放动画。
test('shows a static stage-coloured glow on rows that automation is running', () => {
  assert.match(source, /const automationRunningStage = automationState === 'running' && String\(automationBook\?\.status \|\| ''\)\.toLowerCase\(\) === 'running'/);
  assert.match(source, /is-automation-running/);
  assert.match(source, /data-automation-stage=\{automationRunningStage \|\| undefined\}/);
  assert.match(stylesheet, /\.batch-factory-book-row\.is-automation-running::before/);
  assert.match(stylesheet, /\[data-automation-stage='comic_bootstrap'\]/);
  assert.match(stylesheet, /\[data-automation-stage='comic_images'\]/);
  assert.ok(!stylesheet.includes('bf-book-flow'), '执行中的光带不应再播放流动动画');
});
