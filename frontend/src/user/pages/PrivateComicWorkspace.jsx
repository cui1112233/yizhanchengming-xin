import { useEffect, useRef, useState } from 'react';
import { Alert, Button, Dropdown, Image, Input, Modal, Popconfirm, Progress, Select, Space, Switch, Tooltip, Typography, message } from 'antd';
import { Download, ExternalLink, ImageIcon, MoreHorizontal, Pencil, RefreshCw, Settings2, WandSparkles } from 'lucide-react';

import { getComicPageImageBlob } from '../../shared/api/batchFactory';
import { IMAGE_ASPECT_RATIO_OPTIONS } from '../../shared/imageAspectRatios.js';
import { BatchFactoryComicStylePicker } from './shuihuo/BatchFactoryComicStylePicker';
import { nextComicOutputVisibleCount, resolveComicImageSources } from './shuihuo/batchFactoryComicStatus.js';
import { comicEstimatedPageRange } from './shuihuo/batchFactoryContentRange.js';
import { loadBookAssetDisplayURL } from './shuihuo/batchFactoryNovelListHelpers';
import { PrivateStudioAssetEditor } from './private-studio/PrivateStudioAssetEditor';
import { privateComicGenerationStatus } from './private-studio/privateComicGenerationStatus.js';
import { privateComicAssetUseThreeView } from './private-studio/privateComicThreeView.js';
import { privateComicPrimaryAction } from './private-studio/privateComicPrimaryAction.js';
import { PrivateStudioChips } from './private-studio/PrivateStudioChips';
import { PRIVATE_STUDIO_CONSTRAINT_FIELDS, PrivateStudioConstraintSection, PrivateStudioConstraintsModal } from './private-studio/PrivateStudioConstraintsModal';
import { PrivateStudioHeaderActions } from './private-studio/PrivateStudioHeaderActions';
import { PrivateStudioNote } from './private-studio/PrivateStudioNote';
import { privateComicMinSourceLength, privateComicTitle } from './privateCustomizationComic';
import { canExtractPrivateComicAssets } from './privateComicSingleBookFlow.js';

const COMIC_DIGITS = ['零', '一', '二', '三', '四', '五', '六', '七', '八', '九'];

function text(value) {
  return String(value ?? '').trim();
}

function formatDraftSavedTime(value) {
  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) return '';
  return `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`;
}

function comicPageNumber(value) {
  const number = Number(value);
  if (!Number.isFinite(number) || number <= 0) return String(value);
  if (number < 10) return COMIC_DIGITS[number];
  if (number === 10) return '十';
  if (number < 20) return `十${COMIC_DIGITS[number % 10]}`;
  if (number < 100) return `${COMIC_DIGITS[Math.floor(number / 10)]}十${number % 10 ? COMIC_DIGITS[number % 10] : ''}`;
  return String(number);
}

function presetSelection(item) {
  if (!item) return {};
  return {
    presetId: text(item.id),
    presetName: text(item.name),
    presetSlot: text(item.slot),
    presetVersion: item.version ?? null
  };
}

function optionsFor(catalog, slot) {
  return (Array.isArray(catalog) ? catalog : [])
    .filter(item => item?.slot === slot)
    .map(item => ({ value: item.id, label: `${item.name || item.id}${item.version ? ` · v${item.version}` : ''}`, item }));
}

function constraintOptions(catalog, category) {
  return (Array.isArray(catalog) ? catalog : [])
    .filter(item => item?.kind === 'addon' && item?.constraintCategory === category)
    .map(item => ({ value: item.id, label: `${item.name || item.id}${item.version ? ` · v${item.version}` : ''}`, item }));
}

function selectedConstraint(comic, category) {
  return (comic?.constraints || []).find(item => item?.constraintCategory === category)?.presetId || '';
}

function taskLabel(task) {
  if (!task) return '尚未提交';
  if (task.status === 'succeeded' || task.status === 'completed') return '漫画生成完成';
  if (task.status === 'failed') return '漫画生成失败';
  if (task.status === 'created') return '已创建，准备启动';
  if (task.status === 'idle') return '已就绪，可手动执行或重试';
  return '漫画生成中';
}

function PrivateComicAssetImage({ imageUrl, alt = '', preview = false }) {
  const [display, setDisplay] = useState({ url: '', failed: false });
  useEffect(() => {
    let active = true;
    let ownedUrl = '';
    setDisplay({ url: '', failed: false });
    loadBookAssetDisplayURL(imageUrl).then(result => {
      if (!active) {
        if (result.owned) URL.revokeObjectURL(result.url);
        return;
      }
      if (result.owned) ownedUrl = result.url;
      setDisplay({ url: result.url, failed: false });
    }).catch(() => { if (active) setDisplay({ url: '', failed: true }); });
    return () => {
      active = false;
      if (ownedUrl) URL.revokeObjectURL(ownedUrl);
    };
  }, [imageUrl]);

  if (!display.url) return <span className="batch-factory-asset-image-state">{display.failed ? '图片加载失败' : '正在读取图片…'}</span>;
  return preview
    ? <Image src={display.url} alt={alt} preview={{ mask: '点击放大' }} />
    : <img src={display.url} alt={alt} />;
}

function ComicPageCard({ page, index, taskReady, action, actionKey, onGeneratePage, onSavePagePrompt }) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState({ comicPrompt: '', imagePrompt: '' });
  const [display, setDisplay] = useState({ url: '', failed: false });
  const [original, setOriginal] = useState({ url: '', loading: false, failed: false });
  const [viewerOpen, setViewerOpen] = useState(false);
  const originalURLRef = useRef('');
  const images = Array.isArray(page?.images) ? page.images : [];
  const primary = images.find(image => image?.isPrimary) || images[0] || null;
  // 卡片只取 420px 缩略图；原图留到点击放大或下载本页时才拉取。
  const sources = resolveComicImageSources(primary);
  const generating = action === 'pages' && actionKey === page?.videoId;
  const saving = action === 'save' && actionKey === page?.videoId;
  const busyElsewhere = Boolean(action) && !generating && !saving;
  const hasPrompt = Boolean(text(page?.comicPrompt) || text(page?.imagePrompt));
  const canSave = Boolean(text(draft.comicPrompt) && text(draft.imagePrompt));
  // 页状态只读后端派生字段（promptStatus / imageStatus / generationState），前端不再自行推导一套。
  const backendPromptStatus = text(page?.promptStatus);
  const backendImageStatus = text(page?.imageStatus);
  const backendGenerationState = text(page?.generationState);
  const hasBackendPageStatus = Boolean(backendPromptStatus || backendImageStatus || backendGenerationState);
  const pageStatusTag = !hasBackendPageStatus
    ? (images.length ? '已生成' : hasPrompt ? '待生成' : '待确认')
    : backendGenerationState === 'generating'
      ? '生成中'
      : backendGenerationState === 'failed'
        ? '生成失败'
        : backendImageStatus === 'primary'
          ? '已生成'
          : backendGenerationState === 'succeeded'
            ? '已生成'
            : backendPromptStatus === 'ready'
              ? '待生成'
              : '待确认';

  useEffect(() => {
    let active = true;
    const owned = [];
    setDisplay({ url: '', failed: false });
    if (!sources.thumbnail) return () => {};
    const load = target => getComicPageImageBlob(target).then(blob => {
      if (!active) return;
      const url = URL.createObjectURL(blob);
      owned.push(url);
      setDisplay({ url, failed: false });
    });
    load(sources.thumbnail).catch(() => {
      if (!active) return;
      // 缩略图生成失败时自动回落原图，不让卡片停在占位文案。
      if (sources.original && sources.original !== sources.thumbnail) {
        load(sources.original).catch(() => { if (active) setDisplay({ url: '', failed: true }); });
        return;
      }
      setDisplay({ url: '', failed: true });
    });
    return () => {
      active = false;
      owned.forEach(url => URL.revokeObjectURL(url));
    };
  }, [sources.thumbnail, sources.original]);

  useEffect(() => () => {
    if (originalURLRef.current) URL.revokeObjectURL(originalURLRef.current);
  }, []);

  async function ensureOriginal() {
    if (originalURLRef.current) return originalURLRef.current;
    if (!sources.original) return '';
    setOriginal(current => ({ ...current, loading: true, failed: false }));
    try {
      const blob = await getComicPageImageBlob(sources.original);
      const url = URL.createObjectURL(blob);
      originalURLRef.current = url;
      setOriginal({ url, loading: false, failed: false });
      return url;
    } catch {
      setOriginal({ url: '', loading: false, failed: true });
      return '';
    }
  }

  async function openViewer() {
    if (!sources.original) return;
    setViewerOpen(true);
    await ensureOriginal();
  }

  async function downloadPage() {
    const url = await ensureOriginal();
    if (!url) {
      message.error('原图读取失败，请稍后重试');
      return;
    }
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = `漫画${comicPageNumber(index + 1)}.png`;
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
  }

  function beginEdit() {
    setDraft({ comicPrompt: text(page?.comicPrompt), imagePrompt: text(page?.imagePrompt) });
    setEditing(true);
  }

  async function save() {
    const saved = await onSavePagePrompt?.(page, draft);
    if (saved) setEditing(false);
  }

  return <div className="shot-output-card">
    <div className="shot-output-card-header">
      <span className="private-comic-page-title">漫画{comicPageNumber(index + 1)} · 第 {index + 1} 页</span>
      <span className="private-studio-shot-status">{pageStatusTag}</span>
      <Space size={8} wrap>
        <Button size="small" icon={<Pencil size={15} aria-hidden="true" />} onClick={editing ? () => setEditing(false) : beginEdit}>{editing ? '取消编辑' : '编辑提示词'}</Button>
        <Button
          size="small"
          type="primary"
          icon={<WandSparkles size={15} aria-hidden="true" />}
          loading={generating}
          disabled={!taskReady || !hasPrompt || busyElsewhere || generating}
          title={!taskReady ? '请先提取资产创建任务' : !hasPrompt ? '该页缺少提示词' : busyElsewhere ? '当前有其他操作进行中' : ''}
          onClick={() => onGeneratePage?.(page)}
        >{images.length ? '重新生成图' : '生成图'}</Button>
        {sources.original ? <Button size="small" icon={<Download size={15} aria-hidden="true" />} loading={original.loading} onClick={() => void downloadPage()}>下载本页</Button> : null}
      </Space>
    </div>
    {editing ? <div className="private-comic-page-editor">
      <label>
        <span>漫画提示词</span>
        <Input.TextArea rows={4} value={draft.comicPrompt} placeholder="漫画提示词" onChange={event => setDraft(current => ({ ...current, comicPrompt: event.target.value }))} />
      </label>
      <label>
        <span>图片提示词</span>
        <Input.TextArea rows={4} value={draft.imagePrompt} placeholder="图片提示词" onChange={event => setDraft(current => ({ ...current, imagePrompt: event.target.value }))} />
      </label>
      <Space>
        <Button size="small" type="primary" loading={saving} disabled={!canSave || busyElsewhere} title={!canSave ? '两项提示词都需要填写后才能保存' : ''} onClick={() => void save()}>保存提示词</Button>
        <Button size="small" onClick={() => setEditing(false)}>取消</Button>
      </Space>
      {!canSave ? <small className="entity-media-status">保存需要漫画提示词与图片提示词都不为空。</small> : null}
    </div> : <pre className="shot-output-card-content">{`漫画提示词\n${text(page?.comicPrompt) || '暂无'}\n\n图片提示词\n${text(page?.imagePrompt) || '暂无'}`}</pre>}
    <div className="private-comic-page-image">
      {display.url ? <img src={display.url} alt={`漫画第${index + 1}页`} title="点击查看原图" style={{ cursor: 'zoom-in' }} onClick={() => void openViewer()} /> : <div className="private-comic-page-placeholder">
        <span className="entity-media-status">{display.failed ? '图片加载失败' : images.length ? '正在读取图片…' : '本页还没有生成图片'}</span>
        {!images.length ? <Button
          size="small"
          type="primary"
          icon={<WandSparkles size={15} aria-hidden="true" />}
          loading={generating}
          disabled={!taskReady || !hasPrompt || busyElsewhere || generating}
          onClick={() => onGeneratePage?.(page)}
        >生成图</Button> : null}
      </div>}
    </div>
    <Modal open={viewerOpen} footer={null} centered width="90vw" destroyOnClose onCancel={() => setViewerOpen(false)} title={`漫画${comicPageNumber(index + 1)} · 原图`}>
      {original.url ? <img src={original.url} alt={`漫画第${index + 1}页原图`} style={{ width: '100%', height: 'auto' }} /> : <div className="entity-media-status">{original.failed ? '原图加载失败' : '正在读取原图…'}</div>}
    </Modal>
  </div>;
}

function ComicOutputCards({ pages, taskReady, action, actionKey, onGeneratePage, onSavePagePrompt, onDownloadAll }) {
  const pageList = Array.isArray(pages) ? pages : [];
  const imagedPages = pageList.filter(page => (Array.isArray(page?.images) ? page.images : []).length > 0).length;
  // 首屏只挂 COMIC_OUTPUT_FIRST_SCREEN 张卡片，滚到尾部哨兵再按批追加；
  // 35 页一次性挂载会同时发起 35 个缩略图请求并解码，仍然会卡。
  const [visibleCount, setVisibleCount] = useState(() => nextComicOutputVisibleCount(0, pageList.length));
  const sentinelRef = useRef(null);

  useEffect(() => {
    setVisibleCount(nextComicOutputVisibleCount(0, pageList.length));
  }, [pageList.length]);

  useEffect(() => {
    if (visibleCount >= pageList.length) return undefined;
    const node = sentinelRef.current;
    if (!node || typeof IntersectionObserver === 'undefined') {
      setVisibleCount(pageList.length);
      return undefined;
    }
    const observer = new IntersectionObserver(entries => {
      if (entries.some(entry => entry.isIntersecting)) {
        setVisibleCount(current => nextComicOutputVisibleCount(current, pageList.length));
      }
    }, { rootMargin: '320px' });
    observer.observe(node);
    return () => observer.disconnect();
  }, [visibleCount, pageList.length]);

  const visiblePages = pageList.slice(0, visibleCount);
  return <div className="shot-output-cards">
    <div className="shot-output-toolbar">
      <span>共 {pageList.length} 页 · 已出图 {imagedPages} 页</span>
      <Space size={8}>
        <Button size="small" icon={<Download size={15} aria-hidden="true" />} loading={action === 'download'} disabled={!pageList.length || (action && action !== 'download')} onClick={onDownloadAll}>下载全部</Button>
      </Space>
    </div>
    {visiblePages.map((page, index) => <ComicPageCard
      key={page?.videoId || page?.id || index}
      page={page}
      index={index}
      taskReady={taskReady}
      action={action}
      actionKey={actionKey}
      onGeneratePage={onGeneratePage}
      onSavePagePrompt={onSavePagePrompt}
    />)}
    {visibleCount < pageList.length ? <div ref={sentinelRef} className="shot-output-cards-more" aria-hidden="true" /> : null}
  </div>;
}

export function PrivateComicWorkspace({
  catalog = [],
  modelOptions = {},
  draft = {},
  draftSaveState = { status: 'idle', savedAt: null, reason: '' },
  comicAssets = [],
  comicAssetImages = {},
  comicAssetImageFailures = [],
  comicPages = [],
  readiness = null,
  action = '',
  actionKey = '',
  onExtractAssets,
  onGenerateAssetImage,
  onSaveAsset,
  onUploadAssetImage,
  onSetPrimaryAssetImage,
  onGenerateMissingAssetImages,
  onGeneratePrompts,
  onGeneratePageImage,
  onGenerateAllPageImages,
  onSavePagePrompt,
  onDownloadAll,
  onDraftChange,
  task,
  busy = false,
  error = '',
  onRefresh,
  onOpenBatchFactory,
  onOpenUnifiedConfig,
  onReloadCatalog,
  onRetry,
  onPendingReferenceChange,
  onPersistSettings,
  customizationWorkspace = 'comic',
  onWorkspaceChange
}) {
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [savingSettings, setSavingSettings] = useState(false);
  const [leftPanelWidth, setLeftPanelWidth] = useState(null);
  const [assetPreview, setAssetPreview] = useState(null);
  const [editingAsset, setEditingAsset] = useState(null);
  const [assetSaving, setAssetSaving] = useState(false);
  const [assetImageThreeView, setAssetImageThreeView] = useState(true);
  const workbenchRef = useRef(null);
  const comic = draft.comic && typeof draft.comic === 'object' ? draft.comic : {};

  async function saveEditingAsset() {
    if (!text(editingAsset?.name) || !text(editingAsset?.prompt)) {
      message.warning('请填写资产名称和提示词。');
      return;
    }
    setAssetSaving(true);
    try {
      if (await onSaveAsset?.(assetList.find(asset => asset.id === editingAsset.id) || editingAsset, editingAsset)) setEditingAsset(null);
    } finally {
      setAssetSaving(false);
    }
  }

  async function uploadEditingAssetImage(event) {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file || !editingAsset) return;
    setAssetSaving(true);
    try {
      await onUploadAssetImage?.(editingAsset, file);
    } finally {
      setAssetSaving(false);
    }
  }

  async function makeEditingAssetImagePrimary(image) {
    if (!editingAsset || !image || image.isPrimary) return;
    setAssetSaving(true);
    try {
      await onSetPrimaryAssetImage?.(editingAsset, image);
    } finally {
      setAssetSaving(false);
    }
  }

  function handleResizeStart(event) {
    if (event.button !== 0 || !workbenchRef.current) return;
    event.preventDefault();
    const workbench = workbenchRef.current;
    function resize(moveEvent) {
      const rect = workbench.getBoundingClientRect();
      const nextWidth = Math.min(rect.width - 460 - 8, Math.max(320, moveEvent.clientX - rect.left));
      setLeftPanelWidth(nextWidth);
    }
    function finishResize() {
      window.removeEventListener('pointermove', resize);
      window.removeEventListener('pointerup', finishResize);
      document.body.classList.remove('is-resizing');
    }
    document.body.classList.add('is-resizing');
    window.addEventListener('pointermove', resize);
    window.addEventListener('pointerup', finishResize, { once: true });
  }
  const patchDraft = patch => onDraftChange?.({ ...draft, ...patch });
  const patchComic = patch => patchDraft({ comic: { ...comic, ...patch, enabled: true } });
  // 约束弹窗点「完成」时，已有服务端任务的书必须先真正写回设置并回读通过再关闭；
  // 还没有任务时只保存本地草稿，首次提取资产创建任务时一起提交。
  async function saveSettings() {
    if (savingSettings) return;
    if (!task?.batchId) {
      setSettingsOpen(false);
      return;
    }
    setSavingSettings(true);
    try {
      await onPersistSettings?.();
      setSettingsOpen(false);
    } catch (error) {
      message.error(error?.message || '漫画约束保存失败');
    } finally {
      setSavingSettings(false);
    }
  }
  const pageOptions = optionsFor(catalog, 'batch.comic-page');
  const derivationOptions = optionsFor(catalog, 'batch.comic-image-derivation');
  const promptName = text(comic.presetName) || text(pageOptions.find(option => option.value === comic.presetId)?.item?.name);
  const derivationName = text(comic.imageDerivation?.presetName) || text(derivationOptions.find(option => option.value === comic.imageDerivation?.presetId)?.item?.name);
  const constraintCount = (comic.constraints || []).filter(item => item?.presetId).length;

  const assetList = Array.isArray(comicAssets) ? comicAssets : [];
  const pageList = Array.isArray(comicPages) ? comicPages : [];
  const assetImagesFor = asset => Array.isArray(comicAssetImages?.[asset?.id]) ? comicAssetImages[asset.id] : [];
  const editingAssetImages = editingAsset ? assetImagesFor(editingAsset) : [];
  const editingAssetPrimaryImage = editingAssetImages.find(image => image?.isPrimary) || editingAssetImages[0] || null;
  const editingAssetGenerating = action === 'images' && actionKey === editingAsset?.id;
  const assetsForKind = kind => assetList.filter(asset => text(asset?.kind) === kind);
  const assetImageReady = asset => assetImagesFor(asset).length > 0;
  const assetsWithImages = assetList.filter(assetImageReady).length;
  const missingAssetImages = Math.max(0, assetList.length - assetsWithImages);
  const promptPages = pageList.filter(page => text(page?.comicPrompt) && text(page?.imagePrompt));
  const imagedPages = pageList.filter(page => (Array.isArray(page?.images) ? page.images : []).length > 0);
  // 书级就绪度以后端 comicReadiness 为准；缺失时只回退到当前已加载页，不另算一套口径。
  const readinessValue = readiness && typeof readiness === 'object' ? readiness : null;
  const hasReadiness = Boolean(readinessValue && Number.isFinite(Number(readinessValue.pagesTotal)));
  const pagesTotal = hasReadiness ? Math.max(0, Number(readinessValue.pagesTotal)) : pageList.length;
  const promptReadyCount = hasReadiness ? Math.max(0, Number(readinessValue.promptReadyCount || 0)) : promptPages.length;
  const imageReadyCount = hasReadiness ? Math.max(0, Number(readinessValue.imageReadyCount || 0)) : imagedPages.length;
  const readinessComplete = hasReadiness && pagesTotal > 0 && imageReadyCount >= pagesTotal;
  const pageGenerating = pageList.some(page => text(page?.generationState) === 'generating');
  const pageFailed = pageList.some(page => text(page?.generationState) === 'failed');
  const assetSections = [
    { key: 'character', label: '人物', empty: '提取资产后显示人物信息' },
    { key: 'scene', label: '场景', empty: '提取资产后显示场景信息' },
    { key: 'prop', label: '道具', empty: '提取资产后显示道具信息' }
  ];
  const assetKindLabel = kind => (assetSections.find(section => section.key === text(kind)) || {}).label || '资产';
  const comicAssetChips = assetList.length
    ? assetList.map(asset => ({ key: `asset-${asset.id}`, label: text(asset?.name) || '未命名资产', suffix: assetKindLabel(asset?.kind) }))
    : assetSections.map(section => ({ key: `placeholder-${section.key}`, label: section.label }));
  const pipelineSteps = [
    { key: 'assets', label: '提取资产', done: assetList.length > 0, detail: assetList.length ? `已完成 ${assetList.length} 项` : '未提取' },
    { key: 'assetImages', label: '资产图（可选）', optional: true, done: assetList.length > 0 && assetsWithImages >= assetList.length, detail: assetList.length ? `可选 · 已生成 ${assetsWithImages}/${assetList.length}` : '可选' },
    { key: 'prompts', label: '漫画提示词', done: promptReadyCount > 0, detail: pagesTotal ? `已生成 ${promptReadyCount}/${pagesTotal} 页` : (assetList.length ? '待生成' : '待资产') },
    { key: 'images', label: '漫画图片', done: pagesTotal > 0 && imageReadyCount >= pagesTotal, detail: pagesTotal ? `已出图 ${imageReadyCount}/${pagesTotal}` : '待提示词' }
  ];
  const currentStepKey = pipelineSteps.find(step => !step.done && !step.optional)?.key || '';
  const failureIds = Array.isArray(comicAssetImageFailures) ? comicAssetImageFailures : [];
  const assetImageFailed = asset => failureIds.includes(asset?.id);
  const actionBusy = Boolean(action);
  const taskReady = Boolean(task?.batchId && task?.bookId);
  const canExtract = canExtractPrivateComicAssets({ sourceText: draft.sourceText, busy, action });
  const canUseTaskActions = taskReady && !actionBusy && !busy;
  const canGenerateAssetImages = canUseTaskActions && missingAssetImages > 0;
  const canGeneratePrompts = canUseTaskActions && assetList.length > 0;
  const canGeneratePages = canUseTaskActions && promptPages.length > 0;
  const missingImageCount = Math.max(0, pageList.length - imagedPages.length);
  const allPromptsReady = pageList.length > 0 && promptPages.length >= pageList.length;
  const promptButtonLabel = action === 'prompts'
    ? '生成中…'
    : !assetList.length ? '生成漫画提示词'
      : allPromptsReady ? '重新生成漫画提示词' : '补齐漫画提示词';
  const promptButtonTitle = !taskReady
    ? '请先提取资产创建任务'
    : !assetList.length ? '需先提取资产' : actionBusy ? '当前有其他操作进行中' : '';
  const sourceLength = Array.from(text(draft.sourceText)).length;
  const pageEstimate = comicEstimatedPageRange(sourceLength);
  const draftSaveLabel = draftSaveState.status === 'saving'
    ? '保存中…'
    : draftSaveState.status === 'saved'
      ? `已保存 · ${formatDraftSavedTime(draftSaveState.savedAt) || '—'}`
      : draftSaveState.status === 'failed'
        ? '保存失败'
        : '尚未保存';
  const sourceReady = sourceLength >= privateComicMinSourceLength;
  const primaryAction = privateComicPrimaryAction({
    sourceReady,
    taskReady,
    assetCount: assetList.length,
    pagesTotal,
    promptReadyCount,
    imageReadyCount,
    busy: busy || actionBusy,
    action
  });
  const generationStatus = privateComicGenerationStatus({
    assetCount: assetList.length,
    pagesTotal,
    promptReadyCount,
    imageReadyCount,
    action,
    taskFailed: task?.status === 'failed',
    promptFailed: task?.status === 'failed' && promptReadyCount < pagesTotal,
    imageFailed: pageFailed
  });
  const generationStatusItems = [generationStatus.prompts, generationStatus.images];
  const runPrimaryAction = () => {
    if (primaryAction.key === 'prompts') onGeneratePrompts?.();
    if (primaryAction.key === 'images') onGenerateAllPageImages?.();
  };
  const progress = Number.isFinite(Number(task?.progress)) ? Math.max(0, Math.min(100, Number(task.progress))) : null;
  const comicSourceTitle = text(draft.title) || privateComicTitle(draft.sourceText, '');
  const comicHeaderTitle = comicSourceTitle
    ? `《${comicSourceTitle.replace(/^《|》$/g, '').slice(0, 24)}》`
    : '未命名漫画';
  // 页数取后端 comicReadiness.pagesTotal；时长在漫画链路不存在（见方案 D4），因此不渲染。
  const comicHeaderMeta = sourceReady
    ? ['一段内容', pagesTotal ? `${pagesTotal} 页` : '尚未生成分镜', imageReadyCount ? `已出图 ${imageReadyCount} 页` : ''].filter(Boolean).join(' · ')
    : '等待输入内容';
  const comicHeaderStatus = action === 'assets' || action === 'prompts' || action === 'pages' || pageGenerating
    ? '生成中'
    : task?.status === 'failed' || pageFailed
      ? '生成失败'
      : readinessComplete || task?.status === 'succeeded' || task?.status === 'completed'
        ? '已完成'
        : assetList.length > 0 || promptReadyCount > 0 || pageList.length > 0
          ? '准备生成'
          : '待输入';

  return <section id="private-customization-panel-comic" ref={workbenchRef} style={leftPanelWidth ? { gridTemplateColumns: `${leftPanelWidth}px 8px minmax(460px, 1fr)` } : undefined} className="script-workbench private-comic-workbench private-studio-workbench private-studio-resizable" role="tabpanel" aria-labelledby="private-customization-tab-comic">
    <div className="script-left private-studio-side">
      <div className="script-left-scroll">
        <div className="private-studio-left-fixed">
        <div className="private-customization-brand-row">
          <div className="private-customization-brand"><span className="private-customization-brand-mark">私</span><strong>私人定制</strong></div>
          <span
            className={`private-customization-save-state is-${draftSaveState.status}`}
            role="status"
            aria-live="polite"
            title={draftSaveState.status === 'failed' ? `草稿保存失败：${draftSaveState.reason}` : undefined}
          ><i />{draftSaveLabel}</span>
        </div>
        <div className="private-customization-workspaces" role="tablist" aria-label="私人定制工作区">
          <button type="button" role="tab" id="private-customization-tab-script" aria-controls="private-customization-panel-script" aria-selected={customizationWorkspace === 'script'} className={customizationWorkspace === 'script' ? 'is-active' : ''} onClick={() => onWorkspaceChange?.('script')}>剧本创作</button>
          <button type="button" role="tab" id="private-customization-tab-comic" aria-controls="private-customization-panel-comic" aria-selected={customizationWorkspace === 'comic'} className={customizationWorkspace === 'comic' ? 'is-active' : ''} onClick={() => onWorkspaceChange?.('comic')}>漫画生成</button>
        </div>
        <div className="private-customization-panel-heading">
          <div>
            <span className="private-customization-eyebrow">01 · 粘贴你的内容</span>
            <h2>输入内容</h2>
          </div>
          <span className="private-customization-source-hint">支持 TXT / 长文本</span>
        </div>
        <div className="script-chat-shell private-comic-chat-shell">
          <Input
            aria-label="漫画标题"
            name="title"
            className="private-comic-title-input"
            maxLength={80}
            value={draft.title || (draft.sourceText ? privateComicTitle(draft.sourceText, '') : '')}
            placeholder="漫画标题（默认使用正文第一行）"
            disabled={busy}
            onKeyDown={event => { if (event.key === 'Enter') event.preventDefault(); }}
            onChange={event => patchDraft({ title: event.target.value })}
          />
          <div className="script-source-input">
            <Input.TextArea
              aria-label="漫画正文"
              name="sourceText"
              className="script-chat-textarea"
              autoSize={{ minRows: 4, maxRows: 10 }}
              value={draft.sourceText || ''}
              placeholder="把要生成漫画的正文直接粘贴到这里"
              disabled={busy}
              onChange={event => patchDraft({ sourceText: event.target.value })}
            />
            {text(draft.sourceText) ? (
              <Popconfirm
                title={`清空已粘贴的 ${sourceLength} 字正文？`}
                description="该操作不可撤销。"
                okText="清空正文"
                cancelText="保留正文"
                okButtonProps={{ danger: true }}
                onConfirm={() => patchDraft({ sourceText: '' })}
              >
                <button
                  type="button"
                  className="script-source-clear"
                  aria-label="清空漫画正文"
                  title="清空漫画正文"
                >×</button>
              </Popconfirm>
            ) : null}
          </div>
          <div className="script-chat-options">
            <div className="script-chat-tools">
              <Button
                type="text"
                aria-label="刷新资产与分镜"
                title="刷新资产与分镜"
                icon={<RefreshCw size={16} strokeWidth={1.8} aria-hidden="true" />}
                disabled={!taskReady}
                onClick={onRefresh}
              />
              <Button
                className="private-studio-extract"
                type="primary"
                loading={action === 'assets'}
                aria-label="提取漫画资产"
                disabled={!canExtract}
                title={!sourceReady ? `漫画正文至少需要 ${privateComicMinSourceLength} 个字` : actionBusy ? '当前有其他操作进行中' : ''}
                icon={<WandSparkles size={14} strokeWidth={1.8} aria-hidden="true" />}
                onClick={onExtractAssets}
              >{action === 'assets' ? '提取中…' : assetList.length ? '重新提取资产' : '提取资产'}</Button>
            </div>
          </div>
        </div>
        <div className="private-studio-count-row">
          <span>{sourceLength ? `已读取 ${sourceLength} 字 · 预计约 ${pageEstimate.min}～${pageEstimate.max} 张` : '尚未读取'}</span>
          <span>自动保存</span>
        </div>
        {draftSaveState.status === 'failed' ? (
          <div className="private-studio-draft-error" role="alert">
            草稿保存失败，正文仍在页面上，请勿关闭（{draftSaveState.reason}）
          </div>
        ) : null}

        <div className="script-instruction-status private-comic-instruction">
          <span className="private-comic-instruction-label">流程</span>
          <span className="private-comic-pipeline">
            {pipelineSteps.map((step, index) => (
              <span
                key={step.key}
                className={`private-comic-pipeline-step${step.done ? ' is-done' : ''}${!step.done && step.key === currentStepKey ? ' is-active' : ''}`}
              >
                <b>{index + 1}. {step.label}</b>
                <small>{step.detail}</small>
              </span>
            ))}
          </span>
        </div>

        </div>
        <div className="private-studio-left-scroll-body">
        <div className="entity-section legacy-panel-card">
          <div className="entity-section-header">
            <Typography.Text strong>资产</Typography.Text>
            <Space size={8}>
              <span className="legacy-muted">{assetList.length}</span>
              <Button
                size="small"
                type="text"
                className="entity-add-button"
                loading={action === 'images'}
                disabled={!canGenerateAssetImages}
                title={!taskReady ? '请先提取资产创建任务' : !assetList.length ? '需先提取资产' : !missingAssetImages ? '资产图已全部生成' : actionBusy ? '当前有其他操作进行中' : ''}
                onClick={onGenerateMissingAssetImages}
              >{action === 'images' ? '生成中…' : missingAssetImages ? `生成资产图（可选，缺 ${missingAssetImages}）` : '资产图已齐'}</Button>
            </Space>
          </div>
          <PrivateStudioChips items={comicAssetChips} />
          {assetSections.map(section => {
            const items = assetsForKind(section.key);
            const visible = items.slice(0, 3);
            return <div className="private-comic-asset-row" key={section.key}>
              <b className="private-comic-asset-row-label">{section.label}<span className="legacy-muted">{items.length}</span></b>
              {visible.length ? visible.map(asset => <button
                type="button"
                className="private-comic-asset-name"
                key={asset.id}
                title={`${asset.name || '未命名资产'} · ${assetImageReady(asset) ? '资产图已生成' : assetImageFailed(asset) ? '资产图生成失败，可补齐' : '资产图未生成'}`}
                onClick={() => setEditingAsset({ ...asset })}
              >{asset.name || '未命名资产'}</button>) : <span className="private-comic-asset-empty">{section.empty}</span>}
              {items.length > 3 ? <Button size="small" type="text" className="entity-section-more" onClick={() => setAssetPreview({ title: section.label, items })}>查看全部</Button> : null}
            </div>;
          })}
        </div>
        {error ? <Alert type="error" showIcon message={error} /> : null}
        <Typography.Text type="secondary" className="private-comic-hint">仅处理当前正文</Typography.Text>
        </div>
      </div>
    </div>

    <div
      className="script-resize-handle"
      role="separator"
      aria-label="调整左右面板宽度"
      aria-orientation="vertical"
      onPointerDown={handleResizeStart}
    />
    <div className="script-right private-studio-main">
      <div className="private-customization-canvas-heading">
        <div>
          <span className="private-customization-eyebrow">创作画布</span>
          <div className="private-customization-title-line"><h2>{comicHeaderTitle}</h2><span className="private-customization-ready-badge"><i />{comicHeaderStatus}</span></div>
          <p className="private-customization-work-meta">{comicHeaderMeta}</p>
        </div>
        <div className="private-studio-head-right">
          <span className="private-customization-status-dot"><i />不发布</span>
          <PrivateStudioHeaderActions
            onHistory={onOpenBatchFactory}
            historyDisabled={!task?.batchId}
            historyTitle="漫画历史版本（在批量工厂查看）"
            onSettings={onOpenUnifiedConfig}
          />
        </div>
      </div>
      <div className="script-tabs">
        <Typography.Text strong>漫画生成</Typography.Text>
        <Button
          type="text"
          icon={<Settings2 size={16} aria-hidden="true" />}
          onClick={onOpenUnifiedConfig}
          title="漫画生成统一配置"
        >配置</Button>
        <span className="private-customization-comic-no-publish">不自动发布</span>
        <div style={{ flex: 1 }} />
        {task?.bookId ? <span className="private-comic-task-meta">{task.bookId}</span> : null}
      </div>
      <div className="script-toolbar private-studio-toolbar">
        <div className="private-studio-toolbar-left">
        <div className="private-studio-tool-pill">
          <span className="private-studio-tool-label">文本模型</span>
        <Select className="private-studio-tool-select" style={{ width: 140, flex: '0 0 140px' }} allowClear value={draft.textModelId || undefined} options={modelOptions.text || []} placeholder="文本模型" disabled={busy} onChange={textModelId => patchDraft({ textModelId: textModelId || '' })} />
        </div>
        <div className="private-studio-tool-pill">
          <span className="private-studio-tool-label">图片模型</span>
        <Select className="private-studio-tool-select" style={{ width: 140, flex: '0 0 140px' }} allowClear value={draft.imageModelId || undefined} options={modelOptions.image || []} placeholder="图片模型" disabled={busy} onChange={imageModelId => patchDraft({ imageModelId: imageModelId || '' })} />
        </div>
        </div>
        <div className="private-studio-toolbar-right">
        <Button
          className="script-constraint-button private-studio-tool-pill-button"
          icon={<Settings2 size={16} strokeWidth={1.8} aria-hidden="true" />}
          onClick={() => setSettingsOpen(true)}
          title="约束设置"
        >约束条件</Button>
          <Space>
            <Button
              className="private-studio-tool-primary"
              type="primary"
              icon={<WandSparkles size={16} strokeWidth={1.8} aria-hidden="true" />}
              loading={action === 'prompts' || action === 'pages'}
              disabled={primaryAction.disabled}
              title={!sourceReady ? `漫画正文至少需要 ${privateComicMinSourceLength} 个字` : !taskReady || !assetList.length ? '请先提取资产' : ''}
              onClick={runPrimaryAction}
            >{primaryAction.label}</Button>
            <Dropdown
              trigger={['click']}
              menu={{
                items: [
                  { key: 'prompts', label: promptButtonLabel, icon: <WandSparkles size={14} />, disabled: !canGeneratePrompts },
                  { key: 'pages', label: missingImageCount ? `出图（缺 ${missingImageCount}）` : '批量出图', icon: <WandSparkles size={14} />, disabled: !canGeneratePages || (action === 'pages' && actionKey === '') },
                  { key: 'download', label: action === 'download' ? '正在打包…' : '下载提示词与图片（ZIP）', title: '复用批量工厂的漫画打包下载', icon: <Download size={14} />, disabled: !taskReady || (Boolean(action) && action !== 'download') },
                  { key: 'refresh', label: '刷新', title: '刷新状态', icon: <RefreshCw size={14} />, disabled: !task?.batchId || busy },
                  { key: 'detail', label: '查看详情', title: '查看漫画详情', icon: <ExternalLink size={14} />, disabled: !task?.batchId },
                  { key: 'retry', label: '重试', disabled: task?.status !== 'failed' || busy }
                ],
                onClick: ({ key }) => {
                  if (key === 'prompts') onGeneratePrompts();
                  if (key === 'pages') onGenerateAllPageImages();
                  if (key === 'download') onDownloadAll();
                  if (key === 'refresh') onRefresh();
                  if (key === 'detail') onOpenBatchFactory();
                  if (key === 'retry') onRetry();
                }
              }}
            >
              <Button aria-label="更多操作" title="更多操作" icon={<MoreHorizontal size={17} strokeWidth={1.8} />} />
            </Dropdown>
          </Space>
      </div>
          </div>
      <div className="private-comic-generation-statuses" aria-label="漫画生成状态">
        {generationStatusItems.map(statusItem => (
          <div
            key={statusItem.key}
            className={`private-comic-generation-status is-${statusItem.tone}`}
            role="status"
            aria-live="polite"
          >
            <span className="private-comic-generation-status-icon" aria-hidden="true" />
            <span className="private-comic-generation-status-copy">
              <b>{statusItem.key === 'prompts' ? '漫画提示词' : '漫画图片'}</b>
              <small>{statusItem.label} · {statusItem.detail}</small>
            </span>
            {statusItem.canRetry ? <Button
              size="small"
              className="private-comic-generation-status-retry"
              disabled={busy || actionBusy || !taskReady}
              title={statusItem.key === 'prompts' ? '只补齐尚未完成的漫画提示词' : '只生成尚未完成的漫画图片'}
              onClick={statusItem.key === 'prompts' ? onGeneratePrompts : onGenerateAllPageImages}
            >重试未完成项</Button> : null}
          </div>
        ))}
      </div>
      <div className="private-studio-content">
      <div className="private-studio-center">
      <div className="script-output">
        {task ? <div className="legacy-panel-card private-comic-result-card" role="status" aria-live="polite">
          <div className="private-comic-result-heading">
            <b>{taskLabel(task)}</b>
            {task.bookId ? <span>{`任务已保存 · ${task.bookId}`}</span> : null}
          </div>
          {progress != null ? <Progress percent={progress} status={task.status === 'failed' ? 'exception' : undefined} /> : null}
          {task.message ? <Typography.Paragraph type="secondary">{task.message}</Typography.Paragraph> : null}
          {task.error ? <Typography.Paragraph type="danger">{task.error}</Typography.Paragraph> : null}
        </div> : null}
        {pageList.length ? <ComicOutputCards
          pages={pageList}
          taskReady={taskReady}
          action={action}
          actionKey={actionKey}
          onGeneratePage={onGeneratePageImage}
          onSavePagePrompt={onSavePagePrompt}
          onDownloadAll={onDownloadAll}
        /> : task ? <div className="script-empty legacy-panel-card">
          <div className="script-empty-title">还没有漫画页</div>
          <div className="script-empty-copy">先在左侧提取资产，再点击「生成漫画提示词」；确认提示词后再生成漫画图片。</div>
        </div> : <div className="script-empty legacy-panel-card">
          <div className="script-empty-card">
            <span className="script-empty-dot" aria-hidden="true" />
            <span className="script-empty-ray" aria-hidden="true" />
            <span className="script-empty-line script-empty-line--top" aria-hidden="true" />
            <span className="script-empty-line script-empty-line--bottom" aria-hidden="true" />
          </div>
          <div className="script-empty-title">准备生成漫画</div>
          <div className="script-empty-copy">在左侧粘贴正文并提取人物、场景与道具，再依次生成漫画提示词和漫画图片。</div>
        </div>}
      </div>
      </div>
      <aside className="private-studio-side-preview">
        <PrivateStudioNote mode="comic" />
      </aside>
      </div>
    </div>

    <PrivateStudioConstraintsModal
      open={settingsOpen}
      onCancel={() => setSettingsOpen(false)}
      onOk={() => void saveSettings()}
      okText={task?.batchId ? '保存到任务' : '完成'}
      okButtonProps={{ loading: savingSettings }}
      footerNote="漫画约束只作用于本书的漫画生成，不会改动批量工厂的预设。已有任务时点「保存到任务」会立即写回服务端并回读校验；未出图的页仍会写入打包下载里的提示词文件。"
    >
      <div className="private-customization-comic-settings">
        <PrivateStudioConstraintSection
          title="画风与参考图"
          description="画风库图片会默认作为参考图发给 AI；也可以上传自己的全局参考图。"
        >
          <BatchFactoryComicStylePicker
            batchId={task?.batchId || ''}
            catalog={catalog}
            value={comic}
            disabled={busy}
            onReferenceDataUrl={onPendingReferenceChange}
            onChange={next => patchDraft({ comic: next })}
          />
          {!task?.batchId ? <Typography.Text type="secondary" className="private-customization-reference-note">首次提取资产时会创建任务并上传自定义参考图。</Typography.Text> : null}
        </PrivateStudioConstraintSection>
        <PrivateStudioConstraintSection title="漫画提示词" description="决定分镜文本的组织方式。">
          <Select allowClear value={comic.presetId || undefined} options={pageOptions} placeholder="选择已发布漫画提示词" disabled={busy} onChange={value => patchComic(presetSelection(pageOptions.find(option => option.value === value)?.item))} />
        </PrivateStudioConstraintSection>
        <PrivateStudioConstraintSection title="图片描述词推导" description="决定图片提示词的推导方式。">
          <Select allowClear value={comic.imageDerivation?.presetId || undefined} options={derivationOptions} placeholder="选择图片描述词推导" disabled={busy} onChange={value => patchComic({ imageDerivation: presetSelection(derivationOptions.find(option => option.value === value)?.item) })} />
        </PrivateStudioConstraintSection>
        {PRIVATE_STUDIO_CONSTRAINT_FIELDS.comic.map(field => <PrivateStudioConstraintSection
          key={field.key}
          title={field.label}
          description="选择已发布的漫画约束预设；留空表示不启用。"
        >
          <Select allowClear value={selectedConstraint(comic, field.key) || undefined} options={constraintOptions(catalog, field.key)} placeholder="未选择" onChange={value => {
            const rest = (comic.constraints || []).filter(item => item?.constraintCategory !== field.key);
            const option = constraintOptions(catalog, field.key).find(item => item.value === value);
            patchComic({ constraints: value && option ? [...rest, { ...presetSelection(option.item), constraintCategory: field.key }] : rest });
          }} />
        </PrivateStudioConstraintSection>)}
        <small className="entity-media-status private-comic-advanced-note">{promptName ? `提示词「${promptName}」` : '提示词未选择'} · {derivationName ? `推导「${derivationName}」` : '推导未选择'} · 约束 {constraintCount ? `已配置 ${constraintCount} 项` : '未配置'}</small>
      </div>
    </PrivateStudioConstraintsModal>

    <Modal
      title={assetPreview ? `选择${assetPreview.title} · ${assetPreview.items.length}` : ''}
      open={Boolean(assetPreview)}
      onCancel={() => setAssetPreview(null)}
      footer={null}
      width={520}
    >
      {(assetPreview?.items || []).map(asset => <button type="button" className="private-comic-asset-preview-row" key={asset.id} onClick={() => {
        setAssetPreview(null);
        setEditingAsset({ ...asset });
      }}>
        <span>{asset.name || '未命名资产'}</span>
        <small className="entity-media-status">
          {assetImageReady(asset)
            ? `资产图已生成${assetImagesFor(asset).length > 1 ? ` · ${assetImagesFor(asset).length} 张` : ''}`
            : assetImageFailed(asset) ? '资产图生成失败，可补齐' : '资产图未生成'}
        </small>
      </button>)}
    </Modal>

    <PrivateStudioAssetEditor
      open={Boolean(editingAsset)}
      title={`编辑资产 · ${editingAsset?.name || '未命名资产'}`}
      onClose={() => setEditingAsset(null)}
      width={920}
      basic={<>
        <div className="batch-factory-asset-editor-config-row">
          <label className="batch-factory-asset-editor-config-field is-kind"><span>资产分类</span><Input value={assetKindLabel(editingAsset?.kind)} disabled /></label>
          <label className="batch-factory-asset-editor-config-field is-model"><span>图片模型</span><Select allowClear value={editingAsset?.imageModelId || draft.imageModelId || undefined} placeholder="使用书籍默认" options={modelOptions.image || []} onChange={imageModelId => setEditingAsset(current => ({ ...current, imageModelId: imageModelId || '' }))} /></label>
          <label className="batch-factory-asset-editor-config-field is-ratio"><span>图片画幅</span><Select allowClear value={editingAsset?.imageAspectRatio || '9:16'} placeholder="默认 9:16" options={IMAGE_ASPECT_RATIO_OPTIONS} onChange={imageAspectRatio => setEditingAsset(current => ({ ...current, imageAspectRatio: imageAspectRatio || '' }))} /></label>
        </div>
        <label className="shuihuo-form-label">名称<Input value={editingAsset?.name || ''} onChange={event => setEditingAsset(current => ({ ...current, name: event.target.value }))} /></label>
      </>}
      settings={editingAsset?.kind === 'character' ? <section className="batch-factory-asset-editor-character-section">
        <div className="batch-factory-asset-editor-section-title">人物设定</div>
        <div className="batch-factory-asset-editor-character-fields">
          <label className="shuihuo-form-label">人物事实<Input.TextArea rows={3} value={editingAsset?.facts || ''} onChange={event => setEditingAsset(current => ({ ...current, facts: event.target.value }))} placeholder="身份、性别、年龄阶段、别名等原文事实" /></label>
          <label className="shuihuo-form-label">人物关系<Input.TextArea rows={3} value={editingAsset?.relationships || ''} onChange={event => setEditingAsset(current => ({ ...current, relationships: event.target.value }))} placeholder="与其他人物的关系及原文依据" /></label>
          <label className="shuihuo-form-label batch-factory-asset-editor-character-appearance">人物外形<Input.TextArea rows={4} value={editingAsset?.appearance || editingAsset?.prompt || ''} onChange={event => setEditingAsset(current => ({ ...current, appearance: event.target.value }))} placeholder="脸型、发型、服装、体态等跨镜头一致性锚点" /></label>
        </div>
      </section> : <Typography.Text type="secondary">场景 / 道具暂无独立设定字段，提示词在下方维护。</Typography.Text>}
      prompts={<label className="shuihuo-form-label">合并资产提示词 / AI 生图提示词<Input.TextArea rows={editingAsset?.kind === 'character' ? 6 : 12} value={editingAsset?.prompt || ''} onChange={event => setEditingAsset(current => ({ ...current, prompt: event.target.value }))} placeholder="输入最终合并后的可视化提示词" /></label>}
      ruleNote="规则：这个资产有图片时，生成视频走“图片生视频”，不发文字设定；没图片时，按上面的文字走“文字生视频”。"
      aside={<>
        <div className="batch-factory-asset-editor-panel-head"><b>主图预览</b><span className="entity-media-status">图片 {editingAssetImages.length} 张</span></div>
        {editingAssetPrimaryImage ? <div className="batch-factory-asset-editor-main"><PrivateComicAssetImage imageUrl={editingAssetPrimaryImage.url} alt={`${editingAsset?.name || '资产'} 主图`} preview /></div>
          : <div className="batch-factory-asset-editor-main is-empty"><ImageIcon size={34} aria-hidden="true" /><span>{assetImageFailed(editingAsset) ? '资产图生成失败，可重新生成' : '暂无主图，请上传图片或用 AI 生成'}</span></div>}
        <div className="batch-factory-asset-editor-panel-actions">
          {editingAsset?.kind === 'character' ? <Tooltip title="开启后人物资产图按“正面近照 + 正面/侧面/背面全身”三视图生成，便于分镜保持角色一致。"><span className="batch-factory-asset-editor-three-view"><Switch size="small" checked={assetImageThreeView} onChange={setAssetImageThreeView} disabled={assetSaving || editingAssetGenerating} />人物三视图</span></Tooltip> : null}
          <label className="batch-factory-image-upload"><input type="file" accept="image/png,image/jpeg,image/webp" onChange={uploadEditingAssetImage} disabled={assetSaving || editingAssetGenerating} />上传图片版本</label>
          <Button type="primary" loading={editingAssetGenerating} disabled={assetSaving || editingAssetGenerating || !editingAsset?.id} onClick={() => onGenerateAssetImage?.(editingAsset, privateComicAssetUseThreeView(editingAsset?.kind, assetImageThreeView))}>AI生成</Button>
        </div>
        <div className="batch-factory-asset-editor-thumbs">{editingAssetImages.map(image => <button type="button" key={image.id} className={`batch-factory-asset-editor-thumb${image.id === editingAssetPrimaryImage?.id ? ' is-main' : ''}`} title={image.isPrimary ? '当前主图' : '切换为主图'} disabled={assetSaving || editingAssetGenerating} onClick={() => void makeEditingAssetImagePrimary(image)}><PrivateComicAssetImage imageUrl={image.url} alt={`${editingAsset?.name || '资产'} 版本 ${image.revision || ''}`} />{image.isPrimary ? <span className="batch-factory-asset-editor-thumb-tag">主图</span> : null}</button>)}{!editingAssetImages.length ? <span className="batch-factory-asset-editor-thumbs-empty">还没有图片版本</span> : null}</div>
      </>}
      footer={<Space><Button onClick={() => setEditingAsset(null)}>取消</Button><Button type="primary" loading={assetSaving} onClick={() => void saveEditingAsset()}>保存</Button></Space>}
    />
  </section>;
}

export default PrivateComicWorkspace;
