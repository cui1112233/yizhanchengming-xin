import {
  Alert,
  Button,
  Collapse,
  Empty,
  Input,
  InputNumber,
  Progress,
  Segmented,
  Select,
  Space,
  Tag,
  Typography,
  message
} from 'antd';
import {
  AlertTriangle,
  Archive,
  Film,
  FolderPlus,
  LayoutGrid,
  RotateCcw,
  Save,
  Search,
  Settings2,
  SlidersHorizontal,
  Upload,
  Video
} from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { DirectorPanel } from './DirectorPanel';
import { HookReviewPanel } from './HookReviewPanel';
import { NovelSourceModule } from './NovelSourceModule.jsx';
import { WorkbenchCard } from './WorkbenchCard';
import { actionState } from './batchFactoryV11State.js';
import { unavailableCapabilitySummary } from './capabilityReadiness.js';
import {
  DEFAULT_WORKSPACE_LAYOUT,
  GRID_COLUMNS,
  WORKSPACE_STORAGE_KEY,
  moveWorkspaceItem,
  normalizeLayout,
  resetWorkspaceLayout,
  resizeWorkspaceItem,
  setItemMaximized,
  setWorkspaceItemCollapsed,
  setWorkspaceItemHidden
} from './workspace-layout';
import './batch-factory-v11-workbench.css';

const { TextArea } = Input;

const STATUS_ORDER = ['全部', '待开始', '待审核', 'AI处理中', '待生成', '排队中', '生成中', '异常', '待合并', '已合并', '待上传', '已发布'];

const statusTone = {
  待开始: 'default',
  待审核: 'gold',
  AI处理中: 'processing',
  待生成: 'purple',
  排队中: 'blue',
  生成中: 'processing',
  异常: 'red',
  待合并: 'cyan',
  已合并: 'green',
  待上传: 'geekblue',
  已发布: 'green',
  已完成: 'green'
};

const statusColor = {
  待开始: '#94a3b8',
  待审核: '#d4a72c',
  AI处理中: '#7c6cff',
  待生成: '#8b5cf6',
  排队中: '#3b82f6',
  生成中: '#2f9ee8',
  异常: '#ee7777',
  待合并: '#22b8cf',
  已合并: '#37c98b',
  待上传: '#4f7cff',
  已发布: '#37c98b',
  已完成: '#37c98b'
};

function loadStoredLayout() {
  try {
    return normalizeLayout(JSON.parse(localStorage.getItem(WORKSPACE_STORAGE_KEY) || 'null'));
  } catch (_) {
    return normalizeLayout(DEFAULT_WORKSPACE_LAYOUT);
  }
}

function assetName(item) {
  return typeof item === 'string' ? item : (item?.name || item?.label || item?.id || '未命名资产');
}

function assetPrompt(item) {
  return typeof item === 'string' ? '' : (item?.prompt || item?.visualPrompt || item?.description || '');
}

function AssetPromptGroup({ label, type, items, drafts, onChange, onRefresh, onSave, onGenerateImage, generatedImages = {} }) {
  return <div className="bf11-asset-prompt-group">
    <div className="bf11-asset-prompt-title">
      <Typography.Text strong>{label}</Typography.Text>
      <Space>
        <Button size="small" disabled={!onRefresh} onClick={onRefresh}>重新获取</Button>
        <Button size="small" disabled={!onSave} onClick={() => onSave(type, items, drafts)}>保存</Button>
      </Space>
    </div>
    <div className="bf11-asset-prompt-list">
      {(items || []).map((item, index) => {
        const name = assetName(item);
        const key = `${type}:${item?.id || name}:${index}`;
        return <div className="bf11-asset-prompt-row" key={key}>
          <Input value={name} disabled />
          <TextArea
            rows={3}
            value={drafts[key] ?? assetPrompt(item)}
            placeholder="等待 V11 资产 Prompt 数据"
            onChange={event => onChange(key, event.target.value)}
          />
          {generatedImages[item?.id || name] ? <img className="bf11-asset-preview" src={generatedImages[item?.id || name]} alt={name} /> : null}
          <Button size="small" disabled={!onGenerateImage || !assetPrompt(item) && !drafts[key]} onClick={() => onGenerateImage(item, drafts[key] ?? assetPrompt(item))}>生成图片</Button>
        </div>;
      })}
      {!(items || []).length ? <Typography.Text type="secondary">当前没有服务端资产数据。</Typography.Text> : null}
    </div>
  </div>;
}

export function BatchFactoryV11Workbench({
  batch = null,
  books = [],
  productionStatus = null,
  mergeStatus = null,
  capabilities = {},
  onOpenBatchManager,
  onOpenHistory,
  onOpenBatchSettings,
  onOpenPublishSettings,
  onOpenBookSettings,
  onOpenVideoSettings,
  onRunHook,
  onApproveHook,
  onRunDirector,
  onRunBatchDirector,
  onPreviewFinalPrompt,
  onRunProduction,
  onSaveSource,
  onSaveVideoPrompt,
  onRefreshAssets,
  onSaveAssetPrompts,
  onGenerateShotImage,
  generatedShotImages = {},
  onGenerateAssetImage,
  generatedAssetImages = {},
  onRunMerge,
  onRunUpload
}) {
  const [selectedBookId, setSelectedBookId] = useState(books[0]?.id || '');
  const [selectedVideoId, setSelectedVideoId] = useState(books[0]?.videos?.[0]?.id || '');
  const [currentShotIndex, setCurrentShotIndex] = useState(0);
  const [statusFilter, setStatusFilter] = useState('全部');
  const [bookSearch, setBookSearch] = useState('');
  const [previewTarget, setPreviewTarget] = useState('merged');
  const [editMode, setEditMode] = useState(false);
  const [layout, setLayout] = useState(loadStoredLayout);
  const [assetDrafts, setAssetDrafts] = useState({});
  const [videoPromptDrafts, setVideoPromptDrafts] = useState({});
  const [mergeTimingMode, setMergeTimingMode] = useState('speed');
  const [mergeSpeed, setMergeSpeed] = useState('1.0');
  const [ttsSpeed, setTtsSpeed] = useState(1.7);
  const gridRef = useRef(null);
  const [gridWidth, setGridWidth] = useState(1200);

  const batchCreateAction = actionState(capabilities, 'batch.create');
  const batchReadAction = actionState(capabilities, 'batch.read');
  const settingsAction = actionState(capabilities, 'settings.edit');
  const overrideAction = actionState(capabilities, 'override.edit');
  const productionAction = actionState(capabilities, 'production.submit');
  const mergeAction = actionState(capabilities, 'merge.run');
  const compilerAction = actionState(capabilities, 'compiler.preview');
  const unavailableCapabilities = useMemo(() => unavailableCapabilitySummary(capabilities), [capabilities]);

  const selectedBook = useMemo(
    () => books.find(book => book.id === selectedBookId) || books[0] || null,
    [books, selectedBookId]
  );

  const selectedVideo = useMemo(() => {
    const videos = selectedBook?.videos || [];
    return videos.find(video => video.id === selectedVideoId) || videos[0] || null;
  }, [selectedBook, selectedVideoId]);

  const selectedShots = selectedVideo?.shots || [];
  const selectedShot = selectedShots[currentShotIndex] || selectedShots[0] || null;

  useEffect(() => {
    if (!books.length) {
      setSelectedBookId('');
      setSelectedVideoId('');
      setCurrentShotIndex(0);
      return;
    }
    if (!books.some(book => book.id === selectedBookId)) {
      setSelectedBookId(books[0].id);
      setSelectedVideoId(books[0]?.videos?.[0]?.id || '');
      setCurrentShotIndex(0);
      setPreviewTarget('merged');
    }
  }, [books, selectedBookId]);

  useEffect(() => {
    if (!selectedBook) return;
    if (!(selectedBook.videos || []).some(video => video.id === selectedVideoId)) {
      setSelectedVideoId(selectedBook.videos?.[0]?.id || '');
      setCurrentShotIndex(0);
      setPreviewTarget('merged');
    }
  }, [selectedBook?.id, selectedVideoId]);

  useEffect(() => {
    if (!selectedShots.length) {
      setCurrentShotIndex(0);
      return;
    }
    if (currentShotIndex >= selectedShots.length) setCurrentShotIndex(selectedShots.length - 1);
  }, [selectedVideo?.id, selectedShots.length, currentShotIndex]);

  useEffect(() => {
    const node = gridRef.current;
    if (!node) return undefined;
    const update = () => setGridWidth(node.clientWidth || 1200);
    update();
    if (typeof ResizeObserver === 'undefined') {
      window.addEventListener('resize', update);
      return () => window.removeEventListener('resize', update);
    }
    const observer = new ResizeObserver(update);
    observer.observe(node);
    return () => observer.disconnect();
  }, []);

  const statusCounts = useMemo(() => STATUS_ORDER.reduce((out, status) => {
    out[status] = status === '全部' ? books.length : books.filter(book => book.status === status).length;
    return out;
  }, {}), [books]);

  const filteredBooks = useMemo(() => {
    const query = bookSearch.trim().toLowerCase();
    return books.filter(book => !query || `${book.title || ''} ${book.bookId || ''} ${book.platform || ''}`.toLowerCase().includes(query));
  }, [books, bookSearch]);

  const filterResults = useMemo(
    () => statusFilter === '全部' ? books : books.filter(book => book.status === statusFilter),
    [books, statusFilter]
  );

  const videoProgress = useMemo(() => {
    const all = books.flatMap(book => book.videos || []);
    const taskStatus = new Map();
    for (const job of productionStatus?.jobs || []) {
      for (const task of job.tasks || []) taskStatus.set(task.videoId, task.status);
    }
    const summary = { total: all.length, completed: 0, pending: 0, queued: 0, generating: 0, failed: 0 };
    for (const video of all) {
      const durableStatus = taskStatus.get(video.id);
      if (durableStatus === 'succeeded' || video.status === '已完成') summary.completed += 1;
      else if (durableStatus === 'failed' || video.status === '异常') summary.failed += 1;
      else if (durableStatus === 'queued' || video.status === '排队中') summary.queued += 1;
      else if (durableStatus === 'running' || video.status === '生成中') summary.generating += 1;
      else summary.pending += 1;
    }
    summary.percent = summary.total ? Math.round(((summary.completed + summary.failed) / summary.total) * 100) : 0;
    return summary;
  }, [books, productionStatus]);

  const productionTaskByVideoId = useMemo(() => {
    const map = new Map();
    for (const job of productionStatus?.jobs || []) {
      for (const task of job.tasks || []) {
        if (task?.videoId) map.set(task.videoId, task);
      }
    }
    return map;
  }, [productionStatus]);

  const productionTaskByShotId = useMemo(() => {
    const map = new Map();
    for (const job of productionStatus?.jobs || []) {
      for (const task of job.tasks || []) {
        if (task?.shotId) map.set(task.shotId, task);
      }
    }
    return map;
  }, [productionStatus]);

  const metrics = useMemo(() => ({
    columnWidth: Math.max(20, (gridWidth - ((GRID_COLUMNS - 1) * layout.gap)) / GRID_COLUMNS),
    rowHeight: layout.rowHeight,
    gap: layout.gap
  }), [gridWidth, layout.gap, layout.rowHeight]);

  function selectBook(book) {
    setSelectedBookId(book.id);
    setSelectedVideoId(book.videos?.[0]?.id || '');
    setCurrentShotIndex(0);
    setPreviewTarget('merged');
  }

  function updateCard(id, updater) {
    setLayout(current => updater(current, id));
  }

  function saveLayout() {
    localStorage.setItem(WORKSPACE_STORAGE_KEY, JSON.stringify(layout));
    setEditMode(false);
    message.success('布局已保存');
  }

  function restoreDefault() {
    const next = resetWorkspaceLayout();
    setLayout(next);
    localStorage.setItem(WORKSPACE_STORAGE_KEY, JSON.stringify(next));
    message.success('已恢复默认布局');
  }

  function restoreHiddenCards() {
    setLayout(current => {
      let next = normalizeLayout(current);
      for (const id of Object.keys(next.items)) next = setWorkspaceItemHidden(next, id, false);
      return next;
    });
  }

  function setAssetDraft(key, value) {
    setAssetDrafts(current => ({ ...current, [key]: value }));
  }

  const bookAssets = selectedBook?.assets || {};
  const selectedVideos = selectedBook?.videos || [];
  const selectedHook = selectedBook?.hook || selectedBook?.hookRevision || null;

  const bookCollapseItems = selectedBook ? [
    {
      key: 'source',
      label: <span className="bf11-fold-label"><strong>原文与 Hook</strong><small>先确认来源，再决定开头策略</small></span>,
      children: <NovelSourceModule
        book={selectedBook}
        onSave={onSaveSource}
        hookContent={<HookReviewPanel book={selectedBook} capabilities={capabilities} onRunHook={onRunHook} onApproveHook={onApproveHook} />}
      />
    },
    {
      key: 'assets',
      label: <span className="bf11-fold-label"><strong>管理资产</strong><small>人物与场景仅属于当前小说</small></span>,
      children: <div className="bf11-asset-editor-stack">
        <AssetPromptGroup label="人物 Prompt" type="character" items={bookAssets.characters || []} drafts={assetDrafts} generatedImages={generatedAssetImages.character || {}} onChange={setAssetDraft} onRefresh={() => onRefreshAssets?.(selectedBook)} onSave={(type, items, drafts) => onSaveAssetPrompts?.(selectedBook, type, items, drafts)} onGenerateImage={(item, prompt) => onGenerateAssetImage?.(selectedBook, "character", item, prompt)} />
        <AssetPromptGroup label="场景 Prompt" type="scene" items={bookAssets.scenes || []} drafts={assetDrafts} generatedImages={generatedAssetImages.scene || {}} onChange={setAssetDraft} onRefresh={() => onRefreshAssets?.(selectedBook)} onSave={(type, items, drafts) => onSaveAssetPrompts?.(selectedBook, type, items, drafts)} onGenerateImage={(item, prompt) => onGenerateAssetImage?.(selectedBook, "scene", item, prompt)} />
      </div>
    },
    {
      key: 'constraints',
      label: <span className="bf11-fold-label"><strong>生成约束</strong><small>当前小说的继承与覆盖</small></span>,
      children: <div className="bf11-inline-editor">
        <div className="bf11-inline-editor-title"><Typography.Text strong>生成约束</Typography.Text><Tag>{selectedBook.overrideCount || 0} 项当前小说覆盖</Tag></div>
        <Typography.Paragraph type="secondary">
          画幅、固定单 VIDEO、画质、限制词与负面提示词按“系统 → 批次 → 当前小说 → 单 VIDEO”解析；这里只编辑当前小说层，不会改动其他小说。
        </Typography.Paragraph>
        <Space wrap>
          <Button
            icon={<SlidersHorizontal size={14} />}
            disabled={overrideAction.disabled || !onOpenBookSettings}
            title={overrideAction.disabled ? overrideAction.reason : ''}
            onClick={() => onOpenBookSettings?.(selectedBook)}
          >编辑当前小说约束</Button>
          <Typography.Text type="secondary">实际生效值与兼容性由 Go 在提交前校验。</Typography.Text>
        </Space>
      </div>
    },
    {
      key: 'visual-prompts',
      label: <span className="bf11-fold-label"><strong>画面提示词</strong><small>{selectedVideos.length} 个 VIDEO · 可逐条覆盖</small></span>,
      children: <div className="bf11-video-plan">
        <DirectorPanel
          book={selectedBook}
          capabilities={capabilities}
          onRunDirector={onRunDirector}
        />
        {selectedVideo && selectedShots.length ? <div className="bf11-shot-switcher" data-bf-control="shot-switcher">
          <Button size="small" disabled={currentShotIndex <= 0} onClick={() => setCurrentShotIndex(index => Math.max(0, index - 1))}>上一分镜</Button>
          <Tag color="blue">Shot {currentShotIndex + 1} / {selectedShots.length}</Tag>
          <Typography.Text strong>{selectedShot?.label || selectedShot?.id || '当前分镜'}</Typography.Text>
          <Button size="small" disabled={currentShotIndex >= selectedShots.length - 1} onClick={() => setCurrentShotIndex(index => Math.min(selectedShots.length - 1, index + 1))}>下一分镜</Button>
        </div> : null}
        {selectedShot ? <div className="bf11-shot-context" data-bf-region="active-shot">
          <Space wrap size={[6, 6]}>
            <Tag>当前 Shot：{selectedShot.label || selectedShot.id}</Tag>
            {selectedShot.startSeconds !== undefined && selectedShot.endSeconds !== undefined
              ? <Tag>时间轴 {selectedShot.startSeconds}s – {selectedShot.endSeconds}s</Tag>
              : null}
            <Tag color={statusTone[productionTaskByShotId.get(selectedShot.id)?.status || selectedShot.status]}>
              {productionTaskByShotId.get(selectedShot.id)?.status || selectedShot.status || '待生成'}
            </Tag>
          </Space>
          <Typography.Paragraph type="secondary">
            人物：{(selectedShot.characterRefs || []).join('、') || '沿用当前小说资产'}
            {' · '}场景：{(selectedShot.sceneRefs || []).join('、') || '沿用当前小说资产'}
            {' · '}道具：{(selectedShot.propRefs || []).join('、') || '无'}
          </Typography.Paragraph>
          {productionTaskByShotId.get(selectedShot.id) ? <Typography.Text type="secondary">
            任务：第 {productionTaskByShotId.get(selectedShot.id).attempt || 1} 次尝试
            {productionTaskByShotId.get(selectedShot.id).provider ? ` · ${productionTaskByShotId.get(selectedShot.id).provider}` : ''}
            {productionTaskByShotId.get(selectedShot.id).providerTaskId ? ` · ${productionTaskByShotId.get(selectedShot.id).providerTaskId}` : ''}
            {productionTaskByShotId.get(selectedShot.id).errorMessage ? ` · ${productionTaskByShotId.get(selectedShot.id).errorMessage}` : ''}
          </Typography.Text> : null}
          <div className="bf11-shot-media-grid">
            <div className="bf11-shot-media-slot">
              <Typography.Text strong>画面图</Typography.Text>
              {(generatedShotImages[selectedShot.id] || selectedShot.visualImageUrl)
                ? <img src={generatedShotImages[selectedShot.id] || selectedShot.visualImageUrl} alt={`Shot ${selectedShot.id} 画面图`} />
                : <Typography.Text type="secondary">尚未生成画面图</Typography.Text>}
              <Button size="small" disabled={!onGenerateShotImage || !selectedShot.visualPrompt} onClick={() => onGenerateShotImage?.(selectedBook, selectedVideo, selectedShot)}>使用 API 图片模型生成</Button>
            </div>
            <div className="bf11-shot-media-slot">
              <Typography.Text strong>视频结果</Typography.Text>
              {selectedShot.videoUrl || productionTaskByShotId.get(selectedShot.id)?.mediaUrl
                ? <video controls preload="metadata" src={selectedShot.videoUrl || productionTaskByShotId.get(selectedShot.id)?.mediaUrl} />
                : <Typography.Text type="secondary">尚未生成视频</Typography.Text>}
            </div>
          </div>
        </div> : null}
        {selectedVideo ? <div className="bf11-video-detail">
          <div className="bf11-video-detail-head">
            <Space wrap>
              <strong>{selectedVideo.label || '当前 VIDEO'}</strong>
              {selectedVideo.duration ? <Tag>{selectedVideo.duration}s</Tag> : null}
              <Tag color={statusTone[selectedVideo.status]}>{selectedVideo.status || '未返回状态'}</Tag>
            </Space>
            <Button
              size="small"
              icon={<SlidersHorizontal size={14} />}
              disabled={overrideAction.disabled || !onOpenVideoSettings}
              title={overrideAction.disabled ? overrideAction.reason : ''}
              onClick={() => onOpenVideoSettings?.(selectedBook, selectedVideo)}
            >VIDEO 约束</Button>
          </div>
          <div className="bf11-inline-editor">
            <div className="bf11-inline-editor-title"><Typography.Text strong>画面提示词</Typography.Text><Tag>visualPrompt</Tag></div>
            <TextArea
              rows={6}
              value={videoPromptDrafts[selectedShot?.id || selectedVideo.id] ?? selectedShot?.visualPrompt ?? selectedVideo.visualPrompt ?? ''}
              placeholder="当前服务端未返回当前 Shot 画面提示词"
              onChange={event => setVideoPromptDrafts(current => ({ ...current, [selectedShot?.id || selectedVideo.id]: event.target.value }))}
            />
            <Space wrap>
              <Button disabled={!onSaveVideoPrompt} onClick={() => onSaveVideoPrompt?.(selectedBook, selectedVideo, videoPromptDrafts[selectedShot?.id || selectedVideo.id] ?? selectedShot?.visualPrompt ?? selectedVideo.visualPrompt ?? '', selectedShot)}>保存画面提示词</Button>
              <Button
                disabled={compilerAction.disabled || !onPreviewFinalPrompt}
                title={compilerAction.disabled ? compilerAction.reason : ''}
                onClick={() => onPreviewFinalPrompt?.(selectedBook, selectedVideo, selectedShot)}
              >本次提交预览</Button>
              <Typography.Text type="secondary">保存后只覆盖当前 VIDEO；预览展示 Go 编译后的最终提示词。</Typography.Text>
            </Space>
          </div>
        </div> : <Typography.Text type="secondary">先运行 Director，才能生成可编辑的 VIDEO 画面提示词。</Typography.Text>}
      </div>
    },
    {
      key: 'video-cards',
      label: <span className="bf11-fold-label"><strong>VIDEO 卡片</strong><small>{selectedVideos.length} 条分镜生产卡</small></span>,
      children: <div className="bf11-video-plan">
        <div className="bf11-video-list">
          {selectedVideos.map((video, index) => <button
            key={video.id}
            className={selectedVideo?.id === video.id ? 'is-selected' : ''}
            onClick={() => { setSelectedVideoId(video.id); setPreviewTarget(video.id); }}
          >
            <span><Video size={14} /> {video.label || `VIDEO ${String(index + 1).padStart(2, '0')}`}</span>
            <small>{video.duration ? `${video.duration}s` : '—'}</small>
            <Tag color={statusTone[video.status]}>{video.status || '未返回状态'}</Tag>
          </button>)}
        </div>
        {selectedVideo ? <div className="bf11-video-detail">
          <div className="bf11-video-detail-head">
            <Space wrap>
              <strong>{selectedVideo.label || '当前 VIDEO'}</strong>
              {selectedVideo.duration ? <Tag>{selectedVideo.duration}s</Tag> : null}
              <Tag color={statusTone[selectedVideo.status]}>{selectedVideo.status || '未返回状态'}</Tag>
            </Space>
            <Button
              size="small"
              icon={<SlidersHorizontal size={14} />}
              disabled={overrideAction.disabled || !onOpenVideoSettings}
              title={overrideAction.disabled ? overrideAction.reason : ''}
              onClick={() => onOpenVideoSettings?.(selectedBook, selectedVideo)}
            >编辑 VIDEO</Button>
          </div>
          <div className="bf11-assets compact">
            <div><b>人物</b>{(selectedVideo.characters || []).map(item => <Tag key={assetName(item)}>{assetName(item)}</Tag>)}</div>
            <div><b>场景</b>{(selectedVideo.scenes || []).map(item => <Tag key={assetName(item)}>{assetName(item)}</Tag>)}</div>
          </div>
          <Space wrap>
            <Button
              disabled={productionAction.disabled || !onRunProduction}
              title={productionAction.disabled ? productionAction.reason : ''}
              onClick={() => onRunProduction?.(batch, selectedBook)}
            >生成当前小说</Button>
            {selectedShot && productionTaskByShotId.get(selectedShot.id)?.status === 'failed' ? <Button
              danger
              disabled={productionAction.disabled || !onRunProduction}
              title={productionAction.disabled ? productionAction.reason : ''}
              onClick={() => onRunProduction?.(batch, selectedBook)}
            >重试失败 Shot</Button> : null}
            <Typography.Text type="secondary">提交只会生产当前小说尚未完成的 Shot；成功 Shot 不会重复生成。</Typography.Text>
          </Space>
        </div> : <Typography.Text type="secondary">当前小说没有 VIDEO 数据。</Typography.Text>}
      </div>
    }
  ] : [];

  const cards = [
    {
      id: 'book-list',
      title: '小说列表',
      subtitle: `${filteredBooks.length}/${books.length} 本`,
      content: <div data-bf-card="book-list" className="bf11-book-list-body">
        <Input
          allowClear
          prefix={<Search size={14} />}
          value={bookSearch}
          onChange={event => setBookSearch(event.target.value)}
          placeholder="搜索书名 / Book ID / 平台"
        />
        <div className="bf11-book-list">
          {filteredBooks.map(book => <button
            key={book.id}
            className={`bf11-book-row ${selectedBook?.id === book.id ? 'is-selected' : ''} ${book.status === '异常' ? 'is-error' : ''}`}
            onClick={() => selectBook(book)}
          >
            <span className="bf11-book-index">{String(books.indexOf(book) + 1).padStart(2, '0')}</span>
            <span className="bf11-book-copy">
              <strong>{book.title || book.bookId || book.id}</strong>
              <small>{book.platform || '未知来源'} · {book.bookId || book.id}</small>
            </span>
            <span className="bf11-book-meta">
              <span
                className="bf11-book-status"
                style={{ backgroundColor: statusColor[book.status] || '#94a3b8' }}
                aria-label={`状态：${book.status || '未返回状态'}`}
                title={book.status || '未返回状态'}
              />
            </span>
          </button>)}
          {!filteredBooks.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="没有匹配的小说" /> : null}
        </div>
      </div>
    },
    {
      id: 'book-workbench',
      title: '当前小说工作台',
      subtitle: selectedBook ? `${selectedBook.platform || '未知来源'} · Book ID ${selectedBook.bookId || selectedBook.id}` : '请选择小说',
      content: <div data-bf-card="book-workbench" className="bf11-book-workbench-body">
        {selectedBook ? <>
          <div className="bf11-current-book-head">
            <div>
              <Space wrap size={6}>
                <Typography.Title level={4}>{selectedBook.title || selectedBook.bookId || selectedBook.id}</Typography.Title>
                <Tag color={statusTone[selectedBook.status]}>{selectedBook.status || '未返回状态'}</Tag>
                {selectedBook.overrideCount ? <Tag color="purple">当前小说已覆盖 {selectedBook.overrideCount} 项</Tag> : <Tag>当前层无覆盖</Tag>}
              </Space>
            </div>
            <Button
              icon={<Settings2 size={15} />}
              disabled={overrideAction.disabled || !onOpenBookSettings}
              title={overrideAction.disabled ? overrideAction.reason : ''}
              onClick={() => onOpenBookSettings?.(selectedBook)}
            >当前小说设置</Button>
          </div>
          <Collapse
            className="bf11-book-collapse"
            defaultActiveKey={['source', 'assets', 'constraints', 'visual-prompts', 'video-cards']}
            items={bookCollapseItems}
          />
        </> : <Empty description="请选择小说" />}
      </div>
    },
    {
      id: 'preview',
      title: '成片预览 / 合并成品',
      subtitle: '全工作台共用一个播放器',
      content: <div data-bf-card="preview" className="bf11-preview-body">
        <div className="bf11-preview-tabs">
          <button className={previewTarget === 'merged' ? 'is-active' : ''} onClick={() => setPreviewTarget('merged')}>最终合并</button>
          {selectedVideos.map((video, index) => <button key={video.id} className={previewTarget === video.id ? 'is-active' : ''} onClick={() => { setPreviewTarget(video.id); setSelectedVideoId(video.id); }}>{video.label || `VIDEO ${String(index + 1).padStart(2, '0')}`}</button>)}
        </div>
        <div className="bf11-unified-player" data-bf-player="unified">
          {(() => {
            const latestMerge = (mergeStatus?.jobs || []).slice(-1)[0];
            const selectedPreviewVideo = selectedVideos.find(video => video.id === previewTarget);
            const previewTask = selectedPreviewVideo ? productionTaskByVideoId.get(selectedPreviewVideo.id) : null;
            const activeShotTask = selectedShot ? productionTaskByShotId.get(selectedShot.id) : null;
            const mediaURL = previewTarget === 'merged'
              ? (selectedBook?.mergedUrl || latestMerge?.outputUrl || '')
              : (selectedShot?.videoUrl || selectedShot?.mediaUrl || activeShotTask?.mediaUrl
                || selectedPreviewVideo?.url || selectedPreviewVideo?.mediaUrl || previewTask?.mediaUrl || '');
            return mediaURL ? <video controls preload="metadata" src={mediaURL} style={{ width: '100%', maxHeight: 360 }} /> : <Film size={42} />;
          })()}
          <strong>{previewTarget === 'merged'
            ? (selectedBook?.mergedFileName || '暂无合并成品')
            : (selectedVideos.find(video => video.id === previewTarget)?.label || '当前 VIDEO')}</strong>
          <span>{previewTarget === 'merged' ? '最终合并成品统一在这里预览' : '原始 VIDEO 统一切换到同一个播放器预览'}</span>
          <div className="bf11-player-track"><i /></div>
          <small>{selectedBook?.mergedUrl
            || (mergeStatus?.jobs || []).some(job => job.outputUrl)
            || selectedVideos.find(video => video.id === previewTarget)?.url
            || selectedVideos.find(video => video.id === previewTarget)?.mediaUrl
            || productionTaskByVideoId.get(previewTarget)?.mediaUrl
            ? '媒体地址已返回，可直接播放'
            : '当前暂无可播放媒体地址'}</small>
        </div>
      </div>
    },
    {
      id: 'batch-tools',
      title: '批量工具',
      subtitle: '批量优先 · 异常优先',
      content: <div data-bf-card="batch-tools" className="bf11-batch-tools-body">
        <section className="bf11-tool-section">
          <div className="bf11-tool-section-head"><strong>视频生成进度</strong><Tag>{videoProgress.completed + videoProgress.failed}/{videoProgress.total}</Tag></div>
          <Progress percent={videoProgress.percent} status={videoProgress.failed ? 'exception' : 'active'} />
          <Space wrap size={[4, 4]}>
            <Tag>待生成 {videoProgress.pending}</Tag>
            <Tag color="blue">排队中 {videoProgress.queued}</Tag>
            <Tag color="processing">生成中 {videoProgress.generating}</Tag>
            <Tag color="green">已完成 {videoProgress.completed}</Tag>
            <Tag color="red">失败 {videoProgress.failed}</Tag>
          </Space>
          <div className="bf11-tool-buttons">
            <Button disabled={!onRunBatchDirector} onClick={() => onRunBatchDirector?.(batch)} title={!onRunBatchDirector ? '等待 Director 批量接线' : ''}>批量 Director</Button>
            <Button
              disabled={productionAction.disabled || !onRunProduction}
              title={productionAction.disabled ? productionAction.reason : ''}
              onClick={() => onRunProduction?.(batch)}
            >生成待生成</Button>
          </div>
        </section>

        <section className="bf11-tool-section">
          <div className="bf11-tool-section-head"><strong>批量合并</strong><Tag color="cyan">按导演顺序</Tag></div>
          <Segmented
            block
            value={mergeTimingMode}
            onChange={setMergeTimingMode}
            options={[
              { value: 'speed', label: '固定倍率' },
              { value: 'audio', label: '跟随音频时长' }
            ]}
          />
          {mergeTimingMode === 'speed' ? <Select
            value={mergeSpeed}
            onChange={setMergeSpeed}
            options={[
              { value: '1.0', label: '1.0x' },
              { value: '1.2', label: '1.2x' },
              { value: '1.5', label: '1.5x' },
              { value: '2.0', label: '2.0x' }
            ]}
          /> : <div className="bf11-tts-timing">
            <div><Typography.Text strong>TTS 测时语速</Typography.Text><InputNumber min={0.5} max={3} step={0.1} value={ttsSpeed} onChange={value => setTtsSpeed(Number(value || 1.7))} addonAfter="x" /></div>
            <Typography.Text type="secondary">TTS 只测时，不合入成片；默认 1.7 是 TTS 参数，不是固定视频倍率。</Typography.Text>
          </div>}
          <Button
            disabled={mergeAction.disabled || !onRunMerge}
            title={mergeAction.disabled ? mergeAction.reason : ''}
            onClick={() => onRunMerge?.(batch, { bookId: selectedBook?.id || '', timingMode: mergeTimingMode, speed: mergeSpeed, ttsSpeed })}
          >合并待合并（当前小说）</Button>
        </section>

        <section className="bf11-tool-section">
          <div className="bf11-tool-section-head"><strong>批量上传</strong><Tag>确认后外发</Tag></div>
          <Button icon={<Upload size={14} />} disabled={!onRunUpload} onClick={() => onRunUpload?.(batch)}>上传待上传</Button>
          <Typography.Text type="secondary">点击后选择 121/Yadi；先保存加密账号，再生成确认单，最后由你明确确认提交。</Typography.Text>
        </section>
      </div>
    }
  ];

  const batchPatch = batch?.settingsState?.patch || {};
  const mode = batch?.mode || batchPatch.productionMode;
  const modeLabel = mode ? (mode === 'viral' || mode === 'viral_hook' ? '爆款开头' : '原文直转') : '生产方式未覆盖';
  const provider = batchPatch.videoProvider || 'personal_api';
  const providerLabel = provider === 'doubao_local_executor'
    ? '豆包本地执行器'
    : provider === 'autodl_comfyui' ? 'AutoDL · MiniMax H3' : '个人中心 API · yd2.0-mini';
  const batchSubtitle = [modeLabel, providerLabel, batchPatch.videoModelId, batchPatch.aspectRatio, batchPatch.versionConfigId].filter(Boolean).join(' · ');

  return <div className="batch-factory-workbench bf11-workbench">
    <section className="bf11-batch-header" data-bf-region="batch-header">
      <div className="bf11-batch-title">
        <Space wrap size={8}>
          <Typography.Title level={3}>{batch?.title || 'Batch Factory V11'}</Typography.Title>
          <Tag>{Number(batch?.count ?? books.length)} 本</Tag>
          {batch?.status ? <Tag color="processing">{batch.status}</Tag> : null}
        </Space>
        <Typography.Text type="secondary">{batchSubtitle || '等待服务端生产配置'}</Typography.Text>
      </div>
      <Space wrap>
        <Button
          icon={<FolderPlus size={15} />}
          disabled={batchCreateAction.disabled || !onOpenBatchManager}
          title={batchCreateAction.disabled ? batchCreateAction.reason : ''}
          onClick={() => onOpenBatchManager?.('new')}
        >+ 新建批次</Button>
        <Button
          icon={<Archive size={15} />}
          disabled={batchReadAction.disabled || !onOpenHistory}
          title={batchReadAction.disabled ? batchReadAction.reason : ''}
          onClick={() => onOpenHistory?.()}
        >历史批次</Button>
        <Button
          icon={<Settings2 size={15} />}
          disabled={settingsAction.disabled || !batch || !onOpenBatchSettings}
          title={settingsAction.disabled ? settingsAction.reason : ''}
          onClick={() => onOpenBatchSettings?.(batch)}
        >生产统一设置</Button>
        <Button disabled={!onOpenPublishSettings} onClick={() => onOpenPublishSettings?.(batch)}>发布统一设置</Button>
        {!editMode ? <Button icon={<LayoutGrid size={15} />} onClick={() => setEditMode(true)}>编辑布局</Button> : <>
          <Button type="primary" icon={<Save size={15} />} onClick={saveLayout}>保存布局</Button>
          <Button icon={<RotateCcw size={15} />} onClick={restoreDefault}>恢复默认布局</Button>
          <Button onClick={restoreHiddenCards}>显示隐藏卡片</Button>
        </>}
      </Space>
    </section>

    {unavailableCapabilities.length ? <Alert
      className="bf11-capability-alert"
      type="warning"
      showIcon
      message="当前批量工厂还有未开放能力"
      description={<Space direction="vertical" size={2}>{unavailableCapabilities.map(item => <Typography.Text key={item.key}>· {item.label}：{item.reason}</Typography.Text>)}</Space>}
    /> : null}

    <section className="bf11-status-row">
      <div className="bf11-status-center" data-bf-region="status-center">
        <div className="bf11-region-heading"><strong>批次状态中心</strong><span>只负责筛选入口</span></div>
        <div className="bf11-status-cards">
          {STATUS_ORDER.map(status => <button key={status} className={statusFilter === status ? 'is-active' : ''} onClick={() => setStatusFilter(status)}>
            <span>{status}</span><strong>{statusCounts[status] || 0}</strong>
          </button>)}
        </div>
      </div>
      <div className="bf11-current-filter" data-bf-region="current-filter">
        <div className="bf11-region-heading"><strong>当前筛选</strong><span>{statusFilter} · {filterResults.length} 本</span></div>
        <div className="bf11-filter-results">
          {filterResults.slice(0, 5).map(book => <button key={book.id} onClick={() => selectBook(book)}>
            {book.status === '异常' ? <AlertTriangle size={14} /> : <span className="bf11-filter-dot" />}
            <span>{book.title || book.bookId || book.id}</span>
            <Tag color={statusTone[book.status]}>{book.status || '未返回状态'}</Tag>
          </button>)}
          {filterResults.length > 5 ? <small>还有 {filterResults.length - 5} 本；小说总列表保持完整，点击这里只切换当前小说。</small> : null}
          {!filterResults.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="没有符合状态的小说" /> : null}
        </div>
      </div>
    </section>

    <div
      className={`bf11-workspace-grid ${editMode ? 'is-editing' : ''}`}
      ref={gridRef}
      style={{
        '--bf11-row-height': `${layout.rowHeight}px`,
        '--bf11-grid-gap': `${layout.gap}px`
      }}
    >
      {cards.map(card => <WorkbenchCard
        key={card.id}
        id={card.id}
        title={card.title}
        subtitle={card.subtitle}
        item={layout.items[card.id]}
        editMode={editMode}
        maximized={layout.maximizedId === card.id}
        gridMetrics={metrics}
        onMove={(x, y) => updateCard(card.id, (current, id) => moveWorkspaceItem(current, id, x, y))}
        onResize={(w, h) => updateCard(card.id, (current, id) => resizeWorkspaceItem(current, id, w, h))}
        onCollapse={collapsed => updateCard(card.id, (current, id) => setWorkspaceItemCollapsed(current, id, collapsed))}
        onHide={() => updateCard(card.id, (current, id) => setWorkspaceItemHidden(current, id, true))}
        onMaximize={maximized => updateCard(card.id, (current, id) => setItemMaximized(current, id, maximized))}
      >{card.content}</WorkbenchCard>)}
    </div>
  </div>;
}

export default BatchFactoryV11Workbench;
