import { Button, Empty, Input, Modal, Select, Spin, Upload, message } from 'antd';
import { ArrowLeft, CheckCircle2, Download, ImagePlus, Library, Search, Settings2, Trash2, X } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import {
  cancelAgentGeneration,
  createAgentImageGeneration,
  createAgentTask,
  getAgentAssetBlob,
  getAgentCanvas,
  getAgentGeneration,
  saveAgentCanvas,
  uploadAgentAsset
} from '../../shared/api/agent.js';
import { apiRequest } from '../../shared/api/client.js';
import { getCurrentUsername } from '../../shared/api/auth.js';
import { listAvailableModels, modelSelectOptions } from '../../shared/api/modelCatalog.js';
import {
  assetReferenceErrorMessage,
  buildAssetReferenceRequests,
  normalizeAssetReferenceResults,
  normalizeAssetReferenceSettings,
  validateAssetReferenceInput
} from '../pages/asset-reference-workbench.js';
import './asset-reference-workbench.css';

const { TextArea } = Input;
// 本地缓存按登录账号隔离，避免换账号后还能看到上一个人的提示词与图片。
function storageScope() {
  return getCurrentUsername() || 'anonymous';
}
const projectKey = () => `qiantie:asset-reference:project:${storageScope()}`;
const resultsKey = () => `qiantie:asset-reference:results:${storageScope()}`;
const settingsKey = () => `qiantie:asset-reference:settings:${storageScope()}`;
const NODE_ID = 'asset-reference-image';
const TERMINAL_STATES = new Set(['awaiting_confirmation', 'completed', 'failed', 'cancelled']);
const RATIOS = ['16:9', '9:16', '21:9', '1:1', '3:4', '4:3'];

function readJSON(key, fallback) {
  try { return JSON.parse(localStorage.getItem(key) || '') || fallback; } catch { return fallback; }
}

function persistedCanvasNode() {
  return {
    id: NODE_ID,
    type: 'image',
    position: { x: 0, y: 0 },
    title: '资产参考',
    version: 1,
    sourceNodeIds: [],
    status: 'draft',
    summary: '',
    config: { prompt: '', modelId: '', ratio: '16:9', quality: 'standard', resolution: '1K', count: 1, referenceAssets: [] }
  };
}

function taskResultAssets(task) {
  return (Array.isArray(task?.results) ? task.results : []).map(result => result?.asset || {
    id: result?.assetId,
    kind: result?.kind,
    url: result?.asset?.url,
    name: result?.asset?.name
  });
}

function wait(ms) {
  return new Promise(resolve => window.setTimeout(resolve, ms));
}

export default function AssetReferenceWorkbench({ onBack }) {
  const savedSettings = useMemo(() => normalizeAssetReferenceSettings(readJSON(settingsKey(), { count: 5, resolution: '1K', ratio: '16:9' })), []);
  const [models, setModels] = useState([]);
  const [modelId, setModelId] = useState('');
  const [resolution, setResolution] = useState(savedSettings.resolution);
  const [ratio, setRatio] = useState(savedSettings.ratio);
  const [count, setCount] = useState(savedSettings.count);
  const [prompt, setPrompt] = useState('');
  const [references, setReferences] = useState([]);
  const [results, setResults] = useState(() => readJSON(resultsKey(), []));
  const [query, setQuery] = useState('');
  const [selectedResultIds, setSelectedResultIds] = useState([]);
  const [presetOpen, setPresetOpen] = useState(false);
  const [libraryOpen, setLibraryOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const [running, setRunning] = useState(false);
  const [cancelled, setCancelled] = useState(false);
  const [generationIds, setGenerationIds] = useState([]);
  const projectRef = useRef({ taskId: '', nodeVersion: 1 });
  const cancelledRef = useRef(false);
  // 取消要真的断请求，而不是只把结果丢掉。
  const abortRef = useRef(null);
  const generationIdsRef = useRef([]);

  useEffect(() => {
    let active = true;
    async function initialize() {
      try {
        const imageModels = await listAvailableModels('image');
        if (!active) return;
        setModels(imageModels);
        setModelId(imageModels[0]?.id || '');
        await ensureProject();
      } catch (error) {
        if (active) message.error(error?.message || '资产参考工作台初始化失败');
      } finally {
        if (active) setLoading(false);
      }
    }
    initialize();
    return () => {
      active = false;
      cancelledRef.current = true;
      abortRef.current?.abort();
      abortRef.current = null;
    };
  }, []);

  useEffect(() => {
    localStorage.setItem(settingsKey(), JSON.stringify({ resolution, ratio, count }));
  }, [resolution, ratio, count]);

  useEffect(() => {
    localStorage.setItem(resultsKey(), JSON.stringify(results.slice(0, 200)));
  }, [results]);

  async function ensureProject() {
    let taskId = String(localStorage.getItem(projectKey()) || '').trim();
    let canvas = null;
    if (taskId) {
      try { canvas = (await getAgentCanvas(taskId)).canvas; } catch { taskId = ''; }
    }
    if (!taskId) {
      const created = await createAgentTask({ projectType: 'comic' });
      taskId = String(created?.task?.id || '').trim();
      canvas = created?.task?.canvas;
      if (!taskId || !canvas) throw new Error('未能创建资产参考项目');
      localStorage.setItem(projectKey(), taskId);
    }
    let node = (canvas?.nodes || []).find(item => item.id === NODE_ID && item.type === 'image');
    if (!node) {
      const nextCanvas = { ...canvas, nodes: [...(canvas?.nodes || []), persistedCanvasNode()] };
      canvas = (await saveAgentCanvas(taskId, nextCanvas, Number(canvas?.revision) || 0)).canvas;
      node = (canvas?.nodes || []).find(item => item.id === NODE_ID);
    }
    projectRef.current = { taskId, nodeVersion: Number(node?.version) || 1 };
    return projectRef.current;
  }

  async function uploadReference(file) {
    if (!String(file?.type || '').startsWith('image/')) {
      message.warning('只能添加图片文件');
      return Upload.LIST_IGNORE;
    }
    if (references.length >= 10) {
      message.warning('参考图片最多 10 张');
      return Upload.LIST_IGNORE;
    }
    setUploading(true);
    try {
      const project = projectRef.current.taskId ? projectRef.current : await ensureProject();
      const response = await uploadAgentAsset(project.taskId, file);
      if (!response?.asset) throw new Error('上传结果缺少素材');
      setReferences(current => [...current, { ...response.asset, preview: URL.createObjectURL(file) }]);
      message.success('参考图片已添加');
    } catch (error) {
      message.error(error?.message || '参考图片上传失败');
    } finally {
      setUploading(false);
    }
    return Upload.LIST_IGNORE;
  }

  function removeReference(id) {
    setReferences(current => {
      const target = current.find(item => item.id === id);
      if (target?.preview) URL.revokeObjectURL(target.preview);
      return current.filter(item => item.id !== id);
    });
  }

  async function pollGeneration(taskId, id, request) {
    while (!cancelledRef.current) {
      await wait(2500);
      const task = (await getAgentGeneration(taskId, id, request)).task;
      if (TERMINAL_STATES.has(task?.state)) return task;
    }
    return null;
  }

  async function generate() {
    const validation = validateAssetReferenceInput({ prompt, referenceAssets: references, modelId });
    if (!validation.ok) return message.warning(validation.message || '请输入提示词或添加参考图片');
    if (running) return message.info('正在生成，请勿重复提交');
    setRunning(true);
    setCancelled(false);
    cancelledRef.current = false;
    const controller = new AbortController();
    abortRef.current = controller;
    const requestWithSignal = (path, options = {}) => apiRequest(path, { ...options, signal: controller.signal });
    try {
      const project = projectRef.current.taskId ? projectRef.current : await ensureProject();
      const operationId = globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(36).slice(2)}`;
      const settings = normalizeAssetReferenceSettings({ resolution, ratio, count });
      const requests = buildAssetReferenceRequests({
        ...project, nodeId: NODE_ID, modelId, prompt, ...settings, referenceAssets: references, operationId
      });
      generationIdsRef.current = [];
      setGenerationIds([]);
      const settled = await Promise.allSettled(requests.map(async payload => {
        if (cancelledRef.current) return [];
        const response = await createAgentImageGeneration(project.taskId, payload, requestWithSignal);
        const queued = response?.generation;
        if (queued?.id) {
          generationIdsRef.current = [...generationIdsRef.current, queued.id];
          setGenerationIds([...generationIdsRef.current]);
        }
        let assets = Array.isArray(response?.resultAssets) ? response.resultAssets : [];
        let finalTask = queued;
        if (!assets.length && queued?.id && !TERMINAL_STATES.has(queued.state)) {
          finalTask = await pollGeneration(project.taskId, queued.id, requestWithSignal);
          assets = taskResultAssets(finalTask);
        } else if (!assets.length) {
          assets = taskResultAssets(queued);
        }
        if (cancelledRef.current || finalTask?.state === 'cancelled') return [];
        if (finalTask?.state === 'failed') throw new Error(finalTask.errorMessage || '图片生成失败');
        return assets;
      }));
      if (cancelledRef.current || controller.signal.aborted) return;
      const assets = settled.flatMap(item => item.status === 'fulfilled' ? item.value : []);
      const failures = settled.filter(item => item.status === 'rejected');
      if (!assets.length && failures.length) throw failures[0].reason;
      const normalized = normalizeAssetReferenceResults(assets, { prompt, modelId, ...settings })
        .map(item => ({ ...item, createdAt: new Date().toISOString() }));
      if (!normalized.length) throw new Error('生成服务没有返回图片');
      setResults(current => [...normalized, ...current]);
      if (failures.length) message.warning(`成功生成 ${normalized.length} 张，失败 ${failures.length} 张`);
      else message.success(`成功生成 ${normalized.length} 张图片`);
    } catch (error) {
      if (!cancelledRef.current && !controller.signal.aborted) message.error(assetReferenceErrorMessage(error));
    } finally {
      if (abortRef.current === controller) abortRef.current = null;
      setRunning(false);
      generationIdsRef.current = [];
      setGenerationIds([]);
    }
  }

  async function cancel() {
    cancelledRef.current = true;
    setCancelled(true);
    // 先断掉正在飞的请求，再让服务端把排队中的任务取消。
    abortRef.current?.abort();
    abortRef.current = null;
    const { taskId } = projectRef.current;
    try {
      const ids = generationIdsRef.current.length ? generationIdsRef.current : generationIds;
      if (taskId && ids.length) {
        await Promise.allSettled(ids.map(id => cancelAgentGeneration(taskId, id)));
      }
      message.info('已取消本次操作');
    } catch (error) {
      message.warning(error?.message || '取消请求未确认，结果将不再写入列表');
    } finally {
      setRunning(false);
    }
  }

  async function downloadResult(result) {
    try {
      const blob = await getAgentAssetBlob(result.url);
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement('a');
      anchor.href = url;
      anchor.download = result.name || `${result.id}.png`;
      anchor.click();
      URL.revokeObjectURL(url);
    } catch (error) {
      message.error(error?.message || '图片下载失败');
    }
  }

  async function downloadSelected() {
    const selected = results.filter(item => selectedResultIds.includes(item.id));
    for (const result of selected) await downloadResult(result);
  }

  function removeSelected() {
    setResults(current => current.filter(item => !selectedResultIds.includes(item.id)));
    setSelectedResultIds([]);
    message.success('已从列表移除（不会删除项目素材）');
  }

  function toggleResult(id) {
    setSelectedResultIds(current => current.includes(id) ? current.filter(value => value !== id) : [...current, id]);
  }

  const visibleResults = useMemo(() => {
    const value = query.trim().toLocaleLowerCase('zh-CN');
    if (!value) return results;
    return results.filter(item => `${item.prompt} ${item.modelId} ${item.ratio} ${item.resolution}`.toLocaleLowerCase('zh-CN').includes(value));
  }, [query, results]);

  const visibleIds = visibleResults.map(item => item.id);
  const allVisibleSelected = visibleIds.length > 0 && visibleIds.every(id => selectedResultIds.includes(id));

  function toggleAllVisible() {
    setSelectedResultIds(current => {
      if (allVisibleSelected) return current.filter(id => !visibleIds.includes(id));
      return [...new Set([...current, ...visibleIds])];
    });
  }

  return (
    <main className="asset-reference-workbench">
      <header className="asset-reference-page-head">
        <div className="asset-reference-breadcrumb">
          <Button icon={<ArrowLeft size={16} />} onClick={onBack}>返回</Button>
          <span>插件市场</span><b>›</b><strong>资产参考</strong>
        </div>
        <Button type="text" className="asset-reference-close" aria-label="关闭资产参考" icon={<X size={18} />} onClick={onBack} />
      </header>
      <section className="asset-reference-title">
        <span className="asset-reference-title-icon"><ImagePlus size={22} /></span>
        <div><div className="asset-reference-title-line"><h1>资产参考</h1><span className="asset-reference-ready"><CheckCircle2 size={14} /> 已授权</span></div><p>支持文生图和多图参考，可拖拽或 Ctrl+V 粘贴添加参考图</p></div>
      </section>
      <Spin spinning={loading} tip="正在加载图片模型">
        <div className="asset-reference-layout">
          <section className="asset-reference-results">
            <div className="asset-reference-result-toolbar">
              <Input allowClear value={query} onChange={event => setQuery(event.target.value)} prefix={<Search size={16} />} placeholder="超级搜索：提示词、模型、尺寸、时间..." />
              <Button type="text" disabled={!visibleResults.length} onClick={toggleAllVisible}>{allVisibleSelected ? '取消全选' : '全选'}</Button>
              <Button danger disabled={!results.length} onClick={() => { setResults([]); setSelectedResultIds([]); message.success('已全部从列表移除（不会删除项目素材）'); }}>全部从列表移除</Button>
              <Button type="text" disabled={!selectedResultIds.length} onClick={downloadSelected} icon={<Download size={15} />}>下载</Button>
              <Button type="text" disabled={!selectedResultIds.length} onClick={removeSelected} icon={<Trash2 size={15} />}>批量从列表移除</Button>
            </div>
            <div className="asset-reference-result-label">生成结果 <span>{visibleResults.length}/{results.length}</span></div>
            <div className="asset-reference-result-canvas">
            {visibleResults.length ? <div className="asset-reference-waterfall">{visibleResults.map(item => (
              <article key={`${item.id}-${item.createdAt || ''}`} className={`asset-reference-result-card${selectedResultIds.includes(item.id) ? ' is-selected' : ''}`} onClick={() => toggleResult(item.id)}>
                <img src={item.url} alt={item.prompt || item.name || '生成图片'} />
                <div><strong>{item.prompt || '基于参考图生成'}</strong><small>{item.modelId} · {item.resolution} · {item.ratio}</small></div>
                <Button aria-label="下载图片" icon={<Download size={15} />} onClick={event => { event.stopPropagation(); downloadResult(item); }} />
              </article>
            ))}</div> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="在右侧输入提示词或添加参考图，生成记录会永久保存在这里" />}
            </div>
          </section>
          <aside className="asset-reference-controls">
            <label>模型<Select value={modelId || undefined} onChange={setModelId} options={modelSelectOptions(models, 'image')} placeholder="请选择图片模型" /></label>
            <fieldset><legend>分辨率</legend><div className="asset-reference-resolutions">{['1K', '2K', '4K'].map(value => <button key={value} type="button" className={resolution === value ? 'is-active' : ''} onClick={() => setResolution(value)}>{value}</button>)}</div></fieldset>
            <fieldset><legend>图片尺寸</legend><div className="asset-reference-ratios">{RATIOS.map(value => <button key={value} type="button" className={ratio === value ? 'is-active' : ''} onClick={() => setRatio(value)}><span />{value}</button>)}</div></fieldset>
            <label>数量<Select value={count} onChange={setCount} options={Array.from({ length: 20 }, (_, index) => index + 1).map(value => ({ value, label: String(value) }))} /></label>
            <div className="asset-reference-upload-head"><strong>参考图片</strong><div><Button type="text" size="small" icon={<Settings2 size={14} />} onClick={() => setPresetOpen(true)}>预设</Button><Button type="text" size="small" icon={<Library size={14} />} onClick={() => setLibraryOpen(true)}>参考库</Button></div></div>
            <Upload.Dragger accept="image/*" multiple showUploadList={false} beforeUpload={uploadReference} disabled={uploading || references.length >= 10}>
              <span className="asset-reference-upload-plus">+</span><strong>{uploading ? '正在上传...' : '添加参考图片'}</strong><small>点击上传、或拖拽、Ctrl+V 粘贴到此窗口</small>
            </Upload.Dragger>
            {references.length ? <div className="asset-reference-reference-list">{references.map(item => <div key={item.id}><img src={item.preview || item.url} alt={item.name || '参考图片'} /><button type="button" aria-label="移除参考图片" onClick={() => removeReference(item.id)}><X size={14} /></button></div>)}</div> : null}
            <label className="asset-reference-prompt">提示词<TextArea rows={6} value={prompt} onChange={event => setPrompt(event.target.value)} placeholder="描述你想生成的画面；输入 @ 可引用参考库卡片..." /></label>
            <div className="asset-reference-actions">
              {running ? <Button danger onClick={cancel}>取消</Button> : null}
              <Button className="asset-reference-generate" size="large" loading={running} disabled={!modelId} onClick={generate}>{running ? 'AI生图中...' : 'AI生图'}</Button>
            </div>
            {cancelled ? <p className="asset-reference-cancelled">已取消本次操作</p> : null}
          </aside>
        </div>
      </Spin>
      <Modal title="参数预设" open={presetOpen} onCancel={() => setPresetOpen(false)} footer={<Button type="primary" onClick={() => { localStorage.setItem(settingsKey(), JSON.stringify({ resolution, ratio, count })); setPresetOpen(false); message.success('当前参数已保存为本机预设'); }}>保存当前参数</Button>}>
        <p>保存当前分辨率、图片尺寸和数量，下次进入资产参考时自动恢复。</p>
        <p>{resolution} · {ratio} · {count} 张</p>
      </Modal>
      <Modal title="参考库" open={libraryOpen} onCancel={() => setLibraryOpen(false)} footer={<Button onClick={() => setLibraryOpen(false)}>关闭</Button>}>
        {references.length ? <div className="asset-reference-library">{references.map(item => <img key={item.id} src={item.preview || item.url} alt={item.name || '参考图片'} />)}</div> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有参考图片，请先上传" />}
      </Modal>
    </main>
  );
}
