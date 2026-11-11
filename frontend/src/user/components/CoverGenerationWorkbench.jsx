import { Button, Dropdown, Empty, Input, InputNumber, Modal, Select, Spin, Upload, message } from 'antd';
import {
  ArrowLeft, CheckCircle2, Image as ImageIcon, Save, Sparkles,
  Upload as UploadIcon, Wand2, X
} from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import {
  createAgentImageGeneration,
  createAgentTask,
  getAgentAssetBlob,
  getAgentCanvas,
  getAgentGeneration,
  listAgentGenerations,
  saveAgentCanvas,
  uploadAgentAsset
} from '../../shared/api/agent.js';
import { designCoverPrompt, getCoverPromptConfig } from '../../shared/api/coverGeneration.js';
import { apiRequest } from '../../shared/api/client.js';
import { isAbortError } from '../pages/pluginGeneration.js';
import { createHydrationGuard } from '../pages/hydrationGuard.js';
import { listAvailableModels, modelSelectOptions } from '../../shared/api/modelCatalog.js';
import './cover-generation-workbench.css';

const { TextArea } = Input;
const PROJECT_KEY = 'qiantie:cover-generation:project';
const NODE_ID = 'cover-generation-image';
const TERMINAL_STATES = new Set(['awaiting_confirmation', 'completed', 'failed', 'cancelled']);
// 单张图最长等待 10 分钟：超时给出明确提示，避免无限转圈。
const POLL_INTERVAL_MS = 2500;
const POLL_MAX_ATTEMPTS = Math.ceil((10 * 60 * 1000) / POLL_INTERVAL_MS);
const DEFAULT_RESOLUTIONS = ['1K', '2K', '4K'];
// 8 个尺寸选项沿用洛水面板（bundle 457615-457624）；默认值改为 3:4，
// 因为新的「一键封面指令」面向番茄/七猫竖屏封面，洛水原默认 4:3 是横版。
const COVER_RATIOS = [
  { value: '4:3', label: '4:3 横版' },
  { value: '3:4', label: '3:4 竖版' },
  { value: '3:2', label: '3:2 横版' },
  { value: '2:3', label: '2:3 竖版' },
  { value: '7:10', label: '7:10 竖版' },
  { value: '9:16', label: '9:16 竖版' },
  { value: '16:9', label: '16:9 横版' },
  { value: '1:1', label: '1:1 方形' }
];
// V88 图片通道（batchfactoryv11.SupportedImageAspectRatios）不含 7:10，选中时明确提示，不静默替换。
const UNSUPPORTED_RATIOS = new Set(['7:10']);
const COVER_FONTS = [
  '微软雅黑', '宋体', '黑体', '楷体', '仿宋', '华文行楷', '华文新魏', '华文楷体', '华文宋体', '华文中宋',
  '方正舒体', '方正姚体', '隶书', '幼圆', '思源黑体', '思源宋体', '阿里巴巴普惠体', '站酷高端黑', '站酷快乐体', '优设标题黑'
].map(value => ({ value, label: value }));

function wait(ms) {
  return new Promise(resolve => window.setTimeout(resolve, ms));
}

function persistedCanvasNode() {
  return {
    id: NODE_ID,
    type: 'image',
    position: { x: 0, y: 0 },
    title: '封面生成',
    version: 1,
    sourceNodeIds: [],
    status: 'draft',
    summary: '',
    config: { prompt: '', modelId: '', ratio: '3:4', quality: 'standard', resolution: '1K', count: 1, referenceAssets: [] }
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

function downloadBlob(blob, name) {
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = name;
  anchor.click();
  URL.revokeObjectURL(url);
}

// 洛水桌面端用「原图 / 5M / 10M」保存；浏览器端按目标体积重编码，功能对齐。
async function fitBlobWithin(blob, maxBytes) {
  let bitmap;
  try { bitmap = await createImageBitmap(blob); } catch { return blob; }
  let width = bitmap.width;
  let height = bitmap.height;
  let quality = 0.92;
  for (let attempt = 0; attempt < 14; attempt += 1) {
    const canvas = document.createElement('canvas');
    canvas.width = Math.max(1, Math.round(width));
    canvas.height = Math.max(1, Math.round(height));
    canvas.getContext('2d').drawImage(bitmap, 0, 0, canvas.width, canvas.height);
    const encoded = await new Promise(resolve => canvas.toBlob(resolve, 'image/jpeg', quality));
    if (encoded && encoded.size <= maxBytes) return encoded;
    quality = Math.max(0.45, quality - 0.08);
    if (quality <= 0.45 && Math.min(width, height) > 900) {
      width *= 0.85;
      height *= 0.85;
    }
  }
  return blob;
}

export default function CoverGenerationWorkbench({ onBack }) {
  const [models, setModels] = useState([]);
  const [promptOptions, setPromptOptions] = useState([]);
  const [modelId, setModelId] = useState('');
  const [designPromptId, setDesignPromptId] = useState('');
  const [content, setContent] = useState('');
  const [dramaName, setDramaName] = useState('');
  const [fontStyle, setFontStyle] = useState('');
  // 「一键封面指令」面向番茄/七猫竖屏封面，默认 3:4；仍可在 8 档中切换。
  const [aspectRatio, setAspectRatio] = useState('3:4');
  const [resolution, setResolution] = useState('1K');
  const [count, setCount] = useState(1);
  const [references, setReferences] = useState([]);
  const [results, setResults] = useState([]);
  const [selectedId, setSelectedId] = useState('');
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const [running, setRunning] = useState(false);
  const [progress, setProgress] = useState({ current: 0, total: 0, failed: 0 });
  const [preview, setPreview] = useState(null);
  const [editOpen, setEditOpen] = useState(false);
  const [editInstruction, setEditInstruction] = useState('');
  const [editing, setEditing] = useState(false);
  const [contextMenu, setContextMenu] = useState(null);
  const projectRef = useRef({ taskId: '', nodeVersion: 1 });
  // 模型列表是异步加载的：用户已经选过模型后，回包不再覆盖选择。
  const modelGuardRef = useRef(null);
  if (!modelGuardRef.current) modelGuardRef.current = createHydrationGuard();
  // 生成/改图共用：离开面板时真正 abort，已完成的图片保留在 results。
  const abortRef = useRef(null);
  useEffect(() => () => {
    abortRef.current?.abort();
    abortRef.current = null;
  }, []);

  const selected = useMemo(() => results.find(item => item.id === selectedId) || null, [results, selectedId]);
  const resolutionOptions = useMemo(() => {
    const model = models.find(item => item.id === modelId);
    const supported = model?.capabilities?.resolutions;
    return Array.isArray(supported) && supported.length ? supported : DEFAULT_RESOLUTIONS;
  }, [models, modelId]);

  useEffect(() => {
    let active = true;
    (async () => {
      try {
        const [imageModels, prompts] = await Promise.all([listAvailableModels('image'), getCoverPromptConfig()]);
        if (!active) return;
        setModels(imageModels);
        setPromptOptions(prompts);
        if (modelGuardRef.current.canHydrate()) setModelId(imageModels[0]?.id || '');
      } catch (error) {
        if (active) message.error(error?.message || '封面生成初始化失败');
      } finally {
        if (active) setLoading(false);
      }
    })();
    return () => { active = false; };
  }, []);

  useEffect(() => {
    if (!resolutionOptions.includes(resolution)) setResolution(resolutionOptions[0] || '1K');
  }, [resolutionOptions, resolution]);

  useEffect(() => {
    if (!results.length) return;
    setSelectedId(current => (current && results.some(item => item.id === current) ? current : results[0].id));
  }, [results]);

  // 刷新后从任务历史回读已生成的封面，避免结果只活在内存里。
  useEffect(() => {
    let active = true;
    (async () => {
      const taskId = String(localStorage.getItem(PROJECT_KEY) || '').trim();
      if (!taskId) return;
      try {
        const data = await listAgentGenerations(taskId);
        if (!active) return;
        const seen = new Set();
        const restored = [];
        (Array.isArray(data?.tasks) ? data.tasks : []).forEach(item => {
          taskResultAssets(item).forEach(asset => {
            if (asset?.kind !== 'image' || !String(asset?.id || '').trim()) return;
            if (seen.has(asset.id)) return;
            seen.add(asset.id);
            restored.push({ ...asset, createdAt: asset.createdAt || item.createdAt || '' });
          });
        });
        if (restored.length) setResults(current => (current.length ? current : restored));
      } catch {
        // 历史回读失败不影响新生成，静默处理。
      }
    })();
    return () => { active = false; };
  }, []);

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
      if (!taskId || !canvas) throw new Error('未能创建封面生成项目');
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

  async function readDocument(file) {
    const lowerName = String(file?.name || '').toLowerCase();
    if (!lowerName.endsWith('.txt')) {
      message.warning('Word 文档读取尚未接入，当前请先上传 TXT');
      return Upload.LIST_IGNORE;
    }
    try {
      setContent(await file.text());
      message.success('文件读取成功');
    } catch {
      message.error('文件读取失败');
    }
    return Upload.LIST_IGNORE;
  }

  async function uploadReference(file) {
    if (!String(file?.type || '').startsWith('image/')) {
      message.warning('只能添加图片文件');
      return Upload.LIST_IGNORE;
    }
    if (references.length >= 20) {
      message.warning('参考图片最多 20 张');
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

  function beginRun() {
    if (abortRef.current) return null;
    abortRef.current = new AbortController();
    return abortRef.current;
  }

  function endRun(controller) {
    if (abortRef.current === controller) abortRef.current = null;
  }

  function scopedRequest(controller) {
    return (path, options = {}) => apiRequest(path, { ...options, signal: controller.signal });
  }

  async function pollGeneration(taskId, id, request) {
    for (let attempt = 0; attempt < POLL_MAX_ATTEMPTS; attempt += 1) {
      await wait(POLL_INTERVAL_MS);
      // 等待期间被取消：下一次请求会直接以 AbortError 结束，不再写结果。
      const task = (await getAgentGeneration(taskId, id, request)).task;
      if (TERMINAL_STATES.has(task?.state)) return task;
    }
    throw new Error('生成等待超时：任务仍在后台处理，请稍后重试或到 Agent 任务里查看');
  }

  function cancelRun() {
    const controller = abortRef.current;
    if (!controller) return;
    abortRef.current = null;
    // 真正中止当前所有在飞请求，已完成的结果保留。
    controller.abort();
    message.info('已取消本次生成，已完成的结果会保留');
  }

  async function generate() {
    if (running) return message.info('正在生成，请勿重复提交');
    if (!content.trim() && !dramaName.trim()) return message.warning('请输入原文内容或短剧名称');
    if (!modelId) return message.warning('请选择绘图模型');
    if (UNSUPPORTED_RATIOS.has(aspectRatio)) {
      return message.warning('V88 图片通道暂不支持 7:10，请选择 2:3 竖版');
    }
    const total = Math.min(10, Math.max(1, Number(count) || 1));
    const controller = beginRun();
    if (!controller) return message.info('正在生成，请勿重复提交');
    const request = scopedRequest(controller);
    setRunning(true);
    setProgress({ current: 0, total, failed: 0 });
    try {
      const project = projectRef.current.taskId ? projectRef.current : await ensureProject();
      // 先用 V88 文本模型执行「一键封面指令」得到最终图片提示词，再走图片通道。
      const prompt = await designCoverPrompt({
        content,
        dramaName,
        promptId: designPromptId,
        aspectRatio,
        fontStyle,
        referenceCount: references.length
      }, request);
      if (!prompt.trim()) throw new Error('封面提示词生成失败');
      const operationId = globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(36).slice(2)}`;
      const referenceAssetIds = references
        .filter(item => item?.kind === 'image' && String(item?.id || '').trim())
        .map(item => String(item.id));
      const requests = Array.from({ length: total }, (_, index) => ({
        nodeId: NODE_ID,
        nodeVersion: project.nodeVersion,
        modelId,
        mode: 'text_to_image',
        prompt,
        parameters: { resolution, ratio: aspectRatio, quality: 'standard', count: 1 },
        advanced: {},
        referenceAssetIds,
        idempotencyKey: `cover:${project.taskId}:${operationId}:${index + 1}`
      }));
      let completed = 0;
      let failed = 0;
      // 多张并发出图共用一个 controller：取消时一次性中止全部未完成的，已完成的保留。
      const settled = await Promise.allSettled(requests.map(async payload => {
        try {
          const response = await createAgentImageGeneration(project.taskId, payload, request);
          const queued = response?.generation;
          let assets = Array.isArray(response?.resultAssets) ? response.resultAssets : [];
          let finalTask = queued;
          if (!assets.length && queued?.id && !TERMINAL_STATES.has(queued.state)) {
            finalTask = await pollGeneration(project.taskId, queued.id, request);
            assets = taskResultAssets(finalTask);
          } else if (!assets.length) {
            assets = taskResultAssets(queued);
          }
          if (finalTask?.state === 'failed') throw new Error(finalTask.errorMessage || '图片生成失败');
          const items = assets
            .filter(asset => asset?.kind === 'image' && String(asset?.id || '').trim())
            .map(asset => ({ ...asset, createdAt: new Date().toISOString() }));
          if (items.length) setResults(current => [...current, ...items]);
          return items;
        } catch (error) {
          if (!isAbortError(error)) failed += 1;
          throw error;
        } finally {
          completed += 1;
          setProgress({ current: completed, total, failed });
        }
      }));
      if (controller.signal.aborted) return;
      const failures = settled.filter(item => item.status === 'rejected');
      const successes = settled.reduce((sum, item) => sum + (item.status === 'fulfilled' ? item.value.length : 0), 0);
      if (!successes && failures.length) throw failures[0].reason;
      if (failures.length) message.warning(`成功生成 ${successes} 张，失败 ${failures.length} 张`);
      else message.success(`封面生成成功，共 ${successes} 张`);
    } catch (error) {
      // 主动取消（离开面板）不弹错误提示；真实失败才报。
      if (!isAbortError(error) && !controller.signal.aborted) {
        message.error(error?.message || '封面生成失败');
      }
    } finally {
      endRun(controller);
      setRunning(false);
    }
  }

  async function saveResult(result, mode = 'original') {
    try {
      const blob = await getAgentAssetBlob(result.url);
      const index = Math.max(0, results.findIndex(item => item.id === result.id));
      const baseName = `${dramaName.trim() || '封面'}-${index + 1}`;
      if (mode === 'original') {
        downloadBlob(blob, result.name || `${baseName}.png`);
        message.success('封面保存成功');
        return;
      }
      const target = mode === '5m' ? 5 * 1024 * 1024 : 10 * 1024 * 1024;
      downloadBlob(await fitBlobWithin(blob, target), `${baseName}.jpg`);
      message.success('封面保存成功');
    } catch (error) {
      message.error(error?.message || '保存失败');
    }
  }

  function deleteResult(result) {
    setResults(current => current.filter(item => item.id !== result.id));
    setSelectedId(current => (current === result.id ? '' : current));
    message.success('已从列表移除（不会删除项目素材）');
  }

  async function startEdit() {
    if (!selected) return message.warning('没有可修改的封面');
    if (!editInstruction.trim()) return message.warning('请输入需要修改的内容');
    const controller = beginRun();
    if (!controller) return message.info('正在生成，请勿重复提交');
    const request = scopedRequest(controller);
    setEditing(true);
    try {
      const project = projectRef.current.taskId ? projectRef.current : await ensureProject();
      const operationId = globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(36).slice(2)}`;
      const response = await createAgentImageGeneration(project.taskId, {
        nodeId: NODE_ID,
        nodeVersion: project.nodeVersion,
        modelId,
        mode: 'image_to_image',
        prompt: editInstruction,
        parameters: { resolution, ratio: aspectRatio, quality: 'standard', count: 1 },
        advanced: {},
        referenceAssetIds: [selected.id],
        idempotencyKey: `cover-edit:${project.taskId}:${operationId}`
      }, request);
      const queued = response?.generation;
      let assets = Array.isArray(response?.resultAssets) ? response.resultAssets : [];
      let finalTask = queued;
      if (!assets.length && queued?.id && !TERMINAL_STATES.has(queued.state)) {
        finalTask = await pollGeneration(project.taskId, queued.id, request);
        assets = taskResultAssets(finalTask);
      } else if (!assets.length) {
        assets = taskResultAssets(queued);
      }
      if (finalTask?.state === 'failed') throw new Error(finalTask.errorMessage || '改图失败');
      const replacement = assets.find(asset => asset?.kind === 'image' && asset?.id);
      if (!replacement) throw new Error('改图成功但未返回图片');
      setResults(current => current.map(item => (item.id === selected.id ? { ...replacement, createdAt: new Date().toISOString() } : item)));
      setSelectedId(replacement.id);
      setEditOpen(false);
      setEditInstruction('');
      message.success('改图成功');
    } catch (error) {
      if (!isAbortError(error) && !controller.signal.aborted) message.error(error?.message || '改图失败');
    } finally {
      endRun(controller);
      setEditing(false);
    }
  }

  const saveMenuItems = [
    { key: 'original', label: '原图' },
    { key: '5m', label: '保存为 5M' },
    { key: '10m', label: '保存为 10M' }
  ];

  const contextMenuNode = contextMenu ? (
    <div className="cover-context-menu" style={{ left: contextMenu.x, top: contextMenu.y }} onMouseLeave={() => setContextMenu(null)}>
      <button type="button" onClick={() => { setContextMenu(null); saveResult(contextMenu.result, 'original'); }}>保存原图</button>
      <button type="button" onClick={() => { setContextMenu(null); saveResult(contextMenu.result, '5m'); }}>保存为 5M</button>
      <button type="button" onClick={() => { setContextMenu(null); saveResult(contextMenu.result, '10m'); }}>保存为 10M</button>
      <button type="button" className="is-danger" onClick={() => { setContextMenu(null); deleteResult(contextMenu.result); }}>从列表移除</button>
    </div>
  ) : null;

  return (
    <main className="cover-workbench">
      <header className="cover-page-head">
        <div className="cover-breadcrumb">
          <Button icon={<ArrowLeft size={16} />} onClick={onBack}>返回</Button>
          <span>插件市场</span><b>›</b><strong>封面生成</strong>
        </div>
        <Button type="text" aria-label="关闭封面生成" icon={<X size={18} />} onClick={onBack} />
      </header>
      <section className="cover-title">
        <span className="cover-title-icon"><ImageIcon size={22} /></span>
        <div>
          <div className="cover-title-line"><h1>封面生成</h1><span className="cover-ready"><CheckCircle2 size={14} /> 已授权</span></div>
          <p>上传参考图、选择封面设计、尺寸和模型，生成短剧封面</p>
        </div>
      </section>
      <Spin spinning={loading} tip="正在加载图片模型">
        <div className="cover-layout">
          <section className="cover-controls">
            <div className="cover-field-head">
              <span>原文内容</span>
              <Upload accept=".txt,.doc,.docx" showUploadList={false} beforeUpload={readDocument}>
                <Button size="small" icon={<UploadIcon size={14} />}>上传文档</Button>
              </Upload>
            </div>
            <TextArea
              rows={5}
              value={content}
              onChange={event => setContent(event.target.value)}
              placeholder="请输入原文内容，或上传txt/word文档..."
            />

            <div className="cover-field-head">
              <span>可选参考图片 {references.length ? <em>（{references.length}张）</em> : null}</span>
              {references.length ? <Button type="text" size="small" onClick={() => setReferences([])}>清空</Button> : null}
            </div>
            {references.length ? (
              <div className="cover-reference-list">
                {references.map(item => (
                  <div key={item.id}>
                    <img src={item.preview || item.url} alt={item.name || '参考图片'} title={item.name} />
                    <button type="button" aria-label="移除参考图片" onClick={() => removeReference(item.id)}><X size={12} /></button>
                  </div>
                ))}
              </div>
            ) : null}
            <Upload.Dragger accept="image/*" multiple showUploadList={false} beforeUpload={uploadReference} disabled={uploading || references.length >= 20}>
              <span className="cover-upload-icon"><ImageIcon size={20} color="#f43f5e" /></span>
              <strong>{uploading ? '正在上传...' : references.length ? '继续添加图片' : '点击上传参考图片'}</strong>
              <small>支持多张图片，可上传人物、风格、场景等</small>
            </Upload.Dragger>

            <label className="cover-field">
              <span>选择封面设计</span>
              <Select
                value={designPromptId || undefined}
                onChange={setDesignPromptId}
                allowClear
                placeholder="选择封面设计提示词"
                options={promptOptions.map(item => ({ value: item.id, label: item.name }))}
              />
            </label>

            <div className="cover-row">
              <label className="cover-field">
                <span>尺寸</span>
                <Select value={aspectRatio} onChange={setAspectRatio} options={COVER_RATIOS} />
              </label>
              <label className="cover-field">
                <span>分辨率</span>
                <Select value={resolution} onChange={setResolution} options={resolutionOptions.map(value => ({ value, label: value }))} />
              </label>
              <label className="cover-field">
                <span>并发批量</span>
                <InputNumber min={1} max={10} precision={0} value={count} onChange={value => setCount(Math.min(10, Math.max(1, Number(value) || 1)))} />
              </label>
            </div>
            <p className="cover-hint">生成数量支持 1-10 张，默认 1 张，会并发批量生成</p>
            {UNSUPPORTED_RATIOS.has(aspectRatio) ? <p className="cover-warning">V88 图片通道暂不支持 7:10，请选择 2:3 竖版</p> : null}

            <div className="cover-row">
              <label className="cover-field">
                <span>短剧名称（可选）</span>
                <Input value={dramaName} onChange={event => setDramaName(event.target.value)} placeholder="输入短剧名称" />
              </label>
              <label className="cover-field">
                <span>字体设计（可选）</span>
                <Select
                  value={fontStyle ? [fontStyle] : []}
                  onChange={values => setFontStyle(values[values.length - 1] || '')}
                  allowClear
                  showSearch
                  placeholder="选择或输入字体"
                  options={COVER_FONTS}
                  mode="tags"
                />
              </label>
            </div>

            <label className="cover-field">
              <span>绘图模型</span>
              {models.length ? (
                <Select value={modelId || undefined} onChange={value => { modelGuardRef.current.touch(); setModelId(value); }} options={modelSelectOptions(models, 'image')} placeholder="请选择图片模型" />
              ) : <Input disabled value="尚未添加可用的图片模型" />}
            </label>

            <Button
              className="cover-generate"
              size="large"
              loading={running}
              disabled={running || !models.length}
              icon={<Sparkles size={16} />}
              onClick={generate}
            >
              {running ? `生成中 (${progress.current}/${progress.total || count})` : '开始生成'}
            </Button>
            {running ? (
              <Button className="cover-cancel" danger onClick={cancelRun}>
                取消生成
              </Button>
            ) : null}
          </section>

          <section className="cover-results">
            <div className="cover-results-head">
              <div className="cover-results-title">
                <ImageIcon size={18} />
                <span>生成结果</span>
                {results.length ? <em>{results.length} 张</em> : null}
                {running ? <em>进度 {progress.current}/{progress.total || count}</em> : null}
                {progress.failed > 0 ? <em className="is-failed">失败 {progress.failed} 张</em> : null}
              </div>
              {selected ? (
                <div className="cover-result-actions">
                  <Button type="text" size="small" onClick={() => setPreview(selected)}>预览</Button>
                  <Dropdown menu={{ items: saveMenuItems, onClick: ({ key }) => saveResult(selected, key) }} trigger={['click']}>
                    <Button type="text" size="small" icon={<Save size={15} />}>保存</Button>
                  </Dropdown>
                  <Button type="text" size="small" icon={<Wand2 size={15} />} onClick={() => { setEditInstruction(''); setEditOpen(true); }}>改图</Button>
                </div>
              ) : null}
            </div>
            <div className="cover-result-body">
              {results.length ? (
                <div className="cover-result-grid">
                  {results.map((item, index) => (
                    <div
                      key={item.id}
                      className={`cover-result-card${item.id === selectedId ? ' is-selected' : ''}`}
                      onClick={() => setSelectedId(item.id)}
                      onDoubleClick={() => setPreview(item)}
                      onContextMenu={event => {
                        event.preventDefault();
                        setSelectedId(item.id);
                        setContextMenu({ x: event.clientX, y: event.clientY, result: item });
                      }}
                    >
                      <span className="cover-result-index">{index + 1}</span>
                      {item.id === selectedId ? <span className="cover-result-selected">已选中</span> : null}
                      <img src={item.url} alt={item.name || `封面 ${index + 1}`} />
                    </div>
                  ))}
                </div>
              ) : running ? (
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="正在并发生成封面，完成一张会立即显示..." />
              ) : (
                <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="输入内容后点击生成按钮" />
              )}
            </div>
          </section>
        </div>
      </Spin>

      <Modal title="封面预览" open={Boolean(preview)} footer={<Button onClick={() => setPreview(null)}>关闭</Button>} onCancel={() => setPreview(null)} width={860}>
        {preview ? <div className="cover-preview"><img src={preview.url} alt={preview.name || '封面预览'} /><div><Button icon={<Save size={15} />} onClick={() => saveResult(preview, 'original')}>保存原图</Button></div></div> : null}
      </Modal>

      <Modal
        title="改图"
        open={editOpen}
        onCancel={() => { if (!editing) setEditOpen(false); }}
        footer={(
          <div className="cover-edit-footer">
            <Button onClick={() => setEditOpen(false)} disabled={editing}>取消</Button>
            <Button type="primary" loading={editing} disabled={!editInstruction.trim()} onClick={startEdit}>{editing ? '改图中...' : '开始改图'}</Button>
          </div>
        )}
        width={520}
      >
        {selected ? (
          <div className="cover-edit-body">
            <span className="cover-field-label">当前封面</span>
            <div className="cover-edit-current"><img src={selected.url} alt="当前封面" /></div>
            <span className="cover-field-label">需要修改的内容</span>
            <TextArea
              rows={4}
              value={editInstruction}
              onChange={event => setEditInstruction(event.target.value)}
              placeholder="请描述需要修改的地方，例如：把背景改成蓝色、去掉右边的人物、添加一个金色标题..."
            />
          </div>
        ) : null}
      </Modal>

      {contextMenuNode}
      {contextMenu ? <div className="cover-context-mask" onClick={() => setContextMenu(null)} /> : null}
    </main>
  );
}
