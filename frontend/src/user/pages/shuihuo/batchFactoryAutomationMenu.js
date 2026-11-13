import { batchFactoryAutomationRetryAction } from './batchFactoryBookState.js';

export function batchFactoryAutomationMenu(status, hasBooks, stageRetryCount = 0) {
  const state = String(status?.state || 'idle');
  const active = ['running', 'paused', 'scheduled', 'needs_attention'].includes(state);
  const retry = batchFactoryAutomationRetryAction(status);
  const canManualRetry = !active && stageRetryCount > 0 && retry.disabled;
  const retryItem = canManualRetry
    ? { key: 'retry-stages', label: `重试失败步骤（${stageRetryCount}）`, disabled: false, title: '仅重试失败或已中断的步骤，保留已有成果' }
    : { key: 'retry', ...retry, disabled: state === 'running' || state === 'scheduled' || retry.disabled, title: state === 'running' ? '本轮正在执行，请等待结束后重试失败任务' : retry.disabled ? '暂无可重试的失败任务；未开始的书请先启动自动生产' : '从失败或中断的位置重试，保留已完成成果' };
  return [
    { key: 'start', label: '启动自动生产（选择预设）', disabled: active || !hasBooks },
    { key: 'pause', label: '暂停任务（可继续）', disabled: state !== 'running', title: '暂停后续派发，已有成果保留；已提交的模型任务仍可能返回' },
    { key: 'resume', label: '继续未完成任务', disabled: !['paused', 'needs_attention'].includes(state), title: '沿用本轮任务配置，从未完成的位置继续' },
    retryItem,
    { key: 'cancel', label: '结束本轮任务', danger: true, disabled: !active, title: '结束本轮后续派发并保留成果；需要时重新启动，不会撤销已提交给模型的任务' }
  ];
}
