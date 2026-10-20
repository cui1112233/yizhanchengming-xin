import { InputNumber, Popover, Segmented, Select, Switch } from 'antd';
import { SlidersHorizontal } from 'lucide-react';
import { formatNodeParameterSummary, normalizeNodeConfig } from '../pages/agent-node-config';

const labels = {
  image: { title: '图片参数', ratio: '比例', quality: '画质', resolution: '清晰度', count: '数量' },
  video: { title: '视频参数', mode: '类型', ratio: '比例', resolution: '清晰度', durationSeconds: '时长', count: '数量', generateAudio: '生成声音' },
  text: { title: '文字参数', operation: '类型', length: '长度', creativity: '风格' },
  audio: { title: '音频参数', operation: '类型', speed: '语速', format: '格式' }
};

const choices = {
  quality: [{ value: 'low', label: '快速' }, { value: 'standard', label: '标准' }, { value: 'high', label: '高清' }],
  operationText: [{ value: 'generate', label: '生成' }, { value: 'rewrite', label: '改写' }, { value: 'continue', label: '续写' }],
  operationAudio: [{ value: 'speech', label: '语音' }, { value: 'music', label: '音乐' }, { value: 'effect', label: '音效' }],
  length: [{ value: 'short', label: '精简' }, { value: 'standard', label: '标准' }, { value: 'long', label: '长篇' }],
  creativity: [{ value: 'precise', label: '严谨' }, { value: 'balanced', label: '均衡' }, { value: 'creative', label: '创意' }]
};

function valuesOr(fallback, values) {
  return Array.isArray(values) && values.length ? values : fallback;
}

function optionList(values, suffix = '') {
  return values.map(value => ({ value, label: `${value}${suffix}` }));
}

export function AgentNodeParameterPopover({ kind, config = {}, capabilities = {}, onChange, disabled = false }) {
  const normalized = normalizeNodeConfig(kind, config);
  const parameters = normalized.parameters;
  const update = patch => onChange?.({ parameters: patch });
  const stop = event => event.stopPropagation();
  const maxCount = Math.max(1, Math.min(4, Number(capabilities.maxOutputCount) || (kind === 'image' ? 4 : 1)));
  const countOptions = optionList(Array.from({ length: maxCount }, (_, index) => index + 1), kind === 'image' ? '张' : '个');
  const duration = capabilities.duration || {};
  const durationValues = [5, 10, 15].filter(value => (!duration.min || value >= duration.min) && (!duration.max || value <= duration.max));
  const generationModes = valuesOr(kind === 'video' ? ['text_to_video', 'image_to_video', 'first_frame'] : [], capabilities.generationModes);

  const content = <section className="agent-node-parameter-popover" aria-label={labels[kind]?.title} onMouseDown={stop}>
    {kind === 'image' ? <>
      <label><span>{labels.image.ratio}</span><Segmented block value={parameters.ratio} options={optionList(valuesOr(['16:9', '9:16', '1:1'], capabilities.ratios))} onChange={ratio => update({ ratio })} /></label>
      <label><span>{labels.image.quality}</span><Segmented block value={parameters.quality} options={choices.quality.filter(item => !capabilities.qualityLevels?.length || capabilities.qualityLevels.includes(item.value))} onChange={quality => update({ quality })} /></label>
      <label><span>{labels.image.resolution}</span><Select value={parameters.resolution} options={optionList(valuesOr(['1K', '2K', '4K'], capabilities.resolutions))} onChange={resolution => update({ resolution })} /></label>
      <label><span>{labels.image.count}</span><Select value={parameters.count} options={countOptions} onChange={count => update({ count })} /></label>
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
      <label><span>{labels.audio.operation}</span><Segmented block value={parameters.operation} options={choices.operationAudio} onChange={operation => update({ operation })} /></label>
      <label><span>{labels.audio.speed}</span><InputNumber min={0.5} max={2} step={0.1} value={parameters.speed} onChange={speed => update({ speed })} /></label>
      <label><span>{labels.audio.format}</span><Select value={parameters.format} options={optionList(['mp3', 'wav'])} onChange={format => update({ format })} /></label>
    </> : null}
  </section>;

  return <Popover trigger="click" placement="bottom" content={content}>
    <button type="button" className="agent-node-composer__parameters" aria-label={labels[kind]?.title || '节点参数'} disabled={disabled} onClick={stop}>
      <SlidersHorizontal size={14} /><span>{formatNodeParameterSummary(kind, normalized)}</span>
    </button>
  </Popover>;
}
