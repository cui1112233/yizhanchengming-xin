// 异步回包覆盖本地输入（useEffect 重置）的守卫逻辑测试。
// 场景来源：API 设置页自动保存/渠道刷新/预设勾选；只验证纯逻辑，4001 实测见 docs/acceptance。
import test from 'node:test';
import assert from 'node:assert/strict';

import { createHydrationGuard, createRevisionTracker, sameText } from './hydrationGuard.js';

test('用户在输入中，异步回包到达，值不被覆盖', () => {
  const guard = createHydrationGuard();
  let inputValue = '';
  const applyServerValue = value => { if (guard.canHydrate()) inputValue = value; };

  inputValue = '用户正在输入的新 Key';
  guard.touch();
  applyServerValue('服务端旧的已保存值');
  assert.equal(inputValue, '用户正在输入的新 Key');

  // 只有关闭/重新打开后才允许再回灌一次
  guard.reset();
  applyServerValue('重新打开后的服务端值');
  assert.equal(inputValue, '重新打开后的服务端值');
});

test('取消勾选后，异步回包不会把勾选重置', () => {
  const guard = createHydrationGuard();
  let pending = ['deepseek', 'doubao', 'kimi'];
  const syncFromServer = ids => { if (guard.canHydrate()) pending = [...ids]; };

  pending = pending.filter(id => id !== 'deepseek');
  guard.touch();
  // 保存状态变化触发父组件重渲染，回包内容仍是旧的三项
  syncFromServer(['deepseek', 'doubao', 'kimi']);
  assert.deepEqual(pending, ['doubao', 'kimi']);
});

test('请求飞行中用户又改过：旧回包按版本丢弃', () => {
  const tracker = createRevisionTracker();
  const revisionAtRequest = tracker.current();
  tracker.bump(); // 用户删除或编辑了渠道
  assert.equal(tracker.isCurrent(revisionAtRequest), false);
  assert.equal(tracker.isCurrent(tracker.current()), true);
});

test('异步保存只清理刚提交的草稿，不清用户新输入', () => {
  const submitted = 'sk-old';
  assert.equal(sameText('sk-old', submitted), true);
  assert.equal(sameText('sk-new-typing', submitted), false);
  assert.equal(sameText(undefined, ''), true);
});
