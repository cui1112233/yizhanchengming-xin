import { Button, Drawer, Empty, Form, Input, InputNumber, Modal, Progress, Select, Spin, Tabs, Tooltip, message } from 'antd';
import { BarChart3, BookOpen, Bot, CalendarClock, CheckCircle2, CircleAlert, Download, FileText, LoaderCircle, Menu, NotebookPen, Plus, RotateCcw, Save, Send, Settings2, Sparkles, Square, WandSparkles } from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { listAvailableModels, modelSelectOptions } from '../../shared/api/modelCatalog.js';
import { cancelNovelCreationRun, confirmNovelCreationRun, createNovelCreationProject, createNovelCreationSession, exportNovelCreationProject, getNovelCreationProjectSettings, getNovelCreationRun, listNovelCreationModes, listNovelCreationProjects, listNovelCreationSessions, retryNovelCreationRun, sendNovelCreationMessage, updateNovelCreationProjectSettings } from '../../shared/api/novelCreation.js';
import './NovelCreationPage.css';

const ACTIVE_MODE_KEY = 'novel-creation-active-mode';
const INSPECTOR_WIDTH_KEY = 'novel-creation-inspector-width';
const MODE_INPUT_FIELDS = {
  long_imitation: [{ key: 'referenceText', label: '参考文本', placeholder: '粘贴参考作品片段，只提取节奏与句式，不复制原句' }],
  short_imitation: [{ key: 'referenceText', label: '参考文本', placeholder: '粘贴参考作品片段，只提取节奏与句式，不复制原句' }],
  fanfiction: [{ key: 'existingSetting', label: '既有人物与世界规则', placeholder: '填写原有人物性格、关系与不可违背的世界规则' }],
  continuation: [{ key: 'previousText', label: '被续写的前文', placeholder: '粘贴需要承接的前文结尾，续写不会重写这部分' }]
};
const DEFAULT_MODES = [
  { id: 'long_original', name: '长篇原创', description: '从设定、卷纲到章节持续创作' },
  { id: 'short_original', name: '短篇原创', description: '围绕单一核心冲突完成短篇' },
  { id: 'long_imitation', name: '仿写长篇', description: '提取参考作品风格后创作全新长篇' },
  { id: 'short_imitation', name: '仿写短篇', description: '按参考文本风格完成全新短篇' },
  { id: 'fanfiction', name: '番外创作', description: '基于既有人物与世界规则创作番外' },
  { id: 'continuation', name: '续写创作', description: '承接给定正文继续发展情节' }
];
const TERMINAL = new Set(['succeeded', 'succeeded_with_warnings', 'failed', 'cancelled']);
const STATUS_LABELS = {
  awaiting_confirmation: '创作蓝图待确认', queued: '排队中', planning: '生成创作蓝图', writing: '写作正文', normalizing: '整理正文', auditing: '审核章节', revising: '修订章节', settling: '结算人物与伏笔', indexing: '建立记忆索引', succeeded: '创作完成', succeeded_with_warnings: '完成，有审核提醒', failed: '创作失败', cancelled: '已取消'
};

function parseJSONValue(value) {
  if (!value) return null;
  if (typeof value === 'object') return value;
  try { return JSON.parse(value); } catch (_) { return null; }
}

function clampInspectorWidth(value) {
  return Math.min(800, Math.max(360, Number(value) || 430));
}

function initialMode() {
  return localStorage.getItem(ACTIVE_MODE_KEY) || 'long_original';
}

function initialInspectorWidth() {
  return clampInspectorWidth(localStorage.getItem(INSPECTOR_WIDTH_KEY));
}

function runProgress(snapshot) {
  const events = Array.isArray(snapshot?.events) ? snapshot.events : [];
  return Math.max(0, ...events.map(item => Number(item.progress) || 0));
}

function statusIcon(status) {
  if (status === 'failed' || status === 'cancelled') return <CircleAlert size={16} />;
  if (status === 'succeeded' || status === 'succeeded_with_warnings') return <CheckCircle2 size={16} />;
  if (status === 'awaiting_confirmation') return <FileText size={16} />;
  return <LoaderCircle className="novel-creation-spin" size={16} />;
}

export default function NovelCreationPage() {
  const [modes, setModes] = useState(DEFAULT_MODES);
  const [activeMode, setActiveMode] = useState(initialMode);
  const [projects, setProjects] = useState([]);
  const [projectId, setProjectId] = useState('');
  const [sessions, setSessions] = useState([]);
  const [sessionId, setSessionId] = useState('');
  const [models, setModels] = useState([]);
  const [modelId, setModelId] = useState('');
  const [thinkingLevel, setThinkingLevel] = useState('standard');
  const [draft, setDraft] = useState('');
  const [snapshot, setSnapshot] = useState(null);
  const [selectedChapterId, setSelectedChapterId] = useState('');
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [projectModalOpen, setProjectModalOpen] = useState(false);
  const [toolDrawer, setToolDrawer] = useState('');
  const [inspectorWidth, setInspectorWidth] = useState(initialInspectorWidth);
  const [modeInputs, setModeInputs] = useState({});
  const [blueprintDraft, setBlueprintDraft] = useState('');
  const [checkpointOpen, setCheckpointOpen] = useState(false);
  const [savingSettings, setSavingSettings] = useState(false);
  const [exporting, setExporting] = useState(false);
  const [projectForm] = Form.useForm();
  const [settingsForm] = Form.useForm();
  const resizeRef = useRef(null);
  const currentProject = projects.find(item => item.id === projectId);
  const currentMode = modes.find(item => item.id === activeMode) || modes[0];
  const currentRun = snapshot?.run;
  const modeFields = MODE_INPUT_FIELDS[activeMode] || [];
  const runtime = parseJSONValue(currentRun?.runtimeSnapshot);
  const blueprint = runtime?.foundationPlan || null;
  const messages = Array.isArray(snapshot?.messages) ? snapshot.messages : [];
  const events = Array.isArray(snapshot?.events) ? snapshot.events : [];
  const chapters = Array.isArray(snapshot?.chapters) ? snapshot.chapters : [];
  const selectedChapter = chapters.find(chapter => chapter.id === selectedChapterId) || chapters[0];
  const chapter = selectedChapter;

  useEffect(() => {
    if (chapters.length && !chapters.some(chapter => chapter.id === selectedChapterId)) {
      setSelectedChapterId(chapters[0].id);
    }
  }, [chapters, selectedChapterId]);

  const loadProjects = useCallback(async preferred => {
    const result = await listNovelCreationProjects();
    const next = Array.isArray(result?.projects) ? result.projects : [];
    setProjects(next);
    setProjectId(current => preferred || (next.some(item => item.id === current) ? current : next[0]?.id || ''));
    return next;
  }, []);

  useEffect(() => {
    let alive = true;
    Promise.allSettled([listNovelCreationModes(), listNovelCreationProjects(), listAvailableModels('text')]).then(results => {
      if (!alive) return;
      if (results[0].status === 'fulfilled' && results[0].value?.modes?.length) setModes(results[0].value.modes);
      if (results[1].status === 'fulfilled') {
        const next = results[1].value?.projects || [];
        setProjects(next);
        setProjectId(next[0]?.id || '');
      }
      if (results[2].status === 'fulfilled') {
        const options = modelSelectOptions(results[2].value, 'text');
        setModels(options);
        setModelId(options[0]?.value || '');
      }
      setLoading(false);
    });
    return () => { alive = false; };
  }, []);

  useEffect(() => {
    localStorage.setItem(ACTIVE_MODE_KEY, activeMode);
    if (!projectId) { setSessions([]); setSessionId(''); return; }
    listNovelCreationSessions(projectId).then(result => {
      const next = result?.sessions || [];
      setSessions(next);
      const matching = next.filter(item => item.mode === activeMode);
      setSessionId(current => matching.some(item => item.id === current) ? current : matching[0]?.id || '');
    }).catch(error => message.error(error.message));
  }, [activeMode, projectId]);

  useEffect(() => {
    const runId = currentRun?.id;
    if (!runId || TERMINAL.has(currentRun.status) || currentRun.status === 'awaiting_confirmation') return undefined;
    const controller = new AbortController();
    const timer = window.setInterval(() => {
      getNovelCreationRun(runId, 0, controller.signal).then(setSnapshot).catch(error => { if (!error?.aborted) message.error(error.message); });
    }, 1500);
    return () => { controller.abort(); window.clearInterval(timer); };
  }, [currentRun?.id, currentRun?.status]);

  useEffect(() => {
    const move = event => {
      if (!resizeRef.current) return;
      const width = clampInspectorWidth(window.innerWidth - event.clientX);
      setInspectorWidth(width);
      localStorage.setItem(INSPECTOR_WIDTH_KEY, String(width));
    };
    const stop = () => { resizeRef.current = null; document.body.classList.remove('novel-creation-resizing'); };
    window.addEventListener('pointermove', move);
    window.addEventListener('pointerup', stop);
    return () => { window.removeEventListener('pointermove', move); window.removeEventListener('pointerup', stop); };
  }, []);

  async function createProject(values) {
    const result = await createNovelCreationProject(values);
    await loadProjects(result.project.id);
    setModelId(values.modelId);
    setProjectModalOpen(false);
    projectForm.resetFields();
    message.success('作品已创建');
  }

  async function ensureSession() {
    const existing = sessions.find(item => item.id === sessionId && item.mode === activeMode) || sessions.find(item => item.mode === activeMode);
    if (existing) { setSessionId(existing.id); return existing; }
    const result = await createNovelCreationSession(projectId, { mode: activeMode, title: `${currentMode?.name || '创作'}会话`, contextBudget: 120000 });
    setSessions(current => [result.session, ...current]);
    setSessionId(result.session.id);
    return result.session;
  }

  async function send() {
    if (!projectId) { setProjectModalOpen(true); message.info('先创建一个作品'); return; }
    if (!draft.trim()) { message.warning('请先输入创作要求或正文'); return; }
    if (!modelId) { message.warning('请先选择文本模型'); return; }
    const missing = modeFields.find(field => !String(modeInputs[field.key] || '').trim());
    if (missing) { message.warning(`请先填写${missing.label}`); return; }
    const payloadInputs = Object.fromEntries(modeFields
      .map(field => [field.key, String(modeInputs[field.key] || '').trim()])
      .filter(([, value]) => value));
    setSending(true);
    try {
      const session = await ensureSession();
      const result = await sendNovelCreationMessage(session.id, { projectId, content: draft.trim(), modelId, thinkingLevel, idempotencyKey: crypto.randomUUID(), modeInputs: payloadInputs });
      setDraft('');
      setSnapshot({ project: currentProject, session, run: result.run, messages: [{ role: 'user', content: draft.trim() }], events: [] });
    } catch (error) {
      message.error(error.message);
    } finally {
      setSending(false);
    }
  }

  async function retry() {
    if (!currentRun?.id) return;
    const result = await retryNovelCreationRun(currentRun.id);
    setSnapshot(current => ({ ...current, run: result.run, events: [] }));
  }

  async function cancel() {
    if (!currentRun?.id) return;
    const result = await cancelNovelCreationRun(currentRun.id);
    setSnapshot(current => ({ ...current, run: result.run }));
  }

  async function confirmBlueprint(blueprintOverride) {
    if (!currentRun?.id) return;
    try {
      const result = await confirmNovelCreationRun(currentRun.id, blueprintOverride || null);
      setCheckpointOpen(false);
      setBlueprintDraft('');
      setSnapshot(current => ({ ...current, run: result.run, events: [] }));
      message.success('蓝图已确认，开始写正文');
    } catch (error) {
      message.error(error.message);
    }
  }

  function openBlueprintEditor() {
    setBlueprintDraft(JSON.stringify(blueprint || {}, null, 2));
    setCheckpointOpen(true);
  }

  async function submitBlueprint() {
    try {
      const parsed = blueprintDraft.trim() ? JSON.parse(blueprintDraft) : null;
      await confirmBlueprint(parsed);
    } catch (error) {
      message.error(error instanceof SyntaxError ? '蓝图 JSON 格式无效' : error.message);
    }
  }

  async function exportBook() {
    if (!projectId) return;
    setExporting(true);
    try {
      const blob = await exportNovelCreationProject(projectId, 'markdown');
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `${currentProject?.title || '小说'}.md`;
      document.body.appendChild(link);
      link.click();
      link.remove();
      URL.revokeObjectURL(url);
    } catch (error) {
      message.error(error.message);
    } finally {
      setExporting(false);
    }
  }

  async function openTool(label) {
    setToolDrawer(label);
    if (label !== '写作风格' || !projectId) return;
    try {
      const result = await getNovelCreationProjectSettings(projectId);
      const settings = result?.settings || {};
      settingsForm.setFieldsValue({ styleProfile: settings.styleProfile || '', rules: (settings.rules || []).join('\n') });
    } catch (error) {
      message.error(error.message);
    }
  }

  async function saveSettings() {
    if (!projectId) { message.warning('请先选择作品'); return; }
    setSavingSettings(true);
    try {
      const values = await settingsForm.validateFields();
      const rules = String(values.rules || '').split('\n').map(item => item.trim()).filter(Boolean);
      await updateNovelCreationProjectSettings(projectId, { styleProfile: String(values.styleProfile || '').trim(), rules });
      message.success('写作风格与规则已保存');
    } catch (error) {
      if (error?.errorFields) return;
      message.error(error.message);
    } finally {
      setSavingSettings(false);
    }
  }

  function createSessionFromButton() {
    if (!projectId) {
      setProjectModalOpen(true);
      message.info('先创建一个作品');
      return;
    }
    setSessionId('');
    setSnapshot(null);
    setDraft('');
  }

  const toolbar = useMemo(() => [
    ['写作风格', FileText], ['导入素材', BookOpen], ['Agent 设置', Bot], ['定时写作', CalendarClock], ['数据看板', BarChart3], ['备忘录', NotebookPen]
  ], []);

  if (loading) return <div className="novel-creation-loading"><Spin size="large" /><span>正在打开小说创作</span></div>;

  return <div className="novel-creation-page" style={{ '--novel-inspector-width': `${inspectorWidth}px` }}>
    <main className={`novel-creation-workbench${currentRun ? ' has-run' : ''}`}>
      <aside className="novel-creation-sidebar">
    <header className="novel-creation-launcher">
      <div className="novel-creation-launcher-head"><strong>开始创作</strong></div>
      <div className="novel-creation-mode-tabs">{modes.map(mode => <button key={mode.id} type="button" className={mode.id === activeMode ? 'active' : ''} onClick={() => { setActiveMode(mode.id); setSnapshot(null); setModeInputs({}); }}><Sparkles size={15} />{mode.name}</button>)}</div>
    </header>
        <div className="novel-creation-sidebar-head"><span>我的作品 {projects.length}</span><Tooltip title="新建作品"><Button type="text" icon={<Plus size={17} />} onClick={() => setProjectModalOpen(true)} /></Tooltip></div>
        {projects.length ? <Select className="novel-creation-project-select" value={projectId || undefined} placeholder="创建或选择作品" onChange={setProjectId} options={projects.map(item => ({ value: item.id, label: item.title }))} /> : <p className="novel-creation-session-empty">还没有作品</p>}
        <div className="novel-creation-session-list"><div className="novel-creation-mode-title">会话记录</div>{sessions.filter(item => item.mode === activeMode).length ? sessions.filter(item => item.mode === activeMode).map(item => <button key={item.id} type="button" className={item.id === sessionId ? 'active' : ''} onClick={() => setSessionId(item.id)}><Menu size={14} /><span>{item.title || '未命名会话'}</span></button>) : <p className="novel-creation-session-empty">当前分类暂无会话，点击下方新建</p>}<Button className="novel-creation-new-session" icon={<Plus size={15} />} onClick={createSessionFromButton}>新建会话</Button></div>
        <div className="novel-creation-toolbar">{toolbar.map(([label, Icon]) => <Button key={label} icon={<Icon size={15} />} onClick={() => openTool(label)}>{label}</Button>)}</div>
      </aside>

      <section className="novel-creation-conversation">
        <div className="novel-creation-chat-scroll">
          {!messages.length && !currentRun ? <div className="novel-creation-welcome"><p>告诉我你想写什么——题材、世界观、主角、核心冲突。</p></div> : null}
          {messages.map(item => <article key={item.id || `${item.role}-${item.sequence}-${item.content}`} className={`novel-creation-message ${item.role}`}><b>{item.role === 'user' ? '我' : '创作 Agent'}</b><p>{item.content}</p></article>)}
          {events.filter(item => item.type === 'assistant.delta').map(item => <article key={item.eventId} className="novel-creation-message assistant"><b>创作 Agent</b><p>{item.message}</p></article>)}
          {currentRun ? <div className={`novel-creation-run-card ${currentRun.status}`}><div>{statusIcon(currentRun.status)}<strong>{STATUS_LABELS[currentRun.status] || currentRun.status}</strong><span>{runProgress(snapshot)}%</span></div><Progress percent={runProgress(snapshot)} showInfo={false} /><p>{events.at(-1)?.message || currentRun.errorMessage || '任务已创建，等待执行'}</p>{currentRun.errorMessage ? <div className="novel-creation-error">{currentRun.errorMessage}</div> : null}{currentRun.status === 'awaiting_confirmation' ? <div className="novel-creation-checkpoint"><div><b>{blueprint?.title || '创作蓝图'}</b><p>{blueprint?.coreConflict || '请确认基础设定后再写正文'}</p></div><div className="novel-creation-run-actions"><Button type="primary" icon={<CheckCircle2 size={14} />} onClick={() => confirmBlueprint(null)}>确认蓝图</Button><Button icon={<FileText size={14} />} onClick={openBlueprintEditor}>修改蓝图</Button></div></div> : null}<div className="novel-creation-run-actions">{['failed', 'cancelled', 'succeeded_with_warnings'].includes(currentRun.status) ? <Button icon={<RotateCcw size={14} />} onClick={retry}>从失败处重试</Button> : null}{!TERMINAL.has(currentRun.status) && currentRun.status !== 'awaiting_confirmation' ? <Button icon={<Square size={13} />} onClick={cancel}>停止</Button> : null}</div></div> : null}
        </div>
        <div className="novel-creation-composer">{modeFields.length ? <div className="novel-creation-mode-inputs">{modeFields.map(field => <label key={field.key}><span>{field.label}</span><Input.TextArea rows={2} value={modeInputs[field.key] || ''} placeholder={field.placeholder} onChange={event => setModeInputs(current => ({ ...current, [field.key]: event.target.value }))} /></label>)}</div> : null}<Input.TextArea value={draft} onChange={event => setDraft(event.target.value)} autoSize={{ minRows: 3, maxRows: 9 }} placeholder="输入指令… 可在模型列表中设置思考强度 (Enter 发送，Shift+Enter 换行)" onPressEnter={event => { if (!event.shiftKey) { event.preventDefault(); send(); } }} /><div className="novel-creation-composer-controls"><Select value={modelId || undefined} placeholder="选择文本模型" options={models} onChange={setModelId} /><Select value={thinkingLevel} onChange={setThinkingLevel} options={[{ value: 'fast', label: '快速' }, { value: 'standard', label: '标准思考' }, { value: 'deep', label: '深度思考' }]} /><span>{currentMode?.name}</span><Button type="primary" icon={<Send size={15} />} loading={sending} onClick={send}>开始创作</Button></div></div>
      </section>

      <div className="novel-creation-resizer" onPointerDown={event => { resizeRef.current = event.clientX; document.body.classList.add('novel-creation-resizing'); }} />
      <aside className="novel-creation-inspector">
        <Tabs defaultActiveKey="progress" items={[
          { key: 'progress', label: '创作进度', children: <div className="novel-creation-inspector-panel"><h3>实时执行链</h3>{events.length ? <ol className="novel-creation-timeline">{events.filter(item => item.type !== 'assistant.delta').map(item => <li key={item.eventId} className={item.status}><span /> <div><b>{STATUS_LABELS[item.status] || item.step || item.type}</b><p>{item.message}</p></div></li>)}</ol> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="开始创作后显示每一步状态" />}</div> },
          { key: 'outline', label: '大纲', children: <div className="novel-creation-inspector-panel"><h3>作品设定</h3>{currentProject ? <dl><dt>题材</dt><dd>{currentProject.genre}</dd><dt>世界</dt><dd>{currentProject.world || '等待补充'}</dd><dt>主角</dt><dd>{currentProject.protagonist || '等待补充'}</dd><dt>核心冲突</dt><dd>{currentProject.coreConflict || '等待补充'}</dd></dl> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="请先选择作品" />}</div> },
          { key: 'chapter', label: `章节${chapters.length ? ` (${chapters.length})` : ''}`, children: <div className="novel-creation-inspector-panel"><div className="novel-creation-chapter-tools"><span>{chapters.length ? `已保存 ${chapters.length} 章` : '尚无章节'}</span><Button size="small" icon={<Download size={13} />} loading={exporting} disabled={!chapters.length} onClick={exportBook}>导出整本</Button></div>{chapters.length ? <><div className="novel-creation-chapter-list">{chapters.map(item => <button type="button" key={item.id} className={item.id === chapter?.id ? 'active' : ''} onClick={() => setSelectedChapterId(item.id)}><b>{item.title}</b><span>{item.wordCount} 字</span></button>)}</div>{chapter ? <article className="novel-creation-chapter-body"><h3>{chapter.title}</h3><small>{chapter.wordCount} 字</small><p>{chapter.content}</p></article> : null}</> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={currentRun?.status === 'succeeded' ? '本次没有保存章节正文' : '完成创作后在这里查看章节'} />}</div> }
        ]} />
      </aside>
    </main>

    <Modal open={projectModalOpen} title="新建小说作品" okText="创建作品" cancelText="取消" onCancel={() => setProjectModalOpen(false)} onOk={() => projectForm.submit()} destroyOnClose>
      <Form form={projectForm} layout="vertical" onFinish={createProject} initialValues={{ genre: '都市', modelId, targetChapterWords: 2200 }}>
        <Form.Item name="title" label="作品名称" rules={[{ required: true, message: '请输入作品名称' }]}><Input placeholder="例如：雨夜归人" /></Form.Item>
        <div className="novel-creation-form-grid"><Form.Item name="genre" label="题材" rules={[{ required: true, message: '请输入题材' }]}><Input /></Form.Item><Form.Item name="targetChapterWords" label="每章目标字数"><InputNumber min={800} max={10000} step={100} /></Form.Item></div>
        <Form.Item name="modelId" label="文本模型" rules={[{ required: true, message: '请选择文本模型' }]}><Select options={models} placeholder="使用 API 配置中的文本模型" /></Form.Item>
        <Form.Item name="world" label="世界设定"><Input.TextArea rows={2} /></Form.Item>
        <Form.Item name="protagonist" label="主角设定"><Input.TextArea rows={2} /></Form.Item>
        <Form.Item name="coreConflict" label="核心冲突"><Input.TextArea rows={2} /></Form.Item>
      </Form>
    </Modal>
    <Modal open={checkpointOpen} title="确认或修改创作蓝图" okText="确认蓝图" cancelText="取消" width={760} onCancel={() => setCheckpointOpen(false)} onOk={submitBlueprint}>
      <p className="novel-creation-settings-hint">蓝图会被保存并在后续章节复用。只修改需要调整的字段，保存后立即进入正文写作。</p>
      <Input.TextArea className="novel-creation-blueprint-editor" value={blueprintDraft} onChange={event => setBlueprintDraft(event.target.value)} autoSize={{ minRows: 14, maxRows: 24 }} />
    </Modal>
    <Drawer open={Boolean(toolDrawer)} title={toolDrawer} width={520} onClose={() => setToolDrawer('')}>
      {toolDrawer === '写作风格' ? <Form form={settingsForm} layout="vertical">
        <p className="novel-creation-settings-hint">风格与规则按作品保存，会注入基础设定、章节计划、正文和审核。</p>
        <Form.Item name="styleProfile" label="写作风格"><Input.TextArea rows={5} placeholder="例如：冷峻短句、克制对白、以环境推动情绪" /></Form.Item>
        <Form.Item name="rules" label="创作规则（每行一条）"><Input.TextArea rows={6} placeholder="例如：不出现血腥描写&#10;主角不得无理由背叛同伴" /></Form.Item>
        <Button type="primary" icon={<Save size={15} />} loading={savingSettings} disabled={!projectId} onClick={saveSettings}>保存风格与规则</Button>
      </Form> : <div className="novel-creation-tool-placeholder"><Settings2 size={28} /><h3>{toolDrawer}</h3><p>该面板沿用当前作品和会话，不会另建一套数据。设置保存后会进入下一次创作任务的输入快照。</p></div>}
    </Drawer>
  </div>;
}
