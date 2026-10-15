import { Alert, Button, Tag } from 'antd';
import { useState } from 'react';
import { beginGeneration, cancelGeneration, createGenerationState, finishGeneration } from './generationState';

export function GenerationPanel({ onGenerate }) {
  const [state, setState] = useState(createGenerationState);
  async function run() {
    if (state.status === 'running' || !onGenerate) return;
    const running = beginGeneration(state);
    setState(running);
    try {
      const result = await onGenerate(running.requestId);
      setState(current => finishGeneration(current, result));
    } catch (error) {
      setState(current => finishGeneration(current, null, { message: error?.message || '生成失败', retryable: true }));
    }
  }
  return (
    <section className="novel-panel-generation" aria-label="生成状态">
      <Tag color={state.status === 'failed' ? 'red' : 'blue'}>{state.status}</Tag>
      {state.error && <Alert type="error" showIcon message={state.error.message} />}
      <Button loading={state.status === 'running'} disabled={state.status === 'running'} onClick={run}>开始生成</Button>
      <Button disabled={state.status !== 'running'} onClick={() => setState(current => cancelGeneration(current))}>取消</Button>
    </section>
  );
}

export default GenerationPanel;
