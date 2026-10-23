import { Input } from 'antd';
import { ArrowUp, ImagePlus, Plus } from 'lucide-react';
import { useEffect, useState } from 'react';
import { AuthenticatedAgentImage } from './AuthenticatedAgentImage';
import { AuthenticatedAgentVideo } from './AuthenticatedAgentVideo';
import { AuthenticatedAgentAudio } from './AuthenticatedAgentAudio';
import { videoResultPreview } from '../pages/agent-canvas-video-jobs';
import { submitNodeGeneration } from '../pages/agent-canvas-node-generation';
import { audioOperationNeedsModel, canSubmitAudioNode } from '../pages/agent-canvas-audio-jobs';
import { reconcileNodeModel } from '../pages/agent-model-capabilities';
import { AgentModelPicker } from './AgentModelPicker';
import { AgentNodeParameterPopover } from './AgentNodeParameterPopover';
import { AgentNodeSettingsPanel } from './AgentNodeSettingsPanel';
import AgentScriptNodeComposer from './AgentScriptNodeComposer';

const PROMPT_LABELS = { text: '文字生成要求', image: '图片提示词', video: '视频提示词', audio: '音频生成要求' };
const PROMPT_PLACEHOLDERS = {
  text: '描述要生成或改写的正文内容。',
  image: '输入提示词或添加参考图，描述你想要生成的图片。',
  video: '输入提示词或添加参考图，描述你想要生成的视频。',
  audio: '输入要朗读的文字，或描述音乐、音效。'
};

export default function AgentNodeComposer({
  node, models = [], modelLoadError = '', generationState = {},
  modelsByKind = {}, scriptState = {}, onSourceTextChange, onCommitScriptSource, onRunBatchFactoryStage, onSplitScript, onBatchShots,
  onConfigChange, onPickImageAsset, onPickReferenceAsset, onRemoveReference, onSubmit,
  timingWarning = '',
  videoJob, videoState = {}, videoBusy = false, videoActionError = '',
  onSelectImageCandidate, onConfirmImage,
  onSelectAudioCandidate, onConfirmAudio,
  onCancelVideo, onRetryVideo, onConfirmVideo
}) {
  const kind = node?.data?.type;
  const prompt = String(node?.data?.config?.prompt || '');
  const [draftPrompt, setDraftPrompt] = useState(prompt);
  useEffect(() => {
    setDraftPrompt(prompt);
  }, [node?.id, prompt]);
  if (kind === 'script') return <AgentScriptNodeComposer
    node={node}
    modelsByKind={modelsByKind}
    modelLoadError={modelLoadError}
    scriptState={scriptState}
    onConfigChange={onConfigChange}
    onSourceTextChange={onSourceTextChange}
    onCommitSource={onCommitScriptSource}
    onRunBatchFactoryStage={onRunBatchFactoryStage}
    onSplitScript={onSplitScript}
    onBatchShots={onBatchShots}
  />;
  if (!['text', 'image', 'video', 'audio'].includes(kind)) return null;
  const config = node.data?.config || {};
  const references = Array.isArray(config.referenceAssets) ? config.referenceAssets : [];
  const modelState = reconcileNodeModel(config, models, kind);
  const selectedModel = modelState.modelId;
  const busy = kind === 'video' ? videoBusy : Boolean(generationState.busy);
  const error = kind === 'video'
    ? (videoJob?.errorMessage || videoActionError)
    : String(generationState.error?.message || generationState.error || '');
  const videoResult = kind === 'video' ? videoResultPreview(videoJob) : null;
  const imageCandidates = kind === 'image' && Array.isArray(generationState.candidates) ? generationState.candidates : [];
  const selectedImageID = String(generationState.selectedAssetId || imageCandidates[0]?.id || '');
  const audioCandidates = kind === 'audio' && Array.isArray(generationState.candidates) ? generationState.candidates : [];
  const selectedAudioID = String(generationState.selectedAssetId || audioCandidates[0]?.id || '');
  const audioNeedsModel = kind === 'audio' && audioOperationNeedsModel(config.parameters?.operation || config.operation);
  const audioReady = kind === 'audio' && canSubmitAudioNode(node, selectedModel);
  const isBatchFactoryNode = config.source === 'batch_factory';
  const imageGenerationStatus = kind !== 'image' ? null : (() => {
    if (error) return { tone: 'error', label: '图片生成失败', detail: error };
    if (busy) {
      const state = generationState.presentation?.state;
      if (state === 'queued') return { tone: 'running', label: '正在排队，等待豆包执行器接收' };
      if (state === 'running') return { tone: 'running', label: '正在生成图片' };
      return { tone: 'running', label: generationState.phase || '正在提交生成请求' };
    }
    if (imageCandidates.length) return { tone: 'review', label: '图片已生成，等待选择图片' };
    return null;
  })();
  const submitDisabled = typeof onSubmit !== 'function' || busy || (kind === 'audio'
    ? (!isBatchFactoryNode && !audioReady)
    : (!draftPrompt.trim() || (!isBatchFactoryNode && (!selectedModel || !modelState.available || (kind === 'video' && !videoState.canGenerate)))));
  const update = patch => onConfigChange?.(node.id, patch);
  const stop = event => event.stopPropagation();
  const persistPrompt = () => {
    if (draftPrompt !== prompt) update({ prompt: draftPrompt });
  };

  return <section className="agent-node-composer nodrag nopan" aria-label={`${PROMPT_LABELS[kind]}输入框`} onMouseDown={stop} onDoubleClick={stop}>
    {kind === 'image' ? <button type="button" className="agent-node-composer__source" onPointerDown={stop} onClick={event => { event.preventDefault(); stop(event); onPickImageAsset?.(node.id, event); }}>
      <ImagePlus size={15} />{node.data?.assetRef?.url ? '更换当前图片' : '从收藏添加图片'}
    </button> : null}

    {kind !== 'text' ? <div className="agent-node-composer__references" aria-label="参考素材">
      {references.map((asset, index) => <span className="agent-canvas-node-reference" key={asset.id} title={asset.name || asset.id}>
        {asset.url ? asset.kind === 'video'
          ? <AuthenticatedAgentVideo source={asset.url} aria-label={`视频参考素材 ${index + 1}`} preload="metadata" muted />
          : asset.kind === 'audio'
            ? <span className="agent-canvas-node-reference__audio">音频</span>
            : <AuthenticatedAgentImage source={asset.url} alt={`参考素材 ${index + 1}`} />
          : <ImagePlus size={14} />}
        <i>{index + 1}</i>
        <button type="button" aria-label="移除参考素材" onClick={event => { stop(event); onRemoveReference?.(node.id, asset.id); }}>×</button>
      </span>)}
      <button type="button" className="agent-node-composer__add-reference" aria-label="添加参考素材" onClick={event => { stop(event); onPickReferenceAsset?.(node.id, event); }}><Plus size={18} /></button>
      <span className="agent-node-composer__reference-count">{references.length}/10</span>
    </div> : null}

    {imageGenerationStatus ? <div className={`agent-node-composer__generation-status is-${imageGenerationStatus.tone}`} role="status" aria-live="polite">
      {imageGenerationStatus.tone === 'running' ? <span className="agent-node-composer__generation-spinner" aria-hidden="true" /> : null}
      <div><strong>{imageGenerationStatus.label}</strong>{imageGenerationStatus.detail ? <small>{imageGenerationStatus.detail}</small> : null}</div>
    </div> : null}

    {videoResult && !node.data?.assetRef?.url ? <AuthenticatedAgentVideo source={videoResult.url} className="agent-node-composer__result" aria-label="待确认视频结果" preload="metadata" muted controls /> : null}

    {imageCandidates.length ? <div className="agent-node-composer__candidates" aria-label="生成图片候选">
      <div className="agent-node-composer__candidate-head"><strong>选择生成结果</strong><span>{imageCandidates.length} 张</span></div>
      <div className="agent-node-composer__candidate-list">
        {imageCandidates.map((asset, index) => <button
          type="button"
          key={asset.id}
          className={asset.id === selectedImageID ? 'is-selected' : ''}
          aria-label={`选择候选图片 ${index + 1}`}
          aria-pressed={asset.id === selectedImageID}
          onClick={() => onSelectImageCandidate?.(node.id, asset.id)}
        ><AuthenticatedAgentImage source={asset.url} alt={`候选图片 ${index + 1}`} /><span>{index + 1}</span></button>)}
      </div>
      <div className="agent-node-composer__candidate-actions">
        <button type="button" disabled={busy} onClick={event => submitNodeGeneration({ event, onSubmit, nodeID: node.id, modelID: selectedModel })}>重新生成</button>
        <button type="button" className="is-primary" disabled={busy || !selectedImageID} onClick={() => onConfirmImage?.(node.id, selectedImageID)}>使用这张</button>
      </div>
    </div> : null}

    {audioCandidates.length ? <div className="agent-node-composer__candidates agent-node-composer__audio-candidates" aria-label="生成音频候选">
      <div className="agent-node-composer__candidate-head"><strong>选择音频结果</strong><span>{audioCandidates.length} 条</span></div>
      <div className="agent-node-composer__audio-list">
        {audioCandidates.map((asset, index) => <article key={asset.id} className={asset.id === selectedAudioID ? 'is-selected' : ''}>
          <AuthenticatedAgentAudio source={asset.url} aria-label={`候选音频 ${index + 1}`} preload="metadata" controls />
          <button type="button" aria-pressed={asset.id === selectedAudioID} onClick={() => onSelectAudioCandidate?.(node.id, asset.id)}>选择第 {index + 1} 条</button>
        </article>)}
      </div>
      <div className="agent-node-composer__candidate-actions">
        <button type="button" disabled={busy} onClick={event => submitNodeGeneration({ event, onSubmit, nodeID: node.id, modelID: selectedModel })}>重新执行</button>
        <button type="button" className="is-primary" disabled={busy || !selectedAudioID} onClick={() => onConfirmAudio?.(node.id, selectedAudioID)}>使用这条</button>
      </div>
    </div> : null}

    <Input.TextArea
      aria-label={PROMPT_LABELS[kind]}
      className="agent-node-composer__prompt"
      value={draftPrompt}
      maxLength={kind === 'text' ? 8000 : 4000}
      autoSize={{ minRows: 2, maxRows: 6 }}
      placeholder={PROMPT_PLACEHOLDERS[kind]}
      onChange={event => setDraftPrompt(event.target.value)}
      onBlur={persistPrompt}
    />

    <div className="agent-node-composer__toolbar">
      {kind !== 'audio' || audioNeedsModel ? <AgentModelPicker kind={kind} config={config} models={models} disabled={busy} onChange={update} /> : <span className="agent-node-composer__local-processor">本地音频处理</span>}
      <AgentNodeParameterPopover kind={kind} config={config} capabilities={modelState.model?.capabilities} disabled={busy} onChange={update} />
      <AgentNodeSettingsPanel kind={kind} config={config} models={models} disabled={busy} onChange={update} />

      <span className={`agent-node-composer__state${error ? ' is-error' : ''}`} aria-live="polite">
        {kind === 'image' ? '' : error || (busy ? `正在生成${kind === 'text' ? '文字' : kind === 'video' ? '视频' : '音频'}…` : kind === 'audio' && audioCandidates.length ? '等待选择' : kind === 'video' && videoJob ? videoState.label : generationState.presentation?.message || '')}
      </span>
      {kind === 'video' && videoState.canCancel ? <button type="button" onClick={() => onCancelVideo?.(node.id)} disabled={busy}>取消</button> : null}
      {kind === 'video' && videoState.canRetry ? <button type="button" onClick={() => onRetryVideo?.(node.id)} disabled={busy}>重试</button> : null}
      {kind === 'video' && videoState.canConfirm ? <button type="button" onClick={() => onConfirmVideo?.(node.id)} disabled={busy}>确认</button> : null}
      <button type="button" className="agent-node-composer__submit" aria-label={`生成${kind === 'text' ? '文字' : kind === 'image' ? '图片' : kind === 'video' ? '视频' : '音频'}`} title={modelLoadError || modelState.reason || (selectedModel ? '开始生成' : '请先在 API 配置中添加可用模型')} disabled={submitDisabled} onPointerDown={stop} onClick={event => submitNodeGeneration({ event, onSubmit, nodeID: node.id, modelID: selectedModel, prompt: draftPrompt })}>
        <ArrowUp size={17} /><span>生成</span>
      </button>
    </div>
  </section>;
}
