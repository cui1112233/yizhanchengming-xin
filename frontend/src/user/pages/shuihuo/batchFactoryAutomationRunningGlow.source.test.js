import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');
const stylesheet = fs.readFileSync(path.join(here, '../shuihuo-production.css'), 'utf8');

// 自动生产执行中的光效：整行静态微光定位这本书，流动光带围绕
// “当前阶段对应的格子”边框转动（阶段→格子映射见 AUTOMATION_STAGE_CELL）。
test('glows around the cell that matches the running automation stage', () => {
  assert.match(source, /const automationRunningStage = automationState === 'running' && String\(automationBook\?\.status \|\| ''\)\.toLowerCase\(\) === 'running'/);
  assert.match(source, /const automationActiveCell = AUTOMATION_STAGE_CELL\[automationRunningStage\] \|\| '';/);
  assert.match(source, /comic_assets: 'assets'/);
  assert.match(source, /comic_asset_images: 'assets'/);
  assert.match(source, /comic_bootstrap: 'prompts'/);
  assert.match(source, /comic_images: 'comic'/);
  assert.match(source, /comic_publish: 'actions'/);
  assert.match(source, /is-automation-cell/);
  assert.match(source, /data-automation-stage=\{automationRunningStage \|\| undefined\}/);
  assert.match(stylesheet, /\.batch-factory-book-row\.is-automation-running \.is-automation-cell::after/);
  assert.match(stylesheet, /conic-gradient\(from var\(--bf-flow-angle\)/);
  assert.match(stylesheet, /@keyframes bf-cell-flow/);
  assert.match(stylesheet, /\[data-automation-stage='comic_images'\]\{--bf-flow-color:#fbbf24\}/);
  assert.ok(!stylesheet.includes('.batch-factory-book-row.is-automation-running::before'), '整行顶部静态光条应已删除，动效只保留在格子上');
});
