import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');
const stylesheet = fs.readFileSync(path.join(here, '../shuihuo-production.css'), 'utf8');

// 自动生产执行中要在书籍行上给出可感知的状态：整行流动光边，按阶段换色。
test('flows a stage-coloured glow along rows that automation is running', () => {
  assert.match(source, /const automationRunningStage = String\(automationBook\?\.status \|\| ''\)\.toLowerCase\(\) === 'running'/);
  assert.match(source, /is-automation-running/);
  assert.match(source, /data-automation-stage=\{automationRunningStage \|\| undefined\}/);
  assert.match(stylesheet, /\.batch-factory-book-row\.is-automation-running::before/);
  assert.match(stylesheet, /\[data-automation-stage='comic_bootstrap'\]/);
  assert.match(stylesheet, /\[data-automation-stage='comic_images'\]/);
  assert.match(stylesheet, /@keyframes bf-book-flow/);
});
