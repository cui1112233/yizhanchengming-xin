import { useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { approveAgentDirectorPlan, askAgent, getAgentTask } from '../../shared/api/agent.js';
import { lockCanvasPetInput, runCanvasPetRequest } from '../pages/canvas-pet-service.js';
import { directorPlanPresentation } from '../pages/agent-director-commands.js';
import './CanvasPetPanel.css';

export default function CanvasPetPanel({ taskId, selectedCount, nodeCount, revision, services, onEditSelection, onLocateSelection }) {
  const [draft, setDraft] = useState('');
  const [messages, setMessages] = useState([]);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const [loaded, setLoaded] = useState(false);
  const [reload, setReload] = useState(0);
  const [applied, setApplied] = useState([]);
  const [applying, setApplying] = useState(false);
  const pending = useRef(false);
  const controller = useRef(null);
  const releaseLock = useRef(() => {});
  const mounted = useRef(true);
  useEffect(() => {
    let active = true; mounted.current = true; setLoaded(false); setError('');
    getAgentTask(taskId).then(result => { if (active) { setMessages(result.task?.messages || []); setLoaded(true); } }).catch(failure => { if (active) setError(failure.message || '项目对话读取失败'); });
    return () => { active = false; mounted.current = false; controller.current?.abort(); releaseLock.current(); };
  }, [taskId, reload]);
  async function run(planId) {
    if (pending.current || !loaded || (!planId && !draft.trim())) return;
    pending.current = true; setApplying(Boolean(planId));
    const request = new AbortController(); controller.current = request;
    const content = draft.trim(); setBusy(true); setError('');
    let unlock = () => {};
    try {
      if (planId) {
        services.checkCanApply?.();
        const releaseWrites = services.beginApply?.() || (() => {});
        unlock = releaseWrites;
        const releaseInput = lockCanvasPetInput(window);
        unlock = () => { releaseInput(); releaseWrites(); };
        releaseLock.current = unlock;
      }
      const result = await runCanvasPetRequest({ ...services, taskId, ask: askAgent, approve: approveAgentDirectorPlan }, { prompt: content, planId, signal: request.signal });
      if (!mounted.current || request.signal.aborted) return;
      if (planId) {
        if (!result?.canvas) throw new Error('接口没有返回执行后的画布');
        const accepted = await services.acceptCanvas(result.canvas);
        if (accepted === false) throw new Error('画布有更新，请重新读取');
        if (!mounted.current) return;
        setApplied(items => [...items, planId]);
      } else {
        if (!Array.isArray(result.task?.messages)) throw new Error('接口没有返回项目对话');
        setMessages(result.task.messages); setDraft(value => value.trim() === content ? '' : value);
      }
    } catch (failure) {
      if (mounted.current) setError(request.signal.aborted ? '已停止等待；如已提交，请从项目对话检查结果。' : failure.message || '处理失败，请重试');
    } finally { unlock(); pending.current = false; if (mounted.current) { setBusy(false); setApplying(false); } if (controller.current === request) controller.current = null; }
  }
  return <section className="canvas-pet-panel" aria-label="画布创作服务">
    {applying && createPortal(<div className="canvas-pet-write-lock" role="status">正在应用并保存画布，请稍候…</div>, document.body)}
    <p className="canvas-pet-context">当前画布 · {nodeCount} 张卡片 · 已选 {selectedCount} 张 · 版本 {revision}</p>
    <p>整理故事、搭建卡片和引用关系。先查看修改清单，确认后应用到当前画布。</p>
    <div className="canvas-pet-actions">
      <button disabled={selectedCount !== 1 || busy} onClick={onEditSelection}>编辑选中卡片</button>
      <button disabled={selectedCount !== 1 || busy} onClick={onLocateSelection}>定位选中卡片</button>
    </div>
    <p className="canvas-pet-note">图片、视频和音频在卡片编辑器中选模型、生成并查看结果；提出生成方案不代表素材已生成。</p>
    {!messages.length && <div className="canvas-pet-empty">说说你想做什么。可以先选中故事或参考卡片，让我围绕它整理创作方案。</div>}
    <div className="canvas-pet-history" aria-live="polite">{messages.slice(-16).map((entry, index) => {
      const plan = entry.context?.directorPlan;
      const presentation = plan ? directorPlanPresentation(plan) : null;
      const stale = entry.context?.canvasPetRevision !== undefined && Number(entry.context.canvasPetRevision) !== revision;
      const alreadyApplied = applied.includes(plan?.id) || plan?.status === 'applied';
      return <article key={entry.id || index} className={`canvas-pet-message is-${entry.role}`}>
        <strong>{entry.role === 'user' ? '你' : '创作助手'}</strong><p>{entry.content}</p>
        {presentation && <div className="canvas-pet-plan"><strong>{presentation.title}</strong><p>{presentation.summary}</p><ul>{presentation.items.map(item => <li key={item.key}>{item.label}{item.requiresExecution ? '（需在卡片执行）' : ''}</li>)}</ul>
          {!presentation.executable && <p>该方案包含生成步骤，尚未应用。请先要求只搭建卡片，再从卡片中执行生成。</p>}
          <button disabled={busy || stale || !presentation.executable || alreadyApplied} onClick={() => void run(plan.id)}>{alreadyApplied ? '已应用到画布' : stale ? '画布已更新，请重新出方案' : '确认应用到画布'}</button>
        </div>}
      </article>;
    })}</div>
    {error && <div role="alert">{error}</div>}
    {!loaded && !busy && <button onClick={() => setReload(value => value + 1)}>重新读取项目对话</button>}
    <label className="canvas-pet-composer">创作要求<textarea aria-label="画布创作要求" value={draft} maxLength={6000} rows={3} onChange={event => setDraft(event.target.value)} placeholder="例如：根据选中的故事，搭建角色、场景和镜头卡片" /></label>
    <div className="canvas-pet-actions"><button disabled={busy || !loaded || !draft.trim()} onClick={() => void run()}>{busy ? '正在处理…' : '生成创作方案'}</button>{busy && !applying && <button onClick={() => controller.current?.abort()}>停止等待</button>}</div>
  </section>;
}
