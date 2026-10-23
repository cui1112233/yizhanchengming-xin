import { Select } from 'antd';
import { modelOptionsForProvider, normalizeModelProvider, providerOptionsForModels, reconcileProviderModel } from '../pages/shuihuo/providerModelBinding';

export function AgentBatchModelPicker({ kind, config = {}, models = [], onChange, disabled = false }) {
  const provider = normalizeModelProvider(config.provider || 'personal_api');
  const providerOptions = providerOptionsForModels(models, kind);
  const modelOptions = modelOptionsForProvider(models, kind, provider);
  const modelId = modelOptions.some(option => option.value === config.modelId) ? config.modelId : undefined;
  const label = kind === 'image' ? '图片' : kind === 'video' ? '视频' : '音频';

  return <div className={`agent-batch-model-picker is-${kind}`}>
    {kind === 'image' || kind === 'video' ? <Select
      aria-label={`${label}生成通道`}
      value={provider}
      options={providerOptions}
      disabled={disabled}
      onChange={nextProvider => {
        const next = reconcileProviderModel({ models, kind, provider: nextProvider, modelId: config.modelId });
        onChange?.({ provider: next.provider, modelId: next.modelId });
      }}
    /> : null}
    <Select
      aria-label={`${label}模型`}
      value={modelId}
      placeholder="选择模型"
      options={modelOptions}
      disabled={disabled}
      onChange={nextModelId => onChange?.({ provider, modelId: nextModelId })}
      notFoundContent="当前通道没有可用模型"
    />
  </div>;
}
