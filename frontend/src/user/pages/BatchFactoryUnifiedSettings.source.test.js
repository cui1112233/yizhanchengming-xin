import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const novelList = fs.readFileSync(new URL('./shuihuo/BatchFactoryNovelList.jsx', import.meta.url), 'utf8');
const unifiedModal = fs.readFileSync(new URL('./shuihuo/BatchFactoryUnifiedSettingsModal.jsx', import.meta.url), 'utf8');
const stylePicker = fs.readFileSync(new URL('./shuihuo/BatchFactoryComicStylePicker.jsx', import.meta.url), 'utf8');

test('失败提示按批次汇总且不再逐本弹 toast', () => {
  assert.match(novelList, /batch-factory-failure-summary-\$\{batch\?\.id \|\| 'batch'\}/);
  assert.doesNotMatch(novelList, /key: `batch-factory-failure-\$\{book\.id\}`/);
  assert.match(novelList, /failureSummaryRef\.current\.signature === summarySignature/);
  assert.match(novelList, /collectNewAutomationFailures/);
});

test('保存统一配置后逐本校验 publishSettings 已落库', () => {
  assert.match(novelList, /pushKeys\.includes\('publishSettings'\)/);
  assert.match(novelList, /书级 publishSettings 未落库/);
});

test('组织归属不再宣称无条件自动带入', () => {
  assert.match(unifiedModal, /当前网站配置档未提供组织归属，提交时需在确认单选择/);
  assert.doesNotMatch(unifiedModal, /由网站配置档自动带入，不在这里重复选择。/);
  assert.match(unifiedModal, /if \(\(publish\.organization \|\| ''\) === organization\) return;/);
});

test('画风封面失败降级为占位并记忆失败源，不重复请求', () => {
  assert.match(stylePicker, /const failedPreviewSources = new Set\(\);/);
  assert.match(stylePicker, /failedPreviewSources\.has\(String\(src\)\)/);
  assert.match(stylePicker, /failedPreviewSources\.add\(String\(src\)\)/);
  assert.match(stylePicker, /ImageOff/);
});
