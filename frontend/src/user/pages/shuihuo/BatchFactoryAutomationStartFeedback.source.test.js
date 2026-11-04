import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');

test('explains why immediate start is unavailable while a batch automation is active', () => {
  assert.ok(/\['running', 'paused', 'scheduled', 'needs_attention'\]\.includes\(automationState\)/.test(source), 'the UI must match the backend active-task states');
  assert.ok(/已有任务待处理，请先在任务\/日志中处理当前任务/.test(source), 'the disabled action must explain how to proceed');
});

test('imports the content capture helper used by the content editor', () => {
  assert.ok(/import \{[^}]*contentCaptureCharactersForBook[^}]*\} from '\.\/batchFactoryContentRange'/s.test(source), 'the content editor helper must be imported');
});

test('automation menu exposes the preset chooser instead of presenting it as a schedule-only action', () => {
  assert.match(source, /\{ key: 'start', label: '启动自动生产（选择预设）', disabled: automationActive \|\| automationStartBlocked \}/);
  assert.match(source, /key === 'start' \? openAutomationStart\(\) : runAutomationAction\(key\)/);
  assert.match(source, /placeholder="选择已保存预设"/);
  assert.match(source, /runAutomationAction\('start', \{ immediate: true \}\)/);
});

test('immediate start asks for explicit confirmation before launching', () => {
  assert.match(source, /function confirmImmediateAutomation\(\)/);
  assert.match(source, /Modal\.confirm\(/);
  assert.match(source, /是否自动提交 121：\$\{autoSubmit \? '是' : '否'\}/);
  assert.match(source, /okButtonProps: \{ danger: autoSubmit \}/);
});
