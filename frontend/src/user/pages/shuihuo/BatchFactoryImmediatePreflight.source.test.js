import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');
// 管理员模型页路径以真实路由表为准，避免硬编码一个猜出来的地址。
const adminApp = fs.readFileSync(path.join(here, '../../../admin/App.jsx'), 'utf8');

test('每个立即执行入口都先拉一次只读预检，再带 skipPreflight 真正提交', () => {
  assert.match(source, /import \{ preflightBatchAutomation \} from '\.\.\/\.\.\/\.\.\/shared\/api\/batchFactoryAutomationPreflight\.js'/);
  assert.match(source, /if \(key === 'start' && immediate && !skipPreflight\) \{/);
  assert.match(source, /await preflightBatchAutomation\(batch\.id, \{/);
  assert.match(source, /runMode: automationRunMode/);
  assert.match(source, /skipPreflight: true/);
  // 预检只在这一处发起，没有定时器或轮询包装。
  const calls = source.match(/await preflightBatchAutomation\(/g) || [];
  assert.equal(calls.length, 1);
  assert.doesNotMatch(source, /setInterval\([\s\S]{0,200}preflightBatchAutomation/);
});

test('立即执行按钮不再直接弹确认框，而是打开逐本质检面板', () => {
  assert.match(source, /function confirmImmediateAutomation\(\) \{[\s\S]*void openAutomationPreflight\(\{ bookIds: targetBookIDs, presetId: automationPresetID \}\);/);
  assert.match(source, /<BatchFactoryPreflightModal/);
  assert.match(source, /onRun=\{confirmPreflightRun\}/);
  assert.match(source, /onFix=\{handleAutomationPreflightFix\}/);
  assert.match(source, /setAutomationPreflightOpen\(false\)/);
});

test('质检面板的修复入口分别指向统一配置、单书设置和模型管理页', () => {
  assert.match(source, /setUnifiedSettingsOpen\(true\)/);
  assert.match(source, /setConfigTarget\(\{ book: targetBook \}\)/);
  assert.match(adminApp, /pathname === '\/admin\/shuihuo-models'/);
  assert.match(source, /window\.location\.assign\('\/admin\/shuihuo-models'\)/);
  assert.match(source, /void openAutomationStart\(\)/);
});

test('定时执行的本地快检按漫画配置键判断发布绑定，不再误伤漫画批次', () => {
  assert.match(source, /const publish = \(productionMode === 'comic' \? value\?\.comicPublishSettings : value\?\.publishSettings\) \|\| \{\};/);
  assert.match(source, /if \(key === 'start' && immediate && !skipPreflight\) \{[\s\S]*\r?\n    \}\r?\n    if \(key === 'start'\) \{/);
});
