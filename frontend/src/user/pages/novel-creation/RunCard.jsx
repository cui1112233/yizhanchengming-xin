import { Button, Progress } from 'antd';
import { PHASES, STATUS, TERMINAL, parseJSON, validRun } from './contracts.js';

export default function RunCard({ snapshot, onAction, onBlueprint, busy }) {
  const run = validRun(snapshot?.run); if (!run) return null;
  const runtime = parseJSON(run.runtimeSnapshot), secondary = runtime.confirmationKind === 'next_model_call';
  const unknown = run.errorCode === 'novel_upstream_result_unknown';
  const progress = Math.max(0, ...(snapshot.events || []).map(e => Number(e.progress) || 0));
  const calls = Object.entries(runtime.modelCalls || {});
  const audit = !runtime.operation && ['succeeded', 'succeeded_with_warnings'].includes(run.status) && typeof runtime.finalAudit?.score === 'number' && Number.isFinite(runtime.finalAudit.score) ? runtime.finalAudit : null;
  return <section className="novel-run" aria-label="当前任务"><b>{run.status === 'analyzing' ? '资料分析' : STATUS[run.status] || '处理中'}</b><span>{STATUS[snapshot.executionState] || ''}</span><Progress percent={progress} showInfo />
    {unknown ? <p role="alert">上游结果待确认，已暂停自动重试。请核对原轮次；不会自动重新调用模型。</p> : run.errorMessage && <p role="alert">{run.errorMessage}</p>}
    {audit && <section aria-label="最终审核"><b>最终审核：{audit.score}分</b>{Array.isArray(audit.issues) && audit.issues.length > 0 && <><p>未解决问题：</p><ul>{audit.issues.map((issue, index) => <li key={issue.id || index}>{({ high: '严重', critical: '严重', medium: '一般', low: '轻微' })[issue.severity] || '提示'}：{typeof issue.description === 'string' ? issue.description : '请检查本章内容'}</li>)}</ul></>}</section>}
    {snapshot.executionState === 'save_pending' && <p>模型结果已保留，正在等待恢复保存；恢复使用同一任务，不会重新生成。</p>}
    {run.status === 'awaiting_confirmation' && <><p>{secondary ? '请确认后继续下一次模型调用' : '请确认创作蓝图后继续'}</p>{secondary && runtime.nextModelCall && <p>下一步：{PHASES.find(([phase]) => runtime.callKey?.includes(phase))?.[1] || '资料处理'} · 选择模型：{runtime.nextModelCall.chosenModelId} · 实际模型：{runtime.nextModelCall.provider?.actualModel || '尚无记录'}</p>}{!secondary && <p>{runtime.foundationPlan?.coreConflict || runtime.foundationPlan?.title}</p>}<div className="novel-actions"><Button type="primary" loading={busy} onClick={() => onAction('confirm', secondary ? { confirmationId: runtime.confirmationId } : {})}>{secondary ? '确认下一次模型调用' : '确认蓝图'}</Button>{!secondary && <Button onClick={() => onBlueprint(runtime.foundationPlan || {})}>修改蓝图</Button>}</div></>}
    <div className="novel-actions">{!unknown && (run.status === 'failed' || ['ready', 'save_pending'].includes(snapshot.executionState)) && <Button loading={busy} onClick={() => onAction('retry', {})}>{snapshot.executionState === 'save_pending' ? '恢复保存结果' : '从失败处重试'}</Button>}
      {!TERMINAL.has(run.status) && <Button onClick={() => onAction('cancel', {})}>停止后续步骤</Button>}</div>
    {run.status === 'cancelled' && <small>已停止后续步骤，已发送请求可能仍由上游处理。</small>}
    {calls.length > 0 && <details><summary>实际模型与参数</summary><p>“不支持，未发送”表示提供方不接受此参数，不能视为已生效。</p>{calls.map(([key, call]) => <div className="novel-item" key={key}><b>{PHASES.find(([phase]) => key.includes(phase))?.[1] || '处理步骤'}</b><p>选择：{call.chosenModelId} · 实际：{call.provider?.actualModel || '尚无记录'} · 通道：{call.provider?.channelId || '默认通道'}</p><p>温度 {call.temperature}：{call.temperatureStatus === 'sent' ? '已发送' : '不支持，未发送'} · 输出上限 {call.maxTokens}：{call.maxTokensStatus === 'sent' ? '已发送' : '不支持，未发送'}</p></div>)}</details>}
    {runtime.agentSettings && <details><summary>本次有效Agent设置</summary><p>审核通过分{runtime.agentSettings.passScore} · 最多修订{runtime.agentSettings.maxRevisionRounds}轮 · 基础温度{runtime.agentSettings.baseTemperature} · 温度增量{runtime.agentSettings.temperatureStep}</p><p>长度整理{runtime.agentSettings.enableLengthNorm ? '开启' : '关闭'} · 后处理{runtime.agentSettings.enablePostWrite ? '开启' : '关闭'} · 表达检查{runtime.agentSettings.enableAiTells ? '开启' : '关闭'} · 二次确认{runtime.agentSettings.requireSecondaryConfirm ? '开启' : '关闭'}</p></details>}
  </section>;
}
