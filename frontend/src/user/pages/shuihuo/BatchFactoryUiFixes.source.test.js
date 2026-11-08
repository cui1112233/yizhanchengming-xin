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

test('自动化失败提示移出“编辑内容”列，改到整行底部通栏', () => {
  assert.match(list, /batch-factory-book-row-error-strip/);
  assert.match(list, /<BatchFactoryErrorNotice[\s\S]{0,1600}?重试失败小说/);
  assert.doesNotMatch(list, /message=\{\`\$\{automationBook\.stale ? '历史自动生产失败'/);
});

test('统一错误组件：人话摘要、折叠原文、复制按钮、可换行', () => {
  assert.match(parse, /上游要求参考图 ref_image_0/);
  assert.match(notice, /复制错误信息/);
  assert.match(notice, /展开原始信息/);
  assert.match(notice, /request_id/);
  assert.match(css, /\.batch-factory-error-notice-summary\s*\{[^}]*white-space:\s*normal;[^}]*overflow-wrap:\s*anywhere;[^}]*word-break:\s*break-word/s);
});

test('失败提示两段式：新失败只弹 4.5 秒 toast，持久记录压成默认收起的 28px 单行', () => {
  // 瞬时提示：只在新失败签名变化时弹，带“查看详情”，4.5 秒后自动消失。
  assert.match(list, /duration:\s*4\.5/);
  assert.match(list, /视频生成失败 · 参考图缺失/);
  assert.match(list, /查看详情/);
  assert.match(list, /notifiedAutomationFailureRef/);
  // 持久记录：默认收起、一句话 + 详情按钮，展开后才出现原始信息与重试入口。
  assert.match(list, /<div className="batch-factory-book-row-error-strip">[\s\S]{0,600}?lineMode/);
  assert.match(notice, /const canExpand = lineMode/);
  assert.match(notice, /batch-factory-error-notice-line-text/);
  assert.match(notice, /\{expanded \? '收起' : '详情'\}/);
  assert.match(notice, /const \[expanded, setExpanded\] = useState\(false\)/);
  // 视觉：左侧 3px 状态色 + 单行 28px，不再是整块面板。
  assert.match(css, /\.batch-factory-error-notice\.is-line\s*\{[^}]*min-height:\s*28px/s);
  assert.match(css, /border-left:\s*3px solid/s);
  // 敏感词提示不能被一起压扁。
  assert.doesNotMatch(list, /batch-factory-sensitive-review[\s\S]{0,80}?lineMode/);
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
