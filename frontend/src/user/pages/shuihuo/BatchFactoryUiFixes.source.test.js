import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const list = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');
const notice = fs.readFileSync(path.join(here, 'BatchFactoryErrorNotice.jsx'), 'utf8');
const parse = fs.readFileSync(path.join(here, 'batchFactoryErrorParse.js'), 'utf8');
const novelFetch = fs.readFileSync(path.join(here, '..', 'NovelFetchPage.jsx'), 'utf8');
const css = fs.readFileSync(path.join(here, '..', 'shuihuo-production.css'), 'utf8');

test('P1-A 重试失败步骤：点击即 loading，成功/失败都有 toast，并区分 503/409/网络', () => {
  assert.match(list, /message\.loading\(\{ content: '正在提交重试…'/);
  assert.match(list, /已提交，正在重跑…/);
  assert.match(list, /function retryFailureMessage\(error\)/);
  assert.match(list, /status === 503[\s\S]{0,120}?状态服务暂不可用/);
  assert.match(list, /status === 409[\s\S]{0,120}?HTTP 409/);
  assert.match(list, /Failed to fetch|NetworkError/);
  assert.match(list, /await retryBookStage\(batch\.id, book\.id/);
  assert.match(list, /正在提交重试…[\s\S]{0,4000}?await retryBookStage\(batch\.id/s);
});

test('P1-B 503 退避轮询（5s→15s→30s）与面板可重试按钮', () => {
  assert.match(list, /BACKOFF_MS = \[5000, 15000, 30000\]/);
  assert.match(list, /runtimeStatusFailureRef\.current/);
  assert.match(list, /状态服务暂不可用，请稍后重试/);
  assert.match(list, /onRetry=\{\(\) => loadRuntimeStatus\(\)\}/);
  assert.match(list, /<Button size="small" onClick=\{onRetry\}>重试<\/Button>/);
});

test('P2 敏感词 AI 修复入口保留，并改为打开确认 Modal', () => {
  assert.match(novelFetch, /敏感词 AI 修复/);
  assert.match(novelFetch, /Modal\.confirm\(\{[\s\S]{0,600}?敏感词 AI 修复/);
  assert.match(novelFetch, /当前没有可处理的任务，请先勾选任务/);
  assert.match(novelFetch, />敏感词 AI 修复<\/V88Button>/);
});

test('P2 统一配置面板：视口内高度、内容区内部滚动、sticky footer', () => {
  assert.match(css, /\.batch-factory-unified-settings-modal\s*\{[^}]*max-height:\s*calc\(100vh - 48px\)/s);
  assert.match(css, /\.batch-factory-unified-settings-modal \.ant-modal-body\s*\{[^}]*overflow-y:\s*auto/s);
  assert.match(css, /\.batch-factory-unified-settings-modal \.ant-modal-footer\s*\{[^}]*position:\s*sticky;[^}]*bottom:\s*0/s);
  assert.match(css, /@media \(max-width: 960px\)[\s\S]{0,400}?batch-factory-unified-settings-modal/s);
});

test('P3 批次表操作列右侧固定且背景遮挡滚动内容', () => {
  assert.match(css, /\.batch-factory-novel-list-head > span:nth-child\(7\)[\s\S]{0,260}?position:\s*sticky;[\s\S]{0,120}?right:\s*0/s);
  assert.match(css, /box-shadow:\s*-10px 0 14px -10px/);
});

test('主表不重复展示自动化失败通栏，错误只保留在任务日志和单本详情', () => {
  assert.doesNotMatch(list, /batch-factory-book-row-error-strip/);
  assert.match(list, /function BatchLogs[\s\S]*?<BatchFactoryErrorNotice compact title="自动生产失败"[\s\S]*?function UploadNetwork/);
  assert.match(list, /state\.failedStage\?\.errorMessage[\s\S]{0,500}?<BatchFactoryErrorNotice/);
});

test('漫画合集已经发布后不再展示之前补图阶段的旧失败', () => {
  assert.match(list, /const visibleAutomationError = published \? '' : String\(book\.error \|\| ''\);/);
  assert.match(list, /visibleAutomationError \? ` · \$\{batchFactoryVisibleError\(visibleAutomationError\)\}` : ''/);
  assert.match(list, /visibleAutomationError \? <BatchFactoryErrorNotice compact title="自动生产失败"/);
});

test('统一错误组件：人话摘要、折叠原文、复制按钮、可换行', () => {
  assert.match(parse, /上游要求参考图 ref_image_0/);
  assert.match(notice, /复制错误信息/);
  assert.match(notice, /展开原始信息/);
  assert.match(notice, /request_id/);
  assert.match(css, /\.batch-factory-error-notice-summary\s*\{[^}]*white-space:\s*normal;[^}]*overflow-wrap:\s*anywhere;[^}]*word-break:\s*break-word/s);
});

test('新失败只弹临时提醒，查看详情直接打开任务日志', () => {
  assert.match(list, /duration:\s*3/);
  assert.match(list, /查看详情/);
  assert.match(list, /notifiedAutomationFailureRef/);
  assert.match(list, /onClick=\{\(\) => setLogsOpen\(true\)\}/);
  assert.doesNotMatch(list, /scrollIntoView/);
});

test('敏感内容提示在窄书籍卡片中保持正文可读，操作按钮另起一行', () => {
  assert.match(css, /\.batch-factory-sensitive-review\.ant-alert\s*\{[^}]*flex-wrap:\s*wrap/s);
  assert.match(css, /\.batch-factory-sensitive-review \.ant-alert-content\s*\{[^}]*min-width:\s*0/s);
  assert.match(css, /\.batch-factory-sensitive-review \.ant-alert-action\s*\{[^}]*flex:\s*0 0 100%;[^}]*width:\s*100%/s);
});

test('换开头变体：按 VIDEO 版本区分文案，并带前置条件的生成按钮', () => {
  assert.match(list, /'换开头变体尚未生成'/);
  assert.match(list, /其他版本已有 \$\{otherVersionOpeningCount\} 条/);
  assert.match(list, /currentOpeningCount === 0/);
  assert.match(list, /videos\.length < 2/);
  assert.match(list, /directorRevision/);
  assert.match(list, /openingEnabled !== true/);
  assert.match(list, /onGenerateOpening\?\.\(\)/);
  assert.match(list, /runBookStageAction\(promptBook, 'opening', 'force'\)/);
});
