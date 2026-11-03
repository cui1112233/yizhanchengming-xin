import { useEffect, useState } from 'react';
import { Alert, Button, Drawer, Input, Progress, Select, Space, Typography } from 'antd';
import { Download, ExternalLink, Pencil, RefreshCw, Settings2, WandSparkles } from 'lucide-react';

import { getComicPageImageBlob } from '../../shared/api/batchFactory';
import { BatchFactoryComicStylePicker } from './shuihuo/BatchFactoryComicStylePicker';
import { privateComicMinSourceLength, privateComicTitle } from './privateCustomizationComic';

const COMIC_DIGITS = ['零', '一', '二', '三', '四', '五', '六', '七', '八', '九'];

function text(value) {
  return String(value ?? '').trim();
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
  return '漫画生成中';
}

function ComicPageCard({ page, index, taskReady, action, actionKey, onGeneratePage, onSavePagePrompt }) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState({ comicPrompt: '', imagePrompt: '' });
  const [display, setDisplay] = useState({ url: '', failed: false });
  const images = Array.isArray(page?.images) ? page.images : [];
  const primary = images.find(image => image?.isPrimary) || images[0] || null;
  const generating = action === 'pages' && actionKey === page?.videoId;
  const saving = action === 'save' && actionKey === page?.videoId;
  const busyElsewhere = Boolean(action) && !generating && !saving;
  const hasPrompt = Boolean(text(page?.comicPrompt) || text(page?.imagePrompt));
  const canSave = Boolean(text(draft.comicPrompt) && text(draft.imagePrompt));

  useEffect(() => {
    let active = true;
    let ownedUrl = '';
    setDisplay({ url: '', failed: false });
    if (!primary?.url) return () => {};
    getComicPageImageBlob(primary.url).then(blob => {
      if (!active) return;
      ownedUrl = URL.createObjectURL(blob);
      setDisplay({ url: ownedUrl, failed: false });
    }).catch(() => { if (active) setDisplay({ url: '', failed: true }); });
    return () => {
      active = false;
      if (ownedUrl) URL.revokeObjectURL(ownedUrl);
    };
  }, [primary?.url]);

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
      <Space size={8} wrap>
        <Button size="small" icon={<Pencil size={15} aria-hidden="true" />} onClick={editing ? () => setEditing(false) : beginEdit}>{editing ? '取消编辑' : '编辑提示词'}</Button>
        <Button
          size="small"
          type="primary"
          icon={<WandSparkles size={15} aria-hidden="true" />}
          loading={generating}
          disabled={!taskReady || !hasPrompt || busyElsewhere || generating}
          title={!taskReady ? '请先开始生成漫画创建任务' : !hasPrompt ? '该页缺少提示词' : busyElsewhere ? '当前有其他操作进行中' : ''}
          onClick={() => onGeneratePage?.(page)}
        >{images.length ? '重新生成图' : '生成图'}</Button>
        {display.url ? <Button size="small" icon={<Download size={15} aria-hidden="true" />} href={display.url} download={`漫画${comicPageNumber(index + 1)}.png`}>下载本页</Button> : null}
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
      {display.url ? <img src={display.url} alt={`漫画第${index + 1}页`} /> : <div className="private-comic-page-placeholder">
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
  </div>;
}

function ComicOutputCards({ pages, taskReady, action, actionKey, onGeneratePage, onSavePagePrompt, onDownloadAll }) {
  const pageList = Array.isArray(pages) ? pages : [];
  const imagedPages = pageList.filter(page => (Array.isArray(page?.images) ? page.images : []).length > 0).length;
  return <div className="shot-output-cards">
    <div className="shot-output-toolbar">
      <span>共 {pageList.length} 页 · 已出图 {imagedPages} 页</span>
      <Space size={8}>
        <Button size="small" icon={<Download size={15} aria-hidden="true" />} loading={action === 'download'} disabled={!pageList.length || (action && action !== 'download')} onClick={onDownloadAll}>下载全部</Button>
      </Space>
    </div>
    {pageList.map((page, index) => <ComicPageCard
      key={page?.videoId || page?.id || index}
      page={page}
      index={index}
      taskReady={taskReady}
      action={action}
      actionKey={actionKey}
      onGeneratePage={onGeneratePage}
      onSavePagePrompt={onSavePagePrompt}
    />)}
  </div>;
}

export function PrivateComicWorkspace({
  catalog = [],
  modelOptions = {},
  draft = {},
  comicAssets = [],
  comicAssetImages = {},
  comicAssetImageFailures = [],
  comicPages = [],
  action = '',
  actionKey = '',
  onExtractAssets,
  onGenerateAssetImage,
  onGenerateMissingAssetImages,
  onGeneratePrompts,
  onGeneratePageImage,
  onGenerateAllPageImages,
  onSavePagePrompt,
  onDownloadAll,
  onDraftChange,
  onSubmit,
  task,
  busy = false,
  error = '',
  presetReady = true,
  onRefresh,
  onOpenBatchFactory,
  onOpenUnifiedConfig,
  onReloadCatalog,
  onRetry,
  onPendingReferenceChange
}) {
  const [settingsOpen, setSettingsOpen] = useState(false);
  const comic = draft.comic && typeof draft.comic === 'object' ? draft.comic : {};
  const patchDraft = patch => onDraftChange?.({ ...draft, ...patch });
  const patchComic = patch => patchDraft({ comic: { ...comic, ...patch, enabled: true } });
  const pageOptions = optionsFor(catalog, 'batch.comic-page');
  const derivationOptions = optionsFor(catalog, 'batch.comic-image-derivation');
  const promptName = text(comic.presetName) || text(pageOptions.find(option => option.value === comic.presetId)?.item?.name);
  const derivationName = text(comic.imageDerivation?.presetName) || text(derivationOptions.find(option => option.value === comic.imageDerivation?.presetId)?.item?.name);
  const constraintCount = (comic.constraints || []).filter(item => item?.presetId).length;

  const assetList = Array.isArray(comicAssets) ? comicAssets : [];
  const pageList = Array.isArray(comicPages) ? comicPages : [];
  const assetImagesFor = asset => Array.isArray(comicAssetImages?.[asset?.id]) ? comicAssetImages[asset.id] : [];
  const assetsForKind = kind => assetList.filter(asset => text(asset?.kind) === kind);
  const assetImageReady = asset => assetImagesFor(asset).length > 0;
  const assetsWithImages = assetList.filter(assetImageReady).length;
  const missingAssetImages = Math.max(0, assetList.length - assetsWithImages);
  const promptPages = pageList.filter(page => text(page?.comicPrompt) || text(page?.imagePrompt));
  const imagedPages = pageList.filter(page => (Array.isArray(page?.images) ? page.images : []).length > 0);
  const assetSections = [
    { key: 'character', label: '人物', empty: '提取资产后显示人物信息' },
    { key: 'scene', label: '场景', empty: '提取资产后显示场景信息' },
    { key: 'prop', label: '道具', empty: '提取资产后显示道具信息' }
  ];
  const pipelineSteps = [
    { key: 'assets', label: '提取资产', done: assetList.length > 0, detail: assetList.length ? `已完成 ${assetList.length} 项` : '未提取' },
    { key: 'assetImages', label: '资产图', done: assetList.length > 0 && assetsWithImages >= assetList.length, detail: assetList.length ? `已生成 ${assetsWithImages}/${assetList.length}` : '待资产' },
    { key: 'prompts', label: '分镜提示词', done: promptPages.length > 0, detail: promptPages.length ? `已生成 ${promptPages.length} 页` : (assetList.length ? '待生成' : '待资产') },
    { key: 'images', label: '分镜生图', done: pageList.length > 0 && imagedPages.length >= pageList.length, detail: pageList.length ? `已出图 ${imagedPages.length}/${pageList.length}` : '待提示词' }
  ];
  const currentStepKey = pipelineSteps.find(step => !step.done)?.key || '';
  const failureIds = Array.isArray(comicAssetImageFailures) ? comicAssetImageFailures : [];
  const assetImageFailed = asset => failureIds.includes(asset?.id);
  const actionBusy = Boolean(action);
  const taskReady = Boolean(task?.batchId && task?.bookId);
  const canExtract = taskReady && !actionBusy && !busy;
  const canGenerateAssetImages = canExtract && missingAssetImages > 0;
  const canGeneratePrompts = canExtract && assetList.length > 0;
  const canGeneratePages = canExtract && promptPages.length > 0;
  const missingImageCount = Math.max(0, pageList.length - imagedPages.length);
  const allPromptsReady = pageList.length > 0 && promptPages.length >= pageList.length;
  const promptButtonLabel = action === 'prompts'
    ? '生成中…'
    : !assetList.length ? '生成分镜提示词'
      : allPromptsReady ? '重新生成分镜提示词' : '补齐分镜提示词';
  const promptButtonTitle = !taskReady
    ? '请先开始生成漫画创建任务'
    : !assetList.length ? '需先提取资产' : actionBusy ? '当前有其他操作进行中' : '';
  const sourceLength = Array.from(text(draft.sourceText)).length;
  const sourceReady = sourceLength >= privateComicMinSourceLength;
  const progress = Number.isFinite(Number(task?.progress)) ? Math.max(0, Math.min(100, Number(task.progress))) : null;

  return <section id="private-customization-panel-comic" className="script-workbench private-comic-workbench" role="tabpanel" aria-labelledby="private-customization-tab-comic">
    <div className="script-left">
      <div className="script-left-scroll">
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
              rows={10}
              value={draft.sourceText || ''}
              placeholder="把要生成漫画的正文直接粘贴到这里"
              disabled={busy}
              onChange={event => patchDraft({ sourceText: event.target.value })}
            />
            {text(draft.sourceText) ? (
              <button
                type="button"
                className="script-source-clear"
                aria-label="清空漫画正文"
                title="清空漫画正文"
                onClick={() => patchDraft({ sourceText: '' })}
              >×</button>
            ) : null}
          </div>
          <div className="script-chat-options">
            <span className="private-comic-source-count">正文 {sourceLength} 字（至少 {privateComicMinSourceLength} 字）</span>
            <Space size={8}>
              <Button
                type="text"
                aria-label="刷新资产与分镜"
                title="刷新资产与分镜"
                icon={<RefreshCw size={16} strokeWidth={1.8} aria-hidden="true" />}
                disabled={!taskReady}
                onClick={onRefresh}
              />
              <Button
                className="script-chat-submit"
                type="primary"
                icon={<WandSparkles size={16} strokeWidth={1.8} aria-hidden="true" />}
                loading={action === 'assets'}
                disabled={!canExtract}
                title={!taskReady ? '请先开始生成漫画创建任务' : actionBusy ? '当前有其他操作进行中' : ''}
                onClick={onExtractAssets}
              >{action === 'assets' ? '提取中…' : assetList.length ? '重新提取资产' : '提取资产'}</Button>
            </Space>
          </div>
        </div>

        <div className="script-instruction-status private-comic-instruction">
          <span className="private-comic-instruction-label">当前漫画指令：</span>
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
                title={!taskReady ? '请先开始生成漫画创建任务' : !assetList.length ? '需先提取资产' : !missingAssetImages ? '资产图已全部生成' : actionBusy ? '当前有其他操作进行中' : ''}
                onClick={onGenerateMissingAssetImages}
              >{action === 'images' ? '生成中…' : missingAssetImages ? `补齐资产图（缺 ${missingAssetImages}）` : '资产图已齐'}</Button>
            </Space>
          </div>
          <div className="entity-list">
            {assetSections.map(section => {
              const items = assetsForKind(section.key);
              return <div className="private-comic-asset-group" key={section.key}>
                <div className="private-comic-asset-group-title">
                  <b>{section.label}</b>
                  <span className="legacy-muted">{items.length}</span>
                </div>
                {items.length ? items.map(asset => <div className="entity-card-row" key={asset.id}>
                  <div className="entity-card private-comic-asset-card">
                    <span>{asset.name || '未命名资产'}</span>
                    <small className="entity-media-status">
                      {assetImageReady(asset)
                        ? `资产图已生成${assetImagesFor(asset).length > 1 ? ` · ${assetImagesFor(asset).length} 张` : ''}`
                        : assetImageFailed(asset) ? '资产图生成失败，可补齐' : '资产图未生成'}
                    </small>
                  </div>
                </div>) : <div className="entity-card">{section.empty}</div>}
              </div>;
            })}
          </div>
        </div>

        {error ? <Alert type="error" showIcon message={error} /> : null}
        {!presetReady ? <Alert
          type="warning"
          showIcon
          message={<span>尚未检测到批量工厂的自动化预设，请先创建预设再开始生成漫画。<Button type="link" size="small" href="/batch-factory" target="_self">去批量工厂创建预设</Button></span>}
        /> : null}
        <Typography.Text type="secondary" className="private-comic-hint">只提交当前这篇正文，不会加入批量工厂的其他书籍。</Typography.Text>
      </div>
    </div>

    <div className="script-right">
      <div className="script-tabs">
        <Typography.Text strong>漫画生成</Typography.Text>
        <Button
          type="text"
          icon={<Settings2 size={16} aria-hidden="true" />}
          onClick={onOpenUnifiedConfig}
          title="漫画生成统一配置"
        >统一配置</Button>
        <span className="private-customization-comic-no-publish">不自动发布</span>
        <div style={{ flex: 1 }} />
        {task?.bookId ? <span className="private-comic-task-meta">{task.bookId}</span> : null}
      </div>
      <div className="script-toolbar">
        <Select style={{ width: 140, flex: '0 0 140px' }} allowClear value={draft.textModelId || undefined} options={modelOptions.text || []} placeholder="文本模型" disabled={busy} onChange={textModelId => patchDraft({ textModelId: textModelId || '' })} />
        <Select style={{ width: 140, flex: '0 0 140px' }} allowClear value={draft.imageModelId || undefined} options={modelOptions.image || []} placeholder="图片模型" disabled={busy} onChange={imageModelId => patchDraft({ imageModelId: imageModelId || '' })} />
        <Button
          className="script-constraint-button"
          icon={<Settings2 size={16} strokeWidth={1.8} aria-hidden="true" />}
          onClick={() => setSettingsOpen(true)}
          title="约束设置"
        >约束设置</Button>
        <Space wrap>
          {presetReady
            ? <Button type="primary" icon={<WandSparkles size={16} strokeWidth={1.8} aria-hidden="true" />} loading={busy} disabled={busy || !sourceReady} onClick={onSubmit}>开始生成漫画</Button>
            : <Button type="primary" icon={<WandSparkles size={16} strokeWidth={1.8} aria-hidden="true" />} loading={busy} disabled>开始生成漫画</Button>}
          <Button
            icon={<WandSparkles size={16} strokeWidth={1.8} aria-hidden="true" />}
            loading={action === 'prompts'}
            disabled={!canGeneratePrompts}
            title={promptButtonTitle}
            onClick={onGeneratePrompts}
          >{promptButtonLabel}</Button>
          <Button
            type="primary"
            ghost
            icon={<WandSparkles size={16} strokeWidth={1.8} aria-hidden="true" />}
            loading={action === 'pages' && actionKey === ''}
            disabled={!canGeneratePages || (action === 'pages' && actionKey === '')}
            title={!taskReady ? '请先开始生成漫画创建任务' : !promptPages.length ? '需先生成分镜提示词' : actionBusy ? '当前有其他操作进行中' : ''}
            onClick={onGenerateAllPageImages}
          >{action === 'pages' && actionKey === '' ? '生成中…' : missingImageCount ? `批量出图（缺 ${missingImageCount} 页）` : '批量出图'}</Button>
          <Button icon={<RefreshCw size={16} strokeWidth={1.8} aria-hidden="true" />} disabled={!task?.batchId || busy} onClick={onRefresh}>刷新状态</Button>
          {task?.batchId ? <Button icon={<ExternalLink size={16} strokeWidth={1.8} aria-hidden="true" />} onClick={onOpenBatchFactory}>查看漫画详情</Button> : null}
          {task?.status === 'failed' ? <Button onClick={onRetry} disabled={busy}>重试/继续</Button> : null}
        </Space>
      </div>
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
          <div className="script-empty-copy">先在左侧提取资产，再点击「生成分镜提示词」；随后这里会按页显示提示词与图片。</div>
        </div> : <div className="script-empty legacy-panel-card">
          <div className="script-empty-card">
            <span className="script-empty-dot" aria-hidden="true" />
            <span className="script-empty-ray" aria-hidden="true" />
            <span className="script-empty-line script-empty-line--top" aria-hidden="true" />
            <span className="script-empty-line script-empty-line--bottom" aria-hidden="true" />
          </div>
          <div className="script-empty-title">准备生成漫画</div>
          <div className="script-empty-copy">在左侧粘贴正文，点击工具栏「开始生成漫画」</div>
        </div>}
      </div>
    </div>

    <Drawer
      title="漫画约束设置"
      open={settingsOpen}
      onClose={() => setSettingsOpen(false)}
      width={620}
      rootClassName="private-comic-settings-drawer"
    >
      <div className="private-customization-comic-settings">
        <div className="private-customization-settings-heading"><b>高级参数</b><span>画风、提示词与模型已放到左侧和工具栏，这里保留提示词配置与约束参数</span></div>
        <BatchFactoryComicStylePicker
          batchId={task?.batchId || ''}
          catalog={catalog}
          value={comic}
          disabled={busy}
          onReferenceDataUrl={onPendingReferenceChange}
          onChange={next => patchDraft({ comic: next })}
        />
        {!task?.batchId ? <Typography.Text type="secondary" className="private-customization-reference-note">提交时会先创建任务并上传自定义参考图，再开始生成。</Typography.Text> : null}
        <div className="private-customization-comic-select-grid private-comic-advanced-grid">
          <label className="private-customization-field"><span>漫画提示词</span><Select allowClear value={comic.presetId || undefined} options={pageOptions} placeholder="选择已发布漫画提示词" disabled={busy} onChange={value => patchComic(presetSelection(pageOptions.find(option => option.value === value)?.item))} /></label>
          <label className="private-customization-field"><span>图片描述词推导</span><Select allowClear value={comic.imageDerivation?.presetId || undefined} options={derivationOptions} placeholder="选择图片描述词推导" disabled={busy} onChange={value => patchComic({ imageDerivation: presetSelection(derivationOptions.find(option => option.value === value)?.item) })} /></label>
        </div>
        <small className="entity-media-status private-comic-advanced-note">{promptName ? `提示词「${promptName}」` : '提示词未选择'} · {derivationName ? `推导「${derivationName}」` : '推导未选择'} · 约束 {constraintCount ? `已配置 ${constraintCount} 项` : '未配置'}</small>
        <div className="private-customization-comic-select-grid">
          {['comic_quality', 'comic_restriction', 'comic_negative'].map(category => <label className="private-customization-field" key={category}><span>{category === 'comic_quality' ? '画质约束' : category === 'comic_restriction' ? '画面限制' : '负面提示词'}</span><Select allowClear value={selectedConstraint(comic, category) || undefined} options={constraintOptions(catalog, category)} placeholder="未选择" onChange={value => {
            const rest = (comic.constraints || []).filter(item => item?.constraintCategory !== category);
            const option = constraintOptions(catalog, category).find(item => item.value === value);
            patchComic({ constraints: value && option ? [...rest, { ...presetSelection(option.item), constraintCategory: category }] : rest });
          }} /></label>)}
        </div>
      </div>
    </Drawer>
  </section>;
}

export default PrivateComicWorkspace;
