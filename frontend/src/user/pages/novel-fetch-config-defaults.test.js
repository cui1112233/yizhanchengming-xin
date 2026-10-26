import assert from 'node:assert/strict';
import test from 'node:test';
import React from 'react';
import { Form } from 'antd';
import { renderToStaticMarkup } from 'react-dom/server';

import { NovelFetchAutomationOptions } from './novel-fetch-automation-options.js';
import { withNovelFetchConfigDefaults } from './novel-fetch-config-defaults.js';

test('旧配置缺少自动处理开关时默认开启', () => {
  const result = withNovelFetchConfigDefaults({});

  assert.deepEqual(result.workflow, {
    auto_classify_missing: true,
    auto_fetch_original: true,
    auto_rewrite_after_fetch: true
  });
  assert.equal(result.fetch.auto_detect_platform, true);
  assert.equal(result.sensitive_ai.enabled, true);
  assert.deepEqual(result.sensitive_ai, {
    enabled: true,
    group_hits: true,
    concurrency: 4,
    max_hits_per_task: 80,
    context_chars: 12,
    retries: 1,
    temperature: 0.2
  });
  assert.deepEqual(result.web_submit.version_profiles, { original: '', ai1: '', ai2: '', ai3: '', ai4: '', ai5: '' });
});

test('敏感词 AI 参数保留用户已保存的 REA 字段', () => {
  const result = withNovelFetchConfigDefaults({ sensitive_ai: { group_hits: false, concurrency: 7, max_hits_per_task: 12, context_chars: 20, retries: 3, temperature: 0.7 } });
  assert.equal(result.sensitive_ai.group_hits, false);
  assert.equal(result.sensitive_ai.concurrency, 7);
  assert.equal(result.sensitive_ai.max_hits_per_task, 12);
  assert.equal(result.sensitive_ai.context_chars, 20);
  assert.equal(result.sensitive_ai.retries, 3);
  assert.equal(result.sensitive_ai.temperature, 0.7);
});

test('用户明确关闭的自动处理开关保持关闭', () => {
  const result = withNovelFetchConfigDefaults({
    workflow: {
      auto_classify_missing: false,
      auto_fetch_original: false,
      auto_rewrite_after_fetch: false
    },
    fetch: { auto_detect_platform: false },
    sensitive_ai: { enabled: false }
  });

  assert.equal(result.workflow.auto_classify_missing, false);
  assert.equal(result.workflow.auto_fetch_original, false);
  assert.equal(result.workflow.auto_rewrite_after_fetch, false);
  assert.equal(result.fetch.auto_detect_platform, false);
  assert.equal(result.sensitive_ai.enabled, false);
});

test('自动处理选项把表单值绑定到真实复选框', () => {
  const enabledMarkup = renderToStaticMarkup(React.createElement(
    Form,
    { initialValues: withNovelFetchConfigDefaults({}) },
    React.createElement(NovelFetchAutomationOptions)
  ));
  const disabledMarkup = renderToStaticMarkup(React.createElement(
    Form,
    { initialValues: withNovelFetchConfigDefaults({
      workflow: {
        auto_classify_missing: false,
        auto_fetch_original: false,
        auto_rewrite_after_fetch: false
      },
      fetch: { auto_detect_platform: false },
      sensitive_ai: { enabled: false }
    }) },
    React.createElement(NovelFetchAutomationOptions)
  ));

  assert.equal((enabledMarkup.match(/type="checkbox"/g) || []).length, 5);
  assert.equal((enabledMarkup.match(/checked=""/g) || []).length, 5);
  assert.equal((disabledMarkup.match(/checked=""/g) || []).length, 0);
});
