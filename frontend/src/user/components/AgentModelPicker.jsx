import { Select } from 'antd';
import { Bot, CircleAlert } from 'lucide-react';
import { capabilityDescription, compatibleModels, expectedDuration, reconcileNodeModel } from '../pages/agent-model-capabilities';
import { providerLabel } from '../pages/shuihuo/providerModelBinding';

export function AgentModelPicker({ kind, config = {}, models = [], onChange, disabled = false }) {
  const state = reconcileNodeModel(config, models, kind);
  const choices = compatibleModels(models, { kind, provider: state.provider });
  const options = choices.map(model => ({ value: model.id, label: model.displayName || model.name || model.modelId || model.id, model }));
  if (state.modelId && !state.available) {
    options.unshift({ value: state.modelId, label: `${state.modelId}（当前模型不可用）`, disabled: true, unavailable: true });
  }

  return <Select
    aria-label={`${kind === 'text' ? '文字' : kind === 'image' ? '图片' : kind === 'video' ? '视频' : '音频'}模型`}
    className="agent-model-picker"
    variant="borderless"
    size="small"
    showSearch
    popupMatchSelectWidth={false}
    popupClassName="agent-model-picker__dropdown"
    allowClear={false}
    disabled={disabled}
    value={state.modelId || undefined}
    placeholder={state.reason || '选择模型'}
    notFoundContent="当前通道没有可用模型，请在设置中切换生成通道或到 API 配置添加模型"
    optionFilterProp="label"
    options={options}
    optionRender={option => {
      const model = option.data.model;
      return <div className={`agent-model-picker__option${option.data.unavailable ? ' is-unavailable' : ''}`}>
        <span className="agent-model-picker__icon">{option.data.unavailable ? <CircleAlert size={15} /> : <Bot size={15} />}</span>
        <span><strong>{option.data.label}</strong><small>{model ? `${providerLabel(model.provider)} · ${capabilityDescription(model)}${expectedDuration(model) ? ` · ${expectedDuration(model)}` : ''}` : '当前模型不可用'}</small></span>
      </div>;
    }}
    onChange={modelId => {
      const model = choices.find(item => item.id === modelId);
      if (!model) return;
      onChange?.({ provider: model.provider, modelId, model: modelId });
    }}
  />;
}
