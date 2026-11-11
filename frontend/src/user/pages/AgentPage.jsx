import { Button, Form, Input, Modal, Popconfirm, Segmented, Space, Tag, Tooltip, Typography, message } from 'antd';
import { FilePenLine, MessageSquarePlus, Pencil, RotateCcw, Trash2 } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { askAgent, clearAgentTask, createAgentTask, createMyAgentSkill, deleteAgentTask, deleteMyAgentSkill, getAgentTask, getMyAgentSkill, listAgentSkills, listAgentTasks, renameAgentTask, updateMyAgentSkill } from '../../shared/api/agent';
import { dispatchPetContext, dispatchPetSkills, dispatchPetState } from '../../shared/pet/stacky';

const blankSkill = { name: '', description: '', category: '创作', inputTemplate: '', body: '' };

function formatTaskTime(value) {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '';
  return new Intl.DateTimeFormat('zh-CN', { month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit' }).format(date);
}

function taskSummary(task) {
  const lastMessage = task.messages?.at(-1);
  return {
    id: task.id,
    title: task.title,
    createdAt: task.createdAt,
    updatedAt: task.updatedAt,
    messageCount: task.messages?.length || 0,
    preview: lastMessage?.content?.slice(0, 120) || ''
  };
}

export function AgentPage() {
  const [tasks, setTasks] = useState([]);
  const [activeTaskId, setActiveTaskId] = useState(null);
  const [activeTask, setActiveTask] = useState(null);
  const [skills, setSkills] = useState([]);
  const [selectedSkillIds, setSelectedSkillIds] = useState([]);
  const [source, setSource] = useState('all');
  const [question, setQuestion] = useState('');
  const [composerMenuOpen, setComposerMenuOpen] = useState(false);
  const [composerPanel, setComposerPanel] = useState(null);
  const [selectedMode, setSelectedMode] = useState('创作助手');
  const [selectedExpert, setSelectedExpert] = useState('CM 创作顾问');
  const [attachedFile, setAttachedFile] = useState(null);
  const [loading, setLoading] = useState(true);
  const [taskDetailLoading, setTaskDetailLoading] = useState(false);
  const [asking, setAsking] = useState(false);
  const [skillEditorOpen, setSkillEditorOpen] = useState(false);
  const [editingSkillId, setEditingSkillId] = useState(null);
  const [savingSkill, setSavingSkill] = useState(false);
  const [renameOpen, setRenameOpen] = useState(false);
  const [renameValue, setRenameValue] = useState('');
  const [renaming, setRenaming] = useState(false);
  const [form] = Form.useForm();
  const historyRef = useRef(null);
  const attachmentInputRef = useRef(null);
  const taskRequestRef = useRef(0);
  const initializationRequestRef = useRef(0);
  const skillsRequestRef = useRef(0);
  const activeTaskIdRef = useRef(null);
  const mountedRef = useRef(false);
  const tasksRef = useRef([]);
  const taskDetailLoadingRef = useRef(false);
  const composerTriggerRef = useRef(null);
  const composerMenuId = 'agent-composer-menu';

  function syncTaskUrl(taskId, historyMode = 'replace') {
    const nextUrl = taskId ? `/agent?task=${encodeURIComponent(taskId)}` : '/agent';
    if (historyMode === 'none' || `${window.location.pathname}${window.location.search}` === nextUrl) return;
    if (historyMode === 'push') {
      window.history.pushState({}, '', nextUrl);
    } else {
      window.history.replaceState({}, '', nextUrl);
    }
  }

  function activateTask(task, { historyMode = 'replace' } = {}) {
    activeTaskIdRef.current = task?.id || null;
    setActiveTaskId(task?.id || null);
    setActiveTask(task || null);
    syncTaskUrl(task?.id || null, historyMode);
  }

  async function loadSkills() {
    const requestId = ++skillsRequestRef.current;
    const result = await listAgentSkills();
    if (!mountedRef.current || skillsRequestRef.current !== requestId) return null;
    setSkills(Array.isArray(result.skills) ? result.skills : []);
    return result;
  }

  async function refreshTasks(isCurrent = () => true) {
    const result = await listAgentTasks();
    const nextTasks = Array.isArray(result.tasks) ? result.tasks : [];
    if (!mountedRef.current || !isCurrent()) return null;
    tasksRef.current = nextTasks;
    setTasks(nextTasks);
    return nextTasks;
  }

  async function selectTask(taskId, { historyMode = 'replace' } = {}) {
    syncTaskUrl(taskId, historyMode);
    const requestId = ++taskRequestRef.current;
    activeTaskIdRef.current = taskId;
    setAsking(false);
    setActiveTaskId(taskId);
    setActiveTask(null);
    taskDetailLoadingRef.current = true;
    setTaskDetailLoading(true);
    try {
      const result = await getAgentTask(taskId);
      if (mountedRef.current && taskRequestRef.current === requestId) activateTask(result.task || null, { historyMode: 'none' });
    } catch (error) {
      if (!mountedRef.current || taskRequestRef.current !== requestId) return;
      activateTask(null, { historyMode: 'replace' });
      message.error(error.message || '读取任务失败');
      if (error?.status === 404 || /\b404\b|not found|不存在/i.test(String(error?.message || error || ''))) {
        refreshTasks()
          .then(nextTasks => {
            if (mountedRef.current && taskRequestRef.current === requestId && nextTasks?.[0]) return selectTask(nextTasks[0].id, { historyMode: 'replace' });
            return undefined;
          })
          .catch(() => undefined);
      } else {
        refreshTasks().catch(() => undefined);
      }
    } finally {
      if (mountedRef.current && taskRequestRef.current === requestId) {
        taskDetailLoadingRef.current = false;
        setTaskDetailLoading(false);
      }
    }
  }

  async function createTask(expectedRequestId = taskRequestRef.current, { historyMode = 'push' } = {}) {
    try {
      const result = await createAgentTask();
      const task = result.task;
      if (!task) throw new Error('未能创建任务');
      if (!mountedRef.current || taskRequestRef.current !== expectedRequestId) return null;
      const requestId = ++taskRequestRef.current;
      taskDetailLoadingRef.current = false;
      setTaskDetailLoading(false);
      activateTask(task, { historyMode });
      await refreshTasks(() => mountedRef.current && taskRequestRef.current === requestId && activeTaskIdRef.current === task.id);
      if (!mountedRef.current || taskRequestRef.current !== requestId || activeTaskIdRef.current !== task.id) return null;
      return { task, requestId };
    } catch (error) {
      if (mountedRef.current && taskRequestRef.current === expectedRequestId) message.error(error.message || '新建聊天失败');
      return null;
    }
  }

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
      taskRequestRef.current += 1;
      initializationRequestRef.current += 1;
      skillsRequestRef.current += 1;
    };
  }, []);

  useEffect(() => {
    let cancelled = false;
    const initializationRequestId = ++initializationRequestRef.current;
    const initialTaskRequestId = taskRequestRef.current;
    dispatchPetContext({ page: 'CM 画布' });
    Promise.all([refreshTasks(), loadSkills()])
      .then(([nextTasks]) => {
        if (cancelled
          || initializationRequestRef.current !== initializationRequestId
          || taskRequestRef.current !== initialTaskRequestId) return undefined;
        const requestedTaskId = new URLSearchParams(window.location.search).get('task');
        const requestedTask = nextTasks.find(task => task.id === requestedTaskId);
        if (requestedTask) return selectTask(requestedTask.id, { historyMode: 'replace' });
        if (nextTasks[0]) return selectTask(nextTasks[0].id, { historyMode: 'replace' });
        syncTaskUrl(null, 'replace');
        return undefined;
      })
      .catch(error => {
        if (!cancelled && initializationRequestRef.current === initializationRequestId) {
          message.error(error.message || '读取 CM 画布失败');
        }
      })
      .finally(() => {
        if (!cancelled && initializationRequestRef.current === initializationRequestId) setLoading(false);
      });
    return () => {
      cancelled = true;
      initializationRequestRef.current += 1;
    };
  }, []);

  useEffect(() => {
    function selectHistoryTask() {
      if (!mountedRef.current) return;
      const requestedTaskId = new URLSearchParams(window.location.search).get('task');
      const requestedTask = tasksRef.current.find(task => task.id === requestedTaskId);
      if (requestedTask) {
        if (activeTaskIdRef.current !== requestedTask.id) selectTask(requestedTask.id, { historyMode: 'none' });
        return;
      }

      if (activeTaskIdRef.current !== null || taskDetailLoadingRef.current) {
        taskRequestRef.current += 1;
        taskDetailLoadingRef.current = false;
        setTaskDetailLoading(false);
        setAsking(false);
        activateTask(null, { historyMode: 'none' });
      }
    }

    window.addEventListener('popstate', selectHistoryTask);
    return () => window.removeEventListener('popstate', selectHistoryTask);
  }, []);

  useEffect(() => { dispatchPetSkills(selectedSkillIds); }, [selectedSkillIds]);
  useEffect(() => {
    function closeComposerMenu(event) {
      if (event.key === 'Escape' && composerMenuOpen) {
        setComposerMenuOpen(false);
        setComposerPanel(null);
        composerTriggerRef.current?.focus();
      }
    }

    window.addEventListener('keydown', closeComposerMenu);
    return () => window.removeEventListener('keydown', closeComposerMenu);
  }, [composerMenuOpen]);
  useEffect(() => {
    historyRef.current?.scrollTo({ top: historyRef.current.scrollHeight, behavior: 'smooth' });
  }, [activeTask?.id, activeTask?.messages?.length]);

  const visibleSkills = useMemo(() => skills.filter(skill => source === 'all' || skill.source === source), [skills, source]);
  const selectedSkills = skills.filter(skill => selectedSkillIds.includes(skill.id));
  const messages = activeTask?.messages || [];

  function toggleSkill(id) {
    setSelectedSkillIds(current => current.includes(id) ? current.filter(value => value !== id) : current.length >= 3 ? current : [...current, id]);
  }

  function useTemplate(skill) {
    if (!skill.inputTemplate) return;
    if (question.trim() && !window.confirm('将用技能模板替换当前输入内容吗？')) return;
    setQuestion(skill.inputTemplate);
  }

  function openComposerTool(tool) {
    if (tool === 'file') {
      attachmentInputRef.current?.click();
      setComposerMenuOpen(false);
      return;
    }
    if (tool === 'connector') {
      message.info('当前账号尚未配置连接器。连接器需要由管理员单独接入后才会显示。');
      setComposerMenuOpen(false);
      return;
    }
    setComposerPanel(current => current === tool ? null : tool);
  }

  function readAttachment(event) {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file) return;
    if (!/\.(txt|md|csv|json)$/i.test(file.name)) {
      message.error('当前仅支持 TXT、MD、CSV、JSON 文本文件。');
      return;
    }
    if (file.size > 200 * 1024) {
      message.error('附件不能超过 200 KB。');
      return;
    }
    const reader = new FileReader();
    reader.onload = () => setAttachedFile({ name: file.name, content: String(reader.result || '').slice(0, 20000) });
    reader.onerror = () => message.error('读取附件失败');
    reader.readAsText(file);
  }

  async function sendQuestion() {
    const prompt = question.trim();
    if (!prompt || asking || taskDetailLoadingRef.current) return;
    let requestId = taskRequestRef.current;
    let taskId = activeTask?.id || null;
    function isCurrentTaskRequest() {
      return mountedRef.current && taskRequestRef.current === requestId && activeTaskIdRef.current === taskId;
    }
    setAsking(true);
    setQuestion('');
    dispatchPetState('working');
    try {
      const createdTask = activeTask ? { task: activeTask, requestId } : await createTask(requestId);
      if (!createdTask) return;
      const task = createdTask.task;
      taskId = task.id;
      requestId = createdTask.requestId;
      const result = await askAgent({
        taskId: task.id,
        prompt,
        context: {
          page: 'CM 画布',
          mode: selectedMode,
          expert: selectedExpert,
          attachment: attachedFile ? { name: attachedFile.name, content: attachedFile.content } : undefined
        },
        skillIds: selectedSkillIds
      });
      const nextTask = result.task || { ...task, messages: [...(task.messages || []), result.user, result.assistant] };
      if (!isCurrentTaskRequest()) return;
      setActiveTask(nextTask);
      await refreshTasks(isCurrentTaskRequest);
      if (isCurrentTaskRequest()) dispatchPetState('success');
    } catch (error) {
      if (!isCurrentTaskRequest()) return;
      message.error(error.message || 'CM 暂时无法回答');
      if (taskId) {
        getAgentTask(taskId)
          .then(result => {
            if (isCurrentTaskRequest()) setActiveTask(result.task || null);
          })
          .catch(() => undefined);
      }
      refreshTasks(isCurrentTaskRequest).catch(() => undefined);
      if (isCurrentTaskRequest()) dispatchPetState('error');
    } finally {
      if (isCurrentTaskRequest()) setAsking(false);
    }
  }

  function openRenameTask() {
    if (!activeTask) return;
    setRenameValue(activeTask.title);
    setRenameOpen(true);
  }

  function toggleComposerMenu() {
    setComposerMenuOpen(current => !current);
    setComposerPanel(null);
  }

  async function saveRename() {
    if (!activeTaskId) return;
    const taskId = activeTaskId;
    setRenaming(true);
    try {
      const result = await renameAgentTask(taskId, renameValue);
      const nextTask = result.task;
      if (activeTaskIdRef.current === taskId) setActiveTask(nextTask);
      setRenameOpen(false);
      await refreshTasks();
      message.success('任务已重命名');
    } catch (error) {
      message.error(error.message || '重命名失败');
    } finally {
      setRenaming(false);
    }
  }

  async function clearCurrentTask() {
    if (!activeTaskId) return;
    const taskId = activeTaskId;
    taskRequestRef.current += 1;
    try {
      await clearAgentTask(taskId);
      if (activeTaskIdRef.current === taskId) setActiveTask(current => current ? { ...current, messages: [] } : current);
      await refreshTasks();
      message.success('当前任务已清空');
    } catch (error) {
      message.error(error.message || '清空任务失败');
    }
  }

  async function deleteCurrentTask() {
    if (!activeTaskId) return;
    const taskId = activeTaskId;
    try {
      await deleteAgentTask(taskId);
      const deletedActiveTask = activeTaskIdRef.current === taskId;
      if (deletedActiveTask) {
        taskRequestRef.current += 1;
        activateTask(null, { historyMode: 'replace' });
      }
      const nextTasks = await refreshTasks();
      if (deletedActiveTask && nextTasks[0]) await selectTask(nextTasks[0].id, { historyMode: 'replace' });
      message.success('当前任务已删除');
    } catch (error) {
      message.error(error.message || '删除任务失败');
    }
  }

  function openCreateSkill() {
    setEditingSkillId(null);
    form.setFieldsValue(blankSkill);
    setSkillEditorOpen(true);
  }

  async function openEditSkill(id) {
    try {
      const result = await getMyAgentSkill(id);
      setEditingSkillId(id);
      form.setFieldsValue(result.skill);
      setSkillEditorOpen(true);
    } catch (error) { message.error(error.message || '无法读取技能'); }
  }

  async function saveSkill(values) {
    setSavingSkill(true);
    try {
      if (editingSkillId) await updateMyAgentSkill(editingSkillId, values);
      else await createMyAgentSkill(values);
      await loadSkills();
      setSkillEditorOpen(false);
      message.success('我的技能已保存');
    } catch (error) { message.error(error.message || '保存技能失败'); }
    finally { setSavingSkill(false); }
  }

  async function removeSkill(id) {
    try {
      await deleteMyAgentSkill(id);
      setSelectedSkillIds(current => current.filter(value => value !== id));
      await loadSkills();
      message.success('我的技能已删除');
    } catch (error) { message.error(error.message || '删除技能失败'); }
  }

  return (
    <div className="agent-page utility-workbench">
      <aside className="agent-task-sidebar cm-conversation-frame" aria-label="CM 任务列表">
        <header className="agent-task-sidebar-header">
          <div><Typography.Title level={4}>任务</Typography.Title><Typography.Text>当前账号的 CM 对话</Typography.Text></div>
          <Tooltip title="新建聊天"><Button aria-label="新建聊天" icon={<MessageSquarePlus size={17} />} type="primary" onClick={() => createTask()} /></Tooltip>
        </header>
        <Button className="agent-task-new" icon={<MessageSquarePlus size={16} />} type="primary" block onClick={() => createTask()}>新建聊天</Button>
        <div className="agent-task-list" aria-live="polite">
          {loading ? <div className="agent-task-loading" role="status">正在读取任务列表...</div> : null}
          {!loading && tasks.length === 0 ? <div className="agent-task-empty">还没有任务<br />从一段新对话开始。</div> : null}
          {tasks.map(task => <button className={`agent-task-row${task.id === activeTaskId ? ' active' : ''}`} type="button" key={task.id} onClick={() => selectTask(task.id, { historyMode: 'push' })}>
            <strong>{task.title}</strong><span>{task.preview || '尚未开始对话'}</span><time>{formatTaskTime(task.updatedAt)}</time>
          </button>)}
        </div>
      </aside>
      <section className="agent-workbench cm-conversation-frame" aria-labelledby="agent-workbench-title">
        <header className="agent-workbench-header">
          <div><Typography.Title id="agent-workbench-title" level={3}>{activeTask?.title || 'CM Agent'}</Typography.Title><Typography.Paragraph>{activeTask ? '当前任务的消息只会保存在你的账号中。' : '新建聊天后，CM 会将每段对话单独保存。'}</Typography.Paragraph></div>
          {activeTask ? <Space size={2} className="agent-task-actions">
            <Tooltip title="重命名任务"><Button aria-label="重命名任务" type="text" icon={<Pencil size={17} />} onClick={openRenameTask} /></Tooltip>
            <Popconfirm title="清空当前任务的全部消息？" description="不会影响其他任务。" okText="清空" cancelText="取消" onConfirm={clearCurrentTask}><Tooltip title="清空当前任务"><Button aria-label="清空当前任务" type="text" icon={<RotateCcw size={17} />} disabled={messages.length === 0} /></Tooltip></Popconfirm>
            <Popconfirm title="删除当前任务？" description="删除后无法恢复，不会影响其他任务。" okText="删除" cancelText="取消" okButtonProps={{ danger: true }} onConfirm={deleteCurrentTask}><Tooltip title="删除当前任务"><Button aria-label="删除当前任务" danger type="text" icon={<Trash2 size={17} />} /></Tooltip></Popconfirm>
          </Space> : null}
        </header>
        <div ref={historyRef} className="agent-workbench-history" aria-live="polite">
          {taskDetailLoading ? <div className="agent-workbench-loading" role="status">正在读取任务详情...</div> : null}
          {!loading && !activeTask ? <div className="agent-workbench-empty">新建聊天，开始一段独立的 CM 对话。</div> : null}
          {activeTask && messages.length === 0 ? <div className="agent-workbench-empty">在这里输入需求，CM 会围绕当前任务继续对话。</div> : null}
          {messages.map((entry, index) => <article key={`${entry.createdAt || index}-${entry.role}`} className={`agent-workbench-message agent-workbench-message--${entry.role}`}><strong>{entry.role === 'assistant' ? 'CM' : '你'}</strong><span>{entry.content}</span></article>)}
        </div>
        <div className="agent-workbench-composer">
          <input ref={attachmentInputRef} className="agent-attachment-input" type="file" accept=".txt,.md,.csv,.json,text/plain,text/markdown,text/csv,application/json" onChange={readAttachment} />
          <div className="agent-selected-skills">{selectedSkills.map(skill => <Tag closable key={skill.id} onClose={() => toggleSkill(skill.id)}>{skill.name}</Tag>)}{attachedFile ? <Tag closable onClose={() => setAttachedFile(null)}>附件：{attachedFile.name}</Tag> : null}</div>
          <Input.TextArea value={question} disabled={taskDetailLoading} autoSize={{ minRows: 2, maxRows: 5 }} placeholder={taskDetailLoading ? '正在读取任务详情...' : '今天想让 CM 帮你做什么？可引用内容、调用技能。'} onChange={event => setQuestion(event.target.value)} onPressEnter={event => { if (event.nativeEvent.isComposing) return; if (!event.shiftKey) { event.preventDefault(); sendQuestion(); } }} />
          <div className="agent-composer-actions">
            <div className="agent-composer-plus-wrap">
              <button ref={composerTriggerRef} className="agent-composer-plus" type="button" aria-label="添加上下文或调用能力" aria-expanded={composerMenuOpen} aria-controls={composerMenuId} title="添加上下文或调用能力" onClick={toggleComposerMenu}>+</button>
              {composerMenuOpen ? <div id={composerMenuId} role="menu" className="agent-composer-menu">
                <button role="menuitem" type="button" onClick={() => openComposerTool('file')}>⌇ <span>添加文件</span><i>›</i></button>
                <button role="menuitem" type="button" onClick={() => openComposerTool('mode')}>◌ <span>模式</span><i>›</i></button>
                <button role="menuitem" type="button" onClick={() => openComposerTool('expert')}>◉ <span>专家</span><i>›</i></button>
                <button role="menuitem" type="button" onClick={() => openComposerTool('skill')}>⌘ <span>技能</span><i>›</i></button>
                <button role="menuitem" type="button" onClick={() => openComposerTool('connector')}>⌁ <span>连接器</span><i>›</i></button>
              </div> : null}
            </div>
            <Button type="primary" loading={asking} disabled={!question.trim() || taskDetailLoading} onClick={sendQuestion}>发送</Button>
          </div>
          {composerPanel === 'mode' ? <div className="agent-composer-context"><strong>模式</strong>{['创作助手', '分析拆解', '修改润色'].map(value => <button type="button" key={value} className={selectedMode === value ? 'active' : ''} onClick={() => { setSelectedMode(value); setComposerPanel(null); setComposerMenuOpen(false); }}>{value}</button>)}</div> : null}
          {composerPanel === 'expert' ? <div className="agent-composer-context"><strong>专家</strong>{['CM 创作顾问', '前贴片广告策划', '短剧编剧'].map(value => <button type="button" key={value} className={selectedExpert === value ? 'active' : ''} onClick={() => { setSelectedExpert(value); setComposerPanel(null); setComposerMenuOpen(false); }}>{value}</button>)}</div> : null}
          {composerPanel === 'skill' ? <div className="agent-skill-panel"><header><div><strong>技能</strong><span>最多选择 3 项</span></div><Button size="small" icon={<FilePenLine size={14} />} onClick={openCreateSkill}>新建我的技能</Button></header><Segmented value={source} onChange={setSource} options={[{ label: '全部', value: 'all' }, { label: '平台自带', value: 'system' }, { label: '我的技能', value: 'user' }]} /><div className="agent-skill-list">{visibleSkills.map(skill => <article className={`agent-skill-card${selectedSkillIds.includes(skill.id) ? ' selected' : ''}`} key={skill.id}><div><strong>{skill.name}</strong><Tag>{skill.source === 'system' ? '平台自带' : '我的技能'}</Tag></div><p>{skill.description || '未填写说明'}</p><Space size="small" wrap><Button size="small" type={selectedSkillIds.includes(skill.id) ? 'primary' : 'default'} onClick={() => toggleSkill(skill.id)}>{selectedSkillIds.includes(skill.id) ? '已选用' : '选用'}</Button>{skill.inputTemplate ? <Button size="small" onClick={() => useTemplate(skill)}>填入模板</Button> : null}{skill.source === 'user' ? <Button size="small" onClick={() => openEditSkill(skill.id)}>编辑</Button> : null}{skill.source === 'user' ? <Popconfirm title="删除这项我的技能？" onConfirm={() => removeSkill(skill.id)}><Button size="small" danger>删除</Button></Popconfirm> : null}</Space></article>)}</div></div> : null}
        </div>
      </section>
      <Modal title={editingSkillId ? '编辑我的技能' : '新建我的技能'} open={skillEditorOpen} onCancel={() => setSkillEditorOpen(false)} footer={null} destroyOnClose>
        <Form form={form} layout="vertical" onFinish={saveSkill}><Form.Item label="名称" name="name" rules={[{ required: true }]}><Input /></Form.Item><Form.Item label="说明" name="description"><Input /></Form.Item><Form.Item label="分类" name="category"><Input /></Form.Item><Form.Item label="输入模板" name="inputTemplate"><Input.TextArea autoSize={{ minRows: 3 }} /></Form.Item><Form.Item label="技能规则" name="body" rules={[{ required: true }]}><Input.TextArea autoSize={{ minRows: 8 }} /></Form.Item><Button htmlType="submit" type="primary" loading={savingSkill}>保存</Button></Form>
      </Modal>
      <Modal title="重命名任务" open={renameOpen} onCancel={() => setRenameOpen(false)} onOk={saveRename} okText="保存" cancelText="取消" confirmLoading={renaming} destroyOnClose>
        <Input aria-label="任务名称" autoFocus value={renameValue} maxLength={80} onChange={event => setRenameValue(event.target.value)} onPressEnter={saveRename} />
      </Modal>
    </div>
  );
}
