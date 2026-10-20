import { Button, Form, Input, Modal, Popconfirm, Segmented, Space, Tag, Tooltip, Typography, message } from 'antd';
import { ChevronDown, ChevronUp, Clock3, FilePenLine, MessageCircle, MessageSquarePlus, MoreHorizontal, Pencil, Pin, RotateCcw, Save, SearchCheck, Send, Trash2 } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { askAgent, clearAgentTask, createAgentTask, createMyAgentSkill, deleteAgentTask, deleteMyAgentSkill, getAgentCanvas, getAgentTask, getMyAgentSkill, listAgentExecutions, listAgentSkills, listAgentTasks, renameAgentTask, saveAgentCanvas, updateMyAgentSkill } from '../../shared/api/agent';
import { dispatchPetContext, dispatchPetSkills, dispatchPetState } from '../../shared/pet/stacky';
import SuperOpcLiquidMetalLogo from '../../shared/components/SuperOpcLiquidMetalLogo';
import { AgentCanvas } from '../components/AgentCanvas';
import { buildSelectedNodeContext, canvasSaveAction, canvasSaveStatusText, createCanvasSaveQueue, shouldKeepPendingCreatedTask } from './agent-canvas-state';

const blankSkill = { name: '', description: '', category: '创作', inputTemplate: '', body: '' };
function formatTaskTime(value) {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '';
  return new Intl.DateTimeFormat('zh-CN', { month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit' }).format(date);
}

function taskGroup(value) {
  const date = new Date(value || 0);
  if (Number.isNaN(date.getTime())) return '更早';
  const today = new Date();
  const startToday = new Date(today.getFullYear(), today.getMonth(), today.getDate()).getTime();
  const startYesterday = startToday - 24 * 60 * 60 * 1000;
  if (date.getTime() >= startToday) return '今天';
  if (date.getTime() >= startYesterday) return '昨天';
  return '更早';
}

function createProblemSearchPrompt(text) {
  return [
    '请对下面这段文字进行“找问题”诊断。',
    '要求：',
    '1. 只指出有明确依据的问题，不要为了凑数量而挑错。',
    '2. 优先检查节奏、冲突、人物动机、信息密度、逻辑连续性、结尾钩子与可读性。',
    '3. 按【问题类型｜严重程度｜原文定位｜为什么是问题｜修改建议】组织结果。',
    '4. 先给问题清单，再给最值得优先修改的 1-3 项。',
    '5. 不要直接重写全文，除非我继续要求。',
    '',
    '待检查文字：',
    text
  ].join('\n');
}

export function AgentPageV2({ theme }) {
  const [tasks, setTasks] = useState([]);
  const [activeTaskId, setActiveTaskId] = useState(null);
  const [activeTask, setActiveTask] = useState(null);
  const [invalidCanvasTask, setInvalidCanvasTask] = useState(false);
  const [isCanvasWorkspace, setIsCanvasWorkspace] = useState(() => window.location.pathname === '/agent/canvas');
  const [skills, setSkills] = useState([]);
  const [selectedSkillIds, setSelectedSkillIds] = useState([]);
  const [source, setSource] = useState('all');
  const [question, setQuestion] = useState('');
  const [composerMenuOpen, setComposerMenuOpen] = useState(false);
  const [composerPanel, setComposerPanel] = useState(null);
  const [selectedMode, setSelectedMode] = useState('创作助手');
  const [selectedExpert, setSelectedExpert] = useState('CM 创作顾问');
  const [attachedFile, setAttachedFile] = useState(null);
  const [problemSearchMode, setProblemSearchMode] = useState(false);
  const [taskDrawerOpen, setTaskDrawerOpen] = useState(false);
  const [conversationPanelOpen, setConversationPanelOpen] = useState(true);
  const [loading, setLoading] = useState(true);
  const [taskDetailLoading, setTaskDetailLoading] = useState(false);
  const [canvas, setCanvas] = useState({ version: 1, revision: 0, nodes: [], edges: [] });
  const [canvasSaveStatus, setCanvasSaveStatus] = useState('idle');
  const [canvasViewport, setCanvasViewport] = useState({ x: 0, y: 0, zoom: 1 });
  const [executions, setExecutions] = useState([]);
  const [selectedCanvasNodes, setSelectedCanvasNodes] = useState([]);
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
  const canvasSaveQueueRef = useRef(null);
  const composerMenuId = 'agent-composer-menu';

  function syncTaskUrl(taskId, historyMode = 'replace') {
    const nextUrl = taskId ? `/agent/canvas?taskId=${encodeURIComponent(taskId)}` : '/agent';
    if (historyMode === 'none' || `${window.location.pathname}${window.location.search}` === nextUrl) return;
    if (historyMode === 'push') window.history.pushState({}, '', nextUrl);
    else window.history.replaceState({}, '', nextUrl);
    window.dispatchEvent(new PopStateEvent('popstate'));
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
      if (mountedRef.current && taskRequestRef.current === requestId) {
        setInvalidCanvasTask(!result.task);
        activateTask(result.task || null, { historyMode: 'none' });
      }
    } catch (error) {
      if (!mountedRef.current || taskRequestRef.current !== requestId) return;
      setInvalidCanvasTask(isCanvasWorkspace);
      activateTask(null, { historyMode: isCanvasWorkspace ? 'none' : 'replace' });
      message.error(error.message || '读取任务失败');
      refreshTasks().catch(() => undefined);
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
      setIsCanvasWorkspace(true);
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
    dispatchPetContext({ page: 'Agent 工作区' });
    Promise.all([refreshTasks(), loadSkills()])
      .then(([nextTasks]) => {
        if (cancelled || initializationRequestRef.current !== initializationRequestId || taskRequestRef.current !== initialTaskRequestId) return undefined;
        const params = new URLSearchParams(window.location.search);
        const legacyTaskId = params.get('task');
        const requestedTaskId = isCanvasWorkspace ? params.get('taskId') : legacyTaskId;
        const requestedTask = nextTasks.find(task => task.id === requestedTaskId);
        if (requestedTask) return selectTask(requestedTask.id, { historyMode: 'replace' });
        if (isCanvasWorkspace && requestedTaskId) {
          setInvalidCanvasTask(true);
          return undefined;
        }
        if (legacyTaskId) {
          const nextUrl = `/agent/canvas?taskId=${encodeURIComponent(legacyTaskId)}`;
          window.history.replaceState({}, '', nextUrl);
          setIsCanvasWorkspace(true);
          return undefined;
        }
        syncTaskUrl(null, 'replace');
        return undefined;
      })
      .catch(error => {
        if (!cancelled && initializationRequestRef.current === initializationRequestId) message.error(error.message || '读取 Agent 工作区失败');
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
      const nextIsCanvasWorkspace = window.location.pathname === '/agent/canvas';
      setIsCanvasWorkspace(nextIsCanvasWorkspace);
      const requestedTaskId = nextIsCanvasWorkspace ? new URLSearchParams(window.location.search).get('taskId') : null;
      const requestedTask = tasksRef.current.find(task => task.id === requestedTaskId);
      if (requestedTask) {
        setInvalidCanvasTask(false);
        if (activeTaskIdRef.current !== requestedTask.id) selectTask(requestedTask.id, { historyMode: 'none' });
        return;
      }
      if (nextIsCanvasWorkspace && shouldKeepPendingCreatedTask({ requestedTaskId, activeTaskId: activeTaskIdRef.current, knownTaskIds: tasksRef.current.map(task => task.id) })) {
        setInvalidCanvasTask(false);
        return;
      }
      if (nextIsCanvasWorkspace && requestedTaskId) setInvalidCanvasTask(true);
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
      if (event.key === 'Escape' && (composerMenuOpen || composerPanel)) {
        setComposerMenuOpen(false);
        setComposerPanel(null);
        composerTriggerRef.current?.focus();
      }
    }
    window.addEventListener('keydown', closeComposerMenu);
    return () => window.removeEventListener('keydown', closeComposerMenu);
  }, [composerMenuOpen, composerPanel]);

  useEffect(() => {
    historyRef.current?.scrollTo({ top: historyRef.current.scrollHeight, behavior: 'smooth' });
  }, [activeTask?.id, activeTask?.messages?.length]);

  useEffect(() => {
    let cancelled = false;
    let saveQueue = null;
    canvasSaveQueueRef.current?.dispose();
    canvasSaveQueueRef.current = null;
    if (!isCanvasWorkspace || !activeTaskId) {
      setCanvas({ version: 1, revision: 0, nodes: [], edges: [] });
      setCanvasSaveStatus('idle');
      setCanvasViewport({ x: 0, y: 0, zoom: 1 });
      setExecutions([]);
      setSelectedCanvasNodes([]);
      return undefined;
    }
    Promise.all([getAgentCanvas(activeTaskId), listAgentExecutions(activeTaskId)])
      .then(([canvasResult, executionResult]) => {
        if (cancelled) return;
        const loadedCanvas = canvasResult.canvas || { version: 1, revision: 0, nodes: [], edges: [] };
        setCanvas(loadedCanvas);
        setCanvasViewport(loadedCanvas.viewport || { x: 0, y: 0, zoom: 1 });
        setCanvasSaveStatus('idle');
        saveQueue = createCanvasSaveQueue({
          revision: loadedCanvas.revision,
          save: (nextCanvas, revision) => saveAgentCanvas(activeTaskId, nextCanvas, revision),
          reload: () => getAgentCanvas(activeTaskId),
          onCanvas: setCanvas,
          onStatus: setCanvasSaveStatus
        });
        canvasSaveQueueRef.current = saveQueue;
        setExecutions(Array.isArray(executionResult.executions) ? executionResult.executions : []);
      })
      .catch(error => { if (!cancelled) message.error(error.message || '读取创作画布失败'); });
    return () => {
      cancelled = true;
      saveQueue?.dispose();
      if (canvasSaveQueueRef.current === saveQueue) canvasSaveQueueRef.current = null;
    };
  }, [activeTaskId, isCanvasWorkspace]);

  const visibleSkills = useMemo(() => skills.filter(skill => source === 'all' || skill.source === source), [skills, source]);
  const selectedSkills = skills.filter(skill => selectedSkillIds.includes(skill.id));
  const messages = activeTask?.messages || [];
  const selectedNodeContext = buildSelectedNodeContext(canvas, selectedCanvasNodes);
  const saveAction = canvasSaveAction(canvasSaveStatus);
  const taskGroups = useMemo(() => tasks.reduce((groups, task) => {
    const key = taskGroup(task.updatedAt);
    groups[key].push(task);
    return groups;
  }, { 今天: [], 昨天: [], 更早: [] }), [tasks]);

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
    setComposerMenuOpen(false);
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

  async function sendQuestion({ problemSearch = problemSearchMode } = {}) {
    const rawPrompt = question.trim();
    if (!rawPrompt || asking || taskDetailLoadingRef.current) return;
    const prompt = problemSearch ? createProblemSearchPrompt(rawPrompt) : rawPrompt;
    let requestId = taskRequestRef.current;
    let taskId = activeTask?.id || null;

    function isCurrentTaskRequest() {
      return mountedRef.current && taskRequestRef.current === requestId && activeTaskIdRef.current === taskId;
    }

    setAsking(true);
    setQuestion('');
    dispatchPetState('working', { title: 'CM 正在处理任务' });
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
          page: 'Agent 工作区',
          mode: problemSearch ? '分析拆解' : selectedMode,
          expert: selectedExpert,
          intent: problemSearch ? 'find-problems' : 'chat',
          originalInput: problemSearch ? rawPrompt : undefined,
          selectedNodes: selectedNodeContext,
          attachment: attachedFile ? { name: attachedFile.name, content: attachedFile.content } : undefined
        },
        skillIds: selectedSkillIds
      });
      const nextTask = result.task || { ...task, messages: [...(task.messages || []), result.user, result.assistant] };
      if (!isCurrentTaskRequest()) return;
      setActiveTask(nextTask);
      await refreshTasks(isCurrentTaskRequest);
      if (isCurrentTaskRequest()) dispatchPetState('success', { title: 'CM 已完成本次处理', detail: '回复已写入当前任务对话。' });
    } catch (error) {
      if (!isCurrentTaskRequest()) return;
      message.error(error.message || 'CM 暂时无法回答');
      if (taskId) {
        getAgentTask(taskId)
          .then(result => { if (isCurrentTaskRequest()) setActiveTask(result.task || null); })
          .catch(() => undefined);
      }
      refreshTasks(isCurrentTaskRequest).catch(() => undefined);
      if (isCurrentTaskRequest()) dispatchPetState('error', { title: 'CM 本次处理失败', detail: error.message || '请稍后重试。' });
    } finally {
      if (isCurrentTaskRequest()) setAsking(false);
    }
  }

  function openRenameTask() {
    if (!activeTask) return;
    setRenameValue(activeTask.title);
    setRenameOpen(true);
  }

  async function saveRename() {
    if (!activeTaskId) return;
    const taskId = activeTaskId;
    setRenaming(true);
    try {
      const result = await renameAgentTask(taskId, renameValue);
      if (activeTaskIdRef.current === taskId) setActiveTask(result.task);
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
      if (deletedActiveTask && nextTasks?.[0]) await selectTask(nextTasks[0].id, { historyMode: 'replace' });
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
    } catch (error) {
      message.error(error.message || '无法读取技能');
    }
  }

  async function saveSkill(values) {
    setSavingSkill(true);
    try {
      if (editingSkillId) await updateMyAgentSkill(editingSkillId, values);
      else await createMyAgentSkill(values);
      await loadSkills();
      setSkillEditorOpen(false);
      message.success('我的技能已保存');
    } catch (error) {
      message.error(error.message || '保存技能失败');
    } finally {
      setSavingSkill(false);
    }
  }

  async function removeSkill(id) {
    try {
      await deleteMyAgentSkill(id);
      setSelectedSkillIds(current => current.filter(value => value !== id));
      await loadSkills();
      message.success('我的技能已删除');
    } catch (error) {
      message.error(error.message || '删除技能失败');
    }
  }

  function saveCanvasNow() {
    const queue = canvasSaveQueueRef.current;
    if (!queue || !saveAction.method) return;
    if (saveAction.method === 'retry') void queue.retry();
    else void queue.flush();
  }

  return (
    <div className={`agent-page utility-workbench agent-workspace-v3${isCanvasWorkspace && !taskDrawerOpen ? ' agent-task-drawer-collapsed' : ''}`}>
      {isCanvasWorkspace ? <div className={conversationPanelOpen ? 'agent-canvas-workspace' : 'agent-canvas-workspace is-conversation-collapsed'}>
        <header className="agent-canvas-toolbar">
          <button className="agent-canvas-toolbar-button" type="button" aria-expanded={taskDrawerOpen} aria-controls="agent-project-drawer" onClick={() => setTaskDrawerOpen(open => !open)}>项目</button>
          <button className="agent-canvas-project-title" type="button" onClick={openRenameTask}>{activeTask?.title || '未选择项目'}<Pencil size={14} /></button>
          <span className={`agent-canvas-save-status is-${canvasSaveStatus}`}>{asking ? '正在执行…' : canvasSaveStatusText[canvasSaveStatus]} · {canvas.nodes.length} 个节点</span>
          <Tooltip title={saveAction.label}><span><button className="agent-canvas-toolbar-button" type="button" aria-label={saveAction.label} disabled={saveAction.disabled} onClick={saveCanvasNow}><Save size={14} />{saveAction.label}</button></span></Tooltip>
          <span className="agent-canvas-view-info">缩放 {Math.round((canvasViewport.zoom || 1) * 100)}%</span>
          <button className="agent-canvas-toolbar-button" type="button" aria-pressed={conversationPanelOpen} onClick={() => setConversationPanelOpen(open => !open)}>{conversationPanelOpen ? '收起对话' : '打开对话'}</button>
        </header>

        <aside id="agent-project-drawer" className={`agent-project-drawer${taskDrawerOpen ? ' is-open' : ''}`} aria-hidden={!taskDrawerOpen}>
          <header><strong>项目</strong><Button className="agent-task-new" icon={<MessageSquarePlus size={15} />} onClick={() => { setTaskDrawerOpen(false); createTask(); }}>新建项目</Button></header>
          <div className="agent-task-list" aria-live="polite">
            {loading ? <div className="agent-task-loading" role="status">正在读取项目...</div> : null}
            {!loading && tasks.length === 0 ? <div className="agent-task-empty">还没有项目<br />从一段新对话开始。</div> : null}
            {Object.entries(taskGroups).map(([group, groupTasks]) => groupTasks.length ? <section className="agent-task-group" key={group}><h4>{group}</h4>{groupTasks.map(task => (
              <button className={`agent-task-row${task.id === activeTaskId ? ' active' : ''}`} type="button" key={task.id} onClick={() => { selectTask(task.id, { historyMode: 'push' }); setTaskDrawerOpen(false); }}>
                <strong>{task.title}</strong><span>{task.preview || '尚未开始对话'}</span>
              </button>
            ))}</section> : null)}
          </div>
          {activeTask ? <footer><Button type="text" icon={<RotateCcw size={15} />} disabled={messages.length === 0} onClick={clearCurrentTask}>清空对话</Button><Popconfirm title="删除当前项目？" description="删除后无法恢复，不会影响其他项目。" okText="删除" cancelText="取消" okButtonProps={{ danger: true }} onConfirm={deleteCurrentTask}><Button type="text" danger icon={<Trash2 size={15} />}>删除项目</Button></Popconfirm></footer> : null}
        </aside>

        <main className="agent-canvas-main" aria-label="创作画布">
          <AgentCanvas
            key={activeTaskId}
            taskId={activeTaskId}
            canvas={canvas}
            onSelectionChange={setSelectedCanvasNodes}
            onViewportChange={setCanvasViewport}
            onChange={nextCanvas => {
              if (!activeTaskId) return;
              setCanvas(nextCanvas);
              canvasSaveQueueRef.current?.enqueue(nextCanvas);
            }}
            onServerCanvasChange={async () => {
              if (!activeTaskId) return;
              const result = await getAgentCanvas(activeTaskId);
              const nextCanvas = result.canvas || { version: 1, revision: 0, nodes: [], edges: [] };
              if (canvasSaveQueueRef.current) await canvasSaveQueueRef.current?.acceptServerCanvas(nextCanvas);
              else setCanvas(nextCanvas);
              setCanvasViewport(nextCanvas.viewport || { x: 0, y: 0, zoom: 1 });
            }}
          />
        </main>

        <aside className="agent-conversation-panel" aria-label="项目对话">
          <div ref={historyRef} className="agent-workbench-history" aria-live="polite">
            {taskDetailLoading ? <div className="agent-workbench-loading" role="status">正在读取任务详情...</div> : null}
            {!loading && !activeTask ? <div className="agent-workbench-empty">{invalidCanvasTask ? <><strong>找不到这个任务</strong><span>任务可能已删除或链接已失效。<a href="/agent">返回创作首页</a></span></> : <><SuperOpcLiquidMetalLogo theme={theme} className="agent-workbench-logo" /><strong>一战晟铭 / CM</strong><span>你的 AI 创作工作空间</span></>}</div> : null}
            {activeTask && messages.length === 0 ? <div className="agent-workbench-empty"><SuperOpcLiquidMetalLogo theme={theme} className="agent-workbench-logo" /><strong>一战晟铭 / CM</strong><span>输入需求开始对话；也可以直接开启「找问题」。</span></div> : null}
            {messages.map((entry, index) => <article key={`${entry.createdAt || index}-${entry.role}`} className={`agent-workbench-message agent-workbench-message--${entry.role}`}><strong>{entry.role === 'assistant' ? 'CM' : '你'}</strong><span>{entry.content}</span></article>)}
          </div>
          {executions.length ? <section className="agent-execution-records" aria-label="执行记录"><strong>执行记录</strong>{executions.slice(0, 4).map((execution, index) => <span key={execution.id || index}>{execution.toolName || execution.tool || execution.name || '工具'} · {execution.status || '已完成'}</span>)}</section> : null}
          {selectedNodeContext.length ? <div className="agent-selected-node-references" aria-label="已选节点引用">{selectedNodeContext.map(node => <Tag key={node.id}>{node.title}</Tag>)}</div> : null}

        <div className={`agent-workbench-composer${problemSearchMode ? ' is-problem-search' : ''}`}>
          <input ref={attachmentInputRef} className="agent-attachment-input" type="file" accept=".txt,.md,.csv,.json,text/plain,text/markdown,text/csv,application/json" onChange={readAttachment} />
          <div className="agent-selected-skills">
            {problemSearchMode ? <Tag closable onClose={() => setProblemSearchMode(false)}>找问题模式</Tag> : null}
            {selectedSkills.map(skill => <Tag closable key={skill.id} onClose={() => toggleSkill(skill.id)}>{skill.name}</Tag>)}
            {attachedFile ? <Tag closable onClose={() => setAttachedFile(null)}>附件：{attachedFile.name}</Tag> : null}
          </div>
          <Input.TextArea
            value={question}
            disabled={taskDetailLoading}
            autoSize={{ minRows: 2, maxRows: 8 }}
            placeholder={taskDetailLoading ? '正在读取任务详情...' : problemSearchMode ? '粘贴需要检查的文字，然后点击「找问题」…' : '今天想让 CM 帮你做什么？'}
            onChange={event => setQuestion(event.target.value)}
            onPressEnter={event => {
              if (event.nativeEvent.isComposing) return;
              if (!event.shiftKey) {
                event.preventDefault();
                sendQuestion();
              }
            }}
          />
          <div className="agent-composer-actions">
            <div className="agent-composer-actions-left">
              <div className="agent-composer-plus-wrap">
                <button
                  ref={composerTriggerRef}
                  className="agent-composer-plus"
                  type="button"
                  aria-label="添加上下文或调用能力"
                  aria-expanded={composerMenuOpen}
                  aria-controls={composerMenuId}
                  title="添加上下文或调用能力"
                  onClick={() => { setComposerMenuOpen(open => !open); setComposerPanel(null); }}
                >+</button>
                {composerMenuOpen ? (
                  <div id={composerMenuId} role="menu" className="agent-composer-menu">
                    <button role="menuitem" type="button" onClick={() => openComposerTool('file')}>⌇ <span>添加文件</span><i>›</i></button>
                    <button role="menuitem" type="button" onClick={() => openComposerTool('mode')}>◌ <span>模式</span><i>›</i></button>
                    <button role="menuitem" type="button" onClick={() => openComposerTool('expert')}>◉ <span>专家</span><i>›</i></button>
                    <button role="menuitem" type="button" onClick={() => openComposerTool('skill')}>⌘ <span>技能</span><i>›</i></button>
                    <button role="menuitem" type="button" onClick={() => openComposerTool('connector')}>⌁ <span>连接器</span><i>›</i></button>
                  </div>
                ) : null}
              </div>
              <button className={`agent-context-chip agent-problem-chip${problemSearchMode ? ' active' : ''}`} type="button" onClick={() => setProblemSearchMode(active => !active)}>
                <SearchCheck size={14} />找问题
              </button>
              <button className="agent-context-chip" type="button" onClick={() => openComposerTool('mode')}>{selectedMode}<ChevronDown size={13} /></button>
              <button className="agent-context-chip" type="button" onClick={() => openComposerTool('expert')}>{selectedExpert}<ChevronDown size={13} /></button>
              <button className="agent-context-chip" type="button" onClick={() => openComposerTool('skill')}>技能{selectedSkillIds.length ? ` ${selectedSkillIds.length}` : ''}<ChevronDown size={13} /></button>
            </div>
            <Button
              type="primary"
              icon={problemSearchMode ? <SearchCheck size={15} /> : <Send size={15} />}
              loading={asking}
              disabled={!question.trim() || taskDetailLoading}
              onClick={() => sendQuestion()}
            >{problemSearchMode ? '找问题' : '发送'}</Button>
          </div>

          {composerPanel === 'mode' ? (
            <div className="agent-composer-context"><strong>模式</strong>{['创作助手', '分析拆解', '修改润色'].map(value => <button type="button" key={value} className={selectedMode === value ? 'active' : ''} onClick={() => { setSelectedMode(value); setComposerPanel(null); }}>{value}</button>)}</div>
          ) : null}
          {composerPanel === 'expert' ? (
            <div className="agent-composer-context"><strong>专家</strong>{['CM 创作顾问', '前贴片广告策划', '短剧编剧'].map(value => <button type="button" key={value} className={selectedExpert === value ? 'active' : ''} onClick={() => { setSelectedExpert(value); setComposerPanel(null); }}>{value}</button>)}</div>
          ) : null}
          {composerPanel === 'skill' ? (
            <div className="agent-skill-panel">
              <header><div><strong>技能</strong><span>最多选择 3 项</span></div><Button size="small" icon={<FilePenLine size={14} />} onClick={openCreateSkill}>新建我的技能</Button></header>
              <Segmented value={source} onChange={setSource} options={[{ label: '全部', value: 'all' }, { label: '平台自带', value: 'system' }, { label: '我的技能', value: 'user' }]} />
              <div className="agent-skill-list">
                {visibleSkills.map(skill => (
                  <article className={`agent-skill-card${selectedSkillIds.includes(skill.id) ? ' selected' : ''}`} key={skill.id}>
                    <div><strong>{skill.name}</strong><Tag>{skill.source === 'system' ? '平台自带' : '我的技能'}</Tag></div>
                    <p>{skill.description || '未填写说明'}</p>
                    <Space size="small" wrap>
                      <Button size="small" type={selectedSkillIds.includes(skill.id) ? 'primary' : 'default'} onClick={() => toggleSkill(skill.id)}>{selectedSkillIds.includes(skill.id) ? '已选用' : '选用'}</Button>
                      {skill.inputTemplate ? <Button size="small" onClick={() => useTemplate(skill)}>填入模板</Button> : null}
                      {skill.source === 'user' ? <Button size="small" onClick={() => openEditSkill(skill.id)}>编辑</Button> : null}
                      {skill.source === 'user' ? <Popconfirm title="删除这项我的技能？" onConfirm={() => removeSkill(skill.id)}><Button size="small" danger>删除</Button></Popconfirm> : null}
                    </Space>
                  </article>
                ))}
              </div>
            </div>
          ) : null}
        </div>
        </aside>
      </div>

      : <section className="agent-creation-home" aria-labelledby="agent-creation-home-title">
        <div className="agent-creation-home-inner">
          <Typography.Title id="agent-creation-home-title" level={2}>开始创作</Typography.Title>
          <Typography.Paragraph>告诉 CM 你想完成什么，马上进入专属创作画布。</Typography.Paragraph>
          <div className={`agent-workbench-composer agent-creation-composer${problemSearchMode ? ' is-problem-search' : ''}`}>
            <input ref={attachmentInputRef} className="agent-attachment-input" type="file" accept=".txt,.md,.csv,.json,text/plain,text/markdown,text/csv,application/json" onChange={readAttachment} />
            <div className="agent-selected-skills">{selectedSkills.map(skill => <Tag closable key={skill.id} onClose={() => toggleSkill(skill.id)}>{skill.name}</Tag>)}{attachedFile ? <Tag closable onClose={() => setAttachedFile(null)}>附件：{attachedFile.name}</Tag> : null}</div>
            <Input.TextArea value={question} autoSize={{ minRows: 4, maxRows: 8 }} placeholder="描述你的创作需求…" onChange={event => setQuestion(event.target.value)} onPressEnter={event => { if (!event.shiftKey) { event.preventDefault(); sendQuestion(); } }} />
            <div className="agent-composer-actions"><div className="agent-composer-actions-left"><button className="agent-context-chip" type="button" onClick={() => openComposerTool('mode')}>{selectedMode}<ChevronDown size={13} /></button><button className="agent-context-chip" type="button" onClick={() => openComposerTool('expert')}>{selectedExpert}<ChevronDown size={13} /></button><button className="agent-context-chip" type="button" onClick={() => openComposerTool('skill')}>技能{selectedSkillIds.length ? ` ${selectedSkillIds.length}` : ''}<ChevronDown size={13} /></button><button className="agent-context-chip" type="button" onClick={() => openComposerTool('file')}>添加附件</button></div><Button type="primary" icon={<Send size={15} />} loading={asking} disabled={!question.trim()} onClick={() => sendQuestion()}>开始创作</Button></div>
            {composerPanel === 'mode' ? <div className="agent-composer-context"><strong>模式</strong>{['创作助手', '分析拆解', '修改润色'].map(value => <button type="button" key={value} className={selectedMode === value ? 'active' : ''} onClick={() => { setSelectedMode(value); setComposerPanel(null); }}>{value}</button>)}</div> : null}
            {composerPanel === 'expert' ? <div className="agent-composer-context"><strong>专家</strong>{['CM 创作顾问', '前贴片广告策划', '短剧编剧'].map(value => <button type="button" key={value} className={selectedExpert === value ? 'active' : ''} onClick={() => { setSelectedExpert(value); setComposerPanel(null); }}>{value}</button>)}</div> : null}
            {composerPanel === 'skill' ? <div className="agent-skill-panel"><Segmented value={source} onChange={setSource} options={[{ label: '全部', value: 'all' }, { label: '平台自带', value: 'system' }, { label: '我的技能', value: 'user' }]} /><div className="agent-skill-list">{visibleSkills.map(skill => <article className={`agent-skill-card${selectedSkillIds.includes(skill.id) ? ' selected' : ''}`} key={skill.id}><strong>{skill.name}</strong><p>{skill.description || '未填写说明'}</p><Button size="small" type={selectedSkillIds.includes(skill.id) ? 'primary' : 'default'} onClick={() => toggleSkill(skill.id)}>{selectedSkillIds.includes(skill.id) ? '已选用' : '选用'}</Button></article>)}</div></div> : null}
          </div>
          <section className="agent-recent-projects" aria-label="最近项目"><h3>最近项目</h3>{loading ? <span>正在读取项目…</span> : tasks.slice(0, 6).map(task => <button type="button" key={task.id} onClick={() => { window.history.pushState({}, '', `/agent/canvas?taskId=${encodeURIComponent(task.id)}`); window.dispatchEvent(new PopStateEvent('popstate')); }}><strong>{task.title}</strong><span>{task.preview || '尚未开始对话'}</span></button>)}</section>
        </div>
      </section>}
      <Modal title={editingSkillId ? '编辑我的技能' : '新建我的技能'} open={skillEditorOpen} onCancel={() => setSkillEditorOpen(false)} footer={null} destroyOnClose>
        <Form form={form} layout="vertical" onFinish={saveSkill}>
          <Form.Item label="名称" name="name" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item label="说明" name="description"><Input /></Form.Item>
          <Form.Item label="分类" name="category"><Input /></Form.Item>
          <Form.Item label="输入模板" name="inputTemplate"><Input.TextArea autoSize={{ minRows: 3 }} /></Form.Item>
          <Form.Item label="技能规则" name="body" rules={[{ required: true }]}><Input.TextArea autoSize={{ minRows: 8 }} /></Form.Item>
          <Button htmlType="submit" type="primary" loading={savingSkill}>保存</Button>
        </Form>
      </Modal>

      <Modal title="重命名任务" open={renameOpen} onCancel={() => setRenameOpen(false)} onOk={saveRename} okText="保存" cancelText="取消" confirmLoading={renaming} destroyOnClose>
        <Input aria-label="任务名称" autoFocus value={renameValue} maxLength={80} onChange={event => setRenameValue(event.target.value)} onPressEnter={saveRename} />
      </Modal>
    </div>
  );
}
