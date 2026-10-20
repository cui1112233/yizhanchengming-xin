import { Background, BaseEdge, Handle, MarkerType, Position, ReactFlow, ViewportPortal, applyEdgeChanges, applyNodeChanges, getBezierPath, useUpdateNodeInternals } from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { Button, Dropdown, Input, Modal, Popover, Segmented, Select, Switch, Tooltip } from 'antd';
import {
  AlignCenterHorizontal, AlignCenterVertical, AlignEndHorizontal, AlignEndVertical,
  AlignHorizontalDistributeCenter, AlignStartHorizontal, AlignStartVertical,
  AlignVerticalDistributeCenter, Eye, EyeOff, Focus, GitBranch, Group, Hand, Maximize2,
  ImagePlus, MousePointer2, Plus, Redo2, SlidersHorizontal, Undo2, Ungroup, ZoomIn, ZoomOut
} from 'lucide-react';
import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
import {
  alignCanvasNodes, canvasGroupBounds, canvasKeyboardCommand, commandAvailability,
  connectCanvasNodes, createCanvasNode, distributeCanvasNodes, groupCanvasNodes,
  moveCanvasGroup, selectCanvasNodes, ungroupCanvasNodes
} from '../pages/agent-canvas-commands';
import { createCanvasHistory } from '../pages/agent-canvas-history';
import {
  appearanceCssVariables, applyCanvasEdgeAppearance, canvasEdgeColorOptions, canvasEdgeWidthOptions,
  edgeStyleForAppearance, normalizeCanvasAppearance
} from '../pages/agent-canvas-appearance';
import {
  appendReferenceAsset, attachImageAsset, copyCanvasSelection, createNodeAt, pasteCanvasSelection, removeCanvasNodes,
  readableAssetRequestError, shouldOpenConnectionDrop, syncNodeReferencesFromEdges
} from '../pages/agent-canvas-context-menu';
import { cancelAgentVideoJob, confirmAgentVideoJob, createAgentImageGeneration, createAgentTextGeneration, createAgentVideoJob, getAgentVideoJob, listAgentAssets, listAgentVideoJobs, retryAgentVideoJob, uploadAgentAsset } from '../../shared/api/agent';
import { listAvailableModels, modelSelectOptions } from '../../shared/api/modelCatalog';
import { nodeGenerationIdempotencyKey } from '../pages/agent-canvas-node-generation';
import { canSubmitVideoFromConfig, makeVideoJobIdempotencyKey, mergeLatestVideoJobsByNode, shouldAcceptVideoJobResponse, videoNodeViewState, videoReferenceAssetIDs, videoReferenceCapabilityError } from '../pages/agent-canvas-video-jobs';
import { normalizeCanvasNodeType } from '../pages/agent-canvas-node-presentation';
import { AuthenticatedAgentImage } from './AuthenticatedAgentImage';
import { AuthenticatedAgentVideo } from './AuthenticatedAgentVideo';
import AgentNodeComposer from './AgentNodeComposer';

const LOCAL_DOUBAO_VIDEO_MODEL_OPTION = { value: 'local-doubao-executor-video', label: '豆包本地执行器（视频）' };

function CanvasNode({
  data, id, onManualEdit, onPickImageAsset, onPickReferenceAsset, onRemoveReference,
  onNodeConfigChange, onSubmitText, onSubmitImage, onSubmitVideo, onCancelVideo, onRetryVideo, onConfirmVideo,
  videoJob, videoState, videoBusy, videoActionError, generationState, modelOptions, modelLoadError,
  selected, isEditorOpen, onOpenEditor
}) {
  const updateNodeInternals = useUpdateNodeInternals();
  const stopCanvasEvent = event => event.stopPropagation();
  const handleNodeDoubleClick = event => {
    event.preventDefault();
    event.stopPropagation();
    if (normalizeCanvasNodeType(data.type) === 'text') onManualEdit?.(id);
  };
  const openEditor = event => {
    stopCanvasEvent(event);
    onOpenEditor?.(id);
  };
  useLayoutEffect(() => {
    updateNodeInternals(id);
  }, [id, isEditorOpen, updateNodeInternals]);
  const nodeType = normalizeCanvasNodeType(data.type);
  const composer = isEditorOpen ? <AgentNodeComposer
    node={{ id, data }}
    modelOptions={modelOptions}
    modelLoadError={modelLoadError}
    generationState={generationState}
    onConfigChange={onNodeConfigChange}
    onPickImageAsset={onPickImageAsset}
    onPickReferenceAsset={onPickReferenceAsset}
    onRemoveReference={onRemoveReference}
    onSubmit={nodeType === 'text' ? onSubmitText : nodeType === 'image' ? onSubmitImage : (nodeID, modelID) => onSubmitVideo?.(nodeID, false, modelID)}
    videoJob={videoJob}
    videoState={videoState}
    videoBusy={videoBusy}
    videoActionError={videoActionError}
    onCancelVideo={onCancelVideo}
    onRetryVideo={onRetryVideo}
    onConfirmVideo={onConfirmVideo}
  /> : null;
  return (
    <article
      className={`agent-canvas-node agent-canvas-node--${nodeType} agent-canvas-node--${data.status}${selected ? ' is-selected' : ''}`}
      onDoubleClick={handleNodeDoubleClick}
    >
      <div className="agent-canvas-node-preview-shell">
        <Handle type="target" position={Position.Left} className="agent-canvas-node-target" aria-label="连接到此节点" />
        <div className="agent-canvas-node-card">
          <header><span>{nodeType}</span><strong>{data.title}</strong></header>
          {nodeType === 'image' ? <button type="button" className="agent-canvas-node-image-select" aria-label="展开图片编辑" onMouseDown={stopCanvasEvent} onDoubleClick={stopCanvasEvent} onClick={openEditor}>
            {data.assetRef?.url ? <AuthenticatedAgentImage source={data.assetRef.url} className="agent-canvas-node-media" alt="" /> : <span className="agent-canvas-node-image-placeholder" aria-label="待生成图片"><ImagePlus size={27} /></span>}
          </button> : null}
          {nodeType === 'video' && data.assetRef?.url ? <AuthenticatedAgentVideo source={data.assetRef.url} className="agent-canvas-node-media" aria-label="视频节点预览" preload="metadata" muted controls onClick={openEditor} /> : null}
          {nodeType === 'video' && !data.assetRef?.url ? <button type="button" className="agent-canvas-node-video-placeholder" onMouseDown={stopCanvasEvent} onDoubleClick={stopCanvasEvent} onClick={openEditor}>等待生成视频</button> : null}
          {nodeType === 'audio' ? <button type="button" className="agent-canvas-node-audio-placeholder" onMouseDown={stopCanvasEvent} onDoubleClick={stopCanvasEvent} onClick={openEditor}>音频设置</button> : null}
          {nodeType === 'text' ? <button type="button" className="agent-canvas-node-text-preview" onClick={openEditor}>{data.summary || '等待 Agent 写入结果'}</button> : null}
          <footer><span>v{data.version}</span><span>{data.status}</span></footer>
        </div>
        <Handle type="source" position={Position.Right} className="agent-canvas-node-quick-add" aria-label="拖出连接" />
      </div>
      {composer}
    </article>
  );
}

function CanvasEdge({ id, sourceX, sourceY, sourcePosition, targetX, targetY, targetPosition, data, markerEnd }) {
  const appearance = normalizeCanvasAppearance(data?.appearance);
  const [edgePath] = getBezierPath({ sourceX, sourceY, sourcePosition, targetX, targetY, targetPosition });
  const style = edgeStyleForAppearance(appearance);
  const energyStyle = { ...style, strokeWidth: Number(style.strokeWidth) + 1 };
  return <>
    <BaseEdge id={id} path={edgePath} markerEnd={markerEnd} style={style} />
    {appearance.edgeAnimated ? <path d={edgePath} className="agent-canvas-edge-flow" style={energyStyle} /> : null}
  </>;
}

const canvasEdgeTypes = { canvas: CanvasEdge };

function toFlowNode(node) {
  return {
    id: node.id,
    type: 'agent',
    position: node.position,
    selected: false,
    data: {
      type: normalizeCanvasNodeType(node.type),
      title: node.title,
      summary: node.summary,
      version: node.version,
      sourceNodeIds: node.sourceNodeIds,
      status: node.status,
      assetRef: node.assetRef || null,
      config: node.config || {},
      ...(node.groupId ? { groupId: node.groupId } : {})
    }
  };
}

function toFlowEdge(edge, appearance) {
  const style = edgeStyleForAppearance(appearance);
  return applyCanvasEdgeAppearance({
    ...edge,
    type: 'canvas',
    hidden: !appearance.edgesVisible,
    markerEnd: { type: MarkerType.ArrowClosed, color: style.stroke },
    data: { ...(edge.data || {}), relation: edge.relation || edge.data?.relation || 'reference', appearance }
  }, appearance);
}

function toCanvas(nodes, edges, revision, viewport, appearance) {
  return {
    version: 1,
    revision,
    ...(viewport ? { viewport } : {}),
    appearance: normalizeCanvasAppearance(appearance),
    nodes: nodes.map(node => ({
      id: node.id,
      type: normalizeCanvasNodeType(node.data.type),
      position: { x: node.position.x, y: node.position.y },
      title: node.data.title,
      version: node.data.version,
      sourceNodeIds: [...(node.data.sourceNodeIds || [])],
      status: node.data.status,
      summary: node.data.summary || '',
      ...(node.data.assetRef ? { assetRef: node.data.assetRef } : {}),
      ...(node.data.config && Object.keys(node.data.config).length ? { config: node.data.config } : {}),
      ...(node.data.groupId ? { groupId: node.data.groupId } : {})
    })),
    edges: edges.map(edge => ({ id: edge.id, source: edge.source, target: edge.target, relation: edge.data?.relation || 'reference' }))
  };
}

function canvasSignature(nodes, edges, viewport, appearance) {
  return JSON.stringify(toCanvas(nodes, edges, 0, viewport, appearance));
}

function IconTool({ label, disabled = false, active = false, onClick, children }) {
  return (
    <Tooltip title={label}>
      <span>
        <Button className={active ? 'is-active' : ''} type="text" aria-label={label} disabled={disabled} icon={children} onClick={onClick} />
      </span>
    </Tooltip>
  );
}

export function AgentCanvas({ taskId, canvas, onChange, onAppearanceChange, onSelectionChange, onViewportChange, onServerCanvasChange }) {
  const currentCanvas = canvas || { version: 1, revision: 0, nodes: [], edges: [] };
  const [nodes, setNodes] = useState([]);
  const [edges, setEdges] = useState([]);
  const [selectedIds, setSelectedIds] = useState([]);
  const [interactionMode, setInteractionMode] = useState('select');
  const [flowInstance, setFlowInstance] = useState(null);
  const [viewport, setViewport] = useState(currentCanvas.viewport || { x: 0, y: 0, zoom: 1 });
  const [appearance, setAppearance] = useState(() => normalizeCanvasAppearance(currentCanvas.appearance));
  const [nodeDragActive, setNodeDragActive] = useState(false);
  const [contextMenu, setContextMenu] = useState(null);
  const [contextMenuSubmenu, setContextMenuSubmenu] = useState('');
  const [connectionDropMenu, setConnectionDropMenu] = useState(null);
  const [connectingSourceID, setConnectingSourceID] = useState('');
  const [activeEditorNodeID, setActiveEditorNodeID] = useState('');
  const [assetPanel, setAssetPanel] = useState(null);
  const [assets, setAssets] = useState([]);
  const [assetLoading, setAssetLoading] = useState(false);
  const [assetError, setAssetError] = useState('');
  const [manualNode, setManualNode] = useState(null);
  const [manualNodeDraft, setManualNodeDraft] = useState({ title: '', summary: '', status: 'draft', config: {} });
  const [videoJobs, setVideoJobs] = useState({});
  const [videoJobsLoadedTaskID, setVideoJobsLoadedTaskID] = useState('');
  const [videoJobsLoadError, setVideoJobsLoadError] = useState('');
  const [videoBusyByNode, setVideoBusyByNode] = useState({});
  const [videoActionErrors, setVideoActionErrors] = useState({});
  const [nodeGenerationByNode, setNodeGenerationByNode] = useState({});
  const [configuredTextModels, setConfiguredTextModels] = useState([]);
  const [configuredImageModels, setConfiguredImageModels] = useState([]);
  const [configuredVideoModels, setConfiguredVideoModels] = useState([]);
  const [modelLoadError, setModelLoadError] = useState('');
  const [, setHistoryVersion] = useState(0);
  const rootRef = useRef(null);
  const nodesRef = useRef([]);
  const edgesRef = useRef([]);
  const viewportRef = useRef(viewport);
  const appearanceRef = useRef(appearance);
  const signatureRef = useRef('');
  const groupDragRef = useRef(null);
  const nodeSequenceRef = useRef(0);
  const clipboardRef = useRef(null);
  const uploadInputRef = useRef(null);
  const connectionSourceRef = useRef('');
  const connectionCompletedRef = useRef(false);
  const historyRef = useRef(createCanvasHistory({ nodes: [], edges: [] }));
  const taskIdRef = useRef(taskId);
  taskIdRef.current = taskId;

  const configuredTextModelOptions = useMemo(() => modelSelectOptions(configuredTextModels, 'text'), [configuredTextModels]);
  const configuredImageModelOptions = useMemo(() => modelSelectOptions(configuredImageModels, 'image'), [configuredImageModels]);
  const configuredVideoModelOptions = useMemo(() => {
    const options = modelSelectOptions(configuredVideoModels, 'video');
    return options.some(option => option.value === LOCAL_DOUBAO_VIDEO_MODEL_OPTION.value)
      ? options
      : [...options, LOCAL_DOUBAO_VIDEO_MODEL_OPTION];
  }, [configuredVideoModels]);

  useEffect(() => {
    let active = true;
    Promise.all(['text', 'image', 'video'].map(kind => listAvailableModels(kind)))
      .then(([textModels, imageModels, videoModels]) => {
        if (!active) return;
        setConfiguredTextModels(textModels);
        setConfiguredImageModels(imageModels);
        setConfiguredVideoModels(videoModels);
        setModelLoadError('');
      })
      .catch(error => {
        if (!active) return;
        setConfiguredTextModels([]);
        setConfiguredImageModels([]);
        setConfiguredVideoModels([]);
        setModelLoadError(error?.message || '模型列表加载失败');
      });
    return () => { active = false; };
  }, []);

  function closeNodeComposer() {
    setContextMenu(null);
    setConnectionDropMenu(null);
    setActiveEditorNodeID('');
    setAssetPanel(null);
  }

  const openManualNode = useCallback(node => {
    if (!node) return;
    setContextMenu(null);
    setManualNode(node);
    setManualNodeDraft({
      title: node.data?.title || '',
      summary: node.data?.summary || '',
      status: node.data?.status || 'draft',
      config: { ...(node.data?.config || {}) }
    });
  }, []);
  const handleNodeContentDoubleClick = useCallback(nodeID => {
    openManualNode(nodesRef.current.find(node => node.id === nodeID));
  }, [openManualNode]);
  const canvasNodeTypes = useMemo(() => ({
    agent: props => {
      const videoJob = videoJobs[props.id]?.projectId === taskId ? videoJobs[props.id] : null;
      const kind = normalizeCanvasNodeType(props.data?.type);
      const modelOptions = kind === 'text' ? configuredTextModelOptions : kind === 'image' ? configuredImageModelOptions : configuredVideoModelOptions;
      return <CanvasNode
        {...props}
        isEditorOpen={activeEditorNodeID === props.id}
        onOpenEditor={nodeID => { setAssetPanel(null); setActiveEditorNodeID(nodeID); }}
        onManualEdit={handleNodeContentDoubleClick}
        onPickImageAsset={openImageAssetPicker}
        onPickReferenceAsset={openReferenceAssetPanel}
        onRemoveReference={removeReference}
        onNodeConfigChange={handleNodeConfigChange}
        onSubmitText={submitTextForNode}
        onSubmitImage={submitImageForNode}
        onSubmitVideo={submitVideoForNode}
        onCancelVideo={cancelVideo}
        onRetryVideo={retryVideo}
        onConfirmVideo={confirmVideo}
        videoJob={videoJob}
        videoState={videoNodeViewState(videoJob)}
        videoBusy={videoJobsLoadedTaskID !== taskId || Boolean(videoBusyByNode[props.id])}
        videoActionError={videoActionErrors[props.id] || (videoJobsLoadedTaskID !== taskId ? videoJobsLoadError : '')}
        generationState={nodeGenerationByNode[props.id] || {}}
        modelOptions={modelOptions}
        modelLoadError={modelLoadError}
      />;
    }
  }), [activeEditorNodeID, configuredImageModelOptions, configuredTextModelOptions, configuredVideoModelOptions, handleNodeContentDoubleClick, modelLoadError, nodeGenerationByNode, taskId, videoActionErrors, videoBusyByNode, videoJobs, videoJobsLoadError, videoJobsLoadedTaskID]);
  const availability = useMemo(() => commandAvailability(nodes, selectedIds), [nodes, selectedIds]);
  const groups = useMemo(() => [...new Set(nodes.map(node => node.data?.groupId).filter(Boolean))]
    .map(groupId => ({ groupId, bounds: canvasGroupBounds(nodes, groupId) }))
    .filter(group => group.bounds), [nodes]);
  const contextAvailability = useMemo(
    () => contextMenu ? commandAvailability(nodes, contextMenu.nodeIds) : null,
    [contextMenu, nodes]
  );

  function updateRefs(nextNodes, nextEdges, nextViewport = viewportRef.current, nextAppearance = appearanceRef.current) {
    nodesRef.current = nextNodes;
    edgesRef.current = nextEdges;
    viewportRef.current = nextViewport;
    appearanceRef.current = nextAppearance;
    signatureRef.current = canvasSignature(nextNodes, nextEdges, nextViewport, nextAppearance);
  }

  useEffect(() => {
    const nextNodes = (currentCanvas.nodes || []).map(toFlowNode);
    const nextAppearance = normalizeCanvasAppearance(currentCanvas.appearance);
    const nextEdges = (currentCanvas.edges || []).map(edge => toFlowEdge(edge, nextAppearance));
    const nextViewport = currentCanvas.viewport || { x: 0, y: 0, zoom: 1 };
    if (canvasSignature(nextNodes, nextEdges, nextViewport, nextAppearance) === signatureRef.current) return;
    setNodes(nextNodes);
    setEdges(nextEdges);
    setSelectedIds([]);
    setContextMenu(null);
    setActiveEditorNodeID('');
    setAssetPanel(null);
    setViewport(nextViewport);
    setAppearance(nextAppearance);
    updateRefs(nextNodes, nextEdges, nextViewport, nextAppearance);
    historyRef.current.reset({ nodes: nextNodes, edges: nextEdges });
    setHistoryVersion(value => value + 1);
  }, [canvas]);

  useEffect(() => {
    const requestedTaskId = taskId;
    let active = true;
    setVideoJobs({});
    setVideoJobsLoadedTaskID('');
    setVideoJobsLoadError('');
    setVideoBusyByNode({});
    setVideoActionErrors({});
    setNodeGenerationByNode({});
    closeNodeComposer();
    if (!requestedTaskId) return () => { active = false; };
    listAgentVideoJobs(requestedTaskId)
      .then(result => {
        if (!shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current, active })) return;
        setVideoJobs(current => mergeLatestVideoJobsByNode(current, result?.jobs, requestedTaskId));
        setVideoJobsLoadedTaskID(requestedTaskId);
      })
      .catch(error => {
        if (!shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current, active })) return;
        setVideoJobsLoadError(error?.message || '视频任务状态加载失败，请刷新重试');
      });
    return () => { active = false; };
  }, [taskId]);

  useEffect(() => {
    if (!taskId) return undefined;
    if (videoJobsLoadedTaskID !== taskId) return undefined;
    const requestedTaskId = taskId;
    let active = true;
    const pending = Object.values(videoJobs).filter(job => videoNodeViewState(job).polling);
    if (!pending.length) return undefined;
    const timer = window.setTimeout(async () => {
      const updates = await Promise.all(pending.map(async job => {
        try { return (await getAgentVideoJob(requestedTaskId, job.id)).job; } catch (_) { return null; }
      }));
      if (!shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current, active })) return;
      setVideoJobs(current => mergeLatestVideoJobsByNode(current, updates.filter(Boolean), requestedTaskId));
    }, 2500);
    return () => { active = false; window.clearTimeout(timer); };
  }, [taskId, videoJobs, videoJobsLoadedTaskID]);

  function emitChange(nextNodes, nextEdges, nextViewport = viewportRef.current, nextAppearance = appearanceRef.current) {
    updateRefs(nextNodes, nextEdges, nextViewport, nextAppearance);
    onChange?.(toCanvas(nextNodes, nextEdges, currentCanvas.revision, nextViewport, nextAppearance));
  }

  function commit(nextNodes, nextEdges, { record = true, nextViewport = viewportRef.current } = {}) {
    setNodes(nextNodes);
    setEdges(nextEdges);
    if (record) historyRef.current.push({ nodes: nextNodes, edges: nextEdges });
    setHistoryVersion(value => value + 1);
    emitChange(nextNodes, nextEdges, nextViewport);
  }

  function handleNodesChange(changes) {
    const next = applyNodeChanges(changes, nodesRef.current);
    nodesRef.current = next;
    setNodes(next);
    if (changes.some(change => change.type === 'remove')) {
      const remaining = new Set(next.map(node => node.id));
      const nextEdges = edgesRef.current.filter(edge => remaining.has(edge.source) && remaining.has(edge.target));
      commit(next, nextEdges);
    }
  }

  function handleEdgesChange(changes) {
    const next = applyEdgeChanges(changes, edgesRef.current);
    edgesRef.current = next;
    setEdges(next);
    if (changes.some(change => change.type === 'remove' || change.type === 'add')) {
      commit(syncNodeReferencesFromEdges({ nodes: nodesRef.current, edges: next }), next);
    }
  }

  function handleSelectionChange(selection) {
    const ids = (selection?.nodes || []).map(node => node.id);
    setSelectedIds(ids);
    onSelectionChange?.(ids);
  }

  function handleNodeClick(event, node) {
    setContextMenu(null);
    setAssetPanel(null);
    if (activeEditorNodeID && activeEditorNodeID !== node.id) setActiveEditorNodeID('');
    const next = selectCanvasNodes(
      nodesRef.current,
      node.id,
      event.ctrlKey || event.metaKey || event.shiftKey,
      selectedIds
    );
    const ids = next.filter(item => item.selected).map(item => item.id);
    nodesRef.current = next;
    setNodes(next);
    setSelectedIds(ids);
    onSelectionChange?.(ids);
  }

  function handleNodeDoubleClick(event, node) {
    event.preventDefault();
    if (normalizeCanvasNodeType(node.data?.type) !== 'text') return;
    openManualNode(node);
  }

  function handleManualNodeSave() {
    if (!manualNode) return;
    const title = manualNodeDraft.title.trim();
    if (!title) return;
    const next = nodesRef.current.map(node => node.id === manualNode.id
      ? { ...node, data: { ...node.data, title, summary: manualNodeDraft.summary.trim(), status: manualNodeDraft.status, config: manualNodeDraft.config } }
      : node);
    commit(next, edgesRef.current);
    setManualNode(null);
  }

  function imageAssetPanelAnchor(event) {
    const rootBounds = rootRef.current?.getBoundingClientRect();
    const targetBounds = event?.currentTarget?.getBoundingClientRect();
    if (!rootBounds || !targetBounds) return null;
    const width = 288;
    return {
      left: Math.max(8, Math.min(targetBounds.left - rootBounds.left - width - 12, rootBounds.width - width - 8)),
      top: Math.max(8, Math.min(targetBounds.top - rootBounds.top, rootBounds.height - 260))
    };
  }

  function openImageMaterialPicker(nodeID, mode, event) {
    if (!taskId || !nodesRef.current.some(node => node.id === nodeID && node.data?.type === 'image')) return;
    event?.preventDefault();
    event?.stopPropagation();
    setAssetPanel({ assetKind: 'image', mode, nodeID, position: { x: 0, y: 0 }, anchor: imageAssetPanelAnchor(event) });
    setAssetLoading(true);
    setAssetError('');
    listAgentAssets(taskId, 'image')
      .then(result => setAssets(result.assets || []))
      .catch(error => {
        setAssets([]);
        setAssetError(readableAssetRequestError(error, 'load'));
      })
      .finally(() => setAssetLoading(false));
  }

  function openImageAssetPicker(nodeID, event) {
    openImageMaterialPicker(nodeID, 'image-source', event);
  }

  function openImageReferencePanel(nodeID, event) {
    openImageMaterialPicker(nodeID, 'image-reference', event);
  }

  function openReferenceAssetPanel(nodeID, event) {
    const node = nodesRef.current.find(item => item.id === nodeID);
    if (!taskId || !['image', 'video'].includes(node?.data?.type)) return;
    event?.preventDefault();
    event?.stopPropagation();
    const assetKind = node.data.type === 'image' ? 'image' : '';
    setAssetPanel({ assetKind, mode: 'node-reference', nodeID, position: { x: 0, y: 0 }, anchor: imageAssetPanelAnchor(event) });
    setAssetLoading(true);
    setAssetError('');
    listAgentAssets(taskId, assetKind)
      .then(result => setAssets(result.assets || []))
      .catch(error => {
        setAssets([]);
        setAssetError(readableAssetRequestError(error, 'load'));
      })
      .finally(() => setAssetLoading(false));
  }

  function removeReference(nodeID, assetID) {
    const next = nodesRef.current.map(node => node.id === nodeID
      ? {
        ...node,
        data: {
          ...node.data,
          config: {
            ...(node.data?.config || {}),
            referenceAssets: (node.data?.config?.referenceAssets || []).filter(asset => asset?.id !== assetID),
            autoReferenceAssetIDs: (node.data?.config?.autoReferenceAssetIDs || []).filter(id => id !== assetID)
          }
        }
      }
      : node);
    commit(next, edgesRef.current);
  }

  function handleNodeConfigChange(nodeID, patch) {
    const next = nodesRef.current.map(node => node.id === nodeID
      ? { ...node, data: { ...node.data, config: { ...(node.data?.config || {}), ...patch } } }
      : node);
    commit(next, edgesRef.current);
  }

  function applyGeneratedCanvas(serverCanvas) {
    if (!serverCanvas) return;
    const nextAppearance = normalizeCanvasAppearance(serverCanvas.appearance || appearanceRef.current);
    const nextNodes = (serverCanvas.nodes || []).map(toFlowNode);
    const nextEdges = (serverCanvas.edges || []).map(edge => toFlowEdge(edge, nextAppearance));
    const nextViewport = serverCanvas.viewport || viewportRef.current;
    setNodes(nextNodes);
    setEdges(nextEdges);
    setAppearance(nextAppearance);
    setViewport(nextViewport);
    updateRefs(nextNodes, nextEdges, nextViewport, nextAppearance);
    historyRef.current.push({ nodes: nextNodes, edges: nextEdges });
    setHistoryVersion(value => value + 1);
    onServerCanvasChange?.(serverCanvas);
  }

  async function submitTextForNode(nodeID, selectedModel) {
    const requestedTaskId = taskId;
    const node = nodesRef.current.find(item => item.id === nodeID);
    if (!requestedTaskId || node?.data?.type !== 'text') return;
    const config = node.data?.config || {};
    const attempt = Number(nodeGenerationByNode[nodeID]?.attempt) || 0;
    setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: true, error: '' } }));
    try {
      const result = await createAgentTextGeneration(requestedTaskId, {
        nodeId: nodeID,
        nodeVersion: node.data.version,
        modelId: selectedModel,
        prompt: String(config.prompt || '').trim(),
        idempotencyKey: nodeGenerationIdempotencyKey({ taskId: requestedTaskId, nodeId: nodeID, nodeVersion: node.data.version, kind: 'text', attempt })
      });
      if (requestedTaskId !== taskIdRef.current) return;
      applyGeneratedCanvas(result.canvas);
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { busy: false, error: '', attempt } }));
    } catch (error) {
      if (requestedTaskId !== taskIdRef.current) return;
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { busy: false, error: error?.message || '文字生成失败', attempt: attempt + 1 } }));
    }
  }

  async function submitImageForNode(nodeID, selectedModel) {
    const requestedTaskId = taskId;
    const node = nodesRef.current.find(item => item.id === nodeID);
    if (!requestedTaskId || node?.data?.type !== 'image') return;
    const config = node.data?.config || {};
    const attempt = Number(nodeGenerationByNode[nodeID]?.attempt) || 0;
    const sizes = { '1:1': '1024x1024', '16:9': '1536x1024', '9:16': '1024x1536' };
    setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: true, error: '' } }));
    try {
      const result = await createAgentImageGeneration(requestedTaskId, {
        nodeId: nodeID,
        nodeVersion: node.data.version,
        modelId: selectedModel,
        prompt: String(config.prompt || '').trim(),
        size: sizes[config.ratio] || sizes['16:9'],
        referenceAssetIds: (config.referenceAssets || []).filter(asset => asset?.kind === 'image').map(asset => asset.id),
        idempotencyKey: nodeGenerationIdempotencyKey({ taskId: requestedTaskId, nodeId: nodeID, nodeVersion: node.data.version, kind: 'image', attempt })
      });
      if (requestedTaskId !== taskIdRef.current) return;
      applyGeneratedCanvas(result.canvas);
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { busy: false, error: '', attempt } }));
    } catch (error) {
      if (requestedTaskId !== taskIdRef.current) return;
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { busy: false, error: error?.message || '图片生成失败', attempt: attempt + 1 } }));
    }
  }

  function videoRequestPayloadForNode(node, ignoreReferences = false) {
    const config = node?.data?.config || {};
    return {
      nodeId: node.id,
      idempotencyKey: makeVideoJobIdempotencyKey(),
      prompt: config.prompt || node.data?.summary || '',
      modelKey: config.model || '',
      aspectRatio: config.ratio || '',
      resolution: config.resolution || '',
      duration: config.durationSeconds || 10,
      referenceAssetIds: videoReferenceAssetIDs(config),
      ignoreReferences
    };
  }

  async function submitVideoForNode(nodeID, ignoreReferences = false, selectedModel = '') {
    const requestedTaskId = taskId;
    if (!requestedTaskId || videoJobsLoadedTaskID !== requestedTaskId) return;
    const node = nodesRef.current.find(item => item.id === nodeID);
    if (!node || node.data?.type !== 'video') return;
    const referenceError = videoReferenceCapabilityError(node.data?.config || {});
    if (referenceError) {
      setVideoActionErrors(current => ({ ...current, [nodeID]: referenceError }));
      return;
    }
    const payload = videoRequestPayloadForNode(node, ignoreReferences);
    if (selectedModel) payload.modelKey = selectedModel;
    if (!canSubmitVideoFromConfig({ prompt: payload.prompt })) return;
    setVideoBusyByNode(current => ({ ...current, [nodeID]: true }));
    setVideoActionErrors(current => ({ ...current, [nodeID]: '' }));
    try {
      const result = await createAgentVideoJob(requestedTaskId, payload);
      if (!shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current })) return;
      setVideoJobs(current => mergeLatestVideoJobsByNode(current, [result.job], requestedTaskId));
    } catch (error) {
      if (!shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current })) return;
      if (!ignoreReferences && String(error?.responseText || '').includes('UNSUPPORTED_REFERENCE_IMAGES')) {
        Modal.confirm({
          title: '当前模型不支持参考图',
          content: '继续后会忽略画布里连接的图片，按视频提示词生成。',
          okText: '仍然生成', cancelText: '取消',
          onOk: () => submitVideoForNode(nodeID, true)
        });
      } else {
        setVideoActionErrors(current => ({ ...current, [nodeID]: error?.message || '视频任务创建失败' }));
      }
    } finally {
      if (shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current })) setVideoBusyByNode(current => ({ ...current, [nodeID]: false }));
    }
  }

  async function cancelVideo(nodeID) {
    const requestedTaskId = taskId;
    const job = videoJobs[nodeID];
    if (!requestedTaskId || videoJobsLoadedTaskID !== requestedTaskId || !job) return;
    setVideoBusyByNode(current => ({ ...current, [nodeID]: true }));
    setVideoActionErrors(current => ({ ...current, [nodeID]: '' }));
    try {
      const result = await cancelAgentVideoJob(requestedTaskId, job.id);
      if (!shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current })) return;
      setVideoJobs(current => mergeLatestVideoJobsByNode(current, [result.job], requestedTaskId));
    } catch (error) {
      if (shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current })) setVideoActionErrors(current => ({ ...current, [nodeID]: error?.message || '取消视频任务失败' }));
    } finally {
      if (shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current })) setVideoBusyByNode(current => ({ ...current, [nodeID]: false }));
    }
  }

  async function retryVideo(nodeID) {
    const requestedTaskId = taskId;
    const job = videoJobs[nodeID];
    if (!requestedTaskId || videoJobsLoadedTaskID !== requestedTaskId || !job) return;
    setVideoBusyByNode(current => ({ ...current, [nodeID]: true }));
    setVideoActionErrors(current => ({ ...current, [nodeID]: '' }));
    try {
      const result = await retryAgentVideoJob(requestedTaskId, job.id, makeVideoJobIdempotencyKey());
      if (!shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current })) return;
      setVideoJobs(current => mergeLatestVideoJobsByNode(current, [result.job], requestedTaskId));
    } catch (error) {
      if (shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current })) setVideoActionErrors(current => ({ ...current, [nodeID]: error?.message || '重试视频任务失败' }));
    } finally {
      if (shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current })) setVideoBusyByNode(current => ({ ...current, [nodeID]: false }));
    }
  }

  async function confirmVideo(nodeID) {
    const requestedTaskId = taskId;
    const job = videoJobs[nodeID];
    if (!requestedTaskId || videoJobsLoadedTaskID !== requestedTaskId || !job) return;
    setVideoBusyByNode(current => ({ ...current, [nodeID]: true }));
    setVideoActionErrors(current => ({ ...current, [nodeID]: '' }));
    try {
      const result = await confirmAgentVideoJob(requestedTaskId, job.id);
      if (!shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current })) return;
      setVideoJobs(current => mergeLatestVideoJobsByNode(current, [result.job], requestedTaskId));
      await onServerCanvasChange?.();
    } catch (error) {
      if (shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current })) setVideoActionErrors(current => ({ ...current, [nodeID]: error?.message || '确认回填失败' }));
    } finally {
      if (shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current })) setVideoBusyByNode(current => ({ ...current, [nodeID]: false }));
    }
  }

  function handleConnect(connection) {
    connectionCompletedRef.current = true;
    const next = connectCanvasNodes(edgesRef.current, connection);
    if (next === edgesRef.current) return;
    const nextNodes = syncNodeReferencesFromEdges({ nodes: nodesRef.current, edges: next });
    commit(nextNodes, next.map(edge => toFlowEdge(edge, appearanceRef.current)));
  }

  function handleConnectStart(_event, params) {
    const sourceID = String(params?.nodeId || '');
    connectionSourceRef.current = sourceID;
    connectionCompletedRef.current = false;
    setConnectingSourceID(sourceID);
    setConnectionDropMenu(null);
  }

  function handleConnectEnd(event) {
    const sourceID = connectionSourceRef.current;
    const connectionCompleted = connectionCompletedRef.current;
    const droppedOnHandle = Boolean(event?.target?.closest?.('.react-flow__handle'));
    connectionSourceRef.current = '';
    connectionCompletedRef.current = false;
    setConnectingSourceID('');
    if (!shouldOpenConnectionDrop({ sourceID, connectionCompleted, droppedOnHandle })) return;
    const bounds = rootRef.current?.getBoundingClientRect();
    const point = event?.changedTouches?.[0] || event;
    const screenPosition = { x: Number(point?.clientX) || 0, y: Number(point?.clientY) || 0 };
    const position = flowInstance?.screenToFlowPosition(screenPosition) || { x: 0, y: 0 };
    setConnectionDropMenu({
      sourceID,
      position,
      left: Math.max(8, Math.min(screenPosition.x - (bounds?.left || 0), Math.max(8, (bounds?.width || 200) - 150))),
      top: Math.max(8, Math.min(screenPosition.y - (bounds?.top || 0), Math.max(8, (bounds?.height || 140) - 90)))
    });
  }

  function isValidCanvasConnection(connection) {
    return connectCanvasNodes(edgesRef.current, connection) !== edgesRef.current;
  }

  function handleConnectionDropCreate(kind) {
    if (!connectionDropMenu) return;
    const nextNode = createNodeAt({ id: nextCanvasID(kind), kind, position: connectionDropMenu.position });
    const nextNodes = nodesRef.current.map(node => ({ ...node, selected: false })).concat(nextNode);
    const nextEdges = connectCanvasNodes(edgesRef.current, { source: connectionDropMenu.sourceID, target: nextNode.id });
    if (nextEdges === edgesRef.current) return;
    const syncedNodes = syncNodeReferencesFromEdges({ nodes: nextNodes, edges: nextEdges });
    setSelectedIds([nextNode.id]);
    onSelectionChange?.([nextNode.id]);
    commit(syncedNodes, nextEdges.map(edge => toFlowEdge(edge, appearanceRef.current)));
    setConnectionDropMenu(null);
  }

  function runNodeCommand(command) {
    const next = command(nodesRef.current);
    if (next !== nodesRef.current) commit(next, edgesRef.current);
  }

  function addNode() {
    const bounds = rootRef.current?.getBoundingClientRect();
    const screenPoint = bounds ? { x: bounds.left + bounds.width / 2, y: bounds.top + bounds.height / 2 } : { x: 0, y: 0 };
    const position = flowInstance?.screenToFlowPosition(screenPoint) || { x: 0, y: 0 };
    nodeSequenceRef.current += 1;
    const nextNode = createCanvasNode(`node-${Date.now()}-${nodeSequenceRef.current}`, position);
    const next = nodesRef.current.map(node => ({ ...node, selected: false })).concat(nextNode);
    setSelectedIds([nextNode.id]);
    onSelectionChange?.([nextNode.id]);
    commit(next, edgesRef.current);
  }

  function nextCanvasID(prefix = 'node') {
    nodeSequenceRef.current += 1;
    return `${prefix}-${Date.now()}-${nodeSequenceRef.current}`;
  }

  function openContextMenu(event, target, nodeIds = []) {
    event.preventDefault();
    event.stopPropagation();
    const bounds = rootRef.current?.getBoundingClientRect();
    const clickPosition = { x: event.clientX, y: event.clientY };
    const position = flowInstance?.screenToFlowPosition(clickPosition) || { x: 0, y: 0 };
    const maxX = Math.max(8, (bounds?.width || 216) - 216);
    const maxY = Math.max(8, (bounds?.height || 304) - 304);
    setContextMenuSubmenu('');
    setContextMenu({
      target,
      nodeIds,
      position,
      left: Math.max(8, Math.min((event.clientX - (bounds?.left || 0)), maxX)),
      top: Math.max(8, Math.min((event.clientY - (bounds?.top || 0)), maxY))
    });
  }

  function handlePaneContextMenu(event) {
    const next = nodesRef.current.map(node => ({ ...node, selected: false }));
    nodesRef.current = next;
    setNodes(next);
    setSelectedIds([]);
    onSelectionChange?.([]);
    openContextMenu(event, 'pane');
  }

  function handleNodeContextMenu(event, node) {
    const keepSelection = selectedIds.includes(node.id);
    const nodeIds = keepSelection ? selectedIds : [node.id];
    if (!keepSelection) {
      const next = nodesRef.current.map(item => ({ ...item, selected: item.id === node.id }));
      nodesRef.current = next;
      setNodes(next);
      setSelectedIds(nodeIds);
      onSelectionChange?.(nodeIds);
    }
    openContextMenu(event, 'nodes', nodeIds);
  }

  function handleContextCreate(kind) {
    if (!contextMenu) return;
    const nextNode = createNodeAt({ id: nextCanvasID(kind), kind, position: contextMenu.position });
    const next = nodesRef.current.map(node => ({ ...node, selected: false })).concat(nextNode);
    setSelectedIds([nextNode.id]);
    onSelectionChange?.([nextNode.id]);
    commit(next, edgesRef.current);
    setContextMenu(null);
    setContextMenuSubmenu('');
  }

  async function openAssetPanel(assetKind, replaceNodeId = '') {
    if (!taskId) return;
    setContextMenu(null);
    setAssetPanel({ assetKind, replaceNodeId, position: contextMenu?.position || { x: 0, y: 0 } });
    setAssetLoading(true);
    setAssetError('');
    try {
      const result = await listAgentAssets(taskId, assetKind);
      setAssets(result.assets || []);
    } catch (error) {
      setAssets([]);
      setAssetError(readableAssetRequestError(error, 'load'));
    } finally {
      setAssetLoading(false);
    }
  }

  function beginLocalUpload(assetKind) {
    if (!taskId || !contextMenu) return;
    const position = contextMenu.position;
    setContextMenu(null);
    setContextMenuSubmenu('');
    setAssetPanel({ assetKind, replaceNodeId: '', position });
    setAssetLoading(false);
    setAssetError('');
    uploadInputRef.current?.click();
  }

  function commitAsset(asset) {
    if (!assetPanel || !asset) return;
    if (assetPanel.mode === 'image-reference' || assetPanel.mode === 'node-reference') {
      const next = nodesRef.current.map(node => node.id === assetPanel.nodeID
        ? appendReferenceAsset({ node, asset })
        : node);
      commit(syncNodeReferencesFromEdges({ nodes: next, edges: edgesRef.current }), edgesRef.current);
      const updatedNode = next.find(node => node.id === assetPanel.nodeID);
      if (updatedNode && manualNode?.id === updatedNode.id) {
        setManualNode(updatedNode);
        setManualNodeDraft(current => ({ ...current, config: { ...(updatedNode.data?.config || {}) } }));
      }
    } else if (assetPanel.mode === 'image-source') {
      const next = nodesRef.current.map(node => node.id === assetPanel.nodeID
        ? attachImageAsset({ node, asset })
        : node);
      commit(syncNodeReferencesFromEdges({ nodes: next, edges: edgesRef.current }), edgesRef.current);
      const updatedNode = next.find(node => node.id === assetPanel.nodeID);
      if (updatedNode && manualNode?.id === updatedNode.id) {
        setManualNode(updatedNode);
        setManualNodeDraft(current => ({ ...current, title: updatedNode.data?.title || current.title, config: { ...(updatedNode.data?.config || {}) } }));
      }
    } else if (assetPanel.replaceNodeId) {
      const next = nodesRef.current.map(node => node.id === assetPanel.replaceNodeId
        ? { ...node, data: { ...node.data, type: asset.kind, title: asset.name, assetRef: asset } }
        : node);
      commit(next, edgesRef.current);
    } else {
      const nextNode = createNodeAt({ id: nextCanvasID(asset.kind), kind: asset.kind, position: assetPanel.position, asset });
      const next = nodesRef.current.map(node => ({ ...node, selected: false })).concat(nextNode);
      setSelectedIds([nextNode.id]);
      onSelectionChange?.([nextNode.id]);
      commit(next, edgesRef.current);
    }
    setAssetPanel(null);
  }

  function startAssetUpload() {
    uploadInputRef.current?.click();
  }

  async function handleAssetUpload(event) {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file || !taskId || !assetPanel) return;
    setAssetLoading(true);
    setAssetError('');
    try {
      const result = await uploadAgentAsset(taskId, file);
      commitAsset(result.asset);
    } catch (error) {
      setAssetError(readableAssetRequestError(error, 'upload'));
    } finally {
      setAssetLoading(false);
    }
  }

  function handleContextCopy() {
    if (!contextMenu?.nodeIds.length) return;
    copySelection(contextMenu.nodeIds);
    setContextMenu(null);
  }

  function setCanvasSelection(ids) {
    const selected = new Set(ids || []);
    const next = nodesRef.current.map(node => ({ ...node, selected: selected.has(node.id) }));
    nodesRef.current = next;
    setNodes(next);
    setSelectedIds([...selected]);
    onSelectionChange?.([...selected]);
  }

  function copySelection(ids = selectedIds) {
    clipboardRef.current = copyCanvasSelection({
      nodes: nodesRef.current,
      edges: edgesRef.current,
      selectedIds: ids
    });
    return clipboardRef.current;
  }

  function pasteClipboardAt(position) {
    if (!clipboardRef.current) return false;
    const pasted = pasteCanvasSelection({
      clipboard: clipboardRef.current,
      position,
      nextID: sourceID => nextCanvasID(`copy-${sourceID}`)
    });
    if (!pasted.nodes.length) return false;
    const nextNodes = nodesRef.current.map(node => ({ ...node, selected: false })).concat(pasted.nodes);
    const nextEdges = edgesRef.current.concat(pasted.edges.map(edge => toFlowEdge(edge, appearanceRef.current)));
    const ids = pasted.nodes.map(node => node.id);
    setSelectedIds(ids);
    onSelectionChange?.(ids);
    commit(nextNodes, nextEdges);
    return true;
  }

  function handleContextPaste() {
    if (!contextMenu) return;
    pasteClipboardAt(contextMenu.position);
    setContextMenu(null);
  }

  function deleteCanvasNodes(ids = selectedIds) {
    if (!ids.length) return false;
    if (ids.includes(activeEditorNodeID)) closeNodeComposer();
    const result = removeCanvasNodes({ nodes: nodesRef.current, edges: edgesRef.current, ids });
    setSelectedIds([]);
    onSelectionChange?.([]);
    commit(result.nodes, result.edges);
    return true;
  }

  function handleContextDelete() {
    if (!contextMenu?.nodeIds.length) return;
    deleteCanvasNodes(contextMenu.nodeIds);
    setContextMenu(null);
  }

  function handleContextDisconnect() {
    if (!contextMenu?.nodeIds.length) return;
    const ids = new Set(contextMenu.nodeIds);
    const nextEdges = edgesRef.current.filter(edge => !ids.has(edge.source) && !ids.has(edge.target));
    if (nextEdges.length !== edgesRef.current.length) {
      commit(syncNodeReferencesFromEdges({ nodes: nodesRef.current, edges: nextEdges }), nextEdges);
    }
    setContextMenu(null);
  }

  function handleContextNodeCommand(command) {
    if (!contextMenu?.nodeIds.length) return;
    const next = command(nodesRef.current, contextMenu.nodeIds);
    if (next !== nodesRef.current) commit(next, edgesRef.current);
    setContextMenu(null);
  }

  function applyHistory(direction) {
    const snapshot = direction === 'undo' ? historyRef.current.undo() : historyRef.current.redo();
    const restoredEdges = snapshot.edges.map(edge => toFlowEdge(edge, appearanceRef.current));
    commit(snapshot.nodes, restoredEdges, { record: false });
  }

  function keyboardPastePosition() {
    const selected = nodesRef.current.filter(node => selectedIds.includes(node.id));
    if (selected.length) {
      return {
        x: Math.max(...selected.map(node => Number(node.position?.x) || 0)) + 48,
        y: Math.min(...selected.map(node => Number(node.position?.y) || 0)) + 48
      };
    }
    const bounds = rootRef.current?.getBoundingClientRect();
    return flowInstance?.screenToFlowPosition(bounds ? {
      x: bounds.left + bounds.width / 2,
      y: bounds.top + bounds.height / 2
    } : { x: 0, y: 0 }) || { x: 0, y: 0 };
  }

  function applyKeyboardCommand(command) {
    switch (command) {
      case 'undo':
      case 'redo':
        if ((command === 'undo' && historyRef.current.canUndo()) || (command === 'redo' && historyRef.current.canRedo())) applyHistory(command);
        return;
      default:
    }
  }

  function handleKeyDown(event) {
    if (event.key === 'Escape') {
      event.preventDefault();
      closeNodeComposer();
      return;
    }
    const command = canvasKeyboardCommand({
      ctrlKey: event.ctrlKey, metaKey: event.metaKey, shiftKey: event.shiftKey, key: event.key,
      targetTagName: event.target?.tagName, isContentEditable: event.target?.isContentEditable
    });
    if (!command) return;
    event.preventDefault();
    applyKeyboardCommand(command);
  }

  function handleMoveEnd(_event, nextViewport) {
    setViewport(nextViewport);
    viewportRef.current = nextViewport;
    onViewportChange?.(nextViewport);
    emitChange(nodesRef.current, edgesRef.current, nextViewport);
  }

  function updateAppearance(nextValue) {
    const nextAppearance = normalizeCanvasAppearance({ ...appearanceRef.current, ...nextValue });
    const nextEdges = edgesRef.current.map(edge => toFlowEdge(edge, nextAppearance));
    setAppearance(nextAppearance);
    setEdges(nextEdges);
    onAppearanceChange?.(nextAppearance);
    emitChange(nodesRef.current, nextEdges, viewportRef.current, nextAppearance);
  }

  function startGroupDrag(event, groupId) {
    event.stopPropagation();
    event.currentTarget.setPointerCapture(event.pointerId);
    groupDragRef.current = { groupId, pointerId: event.pointerId, startX: event.clientX, startY: event.clientY, startNodes: nodesRef.current };
  }

  function moveGroupDrag(event) {
    const drag = groupDragRef.current;
    if (!drag || drag.pointerId !== event.pointerId) return;
    const next = moveCanvasGroup(drag.startNodes, drag.groupId, { x: event.clientX - drag.startX, y: event.clientY - drag.startY });
    nodesRef.current = next;
    setNodes(next);
  }

  function endGroupDrag(event) {
    const drag = groupDragRef.current;
    if (!drag || drag.pointerId !== event.pointerId) return;
    groupDragRef.current = null;
    commit(nodesRef.current, edgesRef.current);
  }

  const alignItems = [
    { key: 'left', label: '左对齐', icon: <AlignStartVertical size={15} /> },
    { key: 'centerX', label: '水平居中', icon: <AlignCenterVertical size={15} /> },
    { key: 'right', label: '右对齐', icon: <AlignEndVertical size={15} /> },
    { key: 'top', label: '顶对齐', icon: <AlignStartHorizontal size={15} /> },
    { key: 'centerY', label: '垂直居中', icon: <AlignCenterHorizontal size={15} /> },
    { key: 'bottom', label: '底对齐', icon: <AlignEndHorizontal size={15} /> }
  ];

  return (
    <div ref={rootRef} className={`agent-canvas${connectingSourceID ? ' is-connecting' : ''}`} aria-label="创作画布" tabIndex={0} onKeyDown={handleKeyDown} style={appearanceCssVariables(appearance)}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={canvasNodeTypes}
        edgeTypes={canvasEdgeTypes}
        onInit={setFlowInstance}
        onNodesChange={handleNodesChange}
        onEdgesChange={handleEdgesChange}
        onSelectionChange={handleSelectionChange}
        onNodeClick={handleNodeClick}
        onNodeDoubleClick={handleNodeDoubleClick}
        onPaneClick={closeNodeComposer}
        onPaneContextMenu={handlePaneContextMenu}
        onNodeContextMenu={handleNodeContextMenu}
        onConnect={handleConnect}
        onConnectStart={handleConnectStart}
        onConnectEnd={handleConnectEnd}
        connectionRadius={24}
        isValidConnection={isValidCanvasConnection}
        onNodeDragStart={() => { setNodeDragActive(true); setContextMenu(null); }}
        onNodeDragStop={() => { setNodeDragActive(false); commit(nodesRef.current, edgesRef.current); }}
        minZoom={0.25}
        maxZoom={2}
        fitView={!currentCanvas.viewport}
        defaultViewport={currentCanvas.viewport}
        onMove={(_event, nextViewport) => {
          setViewport(nextViewport);
          viewportRef.current = nextViewport;
          onViewportChange?.(nextViewport);
        }}
        onMoveEnd={handleMoveEnd}
        panOnDrag={interactionMode === 'pan' ? [0, 1] : [1]}
        nodesDraggable={interactionMode !== 'pan'}
        elementsSelectable={interactionMode !== 'pan'}
        selectionOnDrag={interactionMode === 'select'}
        snapToGrid={appearance.alignmentGuides}
        snapGrid={[20, 20]}
        multiSelectionKeyCode="Shift"
      >
        {appearance.backgroundDots ? <Background variant="dots" gap={20} size={1} /> : null}
        <ViewportPortal>
          {groups.map(({ groupId, bounds }) => (
            <section className="agent-canvas-group" key={groupId} style={{ left: bounds.x, top: bounds.y, width: bounds.width, height: bounds.height }}>
              <button
                type="button"
                className="agent-canvas-group-handle"
                onPointerDown={event => startGroupDrag(event, groupId)}
                onPointerMove={moveGroupDrag}
                onPointerUp={endGroupDrag}
                onPointerCancel={endGroupDrag}
              >节点分组</button>
            </section>
          ))}
        </ViewportPortal>
      </ReactFlow>

      {connectionDropMenu ? <section className="agent-canvas-connection-drop-menu" aria-label="添加参考节点" style={{ left: connectionDropMenu.left, top: connectionDropMenu.top }}>
        <strong>添加参考节点</strong>
        <div>
          <button type="button" onClick={() => handleConnectionDropCreate('text')}>文本</button>
          <button type="button" onClick={() => handleConnectionDropCreate('image')}>图片</button>
        </div>
      </section> : null}

      {contextMenu ? <section
        className="agent-canvas-context-menu"
        role="menu"
        aria-label="画布右键菜单"
        style={{ left: contextMenu.left, top: contextMenu.top }}
        onContextMenu={event => event.preventDefault()}
      >
        {contextMenu.target === 'pane' ? <>
          <button
            type="button"
            role="menuitem"
            className="agent-canvas-context-menu-parent"
            onMouseEnter={() => setContextMenuSubmenu('create')}
            onFocus={() => setContextMenuSubmenu('create')}
            onClick={() => setContextMenuSubmenu('create')}
          ><span>新建节点</span><span aria-hidden="true">›</span></button>
          <button
            type="button"
            role="menuitem"
            className="agent-canvas-context-menu-parent"
            onMouseEnter={() => setContextMenuSubmenu('assets')}
            onFocus={() => setContextMenuSubmenu('assets')}
            onClick={() => setContextMenuSubmenu('assets')}
          ><span>从资产库添加</span><span aria-hidden="true">›</span></button>
          <button
            type="button"
            role="menuitem"
            className="agent-canvas-context-menu-parent"
            onMouseEnter={() => setContextMenuSubmenu('upload')}
            onFocus={() => setContextMenuSubmenu('upload')}
            onClick={() => setContextMenuSubmenu('upload')}
          ><span>本地上传</span><span aria-hidden="true">›</span></button>
          <span className="agent-canvas-context-menu-divider" />
          <button type="button" role="menuitem" disabled={!clipboardRef.current} onClick={handleContextPaste}>粘贴</button>
          <button type="button" role="menuitem" disabled={!historyRef.current.canUndo()} onClick={() => { applyHistory('undo'); setContextMenu(null); }}>撤销</button>
          <button type="button" role="menuitem" disabled={!historyRef.current.canRedo()} onClick={() => { applyHistory('redo'); setContextMenu(null); }}>重做</button>
          {contextMenuSubmenu ? <section className="agent-canvas-context-submenu" role="menu" aria-label="画布右键子菜单">
            {contextMenuSubmenu === 'create' ? <>
              <button type="button" role="menuitem" onClick={() => handleContextCreate('text')}>新建文本</button>
              <button type="button" role="menuitem" onClick={() => handleContextCreate('image')}>新建图片</button>
              <button type="button" role="menuitem" onClick={() => handleContextCreate('video')}>新建视频</button>
              <button type="button" role="menuitem" onClick={() => handleContextCreate('audio')}>新建音频</button>
            </> : null}
            {contextMenuSubmenu === 'assets' ? <>
              <button type="button" role="menuitem" onClick={() => openAssetPanel('image')}>图片</button>
              <button type="button" role="menuitem" onClick={() => openAssetPanel('video')}>视频</button>
            </> : null}
            {contextMenuSubmenu === 'upload' ? <>
              <button type="button" role="menuitem" onClick={() => beginLocalUpload('image')}>上传图片</button>
              <button type="button" role="menuitem" onClick={() => beginLocalUpload('video')}>上传视频</button>
            </> : null}
          </section> : null}
        </> : <>
          <p className="agent-canvas-context-menu-title">{contextMenu.nodeIds.length > 1 ? `已选 ${contextMenu.nodeIds.length} 个节点` : '节点操作'}</p>
          <button type="button" role="menuitem" onClick={handleContextCopy}>复制</button>
          {contextMenu.nodeIds.length === 1 && ['image', 'video'].includes(nodes.find(node => node.id === contextMenu.nodeIds[0])?.data?.type) ? <button type="button" role="menuitem" onClick={() => openAssetPanel(nodes.find(node => node.id === contextMenu.nodeIds[0])?.data?.type, contextMenu.nodeIds[0])}>替换素材</button> : null}
          <button type="button" role="menuitem" onClick={handleContextDisconnect}>断开连线</button>
          <span className="agent-canvas-context-menu-divider" />
          <button type="button" role="menuitem" disabled={!contextAvailability?.canGroup} onClick={() => handleContextNodeCommand((current, ids) => groupCanvasNodes(current, ids, `group-${Date.now()}`))}>分组</button>
          <button type="button" role="menuitem" disabled={!contextAvailability?.canUngroup} onClick={() => handleContextNodeCommand(ungroupCanvasNodes)}>取消分组</button>
          <button type="button" role="menuitem" disabled={!contextAvailability?.canAlign} onClick={() => handleContextNodeCommand((current, ids) => alignCanvasNodes(current, ids, 'left'))}>左对齐</button>
          <button type="button" role="menuitem" disabled={!contextAvailability?.canDistribute} onClick={() => handleContextNodeCommand((current, ids) => distributeCanvasNodes(current, ids, 'horizontal'))}>水平等距</button>
          <span className="agent-canvas-context-menu-divider" />
          <button type="button" role="menuitem" className="is-danger" onClick={handleContextDelete}>删除</button>
        </>}
      </section> : null}

      <input ref={uploadInputRef} className="agent-canvas-upload-input" type="file" accept="image/*,video/*" onChange={handleAssetUpload} />
      {assetPanel ? <aside className={`agent-canvas-asset-panel${assetPanel.anchor ? ' is-node-picker' : ''}`} aria-label="项目素材库" style={assetPanel.anchor || undefined}>
        <header><strong>{assetPanel.mode === 'image-source' || assetPanel.mode === 'node-reference' ? '我的收藏' : assetPanel.replaceNodeId ? '替换素材' : '添加素材'}</strong><button type="button" aria-label="关闭素材库" onClick={() => setAssetPanel(null)}>×</button></header>
        {assetPanel.mode === 'image-source' || assetPanel.mode === 'node-reference' ? <p className="agent-canvas-asset-section-title">画布所有素材</p> : null}
        <Button size="small" onClick={startAssetUpload} disabled={assetLoading}>上传{assetPanel.assetKind === 'image' ? '图片' : '视频'}</Button>
        {assetError ? <p className="agent-canvas-asset-error">{assetError}</p> : null}
        {assetLoading ? <p className="agent-canvas-asset-empty">正在读取素材...</p> : null}
        {!assetLoading && !assetError && !assets.length ? <p className="agent-canvas-asset-empty">还没有可用素材</p> : null}
        <div className="agent-canvas-asset-list">{assets.map(asset => <button type="button" key={asset.id} onClick={() => commitAsset(asset)}>{asset.url ? asset.kind === 'video'
          ? <AuthenticatedAgentVideo source={asset.url} aria-label="视频素材缩略图" preload="metadata" muted />
          : <AuthenticatedAgentImage source={asset.url} alt="" />
          : <ImagePlus size={16} />}<strong>{asset.name}</strong><span>{asset.kind === 'image' ? '图片' : '视频'}</span></button>)}</div>
      </aside> : null}

      <Modal
        title="手动编辑节点"
        open={Boolean(manualNode)}
        onCancel={() => setManualNode(null)}
        onOk={handleManualNodeSave}
        okText="保存"
        cancelText="取消"
        destroyOnClose
      >
        <label className="agent-canvas-manual-field">标题<Input value={manualNodeDraft.title} maxLength={191} onChange={event => setManualNodeDraft(current => ({ ...current, title: event.target.value }))} /></label>
        <label className="agent-canvas-manual-field">内容<Input.TextArea value={manualNodeDraft.summary} rows={6} maxLength={8000} onChange={event => setManualNodeDraft(current => ({ ...current, summary: event.target.value }))} /></label>
        <label className="agent-canvas-manual-field">状态<Select value={manualNodeDraft.status} onChange={status => setManualNodeDraft(current => ({ ...current, status }))} options={[{ value: 'draft', label: '草稿' }, { value: 'confirmed', label: '已确认' }, { value: 'failed', label: '需处理' }]} /></label>
      </Modal>

      {appearance.alignmentGuides && nodeDragActive ? <><span className="agent-canvas-guide agent-canvas-guide--vertical" /><span className="agent-canvas-guide agent-canvas-guide--horizontal" /></> : null}

      <nav className="agent-canvas-toolrail" aria-label="画布工具">
        <IconTool label="选择节点" active={interactionMode === 'select'} onClick={() => setInteractionMode('select')}><MousePointer2 size={17} /></IconTool>
        <IconTool label="拖动画布" active={interactionMode === 'pan'} onClick={() => setInteractionMode('pan')}><Hand size={17} /></IconTool>
        <IconTool label="连接节点" active={interactionMode === 'connect'} onClick={() => setInteractionMode('connect')}><GitBranch size={17} /></IconTool>
        <IconTool label={appearance.edgesVisible ? '隐藏连线' : '显示连线'} active={!appearance.edgesVisible} onClick={() => updateAppearance({ edgesVisible: !appearance.edgesVisible })}>{appearance.edgesVisible ? <EyeOff size={17} /> : <Eye size={17} />}</IconTool>
        <IconTool label="添加节点" onClick={addNode}><Plus size={17} /></IconTool>
        <Popover trigger="click" placement="rightTop" overlayClassName="agent-canvas-appearance-popover" content={<section className="agent-canvas-appearance-panel" aria-label="画布外观设置">
          <div className="agent-canvas-appearance-row"><span>连线颜色</span><div className="agent-canvas-color-options">{canvasEdgeColorOptions.map(option => <button key={option.value} type="button" className={`agent-canvas-color-swatch is-${option.value}${appearance.edgeColor === option.value ? ' is-active' : ''}`} aria-label={option.label} aria-pressed={appearance.edgeColor === option.value} onClick={() => updateAppearance({ edgeColor: option.value })} />)}</div></div>
          <div className="agent-canvas-appearance-row"><span>连线动效</span><Switch size="small" checked={appearance.edgeAnimated} onChange={edgeAnimated => updateAppearance({ edgeAnimated })} /></div>
          <div className="agent-canvas-appearance-row agent-canvas-appearance-row--stack"><span>连线粗细</span><Segmented size="small" value={appearance.edgeWidth} options={canvasEdgeWidthOptions} onChange={edgeWidth => updateAppearance({ edgeWidth })} /></div>
          <div className="agent-canvas-appearance-row"><span>背景网点</span><Switch size="small" checked={appearance.backgroundDots} onChange={backgroundDots => updateAppearance({ backgroundDots })} /></div>
          <div className="agent-canvas-appearance-row"><span>辅助线对齐</span><Switch size="small" checked={appearance.alignmentGuides} onChange={alignmentGuides => updateAppearance({ alignmentGuides })} /></div>
        </section>}>
          <span><IconTool label="画布外观"><SlidersHorizontal size={17} /></IconTool></span>
        </Popover>
      </nav>

      <nav className="agent-canvas-editbar" aria-label="节点编辑工具">
        <IconTool label="撤销 Ctrl+Z" disabled={!historyRef.current.canUndo()} onClick={() => applyHistory('undo')}><Undo2 size={17} /></IconTool>
        <IconTool label="重做 Ctrl+Shift+Z" disabled={!historyRef.current.canRedo()} onClick={() => applyHistory('redo')}><Redo2 size={17} /></IconTool>
        <span className="agent-canvas-tool-divider" />
        <IconTool label={availability.canGroup ? '分组' : availability.groupReason} disabled={!availability.canGroup} onClick={() => runNodeCommand(current => groupCanvasNodes(current, selectedIds, `group-${Date.now()}`))}><Group size={17} /></IconTool>
        <IconTool label={availability.canUngroup ? '解组' : availability.ungroupReason} disabled={!availability.canUngroup} onClick={() => runNodeCommand(current => ungroupCanvasNodes(current, selectedIds))}><Ungroup size={17} /></IconTool>
        <Dropdown disabled={!availability.canAlign} menu={{ items: alignItems, onClick: ({ key }) => runNodeCommand(current => alignCanvasNodes(current, selectedIds, key)) }} trigger={['click']}>
          <span><IconTool label={availability.canAlign ? '对齐' : availability.alignReason} disabled={!availability.canAlign}><AlignCenterHorizontal size={17} /></IconTool></span>
        </Dropdown>
        <IconTool label={availability.canDistribute ? '水平等距' : availability.distributeReason} disabled={!availability.canDistribute} onClick={() => runNodeCommand(current => distributeCanvasNodes(current, selectedIds, 'horizontal'))}><AlignHorizontalDistributeCenter size={17} /></IconTool>
        <IconTool label={availability.canDistribute ? '垂直等距' : availability.distributeReason} disabled={!availability.canDistribute} onClick={() => runNodeCommand(current => distributeCanvasNodes(current, selectedIds, 'vertical'))}><AlignVerticalDistributeCenter size={17} /></IconTool>
      </nav>

      <nav className="agent-canvas-zoombar" aria-label="画布缩放工具">
        <IconTool label="缩小" onClick={() => flowInstance?.zoomOut()}><ZoomOut size={17} /></IconTool>
        <span className="agent-canvas-zoom-value">{Math.round((viewport.zoom || 1) * 100)}%</span>
        <IconTool label="放大" onClick={() => flowInstance?.zoomIn()}><ZoomIn size={17} /></IconTool>
        <IconTool label="适合屏幕" onClick={() => flowInstance?.fitView({ padding: 0.18 })}><Maximize2 size={17} /></IconTool>
        <IconTool label={selectedIds.length ? '聚焦选中' : '请先选择节点'} disabled={!selectedIds.length} onClick={() => flowInstance?.fitView({ nodes: nodesRef.current.filter(node => selectedIds.includes(node.id)), padding: 0.3 })}><Focus size={17} /></IconTool>
      </nav>
    </div>
  );
}
