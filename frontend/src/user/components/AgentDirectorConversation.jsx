import { Button, Input, Segmented, Tag, Tooltip } from 'antd';
import { Bot, Check, CircleAlert, ListTree, Sparkles, X } from 'lucide-react';
import { useState } from 'react';
import { directorPlanPresentation } from '../pages/agent-director-commands';
import { normalizeDirectorMode, summarizeDirectorPlan } from '../pages/agent-director-mode';

const modeOptions = [
  { label: '逐步确认', value: 'supervised' },
  { label: '自动执行', value: 'automatic' }
];

export function AgentDirectorConversation({ mode = 'supervised', plans = [], busy = false, onModeChange, onApprove, onReject }) {
  const [rejectingPlanID, setRejectingPlanID] = useState('');
  const [feedback, setFeedback] = useState('');
  const normalizedMode = normalizeDirectorMode(mode);
  const visiblePlans = (Array.isArray(plans) ? plans : []).filter(plan => plan && plan.status !== 'rejected' && plan.status !== 'applied');

  function submitFeedback(planID) {
    const value = feedback.trim();
    if (!value) return;
    onReject?.(planID, value);
    setFeedback('');
    setRejectingPlanID('');
  }

  return <>
    <header className="agent-director-header">
      <div className="agent-director-heading"><span><Bot size={16} /></span><div><strong>导演</strong><small>{normalizedMode === 'automatic' ? '按目标连续推进，遇到缺少配置会停下' : '每次改动画布前先给你确认'}</small></div></div>
      <Segmented
        aria-label="导演执行模式"
        size="small"
        value={normalizedMode}
        options={modeOptions}
        disabled={busy}
        onChange={value => onModeChange?.(normalizeDirectorMode(value))}
      />
    </header>

    {visiblePlans.length ? <section className="agent-director-plans" aria-label="待确认导演方案">
      {visiblePlans.map(plan => {
        const presentation = directorPlanPresentation(plan);
        const summary = summarizeDirectorPlan(plan);
        const rejecting = rejectingPlanID === presentation.id;
        return <article className="agent-director-plan" key={presentation.id}>
          <div className="agent-director-plan__heading"><span><ListTree size={15} /></span><div><strong>{presentation.title}</strong><small>{presentation.summary}</small></div><Tag>{summary.commandCount} 项改动</Tag></div>
          <div className="agent-director-plan__items">
            {presentation.items.map(item => <div className={`${item.destructive ? 'is-destructive' : ''}${item.requiresExecution ? ' is-execution' : ''}`} key={item.key}>
              <span>{item.requiresExecution ? <Sparkles size={13} /> : item.destructive ? <CircleAlert size={13} /> : <Check size={13} />}</span>
              <strong>{item.label}</strong>
            </div>)}
          </div>
          {summary.destructiveCount ? <p className="agent-director-plan__warning"><CircleAlert size={13} />只会移除画布节点，不会删除项目素材。</p> : null}
          {summary.executionCount ? <p className="agent-director-plan__warning"><Sparkles size={13} />包含真实生成任务，缺少模型或通道时会停止。</p> : null}
          {rejecting ? <div className="agent-director-plan__feedback">
            <Input.TextArea autoFocus value={feedback} autoSize={{ minRows: 2, maxRows: 4 }} placeholder="说一下哪里需要调整" onChange={event => setFeedback(event.target.value)} />
            <div><Button size="small" type="text" onClick={() => { setRejectingPlanID(''); setFeedback(''); }}>取消</Button><Button size="small" type="primary" disabled={!feedback.trim()} onClick={() => submitFeedback(presentation.id)}>提交反馈</Button></div>
          </div> : <div className="agent-director-plan__actions">
            <Button size="small" icon={<X size={13} />} onClick={() => setRejectingPlanID(presentation.id)}>反馈调整</Button>
            <Tooltip title={presentation.executable ? '按上面的清单修改画布' : '将先检查模型与生成通道'}><Button size="small" type="primary" icon={<Check size={13} />} loading={busy} onClick={() => onApprove?.(presentation.id)}>确认执行</Button></Tooltip>
          </div>}
        </article>;
      })}
    </section> : null}
  </>;
}
