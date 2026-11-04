import test from 'node:test';
import assert from 'node:assert/strict';
import {
  describeWebSubmitExecution,
  describeWebSubmitReconcile,
  describeWebSubmitVisible,
  webSubmitFailureLines
} from './webSubmitFeedback.js';

test('失败分组大于 0 时绝不提示成功', () => {
  const result = {
    ok: false,
    success_groups: 1,
    failed_groups: 2,
    results: [
      { bookId: 'book-1', version: 'ai1', status: 'submitted' },
      { bookId: 'book-2', version: 'ai1', status: 'error', error: '正文少于最少字数' },
      { bookId: 'book-3', version: 'ai1', status: 'pending_confirmation', error: '等待回读确认' }
    ]
  };
  const feedback = describeWebSubmitExecution(result, '自动提交完成');
  assert.equal(feedback.level, 'warning');
  assert.match(feedback.text, /2 个版本未提交成功/);
  assert.match(feedback.text, /book-2\/ai1：正文少于最少字数/);
  assert.doesNotMatch(feedback.text, /自动提交完成/);
});

test('全部失败时也绝不提示成功', () => {
  const feedback = describeWebSubmitExecution({
    ok: false,
    success_groups: 0,
    failed_groups: 1,
    results: [{ bookId: 'book-1', version: 'ai1', status: 'error', error: '任务不存在' }]
  }, '自动提交完成');
  assert.equal(feedback.level, 'warning');
  assert.match(feedback.text, /1 个版本未提交成功/);
});

test('只有 failed_groups===0 且确有成功提交才提示成功', () => {
  const feedback = describeWebSubmitExecution({
    ok: true,
    success_groups: 2,
    failed_groups: 0,
    results: [{ bookId: 'book-1', version: 'ai1', status: 'submitted' }]
  }, '自动提交完成');
  assert.equal(feedback.level, 'success');
  assert.equal(feedback.text, '自动提交完成');
});

test('全部跳过时提示没有新提交，不报完成', () => {
  const feedback = describeWebSubmitExecution({
    ok: true,
    success_groups: 0,
    failed_groups: 0,
    results: [{ bookId: 'book-1', version: 'ai1', status: 'skipped' }]
  }, '自动提交完成');
  assert.equal(feedback.level, 'warning');
  assert.match(feedback.text, /没有产生新的提交/);
});

test('非提交类响应（如组织目录）保持原有成功文案', () => {
  const feedback = describeWebSubmitExecution({ ok: true, organizations: [{ id: 'org-1' }] }, '已同步组织归属');
  assert.equal(feedback.level, 'success');
  assert.equal(feedback.text, '已同步组织归属');
});

test('登录检查失败时给出可操作的重新登录文案', () => {
  const feedback = describeWebSubmitVisible({
    ok: false,
    result: { checks: { login_session: false, upload_endpoint: '/upload.php' } }
  });
  assert.equal(feedback.level, 'error');
  assert.match(feedback.text, /登录已失效/);
  assert.match(feedback.text, /重新登录并验证/);
});

test('登录检查返回 ok 但会话检查为 false 时也不报可用', () => {
  const feedback = describeWebSubmitVisible({
    ok: true,
    result: { checks: { login_session: false } }
  });
  assert.equal(feedback.level, 'error');
  assert.doesNotMatch(feedback.text, /登录可用/);
});

test('登录检查通过时才提示可用', () => {
  const feedback = describeWebSubmitVisible({ ok: true, result: { checks: { login_session: true } } });
  assert.equal(feedback.level, 'success');
  assert.equal(feedback.text, '视频管理系统登录可用');
});

test('核对待回读时提示等待确认，不报核对完成', () => {
  const feedback = describeWebSubmitReconcile({
    ok: false,
    success_groups: 0,
    failed_groups: 1,
    results: [{ bookId: 'book-1', version: 'ai1', status: 'pending_confirmation' }]
  }, 'ai1 核对完成');
  assert.equal(feedback.level, 'warning');
  assert.match(feedback.text, /仍有 1 个版本等待目标站回读确认/);
  assert.doesNotMatch(feedback.text, /核对完成/);
});

test('失败项摘取最多三条，避免提示刷屏', () => {
  const lines = webSubmitFailureLines({
    results: [
      { bookId: 'a', status: 'error', error: 'e1' },
      { bookId: 'b', status: 'error', error: 'e2' },
      { bookId: 'c', status: 'error', error: 'e3' },
      { bookId: 'd', status: 'error', error: 'e4' }
    ]
  });
  assert.equal(lines.length, 3);
});
