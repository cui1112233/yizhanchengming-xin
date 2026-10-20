import { Input, Select } from 'antd';
import { ArrowUp, ImagePlus, Plus } from 'lucide-react';
import { AuthenticatedAgentImage } from './AuthenticatedAgentImage';
import { AuthenticatedAgentVideo } from './AuthenticatedAgentVideo';
import { videoResultPreview } from '../pages/agent-canvas-video-jobs';
import { submitNodeGeneration } from '../pages/agent-canvas-node-generation';

const PROMPT_LABELS = { text: '文字生成要求', image: '图片提示词', video: '视频提示词' };
const PROMPT_PLACEHOLDERS = {
  text: '描述要生成或改写的正文内容。',
  image: '输入提示词或添加参考图，描述你想要生成的图片。',
  video: '输入提示词或添加参考图，描述你想要生成的视频。'
};

export default function AgentNodeComposer({
  node, modelOptions = [], modelLoadError = '', generationState = {},
  onConfigChange, onPickImageAsset, onPickReferenceAsset, onRemoveReference, onSubmit,
  videoJob, videoState = {}, videoBusy = false, videoActionError = '',
  onCancelVideo, onRetryVideo, onConfirmVideo
}) {
  const kind = node?.data?.type;
  if (!['text', 'image', 'video'].includes(kind)) return null;
  const config = node.data?.config || {};
  const references = Array.isArray(config.referenceAssets) ? config.referenceAssets : [];
  const selectedModel = modelOptions.some(option => option.value === config.model)
    ? config.model
    : (modelOptions[0]?.value || '');
  const prompt = String(config.prompt || '');
  const busy = kind === 'video' ? videoBusy : Boolean(generationState.busy);
  const error = kind === 'video'
    ? (videoJob?.errorMessage || videoActionError)
    : String(generationState.error?.message || generationState.error || '');
  const videoResult = kind === 'video' ? videoResultPreview(videoJob) : null;
  const submitDisabled = busy || !prompt.trim() || !selectedModel || (kind === 'video' && !videoState.canGenerate);
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
      {kind === 'text' ? <Select aria-label="文字模型" className="agent-node-composer__model" variant="borderless" size="small" value={selectedModel || undefined} placeholder={modelLoadError || '选择文字模型'} options={modelOptions} onChange={model => update({ model })} /> : null}
      {kind === 'image' ? <Select aria-label="图片模型" className="agent-node-composer__model" variant="borderless" size="small" value={selectedModel || undefined} placeholder={modelLoadError || '选择图片模型'} options={modelOptions} onChange={model => update({ model })} /> : null}
      {kind === 'video' ? <Select aria-label="视频模型" className="agent-node-composer__model" variant="borderless" size="small" value={selectedModel || undefined} placeholder={modelLoadError || '选择视频模型'} options={modelOptions} onChange={model => update({ model })} /> : null}

      {kind === 'image' ? <Select aria-label="图片比例" variant="borderless" size="small" value={config.ratio || '16:9'} options={['16:9', '9:16', '1:1'].map(value => ({ value, label: value }))} onChange={ratio => update({ ratio })} /> : null}
      {kind === 'video' ? <>
        <Select aria-label="视频比例" variant="borderless" size="small" value={config.ratio || '16:9'} options={[{ value: '16:9', label: '16:9' }, { value: '9:16', label: '9:16' }]} onChange={ratio => update({ ratio })} />
        <Select aria-label="视频时长（秒）" variant="borderless" size="small" value={config.durationSeconds || 10} options={[5, 10, 15].map(value => ({ value, label: `${value}s` }))} onChange={durationSeconds => update({ durationSeconds })} />
        <Select aria-label="视频清晰度" variant="borderless" size="small" value={config.resolution || 'auto'} options={[{ value: 'auto', label: '自动' }, { value: '720p', label: '720p' }, { value: '1080p', label: '1080p' }]} onChange={resolution => update({ resolution })} />
      </> : null}

      <span className={`agent-node-composer__state${error ? ' is-error' : ''}`} aria-live="polite">
        {error || (busy ? `正在生成${kind === 'text' ? '文字' : kind === 'image' ? '图片' : '视频'}…` : kind === 'video' && videoJob ? videoState.state : '')}
      </span>
      {kind === 'video' && videoState.canCancel ? <button type="button" onClick={() => onCancelVideo?.(node.id)} disabled={busy}>取消</button> : null}
      {kind === 'video' && videoState.canRetry ? <button type="button" onClick={() => onRetryVideo?.(node.id)} disabled={busy}>重试</button> : null}
      {kind === 'video' && videoState.canConfirm ? <button type="button" onClick={() => onConfirmVideo?.(node.id)} disabled={busy}>确认</button> : null}
      <button type="button" className="agent-node-composer__submit" aria-label={`生成${kind === 'text' ? '文字' : kind === 'image' ? '图片' : '视频'}`} title={selectedModel ? '开始生成' : '请先在 API 配置中添加可用模型'} disabled={submitDisabled} onPointerDown={stop} onClick={event => submitNodeGeneration({ event, onSubmit, nodeID: node.id, modelID: selectedModel })}>
        <ArrowUp size={17} /><span>生成</span>
      </button>
    </div>
  </section>;
}
