import { Alert, Button, Empty, Input, Select, Switch, Tag, Upload, message } from 'antd';
import { ArrowLeft, Download, ImagePlus, Sparkles, X } from 'lucide-react';
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
import { listAvailableModels, modelSelectOptions } from '../../shared/api/modelCatalog.js';
import {
  assetReferenceErrorMessage,
  buildAssetReferenceRequests,
  normalizeAssetReferenceResults,
  normalizeAssetReferenceSettings,
  validateAssetReferenceInput
} from '../pages/asset-reference-workbench.js';
import { TURNAROUND_TEMPLATES, buildTurnaroundPrompt } from './characterTurnaroundPrompts.js';
import './asset-reference-workbench.css';
import './character-turnaround.css';

const { TextArea } = Input;
const NODE_ID = 'character-turnaround-image';
const PROJECT_KEY = 'qiantie:character-turnaround:project';
const TERMINAL_STATES = new Set(['awaiting_confirmation', 'completed', 'failed', 'cancelled']);
const RESOLUTIONS = ['1K', '2K', '4K'];

function wait(ms) {
  return new Promise(resolve => window.setTimeout(resolve, ms));
}

function persistedCanvasNode() {
  return {
    id: NODE_ID,
    type: 'image',
    position: { x: 0, y: 0 },
    title: '角色三视图',
    version: 1,
    sourceNodeIds: [],
    status: 'draft',
    summary: '',
    config: { prompt: '', modelId: '', ratio: '16:9', quality: 'standard', resolution: '2K', count: 1, referenceAssets: [] }
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

function escapeXml(value) {
  return String(value || '').replace(/[<>&"']/g, ch => ({ '<': '&lt;', '>': '&gt;', '&': '&amp;', '"': '&quot;', "'": '&apos;' }[ch]));
}

// 演练模式的占位图：真实产出格式（data URL），但不发任何生成请求。
function mockTurnaroundImage(name) {
  const label = escapeXml(name || '角色');
  const figure = (x, view) => `
    <g transform="translate(${x} 0)">
      <circle cx="0" cy="150" r="46" fill="#dfe7f2" stroke="#9fb3cc" stroke-width="2"/>
      <rect x="-52" y="205" width="104" height="200" rx="28" fill="#e8eef7" stroke="#9fb3cc" stroke-width="2"/>
      <rect x="-46" y="405" width="38" height="150" rx="16" fill="#dfe7f2" stroke="#9fb3cc" stroke-width="2"/>
      <rect x="8" y="405" width="38" height="150" rx="16" fill="#dfe7f2" stroke="#9fb3cc" stroke-width="2"/>
      <text x="0" y="600" text-anchor="middle" font-family="sans-serif" font-size="24" fill="#6b7f99">${view}</text>
    </g>`;
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="1280" height="720" viewBox="0 0 1280 720">
    <rect width="1280" height="720" fill="#ffffff"/>
    <rect x="30" y="30" width="1220" height="490" fill="none" stroke="#cbd6e5" stroke-width="2" stroke-dasharray="10 8"/>
    <text x="640" y="80" text-anchor="middle" font-family="sans-serif" font-size="30" fill="#33475f">${label} · 三视图（本地演练占位图）</text>
    ${figure(330, '正视图')}
    ${figure(640, '右视图')}
    ${figure(950, '后视图')}
    <text x="640" y="690" text-anchor="middle" font-family="sans-serif" font-size="22" fill="#8a9bb2">演练模式不调用模型，仅用于验证界面与生成编排</text>
  </svg>`;
  return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`;
}

export default function CharacterTurnaroundWorkbench({ onBack }) {
  const mockFromUrl = useMemo(() => new URLSearchParams(window.location.search).get('mock') === '1', []);
  const [mock, setMock] = useState(mockFromUrl);
  const [models, setModels] = useState([]);
  const [modelId, setModelId] = useState('');
  const [resolution, setResolution] = useState('2K');
  const [templateId, setTemplateId] = useState(TURNAROUND_TEMPLATES[0].id);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [references, setReferences] = useState([]);
  const [results, setResults] = useState([]);
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const [running, setRunning] = useState(false);
  const projectRef = useRef({ taskId: '', nodeVersion: 1 });
  const cancelledRef = useRef(false);
  const generationIdsRef = useRef([]);

  const template = TURNAROUND_TEMPLATES.find(item => item.id === templateId) || TURNAROUND_TEMPLATES[0];
  const prompt = buildTurnaroundPrompt(template, { name, description });
  const ratio = template.ratio;

  useEffect(() => {
    let active = true;
    async function initialize() {
      if (mock) {
        setModels([{ id: 'mock-local', name: '本地演练（不调用模型）' }]);
        setModelId('mock-local');
        setLoading(false);
        return;
      }
      try {
        const imageModels = await listAvailableModels('image');
        if (!active) return;
        setModels(imageModels);
        setModelId(imageModels[0]?.id || '');
        await ensureProject();
      } catch (error) {
        if (active) message.error(error?.message || '角色三视图工作台初始化失败');
      } finally {
        if (active) setLoading(false);
      }
    }
    initialize();
    return () => { active = false; cancelledRef.current = true; };
  }, [mock]);

  async function ensureProject() {
    let taskId = String(localStorage.getItem(PROJECT_KEY) || '').trim();
    let canvas = null;
    if (taskId) {
      try { canvas = (await getAgentCanvas(taskId)).canvas; } catch { taskId = ''; }
    }
    if (!taskId) {
      const created = await createAgentTask({ projectType: 'comic' });
      taskId = String(created?.task?.id || '').trim();
      canvas = created?.task?.canvas;
      if (!taskId || !canvas) throw new Error('未能创建角色三视图项目');
      localStorage.setItem(PROJECT_KEY, taskId);
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
    if (references.length >= 3) {
      message.warning('参考图最多 3 张');
      return Upload.LIST_IGNORE;
    }
    if (mock) {
      setReferences(current => [...current, { id: `mock-${Date.now()}-${current.length}`, name: file.name, preview: URL.createObjectURL(file) }]);
      message.success('已添加参考图（演练模式不上传服务器）');
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

  async function pollGeneration(taskId, id) {
    while (!cancelledRef.current) {
      await wait(2500);
      const task = (await getAgentGeneration(taskId, id)).task;
      if (TERMINAL_STATES.has(task?.state)) return task;
    }
    return null;
  }

  async function generate() {
    if (!description.trim() && !name.trim() && !references.length) {
      return message.warning('请填写角色名或角色设定，或添加参考图');
    }
    if (running) return message.info('正在生成，请勿重复提交');

    if (mock) {
      setRunning(true);
      await wait(900);
      const url = mockTurnaroundImage(name || description.slice(0, 8));
      setResults(current => [{
        id: `mock-${Date.now()}`,
        url,
        name: `${name || '角色'}-三视图-演练.svg`,
        prompt,
        modelId: 'mock-local',
        ratio,
        resolution,
        createdAt: new Date().toISOString()
      }, ...current]);
      setRunning(false);
      message.success('演练模式已返回占位图，未调用任何模型');
      return;
    }

    const validation = validateAssetReferenceInput({ prompt, referenceAssets: references, modelId });
    if (!validation.ok) return message.warning(validation.message || '请补充角色设定或参考图');
    setRunning(true);
    cancelledRef.current = false;
    try {
      const project = projectRef.current.taskId ? projectRef.current : await ensureProject();
      const operationId = globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(36).slice(2)}`;
      const settings = normalizeAssetReferenceSettings({ resolution, ratio, count: 1 });
      const [request] = buildAssetReferenceRequests({
        ...project, nodeId: NODE_ID, modelId, prompt, ...settings, referenceAssets: references, operationId
      });
      generationIdsRef.current = [];
      const response = await createAgentImageGeneration(project.taskId, request);
      const queued = response?.generation;
      if (queued?.id) generationIdsRef.current = [queued.id];
      let assets = Array.isArray(response?.resultAssets) ? response.resultAssets : [];
      let finalTask = queued;
      if (!assets.length && queued?.id && !TERMINAL_STATES.has(queued.state)) {
        finalTask = await pollGeneration(project.taskId, queued.id);
        assets = taskResultAssets(finalTask);
      } else if (!assets.length) {
        assets = taskResultAssets(queued);
      }
      if (cancelledRef.current) return;
      if (finalTask?.state === 'failed') throw new Error(finalTask.errorMessage || '图片生成失败');
      if (!assets.length) throw new Error('生成服务没有返回图片');
      const normalized = normalizeAssetReferenceResults(assets, { prompt, modelId, ...settings })
        .map(item => ({ ...item, createdAt: new Date().toISOString() }));
      setResults(current => [...normalized, ...current]);
      message.success('角色三视图已生成');
    } catch (error) {
      if (!cancelledRef.current) message.error(assetReferenceErrorMessage(error));
    } finally {
      setRunning(false);
      generationIdsRef.current = [];
    }
  }

  async function cancel() {
    cancelledRef.current = true;
    const { taskId } = projectRef.current;
    try {
      const ids = generationIdsRef.current;
      if (taskId && ids.length) await Promise.allSettled(ids.map(id => cancelAgentGeneration(taskId, id)));
      message.info('已取消本次生成');
    } finally {
      setRunning(false);
    }
  }

  async function downloadResult(result) {
    try {
      if (String(result.url || '').startsWith('data:')) {
        const anchor = document.createElement('a');
        anchor.href = result.url;
        anchor.download = result.name || 'character-turnaround.svg';
        anchor.click();
        return;
      }
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

  return (
    <main className="asset-reference-workbench">
      <header className="asset-reference-page-head">
        <div className="asset-reference-breadcrumb">
          <Button icon={<ArrowLeft size={16} />} onClick={onBack}>返回</Button>
          <span>插件市场</span><b>›</b><strong>角色三视图</strong>
        </div>
        <Button type="text" className="asset-reference-close" aria-label="关闭角色三视图" icon={<X size={18} />} onClick={onBack} />
      </header>

      <div className="asset-reference-title">
        <div className="asset-reference-title-icon"><Sparkles size={20} /></div>
        <div>
          <div className="asset-reference-title-line">
            <h1>角色三视图</h1>
            <Tag color="purple">图片通道</Tag>
            {mock ? <Tag color="orange">演练模式</Tag> : null}
          </div>
          <p>按角色设定生成正面 / 侧面 / 背面标准三视图，走 V88 自己的图片通道。</p>
        </div>
      </div>

      <div className="asset-reference-layout">
        <section className="asset-reference-controls">
          {mock ? (
            <Alert
              type="warning"
              showIcon
              message="演练模式：不会调用任何图片模型"
              description="点击生成只会返回本地占位图，用于验证界面与生成编排。关闭开关后才会走真实图片通道并消耗额度。"
            />
          ) : null}

          <label>
            <span>提示词模板</span>
            <Select
              value={templateId}
              onChange={setTemplateId}
              options={TURNAROUND_TEMPLATES.map(item => ({ value: item.id, label: item.name }))}
            />
          </label>

          <label>
            <span>角色名</span>
            <Input value={name} onChange={event => setName(event.target.value)} placeholder="例如：沈知意" />
          </label>

          <label>
            <span>角色设定</span>
            <TextArea
              rows={6}
              value={description}
              onChange={event => setDescription(event.target.value)}
              placeholder="外形、年龄、发型、服装、气质、标志性特征；写得越具体，三视图越稳定。"
            />
          </label>

          <label>
            <span>参考图（可选，最多 3 张）</span>
            <div className="asset-reference-reference-list">
              {references.map(item => (
                <div key={item.id}>
                  <img src={item.preview || item.url} alt={item.name || '参考图'} />
                  <button type="button" aria-label="移除参考图" onClick={() => removeReference(item.id)}>×</button>
                </div>
              ))}
              <div className="asset-reference-reference-add">
                <Upload accept="image/*" showUploadList={false} beforeUpload={uploadReference} disabled={uploading || references.length >= 3}>
                  <Button icon={<ImagePlus size={15} />} disabled={uploading || references.length >= 3}>{uploading ? '上传中…' : '添加参考图'}</Button>
                </Upload>
              </div>
            </div>
          </label>

          <label>
            <span>图片模型</span>
            <Select
              value={modelId}
              onChange={setModelId}
              disabled={mock}
              options={mock ? [{ value: 'mock-local', label: '本地演练（不调用模型）' }] : modelSelectOptions(models, 'image')}
              placeholder={loading ? '加载中…' : '请选择图片模型'}
            />
          </label>

          <label>
            <span>分辨率</span>
            <Select value={resolution} onChange={setResolution} options={RESOLUTIONS.map(item => ({ value: item, label: item }))} />
          </label>

          <label>
            <span>演练模式（不调用模型）</span>
            <Switch checked={mock} onChange={setMock} />
          </label>

          <label>
            <span>最终提示词（只读）</span>
            <div className="asset-reference-prompt">
              <TextArea readOnly value={prompt} />
            </div>
          </label>

          <div className="asset-reference-actions">
            <span>比例 {ratio} · 每次 1 张</span>
            {running && !mock ? <Button onClick={cancel}>取消</Button> : null}
            <Button type="primary" className="asset-reference-generate" icon={<Sparkles size={16} />} loading={running} onClick={generate}>
              {mock ? '生成占位图（不消耗额度）' : '生成三视图'}
            </Button>
          </div>
        </section>

        <section className="asset-reference-results">
          <div className="asset-reference-result-label">生成结果 <span>{results.length} 张</span></div>
          <div className="asset-reference-result-canvas">
            {results.length ? (
              <div className="asset-reference-waterfall">
                {results.map(result => (
                  <article key={result.id} className="asset-reference-result-card">
                    <img src={result.url} alt={`${name || '角色'} 三视图`} />
                    <div>
                      <small>{result.ratio} · {result.resolution}</small>
                      <Button size="small" type="text" icon={<Download size={14} />} onClick={() => downloadResult(result)}>下载</Button>
                    </div>
                  </article>
                ))}
              </div>
            ) : (
              <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="填写角色设定后点击生成；演练模式不会消耗额度" />
            )}
          </div>
        </section>
      </div>
    </main>
  );
}

