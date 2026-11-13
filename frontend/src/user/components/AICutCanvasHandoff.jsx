import { Alert, Button, Checkbox, Empty, Modal, Select, Spin, Tag } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { getAICutCreativeBinding, transferAICutCreativeAssets, transferAICutCreativeScript } from '../../shared/api/aicut.js';
import { getAgentCanvas, listAgentAssets } from '../../shared/api/agent.js';
import { apiRequest } from '../../shared/api/client.js';
import { scopedAICutOptions } from '../../shared/api/aicutTimelineScope.js';
import { listProjectTimelines, loadProjectTimeline } from '../../shared/api/aicutTimelines.js';
import { creativeTimelineHint, cutEditorHref } from '../pages/aicutCanvasNavigation.js';
import { creativeAssetSelection, creativeScriptText, defaultCreativeAssetSelection, transferableCreativeAssets, transferableCreativeScriptNodes } from '../pages/aicutCreativeCanvas.js';
import './AICutCanvasHandoff.css';

// This view owns no canvas or save queue. All edits remain in the CM workspace.
export default function AICutCanvasHandoff({ taskId, canvas, selectedNodeIds = [], onFlushCanvas, onNavigate = href => window.location.assign(href) }) {
  const [binding, setBinding] = useState(null);
  const [timelines, setTimelines] = useState([]);
  const [target, setTarget] = useState('');
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [retry, setRetry] = useState(0);
  const [modal, setModal] = useState('');
  const [items, setItems] = useState([]);
  const [selected, setSelected] = useState([]);
  const snapshot = useRef(null);
  const barRef = useRef(null);
  const sessionRef = useRef(null);
  const runtime = useRef(null);
  runtime.current = { taskId, canvas, selectedNodeIds, onFlushCanvas };
  const current = session => session?.active && session === sessionRef.current && runtime.current.taskId === session.taskId;
  useEffect(() => {
    const stack = barRef.current?.parentElement;
    const workspace = stack?.closest('.agent-canvas-workspace');
    if (!stack || !workspace) return;
    const measure = () => workspace.style.setProperty('--agent-workspace-header-height', `${Math.ceil(stack.getBoundingClientRect().height)}px`);
    const observer = new ResizeObserver(measure); observer.observe(stack); measure();
    return () => { observer.disconnect(); workspace.style.removeProperty('--agent-workspace-header-height'); };
  }, [binding, loading, error]);
  useEffect(() => {
    const session = { taskId, active: true }; sessionRef.current = session;
    setBinding(null); setTimelines([]); setTarget(''); setLoading(true); setBusy(false); setModal(''); setError(''); setItems([]); setSelected([]);
    if (!taskId) { setLoading(false); return () => { session.active = false; }; }
    (async () => {
      try {
        const origin = await getAICutCreativeBinding(taskId);
        if (!current(session)) return;
        if (!origin) return;
        if (origin.taskId !== taskId || !origin.projectId) throw Error('画布所属工程数据不完整');
        setBinding(origin);
        const available = await listProjectTimelines(origin.projectId);
        if (!current(session)) return;
        setTimelines(available);
        const hint = creativeTimelineHint(window.location.search, taskId);
        if (available.some(item => item.id === hint)) setTarget(hint);
        else setError('原目标时间线已失效，请重新选择。');
      } catch (failure) { if (current(session)) setError(failure.message || '剪辑工程关联读取失败'); }
      finally { if (current(session)) setLoading(false); }
    })();
    return () => { session.active = false; };
  }, [taskId, retry]);
  async function flush(session) {
    if (!current(session)) throw Error('创作项目已切换');
    if (typeof runtime.current.onFlushCanvas !== 'function') throw Error('画布尚未准备完成');
    await runtime.current.onFlushCanvas();
    if (!current(session)) throw Error('创作项目已切换');
  }
  const ready = binding && target && !loading;
  function selectTarget(value) {
    setTarget(value); setError('');
    const url = new URL(window.location.href);
    if (url.searchParams.get('taskId') === taskId) {
      if (value === 'default') url.searchParams.delete('cutTimeline'); else url.searchParams.set('cutTimeline', value);
      window.history.replaceState({}, '', `${url.pathname}${url.search}${url.hash}`);
    }
  }
  function targetRequest(path, options) {
    return apiRequest(path, scopedAICutOptions(path, options, { projectId: binding.projectId, timelineId: target }));
  }
  async function returnToEditor() {
    const session = sessionRef.current; if (!ready || busy) return;
    setBusy(true); setError('');
    try { await flush(session); if (current(session)) onNavigate(cutEditorHref(binding.projectId, target)); }
    catch (failure) { if (current(session)) setError(failure.message); }
    finally { if (current(session)) setBusy(false); }
  }
  async function choose(kind) {
    const session = sessionRef.current; if (!ready || busy) return;
    setBusy(true); setError(''); setItems([]); setSelected([]); setModal(kind);
    try {
      await flush(session);
      if (kind === 'assets') {
        const result = await listAgentAssets(taskId); if (!current(session)) return;
        const available = transferableCreativeAssets(result.assets);
        setItems(available); setSelected(defaultCreativeAssetSelection(available, runtime.current.canvas?.nodes, runtime.current.selectedNodeIds));
      } else {
        const [saved, timeline] = await Promise.all([getAgentCanvas(taskId), loadProjectTimeline(binding.projectId, target)]);
        if (!current(session)) return;
        const nodes = transferableCreativeScriptNodes(saved.canvas?.nodes);
        snapshot.current = { canvasRevision: saved.canvas?.revision, timelineRevision: timeline.revision };
        setItems(nodes); setSelected(nodes.filter(node => runtime.current.selectedNodeIds.includes(node.id)).map(node => node.id).slice(0, 10));
      }
    } catch (failure) { if (current(session)) setError(failure.message || '交接内容读取失败'); }
    finally { if (current(session)) setBusy(false); }
  }
  async function send() {
    const session = sessionRef.current; if (!ready || busy || !selected.length) return;
    setBusy(true); setError('');
    try {
      await flush(session);
      if (modal === 'assets') {
        await transferAICutCreativeAssets(binding.projectId, creativeAssetSelection(selected, items), targetRequest);
      } else {
        const saved = await getAgentCanvas(taskId); if (!current(session)) return;
        if (saved.canvas?.revision !== snapshot.current?.canvasRevision) throw Error('画布内容已更新，请重新读取并确认剧本。');
        const ids = items.filter(node => selected.includes(node.id)).map(node => node.id);
        await transferAICutCreativeScript(binding.projectId, ids, snapshot.current.timelineRevision, targetRequest);
      }
      if (current(session)) onNavigate(cutEditorHref(binding.projectId, target));
    } catch (failure) { if (current(session)) setError(failure.message || '送入剪辑失败'); }
    finally { if (current(session)) setBusy(false); }
  }
  if (!taskId || (!loading && !binding && !error)) return null;
  const limit = modal === 'assets' ? 30 : 10;
  return <section ref={barRef} className="aicut-canvas-handoff" aria-label="当前画布的剪辑工程">
    <div className="aicut-canvas-handoff__origin"><Tag>同一张画布</Tag><strong>{binding ? binding.projectTitle : '正在核实剪辑工程…'}</strong></div>
    {binding && <><Select aria-label="交接目标时间线" placeholder="选择目标时间线" value={target || undefined} options={timelines.map(item => ({ value: item.id, label: item.title }))} disabled={loading || busy || Boolean(modal)} onChange={selectTarget} />
      <Button disabled={!ready || busy} onClick={() => void returnToEditor()}>返回剪辑</Button>
      <Button disabled={!ready || busy} onClick={() => void choose('script')}>发送剧本</Button>
      <Button type="primary" disabled={!ready || busy} onClick={() => void choose('assets')}>送入剪辑</Button></>}
    {error && !modal && <Alert type="error" message={error} action={<Button size="small" onClick={() => setRetry(value => value + 1)}>重新读取</Button>} />}
    <Modal title={modal === 'assets' ? '送入剪辑' : '发送剧本'} open={Boolean(modal)} onCancel={() => { if (!busy) setModal(''); }} onOk={() => void send()} okText={`确认 ${selected.length} 项并返回剪辑`} cancelText="继续创作" confirmLoading={busy} okButtonProps={{ disabled: busy || !selected.length }} closable={!busy} maskClosable={!busy}>
      <p>{binding?.projectTitle} · {timelines.find(item => item.id === target)?.title}。{modal === 'assets' ? '素材进入工程素材库，已有素材不会重复导入。' : '所选文本按列表顺序追加到目标时间线剧本，已有轨道保持不变。'}</p>
      {error && <Alert type="error" message={error} action={<Button size="small" disabled={busy} onClick={() => void choose(modal)}>重新读取</Button>} />}
      {busy && !items.length ? <Spin /> : !items.length ? <Empty description={modal === 'assets' ? '还没有可用素材，请先上传或生成。' : '还没有填写正文的故事或分镜卡片。'} /> : <Checkbox.Group className="aicut-canvas-handoff__items" value={selected} disabled={busy} onChange={ids => setSelected([...new Set(ids)].slice(0, limit))}>{items.map(item => <Checkbox key={item.id} value={item.id} disabled={busy || (!selected.includes(item.id) && selected.length >= limit)}><strong>{item.name || item.title || item.id}</strong>{modal === 'script' && <span>{creativeScriptText(item).slice(0, 180)}</span>}</Checkbox>)}</Checkbox.Group>}
    </Modal>
  </section>;
}
