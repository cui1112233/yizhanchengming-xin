import { Input, InputNumber, Popover, Select, Switch } from 'antd';
import { Settings2 } from 'lucide-react';
import { providerOptionsForModels, reconcileProviderModel } from '../pages/shuihuo/providerModelBinding';

const channelLabels = { image: '图片生成通道', video: '视频引擎', text: '文本通道', audio: '音频通道' };

export function AgentNodeSettingsPanel({ kind, config = {}, models = [], onChange, disabled = false }) {
  const providerOptions = providerOptionsForModels(models, kind);
  const stop = event => event.stopPropagation();
  const advanced = config.advanced || {};
  const patchAdvanced = patch => onChange?.({ advanced: patch });
  const content = <section className={`agent-node-settings-panel agent-node-settings-panel--${kind}`} aria-label={`${channelLabels[kind]}设置`} onMouseDown={stop} onClick={stop}>
    <label><span>{channelLabels[kind]}</span><Select
      value={config.provider || 'personal_api'}
      options={providerOptions}
      getPopupContainer={trigger => trigger.parentElement}
      onChange={provider => {
        const next = reconcileProviderModel({ models, kind, provider, modelId: config.modelId || config.model });
        onChange?.({ provider: next.provider, modelId: next.modelId, model: next.modelId });
      }}
    /></label>
    {kind === 'image' ? <>
      <label className="is-inline"><span>自动连接结果</span><Switch checked={advanced.autolink !== false} onChange={autolink => patchAdvanced({ autolink })} /></label>
      <label><span>随机种子</span><InputNumber min={-1} precision={0} value={advanced.seed ?? -1} onChange={seed => patchAdvanced({ seed })} /></label>
      <label><span>负面提示词</span><Input value={advanced.negativePrompt || ''} onChange={event => patchAdvanced({ negativePrompt: event.target.value })} /></label>
    </> : null}
    {kind === 'video' ? <>
      <label className="is-inline"><span>自动连接结果</span><Switch checked={advanced.autolink !== false} onChange={autolink => patchAdvanced({ autolink })} /></label>
      <label><span>镜头运动</span><Select allowClear value={advanced.cameraMotion || undefined} options={['自动', '推进', '拉远', '环绕', '平移'].map(value => ({ value, label: value }))} getPopupContainer={trigger => trigger.parentElement} onChange={cameraMotion => patchAdvanced({ cameraMotion })} /></label>
    </> : null}
    {kind === 'text' ? <label><span>结构要求</span><Input value={advanced.structure || ''} onChange={event => patchAdvanced({ structure: event.target.value })} /></label> : null}
    {kind === 'audio' ? <>
      <label><span>声音</span><Input value={config.voice || ''} onChange={event => onChange?.({ voice: event.target.value })} /></label>
      <label><span>情绪</span><Input value={advanced.emotion || ''} onChange={event => patchAdvanced({ emotion: event.target.value })} /></label>
    </> : null}
  </section>;

  return <Popover trigger="click" placement="bottomRight" content={content}>
    <button type="button" className="agent-node-composer__settings" aria-label="设置" title="设置" disabled={disabled} onClick={stop}><Settings2 size={15} /></button>
  </Popover>;
}
