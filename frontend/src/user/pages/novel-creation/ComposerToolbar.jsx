import { Button, InputNumber, Popover, Select, Tooltip } from 'antd';
import { ArrowUp, Brain, CircleDashed, SlidersHorizontal } from 'lucide-react';
import { THINKING } from './contracts.js';
import './ComposerToolbar.css';

export default function ComposerToolbar({ models, modelId, onModelChange, thinkingLevel, onThinkingChange, thinkingLabel = '创作思考强度', contextBudget, onBudgetChange, referenceControl, locked = false, sending = false, disabled = false, hasContent = false, sendLabel = '发送', onSend }) {
  return <div className="novel-composer-toolbar">
    <div className="novel-composer-tools">
      <div className="novel-composer-model"><CircleDashed size={17} aria-hidden="true"/><Select aria-label="创作文本模型" variant="borderless" value={modelId || undefined} options={models} disabled={locked} onChange={onModelChange} placeholder="选择文本模型" popupMatchSelectWidth={320}/></div>
      {referenceControl}
      <Popover trigger="click" placement="topLeft" arrow={false} rootClassName="novel-composer-popover" content={<div className="novel-composer-setting"><b>思考要求</b><p>作为提示词要求传给模型。</p><Select aria-label={thinkingLabel} value={thinkingLevel} options={THINKING} disabled={locked} onChange={onThinkingChange}/></div>}>
        <Tooltip title="思考要求"><Button type="text" aria-label="调整思考要求" className="novel-composer-tool" disabled={locked} icon={<Brain size={17}/>}/></Tooltip>
      </Popover>
      {onBudgetChange && <Popover trigger="click" placement="topLeft" arrow={false} rootClassName="novel-composer-popover" content={<div className="novel-composer-setting"><b>上下文设置</b><p>本轮最多携带的历史内容字数。</p><label>上下文字数<InputNumber aria-label="上下文字数" min={4000} max={64000} step={1000} value={contextBudget} disabled={locked} onChange={onBudgetChange}/></label></div>}>
        <Tooltip title="上下文设置"><Button type="text" aria-label="调整上下文设置" className="novel-composer-tool" disabled={locked} icon={<SlidersHorizontal size={17}/>}/></Tooltip>
      </Popover>}
    </div>
    <Tooltip title={`${sendLabel}（Enter）`}><Button aria-label={sendLabel} type="text" shape="circle" className={`novel-composer-send${hasContent ? ' is-ready' : ''}`} icon={<ArrowUp size={19}/>} loading={sending} disabled={disabled} onClick={onSend}/></Tooltip>
  </div>;
}
