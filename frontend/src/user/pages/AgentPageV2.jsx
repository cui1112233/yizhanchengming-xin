import { Button, Form, Input, Modal, Popconfirm, Segmented, Space, Tag, Tooltip, Typography, message } from 'antd';
import { ArrowRight, ArrowUp, Bot, ChevronDown, ChevronUp, Clock3, FilePenLine, FileText, Film, Image as ImageIcon, MessageCircle, MoreHorizontal, Music, Paperclip, Pin, RotateCcw, SearchCheck, Send, Sparkles, Trash2, X } from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { activateAgentGenerationResult, approveAgentDirectorPlan, askAgent, clearAgentTask, createAgentShotBatchNodes, createAgentTask, createMyAgentSkill, deleteAgentTask, deleteMyAgentSkill, getAgentCanvas, getAgentGenerationSummary, getAgentTask, getMyAgentSkill, listAgentAssets, listAgentExecutions, listAgentGenerations, listAgentScriptShots, listAgentSkills, listAgentTasks, renameAgentTask, saveAgentCanvas, syncAgentBatchFactoryBook, updateMyAgentSkill, uploadAgentAsset } from '../../shared/api/agent';
import { getCurrentUsername } from '../../shared/api/auth';
import { dispatchPetContext, dispatchPetSkills, dispatchPetState } from '../../shared/pet/stacky';
import SuperOpcLiquidMetalLogo from '../../shared/components/SuperOpcLiquidMetalLogo';
import { AgentCanvas } from '../components/AgentCanvas';
import { AgentDirectorConversation } from '../components/AgentDirectorConversation';
import { AgentStoryboardWorkspace } from '../components/AgentStoryboardWorkspace';
import { AgentWorkspaceDrawer } from '../components/AgentWorkspaceDrawer';
import { AgentWorkspaceHeader } from '../components/AgentWorkspaceHeader';
import { AuthenticatedAgentImage } from '../components/AuthenticatedAgentImage';
import { AuthenticatedAgentVideo } from '../components/AuthenticatedAgentVideo';
import { buildSelectedNodeContext, createCanvasSaveQueue, mergePendingCanvas, shouldAcceptCanvasRefresh, shouldKeepPendingCreatedTask } from './agent-canvas-state';
import { clearCanvasDraft, isDraftAtLeastAsFresh, loadCanvasDraft, saveCanvasDraft } from './agent-canvas-draft-storage';
import { detectDirectorMode } from './agent-director-mode';
import { buildStoryboardAssetContext, patchStoryboardNodeConfig } from './agent-storyboard-model';
import { hasStoryboardAssetDrop, readStoryboardAssetDrop } from './agent-storyboard-conversation-drag';
import { STORYBOARD_SPLIT_DEFAULTS, storyboardSplitRatio, storyboardSplitRatioFromPointer } from './agent-storyboard-split';
import { patchStoryboardState } from './agent-storyboard-state';
import { createNodeAt } from './agent-canvas-context-menu';
import { appendUniqueAssetID, classifyComposerAttachment } from './agent-composer-attachments';
import { findNodeMentionCandidates, insertNodeMention } from './agent-node-mentions';
import { normalizeWorkspaceView, workspaceViewForActiveView, workspaceViewPatch, workspaceViewQuery } from './agent-workspace-view';
import { batchFactoryGroupLabel, refreshBatchFactoryBookCanvas, syncBookThenLoadCanvas } from './agent-canvas-batch-factory';
import { agentProjectCapabilities, agentProjectTypeLabel, agentProjectTypeOptions, normalizeAgentProjectType } from './agent-project-type';

const blankSkill = { name: '', description: '', category: '创作', inputTemplate: '', body: '' };
function formatTaskTime(value) {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '';
  return new Intl.DateTimeFormat('zh-CN', { month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit' }).format(date);
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

function ConversationAssetPreview({ asset }) {
  if (asset.kind === 'image' && asset.url) return <AuthenticatedAgentImage source={asset.url} alt="" />;
  if (asset.kind === 'video' && asset.url) return <AuthenticatedAgentVideo source={asset.url} aria-label="视频素材预览" preload="metadata" muted />;
  if (asset.kind === 'audio') return <Music size={15} aria-hidden="true" />;
  if (asset.kind === 'image') return <ImageIcon size={15} aria-hidden="true" />;
  if (asset.kind === 'video') return <Film size={15} aria-hidden="true" />;
  return <FileText size={15} aria-hidden="true" />;
}

export function AgentPageV2({ theme }) {
  const [tasks, setTasks] = useState([]);
  const [activeTaskId, setActiveTaskId] = useState(null);
  const [activeTask, setActiveTask] = useState(null);
  const [invalidCanvasTask, setInvalidCanvasTask] = useState(false);
  const [isCanvasWorkspace, setIsCanvasWorkspace] = useState(() => window.location.pathname === '/agent/canvas');
  const [requestedBatchBookId, setRequestedBatchBookId] = useState(() => new URLSearchParams(window.location.search).get('bookId') || '');
  const [skills, setSkills] = useState([]);
  const [selectedSkillIds, setSelectedSkillIds] = useState([]);
  const [source, setSource] = useState('all');
  const [projectFilter, setProjectFilter] = useState('all');
  const [newProjectType, setNewProjectType] = useState('video');
  const [question, setQuestion] = useState('');
  const [composerMenuOpen, setComposerMenuOpen] = useState(false);
  const [composerPanel, setComposerPanel] = useState(null);
  const [selectedMode, setSelectedMode] = useState('创作助手');
  const [selectedExpert, setSelectedExpert] = useState('CM 创作顾问');
  const [attachedFile, setAttachedFile] = useState(null);
  const [problemSearchMode, setProblemSearchMode] = useState(false);
  const [viewState, setViewState] = useState(() => normalizeWorkspaceView({
    activeView: workspaceViewQuery(new URLSearchParams(window.location.search).get('view'))
  }));
  const [workflowFocusRequest, setWorkflowFocusRequest] = useState(0);
  const [storyboardDirectorResizing, setStoryboardDirectorResizing] = useState(false);
  const [loading, setLoading] = useState(true);
  const [taskDetailLoading, setTaskDetailLoading] = useState(false);
  const [canvas, setCanvas] = useState({ version: 1, revision: 0, nodes: [], edges: [] });
  const [canvasSaveStatus, setCanvasSaveStatus] = useState('idle');
  const [canvasViewport, setCanvasViewport] = useState({ x: 0, y: 0, zoom: 1 });
  const [workspaceAssets, setWorkspaceAssets] = useState([]);
  const [workspaceAssetsLoading, setWorkspaceAssetsLoading] = useState(false);
  const [workspaceAssetsError, setWorkspaceAssetsError] = useState('');
  const [executions, setExecutions] = useState([]);
  const [generations, setGenerations] = useState([]);
  const [generationSummary, setGenerationSummary] = useState(null);
  const [storyboardShots, setStoryboardShots] = useState([]);
  const [storyboardShotsLoading, setStoryboardShotsLoading] = useState(false);
  const [storyboardShotsError, setStoryboardShotsError] = useState('');
  const [selectedCanvasNodes, setSelectedCanvasNodes] = useState([]);
  const [selectedConversationAssetIds, setSelectedConversationAssetIds] = useState([]);
  const [conversationAssetDropActive, setConversationAssetDropActive] = useState(false);
  const [dismissedDirectorPlanIDs, setDismissedDirectorPlanIDs] = useState([]);
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
  const questionInputRef = useRef(null);
  const taskRequestRef = useRef(0);
  const initializationRequestRef = useRef(0);
  const skillsRequestRef = useRef(0);
  const activeTaskIdRef = useRef(null);
  const mountedRef = useRef(false);
  const tasksRef = useRef([]);
  const taskDetailLoadingRef = useRef(false);
  const composerTriggerRef = useRef(null);
  const canvasSaveQueueRef = useRef(null);
  const canvasStateRef = useRef(canvas);
  const canvasRef = useRef(null);
  const pendingWorkflowNodeRef = useRef('');
  const viewStateRef = useRef(viewState);
  const storyboardDirectorResizeRef = useRef(null);
  const composerMenuId = 'agent-composer-menu';

  function replaceCanvas(nextCanvas) {
    canvasStateRef.current = nextCanvas;
    setCanvas(nextCanvas);
  }

  function persistCanvasDraft(nextCanvas, taskId = activeTaskIdRef.current) {
    if (!taskId || !nextCanvas) return;
    const baseRevision = canvasSaveQueueRef.current?.currentRevision?.() ?? (Number(nextCanvas.revision) || 0);
    saveCanvasDraft(window.localStorage, getCurrentUsername(), taskId, nextCanvas, baseRevision);
  }

  const applyCanvasEdit = useCallback(nextCanvas => {
    if (!activeTaskIdRef.current || !nextCanvas) return;
    replaceCanvas(nextCanvas);
    persistCanvasDraft(nextCanvas);
    canvasSaveQueueRef.current?.enqueue(nextCanvas);
  }, []);

  function syncTaskUrl(taskId, historyMode = 'replace') {
    const searchParams = new URLSearchParams(window.location.search);
    searchParams.delete('task');
    if (taskId) {
      searchParams.set('taskId', taskId);
      searchParams.set('view', viewStateRef.current.activeView);
    } else {
      searchParams.delete('taskId');
      searchParams.delete('view');
    }
    const query = searchParams.toString();
    const nextUrl = `${taskId ? '/agent/canvas' : '/agent'}${query ? `?${query}` : ''}`;
    if (historyMode === 'none' || `${window.location.pathname}${window.location.search}` === nextUrl) return;
    if (historyMode === 'push') window.history.pushState({}, '', nextUrl);
    else window.history.replaceState({}, '', nextUrl);
    window.dispatchEvent(new PopStateEvent('popstate'));
  }

  function syncWorkspaceViewUrl(view) {
    const searchParams = new URLSearchParams(window.location.search);
    searchParams.set('view', view);
    const nextUrl = `${window.location.pathname}?${searchParams.toString()}`;
    if (`${window.location.pathname}${window.location.search}` !== nextUrl) window.history.replaceState({}, '', nextUrl);
  }

  function updateWorkspaceView(patch) {
    const nextViewState = workspaceViewPatch(viewStateRef.current, patch);
    viewStateRef.current = nextViewState;
    setViewState(nextViewState);
    if (Object.prototype.hasOwnProperty.call(patch || {}, 'activeView')) syncWorkspaceViewUrl(nextViewState.activeView);
  }

  function openNodeInWorkflow(nodeId) {
    if (!nodeId) return;
    pendingWorkflowNodeRef.current = nodeId;
    setWorkflowFocusRequest(current => current + 1);
    updateWorkspaceView({ activeView: 'workflow' });
  }

  const handleStoryboardSelectionChange = useCallback(nodeId => {
    setSelectedCanvasNodes(current => {
      const next = nodeId ? [nodeId] : [];
      return current.length === next.length && current.every((value, index) => value === next[index]) ? current : next;
    });
  }, []);

  const handleStoryboardNodeConfigChange = useCallback((nodeId, patch) => {
    if (!nodeId) return;
    const currentCanvas = canvasStateRef.current;
    const nextNodes = patchStoryboardNodeConfig(currentCanvas.nodes, nodeId, patch);
    if (nextNodes !== currentCanvas.nodes) applyCanvasEdit({ ...currentCanvas, nodes: nextNodes });
  }, [applyCanvasEdit]);

  const handleStoryboardShotUpdated = useCallback(updated => {
    if (!updated?.id) return;
    setStoryboardShots(current => current.map(shot => shot.id === updated.id ? updated : shot));
  }, []);

  const handleSharedScriptShotsChange = useCallback((scriptNodeId, nextShots) => {
    const normalizedNodeId = String(scriptNodeId || '').trim();
    if (!normalizedNodeId) return;
    const replacements = Array.isArray(nextShots) ? nextShots : [];
    setStoryboardShots(current => [
      ...current.filter(shot => String(shot?.scriptNodeId || '') !== normalizedNodeId),
      ...replacements
    ]);
  }, []);

  const handleStoryboardBatchCreate = useCallback(async payload => {
    const requestedTaskId = activeTaskIdRef.current;
    const requestedQueue = canvasSaveQueueRef.current;
    if (!requestedTaskId || !requestedQueue) throw new Error('画布保存队列尚未准备完成');
    const isCurrentRequest = () => shouldAcceptCanvasRefresh({
      requestedTaskId,
      currentTaskId: activeTaskIdRef.current,
      requestedQueue,
      currentQueue: canvasSaveQueueRef.current,
      mounted: mountedRef.current
    });
    await requestedQueue.flushRequired();
    if (!isCurrentRequest()) return { ignored: true };
    const result = await createAgentShotBatchNodes(requestedTaskId, payload);
    if (!isCurrentRequest()) return { ...result, ignored: true };
    if (result?.canvas) {
      const accepted = await requestedQueue.acceptServerCanvas(result.canvas);
      if (accepted === false || !isCurrentRequest()) return { ...result, ignored: true };
      setCanvasViewport(result.canvas.viewport || { x: 0, y: 0, zoom: 1 });
    }
    const [generationResult, executionResult, summaryResult] = await Promise.all([
      listAgentGenerations(requestedTaskId),
      listAgentExecutions(requestedTaskId),
      getAgentGenerationSummary(requestedTaskId)
    ]);
    if (!isCurrentRequest()) return { ...result, ignored: true };
    setGenerations(Array.isArray(generationResult.tasks) ? generationResult.tasks : []);
    setGenerationSummary(summaryResult?.summary || null);
    setExecutions(Array.isArray(executionResult.executions) ? executionResult.executions : []);
    return result;
  }, []);

  const handleGenerationResultActivate = useCallback(async item => {
    const requestedTaskId = activeTaskIdRef.current;
    const requestedQueue = canvasSaveQueueRef.current;
    const node = canvasStateRef.current.nodes?.find(candidate => String(candidate?.id || '') === String(item?.nodeId || ''));
    const expectedNodeVersion = Number(node?.version ?? node?.data?.version ?? 0);
    if (!requestedTaskId || !requestedQueue) throw new Error('画布保存队列尚未准备完成');
    if (!item?.generationId || !item?.resultId || item?.kind === 'video' || expectedNodeVersion < 1) throw new Error('当前结果不能切换');
    const isCurrentRequest = () => shouldAcceptCanvasRefresh({
      requestedTaskId,
      currentTaskId: activeTaskIdRef.current,
      requestedQueue,
      currentQueue: canvasSaveQueueRef.current,
      mounted: mountedRef.current
    });
    await requestedQueue.flushRequired();
    if (!isCurrentRequest()) return;
    const result = await activateAgentGenerationResult(requestedTaskId, item.generationId, item.resultId, expectedNodeVersion);
    if (!isCurrentRequest()) return;
    if (!result?.canvas) throw new Error('切换结果后没有返回画布');
    const accepted = await requestedQueue.acceptServerCanvas(result.canvas);
    if (accepted === false || !isCurrentRequest()) return;
    setCanvasViewport(result.canvas.viewport || { x: 0, y: 0, zoom: 1 });
    setGenerations(current => current.map(task => String(task?.id || '') === String(result?.task?.id || '') ? result.task : task));
    message.success('已设为当前结果');
  }, []);

  const handleStoryboardStateChange = useCallback(patch => {
    const nextStoryboard = patchStoryboardState(viewStateRef.current.storyboard, patch);
    updateWorkspaceView({ storyboard: nextStoryboard });
  }, []);

  function startStoryboardDirectorResize(event) {
    const bounds = event.currentTarget.parentElement?.getBoundingClientRect();
    if (event.button !== 0 || !bounds || storyboardDirectorResizeRef.current) return;
    storyboardDirectorResizeRef.current = { pointerId: event.pointerId, bounds };
    setStoryboardDirectorResizing(true);
    event.currentTarget.setPointerCapture?.(event.pointerId);
    event.preventDefault();
  }

  useEffect(() => {
    if (!storyboardDirectorResizing) return undefined;
    function moveStoryboardDirectorResize(event) {
      const resize = storyboardDirectorResizeRef.current;
      if (!resize || resize.pointerId !== event.pointerId) return;
      const director = storyboardSplitRatioFromPointer(event.clientX, resize.bounds, 'x', 18, STORYBOARD_SPLIT_DEFAULTS.director);
      handleStoryboardStateChange({ layout: { director } });
      event.preventDefault();
    }
    function finishStoryboardDirectorResize(event) {
      const resize = storyboardDirectorResizeRef.current;
      if (!resize || resize.pointerId !== event.pointerId) return;
      storyboardDirectorResizeRef.current = null;
      setStoryboardDirectorResizing(false);
    }
    window.addEventListener('pointermove', moveStoryboardDirectorResize);
    window.addEventListener('pointerup', finishStoryboardDirectorResize);
    window.addEventListener('pointercancel', finishStoryboardDirectorResize);
    return () => {
      window.removeEventListener('pointermove', moveStoryboardDirectorResize);
      window.removeEventListener('pointerup', finishStoryboardDirectorResize);
      window.removeEventListener('pointercancel', finishStoryboardDirectorResize);
    };
  }, [handleStoryboardStateChange, storyboardDirectorResizing]);

  function resetStoryboardDirectorSplit() {
    handleStoryboardStateChange({ layout: { director: STORYBOARD_SPLIT_DEFAULTS.director } });
  }

  function adjustStoryboardDirectorSplit(event) {
    if (!['ArrowLeft', 'ArrowRight'].includes(event.key)) return;
    event.preventDefault();
    const direction = event.key === 'ArrowLeft' ? -2 : 2;
    const director = storyboardSplitRatio(viewStateRef.current.storyboard.layout.director + direction, 18, STORYBOARD_SPLIT_DEFAULTS.director);
    handleStoryboardStateChange({ layout: { director } });
  }

  const addStoryboardAssetToWorkflow = useCallback(asset => {
    const kind = String(asset?.kind || '').trim();
    if (!['image', 'video', 'audio'].includes(kind)) return;
    const nodeId = `${kind}-${Date.now()}-storyboard`;
    const currentCanvas = canvasStateRef.current;
    const index = Array.isArray(currentCanvas.nodes) ? currentCanvas.nodes.length : 0;
    const nextNode = createNodeAt({
      id: nodeId,
      kind,
      position: { x: 100 + (index % 3) * 320, y: 100 + Math.floor(index / 3) * 250 },
      asset
    });
    applyCanvasEdit({
      ...currentCanvas,
      nodes: [...(currentCanvas.nodes || []).map(node => ({ ...node, selected: false })), nextNode]
    });
    pendingWorkflowNodeRef.current = nodeId;
    handleStoryboardStateChange({ selectedKey: `node:${nodeId}` });
    openNodeInWorkflow(nodeId);
    message.success('素材已加入工作流');
  }, [applyCanvasEdit, handleStoryboardStateChange]);

  const addStoryboardAssetToConversation = useCallback(asset => {
    const id = String(asset?.id || asset?.assetId || '').trim();
    if (!id) return;
    setSelectedConversationAssetIds(current => current.includes(id) ? current : [...current, id]);
  }, []);

  const handleWorkspaceAssetsCreated = useCallback(created => {
    if (!Array.isArray(created) || !created.length) return;
    setWorkspaceAssets(current => {
      const byID = new Map(current.map(asset => [asset.id, asset]));
      for (const asset of created) if (asset?.id) byID.set(asset.id, asset);
      return [...byID.values()];
    });
  }, []);

  const removeStoryboardConversationAsset = useCallback(assetId => {
    setSelectedConversationAssetIds(current => current.filter(id => id !== assetId));
  }, []);

  function handleConversationAssetDragOver(event) {
    if (!hasStoryboardAssetDrop(event.dataTransfer)) return;
    event.preventDefault();
    event.dataTransfer.dropEffect = 'copy';
    setConversationAssetDropActive(true);
  }

  function handleConversationAssetDrop(event) {
    event.preventDefault();
    setConversationAssetDropActive(false);
    const asset = readStoryboardAssetDrop(event.dataTransfer, workspaceAssets);
    if (!asset) return;
    addStoryboardAssetToConversation(asset);
    questionInputRef.current?.focus?.();
    message.success('素材已加入本次对话');
  }

  function handleConversationAssetDragLeave(event) {
    if (event.currentTarget.contains(event.relatedTarget)) return;
    setConversationAssetDropActive(false);
  }

  function handleConversationStarter(prompt) {
    setQuestion(prompt);
    requestAnimationFrame(() => questionInputRef.current?.focus?.());
  }

  useEffect(() => {
    if (viewState.activeView !== 'workflow' || !pendingWorkflowNodeRef.current) return undefined;
    const nodeId = pendingWorkflowNodeRef.current;
    const timer = window.setTimeout(() => {
      canvasRef.current?.locateNode(nodeId);
      pendingWorkflowNodeRef.current = '';
    }, 0);
    return () => window.clearTimeout(timer);
  }, [viewState.activeView, activeTaskId, workflowFocusRequest]);

  function activateTask(task, { historyMode = 'replace' } = {}) {
    activeTaskIdRef.current = task?.id || null;
    setActiveTaskId(task?.id || null);
    setActiveTask(task || null);
    setDismissedDirectorPlanIDs([]);
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

  async function createTask(expectedRequestId = taskRequestRef.current, { historyMode = 'push', projectType = 'video' } = {}) {
    try {
      const result = await createAgentTask({ projectType });
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
          params.delete('task');
          params.set('taskId', legacyTaskId);
          params.set('view', viewStateRef.current.activeView);
          const nextUrl = `/agent/canvas?${params.toString()}`;
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
      const searchParams = new URLSearchParams(window.location.search);
      if (nextIsCanvasWorkspace) {
        const requestedView = searchParams.get('view');
        const nextViewState = requestedView
          ? workspaceViewForActiveView(viewStateRef.current, workspaceViewQuery(requestedView))
          : workspaceViewPatch(viewStateRef.current, { activeView: workspaceViewQuery(requestedView) });
        viewStateRef.current = nextViewState;
        setViewState(nextViewState);
        replaceCanvas({ ...canvasStateRef.current, viewState: nextViewState });
      }
      setIsCanvasWorkspace(nextIsCanvasWorkspace);
      setRequestedBatchBookId(nextIsCanvasWorkspace ? searchParams.get('bookId') || '' : '');
      const requestedTaskId = nextIsCanvasWorkspace ? searchParams.get('taskId') : null;
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
    if (!composerMenuOpen && !composerPanel) return undefined;
    function closeComposerOnOutsidePointer(event) {
      if (event.target?.closest?.('.agent-workbench-composer')) return;
      setComposerMenuOpen(false);
      setComposerPanel(null);
    }
    window.addEventListener('pointerdown', closeComposerOnOutsidePointer);
    return () => window.removeEventListener('pointerdown', closeComposerOnOutsidePointer);
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
      replaceCanvas({ version: 1, revision: 0, nodes: [], edges: [] });
      setCanvasSaveStatus('idle');
      setCanvasViewport({ x: 0, y: 0, zoom: 1 });
      setExecutions([]);
      setGenerations([]);
      setGenerationSummary(null);
      setStoryboardShots([]);
      setSelectedCanvasNodes([]);
      return undefined;
    }
    const requestedBookId = String(requestedBatchBookId || '').trim();
    const canvasRequest = syncBookThenLoadCanvas({
      taskId: activeTaskId,
      bookId: requestedBookId,
      sync: syncAgentBatchFactoryBook,
      getCanvas: getAgentCanvas
    });
    Promise.all([canvasRequest, listAgentExecutions(activeTaskId)])
      .then(([canvasLoad, executionResult]) => {
        if (cancelled) return;
        const canvasResult = canvasLoad.result;
        const loadedCanvas = canvasResult.canvas || { version: 1, revision: 0, nodes: [], edges: [] };
        const loadedViewState = normalizeWorkspaceView(loadedCanvas.viewState);
        const queryView = new URLSearchParams(window.location.search).get('view');
        const nextViewState = queryView
          ? workspaceViewForActiveView(loadedViewState, workspaceViewQuery(queryView))
          : loadedViewState;
        viewStateRef.current = nextViewState;
        setViewState(nextViewState);
        syncWorkspaceViewUrl(nextViewState.activeView);
        saveQueue = createCanvasSaveQueue({
          revision: loadedCanvas.revision,
          save: (nextCanvas, revision) => saveAgentCanvas(activeTaskId, nextCanvas, revision),
          reload: () => getAgentCanvas(activeTaskId),
          onCanvas: replaceCanvas,
          onStatus: setCanvasSaveStatus,
          onSynced: savedCanvas => {
            const draft = loadCanvasDraft(window.localStorage, getCurrentUsername(), activeTaskId);
            if (draft && Number(draft.baseRevision) <= Number(savedCanvas?.revision)) {
              clearCanvasDraft(window.localStorage, getCurrentUsername(), activeTaskId);
            }
          }
        });
        canvasSaveQueueRef.current = saveQueue;
        const serverCanvas = { ...loadedCanvas, viewState: nextViewState };
        const draft = loadCanvasDraft(window.localStorage, getCurrentUsername(), activeTaskId);
        const draftIsCurrent = draft && isDraftAtLeastAsFresh(draft, loadedCanvas);
        // A Batch Factory book can gain projected nodes while this browser is
        // closed. An old empty local draft must never hide or overwrite them.
        const staleEmptyBatchDraft = Boolean(draftIsCurrent && requestedBookId && serverCanvas.nodes?.length && !draft.canvas?.nodes?.length);
        const initialCanvas = draftIsCurrent && !staleEmptyBatchDraft
          ? { ...mergePendingCanvas(serverCanvas, draft.canvas), viewState: nextViewState }
          : serverCanvas;
        if (draft && (!draftIsCurrent || staleEmptyBatchDraft)) {
          clearCanvasDraft(window.localStorage, getCurrentUsername(), activeTaskId);
        }
        replaceCanvas(initialCanvas);
        setCanvasViewport(initialCanvas.viewport || { x: 0, y: 0, zoom: 1 });
        setCanvasSaveStatus(draftIsCurrent && !staleEmptyBatchDraft ? 'dirty' : 'idle');
        if (draftIsCurrent && !staleEmptyBatchDraft) saveQueue.enqueue(initialCanvas);
        setExecutions(Array.isArray(executionResult.executions) ? executionResult.executions : []);
      })
      .catch(error => { if (!cancelled) message.error(error.message || '读取创作画布失败'); });
    return () => {
      cancelled = true;
      void saveQueue?.flush();
      saveQueue?.dispose();
      if (canvasSaveQueueRef.current === saveQueue) canvasSaveQueueRef.current = null;
    };
  }, [activeTaskId, isCanvasWorkspace, requestedBatchBookId]);

  useEffect(() => {
    let cancelled = false;
    let timer = null;
    if (!isCanvasWorkspace || !activeTaskId) {
      setGenerations([]);
      return undefined;
    }
    const refresh = () => Promise.all([listAgentGenerations(activeTaskId), getAgentGenerationSummary(activeTaskId)])
      .then(([generationResult, summaryResult]) => {
        if (cancelled) return;
        setGenerations(Array.isArray(generationResult.tasks) ? generationResult.tasks : []);
        setGenerationSummary(summaryResult?.summary || null);
      })
      .catch(() => {});
    void refresh();
    timer = window.setInterval(refresh, 5000);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [activeTaskId, isCanvasWorkspace]);

  const scriptNodeIDs = useMemo(() => (Array.isArray(canvas.nodes) ? canvas.nodes : [])
    .filter(node => String(node?.type ?? node?.data?.type ?? '') === 'script')
    .map(node => String(node.id || '').trim())
    .filter(Boolean), [canvas.nodes]);
  const scriptNodeKey = scriptNodeIDs.join('\u0000');

  useEffect(() => {
    let cancelled = false;
    let timer = null;
    if (!isCanvasWorkspace || !activeTaskId || !scriptNodeIDs.length) {
      setStoryboardShots([]);
      setStoryboardShotsLoading(false);
      setStoryboardShotsError('');
      return undefined;
    }
    setStoryboardShotsLoading(true);
    setStoryboardShotsError('');
    const refresh = () => Promise.all(scriptNodeIDs.map(nodeId => listAgentScriptShots(activeTaskId, nodeId)))
      .then(results => {
        if (cancelled) return;
        setStoryboardShots(results.flatMap(result => Array.isArray(result?.shots) ? result.shots : []));
        setStoryboardShotsError('');
      })
      .catch(error => {
        if (!cancelled) setStoryboardShotsError(error?.message || '读取镜头失败');
      })
      .finally(() => { if (!cancelled) setStoryboardShotsLoading(false); });
    void refresh();
    timer = window.setInterval(refresh, 5000);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [activeTaskId, isCanvasWorkspace, scriptNodeKey]);

  useEffect(() => {
    if (!isCanvasWorkspace || !activeTaskId) return undefined;
    const preserveAndSync = () => {
      persistCanvasDraft(canvasStateRef.current, activeTaskId);
      const queue = canvasSaveQueueRef.current;
      void queue?.retry();
      void queue?.flush();
    };
    window.addEventListener('pagehide', preserveAndSync);
    window.addEventListener('online', preserveAndSync);
    window.addEventListener('focus', preserveAndSync);
    return () => {
      window.removeEventListener('pagehide', preserveAndSync);
      window.removeEventListener('online', preserveAndSync);
      window.removeEventListener('focus', preserveAndSync);
    };
  }, [activeTaskId, isCanvasWorkspace]);

  useEffect(() => {
    const requestedBookId = String(requestedBatchBookId || '').trim();
    if (!isCanvasWorkspace || !activeTaskId || !requestedBookId) return undefined;
    let cancelled = false;
    let refreshing = false;
    const refreshMappedBook = async () => {
      if (cancelled || refreshing || document.visibilityState !== 'visible') return;
      const requestedTaskId = activeTaskId;
      const requestedQueue = canvasSaveQueueRef.current;
      if (!requestedQueue) return;
      const isCurrentRefresh = () => shouldAcceptCanvasRefresh({
        requestedTaskId,
        currentTaskId: activeTaskIdRef.current,
        requestedQueue,
        currentQueue: canvasSaveQueueRef.current,
        mounted: mountedRef.current
      });
      refreshing = true;
      try {
        const result = await refreshBatchFactoryBookCanvas({
          taskId: requestedTaskId,
          bookId: requestedBookId,
          queue: requestedQueue,
          sync: syncAgentBatchFactoryBook,
          getCanvas: getAgentCanvas,
          isCurrent: isCurrentRefresh
        });
        if (result?.ignored || !isCurrentRefresh()) return;
        const nextCanvas = result?.canvas || { version: 1, revision: 0, nodes: [], edges: [] };
        const accepted = await requestedQueue.acceptServerCanvas(nextCanvas);
        if (accepted === false || !isCurrentRefresh()) return;
        setCanvasViewport(nextCanvas.viewport || { x: 0, y: 0, zoom: 1 });
      } catch {
        // A background refresh should not interrupt active editing. The next
        // return to this tab or an explicit refresh can try again.
      } finally {
        refreshing = false;
      }
    };
    const refreshWhenVisible = () => {
      if (document.visibilityState === 'visible') void refreshMappedBook();
    };
    window.addEventListener('focus', refreshWhenVisible);
    document.addEventListener('visibilitychange', refreshWhenVisible);
    return () => {
      cancelled = true;
      window.removeEventListener('focus', refreshWhenVisible);
      document.removeEventListener('visibilitychange', refreshWhenVisible);
    };
  }, [activeTaskId, isCanvasWorkspace, requestedBatchBookId]);

  useEffect(() => {
    let cancelled = false;
    if (!isCanvasWorkspace || !activeTaskId) {
      setWorkspaceAssets([]);
      setWorkspaceAssetsLoading(false);
      setWorkspaceAssetsError('');
      return undefined;
    }
    setWorkspaceAssetsLoading(true);
    setWorkspaceAssetsError('');
    listAgentAssets(activeTaskId)
      .then(result => {
        if (!cancelled) setWorkspaceAssets(Array.isArray(result.assets) ? result.assets : []);
      })
      .catch(error => {
        if (cancelled) return;
        setWorkspaceAssets([]);
        setWorkspaceAssetsError(error.message || '读取项目素材失败');
      })
      .finally(() => { if (!cancelled) setWorkspaceAssetsLoading(false); });
    return () => { cancelled = true; };
  }, [activeTaskId, isCanvasWorkspace]);

  const visibleSkills = useMemo(() => skills.filter(skill => source === 'all' || skill.source === source), [skills, source]);
  const visibleProjects = useMemo(() => projectFilter === 'all'
    ? tasks
    : tasks.filter(task => normalizeAgentProjectType(task.projectType) === projectFilter), [projectFilter, tasks]);
  const activeProjectType = normalizeAgentProjectType(activeTask?.projectType);
  const activeProjectCapabilities = agentProjectCapabilities(activeProjectType);
  const selectedSkills = skills.filter(skill => selectedSkillIds.includes(skill.id));
  const messages = activeTask?.messages || [];
  const selectedNodeContext = buildSelectedNodeContext(canvas, selectedCanvasNodes);
  const selectedAssetContext = buildStoryboardAssetContext(workspaceAssets, selectedConversationAssetIds);
  const selectedConversationAssets = useMemo(() => {
    const assetsByID = new Map(workspaceAssets.map(asset => [String(asset?.id || asset?.assetId || '').trim(), asset]));
    return selectedAssetContext.map(asset => ({ ...asset, url: String(assetsByID.get(asset.id)?.url || '') }));
  }, [selectedAssetContext, workspaceAssets]);
  const nodeMentionCandidates = useMemo(() => findNodeMentionCandidates(canvas.nodes, question), [canvas.nodes, question]);
  function selectNodeMention(node) {
    setSelectedCanvasNodes(current => current.includes(node.id) ? current : [...current, node.id]);
    setQuestion(current => insertNodeMention(current, node));
    questionInputRef.current?.focus();
  }
  const nodeMentionPicker = nodeMentionCandidates.length ? <div className="agent-node-mention-picker" role="listbox" aria-label="引用画布节点">
    {nodeMentionCandidates.map(node => <button type="button" role="option" key={node.id} onMouseDown={event => event.preventDefault()} onClick={() => selectNodeMention(node)}>
      <span>{node.type === 'image' ? '图片' : node.type === 'video' ? '视频' : node.type === 'audio' ? '音频' : '文本'}</span><strong>{node.title || node.data?.title}</strong><small>{node.summary || node.data?.summary || node.status || node.data?.status || '画布节点'}</small>
    </button>)}
  </div> : null;
  const directorPlans = useMemo(() => messages.flatMap(entry => {
    const plan = entry?.context?.directorPlan;
    return plan && typeof plan === 'object' && !dismissedDirectorPlanIDs.includes(String(plan.id || '')) ? [plan] : [];
  }), [dismissedDirectorPlanIDs, messages]);
  const canvasGroups = useMemo(() => [...new Set((canvas.nodes || []).map(node => node.groupId).filter(Boolean))]
    .map(groupId => ({ id: groupId, title: batchFactoryGroupLabel(groupId, canvas.nodes) })), [canvas.nodes]);
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
    const classification = classifyComposerAttachment(file);
    if (classification.kind === 'unsupported') {
      message.error('仅支持 TXT、MD、CSV、JSON、图片、视频或音频文件。');
      return;
    }
    if (classification.kind === 'too-large') {
      message.error(classification.limit > 1024 * 1024 ? '媒体附件不能超过 128 MB。' : '文本附件不能超过 200 KB。');
      return;
    }
    if (classification.kind === 'media') {
      setAttachedFile({ kind: 'media', name: file.name, file });
      return;
    }
    const reader = new FileReader();
    reader.onload = () => setAttachedFile({ kind: 'text', name: file.name, content: String(reader.result || '').slice(0, 20000) });
    reader.onerror = () => message.error('读取附件失败');
    reader.readAsText(file);
  }

  async function sendQuestion({ problemSearch = problemSearchMode, promptOverride = '', directorFeedback = null } = {}) {
    const rawPrompt = String(promptOverride || question).trim();
    if (!rawPrompt || asking || taskDetailLoadingRef.current) return;
    const prompt = problemSearch ? createProblemSearchPrompt(rawPrompt) : rawPrompt;
    let requestId = taskRequestRef.current;
    let taskId = activeTask?.id || null;

    function isCurrentTaskRequest() {
      return mountedRef.current && taskRequestRef.current === requestId && activeTaskIdRef.current === taskId;
    }

    setAsking(true);
    if (!promptOverride) setQuestion('');
    const isDirectorRequest = viewStateRef.current.activeView === 'storyboard';
    const directorMode = isDirectorRequest
      ? detectDirectorMode(rawPrompt, viewStateRef.current.storyboard.directorMode)
      : 'supervised';
    if (isDirectorRequest && directorMode !== viewStateRef.current.storyboard.directorMode) {
      handleStoryboardStateChange({ directorMode });
    }
    dispatchPetState('working', { title: 'CM 正在处理任务' });
    try {
      const createdTask = activeTask ? { task: activeTask, requestId } : await createTask(requestId, { projectType: newProjectType });
      if (!createdTask) return;
      const task = createdTask.task;
      taskId = task.id;
      requestId = createdTask.requestId;
      let assetContextForRequest = selectedAssetContext;
      if (attachedFile?.kind === 'media') {
        const uploaded = await uploadAgentAsset(task.id, attachedFile.file);
        const asset = uploaded?.asset;
        if (!asset?.id) throw new Error('媒体附件上传没有返回素材');
        handleWorkspaceAssetsCreated([asset]);
        const nextAssetIDs = appendUniqueAssetID(selectedConversationAssetIds, asset.id);
        setSelectedConversationAssetIds(nextAssetIDs);
        assetContextForRequest = buildStoryboardAssetContext([...workspaceAssets, asset], nextAssetIDs);
      }
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
          selectedAssets: assetContextForRequest,
          director: isDirectorRequest ? {
            enabled: true,
            mode: directorMode,
            ...(directorFeedback ? { feedbackToPlanId: directorFeedback.planId } : {})
          } : undefined,
          attachment: attachedFile?.kind === 'text' ? { name: attachedFile.name, content: attachedFile.content } : undefined
        },
        skillIds: selectedSkillIds
      });
      const nextTask = result.task || { ...task, messages: [...(task.messages || []), result.user, result.assistant] };
      if (!isCurrentTaskRequest()) return;
      setActiveTask(nextTask);
      if (result.canvas) {
        const accepted = await canvasSaveQueueRef.current?.acceptServerCanvas(result.canvas);
        if (accepted !== false && isCurrentTaskRequest() && !canvasSaveQueueRef.current) replaceCanvas(result.canvas);
        if (accepted !== false && isCurrentTaskRequest()) setCanvasViewport(result.canvas.viewport || { x: 0, y: 0, zoom: 1 });
      }
      setSelectedConversationAssetIds([]);
      setAttachedFile(null);
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

  function returnToAgentHome() {
    activateTask(null, { historyMode: 'push' });
  }

  function showAllAgentProjects() {
    updateWorkspaceView({ drawerOpen: true, drawerTab: 'nodes' });
  }

  function confirmDeleteCurrentTask() {
    if (!activeTaskId) return;
    Modal.confirm({
      title: '删除当前项目？',
      content: '删除后无法恢复，不会影响其他项目。',
      okText: '删除',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: deleteCurrentTask
    });
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

  async function approveDirectorPlan(planId) {
    const taskId = activeTaskIdRef.current;
    const revision = Number(canvas.revision) || 0;
    if (!taskId || !planId || asking) return;
    const requestId = taskRequestRef.current;
    const isCurrentTask = () => mountedRef.current && taskRequestRef.current === requestId && activeTaskIdRef.current === taskId;
    setAsking(true);
    try {
      const result = await approveAgentDirectorPlan(taskId, planId, revision);
      if (!isCurrentTask()) return;
      const nextCanvas = result?.canvas;
      if (!nextCanvas) throw new Error('导演执行后没有返回画布');
      const accepted = await canvasSaveQueueRef.current?.acceptServerCanvas(nextCanvas);
      if (accepted === false || !isCurrentTask()) return;
      if (!canvasSaveQueueRef.current) replaceCanvas(nextCanvas);
      setCanvasViewport(nextCanvas.viewport || { x: 0, y: 0, zoom: 1 });
      const [taskResult, executionResult] = await Promise.all([getAgentTask(taskId), listAgentExecutions(taskId)]);
      if (!isCurrentTask()) return;
      setActiveTask(taskResult.task || null);
      setExecutions(Array.isArray(executionResult.executions) ? executionResult.executions : []);
      setDismissedDirectorPlanIDs(current => [...new Set([...current, String(planId)])]);
      await refreshTasks(isCurrentTask);
      if (isCurrentTask()) message.success('导演方案已应用到画布');
    } catch (error) {
      if (isCurrentTask()) message.error(error.message || '导演方案未能执行');
    } finally {
      if (isCurrentTask()) setAsking(false);
    }
  }

  return (
    <div className={`agent-page utility-workbench agent-workspace-v3${isCanvasWorkspace && !viewState.drawerOpen ? ' agent-task-drawer-collapsed' : ''}`}>
      {isCanvasWorkspace ? <div
        className={`${viewState.conversationOpen ? 'agent-canvas-workspace' : 'agent-canvas-workspace is-conversation-collapsed'}${viewState.activeView === 'storyboard' ? ' is-storyboard' : ' is-workflow'}${storyboardDirectorResizing ? ' is-resizing-director' : ''}`}
        style={viewState.activeView === 'storyboard' ? { '--storyboard-director-split': `${viewState.storyboard.layout.director}%` } : undefined}
      >
        <AgentWorkspaceHeader
          projectTitle={activeTask?.title}
          activeView={viewState.activeView}
          saveStatus={canvasSaveStatus}
          nodeCount={canvas.nodes.length}
          materialCount={workspaceAssets.length}
          zoom={Math.round((canvasViewport.zoom || 1) * 100)}
          isRunning={asking}
          onRename={openRenameTask}
          onBackHome={returnToAgentHome}
          onShowProjects={showAllAgentProjects}
          onCreateProject={() => createTask(taskRequestRef.current, { projectType: newProjectType })}
          onDeleteProject={confirmDeleteCurrentTask}
          onViewChange={activeView => {
            const nextView = workspaceViewForActiveView(viewStateRef.current, activeView);
            updateWorkspaceView({ activeView: nextView.activeView, drawerOpen: nextView.drawerOpen, conversationOpen: nextView.conversationOpen });
          }}
          onRetrySync={() => void canvasSaveQueueRef.current?.retry()}
        />

        <AgentWorkspaceDrawer
          open={viewState.drawerOpen}
          width={viewState.drawerWidth}
          activeTab={viewState.drawerTab}
          nodes={canvas.nodes}
          groups={canvasGroups}
          assets={workspaceAssets}
          executions={executions}
          shots={storyboardShots}
          generations={generations}
          generationSummary={generationSummary}
          assetsLoading={workspaceAssetsLoading}
          assetsError={workspaceAssetsError}
          tasks={tasks}
          activeTaskId={activeTaskId}
          onTabChange={drawerTab => updateWorkspaceView({ drawerTab })}
          onWidthChange={drawerWidth => updateWorkspaceView({ drawerWidth })}
          onLocateNode={openNodeInWorkflow}
          onOpenDirectorDesk={nodeId => canvasRef.current?.openDirectorDesk(nodeId)}
          onActivateGenerationResult={handleGenerationResultActivate}
          onCreateFromAsset={asset => canvasRef.current?.createFromAsset(asset)}
          onSelectTask={taskId => selectTask(taskId, { historyMode: 'push' })}
          onNewProject={() => createTask(taskRequestRef.current, { projectType: newProjectType })}
          onClose={() => updateWorkspaceView({ drawerOpen: false })}
          footer={activeTask ? <><Button type="text" icon={<RotateCcw size={15} />} disabled={messages.length === 0} onClick={clearCurrentTask}>清空对话</Button><Popconfirm title="删除当前项目？" description="删除后无法恢复，不会影响其他项目。" okText="删除" cancelText="取消" okButtonProps={{ danger: true }} onConfirm={deleteCurrentTask}><Button type="text" danger icon={<Trash2 size={15} />}>删除项目</Button></Popconfirm></> : null}
        />

        {viewState.activeView === 'workflow' ? <main className="agent-canvas-main" aria-label="创作画布">
          <AgentCanvas
            ref={canvasRef}
            key={activeTaskId}
            taskId={activeTaskId}
            projectType={activeProjectType}
            projectCapabilities={activeProjectCapabilities}
            requestedBookId={requestedBatchBookId}
            canvas={canvas}
            executions={executions}
            scriptShots={storyboardShots}
            onScriptShotsChange={handleSharedScriptShotsChange}
            workspaceView={viewState}
            onWorkspaceViewChange={updateWorkspaceView}
            conversationOpen={viewState.conversationOpen}
            onToggleConversation={() => updateWorkspaceView({ conversationOpen: !viewStateRef.current.conversationOpen })}
            onSelectionChange={setSelectedCanvasNodes}
            onViewportChange={setCanvasViewport}
            onViewportCommit={nextViewport => {
              const nextCanvas = { ...canvasStateRef.current, viewport: nextViewport };
              replaceCanvas(nextCanvas);
              persistCanvasDraft(nextCanvas);
            }}
            onFlushCanvas={() => {
              const queue = canvasSaveQueueRef.current;
              if (!queue) throw new Error('画布保存队列尚未准备完成');
              return queue.flushRequired();
            }}
            onExecutionChange={execution => {
              if (!execution?.id) return;
              setExecutions(current => [execution, ...current.filter(item => item.id !== execution.id)]);
            }}
            onAssetsCreated={handleWorkspaceAssetsCreated}
            onChange={applyCanvasEdit}
            onServerCanvasChange={async () => {
              const requestedTaskId = activeTaskId;
              const requestedQueue = canvasSaveQueueRef.current;
              if (!requestedTaskId) return;
              const result = await getAgentCanvas(requestedTaskId);
              const isCurrentRefresh = () => shouldAcceptCanvasRefresh({
                requestedTaskId,
                currentTaskId: activeTaskIdRef.current,
                requestedQueue,
                currentQueue: canvasSaveQueueRef.current,
                mounted: mountedRef.current
              });
              if (!isCurrentRefresh()) return;
              const nextCanvas = result.canvas || { version: 1, revision: 0, nodes: [], edges: [] };
              if (requestedQueue) {
                const accepted = await requestedQueue.acceptServerCanvas(nextCanvas);
                if (!accepted || !isCurrentRefresh()) return;
              }
              else replaceCanvas(nextCanvas);
              setCanvasViewport(nextCanvas.viewport || { x: 0, y: 0, zoom: 1 });
            }}
          />
        </main> : <AgentStoryboardWorkspace
          taskId={activeTaskId}
          nodes={canvas.nodes}
          assets={workspaceAssets}
          shots={storyboardShots}
          generations={generations}
          generationSummary={generationSummary}
          loading={workspaceAssetsLoading}
          error={workspaceAssetsError}
          shotsLoading={storyboardShotsLoading}
          shotsError={storyboardShotsError}
          conversationAssetIds={selectedConversationAssetIds}
          storyboardState={viewState.storyboard}
          onStoryboardStateChange={handleStoryboardStateChange}
          onSelectionChange={handleStoryboardSelectionChange}
          onNodeConfigChange={handleStoryboardNodeConfigChange}
          onShotUpdated={handleStoryboardShotUpdated}
          onCreateShotBatch={handleStoryboardBatchCreate}
           onAddToConversation={addStoryboardAssetToConversation}
           onAddToWorkflow={addStoryboardAssetToWorkflow}
           onAssetsCreated={handleWorkspaceAssetsCreated}
          onOpenNodeInWorkflow={openNodeInWorkflow}
        />}

        <aside className="agent-conversation-panel" aria-label="项目对话">
          {viewState.activeView !== 'storyboard' ? <header className="agent-conversation-header">
            <span><MessageCircle size={16} /></span>
            <div><strong>项目对话</strong><small>{activeTask?.title || activeTask?.name || '新对话'}</small></div>
          </header> : null}
          {viewState.activeView === 'storyboard' ? <AgentDirectorConversation
            mode={viewState.storyboard.directorMode}
            plans={directorPlans}
            busy={asking}
            onModeChange={directorMode => handleStoryboardStateChange({ directorMode })}
            onApprove={approveDirectorPlan}
            onReject={(planId, feedback) => sendQuestion({
              promptOverride: `请根据我对方案「${planId}」的反馈重新整理一个待确认的画布方案：${feedback}`,
              directorFeedback: { planId }
            }).then(() => setDismissedDirectorPlanIDs(current => [...new Set([...current, String(planId)])]))}
          /> : null}
          <div ref={historyRef} className="agent-workbench-history" aria-live="polite">
            {taskDetailLoading ? <div className="agent-workbench-loading" role="status">正在读取任务详情...</div> : null}
            {!loading && !activeTask ? <div className="agent-workbench-empty">{invalidCanvasTask ? <><strong>找不到这个任务</strong><span>任务可能已删除或链接已失效。<a href="/agent">返回创作首页</a></span></> : <><SuperOpcLiquidMetalLogo theme={theme} className="agent-workbench-logo" /><strong>一战晟铭 / CM</strong><span>你的 AI 创作工作空间</span></>}</div> : null}
            {activeTask && messages.length === 0 ? <div className="agent-workbench-empty agent-workbench-empty--conversation"><SuperOpcLiquidMetalLogo theme={theme} className="agent-workbench-logo" /><strong>{viewState.activeView === 'storyboard' ? '让创作导演协助你的灵感' : '从一个创作想法开始'}</strong><span>{viewState.activeView === 'storyboard' ? '选择一个方向，导演会先整理为可确认的步骤。' : '选择一个方向，或直接在下方输入你的创作需求。'}</span><div className="agent-conversation-starters" aria-label="常用创作入口">
              <button className="agent-conversation-starter" type="button" disabled={taskDetailLoading} onClick={() => handleConversationStarter('请从当前画布开始，帮我梳理下一步创作方向和需要补齐的内容。')}><Sparkles size={16} /><span>从画布开始创作</span><ArrowRight size={15} /></button>
              <button className="agent-conversation-starter" type="button" disabled={taskDetailLoading} onClick={() => handleConversationStarter('请基于当前项目素材，帮我生成一份适合继续创作的脚本构思。')}><ImageIcon size={16} /><span>从素材预设开始脚本原创</span><ArrowRight size={15} /></button>
              <button className="agent-conversation-starter" type="button" disabled={taskDetailLoading} onClick={() => handleConversationStarter('我想上传一段故事，请帮我把它改编成可执行的画布和分镜计划。')}><FileText size={16} /><span>上传故事来改编</span><ArrowRight size={15} /></button>
              <button className="agent-conversation-starter" type="button" disabled={taskDetailLoading} onClick={() => handleConversationStarter('请帮我批量优化当前项目的提示词，让画面、角色和镜头描述更清晰一致。')}><FilePenLine size={16} /><span>批量优化提示词</span><ArrowRight size={15} /></button>
            </div></div> : null}
            {messages.map((entry, index) => <article key={`${entry.createdAt || index}-${entry.role}`} className={`agent-workbench-message agent-workbench-message--${entry.role}`}><strong>{entry.role === 'assistant' ? (viewState.activeView === 'storyboard' ? '导演' : 'CM') : '你'}</strong><span>{entry.content}</span></article>)}
          </div>
          {executions.length ? <section className="agent-execution-records" aria-label="执行记录"><strong>执行记录</strong>{executions.slice(0, 4).map((execution, index) => <span key={execution.id || index}>{execution.toolName || execution.tool || execution.name || '工具'} · {execution.status || '已完成'}</span>)}</section> : null}
          {selectedNodeContext.length ? <div className="agent-selected-node-references" aria-label="已选节点引用">{selectedNodeContext.map(node => <Tag key={node.id}>{node.title}</Tag>)}</div> : null}
        <div
          className={`agent-workbench-composer${problemSearchMode ? ' is-problem-search' : ''}${viewState.activeView === 'storyboard' ? ' is-director' : ''}${conversationAssetDropActive ? ' is-asset-drop-target' : ''}`}
          onDragOver={handleConversationAssetDragOver}
          onDragEnter={handleConversationAssetDragOver}
          onDragLeave={handleConversationAssetDragLeave}
          onDrop={handleConversationAssetDrop}
        >
          <div className="agent-composer-overlay" aria-live="polite">
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
          <div className="agent-composer-body">
          <input ref={attachmentInputRef} className="agent-attachment-input" type="file" accept=".txt,.md,.csv,.json,text/plain,text/markdown,text/csv,application/json,image/*,video/*,audio/*" onChange={readAttachment} />
          {selectedConversationAssets.length ? <div className="agent-composer-asset-attachments" aria-label="已加入对话的素材">
            {selectedConversationAssets.map(asset => <div className="agent-composer-asset-attachment" key={asset.id} title={asset.name}>
              <span className={`agent-composer-asset-attachment__preview is-${asset.kind || 'file'}`}><ConversationAssetPreview asset={asset} /></span>
              <span className="agent-composer-asset-attachment__name">{asset.name}</span>
              <button type="button" className="agent-composer-asset-attachment__remove" aria-label={`移除素材 ${asset.name}`} onClick={() => removeStoryboardConversationAsset(asset.id)}><X size={13} /></button>
            </div>)}
          </div> : null}
          <div className="agent-selected-skills">
            {problemSearchMode ? <Tag closable onClose={() => setProblemSearchMode(false)}>找问题模式</Tag> : null}
            {selectedSkills.map(skill => <Tag closable key={skill.id} onClose={() => toggleSkill(skill.id)}>{skill.name}</Tag>)}
            {attachedFile ? <Tag closable onClose={() => setAttachedFile(null)}>{attachedFile.kind === 'media' ? '媒体' : '附件'}：{attachedFile.name}</Tag> : null}
          </div>
          <Input.TextArea
            ref={questionInputRef}
            value={question}
            disabled={taskDetailLoading}
            autoSize={{ minRows: 2, maxRows: 8 }}
            placeholder={taskDetailLoading ? '正在读取任务详情...' : problemSearchMode ? '粘贴需要检查的文字，然后点击「找问题」…' : viewState.activeView === 'storyboard' ? '告诉导演你想完成什么，或引用项目素材…' : '今天想让 CM 帮你做什么？'}
            onChange={event => setQuestion(event.target.value)}
            onPressEnter={event => {
              if (event.nativeEvent.isComposing) return;
              if (!event.shiftKey) {
                event.preventDefault();
                sendQuestion();
              }
            }}
          />
          {nodeMentionPicker}
          <div className="agent-composer-actions">
            <div className="agent-composer-tools">
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
                    {viewState.activeView !== 'storyboard' ? <><button role="menuitem" type="button" onClick={() => openComposerTool('expert')}>◉ <span>专家</span><i>›</i></button>
                    <button role="menuitem" type="button" onClick={() => openComposerTool('skill')}>⌘ <span>技能</span><i>›</i></button>
                    <button role="menuitem" type="button" onClick={() => { setProblemSearchMode(active => !active); setComposerMenuOpen(false); }}><SearchCheck size={14} /> <span>{problemSearchMode ? '退出找问题' : '找问题'}</span><i>›</i></button></> : null}
                    <button role="menuitem" type="button" onClick={() => openComposerTool('connector')}>⌁ <span>连接器</span><i>›</i></button>
                  </div>
                ) : null}
              </div>
              <button className="agent-composer-mode" type="button" onClick={() => openComposerTool('mode')}>{selectedMode}<ChevronDown size={13} /></button>
              <button className="agent-composer-tool" type="button" aria-label="添加附件" title="添加附件" onClick={() => openComposerTool('file')}><Paperclip size={16} /></button>
              {viewState.activeView !== 'storyboard' ? <><button className="agent-composer-tool" type="button" aria-label="选择技能" title="选择技能" onClick={() => openComposerTool('skill')}><Sparkles size={16} /></button>
              <button className="agent-composer-tool" type="button" aria-label="选择专家" title="选择专家" onClick={() => openComposerTool('expert')}><Bot size={16} /></button></> : null}
            </div>
            <Button
              className="agent-composer-send"
              type="primary"
              shape="circle"
              icon={problemSearchMode ? <SearchCheck size={16} /> : <ArrowUp size={16} />}
              aria-label="发送消息"
              title={problemSearchMode ? '开始找问题' : '发送消息'}
              loading={asking}
              disabled={!question.trim() || taskDetailLoading}
              onClick={() => sendQuestion()}
            />
          </div>

          </div>
        </div>
        </aside>
        {viewState.activeView === 'storyboard' && viewState.conversationOpen ? <button
          type="button"
          className="agent-storyboard-workspace-splitter"
          aria-label="调整 AI 助手面板宽度"
          title="拖动调整宽度，双击恢复默认"
          onPointerDown={startStoryboardDirectorResize}
          onDoubleClick={resetStoryboardDirectorSplit}
          onKeyDown={adjustStoryboardDirectorSplit}
        /> : null}
      </div>

      : <section className="agent-creation-home" aria-labelledby="agent-creation-home-title">
        <div className="agent-creation-home-inner">
          <Typography.Title id="agent-creation-home-title" level={2}>开始创作</Typography.Title>
          <Typography.Paragraph>告诉 CM 你想完成什么，马上进入专属创作画布。</Typography.Paragraph>
          <div className={`agent-workbench-composer agent-creation-composer${problemSearchMode ? ' is-problem-search' : ''}`}>
            <input ref={attachmentInputRef} className="agent-attachment-input" type="file" accept=".txt,.md,.csv,.json,text/plain,text/markdown,text/csv,application/json,image/*,video/*,audio/*" onChange={readAttachment} />
            <div className="agent-selected-skills">{selectedSkills.map(skill => <Tag closable key={skill.id} onClose={() => toggleSkill(skill.id)}>{skill.name}</Tag>)}{attachedFile ? <Tag closable onClose={() => setAttachedFile(null)}>{attachedFile.kind === 'media' ? '媒体' : '附件'}：{attachedFile.name}</Tag> : null}</div>
            <Input.TextArea ref={questionInputRef} value={question} autoSize={{ minRows: 4, maxRows: 8 }} placeholder="描述你的创作需求…" onChange={event => setQuestion(event.target.value)} onPressEnter={event => { if (!event.shiftKey) { event.preventDefault(); sendQuestion(); } }} />
            {nodeMentionPicker}
            <div className="agent-creation-type-picker" aria-label="作品类型"><span>作品类型</span><Segmented value={newProjectType} onChange={setNewProjectType} options={agentProjectTypeOptions().map(option => ({ value: option.value, label: option.label }))} /></div>
            <div className="agent-composer-actions"><div className="agent-composer-actions-left"><button className="agent-context-chip" type="button" onClick={() => openComposerTool('mode')}>{selectedMode}<ChevronDown size={13} /></button><button className="agent-context-chip" type="button" onClick={() => openComposerTool('expert')}>{selectedExpert}<ChevronDown size={13} /></button><button className="agent-context-chip" type="button" onClick={() => openComposerTool('skill')}>技能{selectedSkillIds.length ? ` ${selectedSkillIds.length}` : ''}<ChevronDown size={13} /></button><button className="agent-context-chip" type="button" onClick={() => openComposerTool('file')}>添加附件</button></div><Button type="primary" icon={<Send size={15} />} loading={asking} disabled={!question.trim()} onClick={() => sendQuestion()}>开始创作</Button></div>
            {composerPanel === 'mode' ? <div className="agent-composer-context"><strong>模式</strong>{['创作助手', '分析拆解', '修改润色'].map(value => <button type="button" key={value} className={selectedMode === value ? 'active' : ''} onClick={() => { setSelectedMode(value); setComposerPanel(null); }}>{value}</button>)}</div> : null}
            {composerPanel === 'expert' ? <div className="agent-composer-context"><strong>专家</strong>{['CM 创作顾问', '前贴片广告策划', '短剧编剧'].map(value => <button type="button" key={value} className={selectedExpert === value ? 'active' : ''} onClick={() => { setSelectedExpert(value); setComposerPanel(null); }}>{value}</button>)}</div> : null}
            {composerPanel === 'skill' ? <div className="agent-skill-panel"><Segmented value={source} onChange={setSource} options={[{ label: '全部', value: 'all' }, { label: '平台自带', value: 'system' }, { label: '我的技能', value: 'user' }]} /><div className="agent-skill-list">{visibleSkills.map(skill => <article className={`agent-skill-card${selectedSkillIds.includes(skill.id) ? ' selected' : ''}`} key={skill.id}><strong>{skill.name}</strong><p>{skill.description || '未填写说明'}</p><Button size="small" type={selectedSkillIds.includes(skill.id) ? 'primary' : 'default'} onClick={() => toggleSkill(skill.id)}>{selectedSkillIds.includes(skill.id) ? '已选用' : '选用'}</Button></article>)}</div></div> : null}
          </div>
          <section className="agent-recent-projects" aria-label="最近项目"><div className="agent-project-list-head"><h3>最近项目</h3><Segmented size="small" value={projectFilter} onChange={setProjectFilter} options={[{ label: '全部', value: 'all' }, { label: '视频', value: 'video' }, { label: '漫画', value: 'comic' }]} /></div>{loading ? <span>正在读取项目…</span> : visibleProjects.slice(0, 6).map(task => <button type="button" key={task.id} onClick={() => { window.history.pushState({}, '', `/agent/canvas?taskId=${encodeURIComponent(task.id)}`); window.dispatchEvent(new PopStateEvent('popstate')); }}><strong>{task.title}<Tag>{agentProjectTypeLabel(task.projectType)}</Tag></strong><span>{task.preview || '尚未开始对话'}</span></button>)}</section>
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
