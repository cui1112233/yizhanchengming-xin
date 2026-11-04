import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'NovelFetchPage.jsx'), 'utf8');
const css = fs.readFileSync(path.join(here, 'novel-fetch-v88.css'), 'utf8');

test('小说获取容器铺满并使用 V88 页面间距', () => {
  assert.match(css, /\.novel-fetch-workbench\s*\{[^}]*width:\s*100%/s);
  assert.match(css, /\.novel-fetch-workbench\s*\{[^}]*max-width:\s*none/s);
  assert.match(css, /\.novel-fetch-workbench\s*\{[^}]*padding:\s*var\(--v88-space-page\)/s);
  assert.match(css, /\.novel-fetch-workbench\s*\{[^}]*overflow-x:\s*hidden/s);
  assert.match(source, /className="novel-fetch-page novel-fetch-workbench"/);
});

test('小说获取接入 V88 主题，按钮走 ConfigProvider 令牌', () => {
  assert.match(source, /import \{ V88Button \} from '\.\.\/\.\.\/shared\/components\/V88Button'/);
  assert.match(source, /import \{ V88EmptyState \} from '\.\.\/\.\.\/shared\/components\/V88EmptyState'/);
  assert.match(source, /createV88AntTheme\('dark'\)/);
  assert.match(source, /<ConfigProvider theme=\{antTheme\}>/);
  assert.match(css, /\.ant-btn\.v88-btn--primary:not\(:disabled\)/);
  assert.match(css, /\.ant-btn\.v88-btn--danger:not\(:disabled\)/);
});

test('删除/停止类危险操作使用 danger 变体', () => {
  assert.match(source, /variant="danger"[\s\S]{0,240}?>停止选中<\/V88Button>/);
  assert.match(source, /variant="danger"[\s\S]{0,240}?>删除全部<\/V88Button>/);
  assert.match(source, /variant="danger"[\s\S]{0,240}?>删除失败<\/V88Button>/);
  assert.match(source, /variant="danger"[\s\S]{0,240}?>删除选中<\/V88Button>/);
  assert.match(source, /variant="danger"[\s\S]{0,240}?>删除<\/V88Button>/);
  assert.match(source, /variant=\{queueSummary\.enabled \? 'danger' : 'primary'\}/);
});

test('开始处理不再误用 danger，且 loading 绑定真实请求状态', () => {
  assert.doesNotMatch(source, /type="primary" danger loading=\{processing\} onClick=\{handleProcess\}/);
  assert.match(source, /variant="primary" loading=\{processing\} onClick=\{handleProcess\}/);
  assert.match(source, /loading=\{taskActionBusy === 'sensitive-repair'\}/);
  assert.match(source, /loading=\{taskActionBusy === 'retry-selected'\}/);
  assert.match(source, /loading=\{taskActionBusy === 'rules-all'\}/);
  assert.match(source, /loading=\{automationChanging\}/);
  assert.match(source, /loading=\{webSubmitLoading\}/);
  assert.doesNotMatch(source, /loading=\{true\}/);
});

test('空态统一为 V88EmptyState，且只在无数据时出现', () => {
  const emptyCount = (source.match(/<V88EmptyState/g) || []).length;
  assert.ok(emptyCount >= 8, `V88EmptyState 使用次数过少：${emptyCount}`);
  assert.match(source, /schedules\.length \? .* : <V88EmptyState/s);
  assert.match(source, /locale=\{\{ emptyText: <V88EmptyState/s);
  assert.match(source, /filtered\.length \? .* : <V88EmptyState/s);
  assert.doesNotMatch(source, /<Empty image=\{Empty\.PRESENTED_IMAGE_SIMPLE\}/);
});

test('假状态修复：无当前批次时状态卡不再声称「当前批次」', () => {
  assert.match(source, /const processScopeLabel = currentBatch\?\.id \? '当前批次' : '全部可见任务';/);
  assert.match(source, /className="novel-fetch-result-side">\{processScopeLabel\}</);
  assert.doesNotMatch(source, /className="novel-fetch-result-side">当前批次</);
});

test('动效使用 V88 令牌并尊重 prefers-reduced-motion', () => {
  assert.match(css, /transition:[^;]*var\(--v88-motion-base\)[^;]*var\(--v88-ease\)/s);
  assert.match(css, /@media \(prefers-reduced-motion: reduce\)/);
  assert.match(css, /transition-duration:\s*0ms/);
});

test('原有入口一个不少', () => {
  ['开始处理', '开始定时', '定时任务', '敏感词 AI 修复', '刷新任务', '暂停全自动', '开启全自动', '重试当前失败', '处理当前规则', '提交当前批次', '当前批次入厂', '重试全部失败', '规则处理全部', '提交全部', '删除失败', '删除全部', '全选当前结果', '清空选择', '立即重试选中', '生成 AI 改文', '调整选中', '规则处理选中', '进入批量工厂', '提交选中', '删除选中'].forEach(label => {
    assert.ok(source.includes(label), `缺少入口：${label}`);
  });
});
