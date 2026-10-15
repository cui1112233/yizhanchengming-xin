import { Alert, Button, Collapse, Drawer, Form, Input, InputNumber, Modal, Select, Space, Table, Tabs, Tag, Typography, message } from 'antd';
import { Download, Eye, RefreshCw } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import {
  checkWebSubmitEnvironment, fetchNovelContent, getWebSubmitConfig, previewWebSubmit,
  saveWebSubmitConfig, startWebSubmit, syncWebSubmitConfigs, syncWebSubmitStyles
} from '../../shared/api/novelFetch';
import {
  analyzeWorkshopOpening, deleteWorkshopKnowledge, deleteWorkshopTasks, fetchWorkshopOriginal,
  generateWorkshopAi, getWorkshopConfig, getWorkshopKnowledge, getWorkshopKnowledgeSummary,
  getWorkshopTask, listWorkshopTasks, normalizeWorkshopOpening, optimizeWorkshopKnowledge,
  previewWorkshopRules, restoreWorkshopOriginal, retryWorkshopTasks, saveWorkshopConfig,
  saveWorkshopKnowledge, saveWorkshopOpening, suggestWorkshopRules, testWorkshopAi
} from '../../shared/api/novelFetchWorkshop';
import './novel-fetch.css';

const TABS = [
  { key: 'process', label: '处理' },
  { key: 'tasks', label: '任务' },
  { key: 'config', label: '配置' },
  { key: 'knowledge', label: '知识库' },
  { key: 'rules', label: '处理规则' },
  { key: 'submit', label: '网站提交' },
  { key: 'logs', label: '日志' }
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
  const [configForm] = Form.useForm();
  const [knowledgeForm] = Form.useForm();
  const [activeTab, setActiveTab] = useState(initialTab);
  const [processRows, setProcessRows] = useState([]);
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
  const [openingText, setOpeningText] = useState('');
  const [openingItems, setOpeningItems] = useState([]);
  const [configLoading, setConfigLoading] = useState(false);
  const [webSubmitConfig, setWebSubmitConfig] = useState(null);
  const [webSubmitPlan, setWebSubmitPlan] = useState(null);
  const [webSubmitResults, setWebSubmitResults] = useState([]);
  const [webSubmitLoading, setWebSubmitLoading] = useState(false);

  const selectedTask = useMemo(() => selectedTaskKeys[0] || '', [selectedTaskKeys]);

  async function loadTasks() {
    setTasksLoading(true);
    try { setTasks((await listWorkshopTasks()).tasks || []); } catch (error) { message.error(error.message || '读取任务失败'); } finally { setTasksLoading(false); }
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
      configForm.setFieldsValue(data.appConfig || data.config || {});
    } catch (error) { message.error(error.message || '读取配置失败'); } finally { setConfigLoading(false); }
  }
  async function loadWebSubmit() {
    try { setWebSubmitConfig((await getWebSubmitConfig()).config || {}); } catch (error) { message.error(error.message || '读取提交配置失败'); }
  }

  useEffect(() => { loadTasks(); }, []);
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
    const bookIds = parseIds(values.bookIds);
    if (!bookIds.length) return message.warning('请填写至少一个书籍 ID');
    setProcessing(true);
    setProcessRows(bookIds.map(bookId => ({ bookId, status: 'loading' })));
    try {
      const data = await fetchNovelContent({ platform: values.platform, bookIds, maxTxt: values.maxTxt });
      const results = data.results || [];
      setProcessRows(results.map(row => ({ ...row, status: row.status === 'ok' ? 'done' : 'failed' })));
      await loadTasks();
    } catch (error) {
      setProcessRows(bookIds.map(bookId => ({ bookId, status: 'failed', error: error.message })));
      message.error(error.message || '处理失败');
    } finally { setProcessing(false); }
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

  const panels = {
    process: <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <div className="novel-fetch-grid">
        <Form className="novel-fetch-panel" form={processForm} layout="vertical" initialValues={{ platform: 2, maxTxt: 2000 }}>
          <Typography.Title level={5}>获取并处理原文</Typography.Title>
          <Form.Item name="platform" label="平台"><Select options={PLATFORMS} /></Form.Item>
          <Form.Item name="bookIds" label="书籍 ID" rules={[{ required: true, message: '请输入书籍 ID' }]}><Input.TextArea rows={5} placeholder="每行一个 ID，支持空格或逗号分隔" /></Form.Item>
          <Form.Item name="maxTxt" label="截取字数"><InputNumber min={100} max={100000} /></Form.Item>
          <Space><Button type="primary" loading={processing} onClick={handleProcess}>开始处理</Button><Button onClick={() => processForm.resetFields()}>清空</Button></Space>
        </Form>
        <div className="novel-fetch-panel"><Typography.Title level={5}>处理结果</Typography.Title><Table size="small" rowKey="bookId" pagination={false} dataSource={processRows} locale={{ emptyText: '提交任务后在这里查看处理结果' }} columns={[
          { title: '书籍 ID', dataIndex: 'bookId' }, { title: '状态', dataIndex: 'status', render: value => <Tag color={statusColor(value)}>{value === 'done' ? '完成' : value === 'failed' ? '失败' : '处理中'}</Tag> },
          { title: '内容', render: (_, row) => row.data ? <Button size="small" icon={<Download size={14} />} onClick={() => downloadText(`${row.bookId}.txt`, row.data)}>下载</Button> : row.error || '-' }
        ]} /></div>
      </div>
    </Space>,
    tasks: <Space direction="vertical" size={12} style={{ width: '100%' }}>
      <div className="novel-fetch-panel"><Space wrap><Button icon={<RefreshCw size={14} />} loading={tasksLoading} onClick={loadTasks}>刷新任务</Button><Button disabled={!selectedTaskKeys.length} onClick={handleRetryTasks}>批量重试</Button><Button danger disabled={!selectedTaskKeys.length} onClick={handleDeleteTasks}>删除选中</Button><Button disabled={!selectedTask} onClick={handleGenerateAi}>生成 AI 改文</Button></Space></div>
      <Table className="novel-fetch-panel" rowKey="bookId" loading={tasksLoading} dataSource={tasks} rowSelection={{ selectedRowKeys: selectedTaskKeys, onChange: setSelectedTaskKeys }} pagination={{ pageSize: 20 }} columns={[
        { title: '书名', dataIndex: 'bookName' }, { title: '平台', dataIndex: 'platformName' }, { title: '原文状态', dataIndex: 'originalStatus', render: value => <Tag color={statusColor(value)}>{value || '未抓取'}</Tag> },
        { title: 'AI 状态', render: (_, row) => <Tag color={statusColor(row.aiStatus)}>{row.aiGeneratedCount ? `${row.aiGeneratedCount} 版` : row.aiStatus || '未生成'}</Tag> }, { title: '操作', render: (_, row) => <Space><Button size="small" onClick={() => selectTask(row.bookId)}>选择</Button><Button size="small" icon={<Eye size={14} />} onClick={async () => setTaskDetail({ bookId: row.bookId, data: await getWorkshopTask(row.bookId) })}>详情</Button><Button size="small" onClick={() => fetchWorkshopOriginal(row.bookId, row.maxTxt).then(loadTasks)}>重新抓取</Button><Button size="small" onClick={() => restoreWorkshopOriginal(row.bookId).then(loadTasks)}>恢复原文</Button></Space> }
      ]} />
    </Space>,
    config: <Form className="novel-fetch-panel" form={configForm} layout="vertical" style={{ maxWidth: 760 }}><Typography.Title level={5}>配置</Typography.Title><Form.Item name="base_url" label="AI 基础接口"><Input /></Form.Item><Form.Item name="api_key" label="API Key"><Input.Password /></Form.Item><Form.Item name="model" label="模型"><Input /></Form.Item><Form.Item name="default_max_txt" label="默认抓取字数"><InputNumber min={100} max={100000} /></Form.Item><Form.Item name="default_ai_count" label="默认 AI 数量"><InputNumber min={1} max={20} /></Form.Item><Space><Button type="primary" loading={configLoading} onClick={async () => { await saveWorkshopConfig(configForm.getFieldsValue()); message.success('配置已保存'); }}>保存配置</Button><Button onClick={() => testWorkshopAi('rewrite').then(() => message.success('测试接口成功')).catch(error => message.error(error.message || '测试失败'))}>测试接口</Button></Space></Form>,
    knowledge: <Space direction="vertical" size={12} style={{ width: '100%' }}><div className="novel-fetch-panel"><Tabs activeKey={knowledgeKind} onChange={setKnowledgeKind} items={KNOWLEDGE_TABS.map(([key, label]) => ({ key, label: `${label}（${knowledgeSummary[key] || 0}）` }))} /><Space wrap><Button onClick={() => { knowledgeForm.resetFields(); setKnowledgeModal(true); }}>新增条目</Button><Button loading={knowledgeLoading} onClick={loadKnowledge}>刷新</Button>{knowledgeKind === 'opening_phrases' ? <Button onClick={() => normalizeWorkshopOpening().then(loadKnowledge)}>规范化开头词</Button> : null}</Space></div>{knowledgeKind === 'opening_phrases' ? <div className="novel-fetch-panel"><Input.TextArea rows={4} value={openingText} onChange={event => setOpeningText(event.target.value)} placeholder="粘贴爆款开头进行分析" /><Space style={{ marginTop: 8 }}><Button onClick={handleOpeningAnalyze}>分析爆款开头</Button>{openingItems.map((item, index) => <Button key={item.id || index} onClick={() => saveWorkshopOpening(item).then(loadKnowledge)}>保存分析项</Button>)}</Space></div> : null}<Table className="novel-fetch-panel" rowKey="id" loading={knowledgeLoading} dataSource={knowledge.items || []} columns={[{ title: '条目', render: (_, item) => item.title || item.name || item.text || item.id }, { title: '内容', render: (_, item) => String(item.content || item.prompt || item.text || '').slice(0, 100) }, { title: '操作', render: (_, item) => <Space><Button size="small" onClick={() => { knowledgeForm.setFieldsValue(item); setKnowledgeModal(true); }}>编辑</Button><Button size="small" onClick={() => optimizeWorkshopKnowledge(knowledgeKind, item.id).then(loadKnowledge)}>AI 优化</Button><Button size="small" danger onClick={() => deleteWorkshopKnowledge(knowledgeKind, item.id).then(loadKnowledge)}>删除</Button></Space> }]} /><Modal open={knowledgeModal} title="知识条目" onCancel={() => setKnowledgeModal(false)} onOk={saveKnowledge}><Form form={knowledgeForm} layout="vertical"><Form.Item name="id" hidden><Input /></Form.Item><Form.Item name="title" label="标题"><Input /></Form.Item><Form.Item name="content" label="内容"><Input.TextArea rows={7} /></Form.Item></Form></Modal></Space>,
    rules: <Space direction="vertical" size={12} style={{ width: '100%' }}><div className="novel-fetch-panel"><Space wrap><Select value={ruleScope} onChange={setRuleScope} options={[{ value: 'original', label: '原文' }, { value: 'ai', label: 'AI 版本' }]} /><Button disabled={!selectedTaskData?.original} onClick={() => setRuleText(selectedTaskData.original)}>载入当前任务原文</Button><Button type="primary" loading={ruleLoading} onClick={handleRulePreview}>预览规则</Button><Button loading={ruleLoading} onClick={handleRuleSuggest}>AI 规则建议</Button></Space><Input.TextArea style={{ marginTop: 12 }} rows={9} value={ruleText} onChange={event => setRuleText(event.target.value)} placeholder="输入需要处理的文本" /></div>{ruleResult ? <div className="novel-fetch-compare"><Tabs items={[{ key: 'result', label: '处理后文本', children: <Input.TextArea rows={14} readOnly value={ruleResult.result || ''} /> }, { key: 'trace', label: '阶段追踪', children: <Collapse items={(ruleResult.trace || []).map((item, index) => ({ key: String(index), label: item.title || `阶段 ${index + 1}`, children: JSON.stringify(item) }))} /> }, { key: 'suggestions', label: '规则建议', children: <Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }}>{JSON.stringify(ruleResult.suggestions || {}, null, 2)}</Typography.Paragraph> }]} /></div> : null}</Space>,
    submit: <Space direction="vertical" size={12} style={{ width: '100%' }}><div className="novel-fetch-panel"><Typography.Title level={5}>网站提交</Typography.Title><Alert type="info" showIcon message="仅在确认“开始自动提交”后执行外部提交。" /><Space wrap style={{ marginTop: 12 }}><Button loading={webSubmitLoading} onClick={() => handleWebSubmit(checkWebSubmitEnvironment, '环境检测完成')}>环境检测</Button><Button loading={webSubmitLoading} onClick={() => handleWebSubmit(syncWebSubmitConfigs, '同步网站配置完成')}>同步网站配置</Button><Button loading={webSubmitLoading} onClick={() => handleWebSubmit(syncWebSubmitStyles, '同步网站风格完成')}>同步网站风格</Button><Button disabled={!selectedTaskKeys.length} loading={webSubmitLoading} onClick={async () => { const result = await handleWebSubmit(() => previewWebSubmit({ taskIds: selectedTaskKeys }), '提交预览已生成'); if (result) setWebSubmitPlan(result.plan); }}>提交预览</Button><Button type="primary" disabled={!selectedTaskKeys.length} loading={webSubmitLoading} onClick={() => Modal.confirm({ title: '确认开始自动提交', onOk: async () => { const result = await handleWebSubmit(() => startWebSubmit({ taskIds: selectedTaskKeys }), '自动提交完成'); if (result) setWebSubmitResults(result.results || []); } })}>开始自动提交</Button></Space><Typography.Paragraph type="secondary" style={{ marginTop: 12 }}>已选任务：{selectedTaskKeys.length}；配置版本：{(webSubmitConfig?.submitVersions || []).join('、') || '默认'}</Typography.Paragraph>{webSubmitPlan ? <Alert type="success" message="提交预览已生成" description={`提交组：${webSubmitPlan.groups?.length || 0}，跳过：${webSubmitPlan.skipped?.length || 0}`} /> : null}</div><Table className="novel-fetch-panel" rowKey={row => `${row.bookId}-${row.version}`} dataSource={webSubmitResults} locale={{ emptyText: '暂无提交结果' }} columns={[{ title: '任务', dataIndex: 'bookId' }, { title: '版本', dataIndex: 'version' }, { title: '状态', dataIndex: 'status' }, { title: '失败重试', render: (_, row) => row.status === 'error' ? <Button size="small" onClick={() => startWebSubmit({ taskIds: [row.bookId], versions: [row.version] })}>失败重试</Button> : '-' }]} /></Space>,
    logs: <div className="novel-fetch-panel"><Typography.Title level={5}>操作日志</Typography.Title><Select value={selectedTask || undefined} style={{ width: 320, marginBottom: 12 }} placeholder="选择任务查看日志" options={tasks.map(task => ({ value: task.bookId, label: task.bookName || task.bookId }))} onChange={selectTask} />{selectedTaskData ? <Collapse items={(selectedTaskData.logs || []).map((log, index) => ({ key: String(index), label: `${log.time || ''} ${log.event || '操作日志'}`, children: <Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }}>{log.detail || JSON.stringify(log)}</Typography.Paragraph> }))} /> : <Typography.Text type="secondary">选择任务后显示原文、AI、规则和提交的操作日志。</Typography.Text>}</div>
  };

  return <Space className="novel-fetch-page novel-fetch-workbench" direction="vertical" size={16} style={{ width: '100%' }}><div className="novel-fetch-summary"><Typography.Title level={3} style={{ margin: 0 }}>小说获取</Typography.Title><Typography.Text type="secondary">统一处理、任务、配置、知识库、处理规则、网站提交与日志。</Typography.Text></div><Tabs className="novel-fetch-panel" activeKey={activeTab} onChange={changeTab} items={TABS.map(item => ({ ...item, children: panels[item.key] }))} /><Drawer open={Boolean(taskDetail)} title={taskDetail?.bookId ? `任务详情 · ${taskDetail.bookId}` : '任务详情'} width={860} onClose={() => setTaskDetail(null)}>{taskDetail?.data ? <Tabs items={[{ key: 'original', label: '原文', children: <Input.TextArea rows={18} readOnly value={taskDetail.data.original || ''} /> }, { key: 'logs', label: '操作日志', children: <Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }}>{JSON.stringify(taskDetail.data.logs || [], null, 2)}</Typography.Paragraph> }]} /> : null}</Drawer></Space>;
}

export default NovelFetchPage;
