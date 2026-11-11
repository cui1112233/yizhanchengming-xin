// 「自定义预设」选择逻辑的五场景测试：正常 / 空值 / 重复 / 错误 / 取消。
// 只验证纯逻辑；写服务端与界面交互见 docs/acceptance 的实测记录。
import test from 'node:test';
import assert from 'node:assert/strict';

import {
  buildPresetSelectionPayload,
  explicitPresetIds,
  filterPresetsBySelection,
  selectedPresetIds
} from './apiPresetSelection.js';

const PRESETS = [
  { id: 'deepseek', label: 'DeepSeek' },
  { id: 'doubao', label: '豆包' },
  { id: 'kimi', label: 'Kimi' }
];
const ALL = PRESETS.map(preset => preset.id);

test('正常输入：未记录时展示全部，勾选后只展示勾选项', () => {
  assert.deepEqual(selectedPresetIds(undefined, 'text', ALL), ALL);
  const selection = { text: { enabled: ['kimi', 'deepseek'] } };
  // 顺序跟随预置清单，避免卡片顺序随勾选顺序跳动
  assert.deepEqual(selectedPresetIds(selection, 'text', ALL), ['deepseek', 'kimi']);
  assert.deepEqual(filterPresetsBySelection(PRESETS, selection, 'text').map(item => item.id), ['deepseek', 'kimi']);
  // 其他分组不受影响
  assert.deepEqual(selectedPresetIds(selection, 'video', ['seedance']), ['seedance']);
});

test('空值：空对象、坏结构回落全部；显式空数组表示全部取消', () => {
  for (const empty of [undefined, null, {}, { text: null }, { text: { enabled: null } }, { text: { enabled: 'deepseek' } }]) {
    assert.equal(explicitPresetIds(empty, 'text'), null);
    assert.deepEqual(selectedPresetIds(empty, 'text', ALL), ALL);
  }
  const cleared = { text: { enabled: [] } };
  assert.deepEqual(explicitPresetIds(cleared, 'text'), []);
  assert.deepEqual(filterPresetsBySelection(PRESETS, cleared, 'text'), []);
  // 空白 id 视为无效，不占位
  assert.deepEqual(explicitPresetIds({ text: { enabled: ['  ', ''] } }, 'text'), []);
});

test('重复操作：重复勾选同一项不会产生重复记录，连续保存结果一致', () => {
  const first = buildPresetSelectionPayload(undefined, 'text', ['deepseek', 'deepseek', 'kimi']);
  assert.deepEqual(first, { text: { enabled: ['deepseek', 'kimi'] } });
  const second = buildPresetSelectionPayload(first, 'text', ['deepseek', 'kimi']);
  assert.deepEqual(second, first);
  // 重复点击“恢复默认”不改变结果
  assert.deepEqual(buildPresetSelectionPayload(first, 'text', ALL), buildPresetSelectionPayload(first, 'text', ALL));
});

test('错误操作：未知 id 被丢弃，数字 id 归一化为字符串，不污染其他分组', () => {
  const selection = { text: { enabled: ['deepseek', 'not-exist', 42] }, video: { enabled: ['seedance'] } };
  assert.deepEqual(selectedPresetIds(selection, 'text', ALL), ['deepseek']);
  assert.deepEqual(selectedPresetIds({ text: { enabled: [42] } }, 'text', ['42']), ['42']);
  const next = buildPresetSelectionPayload(selection, 'text', ['doubao', '']);
  assert.deepEqual(next, { text: { enabled: ['doubao'] }, video: { enabled: ['seedance'] } });
});

test('取消操作：纯函数不修改入参，取消后原选择保持不变', () => {
  const original = { text: { enabled: ['deepseek'] } };
  const snapshot = JSON.stringify(original);
  const next = buildPresetSelectionPayload(original, 'text', ['doubao']);
  assert.notEqual(next, original);
  assert.equal(JSON.stringify(original), snapshot);
  assert.deepEqual(selectedPresetIds(original, 'text', ALL), ['deepseek']);
  // 取消（不调用 onConfirm）时没有任何写入，原对象仍是旧勾选
  assert.equal(JSON.stringify(original), snapshot);
});
