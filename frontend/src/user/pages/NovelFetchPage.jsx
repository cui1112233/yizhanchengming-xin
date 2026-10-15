import { Alert, Button, Collapse, Descriptions, Drawer, Form, Input, InputNumber, Modal, Select, Space, Table, Tabs, Tag, Typography, message } from 'antd';
import { Download, Eye, RefreshCw } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import {
  checkWebSubmitEnvironment, getWebSubmitConfig, getWebSubmitHistory, previewWebSubmit,
  saveWebSubmitConfig, startWebSubmit, syncWebSubmitConfigs, syncWebSubmitStyles, uploadLogin, getUploadSession
} from '../../shared/api/novelFetch';
import {
  analyzeWorkshopOpening, deleteWorkshopKnowledge, deleteWorkshopTasks, fetchWorkshopOriginal,
  generateWorkshopAi, getWorkshopConfig, getWorkshopKnowledge, getWorkshopKnowledgeSummary,
  getWorkshopPlatforms, getWorkshopTask, getWorkshopJobs, listWorkshopTasks, normalizeWorkshopOpening, optimizeWorkshopKnowledge,
  getWorkshopLogs, listWorkshopSchedules, deleteWorkshopSchedule, saveWorkshopSchedule,
  previewWorkshopRules, restoreWorkshopOriginal, retryWorkshopTasks, saveWorkshopConfig,
  updateWorkshopAiCount, applyWorkshopRules,
  saveWorkshopKnowledge, saveWorkshopOpening, startWorkshopProcess, suggestWorkshopRules, testWorkshopAi, reprocessWorkshopSensitive
} from '../../shared/api/novelFetchWorkshop';
import { readNovelFetchProcessSettings, writeNovelFetchProcessSettings } from './novel-fetch-process-settings';
import './novel-fetch.css';

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
const REWRITE_METHODS = [
  { value: 'auto', label: '自动轮换' },
  { value: 'high_imitation', label: '高仿改写' },
  { value: 'opening_instruction', label: '爆款开头' },
  { value: 'instruction', label: '自定义规则' }
];
const PLATFORMS = [
  { value: 1, label: '黑岩付费' }, { value: 2, label: '番茄付费' }, { value: 3, label: '七猫付费' },
  { value: 4, label: '点众付费' }, { value: 7, label: '番茄免费' }, { value: 15, label: '知乎付费' },
  { value: 20, label: '掌阅付费' }, { value: 26, label: '卓越付费' }, { value: 29, label: '九州书城' }, { value: 31, label: '掌文付费' }
];
const KNOWLEDGE_TABS = [
  ['high_imitation', '高仿库'], ['opening_phrases', '爆款开头'], ['rewrite_templates', '改文模板'],
  ['layout_rules', '排版规则'], ['symbol_rules', '符号规则'], ['chapter_rules', '章节规则']
];

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
function downloadText(name, content) {
  const url = URL.createObjectURL(new Blob([content], { type: 'text/plain;charset=utf-8' }));
  const link = document.createElement('a');
  link.href = url;
  link.download = name;
  link.click();
  URL.revokeObjectURL(url);
}

export function NovelFetchPage() {
  return <LegacyNovelFetchPage />;
}

export function LegacyNovelFetchPage() {
  const [processForm] = Form.useForm();
  const savedProcessSettings = useMemo(() => readNovelFetchProcessSettings(), []);
  const [configForm] = Form.useForm();
  const [knowledgeForm] = Form.useForm();
  const [activeTab, setActiveTab] = useState(initialTab);
  const [processRows, setProcessRows] = useState([]);
  const [platformOptions, setPlatformOptions] = useState(PLATFORMS);
  const [processJob, setProcessJob] = useState(null);
  const [schedules, setSchedules] = useState([]);
  const [scheduleModalOpen, setScheduleModalOpen] = useState(false);
  const [scheduleTime, setScheduleTime] = useState('');
  const [globalLogs, setGlobalLogs] = useState([]);
  const [processing, setProcessing] = useState(false);
  const [tasks, setTasks] = useState([]);
  const [tasksLoading, setTasksLoading] = useState(false);
  const [selectedTaskKeys, setSelectedTaskKeys] = useState([]);
  const [selectedTaskData, setSelectedTaskData] = useState(null);
  const [taskDetail, setTaskDetail] = useState(null);
  const [ruleText, setRuleText] = useState('');
  const [ruleScope, setRuleScope] = useState('original');
  const [ruleResult, setRuleResult] = useState(null);
  const [ruleLoading, setRuleLoading] = useState(false);
  const [knowledgeKind, setKnowledgeKind] = useState('high_imitation');
  const [knowledge, setKnowledge] = useState({ items: [] });
  const [knowledgeSummary, setKnowledgeSummary] = useState({});
  const [knowledgeLoading, setKnowledgeLoading] = useState(false);
  const [knowledgeModal, setKnowledgeModal] = useState(false);
  const [knowledgeSearch, setKnowledgeSearch] = useState('');
  const [openingText, setOpeningText] = useState('');
  const [openingItems, setOpeningItems] = useState([]);
  const [configLoading, setConfigLoading] = useState(false);
  const [webSubmitConfig, setWebSubmitConfig] = useState(null);
  const [webSubmitPlan, setWebSubmitPlan] = useState(null);
  const [webSubmitResults, setWebSubmitResults] = useState([]);
  const [webSubmitHistory, setWebSubmitHistory] = useState([]);
  const [webSubmitLoading, setWebSubmitLoading] = useState(false);
  const [uploadSession, setUploadSession] = useState(null);
  const [loginModalOpen, setLoginModalOpen] = useState(false);
  const [loginUser, setLoginUser] = useState('');
  const [loginPassword, setLoginPassword] = useState('');
  const [versionConfigOpen, setVersionConfigOpen] = useState(false);
  const [versionPrompts, setVersionPrompts] = useState({ rewrite: '', processingRule: '', knowledgeUsage: '' });
  const [uploadProfiles, setUploadProfiles] = useState([]);
  const [selectedUploadProfile, setSelectedUploadProfile] = useState('');
  const [profileBindings, setProfileBindings] = useState({});
  const [taskDate, setTaskDate] = useState(() => new Date().toISOString().slice(0, 10));
  const [batchAiCount, setBatchAiCount] = useState(1);
  const [ruleApplyScope, setRuleApplyScope] = useState('both');
  const [selectedVersions, setSelectedVersions] = useState(() => savedProcessSettings.selectedVersions);
  const [rewriteMethods, setRewriteMethods] = useState(() => savedProcessSettings.rewriteMethods);
  const [sensitiveAiProcessEnabled, setSensitiveAiProcessEnabled] = useState(() => savedProcessSettings.sensitiveAiProcessEnabled);
  const [textModel, setTextModel] = useState(() => savedProcessSettings.textModel);

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
    if (!taskDate) return tasks;
    return tasks.filter(task => String(task.createdAt || task.created_at || task.updatedAt || '').slice(0, 10) === taskDate);
  }, [tasks, taskDate]);
  const processSummary = useMemo(() => {
    const rows = visibleTasks;
    return {
      total: rows.length,
      originalDone: rows.filter(row => String(row.originalStatus || row.original_status).includes('done')).length,
      processing: rows.filter(row => /processing|running|pending/.test(String(row.status || row.aiStatus || row.ai_status))).length,
      aiDone: rows.filter(row => String(row.aiStatus || row.ai_status).includes('done') || Number(row.aiGeneratedCount || row.ai_generated_count) > 0).length,
      failed: rows.filter(row => /fail|error/.test(String(row.status || row.aiStatus || row.ai_status))).length
    };
  }, [visibleTasks]);

  async function loadTasks() {
    setTasksLoading(true);
    try {
      const [taskData, jobData] = await Promise.all([listWorkshopTasks(), getWorkshopJobs(10)]);
      setTasks(taskData.tasks || []);
      if (jobData.latest && Object.keys(jobData.latest).length) setProcessJob(jobData.latest);
    } catch (error) { message.error(error.message || '读取任务失败'); } finally { setTasksLoading(false); }
  }
  async function selectTask(bookId) {
    if (!bookId) return;
    try { setSelectedTaskData(await getWorkshopTask(bookId)); } catch (error) { message.error(error.message || '读取任务详情失败'); }
  }
  async function loadKnowledge() {
    setKnowledgeLoading(true);
    try {
      const [items, summary] = await Promise.all([getWorkshopKnowledge(knowledgeKind), getWorkshopKnowledgeSummary()]);
      setKnowledge(items || { items: [] });
      setKnowledgeSummary(summary || {});
    } catch (error) { message.error(error.message || '读取知识库失败'); } finally { setKnowledgeLoading(false); }
  }
  async function loadConfig() {
    setConfigLoading(true);
    try {
      const data = await getWorkshopConfig();
      const config = data.appConfig || data.config || {};
      configForm.setFieldsValue(config);
      const configuredPlatforms = Array.isArray(data.platforms) ? data.platforms : Array.isArray(config.platforms) ? config.platforms : [];
      if (configuredPlatforms.length) setPlatformOptions(configuredPlatforms.map(item => ({ value: Number(item.id ?? item.platform_id ?? item.value), label: item.name || item.platform_name || item.label || String(item.id ?? item.platform_id ?? item.value) })));
    } catch (error) { message.error(error.message || '读取配置失败'); } finally { setConfigLoading(false); }
  }
  async function loadWebSubmit() {
    try {
      const [configData, historyData] = await Promise.all([getWebSubmitConfig(), getWebSubmitHistory()]);
      setWebSubmitConfig(configData.settings || configData.config || {});
      setWebSubmitHistory(historyData.records || historyData.items || []);
    } catch (error) { message.error(error.message || '读取提交配置失败'); }
  }
  async function loadUploadSession() {
    try { setUploadSession(await getUploadSession()); } catch { setUploadSession({ status: 'offline' }); }
  }
  async function loadSchedules() {
    try { setSchedules((await listWorkshopSchedules()).schedules || []); } catch (error) { message.error(error.message || '读取定时任务失败'); }
  }
  async function loadGlobalLogs() {
    try { setGlobalLogs((await getWorkshopLogs()).logs || []); } catch (error) { message.error(error.message || '读取操作记录失败'); }
  }

  useEffect(() => {
    loadTasks();
    loadUploadSession();
    loadSchedules();
    loadGlobalLogs();
    getWorkshopPlatforms().then(data => {
      const items = data.platforms || data.items || data;
      if (Array.isArray(items) && items.length) setPlatformOptions(items.map(item => ({ value: Number(item.id ?? item.platform_id ?? item.value), label: item.name || item.platform_name || item.label || String(item.id ?? item.platform_id ?? item.value) })));
    }).catch(() => {});
  }, []);
  useEffect(() => {
    if (activeTab === 'config') loadConfig();
    if (activeTab === 'knowledge') loadKnowledge();
    if (activeTab === 'submit') loadWebSubmit();
  }, [activeTab, knowledgeKind]);

  function changeTab(tab) {
    setActiveTab(tab);
    const url = new URL(window.location.href);
    url.searchParams.set('tab', tab);
    window.history.replaceState(null, '', `${url.pathname}${url.search}`);
  }
  async function handleProcess() {
    const values = await processForm.validateFields();
    const inputText = String(values.bookIds || '').trim();
    const bookIds = parseIds(inputText);
    if (!bookIds.length) return message.warning('请填写至少一个书籍 ID');
    setProcessing(true);
    setProcessRows(bookIds.map(bookId => ({ bookId, status: 'loading' })));
    try {
      // The Go provider resolver reads the account's selected model. Persist the
      // visible choice before starting so a local draft cannot silently use a
      // different enabled model.
      if (textModel) await saveWorkshopConfig({ text_model_id: textModel });
      const data = await startWorkshopProcess({
        inputText,
        platform_id: String(values.platform),
        maxTxt: values.maxTxt,
        text_model_id: textModel,
        selected_versions: selectedVersions,
        rewrite_methods: rewriteMethods,
        sensitive_ai_enabled: sensitiveAiProcessEnabled
      });
      setProcessJob(data);
      const results = data.result?.tasks || data.tasks || [];
      setProcessRows(results.length
        ? results.map(row => ({ ...row, bookId: row.bookId || row.book_id, status: row.status === 'ok' || row.originalStatus === 'done' ? 'done' : (row.status || 'failed') }))
        : bookIds.map(bookId => ({ bookId, status: data.status === 'done' ? 'done' : data.status || 'failed', error: data.error || '' })));
      await loadTasks();
    } catch (error) {
      setProcessRows(bookIds.map(bookId => ({ bookId, status: 'failed', error: error.message })));
      message.error(error.message || '处理失败');
    } finally { setProcessing(false); }
  }
  async function handleScheduleProcess() {
    const values = await processForm.validateFields();
    const bookIds = parseIds(values.bookIds);
    if (!bookIds.length) return message.warning('请先填写待处理清单');
    if (!scheduleTime) return message.warning('请选择定时开始时间');
    try {
      await saveWorkshopSchedule({
        run_at: scheduleTime,
        input_text: String(values.bookIds || ''),
        platform_id: String(values.platform),
        max_txt: values.maxTxt,
        selected_versions: selectedVersions,
        rewrite_methods: rewriteMethods,
        sensitive_ai_enabled: sensitiveAiProcessEnabled
      });
      setScheduleModalOpen(false);
      await loadSchedules();
      message.success('定时任务已保存');
    } catch (error) { message.error(error.message || '保存定时任务失败'); }
  }
  async function handleSensitiveRepair() {
    if (!selectedTaskKeys.length) return message.warning('请先在任务页选择任务');
    try { await reprocessWorkshopSensitive(selectedTaskKeys, true); await loadTasks(); message.success('敏感词 AI 修复已加入处理'); } catch (error) { message.error(error.message || '敏感词 AI 修复失败'); }
  }
  async function handleGenerateAi() {
    if (!selectedTask) return message.warning('请先选择一个任务');
    try { await generateWorkshopAi(selectedTask, 1); await selectTask(selectedTask); await loadTasks(); message.success('已生成 AI 版本'); } catch (error) { message.error(error.message || '生成 AI 版本失败'); }
  }
  async function handleRetryTasks() {
    try { await retryWorkshopTasks(selectedTaskKeys); await loadTasks(); message.success('已加入重试'); } catch (error) { message.error(error.message || '批量重试失败'); }
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
    if (!ruleText.trim()) return message.warning('请填写待处理文本');
    setRuleLoading(true);
    try {
      const result = await suggestWorkshopRules(ruleText, 'layout', '优化排版与敏感词处理');
      setRuleResult(current => ({ ...current, suggestions: result }));
    } catch (error) { message.error(error.message || '生成规则建议失败'); } finally { setRuleLoading(false); }
  }
  async function saveRuleConfiguration() {
    if (!ruleText.trim()) return message.warning('请先填写规则内容');
    try {
      await saveWorkshopConfig({ processing_rule_prompt: ruleText, processing_rule_scope: ruleScope });
      message.success('处理规则已保存');
    } catch (error) { message.error(error.message || '保存处理规则失败'); }
  }
  async function saveKnowledge() {
    const item = await knowledgeForm.validateFields();
    try { await saveWorkshopKnowledge(knowledgeKind, item); setKnowledgeModal(false); await loadKnowledge(); message.success('知识条目已保存'); } catch (error) { message.error(error.message || '保存知识条目失败'); }
  }
  async function handleOpeningAnalyze() {
    if (!openingText.trim()) return message.warning('请填写开头文本');
    try { setOpeningItems((await analyzeWorkshopOpening(openingText)).items || []); } catch (error) { message.error(error.message || '分析失败'); }
  }
  async function handleWebSubmit(action, success) {
    setWebSubmitLoading(true);
    try { const result = await action(); if (success) message.success(success); return result; } catch (error) { message.error(error.message || '网站提交操作失败'); return null; } finally { setWebSubmitLoading(false); }
  }
  async function handleUploadLogin() {
    if (!loginUser.trim() || !loginPassword) return message.warning('请输入视频管理系统账号和密码');
    setWebSubmitLoading(true);
    try {
      await uploadLogin({ username: loginUser.trim(), password: loginPassword });
      setLoginPassword('');
      setLoginModalOpen(false);
      await loadUploadSession();
      message.success('视频管理系统登录验证成功');
    } catch (error) { message.error(error.message || '视频管理系统登录失败'); } finally { setWebSubmitLoading(false); }
  }
  async function updateSelectedAiCount() {
    if (!selectedTaskKeys.length) return message.warning('请先选择任务');
    try { await updateWorkshopAiCount(selectedTaskKeys, batchAiCount); await loadTasks(); message.success(`已调整为 ${batchAiCount} 个 AI 版本`); } catch (error) { message.error(error.message || '调整 AI 数量失败'); }
  }
  async function retryFailedTasks() {
    const failed = tasks.filter(task => String(task.status || task.aiStatus || '').includes('fail')).map(task => task.bookId);
    if (!failed.length) return message.info('当前没有失败任务');
    try { await retryWorkshopTasks(failed); await loadTasks(); message.success('失败任务已加入重试'); } catch (error) { message.error(error.message || '重试失败任务失败'); }
  }
  function selectAllTasks() { setSelectedTaskKeys(visibleTasks.map(task => task.bookId || task.book_id).filter(Boolean)); }
  function shiftTaskDate(days) {
    const base = taskDate ? new Date(`${taskDate}T00:00:00`) : new Date();
    base.setDate(base.getDate() + days);
    setTaskDate(base.toISOString().slice(0, 10));
  }
  async function applySelectedRules() {
    if (!selectedTaskKeys.length) return message.warning('请先选择任务');
    try { await applyWorkshopRules(selectedTaskKeys, ruleApplyScope, sensitiveAiProcessEnabled); await loadTasks(); message.success('已完成规则处理'); } catch (error) { message.error(error.message || '规则处理失败'); }
  }
  async function submitVisibleTasks(all = false) {
    const source = all ? tasks : visibleTasks;
    const ids = source.map(task => task.bookId || task.book_id).filter(Boolean);
    if (!ids.length) return message.info('当前没有可提交任务');
    const result = await handleWebSubmit(() => startWebSubmit({ taskIds: ids }), all ? '已提交全部任务' : '已提交当前任务');
    if (result) setWebSubmitResults(result.results || []);
  }
  async function deleteTasksByFilter(filter, emptyMessage) {
    const ids = tasks.filter(filter).map(task => task.bookId || task.book_id).filter(Boolean);
    if (!ids.length) return message.info(emptyMessage);
    Modal.confirm({ title: '确认删除任务', content: '删除后无法恢复，确认继续？', onOk: async () => { await deleteWorkshopTasks(ids); setSelectedTaskKeys(current => current.filter(id => !ids.includes(id))); await loadTasks(); message.success('任务已删除'); } });
  }
  function transferToBatchFactory() { window.location.assign('/batch-factory'); }
  async function saveVersionConfig() {
    try {
      await saveWorkshopConfig({ rewrite: { prompt: versionPrompts.rewrite, processing_rule_prompt: versionPrompts.processingRule, ai_slot_methods: rewriteMethods }, knowledge: { usage_prompt: versionPrompts.knowledgeUsage }, selected_versions: selectedVersions, sensitive_ai_enabled: sensitiveAiProcessEnabled });
      await saveWebSubmitConfig({ selected_profile: selectedUploadProfile, profile_bindings: profileBindings, submit_versions: selectedVersions });
      setVersionConfigOpen(false);
      message.success('版本配置已保存');
    } catch (error) { message.error(error.message || '保存版本配置失败'); }
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
      setProfileBindings(config.profile_bindings && typeof config.profile_bindings === 'object' ? config.profile_bindings : {});
    } catch { /* profile configuration is optional until the upload backend is paired */ }
    setVersionConfigOpen(true);
  }

  const panels = {
    process: <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Form className="novel-fetch-public-workspace" form={processForm} layout="vertical" initialValues={{ platform: savedProcessSettings.platform, maxTxt: savedProcessSettings.maxTxt }} onValuesChange={() => {
        writeNovelFetchProcessSettings({
          ...savedProcessSettings,
          ...processForm.getFieldsValue(),
          selectedVersions,
          rewriteMethods,
          sensitiveAiProcessEnabled,
          textModel
        });
      }}>
        <div className="novel-fetch-public-heading"><div><Typography.Title level={4}>批量输入</Typography.Title></div><Typography.Text type="secondary">平台：{platformOptions.length}</Typography.Text></div>
        <div className="novel-fetch-public-grid novel-fetch-public-grid--three">
          <Form.Item name="platform" label="平台"><Select options={platformOptions} /></Form.Item>
          <label className="novel-fetch-field"><span>文本模型</span><Select value={textModel} onChange={setTextModel} options={[{ value: '[anti-2]gemini-3.8-flash', label: '[anti-2]gemini-3.8-flash' }, { value: 'gpt-4o-mini', label: 'gpt-4o-mini' }]} /><small>执行时使用此模型；修改后作为后续批次默认值</small></label>
        </div>
        <Form.Item name="bookIds" label="待处理清单（一行一本书）" rules={[{ required: true, message: '请输入至少一个书籍 ID' }]}><Input.TextArea className="novel-fetch-public-input" rows={5} placeholder={'每一行：书籍 ID、书名、推荐理由、男女频、标签、评级。\n示例：123456789\t书名\t打脸逆袭、复仇反转\t女频\t现代世情、爽文\tS'} /></Form.Item>
        <section className="novel-fetch-public-section"><Typography.Title level={5}>本次处理设置</Typography.Title><Typography.Text type="secondary">AI 文案（选择本次需要生成的版本）</Typography.Text><div className="novel-fetch-version-grid">{AI_VERSIONS.map(version => <label key={version} className={`novel-fetch-version-option ${selectedVersions.includes(version) ? 'is-selected' : ''}`}><input type="checkbox" checked={selectedVersions.includes(version)} onChange={event => setSelectedVersions(current => event.target.checked ? [...current, version] : current.filter(item => item !== version))} /><span>{AI_VERSION_LABELS[version]}</span></label>)}</div><div className="novel-fetch-public-grid novel-fetch-public-grid--five">{AI_VERSIONS.slice(1).map(version => <label className="novel-fetch-field" key={version}><span>{AI_VERSION_LABELS[version]} 改文方式</span><Select value={rewriteMethods[version]} onChange={value => setRewriteMethods(current => ({ ...current, [version]: value }))} disabled={!selectedVersions.includes(version)} options={REWRITE_METHODS} /></label>)}</div><div className="novel-fetch-public-hint">AI 文案处理优先方案：跟随“配置”中 AI1～AI5 的现有设置。</div></section>
        <section className="novel-fetch-public-section novel-fetch-sensitive"><Typography.Title level={5}>敏感处理</Typography.Title><label><input type="checkbox" checked={sensitiveAiProcessEnabled} onChange={event => setSensitiveAiProcessEnabled(event.target.checked)} /> 启用敏感词 AI 修复（关闭则直接替换）</label></section>
        <div className="novel-fetch-public-actions"><Button type="primary" loading={processing} onClick={handleProcess}>开始处理</Button><Button onClick={() => setScheduleModalOpen(true)}>开始定时</Button><Button onClick={() => { loadSchedules(); setScheduleModalOpen(true); }}>定时任务</Button><Button icon={<RefreshCw size={14} />} loading={tasksLoading} onClick={loadTasks}>刷新任务</Button><Button onClick={() => processForm.resetFields()}>清空</Button><Button onClick={openVersionConfig}>版本配置</Button></div>
      </Form>
      <Modal open={scheduleModalOpen} title="定时任务" okText="保存定时任务" cancelText="取消" onCancel={() => setScheduleModalOpen(false)} onOk={handleScheduleProcess}>
        <Typography.Paragraph type="secondary">保存后由 Go 后端统一调度；不会立即加入正在执行的队列。</Typography.Paragraph>
        <label className="novel-fetch-field"><span>开始时间</span><Input type="datetime-local" value={scheduleTime} onChange={event => setScheduleTime(event.target.value)} /></label>
        <Table size="small" pagination={false} rowKey="id" dataSource={schedules} locale={{ emptyText: '暂无定时任务' }} columns={[{ title: '开始时间', dataIndex: 'run_at' }, { title: '状态', dataIndex: 'status', render: value => value || '待执行' }, { title: '操作', render: (_, row) => <Button size="small" danger onClick={async () => { await deleteWorkshopSchedule(row.id); await loadSchedules(); }}>删除</Button> }]} />
      </Modal>
      <Modal open={versionConfigOpen} title="版本配置" onCancel={() => setVersionConfigOpen(false)} onOk={saveVersionConfig}><Alert type="info" showIcon message="版本配置会保存到当前工作台，切换页面后仍然保留。" /><Typography.Paragraph>版本配置：{selectedVersions.map(version => AI_VERSION_LABELS[version]).join('、') || '未选择'}</Typography.Paragraph><Space wrap><Button onClick={() => handleWebSubmit(syncWebSubmitConfigs, '已同步批量后台配置')}>同步批量后台配置</Button><Button onClick={() => handleWebSubmit(syncWebSubmitStyles, '已同步批量风格类型')}>同步批量风格类型</Button></Space><Form layout="vertical" style={{ marginTop: 12 }}><Form.Item label="默认上传配置档"><Select value={selectedUploadProfile || undefined} placeholder="选择默认配置档" options={uploadProfiles.map(profile => ({ value: String(profile.id || profile.name), label: profile.name || profile.id }))} onChange={setSelectedUploadProfile} /></Form.Item><Typography.Text type="secondary">每个版本可绑定独立配置档；未绑定时使用默认配置档。</Typography.Text>{selectedVersions.map(version => <Form.Item key={version} label={`${AI_VERSION_LABELS[version]} 配置档`}><Select allowClear value={profileBindings[version] || undefined} placeholder="跟随默认配置档" options={uploadProfiles.map(profile => ({ value: String(profile.id || profile.name), label: profile.name || profile.id }))} onChange={value => setProfileBindings(current => ({ ...current, [version]: value || '' }))} /></Form.Item>)}<Form.Item label="改文提示词"><Input.TextArea rows={3} value={versionPrompts.rewrite} onChange={event => setVersionPrompts(current => ({ ...current, rewrite: event.target.value }))} placeholder="本次处理的改文提示词" /></Form.Item><Form.Item label="处理规则提示词"><Input.TextArea rows={3} value={versionPrompts.processingRule} onChange={event => setVersionPrompts(current => ({ ...current, processingRule: event.target.value }))} placeholder="整本书共同遵守的规则" /></Form.Item><Form.Item label="知识库使用提示词"><Input.TextArea rows={3} value={versionPrompts.knowledgeUsage} onChange={event => setVersionPrompts(current => ({ ...current, knowledgeUsage: event.target.value }))} placeholder="知识库的使用要求" /></Form.Item></Form></Modal>
      <div className="novel-fetch-status-layout">
        <div className="novel-fetch-public-result"><div className="novel-fetch-public-result-title">处理状态</div><div className="novel-fetch-stat-grid">
          {[['总数量', processSummary.total], ['原文已解析', processSummary.originalDone], ['处理中', processSummary.processing], ['AI已完成', processSummary.aiDone], ['失败', processSummary.failed]].map(([label, value]) => <div className="novel-fetch-stat-card" key={label}><span>{label}</span><strong>{value}</strong></div>)}
        </div><div className="novel-fetch-progress-row"><span>当前任务</span><b>{processJob?.status || (processing ? 'running' : 'idle')}</b></div><pre>{processRows.length ? JSON.stringify(processRows, null, 2) : '暂无处理结果。提交任务后在这里查看状态。'}</pre></div>
        <div className="novel-fetch-public-result novel-fetch-live-log"><div className="novel-fetch-public-result-title">实时日志</div><div className={`novel-fetch-login-status ${uploadSession?.status === 'ready' ? 'is-ready' : 'is-offline'}`}><span>{uploadSession?.status === 'ready' ? '● 已登录视频管理系统' : '● 未登录视频管理系统'}</span><Button size="small" onClick={() => setLoginModalOpen(true)}>更换账号</Button></div><pre>{processJob?.error ? processJob.error : processJob?.status === 'done' ? '完整流程已完成。' : processing ? '正在执行：读取原文、分类、敏感处理和 AI 版本。' : '暂无正在处理的任务。'}</pre></div>
      </div>
      <Modal open={loginModalOpen} title="登录视频管理系统" confirmLoading={webSubmitLoading} okText="保存并验证" cancelText="取消" onCancel={() => setLoginModalOpen(false)} onOk={handleUploadLogin}><Form layout="vertical"><Form.Item label="账号"><Input value={loginUser} onChange={event => setLoginUser(event.target.value)} autoComplete="username" /></Form.Item><Form.Item label="密码"><Input.Password value={loginPassword} onChange={event => setLoginPassword(event.target.value)} autoComplete="current-password" /></Form.Item></Form></Modal>
    </Space>,
    tasks: <Space direction="vertical" size={12} style={{ width: '100%' }}>
      <div className="novel-fetch-panel"><Space wrap><Button icon={<RefreshCw size={14} />} loading={tasksLoading} onClick={loadTasks}>刷新任务</Button><Button disabled={!selectedTaskKeys.length} onClick={handleRetryTasks}>批量重试</Button><Button danger disabled={!selectedTaskKeys.length} onClick={handleDeleteTasks}>删除选中</Button><Button disabled={!selectedTask} onClick={handleGenerateAi}>生成 AI 改文</Button><Button disabled={!selectedTaskKeys.length} onClick={handleSensitiveRepair}>敏感词 AI 修复</Button></Space></div>
      <div className="novel-fetch-panel"><Space wrap><Typography.Text>查看日期</Typography.Text><Button onClick={() => shiftTaskDate(-1)}>前一天</Button><Input type="date" value={taskDate} onChange={event => setTaskDate(event.target.value)} /><Button onClick={() => setTaskDate(new Date().toISOString().slice(0, 10))}>今天</Button><Button onClick={() => shiftTaskDate(1)}>后一天</Button><Button onClick={() => setTaskDate('')}>默认视图</Button><Button onClick={selectAllTasks}>全选</Button><Button onClick={() => setSelectedTaskKeys([])}>清空</Button><Button onClick={retryFailedTasks}>重试失败</Button><Button onClick={transferToBatchFactory}>进入批量工厂</Button><Button onClick={() => submitVisibleTasks(false)}>提交网络</Button><Button onClick={() => submitVisibleTasks(true)}>提交全部</Button><Typography.Text>AI数量</Typography.Text><InputNumber min={1} max={20} value={batchAiCount} onChange={value => setBatchAiCount(value || 1)} /><Button onClick={updateSelectedAiCount}>调整选中</Button><Select value={ruleApplyScope} onChange={setRuleApplyScope} options={[{ value: 'both', label: '规则处理原文+AI' }, { value: 'original', label: '只处理原文' }, { value: 'ai', label: '只处理AI' }]} /><Button onClick={applySelectedRules}>规则处理选中</Button><Button danger onClick={handleDeleteTasks}>删除选中</Button><Button danger onClick={() => deleteTasksByFilter(task => /fail|error/.test(String(task.status || task.aiStatus || task.ai_status)), '当前没有失败任务')}>删除失败</Button><Button danger onClick={() => deleteTasksByFilter(() => true, '当前没有任务')}>删除全部</Button></Space></div>
      <Typography.Text type="secondary">已选 {selectedTaskKeys.length} 个任务；当前显示 {visibleTasks.length} 个</Typography.Text><Table className="novel-fetch-panel" rowKey={row => row.bookId || row.book_id} loading={tasksLoading} dataSource={visibleTasks} rowSelection={{ selectedRowKeys: selectedTaskKeys, onChange: setSelectedTaskKeys }} pagination={{ pageSize: 20 }} columns={[
        { title: '书名', dataIndex: 'bookName', render: (value, row) => value || row.book_name || row.bookId }, { title: '平台', render: (_, row) => row.platformName || row.platform_name }, { title: '风格', dataIndex: 'style', render: value => value || '待判定' }, { title: '男女频', dataIndex: 'gender', render: value => value || '待确认' }, { title: '原文状态', render: (_, row) => <Tag color={statusColor(row.originalStatus || row.original_status)}>{row.originalStatus || row.original_status || '未抓取'}</Tag> },
        { title: 'AI 状态', render: (_, row) => <Tag color={statusColor(row.aiStatus || row.ai_status)}>{row.aiGeneratedCount || row.ai_generated_count ? `${row.aiGeneratedCount || row.ai_generated_count} 版` : row.aiStatus || row.ai_status || '未生成'}</Tag> }, { title: '网站提交', render: (_, row) => row.site_submit_status || row.siteSubmitStatus || '未提交' }, { title: '状态', render: (_, row) => <Tag color={statusColor(row.status)}>{row.status || '待处理'}</Tag> }, { title: '操作', render: (_, row) => <Space><Button size="small" onClick={() => selectTask(row.bookId || row.book_id)}>选择</Button><Button size="small" icon={<Eye size={14} />} onClick={async () => setTaskDetail({ bookId: row.bookId || row.book_id, data: await getWorkshopTask(row.bookId || row.book_id) })}>详情</Button><Button size="small" onClick={() => fetchWorkshopOriginal(row.bookId || row.book_id, row.maxTxt).then(loadTasks)}>重新抓取</Button><Button size="small" onClick={() => restoreWorkshopOriginal(row.bookId || row.book_id).then(loadTasks)}>恢复原文</Button></Space> }
      ]} />
    </Space>,
    config: <Form className="novel-fetch-panel" form={configForm} layout="vertical" style={{ maxWidth: 980 }}><Typography.Title level={5}>配置</Typography.Title><div className="novel-fetch-config-grid"><Form.Item name="base_url" label="AI 基础接口"><Input /></Form.Item><Form.Item name="api_key" label="API Key"><Input.Password /></Form.Item><Form.Item name="model" label="模型"><Input /></Form.Item><Form.Item name="fetch_endpoint" label="原文抓取接口"><Input placeholder="https://txt.121w.com/api.php" /></Form.Item><Form.Item name="default_max_txt" label="默认抓取字数"><InputNumber min={100} max={100000} /></Form.Item><Form.Item name="default_ai_count" label="默认 AI 数量"><InputNumber min={1} max={20} /></Form.Item><Form.Item name="fetch_concurrency" label="抓取并发"><InputNumber min={1} max={20} /></Form.Item><Form.Item name="fetch_retries" label="抓取重试"><InputNumber min={0} max={10} /></Form.Item><Form.Item name="fetch_timeout" label="抓取超时（秒）"><InputNumber min={1} max={300} /></Form.Item><Form.Item name="process_line_count" label="默认处理行数"><InputNumber min={1} max={100} /></Form.Item><Form.Item name="anchor_line_count" label="锚点行数"><InputNumber min={1} max={100} /></Form.Item><Form.Item name="rewrite_temp" label="改文温度"><InputNumber min={0} max={2} step={0.01} /></Form.Item><Form.Item name="method_sequence" label="改文方式顺序"><Input placeholder="high_imitation,opening_instruction,instruction" /></Form.Item><Form.Item name="opening_phrase_mode" label="爆款开头模式"><Select options={[{ value: 'auto', label: '自动' }, { value: 'off', label: '关闭' }]} /></Form.Item><Form.Item name="high_imitation_mode" label="高仿改写模式"><Select options={[{ value: 'auto', label: '自动' }, { value: 'off', label: '关闭' }]} /></Form.Item></div><Space wrap><Form.Item name="workflow_auto_classify" valuePropName="checked" noStyle><label><input type="checkbox" /> 自动补齐风格/男女频</label></Form.Item><Form.Item name="workflow_auto_fetch" valuePropName="checked" noStyle><label><input type="checkbox" /> 自动抓取原文</label></Form.Item><Form.Item name="workflow_auto_rewrite" valuePropName="checked" noStyle><label><input type="checkbox" /> 抓取后自动生成 AI 文案</label></Form.Item><Form.Item name="fetch_auto_detect_platform" valuePropName="checked" noStyle><label><input type="checkbox" /> 自动识别平台</label></Form.Item><Form.Item name="sensitive_ai_enabled" valuePropName="checked" noStyle><label><input type="checkbox" /> 默认启用敏感词 AI 修复</label></Form.Item></Space><Form.Item name="rewrite_prompt" label="改文提示词"><Input.TextArea rows={5} /></Form.Item><Form.Item name="processing_rule_prompt" label="处理规则提示词"><Input.TextArea rows={5} /></Form.Item><Space><Button type="primary" loading={configLoading} onClick={async () => { await saveWorkshopConfig(configForm.getFieldsValue()); message.success('配置已保存'); }}>保存配置</Button><Button onClick={() => testWorkshopAi('rewrite').then(() => message.success('测试接口成功')).catch(error => message.error(error.message || '测试失败'))}>测试接口</Button></Space></Form>,
    knowledge: <Space direction="vertical" size={12} style={{ width: '100%' }}><div className="novel-fetch-panel"><Tabs activeKey={knowledgeKind} onChange={setKnowledgeKind} items={KNOWLEDGE_TABS.map(([key, label]) => ({ key, label: `${label}（${knowledgeSummary[key] || 0}）` }))} /><Space wrap><Input allowClear value={knowledgeSearch} onChange={event => setKnowledgeSearch(event.target.value)} placeholder="搜索标题或内容" style={{ width: 260 }} /><Button onClick={() => { knowledgeForm.resetFields(); setKnowledgeModal(true); }}>新增条目</Button><Button loading={knowledgeLoading} onClick={loadKnowledge}>刷新</Button>{knowledgeKind === 'opening_phrases' ? <Button onClick={() => normalizeWorkshopOpening().then(loadKnowledge)}>规范化开头词</Button> : null}</Space></div>{knowledgeKind === 'opening_phrases' ? <div className="novel-fetch-panel"><Input.TextArea rows={4} value={openingText} onChange={event => setOpeningText(event.target.value)} placeholder="粘贴爆款开头进行分析" /><Space style={{ marginTop: 8 }}><Button onClick={handleOpeningAnalyze}>分析爆款开头</Button>{openingItems.map((item, index) => <Button key={item.id || index} onClick={() => saveWorkshopOpening(item).then(loadKnowledge)}>保存分析项</Button>)}</Space></div> : null}<Table className="novel-fetch-panel" rowKey="id" loading={knowledgeLoading} dataSource={(knowledge.items || []).filter(item => { const q = knowledgeSearch.trim().toLowerCase(); return !q || `${item.title || item.name || ''} ${item.content || item.prompt || item.text || ''}`.toLowerCase().includes(q); })} columns={[{ title: '条目', render: (_, item) => item.title || item.name || item.text || item.id }, { title: '内容', render: (_, item) => String(item.content || item.prompt || item.text || '').slice(0, 100) }, { title: '操作', render: (_, item) => <Space><Button size="small" onClick={() => { knowledgeForm.setFieldsValue(item); setKnowledgeModal(true); }}>编辑</Button><Button size="small" onClick={() => saveWorkshopKnowledge(knowledgeKind, { ...item, id: undefined, title: `${item.title || item.name || '未命名'}（副本）` }).then(loadKnowledge)}>复制</Button><Button size="small" onClick={() => optimizeWorkshopKnowledge(knowledgeKind, item.id).then(loadKnowledge)}>AI 优化</Button><Button size="small" danger onClick={() => deleteWorkshopKnowledge(knowledgeKind, item.id).then(loadKnowledge)}>删除</Button></Space> }]} /><Modal open={knowledgeModal} title="知识条目" onCancel={() => setKnowledgeModal(false)} onOk={saveKnowledge}><Form form={knowledgeForm} layout="vertical"><Form.Item name="id" hidden><Input /></Form.Item><Form.Item name="title" label="标题"><Input /></Form.Item><Form.Item name="content" label="内容"><Input.TextArea rows={7} /></Form.Item></Form></Modal></Space>,
    rules: <Space direction="vertical" size={12} style={{ width: '100%' }}><div className="novel-fetch-panel"><Space wrap><Select value={ruleScope} onChange={setRuleScope} options={[{ value: 'original', label: '原文' }, { value: 'ai', label: 'AI 版本' }, { value: 'both', label: '原文 + AI 版本' }]} /><Button disabled={!selectedTaskData?.original} onClick={() => setRuleText(selectedTaskData.original)}>载入当前任务原文</Button><Button type="primary" loading={ruleLoading} onClick={handleRulePreview}>预览规则</Button><Button loading={ruleLoading} onClick={handleRuleSuggest}>AI 规则建议</Button><Button onClick={saveRuleConfiguration}>保存规则配置</Button></Space><Input.TextArea style={{ marginTop: 12 }} rows={9} value={ruleText} onChange={event => setRuleText(event.target.value)} placeholder="输入需要处理的文本或整批任务共同规则" /></div>{ruleResult ? <div className="novel-fetch-compare"><Tabs items={[{ key: 'result', label: '处理后文本', children: <Input.TextArea rows={14} readOnly value={ruleResult.result || ''} /> }, { key: 'trace', label: '阶段追踪', children: <Collapse items={(ruleResult.trace || []).map((item, index) => ({ key: String(index), label: item.title || `阶段 ${index + 1}`, children: JSON.stringify(item) }))} /> }, { key: 'suggestions', label: '规则建议', children: <Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }}>{JSON.stringify(ruleResult.suggestions || {}, null, 2)}</Typography.Paragraph> }]} /></div> : null}</Space>,
    submit: <Space direction="vertical" size={12} style={{ width: '100%' }}><div className="novel-fetch-panel"><Typography.Title level={5}>网站提交</Typography.Title><Alert type="info" showIcon message="仅在确认“开始自动提交”后执行外部提交。小说获取和批量工厂共用同一份视频管理系统登录会话。" /><Space wrap style={{ marginTop: 12 }}><Button loading={webSubmitLoading} onClick={() => handleWebSubmit(checkWebSubmitEnvironment, '环境检测完成')}>环境检测</Button><Button loading={webSubmitLoading} onClick={() => handleWebSubmit(syncWebSubmitConfigs, '同步网站配置完成')}>同步网站配置</Button><Button loading={webSubmitLoading} onClick={() => handleWebSubmit(syncWebSubmitStyles, '同步网站风格完成')}>同步网站风格</Button><Button disabled={!selectedTaskKeys.length} loading={webSubmitLoading} onClick={async () => { const result = await handleWebSubmit(() => previewWebSubmit({ taskIds: selectedTaskKeys }), '提交预览已生成'); if (result) setWebSubmitPlan(result.plan); }}>提交预览</Button><Button type="primary" disabled={!selectedTaskKeys.length} loading={webSubmitLoading} onClick={() => Modal.confirm({ title: '确认开始自动提交', onOk: async () => { const result = await handleWebSubmit(() => startWebSubmit({ taskIds: selectedTaskKeys }), '自动提交完成'); if (result) { setWebSubmitResults(result.results || []); setWebSubmitHistory(current => [...(result.results || []), ...current]); } } })}>开始自动提交</Button></Space><div className="novel-fetch-submit-settings"><Form layout="inline"><Form.Item label="组织归属"><Select style={{ width: 220 }} value={webSubmitConfig?.organization || undefined} placeholder="跟随网站配置档" options={(webSubmitConfig?.organization_catalog || []).map(item => ({ value: String(item.id || item.value), label: item.name || item.label || item.id }))} onChange={value => setWebSubmitConfig(current => ({ ...(current || {}), organization: value }))} /></Form.Item><Form.Item label="最少字数"><InputNumber min={0} value={webSubmitConfig?.min_text_chars ?? 0} onChange={value => setWebSubmitConfig(current => ({ ...(current || {}), min_text_chars: value || 0 }))} /></Form.Item><Form.Item label="失败重试"><InputNumber min={0} max={10} value={webSubmitConfig?.retry_times ?? 1} onChange={value => setWebSubmitConfig(current => ({ ...(current || {}), retry_times: value || 0 }))} /></Form.Item><Button onClick={() => handleWebSubmit(() => saveWebSubmitConfig(webSubmitConfig), '提交配置已保存')}>保存提交配置</Button></Form></div><Typography.Paragraph type="secondary" style={{ marginTop: 12 }}>已选任务：{selectedTaskKeys.length}；配置版本：{(webSubmitConfig?.submit_versions || webSubmitConfig?.submitVersions || []).join('、') || '默认'}；可用组织：{webSubmitConfig?.organization_catalog?.length || 0}</Typography.Paragraph>{webSubmitPlan ? <Alert type="success" message="提交预览已生成" description={`提交组：${webSubmitPlan.groups?.length || 0}，跳过：${webSubmitPlan.skipped?.length || 0}`} /> : null}</div><Table className="novel-fetch-panel" rowKey={row => `${row.bookId}-${row.version}`} dataSource={webSubmitResults} locale={{ emptyText: '暂无本次提交结果' }} columns={[{ title: '任务', dataIndex: 'bookId' }, { title: '版本', dataIndex: 'version' }, { title: '状态', dataIndex: 'status' }, { title: '失败重试', render: (_, row) => row.status === 'error' ? <Button size="small" onClick={() => startWebSubmit({ taskIds: [row.bookId], versions: [row.version] })}>失败重试</Button> : '-' }]} /><div className="novel-fetch-panel"><Typography.Title level={5}>提交历史</Typography.Title><Button size="small" loading={webSubmitLoading} onClick={loadWebSubmit}>刷新历史</Button><Table size="small" rowKey={(row, index) => `${row.book_id || row.bookId}-${row.version}-${index}`} dataSource={webSubmitHistory} pagination={{ pageSize: 10 }} columns={[{ title: '书名', dataIndex: 'book_name' }, { title: '版本', dataIndex: 'version' }, { title: '状态', dataIndex: 'status' }, { title: '结果', dataIndex: 'result' }, { title: '时间', dataIndex: 'time' }]} /></div></Space>,
    logs: <div className="novel-fetch-panel"><Typography.Title level={5}>操作日志</Typography.Title><Select value={selectedTask || undefined} style={{ width: 320, marginBottom: 12 }} placeholder="选择任务查看日志" options={tasks.map(task => ({ value: task.bookId, label: task.bookName || task.bookId }))} onChange={selectTask} />{selectedTaskData ? <Collapse items={(selectedTaskData.logs || []).map((log, index) => ({ key: String(index), label: `${log.time || ''} ${log.event || '操作日志'}`, children: <Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }}>{log.detail || JSON.stringify(log)}</Typography.Paragraph> }))} /> : <Table size="small" pagination={{ pageSize: 20 }} rowKey={(row, index) => row.id || index} dataSource={globalLogs} locale={{ emptyText: '暂无操作记录' }} columns={[{ title: '时间', dataIndex: 'time', render: (value, row) => value || row.createdAt || row.created_at }, { title: '事件', dataIndex: 'event' }, { title: '详情', render: (_, row) => String(row.detail || row.message || '').slice(0, 120) }]} />}</div>
  };

  return <Space className="novel-fetch-page novel-fetch-workbench" direction="vertical" size={0} style={{ width: '100%' }}><div className="novel-fetch-public-title"><Typography.Title level={3} style={{ margin: 0 }}>批量原文改文系统 <span className="novel-fetch-version-badge">V78</span></Typography.Title></div><Tabs className="novel-fetch-panel" activeKey={activeTab} onChange={changeTab} items={TABS.map(item => ({ ...item, children: panels[item.key] }))} /><Drawer open={Boolean(taskDetail)} title={taskDetail?.bookId ? `任务详情 · ${taskDetail.bookId}` : '任务详情'} width={860} onClose={() => setTaskDetail(null)}>{taskDetail?.data ? <Tabs items={[{ key: 'summary', label: '任务概览', children: <Descriptions bordered column={2} items={Object.entries(taskDetail.data.meta || {}).slice(0, 12).map(([key, value]) => ({ key, label: key, children: String(value ?? '') }))} /> }, { key: 'original', label: '原文', children: <Input.TextArea rows={18} readOnly value={taskDetail.data.original || ''} /> }, { key: 'ai', label: 'AI 文案', children: <Input.TextArea rows={18} readOnly value={taskDetail.data.aiText || taskDetail.data.ai_text || ''} /> }, { key: 'sensitive', label: '敏感处理', children: <Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }}>{JSON.stringify(taskDetail.data.sensitive || taskDetail.data.sensitiveLogs || taskDetail.data.meta?.sensitiveDetails || {}, null, 2)}</Typography.Paragraph> }, { key: 'logs', label: '操作日志', children: <Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }}>{JSON.stringify(taskDetail.data.logs || [], null, 2)}</Typography.Paragraph> }]} /> : null}</Drawer></Space>;
}

export default NovelFetchPage;
