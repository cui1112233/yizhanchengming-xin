import { InputNumber, Popover, Segmented, Select, Switch } from 'antd';
import { SlidersHorizontal } from 'lucide-react';
import { formatNodeParameterSummary, normalizeNodeConfig } from '../pages/agent-node-config';
import { IMAGE_ASPECT_RATIOS } from '../../shared/imageAspectRatios';

const labels = {
  image: { title: '图片参数', ratio: '比例', quality: '画质', resolution: '清晰度', count: '数量' },
  video: { title: '视频参数', mode: '类型', ratio: '比例', resolution: '清晰度', durationSeconds: '时长', count: '数量', generateAudio: '生成声音' },
  text: { title: '文字参数', operation: '生成类型', length: '文字长度', creativity: '创作风格' },
  audio: { title: '音频参数', operation: '类型', speed: '语速 / 倍速', format: '格式' }
};

const choices = {
  quality: [{ value: 'low', label: '快速' }, { value: 'standard', label: '标准' }, { value: 'high', label: '高清' }],
  operationText: [{ value: 'generate', label: '生成' }, { value: 'rewrite', label: '改写' }, { value: 'continue', label: '续写' }],
  operationAudio: [
    { value: 'speech', label: '语音' }, { value: 'music', label: '音乐' }, { value: 'sound_effect', label: '音效' },
    { value: 'trim', label: '裁剪' }, { value: 'speed', label: '变速' }, { value: 'smart_split', label: '智能分段' }, { value: 'custom_split', label: '自定义分段' }
  ],
  length: [{ value: 'short', label: '精简' }, { value: 'standard', label: '标准' }, { value: 'long', label: '长篇' }],
  creativity: [{ value: 'precise', label: '严谨' }, { value: 'balanced', label: '均衡' }, { value: 'creative', label: '创意' }]
};

const doubaoImageModelForQuality = Object.freeze({
  low: 'Seedream 5.0 Flash',
  standard: 'Seedream 4.5',
  high: 'Seedream 5.0 Pro'
});

function valuesOr(fallback, values) {
  return Array.isArray(values) && values.length ? values : fallback;
}

function optionList(values, suffix = '') {
  return values.map(value => ({ value, label: `${value}${suffix}` }));
}

function ratioShapeStyle(value) {
  const [width, height] = String(value).split(':').map(Number);
  if (!(width > 0) || !(height > 0)) return {};
  const scale = 16 / Math.max(width, height);
  return {
    width: `${Math.max(6, Math.round(width * scale))}px`,
    height: `${Math.max(6, Math.round(height * scale))}px`
  };
}

function OptionGrid({ className = '', value, options, onChange }) {
  return <div className={`agent-node-parameter-popover__option-grid ${className}`.trim()}>
    {options.map(option => <button
      type="button"
      key={option.value}
      className={option.value === value ? 'is-selected' : ''}
      aria-pressed={option.value === value}
      onClick={() => onChange(option.value)}
    >{option.icon || null}<span>{option.label}</span></button>)}
  </div>;
}

export function AgentNodeParameterPopover({ kind, config = {}, capabilities = {}, onChange, disabled = false }) {
  const normalized = normalizeNodeConfig(kind, config);
  const parameters = normalized.parameters;
  const update = patch => onChange?.({ parameters: patch });
  const stop = event => event.stopPropagation();
  const maxCount = Math.max(1, Math.min(4, Number(capabilities.maxOutputCount) || (kind === 'image' ? 4 : 1)));
  const countOptions = optionList(Array.from({ length: maxCount }, (_, index) => index + 1), kind === 'image' ? '张' : '个');
  const imageRatios = valuesOr(IMAGE_ASPECT_RATIOS, capabilities.ratios);
  const imageResolutions = optionList(valuesOr(['1K', '2K', '4K'], capabilities.resolutions));
  const imageQualityLevels = choices.quality.filter(item => !capabilities.qualityLevels?.length || capabilities.qualityLevels.includes(item.value));
  const imageQualityPatch = quality => capabilities.resolutionManagedByModel
    ? { quality, modelId: doubaoImageModelForQuality[quality], model: doubaoImageModelForQuality[quality] }
    : { quality };
  const duration = capabilities.duration || {};
  const durationValues = [5, 10, 15].filter(value => (!duration.min || value >= duration.min) && (!duration.max || value <= duration.max));
  const generationModes = valuesOr(kind === 'video' ? ['text_to_video', 'image_to_video', 'first_frame'] : [], capabilities.generationModes);

  const content = <section className="agent-node-parameter-popover" aria-label={labels[kind]?.title} onMouseDown={stop}>
    {kind === 'image' ? <>
      <div className="agent-node-parameter-popover__field"><span>{labels.image.quality}</span><OptionGrid value={parameters.quality} options={imageQualityLevels} onChange={quality => update(imageQualityPatch(quality))} /></div>
      {capabilities.resolutionManagedByModel
        ? <div className="agent-node-parameter-popover__field"><span>{labels.image.resolution}</span><em>由豆包所选模型决定</em></div>
        : <div className="agent-node-parameter-popover__field"><span>{labels.image.resolution}</span><OptionGrid value={parameters.resolution} options={imageResolutions} onChange={resolution => update({ resolution })} /></div>}
      <div className="agent-node-parameter-popover__field"><span>{labels.image.ratio}</span><OptionGrid className="agent-node-parameter-popover__ratio-grid" value={parameters.ratio} options={imageRatios.map(ratio => ({ value: ratio, label: ratio, icon: <i className="agent-node-parameter-popover__ratio-shape" style={ratioShapeStyle(ratio)} /> }))} onChange={ratio => update({ ratio })} /></div>
      <div className="agent-node-parameter-popover__field"><span>{labels.image.count}</span><OptionGrid value={parameters.count} options={countOptions} onChange={count => update({ count })} /></div>
    </> : null}
    {kind === 'video' ? <>
      <label><span>{labels.video.mode}</span><Select value={normalized.mode || generationModes[0]} options={optionList(generationModes)} onChange={mode => onChange?.({ mode })} /></label>
      <label><span>{labels.video.ratio}</span><Segmented block value={parameters.ratio} options={optionList(valuesOr(['16:9', '9:16'], capabilities.ratios))} onChange={ratio => update({ ratio })} /></label>
      <label><span>{labels.video.resolution}</span><Select value={parameters.resolution} options={optionList(valuesOr(['720P', '1080P'], capabilities.resolutions))} onChange={resolution => update({ resolution })} /></label>
      <label><span>{labels.video.durationSeconds}</span><Select value={parameters.durationSeconds} options={optionList(durationValues.length ? durationValues : [5], 's')} onChange={durationSeconds => update({ durationSeconds })} /></label>
      <label><span>{labels.video.count}</span><Select value={parameters.count} options={countOptions} onChange={count => update({ count })} /></label>
      {capabilities.supportsAudio ? <label className="is-inline"><span>{labels.video.generateAudio}</span><Switch checked={parameters.generateAudio} onChange={generateAudio => update({ generateAudio })} /></label> : null}
    </> : null}
    {kind === 'text' ? <>
      <label><span>{labels.text.operation}</span><Segmented block value={parameters.operation} options={choices.operationText} onChange={operation => update({ operation })} /></label>
      <label><span>{labels.text.length}</span><Segmented block value={parameters.length} options={choices.length} onChange={length => update({ length })} /></label>
      <label><span>{labels.text.creativity}</span><Segmented block value={parameters.creativity} options={choices.creativity} onChange={creativity => update({ creativity })} /></label>
    </> : null}
    {kind === 'audio' ? <>
      <label className="is-wide"><span>{labels.audio.operation}</span><Select value={parameters.operation === 'effect' ? 'sound_effect' : parameters.operation} options={choices.operationAudio} onChange={operation => update({ operation })} /></label>
      {['speech', 'music', 'sound_effect', 'effect', 'speed'].includes(parameters.operation) ? <label><span>{labels.audio.speed}</span><InputNumber min={0.5} max={2} step={0.1} value={parameters.speed} onChange={speed => update({ speed })} /></label> : null}
      {parameters.operation === 'trim' ? <>
        <label><span>开始（毫秒）</span><InputNumber min={0} precision={0} value={parameters.startMs ?? 0} onChange={startMs => update({ startMs })} /></label>
        <label><span>结束（毫秒）</span><InputNumber min={1} precision={0} value={parameters.endMs ?? 5000} onChange={endMs => update({ endMs })} /></label>
      </> : null}
      {parameters.operation === 'smart_split' ? <label><span>每段时长（秒）</span><InputNumber min={2} max={600} precision={0} value={parameters.segmentSeconds ?? 15} onChange={segmentSeconds => update({ segmentSeconds })} /></label> : null}
      {parameters.operation === 'custom_split' ? <>
        <label><span>开始（毫秒）</span><InputNumber min={0} precision={0} value={parameters.ranges?.[0]?.startMs ?? 0} onChange={startMs => update({ ranges: [{ startMs, endMs: parameters.ranges?.[0]?.endMs ?? 5000 }] })} /></label>
        <label><span>结束（毫秒）</span><InputNumber min={1} precision={0} value={parameters.ranges?.[0]?.endMs ?? 5000} onChange={endMs => update({ ranges: [{ startMs: parameters.ranges?.[0]?.startMs ?? 0, endMs }] })} /></label>
      </> : null}
      {!['trim', 'smart_split', 'custom_split'].includes(parameters.operation) ? <label><span>{labels.audio.format}</span><Select value={parameters.format} options={optionList(['mp3', 'wav'])} onChange={format => update({ format })} /></label> : null}
    </> : null}
  </section>;

  return <Popover trigger="click" placement="bottom" content={content}>
    <button type="button" className="agent-node-composer__parameters" aria-label={labels[kind]?.title || '节点参数'} disabled={disabled} onClick={stop}>
      <SlidersHorizontal size={14} /><span>{formatNodeParameterSummary(kind, normalized)}</span>
    </button>
  </Popover>;
}
