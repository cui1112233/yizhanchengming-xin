import { Input } from 'antd';
import { ArrowUp, ImagePlus, Plus } from 'lucide-react';
import { AuthenticatedAgentImage } from './AuthenticatedAgentImage';
import { AuthenticatedAgentVideo } from './AuthenticatedAgentVideo';
import { videoResultPreview } from '../pages/agent-canvas-video-jobs';
import { submitNodeGeneration } from '../pages/agent-canvas-node-generation';
import { reconcileNodeModel } from '../pages/agent-model-capabilities';
import { AgentModelPicker } from './AgentModelPicker';
import { AgentNodeParameterPopover } from './AgentNodeParameterPopover';
import { AgentNodeSettingsPanel } from './AgentNodeSettingsPanel';

const PROMPT_LABELS = { text: '文字生成要求', image: '图片提示词', video: '视频提示词', audio: '音频生成要求' };
const PROMPT_PLACEHOLDERS = {
  text: '描述要生成或改写的正文内容。',
  image: '输入提示词或添加参考图，描述你想要生成的图片。',
  video: '输入提示词或添加参考图，描述你想要生成的视频。',
  audio: '输入要朗读的文字，或描述音乐、音效。'
};

export default function AgentNodeComposer({
  node, models = [], modelLoadError = '', generationState = {},
  onConfigChange, onPickImageAsset, onPickReferenceAsset, onRemoveReference, onSubmit,
  videoJob, videoState = {}, videoBusy = false, videoActionError = '',
  onCancelVideo, onRetryVideo, onConfirmVideo
}) {
  const kind = node?.data?.type;
  if (!['text', 'image', 'video', 'audio'].includes(kind)) return null;
  const config = node.data?.config || {};
  const references = Array.isArray(config.referenceAssets) ? config.referenceAssets : [];
  const modelState = reconcileNodeModel(config, models, kind);
  const selectedModel = modelState.modelId;
  const prompt = String(config.prompt || '');
  const busy = kind === 'video' ? videoBusy : Boolean(generationState.busy);
  const error = kind === 'video'
    ? (videoJob?.errorMessage || videoActionError)
    : String(generationState.error?.message || generationState.error || '');
  const videoResult = kind === 'video' ? videoResultPreview(videoJob) : null;
  const submitDisabled = typeof onSubmit !== 'function' || busy || !prompt.trim() || !selectedModel || !modelState.available || (kind === 'video' && !videoState.canGenerate);
  const update = patch => onConfigChange?.(node.id, patch);
  const stop = event => event.stopPropagation();

  return <section className="agent-node-composer nodrag nopan" aria-label={`${PROMPT_LABELS[kind]}输入框`} onMouseDown={stop} onDoubleClick={stop}>
    {kind === 'image' ? <button type="button" className="agent-node-composer__source" onClick={event => { stop(event); onPickImageAsset?.(node.id, event); }}>
      <ImagePlus size={15} />{node.data?.assetRef?.url ? '更换当前图片' : '从收藏添加图片'}
    </button> : null}

    {kind !== 'text' ? <div className="agent-node-composer__references" aria-label="参考素材">
      {references.map((asset, index) => <span className="agent-canvas-node-reference" key={asset.id} title={asset.name || asset.id}>
        {asset.url ? asset.kind === 'video'
          ? <AuthenticatedAgentVideo source={asset.url} aria-label={`视频参考素材 ${index + 1}`} preload="metadata" muted />
          : <AuthenticatedAgentImage source={asset.url} alt={`参考素材 ${index + 1}`} />
          : <ImagePlus size={14} />}
        <i>{index + 1}</i>
        <button type="button" aria-label="移除参考素材" onClick={event => { stop(event); onRemoveReference?.(node.id, asset.id); }}>×</button>
      </span>)}
      <button type="button" className="agent-node-composer__add-reference" aria-label="添加参考素材" onClick={event => { stop(event); onPickReferenceAsset?.(node.id, event); }}><Plus size={18} /></button>
      <span className="agent-node-composer__reference-count">{references.length}/10</span>
    </div> : null}

    {videoResult && !node.data?.assetRef?.url ? <AuthenticatedAgentVideo source={videoResult.url} className="agent-node-composer__result" aria-label="待确认视频结果" preload="metadata" muted controls /> : null}

    <Input.TextArea
      aria-label={PROMPT_LABELS[kind]}
      className="agent-node-composer__prompt"
      value={prompt}
      maxLength={kind === 'text' ? 8000 : 4000}
      autoSize={{ minRows: 2, maxRows: 6 }}
      placeholder={PROMPT_PLACEHOLDERS[kind]}
      onChange={event => update({ prompt: event.target.value })}
    />

    <div className="agent-node-composer__toolbar">
      <AgentModelPicker kind={kind} config={config} models={models} disabled={busy} onChange={update} />
      <AgentNodeParameterPopover kind={kind} config={config} capabilities={modelState.model?.capabilities} disabled={busy} onChange={update} />
      <AgentNodeSettingsPanel kind={kind} config={config} models={models} disabled={busy} onChange={update} />

      <span className={`agent-node-composer__state${error ? ' is-error' : ''}`} aria-live="polite">
        {error || (busy ? `正在生成${kind === 'text' ? '文字' : kind === 'image' ? '图片' : '视频'}…` : kind === 'video' && videoJob ? videoState.state : '')}
      </span>
      {kind === 'video' && videoState.canCancel ? <button type="button" onClick={() => onCancelVideo?.(node.id)} disabled={busy}>取消</button> : null}
      {kind === 'video' && videoState.canRetry ? <button type="button" onClick={() => onRetryVideo?.(node.id)} disabled={busy}>重试</button> : null}
      {kind === 'video' && videoState.canConfirm ? <button type="button" onClick={() => onConfirmVideo?.(node.id)} disabled={busy}>确认</button> : null}
      <button type="button" className="agent-node-composer__submit" aria-label={`生成${kind === 'text' ? '文字' : kind === 'image' ? '图片' : kind === 'video' ? '视频' : '音频'}`} title={modelLoadError || modelState.reason || (selectedModel ? '开始生成' : '请先在 API 配置中添加可用模型')} disabled={submitDisabled} onPointerDown={stop} onClick={event => submitNodeGeneration({ event, onSubmit, nodeID: node.id, modelID: selectedModel })}>
        <ArrowUp size={17} /><span>生成</span>
      </button>
    </div>
  </section>;
}
