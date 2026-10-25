import { Input, Select } from 'antd';
import { ArrowUp, Check, Eye, FileText, Film, Image as ImageIcon, ImagePlus, Music2, RefreshCw, SlidersHorizontal, Upload } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { AuthenticatedAgentImage } from './AuthenticatedAgentImage';
import { AuthenticatedAgentVideo } from './AuthenticatedAgentVideo';
import { AuthenticatedAgentAudio } from './AuthenticatedAgentAudio';
import { videoResultPreview } from '../pages/agent-canvas-video-jobs';
import { submitNodeGeneration } from '../pages/agent-canvas-node-generation';
import { audioOperationNeedsModel, canSubmitAudioNode } from '../pages/agent-canvas-audio-jobs';
import { reconcileNodeModel } from '../pages/agent-model-capabilities';
import { compactComposerActionPlan } from '../pages/agent-node-composer-actions';
import { AgentModelPicker } from './AgentModelPicker';
import { AgentNodeParameterPopover } from './AgentNodeParameterPopover';
import { AgentNodeSettingsPanel } from './AgentNodeSettingsPanel';
import { AgentImageTools } from './AgentImageTools';
import { AgentVideoTools } from './AgentVideoTools';
import { AgentBatchFactoryPromptSelector } from './AgentBatchFactoryPromptSelector';
import { AgentNodeComposerReferenceAssets } from './AgentNodeComposerReferenceAssets';
import { AgentNodeComposerTopActions } from './AgentNodeComposerTopActions';
import AgentScriptNodeComposer from './AgentScriptNodeComposer';
import { assetMentionTrigger, filterMentionAssets, insertAssetMention, referenceKindsForPromptNode } from '../pages/agent-canvas-asset-input';

const PROMPT_LABELS = { text: '文字生成要求', image: '图片提示词', video: '视频提示词', audio: '音频生成要求' };
const PROMPT_PLACEHOLDERS = {
  text: '描述要生成或改写的正文内容。',
  image: '输入提示词或添加参考图，描述你想要生成的图片。',
  video: '输入提示词或添加参考图，描述你想要生成的视频。',
  audio: '输入要朗读的文字，或描述音乐、音效。'
};

export default function AgentNodeComposer({
  compact = false,
  canvasZoom = 1,
  node, models = [], modelLoadError = '', generationState = {},
  modelsByKind = {}, scriptState = {}, onSourceTextChange, onCommitScriptSource, onRunBatchFactoryStage, onSplitScript, onBatchShots,
  onConfigChange, onPickImageAsset, onPickReferenceAsset, onRemoveReference, onListMentionAssets, onMentionAsset, onSubmit,
  inputPreview, onPreviewInputs,
  timingWarning = '',
  videoJob, videoState = {}, videoBusy = false, videoActionError = '',
  videoToolsRequest = 0,
  settingsRequest = 0,
  onSelectImageCandidate, onConfirmImage,
  onSelectAudioCandidate, onConfirmAudio,
  onCancelGeneration, onRetryGeneration, onCancelVideo, onRetryVideo, onConfirmVideo, onPublishBatchFactory, onPublishOrganizationChange, publishOrganizations = [], publishOrganizationsLoading = false, publishOrganizationsError = '', onImageTool, onVideoTool, onBatchFactoryPromptSaved
}) {
  const kind = node?.data?.type;
  const prompt = String(node?.data?.config?.prompt || '');
  const [draftPrompt, setDraftPrompt] = useState(prompt);
  const [toolsOpen, setToolsOpen] = useState(false);
  const [mention, setMention] = useState(null);
  const [mentionAssets, setMentionAssets] = useState([]);
  const [mentionState, setMentionState] = useState({ loading: false, error: '', active: 0 });
  const promptRef = useRef(null);
  const mentionRequestRef = useRef(0);
  useEffect(() => setDraftPrompt(prompt), [node?.id, prompt]);
  useEffect(() => {
    setToolsOpen(false);
    setMention(null);
    setMentionAssets([]);
    setMentionState({ loading: false, error: '', active: 0 });
  }, [node?.id]);
  useEffect(() => {
    if ((kind === 'image' || kind === 'video') && videoToolsRequest > 0) setToolsOpen(true);
  }, [kind, node?.id, videoToolsRequest]);

  if (kind === 'script') return <AgentScriptNodeComposer
    compact={compact}
    node={node} modelsByKind={modelsByKind} modelLoadError={modelLoadError} scriptState={scriptState}
    onConfigChange={onConfigChange} onSourceTextChange={onSourceTextChange} onCommitSource={onCommitScriptSource}
    onRunBatchFactoryStage={onRunBatchFactoryStage} onSplitScript={onSplitScript} onBatchShots={onBatchShots}
    onBatchFactoryPromptSaved={onBatchFactoryPromptSaved}
  />;
  if (!['text', 'image', 'video', 'audio'].includes(kind)) return null;

  const config = node.data?.config || {};
  const references = Array.isArray(config.referenceAssets) ? config.referenceAssets : [];
  const modelState = reconcileNodeModel(config, models, kind);
  const selectedModel = modelState.modelId;
  const isVideoToolTask = kind === 'video' && String(generationState.mode || '').startsWith('video_');
  const busy = isVideoToolTask ? Boolean(generationState.busy) : kind === 'video' ? videoBusy : Boolean(generationState.busy);
  const error = kind === 'video' && !isVideoToolTask
    ? (videoJob?.errorMessage || videoActionError || generationState.error?.message || generationState.error || '')
    : String(generationState.error?.message || generationState.error || '');
  const videoResult = kind === 'video' ? videoResultPreview(videoJob) : null;
  const imageCandidates = kind === 'image' && Array.isArray(generationState.candidates) ? generationState.candidates : [];
  const selectedImageID = String(generationState.selectedAssetId || imageCandidates[0]?.id || '');
  const audioCandidates = kind === 'audio' && Array.isArray(generationState.candidates) ? generationState.candidates : [];
  const selectedAudioID = String(generationState.selectedAssetId || audioCandidates[0]?.id || '');
  const videoToolCandidates = isVideoToolTask && Array.isArray(generationState.candidates) ? generationState.candidates : [];
  const selectedVideoToolID = String(generationState.selectedAssetId || videoToolCandidates[0]?.id || '');
  const audioNeedsModel = kind === 'audio' && audioOperationNeedsModel(config.parameters?.operation || config.operation);
  const audioReady = kind === 'audio' && canSubmitAudioNode(node, selectedModel);
  const isBatchFactoryNode = config.source === 'batch_factory';
  const isBatchFactoryMerge = isBatchFactoryNode && config.resourceType === 'merge';
  const mergeReady = isBatchFactoryMerge && config.canMerge === true;
  const mergeReason = String(config.mergeReason || '等待所有分镜视频生成完成');
  const publishStatus = String(config.publishStatus || '').trim();
  const publishProgress = config.publishProgress && typeof config.publishProgress === 'object' ? config.publishProgress : null;
  const publishOrganization = String(config.publishOrganization || '').trim();
  const taskActionBusy = Boolean(generationState.actionBusy || videoBusy);
  const canCancelGeneration = (kind !== 'video' || isVideoToolTask) && generationState.presentation?.canCancel;
  const canRetryGeneration = (kind !== 'video' || isVideoToolTask) && generationState.presentation?.canRetry;
  const canRetry = Boolean(canRetryGeneration || (kind === 'video' && videoState.canRetry));
  const actionPlan = compactComposerActionPlan({
    kind, source: config.source, resourceType: config.resourceType,
    canRetry, hasImageCandidates: imageCandidates.length > 0,
    hasAudioCandidates: audioCandidates.length > 0,
    canPublish: config.canPublish === true
  });
  const submitDisabled = typeof onSubmit !== 'function' || busy || (isBatchFactoryMerge
    ? !mergeReady
    : kind === 'audio'
      ? (!isBatchFactoryNode && !audioReady)
      : (!draftPrompt.trim() || (!isBatchFactoryNode && (!selectedModel || !modelState.available || (kind === 'video' && !videoState.canGenerate)))));
  const update = patch => onConfigChange?.(node.id, patch);
  const stop = event => event.stopPropagation();
  const persistPrompt = () => { if (draftPrompt !== prompt) update({ prompt: draftPrompt }); };
  const mentionKinds = referenceKindsForPromptNode(kind);
  const mentionCandidates = filterMentionAssets(mentionAssets, mentionKinds, mention?.query || '');
  const openMention = async nextMention => {
    setMention(nextMention);
    setMentionState({ loading: true, error: '', active: 0 });
    const requestID = mentionRequestRef.current + 1;
    mentionRequestRef.current = requestID;
    try {
      const loaded = await onListMentionAssets?.(node.id);
      if (mentionRequestRef.current !== requestID) return;
      setMentionAssets(Array.isArray(loaded) ? loaded : []);
      setMentionState({ loading: false, error: '', active: 0 });
    } catch (loadError) {
      if (mentionRequestRef.current !== requestID) return;
      setMentionAssets([]);
      setMentionState({ loading: false, error: loadError?.message || '资产读取失败', active: 0 });
    }
  };
  const updatePrompt = event => {
    const value = event.target.value;
    const nextMention = assetMentionTrigger(value, event.target.selectionStart);
    setDraftPrompt(value);
    if (!nextMention || !mentionKinds.length) {
      mentionRequestRef.current += 1;
      setMention(null);
      return;
    }
    if (!mention) openMention(nextMention);
    else {
      setMention(nextMention);
      setMentionState(current => ({ ...current, active: 0 }));
    }
  };
  const chooseMentionAsset = asset => {
    if (!mention) return;
    const inserted = insertAssetMention(draftPrompt, mention, asset);
    setDraftPrompt(inserted.value);
    setMention(null);
    onMentionAsset?.(node.id, asset, inserted.value);
    window.requestAnimationFrame(() => {
      const input = promptRef.current?.resizableTextArea?.textArea || promptRef.current;
      input?.focus?.();
      input?.setSelectionRange?.(inserted.caret, inserted.caret);
    });
  };
  const handlePromptKeyDown = event => {
    if (!mention) return;
    if (event.key === 'Escape') {
      event.preventDefault();
      setMention(null);
      return;
    }
    if (!mentionCandidates.length || !['ArrowDown', 'ArrowUp', 'Enter'].includes(event.key)) return;
    event.preventDefault();
    if (event.key === 'Enter') chooseMentionAsset(mentionCandidates[mentionState.active] || mentionCandidates[0]);
    else setMentionState(current => ({
      ...current,
      active: (current.active + (event.key === 'ArrowDown' ? 1 : -1) + mentionCandidates.length) % mentionCandidates.length
    }));
  };
  const previewInputs = inputPreview?.inputs || null;
  const previewCount = inputKind => Array.isArray(previewInputs?.[inputKind]) ? previewInputs[inputKind].length : 0;
  const previewPrompt = String(previewInputs?.resolvedPrompt || previewInputs?.prompt || '').trim();
  const promptLimit = kind === 'text' ? 8000 : 4000;
  const imageGenerationStatus = kind !== 'image' ? null : (() => {
    if (error) return { tone: 'error', label: '图片生成失败', detail: error };
    if (busy) {
      const state = generationState.presentation?.state;
      if (state === 'queued') return { tone: 'running', label: '正在排队，等待豆包执行器接收' };
      if (state === 'running') return { tone: 'running', label: '正在生成图片' };
      return { tone: 'running', label: generationState.phase || '正在提交生成请求' };
    }
    return imageCandidates.length ? { tone: 'review', label: '图片已生成，等待选择图片' } : null;
  })();

  const topActions = useMemo(() => {
    const direct = [];
    const more = [];
    const action = (id, label, onClick, icon) => ({ id, label, onClick, icon, disabled: busy || taskActionBusy });
    const referenceAssets = {
      id: 'reference-assets',
      content: <AgentNodeComposerReferenceAssets
        references={references}
        disabled={busy}
        onAdd={event => onPickReferenceAsset?.(node.id, event)}
        onRemove={asset => onRemoveReference?.(node.id, asset.id)}
      />
    };
    for (const actionID of actionPlan.primary) {
      if (actionID === 'reference-assets') direct.push(referenceAssets);
      if (actionID === 'prompt-templates') direct.push({
        id: actionID,
        content: <AgentBatchFactoryPromptSelector config={config} disabled={busy} onSaved={onBatchFactoryPromptSaved} display="popover-compact" />
      });
      if (actionID === 'publish-organization') direct.push({
        id: actionID,
        content: <Select
          size="small"
          aria-label="视频管理系统组织归属"
          className="agent-node-composer__organization-select"
          value={publishOrganization || undefined}
          placeholder="选择组织归属"
          loading={publishOrganizationsLoading}
          disabled={busy || publishOrganizationsLoading}
          options={publishOrganizations.map(item => ({ value: item.id, label: item.level ? `${item.name}（${item.level}）` : item.name }))}
          onChange={organization => onPublishOrganizationChange?.(node.id, organization)}
        />
      });
    }
    for (const actionID of actionPlan.more) {
      if (actionID === 'preview-inputs') more.push(action(actionID, '查看实际输入', () => onPreviewInputs?.(node.id, selectedModel), <Eye size={14} />));
      if (actionID === 'retry') more.push(action(actionID, '重试', () => {
        if (kind === 'video' && !isVideoToolTask) onRetryVideo?.(node.id);
        else onRetryGeneration?.(node.id);
      }, <RefreshCw size={14} />));
      if (actionID === 'confirm-image') more.push(action(actionID, '使用这张', () => onConfirmImage?.(node.id, selectedImageID), <Check size={14} />));
      if (actionID === 'confirm-audio') more.push(action(actionID, '使用这条', () => onConfirmAudio?.(node.id, selectedAudioID), <Check size={14} />));
      if (actionID === 'upload-video-management') more.push({
        ...action(actionID, publishStatus === 'uploaded' ? '已上传到视频管理系统' : publishStatus === 'uploading' ? '正在上传' : '上传到视频管理系统', () => onPublishBatchFactory?.(node.id), <Upload size={14} />),
        disabled: busy || taskActionBusy || !publishOrganization || publishStatus === 'uploaded'
      });
    }
    if (kind === 'image') more.push(action('source-image', node.data?.assetRef?.url ? '更换当前图片' : '从收藏添加图片', event => onPickImageAsset?.(node.id, event), <ImagePlus size={14} />));
    if (canCancelGeneration) more.push(action('cancel-generation', '取消', () => onCancelGeneration?.(node.id)));
    if (kind === 'video' && videoState.canCancel) more.push(action('cancel-video', '取消', () => onCancelVideo?.(node.id)));
    if (kind === 'video' && videoState.canConfirm) more.push(action('confirm-video', '确认', () => onConfirmVideo?.(node.id), <Check size={14} />));
    if (kind === 'image' || (kind === 'video' && !isVideoToolTask && !isBatchFactoryMerge)) direct.push({
      id: 'media-tools',
      label: kind === 'image' ? (toolsOpen ? '收起图片处理' : '图片处理') : (toolsOpen ? '收起视频处理' : '视频处理'),
      icon: <SlidersHorizontal size={14} />,
      disabled: busy || taskActionBusy,
      onClick: () => setToolsOpen(open => !open)
    });
    return { direct, more };
  }, [actionPlan.primary.join('|'), actionPlan.more.join('|'), busy, taskActionBusy, references, config, onBatchFactoryPromptSaved, onPickReferenceAsset, onRemoveReference, node.id, onPreviewInputs, selectedModel, kind, isVideoToolTask, isBatchFactoryMerge, onRetryVideo, onRetryGeneration, onConfirmImage, selectedImageID, onConfirmAudio, selectedAudioID, publishStatus, onPublishBatchFactory, onPickImageAsset, node.data?.assetRef?.url, canCancelGeneration, onCancelGeneration, videoState.canCancel, videoState.canConfirm, onCancelVideo, onConfirmVideo, publishOrganization, publishOrganizationsLoading, publishOrganizations, onPublishOrganizationChange, toolsOpen]);

  return <section className={`agent-node-composer agent-node-composer--${kind} nodrag nopan`} style={{ '--agent-canvas-zoom': canvasZoom }} aria-label={`${PROMPT_LABELS[kind]}输入框`} onMouseDown={stop} onDoubleClick={stop}>
    <div className="agent-node-composer__asset-strip">
      <AgentNodeComposerTopActions primaryActions={topActions.direct} moreActions={topActions.more} />
    </div>

    <div className="agent-node-composer__body">
      {kind === 'image' && toolsOpen ? <AgentImageTools sourceAsset={node.data?.assetRef} busy={busy} onApply={input => onImageTool?.(node.id, input)} /> : null}
      {kind === 'video' && !isVideoToolTask && !isBatchFactoryMerge && toolsOpen ? <AgentVideoTools sourceAsset={node.data?.assetRef} busy={busy} onApply={input => onVideoTool?.(node.id, input)} /> : null}
      {imageGenerationStatus ? <div className={`agent-node-composer__generation-status is-${imageGenerationStatus.tone}`} role="status" aria-live="polite">
        {imageGenerationStatus.tone === 'running' ? <span className="agent-node-composer__generation-spinner" aria-hidden="true" /> : null}
        <div><strong>{imageGenerationStatus.label}</strong>{imageGenerationStatus.detail ? <small>{imageGenerationStatus.detail}</small> : null}</div>
      </div> : null}
      {isBatchFactoryMerge ? <div className="agent-node-composer__merge-status">
        <strong>{mergeReady ? '可合成最终视频' : '暂不可合成'}</strong>
        <span>{mergeReady ? '分镜视频已就绪，可按当前主版本合成。' : mergeReason}</span>
        {publishOrganizationsError ? <small className="is-error">{publishOrganizationsError}</small> : null}
        {!publishOrganizationsLoading && !publishOrganizationsError && !publishOrganizations.length ? <small className="is-warning">暂无可用组织，请先在视频管理系统完成登录。</small> : null}
      </div> : <>
        {videoResult && !node.data?.assetRef?.url ? <AuthenticatedAgentVideo source={videoResult.url} className="agent-node-composer__result" aria-label="待确认视频结果" preload="metadata" muted controls /> : null}
        {imageCandidates.length ? <div className="agent-node-composer__candidates" aria-label="生成图片候选">
          <div className="agent-node-composer__candidate-head"><strong>选择生成结果</strong><span>{imageCandidates.length} 张</span></div>
          <div className="agent-node-composer__candidate-list">{imageCandidates.map((asset, index) => <button type="button" key={asset.id} className={asset.id === selectedImageID ? 'is-selected' : ''} aria-label={`选择候选图片 ${index + 1}`} aria-pressed={asset.id === selectedImageID} onClick={() => onSelectImageCandidate?.(node.id, asset.id)}><AuthenticatedAgentImage source={asset.url} alt={`候选图片 ${index + 1}`} /><span>{index + 1}</span></button>)}</div>
          <div className="agent-node-composer__candidate-actions"><button type="button" className="is-primary" disabled={busy || !selectedImageID} onClick={() => onConfirmImage?.(node.id, selectedImageID)}>使用这张</button></div>
        </div> : null}
        {audioCandidates.length ? <div className="agent-node-composer__candidates agent-node-composer__audio-candidates" aria-label="生成音频候选">
          <div className="agent-node-composer__candidate-head"><strong>选择音频结果</strong><span>{audioCandidates.length} 条</span></div>
          <div className="agent-node-composer__audio-list">{audioCandidates.map((asset, index) => <article key={asset.id} className={asset.id === selectedAudioID ? 'is-selected' : ''}><div className="agent-node-composer__audio-preview"><AuthenticatedAgentAudio source={asset.url} aria-label={`候选音频 ${index + 1}`} preload="metadata" controls waveform waveformSourceLabel={asset.name || `候选音频 ${index + 1}`} waveformDownstreamLabel="视频参考是否可用由目标视频模型决定" /></div><button type="button" aria-pressed={asset.id === selectedAudioID} onClick={() => onSelectAudioCandidate?.(node.id, asset.id)}>选择第 {index + 1} 条</button></article>)}</div>
        </div> : null}
        {videoToolCandidates.length ? <div className="agent-node-composer__candidates agent-node-composer__video-candidates" aria-label="视频处理候选">
          <div className="agent-node-composer__candidate-head"><strong>选择处理结果</strong><span>{videoToolCandidates.length} 个</span></div>
          <div className="agent-node-composer__video-list">{videoToolCandidates.map((asset, index) => <article key={asset.id} className={asset.id === selectedVideoToolID ? 'is-selected' : ''}><AuthenticatedAgentVideo source={asset.url} aria-label={`视频处理结果 ${index + 1}`} preload="metadata" muted controls /><button type="button" aria-pressed={asset.id === selectedVideoToolID} onClick={() => onSelectImageCandidate?.(node.id, asset.id)}>选择第 {index + 1} 个</button></article>)}</div>
          <div className="agent-node-composer__candidate-actions"><button type="button" className="is-primary" disabled={busy || !selectedVideoToolID} onClick={() => onConfirmImage?.(node.id, selectedVideoToolID)}>使用这个结果</button></div>
        </div> : null}
        <div className="agent-node-composer__prompt-wrap">
          <div className="agent-node-composer__prompt-label">
            <span>{PROMPT_LABELS[kind]}</span>
            <span className="agent-node-composer__prompt-count">{draftPrompt.length} / {promptLimit}</span>
          </div>
          {mention ? <div className="agent-node-composer__mention-menu" role="listbox" aria-label="选择项目资产" onMouseDown={event => event.preventDefault()}>
            <header><strong>@ 项目资产</strong><span>{mention.query ? `搜索“${mention.query}”` : '输入名称可筛选'}</span></header>
            {mentionState.loading ? <p>正在读取资产...</p> : null}
            {mentionState.error ? <p className="is-error">{mentionState.error}</p> : null}
            {!mentionState.loading && !mentionState.error && !mentionCandidates.length ? <p>没有匹配的资产</p> : null}
            {mentionCandidates.map((asset, index) => {
              const MentionIcon = asset.kind === 'video' ? Film : asset.kind === 'audio' ? Music2 : asset.kind === 'image' ? ImageIcon : FileText;
              return <button type="button" role="option" aria-selected={mentionState.active === index} className={mentionState.active === index ? 'is-active' : ''} key={asset.id || asset.assetId} onMouseDown={event => event.preventDefault()} onClick={() => chooseMentionAsset(asset)}>
                {asset.url && asset.kind === 'image' ? <AuthenticatedAgentImage source={asset.url} alt="" /> : <span><MentionIcon size={15} /></span>}
                <span><strong>{asset.name || '未命名资产'}</strong><small>{asset.kind === 'image' ? '图片' : asset.kind === 'video' ? '视频' : asset.kind === 'audio' ? '音频' : '文本'}</small></span>
              </button>;
            })}
          </div> : null}
          <Input.TextArea ref={promptRef} aria-label={PROMPT_LABELS[kind]} aria-expanded={Boolean(mention)} className="agent-node-composer__prompt" value={draftPrompt} maxLength={promptLimit} autoSize={false} placeholder={`${PROMPT_PLACEHOLDERS[kind]} 输入 @ 可引用项目资产。`} onChange={updatePrompt} onKeyDown={handlePromptKeyDown} onBlur={persistPrompt} />
        </div>
        {inputPreview?.error ? <small className="is-error" role="alert">{inputPreview.error}</small> : null}
        {previewInputs ? <div className="agent-node-composer__input-preview" aria-live="polite">
          <span title={previewPrompt}>最终文字：{previewPrompt || '未提供'}</span>
          <span>来源节点：{Array.isArray(previewInputs.sourceNodeIds) ? previewInputs.sourceNodeIds.length : 0} 个</span>
          <span>参考素材：图片 {previewCount('images')} / 视频 {previewCount('videos')} / 音频 {previewCount('audios')}</span>
          {Array.isArray(previewInputs.warnings) && previewInputs.warnings.length ? <small className="is-warning">{previewInputs.warnings.join('；')}</small> : null}
        </div> : null}
      </>}
    </div>

    <footer className="agent-node-composer__bottom-bar agent-node-composer__toolbar">
      {kind !== 'audio' || audioNeedsModel ? <AgentModelPicker kind={kind} config={config} models={models} disabled={busy} onChange={update} /> : <span className="agent-node-composer__local-processor">本地音频处理</span>}
      <AgentNodeParameterPopover kind={kind} config={config} capabilities={modelState.model?.capabilities} disabled={busy} onChange={update} />
      <AgentNodeSettingsPanel kind={kind} config={config} models={models} disabled={busy} openRequest={settingsRequest} onChange={update} />
      <span className={`agent-node-composer__state${error ? ' is-error' : ''}`} aria-live="polite">{kind === 'image' ? '' : error || (busy ? `正在${isVideoToolTask ? '处理视频' : `生成${kind === 'text' ? '文字' : kind === 'video' ? '视频' : '音频'}`}…` : kind === 'audio' && audioCandidates.length ? '等待选择' : kind === 'video' && isVideoToolTask ? generationState.presentation?.message || '' : kind === 'video' && videoJob ? videoState.label : generationState.presentation?.message || '')}</span>
      {timingWarning ? <small className="is-warning">{timingWarning}</small> : null}
      {isBatchFactoryMerge && publishProgress?.message ? <small className={publishStatus === 'uploaded' ? 'is-success' : 'is-warning'}>{publishProgress.message}</small> : null}
      <button type="button" className="agent-node-composer__submit" aria-label={isBatchFactoryMerge ? '合成最终视频' : `生成${kind === 'text' ? '文字' : kind === 'image' ? '图片' : kind === 'video' ? '视频' : '音频'}`} title={isBatchFactoryMerge ? (mergeReady ? '按当前分镜主版本合成最终视频' : mergeReason) : (modelLoadError || modelState.reason || (selectedModel ? '开始生成' : '请先在 API 配置中添加可用模型'))} disabled={submitDisabled} onPointerDown={stop} onClick={event => submitNodeGeneration({ event, onSubmit, nodeID: node.id, modelID: selectedModel, prompt: draftPrompt })}><ArrowUp size={17} /><span>{isBatchFactoryMerge ? (config.mergeJobId ? '重新合成' : '合成最终视频') : '生成'}</span></button>
    </footer>
  </section>;
}
