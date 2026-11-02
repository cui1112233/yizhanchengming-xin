import React from 'react';
import { Checkbox, Form, Space } from 'antd';

const OPTIONS = [
  { name: ['workflow', 'auto_classify_missing'], label: '自动补齐风格/男女频' },
  { name: ['workflow', 'auto_fetch_original'], label: '自动抓取原文' },
  { name: ['workflow', 'auto_rewrite_after_fetch'], label: '抓取后自动生成 AI 文案' },
  { name: ['fetch', 'auto_detect_platform'], label: '平台自动识别（推荐）' },
  { name: ['sensitive_ai', 'enabled'], label: '启用敏感词 AI 修复' }
];

export function NovelFetchAutomationOptions() {
  return React.createElement(
    Space,
    { wrap: true },
    OPTIONS.map(option => React.createElement(
      Form.Item,
      { name: option.name, valuePropName: 'checked', noStyle: true, key: option.name.join('.') },
      React.createElement(Checkbox, null, option.label)
    ))
  );
}
