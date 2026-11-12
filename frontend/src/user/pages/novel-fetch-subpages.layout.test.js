import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'NovelFetchPage.jsx'), 'utf8');
const css = fs.readFileSync(path.join(here, 'novel-fetch-subpages.css'), 'utf8');
const petCss = fs.readFileSync(path.join(here, '../../shared/pet/cm-penguin-companion.css'), 'utf8');

test('处理规则先展示可直接选择的规则列表，再展示词组设置', () => {
  const listPosition = source.indexOf('className="novel-fetch-rule-list-panel"');
  const detailPosition = source.indexOf('className="novel-fetch-rule-detail-panel"');
  const groupPosition = source.indexOf('className="novel-fetch-rule-group-panel"');

  assert.notEqual(listPosition, -1, '缺少规则列表区域');
  assert.notEqual(detailPosition, -1, '缺少规则编辑区域');
  assert.notEqual(groupPosition, -1, '缺少词组设置区域');
  assert.ok(listPosition < groupPosition, '规则列表应排在词组设置前面');
  assert.ok(detailPosition < groupPosition, '词组设置应与规则编辑放在右侧操作区，避免左栏撑出大块空白');
  assert.match(source, /role="listbox" aria-label="规则列表"/);
  assert.match(source, /aria-selected=\{index === safeRuleIndex\}/);
});

test('右侧词组设置使用紧凑双列布局', () => {
  assert.match(css, /\.novel-fetch-rule-detail-panel \.novel-fetch-rule-group-fields\s*\{[^}]*display:\s*grid[^}]*grid-template-columns:\s*repeat\(2,minmax\(0,1fr\)\)/s);
  assert.match(css, /\.novel-fetch-rule-detail-panel \.novel-fetch-rule-group-checks\s*\{[^}]*grid-column:\s*1\/-1/s);
});

test('规则工作区完全复用应用主题变量', () => {
  assert.match(css, /\.novel-fetch-rule-workspace\s*\{[^}]*background:\s*var\(--legacy-card\)/s);
  assert.match(css, /\.novel-fetch-rule-list-panel\s*\{[^}]*border[^;]*var\(--legacy-border\)/s);
  assert.match(css, /\.novel-fetch-rule-list-button\s*\{[^}]*color:\s*var\(--text-primary\)/s);
  assert.match(css, /\.novel-fetch-rule-list-button\.is-selected\s*\{[^}]*background:\s*var\(--legacy-accent-dim\)/s);
  assert.match(css, /\.novel-fetch-rule-list\s*\{[^}]*max-height:\s*360px[^}]*overflow-y:\s*auto/s);
  assert.match(css, /\.novel-fetch-workbench \.novel-fetch-panel \.ant-input\s*\{[^}]*background:\s*var\(--legacy-input\)\s*!important[^}]*color:\s*var\(--text-primary\)\s*!important/s);
  assert.doesNotMatch(css, /\.novel-fetch-rule-(?:workspace|list-panel|list-button|group-panel)[^{]*\{[^}]*(?:#[0-9a-f]{3,8}|rgba?\()/is);
});

test('小说获取浅色主题覆盖旧深色控件和处理面板', () => {
  assert.match(css, /\[data-theme='light'\] \.novel-fetch-workbench \.ant-input-number[^{]*\{[^}]*background:\s*var\(--legacy-input\)\s*!important/s);
  assert.match(css, /\[data-theme='light'\] \.novel-fetch-workbench \.ant-select-selector[^{]*\{[^}]*background:\s*var\(--legacy-input\)\s*!important/s);
  assert.match(css, /\[data-theme='light'\] \.novel-fetch-workbench \.novel-fetch-process-side[^{]*\{[^}]*background:\s*var\(--legacy-card\)/s);
  assert.match(css, /\[data-theme='light'\] \.novel-fetch-public-heading \.ant-typography[^{]*\{[^}]*color:\s*var\(--text-primary\)/s);
  assert.match(css, /\[data-theme='light'\] \.novel-fetch-task-selection-bar \.ant-typography[^{]*\{[^}]*color:\s*var\(--text-link\)/s);
});

test('任务表选中行在深浅主题下保持可读且固定列颜色一致', () => {
  assert.match(css, /\[data-theme='dark'\] \.novel-fetch-task-table \.ant-table-tbody > tr\.ant-table-row-selected > td[^{]*\{[^}]*background:\s*var\(--status-info-bg\)\s*!important[^}]*color:\s*var\(--text-primary\)\s*!important/s);
  assert.match(css, /\[data-theme='dark'\] \.novel-fetch-task-table \.ant-table-tbody > tr\.ant-table-row-selected:hover > td[^{]*\{[^}]*background:\s*var\(--status-info-bg\)\s*!important/s);
  assert.match(css, /\[data-theme='light'\] \.novel-fetch-task-table \.ant-table-tbody > tr\.ant-table-row-selected > td[^{]*\{[^}]*background:\s*var\(--status-info-bg\)\s*!important[^}]*color:\s*var\(--text-primary\)\s*!important/s);
  assert.match(css, /\[data-theme='dark'\] \.novel-fetch-task-table \.ant-table-row-selected \.ant-table-cell-fix-(?:left|right)[^{]*\{[^}]*background:\s*var\(--status-info-bg\)\s*!important/s);
  assert.match(css, /\[data-theme='light'\] \.novel-fetch-task-table \.ant-table-row-selected \.ant-table-cell-fix-(?:left|right)[^{]*\{[^}]*background:\s*var\(--status-info-bg\)\s*!important/s);
  assert.doesNotMatch(css, /\.novel-fetch-task-table \.ant-table-row-selected \.ant-table-cell-fix-(?:left|right)[^{]*\{[^}]*background:\s*(?:rgba|transparent)/s);
  assert.match(css, /\.novel-fetch-task-table \.ant-table-row-selected \.ant-typography[^{]*\{[^}]*color:\s*var\(--text-primary\)\s*!important/s);
});

test('任务页使用紧凑字号且不影响其他小说获取子页面', () => {
  assert.match(css, /\.novel-fetch-task-page\s*\{[^}]*font-size:\s*13px/s);
  assert.match(css, /\.novel-fetch-task-page \.ant-(?:btn|input|select-selection-item|table|tabs-tab|pagination)[^{]*\{[^}]*font-size:\s*13px/s);
  assert.match(source, /<NovelFetchSubpageShell className="novel-fetch-task-subpage" title="任务"/);
  assert.match(css, /\.novel-fetch-task-subpage \.novel-fetch-subpage-heading h3\.ant-typography\s*\{[^}]*font-size:\s*20px/s);
  assert.match(css, /\.novel-fetch-queue-summary[^}]*font-size:\s*12px/s);
});

test('知识库使用紧凑工作台并把低频提示词收进设置抽屉', () => {
  assert.match(source, /className="novel-fetch-knowledge-workbench"/);
  assert.match(source, /className="novel-fetch-knowledge-editor-scroll"/);
  assert.match(source, /className="novel-fetch-knowledge-editor-actions"/);
  assert.match(source, /title="知识库设置"/);
  assert.match(source, /open=\{knowledgeSettingsOpen\}/);
  assert.doesNotMatch(source, /className="novel-fetch-panel novel-fetch-knowledge-prompt"/);
});

test('知识库完整目录保留旧版全部分类，并可进入对应资料清单', () => {
  assert.match(source, />完整目录</);
  ['爆款开头', '适用风格', '改文模板', '临时指令', '高仿提示词', '高仿参考', '敏感词', '字位符号', '成对补符号', '章节整行', '章节开头'].forEach(label => {
    assert.match(source, new RegExp(`label: '${label}'`));
  });
  assert.match(source, /collection:[ \t]*'profiles'/);
  assert.match(source, /collection:[ \t]*'temporary_instructions'/);
  assert.match(source, /collection:[ \t]*'prompts'/);
  assert.match(source, /collection:[ \t]*'references'/);
});

test('知识库工作台压缩空白并保持移动端单栏', () => {
  assert.match(source, /className=\{`novel-fetch-knowledge-list-viewport/);
  assert.match(css, /\.novel-fetch-knowledge-workbench\s*\{[^}]*border:\s*1px solid var\(--legacy-border\)/s);
  assert.match(css, /\.novel-fetch-knowledge-layout\s*\{[^}]*grid-template-columns:\s*300px minmax\(0,1fr\)/s);
  assert.match(css, /\.novel-fetch-knowledge-library\s*\{[^}]*align-self:\s*start/s);
  assert.match(css, /\.novel-fetch-knowledge-editor-actions\s*\{[^}]*position:\s*sticky[^}]*bottom:\s*0/s);
  assert.match(css, /@media \(max-width:760px\)[\s\S]*\.novel-fetch-knowledge-layout\s*\{[^}]*grid-template-columns:\s*1fr/s);
});

test('知识库资料列表提供位置提示和前后切换', () => {
  assert.match(source, /aria-label="上一条资料"/);
  assert.match(source, /aria-label="下一条资料"/);
  assert.match(source, /className="novel-fetch-knowledge-list-position"/);
});

test('知识库左栏与圆柱导航遵循紧凑主题工作台', () => {
  assert.match(css, /\.novel-fetch-knowledge-layout\s*\{[^}]*grid-template-columns:\s*300px minmax\(0,1fr\)/s);
  assert.match(css, /\.novel-fetch-knowledge-list-position\s*\{[^}]*text-align:\s*center/s);
  assert.match(css, /\.novel-fetch-knowledge-list-actions\s*\{[^}]*display:\s*grid/s);
  assert.doesNotMatch(css, /\.novel-fetch-knowledge-(?:workbench|layout|library)\s*\{[^}]*(?:#[0-9a-f]{3,8}|rgba?\()/is);
});

test('配置页将长提示词放进独立分组', () => {
  const panelStart = source.indexOf('function NovelFetchConfigPanel');
  const panelEnd = source.indexOf('function TaskDetailDrawer');
  const panelSource = source.slice(panelStart, panelEnd);
  const workflowStart = panelSource.indexOf("<section hidden={section !== 'workflow'}>");
  const promptsStart = panelSource.indexOf("<section hidden={section !== 'prompts'}>");
  const supplementStart = panelSource.indexOf("<section hidden={section !== 'supplement'}>");

  assert.match(panelSource, /key: 'prompts', label: '提示词配置'/);
  assert.ok(promptsStart > workflowStart, '提示词配置应位于工作流分组之后');
  assert.ok(supplementStart > promptsStart, '提示词配置应位于补充设置之前');

  const workflowSource = panelSource.slice(workflowStart, promptsStart);
  const promptsSource = panelSource.slice(promptsStart, supplementStart);
  assert.doesNotMatch(workflowSource, /改文提示词|处理规则提示词|知识库使用提示词/);
  assert.match(promptsSource, /name=\{\['rewrite', 'prompt'\]\}/);
  assert.match(promptsSource, /name=\{\['rewrite', 'processing_rule_prompt'\]\}/);
  assert.match(promptsSource, /name=\{\['knowledge', 'usage_prompt'\]\}/);
});

test('小说获取进入单列布局时收起会遮挡表单的宠物', () => {
  assert.match(petCss, /@media \(max-width:\s*760px\)[\s\S]*body:has\(\.novel-fetch-page\) \.cm-penguin-shell\s*\{\s*display:\s*none;\s*\}/s);
});

test('处理页不再保留解析预览步骤', () => {
  assert.doesNotMatch(source, /<strong>解析预览<\/strong>|>解析预览<\/Button>/);
  assert.doesNotMatch(source, /previewWorkshopProcess/);
  assert.match(source, /normalizeNovelFetchInput/);
  assert.match(source, /有效 <b>\{localInputPreview\.validRows\}<\/b> 本/);
  assert.match(source, /\['原文就绪', processSummary\.originalDone\]/);
});

test('任务页提供搜索、默认视图、全量批处理和单本诊断入口', () => {
  assert.match(source, /placeholder="搜索书籍 ID 或书名"/);
  assert.match(source, />默认视图</);
  assert.match(source, />重试全部失败</);
  assert.match(source, />规则处理全部</);
  assert.match(source, />提交全部</);
  assert.match(source, />删除失败</);
  assert.match(source, />删除全部</);
  assert.match(source, /label: `敏感日志/);
  assert.match(source, /label: '提交日志'/);
  assert.match(source, />进入批量工厂</);
});

test('提交准备页分别提供登录可视验证和组织目录同步', () => {
  assert.match(source, /getWebSubmitOrganizations/);
  assert.match(source, /testWebSubmitVisible/);
  assert.match(source, /async function syncWebSubmitOrganizations\(/);
  assert.match(source, /onClick=\{syncWebSubmitOrganizations\}>同步组织归属<\/Button>/);
  assert.match(source, />验证登录可用性<\/Button>/);
});
