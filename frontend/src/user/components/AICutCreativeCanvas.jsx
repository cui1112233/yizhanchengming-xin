import { Alert, Button, Checkbox, Empty, Modal, Space, Spin, Tag, message } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { ensureAICutCreativeCanvas, transferAICutCreativeAssets, transferAICutCreativeScript } from '../../shared/api/aicut.js';
import { getAgentCanvas, listAgentAssets, saveAgentCanvas } from '../../shared/api/agent.js';
import { getCurrentUsername } from '../../shared/api/auth.js';
import { canvasSaveStatusText, createCanvasSaveQueue, mergePendingCanvas } from '../pages/agent-canvas-state.js';
import { clearCanvasDraft, isDraftAtLeastAsFresh, loadCanvasDraft } from '../pages/agent-canvas-draft-storage.js';
import { createCreativeCard, creativeCardTypes, creativeAssetSelection, creativeScriptText, defaultCreativeAssetSelection, persistCreativeCanvasDraft, transferableCreativeAssets, transferableCreativeScriptNodes } from '../pages/aicutCreativeCanvas.js';
import { AgentCanvas } from './AgentCanvas';
import './AICutCreativeCanvas.css';

export default function AICutCreativeCanvas({ projectId, projectTitle, timelineRevision, petServiceScope, onOpenEditor, onTransferred }) {
  const [canvas, setCanvas] = useState(null);
  const [taskId, setTaskId] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [status, setStatus] = useState('idle');
  const [retry, setRetry] = useState(0);
  const [transferOpen, setTransferOpen] = useState(false);
  const [assets, setAssets] = useState([]);
  const [assetIds, setAssetIds] = useState([]);
  const [assetError, setAssetError] = useState('');
  const [assetLoading, setAssetLoading] = useState(false);
  const [transferring, setTransferring] = useState(false);
  const [scriptOpen, setScriptOpen] = useState(false);
  const [scriptNodes, setScriptNodes] = useState([]);
  const [scriptIds, setScriptIds] = useState([]);
  const [scriptLoading, setScriptLoading] = useState(false);
  const [scriptSending, setScriptSending] = useState(false);
  const [scriptError, setScriptError] = useState('');
  const sessionRef = useRef(null);
  const projectRef = useRef(projectId);
  const selectedRef = useRef([]);
  const canvasUIRef = useRef(null);
  projectRef.current = projectId;

  function isCurrent(session) {
    return session && session.active && sessionRef.current === session && projectRef.current === session.projectId;
  }

  useEffect(() => {
    const session = { projectId, username: getCurrentUsername(), active: true, queue: null, canvas: null, taskId: '' };
    sessionRef.current = session;
    setLoading(true); setError(''); setCanvas(null); setTaskId(''); setStatus('idle');
    setTransferOpen(false); setAssets([]); setAssetIds([]); setAssetError(''); setAssetLoading(false); setTransferring(false);
    setScriptOpen(false); setScriptNodes([]); setScriptIds([]); setScriptLoading(false); setScriptSending(false); setScriptError('');
    selectedRef.current = [];
    async function load() {
      try {
        const result = await ensureAICutCreativeCanvas(projectId);
        if (!isCurrent(session)) return;
        if (!result?.taskId || !result?.canvas) throw new Error('创作画布读取不完整，请重试');
        session.taskId = result.taskId;
        const replace = next => {
          session.canvas = next;
          if (isCurrent(session)) setCanvas(next);
        };
        session.queue = createCanvasSaveQueue({
          revision: result.canvas.revision,
          save: async (next, revision) => {
            const result = await saveAgentCanvas(session.taskId, next, revision);
            // Keep newer pending edits recoverable even if the page closes before
            // their next save begins after this successful request.
            persistCreativeCanvasDraft(window.localStorage, session, Number(result?.canvas?.revision) || revision + 1);
            return result;
          },
          reload: () => getAgentCanvas(session.taskId),
          onCanvas: replace,
          onStatus: next => {
            // A previous request can advance the revision while newer edits wait.
            // Rebase their recovery draft before the next request can fail.
            if (next === 'saving') persistCreativeCanvasDraft(window.localStorage, session);
            if (isCurrent(session)) setStatus(next);
          },
          onSynced: saved => {
            const draft = loadCanvasDraft(window.localStorage, session.username, session.taskId);
            if (draft && Number(draft.baseRevision) <= Number(saved.revision)
              && JSON.stringify(draft.canvas.nodes) === JSON.stringify(saved.nodes)
              && JSON.stringify(draft.canvas.edges) === JSON.stringify(saved.edges)
              && JSON.stringify(draft.canvas.viewport) === JSON.stringify(saved.viewport)) {
              clearCanvasDraft(window.localStorage, session.username, session.taskId);
            }
          }
        });
        const draft = loadCanvasDraft(window.localStorage, session.username, session.taskId);
        const restore = draft && isDraftAtLeastAsFresh(draft, result.canvas);
        const initial = restore ? mergePendingCanvas(result.canvas, draft.canvas) : result.canvas;
        if (draft && !restore) clearCanvasDraft(window.localStorage, session.username, session.taskId);
        replace(initial);
        if (restore) session.queue.enqueue(initial);
        setTaskId(session.taskId);
      } catch (failure) {
        if (isCurrent(session)) setError(failure?.message || '创作画布读取失败');
      } finally {
        if (isCurrent(session)) setLoading(false);
      }
    }
    void load();
    const flush = () => { void session.queue?.flushRequired().catch(() => {}); };
    window.addEventListener('pagehide', flush);
    return () => {
      session.active = false;
      window.removeEventListener('pagehide', flush);
      // Finish only this project's queue; the local draft remains if saving fails.
      if (session.queue) void session.queue.flushRequired().catch(() => {}).finally(() => session.queue.dispose());
      if (sessionRef.current === session) sessionRef.current = null;
    };
  }, [projectId, retry]);

  function changeCanvas(next) {
    const session = sessionRef.current;
    if (!isCurrent(session) || !session.queue) return;
    session.canvas = next;
    setCanvas(next);
    persistCreativeCanvasDraft(window.localStorage, session);
    session.queue.enqueue(next);
  }

  async function flushCanvas() {
    const session = sessionRef.current;
    if (!isCurrent(session) || !session.queue) throw new Error('画布尚未准备完成，请稍后再试');
    const result = await session.queue.flushRequired();
    if (!isCurrent(session)) throw new Error('工程已切换');
    return result;
  }

  async function refreshCanvas(incoming) {
    const session = sessionRef.current;
    if (!isCurrent(session)) return;
    const next = incoming?.nodes ? incoming : (await getAgentCanvas(session.taskId)).canvas;
    if (isCurrent(session) && next) await session.queue.acceptServerCanvas(next);
  }

  function addCard(key) {
    const session = sessionRef.current;
    if (!isCurrent(session) || !session.canvas) return;
    const viewport = session.canvas.viewport || { x: 0, y: 0, zoom: 1 };
    const offset = (session.canvas.nodes?.length || 0) % 5 * 32;
    const id = `creative_${globalThis.crypto.randomUUID().replaceAll('-', '')}`;
    const node = createCreativeCard(key, { id, position: { x: (100 - viewport.x) / viewport.zoom + offset, y: (80 - viewport.y) / viewport.zoom + offset } });
    changeCanvas({ ...session.canvas, nodes: [...(session.canvas.nodes || []), node] });
    selectedRef.current = [id];
    window.requestAnimationFrame(() => { if (isCurrent(session)) canvasUIRef.current?.locateNode(id); });
  }

  async function loadAssets() {
    const session = sessionRef.current;
    if (!isCurrent(session) || !session.taskId) return;
    setTransferOpen(true); setAssetLoading(true); setAssetError(''); setAssets([]); setAssetIds([]);
    try {
      await flushCanvas();
      const result = await listAgentAssets(session.taskId);
      if (!isCurrent(session)) return;
      const available = transferableCreativeAssets(result?.assets);
      setAssets(available);
      setAssetIds(defaultCreativeAssetSelection(available, session.canvas.nodes, selectedRef.current));
    } catch (failure) {
      if (isCurrent(session)) setAssetError(failure?.message || '素材读取失败');
    } finally {
      if (isCurrent(session)) setAssetLoading(false);
    }
  }

  async function transfer() {
    const session = sessionRef.current;
    const ids = creativeAssetSelection(assetIds, assets);
    if (!isCurrent(session) || !ids.length || transferring) return;
    setTransferring(true); setAssetError('');
    try {
      await flushCanvas();
      const result = await transferAICutCreativeAssets(session.projectId, ids);
      if (!isCurrent(session)) return;
      message.success(`已送入 ${result?.assets?.length || 0} 个素材`);
      setTransferOpen(false);
      await onTransferred?.({ assets: result?.assets || [] });
    } catch (failure) {
      if (isCurrent(session)) setAssetError(failure?.message || '素材送入失败');
    } finally {
      if (isCurrent(session)) setTransferring(false);
    }
  }

  async function openEditor() {
    try { await flushCanvas(); onOpenEditor?.(); }
    catch (failure) { message.error(failure?.message || '请先保存画布'); }
  }

  async function loadScriptNodes() {
    const session = sessionRef.current;
    if (!isCurrent(session)) return;
    setScriptOpen(true); setScriptLoading(true); setScriptError(''); setScriptNodes([]); setScriptIds([]);
    try {
      await flushCanvas();
      const result = await getAgentCanvas(session.taskId);
      if (!isCurrent(session)) return;
      const nodes = transferableCreativeScriptNodes(result?.canvas?.nodes);
      setScriptNodes(nodes);
      // Only explicitly selected canvas cards are preselected; character cards are never sent automatically.
      setScriptIds(nodes.filter(node => selectedRef.current.includes(node.id)).map(node => node.id).slice(0, 10));
    } catch (failure) {
      if (isCurrent(session)) setScriptError(failure?.message || '剧本读取失败');
    } finally {
      if (isCurrent(session)) setScriptLoading(false);
    }
  }

  async function sendScript() {
    const session = sessionRef.current;
    if (!isCurrent(session) || !scriptIds.length || scriptSending) return;
    setScriptSending(true); setScriptError('');
    try {
      await flushCanvas();
      const saved = await getAgentCanvas(session.taskId);
      if (!isCurrent(session)) return;
      const available = new Set(transferableCreativeScriptNodes(saved?.canvas?.nodes).map(node => node.id));
      if (scriptIds.some(id => !available.has(id))) throw new Error('所选卡片内容已变化，请重新读取后选择');
      const orderedIds = scriptNodes.filter(node => scriptIds.includes(node.id)).map(node => node.id);
      const result = await transferAICutCreativeScript(session.projectId, orderedIds, timelineRevision);
      if (!isCurrent(session)) return;
      setScriptOpen(false); setScriptIds([]);
      message.success('所选内容已追加到剪辑剧本');
      await onTransferred?.({ timeline: result?.timeline });
    } catch (failure) {
      if (isCurrent(session)) setScriptError(failure?.message || '剧本发送失败');
    } finally {
      if (isCurrent(session)) setScriptSending(false);
    }
  }

  return <section className="agent-page utility-workbench agent-workspace-v3 aicut-creative-canvas" aria-label="准备素材">
    <header className="aicut-creative-canvas__toolbar">
      <div className="aicut-creative-canvas__title"><strong>准备素材</strong><span title={projectTitle}>{projectTitle || '当前工程'} · 故事、分镜与生成素材</span></div>
      <Space wrap size={6}>
        <Tag color={status === 'error' || status === 'conflict' ? 'warning' : undefined} role="status">{canvasSaveStatusText[status]}</Tag>
        {status === 'error' && <Button size="small" onClick={() => void sessionRef.current?.queue?.retry()}>重试保存</Button>}
        <Button disabled={!canvas} onClick={() => void openEditor()}>剪已有素材</Button>
        <Button disabled={!canvas} onClick={() => void loadScriptNodes()}>发送剧本</Button>
        <Button type="primary" disabled={!canvas} onClick={() => void loadAssets()}>送入剪辑</Button>
      </Space>
    </header>
    <div className="aicut-creative-canvas__shortcuts"><span>添加卡片</span>{creativeCardTypes.map(item => <Button key={item.key} size="small" disabled={!canvas} onClick={() => addCard(item.key)}>＋{item.label}</Button>)}<span className="aicut-creative-canvas__hint">连接卡片引用内容；双击编辑，右键添加图片、视频或音频。</span></div>
    {status === 'error' && <Alert type="error" showIcon message="保存失败，本机草稿已保留。请重试保存后再生成或送入剪辑。" />}
    <main className="agent-canvas-main aicut-creative-canvas__main">
      {loading ? <div className="aicut-creative-canvas__state"><Spin tip="正在打开创作画布…"><div /></Spin></div> : error ? <div className="aicut-creative-canvas__state"><Alert type="error" message={error} action={<Button onClick={() => setRetry(value => value + 1)}>重试</Button>} /></div> : canvas && <AgentCanvas
        ref={canvasUIRef} key={taskId} taskId={taskId} canvas={canvas} petServiceScope={petServiceScope} projectType="video"
        onChange={changeCanvas} onSelectionChange={ids => { selectedRef.current = ids; }}
        onViewportCommit={viewport => changeCanvas({ ...sessionRef.current.canvas, viewport })}
        onAppearanceChange={appearance => changeCanvas({ ...sessionRef.current.canvas, appearance })}
        onFlushCanvas={flushCanvas} onServerCanvasChange={refreshCanvas}
      />}
    </main>
    <Modal title="送入剪辑" open={transferOpen} onCancel={() => { if (!transferring) setTransferOpen(false); }} onOk={() => void transfer()} okText={`送入 ${assetIds.length} 个素材并进入剪辑`} cancelText="继续创作" confirmLoading={transferring} okButtonProps={{ disabled: !assetIds.length || assetLoading }} closable={!transferring} maskClosable={!transferring}>
      <p>选择要放入当前工程素材库的图片、视频或音频。最多 30 个；重复送入会复用已有素材。</p>
      {assetError && <Alert type="error" showIcon message={assetError} action={<Button size="small" disabled={transferring} onClick={() => void loadAssets()}>重新读取</Button>} />}
      {assetLoading ? <Spin /> : !assets.length && !assetError ? <Empty description="还没有可用素材。请先在画布上传，或在图片、视频、音频卡片中生成素材，再送入剪辑。" /> : <Checkbox.Group className="aicut-creative-canvas__assets" value={assetIds} disabled={transferring} onChange={ids => setAssetIds(creativeAssetSelection(ids, assets))}>{assets.map(asset => <Checkbox key={asset.id} value={asset.id} disabled={transferring || (!assetIds.includes(asset.id) && assetIds.length >= 30)}><span>{asset.name || asset.id}</span><Tag>{({ image: '图片', video: '视频', audio: '音频' })[asset.kind]}</Tag></Checkbox>)}</Checkbox.Group>}
    </Modal>
    <Modal title="发送剧本" open={scriptOpen} onCancel={() => { if (!scriptSending) setScriptOpen(false); }} onOk={() => void sendScript()} okText={`追加 ${scriptIds.length} 张卡片到剪辑剧本`} cancelText="继续创作" confirmLoading={scriptSending} okButtonProps={{ disabled: !scriptIds.length || scriptLoading }} closable={!scriptSending} maskClosable={!scriptSending}>
      <p>勾选要用于剪辑的故事、分镜或文本内容，按下方顺序追加到现有剧本。现有轨道保持不变。最多选择 10 张卡片。</p>
      {scriptError && <Alert type="error" message={scriptError} action={<Button size="small" disabled={scriptSending} onClick={() => void loadScriptNodes()}>重新读取</Button>} />}
      {scriptLoading ? <Spin /> : !scriptNodes.length && !scriptError ? <Empty description="还没有已填写的剧本或文本。请先双击卡片编辑内容，再发送剧本。" /> : <Checkbox.Group className="aicut-creative-canvas__assets" value={scriptIds} disabled={scriptSending} onChange={ids => setScriptIds([...new Set(ids)].slice(0, 10))}>{scriptNodes.map(node => <Checkbox key={node.id} value={node.id} disabled={scriptSending || (!scriptIds.includes(node.id) && scriptIds.length >= 10)}><span><strong>{node.title || '文本卡片'}</strong><br /><span className="aicut-creative-canvas__script-preview">{creativeScriptText(node).slice(0, 180)}</span></span></Checkbox>)}</Checkbox.Group>}
    </Modal>
  </section>;
}
