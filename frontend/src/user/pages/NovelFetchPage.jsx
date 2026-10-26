import { Alert, Button, Checkbox, Collapse, Descriptions, Drawer, Dropdown, Empty, Form, Input, InputNumber, Modal, Select, Space, Table, Tabs, Tag, Tooltip, Typography, message } from 'antd';
import { CalendarClock, ChevronDown, ChevronUp, Download, Eye, Factory, RefreshCw, Settings2 } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import {
  checkWebSubmitEnvironment, getWebSubmitConfig, getWebSubmitHistory, getWebSubmitOrganizations, previewWebSubmit,
  saveWebSubmitConfig, startWebSubmit, syncWebSubmitConfigs, syncWebSubmitStyles, testWebSubmitVisible, uploadLogin, getUploadSession
} from '../../shared/api/novelFetch';
import {
  analyzeWorkshopOpening, deleteWorkshopKnowledge, deleteWorkshopTasks, fetchWorkshopOriginal,
  generateWorkshopAi, getWorkshopConfig, getWorkshopKnowledge, getWorkshopKnowledgeSummary,
  getWorkshopPlatforms, getWorkshopTask, getWorkshopTaskRulesTrace, getWorkshopTaskSensitiveLog, getWorkshopTaskSiteSubmitLog, getWorkshopJob, getWorkshopJobs, listWorkshopTasks, normalizeWorkshopOpening, optimizeWorkshopKnowledge,
  getWorkshopLogs, listWorkshopSchedules, deleteWorkshopSchedule, saveWorkshopSchedule, updateWorkshopSchedule, saveWorkshopBody,
  previewWorkshopRules, restoreWorkshopOriginal, retryWorkshopTasks, saveWorkshopConfig,
  updateWorkshopAiCount, applyWorkshopRules,
  saveWorkshopKnowledge, saveWorkshopOpening, startWorkshopProcess, startWorkshopAutomation, pauseWorkshopAutomation, suggestWorkshopRules, testWorkshopAi, reprocessWorkshopSensitive, getWorkshopRealtimeStatus,
  getWorkshopCurrentBatch, getWorkshopBatches, previewWorkshopProcess, rerunWorkshopBatch, stopWorkshopTasks
} from '../../shared/api/novelFetchWorkshop';
import { createNovelFetchIntake } from '../../shared/api/batchFactoryV11';
import { listAvailableModels as listAvailableTextModels, modelSelectOptions } from '../../shared/api/modelCatalog';
import { readNovelFetchProcessSettings, writeNovelFetchProcessSettings } from './novel-fetch-process-settings';
import { NovelFetchAutomationOptions } from './novel-fetch-automation-options.js';
import { withNovelFetchConfigDefaults, withNovelFetchVersionProfiles } from './novel-fetch-config-defaults.js';
import { novelFetchAiVersions, novelFetchSensitiveComparisons, novelFetchTaskSummaryItems } from './novel-fetch-task-detail.js';
import { novelFetchProgressLines } from './novel-fetch-progress.js';
import { beijingDateKey, currentBatchBookIds, filterNovelFetchTasks, mergeNovelFetchQueueStatus, novelFetchBatchResultSummary, novelFetchCharCount, novelFetchCurrentBatchTaskDetail, novelFetchOriginalFetchRetry, novelFetchPollingPlan, novelFetchQueueSummary, novelFetchStageLabel, novelFetchSubmissionLabel, novelFetchTaskDateKey, novelFetchTaskStatus, novelFetchVersionReady, retainVisibleTaskIds, stoppableCurrentBatchBookIds } from './novel-fetch-status.js';
import { singleVersionRetryPayload, submissionPreviewMatches } from './novel-fetch-submit.js';
import { mergeDefaultUploadProfile, uploadSessionDisplayName } from './novel-fetch-upload-profile.js';
import { NovelFetchCylinderList, NovelFetchMetricStrip, NovelFetchSubpageShell } from './novel-fetch-workbench-ui.jsx';
import { batchFactorySourceText, batchFactoryVersionLabel } from './novel-fetch-batch-factory.js';
import { normalizeNovelFetchInput } from './novel-fetch-input-format.js';
import { dispatchGlobalStatus } from '../../shared/notifications/globalTaskCenter.js';
import { getMemberCenter, getTeamMembers } from '../../shared/api/member.js';
import { listGlobalAccounts } from '../../shared/api/accountAdmin.js';
import { getSelectedNovelFetchAccount, novelFetchAccountOptions, setSelectedNovelFetchAccount } from '../../shared/api/novelFetchAccountScope.js';
import './novel-fetch.css';
import './novel-fetch-subpages.css';

const TABS = [
  { key: 'process', label: '处理' },
  { key: 'tasks', label: '任务' },
  { key: 'config', label: '配置' },
  { key: 'knowledge', label: '知识库' },
  { key: 'rules', label: '处理规则' },
  { key: 'submit', label: '提交' },
  { key: 'logs', label: '记录' }
];
const AI_VERSIONS = ['original', 'ai1', 'ai2', 'ai3', 'ai4', 'ai5'];
const AI_VERSION_LABELS = { original: '原文', ai1: 'AI1', ai2: 'AI2', ai3: 'AI3', ai4: 'AI4', ai5: 'AI5' };
const AUTO_REWRITE_METHODS = { ai1: 'auto', ai2: 'auto', ai3: 'auto', ai4: 'auto', ai5: 'auto' };
const INPUT_FORMAT_OPTIONS = [
  { value: 'smart', label: '智能识别', description: '每行输入 ID 和书名即可', columnPreset: '智能识别', columnOrder: [] },
  { value: 'paid_id_title_gender_reason', label: '付费 ID、书名、男女频、理由', description: '按制表符分列', columnPreset: '付费 ID + 书名 + 男女频 + 理由', columnOrder: ['bookId', 'bookName', 'gender', 'reason'] }
];

function scheduleInputValue(date = new Date(Date.now() + 10 * 60 * 1000)) {
  const pad = value => String(value).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

function formatScheduleTime(value) {
  if (!value) return '未设置';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return String(value);
  return date.toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' });
}

function scheduleStatusLabel(value) {
  return { scheduled: '待执行', dispatching: '正在入队', queued: '已入队', running: '执行中', done: '已完成', success: '已完成', completed: '已完成', upload_pending: '处理完成，待上传', failed: '失败', cancelled: '已取消', stopped: '已停止' }[value] || value || '待执行';
}

function diagnosticValue(value) {
  if (value === undefined || value === null || value === '') return '—';
  if (typeof value === 'object') return JSON.stringify(value, null, 2);
  return String(value);
}

function renderNovelFetchRulesDiagnostic(data) {
  const stages = Array.isArray(data?.stages) ? data.stages : [];
  if (!stages.length) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有保存规则轨迹" />;
  return <Collapse items={stages.map((stage, index) => {
    const item = stage && typeof stage === 'object' ? stage : { value: stage };
    const title = item.title || item.name || item.stage || `阶段 ${index + 1}`;
    const rows = Object.entries(item).filter(([key]) => key !== 'title' && key !== 'name' && key !== 'stage');
    return { key: String(index), label: title, children: rows.length ? <Descriptions size="small" bordered column={1} items={rows.map(([key, value]) => ({ key, label: key, children: <pre className="novel-fetch-diagnostic-value">{diagnosticValue(value)}</pre> }))} /> : <Typography.Text type="secondary">该阶段没有额外记录</Typography.Text> };
  })} />;
}

function renderNovelFetchSubmitDiagnostic(data) {
  const logs = Array.isArray(data?.logs) ? data.logs : [];
  const result = data?.result && typeof data.result === 'object' ? data.result : null;
  if (!logs.length && !result) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有提交记录" />;
  return <Space direction="vertical" size={12} style={{ width: '100%' }}>
    {result ? <Descriptions size="small" bordered column={{ xs: 1, sm: 2 }} items={Object.entries(result).map(([key, value]) => ({ key, label: key, children: diagnosticValue(value) }))} /> : null}
    {logs.length ? <Table size="small" pagination={{ pageSize: 10 }} rowKey={(row, index) => row.id || row.time || row.createdAt || index} dataSource={logs} columns={[{ title: '时间', render: (_, row) => row.time || row.createdAt || row.created_at || '—' }, { title: '状态', dataIndex: 'status', render: value => value || '—' }, { title: '版本', dataIndex: 'version', render: value => value || '—' }, { title: '结果', render: (_, row) => String(row.result || row.error || row.message || '').slice(0, 160) }]} /> : null}
  </Space>;
}
const PLATFORMS = [
  { value: 1, label: '黑岩付费 (1)' }, { value: 2, label: '番茄付费 (2)' }, { value: 3, label: '七猫付费 (3)' },
  { value: 4, label: '点众付费 (4)' }, { value: 7, label: '番茄免费 (7)' }, { value: 15, label: '知乎付费 (15)' },
  { value: 20, label: '掌阅付费 (20)' }, { value: 26, label: '卓越付费 (26)' }, { value: 29, label: '九州书城 (29)' }, { value: 31, label: '掌文付费 (31)' }
];
const KNOWLEDGE_TABS = [
  ['high_imitation', '高仿库'], ['opening_phrases', '爆款开头'], ['rewrite_templates', '改文模板'],
  ['layout_rules', '排版规则'], ['symbol_rules', '符号规则'], ['chapter_rules', '章节规则']
];
const KNOWLEDGE_LIBRARY_TYPES = [
  { value: 'high_imitation', label: '高仿文章库' },
  { value: 'opening_phrases', label: '爆款开头词库' },
  { value: 'rewrite_templates', label: '批量改文指令库' },
  { value: 'layout_rules', label: '排版规则' },
  { value: 'symbol_rules', label: '符号规则' },
  { value: 'chapter_rules', label: '章节规则' }
];

const KNOWLEDGE_COLLECTION_LABELS = {
  all: '全部条目',
  styles: '按适用风格查看',
  profiles: '改文模板',
  temporary_instructions: '临时指令',
  prompts: '高仿提示词',
  references: '高仿参考'
};

function knowledgeCollectionOptions(kind, library) {
  const items = Array.isArray(library?.items) ? library.items : [];
  if (kind === 'opening_phrases') {
    const styles = [...new Set(items.map(item => String(item.style || '').trim()).filter(Boolean))];
    return [
      { value: 'all', label: '全部开头词' },
      { value: 'styles', label: '按适用风格查看' },
      ...styles.map(style => ({ value: `style:${style}`, label: style }))
    ];
  }
  if (kind === 'rewrite_templates') return [
    { value: 'all', label: '全部改文指令' },
    { value: 'profiles', label: '改文模板' },
    { value: 'temporary_instructions', label: '临时指令' }
  ];
  if (kind === 'high_imitation') return [
    { value: 'all', label: '全部高仿资料' },
    { value: 'prompts', label: '高仿提示词' },
    { value: 'references', label: '高仿参考' }
  ];
  return [{ value: 'all', label: '全部条目' }];
}

function filterKnowledgeCollection(items, collection) {
  const source = Array.isArray(items) ? items : [];
  if (!collection || collection === 'all' || collection === 'styles') return source;
  if (collection.startsWith('style:')) return source.filter(item => String(item.style || '').trim() === collection.slice(6));
  return source.filter(item => String(item.collection || '') === collection);
}

function parseRuleEditorValue(value) {
  try {
    const parsed = JSON.parse(String(value || '{}'));
    return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? parsed : {};
  } catch {
    return {};
  }
}

function summarizeWorkshopProgress(rows, fallbackSubmissions = []) {
  const source = Array.isArray(rows) ? rows : [];
  const text = value => String(value ?? '').trim().toLowerCase();
  const isDone = value => /^(done|success|completed|complete|ready|ai_done|uploaded|已完成|完成|成功|已上传)$/.test(text(value));
  const isFailed = value => /fail|error|failed|失败|错误/.test(text(value));
  const originalDone = source.filter(row => isDone(row.originalStatus ?? row.original_status) || Number(row.originalRawChars ?? row.original_raw_chars) > 0).length;
  const aiDone = source.filter(row => isDone(row.aiStatus ?? row.ai_status) || Number(row.aiGeneratedCount ?? row.ai_generated_count) > 0 || (Array.isArray(row.aiFiles ?? row.ai_files) && (row.aiFiles ?? row.ai_files).length > 0)).length;
  const failed = source.filter(row => isFailed(row.status) || isFailed(row.originalStatus ?? row.original_status) || isFailed(row.aiStatus ?? row.ai_status) || isFailed(row.siteSubmitStatus ?? row.site_submit_status)).length;
  const processing = source.filter(row => ['queued', 'processing', 'waiting_retry'].includes(novelFetchTaskStatus(row).status)).length;
  const submissionDone = source.filter(row => isDone(row.siteSubmitStatus ?? row.site_submit_status)).length;
  const fallbackSubmissionDone = fallbackSubmissions.filter(row => isDone(row.status)).length;
  return {
    total: source.length,
    originalDone,
    aiDone,
    processing,
    failed,
    submissionDone: source.length ? submissionDone : fallbackSubmissionDone
  };
}

// 规则编辑器保留 V78 的业务边界：规则结构由 Go 保存，React 只编辑字段，不在前端复制处理逻辑。
function RuleEditor({ kind, value, onChange }) {
  const [groupIndex, setGroupIndex] = useState(0);
  const [ruleIndex, setRuleIndex] = useState(0);
  const [bulkText, setBulkText] = useState('');
  const config = useMemo(() => parseRuleEditorValue(value), [value]);
  const update = (next) => onChange(JSON.stringify(next, null, 2));
  const setPath = (path, nextValue) => {
    const next = { ...config };
    let target = next;
    path.slice(0, -1).forEach(key => { target[key] = { ...(target[key] || {}) }; target = target[key]; });
    target[path[path.length - 1]] = nextValue;
    update(next);
  };
  if (kind === 'sensitive') {
    const groups = Array.isArray(config.groups) ? config.groups : [{ name: '默认词组', enabled: true, apply_to_original: true, apply_to_ai: true, rules: [] }];
    const safeGroupIndex = Math.min(groupIndex, Math.max(groups.length - 1, 0));
    const group = groups[safeGroupIndex] || groups[0];
    const rules = Array.isArray(group?.rules) ? group.rules : [];
    const safeRuleIndex = Math.min(ruleIndex, Math.max(rules.length - 1, 0));
    const rule = rules[safeRuleIndex] || { find: '', replace: '', enabled: true };
    const setGroup = (key, nextValue) => {
      const nextGroups = groups.map((item, index) => index === safeGroupIndex ? { ...item, [key]: nextValue } : item);
      update({ ...config, groups: nextGroups });
    };
    const setRule = (key, nextValue) => {
      const nextGroups = groups.map((item, index) => index === safeGroupIndex ? { ...item, rules: rules.map((entry, itemIndex) => itemIndex === safeRuleIndex ? { ...entry, [key]: nextValue } : entry) } : item);
      update({ ...config, groups: nextGroups });
    };
    const addGroup = () => { const nextGroups = [...groups, { name: `词组${groups.length + 1}`, enabled: true, apply_to_original: true, apply_to_ai: true, rules: [] }]; update({ ...config, groups: nextGroups }); setGroupIndex(nextGroups.length - 1); setRuleIndex(0); };
    const deleteGroup = () => { if (groups.length <= 1) return; const nextGroups = groups.filter((_, index) => index !== safeGroupIndex); update({ ...config, groups: nextGroups }); setGroupIndex(Math.max(0, safeGroupIndex - 1)); setRuleIndex(0); };
    const addRule = () => { const nextRules = [...rules, { find: '', replace: '', enabled: true }]; const nextGroups = groups.map((item, index) => index === safeGroupIndex ? { ...item, rules: nextRules } : item); update({ ...config, groups: nextGroups }); setRuleIndex(nextRules.length - 1); };
    const deleteRule = () => { if (!rules.length) return; const nextRules = rules.filter((_, index) => index !== safeRuleIndex); const nextGroups = groups.map((item, index) => index === safeGroupIndex ? { ...item, rules: nextRules } : item); update({ ...config, groups: nextGroups }); setRuleIndex(Math.max(0, safeRuleIndex - 1)); };
    const addBulk = () => { const additions = bulkText.split(/\r?\n/).map(line => line.trim()).filter(Boolean).map(line => { const [find, replace = ''] = line.split('=>'); return { find: find.trim(), replace: replace.trim(), enabled: true }; }).filter(item => item.find); if (!additions.length) return; const nextGroups = groups.map((item, index) => index === safeGroupIndex ? { ...item, rules: [...rules, ...additions] } : item); update({ ...config, groups: nextGroups }); setBulkText(''); };
    return <div className="novel-fetch-rule-editor novel-fetch-rule-workspace">
      <Alert type="info" showIcon message="敏感词 AI 修复由处理页开关控制；这里维护词组、替换词和适用范围。AI 失败或关闭时使用替换词。" />
      <div className="novel-fetch-rule-workspace-body">
        <aside className="novel-fetch-rule-list-panel">
          <div className="novel-fetch-rule-section-heading"><div><Typography.Title level={5}>规则列表</Typography.Title><Typography.Text type="secondary">{group?.name || '未命名词组'} · {rules.length} 条规则</Typography.Text></div><Button aria-label="新增规则" onClick={addRule}>新增</Button></div>
          <div className="novel-fetch-rule-list" role="listbox" aria-label="规则列表">
            {rules.length ? rules.map((item, index) => <button type="button" role="option" aria-selected={index === safeRuleIndex} className={`novel-fetch-rule-list-button${index === safeRuleIndex ? ' is-selected' : ''}`} key={index} onClick={() => setRuleIndex(index)}><span>{item.find || `敏感词${index + 1}`}</span>{item.enabled === false ? <small>已停用</small> : null}</button>) : <div className="novel-fetch-rule-empty">暂无规则，点击“新增”开始配置</div>}
          </div>
          <Space wrap><Button onClick={addRule}>新增规则</Button><Button danger onClick={deleteRule} disabled={!rules.length}>删除规则</Button></Space>
        </aside>
        <section className="novel-fetch-rule-detail-panel">
          <div className="novel-fetch-rule-section-heading"><div><Typography.Title level={5}>规则编辑</Typography.Title><Typography.Text type="secondary">维护敏感词、替换词和启用状态</Typography.Text></div></div>
          <div className="novel-fetch-rule-fields">
            <label>查找内容 / 敏感词<input value={rule.find || ''} onChange={event => setRule('find', event.target.value)} /></label>
            <label>替换为（AI 失败或关闭时使用）<input value={rule.replace || ''} onChange={event => setRule('replace', event.target.value)} /></label>
            <label className="novel-fetch-check"><input type="checkbox" checked={rule.enabled !== false} onChange={event => setRule('enabled', event.target.checked)} />启用规则</label>
          </div>
          <label className="novel-fetch-rule-bulk">批量添加敏感词<textarea rows={5} value={bulkText} onChange={event => setBulkText(event.target.value)} placeholder="一行一个：敏感词 或 敏感词=>替换词" /></label>
          <Button onClick={addBulk}>批量加入当前词组</Button>
          <section className="novel-fetch-rule-group-panel">
            <Typography.Title level={5}>词组设置</Typography.Title>
            <div className="novel-fetch-rule-group-fields">
              <label>当前词组<select value={safeGroupIndex} onChange={event => { setGroupIndex(Number(event.target.value)); setRuleIndex(0); }}>{groups.map((item, index) => <option value={index} key={index}>{item.name || `词组${index + 1}`}</option>)}</select></label>
              <label>词组名称<input value={group?.name || ''} onChange={event => setGroup('name', event.target.value)} /></label>
              <div className="novel-fetch-rule-group-checks"><label className="novel-fetch-check"><input type="checkbox" checked={group?.enabled !== false} onChange={event => setGroup('enabled', event.target.checked)} />启用</label><label className="novel-fetch-check"><input type="checkbox" checked={group?.apply_to_original !== false} onChange={event => setGroup('apply_to_original', event.target.checked)} />处理原文</label><label className="novel-fetch-check"><input type="checkbox" checked={group?.apply_to_ai !== false} onChange={event => setGroup('apply_to_ai', event.target.checked)} />处理 AI</label></div>
            </div>
            <Space wrap><Button onClick={addGroup}>新增词组</Button><Button danger onClick={deleteGroup} disabled={groups.length <= 1}>删除词组</Button></Space>
          </section>
        </section>
      </div>
    </div>;
  }
  if (kind === 'layout_rules') return <div className="novel-fetch-rule-editor"><Alert type="info" showIcon message="排版规则只负责结构清洗；正文换行由字位符号规则处理。" /><div className="novel-fetch-rule-grid">{[['apply_to_original', '自动处理原文'], ['apply_to_ai', '自动处理 AI 文案'], ['apply_sensitive', '启用敏感词'], ['apply_chapter_cleanup', '启用章节清洗'], ['apply_symbol_rules', '启用字位符号'], ['apply_pair_fill', '启用成对补符号'], ['drop_empty_lines', '删除空行'], ['trim_lines', '清理行首尾空格']].map(([key, label]) => <label className="novel-fetch-check" key={key}><input type="checkbox" checked={config[key] !== false} onChange={event => setPath([key], event.target.checked)} />{label}</label>)}</div><div className="novel-fetch-rule-grid"><label>段首空格数<input type="number" value={config.indent_spaces ?? 0} onChange={event => setPath(['indent_spaces'], Number(event.target.value) || 0)} /></label><label>段落间空行数<input type="number" value={config.blank_lines_before ?? 0} onChange={event => setPath(['blank_lines_before'], Number(event.target.value) || 0)} /></label></div></div>;
  const listKey = kind === 'chapter_exact_rules' || kind === 'chapter_inline_rules' ? kind : kind;
  const list = Array.isArray(config[listKey]) ? config[listKey] : [];
  const index = Math.min(ruleIndex, Math.max(list.length - 1, 0));
  const item = list[index] || { find: '', replace: '', enabled: true };
  const setItem = (key, nextValue) => update({ ...config, [listKey]: list.map((entry, itemIndex) => itemIndex === index ? { ...entry, [key]: nextValue } : entry) });
  const addItem = () => { update({ ...config, [listKey]: [...list, { find: '', replace: '', enabled: true }] }); setRuleIndex(list.length); };
  const deleteItem = () => { if (!list.length) return; update({ ...config, [listKey]: list.filter((_, itemIndex) => itemIndex !== index) }); setRuleIndex(Math.max(0, index - 1)); };
  return <div className="novel-fetch-rule-editor novel-fetch-rule-workspace"><div className="novel-fetch-rule-workspace-body"><aside className="novel-fetch-rule-list-panel"><div className="novel-fetch-rule-section-heading"><div><Typography.Title level={5}>规则列表</Typography.Title><Typography.Text type="secondary">{kind.includes('chapter') ? '章节清洗' : '符号处理'} · {list.length} 条规则</Typography.Text></div><Button onClick={addItem}>新增</Button></div><div className="novel-fetch-rule-list" role="listbox" aria-label="规则列表">{list.length ? list.map((entry, itemIndex) => <button type="button" role="option" aria-selected={itemIndex === index} className={`novel-fetch-rule-list-button${itemIndex === index ? ' is-selected' : ''}`} key={itemIndex} onClick={() => setRuleIndex(itemIndex)}><span>{entry.find || `规则${itemIndex + 1}`}</span>{entry.enabled === false ? <small>已停用</small> : null}</button>) : <div className="novel-fetch-rule-empty">暂无规则，点击“新增”开始配置</div>}</div><Space wrap><Button onClick={addItem}>新增规则</Button><Button danger onClick={deleteItem} disabled={!list.length}>删除规则</Button></Space></aside><section className="novel-fetch-rule-detail-panel"><div className="novel-fetch-rule-section-heading"><div><Typography.Title level={5}>规则编辑</Typography.Title><Typography.Text type="secondary">维护命中内容、替换内容和启用状态</Typography.Text></div></div><div className="novel-fetch-rule-fields"><label>命中内容<input value={item.find || ''} onChange={event => setItem('find', event.target.value)} /></label><label>替换为<input value={item.replace || ''} onChange={event => setItem('replace', event.target.value)} /></label><label className="novel-fetch-check"><input type="checkbox" checked={item.enabled !== false} onChange={event => setItem('enabled', event.target.checked)} />启用规则</label></div></section></div></div>;
}

function initialTab() {
  const tab = new URLSearchParams(window.location.search).get('tab');
  return TABS.some(item => item.key === tab) ? tab : 'process';
}
function parseIds(value) {
  return [...new Set(String(value || '').split(/[\s,，;；]+/).map(item => item.trim()).filter(Boolean))];
}
function statusColor(value) {
  if (String(value || '').includes('fail')) return 'red';
  if (String(value || '').includes('done')) return 'green';
  return 'gold';
}

function TaskStatusTag({ row }) {
  const taskStatus = novelFetchTaskStatus(row);
  return <Tag color={taskStatus.color}>{taskStatus.label}</Tag>;
}
function downloadText(name, content) {
  const url = URL.createObjectURL(new Blob([content], { type: 'text/plain;charset=utf-8' }));
  const link = document.createElement('a');
  link.href = url;
  link.download = name;
  link.click();
  URL.revokeObjectURL(url);
}

function NovelFetchConfigPanel({ form, effectiveValues, dirty, onDirty, versionPanel, section, onSectionChange }) {
  const watchedValues = Form.useWatch([], form) || {};
  const sectionItems = [
    { key: 'workflow', label: '工作流与原文抓取', description: '抓取、改文和自动处理' },
    { key: 'prompts', label: '提示词配置', description: '改文、规则和知识库提示词' },
    { key: 'supplement', label: '补充设置', description: '任务保留和批量策略' },
    { key: 'ai', label: 'AI 接口参数', description: '模型、并发和输出控制' },
    { key: 'catalog', label: '平台与风格目录', description: '平台和风格 JSON 目录' },
    { key: 'versions', label: '版本配置档', description: '提交版本和配置档绑定' }
  ];
  const readPath = (source, path, fallback = '未设置') => {
    const value = path.reduce((current, key) => current?.[key], source);
    if (typeof value === 'boolean') return value ? '开启' : '关闭';
    return value === undefined || value === null || value === '' ? fallback : String(value);
  };
  const effectiveRows = [
    ['截取字数', ['fetch', 'default_max_txt']],
    ['抓取并发', ['fetch', 'concurrency']],
    ['默认 AI 数量', ['rewrite', 'default_ai_count']],
    ['文本模型', ['ai', 'model']],
    ['自动抓取原文', ['workflow', 'auto_fetch_original']]
  ];
  return <div className="novel-fetch-config-workspace">
    <nav className="novel-fetch-config-nav" aria-label="配置分组">
      {sectionItems.map(item => <button type="button" key={item.key} className={section === item.key ? 'is-active' : ''} onClick={() => onSectionChange(item.key)}><strong>{item.label}</strong><span>{item.description}</span></button>)}
    </nav>
    <main className="novel-fetch-config-main">
      <Form className="novel-fetch-panel novel-fetch-config-form" form={form} layout="vertical" onValuesChange={onDirty}>
      <section hidden={section !== 'workflow'}>
        <Typography.Title level={5}>工作流与原文抓取</Typography.Title>
        <div className="novel-fetch-config-grid">
          <Form.Item name={['fetch', 'endpoint']} label="原文接口地址"><Input placeholder="支持 {bookid}、{platform}、{max_txt}" /></Form.Item>
          <Form.Item name={['fetch', 'default_max_txt']} label="截取字数"><InputNumber min={1} max={100000} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['fetch', 'concurrency']} label="抓取并发"><InputNumber min={1} max={20} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['fetch', 'retries']} label="抓取重试"><InputNumber min={0} max={10} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['fetch', 'original_fetch_retry_interval_minutes']} label="抓取失败再试（分钟）"><InputNumber min={1} max={1440} placeholder="默认 60" style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['fetch', 'timeout_seconds']} label="抓取超时（秒）"><InputNumber min={1} max={300} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['rewrite', 'default_ai_count']} label="默认 AI 数量"><InputNumber min={1} max={20} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['rewrite', 'process_line_count']} label="默认处理行数"><InputNumber min={1} max={100} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['rewrite', 'anchor_line_count']} label="锚点行数"><InputNumber min={0} max={100} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['rewrite', 'temperature']} label="改文温度"><InputNumber min={0} max={2} step={0.01} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['rewrite', 'method_sequence_text']} label="出文轮换顺序"><Input placeholder="high_imitation,opening_instruction,instruction" /></Form.Item>
          <Form.Item name={['rewrite', 'default_template_id']} label="默认改文模板"><Input /></Form.Item>
          <Form.Item name={['rewrite', 'opening_phrase_mode']} label="开头词选择"><Select options={[{ value: 'auto', label: '自动匹配' }, { value: 'off', label: '不使用' }]} /></Form.Item>
          <Form.Item name={['rewrite', 'high_imitation_mode']} label="高仿参考选择"><Select options={[{ value: 'auto', label: '自动匹配' }, { value: 'off', label: '不使用' }]} /></Form.Item>
        </div>
        <NovelFetchAutomationOptions />
      </section>
      <section hidden={section !== 'prompts'}>
        <Typography.Title level={5}>提示词配置</Typography.Title>
        <div className="novel-fetch-config-prompts">
          <Form.Item name={['rewrite', 'prompt']} label="改文提示词"><Input.TextArea rows={6} /></Form.Item>
          <Form.Item name={['rewrite', 'processing_rule_prompt']} label="处理规则提示词"><Input.TextArea rows={6} /></Form.Item>
          <Form.Item name={['knowledge', 'usage_prompt']} label="知识库使用提示词"><Input.TextArea rows={5} /></Form.Item>
        </div>
      </section>
      <section hidden={section !== 'supplement'}>
        <Typography.Title level={5}>补充设置</Typography.Title>
        <div className="novel-fetch-config-grid">
          <Form.Item name={['workflow', 'min_original_chars']} label="原文最少字数"><InputNumber min={0} max={100000} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['workflow', 'skip_short_original']} valuePropName="checked" noStyle><label><input type="checkbox" /> 低于最少字数时保留原文并跳过改文</label></Form.Item>
          <Form.Item name={['workflow', 'auto_reclassify_invalid_style']} valuePropName="checked" noStyle><label><input type="checkbox" /> 风格不在当前目录时重新 AI 判断</label></Form.Item>
          <Form.Item name={['workflow', 'auto_sync_site_styles']} valuePropName="checked" noStyle><label><input type="checkbox" /> 处理前自动同步视频管理系统风格</label></Form.Item>
          <Form.Item name={['storage', 'auto_cleanup_enabled']} valuePropName="checked" noStyle><label><input type="checkbox" /> 自动清理过期已完成任务</label></Form.Item>
          <Form.Item name={['storage', 'cleanup_keep_days']} label="任务保留天数"><InputNumber min={1} max={365} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['workflow', 'auto_submit_batch_size']} label="自动提交每批任务数"><InputNumber min={1} max={1000} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['workflow', 'auto_submit_flush_seconds']} label="自动提交等待秒数"><InputNumber min={0} max={3600} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['ai', 'force_serial_batch']} valuePropName="checked" noStyle><label><input type="checkbox" /> AI 改文批量按顺序处理</label></Form.Item>
          <Form.Item name={['sensitive_ai', 'mode']} label="敏感词处理模式"><Select options={[{ value: 'direct', label: '直接替换' }, { value: 'per_line', label: 'AI逐条修改' }, { value: 'group', label: 'AI整组修改' }]} /></Form.Item>
        </div>
      </section>
      <section hidden={section !== 'ai'}>
        <Typography.Title level={5}>AI 接口参数</Typography.Title>
        <div className="novel-fetch-config-grid">
          <Form.Item name={['ai', 'base_url']} label="AI 基础接口"><Input /></Form.Item>
          <Form.Item name={['ai', 'model']} label="AI 模型"><Input /></Form.Item>
          <Form.Item name={['ai', 'timeout_seconds']} label="超时（秒）"><InputNumber min={1} max={900} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['ai', 'max_concurrency']} label="最大并发"><InputNumber min={1} max={50} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['ai', 'retry_times']} label="重试次数"><InputNumber min={0} max={10} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['ai', 'max_tokens']} label="最大 Token"><InputNumber min={1} max={100000} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['ai', 'temperature']} label="Temperature"><InputNumber min={0} max={2} step={0.01} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['ai', 'top_p']} label="Top P"><InputNumber min={0} max={1} step={0.01} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['ai', 'presence_penalty']} label="Presence Penalty"><InputNumber step={0.01} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name={['ai', 'frequency_penalty']} label="Frequency Penalty"><InputNumber step={0.01} style={{ width: '100%' }} /></Form.Item>
        </div>
        <Space wrap>
          <Form.Item name={['ai', 'stream']} valuePropName="checked" noStyle><label><input type="checkbox" /> 流式响应</label></Form.Item>
          <Form.Item name={['ai', 'json_mode']} valuePropName="checked" noStyle><label><input type="checkbox" /> JSON 模式</label></Form.Item>
          <Form.Item name={['ai', 'enable_thinking']} valuePropName="checked" noStyle><label><input type="checkbox" /> 启用思考</label></Form.Item>
          <Form.Item name={['ai', 'disable_thinking']} valuePropName="checked" noStyle><label><input type="checkbox" /> 禁用思考</label></Form.Item>
        </Space>
        <Form.Item name={['ai', 'extra_body_json']} label="额外 JSON"><Input.TextArea rows={5} className="mono" /></Form.Item>
      </section>
      <section hidden={section !== 'catalog'}>
        <Typography.Title level={5}>平台与风格目录</Typography.Title>
        <Form.Item name="platforms" label="平台目录（JSON）"><Input.TextArea rows={6} className="mono" placeholder="由 Go 配置接口返回的平台目录" /></Form.Item>
        <Form.Item name="styles" label="风格目录（JSON）"><Input.TextArea rows={6} className="mono" placeholder="由 Go 配置接口返回的风格目录" /></Form.Item>
      </section>
      </Form>
      {section === 'versions' ? versionPanel : null}
    </main>
    <aside className="novel-fetch-config-summary">
      <div className={`novel-fetch-config-state ${dirty ? 'is-dirty' : 'is-saved'}`}><strong>{dirty ? '有未保存修改' : '配置已同步'}</strong><span>{dirty ? '保存前不会影响后端当前配置' : '当前表单与已加载配置一致'}</span></div>
      <Typography.Title level={5}>当前选择 / 已生效</Typography.Title>
      {effectiveRows.map(([label, path]) => <div className="novel-fetch-effective-row" key={label}><span>{label}</span><strong>{readPath(watchedValues, path)}</strong><small>已生效：{readPath(effectiveValues, path)}</small></div>)}
    </aside>
  </div>;
}

function TaskDetailDrawer({ taskDetail, onClose, onSaved }) {
  const [originalText, setOriginalText] = useState('');
  const [saving, setSaving] = useState(false);
  const [diagnostics, setDiagnostics] = useState({ sensitive: null, rules: null, submit: null });
  const [diagnosticsLoading, setDiagnosticsLoading] = useState(false);
  useEffect(() => {
    const data = taskDetail?.data || {};
    setOriginalText(data.original || data.originalRaw || data.processedText || '');
    const bookId = taskDetail?.bookId;
    if (!bookId) {
      setDiagnostics({ sensitive: null, rules: null, submit: null });
      return undefined;
    }
    let cancelled = false;
    setDiagnosticsLoading(true);
    Promise.allSettled([
      getWorkshopTaskSensitiveLog(bookId),
      getWorkshopTaskRulesTrace(bookId),
      getWorkshopTaskSiteSubmitLog(bookId)
    ]).then(([sensitive, rules, submit]) => {
      if (cancelled) return;
      setDiagnostics({
        sensitive: sensitive.status === 'fulfilled' ? sensitive.value : { error: sensitive.reason?.message || '敏感日志读取失败' },
        rules: rules.status === 'fulfilled' ? rules.value : { error: rules.reason?.message || '规则轨迹读取失败' },
        submit: submit.status === 'fulfilled' ? submit.value : { error: submit.reason?.message || '提交日志读取失败' }
      });
      setDiagnosticsLoading(false);
    });
    return () => { cancelled = true; };
  }, [taskDetail]);
  async function saveOriginal() {
    if (!taskDetail?.bookId) return;
    setSaving(true);
    try {
      await saveWorkshopBody(taskDetail.bookId, 'original', originalText, 'ready');
      message.success('原文已保存');
      await onSaved?.();
    } catch (error) { message.error(error.message || '保存原文失败'); } finally { setSaving(false); }
  }
  const detailData = taskDetail?.data || {};
  const summaryItems = novelFetchTaskSummaryItems(detailData);
  const aiVersions = novelFetchAiVersions(detailData);
  const sensitiveComparison = novelFetchSensitiveComparisons(diagnostics.sensitive || {});
  const sensitiveSummary = sensitiveComparison.summary;
  const sensitiveContent = diagnosticsLoading
    ? <div className="novel-fetch-detail-loading">正在读取敏感日志…</div>
    : diagnostics.sensitive?.error
      ? <Alert type="error" showIcon message="敏感日志读取失败" description={diagnostics.sensitive.error} />
      : <div className="novel-fetch-sensitive-detail">
          <Descriptions className="novel-fetch-detail-summary" bordered size="small" column={{ xs: 1, sm: 2 }} items={[
            { key: 'mode', label: '处理模式', children: sensitiveSummary.mode },
            { key: 'status', label: '处理状态', children: sensitiveSummary.status },
            { key: 'hits', label: '命中数量', children: sensitiveSummary.hitCount },
            { key: 'fixed', label: '已处理数量', children: sensitiveSummary.fixedCount },
            { key: 'hit-time', label: '命中更新时间', children: sensitiveSummary.hitUpdatedAt },
            { key: 'fixed-time', label: '处理更新时间', children: sensitiveSummary.fixedUpdatedAt }
          ]} />
          {sensitiveComparison.items.length ? <div className="novel-fetch-sensitive-list">
            {sensitiveComparison.items.map(item => <section className="novel-fetch-sensitive-card" key={`${item.index}-${item.keyword}`}>
              <header className="novel-fetch-sensitive-card__head">
                <strong>#{item.index} {item.keyword}</strong>
                <Tag color={item.status === '已完成' ? 'success' : 'processing'}>{item.status}</Tag>
              </header>
              <div className="novel-fetch-sensitive-meta">
                <span><small>命中位置</small><strong>{item.position}</strong></span>
                <span><small>处理来源</small><strong>{item.source}</strong></span>
                <span><small>处理模型</small><strong>{item.model}</strong></span>
                <span><small>尝试次数</small><strong>{item.attempts}</strong></span>
              </div>
              {item.error !== '—' ? <Alert type="error" showIcon message={item.error} /> : null}
              <div className="novel-fetch-sensitive-compare">
                <div><span>命中片段</span><pre>{item.before}</pre></div>
                <div><span>处理后片段</span><pre>{item.after}</pre></div>
              </div>
            </section>)}
          </div> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={sensitiveSummary.hitCount > 0 ? '有命中计数，但这条旧任务没有保存片段明细' : '暂无敏感词命中记录'} />}
        </div>;
  return <Drawer open={Boolean(taskDetail)} title={taskDetail?.bookId ? `任务详情 · ${taskDetail.bookId}` : '任务详情'} width={860} onClose={onClose}>
    {taskDetail?.data ? <Tabs key={`${taskDetail.bookId}-${taskDetail.initialTab || 'summary'}`} defaultActiveKey={taskDetail.initialTab || 'summary'} items={[
      { key: 'summary', label: '任务概览', children: <Descriptions className="novel-fetch-detail-summary" bordered column={{ xs: 1, sm: 2 }} items={summaryItems.map(item => ({ ...item, children: item.value }))} /> },
      { key: 'original', label: '原文', children: <><Input.TextArea rows={18} value={originalText} onChange={event => setOriginalText(event.target.value)} /><Space style={{ marginTop: 12 }}><Button loading={saving} type="primary" onClick={saveOriginal}>保存原文</Button><Button onClick={() => setOriginalText(taskDetail.data.original || taskDetail.data.originalRaw || '')}>恢复已保存原文</Button></Space></> },
      { key: 'ai', label: `AI 文案${aiVersions.length ? `（${aiVersions.length}）` : ''}`, children: aiVersions.length ? <Tabs className="novel-fetch-ai-versions" type="card" items={aiVersions.map(version => ({ key: version.key, label: version.label, children: <div className="novel-fetch-ai-copy"><div className="novel-fetch-ai-copy__head"><strong>{version.label}</strong><span>{version.content.length.toLocaleString('zh-CN')} 字</span></div><pre>{version.content}</pre></div> }))} /> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前任务还没有生成 AI 文案" /> },
      { key: 'sensitive', label: `敏感日志${sensitiveSummary.hitCount ? `（${sensitiveSummary.fixedCount}/${sensitiveSummary.hitCount}）` : ''}`, children: sensitiveContent },
      { key: 'rules', label: '规则轨迹', children: diagnosticsLoading ? <Typography.Text type="secondary">正在读取规则轨迹…</Typography.Text> : diagnostics.rules?.error ? <Alert type="error" showIcon message="规则轨迹读取失败" description={diagnostics.rules.error} /> : renderNovelFetchRulesDiagnostic(diagnostics.rules || {}) },
      { key: 'submit', label: '提交日志', children: diagnosticsLoading ? <Typography.Text type="secondary">正在读取提交日志…</Typography.Text> : diagnostics.submit?.error ? <Alert type="error" showIcon message="提交日志读取失败" description={diagnostics.submit.error} /> : renderNovelFetchSubmitDiagnostic(diagnostics.submit || {}) },
      { key: 'logs', label: '操作日志', children: <Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }}>{JSON.stringify(taskDetail.data.logs || [], null, 2)}</Typography.Paragraph> }
    ]} /> : null}
  </Drawer>;
}

function KnowledgeLibraryManager({
  kind, onKindChange, collection, onCollectionChange, collectionOptions, items, loading, search, onSearchChange, selectedId, onSelect,
  draft, onDraftChange, onAdd, onDuplicate, onDelete, onNormalize, onSave, onOptimize,
  improveGoal, onImproveGoalChange
}) {
  const [jsonDraft, setJsonDraft] = useState('{}');
  const filtered = (items || []).filter(item => {
    const query = String(search || '').trim().toLowerCase();
    if (!query) return true;
    return `${item.name || item.title || ''} ${item.category || item.style || ''} ${Array.isArray(item.tags) ? item.tags.join(' ') : item.tags || ''}`.toLowerCase().includes(query);
  });
  const selectedIndex = filtered.findIndex(item => String(item.id) === String(selectedId));
  const currentIndex = Math.max(0, selectedIndex);
  const selectRelative = delta => {
    const nextIndex = Math.max(0, Math.min(filtered.length - 1, currentIndex + delta));
    const next = filtered[nextIndex];
    if (next) onSelect(String(next.id));
  };
  const setField = (field, value) => onDraftChange({ ...(draft || {}), [field]: value });
  const textList = value => Array.isArray(value) ? value.join('、') : String(value || '');
  useEffect(() => { setJsonDraft(JSON.stringify(draft || {}, null, 2)); }, [draft]);
  const applyJsonDraft = () => {
    try {
      const parsed = JSON.parse(jsonDraft || '{}');
      if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) throw new Error('完整字段必须是 JSON 对象');
      onDraftChange(parsed);
      message.success('完整字段已应用到当前草稿');
    } catch (error) { message.error(error.message || '完整字段 JSON 格式不正确'); }
  };
  return <div className="novel-fetch-knowledge-workbench">
    <div className="novel-fetch-knowledge-layout">
    <aside className="novel-fetch-knowledge-library">
      <div className="novel-fetch-knowledge-column-heading"><div><Typography.Title level={5}>资料列表</Typography.Title><Typography.Text type="secondary">{filtered.length} 条资料</Typography.Text></div><Button size="small" type="primary" onClick={onAdd}>新增</Button></div>
      <div className="novel-fetch-knowledge-controls">
        <label>资料库<Select value={kind} onChange={onKindChange} options={KNOWLEDGE_LIBRARY_TYPES} /></label>
        {collectionOptions?.length > 1 ? <label>分类<Select value={collection} onChange={onCollectionChange} options={collectionOptions} /></label> : null}
        <label>搜索<Input value={search} onChange={event => onSearchChange(event.target.value)} placeholder="按名称、分类、标签查找" /></label>
      </div>
      <div className={`novel-fetch-knowledge-list-viewport${filtered.length < 4 ? ' is-compact' : ''}`}>{loading ? <div className="novel-fetch-cylinder-empty">正在读取资料…</div> : filtered.length ? <NovelFetchCylinderList items={filtered} selectedId={selectedId} onSelect={value => onSelect(String(value))} getTitle={item => item.name || item.title || item.id || '未命名'} getMeta={item => `${item.style || item.category || '自动风格'}${Array.isArray(item.tags) && item.tags.length ? ` · ${item.tags.join('、')}` : ''}`} /> : <div className="novel-fetch-cylinder-empty">暂无条目</div>}</div>
      <div className="novel-fetch-knowledge-list-navigator">
        <Tooltip title="上一条资料"><Button aria-label="上一条资料" type="text" icon={<ChevronUp size={16} />} disabled={!filtered.length || currentIndex === 0} onClick={() => selectRelative(-1)} /></Tooltip>
        <span className="novel-fetch-knowledge-list-position">{filtered.length ? `${currentIndex + 1} / ${filtered.length}` : '0 / 0'}</span>
        <Tooltip title="下一条资料"><Button aria-label="下一条资料" type="text" icon={<ChevronDown size={16} />} disabled={!filtered.length || currentIndex >= filtered.length - 1} onClick={() => selectRelative(1)} /></Tooltip>
      </div>
      <div className="novel-fetch-knowledge-list-actions">
        <Button disabled={!selectedId} onClick={onDuplicate}>复制条目</Button>
        <Button danger disabled={!selectedId} onClick={onDelete}>删除条目</Button>
      </div>
      {kind === 'opening_phrases' ? <Button style={{ marginTop: 8 }} onClick={onNormalize}>矫正旧开头词库</Button> : null}
    </aside>
    <section className="novel-fetch-knowledge-editor">
      <div className="novel-fetch-knowledge-editor-heading"><div><Typography.Title level={5}>{draft?.name || draft?.title || '条目编辑'}</Typography.Title><Typography.Text type="secondary">{draft ? '编辑资料内容与调用范围' : '从左侧选择资料或新建条目'}</Typography.Text></div>{draft ? <Tag color={draft.enabled === false ? 'default' : 'success'}>{draft.enabled === false ? '已停用' : '已启用'}</Tag> : null}</div>
      {!draft ? <div className="novel-fetch-knowledge-editor-empty"><Typography.Text type="secondary">没有找到条目，可以新增一个。</Typography.Text><Button type="primary" onClick={onAdd}>新增资料</Button></div> : <>
        <div className="novel-fetch-knowledge-editor-scroll">
        {kind === 'high_imitation' ? <Alert type="info" showIcon message={draft?.collection === 'references' ? '高仿参考用于提供文风样本，只给“高仿文章”方案调用。' : '高仿提示词是单独一套改文方法，只给“高仿文章”方案调用。'} /> : null}
        {kind === 'opening_phrases' ? <Alert type="info" showIcon message="爆款开头词库会配合“批量改文指令库”使用，不单独替代整套改文指令。" /> : null}
        <div className="novel-fetch-knowledge-fields">
          <label>{kind === 'rewrite_templates' ? '模板ID' : '条目ID'}<Input value={draft.id || ''} onChange={event => setField('id', event.target.value)} /></label>
          <label>{kind === 'rewrite_templates' ? '模板名称' : '名称'}<Input value={draft.name || draft.title || ''} onChange={event => setField(draft.name !== undefined ? 'name' : 'title', event.target.value)} /></label>
          {kind !== 'rewrite_templates' ? <label>风格类型<Input value={draft.style || ''} onChange={event => setField('style', event.target.value)} /></label> : null}
          {kind === 'high_imitation' ? <label>男女频<Input value={draft.gender || ''} onChange={event => setField('gender', event.target.value)} /></label> : null}
          {kind === 'opening_phrases' ? <label>绑定指令模板 ID<Input value={textList(draft.linked_template_ids)} onChange={event => setField('linked_template_ids', event.target.value.split(/[、,，\s]+/).filter(Boolean))} /></label> : null}
          {kind === 'rewrite_templates' ? <label>返回方式<Select value={draft.response_mode || 'changed_lines_json'} onChange={value => setField('response_mode', value)} options={[{ value: 'changed_lines_json', label: '只返回改写块 JSON' }, { value: 'plain', label: '直接返回正文' }]} /></label> : null}
          <label>标签<Input value={textList(draft.tags)} onChange={event => setField('tags', event.target.value.split(/[、,，\s]+/).filter(Boolean))} /></label>
        </div>
        <label className="novel-fetch-knowledge-enabled"><Checkbox checked={draft.enabled !== false} onChange={event => setField('enabled', event.target.checked)}>启用当前资料</Checkbox></label>
        {kind === 'high_imitation' && draft?.collection === 'references' ? <label>高仿参考正文<Input.TextArea rows={9} value={draft.reference_text || draft.content || ''} onChange={event => setField(draft.reference_text !== undefined ? 'reference_text' : 'content', event.target.value)} /></label> : null}
        {kind === 'high_imitation' && draft?.collection !== 'references' ? <><label>系统提示词<Input.TextArea rows={7} value={draft.system_prompt || ''} onChange={event => setField('system_prompt', event.target.value)} /></label><label>用户提示词模板<Input.TextArea rows={7} value={draft.user_prompt_template || ''} onChange={event => setField('user_prompt_template', event.target.value)} /></label></> : null}
        {kind === 'opening_phrases' ? <><label>开头词骨架<Input.TextArea rows={8} value={draft.content || ''} onChange={event => setField('content', event.target.value)} /></label><label>槽位 / 适配提示<Input.TextArea rows={3} value={draft.slot_hint || ''} onChange={event => setField('slot_hint', event.target.value)} /></label></> : null}
        {kind === 'rewrite_templates' && draft?.collection !== 'temporary_instructions' ? <><label>系统提示词<Input.TextArea rows={7} value={draft.system_prompt || ''} onChange={event => setField('system_prompt', event.target.value)} /></label><label>用户提示词模板<Input.TextArea rows={7} value={draft.user_prompt_template || ''} onChange={event => setField('user_prompt_template', event.target.value)} /></label><label>可用变量<Input value={textList(draft.required_variables)} onChange={event => setField('required_variables', event.target.value.split(/[、,，\s]+/).filter(Boolean))} /></label></> : null}
        {kind === 'rewrite_templates' && draft?.collection === 'temporary_instructions' ? <label>临时指令内容<Input.TextArea rows={9} value={draft.content || ''} onChange={event => setField('content', event.target.value)} /></label> : null}
        {['layout_rules', 'symbol_rules', 'chapter_rules'].includes(kind) ? <><label>命中内容<Input value={draft.find || ''} onChange={event => setField('find', event.target.value)} /></label><label>替换内容<Input value={draft.replace || ''} onChange={event => setField('replace', event.target.value)} /></label><label>内容 / 提示词<Input.TextArea rows={6} value={draft.content || draft.prompt || ''} onChange={event => setField(draft.prompt !== undefined ? 'prompt' : 'content', event.target.value)} /></label></> : null}
        <Collapse className="novel-fetch-knowledge-advanced" ghost items={[
          { key: 'notes', label: '拆解备注与 AI 优化', children: <div className="novel-fetch-knowledge-advanced-fields">{['high_imitation', 'opening_phrases'].includes(kind) ? <label>拆解备注<Input.TextArea rows={4} value={draft.analysis || ''} onChange={event => setField('analysis', event.target.value)} /></label> : null}<label>AI 优化要求<Input.TextArea rows={3} value={improveGoal} onChange={event => onImproveGoalChange(event.target.value)} placeholder="例如：保留变量，增强短视频开头吸引力" /></label></div> },
          { key: 'json', label: '完整字段（JSON）', children: <div className="novel-fetch-knowledge-json"><Input.TextArea className="mono" rows={8} value={jsonDraft} onChange={event => setJsonDraft(event.target.value)} /><Button onClick={applyJsonDraft}>应用 JSON</Button></div> }
        ]} />
        </div>
        <div className="novel-fetch-knowledge-editor-actions"><Typography.Text type="secondary">修改只会在保存后生效</Typography.Text><Space><Button onClick={onOptimize}>AI 帮我优化</Button><Button type="primary" onClick={onSave}>保存修改</Button></Space></div>
      </>}
    </section>
    </div>
  </div>;
}

export function NovelFetchPage() {
  return <LegacyNovelFetchPage />;
}

export function LegacyNovelFetchPage() {
  const [processForm] = Form.useForm();
  const bookIdsValue = Form.useWatch('bookIds', processForm) || '';
  const [inputFormat, setInputFormat] = useState('smart');
  const [inputPreview, setInputPreview] = useState(null);
  const savedProcessSettings = useMemo(() => readNovelFetchProcessSettings(), []);
  const [configForm] = Form.useForm();
  const [knowledgeForm] = Form.useForm();
  const [activeTab, setActiveTab] = useState(initialTab);
  const [processRows, setProcessRows] = useState([]);
  const [platformOptions, setPlatformOptions] = useState(PLATFORMS);
  const [selectedPlatform, setSelectedPlatform] = useState(() => savedProcessSettings.platform || '');
  const [processJob, setProcessJob] = useState(null);
  const [processJobs, setProcessJobs] = useState([]);
  const [schedules, setSchedules] = useState([]);
  const [scheduleModalOpen, setScheduleModalOpen] = useState(false);
  const [editingScheduleId, setEditingScheduleId] = useState('');
  const [scheduleTime, setScheduleTime] = useState('');
  const [scheduleSaving, setScheduleSaving] = useState(false);
  const scheduleInputRef = useRef(null);
  const [globalLogs, setGlobalLogs] = useState([]);
  const [processing, setProcessing] = useState(false);
  const [tasks, setTasks] = useState([]);
  const [tasksLoading, setTasksLoading] = useState(false);
  const [queueStatus, setQueueStatus] = useState({ counts: {} });
  const [automationChanging, setAutomationChanging] = useState(false);
  const [selectedTaskKeys, setSelectedTaskKeys] = useState([]);
  const [selectedTaskData, setSelectedTaskData] = useState(null);
  const [taskDetail, setTaskDetail] = useState(null);
  const [ruleText, setRuleText] = useState('');
  const [ruleScope, setRuleScope] = useState('original');
  const [ruleKind, setRuleKind] = useState('sensitive');
  const [ruleConfigText, setRuleConfigText] = useState('{}');
  const [ruleResult, setRuleResult] = useState(null);
  const [ruleAiSample, setRuleAiSample] = useState('');
  const [ruleAiGoal, setRuleAiGoal] = useState('');
  const [ruleLoading, setRuleLoading] = useState(false);
  const [knowledgeKind, setKnowledgeKind] = useState('high_imitation');
  const [knowledge, setKnowledge] = useState({ items: [] });
  const [knowledgeSummary, setKnowledgeSummary] = useState({});
  const [knowledgeUsagePrompt, setKnowledgeUsagePrompt] = useState('');
  const [knowledgeLoading, setKnowledgeLoading] = useState(false);
  const [knowledgeModal, setKnowledgeModal] = useState(false);
  const [knowledgeSearch, setKnowledgeSearch] = useState('');
  const [knowledgeDetails, setKnowledgeDetails] = useState('{}');
  const [knowledgeSelectedId, setKnowledgeSelectedId] = useState('');
  const [knowledgeDraft, setKnowledgeDraft] = useState(null);
  const [knowledgeCollection, setKnowledgeCollection] = useState('all');
  const [knowledgeImproveGoal, setKnowledgeImproveGoal] = useState('');
  const [knowledgeView, setKnowledgeView] = useState('library');
  const [knowledgeSettingsOpen, setKnowledgeSettingsOpen] = useState(false);
  const [openingText, setOpeningText] = useState('');
  const [openingStyle, setOpeningStyle] = useState('');
  const [openingExtra, setOpeningExtra] = useState('');
  const [openingResult, setOpeningResult] = useState('');
  const [openingItems, setOpeningItems] = useState([]);
  const [configLoading, setConfigLoading] = useState(false);
  const [configEffective, setConfigEffective] = useState({});
  const [configDirty, setConfigDirty] = useState(false);
  const [webSubmitConfig, setWebSubmitConfig] = useState(null);
  const [webSubmitPlan, setWebSubmitPlan] = useState(null);
  const [webSubmitPreviewTaskIds, setWebSubmitPreviewTaskIds] = useState([]);
  const [webSubmitEnvironment, setWebSubmitEnvironment] = useState(null);
  const [webSubmitResults, setWebSubmitResults] = useState([]);
  const [webSubmitHistory, setWebSubmitHistory] = useState([]);
  const [webSubmitLoading, setWebSubmitLoading] = useState(false);
  const [submitView, setSubmitView] = useState('ready');
  const [uploadSession, setUploadSession] = useState(null);
  const [loginModalOpen, setLoginModalOpen] = useState(false);
  const [loginFormVisible, setLoginFormVisible] = useState(false);
  const [loginUser, setLoginUser] = useState('');
  const [loginPassword, setLoginPassword] = useState('');
  const [uploadProfileDraft, setUploadProfileDraft] = useState('');
  const [uploadProfileLoading, setUploadProfileLoading] = useState(false);
  const [versionPrompts, setVersionPrompts] = useState({ rewrite: '', processingRule: '', knowledgeUsage: '' });
  const [uploadProfiles, setUploadProfiles] = useState([]);
  const [selectedUploadProfile, setSelectedUploadProfile] = useState('');
  const [profileBindings, setProfileBindings] = useState({});
  const [taskDate, setTaskDate] = useState('');
  const [taskSearch, setTaskSearch] = useState('');
  const [batchAiCount, setBatchAiCount] = useState(1);
  const [ruleApplyScope, setRuleApplyScope] = useState('both');
  const [configSection, setConfigSection] = useState('workflow');
  const [selectedVersions, setSelectedVersions] = useState(() => savedProcessSettings.selectedVersions);
  const [batchFactorySourceVersion, setBatchFactorySourceVersion] = useState('sensitive');
  const [rewriteMethods] = useState(AUTO_REWRITE_METHODS);
  const [sensitiveAiProcessEnabled, setSensitiveAiProcessEnabled] = useState(() => savedProcessSettings.sensitiveAiProcessEnabled);
  const [textModel, setTextModel] = useState(() => savedProcessSettings.textModel);
  const [textModelOptions, setTextModelOptions] = useState([]);
  const [currentBatch, setCurrentBatch] = useState(null);
  const [historyBatches, setHistoryBatches] = useState([]);
  const [taskView, setTaskView] = useState('current');
  const [accountRole, setAccountRole] = useState('member');
  const [accountScope, setAccountScope] = useState(() => getSelectedNovelFetchAccount());
  const [accountOptions, setAccountOptions] = useState([]);
  const [documentHidden, setDocumentHidden] = useState(() => typeof document !== 'undefined' && document.hidden);
  const pollingRef = useRef({ realtimeBusy: false, queueBusy: false, tasksBusy: false, realtimeFailures: 0, taskFailures: 0 });
  const taskDataSignatureRef = useRef('');
  const batchDataSignatureRef = useRef('');
  const inputFormatOption = INPUT_FORMAT_OPTIONS.find(item => item.value === inputFormat) || INPUT_FORMAT_OPTIONS[0];
  const inputFormatLabel = inputFormatOption.columnPreset;
  const inputColumnOrder = inputFormatOption.columnOrder;
  const localInputPreview = useMemo(() => normalizeNovelFetchInput(bookIdsValue, { format: inputFormat }), [bookIdsValue, inputFormat]);

  useEffect(() => {
    writeNovelFetchProcessSettings({
      ...savedProcessSettings,
      ...processForm.getFieldsValue(['platform', 'maxTxt']),
      selectedVersions,
      rewriteMethods,
      sensitiveAiProcessEnabled,
      textModel
    });
  }, [processForm, savedProcessSettings, selectedVersions, rewriteMethods, sensitiveAiProcessEnabled, textModel]);

  const selectedTask = useMemo(() => selectedTaskKeys[0] || '', [selectedTaskKeys]);
  const visibleTasks = useMemo(() => {
    return filterNovelFetchTasks(tasks, {
      mode: taskDate ? 'date' : 'default',
      date: taskDate,
      query: taskSearch
    });
  }, [tasks, taskDate, taskSearch]);
  const currentBatchRows = useMemo(() => {
    const rows = currentBatch?.taskStates || currentBatch?.task_states;
    return Array.isArray(rows) ? rows : [];
  }, [currentBatch]);
  const currentBatchIds = useMemo(() => currentBatchBookIds(currentBatch || {}), [currentBatch]);
  const currentBatchStoppableIds = useMemo(() => stoppableCurrentBatchBookIds(currentBatch || {}), [currentBatch]);
  const progressRows = currentBatchRows.length ? currentBatchRows : visibleTasks;
  const liveLogRows = useMemo(() => {
    if (!currentBatchRows.length) return visibleTasks;
    const batchIds = new Set(currentBatchIds);
    const taskRows = tasks.filter(row => batchIds.has(String(row.bookId ?? row.book_id ?? row.id ?? '').trim()));
    return [...currentBatchRows, ...taskRows];
  }, [currentBatchRows, currentBatchIds, tasks, visibleTasks]);
  const processSummary = useMemo(
    () => summarizeWorkshopProgress(progressRows, webSubmitResults),
    [progressRows, webSubmitResults]
  );
  const liveLogLines = useMemo(() => novelFetchProgressLines(processJobs.length ? processJobs : processJob, liveLogRows), [processJobs, processJob, liveLogRows]);
  const queueSummary = useMemo(() => novelFetchQueueSummary(queueStatus), [queueStatus]);
  const progressPercent = value => processSummary.total
    ? Math.min(100, Math.round((Number(value) / processSummary.total) * 100))
    : 0;
  const webSubmitPreviewReady = submissionPreviewMatches(selectedTaskKeys, webSubmitPreviewTaskIds, webSubmitPlan);

  useEffect(() => {
    if (activeTab !== 'tasks' || taskView !== 'current') return;
    setSelectedTaskKeys(current => {
      const next = retainVisibleTaskIds(current, visibleTasks);
      return next.length === current.length ? current : next;
    });
  }, [activeTab, taskView, visibleTasks]);

  async function loadTasks({ silent = false } = {}) {
    // Background polling must keep the current table stable. Toggling the
    // Ant table spinner on every silent poll makes the whole page flash.
    if (!silent) setTasksLoading(true);
    try {
      const [taskData, jobData] = await Promise.all([
        listWorkshopTasks(silent ? { silent: true, suppressGlobalError: true } : {}),
        getWorkshopJobs(10, silent ? { silent: true, suppressGlobalError: true } : {})
      ]);
      const nextTasks = taskData.tasks || [];
      const nextJobs = Array.isArray(jobData.jobs) ? jobData.jobs : [];
      const taskSignature = JSON.stringify(nextTasks);
      const jobSignature = JSON.stringify(nextJobs);
      if (taskSignature !== taskDataSignatureRef.current) {
        taskDataSignatureRef.current = taskSignature;
        setTasks(nextTasks);
      }
      if (jobSignature !== pollingRef.current.jobsSignature) {
        pollingRef.current.jobsSignature = jobSignature;
        setProcessJobs(nextJobs);
      }
      if (jobData.latest && Object.keys(jobData.latest).length) {
        const latestSignature = JSON.stringify(jobData.latest);
        if (latestSignature !== pollingRef.current.latestJobSignature) {
          pollingRef.current.latestJobSignature = latestSignature;
          setProcessJob(jobData.latest);
        }
      }
      return true;
    } catch (error) {
      if (!silent) message.error(error.message || '读取任务失败');
      return false;
    } finally { if (!silent) setTasksLoading(false); }
  }

  async function loadQueueStatus() {
    try {
      const nextStatus = await getWorkshopRealtimeStatus();
      setQueueStatus(current => mergeNovelFetchQueueStatus(current, nextStatus));
      return true;
    } catch {
      // Keep the last confirmed enabled/paused state during a transient poll failure.
      return false;
    }
  }
  async function loadCurrentBatch() {
    try {
      const data = await getWorkshopCurrentBatch();
      const nextBatch = data.batch || null;
      const signature = JSON.stringify(nextBatch);
      if (signature !== batchDataSignatureRef.current) {
        batchDataSignatureRef.current = signature;
        setCurrentBatch(nextBatch);
      }
      return true;
    } catch { return false; }
  }
  async function loadRealtimeJob() {
    try {
      const jobData = await getWorkshopJobs(10, { silent: true, suppressGlobalError: true });
      const nextJobs = Array.isArray(jobData.jobs) ? jobData.jobs : [];
      const jobSignature = JSON.stringify(nextJobs);
      if (jobSignature !== pollingRef.current.jobsSignature) {
        pollingRef.current.jobsSignature = jobSignature;
        setProcessJobs(nextJobs);
      }
      if (jobData.latest && Object.keys(jobData.latest).length) {
        const latestSignature = JSON.stringify(jobData.latest);
        if (latestSignature !== pollingRef.current.latestJobSignature) {
          pollingRef.current.latestJobSignature = latestSignature;
          setProcessJob(jobData.latest);
        }
      }
      return true;
    } catch {
      return false;
    }
  }
  async function loadHistoryBatches() {
    try { const data = await getWorkshopBatches(); setHistoryBatches(data.batches || []); } catch (error) { message.error(error.message || '读取历史批次失败'); }
  }
  async function selectTask(bookId) {
    if (!bookId) return;
    try { setSelectedTaskData(await getWorkshopTask(bookId)); } catch (error) { message.error(error.message || '读取任务详情失败'); }
  }
  async function loadKnowledge() {
    setKnowledgeLoading(true);
    try {
      const [items, summary] = await Promise.all([getWorkshopKnowledge(knowledgeKind), getWorkshopKnowledgeSummary()]);
      const loaded = items?.data || items || { items: [] };
      setKnowledge(loaded);
      const loadedItems = Array.isArray(loaded.items) ? loaded.items : [];
      setKnowledgeSelectedId(current => loadedItems.some(item => String(item.id) === String(current)) ? current : String(loadedItems[0]?.id || ''));
      setKnowledgeSummary(summary?.summary || summary || {});
    } catch (error) { message.error(error.message || '读取知识库失败'); } finally { setKnowledgeLoading(false); }
  }
  async function loadConfig() {
    setConfigLoading(true);
    try {
      const data = await getWorkshopConfig();
      const config = withNovelFetchConfigDefaults(data.appConfig || data.config || {});
      const aiConfig = data.aiConfig?.ai || data.aiConfig || config.ai || {};
      const catalogText = value => {
        if (typeof value === 'string') return value;
        try { return JSON.stringify(value || [], null, 2); } catch { return '[]'; }
      };
      const rewrite = { ...(config.rewrite || {}) };
      rewrite.method_sequence_text = Array.isArray(rewrite.method_sequence)
        ? rewrite.method_sequence.join(',')
        : String(rewrite.method_sequence || '');
      const loadedValues = {
        ...config,
        workflow: { ...(config.workflow || {}) },
        fetch: { ...(config.fetch || {}) },
        rewrite,
        ai: { ...(config.ai || {}), ...aiConfig },
        sensitive_ai: { ...(config.sensitive_ai || {}) },
        storage: { ...(config.storage || {}) },
        knowledge: { ...(config.knowledge || {}) },
        platforms: catalogText(data.platforms || config.platforms),
        styles: catalogText(data.styles || config.styles)
      };
      configForm.setFieldsValue(loadedValues);
      setConfigEffective(loadedValues);
      setConfigDirty(false);
      setKnowledgeUsagePrompt(String(config.knowledge?.usage_prompt || config.knowledge_usage_prompt || ''));
      const configuredPlatforms = Array.isArray(data.platforms) ? data.platforms : Array.isArray(config.platforms) ? config.platforms : [];
      if (configuredPlatforms.length) setPlatformOptions(configuredPlatforms.map(item => { const id = Number(item.id ?? item.platform_id ?? item.value); const name = item.name || item.platform_name || item.label || String(id); return { value: id, label: `${name} (${id})` }; }));
    } catch (error) { message.error(error.message || '读取配置失败'); } finally { setConfigLoading(false); }
  }
  async function loadRuleConfig(kind = ruleKind) {
    try {
      const data = await getWorkshopConfig();
      const config = data.appConfig || data.config || {};
      // The Go API keeps both V78's legacy knowledge view and the canonical
      // library view. An empty legacy object must not hide a populated library
      // after switching tabs or migrating an older account.
      const knowledgeConfig = {};
      [data.knowledgeLibraries, config.knowledgeLibraries, data.knowledge, config.knowledge]
        .filter(candidate => candidate && typeof candidate === 'object' && !Array.isArray(candidate))
        .forEach(candidate => Object.entries(candidate).forEach(([key, value]) => {
          const existing = knowledgeConfig[key];
          const emptyObject = value && typeof value === 'object' && !Array.isArray(value) && Object.keys(value).length === 0;
          if (emptyObject && existing && typeof existing === 'object' && Object.keys(existing).length > 0) return;
          knowledgeConfig[key] = value;
        }));
      const rewrite = config.rewrite || data.rewrite || data.settings?.rewrite || {};
      const chapterRules = knowledgeConfig.chapter_rules || {};
      const symbolRules = knowledgeConfig.symbol_rules || {};
      // V78 stores chapter/symbol rules inside their library documents. Map
      // the editor's visible kind to that canonical object instead of showing
      // an empty placeholder after switching back to this tab.
      const selected = knowledgeConfig[kind]
        || (kind === 'chapter_exact_rules' ? { chapter_exact_rules: chapterRules.chapter_exact_rules || [] } : null)
        || (kind === 'chapter_inline_rules' ? { chapter_inline_rules: chapterRules.chapter_inline_rules || [] } : null)
        || (kind === 'pair_fill_rules' ? { pair_fill_rules: symbolRules.pair_fill_rules || [] } : null)
        || (kind === 'symbol_rules' ? symbolRules : null)
        || (kind === 'sensitive' ? (knowledgeConfig.sensitive_rules ? { groups: [{ name: '系统敏感词', enabled: true, apply_to_original: true, apply_to_ai: true, rules: knowledgeConfig.sensitive_rules }] } : {}) : {})
        || {};
      // 处理规则提示词与规则 JSON 都保存在统一配置里；只恢复 JSON 会让用户切换回来时误以为规则丢失。
      setRuleText(String(config.processing_rule_prompt || rewrite.processing_rule_prompt || data.processing_rule_prompt || data.settings?.processing_rule_prompt || ''));
      setRuleScope(String(config.processing_rule_scope || rewrite.processing_rule_scope || data.processing_rule_scope || 'both'));
      setRuleConfigText(JSON.stringify(selected, null, 2));
    } catch (error) { message.error(error.message || '读取处理规则配置失败'); }
  }
  async function loadWebSubmit() {
    try {
      const [configData, historyData] = await Promise.all([getWebSubmitConfig(), getWebSubmitHistory()]);
      setWebSubmitConfig(configData.settings || configData.config || {});
      setWebSubmitHistory(historyData.records || historyData.items || []);
    } catch (error) { message.error(error.message || '读取提交配置失败'); }
  }
  async function loadUploadSession() {
    try {
      const session = await getUploadSession();
      setUploadSession(session);
      return session;
    } catch {
      const session = { status: 'offline' };
      setUploadSession(session);
      return session;
    }
  }
  async function loadSchedules() {
    try { setSchedules((await listWorkshopSchedules()).schedules || []); } catch (error) { message.error(error.message || '读取定时任务失败'); }
  }
  async function loadGlobalLogs() {
    try { setGlobalLogs((await getWorkshopLogs()).logs || []); } catch (error) { message.error(error.message || '读取操作记录失败'); }
  }

  async function loadAccountScope() {
    try {
      const center = await getMemberCenter();
      const self = center?.member || {};
      let teamMembers = [];
      let globalAccounts = [];
      if (self.role === 'manager') teamMembers = (await getTeamMembers()).members || [];
      if (self.role === 'dev') globalAccounts = (await listGlobalAccounts()).accounts || [];
      const options = novelFetchAccountOptions(self, teamMembers, globalAccounts);
      const selected = options.some(item => item.value === getSelectedNovelFetchAccount()) ? getSelectedNovelFetchAccount() : self.username;
      setAccountRole(self.role || 'member');
      setAccountOptions(options);
      setAccountScope(selected || '');
      setSelectedNovelFetchAccount(selected || '');
    } catch {
      setAccountOptions([]);
    }
  }

  async function changeAccountScope(username) {
    setSelectedNovelFetchAccount(username);
    setAccountScope(username);
    setSelectedTaskKeys([]);
    setSelectedTaskData(null);
    setTaskDetail(null);
    setWebSubmitPlan(null);
    setWebSubmitResults([]);
    await Promise.all([loadTasks(), loadQueueStatus(), loadCurrentBatch(), loadHistoryBatches(), loadUploadSession(), loadSchedules(), loadGlobalLogs(), loadConfig(), loadKnowledge(), loadWebSubmit()]);
  }

  async function toggleAutomation() {
    setAutomationChanging(true);
    try {
      const wasEnabled = queueSummary.enabled;
      const status = wasEnabled ? await pauseWorkshopAutomation() : await startWorkshopAutomation();
      setQueueStatus(status || { counts: {} });
      if (status?.active_job_id) setProcessJob(current => current?.id ? current : { id: status.active_job_id, status: 'queued' });
      message.success(wasEnabled ? '全自动已暂停，当前正在执行的任务会安全完成' : '全自动已开启，待处理和待上传任务会自动进入队列');
      await Promise.all([loadTasks(), loadCurrentBatch()]);
    } catch (error) {
      message.error(error.message || '切换全自动失败');
    } finally {
      setAutomationChanging(false);
    }
  }

  useEffect(() => {
    loadAccountScope();
    loadTasks();
    loadQueueStatus();
    loadCurrentBatch();
    loadHistoryBatches();
    loadUploadSession();
    loadSchedules();
    loadGlobalLogs();
    getWorkshopPlatforms().then(data => {
      const items = data.platforms || data.items || data;
      if (Array.isArray(items) && items.length) setPlatformOptions(items.map(item => { const id = Number(item.id ?? item.platform_id ?? item.value); const name = item.name || item.platform_name || item.label || String(id); return { value: id, label: `${name} (${id})` }; }));
    }).catch(() => {});
    listAvailableTextModels('text').then(models => {
      const options = modelSelectOptions(models, 'text');
      if (options.length) setTextModelOptions(options);
    }).catch(() => {});
  }, []);

  useEffect(() => {
    const onVisibilityChange = () => setDocumentHidden(document.hidden);
    document.addEventListener('visibilitychange', onVisibilityChange);
    return () => document.removeEventListener('visibilitychange', onVisibilityChange);
  }, []);

  useEffect(() => {
    // 运行状态、队列和完整任务列表分开刷新，避免空闲页面每两秒重拉整张任务表。
    const plan = novelFetchPollingPlan({ activeTab, processing, documentHidden, ...pollingRef.current });
    if (!plan.realtimeMs) return undefined;
    let cancelled = false;
    const timers = [];
    const repeat = (busyKey, delayKey, run, failureKey) => {
      const tick = async () => {
        if (cancelled || document.hidden) return;
        if (!pollingRef.current[busyKey]) {
          pollingRef.current[busyKey] = true;
          const ok = await run();
          pollingRef.current[busyKey] = false;
          pollingRef.current[failureKey] = ok ? 0 : pollingRef.current[failureKey] + 1;
        }
        if (cancelled) return;
        const next = novelFetchPollingPlan({ activeTab, processing, documentHidden: document.hidden, ...pollingRef.current })[delayKey];
        if (next) timers.push(window.setTimeout(tick, next));
      };
      timers.push(window.setTimeout(tick, delayKey === 'realtimeMs' ? delayKey && plan.realtimeMs : plan[delayKey]));
    };
    repeat('realtimeBusy', 'realtimeMs', loadRealtimeJob, 'realtimeFailures');
    repeat('queueBusy', 'queueMs', loadQueueStatus, 'realtimeFailures');
    repeat('tasksBusy', 'tasksMs', async () => {
      const [tasksOK, batchOK] = await Promise.all([loadTasks({ silent: true }), loadCurrentBatch()]);
      return tasksOK && batchOK;
    }, 'taskFailures');
    return () => { cancelled = true; timers.forEach(timer => window.clearTimeout(timer)); };
  }, [activeTab, processing, documentHidden]);

  useEffect(() => {
    // 公网 V78 在进入处理台时会把当前空闲/执行状态同步到顶栏；否则本地顶栏会比线上少一段状态文案。
    dispatchGlobalStatus({
      text: processing ? '正在处理任务。' : '暂无正在处理的任务。',
      tone: processing ? 'working' : 'idle'
    });
  }, [processing]);

  useEffect(() => {
    if (activeTab === 'config') { loadConfig(); openVersionConfig(); }
    if (activeTab === 'knowledge' || activeTab === 'rules') { loadKnowledge(); loadConfig(); loadRuleConfig(); }
    if (activeTab === 'rules') setKnowledgeView('rules');
    if (activeTab === 'submit' || activeTab === 'logs') loadWebSubmit();
    if (activeTab === 'logs') setSubmitView('logs');
  }, [activeTab, knowledgeKind, ruleKind]);

  useEffect(() => {
    const item = (knowledge.items || []).find(entry => String(entry.id) === String(knowledgeSelectedId));
    setKnowledgeDraft(item ? { ...item } : null);
  }, [knowledge.items, knowledgeSelectedId]);

  function changeTab(tab) {
    setActiveTab(tab);
    const url = new URL(window.location.href);
    url.searchParams.set('tab', tab);
    window.history.replaceState(null, '', `${url.pathname}${url.search}`);
  }
  async function resolveInputPreview(values) {
    const inputText = String(values.bookIds || '').trim();
    const parsed = normalizeNovelFetchInput(inputText, { format: inputFormat });
    if (!parsed.validRows) {
      message.warning('请填写至少一条有效的书籍 ID');
      return null;
    }
    if (parsed.duplicateIds || parsed.emptyIdRows) {
      message.warning(`请先处理输入中的 ${parsed.duplicateIds} 条重复 ID 和 ${parsed.emptyIdRows} 条空 ID`);
      return null;
    }
    const payload = {
      input_text: inputText,
      max_txt: values.maxTxt,
      input_format: inputFormat,
      column_preset: inputFormatLabel,
      column_order: inputColumnOrder
    };
    const preview = await previewWorkshopProcess(payload);
    setInputPreview({ ...preview, rows: parsed.rows });
    return { inputText, parsed, metadata: { input_format: inputFormat, column_preset: inputFormatLabel, column_order: inputColumnOrder } };
  }
  async function handleProcess() {
    const values = await processForm.validateFields();
    const resolvedInput = await resolveInputPreview(values);
    if (!resolvedInput) return;
    const { inputText, parsed, metadata } = resolvedInput;
    const bookIds = parsed.rows.map(row => row.bookId).filter(Boolean);
    setProcessing(true);
    setProcessRows(bookIds.map(bookId => ({ bookId, status: 'loading' })));
    try {
      // The Go provider resolver reads the account's selected model. Persist the
      // visible choice before starting so a local draft cannot silently use a
      // different enabled model.
      if (textModel) await saveWorkshopConfig({ text_model_id: textModel });
      const started = await startWorkshopProcess({
        inputText,
        platform_id: String(values.platform),
        maxTxt: values.maxTxt,
        text_model_id: textModel,
        selected_versions: selectedVersions,
        rewrite_methods: rewriteMethods,
        sensitive_ai_enabled: sensitiveAiProcessEnabled,
        ...metadata
      });
      setProcessJob(started);
      setProcessJobs([started]);
      let data = started;
      const deadline = Date.now() + 30 * 60 * 1000;
      while (data?.id && !['done', 'failed'].includes(String(data.status || '').toLowerCase()) && Date.now() < deadline) {
        await new Promise(resolve => setTimeout(resolve, 1000));
        data = await getWorkshopJob(data.id);
        setProcessJob(data);
        setProcessJobs([data]);
        await Promise.all([loadTasks(), loadCurrentBatch()]);
      }
      if (String(data?.status || '').toLowerCase() === 'failed') throw new Error(data.error || '后台处理失败');
      if (Date.now() >= deadline) throw new Error('后台处理超时，请稍后在任务列表查看进度');
      const results = data.result?.tasks || data.tasks || [];
      setProcessRows(results.length
        ? results.map(row => ({ ...row, bookId: row.bookId || row.book_id, status: row.status === 'ok' || row.originalStatus === 'done' ? 'done' : (row.status || 'failed') }))
        : bookIds.map(bookId => ({ bookId, status: data.status === 'done' ? 'done' : data.status || 'failed', error: data.error || '' })));
      await loadTasks();
      await Promise.all([loadCurrentBatch(), loadHistoryBatches()]);
    } catch (error) {
      setProcessRows(bookIds.map(bookId => ({ bookId, status: 'failed', error: error.message })));
      message.error(error.message || '处理失败');
    } finally { setProcessing(false); }
  }
  async function handleScheduleProcess() {
    const values = await processForm.validateFields();
    const resolvedInput = editingScheduleId ? null : await resolveInputPreview(values);
    if (!editingScheduleId && !resolvedInput) return;
    if (!scheduleTime) return message.warning('请选择定时开始时间');
    const runAt = new Date(scheduleTime);
    if (Number.isNaN(runAt.getTime()) || runAt.getTime() <= Date.now()) return message.warning('开始时间必须晚于当前时间');
    try {
      setScheduleSaving(true);
      const isEditing = Boolean(editingScheduleId);
      if (isEditing) {
        await updateWorkshopSchedule(editingScheduleId, { run_at: scheduleTime });
      } else {
        await saveWorkshopSchedule({
          run_at: scheduleTime,
          input_text: String(values.bookIds || ''),
          platform_id: String(values.platform),
          max_txt: values.maxTxt,
          text_model_id: textModel,
          selected_versions: selectedVersions,
          rewrite_methods: rewriteMethods,
          sensitive_ai_enabled: sensitiveAiProcessEnabled,
          auto_submit: true,
          ...resolvedInput.metadata
        });
      }
      setScheduleModalOpen(false);
      setEditingScheduleId('');
      setScheduleTime('');
      await loadSchedules();
      message.success(isEditing ? '定时任务时间已调整' : '定时任务已保存');
    } catch (error) { message.error(error.message || (editingScheduleId ? '调整定时任务失败' : '保存定时任务失败')); } finally { setScheduleSaving(false); }
  }
  function openScheduleDialog(loadExisting = false) {
    setEditingScheduleId('');
    setScheduleTime(current => current || scheduleInputValue());
    setScheduleModalOpen(true);
    if (loadExisting) loadSchedules();
  }
  function editSchedule(row) {
    if (row.status && row.status !== 'scheduled') {
      message.info('只有待执行的定时任务可以调整时间');
      return;
    }
    const value = row.run_at || row.runAt;
    const date = new Date(value);
    if (!value || Number.isNaN(date.getTime())) {
      message.error('定时任务时间无效，无法调整');
      return;
    }
    setEditingScheduleId(String(row.id));
    setScheduleTime(scheduleInputValue(date));
    setScheduleModalOpen(true);
  }
  async function handleSensitiveRepair() {
    const ids = selectedTaskKeys.length ? selectedTaskKeys : currentBatchIds;
    if (!ids.length) return message.warning('当前没有可处理的任务');
    try { await reprocessWorkshopSensitive(ids, true); await loadTasks(); message.success('敏感词 AI 修复已加入处理'); } catch (error) { message.error(error.message || '敏感词 AI 修复失败'); }
  }
  async function handleGenerateAi() {
    if (!selectedTask) return message.warning('请先选择一个任务');
    try { await generateWorkshopAi(selectedTask, 1); await selectTask(selectedTask); await loadTasks(); message.success('已生成 AI 版本'); } catch (error) { message.error(error.message || '生成 AI 版本失败'); }
  }
  async function handleRetryTasks() {
    if (!selectedTaskKeys.length) return message.warning('请先选择需要重试的任务');
    try {
      const result = await retryWorkshopTasks(selectedTaskKeys);
      await Promise.all([loadTasks(), loadCurrentBatch(), loadQueueStatus(), loadRealtimeJob()]);
      message.success(`重试已处理：${novelFetchBatchResultSummary(result)}`);
    } catch (error) { message.error(error.message || '批量重试失败'); }
  }
  async function handleDeleteTasks() {
    Modal.confirm({ title: '删除选中任务', content: '删除后无法恢复，确认继续？', onOk: async () => { await deleteWorkshopTasks(selectedTaskKeys); setSelectedTaskKeys([]); await loadTasks(); } });
  }
  async function handleRulePreview() {
    if (!ruleText.trim()) return message.warning('请填写待处理文本');
    setRuleLoading(true);
    try { setRuleResult(await previewWorkshopRules(ruleText, ruleScope)); } catch (error) { message.error(error.message || '预览规则失败'); } finally { setRuleLoading(false); }
  }
  async function handleRuleSuggest() {
    const sample = ruleAiSample.trim() || ruleText.trim();
    if (!sample) return message.warning('请填写 AI 处理样本文案');
    setRuleLoading(true);
    try {
      const result = await suggestWorkshopRules(sample, ruleKind, ruleAiGoal.trim() || '请根据当前规则类型生成可保存的规则建议');
      setRuleResult(current => ({ ...current, suggestions: result }));
    } catch (error) { message.error(error.message || '生成规则建议失败'); } finally { setRuleLoading(false); }
  }
  async function saveRuleConfiguration() {
    if (!ruleText.trim()) return message.warning('请先填写规则内容');
    try {
      let ruleConfig = {};
      if (ruleConfigText.trim()) {
        ruleConfig = JSON.parse(ruleConfigText);
        if (!ruleConfig || Array.isArray(ruleConfig) || typeof ruleConfig !== 'object') throw new Error('规则配置必须是 JSON 对象');
      }
      await saveWorkshopConfig({ processing_rule_prompt: ruleText, processing_rule_scope: ruleScope, knowledge: { [ruleKind]: ruleConfig } });
      message.success('处理规则已保存');
    } catch (error) { message.error(error.message || '保存处理规则失败'); }
  }
  async function saveKnowledge() {
    const values = await knowledgeForm.validateFields();
    try {
      let extra = {};
      if (knowledgeDetails.trim()) {
        extra = JSON.parse(knowledgeDetails);
        if (!extra || Array.isArray(extra) || typeof extra !== 'object') throw new Error('完整字段必须是 JSON 对象');
      }
      delete values.details;
      await saveWorkshopKnowledge(knowledgeKind, { ...extra, ...values });
      setKnowledgeModal(false);
      await loadKnowledge();
      message.success('知识条目已保存');
    } catch (error) { message.error(error.message || '保存知识条目失败'); }
  }
  async function addKnowledgeLibraryItem() {
    const collection = ['profiles', 'temporary_instructions', 'prompts', 'references'].includes(knowledgeCollection) ? knowledgeCollection : '';
    const defaults = knowledgeKind === 'high_imitation'
      ? collection === 'references'
        ? { id: `mimic_reference_${Date.now()}`, name: '新的高仿参考', collection: 'references', gender: '', style: '', tags: [], reference_text: '', analysis: '', enabled: true }
        : { id: `mimic_prompt_${Date.now()}`, name: '新的高仿提示词', collection: 'prompts', gender: '', style: '', tags: [], system_prompt: '', user_prompt_template: '', analysis: '', enabled: true }
      : knowledgeKind === 'opening_phrases'
        ? { id: `opening_${Date.now()}`, name: '新的开头词', style: '', content: '', analysis: '', slot_hint: '', tags: [], enabled: true, linked_template_ids: [] }
        : collection === 'temporary_instructions'
          ? { id: `instruction_${Date.now()}`, name: '新的临时指令', collection: 'temporary_instructions', content: '', enabled: true }
          : { id: `rewrite_${Date.now()}`, name: '新的改文模板', collection: 'profiles', system_prompt: '', user_prompt_template: '', required_variables: [], response_mode: 'changed_lines_json', enabled: true };
    try {
      const result = await saveWorkshopKnowledge(knowledgeKind, defaults);
      await loadKnowledge();
      setKnowledgeSelectedId(String(result?.item?.id || defaults.id));
      message.success('已新增条目');
    } catch (error) { message.error(error.message || '新增条目失败'); }
  }
  async function duplicateKnowledgeLibraryItem() {
    if (!knowledgeDraft) return;
    try {
      const copy = { ...knowledgeDraft, id: `${knowledgeDraft.id || 'copy'}_${Date.now()}`, name: `${knowledgeDraft.name || knowledgeDraft.title || '复制条目'} 副本` };
      const result = await saveWorkshopKnowledge(knowledgeKind, copy);
      await loadKnowledge();
      setKnowledgeSelectedId(String(result?.item?.id || copy.id));
      message.success('已复制条目');
    } catch (error) { message.error(error.message || '复制条目失败'); }
  }
  async function deleteKnowledgeLibraryItem() {
    if (!knowledgeDraft?.id) return;
    Modal.confirm({ title: '确认删除当前知识库条目？', onOk: async () => {
      try { await deleteWorkshopKnowledge(knowledgeKind, knowledgeDraft.id); await loadKnowledge(); message.success('已删除条目'); } catch (error) { message.error(error.message || '删除条目失败'); }
    } });
  }
  async function saveKnowledgeLibraryItem() {
    if (!knowledgeDraft) return;
    try { await saveWorkshopKnowledge(knowledgeKind, knowledgeDraft); await loadKnowledge(); message.success('条目已保存'); } catch (error) { message.error(error.message || '保存条目失败'); }
  }
  async function optimizeKnowledgeLibraryItem() {
    if (!knowledgeDraft?.id) return;
    try { await optimizeWorkshopKnowledge(knowledgeKind, knowledgeDraft.id); await loadKnowledge(); message.success('AI 优化已完成'); } catch (error) { message.error(error.message || 'AI 优化失败'); }
  }
  async function saveKnowledgeUsagePrompt() {
    try {
      await saveWorkshopConfig({ knowledge: { usage_prompt: knowledgeUsagePrompt } });
      message.success('知识库使用提示词已保存');
    } catch (error) { message.error(error.message || '保存知识库使用提示词失败'); }
  }
  async function handleOpeningAnalyze() {
    if (!openingText.trim()) return message.warning('请填写开头文本');
    try {
      const result = await analyzeWorkshopOpening(openingText, { style: openingStyle, extra: openingExtra });
      const items = result.items || [];
      setOpeningItems(items);
      setOpeningResult(JSON.stringify(items, null, 2));
      message.success(`已拆解 ${items.length} 条开头词`);
    } catch (error) { message.error(error.message || '分析失败'); }
  }
  async function handleOpeningSaveAnalyzed() {
    if (!openingItems.length) return message.warning('请先执行 AI 拆解');
    try {
      for (const item of openingItems) await saveWorkshopOpening(item);
      await loadKnowledge();
      message.success(`已将 ${openingItems.length} 条开头词加入知识库`);
    } catch (error) { message.error(error.message || '保存分析结果失败'); }
  }
  async function handleWebSubmit(action, success) {
    setWebSubmitLoading(true);
    try { const result = await action(); if (success) message.success(success); return result; } catch (error) { message.error(error.message || '网站提交操作失败'); return null; } finally { setWebSubmitLoading(false); }
  }
  async function syncWebSubmitOrganizations() {
    const result = await handleWebSubmit(getWebSubmitOrganizations, '已同步组织归属');
    const catalog = Array.isArray(result?.organizations) ? result.organizations : [];
    if (!catalog.length) return;
    setWebSubmitConfig(current => {
      const selected = String(current?.organization || '').trim();
      const stillAvailable = catalog.some(item => String(item.id || item.value || '').trim() === selected);
      return {
        ...(current || {}),
        organization_catalog: catalog,
        organization: stillAvailable ? selected : String(catalog[0].id || catalog[0].value || '')
      };
    });
  }
  async function verifyWebSubmitVisible() {
    const result = await handleWebSubmit(testWebSubmitVisible, '视频管理系统登录可用');
    if (result) setWebSubmitEnvironment(current => ({ ...(current || {}), visible: result }));
  }
  async function handleWebSubmitPreview() {
    if (!selectedTaskKeys.length) return message.warning('请先选择需要提交的任务');
    const result = await handleWebSubmit(() => previewWebSubmit({ taskIds: selectedTaskKeys }), '提交预览已生成');
    if (!result) return;
    setWebSubmitPlan(result.plan);
    setWebSubmitPreviewTaskIds([...selectedTaskKeys]);
  }
  function confirmWebSubmit() {
    if (!webSubmitPreviewReady) return message.warning('请先为当前选中任务生成提交预览');
    Modal.confirm({
      title: '确认开始自动提交',
      content: `将按当前预览提交 ${selectedTaskKeys.length} 个任务，确认继续？`,
      onOk: async () => {
        const result = await handleWebSubmit(() => startWebSubmit({ taskIds: selectedTaskKeys }), '自动提交完成');
        if (!result) return;
        setWebSubmitResults(result.results || []);
        setWebSubmitHistory(current => [...(result.results || []), ...current]);
        setSubmitView('result');
      }
    });
  }
  async function retryWebSubmitVersion(row) {
    const payload = singleVersionRetryPayload(row);
    if (!payload) return message.error('失败记录缺少任务或版本信息');
    const bookId = payload.taskIds[0];
    const result = await handleWebSubmit(() => startWebSubmit(payload), `${row.version} 已重新提交`);
    if (result?.results) setWebSubmitResults(current => current.map(item => (item.bookId || item.book_id) === bookId && item.version === row.version ? (result.results[0] || item) : item));
  }
  async function handleUploadLogin() {
    if (!loginUser.trim() || !loginPassword) return message.warning('请输入视频管理系统账号和密码');
    setWebSubmitLoading(true);
    try {
      await uploadLogin({ username: loginUser.trim(), password: loginPassword });
      setLoginPassword('');
      const session = await loadUploadSession();
      setLoginFormVisible(false);
      if (session?.status === 'ready') await syncUploadProfiles(false);
      message.success('视频管理系统已登录，可以选择默认网站配置档');
    } catch (error) { message.error(error.message || '视频管理系统登录失败'); } finally { setWebSubmitLoading(false); }
  }
  async function openUploadManager() {
    setLoginModalOpen(true);
    setUploadProfileLoading(true);
    try {
      const [session, configData] = await Promise.all([loadUploadSession(), getWebSubmitConfig()]);
      const settings = configData.settings || configData.config || {};
      const profiles = Array.isArray(settings.upload_profiles) ? settings.upload_profiles : [];
      setWebSubmitConfig(settings);
      setUploadProfiles(profiles);
      setSelectedUploadProfile(String(settings.selected_profile || ''));
      setUploadProfileDraft(String(settings.selected_profile || ''));
      setLoginFormVisible(session?.status !== 'ready');
    } catch (error) {
      message.error(error.message || '读取视频管理系统配置失败');
    } finally {
      setUploadProfileLoading(false);
    }
  }
  async function syncUploadProfiles(showSuccess = true) {
    setUploadProfileLoading(true);
    try {
      const result = await syncWebSubmitConfigs();
      const settings = result.settings || result.config || webSubmitConfig || {};
      const profiles = Array.isArray(result.groups) ? result.groups : (Array.isArray(settings.upload_profiles) ? settings.upload_profiles : []);
      setWebSubmitConfig(settings);
      setUploadProfiles(profiles);
      const savedProfile = String(settings.selected_profile || uploadProfileDraft || '');
      setSelectedUploadProfile(savedProfile);
      setUploadProfileDraft(savedProfile);
      if (showSuccess) message.success(`已同步 ${profiles.length} 个网站配置档`);
    } catch (error) {
      message.error(error.message || '同步网站配置档失败');
    } finally {
      setUploadProfileLoading(false);
    }
  }
  async function syncUploadStylesFromManager() {
    setUploadProfileLoading(true);
    try {
      await syncWebSubmitStyles();
      message.success('风格目录已同步');
    } catch (error) {
      message.error(error.message || '同步风格目录失败');
    } finally {
      setUploadProfileLoading(false);
    }
  }
  async function saveUploadProfileAndClose() {
    if (!uploadProfileDraft) return message.warning('请选择默认网站配置档');
    setUploadProfileLoading(true);
    try {
      const latest = await getWebSubmitConfig();
      const settings = latest.settings || latest.config || {};
      const nextSettings = mergeDefaultUploadProfile(settings, uploadProfileDraft);
      await saveWebSubmitConfig(nextSettings);
      setWebSubmitConfig(nextSettings);
      setSelectedUploadProfile(uploadProfileDraft);
      setLoginModalOpen(false);
      message.success('默认网站配置档已保存');
    } catch (error) {
      message.error(error.message || '保存默认网站配置档失败');
    } finally {
      setUploadProfileLoading(false);
    }
  }
  function openAdvancedVersionProfiles() {
    setLoginModalOpen(false);
    setConfigSection('versions');
    changeTab('config');
  }
  async function updateSelectedAiCount() {
    if (!selectedTaskKeys.length) return message.warning('请先选择任务');
    try { await updateWorkshopAiCount(selectedTaskKeys, batchAiCount); await loadTasks(); message.success(`已调整为 ${batchAiCount} 个 AI 版本`); } catch (error) { message.error(error.message || '调整 AI 数量失败'); }
  }
  async function retryFailedTasks() {
    const source = currentBatchRows.length ? currentBatchRows : visibleTasks;
    const failed = source.filter(task => String(task.status || task.aiStatus || task.ai_status || '').includes('fail')).map(task => task.bookId || task.book_id).filter(Boolean);
    if (!failed.length) return message.info('当前没有失败任务');
    try { await retryWorkshopTasks(failed); await loadTasks(); message.success('失败任务已加入重试'); } catch (error) { message.error(error.message || '重试失败任务失败'); }
  }
  async function retryAllFailedTasks() {
    const failed = tasks
      .filter(task => novelFetchTaskStatus(task).status === 'failed')
      .map(task => task.bookId || task.book_id)
      .filter(Boolean);
    if (!failed.length) return message.info('全部任务中没有失败项');
    try { await retryWorkshopTasks(failed); await loadTasks(); message.success(`已将 ${failed.length} 个失败任务加入重试`); } catch (error) { message.error(error.message || '重试全部失败任务失败'); }
  }
  function selectAllTasks() { setSelectedTaskKeys(visibleTasks.map(task => task.bookId || task.book_id).filter(Boolean)); }
  function shiftTaskDate(days) {
    const base = taskDate ? new Date(`${taskDate}T00:00:00+08:00`) : new Date();
    base.setUTCDate(base.getUTCDate() + days);
    setTaskDate(beijingDateKey(base));
  }
  async function applySelectedRules() {
    if (!selectedTaskKeys.length) return message.warning('请先选择任务');
    try { await applyWorkshopRules(selectedTaskKeys, ruleApplyScope, sensitiveAiProcessEnabled); await loadTasks(); message.success('已完成规则处理'); } catch (error) { message.error(error.message || '规则处理失败'); }
  }
  async function applyAllRules() {
    const ids = currentBatchIds;
    if (!ids.length) return message.warning('当前没有可处理任务');
    try { await applyWorkshopRules(ids, ruleApplyScope, sensitiveAiProcessEnabled); await loadTasks(); message.success('已完成全部任务规则处理'); } catch (error) { message.error(error.message || '规则处理失败'); }
  }
  async function applyRulesToAllTasks() {
    const ids = tasks.map(task => task.bookId || task.book_id).filter(Boolean);
    if (!ids.length) return message.warning('当前没有可处理任务');
    try { await applyWorkshopRules(ids, ruleApplyScope, sensitiveAiProcessEnabled); await loadTasks(); message.success(`已处理全部 ${ids.length} 个任务`); } catch (error) { message.error(error.message || '规则处理全部任务失败'); }
  }
  async function submitVisibleTasks(all = false) {
    const ids = currentBatchIds.length
      ? currentBatchIds
      : visibleTasks.map(task => task.bookId || task.book_id).filter(Boolean);
    if (!ids.length) return message.info('当前没有可提交任务');
    const result = await handleWebSubmit(() => startWebSubmit({ taskIds: ids }), all ? '已提交全部任务' : '已提交当前任务');
    if (result) setWebSubmitResults(result.results || []);
  }
  async function submitAllTasks() {
    const ids = tasks.map(task => task.bookId || task.book_id).filter(Boolean);
    if (!ids.length) return message.info('当前没有可提交任务');
    const result = await handleWebSubmit(() => startWebSubmit({ taskIds: ids }), `已提交全部 ${ids.length} 个任务`);
    if (result) setWebSubmitResults(result.results || []);
  }
  async function deleteTasksByFilter(filter, emptyMessage) {
    const ids = tasks.filter(filter).map(task => task.bookId || task.book_id).filter(Boolean);
    if (!ids.length) return message.info(emptyMessage);
    Modal.confirm({ title: '确认删除任务', content: '删除后无法恢复，确认继续？', onOk: async () => { await deleteWorkshopTasks(ids); setSelectedTaskKeys(current => current.filter(id => !ids.includes(id))); await loadTasks(); message.success('任务已删除'); } });
  }
  async function transferToBatchFactory() {
    const transferIds = selectedTaskKeys.length ? selectedTaskKeys : currentBatchIds;
    if (!transferIds.length) {
      message.warning('请先选择要进入批量工厂的任务');
      return;
    }
    setTasksLoading(true);
    try {
      // 批量工厂必须接收已经确认的正文版本，而不是只带 bookId 跳转；
      // 否则新批次会显示“已登记”却没有可生产的 sourceText。
      const rowSource = [...tasks, ...currentBatchRows];
      const selectedRows = transferIds.map(bookId => rowSource.find(task => (task.bookId || task.book_id) === bookId) || { bookId });
      const details = await Promise.all(selectedRows.map(row => getWorkshopTask(row.bookId || row.book_id)));
      const books = details.map((detail, index) => {
        const row = selectedRows[index] || {};
        const meta = detail?.meta || {};
        const requestedVersion = batchFactorySourceVersion || 'sensitive';
        const sourceText = batchFactorySourceText(detail, requestedVersion);
        if (!sourceText) throw new Error(`《${row.bookName || row.book_name || row.bookId || row.book_id}》没有可用正文`);
        const bookId = String(detail?.bookId || row.bookId || row.book_id || '').trim();
        const title = String(meta.bookName || detail?.bookName || row.bookName || row.book_name || bookId).trim();
        return {
          id: bookId,
          bookId,
          sourceTaskId: bookId,
          title,
          platform: String(meta.platformId || row.platformId || row.platform_id || ''),
          sourceText,
          txtText: sourceText,
          sourceMetadata: {
            sourceMode: 'novel_fetch',
            sourceContentVersion: requestedVersion,
            style: row.style || detail?.style || meta.style || '',
            gender: row.gender || detail?.gender || meta.gender || '',
            bookName: title,
            originalRawChars: detail?.originalRawChars || 0,
            processedChars: detail?.processedChars || sourceText.length
          }
        };
      });
      const result = await createNovelFetchIntake({
        books,
        metadata: { sourceMode: 'novel_fetch', sourceContentVersion: batchFactorySourceVersion, selectedTaskIds: transferIds }
      });
      const intake = result?.intake || result;
      if (!intake?.id) throw new Error('批量工厂接收失败：没有返回 intake');
      const redirectTo = result?.redirectTo || `/batch-factory?intake=${encodeURIComponent(intake.id)}`;
      window.location.assign(redirectTo);
    } catch (error) {
      message.error(error.message || '进入批量工厂失败');
    } finally {
      setTasksLoading(false);
    }
  }
  async function saveVersionConfig() {
    try {
      await saveWorkshopConfig({ rewrite: { prompt: versionPrompts.rewrite, processing_rule_prompt: versionPrompts.processingRule, ai_slot_methods: rewriteMethods }, knowledge: { usage_prompt: versionPrompts.knowledgeUsage }, selected_versions: selectedVersions, sensitive_ai_enabled: sensitiveAiProcessEnabled });
      const versionProfiles = withNovelFetchVersionProfiles(profileBindings);
      await saveWebSubmitConfig({ selected_profile: selectedUploadProfile, profile_bindings: versionProfiles, version_profiles: versionProfiles, submit_versions: selectedVersions });
      message.success('版本配置已保存');
    } catch (error) { message.error(error.message || '保存版本配置失败'); }
  }
  async function saveConfig() {
    try {
      const values = configForm.getFieldsValue(true);
      const parseCatalog = value => {
        if (Array.isArray(value)) return value;
        if (!String(value || '').trim()) return [];
        const parsed = JSON.parse(value);
        if (!Array.isArray(parsed)) throw new Error('平台/风格目录必须是 JSON 数组');
        return parsed;
      };
      const rewrite = { ...(values.rewrite || {}) };
      rewrite.method_sequence = String(rewrite.method_sequence_text || '')
        .split(',').map(item => item.trim()).filter(Boolean);
      delete rewrite.method_sequence_text;
      await saveWorkshopConfig({ ...values, rewrite, platforms: parseCatalog(values.platforms), styles: parseCatalog(values.styles) });
      setConfigEffective(values);
      setConfigDirty(false);
      message.success('配置已保存');
    } catch (error) { message.error(error.message || '保存配置失败'); }
  }
  async function openVersionConfig() {
    try {
      const data = await getWorkshopConfig();
      const rewrite = data.appConfig?.rewrite || data.config?.rewrite || {};
      const knowledgeConfig = data.appConfig?.knowledge || data.config?.knowledge || {};
      setVersionPrompts({ rewrite: rewrite.prompt || '', processingRule: rewrite.processing_rule_prompt || '', knowledgeUsage: knowledgeConfig.usage_prompt || '' });
    } catch { /* keep an empty draft when the optional config endpoint is unavailable */ }
    try {
      const submit = await getWebSubmitConfig();
      const config = submit.settings || submit.config || {};
      setUploadProfiles(Array.isArray(config.upload_profiles) ? config.upload_profiles : []);
      setSelectedUploadProfile(String(config.selected_profile || ''));
      setProfileBindings(withNovelFetchVersionProfiles(config.version_profiles || config.profile_bindings));
    } catch { /* profile configuration is optional until the upload backend is paired */ }
  }

  const knowledgeTabCount = key => ({
    high_imitation: knowledgeSummary.high_imitation || knowledgeSummary.high_imitation_prompts || 0,
    opening_phrases: knowledgeSummary.opening_phrases || 0,
    rewrite_templates: knowledgeSummary.rewrite_templates || knowledgeSummary.rewrite_profiles || 0,
    layout_rules: knowledgeSummary.sensitive_rules || 0,
    symbol_rules: knowledgeSummary.symbol_rules || 0,
    chapter_rules: (knowledgeSummary.chapter_exact_rules || 0) + (knowledgeSummary.chapter_inline_rules || 0)
  }[key] || 0);

  const knowledgeCollectionOptionsForKind = knowledgeCollectionOptions(knowledgeKind, knowledge);
  const visibleKnowledgeItems = filterKnowledgeCollection(knowledge.items, knowledgeCollection);
  const openKnowledgeCollection = (kind, collection = 'all') => {
    setKnowledgeKind(kind);
    setKnowledgeCollection(collection);
    setKnowledgeSelectedId('');
    setKnowledgeSearch('');
  };
  const knowledgeLibraryMetrics = [
    { key: 'opening_phrases', label: '爆款开头', value: knowledgeSummary.opening_phrases, description: '开头词库', onClick: () => openKnowledgeCollection('opening_phrases') },
    { key: 'opening_styles', label: '适用风格', value: knowledgeSummary.opening_styles, description: '开头词风格', onClick: () => openKnowledgeCollection('opening_phrases', 'styles') },
    { key: 'rewrite_profiles', label: '改文模板', value: knowledgeSummary.rewrite_profiles, description: '可复用模板', onClick: () => openKnowledgeCollection('rewrite_templates', 'profiles') },
    { key: 'temporary_instructions', label: '临时指令', value: knowledgeSummary.temporary_instructions, description: '单次补充规则', onClick: () => openKnowledgeCollection('rewrite_templates', 'temporary_instructions') },
    { key: 'high_imitation_prompts', label: '高仿提示词', value: knowledgeSummary.high_imitation_prompts, description: '高仿方法', onClick: () => openKnowledgeCollection('high_imitation', 'prompts') },
    { key: 'high_imitation_references', label: '高仿参考', value: knowledgeSummary.high_imitation_references, description: '高仿样本', onClick: () => openKnowledgeCollection('high_imitation', 'references') }
  ];
  const knowledgeRuleMetrics = [
    { key: 'sensitive_rules', label: '敏感词', value: knowledgeSummary.sensitive_rules, description: '敏感词替换', onClick: () => openKnowledgeCollection('layout_rules') },
    { key: 'symbol_rules', label: '字位符号', value: knowledgeSummary.symbol_rules, description: '符号替换', onClick: () => openKnowledgeCollection('symbol_rules') },
    { key: 'pair_fill_rules', label: '成对补符号', value: knowledgeSummary.pair_fill_rules, description: '补齐规则', onClick: () => openKnowledgeCollection('symbol_rules') },
    { key: 'chapter_exact_rules', label: '章节整行', value: knowledgeSummary.chapter_exact_rules, description: '整行清洗', onClick: () => openKnowledgeCollection('chapter_rules') },
    { key: 'chapter_inline_rules', label: '章节开头', value: knowledgeSummary.chapter_inline_rules, description: '行内清洗', onClick: () => openKnowledgeCollection('chapter_rules') }
  ];
  const knowledgePanel = <Space direction="vertical" size={12} style={{ width: '100%' }}>
    <section className="novel-fetch-knowledge-catalog">
      <div className="novel-fetch-knowledge-catalog-heading"><div><Typography.Title level={5}>完整目录</Typography.Title><Typography.Text type="secondary">每类资料都保留独立入口，点击可直接查看和编辑。</Typography.Text></div><Typography.Text type="secondary">当前：{KNOWLEDGE_COLLECTION_LABELS[knowledgeCollection] || knowledgeCollection}</Typography.Text></div>
      <div className="novel-fetch-knowledge-catalog-groups">
        <div className="novel-fetch-knowledge-catalog-group"><span>资料库</span><NovelFetchMetricStrip items={knowledgeLibraryMetrics.map(item => ({ ...item, active: item.key === knowledgeCollection || (knowledgeKind === 'opening_phrases' && item.key === 'opening_phrases' && knowledgeCollection === 'all') }))} /></div>
        <div className="novel-fetch-knowledge-catalog-group"><span>处理规则</span><NovelFetchMetricStrip items={knowledgeRuleMetrics} /></div>
      </div>
    </section>
    <KnowledgeLibraryManager kind={knowledgeKind} onKindChange={value => { setKnowledgeKind(value); setKnowledgeCollection('all'); setKnowledgeSelectedId(''); }} collection={knowledgeCollection} onCollectionChange={value => { setKnowledgeCollection(value); setKnowledgeSelectedId(''); }} collectionOptions={knowledgeCollectionOptionsForKind} items={visibleKnowledgeItems} loading={knowledgeLoading} search={knowledgeSearch} onSearchChange={setKnowledgeSearch} selectedId={knowledgeSelectedId} onSelect={setKnowledgeSelectedId} draft={knowledgeDraft} onDraftChange={setKnowledgeDraft} onAdd={addKnowledgeLibraryItem} onDuplicate={duplicateKnowledgeLibraryItem} onDelete={deleteKnowledgeLibraryItem} onNormalize={() => normalizeWorkshopOpening().then(loadKnowledge)} onSave={saveKnowledgeLibraryItem} onOptimize={optimizeKnowledgeLibraryItem} improveGoal={knowledgeImproveGoal} onImproveGoalChange={setKnowledgeImproveGoal} />
    <Drawer title="知识库设置" width={520} open={knowledgeSettingsOpen} onClose={() => setKnowledgeSettingsOpen(false)} extra={<Button type="primary" onClick={saveKnowledgeUsagePrompt}>保存设置</Button>}><Typography.Title level={5}>知识库使用提示词</Typography.Title><Typography.Paragraph type="secondary">告诉 AI 如何参考资料，但不得照搬示例内容。这里只需要在整体使用方式变化时调整。</Typography.Paragraph><Input.TextArea rows={10} value={knowledgeUsagePrompt} onChange={event => setKnowledgeUsagePrompt(event.target.value)} placeholder="请输入知识库使用要求" /></Drawer>
  </Space>;
  const openingPanel = <div className="novel-fetch-panel novel-fetch-opening-analyzer"><div className="novel-fetch-toolbar"><div><Typography.Title level={5}>爆款开头拆解入库</Typography.Title><Typography.Text type="secondary">拆解结构、情绪和钩子后再确认入库。</Typography.Text></div><span className="novel-fetch-toolbar-spacer" /><Button onClick={() => normalizeWorkshopOpening().then(loadKnowledge)}>规范化现有开头词</Button></div><div className="novel-fetch-opening-fields"><Input value={openingStyle} onChange={event => setOpeningStyle(event.target.value)} placeholder="适用风格类型（可选）" /><Input value={openingExtra} onChange={event => setOpeningExtra(event.target.value)} placeholder="补充要求（可选）" /></div><div className="novel-fetch-opening-compare"><label>爆款开头原文<Input.TextArea rows={9} value={openingText} onChange={event => setOpeningText(event.target.value)} /></label><label>拆解结果（可编辑）<Input.TextArea rows={9} value={openingResult} onChange={event => setOpeningResult(event.target.value)} className="mono" /></label></div><Space wrap><Button type="primary" onClick={handleOpeningAnalyze}>AI 拆解开头词</Button><Button onClick={handleOpeningSaveAnalyzed}>确认加入开头词库</Button><Button onClick={() => { setOpeningText(''); setOpeningStyle(''); setOpeningExtra(''); setOpeningResult(''); setOpeningItems([]); }}>清空</Button></Space></div>;
  const versionConfigPanel = <div className="novel-fetch-panel novel-fetch-version-config-panel">
    <div className="novel-fetch-public-heading"><div><Typography.Title level={4}>版本对应配置档</Typography.Title><Typography.Text type="secondary">统一管理视频管理系统配置档、版本提示词和提交版本。</Typography.Text></div></div>
    <Space wrap><Button onClick={() => handleWebSubmit(syncWebSubmitConfigs, '已同步批量后台配置')}>同步批量后台配置</Button><Button onClick={syncWebSubmitOrganizations}>同步组织归属</Button><Button onClick={() => handleWebSubmit(syncWebSubmitStyles, '已同步批量风格类型')}>同步批量风格类型</Button></Space>
    <div className="novel-fetch-version-profile-grid"><label>默认上传配置档<Select value={selectedUploadProfile || undefined} placeholder="选择默认配置档" options={uploadProfiles.map(profile => ({ value: String(profile.id || profile.name), label: profile.name || profile.id }))} onChange={setSelectedUploadProfile} /></label>{AI_VERSIONS.map(version => <label key={version}>{AI_VERSION_LABELS[version]} 配置档<Select allowClear value={profileBindings[version] || undefined} placeholder="跟随默认配置档" options={uploadProfiles.map(profile => ({ value: String(profile.id || profile.name), label: profile.name || profile.id }))} onChange={value => setProfileBindings(current => ({ ...withNovelFetchVersionProfiles(current), [version]: value || '' }))} /></label>)}</div><Typography.Text type="secondary">原文和 AI1-AI5 均可单独指定；未绑定版本统一跟随默认配置档。</Typography.Text><div className="novel-fetch-version-prompts"><label>改文提示词<Input.TextArea rows={3} value={versionPrompts.rewrite} onChange={event => setVersionPrompts(current => ({ ...current, rewrite: event.target.value }))} placeholder="本次处理的改文提示词" /></label><label>处理规则提示词<Input.TextArea rows={3} value={versionPrompts.processingRule} onChange={event => setVersionPrompts(current => ({ ...current, processingRule: event.target.value }))} placeholder="整本书共同遵守的规则" /></label><label>知识库使用提示词<Input.TextArea rows={3} value={versionPrompts.knowledgeUsage} onChange={event => setVersionPrompts(current => ({ ...current, knowledgeUsage: event.target.value }))} placeholder="知识库的使用要求" /></label></div><Button type="primary" onClick={saveVersionConfig}>保存版本配置</Button>
  </div>;
  const submissionStatus = [
    { key: 'environment', label: '环境', value: webSubmitEnvironment ? '已检测' : '待检测' },
    { key: 'config', label: '配置', value: webSubmitConfig ? '已加载' : '待加载' },
    { key: 'preview', label: '预览', value: webSubmitPreviewReady ? '当前选择已预览' : '待预览' },
    { key: 'execute', label: '执行', value: webSubmitResults.length ? `${webSubmitResults.length} 条结果` : '未执行' }
  ];
  const submitReadyPanel = <Space direction="vertical" size={12} style={{ width: '100%' }}>
    <Alert type="info" showIcon message="仅在确认“开始自动提交”后执行外部提交。小说获取和批量工厂共用同一份视频管理系统登录会话。" />
    <div className="novel-fetch-panel novel-fetch-submit-actions"><Typography.Title level={5}>提交前检查</Typography.Title><Space wrap><Button loading={webSubmitLoading} onClick={async () => { const result = await handleWebSubmit(checkWebSubmitEnvironment, '环境检测完成'); if (result) setWebSubmitEnvironment(result); }}>环境检测</Button><Button loading={webSubmitLoading} onClick={verifyWebSubmitVisible}>验证登录可用性</Button><Button loading={webSubmitLoading} onClick={() => handleWebSubmit(syncWebSubmitConfigs, '已同步批量后台配置')}>同步批量后台配置</Button><Button loading={webSubmitLoading} onClick={syncWebSubmitOrganizations}>同步组织归属</Button><Button loading={webSubmitLoading} onClick={() => handleWebSubmit(syncWebSubmitStyles, '同步网站风格完成')}>同步网站风格</Button><Button disabled={!selectedTaskKeys.length} loading={webSubmitLoading} onClick={handleWebSubmitPreview}>生成提交预览</Button><Button type="primary" disabled={!webSubmitPreviewReady} loading={webSubmitLoading} onClick={confirmWebSubmit}>开始自动提交</Button></Space></div>
    <div className="novel-fetch-panel novel-fetch-submit-settings"><Typography.Title level={5}>提交配置</Typography.Title><div className="novel-fetch-submit-config-grid"><label>组织归属<Select value={webSubmitConfig?.organization || undefined} placeholder="跟随网站配置档" options={(webSubmitConfig?.organization_catalog || []).map(item => ({ value: String(item.id || item.value), label: item.name || item.label || item.id }))} onChange={value => setWebSubmitConfig(current => ({ ...(current || {}), organization: value }))} /></label><label>最少字数<InputNumber min={0} value={webSubmitConfig?.min_text_chars ?? 0} onChange={value => setWebSubmitConfig(current => ({ ...(current || {}), min_text_chars: value || 0 }))} /></label><label>失败重试<InputNumber min={0} max={10} value={webSubmitConfig?.retry_times ?? 1} onChange={value => setWebSubmitConfig(current => ({ ...(current || {}), retry_times: value || 0 }))} /></label><label className="novel-fetch-submit-check"><Checkbox checked={webSubmitConfig?.skip_submitted !== false} onChange={event => setWebSubmitConfig(current => ({ ...(current || {}), skip_submitted: event.target.checked }))}>跳过已提交版本</Checkbox></label></div><Button onClick={() => handleWebSubmit(() => saveWebSubmitConfig(webSubmitConfig), '提交配置已保存')}>保存提交配置</Button></div>
    <div className="novel-fetch-panel novel-fetch-submit-summary"><Typography.Text>已选任务 <strong>{selectedTaskKeys.length}</strong></Typography.Text><Typography.Text>配置版本 <strong>{(webSubmitConfig?.submit_versions || webSubmitConfig?.submitVersions || []).join('、') || '默认'}</strong></Typography.Text><Typography.Text>可用组织 <strong>{webSubmitConfig?.organization_catalog?.length || 0}</strong></Typography.Text></div>
    {webSubmitPlan ? <><Alert type={webSubmitPreviewReady ? 'success' : 'warning'} message={webSubmitPreviewReady ? '当前选择的提交预览已生成' : '任务选择已变化，请重新生成预览'} description={`提交组：${webSubmitPlan.groups?.length || 0}，跳过：${webSubmitPlan.skipped?.length || 0}`} /><Table className="novel-fetch-panel" size="small" rowKey={(group, index) => group.group_id || index} pagination={false} dataSource={webSubmitPlan.groups || []} columns={[{ title: '版本', dataIndex: 'version' }, { title: '配置档', render: (_, group) => group.summary?.profile_name || group.summary?.profile_id || '默认配置' }, { title: '组织', render: (_, group) => group.summary?.organization || '未设置' }, { title: '风格', render: (_, group) => group.summary?.style || '未分类' }, { title: '任务数', render: (_, group) => group.items?.length || 0 }]} /><Table className="novel-fetch-panel" size="small" rowKey={(item, index) => `${item.id || index}-${item.version || ''}`} pagination={false} locale={{ emptyText: '没有跳过项' }} dataSource={webSubmitPlan.skipped || []} columns={[{ title: '跳过任务', dataIndex: 'id' }, { title: '版本', dataIndex: 'version' }, { title: '状态', dataIndex: 'status' }, { title: '原因', dataIndex: 'error' }]} /></> : null}
  </Space>;
  const submitResultPanel = <Table className="novel-fetch-panel" scroll={{ x: 720 }} rowKey={(row, index) => `${row.bookId || row.book_id}-${row.version}-${index}`} dataSource={webSubmitResults} locale={{ emptyText: '暂无本次提交结果' }} columns={[{ title: '任务', render: (_, row) => row.bookId || row.book_id }, { title: '版本', dataIndex: 'version' }, { title: '状态', dataIndex: 'status' }, { title: '结果', render: (_, row) => row.result || row.message || row.error || '-' }, { title: '操作', render: (_, row) => /error|fail|失败/i.test(String(row.status || '')) ? <Button size="small" loading={webSubmitLoading} onClick={() => retryWebSubmitVersion(row)}>重试该版本</Button> : '-' }]} />;
  const submitHistoryPanel = <div className="novel-fetch-panel"><div className="novel-fetch-toolbar"><Typography.Title level={5}>提交历史</Typography.Title><span className="novel-fetch-toolbar-spacer" /><Button size="small" icon={<RefreshCw size={14} />} loading={webSubmitLoading} onClick={loadWebSubmit}>刷新历史</Button></div><Table size="small" scroll={{ x: 760 }} rowKey={(row, index) => `${row.book_id || row.bookId}-${row.version}-${index}`} dataSource={webSubmitHistory} pagination={{ pageSize: 10 }} columns={[{ title: '书名', render: (_, row) => row.book_name || row.bookName || row.book_id || row.bookId }, { title: '版本', dataIndex: 'version' }, { title: '状态', dataIndex: 'status' }, { title: '结果', dataIndex: 'result' }, { title: '时间', render: (_, row) => row.time || row.created_at || row.createdAt }]} /></div>;
  const submitLogsPanel = <div className="novel-fetch-panel"><div className="novel-fetch-toolbar"><Typography.Title level={5}>操作日志</Typography.Title><span className="novel-fetch-toolbar-spacer" /><Button size="small" icon={<RefreshCw size={14} />} onClick={loadGlobalLogs}>刷新日志</Button></div><Select value={selectedTask || undefined} className="novel-fetch-log-task-select" placeholder="选择任务查看日志" options={tasks.map(task => ({ value: task.bookId || task.book_id, label: task.bookName || task.book_name || task.bookId || task.book_id }))} onChange={selectTask} />{selectedTaskData ? <Collapse items={(selectedTaskData.logs || []).map((log, index) => ({ key: String(index), label: `${log.time || ''} ${log.event || '操作日志'}`, children: <Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }}>{log.detail || JSON.stringify(log)}</Typography.Paragraph> }))} /> : <Table size="small" pagination={{ pageSize: 20 }} rowKey={(row, index) => row.id || index} dataSource={globalLogs} locale={{ emptyText: '暂无操作记录' }} columns={[{ title: '时间', dataIndex: 'time', render: (value, row) => value || row.createdAt || row.created_at }, { title: '事件', dataIndex: 'event' }, { title: '详情', render: (_, row) => String(row.detail || row.message || '').slice(0, 120) }]} />}</div>;
  const submissionPanel = <Space direction="vertical" size={12} style={{ width: '100%' }}><NovelFetchMetricStrip items={submissionStatus} /><Tabs className="novel-fetch-section-tabs" activeKey={submitView} onChange={setSubmitView} items={[{ key: 'ready', label: '提交准备', children: submitReadyPanel }, { key: 'result', label: `本次结果${webSubmitResults.length ? ` (${webSubmitResults.length})` : ''}`, children: submitResultPanel }, { key: 'history', label: '提交历史', children: submitHistoryPanel }, { key: 'logs', label: '操作日志', children: submitLogsPanel }]} /></Space>;
  const currentBatchMenuItems = [
    { key: 'retry-failed', label: '重试当前批次失败', onClick: retryFailedTasks },
    { key: 'factory', label: '当前批次进入批量工厂', onClick: transferToBatchFactory },
    { key: 'rules', label: '规则处理当前批次', onClick: applyAllRules },
    { key: 'submit', label: '提交当前批次', onClick: () => submitVisibleTasks(false) },
    { key: 'resubmit', label: '重新提交当前批次', onClick: () => submitVisibleTasks(true) },
    { type: 'divider' },
    { key: 'delete-failed', danger: true, label: '删除失败', onClick: () => deleteTasksByFilter(task => /fail|error/.test(String(task.status || task.aiStatus || task.ai_status)), '当前没有失败任务') },
    { key: 'delete-all', danger: true, label: '删除全部', onClick: () => deleteTasksByFilter(() => true, '当前没有任务') }
  ];
  const panels = {
    process: <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <div className="novel-fetch-process-grid">
      <div className="novel-fetch-process-main">
      <Form className="novel-fetch-public-workspace" form={processForm} layout="vertical" initialValues={{ platform: savedProcessSettings.platform, maxTxt: savedProcessSettings.maxTxt }} onValuesChange={changedValues => {
        if ('bookIds' in changedValues) setInputPreview(null);
        writeNovelFetchProcessSettings({
          ...savedProcessSettings,
          ...processForm.getFieldsValue(),
          selectedVersions,
          rewriteMethods,
          sensitiveAiProcessEnabled,
          textModel
        });
      }}>
        <div className="novel-fetch-public-heading"><div><Typography.Title level={4}>批量输入</Typography.Title></div><Typography.Text type="secondary">平台ID：{selectedPlatform || '—'}</Typography.Text></div>
        <div className="novel-fetch-public-grid novel-fetch-public-grid--three">
          <Form.Item name="platform" label="平台"><Select options={platformOptions} onChange={value => setSelectedPlatform(String(value || ''))} /></Form.Item>
          <label className="novel-fetch-field"><span>文本模型</span><Select value={textModel} onChange={setTextModel} options={textModelOptions.length ? textModelOptions : [{ value: textModel || '[anti-2]gemini-3.8-flash', label: textModel || '[anti-2]gemini-3.8-flash' }]} /></label>
          <div className="novel-fetch-field novel-fetch-top-login"><span>视频管理系统登录</span><div className="novel-fetch-login-inline"><span className={uploadSession?.status === 'ready' ? 'is-ready' : ''}>● {uploadSessionDisplayName(uploadSession)}</span><Button size="small" onClick={openUploadManager}>管理</Button></div></div>
        </div>
        <div className="novel-fetch-public-grid novel-fetch-public-grid--two novel-fetch-input-format-row"><label className="novel-fetch-field"><span>输入格式</span><Select value={inputFormat} options={INPUT_FORMAT_OPTIONS.map(item => ({ value: item.value, label: item.label }))} onChange={value => { setInputFormat(value); setInputPreview(null); }} /></label><div className="novel-fetch-input-summary"><span>有效 <b>{localInputPreview.validRows}</b> 本</span>{localInputPreview.duplicateIds ? <span className="is-warning">重复 {localInputPreview.duplicateIds}</span> : null}{localInputPreview.emptyIdRows ? <span className="is-warning">空 ID {localInputPreview.emptyIdRows}</span> : null}</div></div>
        <Form.Item name="bookIds" label="待处理清单（一行一本书）" rules={[{ required: true, message: '请输入至少一个书籍 ID' }]}><Input.TextArea className={`novel-fetch-public-input ${String(bookIdsValue).trim() ? 'has-content' : ''}`} rows={5} placeholder={inputFormatOption.description} /></Form.Item>
        <div className="novel-fetch-input-preview"><div><strong>解析预览</strong>{inputPreview ? <span>已识别 {inputPreview.unique_tasks ?? inputPreview.validRows ?? 0} 本</span> : <span>尚未生成</span>}</div><Space><Button disabled={!String(bookIdsValue).trim()} onClick={async () => { const values = await processForm.validateFields(); await resolveInputPreview(values); }}>解析预览</Button>{inputPreview?.rows?.length ? <Tag>{inputPreview.rows.slice(0, 3).map(row => row.bookName || row.bookId).join('、')}{inputPreview.rows.length > 3 ? '…' : ''}</Tag> : null}</Space></div>
        <section className="novel-fetch-public-section novel-fetch-settings-row"><div className="novel-fetch-settings-col"><span className="novel-fetch-settings-label">本次处理设置</span><div className="novel-fetch-content-limit"><span className="novel-fetch-limit-chip"><span>截取字数：</span><InputNumber min={100} max={100000} value={processForm.getFieldValue('maxTxt') || 4000} onChange={value => processForm.setFieldValue('maxTxt', value || 4000)} /></span></div></div><div className="novel-fetch-settings-col"><span className="novel-fetch-settings-label">AI 文案版本</span><div className="novel-fetch-version-grid">{AI_VERSIONS.map(version => <label key={version} className={`novel-fetch-version-option ${selectedVersions.includes(version) ? 'is-selected' : ''}`}><input type="checkbox" checked={selectedVersions.includes(version)} onChange={event => setSelectedVersions(current => event.target.checked ? [...current, version] : current.filter(item => item !== version))} /><span>{AI_VERSION_LABELS[version]}</span></label>)}</div></div></section>
        <div className="novel-fetch-public-grid novel-fetch-public-grid--two novel-fetch-process-policy-row"><label className="novel-fetch-field"><span>改文方式</span><div className="novel-fetch-readonly-field">AI1 自动轮换 <span aria-hidden="true">⌄</span></div></label><label className="novel-fetch-field novel-fetch-sensitive-field"><span>敏感处理</span><span className="novel-fetch-check-line"><input type="checkbox" checked={sensitiveAiProcessEnabled} onChange={event => setSensitiveAiProcessEnabled(event.target.checked)} /> 启用敏感词 AI 修复（关闭则直接替换）</span></label></div>
        <div className="novel-fetch-public-actions"><Button type="primary" danger loading={processing} onClick={handleProcess}>开始处理</Button><Button onClick={() => openScheduleDialog()}>开始定时</Button><Button onClick={() => openScheduleDialog(true)}>定时任务</Button><Button onClick={handleSensitiveRepair}>敏感词 AI 修复</Button><Button icon={<RefreshCw size={14} />} loading={tasksLoading} onClick={() => { loadTasks(); loadCurrentBatch(); loadHistoryBatches(); }}>刷新任务</Button></div>
      </Form>
      <Modal className="novel-fetch-schedule-modal" open={scheduleModalOpen} width={640} title={<div className="novel-fetch-schedule-title"><span className="novel-fetch-schedule-icon"><CalendarClock size={18} /></span><div><strong>定时任务</strong><small>一次性任务 · 由 Go 后端统一调度</small></div></div>} okText={editingScheduleId ? '保存调整' : '保存定时任务'} cancelText="取消" confirmLoading={scheduleSaving} onCancel={() => { setEditingScheduleId(''); setScheduleModalOpen(false); }} onOk={handleScheduleProcess}>
        <div className="novel-fetch-schedule-note"><strong>保存后不会立即执行</strong><span>到达设定时间后加入处理队列，处理完成会默认自动上传到视频管理系统。</span></div>
        <div className="novel-fetch-schedule-create">
          <div className="novel-fetch-schedule-create-heading"><div><strong>{editingScheduleId ? '调整定时任务' : '新建定时任务'}</strong><span>{editingScheduleId ? '只调整开始时间，不改变原任务内容' : `当前清单 ${parseIds(bookIdsValue).length} 本 · ${selectedVersions.map(version => AI_VERSION_LABELS[version]).join('、') || '未选择版本'}`}</span></div><span className="novel-fetch-schedule-timezone">本机时间</span></div>
          <label className="novel-fetch-field"><span>开始时间</span><div className="novel-fetch-schedule-input-wrap"><Input ref={scheduleInputRef} type="datetime-local" min={scheduleInputValue(new Date(Date.now() + 60 * 1000))} value={scheduleTime} onChange={event => setScheduleTime(event.target.value)} onClick={event => event.currentTarget.showPicker?.()} onKeyDown={event => { if (event.key === 'Enter') event.currentTarget.showPicker?.(); }} /><CalendarClock className="novel-fetch-schedule-input-icon" size={16} aria-hidden="true" /></div></label>
        </div>
        <div className="novel-fetch-schedule-list-heading"><strong>已保存任务</strong><span>{schedules.length} 个</span></div>
        {schedules.length ? <div className="novel-fetch-schedule-list">{schedules.map(row => <div className="novel-fetch-schedule-row" key={row.id}><div className="novel-fetch-schedule-row-main"><CalendarClock size={15} /><div><strong>{formatScheduleTime(row.run_at || row.runAt)}</strong><small>创建于 {formatScheduleTime(row.created_at || row.createdAt)} · 默认自动上传</small></div></div><span className={`novel-fetch-schedule-status is-${row.status || 'scheduled'}`}>{scheduleStatusLabel(row.status)}</span><Space className="novel-fetch-schedule-row-actions" size={6}>{(!row.status || row.status === 'scheduled') && <Button size="small" onClick={() => editSchedule(row)}>调整时间</Button>}<Button size="small" danger onClick={async () => { await deleteWorkshopSchedule(row.id); await loadSchedules(); }}>删除</Button></Space></div>)}</div> : <div className="novel-fetch-schedule-empty"><CalendarClock size={24} /><strong>还没有定时任务</strong><span>选择开始时间并保存后，任务会显示在这里。</span></div>}
       </Modal>
        </div>
      <div className="novel-fetch-process-side">
      <div className={`novel-fetch-automation-bar ${queueSummary.enabled ? 'is-enabled' : 'is-paused'}`}>
        <div className="novel-fetch-automation-heading"><div><strong>全自动队列</strong><span>{queueSummary.enabled ? '持续处理、自动上传并回读确认' : '已暂停，不会派发新的自动任务'}</span></div><Button type={queueSummary.enabled ? 'default' : 'primary'} danger={queueSummary.enabled} loading={automationChanging} onClick={toggleAutomation}>{queueSummary.enabled ? '暂停全自动' : '开启全自动'}</Button></div>
        <div className="novel-fetch-automation-facts">
          <span>已选 <b>{selectedTaskKeys.length}</b> 个</span><span>自动队列 <b>{queueSummary.enabled ? '已开启' : '已暂停'}</b></span><span>当前任务 <b>{queueSummary.activeJobId || '无'}</b></span><span>处理中 <b>{queueSummary.running}</b> 个</span><span>排队 <b>{queueSummary.queued}</b> 个</span>
        </div>
        <div className="novel-fetch-automation-foot"><span>最近任务 {queueSummary.latestJob?.status || '暂无'}：{queueSummary.latestJob?.id || '—'}</span><span>刷新 {queueSummary.updatedAt ? formatScheduleTime(queueSummary.updatedAt) : '—'}</span></div>
      </div>
      <div className="novel-fetch-status-layout">
        <div className="novel-fetch-public-result"><div className="novel-fetch-result-heading"><div className="novel-fetch-public-result-title">处理状态</div><span className="novel-fetch-result-side">当前批次</span></div><div className="novel-fetch-stat-grid">
          {[['总数量', processSummary.total], ['原文就绪', processSummary.originalDone], ['处理中', processSummary.processing], ['已完成', processSummary.aiDone], ['失败', processSummary.failed]].map(([label, value]) => <div className="novel-fetch-stat-card" key={label}><span>{label}</span><strong>{value}</strong></div>)}
        </div><div className="novel-fetch-progress-list"><div><span>原文获取</span><i style={{ '--progress': `${progressPercent(processSummary.originalDone)}%` }} /><b>{processSummary.originalDone} / {processSummary.total}</b></div><div><span>AI 文案生成</span><i style={{ '--progress': `${progressPercent(processSummary.aiDone)}%` }} /><b>{processSummary.aiDone} / {processSummary.total}</b></div><div><span>视频管理系统提交</span><i style={{ '--progress': `${progressPercent(processSummary.submissionDone)}%` }} /><b>{processSummary.submissionDone} / {processSummary.total}</b></div></div>
        <div className="novel-fetch-live-log-title">实时日志</div><div className="novel-fetch-live-log"><pre>{liveLogLines.join('\n')}</pre></div></div>
      </div>
      <div className="novel-fetch-public-actions novel-fetch-process-secondary-actions"><Button onClick={retryFailedTasks}>重试当前失败</Button><Button onClick={applyAllRules}>处理当前规则</Button><Button onClick={() => submitVisibleTasks(false)}>提交当前批次</Button><label className="novel-fetch-factory-version-picker"><span>入厂内容</span><Select aria-label="入厂内容版本" value={batchFactorySourceVersion} onChange={setBatchFactorySourceVersion} options={['sensitive', 'original', 'ai1', 'ai2', 'ai3', 'ai4', 'ai5'].map(value => ({ value, label: batchFactoryVersionLabel(value) }))} /></label><Button onClick={transferToBatchFactory}>当前批次入厂</Button></div>
      <div className="novel-fetch-workbench-shortcuts" aria-label="工作台功能入口">
        <button type="button" className="novel-fetch-workbench-shortcut" onClick={() => changeTab('knowledge')}>
          <strong>知识库</strong><span>知识库 + 处理规则</span>
        </button>
        <button type="button" className="novel-fetch-workbench-shortcut" onClick={() => changeTab('config')}>
          <strong>配置</strong><span>打开原配置页面</span>
        </button>
        <button type="button" className="novel-fetch-workbench-shortcut" onClick={() => changeTab('submit')}>
          <strong>提交与记录</strong><span>提交、历史、日志</span>
        </button>
       </div>
       </div>
       </div>
       {currentBatch ? <div className="novel-fetch-panel novel-fetch-current-batch"><div className="novel-fetch-current-batch-heading"><Typography.Title level={5}>当前批次</Typography.Title><Space><Button size="small" icon={<RefreshCw size={14} />} onClick={loadCurrentBatch}>刷新</Button><Button size="small" danger disabled={!currentBatchStoppableIds.length} onClick={() => Modal.confirm({ title: '停止当前批次？', content: '排队中和正在执行的任务会停止，已完成任务不会改变。', okText: '确认停止', okButtonProps: { danger: true }, onOk: async () => { await stopWorkshopTasks(currentBatchStoppableIds); await Promise.all([loadTasks(), loadCurrentBatch()]); message.success('已停止当前批次'); } })}>停止处理</Button><Button size="small" onClick={() => { setTaskSearch(''); setTaskDate(''); setTaskView('current'); changeTab('tasks'); }}>查看全部任务</Button></Space></div><Typography.Text type="secondary">{currentBatch.id} · {currentBatchIds.length} 本</Typography.Text><Table className="novel-fetch-current-batch-table" size="small" pagination={false} scroll={{ x: 1060 }} rowKey={row => row.bookId || row.book_id} dataSource={currentBatch.taskStates || currentBatch.task_states || []} columns={[{ title: 'ID', render: (_, row) => row.id || row.index || row.bookId || row.book_id }, { title: '书名', render: (_, row) => row.bookName || row.book_name || row.bookId || row.book_id }, { title: 'Book ID', render: (_, row) => row.bookId || row.book_id }, { title: 'AI1', render: (_, row) => novelFetchVersionReady(row, 'ai1') ? '✓' : '×' }, { title: '网站提交', render: (_, row) => row.siteSubmitStatus || row.site_submit_status || '○' }, { title: '状态', render: (_, row) => <TaskStatusTag row={row} /> }, { title: '字数', render: (_, row) => novelFetchCharCount(row) }, { title: '创建时间', render: (_, row) => row.createdAt || row.created_at || '-' }, { title: '操作', render: (_, row) => <Button size="small" onClick={async () => { const target = novelFetchCurrentBatchTaskDetail(row); if (!target) return; setTaskDetail({ ...target, data: await getWorkshopTask(target.bookId) }); }}>查看</Button> }]} /></div> : null}
       <Modal className="novel-fetch-upload-manager-modal" open={loginModalOpen} width={620} title="视频管理系统" cancelText="取消" okText="保存并关闭" confirmLoading={uploadProfileLoading} okButtonProps={{ disabled: uploadSession?.status !== 'ready' || !uploadProfileDraft }} onCancel={() => { setLoginModalOpen(false); setLoginFormVisible(false); setLoginPassword(''); }} onOk={saveUploadProfileAndClose}>
        <div className={`novel-fetch-upload-manager-status ${uploadSession?.status === 'ready' ? 'is-ready' : ''}`}>
          <div><strong>{uploadSession?.status === 'ready' ? uploadSessionDisplayName(uploadSession) : '尚未登录'}</strong><span>{uploadSession?.status === 'ready' ? '登录会话可用于配置同步和网站提交' : '请先登录视频管理系统，再同步网站配置档'}</span></div>
          <Button size="small" onClick={() => setLoginFormVisible(value => !value)}>{uploadSession?.status === 'ready' ? '更换账号' : '登录'}</Button>
        </div>
        {loginFormVisible ? <Form className="novel-fetch-upload-manager-login" layout="vertical">
          <div className="novel-fetch-upload-manager-login-fields"><Form.Item label="账号"><Input value={loginUser} onChange={event => setLoginUser(event.target.value)} autoComplete="username" /></Form.Item><Form.Item label="密码"><Input.Password value={loginPassword} onChange={event => setLoginPassword(event.target.value)} autoComplete="current-password" /></Form.Item></div>
          <Button type="primary" loading={webSubmitLoading} onClick={handleUploadLogin}>登录并验证</Button>
        </Form> : null}
        <section className="novel-fetch-upload-profile-section">
          <div className="novel-fetch-upload-profile-heading"><div><strong>默认网站配置档</strong><span>日常提交默认使用；需要分版本时再进入高级配置。</span></div><Tag color={uploadProfiles.length ? 'blue' : 'default'}>{uploadProfiles.length} 个可用</Tag></div>
          <Select showSearch value={uploadProfileDraft || undefined} loading={uploadProfileLoading} disabled={uploadSession?.status !== 'ready'} placeholder={uploadSession?.status === 'ready' ? '同步网站配置档后选择' : '请先登录视频管理系统'} options={uploadProfiles.map(profile => ({ value: String(profile.id || profile.name), label: profile.name || profile.id }))} onChange={setUploadProfileDraft} />
          <div className="novel-fetch-upload-profile-actions"><Button disabled={uploadSession?.status !== 'ready'} loading={uploadProfileLoading} onClick={() => syncUploadProfiles(true)}>同步网站配置档</Button><Button disabled={uploadSession?.status !== 'ready'} loading={uploadProfileLoading} onClick={syncUploadStylesFromManager}>同步风格目录</Button><Button type="link" onClick={openAdvancedVersionProfiles}>调整版本对应关系</Button></div>
        </section>
      </Modal>
    </Space>,
    tasks: <div className="novel-fetch-task-page">
      <Tabs className="novel-fetch-section-tabs" activeKey={taskView} onChange={value => { setTaskView(value); if (value === 'history') loadHistoryBatches(); }} items={[{ key: 'current', label: '当前任务' }, { key: 'history', label: '历史批次' }]} />
      {taskView === 'current' ? <>
        <div className="novel-fetch-task-overview">
          <div className="novel-fetch-queue-summary"><Typography.Text type="secondary">自动队列</Typography.Text><Tag color={queueSummary.running ? 'processing' : queueSummary.queued || queueSummary.waitingRetry ? 'gold' : 'default'}>{queueSummary.label}</Tag><span>执行中 {queueSummary.running}</span><span>排队 {queueSummary.queued}</span><span>等待重试 {queueSummary.waitingRetry}</span>{processJob?.id ? <small>最近任务 {processJob.id}</small> : null}{queueSummary.updatedAt ? <small>更新 {formatScheduleTime(queueSummary.updatedAt)}</small> : null}</div>
          <div><Typography.Text type="secondary">默认范围</Typography.Text><strong>近 3 天成功 + 历史未完成</strong></div>
          <div><Typography.Text type="secondary">全部任务</Typography.Text><strong>{tasks.length}</strong></div>
          <div><Typography.Text type="secondary">当前显示</Typography.Text><strong>{visibleTasks.length}</strong></div>
        </div>
        <div className="novel-fetch-toolbar novel-fetch-task-date-toolbar">
          <Input allowClear className="novel-fetch-task-search" value={taskSearch} onChange={event => setTaskSearch(event.target.value)} placeholder="搜索书籍 ID 或书名" />
          <Typography.Text type="secondary">查看日期</Typography.Text>
          <Button onClick={() => shiftTaskDate(-1)}>前一天</Button>
          <Input className="novel-fetch-date-input" type="date" value={taskDate} onChange={event => setTaskDate(event.target.value)} />
          <Button onClick={() => setTaskDate(beijingDateKey())}>今天</Button>
          <Button onClick={() => shiftTaskDate(1)}>后一天</Button>
          <Button type={!taskDate ? 'primary' : 'default'} onClick={() => setTaskDate('')}>默认视图</Button>
          <span className="novel-fetch-toolbar-spacer" />
          <Button onClick={selectAllTasks}>全选当前结果</Button>
          <Button onClick={() => setSelectedTaskKeys([])}>清空选择</Button>
        </div>
        <div className="novel-fetch-toolbar novel-fetch-task-global-actions">
          <Typography.Text strong>全部任务操作</Typography.Text>
          <Button onClick={retryAllFailedTasks}>重试全部失败</Button>
          <Select className="novel-fetch-rule-scope" value={ruleApplyScope} onChange={setRuleApplyScope} options={[{ value: 'both', label: '规则处理原文 + AI' }, { value: 'original', label: '只处理原文' }, { value: 'ai', label: '只处理 AI' }]} />
          <Button onClick={applyRulesToAllTasks}>规则处理全部</Button>
          <Button type="primary" onClick={() => Modal.confirm({ title: '提交全部任务？', content: `将提交任务列表中的 ${tasks.length} 个任务。`, onOk: submitAllTasks })}>提交全部</Button>
          <span className="novel-fetch-toolbar-spacer" />
          <Button danger onClick={() => deleteTasksByFilter(task => novelFetchTaskStatus(task).status === 'failed', '当前没有失败任务')}>删除失败</Button>
          <Button danger onClick={() => deleteTasksByFilter(() => true, '当前没有任务')}>删除全部</Button>
        </div>
        {selectedTaskKeys.length ? <div className="novel-fetch-task-selection-bar">
          <Typography.Text strong>已选 {selectedTaskKeys.length} 个任务</Typography.Text>
          <Button onClick={handleRetryTasks}>立即重试选中</Button>
          <Button danger onClick={async () => { await stopWorkshopTasks(selectedTaskKeys); await loadTasks(); message.success('已停止选中任务'); }}>停止选中</Button>
          <Button onClick={handleGenerateAi}>生成 AI 改文</Button>
          <Button onClick={handleSensitiveRepair}>敏感词 AI 修复</Button>
          <Typography.Text type="secondary">AI 数量</Typography.Text>
          <InputNumber min={1} max={20} value={batchAiCount} onChange={value => setBatchAiCount(value || 1)} />
          <Button onClick={updateSelectedAiCount}>调整选中</Button>
          <Button onClick={applySelectedRules}>规则处理选中</Button>
          <Button icon={<Factory size={16} />} title={`入厂内容：${batchFactoryVersionLabel(batchFactorySourceVersion)}`} onClick={transferToBatchFactory}>进入批量工厂</Button>
          <Button type="primary" onClick={() => changeTab('submit')}>提交选中</Button>
          <Button danger onClick={handleDeleteTasks}>删除选中</Button>
        </div> : null}
        <Table className="novel-fetch-task-table" scroll={{ x: 1680 }} rowKey={row => row.bookId || row.book_id} loading={tasksLoading} dataSource={visibleTasks} rowSelection={{ selectedRowKeys: selectedTaskKeys, onChange: setSelectedTaskKeys }} pagination={{ pageSize: 20, showSizeChanger: true }} columns={[
          { title: '书籍 ID', width: 130, fixed: 'left', render: (_, row) => row.bookId || row.book_id },
          { title: '书名', width: 180, render: (_, row) => row.bookName || row.book_name || row.bookId || row.book_id },
          { title: '平台', width: 110, render: (_, row) => row.platformName || row.platform_name || '-' },
          { title: '风格', width: 110, dataIndex: 'style', render: value => value || '待判定' },
          { title: '男女频', width: 90, dataIndex: 'gender', render: value => value || '待确认' },
          { title: 'AI 判断', width: 110, render: (_, row) => row.classify_status === 'done' || row.classifyStatus === 'done' || (row.gender && row.style && row.style !== '待判定') ? '已完成' : row.classify_error || row.classifyError || '待判断' },
          { title: '原文', width: 220, render: (_, row) => { const retry = novelFetchOriginalFetchRetry(row); return <Space size={4} direction="vertical"><Tag color={statusColor(row.originalStatus || row.original_status)}>{novelFetchStageLabel(row.originalStatus || row.original_status || '未抓取')}</Tag><Typography.Text type="secondary">{novelFetchCharCount(row)}</Typography.Text>{retry.nextRetryAt ? <Typography.Text type="warning" title={retry.error || undefined}>下次自动抓取：{formatScheduleTime(retry.nextRetryAt)}{retry.attempts ? `（第 ${retry.attempts} 次）` : ''}</Typography.Text> : null}</Space>; } },
          { title: 'AI 文案', width: 120, render: (_, row) => <Tag color={statusColor(row.aiStatus || row.ai_status)}>{row.aiGeneratedCount || row.ai_generated_count ? `${row.aiGeneratedCount || row.ai_generated_count} 版` : novelFetchStageLabel(row.aiStatus || row.ai_status || '未生成')}</Tag> },
          { title: '网站提交', width: 120, render: (_, row) => novelFetchSubmissionLabel(row) },
          { title: '创建/推送时间', width: 190, render: (_, row) => row.push_date || row.pushDate || row.createdAt || row.created_at || '-' },
          { title: '状态', width: 120, render: (_, row) => <TaskStatusTag row={row} /> },
          { title: '操作', width: 390, fixed: 'right', render: (_, row) => {
            const bookId = row.bookId || row.book_id;
            const openDetail = async initialTab => setTaskDetail({ bookId, initialTab, data: await getWorkshopTask(bookId) });
            return <Space wrap>
              <Button size="small" icon={<Eye size={14} />} onClick={() => openDetail('summary')}>详情</Button>
              <Button size="small" onClick={() => openDetail('sensitive')}>敏感日志</Button>
              <Button size="small" onClick={() => openDetail('submit')}>提交日志</Button>
              <Button size="small" onClick={() => fetchWorkshopOriginal(bookId, row.maxTxt || row.max_txt).then(loadTasks)}>重新抓取</Button>
              <Button size="small" onClick={() => generateWorkshopAi(bookId, 1).then(loadTasks)}>生成 AI</Button>
              <Button size="small" onClick={() => restoreWorkshopOriginal(bookId).then(loadTasks)}>恢复原文</Button>
            </Space>;
          } }
        ]} />
        <div className="novel-fetch-task-summary"><span>当前批次 <strong>{currentBatch?.id || '—'}</strong></span><span>完成 <strong>{processSummary.aiDone}</strong></span><span>执行中 <strong>{processSummary.processing}</strong></span><span>失败 <strong>{processSummary.failed}</strong></span><Dropdown trigger={['click']} menu={{ items: currentBatchMenuItems }}><Button className="novel-fetch-task-batch-menu">当前批次操作</Button></Dropdown></div>
      </> : <div className="novel-fetch-history-panel"><div className="novel-fetch-toolbar"><Typography.Title level={5}>历史批次</Typography.Title><span className="novel-fetch-toolbar-spacer" /><Button icon={<RefreshCw size={14} />} loading={tasksLoading} onClick={loadHistoryBatches}>刷新历史批次</Button></div><Table scroll={{ x: 720 }} size="small" rowKey="id" dataSource={historyBatches.filter(batch => batch.id !== currentBatch?.id)} columns={[{ title: '批次', dataIndex: 'id' }, { title: '书籍数', render: (_, row) => row.taskIds?.length || 0 }, { title: '状态', dataIndex: 'status' }, { title: '操作', render: (_, row) => <Space><Button size="small" onClick={async () => { const data = await rerunWorkshopBatch(row.id, 'all'); processForm.setFieldValue('bookIds', data.input_snapshot || ''); changeTab('process'); message.success('历史批次已载入处理清单'); }}>全部重跑</Button><Button size="small" disabled={!row.taskIds?.length} onClick={async () => { const data = await rerunWorkshopBatch(row.id, 'abnormal'); processForm.setFieldValue('bookIds', data.input_snapshot || ''); changeTab('process'); message.success('异常任务已载入处理清单'); }}>重跑异常</Button></Space> }]} /></div>}
    </div>,
     config: <NovelFetchConfigPanel form={configForm} effectiveValues={configEffective} dirty={configDirty} onDirty={() => setConfigDirty(true)} versionPanel={versionConfigPanel} section={configSection} onSectionChange={setConfigSection} />,
    knowledge: <Space direction="vertical" size={12} style={{ width: '100%' }}><div className="novel-fetch-panel"><Typography.Title level={5}>知识库使用提示词</Typography.Title><Input.TextArea rows={4} value={knowledgeUsagePrompt} onChange={event => setKnowledgeUsagePrompt(event.target.value)} placeholder="告诉 AI 如何参考知识库，且不得照搬示例内容" /><Button type="primary" style={{ marginTop: 8 }} onClick={saveKnowledgeUsagePrompt}>保存知识库配置</Button></div><div className="novel-fetch-panel"><Typography.Title level={5}>三大知识库</Typography.Title><Typography.Text type="secondary">资料库只供 AI 自动取用，任务处理仍然全自动</Typography.Text><div className="novel-fetch-knowledge-summary">{[['开头词', knowledgeSummary.opening_phrases], ['改文模板', knowledgeSummary.rewrite_profiles], ['临时指令', knowledgeSummary.temporary_instructions], ['高仿提示词', knowledgeSummary.high_imitation_prompts], ['高仿参考', knowledgeSummary.high_imitation_references], ['敏感词', knowledgeSummary.sensitive_rules], ['字位符号', knowledgeSummary.symbol_rules], ['成对补符号', knowledgeSummary.pair_fill_rules], ['章节整行', knowledgeSummary.chapter_exact_rules], ['章节开头', knowledgeSummary.chapter_inline_rules]].map(([label, count]) => <Tag key={label}>{label} {Number(count || 0)}</Tag>)}</div></div><div className="novel-fetch-panel"><Tabs activeKey={knowledgeKind} onChange={setKnowledgeKind} items={KNOWLEDGE_TABS.map(([key, label]) => ({ key, label: `${label}（${knowledgeTabCount(key)}）` }))} /><Space wrap><Input allowClear value={knowledgeSearch} onChange={event => setKnowledgeSearch(event.target.value)} placeholder="搜索标题、分类、标签或内容" style={{ width: 300 }} /><Button onClick={() => { knowledgeForm.resetFields(); setKnowledgeDetails('{}'); setKnowledgeModal(true); }}>新增条目</Button><Button loading={knowledgeLoading} onClick={loadKnowledge}>刷新</Button>{knowledgeKind === 'opening_phrases' ? <Button onClick={() => normalizeWorkshopOpening().then(loadKnowledge)}>规范化开头词</Button> : null}</Space></div>{knowledgeKind === 'opening_phrases' ? <div className="novel-fetch-panel"><Typography.Title level={5}>爆款开头拆解入库</Typography.Title><Space wrap style={{ width: '100%' }}><Input value={openingStyle} onChange={event => setOpeningStyle(event.target.value)} placeholder="适用风格类型（可选）" style={{ flex: 1, minWidth: 240 }} /><Input value={openingExtra} onChange={event => setOpeningExtra(event.target.value)} placeholder="补充要求（可选）" style={{ flex: 1, minWidth: 240 }} /></Space><Input.TextArea rows={5} value={openingText} onChange={event => setOpeningText(event.target.value)} placeholder="爆款开头原文" style={{ marginTop: 12 }} /><Input.TextArea rows={5} value={openingResult} onChange={event => setOpeningResult(event.target.value)} placeholder="拆解结果（分析后可编辑）" className="mono" style={{ marginTop: 12 }} /><Space style={{ marginTop: 8 }} wrap><Button type="primary" onClick={handleOpeningAnalyze}>AI 拆解开头词</Button><Button onClick={handleOpeningSaveAnalyzed}>确认加入开头词库</Button><Button onClick={() => { setOpeningText(''); setOpeningStyle(''); setOpeningExtra(''); setOpeningResult(''); setOpeningItems([]); }}>清空</Button></Space></div> : null}<Table className="novel-fetch-panel" rowKey="id" loading={knowledgeLoading} dataSource={(knowledge.items || []).filter(item => { const q = knowledgeSearch.trim().toLowerCase(); return !q || `${item.title || item.name || ''} ${item.category || item.style || ''} ${item.tags || ''} ${item.content || item.prompt || item.text || ''}`.toLowerCase().includes(q); })} columns={[{ title: '条目', render: (_, item) => item.title || item.name || item.text || item.id }, { title: '分类/标签', render: (_, item) => `${item.category || item.style || '自动'} ${Array.isArray(item.tags) ? item.tags.join('、') : item.tags || ''}` }, { title: '内容', render: (_, item) => String(item.content || item.prompt || item.text || item.find || '').slice(0, 120) }, { title: '操作', render: (_, item) => <Space><Button size="small" onClick={() => { const { id, ...details } = item; knowledgeForm.setFieldsValue({ ...item, details: JSON.stringify(details, null, 2) }); setKnowledgeDetails(JSON.stringify(details, null, 2)); setKnowledgeModal(true); }}>编辑</Button><Button size="small" onClick={() => saveWorkshopKnowledge(knowledgeKind, { ...item, id: undefined, title: `${item.title || item.name || '未命名'}（副本）` }).then(loadKnowledge)}>复制</Button><Button size="small" onClick={() => optimizeWorkshopKnowledge(knowledgeKind, item.id).then(loadKnowledge)}>AI 优化</Button><Button size="small" danger onClick={() => deleteWorkshopKnowledge(knowledgeKind, item.id).then(loadKnowledge)}>删除</Button></Space> }]} /><Modal open={knowledgeModal} title={`知识条目 · ${KNOWLEDGE_TABS.find(([key]) => key === knowledgeKind)?.[1] || ''}`} onCancel={() => setKnowledgeModal(false)} onOk={saveKnowledge}><Form form={knowledgeForm} layout="vertical"><Form.Item name="id" hidden><Input /></Form.Item><Form.Item name="title" label="名称/标题"><Input /></Form.Item><Space style={{ width: '100%' }} size={12}><Form.Item name="category" label="分类" style={{ flex: 1 }}><Input /></Form.Item><Form.Item name="style" label="适用风格" style={{ flex: 1 }}><Input /></Form.Item><Form.Item name="tags" label="标签" style={{ flex: 1 }}><Input placeholder="用逗号分隔" /></Form.Item></Space><Form.Item name="find" label="命中内容（规则库适用）"><Input /></Form.Item><Form.Item name="replace" label="替换内容（规则库适用）"><Input /></Form.Item><Form.Item name="content" label="内容/提示词"><Input.TextArea rows={7} /></Form.Item><Form.Item label="完整字段（JSON，可保留任意 V78 字段）"><Input.TextArea rows={8} value={knowledgeDetails} onChange={event => setKnowledgeDetails(event.target.value)} className="mono" /></Form.Item></Form></Modal></Space>,
    rules: <Space direction="vertical" size={12} style={{ width: '100%' }}><div className="novel-fetch-panel"><Space wrap><Select value={ruleKind} onChange={value => { setRuleKind(value); loadRuleConfig(value); }} options={[{ value: 'sensitive', label: '敏感词规则' }, { value: 'layout_rules', label: '批量排版规则' }, { value: 'chapter_exact_rules', label: '章节清洗：整行删除' }, { value: 'chapter_inline_rules', label: '章节清洗：行内命中' }, { value: 'symbol_rules', label: '字位符号规则' }, { value: 'pair_fill_rules', label: '成对补符号规则' }]} /><Select value={ruleScope} onChange={setRuleScope} options={[{ value: 'original', label: '原文' }, { value: 'ai', label: 'AI 版本' }, { value: 'both', label: '原文 + AI 版本' }]} /><Button disabled={!selectedTaskData?.original} onClick={() => setRuleText(selectedTaskData.original)}>载入当前任务原文</Button><Button type="primary" loading={ruleLoading} onClick={handleRulePreview}>预览规则</Button><Button loading={ruleLoading} onClick={handleRuleSuggest}>AI 规则建议</Button><Button onClick={saveRuleConfiguration}>保存规则配置</Button></Space><Typography.Text type="secondary">抓取原文后自动处理，AI 返回后也自动处理；当前编辑器保存到 Go 配置，处理页开关决定是否执行。</Typography.Text><RuleEditor kind={ruleKind} value={ruleConfigText} onChange={setRuleConfigText} /><div className="novel-fetch-rule-editor"><Typography.Title level={5}>规则 AI 助手</Typography.Title><Input.TextArea rows={4} value={ruleAiSample} onChange={event => setRuleAiSample(event.target.value)} placeholder="粘贴一段需要分析的原文或 AI 文案" /><Input.TextArea rows={3} value={ruleAiGoal} onChange={event => setRuleAiGoal(event.target.value)} placeholder="例如：提取敏感词、识别章节序号、建议排版参数" /><Space><Button type="primary" loading={ruleLoading} onClick={handleRuleSuggest}>AI 生成当前规则</Button>{ruleResult?.suggestions ? <Button onClick={() => setRuleConfigText(JSON.stringify(ruleResult.suggestions, null, 2))}>加入对应规则</Button> : null}</Space></div><details className="novel-fetch-advanced-rule"><summary>高级 JSON（与 V78 字段完全兼容）</summary><Input.TextArea style={{ marginTop: 12 }} rows={12} value={ruleConfigText} onChange={event => setRuleConfigText(event.target.value)} className="mono" placeholder="当前规则配置 JSON" /></details><Input.TextArea style={{ marginTop: 12 }} rows={9} value={ruleText} onChange={event => setRuleText(event.target.value)} placeholder="输入需要处理的文本或整批任务共同规则" /></div>{ruleResult ? <div className="novel-fetch-compare"><Tabs items={[{ key: 'result', label: '处理后文本', children: <Input.TextArea rows={14} readOnly value={ruleResult.result || ''} /> }, { key: 'trace', label: '阶段追踪', children: <Collapse items={(ruleResult.trace || []).map((item, index) => ({ key: String(index), label: item.title || `阶段 ${index + 1}`, children: JSON.stringify(item) }))} /> }, { key: 'suggestions', label: '规则建议', children: <Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }}>{JSON.stringify(ruleResult.suggestions || {}, null, 2)}</Typography.Paragraph> }]} /></div> : null}</Space>,
    submit: <Space direction="vertical" size={12} style={{ width: '100%' }}><div className="novel-fetch-panel"><Typography.Title level={5}>网站提交</Typography.Title><Alert type="info" showIcon message="仅在确认“开始自动提交”后执行外部提交。小说获取和批量工厂共用同一份视频管理系统登录会话。" /><Space wrap style={{ marginTop: 12 }}><Button loading={webSubmitLoading} onClick={() => handleWebSubmit(checkWebSubmitEnvironment, '环境检测完成')}>环境检测</Button><Button loading={webSubmitLoading} onClick={() => handleWebSubmit(syncWebSubmitConfigs, '已同步批量后台配置')}>同步批量后台配置</Button><Button loading={webSubmitLoading} onClick={() => handleWebSubmit(syncWebSubmitConfigs, '已同步组织归属')}>同步组织归属</Button><Button loading={webSubmitLoading} onClick={() => handleWebSubmit(syncWebSubmitStyles, '同步网站风格完成')}>同步网站风格</Button><Button disabled={!selectedTaskKeys.length} loading={webSubmitLoading} onClick={async () => { const result = await handleWebSubmit(() => previewWebSubmit({ taskIds: selectedTaskKeys }), '提交预览已生成'); if (result) setWebSubmitPlan(result.plan); }}>提交预览</Button><Button type="primary" disabled={!selectedTaskKeys.length} loading={webSubmitLoading} onClick={() => Modal.confirm({ title: '确认开始自动提交', onOk: async () => { const result = await handleWebSubmit(() => startWebSubmit({ taskIds: selectedTaskKeys }), '自动提交完成'); if (result) { setWebSubmitResults(result.results || []); setWebSubmitHistory(current => [...(result.results || []), ...current]); } } })}>开始自动提交</Button></Space><div className="novel-fetch-submit-settings"><Form layout="inline"><Form.Item label="组织归属"><Select style={{ width: 220 }} value={webSubmitConfig?.organization || undefined} placeholder="跟随网站配置档" options={(webSubmitConfig?.organization_catalog || []).map(item => ({ value: String(item.id || item.value), label: item.name || item.label || item.id }))} onChange={value => setWebSubmitConfig(current => ({ ...(current || {}), organization: value }))} /></Form.Item><Form.Item label="最少字数"><InputNumber min={0} value={webSubmitConfig?.min_text_chars ?? 0} onChange={value => setWebSubmitConfig(current => ({ ...(current || {}), min_text_chars: value || 0 }))} /></Form.Item><Form.Item label="失败重试"><InputNumber min={0} max={10} value={webSubmitConfig?.retry_times ?? 1} onChange={value => setWebSubmitConfig(current => ({ ...(current || {}), retry_times: value || 0 }))} /></Form.Item><Form.Item><Checkbox checked={webSubmitConfig?.skip_submitted !== false} onChange={event => setWebSubmitConfig(current => ({ ...(current || {}), skip_submitted: event.target.checked }))}>跳过已提交版本</Checkbox></Form.Item><Button onClick={() => handleWebSubmit(() => saveWebSubmitConfig(webSubmitConfig), '提交配置已保存')}>保存提交配置</Button></Form></div><Typography.Paragraph type="secondary" style={{ marginTop: 12 }}>已选任务：{selectedTaskKeys.length}；配置版本：{(webSubmitConfig?.submit_versions || webSubmitConfig?.submitVersions || []).join('、') || '默认'}；可用组织：{webSubmitConfig?.organization_catalog?.length || 0}</Typography.Paragraph>{webSubmitPlan ? <Alert type="success" message="提交预览已生成" description={`提交组：${webSubmitPlan.groups?.length || 0}，跳过：${webSubmitPlan.skipped?.length || 0}`} /> : null}</div><Table className="novel-fetch-panel" rowKey={row => `${row.bookId}-${row.version}`} dataSource={webSubmitResults} locale={{ emptyText: '暂无本次提交结果' }} columns={[{ title: '任务', dataIndex: 'bookId' }, { title: '版本', dataIndex: 'version' }, { title: '状态', dataIndex: 'status' }, { title: '失败重试', render: (_, row) => row.status === 'error' ? <Button size="small" onClick={() => startWebSubmit({ taskIds: [row.bookId], versions: [row.version] })}>失败重试</Button> : '-' }]} /><div className="novel-fetch-panel"><Typography.Title level={5}>提交历史</Typography.Title><Button size="small" loading={webSubmitLoading} onClick={loadWebSubmit}>刷新历史</Button><Table size="small" rowKey={(row, index) => `${row.book_id || row.bookId}-${row.version}-${index}`} dataSource={webSubmitHistory} pagination={{ pageSize: 10 }} columns={[{ title: '书名', dataIndex: 'book_name' }, { title: '版本', dataIndex: 'version' }, { title: '状态', dataIndex: 'status' }, { title: '结果', dataIndex: 'result' }, { title: '时间', dataIndex: 'time' }]} /></div></Space>,
    logs: <div className="novel-fetch-panel"><Typography.Title level={5}>操作日志</Typography.Title><Select value={selectedTask || undefined} style={{ width: 320, marginBottom: 12 }} placeholder="选择任务查看日志" options={tasks.map(task => ({ value: task.bookId, label: task.bookName || task.bookId }))} onChange={selectTask} />{selectedTaskData ? <Collapse items={(selectedTaskData.logs || []).map((log, index) => ({ key: String(index), label: `${log.time || ''} ${log.event || '操作日志'}`, children: <Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }}>{log.detail || JSON.stringify(log)}</Typography.Paragraph> }))} /> : <Table size="small" pagination={{ pageSize: 20 }} rowKey={(row, index) => row.id || index} dataSource={globalLogs} locale={{ emptyText: '暂无操作记录' }} columns={[{ title: '时间', dataIndex: 'time', render: (value, row) => value || row.createdAt || row.created_at }, { title: '事件', dataIndex: 'event' }, { title: '详情', render: (_, row) => String(row.detail || row.message || '').slice(0, 120) }]} />}</div>
  };

  // 公网 V78 的知识库页不是单纯的表格：左侧负责库类型/条目操作，右侧负责编辑当前条目。
  // 覆盖旧兼容面板，避免两个入口继续产生不同的数据行为。
  panels.knowledge = knowledgePanel;
  panels.submit = submissionPanel;

  const activePanel = activeTab === 'process'
    ? panels.process
    : activeTab === 'knowledge' || activeTab === 'rules'
      ? <NovelFetchSubpageShell title="知识库" description="维护 AI 参考资料、爆款开头和自动处理规则" onBack={() => changeTab('process')} actions={<><Button icon={<Settings2 size={14} />} onClick={() => setKnowledgeSettingsOpen(true)}>知识库设置</Button><Button icon={<RefreshCw size={14} />} loading={knowledgeLoading} onClick={loadKnowledge}>刷新</Button><Button type="primary" onClick={addKnowledgeLibraryItem}>新增资料</Button></>}><Tabs className="novel-fetch-section-tabs" activeKey={knowledgeView} onChange={setKnowledgeView} items={[{ key: 'library', label: '资料库', children: panels.knowledge }, { key: 'opening', label: '爆款开头拆解', children: openingPanel }, { key: 'rules', label: '处理规则', children: panels.rules }]} /></NovelFetchSubpageShell>
      : activeTab === 'tasks'
        ? <NovelFetchSubpageShell className="novel-fetch-task-subpage" title="任务" description="查看当前任务、历史批次和每本书的处理详情" onBack={() => changeTab('process')} actions={<Button icon={<RefreshCw size={14} />} loading={tasksLoading} onClick={() => { loadTasks(); loadCurrentBatch(); loadHistoryBatches(); }}>刷新任务</Button>}>{panels.tasks}</NovelFetchSubpageShell>
      : activeTab === 'config'
        ? <NovelFetchSubpageShell title="配置" description="统一管理抓取、处理、AI 和版本配置" onBack={() => changeTab('process')} actions={<><Tag color={configDirty ? 'gold' : 'green'}>{configDirty ? '未保存' : '已同步'}</Tag><Button onClick={() => testWorkshopAi('rewrite').then(() => message.success('测试接口成功')).catch(error => message.error(error.message || '测试失败'))}>测试文本模型</Button><Button type="primary" loading={configLoading} onClick={saveConfig}>保存配置</Button></>}>{panels.config}</NovelFetchSubpageShell>
      : activeTab === 'submit' || activeTab === 'logs'
        ? <NovelFetchSubpageShell title="提交与记录" description="预览确认后提交，并集中查看结果、历史和操作日志" onBack={() => changeTab('process')} actions={<Button icon={<RefreshCw size={14} />} loading={webSubmitLoading} onClick={() => { loadWebSubmit(); loadGlobalLogs(); }}>刷新状态</Button>}>{panels.submit}</NovelFetchSubpageShell>
    : <div className="novel-fetch-workbench-subpage">
        <div className="novel-fetch-workbench-subpage-header">
          <Button onClick={() => changeTab('process')}>返回处理</Button>
          <Typography.Title level={3} style={{ margin: 0 }}>{TABS.find(item => item.key === activeTab)?.label || '处理'}</Typography.Title>
        </div>
        {panels[activeTab] || panels.process}
      </div>;
  return <Space className="novel-fetch-page novel-fetch-workbench" direction="vertical" size={0} style={{ width: '100%' }}>
    {['dev', 'manager'].includes(accountRole) && accountOptions.length > 1 ? <div className="novel-fetch-account-scope">
      <div><Typography.Text type="secondary">当前查看账号</Typography.Text><strong>{accountOptions.find(item => item.value === accountScope)?.label || accountScope}</strong></div>
      <Select value={accountScope || undefined} options={accountOptions} onChange={changeAccountScope} popupMatchSelectWidth={280} aria-label="切换小说任务账号" />
      <Typography.Text type="secondary">{accountRole === 'dev' ? '开发者可查看全部账号' : '仅显示你的直属组员'}</Typography.Text>
    </div> : null}
    {activePanel}
    <TaskDetailDrawer taskDetail={taskDetail} onClose={() => setTaskDetail(null)} onSaved={loadTasks} />
  </Space>;
}

export default NovelFetchPage;
