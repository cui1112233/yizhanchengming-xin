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
  // 自动生产窗口与“立即执行已选”各自先发起一次只读预检。
  const calls = source.match(/await preflightBatchAutomation\(/g) || [];
  assert.equal(calls.length, 2);
  const selected = source.slice(source.indexOf('async function runSelectedProduction()'), source.indexOf('const effectiveExecutionConfig', source.indexOf('async function runSelectedProduction()')));
  assert.ok(selected.includes('await preflightBatchAutomation(batch.id'));
  assert.ok(selected.includes('book?.passed === true'));
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

test('文本批次预检使用文本规则，立即执行不受视频执行器开关阻止', () => {
  assert.ok(source.includes("if (productionMode === 'text') return textModelPreflightReasons(settings)"));
  assert.ok(source.includes("batchProductionMode === 'video' && automationRunMode !== 'storyboard_only'"));
});

test('定时执行的本地快检按漫画配置键判断发布绑定，不再误伤漫画批次', () => {
  assert.match(source, /const publish = \(productionMode === 'comic' \? value\?\.comicPublishSettings : value\?\.publishSettings\) \|\| \{\};/);
  assert.match(source, /if \(key === 'start' && immediate && !skipPreflight\) \{[\s\S]*\r?\n    \}\r?\n    if \(key === 'start'\) \{/);
});
