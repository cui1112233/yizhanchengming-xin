import { Background, BaseEdge, Handle, MiniMap, Position, ReactFlow, ViewportPortal, applyEdgeChanges, applyNodeChanges, getBezierPath, useUpdateNodeInternals } from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { Button, Dropdown, Input, Modal, Popover, Segmented, Select, Switch, Tag, Tooltip, message } from 'antd';
import {
  AlignCenterHorizontal, AlignCenterVertical, AlignEndHorizontal, AlignEndVertical,
  AlignHorizontalDistributeCenter, AlignStartHorizontal, AlignStartVertical,
  AlignVerticalDistributeCenter, GitBranch, Group, Hand,
  Clapperboard, FileText, Film, GripHorizontal, Image as ImageIcon,
  Download, Folder, ImagePlus, MoreHorizontal, MousePointer2, Music2, Plus, Scissors, SlidersHorizontal, Sparkles, Ungroup
} from 'lucide-react';
import { createContext, forwardRef, lazy, Suspense, useCallback, useContext, useEffect, useImperativeHandle, useLayoutEffect, useMemo, useRef, useState } from 'react';
import {
  alignCanvasNodes, canvasGroupBounds, canvasKeyboardCommand, commandAvailability,
  connectCanvasNodes, distributeCanvasNodes, groupCanvasNodes,
  isCanvasMultiSelectionEvent, moveCanvasGroup, removeCanvasEdges, sameCanvasSelection, selectCanvasNodes, ungroupCanvasNodes
} from '../pages/agent-canvas-commands';
import { createCanvasHistory } from '../pages/agent-canvas-history';
import { layoutCanvas } from '../pages/agent-canvas-layout';
import {
  appearanceCssVariables, applyCanvasEdgeAppearance, canvasEdgeColorOptions, canvasEdgeWidthOptions,
  edgeStyleForAppearance, normalizeCanvasAppearance
} from '../pages/agent-canvas-appearance';
import {
  appendReferenceAsset, attachImageAsset, copyCanvasSelection, createNodeAt, pasteCanvasSelection, removeCanvasNodes, removeReferenceAsset,
  readableAssetRequestError, shouldOpenConnectionDrop
} from '../pages/agent-canvas-context-menu';
import { connectionDropChoices, createTypedCanvasEdge, syncCanvasEffectiveReferences } from '../pages/agent-canvas-connections';
import { cancelAgentGeneration, cancelAgentVideoJob, confirmAgentGeneration, confirmAgentImageGeneration, confirmAgentVideoJob, createAgentAudioGeneration, createAgentImageGeneration, createAgentImageTool, createAgentScriptSplit, createAgentShotBatchNodes, createAgentTextGeneration, createAgentVideoJob, createAgentVideoTool, getAgentGeneration, getAgentGenerationSummary, getAgentVideoJob, listAgentAssets, listAgentGenerations, listAgentScriptShots, listAgentVideoJobs, previewAgentCanvasInputs, removeAgentBatchFactoryNode, retryAgentGeneration, retryAgentVideoJob, syncAgentBatchFactoryBook, uploadAgentAsset } from '../../shared/api/agent';
import { generateBookAssetImages, rewriteWorkingFront, runBookStage, saveBookOverride, saveVideoOverride, setPrimaryBookAssetImage, submitBookMerge, submitBookTo121, updateBookAsset, updateBookMetadata, updateBookSource } from '../../shared/api/batchFactoryV11';
import { listAvailableModels } from '../../shared/api/modelCatalog';
import { buildNodeGenerationRequest, nextNodeGenerationAttempt, nodeExecutionReview } from '../pages/agent-canvas-node-generation';
import { audioJobRequestForNode, canSubmitAudioNode } from '../pages/agent-canvas-audio-jobs';
import { activeGenerationTaskForNode, generationPresentation, generationSummaryText, latestGenerationTaskForNode, mergeGenerationTask, mergeLatestGenerationTasksByNode, shouldAcceptGenerationResponse, shouldPollGeneration } from '../pages/agent-generation-state';
import { normalizeNodeConfig, patchNodeConfig } from '../pages/agent-node-config';
import { inheritedModelSelection, shareModelSelection } from '../pages/agent-canvas-model-selection';
import { modelInputCapabilityError } from '../pages/agent-model-capabilities';
import { canSubmitVideoFromConfig, makeVideoJobIdempotencyKey, mergeLatestVideoJobsByNode, shouldAcceptVideoJobResponse, videoNodeViewState, videoReferenceAssetIDs, videoReferenceCapabilityError } from '../pages/agent-canvas-video-jobs';
import { canvasNodeStatusLabel, canvasNodeTypeLabel, normalizeCanvasNodeType, usesPinnedBatchFactoryComposer } from '../pages/agent-canvas-node-presentation';
import { retainCanvasAssetPanel, retainCanvasEditorNodeID, viewportHasChanged } from '../pages/agent-canvas-state';
import { shouldRecoverCanvasViewport } from '../pages/agent-canvas-viewport';
import { batchFactoryGroupLabel, batchFactoryMappedAction, bookGroupId, focusRequestedBookGroup, isBatchFactoryNode, latestBookRevision, mappedBatchFactoryBooksNeedingRefresh, mappedNodeStatePresentation, projectedNodePresentation, removeBatchFactoryCanvasNodes, runMappedBatchFactoryScriptStage, runMappedBatchFactoryVideoProduction, videoNodeTiming } from '../pages/agent-canvas-batch-factory';
import { invalidateScriptSplitForSource, normalizeScriptShots, normalizeScriptShotsForConfig, runAfterDurableCanvasSave, scriptShotsByNode as groupScriptShotsByNode, scriptSourceMarker, scriptSplitReady } from '../pages/agent-script-shots';
import { AuthenticatedAgentImage } from './AuthenticatedAgentImage';
import { AuthenticatedAgentVideo } from './AuthenticatedAgentVideo';
import { AuthenticatedAgentAudio } from './AuthenticatedAgentAudio';
import { AgentCanvasControls } from './AgentCanvasControls';
import { AgentDirectorDeskNode } from './AgentDirectorDeskNode';
import { createDirectorDeskHostAdapter } from './director-desk/DirectorDeskHostAdapter';
import { insertDirectorCapture as insertDirectorCaptureState } from './director-desk/director-capture';
import AgentNodeComposer from './AgentNodeComposer';

const DirectorDeskApp = lazy(() => import('./director-desk/DirectorDeskApp'));
export { insertDirectorCapture } from './director-desk/director-capture';

function nodeModelInputCapabilityError(node, selectedModel, models) {
  const kind = normalizeCanvasNodeType(node?.data?.type);
  const config = normalizeNodeConfig(kind, node?.data?.config || {});
  const modelID = String(selectedModel || config.modelId || '').trim();
  const model = (Array.isArray(models) ? models : []).find(item => String(item?.id || item?.modelId || '').trim() === modelID);
  const references = Array.isArray(config.referenceAssets) ? config.referenceAssets : [];
  const count = expectedKind => references.filter(item => String(item?.kind || '').trim() === expectedKind).length;
  return modelInputCapabilityError(model, {
    prompt: config.prompt || node?.data?.summary || '',
    imageReferences: count('image'),
    videoReferences: count('video'),
    audioReferences: count('audio')
  });
}

function videoJobGenerationPresentation(job) {
  const state = {
    preparing: 'queued', submitting: 'queued', queued: 'queued', running: 'running',
    result_ready: 'running', ingesting: 'running', canceling: 'running',
    awaiting_confirmation: 'awaiting_confirmation', failed: 'failed', canceled: 'cancelled',
    abandoned: 'failed', applied: 'completed'
  }[String(job?.state || '')] || 'draft';
  return generationPresentation({ state, errorMessage: job?.errorMessage });
}

function CanvasNodeTypeIcon({ type }) {
  const icons = {
    text: FileText,
    script: Clapperboard,
    image: ImageIcon,
    video: Film,
    audio: Music2,
    director: Clapperboard
  };
  const Icon = icons[normalizeCanvasNodeType(type)] || FileText;
  return <span className={`agent-canvas-node-type-icon is-${normalizeCanvasNodeType(type)}`} aria-hidden="true"><Icon size={14} strokeWidth={1.8} /></span>;
}

function AgentAudioQuickTools({ source, onOperation, onMore }) {
  const actions = [
    { operation: 'trim', label: '截取', icon: <Scissors size={14} /> },
    { operation: 'custom_split', label: '自定义切分', icon: <SlidersHorizontal size={14} /> },
    { operation: 'smart_split', label: '智能切分', icon: <Sparkles size={14} /> },
    { operation: 'speed', label: '变速', icon: <span className="agent-audio-quick-tools__speed">◔</span> }
  ];
  return <div className="agent-audio-quick-tools nodrag nopan" onMouseDown={event => event.stopPropagation()} onClick={event => event.stopPropagation()}>
    {actions.map(action => <button type="button" key={action.operation} title={action.label} onClick={() => onOperation?.(action.operation)}>{action.icon}<span>{action.label}</span></button>)}
    <button type="button" title="更多音频操作" aria-label="更多音频操作" onClick={onMore}><MoreHorizontal size={15} /><span>更多</span></button>
    {source ? <a href={source} download title="下载音频" aria-label="下载音频"><Download size={15} /></a> : null}
  </div>;
}

function CanvasNode({
  data, id, onManualEdit, onPickImageAsset, onPickReferenceAsset, onRemoveReference,
  onNodeConfigChange, onSubmitText, onSubmitImage, onSubmitVideo, onSubmitAudio, onImageTool, onVideoTool, onCancelGeneration, onRetryGeneration, onCancelVideo, onRetryVideo, onConfirmVideo,
  onSelectImageCandidate, onConfirmImage, onSelectAudioCandidate, onConfirmAudio, onPublishBatchFactory,
  onSourceTextChange, onCommitScriptSource, onRunBatchFactoryStage, onSplitScript, onBatchShots, scriptState, modelsByKind,
  videoJob, videoState, videoBusy, videoActionError, generationState, models, modelLoadError, inputPreview, onPreviewInputs, timingWarning, taskPresentation,
  selected, isEditorOpen, onOpenEditor, sourcePresentation, onRemoveFromCanvas, onSaveDirectorDeskScreenshot, onOpenDirectorDesk, canvasZoom = 1
}) {
  const updateNodeInternals = useUpdateNodeInternals();
  const stopCanvasEvent = event => {
    if (!isCanvasMultiSelectionEvent(event)) event.stopPropagation();
  };
  const handleNodeDoubleClick = event => {
    event.preventDefault();
    event.stopPropagation();
    if (normalizeCanvasNodeType(data.type) === 'text') {
      onManualEdit?.(id);
      return;
    }
    onOpenEditor?.(id);
  };
  const openEditor = event => {
    if (isCanvasMultiSelectionEvent(event)) return;
    event.preventDefault();
    stopCanvasEvent(event);
    onOpenEditor?.(id);
  };
  const nodeType = normalizeCanvasNodeType(data.type);
  const isPinnedBatchFactoryComposer = usesPinnedBatchFactoryComposer(data);
  const applyAudioOperation = operation => {
    if (nodeType !== 'audio') return;
    onNodeConfigChange?.(id, { parameters: { ...(data.config?.parameters || {}), operation } });
    onOpenEditor?.(id);
  };
  useLayoutEffect(() => {
    updateNodeInternals(id);
  }, [id, isEditorOpen, isPinnedBatchFactoryComposer, updateNodeInternals]);
  const composer = (isEditorOpen || isPinnedBatchFactoryComposer) ? <AgentNodeComposer
    compact={isPinnedBatchFactoryComposer}
    canvasZoom={canvasZoom}
    node={{ id, data }}
    models={models}
    modelLoadError={modelLoadError}
    generationState={generationState}
    modelsByKind={modelsByKind}
    scriptState={scriptState}
    onSourceTextChange={onSourceTextChange}
    onCommitScriptSource={onCommitScriptSource}
    onRunBatchFactoryStage={onRunBatchFactoryStage}
    onSplitScript={onSplitScript}
    onBatchShots={onBatchShots}
    onConfigChange={onNodeConfigChange}
    inputPreview={inputPreview}
    onPreviewInputs={onPreviewInputs}
    timingWarning={timingWarning}
    onPickImageAsset={onPickImageAsset}
    onPickReferenceAsset={onPickReferenceAsset}
    onRemoveReference={onRemoveReference}
    onSubmit={nodeType === 'text' ? onSubmitText : nodeType === 'image' ? onSubmitImage : nodeType === 'video' ? (nodeID, modelID, prompt) => onSubmitVideo?.(nodeID, false, modelID, prompt) : onSubmitAudio}
    onImageTool={onImageTool}
    onVideoTool={onVideoTool}
    onSelectImageCandidate={onSelectImageCandidate}
    onConfirmImage={onConfirmImage}
    onSelectAudioCandidate={onSelectAudioCandidate}
    onConfirmAudio={onConfirmAudio}
    onPublishBatchFactory={onPublishBatchFactory}
    onCancelGeneration={onCancelGeneration}
    onRetryGeneration={onRetryGeneration}
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
      className={`agent-canvas-node agent-canvas-node--${nodeType} agent-canvas-node--${data.status}${selected ? ' is-selected' : ''}${isEditorOpen ? ' is-editor-open' : ''}${taskPresentation ? ' has-task-status' : ''}`}
      onDoubleClick={handleNodeDoubleClick}
    >
      <div className="agent-canvas-node-preview-shell">
        <Handle type="target" position={Position.Left} className="agent-canvas-node-target" aria-label="连接到此节点" />
        {isPinnedBatchFactoryComposer ? <div className="agent-canvas-node-pinned-composer">
          <header className="agent-canvas-node-card-header agent-canvas-node-pinned-header">
            <span className="agent-canvas-node-drag-handle" aria-label="拖动节点"><GripHorizontal size={14} /></span>
            <CanvasNodeTypeIcon type={nodeType} /><strong>{data.title}</strong><Tag color="blue" className="nodrag nopan" onMouseDown={stopCanvasEvent}>批量工厂</Tag>
          </header>
          <div className="agent-canvas-node-pinned-composer__body">{composer}</div>
          <footer className="agent-canvas-node-pinned-footer"><span>v{data.version}</span><span className={`agent-canvas-node-task-state is-${taskPresentation?.state || data.status}`}>{taskPresentation?.errorMessage || taskPresentation?.message || canvasNodeStatusLabel(data.status)}</span><button type="button" className="nodrag nopan" onMouseDown={stopCanvasEvent} onClick={event => { stopCanvasEvent(event); onRemoveFromCanvas?.(id); }}>从画布移除</button></footer>
          <Handle type="source" position={Position.Right} className="agent-canvas-node-quick-add" aria-label="拖出连接" />
        </div> : <div className="agent-canvas-node-card">
          {nodeType === 'audio' && data.assetRef?.url && selected ? <AgentAudioQuickTools source={data.assetRef.url} onOperation={applyAudioOperation} onMore={openEditor} /> : null}
          <header className="agent-canvas-node-card-header">
            <span className="agent-canvas-node-drag-handle" aria-label="拖动节点"><GripHorizontal size={14} /></span>
            <CanvasNodeTypeIcon type={nodeType} />
            <button
              type="button"
              className="nodrag nopan agent-canvas-node-card-header-action"
              aria-label={`编辑${canvasNodeTypeLabel(nodeType)}节点`}
              onMouseDown={stopCanvasEvent}
              onClick={openEditor}
            ><strong>{data.title}</strong><span>{canvasNodeTypeLabel(nodeType)}</span></button>
            {sourcePresentation?.sourceLabel ? <Tag color="blue" className="nodrag nopan" onMouseDown={stopCanvasEvent}>{sourcePresentation.sourceLabel}</Tag> : null}
          </header>
          {nodeType === 'image' ? <div className="agent-canvas-node-image-select agent-canvas-node-drag-surface" aria-label="图片节点预览">
            {data.assetRef?.url ? <AuthenticatedAgentImage source={data.assetRef.url} className="agent-canvas-node-media" alt="" draggable={false} /> : <span className="agent-canvas-node-image-placeholder" aria-label="待生成图片"><ImagePlus size={27} /></span>}
          </div> : null}
          {nodeType === 'video' && data.assetRef?.url ? <AuthenticatedAgentVideo source={data.assetRef.url} className="agent-canvas-node-media agent-canvas-node-drag-surface" aria-label="视频节点预览" preload="metadata" muted controls draggable={false} onDoubleClick={stopCanvasEvent} /> : null}
          {nodeType === 'video' && !data.assetRef?.url ? <div className="agent-canvas-node-video-placeholder agent-canvas-node-drag-surface">等待生成视频</div> : null}
          {nodeType === 'audio' && data.assetRef?.url ? <div className="agent-canvas-node-audio-preview agent-canvas-node-drag-surface">
            <AuthenticatedAgentAudio source={data.assetRef.url} className="nodrag nopan" aria-label="音频节点预览" preload="metadata" controls waveform waveformSourceLabel={data.title || '音频节点'} onDoubleClick={stopCanvasEvent} />
            <button type="button" className="nodrag nopan" onMouseDown={stopCanvasEvent} onClick={openEditor}>编辑音频</button>
          </div> : null}
          {nodeType === 'audio' && !data.assetRef?.url ? <div className="agent-canvas-node-audio-placeholder agent-canvas-node-drag-surface">等待生成或添加音频</div> : null}
          {nodeType === 'director' ? <AgentDirectorDeskNode
            node={{ id, data }}
            onOpen={onOpenDirectorDesk}
            onConfigChange={onNodeConfigChange}
            onPickReferenceAsset={onPickReferenceAsset}
            onRemoveReference={onRemoveReference}
            onSaveScreenshot={onSaveDirectorDeskScreenshot}
          /> : null}
          {nodeType === 'text' ? <div className="agent-canvas-node-text-preview agent-canvas-node-drag-surface">{data.summary || '等待 Agent 写入结果'}</div> : null}
          {nodeType === 'script' ? <div className="agent-canvas-node-text-preview agent-canvas-node-drag-surface">{data.config?.sourceText || '等待填写剧本正文'}</div> : null}
          <footer><span>v{data.version}</span><span className={`agent-canvas-node-task-state is-${taskPresentation?.state || data.status}`}>{taskPresentation?.errorMessage || taskPresentation?.message || canvasNodeStatusLabel(data.status)}</span>{sourcePresentation?.sourceLabel ? <button type="button" className="nodrag nopan" onMouseDown={stopCanvasEvent} onClick={event => { stopCanvasEvent(event); onRemoveFromCanvas?.(id); }}>从画布移除</button> : null}</footer>
        </div>}
        {!isPinnedBatchFactoryComposer ? <Handle type="source" position={Position.Right} className="agent-canvas-node-quick-add" aria-label="拖出连接" /> : null}
      </div>
      {!isPinnedBatchFactoryComposer ? composer : null}
    </article>
  );
}

const CanvasNodeContext = createContext(null);

function CanvasNodeRenderer(props) {
  const bindings = useContext(CanvasNodeContext);
  if (!bindings) return null;
  const {
    activeEditorNodeID, canvasZoom, configuredAudioModels, configuredImageModels, configuredTextModels, configuredVideoModels,
    executions, generationTasksByNode, handleNodeContentDoubleClick, modelLoadError, nodeGenerationByNode,
    openNodeComposer, scriptShotsByNode, sharedScriptShotsByNode, taskId, videoActionErrors, videoBusyByNode,
    videoJobs, videoJobsLoadError, videoJobsLoadedTaskID,
    openImageAssetPicker, openReferenceAssetPanel, removeReference, handleNodeConfigChange, saveDirectorDeskScreenshot, openDirectorDesk,
    submitTextForNode, submitImageForNode, submitImageTool, submitVideoTool, submitVideoForNode, submitAudioForNode,
    selectImageCandidate, confirmImage, selectAudioCandidate, confirmAudio,
    changeScriptSource, commitScriptSource, runBatchFactoryScriptStageForNode, splitScriptForNode, batchShotsForNode, cancelGeneration, retryGeneration, cancelVideo, retryVideo, confirmVideo, publishBatchFactoryMerge, canvasInputPreviewByNode, previewInputsForNode, removeBatchFactoryNode
  } = bindings;
  const videoJob = videoJobs[props.id]?.projectId === taskId ? videoJobs[props.id] : null;
  const kind = normalizeCanvasNodeType(props.data?.type);
  const models = kind === 'text' || kind === 'script' ? configuredTextModels : kind === 'image' ? configuredImageModels : kind === 'video' ? configuredVideoModels : configuredAudioModels;
  const modelsByKind = { text: configuredTextModels, image: configuredImageModels, video: configuredVideoModels, audio: configuredAudioModels };
  const restoredReview = kind === 'image' ? nodeExecutionReview(executions, props.id) : null;
  const localGeneration = nodeGenerationByNode[props.id] || {};
  const generationTask = generationTasksByNode[props.id]?.projectId === taskId ? generationTasksByNode[props.id] : null;
  const loadedScriptState = scriptShotsByNode[props.id] || {};
  const sharedShots = sharedScriptShotsByNode[props.id];
  const effectiveScriptState = Array.isArray(sharedShots) ? { ...loadedScriptState, shots: sharedShots } : loadedScriptState;
  const scriptState = kind === 'script'
    ? normalizeScriptShotsForConfig(props.data?.config || {}, {
      shots: effectiveScriptState.shots || [],
      ...(generationTask ? { generation: generationTask } : {})
    })
    : effectiveScriptState;
  const mergedGeneration = mergeGenerationTask({ id: props.id, data: props.data }, generationTask, localGeneration.selectedAssetId);
  const presentation = mergedGeneration.task ? generationPresentation(mergedGeneration.task) : null;
  const mappedPresentation = mappedNodeStatePresentation(props.data);
  const taskPresentation = mappedPresentation || presentation || (kind === 'video' && videoJob ? videoJobGenerationPresentation(videoJob) : null);
  const taskCandidates = mergedGeneration.task?.state === 'awaiting_confirmation'
    ? mergedGeneration.resultOptions.map(result => ({ ...result.asset, ...result, id: result.assetId, resultId: result.id }))
    : [];
  const candidates = localGeneration.candidates?.length
    ? localGeneration.candidates
    : restoredReview?.candidates?.length ? restoredReview.candidates : taskCandidates;
  const effectivePresentation = mappedPresentation || presentation;
  const generationState = {
    ...(restoredReview || {}),
    ...(mergedGeneration.task ? {
      generationId: mergedGeneration.task.id,
      status: mergedGeneration.task.state,
      mode: mergedGeneration.task.mode,
      presentation,
      selectedAssetId: mergedGeneration.previewAssetId
    } : {}),
    ...localGeneration,
    ...(mappedPresentation ? { presentation: mappedPresentation } : {}),
    candidates,
    busy: Boolean(localGeneration.busy || effectivePresentation?.busy),
    error: localGeneration.error || mappedPresentation?.errorMessage || (mergedGeneration.task?.state === 'failed' ? mergedGeneration.task.errorMessage : '')
  };
  return <CanvasNode
    {...props}
    isEditorOpen={activeEditorNodeID === props.id}
    onOpenEditor={openNodeComposer}
    onManualEdit={handleNodeContentDoubleClick}
    onPickImageAsset={openImageAssetPicker}
    onPickReferenceAsset={openReferenceAssetPanel}
    onRemoveReference={removeReference}
    onSaveDirectorDeskScreenshot={saveDirectorDeskScreenshot}
    onOpenDirectorDesk={openDirectorDesk}
    onNodeConfigChange={handleNodeConfigChange}
    inputPreview={canvasInputPreviewByNode[props.id]}
    onPreviewInputs={previewInputsForNode}
    timingWarning={kind === 'video' ? videoNodeTiming({
      audioSeconds: Number(props.data?.config?.connectedAudioDurationSeconds || 0),
      selectedSeconds: Number(props.data?.config?.durationOverrideSeconds || 0),
      allowedSeconds: props.data?.config?.allowedDurationSeconds
    }).warning : ''}
    onSubmitText={submitTextForNode}
    onSubmitImage={submitImageForNode}
    onImageTool={submitImageTool}
    onVideoTool={submitVideoTool}
    onSubmitVideo={submitVideoForNode}
    onSubmitAudio={submitAudioForNode}
    onSelectImageCandidate={selectImageCandidate}
    onConfirmImage={confirmImage}
    onSelectAudioCandidate={selectAudioCandidate}
    onConfirmAudio={confirmAudio}
    onPublishBatchFactory={publishBatchFactoryMerge}
    onSourceTextChange={changeScriptSource}
    onCommitScriptSource={commitScriptSource}
    onRunBatchFactoryStage={runBatchFactoryScriptStageForNode}
    onSplitScript={splitScriptForNode}
    onBatchShots={batchShotsForNode}
    scriptState={scriptState}
    modelsByKind={modelsByKind}
    onCancelGeneration={cancelGeneration}
    onRetryGeneration={retryGeneration}
    onCancelVideo={cancelVideo}
    onRetryVideo={retryVideo}
    onConfirmVideo={confirmVideo}
    videoJob={videoJob}
    videoState={mappedPresentation && kind === 'video'
      ? { ...videoNodeViewState(videoJob), label: mappedPresentation.errorMessage || mappedPresentation.state, canRetry: mappedPresentation.canRetry }
      : videoNodeViewState(videoJob)}
    videoBusy={videoJobsLoadedTaskID !== taskId || Boolean(videoBusyByNode[props.id])}
    videoActionError={videoActionErrors[props.id] || (videoJobsLoadedTaskID !== taskId ? videoJobsLoadError : '')}
    generationState={generationState}
    taskPresentation={taskPresentation}
    models={models}
    modelLoadError={modelLoadError}
    sourcePresentation={projectedNodePresentation(props.data)}
    onRemoveFromCanvas={removeBatchFactoryNode}
    canvasZoom={canvasZoom}
  />;
}

const canvasNodeTypes = { agent: CanvasNodeRenderer };

function CanvasEdge({ id, sourceX, sourceY, sourcePosition, targetX, targetY, targetPosition, data }) {
  const appearance = normalizeCanvasAppearance(data?.appearance);
  const [edgePath] = getBezierPath({ sourceX, sourceY, sourcePosition, targetX, targetY, targetPosition });
  const style = edgeStyleForAppearance(appearance);
  const energyStyle = { ...style, strokeWidth: Number(style.strokeWidth) + 1 };
  return <>
    <BaseEdge id={id} path={edgePath} style={style} />
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
      config: normalizeNodeConfig(normalizeCanvasNodeType(node.type), node.config || {}),
      ...(node.groupId ? { groupId: node.groupId } : {})
    }
  };
}

function toFlowEdge(edge, appearance) {
  return applyCanvasEdgeAppearance({
    ...edge,
    type: 'canvas',
    data: { ...(edge.data || {}), relation: edge.relation || edge.data?.relation || 'reference', appearance }
  }, appearance);
}

function toCanvas(nodes, edges, revision, viewport, appearance, viewState) {
  return {
    version: 1,
    revision,
    ...(viewport ? { viewport } : {}),
    ...(viewState ? { viewState } : {}),
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

export const AgentCanvas = forwardRef(function AgentCanvas({ taskId, canvas, executions = [], scriptShots = [], onScriptShotsChange, workspaceView: workspaceViewState, onWorkspaceViewChange, conversationOpen, onToggleConversation, onChange, onAppearanceChange, onSelectionChange, onViewportChange, onViewportCommit, onFlushCanvas, onServerCanvasChange, onExecutionChange, onAssetsCreated, requestedBookId = '' }, ref) {
  const currentCanvas = canvas || { version: 1, revision: 0, nodes: [], edges: [] };
  const workspaceView = workspaceViewState || { snapToGrid: false, showMinimap: true, showConnections: true };
  const [nodes, setNodes] = useState([]);
  const [edges, setEdges] = useState([]);
  const [selectedIds, setSelectedIds] = useState([]);
  const [selectedEdgeIds, setSelectedEdgeIds] = useState([]);
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
  const [directorDeskNodeID, setDirectorDeskNodeID] = useState('');
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
  const [canvasInputPreviewByNode, setCanvasInputPreviewByNode] = useState({});
  const [generationTasksByNode, setGenerationTasksByNode] = useState({});
  const [generationTasksLoadedTaskID, setGenerationTasksLoadedTaskID] = useState('');
  const [generationTasksLoadError, setGenerationTasksLoadError] = useState('');
  const [generationSummary, setGenerationSummary] = useState(null);
  const [collapsedGroups, setCollapsedGroups] = useState(() => new Set());
  const [scriptShotsByNode, setScriptShotsByNode] = useState({});
  const sharedScriptShotsByNode = useMemo(() => groupScriptShotsByNode(scriptShots), [scriptShots]);
  const [configuredTextModels, setConfiguredTextModels] = useState([]);
  const [configuredImageModels, setConfiguredImageModels] = useState([]);
  const [configuredVideoModels, setConfiguredVideoModels] = useState([]);
  const [configuredAudioModels, setConfiguredAudioModels] = useState([]);
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
  const focusedBookGroupRef = useRef('');
  const viewportRecoveryKeyRef = useRef('');
  const viewportRecoveryTimerRef = useRef(null);
  const suppressViewportCommitRef = useRef(false);
  const scriptSourceSaveByNodeRef = useRef({});
  const batchFactoryStatusPollingRef = useRef(false);
  const onServerCanvasChangeRef = useRef(onServerCanvasChange);
  taskIdRef.current = taskId;
  onServerCanvasChangeRef.current = onServerCanvasChange;
  const pendingMappedBooks = useMemo(() => mappedBatchFactoryBooksNeedingRefresh(nodes), [nodes]);
  const pendingMappedBookKey = useMemo(() => pendingMappedBooks.map(({ batchId, bookId }) => `${batchId}\u0000${bookId}`).join('\u0001'), [pendingMappedBooks]);

  useEffect(() => {
    let active = true;
    Promise.all(['text', 'image', 'video', 'audio'].map(kind => listAvailableModels(kind)))
      .then(([textModels, imageModels, videoModels, audioModels]) => {
        if (!active) return;
        setConfiguredTextModels(textModels);
        setConfiguredImageModels(imageModels);
        setConfiguredVideoModels(videoModels);
        setConfiguredAudioModels(audioModels);
        setModelLoadError('');
      })
      .catch(error => {
        if (!active) return;
        setConfiguredTextModels([]);
        setConfiguredImageModels([]);
        setConfiguredVideoModels([]);
        setConfiguredAudioModels([]);
        setModelLoadError(error?.message || '模型列表加载失败');
      });
    return () => { active = false; };
  }, []);

  function closeNodeComposer(event) {
    // React Flow can report nested controls as pane clicks. Only its blank pane closes the editor.
    if (event && !event.target?.classList?.contains('react-flow__pane')) return;
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
  const loadScriptShots = useCallback(async nodeID => {
    const requestedTaskId = taskId;
    if (!requestedTaskId || !nodeID) return;
    setScriptShotsByNode(current => ({ ...current, [nodeID]: { ...(current[nodeID] || {}), state: 'loading', errorMessage: '' } }));
    try {
      const result = await listAgentScriptShots(requestedTaskId, nodeID);
      if (requestedTaskId !== taskIdRef.current) return;
      const nodeConfig = nodesRef.current.find(node => node.id === nodeID)?.data?.config || {};
      const normalized = normalizeScriptShotsForConfig(nodeConfig, result);
      setScriptShotsByNode(current => ({
        ...current,
        [nodeID]: normalized
      }));
      onScriptShotsChange?.(nodeID, normalized.shots);
    } catch (error) {
      if (requestedTaskId !== taskIdRef.current) return;
      setScriptShotsByNode(current => ({ ...current, [nodeID]: { ...(current[nodeID] || {}), state: 'failed', errorMessage: error?.message || '镜头读取失败' } }));
    }
  }, [onScriptShotsChange, taskId]);
  const openNodeComposer = useCallback(nodeID => {
    setAssetPanel(null);
    setActiveEditorNodeID(nodeID);
    if (normalizeCanvasNodeType(nodesRef.current.find(node => node.id === nodeID)?.data?.type) === 'script') void loadScriptShots(nodeID);
  }, [loadScriptShots]);
  async function previewInputsForNode(nodeID, selectedModel) {
    const requestedTaskId = taskId;
    const node = nodesRef.current.find(item => item.id === nodeID);
    if (!requestedTaskId || !node) return;
    const config = node.data?.config || {};
    setCanvasInputPreviewByNode(current => ({ ...current, [nodeID]: { loading: true, error: '', inputs: null } }));
    try {
      const result = await previewAgentCanvasInputs(requestedTaskId, {
        nodeId: nodeID,
        provider: config.provider,
        modelId: selectedModel || config.modelId
      });
      if (requestedTaskId !== taskIdRef.current) return;
      setCanvasInputPreviewByNode(current => ({ ...current, [nodeID]: { loading: false, error: '', inputs: result?.inputs || result || null } }));
    } catch (error) {
      if (requestedTaskId !== taskIdRef.current) return;
      setCanvasInputPreviewByNode(current => ({ ...current, [nodeID]: { loading: false, error: error?.message || '画布输入解析失败', inputs: null } }));
    }
  }
  function persistGenerationDraftPrompt(nodeID, draftPrompt) {
    const currentNode = nodesRef.current.find(item => item.id === nodeID);
    if (!currentNode || typeof draftPrompt !== 'string' || String(currentNode.data?.config?.prompt || '') === draftPrompt) return currentNode;
    const kind = normalizeCanvasNodeType(currentNode.data?.type);
    const next = nodesRef.current.map(node => node.id === nodeID
      ? { ...node, data: { ...node.data, config: patchNodeConfig(kind, node.data?.config || {}, { prompt: draftPrompt }) } }
      : node);
    commit(next, edgesRef.current);
    return nodesRef.current.find(item => item.id === nodeID);
  }
  async function validateGenerationInputs(nodeID, selectedModel) {
    const requestedTaskId = taskId;
    if (!requestedTaskId || !nodeID) throw new Error('画布任务尚未准备完成，未提交生成请求');
    setCanvasInputPreviewByNode(current => ({ ...current, [nodeID]: { loading: true, error: '', inputs: null } }));
    try {
      const result = await runAfterDurableCanvasSave(onFlushCanvas, () => {
        const node = nodesRef.current.find(item => item.id === nodeID);
        return previewAgentCanvasInputs(requestedTaskId, {
          nodeId: nodeID,
          provider: node?.data?.config?.provider,
          modelId: selectedModel || node?.data?.config?.modelId
        });
      });
      if (requestedTaskId !== taskIdRef.current) return false;
      setCanvasInputPreviewByNode(current => ({ ...current, [nodeID]: { loading: false, error: '', inputs: result?.inputs || result || null } }));
      return true;
    } catch (error) {
      if (requestedTaskId === taskIdRef.current) setCanvasInputPreviewByNode(current => ({ ...current, [nodeID]: { loading: false, error: error?.message || '画布输入解析失败', inputs: null } }));
      throw error;
    }
  }

  async function publishBatchFactoryMerge(nodeID) {
    const requestedTaskId = taskIdRef.current;
    const node = nodesRef.current.find(item => item.id === nodeID);
    const action = batchFactoryMappedAction(node);
    const config = node?.data?.config || {};
    if (!requestedTaskId || action?.type !== 'merge') return;
    setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: true, actionBusy: true, error: '' } }));
    try {
      await submitBookTo121(action.batchId, action.bookId, {
        ...(String(config.publishOrganization || '').trim() ? { organization: String(config.publishOrganization).trim() } : {})
      });
      if (requestedTaskId !== taskIdRef.current) return;
      await syncAgentBatchFactoryBook(requestedTaskId, action.bookId);
      if (requestedTaskId === taskIdRef.current) await onServerCanvasChange?.();
    } catch (error) {
      if (requestedTaskId === taskIdRef.current) {
        setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], error: error?.message || '上传到视频管理系统失败' } }));
      }
    } finally {
      if (requestedTaskId === taskIdRef.current) {
        setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: false, actionBusy: false } }));
      }
    }
  }

  const canvasZoom = Number.isFinite(Number(viewport?.zoom)) ? Number(viewport.zoom) : 1;
  const canvasNodeContextValue = useMemo(() => ({
    activeEditorNodeID, canvasZoom, configuredAudioModels, configuredImageModels, configuredTextModels, configuredVideoModels,
    executions, generationTasksByNode, handleNodeContentDoubleClick, modelLoadError, nodeGenerationByNode, canvasInputPreviewByNode,
    openNodeComposer, scriptShotsByNode, sharedScriptShotsByNode, taskId, videoActionErrors, videoBusyByNode,
    videoJobs, videoJobsLoadError, videoJobsLoadedTaskID,
    openImageAssetPicker, openReferenceAssetPanel, removeReference, handleNodeConfigChange, saveDirectorDeskScreenshot, openDirectorDesk,
    submitTextForNode, submitImageForNode, submitImageTool, submitVideoTool, submitVideoForNode, submitAudioForNode,
    selectImageCandidate, confirmImage, selectAudioCandidate, confirmAudio,
    changeScriptSource, commitScriptSource, runBatchFactoryScriptStageForNode, splitScriptForNode, batchShotsForNode, cancelGeneration, retryGeneration, cancelVideo, retryVideo, confirmVideo, publishBatchFactoryMerge, previewInputsForNode, removeBatchFactoryNode
  }), [activeEditorNodeID, batchShotsForNode, cancelGeneration, cancelVideo, canvasInputPreviewByNode, canvasZoom, changeScriptSource, commitScriptSource, configuredAudioModels, configuredImageModels, configuredTextModels, configuredVideoModels, confirmAudio, confirmImage, confirmVideo, executions, generationTasksByNode, handleNodeContentDoubleClick, handleNodeConfigChange, modelLoadError, nodeGenerationByNode, openDirectorDesk, openImageAssetPicker, openNodeComposer, openReferenceAssetPanel, publishBatchFactoryMerge, removeBatchFactoryNode, removeReference, retryGeneration, retryVideo, runBatchFactoryScriptStageForNode, saveDirectorDeskScreenshot, scriptShotsByNode, selectAudioCandidate, selectImageCandidate, sharedScriptShotsByNode, splitScriptForNode, submitAudioForNode, submitImageForNode, submitImageTool, submitTextForNode, submitVideoForNode, submitVideoTool, taskId, videoActionErrors, videoBusyByNode, videoJobs, videoJobsLoadError, videoJobsLoadedTaskID]);
  const availability = useMemo(() => commandAvailability(nodes, selectedIds), [nodes, selectedIds]);
  const groups = useMemo(() => [...new Set(nodes.map(node => node.data?.groupId).filter(Boolean))]
    .map(groupId => ({ groupId, bounds: canvasGroupBounds(nodes, groupId) }))
    .filter(group => group.bounds), [nodes]);
  const contextAvailability = useMemo(
    () => contextMenu ? commandAvailability(nodes, contextMenu.nodeIds) : null,
    [contextMenu, nodes]
  );
  const visibleNodes = useMemo(() => nodes.map(node => ({ ...node, hidden: collapsedGroups.has(node.data?.groupId) })), [collapsedGroups, nodes]);
  const visibleEdges = useMemo(() => edges.map(edge => ({
    ...edge,
    hidden: !workspaceView.showConnections || collapsedGroups.has(nodes.find(node => node.id === edge.source)?.data?.groupId) || collapsedGroups.has(nodes.find(node => node.id === edge.target)?.data?.groupId)
  })), [collapsedGroups, edges, nodes, workspaceView.showConnections]);

  function toggleGroup(groupId) {
    setCollapsedGroups(current => {
      const next = new Set(current);
      if (next.has(groupId)) next.delete(groupId);
      else next.add(groupId);
      return next;
    });
  }

  function locateGroup(groupId) {
    const groupNodes = nodesRef.current.filter(node => node.data?.groupId === groupId);
    if (!groupNodes.length) return;
    setCollapsedGroups(current => {
      if (!current.has(groupId)) return current;
      const next = new Set(current);
      next.delete(groupId);
      return next;
    });
    window.requestAnimationFrame(() => flowInstance?.fitView?.({ nodes: groupNodes, padding: 0.22, duration: 260 }));
  }

  useEffect(() => {
    const groupId = bookGroupId(requestedBookId);
    if (!groupId || focusedBookGroupRef.current === groupId) return;
    const focusedNodes = focusRequestedBookGroup({
      bookId: requestedBookId,
      nodes,
      focus: groupNodes => locateGroup(groupNodes[0]?.data?.groupId)
    });
    if (focusedNodes.length) focusedBookGroupRef.current = groupId;
  }, [nodes, requestedBookId]);

  function updateRefs(nextNodes, nextEdges, nextViewport = viewportRef.current, nextAppearance = appearanceRef.current) {
    nodesRef.current = nextNodes;
    edgesRef.current = nextEdges;
    viewportRef.current = nextViewport;
    appearanceRef.current = nextAppearance;
    signatureRef.current = canvasSignature(nextNodes, nextEdges, nextViewport, nextAppearance);
  }

  useEffect(() => {
    const mappedNodes = (currentCanvas.nodes || []).map(toFlowNode);
    const nextAppearance = normalizeCanvasAppearance(currentCanvas.appearance);
    const nextEdges = (currentCanvas.edges || []).map(edge => toFlowEdge(edge, nextAppearance));
    const nextNodes = syncCanvasEffectiveReferences(mappedNodes, nextEdges);
    const nextViewport = currentCanvas.viewport || { x: 0, y: 0, zoom: 1 };
    if (canvasSignature(nextNodes, nextEdges, nextViewport, nextAppearance) === signatureRef.current) return;
    setNodes(nextNodes);
    setEdges(nextEdges);
    setSelectedIds([]);
    setSelectedEdgeIds([]);
    setContextMenu(null);
    setActiveEditorNodeID(current => retainCanvasEditorNodeID(current, nextNodes));
    setAssetPanel(current => retainCanvasAssetPanel(current, nextNodes));
    setViewport(nextViewport);
    setAppearance(nextAppearance);
    updateRefs(nextNodes, nextEdges, nextViewport, nextAppearance);
    historyRef.current.reset({ nodes: nextNodes, edges: nextEdges });
    setHistoryVersion(value => value + 1);
  }, [canvas]);

  useEffect(() => {
    if (!flowInstance || !currentCanvas.viewport || !nodes.length) return undefined;
    const recoveryKey = `${taskId}\u0000${nodes.map(node => node.id).sort().join('\u0001')}`;
    if (viewportRecoveryKeyRef.current === recoveryKey) return undefined;
    viewportRecoveryKeyRef.current = recoveryKey;
    const frame = window.requestAnimationFrame(() => {
      const recoverableNodes = (flowInstance.getNodes?.() || nodesRef.current).filter(node => !node.hidden);
      const bounds = rootRef.current?.getBoundingClientRect();
      if (!bounds || !shouldRecoverCanvasViewport(recoverableNodes, viewportRef.current, bounds)) return;
      suppressViewportCommitRef.current = true;
      window.clearTimeout(viewportRecoveryTimerRef.current);
      flowInstance.fitView({ nodes: recoverableNodes, padding: 0.18, duration: 260 });
      viewportRecoveryTimerRef.current = window.setTimeout(() => {
        suppressViewportCommitRef.current = false;
      }, 800);
    });
    return () => window.cancelAnimationFrame(frame);
  }, [currentCanvas.viewport, flowInstance, nodes, taskId]);

  useEffect(() => () => window.clearTimeout(viewportRecoveryTimerRef.current), []);

  useEffect(() => {
    const requestedTaskId = taskId;
    let active = true;
    setVideoJobs({});
    setVideoJobsLoadedTaskID('');
    setVideoJobsLoadError('');
    setVideoBusyByNode({});
    setVideoActionErrors({});
    setNodeGenerationByNode({});
    setGenerationTasksByNode({});
    setGenerationTasksLoadedTaskID('');
    setGenerationTasksLoadError('');
    setGenerationSummary(null);
    setScriptShotsByNode({});
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
    listAgentGenerations(requestedTaskId)
      .then(result => {
        if (!shouldAcceptGenerationResponse({ requestedTaskId, currentTaskId: taskIdRef.current, active })) return;
        setGenerationTasksByNode(current => mergeLatestGenerationTasksByNode(current, result?.tasks, requestedTaskId));
        setGenerationTasksLoadedTaskID(requestedTaskId);
      })
      .catch(error => {
        if (!shouldAcceptGenerationResponse({ requestedTaskId, currentTaskId: taskIdRef.current, active })) return;
        setGenerationTasksLoadError(error?.message || '生成任务状态加载失败，请刷新重试');
      });
    void refreshGenerationSummary(requestedTaskId, active);
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

  useEffect(() => {
    if (!taskId || generationTasksLoadedTaskID !== taskId) return undefined;
    const requestedTaskId = taskId;
    let active = true;
    const pending = Object.values(generationTasksByNode).filter(shouldPollGeneration);
    if (!pending.length) return undefined;
    const timer = window.setTimeout(async () => {
      const updates = await Promise.all(pending.map(async generation => {
        try { return (await getAgentGeneration(requestedTaskId, generation.id)).task; } catch (_) { return null; }
      }));
      if (!shouldAcceptGenerationResponse({ requestedTaskId, currentTaskId: taskIdRef.current, active })) return;
      setGenerationTasksByNode(current => mergeLatestGenerationTasksByNode(current, updates.filter(Boolean), requestedTaskId));
      void refreshGenerationSummary(requestedTaskId, active);
    }, 2500);
    return () => { active = false; window.clearTimeout(timer); };
  }, [generationTasksByNode, generationTasksLoadedTaskID, taskId]);

  useEffect(() => {
    if (!taskId || !pendingMappedBooks.length) return undefined;
    const requestedTaskId = taskId;
    let active = true;
    const timer = window.setTimeout(async () => {
      if (batchFactoryStatusPollingRef.current) return;
      batchFactoryStatusPollingRef.current = true;
      try {
        await Promise.all(pendingMappedBooks.map(({ bookId }) => syncAgentBatchFactoryBook(requestedTaskId, bookId)));
        if (active && requestedTaskId === taskIdRef.current) await onServerCanvasChangeRef.current?.();
      } catch (_) {
        // Keep the existing truthful running state and retry on the next poll.
      } finally {
        batchFactoryStatusPollingRef.current = false;
      }
    }, 2500);
    return () => { active = false; window.clearTimeout(timer); };
  }, [pendingMappedBookKey, taskId]);

  function emitChange(nextNodes, nextEdges, nextViewport = viewportRef.current, nextAppearance = appearanceRef.current) {
    updateRefs(nextNodes, nextEdges, nextViewport, nextAppearance);
    onChange?.(toCanvas(nextNodes, nextEdges, currentCanvas.revision, nextViewport, nextAppearance, currentCanvas.viewState));
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
      commit(syncCanvasEffectiveReferences(nodesRef.current, next), next);
    }
  }

  function handleSelectionChange(selection) {
    const nodeIds = (selection?.nodes || []).map(node => node.id);
    const edgeIds = (selection?.edges || []).map(edge => edge.id);
    if (!sameCanvasSelection(selectedIds, nodeIds)) {
      setSelectedIds(nodeIds);
      onSelectionChange?.(nodeIds);
    }
    if (!sameCanvasSelection(selectedEdgeIds, edgeIds)) setSelectedEdgeIds(edgeIds);
  }

  function focusCanvasAfterSelection() {
    window.requestAnimationFrame(() => rootRef.current?.focus({ preventScroll: true }));
  }

  function selectCanvasEdge(edgeID) {
    const nextNodes = nodesRef.current.map(node => ({ ...node, selected: false }));
    const nextEdges = edgesRef.current.map(edge => ({ ...edge, selected: edge.id === edgeID }));
    nodesRef.current = nextNodes;
    edgesRef.current = nextEdges;
    setNodes(nextNodes);
    setEdges(nextEdges);
    setSelectedIds([]);
    setSelectedEdgeIds([edgeID]);
    onSelectionChange?.([]);
  }

  function handleEdgeClick(event, edge) {
    focusCanvasAfterSelection();
    setContextMenu(null);
    setAssetPanel(null);
    if (activeEditorNodeID) setActiveEditorNodeID('');
    window.requestAnimationFrame(() => selectCanvasEdge(edge.id));
  }

  function handleNodeClick(event, node) {
    // Controls inside a node can opt out of the canvas-level click cleanup.
    const isMultiSelection = isCanvasMultiSelectionEvent(event);
    if (event.defaultPrevented || (!isMultiSelection && event.target?.closest?.('.agent-node-composer'))) return;
    focusCanvasAfterSelection();
    setContextMenu(null);
    setAssetPanel(null);
    if (activeEditorNodeID && activeEditorNodeID !== node.id) setActiveEditorNodeID('');
    const next = selectCanvasNodes(
      nodesRef.current,
      node.id,
      isMultiSelection,
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
    if (!taskId || !nodesRef.current.some(node => node.id === nodeID && normalizeCanvasNodeType(node.data?.type) === 'image')) return;
    event?.preventDefault();
    event?.stopPropagation();
    const panel = { assetKind: 'image', mode, nodeID, position: { x: 0, y: 0 }, anchor: imageAssetPanelAnchor(event) };
    // React Flow completes its node click after this nested control; open after that cleanup.
    window.requestAnimationFrame(() => setAssetPanel(panel));
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
    if (!taskId || !['image', 'video', 'audio', 'director'].includes(node?.data?.type)) return;
    event?.preventDefault();
    event?.stopPropagation();
    const assetKind = node.data.type === 'image' || node.data.type === 'director' ? 'image' : node.data.type === 'audio' ? 'audio' : '';
    const panel = { assetKind, mode: 'node-reference', nodeID, position: { x: 0, y: 0 }, anchor: imageAssetPanelAnchor(event) };
    window.requestAnimationFrame(() => setAssetPanel(panel));
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
      ? removeReferenceAsset({ node, assetID })
      : node);
    commit(next, edgesRef.current);
  }

  function handleNodeConfigChange(nodeID, patch) {
    const selected = nodesRef.current.find(node => node.id === nodeID);
    const kind = normalizeCanvasNodeType(selected?.data?.type);
    const next = Object.hasOwn(patch, 'provider')
      ? shareModelSelection(nodesRef.current, nodeID, patch)
      : nodesRef.current.map(node => node.id === nodeID
        ? { ...node, data: { ...node.data, config: patchNodeConfig(kind, node.data?.config || {}, patch) } }
        : node);
    commit(next, edgesRef.current);
  }

  function changeScriptSource(nodeID, sourceText) {
    const currentNode = nodesRef.current.find(node => node.id === nodeID && node.data?.type === 'script');
    if (!currentNode || String(currentNode.data?.config?.sourceText || '') === String(sourceText ?? '')) return;
    const next = nodesRef.current.map(node => node.id === nodeID
      ? { ...node, data: { ...node.data, config: invalidateScriptSplitForSource(node.data?.config || {}, sourceText) } }
      : node);
    setScriptShotsByNode(current => ({
      ...current,
      [nodeID]: { state: 'stale', errorMessage: '', generationId: '', shots: [] }
    }));
    onScriptShotsChange?.(nodeID, []);
    commit(next, edgesRef.current, { record: false });
  }

  async function commitScriptSource(nodeID, sourceText) {
    const requestedTaskId = taskIdRef.current;
    const node = nodesRef.current.find(item => item.id === nodeID && item.data?.type === 'script');
    const config = node?.data?.config || {};
    if (!requestedTaskId || config.source !== 'batch_factory' || config.resourceType !== 'book_script') return;
    if (scriptSourceSaveByNodeRef.current[nodeID]) return scriptSourceSaveByNodeRef.current[nodeID];
    const save = (async () => {
      try {
        const saved = await updateBookSource(config.batchId, config.bookId, {
          sourceText: String(sourceText || '').trim(),
          replaceSource: true,
          expectedRevision: Number(config.resourceRevision || 0)
        });
        if (requestedTaskId !== taskIdRef.current) return;
        await syncAgentBatchFactoryBook(requestedTaskId, config.bookId);
        if (requestedTaskId === taskIdRef.current) await onServerCanvasChange?.();
        return saved?.book || saved;
      } catch (error) {
        if (requestedTaskId === taskIdRef.current) {
          setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], error: error?.message || '剧本正文保存失败' } }));
        }
        throw error;
      }
    })();
    scriptSourceSaveByNodeRef.current[nodeID] = save;
    try {
      return await save;
    } finally {
      if (scriptSourceSaveByNodeRef.current[nodeID] === save) delete scriptSourceSaveByNodeRef.current[nodeID];
    }
  }

  async function runBatchFactoryScriptStageForNode(nodeID, stage, textModelID = '', options = {}) {
    const requestedTaskId = taskIdRef.current;
    const savedSource = await scriptSourceSaveByNodeRef.current[nodeID];
    const node = nodesRef.current.find(item => item.id === nodeID && item.data?.type === 'script');
    const action = batchFactoryMappedAction(node);
    const config = node?.data?.config || {};
    if (!requestedTaskId || action?.type !== 'script') return false;
    setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: true, actionBusy: true, error: '' } }));
    try {
      const selectedModelID = String(textModelID || config.modelId || config.textModelId || '').trim();
      let currentRevision = latestBookRevision(savedSource, config.resourceRevision);
      const contentRangeLines = Math.min(500, Math.max(1, Number(options.contentRangeLines) || Number(config.contentRangeLines) || 5));
      if (selectedModelID) {
        const savedSettings = await saveBookOverride(action.batchId, action.bookId, {
          patch: { textModelId: selectedModelID },
          expectedRevision: currentRevision
        });
        currentRevision = latestBookRevision(savedSettings, currentRevision);
      }
      if (stage === 'settings' || contentRangeLines !== Number(config.contentRangeLines || 5)) {
        const savedMetadata = await updateBookMetadata(action.batchId, action.bookId, {
          metadataPatch: { contentRangeLines },
          expectedRevision: currentRevision
        });
        currentRevision = latestBookRevision(savedMetadata, currentRevision);
      }
      if (stage === 'viral') {
        const result = await rewriteWorkingFront(action.batchId, action.bookId, String(options.sourceText || config.sourceText || ''), {
          textModelId: selectedModelID
        });
        if (requestedTaskId !== taskIdRef.current) return null;
        await syncAgentBatchFactoryBook(requestedTaskId, action.bookId);
        if (requestedTaskId === taskIdRef.current) await onServerCanvasChange?.();
        return { candidate: String(result?.candidate || '').trim() };
      }
      if (stage === 'settings') {
        await syncAgentBatchFactoryBook(requestedTaskId, action.bookId);
        if (requestedTaskId === taskIdRef.current) await onServerCanvasChange?.();
        return true;
      }
      await runMappedBatchFactoryScriptStage({
        action,
        stage,
        requestId: mappedActionRequestID(action),
        options: { contentRangeLines },
        runStage: runBookStage
      });
      if (requestedTaskId !== taskIdRef.current) return true;
      await syncAgentBatchFactoryBook(requestedTaskId, action.bookId);
      if (requestedTaskId === taskIdRef.current) await onServerCanvasChange?.();
      return true;
    } catch (error) {
      if (requestedTaskId === taskIdRef.current) {
        setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], error: error?.message || '批量工厂步骤执行失败' } }));
      }
      throw error;
    } finally {
      if (requestedTaskId === taskIdRef.current) {
        setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: false, actionBusy: false } }));
      }
    }
  }

  async function splitScriptForNode(nodeID, input) {
    const requestedTaskId = taskId;
    if (!requestedTaskId) return;
    setScriptShotsByNode(current => ({ ...current, [nodeID]: { ...(current[nodeID] || {}), state: 'running', errorMessage: '' } }));
    try {
      const result = await runAfterDurableCanvasSave(onFlushCanvas, () => {
        const node = nodesRef.current.find(item => item.id === nodeID && item.data?.type === 'script');
        return createAgentScriptSplit(requestedTaskId, {
          nodeId: nodeID,
          nodeVersion: node?.data?.version || input.nodeVersion || 1,
          modelId: input.modelId
        });
      });
      if (requestedTaskId !== taskIdRef.current) return;
      const currentSource = String(nodesRef.current.find(item => item.id === nodeID)?.data?.config?.sourceText || '');
      if (currentSource !== String(input.sourceText || '')) {
        setScriptShotsByNode(current => ({ ...current, [nodeID]: { state: 'stale', errorMessage: '', generationId: '', shots: [] } }));
        return;
      }
      const normalized = normalizeScriptShots(result);
      setScriptShotsByNode(current => ({ ...current, [nodeID]: normalized }));
      onScriptShotsChange?.(nodeID, normalized.shots);
      mergeGenerationForProject(requestedTaskId, result.generation);
      handleNodeConfigChange(nodeID, {
        sourceText: input.sourceText,
        splitModelId: input.modelId,
        splitState: { state: normalized.state, errorMessage: normalized.errorMessage, generationId: normalized.generationId, sourceMarker: scriptSourceMarker(input.sourceText) }
      });
    } catch (error) {
      if (requestedTaskId !== taskIdRef.current) return;
      const saveFailed = error?.code === 'CANVAS_SAVE_REQUIRED';
      const errorMessage = error?.message || (saveFailed ? '画布保存失败，未开始拆分' : '剧本拆分失败');
      setScriptShotsByNode(current => ({ ...current, [nodeID]: { state: saveFailed ? 'stale' : 'failed', errorMessage, generationId: '', shots: [] } }));
      if (!saveFailed) {
        handleNodeConfigChange(nodeID, { splitState: { state: 'failed', errorMessage, generationId: '', sourceMarker: scriptSourceMarker(input.sourceText) } });
        void refreshGenerationTasks(requestedTaskId);
      }
    }
  }

  async function batchShotsForNode(nodeID, payload) {
    const requestedTaskId = taskId;
    if (!requestedTaskId) return null;
    const result = await runAfterDurableCanvasSave(onFlushCanvas, () => {
      const node = nodesRef.current.find(item => item.id === nodeID && item.data?.type === 'script');
      if (!node || !scriptSplitReady(node.data?.config || {})) {
        const error = new Error('剧本正文已变化，请重新拆分后再批量生成');
        error.code = 'SCRIPT_SPLIT_STALE';
        throw error;
      }
      return createAgentShotBatchNodes(requestedTaskId, payload);
    });
    if (requestedTaskId !== taskIdRef.current) return null;
    applyGeneratedCanvas(result.canvas);
    void refreshGenerationTasks(requestedTaskId);
    if (payload.kind === 'video') {
      listAgentVideoJobs(requestedTaskId).then(response => {
        if (shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current })) {
          setVideoJobs(current => mergeLatestVideoJobsByNode(current, response?.jobs, requestedTaskId));
        }
      }).catch(() => {});
    }
    return result;
  }

  function applyGeneratedCanvas(serverCanvas) {
    if (!serverCanvas) return;
    const nextAppearance = normalizeCanvasAppearance(serverCanvas.appearance || appearanceRef.current);
    const mappedNodes = (serverCanvas.nodes || []).map(toFlowNode);
    const nextEdges = (serverCanvas.edges || []).map(edge => toFlowEdge(edge, nextAppearance));
    const nextNodes = syncCanvasEffectiveReferences(mappedNodes, nextEdges);
    const nextViewport = serverCanvas.viewport || viewportRef.current;
    setNodes(nextNodes);
    setEdges(nextEdges);
    setAppearance(nextAppearance);
    setViewport(nextViewport);
    updateRefs(nextNodes, nextEdges, nextViewport, nextAppearance);
    historyRef.current.push({ nodes: nextNodes, edges: nextEdges });
    setHistoryVersion(value => value + 1);
    onChange?.(toCanvas(nextNodes, nextEdges, serverCanvas.revision, nextViewport, nextAppearance, serverCanvas.viewState || currentCanvas.viewState));
    onServerCanvasChange?.(serverCanvas);
  }

  function mergeGenerationForProject(projectID, generation) {
    if (!generation || !shouldAcceptGenerationResponse({ requestedTaskId: projectID, currentTaskId: taskIdRef.current })) return;
    setGenerationTasksByNode(current => mergeLatestGenerationTasksByNode(current, [generation], projectID));
    void refreshGenerationSummary(projectID);
  }

  async function refreshGenerationSummary(projectID, active = true) {
    try {
      const result = await getAgentGenerationSummary(projectID);
      if (!shouldAcceptGenerationResponse({ requestedTaskId: projectID, currentTaskId: taskIdRef.current, active })) return;
      setGenerationSummary(result?.summary || null);
    } catch (_) {
      // The per-node list remains usable when the optional compact summary is unavailable.
    }
  }

  async function refreshGenerationTasks(projectID) {
    try {
      const result = await listAgentGenerations(projectID);
      if (!shouldAcceptGenerationResponse({ requestedTaskId: projectID, currentTaskId: taskIdRef.current })) return;
      setGenerationTasksByNode(current => mergeLatestGenerationTasksByNode(current, result?.tasks, projectID));
      setGenerationTasksLoadedTaskID(projectID);
      setGenerationTasksLoadError('');
	  void refreshGenerationSummary(projectID);
	  return result?.tasks || [];
    } catch (error) {
      if (shouldAcceptGenerationResponse({ requestedTaskId: projectID, currentTaskId: taskIdRef.current })) {
        setGenerationTasksLoadError(error?.message || '生成任务状态加载失败，请刷新重试');
      }
	  return [];
    }
  }

  function mappedActionRequestID(action) {
    const random = globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(16).slice(2)}`;
    return `agent-batch-factory-${action.type}-${action.resourceId || 'book'}-${random}`;
  }

  function mappedImageAssetIDsForVideo(nodeID) {
    const ids = [];
    for (const edge of edgesRef.current) {
      if (edge.target !== nodeID) continue;
      const source = nodesRef.current.find(item => item.id === edge.source);
      const config = source?.data?.config || {};
      if (source?.data?.type !== 'image' || config.source !== 'batch_factory' || config.resourceType !== 'book_asset') continue;
      const assetID = String(config.resourceId || '').trim();
      if (assetID && !ids.includes(assetID)) ids.push(assetID);
    }
    return ids;
  }

  async function runMappedBatchFactoryAction(node, selectedModel = '', draftPrompt = '') {
    const action = batchFactoryMappedAction(node);
    if (!action) return false;
    const requestedTaskId = taskIdRef.current;
    if (!requestedTaskId) return true;
    const requestId = mappedActionRequestID(action);
    const config = node.data?.config || {};
    const prompt = String(draftPrompt || config.prompt || '').trim();
    const selectedProvider = String(config.provider || action.provider || '').trim();
    const selectedModelID = String(selectedModel || config.modelId || action.modelId || '').trim();
    try {
      switch (action.type) {
        case 'image':
          await updateBookAsset(action.batchId, action.bookId, action.resourceId, {
            name: String(config.assetName || node.data?.title || '').trim(),
            formalName: String(config.formalName || '').trim(),
            facts: String(config.facts || '').trim(),
            appearance: String(config.appearance || '').trim(),
            relationships: String(config.relationships || '').trim(),
            prompt,
            imageModelId: selectedModelID,
            imageAspectRatio: String(config.imageAspectRatio || config.ratio || '').trim(),
            expectedRevision: Number(config.resourceRevision || 0)
          });
          await generateBookAssetImages(action.batchId, action.bookId, { assetIds: [action.resourceId], ...(selectedModelID ? { modelId: selectedModelID } : {}), ...(selectedProvider ? { provider: selectedProvider } : {}) });
          break;
        case 'video':
          {
            const patch = {
              ...(prompt ? { videoPrompt: prompt } : {}),
              ...(selectedModelID ? { videoModelId: selectedModelID } : {}),
              ...(selectedProvider ? { videoProvider: selectedProvider } : {}),
              ...(action.durationSeconds > 0 ? { duration: action.durationSeconds } : {})
            };
            const selectedAssetIDs = mappedImageAssetIDsForVideo(node.id);
            if (selectedAssetIDs.length) patch.assetSelection = { autoAssetIds: [], addedAssetIds: selectedAssetIDs, excludedAssetIds: [] };
            if (Object.keys(patch).length) {
              await saveVideoOverride(action.batchId, action.bookId, action.resourceId, {
                patch, expectedRevision: action.expectedRevision
              });
            }
          }
          await runMappedBatchFactoryVideoProduction({
            action: { ...action, provider: selectedProvider || action.provider, modelId: selectedModelID || action.modelId },
            requestId,
            runStage: runBookStage
          });
          break;
        case 'merge':
          if (!action.canMerge) throw new Error(action.mergeReason || '等待所有分镜视频生成完成');
          await submitBookMerge(action.batchId, action.bookId, { requestId });
          break;
        case 'audio':
          await submitBookMerge(action.batchId, action.bookId, { requestId, audioArtifactId: action.resourceId });
          break;
        default:
          break;
      }
      if (requestedTaskId !== taskIdRef.current) return true;
      await syncAgentBatchFactoryBook(requestedTaskId, action.bookId);
      if (requestedTaskId !== taskIdRef.current) return true;
      await onServerCanvasChange?.();
      setNodeGenerationByNode(current => ({ ...current, [node.id]: { ...current[node.id], busy: false, actionBusy: false, error: '' } }));
      setVideoBusyByNode(current => ({ ...current, [node.id]: false }));
    } catch (error) {
      if (requestedTaskId === taskIdRef.current) {
        const messageText = error?.message || '批量工厂操作失败';
        setNodeGenerationByNode(current => ({ ...current, [node.id]: { ...current[node.id], busy: false, actionBusy: false, error: messageText } }));
        setVideoBusyByNode(current => ({ ...current, [node.id]: false }));
        setVideoActionErrors(current => ({ ...current, [node.id]: messageText }));
      }
    }
    return true;
  }

  async function submitTextForNode(nodeID, selectedModel, draftPrompt) {
    const requestedTaskId = taskId;
    const node = persistGenerationDraftPrompt(nodeID, draftPrompt);
    if (!requestedTaskId || node?.data?.type !== 'text') return;
    const capabilityError = nodeModelInputCapabilityError(node, selectedModel, configuredTextModels);
    if (capabilityError) {
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: false, error: capabilityError } }));
      return;
    }
    try {
      if (!await validateGenerationInputs(nodeID, selectedModel)) return;
    } catch (error) {
      if (requestedTaskId === taskIdRef.current) setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: false, error: error?.message || '画布输入校验失败' } }));
      return;
    }
    const attempt = nextNodeGenerationAttempt({ localAttempt: nodeGenerationByNode[nodeID]?.attempt, task: generationTasksByNode[nodeID], nodeId: nodeID, kind: 'text' });
    setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: true, error: '' } }));
    try {
      const result = await createAgentTextGeneration(requestedTaskId, buildNodeGenerationRequest(node, {
        taskId: requestedTaskId, attempt, modelId: selectedModel
      }));
      if (requestedTaskId !== taskIdRef.current) return;
      mergeGenerationForProject(requestedTaskId, result.generation);
      applyGeneratedCanvas(result.canvas);
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { busy: false, error: '', attempt } }));
    } catch (error) {
      if (requestedTaskId !== taskIdRef.current) return;
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { busy: false, error: error?.message || '文字生成失败', attempt: attempt + 1 } }));
      void refreshGenerationTasks(requestedTaskId);
    }
  }

  async function submitImageForNode(nodeID, selectedModel, draftPrompt) {
    const requestedTaskId = taskId;
    const mappedNode = nodesRef.current.find(item => item.id === nodeID);
    if (isBatchFactoryNode(mappedNode?.data)) {
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: true, error: '' } }));
      await runMappedBatchFactoryAction(mappedNode, selectedModel, draftPrompt);
      return;
    }
    const node = persistGenerationDraftPrompt(nodeID, draftPrompt);
    if (!requestedTaskId || node?.data?.type !== 'image') return;
    const capabilityError = nodeModelInputCapabilityError(node, selectedModel, configuredImageModels);
    if (capabilityError) {
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: false, error: capabilityError, phase: '' } }));
      return;
    }
    try {
      if (!await validateGenerationInputs(nodeID, selectedModel)) return;
    } catch (error) {
      if (requestedTaskId === taskIdRef.current) setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: false, error: error?.message || '画布输入校验失败', phase: '' } }));
      return;
    }
    const latestTasks = await refreshGenerationTasks(requestedTaskId);
    if (requestedTaskId !== taskIdRef.current) return;
    const latestTask = latestGenerationTaskForNode(latestTasks, nodeID, 'image');
    const attempt = nextNodeGenerationAttempt({ localAttempt: nodeGenerationByNode[nodeID]?.attempt, task: latestTask || generationTasksByNode[nodeID], nodeId: nodeID, kind: 'image' });
    setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: true, error: '', phase: '正在提交生成请求' } }));
    window.setTimeout(() => {
      if (requestedTaskId !== taskIdRef.current) return;
      void refreshGenerationTasks(requestedTaskId);
      setNodeGenerationByNode(current => current[nodeID]?.busy
        ? { ...current, [nodeID]: { ...current[nodeID], phase: '正在等待生成服务返回' } }
        : current);
    }, 800);
    try {
      const result = await createAgentImageGeneration(requestedTaskId, buildNodeGenerationRequest(node, {
        taskId: requestedTaskId, attempt, modelId: selectedModel
      }));
      if (requestedTaskId !== taskIdRef.current) return;
      const queuedGeneration = result.generation;
      const candidates = Array.isArray(result.resultAssets) ? result.resultAssets : [];
      mergeGenerationForProject(requestedTaskId, queuedGeneration);
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: {
        busy: false, error: '', phase: '', attempt, generationId: queuedGeneration?.id || '', executionId: result.execution?.id || '', status: queuedGeneration?.state || result.execution?.status || 'queued',
        candidates, selectedAssetId: candidates[0]?.id || ''
      } }));
      onExecutionChange?.(result.execution);
      onAssetsCreated?.(candidates);
    } catch (error) {
      if (requestedTaskId !== taskIdRef.current) return;
      const tasks = await refreshGenerationTasks(requestedTaskId);
      const activeTask = activeGenerationTaskForNode(tasks, nodeID, 'image');
      const latestTask = latestGenerationTaskForNode(tasks, nodeID, 'image');
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: activeTask
        ? { ...current[nodeID], busy: false, error: '', phase: '', attempt, generationId: activeTask.id, status: activeTask.state }
        : latestTask?.state === 'failed'
          ? { ...current[nodeID], busy: false, error: '', phase: '', attempt: attempt + 1, generationId: latestTask.id, status: latestTask.state }
          : { ...current[nodeID], busy: false, error: error?.message || '图片生成失败', phase: '', attempt: attempt + 1 }
      }));
    }
  }

  async function submitImageTool(nodeID, input) {
    const requestedTaskId = taskId;
    const node = nodesRef.current.find(item => item.id === nodeID);
    if (!requestedTaskId || node?.data?.type !== 'image' || !input?.sourceAssetId) return;
    const latestTasks = await refreshGenerationTasks(requestedTaskId);
    if (requestedTaskId !== taskIdRef.current) return;
    const latestTask = latestGenerationTaskForNode(latestTasks, nodeID, 'image');
    const attempt = nextNodeGenerationAttempt({ localAttempt: nodeGenerationByNode[nodeID]?.attempt, task: latestTask || generationTasksByNode[nodeID], nodeId: nodeID, kind: 'image' });
    setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: true, error: '', phase: '正在提交图片编辑任务' } }));
    try {
      const task = await createAgentImageTool(requestedTaskId, {
        ...input,
        nodeId: nodeID,
        nodeVersion: Number(node.data.version) || 0,
        idempotencyKey: `agent-image-tool:${requestedTaskId}:${nodeID}:${Number(node.data.version) || 0}:${input.operation}:${Date.now()}:${attempt}`
      });
      if (requestedTaskId !== taskIdRef.current) return;
      mergeGenerationForProject(requestedTaskId, task.task);
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: {
        ...current[nodeID], busy: false, error: '', phase: '', attempt, generationId: task.task?.id || '', status: task.task?.state || 'queued', candidates: [], selectedAssetId: ''
      } }));
    } catch (error) {
      if (requestedTaskId !== taskIdRef.current) return;
      const tasks = await refreshGenerationTasks(requestedTaskId);
      const activeTask = activeGenerationTaskForNode(tasks, nodeID, 'image');
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: activeTask
        ? { ...current[nodeID], busy: false, error: '', phase: '', attempt, generationId: activeTask.id, status: activeTask.state }
        : { ...current[nodeID], busy: false, error: error?.message || '图片编辑提交失败', phase: '', attempt: attempt + 1 }
      }));
    }
  }

  async function submitVideoTool(nodeID, input) {
    const requestedTaskId = taskId;
    const source = nodesRef.current.find(item => item.id === nodeID && item.data?.type === 'video');
    const sourceAssetId = String(input?.sourceAssetId || '').trim();
    const targetKind = input?.operation === 'extract_frames' ? 'image' : input?.operation === 'separate_audio' ? 'audio' : 'video';
    if (!requestedTaskId || !source || !sourceAssetId) return;
    const labels = { trim: '视频剪辑结果', crop: '视频裁剪结果', extract_frames: '视频抽帧结果', separate_audio: '视频原声音频' };
    const target = createNodeAt({
      id: nextCanvasID(targetKind), kind: targetKind,
      position: { x: (Number(source.position?.x) || 0) + 390, y: Number(source.position?.y) || 0 },
      config: inheritedModelSelection(nodesRef.current, targetKind)
    });
    target.data.title = labels[input.operation] || '视频处理结果';
    target.data.sourceNodeIds = [source.id];
    const edge = toFlowEdge({ id: `edge-${source.id}-${target.id}-processing`, source: source.id, target: target.id, relation: 'reference' }, appearanceRef.current);
    const nextEdges = edgesRef.current.concat(edge);
    const nextNodes = syncCanvasEffectiveReferences(nodesRef.current.map(node => ({ ...node, selected: false })).concat(target), nextEdges);
    setSelectedIds([target.id]);
    onSelectionChange?.([target.id]);
    commit(nextNodes, nextEdges);
    setNodeGenerationByNode(current => ({ ...current, [target.id]: { busy: true, error: '', phase: '正在保存视频处理任务' } }));
    try {
      const result = await runAfterDurableCanvasSave(onFlushCanvas, () => createAgentVideoTool(requestedTaskId, {
        ...input,
        nodeId: target.id,
        nodeVersion: Number(target.data.version) || 1,
        idempotencyKey: `agent-video-tool:${requestedTaskId}:${target.id}:${input.operation}:${Date.now()}`
      }));
      if (requestedTaskId !== taskIdRef.current) return;
      mergeGenerationForProject(requestedTaskId, result.task);
      setNodeGenerationByNode(current => ({ ...current, [target.id]: { busy: false, error: '', phase: '', generationId: result.task?.id || '', status: result.task?.state || 'queued', candidates: [], selectedAssetId: '' } }));
      setActiveEditorNodeID(target.id);
    } catch (error) {
      if (requestedTaskId !== taskIdRef.current) return;
      setNodeGenerationByNode(current => ({ ...current, [target.id]: { ...current[target.id], busy: false, error: error?.message || '视频处理任务提交失败', phase: '' } }));
    }
  }

  function selectImageCandidate(nodeID, assetID) {
    setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], selectedAssetId: assetID } }));
  }

  async function confirmImage(nodeID, assetID) {
    const requestedTaskId = taskId;
    const mappedNode = nodesRef.current.find(item => item.id === nodeID);
    const mappedAction = batchFactoryMappedAction(mappedNode);
    if (mappedAction?.type === 'image' && assetID) {
      try {
        await setPrimaryBookAssetImage(mappedAction.batchId, mappedAction.bookId, mappedAction.resourceId, assetID);
        if (requestedTaskId === taskIdRef.current) {
          await syncAgentBatchFactoryBook(requestedTaskId, mappedAction.bookId);
          await onServerCanvasChange?.();
        }
      } catch (error) {
        if (requestedTaskId === taskIdRef.current) setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: false, error: error?.message || '设置批量工厂主版本失败' } }));
      }
      return;
    }
    const local = nodeGenerationByNode[nodeID] || {};
    const restored = nodeExecutionReview(executions, nodeID) || {};
    const executionID = local.executionId || restored.executionId;
    const task = generationTasksByNode[nodeID];
    const generationID = local.generationId || task?.id;
    const resultID = task?.results?.find(item => item.assetId === assetID || item.id === assetID)?.id || assetID;
    if (!requestedTaskId || (!generationID && !executionID) || !assetID) return;
    setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: true, error: '' } }));
    try {
      const result = generationID
        ? await confirmAgentGeneration(requestedTaskId, generationID, resultID)
        : await confirmAgentImageGeneration(requestedTaskId, executionID, assetID);
      if (requestedTaskId !== taskIdRef.current) return;
      mergeGenerationForProject(requestedTaskId, result.task || result.generation);
      applyGeneratedCanvas(result.canvas);
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { busy: false, error: '', attempt: Number(current[nodeID]?.attempt) || 0, status: 'completed', candidates: [], selectedAssetId: '' } }));
      onExecutionChange?.(result.execution);
    } catch (error) {
      if (requestedTaskId !== taskIdRef.current) return;
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: false, error: error?.message || '图片确认失败' } }));
    }
  }

  async function submitAudioForNode(nodeID, selectedModel, draftPrompt) {
    const requestedTaskId = taskId;
    const mappedNode = nodesRef.current.find(item => item.id === nodeID);
    if (isBatchFactoryNode(mappedNode?.data)) {
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: true, error: '' } }));
      await runMappedBatchFactoryAction(mappedNode, selectedModel, draftPrompt);
      return;
    }
    const node = persistGenerationDraftPrompt(nodeID, draftPrompt);
    if (!requestedTaskId || node?.data?.type !== 'audio' || !canSubmitAudioNode(node, selectedModel)) return;
    const capabilityError = nodeModelInputCapabilityError(node, selectedModel, configuredAudioModels);
    if (capabilityError) {
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: false, error: capabilityError } }));
      return;
    }
    try {
      if (!await validateGenerationInputs(nodeID, selectedModel)) return;
    } catch (error) {
      if (requestedTaskId === taskIdRef.current) setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: false, error: error?.message || '画布输入校验失败' } }));
      return;
    }
    const attempt = nextNodeGenerationAttempt({ localAttempt: nodeGenerationByNode[nodeID]?.attempt, task: generationTasksByNode[nodeID], nodeId: nodeID, kind: 'audio' });
    setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: true, error: '' } }));
    try {
      const result = await createAgentAudioGeneration(requestedTaskId, audioJobRequestForNode(node, {
        taskId: requestedTaskId, attempt, modelId: selectedModel
      }));
      if (requestedTaskId !== taskIdRef.current) return;
      const task = result.task;
      const candidates = (Array.isArray(result.assets) && result.assets.length ? result.assets : (task?.results || []).map(item => ({ ...item.asset, id: item.assetId, resultId: item.id })))
        .map((asset, index) => ({ ...asset, id: asset.id || asset.assetId, resultId: asset.resultId || task?.results?.[index]?.id || asset.id || asset.assetId }));
      mergeGenerationForProject(requestedTaskId, task);
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: {
        busy: false, error: '', attempt, generationId: task?.id || '', status: task?.state || 'awaiting_confirmation',
        candidates, selectedAssetId: candidates[0]?.id || ''
      } }));
      onAssetsCreated?.(candidates);
    } catch (error) {
      if (requestedTaskId !== taskIdRef.current) return;
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: false, error: error?.message || '音频任务执行失败', attempt: attempt + 1 } }));
      void refreshGenerationTasks(requestedTaskId);
    }
  }

  function selectAudioCandidate(nodeID, assetID) {
    setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], selectedAssetId: assetID } }));
  }

  async function confirmAudio(nodeID, assetID) {
    const requestedTaskId = taskId;
    const local = nodeGenerationByNode[nodeID] || {};
    const task = generationTasksByNode[nodeID];
    const generationID = local.generationId || task?.id;
    const candidate = (local.candidates || []).find(item => item.id === assetID);
    const resultID = candidate?.resultId || task?.results?.find(item => item.assetId === assetID || item.id === assetID)?.id || assetID;
    if (!requestedTaskId || !generationID || !resultID) return;
    setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: true, error: '' } }));
    try {
      const result = await confirmAgentGeneration(requestedTaskId, generationID, resultID);
      if (requestedTaskId !== taskIdRef.current) return;
      mergeGenerationForProject(requestedTaskId, result.task);
      applyGeneratedCanvas(result.canvas);
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { busy: false, error: '', attempt: Number(current[nodeID]?.attempt) || 0, status: 'completed', candidates: [], selectedAssetId: '' } }));
    } catch (error) {
      if (requestedTaskId !== taskIdRef.current) return;
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], busy: false, error: error?.message || '音频确认失败' } }));
    }
  }

  function makeGenerationRetryIdempotencyKey() {
    const random = globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(16).slice(2)}`;
    return `agent-generation-retry-${random}`;
  }

  async function cancelGeneration(nodeID) {
    const requestedTaskId = taskId;
    const task = generationTasksByNode[nodeID];
    if (!requestedTaskId || !task || (task.kind === 'video' && !String(task.mode || '').startsWith('video_'))) return;
    setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], actionBusy: true, error: '' } }));
    try {
      const result = await cancelAgentGeneration(requestedTaskId, task.id);
      if (requestedTaskId !== taskIdRef.current) return;
      const nextTask = result.task || result.generation;
      mergeGenerationForProject(requestedTaskId, nextTask);
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], actionBusy: false, error: '', generationId: nextTask?.id || task.id, status: nextTask?.state || 'cancelled' } }));
    } catch (error) {
      if (requestedTaskId !== taskIdRef.current) return;
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], actionBusy: false, error: error?.message || '取消生成任务失败' } }));
    }
  }

  async function retryGeneration(nodeID) {
    const requestedTaskId = taskId;
    const mappedNode = nodesRef.current.find(item => item.id === nodeID);
    if (isBatchFactoryNode(mappedNode?.data)) {
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], actionBusy: true, error: '' } }));
      await runMappedBatchFactoryAction(mappedNode);
      return;
    }
    const task = generationTasksByNode[nodeID];
    if (!requestedTaskId || !task || (task.kind === 'video' && !String(task.mode || '').startsWith('video_'))) return;
    setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], actionBusy: true, error: '' } }));
    try {
      const result = await retryAgentGeneration(requestedTaskId, task.id, makeGenerationRetryIdempotencyKey());
      if (requestedTaskId !== taskIdRef.current) return;
      const nextTask = result.task || result.generation;
      mergeGenerationForProject(requestedTaskId, nextTask);
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], actionBusy: false, error: '', generationId: nextTask?.id || '', status: nextTask?.state || 'queued' } }));
    } catch (error) {
      if (requestedTaskId !== taskIdRef.current) return;
      setNodeGenerationByNode(current => ({ ...current, [nodeID]: { ...current[nodeID], actionBusy: false, error: error?.message || '重试生成任务失败' } }));
    }
  }

  function videoRequestPayloadForNode(node, ignoreReferences = false) {
    return {
      ...buildNodeGenerationRequest(node, { taskId }),
      idempotencyKey: makeVideoJobIdempotencyKey(),
      prompt: node?.data?.config?.prompt || node?.data?.summary || '',
      referenceAssetIds: videoReferenceAssetIDs(node?.data?.config || {}),
      ignoreReferences
    };
  }

  async function submitVideoForNode(nodeID, ignoreReferences = false, selectedModel = '', draftPrompt) {
    const requestedTaskId = taskId;
    const mappedNode = nodesRef.current.find(item => item.id === nodeID);
    if (isBatchFactoryNode(mappedNode?.data)) {
      setVideoBusyByNode(current => ({ ...current, [nodeID]: true }));
      setVideoActionErrors(current => ({ ...current, [nodeID]: '' }));
      await runMappedBatchFactoryAction(mappedNode, selectedModel, draftPrompt);
      if (requestedTaskId === taskIdRef.current) setVideoBusyByNode(current => ({ ...current, [nodeID]: false }));
      return;
    }
    if (!requestedTaskId || videoJobsLoadedTaskID !== requestedTaskId) return;
    const node = persistGenerationDraftPrompt(nodeID, draftPrompt);
    if (!node || node.data?.type !== 'video') return;
    const capabilityError = nodeModelInputCapabilityError(node, selectedModel, configuredVideoModels);
    if (capabilityError) {
      setVideoActionErrors(current => ({ ...current, [nodeID]: capabilityError }));
      return;
    }
    const referenceError = videoReferenceCapabilityError(node.data?.config || {});
    if (referenceError) {
      setVideoActionErrors(current => ({ ...current, [nodeID]: referenceError }));
      return;
    }
    try {
      if (!await validateGenerationInputs(nodeID, selectedModel)) return;
    } catch (error) {
      if (requestedTaskId === taskIdRef.current) setVideoActionErrors(current => ({ ...current, [nodeID]: error?.message || '画布输入校验失败' }));
      return;
    }
    const payload = videoRequestPayloadForNode(node, ignoreReferences);
    if (selectedModel) payload.modelId = selectedModel;
    if (!canSubmitVideoFromConfig({ prompt: payload.prompt })) return;
    setVideoBusyByNode(current => ({ ...current, [nodeID]: true }));
    setVideoActionErrors(current => ({ ...current, [nodeID]: '' }));
    try {
      const result = await createAgentVideoJob(requestedTaskId, payload);
      if (!shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current })) return;
      setVideoJobs(current => mergeLatestVideoJobsByNode(current, [result.job], requestedTaskId));
      void refreshGenerationTasks(requestedTaskId);
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
    const mappedNode = nodesRef.current.find(item => item.id === nodeID);
    if (isBatchFactoryNode(mappedNode?.data)) {
      setVideoBusyByNode(current => ({ ...current, [nodeID]: true }));
      setVideoActionErrors(current => ({ ...current, [nodeID]: '' }));
      await runMappedBatchFactoryAction(mappedNode);
      return;
    }
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
      void refreshGenerationTasks(requestedTaskId);
      await onServerCanvasChange?.();
    } catch (error) {
      if (shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current })) setVideoActionErrors(current => ({ ...current, [nodeID]: error?.message || '确认回填失败' }));
    } finally {
      if (shouldAcceptVideoJobResponse({ requestedTaskId, currentTaskId: taskIdRef.current })) setVideoBusyByNode(current => ({ ...current, [nodeID]: false }));
    }
  }

  function handleConnect(connection) {
    connectionCompletedRef.current = true;
    const typedEdge = createTypedCanvasEdge(connection, nodesRef.current);
    if (!typedEdge) return;
    const next = connectCanvasNodes(edgesRef.current, typedEdge);
    if (next === edgesRef.current) return;
    const nextNodes = syncCanvasEffectiveReferences(nodesRef.current, next);
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
    const typedEdge = createTypedCanvasEdge(connection, nodesRef.current);
    return Boolean(typedEdge) && connectCanvasNodes(edgesRef.current, typedEdge) !== edgesRef.current;
  }

  function handleConnectionDropCreate(kind) {
    if (!connectionDropMenu) return;
    const nextNode = createNodeAt({ id: nextCanvasID(kind), kind, position: connectionDropMenu.position, config: inheritedModelSelection(nodesRef.current, kind) });
    const nextNodes = nodesRef.current.map(node => ({ ...node, selected: false })).concat(nextNode);
    const typedEdge = createTypedCanvasEdge({ source: connectionDropMenu.sourceID, target: nextNode.id }, nextNodes);
    const nextEdges = typedEdge ? connectCanvasNodes(edgesRef.current, typedEdge) : edgesRef.current;
    if (nextEdges === edgesRef.current) return;
    const syncedNodes = syncCanvasEffectiveReferences(nextNodes, nextEdges);
    setSelectedIds([nextNode.id]);
    onSelectionChange?.([nextNode.id]);
    commit(syncedNodes, nextEdges.map(edge => toFlowEdge(edge, appearanceRef.current)));
    setConnectionDropMenu(null);
  }

  function runNodeCommand(command) {
    const next = command(nodesRef.current);
    if (next !== nodesRef.current) commit(next, edgesRef.current);
  }

  function addNode(kind = 'script') {
    const bounds = rootRef.current?.getBoundingClientRect();
    const screenPoint = bounds ? { x: bounds.left + bounds.width / 2, y: bounds.top + bounds.height / 2 } : { x: 0, y: 0 };
    const position = flowInstance?.screenToFlowPosition(screenPoint) || { x: 0, y: 0 };
    const nextNode = createNodeAt({ id: nextCanvasID(kind), kind, position, config: inheritedModelSelection(nodesRef.current, kind) });
    const next = nodesRef.current.map(node => ({ ...node, selected: false })).concat(nextNode);
    setSelectedIds([nextNode.id]);
    onSelectionChange?.([nextNode.id]);
    commit(next, edgesRef.current);
  }

  function nextCanvasID(prefix = 'node') {
    nodeSequenceRef.current += 1;
    return `${prefix}-${Date.now()}-${nodeSequenceRef.current}`;
  }

  function openContextMenu(event, target, nodeIds = [], edgeIds = []) {
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
      edgeIds,
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

  function handleEdgeContextMenu(event, edge) {
    selectCanvasEdge(edge.id);
    openContextMenu(event, 'edges', [], [edge.id]);
  }

  function handleContextCreate(kind) {
    if (!contextMenu) return;
    const nextNode = createNodeAt({ id: nextCanvasID(kind), kind, position: contextMenu.position, config: inheritedModelSelection(nodesRef.current, kind) });
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
      commit(syncCanvasEffectiveReferences(next, edgesRef.current), edgesRef.current);
      const updatedNode = next.find(node => node.id === assetPanel.nodeID);
      if (updatedNode && manualNode?.id === updatedNode.id) {
        setManualNode(updatedNode);
        setManualNodeDraft(current => ({ ...current, config: { ...(updatedNode.data?.config || {}) } }));
      }
    } else if (assetPanel.mode === 'image-source') {
      const next = nodesRef.current.map(node => node.id === assetPanel.nodeID
        ? attachImageAsset({ node, asset })
        : node);
      commit(syncCanvasEffectiveReferences(next, edgesRef.current), edgesRef.current);
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
      const nextNode = createNodeAt({ id: nextCanvasID(asset.kind), kind: asset.kind, position: assetPanel.position, asset, config: inheritedModelSelection(nodesRef.current, asset.kind) });
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
    const mappedNodeIDs = ids.filter(nodeID => isBatchFactoryNode(nodesRef.current.find(node => node.id === nodeID)?.data));
    if (mappedNodeIDs.length) {
      const localNodeIDs = ids.filter(nodeID => !mappedNodeIDs.includes(nodeID));
      if (localNodeIDs.length) {
        const result = removeCanvasNodes({ nodes: nodesRef.current, edges: edgesRef.current, ids: localNodeIDs });
        setSelectedIds([]);
        onSelectionChange?.([]);
        commit(result.nodes, result.edges);
      }
      void removeBatchFactoryNodes(mappedNodeIDs);
      return true;
    }
    if (ids.includes(activeEditorNodeID)) closeNodeComposer();
    const result = removeCanvasNodes({ nodes: nodesRef.current, edges: edgesRef.current, ids });
    setSelectedIds([]);
    onSelectionChange?.([]);
    commit(result.nodes, result.edges);
    return true;
  }

  async function removeBatchFactoryNodes(nodeIDs) {
    const requestedTaskId = taskIdRef.current;
    const uniqueIDs = [...new Set((nodeIDs || []).map(value => String(value || '').trim()).filter(Boolean))];
    if (!requestedTaskId || !uniqueIDs.length) return;
    try {
      await removeBatchFactoryCanvasNodes({
        taskId: requestedTaskId,
        nodeIds: uniqueIDs,
        flush: onFlushCanvas,
        remove: removeAgentBatchFactoryNode,
        refresh: onServerCanvasChange,
        isCurrent: () => requestedTaskId === taskIdRef.current
      });
      if (requestedTaskId !== taskIdRef.current) return;
      message.success(uniqueIDs.length === 1 ? '已从画布移除书籍节点' : '已从画布移除书籍节点');
    } catch (error) {
      if (requestedTaskId === taskIdRef.current) message.error(error?.message || '从画布移除失败');
    }
  }

  function removeBatchFactoryNode(nodeID) {
    void removeBatchFactoryNodes([nodeID]);
  }

  function deleteCanvasEdges(ids = selectedEdgeIds) {
    if (!ids.length) return false;
    const nextEdges = removeCanvasEdges(edgesRef.current, ids);
    if (nextEdges === edgesRef.current) return false;
    setSelectedEdgeIds([]);
    commit(syncCanvasEffectiveReferences(nodesRef.current, nextEdges), nextEdges);
    return true;
  }

  function handleContextDelete() {
    if (!contextMenu) return;
    if (contextMenu.edgeIds?.length) deleteCanvasEdges(contextMenu.edgeIds);
    else if (contextMenu.nodeIds.length) deleteCanvasNodes(contextMenu.nodeIds);
    setContextMenu(null);
  }

  function handleContextDisconnect() {
    if (!contextMenu?.nodeIds.length) return;
    const ids = new Set(contextMenu.nodeIds);
    const nextEdges = edgesRef.current.filter(edge => !ids.has(edge.source) && !ids.has(edge.target));
    if (nextEdges.length !== edgesRef.current.length) {
      commit(syncCanvasEffectiveReferences(nodesRef.current, nextEdges), nextEdges);
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
      case 'delete':
        if (!deleteCanvasEdges()) deleteCanvasNodes();
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
    const suppressViewportCommit = suppressViewportCommitRef.current;
    if (suppressViewportCommit) {
      suppressViewportCommitRef.current = false;
      window.clearTimeout(viewportRecoveryTimerRef.current);
    }
    setViewport(nextViewport);
    viewportRef.current = nextViewport;
    onViewportChange?.(nextViewport);
    if (!suppressViewportCommit) onViewportCommit?.(nextViewport);
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

  function addAssetNode(asset, position) {
    if (!asset) return;
    const nextNode = createNodeAt({ id: nextCanvasID(asset.kind), kind: asset.kind, position, asset, config: inheritedModelSelection(nodesRef.current, asset.kind) });
    const nextNodes = nodesRef.current.map(node => ({ ...node, selected: false })).concat(nextNode);
    setSelectedIds([nextNode.id]);
    onSelectionChange?.([nextNode.id]);
    commit(nextNodes, edgesRef.current);
  }

  function openDirectorDesk(nodeID) {
    const director = nodesRef.current.find(node => node.id === nodeID && node.data?.type === 'director');
    if (director) setDirectorDeskNodeID(nodeID);
  }

  function commitDirectorCapture(nodeID, asset) {
    const next = insertDirectorCaptureState({
      nodes: nodesRef.current,
      edges: edgesRef.current,
      directorNodeID: nodeID,
      asset
    });
    setSelectedIds([next.imageNode.id]);
    onSelectionChange?.([next.imageNode.id]);
    commit(next.nodes, next.edges);
    message.success('3D 构图已加入项目素材和画布');
    return asset;
  }

  async function saveDirectorDeskScreenshot(nodeID, file) {
    const requestedTaskId = taskIdRef.current;
    const source = nodesRef.current.find(node => node.id === nodeID && node.data?.type === 'director');
    if (!requestedTaskId || !source || !(file instanceof File)) throw new Error('导演台预演图无法保存');
    try {
      const result = await uploadAgentAsset(requestedTaskId, file);
      if (requestedTaskId !== taskIdRef.current) throw new Error('项目已切换，未将构图加入当前画布');
      const asset = result?.asset;
      if (!asset?.id || asset.kind !== 'image') throw new Error('预演图保存失败');
      return commitDirectorCapture(nodeID, asset);
    } catch (error) {
      message.error(error?.message || '预演图保存失败');
      throw error;
    }
  }

  function canvasCenterPosition() {
    const bounds = rootRef.current?.getBoundingClientRect();
    if (!bounds || !flowInstance?.screenToFlowPosition) return { x: 80, y: 80 };
    return flowInstance.screenToFlowPosition({ x: bounds.left + bounds.width / 2, y: bounds.top + bounds.height / 2 });
  }

  function locateNode(nodeID) {
    const node = nodesRef.current.find(item => item.id === nodeID);
    if (!node) return false;
    const nextNodes = nodesRef.current.map(item => ({ ...item, selected: item.id === nodeID }));
    setNodes(nextNodes);
    nodesRef.current = nextNodes;
    setSelectedIds([nodeID]);
    onSelectionChange?.([nodeID]);
    const internalNode = flowInstance?.getInternalNode?.(nodeID);
    const nodePosition = internalNode?.internals?.positionAbsolute || node.position;
    const nodeType = normalizeCanvasNodeType(node.data?.type);
    const fallbackWidth = ['image', 'video', 'audio'].includes(nodeType) ? 340 : 280;
    const fallbackHeight = ['image', 'video'].includes(nodeType) ? 238 : nodeType === 'audio' ? 128 : 114;
    const nodeWidth = Number(internalNode?.measured?.width) || fallbackWidth;
    const nodeHeight = Number(internalNode?.measured?.height) || fallbackHeight;
    const currentZoom = Number(viewportRef.current?.zoom);
    const zoom = Number.isFinite(currentZoom) ? Math.min(1.25, Math.max(0.75, currentZoom)) : 1;
    flowInstance?.setCenter?.(nodePosition.x + nodeWidth / 2, nodePosition.y + nodeHeight / 2, { zoom, duration: 260 });
    return true;
  }

  async function handleExternalAssetDrop(event) {
    const payload = event.dataTransfer.getData('application/x-agent-asset');
    if (!payload || !taskId) return;
    event.preventDefault();
    try {
      const reference = JSON.parse(payload);
      const result = await listAgentAssets(taskId, reference.kind || '');
      const asset = (result.assets || []).find(item => item.id === reference.id);
      if (!asset) throw new Error('素材已不存在');
      const position = flowInstance?.screenToFlowPosition
        ? flowInstance.screenToFlowPosition({ x: event.clientX, y: event.clientY })
        : canvasCenterPosition();
      addAssetNode(asset, position);
    } catch (error) {
      message.error(error.message || '添加素材失败');
    }
  }

  useImperativeHandle(ref, () => ({
    locateNode,
    openDirectorDesk,
    createFromAsset(asset) {
      addAssetNode(asset, canvasCenterPosition());
    }
  }));

  const directorDeskAdapter = useMemo(() => {
    if (!directorDeskNodeID) return null;
    return createDirectorDeskHostAdapter({
      nodeID: directorDeskNodeID,
      getNode: nodeID => nodesRef.current.find(node => node.id === nodeID),
      saveProject: async (nodeID, project) => handleNodeConfigChange(nodeID, { directorDesk: project }),
      listImages: () => listAgentAssets(taskIdRef.current, 'image'),
      uploadCapture: async file => {
        const requestedTaskId = taskIdRef.current;
        const result = await uploadAgentAsset(requestedTaskId, file);
        if (requestedTaskId !== taskIdRef.current) throw new Error('项目已切换，无法保存 3D 构图');
        return result?.asset;
      },
      onClose: () => setDirectorDeskNodeID('')
    });
  }, [directorDeskNodeID]);

  const alignItems = [
    { key: 'left', label: '左对齐', icon: <AlignStartVertical size={15} /> },
    { key: 'centerX', label: '水平居中', icon: <AlignCenterVertical size={15} /> },
    { key: 'right', label: '右对齐', icon: <AlignEndVertical size={15} /> },
    { key: 'top', label: '顶对齐', icon: <AlignStartHorizontal size={15} /> },
    { key: 'centerY', label: '垂直居中', icon: <AlignCenterHorizontal size={15} /> },
    { key: 'bottom', label: '底对齐', icon: <AlignEndHorizontal size={15} /> }
  ];

  return (
    <div
      ref={rootRef}
      className={`agent-canvas${connectingSourceID ? ' is-connecting' : ''}`}
      aria-label="创作画布"
      tabIndex={0}
      onKeyDownCapture={handleKeyDown}
      onDragOver={event => {
        if (event.dataTransfer.types.includes('application/x-agent-asset')) event.preventDefault();
      }}
      onDrop={handleExternalAssetDrop}
      style={appearanceCssVariables(appearance)}
    >
      <CanvasNodeContext.Provider value={canvasNodeContextValue}>
      <ReactFlow
        nodes={visibleNodes}
        edges={visibleEdges}
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
        onEdgeClick={handleEdgeClick}
        onEdgeContextMenu={handleEdgeContextMenu}
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
          if (!viewportHasChanged(viewportRef.current, nextViewport)) return;
          setViewport(nextViewport);
          viewportRef.current = nextViewport;
          onViewportChange?.(nextViewport);
        }}
        onMoveEnd={handleMoveEnd}
        panOnDrag={interactionMode === 'pan' ? [0, 1] : [1]}
        nodesDraggable={interactionMode !== 'pan'}
        elementsSelectable={interactionMode !== 'pan'}
        selectionOnDrag={interactionMode === 'select'}
        snapToGrid={workspaceView.snapToGrid}
        snapGrid={[20, 20]}
        multiSelectionKeyCode="Shift"
        deleteKeyCode={null}
      >
        {appearance.backgroundDots ? <Background variant="dots" gap={20} size={1} /> : null}
        {workspaceView.showMinimap ? <MiniMap className="agent-canvas-minimap" nodeColor="#526277" maskColor="rgba(7, 13, 21, 0.72)" pannable zoomable /> : null}
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
              ><Folder size={12} />{batchFactoryGroupLabel(groupId)}</button>
              <Button size="small" aria-label={collapsedGroups.has(groupId) ? '展开分组' : '收起分组'} onClick={() => toggleGroup(groupId)}>{collapsedGroups.has(groupId) ? '展开' : '收起'}</Button>
              <Button size="small" aria-label="定位分组" onClick={() => locateGroup(groupId)}>定位</Button>
            </section>
          ))}
        </ViewportPortal>
      </ReactFlow>
      </CanvasNodeContext.Provider>

      {directorDeskAdapter ? <Suspense fallback={<div className="agent-director-desk-loading">正在加载 3D导演台…</div>}>
        <DirectorDeskApp
          adapter={directorDeskAdapter}
          onCapture={asset => commitDirectorCapture(directorDeskNodeID, asset)}
        />
      </Suspense> : null}

      {generationTasksLoadError ? <p className="agent-canvas-global-error" role="alert">{generationTasksLoadError}</p> : null}
      {generationSummary ? <p className="agent-canvas-generation-summary" role="status" aria-live="polite">{generationSummaryText(generationSummary)}</p> : null}

      {connectionDropMenu ? <section className="agent-canvas-connection-drop-menu" aria-label="新建后续节点" style={{ left: connectionDropMenu.left, top: connectionDropMenu.top }}>
        <strong>新建后续节点</strong>
        <div>
          {connectionDropChoices(nodes.find(sourceNode => sourceNode.id === connectionDropMenu.sourceID)).map(choice => (
            <button type="button" key={choice.kind} onClick={() => handleConnectionDropCreate(choice.kind)}>{choice.label}</button>
          ))}
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
              <button type="button" role="menuitem" onClick={() => handleContextCreate('script')}>新建剧本</button>
              <button type="button" role="menuitem" onClick={() => handleContextCreate('text')}>新建文本</button>
              <button type="button" role="menuitem" onClick={() => handleContextCreate('image')}>新建图片</button>
              <button type="button" role="menuitem" onClick={() => handleContextCreate('video')}>新建视频</button>
              <button type="button" role="menuitem" onClick={() => handleContextCreate('audio')}>新建音频</button>
              <button type="button" role="menuitem" onClick={() => handleContextCreate('director')}>新建导演台</button>
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
        </> : contextMenu.target === 'edges' ? <>
          <p className="agent-canvas-context-menu-title">连线操作</p>
          <button type="button" role="menuitem" className="is-danger" onClick={handleContextDelete}>删除连线</button>
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
          <button type="button" role="menuitem" className="is-danger" onClick={handleContextDelete}>{contextMenu.nodeIds.length === 1 && isBatchFactoryNode(nodes.find(node => node.id === contextMenu.nodeIds[0])?.data) ? '从画布移除' : '删除'}</button>
        </>}
      </section> : null}

      <input ref={uploadInputRef} className="agent-canvas-upload-input" type="file" accept="image/*,video/*,audio/*" onChange={handleAssetUpload} />
      {assetPanel ? <aside className={`agent-canvas-asset-panel${assetPanel.anchor ? ' is-node-picker' : ''}`} aria-label="项目素材库" style={assetPanel.anchor || undefined}>
        <header><strong>{assetPanel.mode === 'image-source' || assetPanel.mode === 'node-reference' ? '我的收藏' : assetPanel.replaceNodeId ? '替换素材' : '添加素材'}</strong><button type="button" aria-label="关闭素材库" onClick={() => setAssetPanel(null)}>×</button></header>
        {assetPanel.mode === 'image-source' || assetPanel.mode === 'node-reference' ? <p className="agent-canvas-asset-section-title">画布所有素材</p> : null}
        <Button size="small" onClick={startAssetUpload} disabled={assetLoading}>上传{assetPanel.assetKind === 'image' ? '图片' : assetPanel.assetKind === 'audio' ? '音频' : '视频'}</Button>
        {assetError ? <p className="agent-canvas-asset-error">{assetError}</p> : null}
        {assetLoading ? <p className="agent-canvas-asset-empty">正在读取素材...</p> : null}
        {!assetLoading && !assetError && !assets.length ? <p className="agent-canvas-asset-empty">还没有可用素材</p> : null}
        <div className="agent-canvas-asset-list">{assets.map(asset => <button type="button" key={asset.id} onClick={() => commitAsset(asset)}>{asset.url ? asset.kind === 'video'
          ? <AuthenticatedAgentVideo source={asset.url} aria-label="视频素材缩略图" preload="metadata" muted />
          : asset.kind === 'audio'
            ? <Music2 size={18} />
            : <AuthenticatedAgentImage source={asset.url} alt="" />
          : <ImagePlus size={16} />}<strong>{asset.name}</strong><span>{asset.kind === 'image' ? '图片' : asset.kind === 'audio' ? '音频' : '视频'}</span></button>)}</div>
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
        <Dropdown trigger={['click']} menu={{ items: [
          { key: 'script', label: '新建剧本' }, { key: 'text', label: '新建文本' }, { key: 'image', label: '新建图片' }, { key: 'video', label: '新建视频' }, { key: 'audio', label: '新建音频' }, { key: 'director', label: '新建导演台' }
        ], onClick: ({ key }) => addNode(key) }}>
          <span><IconTool label="添加节点"><Plus size={17} /></IconTool></span>
        </Dropdown>
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

      {selectedIds.length > 0 ? <nav className="agent-canvas-editbar" aria-label="节点编辑工具">
        <IconTool label={availability.canGroup ? '分组' : availability.groupReason} disabled={!availability.canGroup} onClick={() => runNodeCommand(current => groupCanvasNodes(current, selectedIds, `group-${Date.now()}`))}><Group size={17} /></IconTool>
        <IconTool label={availability.canUngroup ? '解组' : availability.ungroupReason} disabled={!availability.canUngroup} onClick={() => runNodeCommand(current => ungroupCanvasNodes(current, selectedIds))}><Ungroup size={17} /></IconTool>
        <Dropdown disabled={!availability.canAlign} menu={{ items: alignItems, onClick: ({ key }) => runNodeCommand(current => alignCanvasNodes(current, selectedIds, key)) }} trigger={['click']}>
          <span><IconTool label={availability.canAlign ? '对齐' : availability.alignReason} disabled={!availability.canAlign}><AlignCenterHorizontal size={17} /></IconTool></span>
        </Dropdown>
        <IconTool label={availability.canDistribute ? '水平等距' : availability.distributeReason} disabled={!availability.canDistribute} onClick={() => runNodeCommand(current => distributeCanvasNodes(current, selectedIds, 'horizontal'))}><AlignHorizontalDistributeCenter size={17} /></IconTool>
        <IconTool label={availability.canDistribute ? '垂直等距' : availability.distributeReason} disabled={!availability.canDistribute} onClick={() => runNodeCommand(current => distributeCanvasNodes(current, selectedIds, 'vertical'))}><AlignVerticalDistributeCenter size={17} /></IconTool>
      </nav> : null}

      <AgentCanvasControls
        workspaceView={workspaceView}
        zoom={Math.round((viewport.zoom || 1) * 100)}
        hasSelection={selectedIds.length > 0}
        conversationOpen={conversationOpen}
        onToggleConversation={onToggleConversation}
        onWorkspaceViewChange={onWorkspaceViewChange}
        onAutoLayout={direction => commit(layoutCanvas(nodesRef.current, edgesRef.current, direction), edgesRef.current)}
        onZoomOut={() => flowInstance?.zoomOut()}
        onZoomIn={() => flowInstance?.zoomIn()}
        onFitView={() => flowInstance?.fitView({ padding: 0.18 })}
        onFocusSelection={() => flowInstance?.fitView({ nodes: nodesRef.current.filter(node => selectedIds.includes(node.id)), padding: 0.3 })}
      />
    </div>
  );
});
