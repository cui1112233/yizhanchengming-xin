import { h3PromptEditRequest } from './h3PromptEditing.js';
import { batchAccessErrorMessage } from './batchFactoryBatchAccessNotice.js';
import { readStoryboardPromptEditableText, storyboardPromptTextForSave } from './storyboardPromptEditor.js';
import { measureH3VideoLines } from './h3LineAudio.js';
import { buildBatchFactoryH3Constraints } from './batchFactoryH3Constraints.js';
import { prependSmartUnifiedStyle, smartUnifiedAnalysisForBook, smartUnifiedDisplayEnabled } from './batchFactorySmartUnified.js';
import { publicationMetadataValue } from './batchFactoryPublicationMetadata.js';
import { resolvePrecompiledStoryboardAssets, resolvePrecompiledStoryboardFrame, resolvePrecompiledVideoWorkspace } from './batchFactoryPrecompiledStoryboard.js';
import { batchFactoryRuntimePollDelay, comicRowResultPollDelay, finishComicPageGeneration, nextViewportLoaded } from './batchFactoryPerformance.js';
import { IMAGE_ASPECT_RATIO_OPTIONS, normalizeImageAspectRatio } from '../../../shared/imageAspectRatios';
import {
  ArrowLeftOutlined,
  BarsOutlined,
  CloudUploadOutlined,
  CloseOutlined,
	CopyOutlined,
	DownloadOutlined,
  FileTextOutlined,
  ReloadOutlined,
	LeftOutlined,
  PictureOutlined,
	RightOutlined,
  SettingOutlined,
  UploadOutlined
  ,DeleteOutlined
} from '@ant-design/icons';
import {
  Alert,
  Button,
  Card,
  Checkbox,
  Collapse,
  Descriptions,
  Dropdown,
  Input,
  InputNumber as NativeInputNumber,
  Image,
  Modal,
  Popconfirm,
  Select,
  Space,
  Switch,
  Tag,
  Tabs,
  Tooltip,
	Typography,
	Upload as AntUpload,
  message
} from 'antd';
import { useEffect, useMemo, useRef, useState } from 'react';

const InputNumber = NativeInputNumber;
import {
  createBookAsset,
	createNovelFetchIntake,
  appendNovelFetchIntake,
	backfillBatch121Metadata,
	  cancelBatchProduction,
  fetchBookOriginal,
  getCapabilities,
  getBatch,
  bindBatchFactoryAgentProject,
  getBookAssetImageBlob,
  getComicPageImageBlob,
	getComicReferencePreviewBlob,
	uploadComicGlobalReference,
  generateBookAssetImages,
  getBookAssetImageGenerationTask,
  downloadComicPromptZip,
	downloadComicCollectionImageZip,
  getComicPagePromptPreview,
  listComicPages,
	getComicSourceSegments,
	segmentComicSource,
	splitComicSourceSegment,
	setComicPageReferenceImages,
	applyComicPageReferenceImages,
  setPrimaryComicPageImage,
  createComicCollection,
  listComicCollections,
	getN8ComicPublishStatus,
	getN8ComicPublishAudioCategories,
  getN8ComicPublishAudios,
	searchN8ComicWorks,
	publishComicCollection,
	getComicCollectionPublishRecord,
  listBookAssets,
  listBookAssetImages,
  getConfigVersions,
  getFinalPrompt,
  getH3Trace,
	createDirectorPromptCandidate,
	getDirectorPromptCandidate,
	acceptDirectorPromptCandidate,
	cancelDirectorPromptCandidate,
	measureH3Audio,
	compileH3Video,
  getDraft,
  getBatchRuntimeSummary,
  getBatchAutomationStatus,
  listAutomationPresets,
  startBatchAutomation,
  pauseBatchAutomation,
  resumeBatchAutomation,
  retryBatchAutomation,
  cancelBatchAutomation,
  getChangeImpact,
  rewriteWorkingFront,
  listSystemPresetCatalog,
  runBatchDirector,
  runBookStage,
  retryBookStage,
  saveBatchSettings,
  saveBookOverride,
  updateBookAsset,
  updateBookMetadata,
  updateComicPage,
  updateBookSource,
  uploadBookAssetImage,
  uploadBookVideoMaterial,
  uploadBookMergeMaterial,
  setPrimaryBookAssetImage,
  saveDraft,
	  saveVideoOverride,
  deleteProductionTask,
	deleteBatchFactoryBook,
  submitBatchMerge,
  submitBookMerge,
  submitBatchProduction,
  submitBookTo121,
  classifyBookPublishMetadata,
  get121OrganizationOptions,
  getBatchFactoryMergeCoverBlob,
  isBatchFactoryMergeCoverURL
} from '../../../shared/api/batchFactory';
import { getConfig } from '../../../shared/api/config';
import { isApiTimeoutError } from '../../../shared/api/client';
import { getCurrentAccount } from '../../../shared/api/auth';
import { listAvailableModels } from '../../../shared/api/modelCatalog';
import { textToSpeech } from '../../../shared/api/tts';
import { BatchFactoryUnifiedSettingsModal } from './BatchFactoryUnifiedSettingsModal';
import { preflightBatchAutomation } from '../../../shared/api/batchFactoryAutomationPreflight.js';
import { BatchFactoryPreflightModal } from './BatchFactoryPreflightModal';
import { BatchFactoryErrorNotice } from './BatchFactoryErrorNotice';
import { CmLoader } from './CmLoader.jsx';
import { PrivateStudioAssetEditor } from '../private-studio/PrivateStudioAssetEditor';
import '../private-studio/private-studio-asset-editor.css';
import { BatchFactoryBookSettingsModal } from './BatchFactoryBookSettingsModal';
import { audioPlanningRequired, normalizeAudioSource, normalizeAudioSourceSettings } from './batchFactoryAudioSource';
import { BatchFactoryGiantMaterialPendingProgress } from './BatchFactoryGiantMaterialPendingProgress';
import { comicImageLoadPlan, comicInlinePresentation, resolveComicImageSources } from './batchFactoryComicStatus.js';
import { shouldAutoClassifyBook, automationBookForCard, automationBookIsActive, automationBookNeedsRetry, batchFactoryAutomationRetryAction, batchFactoryBatchProgress, batchFactoryBookState, batchFactoryBookTimeline, batchFactoryNovelTableRow, batchFactoryVideoProgress, batchFactoryVisibleError } from './batchFactoryBookState';
import { batchFactoryCleanSourceText, batchFactoryPreviewText, comicAnalysisCharactersForBook, comicEstimatedPageRange, batchFactoryProductionText, batchFactorySourceText, contentCaptureCharactersForBook, contentRangeLinesForBook, inheritedContentSettingsForNewBook, publishContentWithWorkingFront, sourceTextAfterViralAdoption } from './batchFactoryContentRange';
import { assetsForProductionMode } from './batchFactoryAssetMode';
import { comicPageErrorMessage, comicPageGenerationPayload, comicPromptForImageAspect, generateComicPageWithConflict, runComicPagePool } from './batchFactoryComicMode';
import { batchFactoryShouldPrepareAudio, batchFactoryShouldRecompileVideoPrompts } from './batchFactoryProductionMode';
import { ComicSourceContent } from './ComicSourceContent.jsx';
import { batchMediaCounts as runtimeBatchMediaCounts, batchFactoryNovelCopyText, bookVideoVersions, resolveBookEditableContent as runtimeResolveBookEditableContent, resolveBookProductionText as runtimeResolveBookProductionText } from './batchFactoryRuntimeLogic';
import { downloadProductionContent } from './batchFactoryContentDownload.js';
import { realVideoDurationSeconds, formatDurationSeconds } from './batchFactoryVideoMeta';
import { getWorkshopPlatforms } from '../../../shared/api/novelFetchWorkshop';
import { ProductionMediaBoundary } from '../batch-factory-v11/ProductionMediaBoundary';
import { checkWebSubmitEnvironment, getWebSubmitConfig, testWebSubmitVisible } from '../../../shared/api/novelFetch';
import { YIZHANSHENGMING_EXECUTOR_PROVIDER, localExecutorTaskToast } from './batchFactoryLocalExecutor.js';
import { batchFactoryPlatformOptions } from './batchFactoryPlatformOptions';
import { currentMergeJobsByOpening, publishMergeFiles } from './batchFactoryMergeGroups.js';
import { BOOK_CONFIG_REGIONS, bookAssetSummary, bookConfigRegionStatus } from './batchFactoryBookConfigRegions';
import { videoProviderForModel } from './videoProviderBinding';
import { modelOptionsForProvider } from './providerModelBinding';
import { stripBaseSetupSection } from '../scriptFinalSegment';
import { formatBeijingDatetimeLocal, normalizeAutomationConcurrency, parseBeijingDatetimeLocal } from './batchFactoryAutomationSchedule';
import { agentCanvasURL, bindBatchFactoryBookAndNavigate } from '../agent-canvas-batch-factory.js';
import { DEFAULT_AUDIO_CATEGORIES, chooseN8Audio, producerAccountOptions, readComicPublishDirectoryCache, writeComicPublishDirectoryCache } from './batchFactoryComicPublishAccounts.js';
import { ComicReferenceChips } from './ComicReferenceChips.jsx';
import { removablePageReferences } from './comicReferenceMaterials.js';
import {
  batchFactoryAudioFingerprint, batchFactoryBookAudioInput, batchFactoryBookTts, generateBatchFactoryBookAudioMeasurement, blobToBase64, saveBookOverrideWithRetry, batchFactoryContentMode, batchFactorySourceVersionLabel, batchFactorySourceVersionClass, batchFactorySourceSummary, batchFactoryWorkbenchConfigRegions, value, bookPlatformName, sensitiveProcessingDetail, batchFactorySourceKind, isBatchFactoryContentReadRunning, batchFactorySourceStatus, resultData, waitForComicPageResult, waitForBookAssetImages, sameSettingsPatch, h3DirectorCards, h3DirectorDocument, formatH3DirectorCardPrompt, finalVideoPromptTemplate, requestID, stageActionKey, capability, assetName, assetDisplayLabel, assetIdentitySummary, assetPrompt, effectiveBookSettings, baseSetupEnabled, h3CompilationStatus, h3ConstraintPreview, storyboardAssetPromptText, composeStoryboardDisplayPrompt, promptPreview, selectedVideoPromptLabel, usesH3VideoPrompt, effectiveBookAssetRules, effectiveBookComicRules, decompressionCountForUpload, isBookUploadedTo121, uploadProgressForBook, completePublishMapping, hasClassificationValues, classificationLabel, classificationDetail, primaryMediaVersion, storyboardVideoLabel, bookMergeJobs, mergeJobSpeed, defaultBookMerge, latestBookMerge, readImageAsDataURL, readStoryboardAssetSelection, loadBookAssetDisplayURL, comicPageReferencePreviewURL, fileAsDataURL, withStoryboardAssetMentions, compactUnifiedAssetDescriptions, separateStoryboardAssetMentions, collapseRepeatedStoryboardAssetMentions, renderStoryboardPromptHTML, contentEditableTextOffset, contentEditablePromptText, setContentEditableTextOffset, batchFactoryPassiveMediaUrl, batchFactoryPassivePosterUrl, batchFactoryIsMergeMedia, batchFactoryParseMediaRatio, batchFactoryMediaOrientation, batchFactoryRatioText, batchFactoryStableMediaLabel, batchFactoryAssetRecordsRefreshKey, localStartTime, localizeShotKeywords
} from './batchFactoryNovelListHelpers.js';
import { automationFailureNotificationsReady, automationFailureSummaryContent, collectNewAutomationFailures } from './batchFactoryFailureNotice.js';



// 统一配置只写“配置”，从不写“内容”：这里出现的键全部是可回填的生成参数、
// 模型选择、约束模板或发布策略。aiPromptConfig 存的是提示词预设引用
// （presetId/presetVersion/presetSlot/constraintCategory）与已发布模板 body，
// 属于人工维护的配置，不含任何一本书已生成的分镜/提示词/图片/视频文本。
const GENERATED_CONTENT_SETTINGS_KEYS = [
  'audioLineDurations', 'audioDurationFingerprint', 'audioDurationManual', 'audioDurationSeconds',
  'ttsFingerprint', 'ttsAudioArtifactId', 'h3StyleAnalysis', 'starredCharacterNames',
  'storyTimelineVersion', 'contentRangeLines', 'contentCaptureCharacters'
];

const UNIFIED_CONFIGURATION_KEYS = [
  'textModelId', 'imageModelId', 'imageProvider', 'videoModelId', 'videoProvider', 'aspectRatio', 'imageAspectRatio', 'videoAspectRatio', 'videoResolution', 'productionMode', 'imageConcurrency', 'assetImagesEnabled',
	'sensitiveAIEnabled', 'sensitiveAIModels', 'sensitiveAIOnFailure',
  'storyboardDurationLimit', 'maxVideoDuration', 'audioPlanningEnabled',
	'audioSource', 'audioMergeEnabled', 'tts', 'publishRewriteEnabled', 'publishSettings', 'comicPublishSettings', 'aiPromptConfig', 'openingEnabled', 'openingCount',
	'watermarkEnabled', 'watermarkText'
];









const BATCH_FACTORY_TABLE_COLUMNS = [
  { label: '序号', minWidth: 56 },
  { label: '编辑内容', minWidth: 300 },
  { label: '单书配置', minWidth: 176 },
  { label: '分镜提示词', minWidth: 330 },
  { label: '分镜视频', minWidth: 300 },
  { label: '操作', minWidth: 220 }
];








function NovelMetadata({ books, createdAt, selectedBookIds, onSelectionChange, onViewBook, onRunBook, platformNames, platformOptions, productionStatus, mergeStatus, stageSummaries, batchId, onContentReady, onBackfill121Metadata, backfilling121Metadata, onAddBooks }) {
  const selected = new Set(selectedBookIds);
  const allSelected = books.length > 0 && books.every(book => selected.has(book.id));
  const selectedCount = selectedBookIds.length;
  const toggleAll = checked => onSelectionChange(checked ? books.map(book => book.id) : []);
  const pendingContentBooks = books.filter(isBatchFactoryContentReadRunning);
  const copyAllBooks = async event => {
    event?.stopPropagation?.();
    if (!books.length) {
      message.warning('当前批次没有可复制的小说');
      return;
    }
    const text = batchFactoryNovelCopyText(books, platformNames);
    try {
      await navigator.clipboard.writeText(text);
      message.success(`已复制 ${books.length} 本书`);
    } catch (error) {
      message.error('复制失败，请检查浏览器剪贴板权限后重试');
    }
  };
  return <div className="batch-factory-novel-list batch-factory-novel-fetch-list batch-factory-novel-list-v13" role="table" aria-label="小说列表">
    {pendingContentBooks.length ? <div className="batch-factory-novel-pending-strip is-summary" role="status">
      <span className="batch-factory-novel-pending-count">正在后台读取 {pendingContentBooks.length} 本正文，详细状态和恢复操作请看对应书籍的“入厂内容”。</span>
    </div> : null}
    <div className="batch-factory-novel-list-toolbar">
      <Space size={8}>
        <Checkbox checked={allSelected} indeterminate={!allSelected && selected.size > 0} onChange={event => toggleAll(event.target.checked)} aria-label="全选小说"><b>序号</b></Checkbox>
        {selectedCount > 0 && <span className="batch-factory-novel-selected-count">已选 {selectedCount} 本</span>}
		<Button size="small" loading={backfilling121Metadata} disabled={!batchId || backfilling121Metadata} onClick={onBackfill121Metadata}>补全 121 男女频</Button>
      </Space>
    </div>
    <div className="batch-factory-novel-list-head" role="row">
      <span className="is-order">#</span>
      <span className="is-novel-head">
        <span>小说</span>
        <Tooltip title="复制全部书名 / 书城ID / 书城">
          <Button
            type="text"
            size="small"
            className="batch-factory-copy-books"
            aria-label="复制全部书名 / 书城ID / 书城"
            icon={<CopyOutlined />}
            onClick={copyAllBooks}
          />
        </Tooltip>
      </span><span>入厂内容</span><span>风格</span><span>男女频</span><span>状态</span><span>操作</span>
    </div>
    {books.map((book, index) => {
      const row = batchFactoryNovelTableRow(book, index, createdAt, { productionStatus, mergeStatus, stageSummary: stageSummaries?.[book.id] });
      const metadata = book.sourceMetadata || {};
      const sourceSummary = batchFactorySourceSummary(book);
      const sourceStatus = batchFactorySourceStatus(book);
      return <div className="batch-factory-novel-list-row" key={book.id} role="row">
        <span className="is-order"><Checkbox checked={selected.has(book.id)} onChange={event => onSelectionChange(event.target.checked ? [...selected, book.id] : [...selected].filter(id => id !== book.id))} aria-label={`选择 ${row.title}`} /><b>{String(index + 1).padStart(2, '0')}</b></span>
        <span className="is-book"><strong title={row.title}>{row.title}</strong><small>ID {row.bookId} · {bookPlatformName(book, platformNames)} · {batchFactorySourceKind(book)}</small></span>
        <span className={`batch-factory-content-version ${sourceSummary.ready ? 'is-ready' : 'is-missing'}`} title={`${batchFactorySourceVersionLabel(book)} · ${sourceSummary.text}`}><i className={`batch-factory-content-mode is-${batchFactorySourceVersionClass(book)}`}>{batchFactorySourceVersionLabel(book)}</i></span>
        <span>{publicationMetadataValue(metadata, 'style')}</span>
        <span>{publicationMetadataValue(metadata, 'gender')}</span>
        <span className={`batch-factory-source-status is-${sourceStatus.tone} ${row.status === '处理中' ? 'is-scheduled' : ''}`} title={`${sourceStatus.detail}${row.state?.detail ? ` · ${row.state.detail}` : ''}`}><b>{row.state?.published ? '已发布' : sourceStatus.label}</b><small>{row.state?.published ? sourceStatus.label : row.status}{row.state?.manual ? ' · 已手调' : ''}</small></span>
        <span className="batch-factory-novel-actions"><Button size="small" onClick={() => onViewBook(book)}>查看</Button>{onRunBook ? <Button size="small" type="link" onClick={() => onRunBook(book)}>执行这本</Button> : null}</span>
      </div>;
    })}
    <AddBooksRow books={books} platformOptions={platformOptions} onAddBooks={onAddBooks} />
  </div>;
}

function AddBooksRow({ books, platformOptions = [], onAddBooks }) {
  const [open, setOpen] = useState(false);
  const [inputText, setInputText] = useState('');
  const [platform, setPlatform] = useState('');
  const [allowDuplicate, setAllowDuplicate] = useState(false);
  const [busy, setBusy] = useState(false);
  const options = Array.isArray(platformOptions) && platformOptions.length
    ? platformOptions
    : batchFactoryPlatformOptions([], { fallback: true });
  useEffect(() => {
    if (!platform && options.length) setPlatform(String(options[0].value));
  }, [options, platform]);
  const parsed = useMemo(() => String(inputText || '').split(/\r?\n/).map(line => line.trim()).filter(Boolean).map(line => {
    const match = line.match(/^(\d{1,25})(?:\s+|[，,、|])?(.*)$/);
    return match ? { bookId: match[1], title: String(match[2] || '').trim() || `小说 ${match[1]}` } : null;
  }).filter(Boolean), [inputText]);
  const existing = useMemo(() => new Set(books.map(book => String(book?.bookId || '').trim()).filter(Boolean)), [books]);
  const duplicates = parsed.filter(item => existing.has(item.bookId));
  async function submit() {
    if (!parsed.length) { message.warning('请按每行“书 ID”或“书 ID + 书名”输入'); return; }
    if (!platform) { message.warning('请选择书城'); return; }
    if (duplicates.length && !allowDuplicate) { message.warning(`发现 ${duplicates.length} 本已存在的书，勾选“重复 ID 仍然添加”后再提交`); return; }
    setBusy(true);
    try {
      await onAddBooks(parsed, platform, allowDuplicate);
      setInputText('');
      setAllowDuplicate(false);
      setOpen(false);
    } catch (error) {
      message.error(error?.message || '添加新书失败');
    } finally { setBusy(false); }
  }
  return <>
    <div className="batch-factory-novel-list-row batch-factory-novel-add-row" role="row">
      <span className="is-order"><b>＋</b></span>
      <span className="is-book"><Button type="link" size="small" onClick={() => setOpen(true)}>添加新书</Button><small>输入多个书 ID 或“书 ID + 书名”</small></span>
      <span>—</span><span>—</span><span>—</span><span>—</span><span>选择书城后添加</span>
    </div>
    <Modal className="batch-factory-add-books-modal" title={<span className="batch-factory-add-books-title"><b>添加新书</b><small>写入当前批次小说列表</small></span>} open={open} onCancel={() => setOpen(false)} onOk={submit} confirmLoading={busy} okText="添加到当前批次" width={620} destroyOnClose>
      <Space direction="vertical" size={12} style={{ width: '100%' }}>
        <Alert type="info" showIcon message="添加后自动获取正文与分类信息" description="新书会继承当前批次配置，自动获取正文并识别风格、男女频；当前已有任务时，正文获取成功的新书会排在后面继续生产。" />
        <label className="batch-factory-engine-field"><span><b>书城</b></span><Select value={platform || undefined} onChange={value => setPlatform(String(value || ''))} options={options.map(item => ({ value: String(item.value), label: item.label }))} placeholder="请选择书城" style={{ width: '100%' }} /></label>
        <label className="batch-factory-engine-field"><span><b>书 ID / 书名</b><small>每行一本，例如：2091926811672429633\t书名</small></span><Input.TextArea rows={8} value={inputText} onChange={event => setInputText(event.target.value)} placeholder={'2091926811672429633\n209208048803840415\t第二本小说'} /></label>
        {duplicates.length ? <Alert type="warning" showIcon message={`发现 ${duplicates.length} 本已存在的书`} description={duplicates.map(item => item.bookId).join('、')} /> : null}
        <Checkbox checked={allowDuplicate} onChange={event => setAllowDuplicate(event.target.checked)}>重复 ID 仍然添加</Checkbox>
      </Space>
    </Modal>
  </>;
}


function AssetEditor({ book, batchId, onSaved, onGenerate, onRegenerate, onRetry, onAssetPromptChange, assetRules, productionMode = 'video', canGenerate, generateReason, generating, engineSettings }) {
  const [assets, setAssets] = useState([]);
  const [loading, setLoading] = useState(false);
  const [savingId, setSavingId] = useState('');
  const [kind, setKind] = useState('character');
  const [libraryKind, setLibraryKind] = useState('character');
  const [search, setSearch] = useState('');
  const [assetPromptCatalog, setAssetPromptCatalog] = useState([]);
  const [assetPromptLoading, setAssetPromptLoading] = useState(false);
  const [selectedAssetIds, setSelectedAssetIds] = useState([]);
  const [imageGenerating, setImageGenerating] = useState(false);
  const [imageGenerationResult, setImageGenerationResult] = useState(null);
  const [assetImageThreeView, setAssetImageThreeView] = useState(true);
  const [assetImages, setAssetImages] = useState({});
  const [editing, setEditing] = useState(null);
	const [imageModels, setImageModels] = useState([]);
	const [imageModelsLoading, setImageModelsLoading] = useState(false);
	const [starredCharacterNames, setStarredCharacterNames] = useState([]);
  const assetRecordsRefreshKey = batchFactoryAssetRecordsRefreshKey(book);
  const kindOptions = [{ value: 'character', label: '人物' }, { value: 'scene', label: '场景' }, { value: 'prop', label: '道具' }];
  const categoryLabel = value => kindOptions.find(item => item.value === value)?.label || '资产';
  async function loadImages(nextAssets, { quiet = false } = {}) {
    const entries = await Promise.all((nextAssets || []).map(async asset => {
      try { return [asset.id, resultData(await listBookAssetImages(batchId, book.id, asset.id), 'images')]; }
      catch (error) { if (!quiet) message.error(error?.message || '读取资产图片版本失败'); return [asset.id, []]; }
    }));
    setAssetImages(Object.fromEntries(entries.map(([assetId, images]) => [assetId, Array.isArray(images) ? images : []])));
  }
  const loadAssets = async ({ quiet = false } = {}) => {
    if (!batchId || !book?.id) return;
    setLoading(true);
    try {
      const response = await listBookAssets(batchId, book.id);
      const next = resultData(response, 'assets');
      const saved = Array.isArray(next) ? next : [];
      setAssets(saved);
      await loadImages(saved, { quiet: true });
    } catch (error) {
      if (!quiet) message.error(error?.message || '读取当前小说资产失败');
    } finally { setLoading(false); }
  };
  useEffect(() => {
    setAssets(Array.isArray(book?.assetRecords) ? book.assetRecords : []);
    setSearch('');
    setSelectedAssetIds([]);
    loadAssets({ quiet: true });
  }, [batchId, book?.id, assetRecordsRefreshKey]);
	useEffect(() => { setStarredCharacterNames(Array.isArray(book?.settingsState?.patch?.starredCharacterNames) ? book.settingsState.patch.starredCharacterNames : []); }, [book?.id, book?.revision, book?.settingsState?.revision]);
	useEffect(() => {
		let active = true;
		setImageModelsLoading(true);
		listAvailableModels('image').then(models => {
			if (active) setImageModels(models);
		}).catch(error => {
			if (active) message.error(error?.message || '读取图片模型失败');
		}).finally(() => {
			if (active) setImageModelsLoading(false);
		});
		return () => { active = false; };
	}, []);
  useEffect(() => {
    let active = true;
    setAssetPromptLoading(true);
    const modules = productionMode === 'comic' ? ['batch-factory'] : ['script', 'batch-factory'];
    Promise.all(modules.map(module => listSystemPresetCatalog(module))).then(results => {
      const records = results.flatMap(result => Array.isArray(result?.catalog) ? result.catalog : []);
      if (active) setAssetPromptCatalog(records.filter(item => productionMode === 'comic'
        ? item?.slot === 'batch.comic-assets'
        : item?.slot === 'script.asset-extraction' || item?.id === 'batch-character-meta'
      ));
    }).catch(error => { if (active) message.error(error?.message || '读取资产提示词失败'); }).finally(() => { if (active) setAssetPromptLoading(false); });
    return () => { active = false; };
  }, [book?.id, productionMode]);
  const textModelId = String(engineSettings?.textModelId || '').trim();
  const imageModelId = String(engineSettings?.imageModelId || '').trim();
  const imageProvider = String(engineSettings?.imageProvider || 'personal_api').trim();
  // 资产画幅默认值跟随全局七档常量的默认（normalizeImageAspectRatio 的回落值），不再写死 16:9。
  const imageAspectRatio = normalizeImageAspectRatio('');
	const imageModelOptions = useMemo(() => modelOptionsForProvider(imageModels, 'image', imageProvider), [imageModels, imageProvider]);
  const activeAssetSelection = productionMode === 'comic' ? assetRules?.assets : assetRules?.extraction;
  const currentAssetPromptId = String(activeAssetSelection?.presetId || '').trim();
  const assetPromptOptions = (currentAssetPromptId && !assetPromptCatalog.some(item => item.id === currentAssetPromptId)
    ? [{ id: currentAssetPromptId, name: activeAssetSelection?.presetName || '预设已失效', version: activeAssetSelection?.presetVersion || 1, disabled: true }, ...assetPromptCatalog]
    : assetPromptCatalog).map(item => ({ value: item.id, label: `${item.name || item.id} · v${item.version || 1}`, disabled: item.disabled === true }));
  const modeAssets = assetsForProductionMode(assets, productionMode);
  const visibleAssets = modeAssets.filter(asset => asset.kind === kind).filter(asset => {
    const query = search.trim().toLowerCase();
    return !query || `${asset.name} ${asset.prompt}`.toLowerCase().includes(query);
  });
  const libraryImages = modeAssets.filter(asset => asset.kind === libraryKind).flatMap(asset => (assetImages[asset.id] || []).map(image => ({ image, asset }))).filter(({ asset }) => {
    const query = search.trim().toLowerCase();
    return !query || `${asset.name} ${asset.prompt}`.toLowerCase().includes(query);
  });
  async function saveAsset({ close = true } = {}) {
    if (!editing?.name?.trim() || !editing?.prompt?.trim()) { message.warning('请填写资产名称和提示词。'); return null; }
    const existing = assets.find(asset => asset.id === editing.id);
    setSavingId(existing?.id || 'new');
    try {
      const assetDetails = {
        name: editing.name,
        formalName: editing.name,
        facts: editing.facts,
        appearance: editing.appearance || editing.prompt,
        relationships: editing.relationships,
		prompt: editing.prompt,
		imageModelId: editing.imageModelId || '',
		imageAspectRatio: editing.imageAspectRatio || ''
      };
      const result = existing
        ? await updateBookAsset(batchId, book.id, existing.id, { ...assetDetails, expectedRevision: Number(existing.revision || 0) })
        : await createBookAsset(batchId, book.id, { kind: editing.kind, source: productionMode === 'comic' ? 'comic' : 'manual', ...assetDetails });
      const savedAsset = resultData(result, 'asset');
      if (close) setEditing(null);
      await loadAssets({ quiet: true });
      await onSaved?.();
      message.success('当前小说资产已保存。');
      return savedAsset?.id ? savedAsset : null;
    } catch (error) { message.error(error?.message || '保存资产失败'); return null; } finally { setSavingId(''); }
  }
  async function runPreset() {
    try {
      if (!textModelId) { message.warning('请先选择个人中心已启用的文本模型。'); return; }
      await onGenerate?.(textModelId);
      await loadAssets({ quiet: true });
    } catch (_) { /* runAi already reports its business error */ }
  }
  async function chooseAssetPrompt(presetId) {
    const preset = assetPromptCatalog.find(item => item.id === presetId);
    try {
      await onAssetPromptChange?.(preset ? { presetId: preset.id, presetName: preset.name, presetSlot: preset.slot, presetVersion: preset.version, constraintCategory: preset.constraintCategory || '' } : { presetId: '' });
      message.success(preset?.id === 'batch-assets-h3' ? '人物场景道具方案已选中，点击“重新生成资产”应用到当前书。' : '资产提示词已保存。');
    } catch (error) { message.error(error?.message || '保存资产提示词失败'); }
  }
  function toggleAssetSelection(assetId, checked) {
    setSelectedAssetIds(current => checked ? [...new Set([...current, assetId])] : current.filter(id => id !== assetId));
  }
	async function toggleCharacterStar(asset) {
		if (asset?.kind !== 'character') return;
		const name = assetName(asset);
		const starred = starredCharacterNames.includes(name);
		const next = starred ? starredCharacterNames.filter(item => item !== name) : [...new Set([...starredCharacterNames, name])];
		setSavingId(`star-${asset.id}`);
		try {
			await saveBookOverride(batchId, book.id, { patch: { starredCharacterNames: next }, expectedRevision: Number(book?.revision || 0) });
			setStarredCharacterNames(next);
			await onSaved?.();
			message.success(starred ? `已取消“${name}”的星标聚焦。` : `已将“${name}”设为星标聚焦人物。`);
		} catch (error) { message.error(error?.message || '保存星标人物失败'); } finally { setSavingId(''); }
	}
  async function generateImages(assetIds = selectedAssetIds, candidateAssets = assets, useThreeView = assetImageThreeView) {
    if (!assetIds.length) { message.warning('请先选择要生成图片的资产。'); return false; }
	const selected = assetIds.map(assetId => candidateAssets.find(asset => asset.id === assetId)).filter(Boolean);
	if (selected.length !== assetIds.length || selected.some(asset => !(asset.imageModelId || imageModelId))) {
		message.warning('请为每个资产选择图片模型，或在书籍引擎配置中设置默认图片模型。');
		return false;
	}
    setImageGenerationResult(null);
    setImageGenerating(true);
    try {
      const result = await generateBookAssetImages(batchId, book.id, {
        assetIds,
        modelId: imageModelId,
        provider: imageProvider,
        imageAspectRatio,
        useThreeView: Boolean(useThreeView)
      });
      const initialImages = Array.isArray(result?.images) ? result.images : [];
      let savedImages = initialImages;
      if (result?.status === 'generating' || result?.status === 'processing' || initialImages.length < assetIds.length) {
        message.loading({ key: `asset-image-${book.id}`, content: `生成 ${assetIds.length} 张资产图片中…`, duration: 0 });
        savedImages = await waitForBookAssetImages(batchId, book.id, assetIds, result?.taskId || '');
      }
      await loadAssets({ quiet: true });
      await onSaved?.();
      const savedCount = savedImages.reduce((count, item) => count + (item.images?.length ? 1 : 0), 0);
      const failedCount = Math.max(0, assetIds.length - savedCount);
      if (failedCount > 0) {
        const summary = `已回填 ${savedCount} 个资产图片，另有 ${failedCount} 个生成失败，可单独重试。`;
        setImageGenerationResult({ type: 'warning', message: summary });
        message.warning({ key: `asset-image-${book.id}`, content: summary });
      } else {
        const summary = `已保存 ${savedCount} 个当前书资产图片版本。`;
        setImageGenerationResult({ type: 'success', message: summary });
        message.success({ key: `asset-image-${book.id}`, content: summary });
      }
      return true;
    } catch (error) {
      const failure = error?.message || '图片生成失败';
      setImageGenerationResult({ type: 'error', message: failure });
      message.error({ key: `asset-image-${book.id}`, content: failure });
      return false;
    } finally { setImageGenerating(false); }
  }
  async function generateEditedAssetImage() {
    const savedAsset = await saveAsset({ close: false });
    if (!savedAsset) return;
	if (await generateImages([savedAsset.id], [savedAsset], assetImageThreeView)) setEditing(null);
  }
  async function uploadImage(asset, event) {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file) return;
    if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type)) { message.error('仅支持 PNG、JPG、WEBP 图片。'); return; }
    setSavingId(`upload-${asset.id}`);
    try {
      await uploadBookAssetImage(batchId, book.id, asset.id, await readImageAsDataURL(file));
      await loadImages(assets, { quiet: true });
      await onSaved?.();
      message.success('图片版本已保存到当前小说资产。');
    } catch (error) { message.error(error?.message || '上传资产图片失败'); } finally { setSavingId(''); }
  }
  async function makePrimary(asset, image) {
    if (image.isPrimary) return;
    setSavingId(`primary-${image.id}`);
    try {
      await setPrimaryBookAssetImage(batchId, book.id, asset.id, image.id);
      await loadImages(assets, { quiet: true });
      await onSaved?.();
      message.success('已切换为当前资产主图。');
    } catch (error) { message.error(error?.message || '切换主图失败'); } finally { setSavingId(''); }
  }
  return <section className="batch-factory-preset-shell">
    <Alert type="info" showIcon message="当前小说的人物、场景与道具" description={productionMode === 'comic' ? '这里只显示本次漫画自动资产；重新提取会替换上一批漫画资产，不会无限叠加。旧手动和视频资产仍保留在其它模式中。' : '这里只显示视频与手动资产；漫画自动资产单独保留在漫画模式。'} />
    {imageGenerationResult ? <Alert type={imageGenerationResult.type} showIcon closable onClose={() => setImageGenerationResult(null)} message="资产图片生成结果" description={imageGenerationResult.message} /> : null}
    <div className="shuihuo-preset-toolbar batch-factory-preset-toolbar">
      <Tooltip title={productionMode === 'comic' ? '选择当前书的漫画资产提示词；模型统一在引擎配置设置。' : '选择当前书的人物、场景与道具提取提示词；模型统一在引擎配置设置。'}><Select value={currentAssetPromptId || undefined} loading={assetPromptLoading} placeholder={productionMode === 'comic' ? '选择漫画资产提示词' : '选择资产提示词'} style={{ minWidth: 230 }} options={assetPromptOptions} onChange={chooseAssetPrompt} /></Tooltip>
      <Tooltip title={canGenerate ? (productionMode === 'comic' ? '提取当前书的漫画人物、场景、道具资产；不会覆盖视频或手动资产。' : '仅在当前书没有文案时提取人物、场景、道具；已有手动资产不会被覆盖。') : generateReason}><Button type="primary" loading={generating} disabled={!canGenerate || generating} onClick={runPreset}>{productionMode === 'comic' ? '提取漫画资产' : '提取资产'}</Button></Tooltip>
      <Tooltip title={productionMode === 'comic' ? '根据当前书内容重新提取漫画资产；视频和手动资产保留。' : '基于当前书内容重新生成 AI 文案和自动资产；手动编辑的资产提示词保留。'}><Button loading={generating} disabled={!canGenerate || generating || !textModelId} onClick={() => onRegenerate?.(textModelId)}>{productionMode === 'comic' ? '重新提取漫画资产' : '重新生成资产'}</Button></Tooltip>
      <Tooltip title="只重跑当前小说最后失败的阶段。"><Button disabled={!canGenerate || generating || !textModelId} onClick={() => onRetry?.(textModelId)}>{productionMode === 'comic' ? '重试漫画资产' : '重试资产'}</Button></Tooltip>
      <Button loading={loading} onClick={() => loadAssets()}>刷新资产</Button>
      <Tooltip title={`图片模型优先使用资产单独配置，未配置时继承书籍引擎；资产画幅未单独选择时使用默认画幅 ${imageAspectRatio}。`}><Button type="primary" loading={imageGenerating} disabled={imageGenerating || !selectedAssetIds.length} onClick={() => generateImages()}>AI生成（已选 {selectedAssetIds.length}）</Button></Tooltip>
      <Button onClick={() => setEditing({ kind, name: '', prompt: '' })}>手动添加</Button>
    </div>
    {!imageModelId ? <Alert type="info" showIcon message="尚未配置书籍默认图片模型" description="已为资产单独选择图片模型的资产仍可生成；其余资产请在编辑窗口选择模型，或在书籍引擎配置中设置默认模型。" /> : null}
    <div className="shuihuo-preset-layout batch-factory-preset-layout">
      <aside className="shuihuo-preset-list">
        <div className="shuihuo-preset-list-head"><strong>资产列表 ({visibleAssets.length})</strong><Select size="small" value={kind} onChange={setKind} options={kindOptions} /><Input.Search value={search} onChange={event => setSearch(event.target.value)} placeholder="搜索名称或提示词" allowClear /></div>
        <div className="shuihuo-prompt-list">
          {visibleAssets.map(asset => <article className={`shuihuo-prompt-row${selectedAssetIds.includes(asset.id) ? ' is-selected' : ''}${asset.kind === 'character' && starredCharacterNames.includes(assetName(asset)) ? ' is-starred' : ''}`} key={asset.id}><div className="shuihuo-prompt-row-head"><Checkbox checked={selectedAssetIds.includes(asset.id)} onChange={event => toggleAssetSelection(asset.id, event.target.checked)}>生成</Checkbox><Tag>{categoryLabel(asset.kind)}</Tag><strong title={assetIdentitySummary(asset) || undefined}>{assetDisplayLabel(asset)}</strong>{asset.kind === 'character' ? <Tooltip title={starredCharacterNames.includes(assetName(asset)) ? '取消星标人物聚焦' : '设为星标人物聚焦'}><Button type="text" size="small" className="batch-factory-character-star" loading={savingId === `star-${asset.id}`} onClick={() => toggleCharacterStar(asset)} aria-label={`${assetDisplayLabel(asset)} 星标人物聚焦`}>★</Button></Tooltip> : null}<Button type="text" size="small" onClick={() => setEditing({ ...asset })}>编辑</Button></div><p>{assetPrompt(asset) || '尚未填写可视化提示词'}</p></article>)}
          {!loading && !visibleAssets.length ? <div className="shuihuo-prompt-empty"><b>＋</b><p>暂无提示词，请点击“智能预设”或手动添加。</p><Button onClick={() => setEditing({ kind, name: '', prompt: '' })}>添加提示词</Button></div> : null}
        </div>
      </aside>
      <main className="shuihuo-preset-library">
        <div className="shuihuo-asset-tabs">{kindOptions.map(option => <button type="button" className={libraryKind === option.value ? 'active' : ''} onClick={() => setLibraryKind(option.value)} key={option.value}>{option.label}</button>)}</div>
        <Input.Search value={search} onChange={event => setSearch(event.target.value)} placeholder={`搜索${categoryLabel(libraryKind)}图片...`} className="shuihuo-preset-search" allowClear />
        <div className="shuihuo-image-library-grid">{libraryImages.map(({ image, asset }) => <article className="shuihuo-image-card" key={image.id}><div className="shuihuo-image-card-visual"><AuthenticatedBookAssetImage imageUrl={image.url} alt={`${assetDisplayLabel(asset)} 图片版本`} /></div><div className="shuihuo-image-card-title"><strong title={assetIdentitySummary(asset) || undefined}>{assetDisplayLabel(asset)}</strong><Tag>{image.isPrimary ? '主图' : `版本 ${image.revision || ''}`}</Tag></div><p>{assetPrompt(asset) || '生成时未保存提示词快照'}</p><div className="shuihuo-image-card-actions">{!image.isPrimary ? <Button type="text" size="small" loading={savingId === `primary-${image.id}`} onClick={() => makePrimary(asset, image)}>设主图</Button> : <span>当前主图</span>}<label className="batch-factory-image-upload"><input type="file" accept="image/png,image/jpeg,image/webp" onChange={event => uploadImage(asset, event)} disabled={savingId !== ''} />上传版本</label></div></article>)}{!libraryImages.length ? <div className="shuihuo-image-empty"><b>图片库</b><p>暂无已生成图片</p><span>在左侧勾选人物、场景或道具后，点击顶部“AI生成”。</span></div> : null}</div>
      </main>
    </div>
    <PrivateStudioAssetEditor
      open={Boolean(editing)}
      title={editing?.id ? `编辑资产 · ${assetName(editing)}` : `添加${categoryLabel(editing?.kind || kind)}`}
      onClose={() => setEditing(null)}
      width={920}
      basic={<>
        <div className="batch-factory-asset-editor-config-row">
          <label className="batch-factory-asset-editor-config-field is-kind"><span>资产分类</span><Select value={editing?.kind} options={kindOptions} onChange={value => setEditing(current => ({ ...current, kind: value }))} /></label>
          <label className="batch-factory-asset-editor-config-field is-model"><span>图片模型</span><Select allowClear value={editing?.imageModelId || imageModelId || undefined} loading={imageModelsLoading} placeholder={imageModelOptions.length ? '使用书籍默认' : '未配置图片模型'} options={imageModelOptions} onChange={imageModelId => setEditing(current => ({ ...current, imageModelId: imageModelId || '' }))} /></label>
          <label className="batch-factory-asset-editor-config-field is-ratio"><span>图片画幅</span><Select allowClear value={editing?.imageAspectRatio || imageAspectRatio || undefined} placeholder={`默认 ${imageAspectRatio}`} options={IMAGE_ASPECT_RATIO_OPTIONS} onChange={imageAspectRatio => setEditing(current => ({ ...current, imageAspectRatio: imageAspectRatio || '' }))} /></label>
        </div>
        <label className="shuihuo-form-label">名称<Input value={editing?.name || ''} onChange={event => setEditing(current => ({ ...current, name: event.target.value }))} /></label>
      </>}
      settings={editing?.kind === 'character' ? <section className="batch-factory-asset-editor-character-section">
        <div className="batch-factory-asset-editor-section-title">人物设定</div>
        <div className="batch-factory-asset-editor-character-fields">
          <label className="shuihuo-form-label">人物事实<Input.TextArea rows={3} value={editing?.facts || ''} onChange={event => setEditing(current => ({ ...current, facts: event.target.value }))} placeholder="身份、性别、年龄阶段、别名等原文事实" /></label>
          <label className="shuihuo-form-label">人物关系<Input.TextArea rows={3} value={editing?.relationships || ''} onChange={event => setEditing(current => ({ ...current, relationships: event.target.value }))} placeholder="与其他人物的关系及原文依据" /></label>
          <label className="shuihuo-form-label batch-factory-asset-editor-character-appearance">人物外形<Input.TextArea rows={4} value={editing?.appearance || editing?.prompt || ''} onChange={event => setEditing(current => ({ ...current, appearance: event.target.value }))} placeholder="脸型、发型、服装、体态等跨镜头一致性锚点" /></label>
        </div>
      </section> : <Typography.Text type="secondary">场景 / 道具暂无独立设定字段，提示词在下方维护；现有字段一个未删，也没有新造字段。</Typography.Text>}
      prompts={<label className="shuihuo-form-label">合并资产提示词 / AI 生图提示词<Input.TextArea rows={editing?.kind === 'character' ? 6 : 12} value={editing?.prompt || ''} onChange={event => setEditing(current => ({ ...current, prompt: event.target.value }))} placeholder="输入最终合并后的可视化提示词" /></label>}
      ruleNote="规则：这个资产有图片时，生成视频走“图片生视频”，不发文字设定；没图片时，按上面的文字走“文字生视频”。"
      aside={<>
        {editing?.id ? (() => {
          const imageList = assetImages[editing.id] || [];
          const mainImage = imageList.find(item => item.isPrimary) || imageList[0] || null;
          const panelBusy = savingId !== '' || imageGenerating;
          return <>
            <div className="batch-factory-asset-editor-panel-head"><b>主图预览</b></div>
            {mainImage ? <div className="batch-factory-asset-editor-main"><AuthenticatedBookAssetImage imageUrl={mainImage.url} alt={`${assetName(editing)} 主图`} preview /></div>
              : <div className="batch-factory-asset-editor-main is-empty"><PictureOutlined /><span>暂无主图，请上传图片或用 AI 生成</span></div>}
            <div className="batch-factory-asset-editor-panel-actions">
              {editing?.kind === 'character' ? <Tooltip title="开启后人物资产图按“正面近照 + 正面/侧面/背面全身”三视图生成，便于分镜保持角色一致。"><span className="batch-factory-asset-editor-three-view"><Switch size="small" checked={assetImageThreeView} onChange={setAssetImageThreeView} disabled={panelBusy} />人物三视图</span></Tooltip> : null}
              <label className="batch-factory-image-upload"><input type="file" accept="image/png,image/jpeg,image/webp" onChange={event => uploadImage(editing, event)} disabled={panelBusy} />上传图片版本</label>
              <Tooltip title={`先保存当前提示词；图片模型未单独选择时继承书籍引擎，资产画幅未单独选择时使用默认画幅 ${imageAspectRatio}。`}><Button type="primary" loading={imageGenerating} disabled={panelBusy} onClick={generateEditedAssetImage}>AI生成</Button></Tooltip>
            </div>
            <div className="batch-factory-asset-editor-thumbs">{imageList.map(item => <button type="button" key={item.id} className={`batch-factory-asset-editor-thumb${item.id === mainImage?.id ? ' is-main' : ''}`} title={item.isPrimary ? '当前主图' : '切换为主图'} disabled={panelBusy} onClick={() => makePrimary(editing, item)}><AuthenticatedBookAssetImage imageUrl={item.url} alt={`${assetName(editing)} 版本 ${item.revision || ''}`} />{item.isPrimary ? <span className="batch-factory-asset-editor-thumb-tag">主图</span> : null}</button>)}{!imageList.length ? <span className="batch-factory-asset-editor-thumbs-empty">还没有图片版本</span> : null}</div>
          </>;
        })() : <div className="batch-factory-asset-editor-main is-empty"><PictureOutlined /><span>先点右下角“保存”建立这个资产，<br />然后才能上传或生成图片</span></div>}
      </>}
      footer={<Space><Button onClick={() => setEditing(null)}>取消</Button><Button type="primary" loading={savingId !== ''} onClick={saveAsset}>保存</Button></Space>}
    />
  </section>;
}






function AuthenticatedBookAssetImage({ imageUrl, alt = '', preview = false }) {
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

  if (!display.url) return <span className="batch-factory-asset-image-state">{display.failed ? '图片加载失败' : ''}</span>;
  return preview
    ? <Image src={display.url} alt={alt} preview={{ mask: '点击放大' }} />
    : <img src={display.url} alt={alt} />;
}

function comicImageFailureText(error) {
  const status = typeof error?.status === 'number' ? error.status : 0;
  if (status === 401 || status === 403) return '登录状态已失效，请刷新页面';
  if (status === 404) return '图片已不在存储中';
  return '图片加载失败，点击重试';
}

function AuthenticatedComicPageImage({ imageUrl, thumbnailUrl = '', alt = '', preview = false, lazy = false }) {
  const hostRef = useRef(null);
  // 列表行只取 420px 缩略图；详情先显示缩略图，再静默替换成原图。
  const sources = resolveComicImageSources({ url: imageUrl, thumbnailUrl });
  const loadPlan = comicImageLoadPlan({ url: imageUrl, thumbnailUrl }, preview);
  const [visible, setVisible] = useState(!lazy);
  const [display, setDisplay] = useState({ url: '', failed: false, message: '' });
  const [retryToken, setRetryToken] = useState(0);
  useEffect(() => {
    if (!lazy) { setVisible(true); return undefined; }
    if (visible) return undefined;
    const node = hostRef.current;
    if (!node || typeof IntersectionObserver === 'undefined') { setVisible(true); return undefined; }
    const observer = new IntersectionObserver(entries => {
      if (entries.some(entry => entry.isIntersecting)) setVisible(true);
    }, { rootMargin: '240px' });
    observer.observe(node);
    return () => observer.disconnect();
  }, [lazy, visible]);
  useEffect(() => {
    if (!visible || loadPlan.length === 0) return undefined;
    let active = true;
    let ownedUrl = '';
    setDisplay({ url: '', failed: false, message: '' });
    void (async () => {
      let loaded = false;
      let lastError;
      for (const source of loadPlan) {
        try {
          const blob = await getComicPageImageBlob(source);
          if (!active) return;
          const nextUrl = URL.createObjectURL(blob);
          const previousUrl = ownedUrl;
          ownedUrl = nextUrl;
          loaded = true;
          setDisplay({ url: nextUrl, failed: false, message: '' });
          if (previousUrl) URL.revokeObjectURL(previousUrl);
        } catch (error) {
          lastError = error;
        }
      }
      if (active && !loaded) setDisplay({ url: '', failed: true, message: comicImageFailureText(lastError) });
    })();
    return () => {
      active = false;
      if (ownedUrl) URL.revokeObjectURL(ownedUrl);
    };
  }, [visible, retryToken, preview, sources.thumbnail, sources.original]);
  if (!visible) return <span ref={hostRef} className="batch-factory-comic-image-state" aria-hidden="true" />;
  if (!display.url) {
    return display.failed
      ? <span className="batch-factory-comic-image-state" role="button" tabIndex={0} title="点击重新加载" style={{ cursor: 'pointer' }} onClick={() => setRetryToken(value => value + 1)}>{display.message}</span>
      : <span className="batch-factory-comic-image-state">正在读取图片…</span>;
  }
  return preview ? <Image src={display.url} alt={alt} preview={{ mask: '点击放大' }} /> : <img src={display.url} alt={alt} />;
}


function AuthenticatedComicPageReference({ referenceURL }) {
  const [displayURL, setDisplayURL] = useState('');
  useEffect(() => {
    let active = true;
    let objectURL = '';
    const previewURL = comicPageReferencePreviewURL(referenceURL);
    if (!previewURL) { setDisplayURL(''); return undefined; }
    getComicReferencePreviewBlob(previewURL).then(blob => {
      if (!active) return;
      objectURL = URL.createObjectURL(blob);
      setDisplayURL(objectURL);
    }).catch(() => { if (active) setDisplayURL(''); });
    return () => { active = false; if (objectURL) URL.revokeObjectURL(objectURL); };
  }, [referenceURL]);
  return displayURL ? <Image width={48} height={48} src={displayURL} alt="本页参考图" preview={{ mask: '查看参考图' }} /> : null;
}


function useAuthenticatedBookAssetThumbs(batchId, book) {
  const [thumbs, setThumbs] = useState({});
  const records = (book?.assetRecords || []).filter(asset => asset?.id);
  const recordKey = records.map(asset => `${asset.id}:${asset.revision || 0}`).join('|');
  useEffect(() => {
    let active = true;
    if (!batchId || !book?.id || !records.length) {
      setThumbs({});
      return () => { active = false; };
    }
    Promise.all(records.map(async asset => {
      try {
        const response = await listBookAssetImages(batchId, book.id, asset.id);
        const images = resultData(response, 'images');
        const primary = Array.isArray(images) ? (images.find(image => image?.isPrimary) || images[0]) : null;
        // Keep only the protected URL here. Downloading every original image
        // while the table mounts fills the browser connection pool and can
        // delay even the next page's JavaScript. The image component fetches
        // the URL only when a visible prompt actually renders that asset.
        const imageUrl = primary?.thumbnailUrl || primary?.thumbnail_url || primary?.url || primary?.imageUrl || primary?.image_url || primary?.path || '';
        return [String(asset.id), imageUrl];
      } catch (_) {
        return [String(asset.id), ''];
      }
    })).then(entries => { if (active) setThumbs(Object.fromEntries(entries)); });
    return () => { active = false; };
  }, [batchId, book?.id, book?.revision, recordKey]);
  return thumbs;
}






function renderStoryboardPromptTextPart(part, keyPrefix) {
  return String(part || '').split('\n').flatMap((line, lineIndex) => [
    lineIndex ? <br key={`${keyPrefix}-br-${lineIndex}`} /> : null,
    <span key={`${keyPrefix}-text-${lineIndex}`}>{line}</span>
  ].filter(Boolean));
}

function renderStoryboardPromptWithAssetMentions(prompt, book, video, assetImages = {}) {
  let text = compactUnifiedAssetDescriptions(withStoryboardAssetMentions(prompt, book, video), book);
  // Director/SD cards from older runs often contain the asset name followed by
  // its full appearance description instead of an @ reference. Normalize only
  // the durable asset name for display; the description remains readable after
  // the thumbnail chip and no regeneration is needed.
  (book?.assetRecords || [])
    .map(asset => assetName(asset))
    .filter(Boolean)
    .sort((left, right) => right.length - left.length)
    .forEach(name => {
      const escaped = name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      text = text.replace(new RegExp(`(^|[^@])${escaped}`, 'g'), `$1@${name}`);
    });
  text = collapseRepeatedStoryboardAssetMentions(text, book);
  text = separateStoryboardAssetMentions(text, book);
  const selectedIds = readStoryboardAssetSelection(book, video).selectedIds;
  const assets = (book?.assetRecords || [])
    .filter(asset => selectedIds.has(String(asset?.id || '')) || text.includes(`@${assetName(asset)}`))
    .map(asset => ({ ...asset, mention: `@${assetName(asset)}`, imageUrl: assetImages[String(asset.id || '')] || asset?.imageUrl || asset?.image_url || asset?.thumbnailUrl || '' }))
    .filter(asset => asset.mention.length > 1)
    .sort((left, right) => right.mention.length - left.mention.length);
  if (!text || !assets.length) return text;
  const pattern = new RegExp(`(${assets.map(asset => asset.mention.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|')})`, 'g');
  const assetByMention = new Map(assets.map(asset => [asset.mention, asset]));
  return text.split(pattern).map((part, index) => {
    const asset = assetByMention.get(part);
    return asset
      ? <span key={`${asset.id || asset.mention}-${index}`} contentEditable={false} className={`batch-factory-asset-mention is-${asset.kind || 'asset'}`} title={`${asset.kind || '资产'}：${assetDisplayLabel(asset)}${assetIdentitySummary(asset) ? ` · ${assetIdentitySummary(asset)}` : ''}`}><span className="batch-factory-asset-mention-avatar">{asset.imageUrl ? <AuthenticatedBookAssetImage imageUrl={asset.imageUrl} alt="" /> : asset.kind === 'character' ? '人' : asset.kind === 'scene' ? '景' : '物'}</span>{assetDisplayLabel(asset)}</span>
      : renderStoryboardPromptTextPart(part, `prompt-text-${index}`);
  });
}






function InlineStoryboardAssets({ book, batchId, selectedVideoId: controlledSelectedVideoId = '', onSelectedVideoChange, onManage, onSaved }) {
  const videos = book?.videos || [];
  const precompiledWorkspace = resolvePrecompiledVideoWorkspace(book);
  const [localSelectedVideoId, setLocalSelectedVideoId] = useState(videos[0]?.id || '');
  const [savingAssetId, setSavingAssetId] = useState('');
  const [starredCharacterNames, setStarredCharacterNames] = useState([]);
  const clickTimers = useRef(new Map());
  const selectedVideoId = controlledSelectedVideoId || localSelectedVideoId;
  const selectVideo = nextId => {
    const value = String(nextId || '');
    setLocalSelectedVideoId(value);
    onSelectedVideoChange?.(value);
  };
  useEffect(() => { setStarredCharacterNames(Array.isArray(book?.settingsState?.patch?.starredCharacterNames) ? book.settingsState.patch.starredCharacterNames : []); }, [book?.id, book?.revision, book?.settingsState?.revision]);
  useEffect(() => () => { for (const timer of clickTimers.current.values()) clearTimeout(timer); clickTimers.current.clear(); }, []);
  useEffect(() => {
    const h3Frame = resolvePrecompiledStoryboardFrame(book, selectedVideoId);
    const next = h3Frame?.key || (videos.some(video => video.id === selectedVideoId) ? selectedVideoId : (videos[0]?.id || ''));
    if (next !== selectedVideoId) selectVideo(next);
  }, [book?.id, videos.map(video => video.id).join('|')]);
  const selectedVideo = videos.find(video => video.id === selectedVideoId) || videos[0];
  const selectedPrecompiledFrame = resolvePrecompiledStoryboardFrame(book, selectedVideoId);
  const selection = readStoryboardAssetSelection(book, selectedVideo);
  const precompiledAssets = resolvePrecompiledStoryboardAssets(book);
  const showingPrecompiledAssets = !selectedVideo && precompiledAssets.length > 0;
  const assets = selectedVideo
    ? (book?.assetRecords || []).filter(asset => selection.candidateIds.includes(String(asset.id || '')))
    : precompiledAssets;
  const currentIndex = Math.max(0, videos.findIndex(video => video.id === selectedVideo?.id));
  const switchVideo = offset => selectVideo(videos[(currentIndex + offset + videos.length) % videos.length]?.id || '');
  const switchPrecompiledFrame = offset => {
    const next = precompiledWorkspace.frames?.[(selectedPrecompiledFrame?.index || 0) + offset];
    if (next) selectVideo(next.key);
  };
  const assetGroups = [
    ['character', '人物'],
    ['scene', '场景'],
    ['prop', '道具']
  ].map(([kind, label]) => ({ kind, label, items: assets.filter(asset => asset.kind === kind) }));

  async function saveStar(asset, starred) {
    if (asset?.kind !== 'character') return false;
    const name = assetName(asset);
    const next = starred ? [...new Set([...starredCharacterNames, name])] : starredCharacterNames.filter(item => item !== name);
    try {
      await saveBookOverride(batchId, book.id, { patch: { starredCharacterNames: next }, expectedRevision: Number(book?.revision || 0) });
      setStarredCharacterNames(next);
      await onSaved?.();
      return true;
    } catch (error) { message.error(error?.message || '保存星标人物失败'); return false; }
  }

  async function toggleAsset(asset, { clearStar = false } = {}) {
    if (!selectedVideo || savingAssetId) return;
    const assetId = String(asset.id || '');
    if (!assetId) return;
    const isSelected = selection.selectedIds.has(assetId);
    const autoAssetIds = selection.autoAssetIds;
    const addedAssetIds = isSelected ? selection.addedAssetIds.filter(id => id !== assetId) : [...new Set([...selection.addedAssetIds, ...(autoAssetIds.includes(assetId) ? [] : [assetId])])];
    const excludedAssetIds = isSelected ? [...new Set([...selection.excludedAssetIds, assetId])] : selection.excludedAssetIds.filter(id => id !== assetId);
    setSavingAssetId(assetId);
    try {
      if (clearStar && !(await saveStar(asset, false))) return;
      await saveVideoOverride(batchId, book.id, selectedVideo.id, { patch: { assetSelection: { autoAssetIds, addedAssetIds, excludedAssetIds } }, expectedRevision: Number(selectedVideo.revision || 0) });
      await onSaved?.();
      message.success(isSelected ? `已停用“${assetName(asset)}”，不会进入当前分镜的画面和视频生成。` : `已启用“${assetName(asset)}”。`);
    } catch (error) { message.error(error?.message || '保存分镜资产选择失败'); } finally { setSavingAssetId(''); }
  }

  function onAssetClick(asset) {
    const assetId = String(asset?.id || '');
    if (!assetId) return;
    if (asset.kind === 'character' && starredCharacterNames.includes(assetName(asset))) {
      toggleAsset(asset, { clearStar: true });
      return;
    }
    clickTimers.current.set(assetId, setTimeout(() => { clickTimers.current.delete(assetId); toggleAsset(asset); }, 220));
  }

  async function onAssetDoubleClick(asset) {
    const assetId = String(asset?.id || '');
    const timer = clickTimers.current.get(assetId);
    if (timer) { clearTimeout(timer); clickTimers.current.delete(assetId); }
    if (asset.kind !== 'character') return;
    setSavingAssetId(assetId);
    try {
      const starred = !starredCharacterNames.includes(assetName(asset));
      if (await saveStar(asset, starred)) message.success(starred ? `“${assetName(asset)}”已设为星标聚焦人物。` : `已取消“${assetName(asset)}”的星标聚焦。`);
    } finally { setSavingAssetId(''); }
  }

  if (!selectedVideo && !showingPrecompiledAssets) return <div className="batch-factory-cell-shell batch-factory-inline-assets is-empty"><div className="batch-factory-cell-empty">AI 推理后会在这里显示当前书的分镜资产。</div></div>;

  const selectedCount = showingPrecompiledAssets ? assets.length : assets.filter(asset => selection.selectedIds.has(String(asset.id))).length;
  return <div className="batch-factory-cell-shell batch-factory-inline-assets">
    <header className="batch-factory-cell-head">
      <div><b>{showingPrecompiledAssets ? `导演分镜卡 ${(selectedPrecompiledFrame?.index || 0) + 1} · 资产` : storyboardVideoLabel(selectedVideo, currentIndex)}</b><small>{showingPrecompiledAssets ? '导演分镜卡已提取，等待 VIDEO 编译' : `${selectedCount}/${assets.length || 0} 已启用`}</small></div>
      <span className="batch-factory-cell-state">{assets.length ? '资产就绪' : '待提取'}</span>
    </header>
    <div className="batch-factory-cell-content batch-factory-asset-content">
      {assets.length ? assetGroups.map(group => group.items.length ? <section className="batch-factory-asset-group-inline" key={group.kind}>
        <div className="batch-factory-asset-group-label"><b>{group.label}</b><span>{group.items.length}</span></div>
        <div className="batch-factory-storyboard-asset-cards">{group.items.map(asset => {
          const selected = showingPrecompiledAssets || selection.selectedIds.has(String(asset.id));
          const starred = !showingPrecompiledAssets && selected && asset.kind === 'character' && starredCharacterNames.includes(assetName(asset));
          const onClick = showingPrecompiledAssets ? () => onManage?.() : () => onAssetClick(asset);
          const onDoubleClick = showingPrecompiledAssets ? undefined : () => onAssetDoubleClick(asset);
          const title = showingPrecompiledAssets ? '导演分镜卡已提取；最终 VIDEO 编译后可按分镜调整资产。点击管理全部资产。' : starred ? '星标人物聚焦；点击后取消星标并停用，双击可取消星标。' : selected ? '点击停用；双击人物卡设为星标聚焦。' : '点击启用：用于当前分镜的图片和视频生成';
          return <button type="button" key={asset.id} disabled={savingAssetId === asset.id} onClick={onClick} onDoubleClick={onDoubleClick} className={`batch-factory-storyboard-asset-card ${starred ? 'is-starred' : selected ? 'is-selected' : 'is-muted'}`} title={title}>
            <span>{group.label}</span><b title={assetIdentitySummary(asset) || undefined}>{assetDisplayLabel(asset)}</b>{starred ? <i aria-label="星标人物">★</i> : null}
          </button>;
        })}</div>
      </section> : null) : <div className="batch-factory-cell-empty">当前分镜尚未识别到需要调用的资产。</div>}
      <div className="batch-factory-asset-summary">
        <span>{assetGroups[0].items.length} 人物</span><span>{assetGroups[1].items.length} 场景</span><span>{assetGroups[2].items.length} 道具</span>
        <button type="button" onClick={() => onManage?.(selectedVideo?.id || selectedPrecompiledFrame?.key)}>管理全部资产</button>
      </div>
    </div>
    {showingPrecompiledAssets ? <div className="batch-factory-cell-pager"><button type="button" aria-label="上一张导演分镜卡资产" disabled={!selectedPrecompiledFrame || selectedPrecompiledFrame.index === 0} onClick={() => switchPrecompiledFrame(-1)}><LeftOutlined /></button><span>{(selectedPrecompiledFrame?.index || 0) + 1} / {precompiledWorkspace.frames.length}</span><button type="button" aria-label="下一张导演分镜卡资产" disabled={!selectedPrecompiledFrame || selectedPrecompiledFrame.index >= precompiledWorkspace.frames.length - 1} onClick={() => switchPrecompiledFrame(1)}><RightOutlined /></button></div> : <div className="batch-factory-cell-pager">
      <button type="button" aria-label="上一分镜资产" disabled={currentIndex === 0} onClick={() => switchVideo(-1)}><LeftOutlined /></button>
      <span>{currentIndex + 1} / {videos.length}</span>
      <button type="button" aria-label="下一分镜资产" disabled={currentIndex >= videos.length - 1} onClick={() => switchVideo(1)}><RightOutlined /></button>
    </div>}
  </div>;
}

function useCompiledVideoPrompt(batchId, book, video, settingsRevision = 0, { enabled = true } = {}) {
  const [displayPrompt, setDisplayPrompt] = useState('');
  const [compiledPrompt, setCompiledPrompt] = useState('');
  const [baseSetupPrompt, setBaseSetupPrompt] = useState('');
  const [smartUnifiedPending, setSmartUnifiedPending] = useState(false);
  const [compilerError, setCompilerError] = useState('');
  const [loading, setLoading] = useState(false);
  const bookRevision = Number(book?.revision || 0);
  const videoRevision = Number(video?.revision || 0);
  useEffect(() => {
    let active = true;
    if (!enabled || !batchId || !book?.id || !video?.id) {
      setDisplayPrompt('');
      setCompiledPrompt('');
      setBaseSetupPrompt('');
      setSmartUnifiedPending(false);
      setCompilerError('');
      setLoading(false);
      return () => { active = false; };
    }
    setDisplayPrompt('');
    setCompiledPrompt('');
    setBaseSetupPrompt('');
    setSmartUnifiedPending(false);
    setCompilerError('');
    setLoading(true);
    getFinalPrompt(batchId, book.id, video.id, { silent: true, suppressGlobalError: true }).then(response => {
      if (!active) return;
	  setDisplayPrompt(String(response?.finalPrompt?.displayPrompt || ''));
	  setCompiledPrompt(String(response?.finalPrompt?.compiledPrompt || ''));
      setBaseSetupPrompt(String(response?.finalPrompt?.baseSetupPrompt || ''));
      setSmartUnifiedPending(response?.finalPrompt?.smartUnifiedPending === true);
    }).catch(error => {
      // Never silently fall back to the stored director text here. That text
      // may have been produced by an older preset and would make a failed
      // compiler read look like the current preset was used successfully.
      if (active) {
        setDisplayPrompt('');
        setCompiledPrompt('');
        setBaseSetupPrompt('');
        setCompilerError(String(error?.message || '最终提示词读取失败，请刷新页面后重试'));
      }
    }).finally(() => {
      if (active) setLoading(false);
    });
    return () => { active = false; };
  }, [enabled, batchId, book?.id, bookRevision, video?.id, videoRevision, settingsRevision]);
  return { displayPrompt, compiledPrompt, baseSetupPrompt, smartUnifiedPending, compilerError, loading };
}

function DeferredInlineBookPrompts(props) {
  const placeholderRef = useRef(null);
  const [loaded, setLoaded] = useState(false);
  useEffect(() => {
    if (loaded) return undefined;
    const node = placeholderRef.current;
    if (!node || typeof IntersectionObserver === 'undefined') {
      setLoaded(true);
      return undefined;
    }
    const observer = new IntersectionObserver(entries => {
      if (entries.some(entry => entry.isIntersecting)) setLoaded(true);
    }, { rootMargin: '600px 0px' });
    observer.observe(node);
    return () => observer.disconnect();
  }, [loaded]);
  if (loaded) return <InlineBookPrompts {...props} />;
  return <div ref={placeholderRef} className="shuihuo-workbench-cell batch-factory-storyboard-cell">
    <div className="batch-factory-cell-empty">滚动到附近后加载分镜提示词。</div>
  </div>;
}

function InlineBookPrompts({ batch, book, batchId, settingsRevision, productionMode = 'video', selectedVideoId: controlledSelectedVideoId = '', onSelectedVideoChange, onManage }) {
  const videos = book?.videos || [];
	const h3Cards = h3DirectorCards(book);
  const precompiledWorkspace = resolvePrecompiledVideoWorkspace(book);
  const assetThumbs = useAuthenticatedBookAssetThumbs(batchId, book);
  const smartUnifiedAnalysis = smartUnifiedAnalysisForBook(book);
  const showSmartUnified = smartUnifiedDisplayEnabled(batch, book) && smartUnifiedAnalysis;
  const h3CompilationMessage = h3CompilationStatus(effectiveBookSettings(batch, book));
  const [localSelectedVideoId, setLocalSelectedVideoId] = useState(videos[0]?.id || '');
  const promptKind = productionMode === 'comic' ? 'visual' : 'video';
  const selectedVideoId = controlledSelectedVideoId || localSelectedVideoId;
  const selectVideo = nextId => {
    const value = String(nextId || '');
    setLocalSelectedVideoId(value);
    onSelectedVideoChange?.(value);
  };
  const selectedIndex = Math.max(0, videos.findIndex(video => video.id === selectedVideoId));
  const video = videos[selectedIndex] || null;
  const comicPrompt = comicPromptForImageAspect(effectiveBookSettings(batch, book), video?.comicPrompt || '');
  const hasVisualPrompt = Boolean(comicPrompt.trim());
  // 分镜卡直接读取 Go 编译后的最终 Prompt，让已开启的智能统一、基础设定、
  // 画质、限制和负面规则在卡片中可见；弹窗与卡片使用同一份编译结果。
  const currentSettings = effectiveBookSettings(batch, book);
  const baseSetupEnabledForBook = baseSetupEnabled(currentSettings);
  // 漫画模式没有导演分镜，读取视频最终提示词只会得到业务错误，直接不请求。
  // productionMode 缺省为 'video'：模式数据未加载时仍按视频模式请求。
  const finalPromptEnabled = productionMode !== 'comic';
  const { displayPrompt, compiledPrompt, baseSetupPrompt, smartUnifiedPending, compilerError, loading } = useCompiledVideoPrompt(batchId, book, video, settingsRevision, { enabled: finalPromptEnabled });
  // 智能统一是增强层。视觉基线尚未取得时，仍须展示已经编译的 VIDEO
  // 提示词，并明确标出基线状态，不能把一条可执行分镜伪装成“什么都没有”。
  const rawCompiledPrompt = compiledPrompt || displayPrompt;
  // 老卡片可能把“基础设定：人物/场景/道具”直接写进 VIDEO 正文；即使
  // Go 返回了兼容结果，展示层也必须先去掉这段旧正文，再注入当前资产表。
  const cleanedCompiledPrompt = stripBaseSetupSection(rawCompiledPrompt);
  const videoPrompt = withStoryboardAssetMentions(cleanedCompiledPrompt, book, video);
  const assetPromptText = baseSetupEnabledForBook
    ? (storyboardAssetPromptText(book, video) || String(baseSetupPrompt || '').trim())
    : '';
  const unifiedStyleText = showSmartUnified ? String(smartUnifiedAnalysis.prompt || '').trim() : '';
  const videoPromptWithUnifiedStyle = prependSmartUnifiedStyle(videoPrompt, unifiedStyleText);
  const displayVideoPrompt = baseSetupEnabledForBook
    ? composeStoryboardDisplayPrompt(videoPromptWithUnifiedStyle, assetPromptText || storyboardAssetPromptText(book, video))
    : videoPromptWithUnifiedStyle;
  const activePrompt = promptKind === 'visual' ? comicPrompt : displayVideoPrompt;

  useEffect(() => {
    const first = book?.videos?.[0]?.id || '';
    if (!controlledSelectedVideoId) setLocalSelectedVideoId(first);
  }, [book?.id]);

  useEffect(() => {
    if (!videos.some(item => item.id === selectedVideoId) && videos[0]?.id) selectVideo(videos[0].id);
  }, [videos.map(item => item.id).join('|')]);

  function move(direction) {
    const next = videos[selectedIndex + direction];
    if (next) selectVideo(next.id);
  }

  if (!video && h3Cards.length) {
		const h3Frame = resolvePrecompiledStoryboardFrame(book, selectedVideoId);
		const h3Card = h3Frame?.card || {};
		const h3CardIndex = h3Frame?.index || 0;
		const h3ConstraintRows = h3ConstraintPreview(effectiveBookSettings(batch, book));
		const h3ConstraintLabels = h3ConstraintRows.map(row => row.split('：', 1)[0]).join(' · ');
    const h3AssetPromptText = baseSetupEnabled(currentSettings) ? storyboardAssetPromptText(book, videos[selectedIndex] || null) : '';
		const h3VideoPromptLabel = selectedVideoPromptLabel(effectiveBookSettings(batch, book));
		const selectH3Frame = index => {
			const next = precompiledWorkspace.frames?.[index];
			if (next) selectVideo(next.key);
		};
		const camera = h3Card.camera || {};
		return <div className="shuihuo-workbench-cell shuihuo-prompt-cell batch-factory-book-prompt-cell">
			<div className="batch-factory-cell-shell batch-factory-prompt-entry-card is-h3-director-card">
				<header className="batch-factory-cell-head"><div><b>导演分镜卡 {h3CardIndex + 1}</b><small>视频提示词：{h3VideoPromptLabel} · 已提取，尚未编译 VIDEO</small></div></header>
				<button type="button" className="batch-factory-cell-content batch-factory-prompt-entry-content" aria-label="打开分镜提示词编辑" onClick={() => onManage?.(h3Frame?.key || '')}>
					<span className="batch-factory-prompt-status">导演分镜卡已提取</span>
					<p>{String(h3Card.source_text || '')}{h3Card.action ? `\n\n动作：${h3Card.action}` : ''}{camera.shot_size ? `\n机位：${camera.shot_size}${camera.shot_angle ? ` · ${camera.shot_angle}` : ''}` : ''}{h3AssetPromptText ? `\n\n【资产提示词】\n${h3AssetPromptText}` : ''}{h3ConstraintRows.length ? `\n\n【当前已启用约束】\n${h3ConstraintRows.join('\n')}` : ''}</p>
					<small title={[h3AssetPromptText, ...h3ConstraintRows].filter(Boolean).join('\n')}>{h3AssetPromptText ? `资产提示词：${promptPreview(h3AssetPromptText)}` : '暂无资产提示词'} · {h3ConstraintLabels ? `约束：${h3ConstraintLabels}` : '未启用约束'} · {h3CompilationMessage}</small>
				</button>
				<div className="batch-factory-cell-pager">
					<button type="button" aria-label="上一张导演分镜卡" disabled={h3CardIndex === 0} onClick={() => selectH3Frame(h3CardIndex - 1)}><LeftOutlined /></button>
					<span>{h3CardIndex + 1} / {h3Cards.length}</span>
					<button type="button" aria-label="下一张导演分镜卡" disabled={h3CardIndex >= h3Cards.length - 1} onClick={() => selectH3Frame(h3CardIndex + 1)}><RightOutlined /></button>
				</div>
			</div>
		</div>;
	}

  if (!video) return <div className="shuihuo-workbench-cell shuihuo-prompt-cell batch-factory-book-prompt-cell"><div className="batch-factory-cell-shell"><div className="batch-factory-cell-empty">AI 推理后会在这里显示当前书的分镜提示词。</div></div></div>;

  const stateLabel = promptKind === 'visual' ? (hasVisualPrompt ? '漫画已生成' : '待生成') : loading ? '编译中' : videoPrompt.trim() ? (smartUnifiedPending ? '基线未获取' : '最终已编译') : '待生成';
  const videoPromptLabel = selectedVideoPromptLabel(effectiveBookSettings(batch, book));

  return <div className="shuihuo-workbench-cell shuihuo-prompt-cell batch-factory-book-prompt-cell">
    <div className="batch-factory-cell-shell batch-factory-prompt-entry-card">
      <header className="batch-factory-cell-head">
        <div><b>{productionMode === 'comic' ? `漫画页 ${String(selectedIndex + 1).padStart(2, '0')}` : storyboardVideoLabel(video, selectedIndex)}</b><small>{promptKind === 'visual' ? '漫画提示词' : '最终视频提示词'}</small></div>
        <span className={`batch-factory-prompt-mode-label is-${productionMode}`}>{productionMode === 'comic' ? '漫画提示词' : '视频提示词'}</span>
      </header>
      <button type="button" className="batch-factory-cell-content batch-factory-prompt-entry-content" aria-label="打开分镜提示词编辑" onClick={() => onManage?.(video.id)}>
        <span className="batch-factory-prompt-status">{stateLabel}</span>
        <span className="batch-factory-prompt-entry-rich-view">
        {activePrompt ? renderStoryboardPromptWithAssetMentions(activePrompt, book, video, assetThumbs) : (compilerError && promptKind === 'video' ? `最终提示词读取失败：${compilerError}` : smartUnifiedPending && promptKind === 'video' ? '智能统一视觉基线尚未取得；不影响本分镜的提示词使用' : loading && promptKind === 'video' ? '正在编译最终视频提示词…' : promptKind === 'visual' ? '请生成漫画提示词再查看' : '请生成视频提示词再查看')}
        </span>
      </button>
      <div className="batch-factory-cell-pager">
        <button type="button" aria-label="上一分镜" disabled={selectedIndex === 0} onClick={() => move(-1)}><LeftOutlined /></button>
        <span>{selectedIndex + 1} / {videos.length}</span>
        <button type="button" aria-label="下一分镜" disabled={selectedIndex >= videos.length - 1} onClick={() => move(1)}><RightOutlined /></button>
      </div>
    </div>
  </div>;
}

function StoryboardVideoNavigator({ videos, selectedVideoId, onSelect }) {
	const index = Math.max(0, videos.findIndex(video => video.id === selectedVideoId));
	const hasVideos = videos.length > 0;
	return <Space wrap className="batch-factory-storyboard-navigator"><Tooltip title="上一分镜"><Button aria-label="上一分镜" icon={<LeftOutlined />} disabled={!hasVideos || index === 0} onClick={() => onSelect(videos[index - 1]?.id)} /></Tooltip><Select value={selectedVideoId || undefined} onChange={onSelect} style={{ minWidth: 260 }} placeholder="选择分镜 / VIDEO" options={videos.map((video, videoIndex) => ({ value: video.id, label: storyboardVideoLabel(video, videoIndex) }))} /><Tooltip title="下一分镜"><Button aria-label="下一分镜" icon={<RightOutlined />} disabled={!hasVideos || index >= videos.length - 1} onClick={() => onSelect(videos[index + 1]?.id)} /></Tooltip></Space>;
}

// openingVariants 按 VIDEO 版本存（不是按书存），所以“当前版本没有”不等于“整本书没有”。
function openingVariantSuccessCount(value) {
  const list = Array.isArray(value) ? value : (Array.isArray(value?.variants) ? value.variants : []);
  return list.filter(item => String(item?.status || '').toLowerCase() === 'success').length;
}

function PromptPanel({ book, batch, batchId, settingsRevision, productionMode = 'video', initialVideoId = '', isDeveloper = false, onSaved, onCompile, onRegenerateVisual, onRegenerateAllVisual, onRetry, onGenerateVideo, onGenerateVisual, onDownloadAllComicZip, comicZipDownloading = false, regenerating, productionAvailable = false, productionReason = '', onGenerateOpening, generatingOpening = false }) {
  const videos = book?.videos || [];
  const precompiledWorkspace = resolvePrecompiledVideoWorkspace(book);
	const [selectedVideoId, setSelectedVideoId] = useState(initialVideoId || videos[0]?.id || '');
  const promptKind = productionMode === 'comic' ? 'visual' : 'video';
  const [editing, setEditing] = useState(false);
  const [editingSnapshot, setEditingSnapshot] = useState(null);
  const [saving, setSaving] = useState(false);
  const [videoPrompt, setVideoPrompt] = useState('');
  const [comicDirectorPrompt, setComicDirectorPrompt] = useState('');
  const [comicPrompt, setComicPrompt] = useState('');
  const [comicPromptPreview, setComicPromptPreview] = useState({ loading: false, compiledPrompt: '', referenceImageUrls: [], referenceImages: [], error: '' });
  const [removingReferenceURL, setRemovingReferenceURL] = useState('');
  const [assetMentionMenu, setAssetMentionMenu] = useState(null);
  const promptInputRef = useRef(null);
  const promptEditableRef = useRef(null);
  const [h3TraceOpen, setH3TraceOpen] = useState(false);
  const [openingVariantsOpen, setOpeningVariantsOpen] = useState(false);
  const [openingBusy, setOpeningBusy] = useState(false);
  const [h3TraceBusy, setH3TraceBusy] = useState(false);
  const [h3Trace, setH3Trace] = useState(null);
  const [h3TraceError, setH3TraceError] = useState('');
  const [promptEditTrace, setPromptEditTrace] = useState(null);
  const [aiOriginalOpen, setAiOriginalOpen] = useState(false);
  const [directorPromptCandidate, setDirectorPromptCandidate] = useState(null);
  const [candidateModalOpen, setCandidateModalOpen] = useState(false);
  const [candidateBusy, setCandidateBusy] = useState(false);
  const assetThumbs = useAuthenticatedBookAssetThumbs(batchId, book);
  const selectedIndex = Math.max(0, videos.findIndex(video => video.id === selectedVideoId));
  const selectedVideo = videos[selectedIndex] || null;
  const selectedPrecompiledFrame = resolvePrecompiledStoryboardFrame(book, selectedVideoId || initialVideoId);
  const h3Document = h3DirectorDocument(book);
  const engineSettings = effectiveBookSettings(batch, book);
  const currentOpeningVariants = selectedVideo?.settingsState?.patch?.openingVariants || [];
  const currentOpeningCount = openingVariantSuccessCount(currentOpeningVariants);
  const otherVersionOpeningCount = videos.reduce((total, video) => (
    String(video?.id || '') === String(selectedVideo?.id || '')
      ? total
      : total + openingVariantSuccessCount(video?.settingsState?.patch?.openingVariants)
  ), 0);
  const openingMissing = [
    videos.length < 2 ? '分镜数需≥2' : '',
    !(book?.directorRevision?.id || book?.directorRevision) ? '需先有 DirectorRevision（先生成文案）' : '',
    engineSettings.openingEnabled !== true ? '需在统一配置开启「换开头」' : ''
  ].filter(Boolean);
  const openingReady = openingMissing.length === 0;
  async function generateOpeningVariants() {
    if (openingBusy || !openingReady) return;
    setOpeningBusy(true);
    try {
      await onGenerateOpening?.();
      // 生成是异步阶段任务，轮询本书版本直到变体落库；避免因为状态回读慢而误报“没反应”。
      for (let attempt = 0; attempt < 12; attempt += 1) {
        await new Promise(resolve => setTimeout(resolve, 1500));
        const response = await getBatch(batchId);
        const latestBook = (response?.batch?.books || []).find(item => String(item?.id || '') === String(book?.id || ''));
        const latestVideo = (latestBook?.videos || []).find(item => String(item?.id || '') === String(selectedVideo?.id || ''));
        if (openingVariantSuccessCount(latestVideo?.settingsState?.patch?.openingVariants) > currentOpeningCount) {
          message.success('换开头变体已生成，可在本弹窗查看。');
          await onSaved?.();
          return;
        }
      }
      message.warning('已提交生成，但当前版本暂未读到新的换开头变体；请稍后刷新或查看任务日志。');
      await onSaved?.();
    } catch (error) {
      message.error(error?.message || '换开头变体生成失败');
    } finally {
      setOpeningBusy(false);
    }
  }
  // 漫画模式没有导演分镜，不请求视频最终提示词；模式缺省为 'video'。
  const { displayPrompt, compiledPrompt, smartUnifiedPending, compilerError, loading: compilingPrompt } = useCompiledVideoPrompt(batchId, book, selectedVideo, settingsRevision, { enabled: productionMode !== 'comic' });
  const smartUnifiedAnalysis = smartUnifiedAnalysisForBook(book);
  const showSmartUnified = smartUnifiedDisplayEnabled(batch, book) && smartUnifiedAnalysis;
  const hasVisualPrompt = Boolean(String(comicPrompt || '').trim());
  useEffect(() => {
    const first = book?.videos?.find(video => video.id === initialVideoId) || book?.videos?.[0] || null;
    const firstH3Frame = resolvePrecompiledStoryboardFrame(book, initialVideoId);
    setSelectedVideoId(first?.id || firstH3Frame?.key || '');
    setVideoPrompt(first?.videoPrompt || '');
	setComicDirectorPrompt(comicPromptForImageAspect(engineSettings, first?.comicPrompt || ''));
	setComicPrompt(comicPromptForImageAspect(engineSettings, first?.comicImagePrompt || first?.comicPrompt || ''));
    setEditing(false);
  }, [book?.id, initialVideoId]);
  useEffect(() => {
    setVideoPrompt(selectedVideo?.videoPrompt || '');
	setComicDirectorPrompt(comicPromptForImageAspect(engineSettings, selectedVideo?.comicPrompt || ''));
	setComicPrompt(comicPromptForImageAspect(engineSettings, selectedVideo?.comicImagePrompt || selectedVideo?.comicPrompt || ''));
    setEditing(false);
    setAssetMentionMenu(null);
  }, [selectedVideoId, selectedVideo?.revision]);
  useEffect(() => {
    if (promptKind !== 'visual' || !isDeveloper || !batchId || !book?.id || !selectedVideo?.id) {
      setComicPromptPreview({ loading: false, compiledPrompt: '', referenceImageUrls: [], referenceImages: [], error: '' });
      return undefined;
    }
    let active = true;
    setComicPromptPreview(previous => ({ ...previous, loading: true, error: '' }));
    getComicPagePromptPreview(batchId, book.id, selectedVideo.id, {
      comicPrompt: comicPromptForImageAspect(engineSettings, selectedVideo.comicPrompt || ''),
      imagePrompt: comicPromptForImageAspect(engineSettings, selectedVideo.comicImagePrompt || selectedVideo.comicPrompt || '')
    }).then(result => {
      if (!active) return;
      setComicPromptPreview({
        loading: false,
        compiledPrompt: String(result?.compiledPrompt || ''),
        referenceImageUrls: Array.isArray(result?.referenceImageUrls) ? result.referenceImageUrls : [],
        referenceImages: Array.isArray(result?.referenceImages) ? result.referenceImages : [],
        error: ''
      });
    }).catch(error => {
      if (!active) return;
      setComicPromptPreview({ loading: false, compiledPrompt: '', referenceImageUrls: [], referenceImages: [], error: error?.message || '完整提示词读取失败' });
    });
    return () => { active = false; };
  }, [promptKind, isDeveloper, batchId, book?.id, selectedVideo?.id, selectedVideo?.revision, settingsRevision]);
  async function removeComicPageReference(reference) {
    if (!reference?.removable || !selectedVideo?.id || removingReferenceURL) return;
    setRemovingReferenceURL(reference.url);
    try {
      const nextReferences = removablePageReferences(comicPromptPreview.referenceImages, reference.url);
      await setComicPageReferenceImages(batchId, book.id, selectedVideo.id, nextReferences);
      setComicPromptPreview(previous => ({
        ...previous,
        referenceImageUrls: previous.referenceImageUrls.filter(item => item !== reference.url),
        referenceImages: previous.referenceImages.filter(item => item.url !== reference.url)
      }));
      message.success('当前页手动参考图已移除');
      await onSaved?.();
    } catch (error) {
      message.error(error?.message || '移除当前页参考图失败');
    } finally {
      setRemovingReferenceURL('');
    }
  }
  useEffect(() => {
    let active = true;
    if (!batchId || !book?.id) return () => { active = false; };
    getDirectorPromptCandidate(batchId, book.id)
      .then(response => {
        if (!active) return;
        setDirectorPromptCandidate(response?.candidate || response || null);
      })
      .catch(() => {
        if (active) setDirectorPromptCandidate(null);
      });
    return () => { active = false; };
  }, [batchId, book?.id, book?.revision]);
  useEffect(() => {
    if (!editing || promptKind !== 'video' || !promptEditableRef.current) return;
    promptEditableRef.current.innerHTML = renderStoryboardPromptHTML(videoPrompt, book, selectedVideo, assetThumbs);
  }, [editing, promptKind, selectedVideoId, selectedVideo?.revision, promptEditTrace, assetThumbs]);
  async function savePrompt() {
    if (!selectedVideo) return;
    setSaving(true);
    try {
      const savedVideoPrompt = promptKind === 'video'
        ? storyboardPromptTextForSave(promptEditableRef.current, videoPrompt)
        : videoPrompt;
      if (promptKind === 'video') setVideoPrompt(savedVideoPrompt);
		const patch = promptKind === 'visual' ? { comicPrompt: comicDirectorPrompt, comicImagePrompt: comicPrompt } : { videoPrompt: savedVideoPrompt };
      if (promptKind === 'visual') {
        // The comic_pages row is the single source of truth for page prompts.
        // Write it first so generation and the private-customization workspace
        // read the same values; only then mirror them onto the video patch for
        // the existing batch-factory display.
        const director = String(comicDirectorPrompt || '').trim();
        const image = String(comicPrompt || '').trim();
        if (!director || !image) {
          message.warning('漫画提示词与图片提示词都不能为空');
          setSaving(false);
          return;
        }
        await updateComicPage(batchId, book.id, selectedVideo.id, { comicPrompt: director, imagePrompt: image });
      }
      await saveVideoOverride(batchId, book.id, selectedVideo.id, { patch, expectedRevision: Number(selectedVideo.revision || 0) });
      let h3CompileWarning = '';
      if (promptKind === 'video' && (book?.directorRevision?.output?.h3_director || book?.directorRevision?.output?.h3Director) && promptEditTrace) {
        try {
          await compileH3Video(batchId, book.id, h3PromptEditRequest(promptEditTrace, selectedIndex, savedVideoPrompt));
        } catch (error) {
          h3CompileWarning = error?.message || '最终 VIDEO 编译记录暂未刷新';
        }
      }
      let refreshWarning = '';
      try {
        await onSaved?.();
      } catch (error) {
        refreshWarning = error?.message || '页面数据暂未刷新';
      }
      setEditing(false);
      setEditingSnapshot(null);
      message.success(promptKind === 'visual' ? '当前分镜漫画提示词已保存。' : '当前分镜视频提示词已保存。');
      if (h3CompileWarning) message.warning(`提示词已保存；${h3CompileWarning}`);
      if (refreshWarning) message.warning(`提示词已保存；${refreshWarning}`);
    } catch (error) { message.error(error?.message || '保存分镜提示词失败'); } finally { setSaving(false); }
  }
  async function openH3Trace() {
    setH3TraceOpen(true);
    setH3TraceBusy(true);
    setH3TraceError('');
    try {
      setH3Trace(await getH3Trace(batchId, book.id));
    } catch (error) {
      setH3Trace(null);
      setH3TraceError(error?.message || '读取 H3 Trace 失败');
    } finally {
      setH3TraceBusy(false);
    }
  }
  async function createPromptCandidate() {
    if (!selectedVideo || candidateBusy) return;
    setCandidateBusy(true);
    try {
      const response = await createDirectorPromptCandidate(batchId, book.id, Number(book.revision || 0));
      const candidate = response?.candidate || response;
      setDirectorPromptCandidate(candidate || null);
      setCandidateModalOpen(true);
    } catch (error) {
      message.error(error?.message || '重新生成 AI 失败');
    } finally {
      setCandidateBusy(false);
    }
  }
  async function acceptPromptCandidate() {
    const candidate = directorPromptCandidate;
    if (!selectedVideo || !candidate?.id || candidateBusy) return;
    setCandidateBusy(true);
    try {
      await acceptDirectorPromptCandidate(batchId, book.id, selectedVideo.id, candidate.id, Number(selectedVideo.revision || 0));
      await onSaved?.();
      setDirectorPromptCandidate(null);
      setCandidateModalOpen(false);
      message.success('已采用新的 AI 结果。');
    } catch (error) {
      message.error(error?.message || '采用新结果失败');
    } finally {
      setCandidateBusy(false);
    }
  }
  async function cancelPromptCandidate() {
    if (!directorPromptCandidate || candidateBusy) {
      setCandidateModalOpen(false);
      return;
    }
    setCandidateBusy(true);
    try {
      const response = await getBatch(batchId);
      const latestBatch = response?.batch || response;
      const latestBook = (latestBatch?.books || []).find(item => item?.id === book.id);
      await cancelDirectorPromptCandidate(batchId, book.id, Number(latestBook?.revision || book.revision || 0));
      await onSaved?.();
      setDirectorPromptCandidate(null);
      setCandidateModalOpen(false);
      message.success('已取消新的 AI 结果。');
    } catch (error) {
      message.error(error?.message || '取消新的 AI 结果失败');
    } finally {
      setCandidateBusy(false);
    }
  }
  function move(direction) {
    const next = videos[selectedIndex + direction];
    if (next) setSelectedVideoId(next.id);
  }

  function movePrecompiledFrame(direction) {
    const frame = precompiledWorkspace.frames?.[(selectedPrecompiledFrame?.index || 0) + direction];
    if (frame) setSelectedVideoId(frame.key);
  }

  function updatePromptValue(event) {
    const nextValue = event.target.value;
    if (promptKind === 'visual') { setComicPrompt(nextValue); return; }
    setVideoPrompt(nextValue);
    const cursor = Number(event.target.selectionStart || nextValue.length);
    const match = nextValue.slice(0, cursor).match(/@([^\s@]*)$/);
    if (!match) { setAssetMentionMenu(null); return; }
    const query = match[1].toLowerCase();
    const candidates = (book?.assetRecords || []).filter(asset => asset?.id && String(assetName(asset) || '').toLowerCase().includes(query));
    setAssetMentionMenu({ start: cursor - match[0].length, end: cursor, candidates });
  }

  function updateEditablePromptValue(event) {
    const root = event.currentTarget;
    const nextValue = contentEditablePromptText(root);
    setVideoPrompt(nextValue);
    updateAssetMentionMenu(root, nextValue);
  }

  function updateAssetMentionMenu(root, value = String(root?.innerText || '').replace(/\u00a0/g, ' ')) {
    let cursor = contentEditableTextOffset(root);
    let match = value.slice(0, cursor).match(/@([^\s@]*)$/);
    // Some browsers keep the old selection when a contenteditable is
    // rehydrated with rich mention spans.  If the actual input ends in an @
    // query, prefer that real text position instead of silently closing the
    // picker because the stale selection points into the previous span.
    const trailingMatch = value.match(/@([^\s@]*)$/);
    if (!match && trailingMatch) {
      cursor = value.length;
      match = trailingMatch;
    }
    if (!match) { setAssetMentionMenu(null); return; }
    const query = match[1].toLowerCase();
    const candidates = (book?.assetRecords || []).filter(asset => asset?.id && String(assetName(asset) || '').toLowerCase().includes(query));
    const selection = typeof window !== 'undefined' ? window.getSelection() : null;
    const range = selection?.rangeCount ? selection.getRangeAt(0) : null;
    const parent = root?.parentElement;
    const caretRect = range?.getBoundingClientRect?.();
    const parentRect = parent?.getBoundingClientRect?.();
    const left = caretRect && parentRect ? Math.max(8, caretRect.left - parentRect.left) : 8;
    const top = caretRect && parentRect ? Math.max(8, caretRect.bottom - parentRect.top + 6) : 8;
    setAssetMentionMenu({ start: cursor - match[0].length, end: cursor, candidates, left, top });
  }

  function handleEditablePromptKeyUp(event) {
    if (event.key === '@' || event.key.length === 1) updateAssetMentionMenu(event.currentTarget);
  }

  function handleEditablePromptKeyDown(event) {
    if (event.key === '@') requestAnimationFrame(() => updateAssetMentionMenu(event.currentTarget));
  }

  function cancelEditing() {
    if (editingSnapshot) {
      setVideoPrompt(editingSnapshot.videoPrompt);
      setComicDirectorPrompt(editingSnapshot.comicDirectorPrompt);
      setComicPrompt(editingSnapshot.comicPrompt);
    }
    setEditingSnapshot(null);
    setAssetMentionMenu(null);
    setEditing(false);
  }

  async function toggleEditing() {
    if (editing) {
      setEditing(false);
      setAssetMentionMenu(null);
      return;
    }
    setEditingSnapshot({ videoPrompt, comicDirectorPrompt, comicPrompt });
    if (promptKind === 'video' && (book?.directorRevision?.output?.h3_director || book?.directorRevision?.output?.h3Director)) {
      try {
        const trace = await getH3Trace(batchId, book.id);
        const segment = trace?.compilation?.compilation?.segments?.[selectedIndex];
        if (!segment) throw new Error('分镜编译记录缺失');
        setPromptEditTrace(trace);
        setVideoPrompt(segment.compiled_prompt);
      } catch (_) { setPromptEditTrace(null); setVideoPrompt(submittedVideoPrompt); }
    } else if (promptKind === 'video') setVideoPrompt(submittedVideoPrompt);
    setEditing(true);
  }

  function insertAssetMention(asset) {
    const menu = assetMentionMenu;
    if (!menu || !asset) return;
    const value = videoPrompt;
    const mention = `@${assetName(asset)} `;
    const next = `${value.slice(0, menu.start)}${mention}${value.slice(menu.end)}`;
    setVideoPrompt(next);
    setAssetMentionMenu(null);
    requestAnimationFrame(() => {
      if (promptEditableRef.current) {
        promptEditableRef.current.innerHTML = renderStoryboardPromptHTML(next, book, selectedVideo, assetThumbs);
        setContentEditableTextOffset(promptEditableRef.current, menu.start + mention.length);
        return;
      }
      const input = promptInputRef.current?.resizableTextArea?.textArea || promptInputRef.current?.textArea;
      if (input) { input.focus(); const position = menu.start + mention.length; input.setSelectionRange(position, position); }
    });
  }
	  const activeLabel = promptKind === 'visual' ? '漫画提示词' : '分镜视频提示词';
  const rawSubmittedVideoPrompt = compiledPrompt || displayPrompt;
  const rawPromptWithUnifiedStyle = showSmartUnified
    ? prependSmartUnifiedStyle(rawSubmittedVideoPrompt, smartUnifiedAnalysis.prompt)
    : rawSubmittedVideoPrompt;
  const submittedVideoPrompt = stripBaseSetupSection(rawPromptWithUnifiedStyle);
  const displaySubmittedVideoPrompt = baseSetupEnabled(engineSettings)
    ? composeStoryboardDisplayPrompt(submittedVideoPrompt, storyboardAssetPromptText(book, selectedVideo))
    : submittedVideoPrompt;
  const activeValue = promptKind === 'visual' ? comicPrompt : editing ? videoPrompt : displaySubmittedVideoPrompt;
  const activeReady = promptKind === 'visual' ? hasVisualPrompt : Boolean(displaySubmittedVideoPrompt.trim());
  const aiOriginalVideoPrompt = String(selectedVideo?.aiOriginalVideoPrompt || selectedVideo?.videoPrompt || '');
  const candidateVideoPrompt = String(directorPromptCandidate?.videoPrompts?.[selectedVideo?.id] || '');
  const h3Segments = h3Trace?.compilation?.compilation?.segments || [];
  const h3Segment = h3Segments[selectedIndex] || null;
	  const h3ConstraintRows = h3ConstraintPreview(engineSettings);
  const h3AssetPromptText = baseSetupEnabled(engineSettings) ? storyboardAssetPromptText(book, selectedVideo) : '';
  const apiDurationSeconds = Number(selectedVideo?.durationSeconds || 0);
  const realDurationSeconds = realVideoDurationSeconds(submittedVideoPrompt);
  const durationLabel = realDurationSeconds !== null
    ? `${realDurationSeconds.toFixed(3)}s / ${apiDurationSeconds}s`
    : apiDurationSeconds ? `${apiDurationSeconds}s` : '—';
  const textModelLabel = String(engineSettings.textModelId || '—');
  const videoModelLabel = String(engineSettings.videoModelId || '—');
  if (!selectedVideo && selectedPrecompiledFrame) {
    const card = selectedPrecompiledFrame.card || {};
    return <div className="batch-factory-prompt-modal-stack">
      <div className="batch-factory-prompt-modal-head"><Space><Tooltip title="上一张导演分镜卡"><Button aria-label="上一张导演分镜卡" icon={<LeftOutlined />} disabled={selectedPrecompiledFrame.index === 0} onClick={() => movePrecompiledFrame(-1)} /></Tooltip><span className="batch-factory-prompt-modal-index">{selectedPrecompiledFrame.index + 1}/{precompiledWorkspace.frames.length}</span><Tooltip title="下一张导演分镜卡"><Button aria-label="下一张导演分镜卡" icon={<RightOutlined />} disabled={selectedPrecompiledFrame.index >= precompiledWorkspace.frames.length - 1} onClick={() => movePrecompiledFrame(1)} /></Tooltip></Space><span className="batch-factory-prompt-status">导演分镜卡 · 待编译 VIDEO</span></div>
      <Alert type="info" showIcon message="当前是导演分镜结构，不是最终 VIDEO 提示词" description="提示词预设决定内容和格式；视频模型只负责最终生成。开启跟随配音后会先读取每行真实时长，再按 10/15 秒确定性切段并填充最终 VIDEO Prompt。" />
			<label className="shuihuo-form-label batch-factory-prompt-editor-label">分镜视频提示词<Input.TextArea rows={18} readOnly value={`${formatH3DirectorCardPrompt(h3Document, card)}${h3AssetPromptText ? `\n\n【资产提示词】\n${h3AssetPromptText}` : ''}${h3ConstraintRows.length ? `\n\n【当前已启用约束】\n${h3ConstraintRows.join('\n')}` : ''}`} /></label>
      <div className="batch-factory-prompt-modal-actions">
        <Tooltip title="等待最终 VIDEO 编译后才能保存"><Button type="primary" disabled>保存</Button></Tooltip>
        <Tooltip title="复用当前导演分镜卡、真实配音时长和当前约束设置，编译最终 VIDEO 提示词；不会重新调用导演模型。"><Button type="primary" loading={regenerating} disabled={regenerating || !onCompile} onClick={onCompile}>编译最终提示词</Button></Tooltip>
        <Tooltip title="等待最终 VIDEO 编译后才能编辑"><Button disabled>编辑</Button></Tooltip>
        <Tooltip title="等待最终 VIDEO 编译"><Button disabled>生成视频</Button></Tooltip>
        <Button type="text" disabled={regenerating} onClick={() => onRetry?.()}>重试</Button>
      </div>
      <p className="shuihuo-modal-note">此处与当前行的资产、分镜视频共用同一张导演分镜卡；最终 VIDEO 生成前，不会伪造可播放视频。</p>
      <Modal title="分镜编译与提交 Trace" open={h3TraceOpen} onCancel={() => setH3TraceOpen(false)} footer={null} width={980}>{h3TraceBusy ? <Alert type="info" showIcon message="正在读取编译 Trace…" /> : h3TraceError ? <Alert type="error" showIcon message="编译 Trace 读取失败" description={h3TraceError} /> : <Alert type="info" showIcon message="当前导演卡尚未产生最终 VIDEO 编译记录" />}</Modal>
    </div>;
  }
  return <div className="batch-factory-prompt-modal-stack">
    <div className="batch-factory-prompt-modal-head">
      <Space>
        <Tooltip title="上一分镜"><Button aria-label="上一分镜" icon={<LeftOutlined />} disabled={selectedIndex === 0} onClick={() => move(-1)} /></Tooltip>
        <span className="batch-factory-prompt-modal-index">{videos.length ? `${selectedIndex + 1}/${videos.length}` : '0/0'}</span>
        <Tooltip title="下一分镜"><Button aria-label="下一分镜" icon={<RightOutlined />} disabled={!videos.length || selectedIndex >= videos.length - 1} onClick={() => move(1)} /></Tooltip>
      </Space>
      <span className={`batch-factory-prompt-mode-label is-${productionMode}`}>{productionMode === 'comic' ? '漫画提示词' : '视频提示词'}</span>
    </div>
    <div className="batch-factory-prompt-meta-bar">
      <span className="batch-factory-prompt-meta-item"><b>真实/API时长：</b>{durationLabel}</span>
      <span className="batch-factory-prompt-meta-item"><b>文本模型：</b>{textModelLabel}</span>
      <span className="batch-factory-prompt-meta-item"><b>视频模型：</b>{videoModelLabel}</span>
    </div>
    {!activeReady ? <Alert type={compilerError && promptKind === 'video' ? 'error' : 'info'} showIcon message={compilerError && promptKind === 'video' ? '最终视频提示词读取失败' : smartUnifiedPending && promptKind === 'video' ? '智能统一待分析' : promptKind === 'visual' ? '请生成漫画提示词再查看' : '请生成视频提示词再查看'} description={compilerError && promptKind === 'video' ? '没有使用旧的已保存提示词冒充当前结果，请刷新页面后重试。' : smartUnifiedPending && promptKind === 'video' ? '智能统一与资产提示词分开获取；未取得基线不会阻断资产、分镜或视频。可在资产区单独刷新。' : `当前处于${productionMode === 'comic' ? '漫画' : '视频'}生产模式，请先生成对应提示词。`} /> : <>
      <label className="shuihuo-form-label batch-factory-prompt-editor-label">
        {activeLabel}
        {promptKind === 'visual' && !editing ? <div className="batch-factory-comic-prompt-preview">
          <div className="batch-factory-comic-prompt-section"><b>分格导演描述</b><Input.TextArea rows={10} readOnly value={comicDirectorPrompt} /></div>
          <div className="batch-factory-comic-prompt-section"><b>本页基础出图描述</b><Input.TextArea rows={10} readOnly value={comicPrompt} /></div>
          {isDeveloper ? <div className="batch-factory-comic-developer-preview">
            <Alert type="info" showIcon message="最终完整出图提示词" description="完整提示词会包含画风、全局前后缀、画幅与分格结构、负面限制，以及本页分格导演描述和基础出图描述。" />
            {comicPromptPreview.error ? <Alert type="error" showIcon message="完整提示词读取失败" description={comicPromptPreview.error} /> : <Input.TextArea rows={18} readOnly value={comicPromptPreview.compiledPrompt} placeholder={comicPromptPreview.loading ? '正在整理最终完整提示词…' : '当前没有可展示的完整提示词'} />}
            <ComicReferenceChips referenceImages={comicPromptPreview.referenceImages} onRemove={removeComicPageReference} removingURL={removingReferenceURL} />
          </div> : null}
        </div> : editing && promptKind === 'visual' ? <div className="batch-factory-comic-prompt-preview">
          <div className="batch-factory-prompt-editor-wrap"><span>分格导演描述（可编辑）</span><Input.TextArea rows={10} value={comicDirectorPrompt} onChange={event => setComicDirectorPrompt(event.target.value)} placeholder="描述本页分格、镜头、人物动作和对白安排" /></div>
          <div className="batch-factory-prompt-editor-wrap"><span>本页基础出图描述（可编辑）</span><Input.TextArea ref={promptInputRef} rows={14} value={comicPrompt} onChange={event => setComicPrompt(event.target.value)} placeholder="描述当前漫画页要生成的画面内容" /></div>
        </div> : !editing && promptKind === 'video' ? <div className="batch-factory-prompt-rich-view">{renderStoryboardPromptWithAssetMentions(activeValue, book, selectedVideo, assetThumbs)}</div> : editing && promptKind === 'video' ? <div className="batch-factory-prompt-editor-rich-wrap">
          <div ref={promptEditableRef} className="batch-factory-prompt-rich-view is-editing" contentEditable suppressContentEditableWarning role="textbox" aria-label="可编辑分镜视频提示词" onInput={updateEditablePromptValue} onKeyDown={handleEditablePromptKeyDown} onKeyUp={handleEditablePromptKeyUp} onBlur={() => setTimeout(() => setAssetMentionMenu(null), 120)} />
          {assetMentionMenu ? <div className="batch-factory-asset-mention-menu" style={{ left: assetMentionMenu.left, top: assetMentionMenu.top }}>{assetMentionMenu.candidates?.length ? assetMentionMenu.candidates.map(asset => { const imageUrl = assetThumbs[String(asset.id)] || asset.imageUrl || asset.image_url || asset.thumbnailUrl; const kindLabel = asset.kind === 'character' ? '角色' : asset.kind === 'scene' ? '场景' : '道具'; return <button type="button" key={asset.id} onMouseDown={event => event.preventDefault()} onClick={() => insertAssetMention(asset)}>{imageUrl ? <AuthenticatedBookAssetImage imageUrl={imageUrl} alt="" /> : <span className="batch-factory-asset-mention-fallback" aria-hidden="true">{asset.kind === 'character' ? '人' : asset.kind === 'scene' ? '景' : '物'}</span>}<span>{assetDisplayLabel(asset)}</span><small>{kindLabel}</small></button>; }) : <div className="batch-factory-asset-mention-empty">暂无可用资产，请先在资产设置添加人物、场景或道具</div>}</div> : null}
        </div> : null}
      </label>
      <div className="batch-factory-prompt-modal-actions">
        <Button type="primary" loading={saving} disabled={!editing || !selectedVideo} onClick={savePrompt}>保存</Button>
        <Tooltip title={promptKind === 'visual' ? '按当前小说原文重新生成漫画提示词。' : '生成一份新的 AI 候选结果；采用前不会修改当前提示词。'}><Button icon={promptKind === 'visual' ? <ReloadOutlined /> : null} loading={promptKind === 'visual' ? regenerating : candidateBusy} disabled={promptKind === 'visual' ? regenerating || !onRegenerateVisual : candidateBusy || !selectedVideo} onClick={promptKind === 'visual' ? onRegenerateVisual : createPromptCandidate}>{promptKind === 'visual' ? '重新生成漫画提示词' : '重新生成 AI'}</Button></Tooltip>
        {promptKind === 'visual' ? <Popconfirm title="重新生成全部漫画提示词" description="这会按当前原文和提示词配置，覆盖本书所有漫画页的提示词。" okText="重新生成" cancelText="取消" onConfirm={onRegenerateAllVisual}><Button icon={<ReloadOutlined />} loading={regenerating} disabled={regenerating || !onRegenerateAllVisual}>重新生成全部漫画提示词</Button></Popconfirm> : null}
        {promptKind === 'visual' ? <Tooltip title="只下载本书全部最终完整出图提示词 Markdown 和资产图片"><Button loading={comicZipDownloading} disabled={comicZipDownloading || regenerating || !onDownloadAllComicZip} onClick={onDownloadAllComicZip}>下载最终提示词+资产图片 ZIP</Button></Tooltip> : null}
        {editing ? <Button onClick={cancelEditing}>取消编辑</Button> : null}
        {!editing ? <Button onClick={toggleEditing}>编辑</Button> : null}
        {promptKind === 'video' && selectedIndex === 0 && engineSettings.openingEnabled === true ? <Button onClick={() => setOpeningVariantsOpen(true)}>查看换开头提示词</Button> : null}
        {promptKind === 'video' ? <Button onClick={() => setAiOriginalOpen(true)}>查看 AI 原始输出</Button> : null}
        <Button type="text" disabled={regenerating} onClick={() => onRetry?.(selectedVideo?.id)}>重试</Button>
        <span className="batch-factory-prompt-primary-action">{promptKind === 'visual' ? <Button type="primary" disabled={!hasVisualPrompt || regenerating} loading={regenerating} onClick={() => onGenerateVisual?.(selectedVideo?.id, comicPrompt)}>生成本页漫画</Button> : <Tooltip title={productionAvailable ? '为当前分镜生成视频候选版本。' : productionReason}><Button type="primary" disabled={!productionAvailable || regenerating} loading={regenerating} onClick={() => onGenerateVideo?.(selectedVideo?.id)}>生成视频</Button></Tooltip>}</span>
      </div>
    </>}
    <p className="shuihuo-modal-note">{promptKind === 'visual' ? '漫画版本从当前书的漫画资产和提示词预设中生成。' : editing ? '编辑的是当前分镜的视频描述；保存后会按已启用资产与约束重新编译最终视频提示词。' : compilingPrompt ? '正在读取已启用资产与约束编译的最终视频提示词。' : '视频版本只在当前分镜内生成和切换，不会覆盖其他分镜。'}</p>
    <Modal title="AI 原始输出" open={aiOriginalOpen} onCancel={() => setAiOriginalOpen(false)} footer={null} width={860}>
      <Input.TextArea rows={18} readOnly value={localizeShotKeywords(aiOriginalVideoPrompt)} placeholder="当前分镜尚未有 AI 原始输出" />
    </Modal>
    <Modal title="重新生成 AI 对比" open={candidateModalOpen} onCancel={cancelPromptCandidate} footer={null} width={1080} destroyOnClose>
      {candidateVideoPrompt ? <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'minmax(0, 1fr) minmax(0, 1fr)', gap: 16 }}>
          <label className="shuihuo-form-label">当前分镜视频提示词<Input.TextArea rows={18} readOnly value={String(selectedVideo?.videoPrompt || videoPrompt || '')} /></label>
          <label className="shuihuo-form-label">新的 AI 重生成候选<Input.TextArea rows={18} readOnly value={candidateVideoPrompt} /></label>
        </div>
        <Space><Button type="primary" loading={candidateBusy} onClick={acceptPromptCandidate}>采用新结果</Button><Button loading={candidateBusy} onClick={cancelPromptCandidate}>取消</Button></Space>
      </Space> : <Alert type="info" showIcon message="当前分镜没有可用的新 AI 候选" />}
    </Modal>
    <Modal title="分镜编译与提交 Trace" open={h3TraceOpen} onCancel={() => setH3TraceOpen(false)} footer={null} width={980}>
      {h3TraceBusy ? <Alert type="info" showIcon message="正在读取编译 Trace…" /> : h3TraceError ? <Alert type="error" showIcon message="编译 Trace 读取失败" description={h3TraceError} /> : h3Trace?.legacy ? <Alert type="warning" showIcon message="旧数据只读" description={h3Trace?.notice || '旧版导演数据，无完整结构化 Trace'} /> : h3Segment ? <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <Descriptions size="small" bordered column={2} items={[
          { key: 'director', label: '导演修订', children: h3Trace?.director_revision_id || '—' },
          { key: 'compilation', label: '编译修订', children: h3Trace?.compilation?.id || '—' },
          { key: 'segment', label: '分段', children: h3Segment.segment_key || `VIDEO ${selectedIndex + 1}` },
          { key: 'duration', label: '确定性时长', children: `${h3Segment.canonical_duration_ms || 0}ms → ${h3Segment.request_duration_ms || 0}ms` },
          { key: 'hash', label: 'Prompt Hash', children: h3Segment.compiled_prompt_hash || '—', span: 2 }
        ]} />
        <label className="shuihuo-form-label">实际提交视频模型的最终 VIDEO Prompt<Input.TextArea rows={16} readOnly value={h3Segment.compiled_prompt || ''} /></label>
        <label className="shuihuo-form-label">Compile Trace<Input.TextArea rows={12} readOnly value={JSON.stringify(h3Segment.compile_trace || {}, null, 2)} /></label>
      </Space> : <Alert type="info" showIcon message="当前分镜尚无编译记录" />}
    </Modal>
    <Modal title="分镜一 | 换开头提示词" open={openingVariantsOpen} onCancel={() => setOpeningVariantsOpen(false)} footer={null} width={860}>
      <Space direction="vertical" size={12} style={{ width: '100%' }}>
        <Card size="small" title="分镜一（原始）"><pre style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{submittedVideoPrompt || '—'}</pre></Card>
        {(selectedVideo?.settingsState?.patch?.openingVariants || []).map(variant => (
          <Card key={variant.index} size="small" title={variant.label || `分镜一 | 换开头${variant.index}`} extra={<Tag color={variant.status === 'success' ? 'success' : 'error'}>{variant.status === 'success' ? '成功' : '失败'}</Tag>}>
            {variant.status === 'success' ? <pre style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{variant.prompt}</pre> : <Alert type="error" showIcon message="AI 未生成出该变体，已阻断后续生产" description={variant.failureReason || '历史记录未保留失败原因；请在任务/日志中查看并重试失败步骤。'} />}
          </Card>
        ))}
        {currentOpeningCount === 0 ? (
          <Alert
            type="info"
            showIcon
            message={otherVersionOpeningCount > 0
              ? `当前版本（VIDEO${String(selectedIndex + 1).padStart(2, '0')}）未生成换开头变体（其他版本已有 ${otherVersionOpeningCount} 条），可切换上方版本查看。`
              : '换开头变体尚未生成'}
            description={openingReady ? '将按当前分镜一提示词生成换开头变体。' : `暂不能生成：${openingMissing.join('；')}`}
            action={(
              <Button
                size="small"
                type="primary"
                loading={openingBusy || generatingOpening}
                disabled={!openingReady || openingBusy || generatingOpening}
                title={openingReady ? '生成换开头变体' : `暂不能生成：${openingMissing.join('；')}`}
                onClick={generateOpeningVariants}
              >生成换开头变体</Button>
            )}
          />
        ) : null}
      </Space>
    </Modal>
  </div>;
}
function PassiveMediaPoster({ className = 'batch-factory-media-poster', label, detail = '已生成 · 点击查看' }) {
  return <span className={className} role="img" aria-label={label}><PictureOutlined /><b>{label}</b><small>{detail}</small></span>;
}

// merge-cover 需要鉴权头，<img> 无法直接携带。这里用已鉴权的 API 客户端
// 取小尺寸封面 blob，并按 URL 模块级缓存，避免列表和版本面板重复请求。
const mergeCoverCache = new Map();

function loadMergeCoverObjectURL(url) {
  const cached = mergeCoverCache.get(url);
  if (cached) return cached;
  const pending = getBatchFactoryMergeCoverBlob(url).then(blob => {
    if (typeof URL?.createObjectURL !== 'function') throw new Error('当前浏览器无法创建封面预览地址');
    if (mergeCoverCache.size >= 64) mergeCoverCache.delete(mergeCoverCache.keys().next().value);
    return URL.createObjectURL(blob);
  }).catch(error => {
    mergeCoverCache.delete(url);
    throw error;
  });
  mergeCoverCache.set(url, pending);
  return pending;
}

function MergeCoverThumb({ url, label, alt = '' }) {
  const [src, setSrc] = useState('');
  useEffect(() => {
    let cancelled = false;
    setSrc('');
    if (!url) return undefined;
    loadMergeCoverObjectURL(url)
      .then(objectURL => { if (!cancelled) setSrc(objectURL); })
      .catch(() => { if (!cancelled) setSrc(''); });
    return () => { cancelled = true; };
  }, [url]);
  if (!src) return <PictureOutlined aria-label={label || '封面加载中'} />;
  return <img src={src} alt={alt || label || ''} loading="lazy" />;
}

function PassiveMediaPreview({
  media,
  className = 'batch-factory-media-poster',
  label = '视频预览',
  detail = '已生成',
  onLoadedMetadata
}) {
  const mediaUrl = batchFactoryPassiveMediaUrl(media);
  const posterUrl = batchFactoryPassivePosterUrl(media);
  const mergeMedia = batchFactoryIsMergeMedia(media);

  if (!mediaUrl && !posterUrl) {
    return <PassiveMediaPoster className={className} label={label} detail={detail} />;
  }

  const primeVideoFrame = event => {
    onLoadedMetadata?.(event);
    const element = event.currentTarget;
    const duration = Number(element?.duration || 0);
    const target = Number.isFinite(duration) && duration > 0 ? Math.min(0.08, duration / 2) : 0.05;
    try {
      if (element && Number(element.currentTime || 0) < 0.001 && target > 0) element.currentTime = target;
    } catch (_) {}
  };

  const ensureDecodedFrame = event => {
    const element = event.currentTarget;
    const duration = Number(element?.duration || 0);
    const target = Number.isFinite(duration) && duration > 0 ? Math.min(0.08, duration / 2) : 0.05;
    try {
      if (element && Number(element.currentTime || 0) < 0.001 && target > 0) element.currentTime = target;
    } catch (_) {}
  };

  return <span className={`${className} has-media`} role="img" aria-label={label}>
    {posterUrl
      ? (isBatchFactoryMergeCoverURL(posterUrl)
        ? <MergeCoverThumb url={posterUrl} label={label} />
        : <img src={posterUrl} alt="" loading="lazy" />)
      : mergeMedia
        // 合成成片绝不拿整段 mp4 当缩略图；封面不可用时只显示占位。
        ? <PassiveMediaPoster className={className} label={label} detail={detail} />
        : <ProductionMediaBoundary showDownload={false}><video
            src={mediaUrl}
            muted
            playsInline
            preload="metadata"
            disablePictureInPicture
            tabIndex={-1}
            aria-hidden="true"
            onLoadedMetadata={primeVideoFrame}
            onLoadedData={ensureDecodedFrame}
          /></ProductionMediaBoundary>}
  </span>;
}





function InlineMediaLibrary({ book, productionStatus, mergeJob, aspectRatio = '16:9', h3CompilationMessage = '将按 10 秒确定性切段并编译最终 VIDEO', selectedVideoId: controlledSelectedVideoId = '', onSelectedVideoChange, onManage, onOpenMerge }) {
  const videos = book?.videos || [];
  const precompiledWorkspace = resolvePrecompiledVideoWorkspace(book);
  // 历史轮次（orphaned）的成功任务按 ordinal 挂回同一分镜，作为「上一版」展示。
  const mediaVersions = bookVideoVersions(book, productionStatus);
  const videoProgress = batchFactoryVideoProgress(book, productionStatus);
  const [localSelectedVideoId, setLocalSelectedVideoId] = useState(videos[0]?.id || '');
  const selectedVideoId = controlledSelectedVideoId || localSelectedVideoId;
  const selectVideo = nextId => {
    const value = String(nextId || '');
    setLocalSelectedVideoId(value);
    onSelectedVideoChange?.(value);
  };
  const selectedIndex = Math.max(0, videos.findIndex(video => video.id === selectedVideoId));
  const video = videos[selectedIndex] || null;
  const versions = video ? (mediaVersions.get(video.id) || []) : [];
  const primary = primaryMediaVersion(video, versions);
  const currentProgress = video ? videoProgress.byVideo.get(String(video.id || '')) : null;
  const mediaFrameLayout = batchFactoryMediaOrientation(batchFactoryParseMediaRatio(video?.settingsState?.patch?.aspectRatio || video?.aspectRatio || aspectRatio, 16 / 9));
  const candidates = versions.filter(version => version.id !== primary?.id).slice(0, 4);
  useEffect(() => { if (!controlledSelectedVideoId) setLocalSelectedVideoId(book?.videos?.[0]?.id || ''); }, [book?.id]);
  useEffect(() => { if (!videos.some(item => item.id === selectedVideoId) && videos[0]?.id) selectVideo(videos[0].id); }, [videos.map(item => item.id).join('|')]);
  function move(direction) { const next = videos[selectedIndex + direction]; if (next) selectVideo(next.id); }
  if (!video && precompiledWorkspace.status === 'awaiting_compilation') return <button type="button" className="batch-factory-cell-shell batch-factory-inline-media is-empty" title="打开待编译的分镜视频" onClick={() => onManage?.(selectedVideoId)}>
    <header className="batch-factory-cell-head"><div><b>分镜视频</b><small>导演分镜卡已提取，尚未编译 VIDEO</small></div><span className="batch-factory-cell-state">待编译</span></header>
    <div className="batch-factory-cell-content batch-factory-video-content"><div className="batch-factory-cell-empty"><CloudUploadOutlined /><span>点击查看待编译的分镜视频；{h3CompilationMessage}。</span></div></div>
    <div className="batch-factory-cell-pager"><span>{precompiledWorkspace.cards.length} 张导演分镜卡待编译 · 点击进入</span></div>
  </button>;
  if (!video) return <div className="batch-factory-cell-shell batch-factory-inline-media is-empty"><div className="batch-factory-cell-empty">AI 推理后显示该书的分镜视频。</div></div>;
  const running = currentProgress?.status === 'running' || currentProgress?.status === 'queued';
  const mergeStatus = String(mergeJob?.status || '').toLowerCase();
  const mergeReady = mergeStatus === 'succeeded' && String(mergeJob?.outputUrl || '').trim();
  const mergeText = mergeReady ? `最终成片 ${batchFactoryStableMediaLabel('M', mergeJob.id)} · ${mergeJobSpeed(mergeJob).toFixed(1)}x`
    : mergeStatus === 'running' || mergeStatus === 'queued' ? '最终成片 · 合成中'
    : mergeStatus === 'failed' ? '最终成片 · 合成失败' : '最终成片 · 尚未合成';
  return <div className={`batch-factory-cell-shell batch-factory-inline-media is-${mediaFrameLayout}`}>
    <header className="batch-factory-cell-head">
      <div><b>{storyboardVideoLabel(video, selectedIndex)}</b><small>{running ? '正在生成' : primary?.mediaUrl ? '当前主版本' : '尚无可用视频'}</small></div>
      <span className={`batch-factory-cell-state${primary?.mediaUrl ? ' is-ready' : running ? ' is-running' : ''}`}>{primary?.mediaUrl ? batchFactoryStableMediaLabel('V', primary.id) : running ? '生成中' : '待生成'}</span>
    </header>
    <div className="batch-factory-cell-content batch-factory-video-content batch-factory-video-content-v13">
      <button type="button" className={`batch-factory-inline-media-primary${running ? ' is-producing' : ''}`} onClick={() => onManage?.(video.id)}>
        {primary?.mediaUrl ? <PassiveMediaPreview className="batch-factory-inline-media-poster" media={primary} label={`分镜 ${selectedIndex + 1} 已生成`} /> : <>{running ? <CmLoader size={28} title="分镜视频生成中" /> : <CloudUploadOutlined />}<span>{currentProgress?.message || '当前分镜暂无视频'}</span></>}
      </button>
      <div className="batch-factory-inline-media-versions"><span>候选</span>{candidates.map((candidate, index) => <button type="button" key={candidate.id} className="batch-factory-inline-media-version" onClick={() => onManage?.(video.id)}>{candidate.mediaUrl ? <PassiveMediaPreview className="batch-factory-inline-media-poster is-candidate" media={candidate} label={`候选 ${index + 1}`} detail="已生成" /> : <PictureOutlined />}</button>)}<button type="button" className="batch-factory-inline-media-version is-more" onClick={() => onManage?.(video.id)} title="打开完整片段库">•••</button></div>
      <button type="button" className={`batch-factory-inline-merge-status is-${mergeStatus || 'idle'}`} onClick={() => onOpenMerge?.()}><span>{mergeText}</span><b>查看</b></button>
    </div>
    <div className="batch-factory-cell-pager"><button type="button" disabled={selectedIndex === 0} onClick={() => move(-1)}><LeftOutlined /></button><span>{selectedIndex + 1} / {videos.length}</span><button type="button" disabled={selectedIndex >= videos.length - 1} onClick={() => move(1)}><RightOutlined /></button></div>
  </div>;
}

function InlineComicLibrary({ batchId, book, selectedVideoId: controlledSelectedVideoId = '', onSelectedVideoChange, onManage, onChanged, refreshNonce = 0, followUntil = 0 }) {
  const videos = book?.videos || [];
  const [pages, setPages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [loadError, setLoadError] = useState('');
  const [reloadKey, setReloadKey] = useState(0);
  const [localSelectedVideoId, setLocalSelectedVideoId] = useState(videos[0]?.id || '');
  const selectedVideoId = controlledSelectedVideoId || localSelectedVideoId;
  const selectedIndex = Math.max(0, videos.findIndex(video => video.id === selectedVideoId));
  const selectedVideo = videos[selectedIndex] || null;
  const page = pages.find(item => item.videoId === selectedVideo?.id) || null;
  const primary = page?.images?.find(image => image.isPrimary) || null;
  const candidates = (page?.images || []).filter(image => image.id !== primary?.id).slice(-4).reverse();
  const presentation = comicInlinePresentation({ page, selectedVideo, loading, error: loadError, hasPrimary: Boolean(primary?.url) });
  // 列表首屏只取状态与图片地址：提示词正文单本约 300KB，行内缩略图完全用不到。
  const inlineView = { view: 'lite', silent: true, suppressGlobalError: true };
  const selectVideo = nextId => {
    setLocalSelectedVideoId(nextId);
    onSelectedVideoChange?.(nextId);
  };
  useEffect(() => {
    let active = true;
    if (!batchId || !book?.id) return () => { active = false; };
    setLoading(true);
    // 失败时保留上一轮数据，只有确实没有数据时才显示“读取失败”，
    // 避免一次网络抖动把整行永久清空成“暂无漫画页”。
    listComicPages(batchId, book.id, inlineView).then(pageResponse => {
      if (active) { setPages(resultData(pageResponse, 'pages') || []); setLoadError(''); }
    }).catch(error => { if (active) setLoadError(error?.message || '漫画页读取失败'); }).finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [batchId, book?.id, refreshNonce, reloadKey]);
  // 该行自己发现“正在生成”就每秒级刷新自己，不依赖打开媒体库时的全局 nonce。
  const resultPollDelay = comicRowResultPollDelay(pages);
  // 刚点过重试时，服务端还没把页面标成 generating，这里先按 4 秒跟随刷新一段时间。
  const [, setFollowTick] = useState(0);
  useEffect(() => {
    if (!followUntil) return undefined;
    const timer = setInterval(() => setFollowTick(value => value + 1), 2000);
    return () => clearInterval(timer);
  }, [followUntil]);
  const followActive = Boolean(followUntil) && Date.now() < followUntil;
  const pollDelay = resultPollDelay || (followActive ? 4000 : 0);
  useEffect(() => {
    if (!pollDelay || !batchId || !book?.id) return undefined;
    const timer = setInterval(() => {
      if (typeof document !== 'undefined' && document.hidden) return;
      listComicPages(batchId, book.id, inlineView).then(pageResponse => {
        setPages(resultData(pageResponse, 'pages') || []);
        setLoadError('');
      }).catch(() => {});
    }, pollDelay);
    return () => clearInterval(timer);
  }, [pollDelay, batchId, book?.id]);
  useEffect(() => {
    if (!videos.some(video => video.id === selectedVideoId) && videos[0]?.id) selectVideo(videos[0].id);
  }, [videos.map(video => video.id).join('|'), selectedVideoId]);
  if (!selectedVideo) return <div className="batch-factory-cell-shell batch-factory-inline-media is-empty"><div className="batch-factory-cell-empty">提取漫画提示词后显示漫画页。</div></div>;
  const generatingPageCount = pages.filter(item => item?.generationState === 'generating').length;
  return <div className="batch-factory-cell-shell batch-factory-inline-comic">
    <header className="batch-factory-cell-head"><div><b>漫画页 {String(selectedIndex + 1).padStart(2, '0')}</b><small>{presentation.statusText}</small>{generatingPageCount > 0 ? <small className="batch-factory-comic-stale batch-factory-comic-running"><CmLoader size={13} title="正在生成" />正在生成 {generatingPageCount} 页…</small> : (followActive ? <small className="batch-factory-comic-stale batch-factory-comic-running"><CmLoader size={13} tone="pending" title="等待开始" />已提交重试，等待开始…</small> : null)}{page?.configStale ? <small className="batch-factory-comic-stale">原文已变化，建议重新分析分页与提示词</small> : null}</div><span className={`batch-factory-cell-state${primary ? ' is-ready' : ''}`}>{presentation.stateText}</span></header>
    <div className={`batch-factory-inline-comic-stage${generatingPageCount > 0 ? ' is-busy' : (followActive ? ' is-pending' : '')}`}>
      {page?.generationError && !primary ? <BatchFactoryErrorNotice lineMode title="漫画页生成失败" error={page.generationError} /> : null}
      <button type="button" className="batch-factory-inline-comic-primary" onClick={() => (presentation.retry ? setReloadKey(key => key + 1) : onManage?.(selectedVideo.id))}>
        {primary?.url ? <AuthenticatedComicPageImage imageUrl={primary.url} thumbnailUrl={primary.thumbnailUrl} alt={`漫画页 ${selectedIndex + 1}`} lazy /> : <>{presentation.tone === 'busy' ? <CmLoader size={30} title="漫画页生成中" /> : <PictureOutlined />}<span>{presentation.placeholder}</span></>}
      </button>
    </div>
    <div className="batch-factory-inline-media-versions"><span>候选</span>{candidates.map(image => <button type="button" className="batch-factory-inline-media-version" key={image.id} onClick={() => onManage?.(selectedVideo.id)}><AuthenticatedComicPageImage imageUrl={image.url} thumbnailUrl={image.thumbnailUrl} alt="漫画候选图" lazy /></button>)}<button type="button" className="batch-factory-inline-media-version is-more" onClick={() => onManage?.(selectedVideo.id)} title="打开漫画媒体库">•••</button></div>
    <div className="batch-factory-cell-pager"><button type="button" disabled={selectedIndex === 0} onClick={() => selectVideo(videos[selectedIndex - 1]?.id)}><LeftOutlined /></button><span>{selectedIndex + 1} / {videos.length}</span><button type="button" disabled={selectedIndex >= videos.length - 1} onClick={() => selectVideo(videos[selectedIndex + 1]?.id)}><RightOutlined /></button></div>
  </div>;
}

const COMIC_PAGE_GENERATION_LIMIT = 35;
// 单本整本出图的页级并发；两本书同时运行时合计不超过上游 8 路的总上限。
const COMIC_PAGE_CONCURRENCY = 4;

function ComicMediaPanel({ batchId, book, comicPublishSettings = {}, initialVideoId = '', onSaved, onGeneratePage, generatingPage = false, refreshNonce = 0, onClose }) {
  const videos = book?.videos || [];
  const [pages, setPages] = useState([]);
  // 403/404 时 pages 会是空数组；必须和「真的 0 页」区分开，
  // 否则界面会把「无权访问」显示成 0/35。
  const [pageLoadError, setPageLoadError] = useState('');
  const [segmentation, setSegmentation] = useState(null);
  const [collections, setCollections] = useState([]);
  const [activeView, setActiveView] = useState('pages');
	const [selectedVideoId, setSelectedVideoId] = useState(initialVideoId || videos[0]?.id || '');
	const [collectionStart, setCollectionStart] = useState(1);
	const [collectionEnd, setCollectionEnd] = useState(Math.min(videos.length || 1, 35));
	const [collectionRangeTouched, setCollectionRangeTouched] = useState(false);
  const [busy, setBusy] = useState('');
	const [referenceUploading, setReferenceUploading] = useState(false);
	const [publishTarget, setPublishTarget] = useState(null);
	// 发布入口不能因为状态目录短暂读取失败而被锁死；提交时仍由后端做最终校验。
	const [publishConfigured, setPublishConfigured] = useState(true);
	const [publishStatusLoaded, setPublishStatusLoaded] = useState(false);
	const cachedPublishDirectory = readComicPublishDirectoryCache();
	const [producerAccounts, setProducerAccounts] = useState(cachedPublishDirectory?.producerAccounts || []);
	const [organizations, setOrganizations] = useState(cachedPublishDirectory?.organizations || []);
	const [audioCategories, setAudioCategories] = useState(cachedPublishDirectory?.audioCategories?.length ? cachedPublishDirectory.audioCategories : DEFAULT_AUDIO_CATEGORIES);
	const [producerAccountOrganizations, setProducerAccountOrganizations] = useState(cachedPublishDirectory?.producerAccountOrganizations || {});
	const [producerAccountsWarning, setProducerAccountsWarning] = useState('');
	const [publishAudios, setPublishAudios] = useState([]);
	const [publishedByCollection, setPublishedByCollection] = useState({});
	const [publishForm, setPublishForm] = useState({ producerAdminName: '', organizationId: '1', audioCategoryId: null, specifiedAudioId: null, specifiedAudioName: '', workIdsText: '' });
	const [workOptions, setWorkOptions] = useState([]);
	const [workSearchLoading, setWorkSearchLoading] = useState(false);
	const [workSearchError, setWorkSearchError] = useState('');
  const selectedIndex = Math.max(0, videos.findIndex(video => video.id === selectedVideoId));
  const selectedVideo = videos[selectedIndex] || null;
  const selectedPage = pages.find(page => page.videoId === selectedVideo?.id) || null;
  const sourceSegment = segmentation?.segments?.find(segment => segment.videoId === selectedVideo?.id) || null;
  const primary = selectedPage?.images?.find(image => image.isPrimary) || null;
  const readyComicPageCount = videos.filter(video => pages.find(page => page.videoId === video.id)?.images?.some(image => image.isPrimary)).length;
  const comicCompletionPercent = videos.length ? Math.round((readyComicPageCount / videos.length) * 100) : 0;
  // 只有「读取失败且一页都没拿到」才按无权访问展示；已有数据时保留原数据。
  const pageAccessMessage = pageLoadError && !pages.length ? pageLoadError : '';
  const readyComicPageNumbers = videos.reduce((numbers, video, index) => {
    if (pages.find(page => page.videoId === video.id)?.images?.some(image => image.isPrimary)) numbers.push(index + 1);
    return numbers;
  }, []);
  const readyComicPageSummary = readyComicPageNumbers.length
    ? `已完成第 ${readyComicPageNumbers.slice(0, 6).join('、')}${readyComicPageNumbers.length > 6 ? ' 等页' : ' 页'}`
    : '暂未完成任何页';
  const selectedComicPageStatus = primary
    ? '当前页主图已就绪'
    : selectedPage?.images?.length
      ? '当前页已有候选图，等待选择主图'
      : readyComicPageCount
        ? `当前第 ${selectedIndex + 1} 页尚未完成，系统仍在继续生成；${readyComicPageSummary}`
        : '当前页尚未完成，系统会继续自动生成';
  async function load() {
    const [pageResponse, collectionResponse, sourceResponse] = await Promise.all([
      listComicPages(batchId, book.id),
      listComicCollections(batchId, book.id),
      getComicSourceSegments(batchId, book.id, { silent: true, suppressGlobalError: true }).catch(() => ({ segmentation: null }))
    ]);
    // 先渲染分页、合集和原文分段；发布记录稍后补，不能阻塞首屏数据。
    const nextCollections = resultData(collectionResponse, 'collections') || [];
    setPages(resultData(pageResponse, 'pages') || []);
    setPageLoadError('');
    setCollections(nextCollections);
    setSegmentation(resultData(sourceResponse, 'segmentation') || null);
    const publishRecords = await Promise.all(nextCollections.map(collection => getComicCollectionPublishRecord(batchId, book.id, collection.id, { silent: true, suppressGlobalError: true }).catch(() => null)));
    const nextPublished = {};
    nextCollections.forEach((collection, index) => {
      const record = publishRecords[index]?.publish;
      if (publishRecords[index]?.published === true && record) nextPublished[collection.id] = record;
    });
    setPublishedByCollection(nextPublished);
  }
	useEffect(() => {
		setSelectedVideoId(initialVideoId || videos[0]?.id || '');
		setCollectionStart(1);
		setCollectionEnd(Math.min(videos.length || 1, 35));
		setCollectionRangeTouched(false);
	}, [batchId, book?.id, initialVideoId]);
	useEffect(() => {
		load().catch(error => {
			setPageLoadError(batchAccessErrorMessage(error));
			message.error(error?.message || '读取漫画媒体库失败');
		});
	}, [batchId, book?.id, refreshNonce]);
	useEffect(() => {
		const organizationId = String(publishForm.organizationId || '').trim();
		const audioCategoryId = String(publishForm.audioCategoryId || '').trim();
		if (!publishTarget || !organizationId || !audioCategoryId) { setPublishAudios([]); return undefined; }
		let active = true;
		getN8ComicPublishAudios({ organizationId, audioCategoryId }, { silent: true, suppressGlobalError: true }).then(result => {
			if (!active) return;
			const candidates = Array.isArray(result?.audios) ? result.audios : [];
			setPublishAudios(candidates);
			const selected = chooseN8Audio(candidates, publishForm.specifiedAudioId);
			setPublishForm(current => ({ ...current, specifiedAudioId: selected?.id || null, specifiedAudioName: selected?.name || '' }));
		}).catch(() => { if (active) { setPublishAudios([]); setPublishForm(current => ({ ...current, specifiedAudioId: null, specifiedAudioName: '' })); } });
		return () => { active = false; };
	}, [publishTarget?.id, publishForm.organizationId, publishForm.audioCategoryId]);
	useEffect(() => {
		if (!collectionRangeTouched && collectionEnd === 1 && videos.length > 1) {
			setCollectionEnd(Math.min(videos.length, 35));
		}
	}, [videos.length, collectionEnd, collectionRangeTouched]);
	useEffect(() => {
		let active = true;
		getN8ComicPublishStatus({ silent: true, suppressGlobalError: true }).then(response => {
			if (!active) return;
			setPublishConfigured(response?.configured === true);
			const accounts = Array.isArray(response?.producerAccounts) ? response.producerAccounts : [];
			setProducerAccounts(accounts);
			setOrganizations(Array.isArray(response?.organizations) ? response.organizations : []);
			const categories = Array.isArray(response?.audioCategories) ? response.audioCategories : [];
			if (categories.length) setAudioCategories(categories);
			setProducerAccountOrganizations(response?.producerAccountOrganizations && typeof response.producerAccountOrganizations === 'object' ? response.producerAccountOrganizations : {});
			const latestDirectory = readComicPublishDirectoryCache();
			writeComicPublishDirectoryCache({ ...response, audioCategories: categories.length ? categories : (latestDirectory?.audioCategories || cachedPublishDirectory?.audioCategories || []) });
			setPublishForm(current => ({
				...current,
				producerAdminName: accounts.includes(String(current.producerAdminName || '').trim()) ? current.producerAdminName : (accounts[0] || ''),
				audioCategoryId: categories.some(item => Number(item.id) === Number(current.audioCategoryId)) ? Number(current.audioCategoryId) : (categories[0]?.id || null)
			}));
			setProducerAccountsWarning(String(response?.producerAccountsWarning || ''));
	}).catch(() => { if (active) setPublishConfigured(true); }).finally(() => { if (active) setPublishStatusLoaded(true); });
		getN8ComicPublishAudioCategories({ silent: true, suppressGlobalError: true }).then(response => {
			if (!active) return;
			const categories = Array.isArray(response?.audioCategories) ? response.audioCategories : [];
			if (!categories.length) return;
			setAudioCategories(categories);
			setPublishForm(current => ({
				...current,
				audioCategoryId: categories.some(item => Number(item.id) === Number(current.audioCategoryId))
					? Number(current.audioCategoryId)
					: Number(categories[0]?.id || 0) || null
			}));
			const latestDirectory = readComicPublishDirectoryCache();
			writeComicPublishDirectoryCache({ ...(latestDirectory || cachedPublishDirectory || {}), audioCategories: categories });
		}).catch(() => {});
		return () => { active = false; };
	}, [batchId]);
  async function choosePrimary(image) {
    if (!selectedPage || image?.isPrimary) return;
    setBusy(image.id);
    try {
      await setPrimaryComicPageImage(batchId, book.id, selectedPage.id, image.id);
      await load();
      await onSaved?.();
      message.success('当前漫画页主图已切换。');
    } catch (error) { message.error(error?.message || '切换漫画主图失败'); } finally { setBusy(''); }
  }
	async function splitSource(payload) {
    const response = await splitComicSourceSegment(batchId, book.id, selectedVideo.id, payload);
    setSegmentation(resultData(response, 'segmentation') || null);
    await onSaved?.();
    message.success('原文已按安全位置拆分，相关页面等待重新生成提示词。');
  }
	async function savePageReferences(referenceImageUrls, successText) {
		if (!selectedVideoId) return;
		setBusy('reference');
		try {
			await setComicPageReferenceImages(batchId, book.id, selectedVideoId, referenceImageUrls);
			await load();
			await onSaved?.();
			message.success(successText);
		} catch (error) { message.error(error?.message || '保存当前页参考图失败'); }
		finally { setBusy(''); }
	}
	async function resegmentSource() {
		if (!batchId || !book?.id || busy) return;
		setBusy('segmentation');
		try {
			const response = await segmentComicSource(batchId, book.id);
			setSegmentation(resultData(response, 'segmentation') || null);
			await load();
			await onSaved?.();
			message.success('已按剧情重新分析漫画分段；请重新生成全部漫画提示词。');
		} catch (error) {
			message.error(error?.message || '重新分析漫画分段失败');
		} finally {
			setBusy('');
		}
	}
	async function uploadPageReference(file) {
		setReferenceUploading(true);
		try {
			const dataUrl = await fileAsDataURL(file);
			const uploaded = await uploadComicGlobalReference(batchId, dataUrl);
			const referenceURL = String(uploaded?.reference?.url || '').trim();
			if (!referenceURL) throw new Error('参考图上传后没有可用地址');
			await savePageReferences([referenceURL], '当前漫画页参考图已保存。');
		} catch (error) { message.error(error?.message || '上传当前页参考图失败'); }
		finally { setReferenceUploading(false); }
		return false;
	}
	async function applyPageReferenceToAll() {
		if (!selectedVideoId) return;
		setBusy('reference-all');
		try {
			const response = await applyComicPageReferenceImages(batchId, book.id, selectedVideoId);
			await load();
			await onSaved?.();
			message.success(`当前页参考图已应用到 ${Number(response?.updated || 0)} 个漫画页。`);
		} catch (error) { message.error(error?.message || '应用全部漫画页失败'); }
		finally { setBusy(''); }
	}
	async function makeCollection() {
		setBusy('collection');
		try {
			await createComicCollection(batchId, book.id, { videoIds: selectedCollectionVideoIds });
			await load();
			message.success(`漫画合集已收集第 ${collectionStart}-${collectionEnd} 页主图。`);
		} catch (error) { message.error(error?.message || '生成漫画合集失败'); } finally { setBusy(''); }
	}
	async function downloadCollectionImages(collection) {
		if (!collection?.id || busy) return;
		const requestKey = `collection-download-${collection.id}`;
		setBusy(requestKey);
		try {
			const blob = await downloadComicCollectionImageZip(batchId, book.id, collection.id);
			const objectURL = URL.createObjectURL(blob);
			const anchor = document.createElement('a');
			anchor.href = objectURL;
			anchor.download = `${String(book.title || '漫画合集').replace(/[\\/:*?"<>|]/g, '_').trim() || '漫画合集'}.zip`;
			document.body.appendChild(anchor);
			anchor.click();
			anchor.remove();
			window.setTimeout(() => URL.revokeObjectURL(objectURL), 1000);
			message.success('漫画合集图片已开始打包下载。');
		} catch (error) {
			message.error(error?.message || '下载漫画合集图片失败');
		} finally {
			setBusy(current => current === requestKey ? '' : current);
		}
	}
	async function regenerateCurrentPage() {
		if (!selectedVideoId || !onGeneratePage) return;
		await onGeneratePage?.(selectedVideoId, selectedVideo?.comicImagePrompt || selectedVideo?.comicPrompt || '');
		await load();
	}
	async function searchComicWorks(query, autoSelect = false) {
		const keyword = String(query || '').trim();
		if (!keyword) {
			setWorkOptions([]);
			setWorkSearchError('');
			return;
		}
		setWorkSearchLoading(true);
		setWorkSearchError('');
		try {
			const response = await searchN8ComicWorks(keyword, { silent: true, suppressGlobalError: true });
			const works = Array.isArray(response?.works) ? response.works.filter(work => Number(work?.id) > 0 && String(work?.name || '').trim()) : [];
			setWorkOptions(works);
			if (autoSelect) {
				const exact = works.find(work => String(work.name).trim() === keyword);
				const selected = exact || (works.length === 1 ? works[0] : null);
				if (selected) setPublishForm(current => ({ ...current, workIdsText: String(selected.id) }));
			}
		} catch (error) {
			setWorkOptions([]);
			setWorkSearchError(error?.message || '作品查询失败，可手动填写作品 ID。');
		} finally {
			setWorkSearchLoading(false);
		}
	}
	function openComicPublish(collection) {
		setPublishTarget(collection);
		const savedProducer = String(comicPublishSettings.producerAdminName || '').trim();
		const savedAudioCategoryId = Number(comicPublishSettings.audioCategoryId || 0);
		const savedAudioCategoryName = String(comicPublishSettings.audioCategoryName || '').trim();
		const selectedAudioCategory = audioCategories.find(item => Number(item.id) === savedAudioCategoryId || item.name === savedAudioCategoryName) || audioCategories[0];
		setPublishForm({
			// 不能用账号目录第一个账号兜底，否则会把当前书已配置的创意者静默替换掉。
			producerAdminName: savedProducer,
			organizationId: String(comicPublishSettings.organizationId || ''),
			audioCategoryId: Number(selectedAudioCategory?.id || 0) || null,
			specifiedAudioId: Number(comicPublishSettings.specifiedAudioId || 0) || null,
			specifiedAudioName: String(comicPublishSettings.specifiedAudioName || ''),
			workIdsText: String(comicPublishSettings.workIdsText || '')
		});
		setWorkOptions([]);
		setWorkSearchError('');
		searchComicWorks(book?.title, !String(comicPublishSettings.workIdsText || '').trim());
	}
	function parsedWorkIDs() {
		const tokens = String(publishForm.workIdsText || '').split(/[,，\s]+/).map(value => value.trim()).filter(Boolean);
		if (!tokens.length || tokens.some(value => !/^\d+$/.test(value) || Number(value) <= 0)) return [];
		return [...new Set(tokens.map(Number))];
	}
	async function confirmComicPublish() {
		const workIds = parsedWorkIDs();
		const selectedAudioCategoryId = Number(publishForm.audioCategoryId || 0);
		const selectedAudioId = Number(publishForm.specifiedAudioId || 0);
		const audioCategoryIsEnabled = audioCategories.some(item => Number(item.id) === selectedAudioCategoryId);
		if (!publishTarget || !String(publishForm.producerAdminName || '').trim() || !String(publishForm.organizationId || '').trim() || !audioCategoryIsEnabled || !selectedAudioId || !workIds.length) {
			message.error('请填写制作者账号、组织、音频分类 ID 和作品 ID。');
			return;
		}
		setBusy(`publish-${publishTarget.id}`);
		try {
			const response = await publishComicCollection(batchId, book.id, publishTarget.id, {
				producerAdminName: String(publishForm.producerAdminName).trim(),
				organizationId: String(publishForm.organizationId || '1'),
				audioCategoryId: Number(publishForm.audioCategoryId),
				audioId: selectedAudioId,
				workIds
			});
			const published = resultData(response, 'publish');
			setPublishedByCollection(current => ({ ...current, [publishTarget.id]: published }));
			setPublishTarget(null);
			message.success(`漫画合集已发布：${published?.name || book.title}`);
		} catch (error) { message.error(error?.message || '发布漫画合集失败'); }
		finally { setBusy(''); }
	}
	const publishDate = new Intl.DateTimeFormat('zh-CN', { timeZone: 'Asia/Shanghai', month: '2-digit', day: '2-digit' }).format(new Date()).replace(/\D/g, '');
	const publishNamePreview = `${String(publishForm.producerAdminName || '').trim()}${publishDate}-${book?.title || ''}`;
	const publishPageCount = publishTarget?.items?.length || 0;
	const producerOptions = producerAccountOptions(producerAccounts, publishForm.producerAdminName, publishForm.organizationId, producerAccountOrganizations);
	const selectedProducer = String(publishForm.producerAdminName || '').trim();
	const selectedProducerOrganizations = producerAccountOrganizations?.[selectedProducer];
	const producerMatchesOrganization = !Array.isArray(selectedProducerOrganizations) || selectedProducerOrganizations.length === 0 || selectedProducerOrganizations.includes(String(publishForm.organizationId || ''));
	const canPublish = publishConfigured && publishPageCount >= 2 && publishPageCount <= 35 && selectedProducer && producerMatchesOrganization && String(publishForm.organizationId || '').trim() && Number(publishForm.audioCategoryId || 0) > 0 && Number(publishForm.specifiedAudioId || 0) > 0 && parsedWorkIDs().length > 0;
	const audioCategoryOptions = audioCategories.map(item => ({ value: String(item.id), label: `${item.name}（ID ${item.id}）` }));
	const InputNumber = props => props.placeholder === '启用中的音频分类 ID' && audioCategoryOptions.length ? <Select showSearch optionFilterProp="label" value={String(props.value || '') || undefined} options={audioCategoryOptions} placeholder="请选择 N8 启用中的音频分类" placement="bottomLeft" getPopupContainer={() => document.body} onChange={value => props.onChange?.(Number(value))} style={props.style} /> : <NativeInputNumber {...props} />;
	const audioSelect = <label className="batch-factory-engine-field"><span><b>指定音频</b></span><Select value={String(publishForm.specifiedAudioId || '') || undefined} showSearch optionFilterProp="label" options={publishAudios.map(item => ({ value: String(item.id), label: `${item.name}（ID ${item.id}）` }))} placeholder="当前分类和组织下没有可用指定音频" onChange={specifiedAudioId => { const selected = publishAudios.find(item => String(item.id) === String(specifiedAudioId)); setPublishForm(current => ({ ...current, specifiedAudioId: selected?.id || null, specifiedAudioName: selected?.name || '' })); }} /></label>;
		const workSelect = <label className="batch-factory-engine-field"><span><b>作品</b></span><Select showSearch optionFilterProp="label" filterOption={false} loading={workSearchLoading} value={parsedWorkIDs().length === 1 ? String(parsedWorkIDs()[0]) : undefined} options={workOptions.map(work => ({ value: String(work.id), label: `${work.name}（ID ${work.id}）` }))} onSearch={query => searchComicWorks(query)} onChange={value => setPublishForm(current => ({ ...current, workIdsText: String(value || '') }))} placeholder={workSearchLoading ? '正在查询作品' : '输入书名搜索作品'} notFoundContent={workSearchError || '没有找到匹配作品'} /></label>;
	const maxCollectionPage = Math.max(1, videos.length);
	const normalizedCollectionStart = Math.min(Math.max(1, Number(collectionStart) || 1), maxCollectionPage);
	const normalizedCollectionEnd = Math.min(Math.max(normalizedCollectionStart, Number(collectionEnd) || normalizedCollectionStart), maxCollectionPage);
	const selectedCollectionVideos = videos.slice(normalizedCollectionStart - 1, normalizedCollectionEnd);
	const selectedCollectionVideoIds = selectedCollectionVideos.map(video => video.id);
	const selectedReadyPageCount = selectedCollectionVideos.filter(video => pages.find(page => page.videoId === video.id)?.images?.some(image => image.isPrimary)).length;
	const selectedCollectionPageCount = selectedCollectionVideoIds.length;
	const canMakeCollection = selectedCollectionPageCount >= 2 && selectedCollectionPageCount <= 35 && selectedReadyPageCount === selectedCollectionPageCount;
	function updateCollectionRange(start, end) {
		const maxPage = Math.max(1, videos.length);
		const nextStart = Math.min(Math.max(1, Number(start) || 1), maxPage);
		const nextEnd = Math.min(Math.max(nextStart, Number(end) || nextStart), maxPage);
		setCollectionRangeTouched(true);
		setCollectionStart(nextStart);
		setCollectionEnd(nextEnd);
	}
	const pageTab = <div className="batch-factory-comic-page-controls">
		{selectedPage?.generationError && !primary ? <BatchFactoryErrorNotice compact title="漫画页生成失败" error={selectedPage.generationError} /> : null}
		<section className="batch-factory-comic-media-section batch-factory-comic-media-candidates">
			<header><div><b>候选图</b><small>选择一张作为本页主图</small></div><span>{selectedPage?.images?.length || 0} 张</span></header>
			<div>{(selectedPage?.images || []).map((image, index) => <button type="button" className={image.isPrimary ? 'is-primary' : ''} key={image.id} disabled={busy === image.id} onClick={() => choosePrimary(image)}><AuthenticatedComicPageImage imageUrl={image.url} thumbnailUrl={image.thumbnailUrl} alt={`候选图 ${index + 1}`} /><span>{image.isPrimary ? '当前主图' : `候选 ${index + 1}`}</span></button>)}{!selectedPage?.images?.length ? <p>生成当前漫画页后，候选图会保留在这里。</p> : null}</div>
		</section>
		<section className="batch-factory-comic-media-section batch-factory-comic-page-reference"><header><div><b>本页参考图</b><small>约束人物、服装和画风</small></div><span>{selectedPage?.referenceImageUrls?.length ? `已设置 ${selectedPage.referenceImageUrls.length} 张` : '未设置'}</span></header><Space wrap>{(selectedPage?.referenceImageUrls || []).map(referenceURL => <AuthenticatedComicPageReference key={referenceURL} referenceURL={referenceURL} />)}<AntUpload accept="image/png,image/jpeg,image/webp" showUploadList={false} beforeUpload={uploadPageReference} disabled={!selectedVideoId || referenceUploading || Boolean(busy)}><Button size="small" icon={<UploadOutlined />} loading={referenceUploading}>上传参考图</Button></AntUpload><Button size="small" disabled={!selectedPage?.referenceImageUrls?.length || Boolean(busy)} loading={busy === 'reference-all'} onClick={applyPageReferenceToAll}>应用到全部</Button>{selectedPage?.referenceImageUrls?.length ? <Button size="small" danger disabled={Boolean(busy)} loading={busy === 'reference'} onClick={() => savePageReferences([], '当前漫画页参考图已清空。')}>清空</Button> : null}</Space></section>
		<section className="batch-factory-comic-media-section batch-factory-comic-media-source"><ComicSourceContent segment={sourceSegment} ordinal={selectedIndex} onSplit={splitSource} embedded /></section>
	</div>;
	const collectionTab = <div className="batch-factory-comic-collections"><div className="batch-factory-comic-collection-actions"><div><b>漫画合集</b><small>按页码范围收集当前主图；同一本书可以分批生成多个合集。</small></div><Tag color={publishConfigured ? 'success' : 'warning'}>{publishStatusLoaded ? (publishConfigured ? '漫画后台已连接' : '漫画后台未配置') : '检查连接中'}</Tag></div><section className="batch-factory-comic-collection-range"><header><div><b>生成漫画合集</b><small>发布接口支持 2–35 页，未选中的页不会进入本合集。</small></div><Tag color={canMakeCollection ? 'success' : 'warning'}>{selectedCollectionPageCount} 页 · 已就绪 {selectedReadyPageCount} 页</Tag></header><div className="batch-factory-comic-collection-range-fields"><span>第</span><InputNumber min={1} max={maxCollectionPage} value={normalizedCollectionStart} onChange={value => updateCollectionRange(value, normalizedCollectionEnd)} /><span>页至</span><InputNumber min={normalizedCollectionStart} max={maxCollectionPage} value={normalizedCollectionEnd} onChange={value => updateCollectionRange(normalizedCollectionStart, value)} /><span>页</span></div><div className="batch-factory-comic-collection-range-shortcuts"><Button type="link" size="small" onClick={() => updateCollectionRange(1, Math.min(maxCollectionPage, 35))}>前 35 页</Button><Button type="link" size="small" onClick={() => updateCollectionRange(Math.max(1, maxCollectionPage - 34), maxCollectionPage)}>后 35 页</Button></div><Button className="batch-factory-comic-collection-range-submit" type="primary" loading={busy === 'collection'} disabled={!canMakeCollection || Boolean(busy)} onClick={makeCollection}>整理所选页为合集</Button></section>{collections.slice().reverse().map(collection => { const published = publishedByCollection[collection.id]; return <section key={collection.id}><header><b>{new Date(collection.createdAt).toLocaleString()}</b><Space><Tag color="success">{collection.items?.length || 0} 页</Tag><Button size="small" icon={<DownloadOutlined />} disabled={!collection.items?.length || Boolean(busy)} loading={busy === `collection-download-${collection.id}`} onClick={() => downloadCollectionImages(collection)}>下载图片</Button>{published ? <Tag color="blue">已发布 #{published.carouselId}</Tag> : <Button size="small" disabled={!publishConfigured || (collection.items?.length || 0) < 2 || (collection.items?.length || 0) > 35} loading={busy === `publish-${collection.id}`} onClick={() => openComicPublish(collection)}>发布到漫画后台</Button>}</Space></header><div>{(collection.items || []).map(item => <figure key={`${collection.id}-${item.ordinal}`}><AuthenticatedComicPageImage imageUrl={item.url} thumbnailUrl={item.thumbnailUrl} alt={`第 ${item.ordinal} 页`} lazy /><figcaption>第 ${item.ordinal} 页</figcaption></figure>)}</div></section>; })}{!collections.length ? <div className="batch-factory-cell-empty">先选择页码范围，再整理当前选中页的主图。</div> : null}</div>;
	return <><div className="batch-factory-comic-media-workbench">
		<section className="batch-factory-comic-media-preview">{primary?.url ? <AuthenticatedComicPageImage imageUrl={primary.url} thumbnailUrl={primary.thumbnailUrl} alt="当前漫画主图" preview /> : <div className="batch-factory-comic-media-preview-empty"><PictureOutlined /><strong>{selectedPage ? '当前漫画页还没有主图' : '当前分镜还没有漫画候选图'}</strong><span>{selectedComicPageStatus}</span><small>{pageAccessMessage || `整体完成度 ${readyComicPageCount} / ${videos.length} 页（${comicCompletionPercent}%）`}</small></div>}</section>
		<aside className="batch-factory-comic-media-tools">
		<header className="batch-factory-comic-media-header"><div className="batch-factory-comic-media-title"><small>媒体库</small><b>{book?.title || '未命名作品'}</b><span>{selectedVideo ? `漫画页 ${String(selectedIndex + 1).padStart(2, '0')}` : '漫画页'} · {primary ? '主图已就绪' : '尚未选择主图'}</span>{selectedPage?.configStale ? <em>原文已变化，建议重新分析分页与提示词</em> : null}</div>{pageAccessMessage ? <div className="batch-factory-comic-media-completion" role="status" aria-label={pageAccessMessage}><span>完成度</span><b>无数据</b><i>—</i></div> : <div className="batch-factory-comic-media-completion" aria-label={`完成度 ${readyComicPageCount}/${videos.length} 页`}><span>完成度</span><b>{readyComicPageCount} / {videos.length} 页</b><i>{comicCompletionPercent}%</i><div className="batch-factory-comic-media-completion-track"><span style={{ width: `${comicCompletionPercent}%` }} /></div></div>}<Space><Tooltip title="让 AI 按剧情重新分析小说分段"><Button loading={busy === 'segmentation'} disabled={Boolean(busy)} onClick={resegmentSource}>重新分析分段</Button></Tooltip><Tooltip title={canMakeCollection ? `按当前选择的第 ${normalizedCollectionStart}-${normalizedCollectionEnd} 页生成漫画合集` : `所选页需要全部有主图，且合集必须为 2–35 页`}><Button type="primary" loading={busy === 'collection'} disabled={!canMakeCollection || Boolean(busy)} onClick={makeCollection}>生成漫画合集（所选页）</Button></Tooltip>{onClose ? <Button type="text" aria-label="关闭漫画媒体库" icon={<CloseOutlined />} onClick={onClose} /> : null}</Space></header>
			<Tabs className="batch-factory-comic-media-tabs" activeKey={activeView} onChange={setActiveView} items={[{ key: 'pages', label: `漫画页 ${videos.length}` }, { key: 'collections', label: `漫画合集 ${collections.length}` }]} />
			<div className="batch-factory-comic-media-scroll">{activeView === 'pages' ? pageTab : collectionTab}</div>
			<footer className="batch-factory-comic-media-footer"><div><Button disabled={selectedIndex === 0} aria-label="上一页" icon={<LeftOutlined />} onClick={() => setSelectedVideoId(videos[selectedIndex - 1]?.id)} /><span><b>{String(selectedIndex + 1).padStart(2, '0')}</b> / {String(videos.length).padStart(2, '0')}</span><Button disabled={selectedIndex >= videos.length - 1} aria-label="下一页" icon={<RightOutlined />} onClick={() => setSelectedVideoId(videos[selectedIndex + 1]?.id)} /></div><Button icon={<ReloadOutlined />} loading={generatingPage} disabled={!selectedVideoId || Boolean(busy) || generatingPage || !onGeneratePage} onClick={regenerateCurrentPage}>重新生成本页</Button></footer>
		</aside>
	</div><Modal title="确认发布漫画合集" open={Boolean(publishTarget)} onCancel={() => setPublishTarget(null)} onOk={confirmComicPublish} okText="确认发布" confirmLoading={String(busy).startsWith('publish-')} okButtonProps={{ disabled: !canPublish }} destroyOnClose><Space direction="vertical" size={14} style={{ width: '100%' }}><Alert type={publishConfigured ? 'info' : 'warning'} showIcon message={publishConfigured ? '漫画后台接口已连接' : '服务器尚未配置漫画后台接口密钥'} description="发布后会按当前主图顺序上传，不会上传候选图。每个合集只能创建一次，重复点击会返回原发布记录。" />{producerAccountsWarning ? <Alert type="warning" showIcon message={producerAccountsWarning} /> : null}<Descriptions size="small" column={1} bordered items={[{ key: 'name', label: '合集名称', children: publishNamePreview || '选择制作者账号后生成' }, { key: 'pages', label: '漫画页', children: `${publishPageCount} 页（接口要求 2 到 35 页）` }, { key: 'organization', label: '所属组织', children: organizations.find(item => String(item.id) === String(publishForm.organizationId))?.name ? `${organizations.find(item => String(item.id) === String(publishForm.organizationId)).name}（组织 ${publishForm.organizationId}）` : `组织 ${publishForm.organizationId || '未填'}` }, { key: 'category', label: '音频分类', children: comicPublishSettings.audioCategoryName ? `${comicPublishSettings.audioCategoryName}（ID ${publishForm.audioCategoryId || '未填'}）` : `ID ${publishForm.audioCategoryId || '未填'}` }]} /><label className="batch-factory-engine-field"><span><b>制作者账号（创意者）</b></span><Select showSearch optionFilterProp="label" value={publishForm.producerAdminName || undefined} options={producerOptions} disabled={publishStatusLoaded && producerOptions.length === 0} onChange={producerAdminName => setPublishForm(current => ({ ...current, producerAdminName }))} placeholder={producerOptions.length ? '请选择制作者账号' : '账号目录尚未配置'} /></label><label className="batch-factory-engine-field"><span><b>所属组织</b></span><Select value={String(publishForm.organizationId || '') || undefined} options={organizations.map(item => ({ value: String(item.id), label: `${item.name}（组织 ${item.id}）` }))} placeholder="请选择 N8 组织" onChange={organizationId => setPublishForm(current => ({ ...current, organizationId: String(organizationId || '') }))} /></label><label className="batch-factory-engine-field"><span><b>音频分类 ID</b></span><InputNumber min={1} precision={0} value={publishForm.audioCategoryId} onChange={audioCategoryId => setPublishForm(current => ({ ...current, audioCategoryId }))} placeholder="启用中的音频分类 ID" style={{ width: '100%' }} /></label>{workSelect}<label className="batch-factory-engine-field"><span><b>手动作品 ID（备用）</b></span><Input value={publishForm.workIdsText} onChange={event => setPublishForm(current => ({ ...current, workIdsText: event.target.value }))} placeholder="查询不到时填写一个或多个作品 ID" /></label></Space></Modal></>;
}

function MediaVersionPanel({ book, batchId, productionStatus, mergeJob, mergeJobs = [], mergeSettings = {}, mergeSpeed, onMergeSpeedChange, onMerge, onUpload, mergeAvailable = false, mergeReason = '', merging = false, onSaved, onDeleted, onRegenerate, onRetry, regenerating, productionAvailable = false, productionReason = '', initialVideoId = '', initialTab = 'clips', onOpenDirector }) {
  const videos = book?.videos || [];
  const videoProgress = batchFactoryVideoProgress(book, productionStatus);
  const mediaVersions = bookVideoVersions(book, productionStatus);
  const [viewerKind, setViewerKind] = useState(initialTab === 'merges' ? 'merge' : 'video');
  const [selectedVideoId, setSelectedVideoId] = useState(initialVideoId || videos[0]?.id || '');
  const [selectedVersionId, setSelectedVersionId] = useState('');
  const [selectedMergePreviewId, setSelectedMergePreviewId] = useState('');
  const [savingTaskId, setSavingTaskId] = useState('');
  const [playRequested, setPlayRequested] = useState(false);
  const [mediaMetrics, setMediaMetrics] = useState({});
  const [primaryOverrides, setPrimaryOverrides] = useState({});
  const [followAudio, setFollowAudio] = useState(normalizeAudioSource(mergeSettings.audioSource) !== 'original' && mergeSettings.audioMergeEnabled === true);
  const [audioDurationSeconds, setAudioDurationSeconds] = useState(Number(mergeSettings.audioDurationSeconds || 0));
  const [audioBusy, setAudioBusy] = useState(false);
  const [pendingUploadKey, setPendingUploadKey] = useState('');
  const [sheetAnchor, setSheetAnchor] = useState('collapsed');
  const [sheetOffset, setSheetOffset] = useState(null);
  const [sheetDragging, setSheetDragging] = useState(false);
  const sheetRef = useRef(null);
  const sheetDragRef = useRef(null);
  const uploadInputRefs = useRef({});
  const mergeUploadInputRef = useRef(null);

  const completedMerges = mergeJobs.filter(job => job?.status === 'succeeded' && String(job?.outputUrl || '').trim());
  const mergeGroups = currentMergeJobsByOpening(mergeJobs);
  const mergeJobsNewest = mergeGroups.current;
  const currentMergeJob = mergeJobsNewest[0] || mergeJob || null;
  const historicalMergeCount = mergeGroups.historyCount;
  const selectedUpload = book?.settingsState?.patch?.primaryUploadSource || {};
  const selectedMerge = selectedUpload?.kind === 'merged'
    ? completedMerges.find(job => job.id === selectedUpload.mergeJobId) || null
    : defaultBookMerge(mergeJobs);
  const mergedReady = selectedMerge?.status === 'succeeded' && String(selectedMerge?.outputUrl || '').trim();
  const uploadVideo = selectedUpload?.kind === 'video' ? videos.find(video => video.id === selectedUpload.videoId) : null;
  const uploadVersion = uploadVideo ? (mediaVersions.get(uploadVideo.id) || []).find(version => version.id === selectedUpload.taskId && !version.isHistory) || null : null;
  const defaultRatio = batchFactoryParseMediaRatio(mergeSettings.aspectRatio, 9 / 16);

  function label(video) {
    return video?.label || video?.title || `分镜 ${videos.findIndex(item => item.id === video?.id) + 1}`;
  }

  function productionProgressFor(video) {
    return videoProgress.byVideo.get(String(video?.id || '')) || null;
  }

  function fallbackRatioFor(video) {
    return batchFactoryParseMediaRatio(
      video?.settingsState?.patch?.aspectRatio || video?.aspectRatio || mergeSettings.aspectRatio,
      defaultRatio
        )}

  function mediaInfo(version, video) {
    const cached = mediaMetrics[String(version?.id || '')] || {};
    const width = Number(cached.width || version?.width || version?.videoWidth || 0);
    const height = Number(cached.height || version?.height || version?.videoHeight || 0);
    const ratio = width > 0 && height > 0 ? width / height : fallbackRatioFor(video);
    return {
      width,
      height,
      ratio,
      orientation: batchFactoryMediaOrientation(ratio),
      ratioText: batchFactoryRatioText(width, height, ratio)
    };
  }

  function mergeInfo(job) {
    const key = `merge:${String(job?.id || '')}`;
    const cached = mediaMetrics[key] || {};
    const width = Number(cached.width || job?.width || job?.videoWidth || 0);
    const height = Number(cached.height || job?.height || job?.videoHeight || 0);
    const ratio = width > 0 && height > 0 ? width / height : defaultRatio;
    return {
      key,
      width,
      height,
      duration: Number(cached.duration || job?.durationSeconds || 0),
      ratio,
      orientation: batchFactoryMediaOrientation(ratio),
      ratioText: batchFactoryRatioText(width, height, ratio)
    };
  }

  function rememberMediaMetrics(key, event) {
    const element = event.currentTarget;
    if (!(element?.videoWidth > 0 && element?.videoHeight > 0)) return;
    const next = {
      width: element.videoWidth,
      height: element.videoHeight,
      duration: Number.isFinite(element.duration) ? element.duration : 0
    };
    setMediaMetrics(current => {
      const previous = current[key];
      if (previous?.width === next.width && previous?.height === next.height && previous?.duration === next.duration) return current;
      return { ...current, [key]: next };
    });
  }

  function primeViewerFrame(key, event) {
    rememberMediaMetrics(key, event);
    const element = event.currentTarget;
    if (!element || !element.paused || Number(element.currentTime || 0) >= 0.001) return;
    const duration = Number(element.duration || 0);
    const target = Number.isFinite(duration) && duration > 0 ? Math.min(0.06, duration / 2) : 0.04;
    try {
      if (target > 0) element.currentTime = target;
    } catch (_) {}
  }

  function primaryFor(video, list) {
    const overrideId = primaryOverrides[String(video?.id || '')];
    return (overrideId ? list.find(version => version.id === overrideId) : null) || primaryMediaVersion(video, list);
  }

  function videoState(video) {
    const list = mediaVersions.get(video.id) || [];
    const main = primaryFor(video, list);
    const progress = productionProgressFor(video);
    const active = ['running', 'queued'].includes(String(progress?.status || '').toLowerCase());
    const failed = String(progress?.status || '').toLowerCase() === 'failed';
    return { list, main, progress, active, failed, ready: Boolean(main?.mediaUrl) };
  }

  const selectedVideo = videos.find(video => video.id === selectedVideoId) || videos[0] || null;
  const selectedVideoState = selectedVideo ? videoState(selectedVideo) : { list: [], main: null };
  const selectedVersion = selectedVideoState.list.find(version => version.id === selectedVersionId)
    || selectedVideoState.main
    || selectedVideoState.list[0]
    || null;
  // Do not let a stale selected history item become the visible current
  // output after refresh.  History is only available from the task/log view;
  // this rail always follows the newest persisted attempt.
  const mergePreview = mergeJobsNewest.find(job => job.id === selectedMergePreviewId)
    || mergeJobsNewest.find(job => job.id === selectedMerge?.id)
    || currentMergeJob
    || selectedMerge
    || null;

  const readyVideos = videos.filter(video => videoState(video).ready);
  const missingVideos = videos.filter(video => !videoState(video).ready);
  const mergeReady = videos.length > 0 && readyVideos.length === videos.length;

  useEffect(() => {
    const next = videos.find(video => video.id === initialVideoId) || videos[0] || null;
    const precompiledFrame = resolvePrecompiledStoryboardFrame(book, initialVideoId);
    setViewerKind(initialTab === 'merges' ? 'merge' : 'video');
    setSelectedVideoId(next?.id || precompiledFrame?.key || '');
    setSelectedVersionId('');
    setSelectedMergePreviewId('');
    setPlayRequested(false);
    setMediaMetrics({});
    setPrimaryOverrides({});
    setSheetAnchor('collapsed');
    setSheetOffset(null);
  }, [book?.id, initialVideoId, initialTab]);

  useEffect(() => {
    setFollowAudio(normalizeAudioSource(mergeSettings.audioSource) === 'original' ? false : mergeSettings.audioMergeEnabled === true);
    setAudioDurationSeconds(Number(mergeSettings.audioDurationSeconds || 0));
  }, [book?.id, mergeSettings.audioSource, mergeSettings.audioMergeEnabled, mergeSettings.audioDurationSeconds]);

  useEffect(() => {
    if (!selectedVideo) {
      if (selectedVersionId) setSelectedVersionId('');
      return;
    }
    if (selectedVersionId && selectedVideoState.list.some(version => version.id === selectedVersionId)) return;
    setSelectedVersionId(selectedVideoState.main?.id || selectedVideoState.list.at(-1)?.id || '');
  }, [selectedVideoId, selectedVersionId, selectedVideoState.main?.id, selectedVideoState.list.map(version => version.id).join('|')]);

  useEffect(() => {
    if (selectedMergePreviewId && mergeJobsNewest.some(job => job.id === selectedMergePreviewId)) return;
    const selectedCurrentMerge = mergeJobsNewest.find(job => job.id === selectedMerge?.id);
    setSelectedMergePreviewId(selectedCurrentMerge?.id || mergeJobsNewest[0]?.id || '');
  }, [selectedMergePreviewId, selectedMerge?.id, mergeJobsNewest.map(job => job.id).join('|')]);

  function uploadSourceKey(source = selectedUpload) {
    if (source?.kind === 'video' && source.videoId && source.taskId) return `video:${source.videoId}:${source.taskId}`;
    if (source?.kind === 'merged' && source.mergeJobId) return `merged:${source.mergeJobId}`;
    if (selectedMerge?.id) return `merged:${selectedMerge.id}`;
    return '';
  }

  useEffect(() => {
    setPendingUploadKey(uploadSourceKey());
  }, [book?.revision, selectedUpload?.kind, selectedUpload?.videoId, selectedUpload?.taskId, selectedUpload?.mergeJobId, selectedMerge?.id]);

  function selectVideoPreview(video, version = null, play = true) {
    if (!video) return;
    const state = videoState(video);
    const nextVersion = version || state.main || state.list.at(-1) || null;
    setViewerKind('video');
    setSelectedVideoId(video.id);
    setSelectedVersionId(nextVersion?.id || '');
    setPlayRequested(Boolean(play && nextVersion?.mediaUrl));
  }

  function selectMergePreview(job, play = true) {
    if (!job) return;
    setViewerKind('merge');
    setSelectedMergePreviewId(job.id);
    setPlayRequested(Boolean(play && job.status === 'succeeded' && String(job.outputUrl || '').trim()));
  }

  async function choosePrimary(video, version) {
    if (!video || !version || savingTaskId) return;
    const key = String(video.id);
    const previous = primaryOverrides[key];
    setPrimaryOverrides(current => ({ ...current, [key]: version.id }));
    setSavingTaskId(version.id);
    setSelectedVideoId(video.id);
    setSelectedVersionId(version.id);
    setViewerKind('video');
    setPlayRequested(Boolean(version.mediaUrl));
    try {
      await saveVideoOverride(batchId, book.id, video.id, { patch: { primaryMediaTaskId: version.id }, expectedRevision: Number(video.revision || 0) });
      await onSaved?.();
      message.success(`${label(video)} 已切换到 ${batchFactoryStableMediaLabel('V', version.id)}；下一次合成将读取它。`);
    } catch (error) {
      setPrimaryOverrides(current => {
        const next = { ...current };
        if (previous) next[key] = previous;
        else delete next[key];
        return next;
      });
      message.error(error?.message || '切换分镜主版本失败');
    } finally {
      setSavingTaskId('');
    }
  }

  async function uploadVideoMaterial(video, event) {
    const input = event.currentTarget;
    const file = input.files?.[0];
    input.value = '';
    if (!file) return;
    if (file.type !== 'video/mp4' && !/\.mp4$/i.test(file.name || '')) {
      message.error('请上传 MP4 视频');
      return;
    }
    const uploadKey = `upload-material-${video.id}`;
    setSavingTaskId(uploadKey);
    try {
      await uploadBookVideoMaterial(batchId, book.id, video.id, file);
      await onSaved?.();
      message.success('素材已上传，并已设为当前分镜主版本。');
    } catch (error) {
      message.error(error?.message || '上传视频素材失败');
    } finally {
      setSavingTaskId('');
    }
  }

  async function uploadMergeMaterial(event) {
    const input = event.currentTarget;
    const file = input.files?.[0];
    input.value = '';
    if (!file) return;
    if (file.type !== 'video/mp4' && !/\.mp4$/i.test(file.name || '')) {
      message.error('请上传 MP4 视频');
      return;
    }
    setSavingTaskId('upload-merge');
    try {
      await uploadBookMergeMaterial(batchId, book.id, file);
      await onSaved?.();
      message.success('成片已上传，可在合成成片中查看。');
    } catch (error) {
      message.error(error?.message || '上传成片失败');
    } finally {
      setSavingTaskId('');
    }
  }

  async function removeCandidate(video, version) {
    if (!video || !version || savingTaskId) return;
    setSavingTaskId(`delete-${version.id}`);
    try {
      await deleteProductionTask(batchId, book.id, video.id, version.id);
      await onDeleted?.();
      message.success('候选视频已从当前书片段库移除。');
    } catch (error) {
      message.error(error?.message || '删除候选视频失败');
    } finally {
      setSavingTaskId('');
    }
  }

  async function chooseUploadSource(source) {
    if (!book?.id || savingTaskId) return;
    setSavingTaskId(source.kind === 'merged' ? `upload-merged-${source.mergeJobId}` : `upload-video-${source.videoId}`);
    try {
      await saveBookOverride(batchId, book.id, { patch: { primaryUploadSource: source }, expectedRevision: Number(book.revision || 0) });
      await onSaved?.();
      setPendingUploadKey(uploadSourceKey(source));
      message.success(source.kind === 'merged'
        ? '合成成片已设为 121 上传主视频。'
        : `${source.videoId === videos[0]?.id ? '首分镜已设为上传主视频和作品封面。' : '分镜主版本已设为 121 上传主视频。'}`);
    } catch (error) {
      message.error(error?.message || '设置上传主视频失败');
    } finally {
      setSavingTaskId('');
    }
  }

  async function applyPendingUploadSource() {
    if (!pendingUploadKey) return;
    const [kind, first, second] = pendingUploadKey.split(':');
    if (kind === 'merged' && first) {
      await chooseUploadSource({ kind: 'merged', mergeJobId: first });
      return;
    }
    if (kind === 'video' && first && second) {
      await chooseUploadSource({ kind: 'video', videoId: first, taskId: second });
    }
  }

  const uploadOptions = [
    ...completedMerges.map(job => ({
      value: `merged:${job.id}`,
      label: `成片 ${batchFactoryStableMediaLabel('M', job.id)} · ${mergeJobSpeed(job).toFixed(1)}x`
    })),
    ...videos.flatMap(video => {
      const state = videoState(video);
      return state.main?.mediaUrl && !state.main.isHistory ? [{
        value: `video:${video.id}:${state.main.id}`,
        label: `${label(video)} · 主 ${batchFactoryStableMediaLabel('V', state.main.id)}`
      }] : [];
    })
  ];

  const uploadSourceText = selectedUpload?.kind === 'video'
    ? (uploadVideo && uploadVersion
      ? `${label(uploadVideo)} · ${batchFactoryStableMediaLabel('V', uploadVersion.id)}`
      : '所选分镜来源暂不可用')
    : mergedReady
      ? `合成成片 · ${batchFactoryStableMediaLabel('M', selectedMerge?.id)} · ${mergeJobSpeed(selectedMerge).toFixed(1)}x`
      : '尚未选择可上传视频';

  function visibleHeightForSheetAnchor(anchor, height) {
    if (anchor === 'full') return height;
    if (anchor === 'half') return Math.min(320, height - 74);
    return 74;
  }

  // The media library is rendered in an Ant Modal portal.  On some viewport
  // sizes its visual origin differs from the CSS offset parent, so a simple
  // `height - visible` transform leaves the sheet floating midway down the
  // modal.  Derive each snap point from the visible workspace instead.
  function sheetGeometry() {
    const sheet = sheetRef.current;
    const parent = sheet?.parentElement;
    if (!sheet || !parent) return null;
    const styleOffset = Number(String(sheet.style.transform || '').match(/translateY\(([-\d.]+)px\)/)?.[1] || 0);
    const rect = sheet.getBoundingClientRect();
    return {
      height: rect.height || 520,
      rawTop: rect.top - styleOffset,
      parentBottom: parent.getBoundingClientRect().bottom,
      appliedOffset: styleOffset
    };
  }

  function sheetAnchors(height) {
    const geometry = sheetGeometry();
    if (geometry) {
      const anchorOffset = anchor => geometry.parentBottom - visibleHeightForSheetAnchor(anchor, geometry.height) - geometry.rawTop;
      return { full: anchorOffset('full'), half: anchorOffset('half'), collapsed: anchorOffset('collapsed') };
    }
    const full = 0;
    const collapsed = Math.max(0, height - 74);
    const half = Math.max(0, height - Math.min(320, height - 74));
    return { full, half, collapsed };
  }

  function sheetAnchorOffset(anchor = sheetAnchor) {
    const height = sheetRef.current?.getBoundingClientRect().height || 520;
    return sheetAnchors(height)[anchor] ?? 0;
  }

  function rubberBand(value, min, max) {
    if (value < min) {
      const distance = min - value;
      return min - (distance * 0.28) / (1 + distance / 260);
    }
    if (value > max) {
      const distance = value - max;
      return max + (distance * 0.28) / (1 + distance / 260);
    }
    return value;
  }

  function startSheetDrag(event) {
    if (event.button !== undefined && event.button !== 0) return;
    const startOffset = sheetOffset == null
      ? (sheetGeometry()?.appliedOffset ?? sheetAnchorOffset())
      : sheetOffset;
    sheetDragRef.current = {
      pointerId: event.pointerId,
      startY: event.clientY,
      startOffset,
      lastOffset: startOffset,
      lastY: event.clientY,
      lastTime: performance.now(),
      velocity: 0
    };
    setSheetDragging(true);
    event.currentTarget.setPointerCapture?.(event.pointerId);
  }

  function moveSheetDrag(event) {
    const drag = sheetDragRef.current;
    if (!drag || drag.pointerId !== event.pointerId) return;
    const now = performance.now();
    const deltaTime = Math.max(1, now - drag.lastTime);
    const instantVelocity = (event.clientY - drag.lastY) / deltaTime;
    drag.velocity = drag.velocity * 0.72 + instantVelocity * 0.28;
    drag.lastY = event.clientY;
    drag.lastTime = now;
    const height = sheetRef.current?.getBoundingClientRect().height || 520;
    const anchors = sheetAnchors(height);
    const nextOffset = rubberBand(drag.startOffset + event.clientY - drag.startY, anchors.full, anchors.collapsed);
    drag.lastOffset = nextOffset;
    setSheetOffset(nextOffset);
  }

  function endSheetDrag(event) {
    const drag = sheetDragRef.current;
    if (!drag || (event?.pointerId != null && drag.pointerId !== event.pointerId)) return;
    const height = sheetRef.current?.getBoundingClientRect().height || 520;
    const anchors = sheetAnchors(height);
    const current = Number.isFinite(drag.lastOffset)
      ? drag.lastOffset
      : (sheetOffset == null ? sheetAnchorOffset() : sheetOffset);
    const velocity = drag.velocity;
    const order = ['full', 'half', 'collapsed'];
    let target;
    if (Math.abs(velocity) > 0.42) {
      const currentIndex = order.indexOf(sheetAnchor);
      const nextIndex = velocity < 0 ? Math.max(0, currentIndex - 1) : Math.min(order.length - 1, currentIndex + 1);
      target = order[nextIndex];
    } else {
      target = order.reduce((best, name) => Math.abs(current - anchors[name]) < Math.abs(current - anchors[best]) ? name : best, 'collapsed');
    }
    sheetDragRef.current = null;
    setSheetDragging(false);
    setSheetAnchor(target);
    setSheetOffset(null);
  }

  function toggleSheet() {
    setSheetOffset(null);
    setSheetAnchor(current => current === 'collapsed' ? 'half' : current === 'half' ? 'full' : 'collapsed');
  }

  async function persistMediaAudioPatch(patch) {
    const result = await saveBookOverrideWithRetry(batchId, book.id, Number(book?.revision || 0), { patch });
    await onSaved?.();
    return result;
  }

  async function measureMergeAudioDuration({ force = false } = {}) {
    if (audioBusy) return Number(audioDurationSeconds || 0);
    const input = batchFactoryBookAudioInput(book);
    const tts = await batchFactoryBookTts(mergeSettings);
    const fingerprint = batchFactoryAudioFingerprint(input, tts);
    const savedFingerprint = String(mergeSettings.audioDurationFingerprint || '');
    const savedSeconds = Number(mergeSettings.audioDurationSeconds || 0);
    if (!force && savedSeconds > 0 && (mergeSettings.audioDurationManual === true || savedFingerprint === fingerprint)) {
      setAudioDurationSeconds(savedSeconds);
      return savedSeconds;
    }
    setAudioBusy(true);
    try {
      const measured = await generateBatchFactoryBookAudioMeasurement(book, mergeSettings);
      setAudioDurationSeconds(measured.durationSeconds);
      await persistMediaAudioPatch({ audioDurationSeconds: measured.durationSeconds, audioDurationFingerprint: measured.fingerprint, audioDurationManual: false });
      message.success(`已读取当前书配音时长：${measured.durationSeconds.toFixed(2)} 秒`);
      return measured.durationSeconds;
    } catch (error) {
      message.error(error?.message || '读取当前书配音时长失败；仍可手动填写秒数');
      return 0;
    } finally { setAudioBusy(false); }
  }

  async function toggleFollowAudio(checked) {
    if (normalizeAudioSource(mergeSettings.audioSource) === 'original') { message.info('当前为原声模式，不会合入 TTS。'); return; }
    setFollowAudio(checked);
    try {
      await persistMediaAudioPatch({ audioMergeEnabled: checked });
      if (checked && Number(audioDurationSeconds || 0) <= 0) await measureMergeAudioDuration();
    } catch (error) { message.error(error?.message || '保存合并跟随配音设置失败'); }
  }

  async function persistManualAudioDuration() {
    const seconds = Number(audioDurationSeconds || 0);
    if (!(seconds > 0)) return;
    try {
      await persistMediaAudioPatch({ audioDurationSeconds: seconds, audioDurationManual: true });
    } catch (error) { message.error(error?.message || '保存配音时长失败'); }
  }

  async function runCurrentMerge() {
    if (!mergeReady || !mergeAvailable || merging) return;
    let seconds = Number(audioDurationSeconds || 0);
    if (followAudio && seconds <= 0) seconds = await measureMergeAudioDuration();
    if (followAudio && seconds <= 0) return;
    onMerge?.({ speed: Number(mergeSpeed || 1), followAudio, audioDurationSeconds: seconds });
  }

  function renderThumb(media, info, labelText, className = 'batch-factory-media-pickstation-thumb') {
    return <PassiveMediaPreview
      media={media}
      className={`${className} is-${info.orientation}`}
      label={labelText}
      onLoadedMetadata={event => rememberMediaMetrics(String(media?.id || ''), event)}
    />;
  }

  function renderMergeRail() {
    const current = mergePreview || currentMergeJob || selectedMerge || null;
    const currentInfo = mergeInfo(current);
    return <section className={`batch-factory-media-pickstation-rail-card is-merge${viewerKind === 'merge' ? ' is-active' : ''}`}>
      <header>
        <div><span className="batch-factory-media-pickstation-index">成</span><b>合成成片</b></div>
        <small>{current ? `${mergeJobsNewest.length} 个上传组成${historicalMergeCount ? ` · 历史 ${historicalMergeCount} 条` : ''}` : '尚无成片'}</small>
      </header>
      {current ? <button type="button" className="batch-factory-media-pickstation-main" onClick={() => selectMergePreview(current)}>
        {current.status === 'succeeded' && String(current.outputUrl || '').trim()
          ? renderThumb(current, currentInfo, `合成成片 ${batchFactoryStableMediaLabel('M', current.id)}`)
          : <PassiveMediaPoster className="batch-factory-media-pickstation-thumb is-landscape" label="等待成片" detail={current.status === 'failed' ? '合成失败' : '处理中'} />}
        <span><b>{current.openingLabel || '原开头'} · {batchFactoryStableMediaLabel('M', current.id)}</b><small>{formatDurationSeconds(currentInfo.duration)} · {mergeJobSpeed(current).toFixed(1)}x · {currentInfo.ratioText}</small><em>{mergeJobsNewest.length > 1 ? '上传组成' : selectedMerge?.id === current.id ? '当前上传来源' : '合成成片'}</em></span>
      </button> : <div className="batch-factory-media-pickstation-empty">全部分镜选定主版本后，可在底部发起合成。</div>}
      <div className="batch-factory-media-pickstation-versions">
        {mergeJobsNewest.map(job => {
          const info = mergeInfo(job);
          const selected = job.id === current?.id;
          return <button
            type="button"
            key={job.id}
            className={`batch-factory-media-pickstation-version${selected ? ' is-selected' : ''}`}
            style={{ '--bf-media-ratio': info.ratio }}
            onClick={() => { setSelectedMergePreviewId(job.id); selectMergePreview(job); }}
            title={`查看 ${job.openingLabel} ${batchFactoryStableMediaLabel('M', job.id)}`}
          >
            <PassiveMediaPreview media={job} className="batch-factory-media-pickstation-version-media" label={`${job.openingLabel} ${batchFactoryStableMediaLabel('M', job.id)}`} onLoadedMetadata={event => rememberMediaMetrics(info.key, event)} />
            <span>{job.openingLabel}</span>
          </button>;
        })}
        <Tooltip title="上传成片">
          <span className="batch-factory-media-pickstation-add" title="上传成片">
            <input ref={el => { mergeUploadInputRef.current = el; }} type="file" accept="video/mp4,.mp4" hidden onChange={uploadMergeMaterial} />
            <Button size="small" loading={savingTaskId === 'upload-merge'} disabled={savingTaskId === 'upload-merge'} onClick={() => mergeUploadInputRef.current?.click()}>＋</Button>
          </span>
        </Tooltip>
      </div>
    </section>;
  }

  function renderVideoRail(video, index) {
    const state = videoState(video);
    const main = state.main;
    const info = mediaInfo(main, video);
    const active = viewerKind === 'video' && video.id === selectedVideoId;
    const uploadMain = selectedUpload?.kind === 'video' && selectedUpload.videoId === video.id && selectedUpload.taskId === main?.id;
    return <section key={video.id} className={`batch-factory-media-pickstation-rail-card${active ? ' is-active' : ''}`}>
      <header>
        <div><span className="batch-factory-media-pickstation-index">{String(index + 1).padStart(2, '0')}</span><b>{label(video)}</b></div>
        <small>{main ? `主 ${batchFactoryStableMediaLabel('V', main.id)}${main.isHistory ? ' · 上一版' : ''}` : state.active ? '生成中' : state.failed ? '失败' : '待生成'}</small>
      </header>
      {main?.mediaUrl ? <button type="button" className="batch-factory-media-pickstation-main" onClick={() => selectVideoPreview(video, main)}>
        {renderThumb(main, info, `${label(video)} 主版本`)}
        <span><b>{batchFactoryStableMediaLabel('V', main.id)} · {info.ratioText}</b><small>{formatDurationSeconds(main.actualDurationSeconds)} · {state.list.length} 个版本</small><em>{uploadMain ? '上传来源' : '合成主版本'}</em></span>
      </button> : <button type="button" className="batch-factory-media-pickstation-main is-empty" onClick={() => selectVideoPreview(video, null, false)}>
        <PassiveMediaPoster className="batch-factory-media-pickstation-thumb is-landscape" label={`分镜 ${index + 1}`} detail={state.active ? '生成中' : state.failed ? '生成失败' : '暂无视频'} />
        <span><b>{state.progress?.message || '暂无可用主版本'}</b><small>{state.active ? '任务会自动刷新' : '生成成功后显示版本'}</small></span>
      </button>}
      <div className="batch-factory-media-pickstation-versions">
        {state.list.map(version => {
          const versionInfo = mediaInfo(version, video);
          const isPrimary = version.id === main?.id;
          const isUpload = selectedUpload?.kind === 'video' && selectedUpload.videoId === video.id && selectedUpload.taskId === version.id;
          return <span className="batch-factory-media-pickstation-version-wrap" key={version.id}>
            <button
              type="button"
              className={`batch-factory-media-pickstation-version${isPrimary ? ' is-selected' : ''}`}
              style={{ '--bf-media-ratio': versionInfo.ratio }}
              onClick={() => choosePrimary(video, version)}
              disabled={savingTaskId === version.id}
              title={isPrimary ? '当前合成主版本' : `切换到 ${batchFactoryStableMediaLabel('V', version.id)} 并作为合成主版本`}
            >
              <PassiveMediaPreview media={version} className="batch-factory-media-pickstation-version-media" label={batchFactoryStableMediaLabel('V', version.id)} onLoadedMetadata={event => rememberMediaMetrics(String(version.id), event)} />
              <span>{batchFactoryStableMediaLabel('V', version.id)}{version.isHistory ? ' · 上一版' : ''}{!isPrimary && !version.isHistory ? ' · 其它版本' : ''}</span>
              {isUpload ? <i>传</i> : null}
            </button>
            {!isPrimary && !isUpload && !version.isHistory ? <Popconfirm title="移除候选版本？" description="只移除当前书片段库引用，不删除远端媒体或历史合成记录。" okText="移除" cancelText="取消" onConfirm={() => removeCandidate(video, version)}>
              <button type="button" className="batch-factory-media-pickstation-version-remove" aria-label={`移除 ${batchFactoryStableMediaLabel('V', version.id)}`}>×</button>
            </Popconfirm> : null}
          </span>;
        })}
        <Tooltip title="上传素材">
          <span className="batch-factory-media-pickstation-add" title="上传素材">
            <input ref={element => { uploadInputRefs.current[video.id] = element; }} type="file" accept="video/mp4,.mp4" hidden disabled={savingTaskId === `upload-material-${video.id}`} onChange={event => uploadVideoMaterial(video, event)} />
            <Button size="small" loading={savingTaskId === `upload-material-${video.id}`} disabled={savingTaskId === `upload-material-${video.id}`} onClick={() => uploadInputRefs.current[video.id]?.click()}>＋</Button>
          </span>
        </Tooltip>
        {state.failed ? <Tooltip title={productionAvailable ? '重试当前分镜最后失败的视频任务。' : productionReason}><Button size="small" danger disabled={!productionAvailable || regenerating} onClick={() => onRetry?.(video.id)}>重试</Button></Tooltip> : null}
      </div>
    </section>;
  }

  const viewerMerge = viewerKind === 'merge' ? mergePreview : null;
  const viewerVideo = viewerKind === 'video' ? selectedVideo : null;
  const viewerVersion = viewerKind === 'video' ? selectedVersion : null;
  const viewerInfo = viewerKind === 'merge' ? mergeInfo(viewerMerge) : mediaInfo(viewerVersion, viewerVideo);
  const viewerUrl = viewerKind === 'merge'
    ? batchFactoryPassiveMediaUrl(viewerMerge)
    : String(viewerVersion?.mediaUrl || '').trim();
  const viewerTitle = viewerKind === 'merge'
    ? (viewerMerge ? `合成成片 ${batchFactoryStableMediaLabel('M', viewerMerge.id)}` : '合成成片')
    : (viewerVideo ? label(viewerVideo) : '分镜');
  const viewerMeta = viewerKind === 'merge'
    ? (viewerMerge ? `${formatDurationSeconds(viewerInfo.duration)} · ${mergeJobSpeed(viewerMerge).toFixed(1)}x · ${viewerInfo.ratioText}` : '—')
    : (viewerVersion ? `${batchFactoryStableMediaLabel('V', viewerVersion.id)} · ${viewerInfo.ratioText}` : '—');

  const primaryVideoDurationRows = videos.map(video => {
    const main = videoState(video).main;
    if (!main) return { actualSeconds: 0, plannedSeconds: 0, exact: false };
    const actual = Number(main.actualDurationSeconds || mediaMetrics[String(main.id || '')]?.duration || 0);
    if (actual > 0) return { actualSeconds: actual, plannedSeconds: actual, exact: true };
    return { actualSeconds: 0, plannedSeconds: Number(main.requestedDurationSeconds || video.durationSeconds || 0), exact: false };
  });
  const videoTotalSeconds = primaryVideoDurationRows.reduce((sum, item) => sum + (Number.isFinite(item.plannedSeconds) ? item.plannedSeconds : 0), 0);
  const videoTotalIsExact = primaryVideoDurationRows.length > 0 && primaryVideoDurationRows.every(item => item.exact && item.actualSeconds > 0);
  // 只有全部主版本都读到真实时长才显示合计；否则显示「未知」，不拿计划值凑数。
  const videoTotalDisplay = videoTotalIsExact ? `${videoTotalSeconds.toFixed(1)}s` : '未知';
  const autoMergeRateRaw = Number(audioDurationSeconds || 0) > 0 && videoTotalSeconds > 0 ? videoTotalSeconds / Number(audioDurationSeconds) : 0;
  const autoMergeRate = autoMergeRateRaw > 0 ? Math.min(4, Math.max(0.5, autoMergeRateRaw)) : 0;
  const estimatedFinalSeconds = autoMergeRate > 0 ? videoTotalSeconds / autoMergeRate : 0;

  const mergeTooltip = !mergeAvailable
    ? mergeReason
    : !mergeReady
      ? `还有 ${missingVideos.length} 个分镜没有可用主版本`
      : followAudio && Number(audioDurationSeconds || 0) <= 0
          ? '点击合成时会先自动生成配音并读取真实时长'
          : '按各分镜当前主版本合成；不会覆盖历史成片。';

  const currentSheetOffset = sheetOffset == null ? sheetAnchorOffset() : sheetOffset;

  const precompiledWorkspace = resolvePrecompiledVideoWorkspace(book);
  const selectedPrecompiledFrame = resolvePrecompiledStoryboardFrame(book, selectedVideoId);
  const movePrecompiledFrame = direction => {
    const frame = precompiledWorkspace.frames?.[(selectedPrecompiledFrame?.index || 0) + direction];
    if (frame) setSelectedVideoId(frame.key);
  };
  if (precompiledWorkspace.status === 'awaiting_compilation') {
    const card = selectedPrecompiledFrame?.card || {};
    const camera = card?.camera || {};
    const currentIndex = selectedPrecompiledFrame?.index || 0;
    return <div className="batch-factory-media-pickstation batch-factory-media-precompiled-workspace">
      <div className="batch-factory-media-pickstation-topline">
        <span>左侧当前分镜预览 · 右侧选择导演卡；编译完成后自动切换为 VIDEO 版本库</span>
        <b>0/{precompiledWorkspace.frames.length} 分镜已就绪</b>
      </div>
      <div className="batch-factory-media-pickstation-workspace">
        <section className="batch-factory-media-pickstation-viewer">
          <header>
            <div><b>VIDEO {String(currentIndex + 1).padStart(2, '0')}</b><small>导演分镜卡 · 待编译</small></div>
            <Tag color="default">待生成</Tag>
          </header>
          <div className="batch-factory-media-pickstation-stage">
            <div className="batch-factory-media-pickstation-frame is-landscape" style={{ '--bf-media-ratio': 16 / 9 }}>
              <div className="batch-factory-media-primary-empty"><PictureOutlined /><span>当前分镜暂不可播放视频</span></div>
            </div>
          </div>
          <footer><span>当前为导演分镜卡 {(currentIndex || 0) + 1}；{h3CompilationStatus(mergeSettings)} Prompt。</span><small>待编译</small></footer>
        </section>
        <aside className="batch-factory-media-pickstation-rail">
          <section className="batch-factory-media-pickstation-rail-card">
            <header><div><span className="batch-factory-media-pickstation-index">成</span><b>合成成片</b></div><small>尚无成片</small></header>
            <div className="batch-factory-media-pickstation-main is-empty"><PassiveMediaPoster className="batch-factory-media-pickstation-thumb is-landscape" label="合成成片" detail="全部 VIDEO 编译并生成后可合成" /><span><b>等待 VIDEO 生成</b><small>不会在导演阶段伪造成片</small></span></div>
          </section>
          {precompiledWorkspace.frames.map((frame, index) => {
            const frameCard = frame.card || {};
            const active = frame.key === selectedPrecompiledFrame?.key;
            return <section key={frame.key} className={`batch-factory-media-pickstation-rail-card${active ? ' is-active' : ''}`}>
              <header><div><span className="batch-factory-media-pickstation-index">{String(index + 1).padStart(2, '0')}</span><b>VIDEO {String(index + 1).padStart(2, '0')}</b></div><small>待生成</small></header>
              <button type="button" className="batch-factory-media-pickstation-main is-empty" onClick={() => setSelectedVideoId(frame.key)}>
                <PassiveMediaPoster className="batch-factory-media-pickstation-thumb is-landscape" label={`分镜 ${index + 1}`} detail="暂无视频" />
                <span><b>{String(frameCard.source_text || '导演分镜卡')}</b><small>{frameCard.action ? `动作：${frameCard.action}` : '等待最终 VIDEO 编译'}</small></span>
              </button>
              <div className="batch-factory-media-pickstation-versions"><Button size="small" type={active ? 'primary' : 'default'} onClick={() => onOpenDirector?.(frame.key)}>打开分镜提示词</Button></div>
            </section>;
          })}
        </aside>
      </div>
      <section className="batch-factory-media-bottom-sheet is-precompiled-collapsed">
        <div className="batch-factory-media-bottom-sheet-bar">
          <div className="batch-factory-media-bottom-sheet-ready"><span><PictureOutlined /></span><div><b>0 / {precompiledWorkspace.frames.length} 分镜已选好</b><small>先完成最终 VIDEO 编译，再生成、选择与合成视频。</small></div></div>
          <div className="batch-factory-media-bottom-sheet-quick"><Button size="small" onClick={() => onOpenDirector?.(selectedPrecompiledFrame?.key || '')}>查看导演提示词</Button><Tooltip title="等待最终 VIDEO 编译"><Button type="primary" disabled>合成当前书</Button></Tooltip><Tooltip title="等待可上传视频"><Button size="small" className="batch-factory-upload-network-button" disabled>上传网络</Button></Tooltip></div>
        </div>
      </section>
    </div>;
  }

  return <div className="batch-factory-media-pickstation">
    <div className="batch-factory-media-pickstation-topline">
      <span>左侧看片 · 右侧选片 · 小版本点击后立即成为该分镜的合成主版本</span>
      <b>{readyVideos.length}/{videos.length} 分镜已就绪</b>
    </div>

    <div className="batch-factory-media-pickstation-workspace">
      <section className="batch-factory-media-pickstation-viewer">
        <header>
          <div><b>{viewerTitle}</b><small>{viewerMeta}</small></div>
          <Tag color={viewerKind === 'merge' ? 'gold' : 'success'}>{viewerKind === 'merge' ? '成片预览' : '合成主版本'}</Tag>
        </header>
        <div className="batch-factory-media-pickstation-stage">
          <div className={`batch-factory-media-pickstation-frame is-${viewerInfo.orientation}`} style={{ '--bf-media-ratio': viewerInfo.ratio }}>
            {/* 合成成片等待首帧时先显示同一张 merge-cover，避免黑屏 0:00。 */}
            {viewerKind === 'merge' && viewerMerge
              ? <MergeCoverThumb url={batchFactoryPassivePosterUrl(viewerMerge)} label="合成成片封面" alt="" />
              : null}
            {viewerUrl ? <ProductionMediaBoundary showDownload={false}><video
              key={`${viewerKind}:${viewerMerge?.id || viewerVersion?.id || ''}`}
              src={viewerUrl}
              controls
              preload="metadata"
              playsInline
              autoPlay={playRequested}
              onLoadedMetadata={event => primeViewerFrame(viewerKind === 'merge' ? viewerInfo.key : String(viewerVersion?.id || ''), event)}
              onLoadedData={event => primeViewerFrame(viewerKind === 'merge' ? viewerInfo.key : String(viewerVersion?.id || ''), event)}
            /></ProductionMediaBoundary> : <div className="batch-factory-media-primary-empty">
              {viewerVideo && videoState(viewerVideo).active ? <CmLoader size={30} title="视频生成中" /> : <PictureOutlined />}
              <span>{viewerKind === 'merge' ? '当前没有可播放的合成成片' : '当前分镜暂无可播放视频'}</span>
            </div>}
          </div>
        </div>
        <footer>
          {viewerKind === 'video' && viewerVideo ? <span>正在看 <b>{viewerVersion ? batchFactoryStableMediaLabel('V', viewerVersion.id) : '—'}</b>　｜　合成用 <b>{videoState(viewerVideo).main ? batchFactoryStableMediaLabel('V', videoState(viewerVideo).main.id) : '未选择'}</b></span>
            : <span>历史成片只用于预览和上传选择，不会自动改写分镜。</span>}
          <small>{viewerInfo.width > 0 && viewerInfo.height > 0 ? `${viewerInfo.width} × ${viewerInfo.height} · ` : ''}{viewerInfo.ratioText}</small>
        </footer>
      </section>

      <aside className="batch-factory-media-pickstation-rail">
        {renderMergeRail()}
        {videos.map(renderVideoRail)}
      </aside>
    </div>

    <section
      ref={sheetRef}
      className={`batch-factory-media-bottom-sheet is-${sheetAnchor}${sheetDragging ? ' is-dragging' : ''}`}
      style={{ transform: `translateY(${currentSheetOffset}px)` }}
    >
      <div
        className="batch-factory-media-bottom-sheet-handle"
        onPointerDown={startSheetDrag}
        onPointerMove={moveSheetDrag}
        onPointerUp={endSheetDrag}
        onPointerCancel={endSheetDrag}
      ><i /></div>

      <div className="batch-factory-media-bottom-sheet-bar">
        <div className="batch-factory-media-bottom-sheet-ready">
          <span className={mergeReady ? 'is-ready' : ''}><PictureOutlined /></span>
          <div><b>{readyVideos.length} / {videos.length} 分镜已选好</b><small>{viewerKind === 'video' && viewerVideo ? `当前：${label(viewerVideo)} · ${videoState(viewerVideo).main ? batchFactoryStableMediaLabel('V', videoState(viewerVideo).main.id) : '未选主版本'}` : uploadSourceText}</small></div>
        </div>
        <div className="batch-factory-media-bottom-sheet-quick">
          {!followAudio ? <InputNumber min={0.5} max={4} step={0.1} value={mergeSpeed} onChange={value => onMergeSpeedChange?.(Number(value || 1))} addonAfter="x" /> : <span className="batch-factory-audio-quick-state">{audioBusy ? '读取配音…' : `配音 ${Number(audioDurationSeconds || 0).toFixed(2)}s · 自动 ${autoMergeRate > 0 ? autoMergeRate.toFixed(3) : '—'}x`}</span>}
          <Button size="small" onClick={toggleSheet}>{sheetAnchor === 'collapsed' ? '展开' : sheetAnchor === 'half' ? '全展开' : '收起'}</Button>
          <Tooltip title={mergeTooltip}><Button type="primary" loading={merging} disabled={!mergeAvailable || merging || !mergeReady} onClick={runCurrentMerge}>合成当前书</Button></Tooltip>
          <Tooltip title="上传当前小说"><Button size="small" className="batch-factory-upload-network-button" onClick={onUpload} disabled={!mergedReady && !uploadVersion?.mediaUrl}>上传网络</Button></Tooltip>
        </div>
      </div>

      <div className="batch-factory-media-bottom-sheet-content">
        <div className="batch-factory-media-bottom-sheet-grid">
          <section>
            <h3>合成控制</h3>
            <div className="batch-factory-media-bottom-sheet-row"><span>本次倍率</span>{followAudio ? <b>{autoMergeRate > 0 ? `自动 · ${autoMergeRate.toFixed(3)}x` : '自动计算'}</b> : <InputNumber min={0.5} max={4} step={0.1} value={mergeSpeed} onChange={value => onMergeSpeedChange?.(Number(value || 1))} addonAfter="x" />}</div>
            <div className="batch-factory-media-bottom-sheet-row"><span>合并跟随配音</span><div className="batch-factory-follow-audio-control"><Switch checked={followAudio} loading={audioBusy} onChange={toggleFollowAudio} /><small>可事后开启并自动读取当前书配音时长</small></div></div>
            {followAudio ? <><div className="batch-factory-media-bottom-sheet-row"><span>配音目标</span><div className="batch-factory-audio-duration-edit"><InputNumber min={0.1} step={0.01} value={audioDurationSeconds || undefined} onChange={value => setAudioDurationSeconds(Number(value || 0))} onBlur={persistManualAudioDuration} addonAfter="s" /><Button size="small" loading={audioBusy} onClick={() => measureMergeAudioDuration({ force: true })}>重新读取</Button></div></div><div className="batch-factory-media-bottom-sheet-row"><span>主视频总时长</span><b>{videoTotalDisplay}</b></div><div className="batch-factory-media-bottom-sheet-row"><span>自动倍率</span><b>{autoMergeRate > 0 ? `${autoMergeRate.toFixed(3)}x` : '—'}</b></div><div className="batch-factory-media-bottom-sheet-row"><span>预计成片</span><b>{estimatedFinalSeconds > 0 ? `${estimatedFinalSeconds.toFixed(2)}s` : '—'}</b></div></> : null}
            <div className="batch-factory-media-bottom-sheet-actions"><Tooltip title={mergeTooltip}><Button type="primary" loading={merging || audioBusy} disabled={!mergeAvailable || merging || !mergeReady} onClick={runCurrentMerge}>{mergeJob?.status === 'succeeded' ? '重新合成' : '合成当前书'}</Button></Tooltip></div>
          </section>

          <section>
            <h3>上传与成片</h3>
            <div className="batch-factory-media-bottom-sheet-row"><span>当前上传来源</span><b>{uploadSourceText}</b></div>
            <Select value={pendingUploadKey || undefined} onChange={setPendingUploadKey} options={uploadOptions} placeholder="选择成片或分镜主版本" style={{ width: '100%' }} />
            <div className="batch-factory-media-bottom-sheet-actions">
              <Button onClick={applyPendingUploadSource} disabled={!pendingUploadKey || pendingUploadKey === uploadSourceKey() || Boolean(savingTaskId)} loading={Boolean(savingTaskId)}>设为上传来源</Button>
              <Tooltip title="上传当前小说"><Button className="batch-factory-upload-network-button" onClick={onUpload} disabled={!mergedReady && !uploadVersion?.mediaUrl}>上传网络</Button></Tooltip>
            </div>
          </section>

          <section className="is-wide">
            <h3>本次合成将使用</h3>
            <div className="batch-factory-media-bottom-sheet-snapshot">
              {videos.map((video, index) => {
                const main = videoState(video).main;
                return <button type="button" key={video.id} onClick={() => selectVideoPreview(video, main)}><b>分镜 {String(index + 1).padStart(2, '0')}</b><span>{main ? batchFactoryStableMediaLabel('V', main.id) : '未就绪'}</span></button>;
              })}
            </div>
          </section>
        </div>
        <p className="batch-factory-media-bottom-sheet-note">拖拽柄支持收起 / 半展开 / 全展开。超出边界时使用阻尼橡皮筋，松手根据手势速度选择吸附锚点。上传网络继续进入现有视频管理系统登录、组织、来源书城、解压数量与重复上传检查。</p>
      </div>
    </section>
  </div>;
}


// 执行日志是只追加的历史。runtime-summary 轮询只带回每本书最近一段日志，
// 如果每次轮询都整体覆盖，用户刚看到的日志会在下一次刷新时消失（闪空）。
// 这里按「时间 + 阶段 + 文案」去重合并，每本最多保留最近 300 条：
// 打开面板立即渲染已有快照，后续轮询/手动刷新只追加，不清空、不闪回加载态。
function mergeBookExecutionLogs(previousLogs, nextLogs) {
  const merged = new Map();
  const push = item => {
    if (!item || typeof item !== 'object') return;
    const key = `${item.at || ''}|${item.stage || ''}|${item.message || ''}`;
    if (!merged.has(key)) merged.set(key, item);
  };
  (Array.isArray(previousLogs) ? previousLogs : []).forEach(push);
  (Array.isArray(nextLogs) ? nextLogs : []).forEach(push);
  return Array.from(merged.values()).slice(-300);
}

function DeferredInlineComicLibrary(props) {
  const placeholderRef = useRef(null);
  const [loaded, setLoaded] = useState(false);
  useEffect(() => {
    if (loaded) return undefined;
    const node = placeholderRef.current;
    if (!node || typeof IntersectionObserver === 'undefined') {
      setLoaded(true);
      return undefined;
    }
    const observer = new IntersectionObserver(entries => {
      setLoaded(current => nextViewportLoaded(current, entries));
    }, { rootMargin: '600px 0px' });
    observer.observe(node);
    return () => observer.disconnect();
  }, [loaded]);
  if (loaded) return <InlineComicLibrary {...props} />;
  return <div ref={placeholderRef} className="batch-factory-cell-shell batch-factory-inline-media is-empty">
    <div className="batch-factory-cell-empty">滚动到附近后加载漫画页。</div>
  </div>;
}

function mergeAutomationLogSnapshot(previous, next) {
  if (!next || typeof next !== 'object') return previous;
  const previousBooks = Array.isArray(previous?.books) ? previous.books : [];
  const nextBooks = Array.isArray(next?.books) ? next.books : [];
  const seen = new Set();
  const books = nextBooks.map(book => {
    const key = String(book?.bookId || book?.id || '');
    seen.add(key);
    const before = previousBooks.find(item => String(item?.bookId || item?.id || '') === key);
    return { ...book, logs: mergeBookExecutionLogs(before?.logs, book?.logs) };
  });
  // 已经从聚合接口消失的书仍保留它已加载的日志，避免“刷新后日志变少”。
  previousBooks.forEach(book => {
    const key = String(book?.bookId || book?.id || '');
    if (seen.has(key)) return;
    const logs = mergeBookExecutionLogs(book?.logs, []);
    if (logs.length) books.push({ ...book, logs });
  });
  return { ...next, books };
}

function BatchLogs({ automationStatus, productionStatus, mergeStatus, error, books = [], productionMode = 'video', onRetry, syncing = false, syncedAt = '' }) {
  const [expandedBooks, setExpandedBooks] = useState({});
  const stepStatusLabel = status => ({ completed: '已完成', running: '执行中', failed: '失败', pending: '等待' })[status] || status;
  const formatLogTime = value => {
    if (!value) return '—';
    const parsed = new Date(value);
    return Number.isNaN(parsed.getTime()) ? String(value) : parsed.toLocaleTimeString('zh-CN', { hour12: false });
  };
  const automationBookRecord = entry => books.find(item => item?.id === entry?.bookId)
    || books.find(item => String(item?.bookId || '') === String(entry?.bookId || ''));
  const jobs = productionMode === 'comic' ? [] : (productionStatus?.jobs || []);
  const merges = productionMode === 'comic' ? [] : (mergeStatus?.jobs || []);
	const activeMerge = merges.find(merge => ['running', 'queued'].includes(String(merge?.status || '').toLowerCase()) && String(merge?.progressPhase || '').toLowerCase() !== 'queued');
	const queuedMerges = merges.filter(merge => String(merge?.status || '').toLowerCase() === 'queued').slice().sort((left, right) => String(left?.createdAt || '').localeCompare(String(right?.createdAt || '')));
	const mergeQueueDetail = merge => {
	  const status = String(merge?.status || '').toLowerCase();
	  if (status === 'running') return `当前工位 · ${merge.progressPhase || '处理中'} · 已处理 ${Number(merge.progressCurrent || 0)} / ${Number(merge.progressTotal || 0)} 个分镜`;
	  if (status === 'queued') {
	    const ahead = Math.max(0, queuedMerges.findIndex(item => item?.id === merge?.id)) + (activeMerge ? 1 : 0);
	    return `合成队列 · 前方 ${ahead} 项${activeMerge ? ` · 当前工位 · ${activeMerge.progressPhase || '处理中'}` : ' · 等待工位'}`;
	  }
	  return merge.outputUrl || batchFactoryVisibleError(merge.errorMessage) || '等待合并结果';
	};
	const automationBookDetail = book => {
	  if (String(book?.stage || '').toLowerCase() !== 'merge') return '';
	  const bookID = String(book?.bookId || '');
	  const bookMerges = merges.filter(merge => String(merge?.bookId || '') === bookID).slice().sort((left, right) => String(right?.createdAt || '').localeCompare(String(left?.createdAt || '')));
	  const activeBookMerge = bookMerges.find(merge => ['running', 'queued'].includes(String(merge?.status || '').toLowerCase())) || bookMerges[0];
	  return activeBookMerge ? mergeQueueDetail(activeBookMerge) : '';
	};
	const automationStageLabel = stage => ({
		comic_assets: '漫画资产',
		comic_asset_images: '漫画资产出图',
		comic_bootstrap: '漫画分段与提示词',
		comic_prompt_recovery: '补齐漫画提示词',
		comic_resume: '恢复漫画生产',
		comic_images: '漫画图片',
		comic_publish: '发布漫画合集',
		audio_planning: '配音与逐行时长',
		waiting_submit: '等待发布',
	})[String(stage || '').trim()] || String(stage || '等待自动生产');
	// 状态枚举直接渲染会把 failed/running 这类英文露到界面上，统一做中文映射。
	const automationStatusLabel = status => {
		const key = String(status || '').trim().toLowerCase();
		const labels = {
			pending: '等待中',
			queued: '排队中',
			ready: '待执行',
			running: '执行中',
			processing: '执行中',
			generating: '生成中',
			completed: '已完成',
			succeeded: '成功',
			success: '成功',
			failed: '失败',
			error: '失败',
			blocked: '已阻塞',
			stopped: '已停止',
			paused: '已暂停',
			cancelled: '已取消',
			canceled: '已取消',
			needs_attention: '待处理',
			skipped: '已跳过',
		};
		return labels[key] || String(status || '等待中');
	};
  const automationBooks = Array.isArray(automationStatus?.books) ? automationStatus.books : [];
  const automationCounts = automationStatus?.counts || {};
  const runMode = String(automationStatus?.runMode || 'full_submit');
  const modeDescription = productionMode === 'comic'
    ? runMode === 'storyboard_only'
      ? '当前任务只提取漫画资产、规划整本分页并生成每页提示词，不会出图或发布。'
      : runMode === 'video_no_submit'
        ? '当前任务会生成缺少主图的漫画页并整理漫画合集，不会发布到漫画后台。'
        : '当前任务会生成缺少主图的漫画页、整理合集，并按统一发布配置提交漫画后台。'
    : runMode === 'storyboard_only'
      ? '当前任务仅生成分镜与最终 VIDEO Prompt，停在“待生成视频”；不会提交视频、合成或上传。'
      : runMode === 'video_no_submit'
        ? '当前任务会生成视频与合成成片，停在“待上传”；刷新只回读状态，不会额外提交新任务。'
        : '自动生产会在成片后提交视频管理系统，并等待视频管理系统回读确认；未确认前不会显示上传成功。';
  const giantWaitingBooks = books.filter(book => book?.sourceMetadata?.giantMaterialId && isBatchFactoryContentReadRunning(book));
  const isComicRetryQueued = book => book?.retryRequested === true && !['failed', 'blocked', 'stopped', 'completed'].includes(String(book?.status || '').toLowerCase());
  const automationLogCount = automationBooks.reduce((total, book) => total + (Array.isArray(book?.logs) ? book.logs.length : 0), 0);
  const latestLogAt = automationBooks.reduce((latest, book) => {
    const bookLogs = Array.isArray(book?.logs) ? book.logs : [];
    bookLogs.forEach(item => {
      const time = new Date(item?.at || '').getTime();
      if (Number.isFinite(time) && time > latest) latest = time;
    });
    return latest;
  }, 0);
  return <><div className="batch-factory-log-sync" data-testid="batch-log-sync-status">
    <span>已加载 {automationBooks.length} 本任务 · {automationLogCount} 条执行日志{latestLogAt ? ` · 最近一条 ${formatLogTime(latestLogAt)}` : ''}{syncedAt ? ` · 同步于 ${syncedAt}` : ''}</span>
    <span>{syncing ? '后台同步中，已加载内容保持可见' : '已显示当前快照，可手动刷新'}</span>
  </div>
    {error ? <Alert type="warning" showIcon message="更早的日志加载失败或超时，已保留当前内容" description={error} action={onRetry ? <Button size="small" onClick={onRetry}>重试</Button> : null} /> : <Alert type="info" showIcon message={productionMode === 'comic' ? '实时读取漫画自动生产与发布状态' : '实时读取 V12 自动生产、视频与合并状态'} description={modeDescription} />}
    {automationStatus?.state && automationStatus.state !== 'idle' ? <Alert type={['paused', 'needs_attention', 'unavailable'].includes(automationStatus.state) ? 'warning' : automationStatus.state === 'completed' ? 'success' : 'info'} showIcon message={automationStatus.state === 'unavailable' ? '自动生产状态暂不可读' : automationStatus.state === 'paused' ? '自动生产 · 已暂停' : `自动生产 · ${automationStatusLabel(automationStatus.state)}`} description={automationStatus.state === 'unavailable' ? '不影响当前书的手动提取、配音、VIDEO 或上传操作。' : automationStatus.state === 'paused' ? `${automationStatus.lastError || '失败项最终自动补跑后仍未成功，已暂停等待处理。'}；并发 ${Number(automationStatus.concurrency || 2)} 本；就绪 ${Number(automationCounts.ready || 0)} / ${Number(automationCounts.total || automationBooks.length)}；失败 ${Number(automationCounts.failed || 0)}；阻塞 ${Number(automationCounts.blocked || 0)}` : `并发 ${Number(automationStatus.concurrency || 2)} 本；就绪 ${Number(automationCounts.ready || 0)} / ${Number(automationCounts.total || automationBooks.length)}；执行中 ${Number(automationCounts.running || 0)}；失败 ${Number(automationCounts.failed || 0)}；阻塞 ${Number(automationCounts.blocked || 0)}`} /> : null}
	<div className="batch-factory-log-list">{giantWaitingBooks.map(book => <section key={`giant-source-${book.id}`}><strong>巨量素材 · {book.title || book.bookId}</strong><span>等待正文读取</span><p>{book.sourceMetadata?.originalReadStage === 'ocr' ? '已进入滚屏 OCR 队列，等待执行器领取；正文回填后自动继续生产。' : book.sourceMetadata?.originalReadStage === 'failed' ? '书城原文获取失败；可手动重试或使用 OCR。' : '正在优先通过书城读取正文。'}</p></section>)}{automationBooks.map(book => {
		const retryQueued = isComicRetryQueued(book);
		const record = automationBookRecord(book);
		const timeline = record ? batchFactoryBookTimeline(record, { productionStatus, mergeStatus, productionMode }) : [];
			const published = productionMode === 'comic'
				&& String(book.status || '').toLowerCase() === 'completed'
				&& String(book.stage || '').toLowerCase() === 'published';
			const visibleAutomationError = published ? '' : String(book.error || '');
			const displayTimeline = published
				? timeline.map(step => ({ ...step, status: 'completed', detail: step.key === 'complete' ? '漫画合集已发布' : step.detail }))
				: timeline;
			const finishedSteps = displayTimeline.filter(step => step.status === 'completed').length;
		const percent = displayTimeline.length ? Math.round((finishedSteps / displayTimeline.length) * 100) : (published ? 100 : 0);
		const logs = Array.isArray(book.logs) ? book.logs : [];
		const expanded = Boolean(expandedBooks[book.bookId]);
		return <section key={`automation-${book.bookId}`}>
			<strong>自动生产 · {book.title || book.bookId}</strong>
			<span>{automationStageLabel(book.stage)} · {retryQueued ? '重试排队中' : automationStatusLabel(book.status)}{book.status === 'failed' && book.failureCategoryLabel ? <em style={{ marginLeft: 6, padding: '1px 8px', borderRadius: 10, fontStyle: 'normal', fontSize: 12, border: '1px solid currentColor', color: book.failureCategory === 'billed_not_retrieved' ? '#f59e0b' : book.failureCategory === 'config_missing' || book.failureCategory === 'external_work_missing' ? '#ef4444' : '#38bdf8' }}>{book.failureCategoryLabel}</em> : null} · 完成率 {percent}% · {book.updatedAt || '—'}</span>
			<p>{retryQueued ? '已接受手动重试，正在等待生产工位；无需再次点击重试。' : automationBookDetail(book) || batchFactoryVisibleError(book.message) || '等待自动生产'}{Number(book.retryCount || 0) > 0 ? ` · 自动重试 ${Math.min(3, Number(book.retryCount || 0))}/3` : ''}{book.retryAt ? ` · 下次重试：北京时间 ${formatBeijingDatetimeLocal(book.retryAt)}` : ''}{visibleAutomationError ? ` · ${batchFactoryVisibleError(visibleAutomationError)}` : ''}</p>
			{visibleAutomationError ? <BatchFactoryErrorNotice compact title="自动生产失败" stageText={automationStageLabel(book.stage)} error={visibleAutomationError} /> : null}
			<Button size="small" type="link" onClick={() => setExpandedBooks(current => ({ ...current, [book.bookId]: !current?.[book.bookId] }))}>{expanded ? '收起执行日志' : '展开执行日志'}</Button>
			{expanded ? <div className="batch-factory-log-detail" style={{ marginTop: 4, padding: '8px 10px', background: 'var(--surface-2)', borderRadius: 8 }}>
				<div style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>链路执行（完成率 {percent}%）</div>
				{timeline.length ? timeline.map(step => <div key={step.key} style={{ fontSize: 12, lineHeight: '20px' }}>{stepStatusLabel(step.status)} · {step.label}{step.detail ? `：${step.detail}` : ''}</div>) : <div style={{ fontSize: 12 }}>暂无链路数据</div>}
				<div style={{ fontSize: 12, color: 'var(--text-tertiary)', marginTop: 8 }}>执行日志（最近 {logs.length} 条）</div>
				{logs.length ? logs.map((item, index) => <div key={`${item?.at || index}-${index}`} style={{ fontSize: 12, lineHeight: '20px' }}><span style={{ color: 'var(--text-tertiary)' }}>{formatLogTime(item?.at)}</span> · {automationStageLabel(item?.stage)} · {item?.message}</div>) : <div style={{ fontSize: 12 }}>暂无执行日志；去敏、审核拒绝与二次降级重试会记录在这里。</div>}
			</div> : null}
		</section>;
	})}{jobs.map(job => <section key={job.id}><strong>生成任务 · {automationStatusLabel(job.status)}</strong><span>{job.bookId || '批量任务'} · {job.updatedAt || job.createdAt || '—'}</span>{(job.tasks || []).map(task => { const durations = [['目标', task.targetDurationSeconds], ['请求', task.requestedDurationSeconds], ['实际', task.actualDurationSeconds]].filter(([, value]) => Number(value) > 0).map(([label, value]) => `${label} ${Number(value).toFixed(2).replace(/\.00$/, '')}s`).join(' · '); return <p key={task.id}>{task.videoId} · {automationStatusLabel(task.status)}{durations ? ` · ${durations}` : ''}{task.errorMessage ? ` · ${batchFactoryVisibleError(task.errorMessage)}` : ''}</p>; })}</section>)}{merges.map(merge => <section key={merge.id}><strong>合并任务 · {automationStatusLabel(merge.status)}</strong><span>{merge.updatedAt || merge.createdAt || '—'}</span><p>{mergeQueueDetail(merge)}</p></section>)}{!giantWaitingBooks.length && !automationBooks.length && !jobs.length && !merges.length ? <p>暂无执行日志{productionMode === 'comic' ? '：当前没有漫画自动生产或发布任务。' : '：当前没有自动生产、视频或合并任务。'}</p> : null}</div>
  </>;
}


function UploadNetwork({ batch, books, selectedBookIds, productionStatus, mergeStatus, mode, onClose, onOpenPublish, onRefresh }) {
  const [account, setAccount] = useState(null);
  const [publishSession, setPublishSession] = useState(null);
  const [manifest, setManifest] = useState(null);
  const [results, setResults] = useState([]);
  const [busy, setBusy] = useState(false);
  const [uploadingBookId, setUploadingBookId] = useState('');
  const [organizations, setOrganizations] = useState([]);
  const [organizationID, setOrganizationID] = useState('');
  const [organizationsError, setOrganizationsError] = useState('');
  const reupload = mode === 'reupload';
  const targetBooks = mode === 'all' ? books : books.filter(book => selectedBookIds.includes(book.id));
  const uploadMergeFiles = book => publishMergeFiles(
    book?.bookId,
    bookMergeJobs(mergeStatus, book?.id),
    effectiveBookSettings(batch, book),
    book?.videos?.length || 0
  );
  const hasBookUploadSource = book => uploadMergeFiles(book).ready;
  const uploadMediaReady = targetBooks.length > 0 && targetBooks.every(hasBookUploadSource);
  const publishSettings = batch?.settingsState?.patch?.publishSettings || {};
  const defaultOrganizationIDs = [...new Set(targetBooks
    .map(book => String(effectiveBookSettings(batch, book).publishSettings?.organization || '').trim())
    .filter(Boolean))];
  const defaultOrganizationID = defaultOrganizationIDs.length === 1 ? defaultOrganizationIDs[0] : '';
  const organizationSelectionMixed = defaultOrganizationIDs.length > 1;
  const hasSingleBookPublishOverride = targetBooks.some(book => Object.hasOwn(book?.settingsState?.patch || {}, 'publishSettings') || Object.hasOwn(book?.settingsState?.patch || {}, 'publishRewriteEnabled'));
  const materialReuseLabel = publishSettings.materialReuse === true ? '复用' : '不复用';
  const horizontalFlipLabel = publishSettings.horizontalFlip === true ? '翻转' : '不翻转';
  const missingPublishMapping = targetBooks.filter(book => {
    const settings = effectiveBookSettings(batch, book).publishSettings || {};
    return !completePublishMapping(settings, book);
  });
  const missingDecompression = targetBooks.filter(book => decompressionCountForUpload(batch, book) < 1);
  const alreadyUploaded = targetBooks.filter(book => isBookUploadedTo121(book));
  const requiresReupload = alreadyUploaded.length > 0 && !reupload;
  async function verify121Session() {
    const [config, environment] = await Promise.all([getWebSubmitConfig(), checkWebSubmitEnvironment()]);
    const settings = config?.settings || config?.config || {};
    setAccount(settings);
    const session = Boolean((environment?.checks || []).some(check => ['视频管理系统登录会话', '121 后台登录会话', '目标站登录会话'].includes(check.name) && check.ok));
    if (!session) {
      setPublishSession({ environment, visible: null });
      return false;
    }
    const visible = await testWebSubmitVisible();
    setPublishSession({ environment, visible });
    return visible?.ok === true;
  }
  useEffect(() => {
    if (!mode) return undefined;
    setManifest(null);
    setResults([]);
    setOrganizationID(defaultOrganizationID);
    let active = true;
    Promise.all([verify121Session(), get121OrganizationOptions()]).then(([, response]) => {
      if (!active) return;
      setOrganizations(Array.isArray(response?.organizations) ? response.organizations : []);
    }).catch(error => { if (active) { setOrganizationsError(error?.message || '121 组织目录读取失败'); setPublishSession({ error: error?.message || '121 后台会话验证失败' }); } });
    return () => { active = false; };
  }, [mode, defaultOrganizationID]);
  useEffect(() => {
    if (!uploadingBookId || !onRefresh) return undefined;
    const timer = setInterval(() => {
      if (document.hidden) return; // 切到后台时不刷，上传在服务端照常进行
      onRefresh().catch(() => {});
    }, 5000);
    return () => clearInterval(timer);
  }, [uploadingBookId, onRefresh]);
  const publishSessionReady = Boolean(publishSession?.environment?.ok && publishSession?.visible?.ok);
  async function prepareUpload() {
    setBusy(true);
    try {
      if (!(await verify121Session())) throw new Error('请先在发布统一登录并验证 121 后台账号');
      if (missingPublishMapping.length) throw new Error(`当前书缺少来源书城，无法上传：${missingPublishMapping.map(book => book.title || book.bookId).join('、')}`);
      if (missingDecompression.length) throw new Error(`请添加解压，才能上传 AI 前贴视频：${missingDecompression.map(book => book.title || book.bookId).join('、')}`);
      if (requiresReupload) throw new Error(`当前书已上传，请使用重新上传：${alreadyUploaded.map(book => book.title || book.bookId).join('、')}`);
      setManifest(targetBooks.map(book => {
        const effectiveSettings = effectiveBookSettings(batch, book);
        const mergeFiles = uploadMergeFiles(book);
        return {
          id: book.id,
          title: book.title || book.bookId,
          txt: `${book.bookId}.txt`,
          mp4Files: mergeFiles.files,
          videoCount: mergeFiles.files.length,
          videoType: effectiveSettings.publishSettings?.uploadVideoType === 'individual' ? '独立 VIDEO' : '最终合成视频',
          jieyaNum: decompressionCountForUpload(batch, book)
        };
      }));
    } catch (error) { message.error(error?.message || '生成上传清单失败'); } finally { setBusy(false); }
  }
  async function submitTo121() {
    setBusy(true);
    const submitted = [];
    try {
      for (const book of targetBooks) {
        try {
          setUploadingBookId(book.id);
          setResults(current => [...current.filter(item => item.id !== book.id), { id: book.id, title: book.title || book.bookId, pending: true, message: reupload ? '再次上传中：正在验证 121 会话、上传 AI 前贴与提交 TXT' : '正在上传：正在验证 121 会话、上传 AI 前贴与提交 TXT' }]);
          const result = await submitBookTo121(batch.id, book.id, { organization: organizationID, category: 'NEW_BOOK', startTime: localStartTime(), textModelId: effectiveBookSettings(batch, book).textModelId, reupload });
          submitted.push({ id: book.id, title: book.title || book.bookId, ok: true, result: resultData(result) });
          await onRefresh?.();
        } catch (error) {
          submitted.push({ id: book.id, title: book.title || book.bookId, ok: false, error: error?.message || '121 提交失败' });
          // Each book has its own confirmation and readback. One remote
          // rejection must not hold the remaining ready books in this queue.
        }
      }
      setResults(submitted);
      const failed = submitted.filter(item => !item.ok);
      if (failed.length) message.warning(`${failed.length} 本提交失败，其余 ${submitted.length - failed.length} 本已继续提交；失败原因已逐本保留。`);
      else message.success(submitted.every(item => item.result?.status === 'confirmed') ? '已提交到 121，并已完成后台列表回读。' : '已提交到 121，正在等待后台列表回读。');
    } catch (error) { message.error(error?.message || '121 提交失败'); } finally { setUploadingBookId(''); setBusy(false); }
  }
  const liveUploadBook = books.find(book => book.id === uploadingBookId);
  const liveUploadProgress = uploadProgressForBook(liveUploadBook);
  return <Modal title={`上传网络 · ${mode === 'all' ? '提交全部' : reupload ? '重新上传' : '提交选中'}`} open={Boolean(mode)} onCancel={onClose} footer={null} width={700} destroyOnClose>
    <Space direction="vertical" size={14} style={{ width: '100%' }}>
      <Alert type="info" showIcon message={`已选择 ${targetBooks.length} 本小说`} description="每本书只提交一次正文；开启换开头且至少有两个分镜时，会把原开头和全部换开头完整成片一起上传。提交后以视频管理系统后台列表回读为准。" />
      <Alert type="info" showIcon message="本次上传素材设置" description={hasSingleBookPublishOverride ? '上传时按每本书的单书发布覆盖优先、批量统一设置兜底；生效值会随确认单冻结，不会修改 121 后台配置方案。' : `本次素材使用：${materialReuseLabel}；本次水平翻转：${horizontalFlipLabel}。这两个值会随确认单冻结，不会修改 121 后台配置方案。`} />
      {missingPublishMapping.length ? <Alert type="warning" showIcon message="存在缺少来源书城的书" description={missingPublishMapping.map(book => book.title || book.bookId).join('、')} /> : null}
      <Alert type="info" showIcon message="逐本识别发布信息" description="提交前会按每本书正文自动识别男女频、121 风格和标签；已有手动或已识别结果会直接复用。识别失败会终止当前书，不会改用批量默认值。" />
      {missingDecompression.length ? <Alert type="warning" showIcon message="请添加解压" description={`以下小说的解压数量为 0，不能附带 AI 前贴视频上传：${missingDecompression.map(book => book.title || book.bookId).join('、')}`} /> : null}
      {requiresReupload ? <Alert type="warning" showIcon message="存在已上传小说" description="普通上传不会重复提交已成功回读的小说。请从该书“查看资料”中选择重新上传。" /> : null}
      {!uploadMediaReady ? <Alert type="warning" showIcon message="存在尚未准备好的上传主视频" description="每本书需要先在片段库选择可用分镜主版本，或完成该书的最终合成成片；确认单不会提交空视频或未完成视频。" /> : null}
      <Tag
        className={`batch-factory-publish-session-tag ${publishSessionReady ? 'is-ready' : 'is-unverified'}`}
        color={publishSessionReady ? 'green' : 'error'}
      >
        {publishSessionReady ? `121 后台已登录：${account?.username || '当前账号'}` : '121 后台账号尚未登录或未验证'}
      </Tag>
      {publishSession?.error ? <Alert type="warning" showIcon message="121 后台会话验证失败" description={publishSession.error} /> : null}
      {!publishSessionReady ? <Button onClick={onOpenPublish}>前往发布统一登录并验证</Button> : null}
      {organizationsError ? <Alert type="warning" showIcon message="121 组织目录读取失败" description={organizationsError} /> : null}
      {organizationSelectionMixed ? <Alert type="warning" showIcon message="所选小说存在不同的单书组织覆盖" description="请为本次上传统一选择一个组织归属；此处改选只作用于本次确认单，不会回写任何统一或单书配置。" /> : null}
      <Select value={organizationID || undefined} placeholder="请选择 121 组织归属" style={{ width: '100%' }} onChange={setOrganizationID} options={organizations.map(item => ({ value: item.id, label: item.level ? `${item.name}（${item.level}）` : item.name }))} disabled={!publishSessionReady || busy} />
      {!manifest ? (() => {
        const disabledReasons = [];
        if (!targetBooks.length) disabledReasons.push('本次没有可上传的小说');
        if (!publishSessionReady) disabledReasons.push('121 后台账号尚未登录或未验证');
        if (!uploadMediaReady) disabledReasons.push('存在尚未准备好的上传主视频');
        if (missingPublishMapping.length > 0) disabledReasons.push('存在发布信息没填完整的小说');
        if (missingDecompression.length > 0) disabledReasons.push('存在解压码数量不足的小说');
        if (requiresReupload) disabledReasons.push('所选小说已上传过，请从该书“查看资料”里选择重新上传');
        if (!organizationID) disabledReasons.push('请先选择 121 组织归属');
        return (
          <Tooltip title={disabledReasons.length ? `暂时不能生成：${disabledReasons.join('；')}` : ''}>
            <span>
              <Button type="primary" onClick={prepareUpload} loading={busy} disabled={disabledReasons.length > 0}>生成上传清单</Button>
            </span>
          </Tooltip>
        );
      })() : null}
      {manifest ? <><Descriptions size="small" column={1} bordered items={manifest.map(item => ({ key: item.id, label: item.title, children: <span>{item.txt} + {item.mp4Files.join('、')}（{item.videoCount} 个{item.videoType}，解压 {item.jieyaNum}）</span> }))} /><Alert type="warning" showIcon message={reupload ? '确认后将重新提交到视频管理系统' : '确认后将直接提交到视频管理系统'} description="提交到接口只表示对方已接收；本页会显示等待视频管理系统后台列表回读，不能替代后台完成状态。" /><Button danger type="primary" onClick={submitTo121} loading={busy}>{reupload ? '确认重新上传到视频管理系统' : '确认并上传到视频管理系统'}</Button></> : null}
      {liveUploadBook ? <Alert type="info" showIcon message={`${liveUploadBook.title || liveUploadBook.bookId}：${liveUploadProgress?.message || '正在创建上传任务'}`} description={liveUploadProgress?.phase ? `当前阶段：${liveUploadProgress.phase}` : '正在等待服务端返回当前阶段。'} /> : null}
      {results.map(item => <Alert key={item.id} type={item.pending ? 'info' : item.ok ? 'info' : 'error'} showIcon message={`${item.title}：${item.pending ? item.message : item.ok ? (item.result?.status === 'confirmed' ? '已提交，视频管理系统已回读' : '已提交，等待视频管理系统回读') : '提交失败'}`} description={item.pending ? '上传过程中的实际阶段会写入当前书状态；完成或失败后自动回读。' : item.ok ? `${item.result?.sourceTextFile || ''} + ${(item.result?.aiHeadVideoFiles || [item.result?.aiHeadVideoFile]).filter(Boolean).join('、')}${item.result?.receipt?.remote_record?.detail ? `；${item.result.receipt.remote_record.detail}` : ''}` : item.error} />)}
    </Space>
  </Modal>;
}

export function BatchFactoryNovelList({ batch, onBack, onBatchChanged, openAutomationOnLoad = false }) {
  const [novelListOpen, setNovelListOpen] = useState(false);
  const [isDeveloper, setIsDeveloper] = useState(false);
  const [viewingBook, setViewingBookRaw] = useState(null);
  const [sensitiveDetailBook, setSensitiveDetailBook] = useState(null);
  const [rawSourceBook, setRawSourceBook] = useState(null);
  const [sourceEditorBook, setSourceEditorBook] = useState(null);
  const [sourceEditorValue, setSourceEditorValue] = useState('');
  const [sourceSaving, setSourceSaving] = useState(false);
  // 小说详情只展示抓取后保存的完整原文；生产草稿只能在“编辑生产内容”中查看和修改。
  const setViewingBook = book => setViewingBookRaw(book ? { ...book, workingFrontContent: '' } : null);
  const [stageSummaries, setStageSummaries] = useState({});
  const [metadataBook, setMetadataBook] = useState(null);
  const [metadataValue, setMetadataValue] = useState({});
  const [metadataSaving, setMetadataSaving] = useState(false);
  const [assetBook, setAssetBook] = useState(null);
  const [configTarget, setConfigTarget] = useState(null);
  const [editingContentBook, setEditingContentBook] = useState(null);
  const [editingContentTitle, setEditingContentTitle] = useState('');
  const [editingContentSourceBookId, setEditingContentSourceBookId] = useState('');
  const [editingContentValue, setEditingContentValue] = useState('');
  const [editingContentMode, setEditingContentMode] = useState('original');
  const [editingContentRangeLines, setEditingContentRangeLines] = useState(5);
  const [editingContentAnalysisCharacters, setEditingContentAnalysisCharacters] = useState(2000);
  const [editingContentCaptureCharacters, setEditingContentCaptureCharacters] = useState(4000);
  const [viralCandidate, setViralCandidate] = useState('');
  const [derivedOpeningOptions, setDerivedOpeningOptions] = useState([]);
  const [derivedOpeningPresetId, setDerivedOpeningPresetId] = useState('');
  const [contentSaving, setContentSaving] = useState(false);
  const [rewritingFront, setRewritingFront] = useState(false);
	const [promptBook, setPromptBook] = useState(null);
	const [promptVideoId, setPromptVideoId] = useState('');
	const [mediaBook, setMediaBook] = useState(null);
	const [mediaVideoId, setMediaVideoId] = useState('');
	const [mediaStartTab, setMediaStartTab] = useState('clips');
	const [pendingMediaPrompt, setPendingMediaPrompt] = useState(null);
	const [rowStoryboardSelection, setRowStoryboardSelection] = useState({});
  const [unifiedSettingsOpen, setUnifiedSettingsOpen] = useState(false);
  // 保存统一配置后关弹窗立刻返回；配音测量/提示词重编译在后台按“每本书自己的
  // 引擎配置”继续跑，进度与失败都回显在这里，不再阻塞弹窗。
  const [backgroundSettingsTask, setBackgroundSettingsTask] = useState(null);
  const backgroundSettingsRunRef = useRef(0);
  const [logsOpen, setLogsOpen] = useState(false);
  const [logsLoading, setLogsLoading] = useState(false);
  const [logsError, setLogsError] = useState('');
  // 最近一次成功同步任务状态的时间。日志入口只读这个快照，不再“打开才请求”，
  // 因此必须把“已加载内容的时间点”显示出来，避免用户以为页面停在加载中。
  const [logsSyncedAt, setLogsSyncedAt] = useState('');
  const [automationStatus, setAutomationStatus] = useState({ state: 'idle', counts: { total: 0, ready: 0, running: 0, pending: 0, failed: 0, blocked: 0 } });
  // 新失败只弹一次临时提醒；完整错误统一在“任务/日志”和单本详情中查看。
  const notifiedAutomationFailureRef = useRef({});
  const failureToastReadyRef = useRef(false);
  const failureSummaryRef = useRef({ signature: '', at: 0 });
  const [automationBusy, setAutomationBusy] = useState('');
  const automationBusyAtRef = useRef(0);
  const [automationStartOpen, setAutomationStartOpen] = useState(false);
  const automationAutoOpenedRef = useRef(false);
  const [automationPresets, setAutomationPresets] = useState([]);
  const [automationPresetID, setAutomationPresetID] = useState('');
  const [automationRunMode, setAutomationRunMode] = useState('full_submit');
  const [automationForceRegenerate, setAutomationForceRegenerate] = useState(false);
  const [automationScheduleMode, setAutomationScheduleMode] = useState('execute');
  const [automationBookScope, setAutomationBookScope] = useState('all');
  const [automationScheduledAt, setAutomationScheduledAt] = useState('');
  const [automationConcurrency, setAutomationConcurrency] = useState(4);
  // 执行前质检面板：只在点击「立即执行」时拉一次只读预检，不参与任何轮询。
  const [automationPreflightOpen, setAutomationPreflightOpen] = useState(false);
  const [automationPreflightLoading, setAutomationPreflightLoading] = useState(false);
  const [automationPreflightError, setAutomationPreflightError] = useState('');
  const [automationPreflightResult, setAutomationPreflightResult] = useState(null);
  const [automationPreflightPayload, setAutomationPreflightPayload] = useState(null);
  const [contentRangeLines, setContentRangeLines] = useState(5);
  const [comicAnalysisCharacters, setComicAnalysisCharacters] = useState(2000);
  const [contentCaptureCharacters, setContentCaptureCharacters] = useState(4000);
  const [contentSettingsSaving, setContentSettingsSaving] = useState(false);
  const [productionStatus, setProductionStatus] = useState(null);
	const [activeProductionRequestID, setActiveProductionRequestID] = useState('');
  const [mergeStatus, setMergeStatus] = useState(null);
  const [mergeSpeed, setMergeSpeed] = useState(1);
  const [capabilities, setCapabilities] = useState({});
  const [configVersions, setConfigVersions] = useState([]);
  const [configVersionsError, setConfigVersionsError] = useState(null);
  const [actionBusy, setActionBusy] = useState('');
  // 出图任务按书 + 页记录忙碌状态，避免 A 书出图时把 B 书的按钮一起变灰。
  const [comicTaskKeys, setComicTaskKeys] = useState(() => new Set());
  const [comicRuns, setComicRuns] = useState({});
  const comicRunControllersRef = useRef(new Map());
  const comicGenerationInFlightRef = useRef(new Set());
  const [comicPageRefreshNonce, setComicPageRefreshNonce] = useState(0);
  // 书级/页级重试提交后，服务端要过几秒才把页面标记为 generating。
  // 这段时间里前端主动跟随刷新，避免用户点了重试却看不到任何执行迹象。
  const [comicFollowUntil, setComicFollowUntil] = useState(0);
  const [uploadMode, setUploadMode] = useState('');
  const [comicBatchPublishOpen, setComicBatchPublishOpen] = useState(false);
  const [selectedBookIds, setSelectedBookIds] = useState([]);
  const [platformNames, setPlatformNames] = useState({});
  const [platformOptions, setPlatformOptions] = useState([]);
  const [columnWidths, setColumnWidths] = useState(null);
  const resizeActiveRef = useRef(false);
  const resizeStateRef = useRef(null);
  const runtimeStatusRequestRef = useRef(null);
  // 在途状态请求只属于发起它的批次；切批次后旧请求既不能复用，也不能回写。
  const runtimeStatusBatchRef = useRef('');
  // 状态接口连续失败次数：503 时据此退避轮询（5s → 15s → 30s 封顶）。
  const runtimeStatusFailureRef = useRef(0);
  const localExecutorTaskStatusesRef = useRef(new Map());
  const localExecutorTasksInitializedRef = useRef(false);
  const autoClassificationInFlightRef = useRef(new Set());
  const autoClassificationAttemptedRef = useRef(new Set());
  const books = Array.isArray(batch?.books) ? batch.books : [];
  const batchProductionMode = batch?.settingsState?.patch?.productionMode === 'comic' ? 'comic' : 'video';
  // 新失败的那一刻只弹一次顶部 toast（4.5 秒），不占版面；已有失败在页面打开时
  // 不重复弹。表单里的持久记录由下面的行内 28px 细线承担。
  useEffect(() => {
    // 初始值是 idle 空对象；必须等书目状态实际返回后，才建立“首屏不提示”的基线。
    // 否则下一次把历史失败载入时会被错当作刚发生的新失败。
    if (!failureToastReadyRef.current && !automationFailureNotificationsReady(automationStatus)) return;
    const notified = notifiedAutomationFailureRef.current;
    const failedEntries = collectNewAutomationFailures({
      books,
      automationStatus,
      batchId: batch?.id,
      notified,
      firstPass: !failureToastReadyRef.current,
      resolveEntry: automationBookForCard
    });
    failureToastReadyRef.current = true;
    if (!failedEntries.length) return;
    const first = failedEntries[0];
    const content = automationFailureSummaryContent(failedEntries);
    // 同一批失败只保留一条汇总提示：N 本书失败 = 1 条，不再逐本堆叠遮挡。
    const summarySignature = `${failedEntries.length}:${failedEntries.map(item => item.signature).join('|')}`;
    const now = Date.now();
    if (failureSummaryRef.current.signature === summarySignature && now - failureSummaryRef.current.at < 15000) return;
    failureSummaryRef.current = { signature: summarySignature, at: now };
    message.error({
      key: `batch-factory-failure-summary-${batch?.id || 'batch'}`,
      duration: 3,
      content: (
        <span>
          {content}
          <Button
            type="link"
            size="small"
            style={{ marginLeft: 4, padding: 0, height: 'auto', fontSize: 12 }}
            onClick={() => setLogsOpen(true)}
          >
            查看详情
          </Button>
        </span>
      )
    });
  }, [automationStatus, batch?.id, books]);
  const editingContentProductionMode = editingContentBook && effectiveBookSettings(batch, editingContentBook).productionMode === 'comic' ? 'comic' : 'video';
  useEffect(() => {
    if (!batch?.id || !promptBook?.id || actionBusy !== stageActionKey('comic', promptBook.id)) return undefined;
    let active = true;
    let polling = false;
    let lastComicRunStamp = '';
    const pollComicPromptBook = async () => {
      if (!active || polling) return;
      if (typeof document !== 'undefined' && document.hidden) return;
      polling = true;
      try {
        // 整批 getBatch 单次约 6MB。生成期间每 1.5 秒拉一次会把主线程和服务器
        // 都拖死；先用轻量 runtime-summary 判断该书的 comic run 是否变化，
        // 只有状态/时间戳变化时才拉整批数据。
        const summary = await getBatchRuntimeSummary(batch.id, { silent: true, suppressGlobalError: true });
        const comicRun = (summary?.stageSummaries?.[promptBook.id]?.runs || []).find(run => run.stage === 'comic');
        const stamp = comicRun ? `${comicRun.status}|${comicRun.updatedAt || comicRun.createdAt || ''}` : '';
        if (!stamp || stamp === lastComicRunStamp) return;
        lastComicRunStamp = stamp;
        const response = await getBatch(batch.id, { silent: true, suppressGlobalError: true });
        const latestBatch = response?.batch || response;
        const latestBook = (latestBatch?.books || []).find(item => item?.id === promptBook.id);
        if (active && latestBook) {
          setPromptBook(current => current?.revision === latestBook.revision ? current : latestBook);
        }
      } catch {
        // The stage request remains the source of truth; a transient poll failure
        // must not interrupt the running per-page generation.
      } finally {
        polling = false;
      }
    };
    pollComicPromptBook();
    const timer = setInterval(pollComicPromptBook, 4000);
    return () => {
      active = false;
      clearInterval(timer);
    };
  }, [actionBusy, batch?.id, promptBook?.id, onBatchChanged]);
  useEffect(() => {
    let active = true;
    getCurrentAccount()
      .then(account => { if (active) setIsDeveloper(account?.role === 'dev'); })
      .catch(() => { if (active) setIsDeveloper(false); });
    return () => { active = false; };
  }, []);
  useEffect(() => {
    if (!mediaBook) return;
    const refreshed = books.find(book => book.id === mediaBook.id);
    if (refreshed && refreshed.revision !== mediaBook.revision) setMediaBook(refreshed);
  }, [books, mediaBook]);
	useEffect(() => {
		if (!pendingMediaPrompt || mediaBook) return;
		const nextBook = books.find(book => book.id === pendingMediaPrompt.bookId);
		if (nextBook) {
			setPromptVideoId(pendingMediaPrompt.frameKey || '');
			setPromptBook(nextBook);
		}
		setPendingMediaPrompt(null);
	}, [books, mediaBook, pendingMediaPrompt]);
	const batchProgress = batchFactoryBatchProgress(books, { productionStatus, mergeStatus, stageSummaries, productionMode: batchProductionMode });
	const progressSegments = [
		{ key: 'uploaded', label: '上传成功', count: batchProgress.uploaded },
		{ key: 'failed', label: '异常', count: batchProgress.failed },
		{ key: 'running', label: '执行中', count: batchProgress.running },
		{ key: 'awaiting-upload', label: '待上传', count: batchProgress.awaitingUpload },
		{ key: 'pending', label: '待开始', count: batchProgress.pending }
	].filter(segment => segment.count > 0).map(segment => ({ ...segment, percent: batchProgress.total ? (segment.count / batchProgress.total) * 100 : 0 }));
  const mediaCounts = runtimeBatchMediaCounts(books, productionStatus, mergeStatus);
  const runCapability = capability(capabilities, 'director.run');
  const workingFrontCapability = capability(capabilities, 'working-front.viral');
  const compilerCapability = capability(capabilities, 'compiler.preview');
  const productionCapability = capability(capabilities, 'production.submit');
	const cancelCapability = capability(capabilities, 'production.cancel');
  const mergeCapability = capability(capabilities, 'merge.run');
  const automationState = String(automationStatus?.state || 'idle');
  const automationCounts = automationStatus?.counts || { total: 0, ready: 0, running: 0, pending: 0, failed: 0, blocked: 0 };
  const automationActive = ['running', 'paused', 'scheduled', 'needs_attention'].includes(automationState);
  const automationStartBlockedLabel = '已有任务待处理，请先在任务/日志中处理当前任务';
  const automationStartButtonLabel = automationState === 'scheduled' ? '已有预约' : automationState === 'running' ? '运行中' : '待处理';

  function localExecutorToastKey() { return `batch-factory-local-executor-${batch?.id || 'current'}`; }
  function showLocalExecutorSubmittingToast() {
    message.loading({ key: localExecutorToastKey(), content: '正在提交给一战晟铭执行器…', duration: 0 });
  }
  function clearLocalExecutorSubmittingToast() {
    message.destroy(localExecutorToastKey());
  }

  useEffect(() => {
    localExecutorTaskStatusesRef.current = new Map();
    localExecutorTasksInitializedRef.current = false;
  }, [batch?.id]);

  useEffect(() => {
    if (!productionStatus) return;
    const tasks = (productionStatus?.jobs || []).flatMap(job => job?.tasks || [])
      .filter(task => task?.provider === YIZHANSHENGMING_EXECUTOR_PROVIDER);
    const current = new Map(tasks.map(task => [String(task.id || `${task.bookId || ''}:${task.videoId || ''}`), String(task.status || '').toLowerCase()]));
    if (!localExecutorTasksInitializedRef.current) {
      localExecutorTaskStatusesRef.current = current;
      localExecutorTasksInitializedRef.current = true;
      return;
    }
    tasks.forEach(task => {
      const taskID = String(task.id || `${task.bookId || ''}:${task.videoId || ''}`);
      const status = String(task.status || '').toLowerCase();
      if (localExecutorTaskStatusesRef.current.get(taskID) === status) return;
      const notice = localExecutorTaskToast(task);
      if (!notice) return;
      const options = { key: localExecutorToastKey(), content: notice.text };
      if (notice.type === 'loading') message.loading({ ...options, duration: 0 });
      else message[notice.type](options);
    });
    localExecutorTaskStatusesRef.current = current;
  }, [productionStatus, batch?.id]);

  async function refreshBatch() {
    return await onBatchChanged?.();
  }

  function markComicTask(key, busy) {
    setComicTaskKeys(current => {
      const next = new Set(current);
      if (busy) next.add(key);
      else next.delete(key);
      return next;
    });
  }

  function isComicTaskBusy(key) {
    return comicTaskKeys.has(key);
  }
  function updateComicRun(bookId, patch) {
    setComicRuns(current => {
      const previous = current[bookId];
      if (!previous && !patch) return current;
      return { ...current, [bookId]: { ...(previous || {}), ...patch } };
    });
  }
  function stopComicRun(bookId) {
    const controller = comicRunControllersRef.current.get(bookId);
    if (!controller) return;
    controller.abort();
    updateComicRun(bookId, { running: false, stopped: true });
    message.info('正在停止整本漫画出图…（已完成页会保留，未完成页可重试）');
  }


  async function openBookInAgent(book) {
    if (!batch?.id || !book?.id || actionBusy) return;
    setActionBusy(`agent-${book.id}`);
    try {
      await bindBatchFactoryBookAndNavigate({
        batchId: batch.id,
        bookId: book.id,
        bind: bindBatchFactoryAgentProject,
        navigate: url => window.location.assign(url)
      });
    } catch (error) {
      message.error(error?.message || '打开 Agent 画布失败');
      setActionBusy('');
    }
  }

  async function openBatchInAgent() {
    if (!batch?.id || !books.length || actionBusy) return;
    setActionBusy('agent-all');
    try {
      const binding = await bindBatchFactoryAgentProject(batch.id);
      const projectId = String(binding?.projectId || '').trim();
      if (!projectId) throw new Error('Agent 项目绑定失败');
      window.location.assign(agentCanvasURL({ projectId, batchId: batch.id }));
    } catch (error) {
      message.error(error?.message || '打开 Agent 画布失败');
      setActionBusy('');
    }
  }

  async function openSelectedInAgent() {
    if (selectedBookIds.length !== 1) return;
    const selectedBook = books.find(book => String(book.id) === String(selectedBookIds[0]));
    if (selectedBook) await openBookInAgent(selectedBook);
  }

  async function addBooksToBatch(items, platform, allowDuplicate) {
    if (!batch?.id) throw new Error('当前批次尚未加载完成，请刷新后重试');
    if (!items?.length) throw new Error('没有可添加的书籍');
    const option = platformOptions.find(item => String(item.value) === String(platform));
    // 小说列表可以在没有打开“自动生产”弹窗的情况下直接新增书。此时页面状态仍是
    // 5 行默认值，不能拿它覆盖已经运行中的批次配置；新增书应继承本批次现有书的
    // 正文范围与抓取长度，空批次才使用当前页面选择。
    const inheritedContent = inheritedContentSettingsForNewBook(books, contentRangeLines, contentCaptureCharacters);
    const booksInput = items.map(item => ({
      id: item.bookId,
      bookId: item.bookId,
      title: item.title,
      platform: String(platform),
      sourceMetadata: { platformId: String(platform), platformName: option?.label || String(platform), sourceMode: 'manual_original', contentRangeLines: inheritedContent.contentRangeLines, contentCaptureCharacters: inheritedContent.contentCaptureCharacters, queueStatus: 'manual_pending' }
    }));
    const beforeBookIDs = new Set(books.map(book => String(book?.id || '')));
    const intakeResult = await createNovelFetchIntake({ books: booksInput, metadata: { sourceMode: 'manual_original', platformId: String(platform), platformName: option?.label || String(platform), contentRangeLines: inheritedContent.contentRangeLines, contentCaptureCharacters: inheritedContent.contentCaptureCharacters } });
    const intake = resultData(intakeResult, 'intake');
    if (!intake?.id) throw new Error('新增书籍暂存失败');
    const appendedResult = await appendNovelFetchIntake(batch.id, intake.id, { allowDuplicate: allowDuplicate === true });
    const appendedBatch = resultData(appendedResult, 'batch') || resultData(await getBatch(batch.id), 'batch');
    const newlyAddedBooks = (appendedBatch?.books || []).filter(book => !beforeBookIDs.has(String(book?.id || '')) && booksInput.some(item => String(item.bookId) === String(book?.bookId)));
    const fetchFailures = [];
    const fetchedBookIDs = [];
    const fetchOneBook = async book => {
      try {
        const fetched = await fetchBookOriginal(batch.id, book.id, { maxTxt: inheritedContent.contentCaptureCharacters, platform: String(platform) });
        const metadata = resultData(fetched, 'sourceMetadata') || resultData(fetched?.book, 'sourceMetadata');
        const styleFailed = String(metadata?.styleStatus || '').trim().toLowerCase() === 'failed';
        const genderMissing = !String(metadata?.gender || '').trim();
        if (styleFailed || genderMissing) {
          return { book, failure: `${book.title || book.bookId}: ${styleFailed ? (metadata?.styleError || '风格识别失败') : '男女频识别失败'}` };
        }
        return { book, id: String(book.id) };
      } catch (error) {
        return { book, failure: `${book.title || book.bookId}: ${error?.message || '获取失败'}` };
      }
    };
    for (let offset = 0; offset < newlyAddedBooks.length; offset += 4) {
      const results = await Promise.all(newlyAddedBooks.slice(offset, offset + 4).map(fetchOneBook));
      results.forEach(result => {
        if (result.failure) fetchFailures.push(result.failure);
        else if (result.id) fetchedBookIDs.push(result.id);
      });
    }
    await refreshBatch();
    // 已启动自动生产的批次不能让新增书停在“正文已获取”。沿用该批次已保存的
    // 预设，只派发新增书，避免已经完成或执行中的旧书被重复提交。
    let automationStarted = 0;
    if (fetchedBookIDs.length) {
      try {
        const statusResult = await getBatchAutomationStatus(batch.id);
        const latestAutomation = resultData(statusResult, 'automation');
        const latestState = String(latestAutomation?.state || '').toLowerCase();
        const presetID = String(latestAutomation?.preset?.id || '').trim();
        if (presetID && ['running', 'completed'].includes(latestState)) {
          const started = await startBatchAutomation(batch.id, {
            presetId: presetID,
            runMode: String(latestAutomation?.runMode || 'full_submit'),
            autoPublish: latestAutomation?.autoPublish === true,
            concurrency: normalizeAutomationConcurrency(latestAutomation?.concurrency),
            bookIds: fetchedBookIDs
          });
          setAutomationStatus(resultData(started, 'automation') || latestAutomation);
          automationStarted = fetchedBookIDs.length;
        }
      } catch (error) {
        message.warning(`新书已加入，但没有接入自动生产：${error?.message || '可在“自动生产”中重试'}`);
      }
    }
    if (fetchFailures.length) {
      message.warning(`已添加 ${newlyAddedBooks.length} 本书；${fetchFailures.length} 本原文或分类尚未完成，可在书籍详情中查看状态并重试。`);
    } else {
      message.success(automationStarted ? `已添加 ${newlyAddedBooks.length} 本书，并已接入自动生产` : `已添加并获取 ${newlyAddedBooks.length} 本书的正文与分类信息`);
    }
  }
  async function deleteBook(book) {
    if (!batch?.id || !book?.id || actionBusy) return;
    setActionBusy(`delete-${book.id}`);
    try {
      await deleteBatchFactoryBook(batch.id, book.id);
      setViewingBook(null);
      setSelectedBookIds(current => current.filter(id => id !== book.id));
      await refreshBatch();
      message.success('已删除当前小说及其本地生产记录；121 内容未受影响');
    } catch (error) { message.error(error?.message || '删除小说失败'); }
    finally { setActionBusy(''); }
  }
  async function fetchMissingBookSource(book) {
    if (!batch?.id || !book?.id || actionBusy || batchFactorySourceText(book?.sourceText)) return;
    setActionBusy(`source-${book.id}`);
    try {
      await fetchBookOriginal(batch.id, book.id);
      await refreshBatch();
      message.success('已获取并写入当前小说正文');
    } catch (error) {
      message.error(error?.message || '获取正文失败');
    } finally {
      setActionBusy('');
    }
  }
  async function ensureBookAudioDuration(book, { force = false, quiet = false, batchSnapshot = batch } = {}) {
    const currentBatch = batchSnapshot || batch;
    const settings = effectiveBookSettings(currentBatch, book);
    if (!audioPlanningRequired(settings)) return Number(settings.audioDurationSeconds || 0);
    const input = batchFactoryBookAudioInput(book);
    const tts = await batchFactoryBookTts(settings);
    const fingerprint = batchFactoryAudioFingerprint(input, tts);
    const currentSeconds = Number(settings.audioDurationSeconds || 0);
    if (!force && currentSeconds > 0 && (settings.audioDurationManual === true || String(settings.audioDurationFingerprint || '') === fingerprint)) return currentSeconds;
    if (!quiet) message.info(`正在为《${book.title || book.bookId || '当前小说'}》读取真实配音时长…`);
    try {
      const measured = await generateBatchFactoryBookAudioMeasurement(book, settings);
      await saveBookOverrideWithRetry(currentBatch.id, book.id, Number(book.revision || 0), { patch: { audioDurationSeconds: measured.durationSeconds, audioDurationFingerprint: measured.fingerprint, audioLineDurations: measured.lineDurations, audioDurationManual: false } });
      return measured.durationSeconds;
    } catch (error) {
      if (currentSeconds > 0) {
        if (!quiet) message.warning(`配音时长重新读取失败，继续使用已保存的配音时长 ${currentSeconds.toFixed(2)} 秒。`);
        return currentSeconds;
      }
      throw error;
    }
  }

  async function prepareAudioPlanningForBatch(currentBatch) {
    const candidates = (currentBatch?.books || []).filter(book => {
      const settings = effectiveBookSettings(currentBatch, book);
      return audioPlanningRequired(settings);
    });
    if (!candidates.length) return;
    message.info(`正在逐本生成配音并读取 ${candidates.length} 本小说的真实时长…`);
    const failed = [];
    for (const book of candidates) {
      try {
        await ensureBookAudioDuration(book, { force: true, quiet: true, batchSnapshot: currentBatch });
      } catch (error) {
        failed.push(`${book.title || book.bookId || '当前小说'}：${error?.message || '配音时长读取失败'}`);
      }
    }
    if (failed.length) {
      message.warning(`已完成 ${candidates.length - failed.length}/${candidates.length} 本配音时长读取；${failed.join('；')}`);
      return;
    }
    message.success(`已生成并读取 ${candidates.length} 本小说的真实配音时长。`);
  }

  async function syncUnifiedSettingsToBooks(currentBatch) {
    const unifiedPatch = currentBatch?.settingsState?.patch || {};
    // “保存统一配置”是批量改模型的唯一入口：把面板里实际有值的字段批量写进
    // 该批次每一本书（包括此前单书单独改过的），这样不会有书还跑在旧模型上。
    // 内容键一律不进推送集合：已生成的分镜/提示词/配音测量/图片视频产物只由
    // 生成流程写，统一配置保存不得触碰，否则会把已完成阶段变成“待重跑”。
    const pushKeys = UNIFIED_CONFIGURATION_KEYS
      .filter(key => !GENERATED_CONTENT_SETTINGS_KEYS.includes(key))
      .filter(key => Object.hasOwn(unifiedPatch, key));
    if (!pushKeys.length) return;
    const pushPatch = Object.fromEntries(pushKeys.map(key => [key, unifiedPatch[key]]));
    const candidates = currentBatch?.books || [];
    const results = await Promise.allSettled(candidates.map(book => saveBookOverrideWithRetry(
      currentBatch.id,
      book.id,
      Number(book.revision || 0),
      { patch: pushPatch }
    )));
    const failed = results.flatMap((result, index) => result.status === 'rejected'
      ? [`${candidates[index]?.title || candidates[index]?.bookId || '当前小说'}：${result.reason?.message || '同步失败'}`]
      : []);
    if (pushKeys.includes('publishSettings')) {
      // PUT 成功不代表书级 patch 真的落下；重新拉取一次，逐本断言
      // publishSettings 已存在，避免“提示成功但书里没有”。
      const latestBatch = resultData(await getBatch(currentBatch.id), 'batch');
      (latestBatch?.books || []).forEach(book => {
        if (!Object.hasOwn(book?.settingsState?.patch || {}, 'publishSettings')) {
          failed.push(`${book.title || book.bookId || book.id}：书级 publishSettings 未落库`);
        }
      });
    }
    if (failed.length) throw new Error(`统一配置已保存，但 ${failed.length} 本小说未完成覆盖：${failed.join('；')}`);
  }
  async function saveProductionMode(value) {
    const productionMode = value === 'comic' ? 'comic' : 'video';
    try {
      await saveBatchSettings(batch.id, {
        patch: { productionMode },
        expectedRevision: Number(batch?.settingsState?.revision || 0)
      }, { suppressGlobalError: true });
    } catch (error) {
      if (Number(error?.status) !== 409) throw error;
      const latestResult = await getBatch(batch.id);
      const latestBatch = resultData(latestResult, 'batch');
      await saveBatchSettings(batch.id, {
        patch: { productionMode },
        expectedRevision: Number(latestBatch?.settingsState?.revision || 0)
      }, { suppressGlobalError: true });
    }

    const savedResult = await getBatch(batch.id);
    const savedBatch = resultData(savedResult, 'batch');
    const booksWithModeOverride = (savedBatch?.books || []).filter(book => Object.hasOwn(book?.settingsState?.patch || {}, 'productionMode'));
    const overrideResults = await Promise.allSettled(booksWithModeOverride.map(book => saveBookOverrideWithRetry(
      savedBatch.id,
      book.id,
      Number(book.revision || 0),
      { patch: {}, restoreKeys: ['productionMode'] }
    )));
    await refreshBatch();
    const failedOverrides = overrideResults
      .map((result, index) => ({ result, book: booksWithModeOverride[index] }))
      .filter(item => item.result.status === 'rejected')
      .map(item => item.book?.title || item.book?.bookId || '当前小说');
    return { failedBookTitles: failedOverrides };
  }
	async function recompileUnifiedH3Prompts(currentBatch) {
		const candidates = (currentBatch?.books || []).filter(book => h3DirectorCards(book).length > 0);
		const result = { compiled: 0, manual: 0, waitingForAudio: 0, failed: [] };
		for (const book of candidates) {
			try {
				const outcome = await compileBookH3Videos(book, { interactive: false, allowAudioSynthesis: true });
				if (outcome?.skipped === 'manual') result.manual += 1;
				else if (outcome?.skipped === 'audio') result.waitingForAudio += 1;
				else if (outcome?.compiled) result.compiled += 1;
			} catch (error) {
				result.failed.push(`${book.title || book.bookId || '当前小说'}：${error?.message || '重新编译失败'}`);
			}
		}
		return result;
	}
	async function compileBookH3Videos(book, { interactive = true, allowAudioSynthesis = true } = {}) {
		const latestResponse = await getBatch(batch.id);
		const latestBatch = latestResponse?.batch || latestResponse;
		const latestBook = (latestBatch?.books || []).find(item => item?.id === book.id);
		const director = latestBook?.directorRevision;
		const h3Document = director?.output?.h3_director || director?.output?.h3Director;
		if (!director?.id || !h3Document) return;
		const settings = effectiveBookSettings(latestBatch, latestBook);
		let priorTrace = null;
		try { priorTrace = await getH3Trace(latestBatch.id, latestBook.id); }
		catch (error) { if (Number(error?.status) !== 404 && !String(error?.message || '').includes('404')) throw error; }
		const hasManualPrompts = (priorTrace?.compilation?.compilation?.segments || []).some(segment => segment.compile_trace?.editable_copy_source === 'user_final_prompt');
		if (hasManualPrompts) {
			if (!interactive) return { skipped: 'manual' };
			const confirmed = await new Promise(resolve => Modal.confirm({ title: '重新编译会替换手动编辑的分镜提示词', content: '历史提交记录保留。是否按当前配置重新生成完整提示词？', okText: '确认重新编译', cancelText: '保留手动提示词', onOk: () => resolve(true), onCancel: () => resolve(false) }));
			if (!confirmed) throw new Error('已保留手动提示词，当前配置尚未应用到最终分镜');
		}
		let audioResult = null;
    if (audioPlanningRequired(settings)) {
			const tts = await batchFactoryBookTts(settings);
			const previousMeasurement = priorTrace?.timeline?.timeline?.audio_measurement;
			if (!allowAudioSynthesis && !previousMeasurement?.asset_id) return { skipped: 'audio' };
			audioResult = await measureH3VideoLines({ directorId: director.id, document: h3Document, tts,
				previous: previousMeasurement,
				synthesize: textToSpeech, encode: blobToBase64,
				measure: payload => measureH3Audio(latestBatch.id, latestBook.id, payload) });
		}
		const promptConfig = settings?.aiPromptConfig || {};
		const videoPreset = promptConfig.video || {};
		const videoPresetBody = String(videoPreset.body || '');
		const videoPromptTemplate = finalVideoPromptTemplate(videoPresetBody);
		const constraints = promptConfig.constraints || {};
		const h3Constraints = buildBatchFactoryH3Constraints(constraints);
		const maxSegmentSeconds = Number(settings.storyboardDurationLimit) === 15 ? 15 : 10;
		await compileH3Video(latestBatch.id, latestBook.id, {
			director_revision_id: director.id,
			audio_asset_id: audioResult?.audio_asset_id || audioResult?.audioMeasurement?.measurement?.asset_id || audioResult?.audio_measurement?.measurement?.asset_id,
			allow_semantic_timeline: !audioPlanningRequired(settings),
			preset: {
				key: String(videoPreset.presetKey || videoPreset.presetId || videoPreset.id || 'h3-video-normal'),
				revision: Math.max(1, Number(videoPreset.presetVersion || videoPreset.version || 1)),
				format: 'h3-structured-v1',
				max_segment_ms: maxSegmentSeconds * 1000,
				request_duration_mode: 'ceil-second',
				prompt_template: videoPromptTemplate,
				output_constraints: videoPromptTemplate ? '按冻结导演数据和所选最终模板生成当前 VIDEO 提示词。' : (videoPresetBody || '按结构化时间线输出当前 VIDEO，保持人物、动作、机位和场景连续性。')
			},
			...h3Constraints,
			editable_copy_overrides: {},
			expected_compilation_id: priorTrace?.compilation?.id || '',
			allow_replace_manual_prompts: hasManualPrompts
		});
		return { compiled: true };
	}
	async function refreshAfterBookSettingsSaved() {
		const shouldRecompileH3 = ['constraints', 'video', 'media'].includes(configTarget?.region)
			&& Boolean(configTarget?.book?.directorRevision?.output?.h3_director || configTarget?.book?.directorRevision?.output?.h3Director);
		if (shouldRecompileH3) {
			try {
				message.info('设置已保存，正在复用现有导演数据重新编译最终 VIDEO Prompt…');
				await runBookStageAction(configTarget.book, 'director', 'compile');
				message.success('设置已生效，最终 VIDEO Prompt 已重新编译。');
			} catch (error) {
				message.error(error?.message ? `设置已保存，但 H3 重编译失败：${error.message}` : '设置已保存，但 H3 重编译失败');
			}
		}
		await refreshBatch();
	}
  async function resetBookAudioMeasurement(book, expectedRevision = book?.revision) {
    return saveBookOverrideWithRetry(batch.id, book.id, Number(expectedRevision || 0), { patch: { audioDurationSeconds: 0, audioDurationFingerprint: '', audioDurationManual: false } });
  }
  async function loadRuntimeStatus({ quiet = false, runtimeCapabilities = capabilities } = {}) {
    if (!batch?.id) return;
    const requestBatchID = batch.id;
    // 切批次时丢弃上一个批次的在途请求，避免把旧批次的 automation.counts
    // 写进当前批次（表现为「自动就绪」等数字串批次）。
    if (runtimeStatusBatchRef.current !== requestBatchID) {
      runtimeStatusBatchRef.current = requestBatchID;
      runtimeStatusRequestRef.current = null;
    }
    // 自动化轮询、生产抽屉和首次加载可能在同一时刻触发；复用同一个请求，
    // 让浏览器也始终只有一条批量运行状态读取在飞行。
    if (runtimeStatusRequestRef.current) return runtimeStatusRequestRef.current;
    setLogsLoading(true);
    const request = (async () => {
      try {
        // 一个批量摘要替代“状态 + 合并 + 每本阶段”的浏览器 N+2 次请求。
        // 服务端合并读取并短缓存，页面只消费同一时刻的快照。
        const runtime = await getBatchRuntimeSummary(requestBatchID, { suppressGlobalError: true });
        // 请求返回时批次已经切走：丢弃结果，不写入当前批次的状态。
        if (runtimeStatusBatchRef.current !== requestBatchID) return null;
        setProductionStatus(runtime?.production || { jobs: [] });
        setMergeStatus(runtime?.merge || { jobs: [] });
        setStageSummaries(runtime?.stageSummaries || {});
        const nextAutomation = runtime?.automation || { state: 'idle', counts: { total: 0, ready: 0, running: 0, pending: 0, failed: 0, blocked: 0 } };
        // 只追加合并日志，不覆盖已显示内容；页面轮询是唯一数据来源，
        // 打开日志面板时不会重新等待一个大包。
        setAutomationStatus(current => mergeAutomationLogSnapshot(current, nextAutomation));
        setLogsError('');
        setLogsSyncedAt(new Date().toLocaleTimeString('zh-CN', { hour12: false }));
        runtimeStatusFailureRef.current = 0;
        return runtime;
      } catch (error) {
        if (runtimeStatusBatchRef.current !== requestBatchID) return null;
        const endpoint = String(error?.source || '').trim();
        const httpStatus = Number(error?.status);
        const status = Number.isInteger(httpStatus) ? `HTTP ${httpStatus}` : '';
        // 503 是“状态服务暂不可用”，不能显示成“暂无任务”，也不该被当成空数据。
        runtimeStatusFailureRef.current = httpStatus === 503 ? (runtimeStatusFailureRef.current || 0) + 1 : 0;
        const timeoutSeconds = Math.max(1, Math.round(Number(error?.timeoutMs || 0) / 1000));
        // 网关 502/504 返回的是 nginx 的 HTML 错误页，不能原样塞进面板。
        const rawMessage = String(error?.message || '').trim();
        const looksLikeHTML = /<html|<!doctype|<head>|<\/body>/i.test(rawMessage);
        const friendlyMessage = looksLikeHTML || httpStatus >= 500
          ? '服务暂时不可用（服务重启或网关异常），请稍后点重试'
          : rawMessage || '读取任务状态失败';
        const text = isApiTimeoutError(error)
          ? [`任务状态读取超时（${timeoutSeconds} 秒），已保留上次结果，可点击重试`, endpoint ? `状态接口：${endpoint}` : ''].filter(Boolean).join(' · ')
          : httpStatus === 503
            ? [ '状态服务暂不可用，请稍后重试', status, endpoint ? `状态接口：${endpoint}` : ''].filter(Boolean).join(' · ')
            : [friendlyMessage, status, endpoint ? `状态接口：${endpoint}` : ''].filter(Boolean).join(' · ');
        setLogsError(text);
        if (!quiet) message.error(text);
        return undefined;
      } finally {
        setLogsLoading(false);
      }
    })();
    runtimeStatusRequestRef.current = request;
    try { return await request; }
    finally {
      if (runtimeStatusRequestRef.current === request) runtimeStatusRequestRef.current = null;
    }
  }
  useEffect(() => {
    if (!batch?.id) return;
    let active = true;
    Promise.all([getCapabilities(), getConfigVersions()]).then(async ([nextCapabilities, nextVersions]) => {
      if (!active) return;
      setCapabilities(nextCapabilities || {});
      setConfigVersions(resultData(nextVersions, 'configVersions') || []);
      setConfigVersionsError(null);
      await loadRuntimeStatus({ quiet: true, runtimeCapabilities: nextCapabilities || {} });
    }).catch(error => { if (active) setConfigVersionsError(error); });
    return () => { active = false; };
  }, [batch?.id]);
  useEffect(() => {
    if (!batch?.id) return undefined;
    let active = true;
    let polling = false;
    let wasRunning = false;
    let automationState = 'idle';
    let timer = null;
    const clearTimer = () => { if (timer) { clearTimeout(timer); timer = null; } };
    const poll = async () => {
      if (polling) return;
      polling = true;
      try {
        const runtime = await loadRuntimeStatus({ quiet: true });
        if (!active) return;
        const next = runtime?.automation;
        automationState = next?.state || 'idle';
        // 轮询同样只追加合并日志，保证“先显示的内容不会被下一次轮询清空”。
        setAutomationStatus(current => mergeAutomationLogSnapshot(current, next || { state: 'idle', counts: { total: 0, ready: 0, running: 0, queued: 0, pending: 0, failed: 0, blocked: 0 } }));
        if (next?.state === 'running') {
          wasRunning = true;
          await onBatchChanged?.();
        } else if (wasRunning) {
          // 任务刚从“执行中”结束，补刷最后一次，拿到最终书籍状态，
          // 避免页面停在结束前的旧画面（如 1 成功 5 异常）。
          wasRunning = false;
          await onBatchChanged?.();
        }
      } catch (error) {
        if (active && Number(error?.status || 0) !== 404) console.error('[batch-factory-automation] status read failed', error);
      } finally { polling = false; }
    };
    const schedule = () => {
      if (!active) return;
      clearTimer();
      if (document.hidden) return;
      const failures = Number(runtimeStatusFailureRef.current || 0);
      // 运行中保持 8 秒刷新；停止或空闲时降为 60 秒，避免 20 本书的
      // 汇总状态在没有任务时仍持续给数据库施压。
      const delay = batchFactoryRuntimePollDelay(automationState, failures);
      timer = setTimeout(() => { void poll().then(schedule); }, delay);
    };
    void poll().then(schedule);
    // 页面切到后台（最小化/切标签）时暂停轮询，回前台立刻补刷一次再继续，
    // 避免多人多标签把小服务器刷爆。页面一打开就在后台则等回前台再启动。
    const onVisibilityChange = () => {
      if (document.hidden) {
        clearTimer();
      } else {
        void poll().then(schedule);
      }
    };
    document.addEventListener('visibilitychange', onVisibilityChange);
    return () => {
      active = false;
      clearTimer();
      document.removeEventListener('visibilitychange', onVisibilityChange);
    };
  }, [batch?.id]);
  useEffect(() => {
    if (!batch?.id || automationStatus?.state !== 'running') return undefined;
    // 漫画图片是逐页写回的，不能等整本合集完成后才刷新媒体库。
    // 这里只刷新页面索引，不重复提交生成请求；后台已完成的第 N 张
    // 会在几秒内出现在列表和媒体库里，当前未完成页也会继续显示进行中。
    const timer = setInterval(() => {
      // 页面在后台时不需要刷新：整表重取会让隐藏标签页持续占 CPU，
      // 切回来时下一次 tick 会自动补上最新数据。
      if (typeof document !== 'undefined' && document.hidden) return;
      setComicPageRefreshNonce(value => value + 1);
    }, 5000);
    return () => clearInterval(timer);
  }, [batch?.id, automationStatus?.state]);
  useEffect(() => {
    let active = true;
    getWorkshopPlatforms({ suppressGlobalError: true }).then(result => {
      if (!active) return;
      const options = batchFactoryPlatformOptions(result?.platforms);
      setPlatformOptions(options);
      setPlatformNames(Object.fromEntries(options.map(option => [String(option.value), option.label])));
    }).catch(() => { if (active) setPlatformNames({}); });
    return () => { active = false; };
  }, []);
  useEffect(() => { setSelectedBookIds(current => current.filter(id => books.some(book => book.id === id))); }, [batch?.id, books.length]);
  useEffect(() => {
    if (!assetBook) return;
    const refreshed = books.find(book => book.id === assetBook.id);
    if (refreshed && refreshed !== assetBook) setAssetBook(refreshed);
  }, [batch?.id, assetBook?.id, books]);
  useEffect(() => {
    if (!promptBook) return;
    const refreshed = books.find(book => book.id === promptBook.id);
    if (refreshed && refreshed !== promptBook) setPromptBook(refreshed);
  }, [batch?.id, promptBook?.id, books]);
  useEffect(() => {
    if (!configTarget?.book) return;
    const refreshed = books.find(book => book.id === configTarget.book.id);
    if (refreshed && refreshed !== configTarget.book) setConfigTarget(current => current ? { ...current, book: refreshed } : current);
  }, [batch?.id, configTarget?.book?.id, books]);
  useEffect(() => {
    if (!editingContentBook) return;
    const refreshed = books.find(book => book.id === editingContentBook.id);
    if (!refreshed || refreshed === editingContentBook) return;
    setEditingContentBook(refreshed);
    if (!String(editingContentValue || '').trim()) setEditingContentValue(runtimeResolveBookProductionText(refreshed));
  }, [batch?.id, books, editingContentBook?.id]);

  async function openContentEditor(book) {
    const currentBook = books.find(item => item?.id === book?.id) || book;
    const rangeLines = contentRangeLinesForBook(currentBook);
    const productionMode = effectiveBookSettings(batch, currentBook).productionMode === 'comic' ? 'comic' : 'video';
    setEditingContentBook(currentBook);
    setEditingContentTitle(String(currentBook?.title || ''));
    setEditingContentSourceBookId(String(currentBook?.bookId || ''));
    setEditingContentValue(runtimeResolveBookEditableContent(currentBook, productionMode));
    setEditingContentMode(batchFactoryContentMode(currentBook));
    setEditingContentRangeLines(rangeLines);
    setEditingContentAnalysisCharacters(comicAnalysisCharactersForBook(currentBook));
    setEditingContentCaptureCharacters(contentCaptureCharactersForBook(currentBook));
    setViralCandidate('');
    setDerivedOpeningOptions([]);
    setDerivedOpeningPresetId('');
    const [draftRequest, catalogRequest] = await Promise.allSettled([
      getDraft({ key: `working-front-candidate:${currentBook.id}`, kind: 'working-front-viral-candidate', scope: batch.id }, { suppressGlobalError: true }),
      listSystemPresetCatalog('batch-factory')
    ]);
    if (draftRequest.status === 'fulfilled') {
      setViralCandidate(String(resultData(draftRequest.value, 'draft')?.content || ''));
    } else if (draftRequest.reason?.status && draftRequest.reason.status !== 404) {
      message.error(draftRequest.reason?.message || '读取爆款候选失败');
    }
    if (catalogRequest.status !== 'fulfilled') {
      message.error(catalogRequest.reason?.message || '读取衍生开篇提示词失败');
      return;
    }
    const allowedSlots = new Set(['batch.hook-adaptation', 'batch.original-director', 'batch.viral-director']);
    const options = (Array.isArray(catalogRequest.value?.catalog) ? catalogRequest.value.catalog : [])
      .filter(item => allowedSlots.has(String(item?.slot || '')))
      .map(item => ({ value: item.id, label: `${item.name || item.id} · v${item.version || 1}` }));
    setDerivedOpeningOptions(options);
    const inheritedId = String(batch?.settingsState?.patch?.aiPromptConfig?.derivedOpening?.presetId || '');
    setDerivedOpeningPresetId(options.some(option => option.value === inheritedId) ? inheritedId : (options[0]?.value || ''));
  }

  function openSourceEditor(book) {
    const currentBook = books.find(item => item?.id === book?.id) || book;
    setSourceEditorBook(currentBook);
    setSourceEditorValue(batchFactoryCleanSourceText(currentBook.sourceText));
  }

  async function saveCompleteSource(book, nextSourceText) {
    if (!book || !batch?.id) return;
    const sourceText = sourceTextAfterViralAdoption(nextSourceText);
    if (!sourceText) { message.error('完整原文不能为空'); return; }
    setSourceSaving(true);
    try {
      const save = revision => updateBookSource(batch.id, book.id, { sourceText, replaceSource: true, expectedRevision: Number(revision || 0) });
      try {
        await save(book.revision);
      } catch (error) {
        if (Number(error?.status) !== 409) throw error;
        const latestResult = await getBatch(batch.id);
        const latestBatch = resultData(latestResult, 'batch');
        const latestBook = (latestBatch?.books || []).find(item => item?.id === book.id);
        if (!latestBook) throw error;
        await save(latestBook.revision);
      }
      // The accepted complete source is authoritative. Remove a previous
      // first-lines production draft so it cannot override this book's new
      // source during subsequent production or a rewrite-enabled upload.
      await saveDraft({ key: `working-front:${book.id}`, kind: 'working-front-content', scope: batch.id, content: '' });
      await refreshBatch();
      setViewingBook({ ...book, sourceText });
      setSourceEditorBook(null);
      setEditingContentBook(null);
      setViralCandidate('');
      message.success('完整原文已保存；下游内容将按新原文重新生成');
    } catch (error) {
      message.error(error?.message || '保存完整原文失败');
    } finally {
      setSourceSaving(false);
    }
  }

  async function saveSourceEditor() {
    return saveCompleteSource(sourceEditorBook, sourceEditorValue);
  }

  async function adoptViralCandidate() {
    return saveCompleteSource(editingContentBook || sourceEditorBook, viralCandidate);
  }
  async function saveWorkingContent() {
    if (!editingContentBook) return;
    const nextTitle = String(editingContentTitle || '').trim();
    if (!nextTitle) { message.error('书名不能为空'); return; }
    const nextSourceBookId = String(editingContentSourceBookId || '').trim();
    if (nextSourceBookId && !/^\d{3,20}$/.test(nextSourceBookId)) { message.error('书号只能填写 3-20 位数字'); return; }
    const sourceBookIdChanged = Boolean(nextSourceBookId) && nextSourceBookId !== String(editingContentBook?.bookId || '');
    setContentSaving(true);
    try {
      await saveDraft({ key: `working-front:${editingContentBook.id}`, kind: 'working-front-content', scope: batch.id, content: editingContentValue.trim() });
      const analysisCharacters = Math.min(100000, Math.max(100, Number(editingContentAnalysisCharacters) || comicAnalysisCharactersForBook(editingContentBook)));
      const captureCharacters = Math.min(100000, Math.max(100, Number(editingContentCaptureCharacters) || contentCaptureCharactersForBook(editingContentBook)));
      if (editingContentProductionMode === 'video') {
        const rangeLines = Math.min(500, Math.max(1, Number(editingContentRangeLines) || contentRangeLinesForBook(editingContentBook)));
        const metadataResult = await updateBookMetadata(batch.id, editingContentBook.id, { title: nextTitle, sourceBookId: nextSourceBookId || undefined, metadata: { ...(editingContentBook.sourceMetadata || {}), productionContentMode: editingContentMode, contentRangeLines: rangeLines, contentCaptureCharacters: captureCharacters }, expectedRevision: Number(editingContentBook.revision || 0) });
        const updatedBook = resultData(metadataResult, 'book') || metadataResult;
        await resetBookAudioMeasurement(editingContentBook, Number(updatedBook?.revision || editingContentBook.revision || 0));
        message.success(`当前小说的视频生产内容已保存（${rangeLines} 行）；原文不会被覆盖。${sourceBookIdChanged ? '书号已更新，可直接点击重试重新抓取。' : ''}`);
      } else {
        const metadataResult = await updateBookMetadata(batch.id, editingContentBook.id, { title: nextTitle, sourceBookId: nextSourceBookId || undefined, metadata: { ...(editingContentBook.sourceMetadata || {}), productionContentMode: editingContentMode, comicAnalysisCharacters: analysisCharacters, contentCaptureCharacters: captureCharacters }, expectedRevision: Number(editingContentBook.revision || 0) });
        const updatedBook = resultData(metadataResult, 'book') || metadataResult;
        message.success(`当前小说的漫画生产正文已保存；漫画分析 ${analysisCharacters} 字，内容截取 ${captureCharacters} 字。${sourceBookIdChanged ? '书号已更新，可直接点击重试重新抓取。' : ''}`);
      }
      await refreshBatch();
      setEditingContentBook(null);
    } catch (error) { message.error(error?.message || '保存生产内容失败'); } finally { setContentSaving(false); }
  }
  async function createViralCandidate() {
    const targetBook = editingContentBook || sourceEditorBook;
    const sourceText = editingContentBook ? editingContentValue : sourceEditorValue;
    if (!targetBook || rewritingFront) return;
    setRewritingFront(true);
    try {
      const result = await rewriteWorkingFront(batch.id, targetBook.id, sourceText, {
        textModelId: effectiveBookSettings(batch, targetBook).textModelId,
        promptPresetId: derivedOpeningPresetId
      });
      const candidate = String(resultData(result, 'candidate') || '').trim();
      setViralCandidate(candidate);
      message.success('爆款候选已生成；录用后会替换完整正文，工作区仍只显示所选行数。');
    } catch (error) { message.error(error?.message || '生成爆款候选失败'); } finally { setRewritingFront(false); }
  }
  async function cancelViralCandidate() {
    const targetBook = editingContentBook || sourceEditorBook;
    if (!targetBook) return;
    try {
      await saveDraft({ key: `working-front-candidate:${targetBook.id}`, kind: 'working-front-viral-candidate', scope: batch.id, content: '' });
      setViralCandidate('');
      message.info('已取消本次爆款候选，完整正文未改动。');
    } catch (error) { message.error(error?.message || '取消爆款候选失败'); }
  }
  async function refreshAssetPresetSnapshot(book) {
    const aiPromptConfig = book?.settingsState?.patch?.aiPromptConfig;
    if (!aiPromptConfig || typeof aiPromptConfig !== 'object') return;
    // Older local builds persisted only preset metadata. Re-save through the
    // current Node bridge so the selected published body is frozen before the
    // asset-only stage asks the text provider to extract anything.
    await saveBookOverride(batch.id, book.id, {
      patch: { aiPromptConfig },
      expectedRevision: Number(book.revision || 0)
    });
    await refreshBatch();
  }
  async function runBookStageAction(book, stage, mode = 'missing', videoId = '', textModelId = '') {
    if (!batch?.id || !book?.id || actionBusy) return;
    setActionBusy(stageActionKey(stage, book.id));
    try {
      const settings = effectiveBookSettings(batch, book);
      const h3VideoPrompt = usesH3VideoPrompt(settings);
      if (stage === 'director' && mode === 'compile' && h3VideoPrompt && h3DirectorCards(book).length) {
        await compileBookH3Videos(book);
        await Promise.all([refreshBatch(), loadRuntimeStatus({ quiet: true })]);
        message.success('已复用统一导演卡编译最终 VIDEO Prompt。');
        return;
      }
      const requestedMode = stage === 'director' && mode === 'compile' ? 'missing' : mode;
      if (stage === 'director' && audioPlanningRequired(settings)) {
        await ensureBookAudioDuration(book);
      }
      const videoProvider = stage === 'video'
        ? videoProviderForModel(settings.videoModelId, settings.videoProvider)
        : '';
      if (videoProvider === YIZHANSHENGMING_EXECUTOR_PROVIDER) showLocalExecutorSubmittingToast();
      const stageResult = await runBookStage(batch.id, book.id, stage, {
        mode: requestedMode,
        // The selected video prompt preset chooses the director contract:
        // H3 uses the structured H3 document, while SD uses the SD text
        // contract. The video model remains an execution-only setting.
        h3: h3VideoPrompt,
        videoId,
        textModelId: textModelId || settings.textModelId,
        ...(stage === 'video' ? {
          provider: videoProvider,
          videoModelId: settings.videoModelId
        } : {}),
        requestId: requestID(`bf11-${stage}-${mode}`)
      });
      if (stage === 'director' && h3VideoPrompt) await compileBookH3Videos(book);
      // The modal keeps its own book snapshot while it is open.  Refresh the
      // exact book immediately after a director run so a newly-created
      // director revision (and its rebuilt H3 cards) is visible without
      // closing/reopening the modal or relying on a parent render race.
      if (stage === 'director') {
        const latestResponse = await getBatch(batch.id);
        const latestBatch = latestResponse?.batch || latestResponse;
        const latestBook = (latestBatch?.books || []).find(item => item?.id === book.id);
        if (latestBook) setPromptBook(latestBook);
      }
      await Promise.all([refreshBatch(), loadRuntimeStatus({ quiet: true })]);
      const labels = { assets: settings.productionMode === 'comic' ? (mode === 'force' ? '已重新提取漫画资产；视频和手动资产未改动' : '已提取漫画资产') : (mode === 'force' ? '已重新生成资产；分镜和 VIDEO 未改动' : '已生成资产'), comic: mode === 'bootstrap' ? '已完成漫画原文分页和每页提示词' : '已重新生成漫画提示词', director: mode === 'force' ? '已重新生成文案' : '已生成文案', opening: '已提交换开头变体生成', image: mode === 'force' ? '已重新生成图片' : '已生成图片', video: mode === 'force' ? '已创建新的视频候选版本' : '已提交视频生成' };
      const baseLabel = labels[stage] || '当前小说阶段已提交。';
      if (stage === 'director' && stageResult?.openingVariants?.triggered) {
        const ov = stageResult.openingVariants;
        if (ov.succeeded) message.success(`${baseLabel}；已自动生成 ${ov.generated} 个换开头变体，可在分镜一点击“查看换开头提示词”。`);
        else message.warning(`${baseLabel}；但换开头变体未生成：${ov.reason || '可重试失败步骤'}`);
      } else {
        message.success(baseLabel);
      }
    } catch (error) {
      // A provider can reject a request synchronously (quota, moderation, or
      // invalid input).  Reload the persisted stage run as well as the book so
      // the error is visible in the card and task timeline immediately.
      // End the visible action before any best-effort readback.  A slow status
      // endpoint must never leave a rejected upstream request spinning.
      setActionBusy('');
      message.error(error?.message || '当前小说阶段执行失败');
      void Promise.all([refreshBatch(), loadRuntimeStatus({ quiet: true })]).catch(() => {});
    } finally { setActionBusy(''); }
  }
  // 重试失败的提示必须区分“状态服务不可用 / 冲突 / 网络”，否则用户只能看到
  // 一句没有上下文的失败 toast。状态接口 503 时任务并未提交，可以安全再点。
  function retryFailureMessage(error) {
    const status = Number(error?.status || 0);
    if (status === 503) return '状态服务暂不可用，请稍后重试（HTTP 503）；本次未提交重试。';
    if (status === 409) {
      // 409 既可能是版本冲突，也可能是后端的业务拒绝（例如没有失败阶段）。
      // 后端返回的 message 更具体，直接展示，避免统一覆盖成“正在执行”。
      const detail = String(error?.message || '').trim();
      if (detail && !/^conflict$/i.test(detail)) return `${detail}（HTTP 409），请刷新后再重试。`;
      return '该小说正在执行或状态已被更新（HTTP 409），请刷新后再重试。';
    }
    if (status === 404) return '未找到可重试的失败步骤（HTTP 404），请先刷新状态。';
    if (!status && /Failed to fetch|NetworkError|network request failed|无法连接/i.test(String(error?.message || ''))) {
      return `网络请求失败：${error?.message || '请检查网络后重试'}`;
    }
    return error?.message || `重试提交失败${status ? `（HTTP ${status}）` : ''}`;
  }
  async function retryLastFailedStage(book, videoId = '', textModelId = '') {
    if (!batch?.id || !book?.id || actionBusy) return;
    // A running server job is the single retry owner.  Letting this modal
    // create a second request races that job and, worse, makes a provider
    // failure look like the client simply found no retryable stage.
    const automationRetry = automationBookNeedsRetry(automationStatus, book.id);
    if (automationRetry) {
      setActionBusy(stageActionKey('retry', book.id));
      const retryToastKey = `batch-retry-${book.id}`;
      // 点击即反馈：POST 之前就把 loading toast 挂出来，避免“点了完全静默”。
      message.loading({ content: '正在提交重试…', key: retryToastKey, duration: 0 });
      try {
        const retryResponse = await retryBatchAutomation(batch.id, [book.id]);
        const retryBook = (retryResponse?.automation?.books || []).find(item => String(item?.bookId || '') === String(book.id || ''));
        const stageLabels = { assets: '漫画资产', comic_assets: '漫画资产图', comic_bootstrap: '分页与漫画提示词', comic_prompt_recovery: '缺失漫画提示词', comic_images: '漫画图片', comic_publish: '漫画合集发布' };
        const stageLabel = stageLabels[String(retryBook?.stage || automationRetry.stage || '').trim()] || '失败阶段';
        // 提交成功立刻回 toast，刷新回读放后台，避免慢接口把成功提示吞掉。
        message.success({ content: `已提交，正在重跑…（${stageLabel}阶段；任务日志会持续更新）`, key: retryToastKey });
        void Promise.allSettled([refreshBatch(), loadRuntimeStatus({ quiet: true })]);
      } catch (error) {
        message.error({ content: retryFailureMessage(error), key: retryToastKey });
        void Promise.allSettled([refreshBatch(), loadRuntimeStatus({ quiet: true })]);
      } finally { setActionBusy(''); }
      return;
    }
    if (automationBookIsActive(automationStatus, book.id)) {
      message.info('服务端自动生产正在处理该小说；会自动重试，无需手动操作。');
      return;
    }
    setActionBusy(stageActionKey('retry', book.id));
    const retryToastKey = `batch-retry-${book.id}`;
    message.loading({ content: '正在提交重试…', key: retryToastKey, duration: 0 });
    try {
      // The runtime summary is a display cache. Historical provider failures
      // can exist before it has indexed the corresponding stage row, so the
      // retry endpoint remains the source of truth for eligibility.
      const runtime = await loadRuntimeStatus({ quiet: true });
      const summary = runtime?.stageSummaries?.[book.id] || stageSummaries[book.id];
      const videoProgress = batchFactoryVideoProgress(book, runtime?.production || productionStatus);
      const effectiveVideoID = String(videoId || [...videoProgress.byVideo.entries()].find(([, item]) => item?.status === 'failed')?.[0] || '');
      const failedVideo = effectiveVideoID ? videoProgress.byVideo.get(effectiveVideoID) : null;
      // A video can fail only after its initial stage submission has succeeded.
      // In that case there is no failed stage-run row, but the persisted task is
      // still an explicit retry target for this storyboard.
      const retryAsVideo = failedVideo?.status === 'failed';
      const settings = effectiveBookSettings(batch, book);
      if (summary?.lastFailed?.stage === 'director' && audioPlanningRequired(settings)) await ensureBookAudioDuration(book);
      await retryBookStage(batch.id, book.id, {
        h3: usesH3VideoPrompt(settings),
        videoId: effectiveVideoID,
        textModelId: textModelId || settings.textModelId,
        ...(summary?.lastFailed?.stage === 'video' || retryAsVideo || !summary?.lastFailed ? {
          provider: videoProviderForModel(settings.videoModelId, settings.videoProvider),
          videoModelId: settings.videoModelId
        } : {}),
        requestId: requestID('bf11-book-retry')
      });
		if (summary?.lastFailed?.stage === 'director' && usesH3VideoPrompt(settings)) await compileBookH3Videos(book);
      message.success({ content: '已提交，正在重跑…', key: retryToastKey });
      setComicFollowUntil(Date.now() + 180000);
      setComicPageRefreshNonce(nonce => nonce + 1);
      // 状态回读放后台：慢/503 的状态接口不能再让重试看起来“没反应”。
      void Promise.allSettled([refreshBatch(), loadRuntimeStatus({ quiet: true })]);
    } catch (error) {
      message.error({ content: retryFailureMessage(error), key: retryToastKey });
      void Promise.allSettled([refreshBatch(), loadRuntimeStatus({ quiet: true })]);
    } finally { setActionBusy(''); }
  }
  async function runAi(scope, textModelId = '') {
    if (!batch?.id || actionBusy) return;
    if (scope?.id) { await runBookStageAction(scope, 'director', 'missing', '', textModelId); return; }
    setActionBusy('director');
    try {
      const audioBooks = books.filter(book => { const settings = effectiveBookSettings(batch, book); return audioPlanningRequired(settings); });
      if (audioBooks.length) message.info(`正在逐本读取 ${audioBooks.length} 本小说的真实配音时长…`);
      for (const book of audioBooks) await ensureBookAudioDuration(book, { quiet: true });
      await runBatchDirector(batch.id, { textModelId: textModelId || batch?.settingsState?.patch?.textModelId });
      await refreshBatch();
      message.success('批量 AI 推理已提交并回读当前作品。');
      setAiOpen(false);
    } catch (error) { message.error(error?.message || 'AI 推理失败'); } finally { setActionBusy(''); }
  }
  async function runProduction(book = null) {
    if (!batch?.id || actionBusy) return;
    setActionBusy('production');
    try {
      const operationRequestID = requestID(book ? 'bf11-book-production' : 'bf11-batch-production');
      const settings = book ? effectiveBookSettings(batch, book) : (batch?.settingsState?.patch || {});
      const provider = videoProviderForModel(settings.videoModelId, settings.videoProvider);
      if (provider === YIZHANSHENGMING_EXECUTOR_PROVIDER) showLocalExecutorSubmittingToast();
      if (book) await runBookStage(batch.id, book.id, 'video', { mode: 'missing', provider, videoModelId: settings.videoModelId, requestId: operationRequestID });
      else await submitBatchProduction(batch.id, operationRequestID, provider, settings.videoModelId);
      setActiveProductionRequestID(operationRequestID);
      await Promise.all([refreshBatch(), loadRuntimeStatus({ quiet: true })]);
      message.success(book ? '当前小说的视频任务已提交。' : '批量视频任务已提交。');
    } catch (error) { message.error(error?.message || '提交视频任务失败'); } finally { setActionBusy(''); }
  }
  async function generateBookComicPage(book, videoId, prompt = '') {
    if (!batch?.id || !book?.id || !videoId) return;
    const requestKey = `page:${book.id}:${videoId}`;
    if (comicGenerationInFlightRef.current.has(requestKey)) return;
	const settings = effectiveBookSettings(batch, book);
	const page = book.videos?.find(video => video.id === videoId);
	const pageIndex = (book.videos || []).findIndex(video => video.id === videoId);
	if (pageIndex >= COMIC_PAGE_GENERATION_LIMIT) { message.error(`生成漫画页失败：单本最多处理前 ${COMIC_PAGE_GENERATION_LIMIT} 页。请调整分页后重试。`); return; }
	const payload = comicPageGenerationPayload(settings, page?.comicPrompt || prompt, prompt || page?.comicImagePrompt || page?.comicPrompt);
    if (settings.productionMode !== 'comic') { message.error('生成漫画页失败：当前小说不是漫画生成模式。请先在统一配置中切换模式。'); return; }
    if (!payload.imagePrompt) { message.error('生成漫画页失败：当前分镜还没有图片提示词。请先生成分镜提示词。'); return; }
    if (!payload.modelId) { message.error('生成漫画页失败：还没有选择图片模型。请先在统一配置里选择图片模型。'); return; }
    comicGenerationInFlightRef.current.add(requestKey);
    const busyKey = `comic-image-${book.id}-${videoId}`;
    markComicTask(busyKey, true);
    try {
      message.loading({ key: busyKey, content: '正在生成当前漫画页…（1 页）', duration: 0 });
      const outcome = await generateComicPageWithConflict(batch.id, book.id, videoId, payload);
      if (outcome.conflict) {
        message.info({ key: busyKey, content: '有 1 页正在生成中，已启动轮询…（1 页）', duration: 0 });
      }
      await finishComicPageGeneration({
        waitForResult: () => waitForComicPageResult(batch.id, book.id, videoId),
        revealResult: () => {
          setComicPageRefreshNonce(value => value + 1);
          setMediaStartTab('comic');
          setMediaVideoId(videoId);
          setMediaBook(book);
          message.success({ key: busyKey, content: '已生成当前漫画页（1 页）；第一张已设为主图。' });
        },
        refreshBatch,
      });
    } catch (error) {
      message.error({ key: busyKey, content: comicPageErrorMessage(error, '生成漫画页') });
    } finally { comicGenerationInFlightRef.current.delete(requestKey); clearLocalExecutorSubmittingToast(); markComicTask(busyKey, false); }
  }

  async function downloadBookComicZip(book) {
    if (!batch?.id || !book?.id || actionBusy) return;
    const requestKey = `comic-zip-${book.id}`;
    setActionBusy(requestKey);
    try {
      const blob = await downloadComicPromptZip(batch.id, book.id);
      const objectURL = URL.createObjectURL(blob);
      const anchor = document.createElement('a');
      anchor.href = objectURL;
      anchor.download = `${String(book.title || '漫画').replace(/[\\/:*?"<>|]/g, '_')}_最终提示词_资产图片.zip`;
      document.body.appendChild(anchor);
      anchor.click();
      anchor.remove();
      window.setTimeout(() => URL.revokeObjectURL(objectURL), 1000);
      message.success('已开始下载最终完整出图提示词和资产图片 ZIP。');
    } catch (error) {
      message.error(error?.message || '下载漫画提示词和图片失败');
    } finally {
      setActionBusy(current => current === requestKey ? '' : current);
    }
  }
  async function generateBookComic(book) {
    if (!batch?.id || !book?.id) return;
    const requestKey = `book:${book.id}`;
    if (comicGenerationInFlightRef.current.has(requestKey)) return;
    const settings = effectiveBookSettings(batch, book);
    if (settings.productionMode !== 'comic') { message.error('生成整本漫画失败：当前小说不是漫画生成模式。请先在统一配置中切换模式。'); return; }
    if (!settings.imageModelId) { message.error('生成整本漫画失败：还没有选择图片模型。请先在统一配置里选择图片模型。'); return; }
    const allVideos = book.videos || [];
    const videos = allVideos.slice(0, COMIC_PAGE_GENERATION_LIMIT);
    if (!videos.length) { message.info('没有可生成的漫画页。请先生成分页与分镜提示词。'); return; }
    if (comicRunControllersRef.current.has(book.id)) { message.info('这本书的整本漫画任务还在运行，可直接停止或等待完成。'); return; }
    comicGenerationInFlightRef.current.add(requestKey);
    const busyKey = `comic-book-${book.id}`;
    const controller = new AbortController();
    comicRunControllersRef.current.set(book.id, controller);
    markComicTask(busyKey, true);
    updateComicRun(book.id, {
      running: true,
      stopped: false,
      failed: [],
      failedCount: 0,
      total: videos.length,
      completed: 0,
      bookTitle: book.title || '',
      truncated: allVideos.length > videos.length,
      skipped: Math.max(0, allVideos.length - videos.length)
    });
    let readyPageIds = new Set();
    let focusVideoId = '';
    let conflictCount = 0;
    try {
      const pageResponse = await listComicPages(batch.id, book.id, { silent: true, suppressGlobalError: true });
      const existingPages = resultData(pageResponse, 'pages') || [];
      readyPageIds = new Set(existingPages
        .filter(page => page?.images?.some(image => image?.isPrimary))
        .map(page => page.videoId)
        .filter(videoId => videos.some(video => video.id === videoId)));
      const targets = videos.filter(video => !readyPageIds.has(video.id));
      focusVideoId = targets[0]?.id || videos[0]?.id || '';
      updateComicRun(book.id, { completed: readyPageIds.size, pending: targets.length });
      message.loading({ key: busyKey, content: `正在生成整本漫画…（${videos.length} 页）`, duration: 0 });
      const result = await runComicPagePool(targets, {
        limit: COMIC_PAGE_CONCURRENCY,
        signal: controller.signal,
        worker: async video => {
          const payload = comicPageGenerationPayload(settings, video.comicPrompt || '', video.comicImagePrompt || video.comicPrompt || '');
          if (!payload.imagePrompt) throw new Error('缺少图片提示词，请先生成分镜提示词');
          const outcome = await generateComicPageWithConflict(batch.id, book.id, video.id, payload);
          if (outcome.conflict) {
            conflictCount += 1;
            message.info({ key: `comic-inflight-${book.id}`, content: `有 ${conflictCount} 页正在生成中，已启动轮询…（${videos.length} 页）`, duration: 3 });
          }
          // 409 表示该页已在后台生成中：同样进入轮询，等图片回填后继续下一批。
          await waitForComicPageResult(batch.id, book.id, video.id, { signal: controller.signal });
        },
        onProgress: progress => {
          updateComicRun(book.id, {
            completed: readyPageIds.size + progress.completed,
            failedCount: progress.failed,
            total: videos.length
          });
        }
      });
      const completedTotal = readyPageIds.size + result.completed.length;
      const failed = result.failed.map(({ item, error }) => ({
        label: item.label || item.id || '当前页',
        reason: comicPageErrorMessage(error, '生成漫画页')
      }));
      updateComicRun(book.id, { running: false, stopped: result.stopped, failed, failedCount: failed.length, completed: completedTotal, inFlight: false });
      await refreshBatch().catch(() => {});
      setComicPageRefreshNonce(value => value + 1);
      setMediaStartTab('comic');
      setMediaVideoId(focusVideoId);
      setMediaBook(book);
      if (result.stopped) {
        message.info({ key: busyKey, content: `已停止整本漫画出图（完成 ${completedTotal}/${videos.length} 页，未完成页可重试）。` });
        return;
      }
      if (failed.length) {
        message.error({ key: busyKey, content: `生成整本漫画失败：${failed.length} 页未完成（完成 ${completedTotal}/${videos.length} 页）。可展开失败清单后按页重试。` });
        return;
      }
      await createComicCollection(batch.id, book.id, { videoIds: videos.map(video => video.id) });
      await loadRuntimeStatus({ quiet: true });
      message.success({ key: busyKey, content: `已生成整本漫画（${videos.length} 页）${conflictCount ? `，其中 ${conflictCount} 页为在途生成回填` : ''}；已整理漫画合集。` });
    } catch (error) {
      updateComicRun(book.id, { running: false, failedCount: 0 });
      message.error({ key: busyKey, content: comicPageErrorMessage(error, '生成整本漫画') });
    } finally {
      comicGenerationInFlightRef.current.delete(requestKey);
      comicRunControllersRef.current.delete(book.id);
      clearLocalExecutorSubmittingToast();
      markComicTask(busyKey, false);
    }
  }
  async function runSelectedProduction() {
    if (!batch?.id || actionBusy) return;
    const selectedBooks = books.filter(book => selectedBookIds.includes(book.id));
    if (!selectedBooks.length) {
      message.info('请先在小说列表中勾选要立即执行的小说。');
      return;
    }
    if (batchProductionMode === 'comic') {
      try {
        const presets = automationPresets.length ? automationPresets : await loadAutomationPresets();
        const activePresetID = automationActive ? String(automationStatus?.preset?.id || '').trim() : '';
        const presetID = activePresetID || automationPresetID || presets[0]?.id || '';
        if (!presetID) {
          message.warning('请先在自动生产中选择一个漫画预设。');
          return;
        }
        setAutomationPresetID(presetID);
        setAutomationBookScope('selected');
        await runAutomationAction('start', { immediate: true, bookIds: selectedBookIds, presetId: presetID });
      } catch (error) {
        message.error(error?.message || '启动已选漫画任务失败');
      }
      return;
    }
    setActionBusy('selected-production');
    const operationRequestID = requestID('bf11-selected-production');
    const failedTitles = [];
    try {
      for (const book of selectedBooks) {
        const settings = effectiveBookSettings(batch, book);
        const provider = videoProviderForModel(settings.videoModelId, settings.videoProvider);
        try {
          if (provider === YIZHANSHENGMING_EXECUTOR_PROVIDER) showLocalExecutorSubmittingToast();
          await runBookStage(batch.id, book.id, 'video', {
            mode: 'missing',
            provider,
            videoModelId: settings.videoModelId,
            requestId: operationRequestID
          });
        } catch (error) {
          failedTitles.push(book.title || `小说 ${book.bookId || book.id}`);
        }
      }
      setActiveProductionRequestID(operationRequestID);
      await Promise.all([refreshBatch(), loadRuntimeStatus({ quiet: true })]);
      const completed = selectedBooks.length - failedTitles.length;
      if (failedTitles.length) {
        message.warning(`已提交 ${completed}/${selectedBooks.length} 本；未提交：${failedTitles.join('、')}`);
      } else {
        message.success(`已立即提交选中的 ${selectedBooks.length} 本小说视频任务。`);
      }
    } catch (error) { message.error(error?.message || '提交已选小说的视频任务失败'); } finally { setActionBusy(''); }
  }
  function mergePayload(prefix, book = null, speed = 1, overrides = {}) {
    const settings = normalizeAudioSourceSettings(book ? effectiveBookSettings(batch, book) : (batch?.settingsState?.patch || {}));
    const followAudio = normalizeAudioSource(settings.audioSource) === 'original' ? false : (Object.hasOwn(overrides, 'followAudio') ? overrides.followAudio === true : settings.audioMergeEnabled === true);
    const audioDurationSeconds = Object.hasOwn(overrides, 'audioDurationSeconds')
      ? Number(overrides.audioDurationSeconds || 0)
      : Number(settings.audioDurationSeconds || 0);
    return {
      requestId: requestID(prefix),
      timingMode: followAudio ? 'audio' : 'speed',
      audioSource: normalizeAudioSource(settings.audioSource),
      audioArtifactId: String(settings.ttsAudioArtifactId || '').trim(),
      speed: followAudio ? 0 : Number(overrides.speed ?? speed ?? 1),
      audioDurationSeconds: followAudio ? audioDurationSeconds : 0
    };
  }
  async function runMerge() {
    if (!batch?.id || actionBusy) return;
    setActionBusy('merge');
    try {
      await submitBatchMerge(batch.id, mergePayload('bf11-merge'));
      await loadRuntimeStatus({ quiet: true });
      message.success('批量合并任务已提交。');
    } catch (error) { message.error(error?.message || '提交批量合并失败'); } finally { setActionBusy(''); }
  }
  async function runBookMerge(book, overrides = {}) {
    if (!batch?.id || !book?.id || actionBusy) return;
    setActionBusy(`merge-${book.id}`);
    try {
      await submitBookMerge(batch.id, book.id, mergePayload('bf11-book-merge', book, Number(overrides.speed ?? mergeSpeed ?? 1), overrides));
      await loadRuntimeStatus({ quiet: true });
      message.success('当前小说最终合成任务已提交。');
    } catch (error) { message.error(error?.message || '提交当前小说合成失败'); } finally { setActionBusy(''); }
  }
	async function cancelProduction() {
		if (!batch?.id || actionBusy || !cancelCapability.available) return;
		if (!activeProductionRequestID) {
			message.info('请在本次页面中先发起一个批量视频操作；取消只会作用于该次操作。');
			return;
		}
		setActionBusy('cancel');
		try {
			const result = await cancelBatchProduction(batch.id, activeProductionRequestID);
			await loadRuntimeStatus({ quiet: true });
			const cancelled = resultData(result, 'cancelledTaskIds') || [];
			const skipped = resultData(result, 'skipped') || [];
			if (cancelled.length) message.success(`已取消 ${cancelled.length} 个本地执行器任务。`);
			else if (skipped.length) message.warning(`没有可取消的本地任务：${skipped.map(item => item.reason).filter(Boolean).join('；')}`);
			else message.info('当前没有运行中的视频任务。');
		} catch (error) { message.error(error?.message || '取消视频任务失败'); } finally { setActionBusy(''); }
	}
  async function loadAutomationPresets() {
    const result = await listAutomationPresets();
    const values = resultData(result, 'presets') || [];
    setAutomationPresets(values);
    return values;
  }
  // 行内「执行这本」：把范围锁定到这一本再打开启动弹窗，预检也只查这一本。
  function runSingleBook(book) {
    if (!book?.id) return;
    setSelectedBookIds([book.id]);
    void openAutomationStart([book.id]);
  }
  // explicitBookIds：行内「执行这本」直接指定单本，不依赖列表勾选状态
  // （用户点行不会勾选复选框，只传 selectedBookIds 会退回“全部小说”）。
  async function openAutomationStart(explicitBookIds = null) {
    try {
      const firstBook = books[0];
      setContentRangeLines(Number(firstBook?.sourceMetadata?.contentRangeLines || 5));
      setComicAnalysisCharacters(Number(firstBook?.sourceMetadata?.comicAnalysisCharacters || 2000));
      setContentCaptureCharacters(Number(firstBook?.sourceMetadata?.contentCaptureCharacters || 4000));
      const values = await loadAutomationPresets();
      setAutomationPresetID(current => current || values[0]?.id || '');
      setAutomationScheduledAt(current => current || formatBeijingDatetimeLocal(new Date(Date.now() + 10 * 60 * 1000).toISOString()));
      const scopeBookIds = Array.isArray(explicitBookIds) ? explicitBookIds : selectedBookIds;
      setAutomationBookScope(scopeBookIds.length ? 'selected' : 'all');
      setAutomationStartOpen(true);
    } catch (error) { message.error(error?.message || '读取自动化预设失败'); }
  }
  async function saveContentSettings(bookIds = []) {
    if (!batch?.id || !books.length || contentSettingsSaving) return false;
    const rangeLines = Math.min(500, Math.max(1, Number(contentRangeLines) || 5));
    const analysisCharacters = Math.min(100000, Math.max(100, Number(comicAnalysisCharacters) || 2000));
    const captureCharacters = Math.min(100000, Math.max(1000, Number(contentCaptureCharacters) || 4000));
    setContentSettingsSaving(true);
    try {
      const targetBooks = bookIds.length ? books.filter(book => bookIds.includes(book.id)) : books;
      const changedBooks = targetBooks.filter(book => {
        const metadata = book.sourceMetadata || {};
        return Number(metadata.comicAnalysisCharacters || 2000) !== analysisCharacters
          || Number(metadata.contentCaptureCharacters || 4000) !== captureCharacters
          || (batchProductionMode === 'video' && Number(metadata.contentRangeLines || 5) !== rangeLines);
      });
      await Promise.all(changedBooks.map(book => {
        const metadata = { ...(book.sourceMetadata || {}), comicAnalysisCharacters: analysisCharacters, contentCaptureCharacters: captureCharacters };
        if (batchProductionMode === 'video') metadata.contentRangeLines = rangeLines;
        return updateBookMetadata(batch.id, book.id, { metadata, expectedRevision: Number(book.revision || 0) });
      }));
      if (changedBooks.length) await refreshBatch();
      setContentRangeLines(rangeLines);
		setComicAnalysisCharacters(analysisCharacters);
      setContentCaptureCharacters(captureCharacters);
	  message.success(batchProductionMode === 'comic'
		? `内容设置已保存：漫画分析 ${analysisCharacters} 字，内容截取 ${captureCharacters} 字。`
		: `内容设置已保存：范围 ${rangeLines} 行，截取 ${captureCharacters} 字。`);
      return true;
    } catch (error) {
      message.error(error?.message || '内容范围和截取设置保存失败');
      return false;
    } finally { setContentSettingsSaving(false); }
  }
  useEffect(() => {
    if (openAutomationOnLoad && !automationAutoOpenedRef.current) {
      automationAutoOpenedRef.current = true;
      void openAutomationStart();
    }
    // The parent only sets this flag for a newly created novel-fetch batch.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [openAutomationOnLoad]);
  function automationPreflightReasons(targetBooks, runMode, productionMode) {
    const reasons = [];
    const selectedBooks = Array.isArray(targetBooks) && targetBooks.length ? targetBooks : books;
    const settings = selectedBooks.map(book => effectiveBookSettings(batch, book));
    if (!runCapability.available) reasons.push(`文本/资产执行器：${runCapability.reason || '未配置'}`);
    if (productionMode === 'comic' && !['storyboard_only'].includes(runMode) && settings.some(value => !String(value?.imageModelId || '').trim())) reasons.push('图片模型：至少有一本书未配置图片模型');
    if (productionMode !== 'comic' && !['storyboard_only'].includes(runMode) && settings.some(value => !String(value?.videoModelId || '').trim())) reasons.push('视频模型：至少有一本书未配置视频模型');
    if (settings.some(value => !String(value?.textModelId || '').trim())) reasons.push('文本模型：至少有一本书未配置文本模型');
    if (runMode === 'full_submit') {
      // 漫画走 comicPublishSettings，视频才走 publishSettings；统一只查后者会
      // 把已绑定组织的漫画批次误判为“发布配置未绑定”。
      const hasPublishProfile = settings.every(value => {
        const publish = (productionMode === 'comic' ? value?.comicPublishSettings : value?.publishSettings) || {};
        return String(publish.websiteProfileId || publish.versionProfile || publish.organization || publish.organizationId || '').trim() !== '';
      });
      if (!hasPublishProfile) reasons.push('发布配置：请先绑定网站配置档或组织归属');
      if (productionMode !== 'comic' && !productionCapability.available) reasons.push(`视频执行器：${productionCapability.reason || '未配置'}`);
    }
    return [...new Set(reasons)];
  }
  // 服务端只读预检：把“点击后才 blocked”的原因提前到点击时展示。
  async function openAutomationPreflight({ bookIds, presetId }) {
    if (!batch?.id) { message.error('页面批次尚未加载，请刷新后重试'); return; }
    setAutomationPreflightPayload({ bookIds, presetId });
    setAutomationPreflightResult(null);
    setAutomationPreflightError('');
    setAutomationPreflightLoading(true);
    setAutomationPreflightOpen(true);
    try {
      const result = await preflightBatchAutomation(batch.id, {
        presetId,
        runMode: automationRunMode,
        productionMode: batchProductionMode,
        bookIds,
        queueIfActive: automationActive && Array.isArray(bookIds) && bookIds.length > 0
      });
      setAutomationPreflightResult(resultData(result, 'preflight'));
    } catch (error) {
      setAutomationPreflightError(error?.message || '执行前预检失败');
    } finally {
      setAutomationPreflightLoading(false);
    }
  }
  function confirmPreflightRun(runnableBookIDs) {
    const payload = automationPreflightPayload;
    setAutomationPreflightOpen(false);
    void runAutomationAction('start', {
      immediate: true,
      bookIds: runnableBookIDs,
      presetId: payload?.presetId || '',
      skipPreflight: true
    });
  }
  function handleAutomationPreflightFix(fixTarget, bookId) {
    setAutomationPreflightOpen(false);
    // 水货生产模型页路由以 frontend/src/admin/App.jsx 的路由表为准（/admin/shuihuo-models）。
    if (fixTarget === 'admin_models') { window.location.assign('/admin/shuihuo-models'); return; }
    if (fixTarget === 'book_settings') {
      const targetBook = books.find(book => book.id === bookId);
      if (targetBook) { setConfigTarget({ book: targetBook }); return; }
    }
    if (fixTarget === 'preset') { void openAutomationStart(); return; }
    setUnifiedSettingsOpen(true);
  }
  function confirmImmediateAutomation() {
    const targetBookIDs = automationBookScope === 'selected' ? selectedBookIds : books.map(book => book.id);
    if (!automationPresetID) { message.warning('请先选择自动化预设'); return; }
    if (!targetBookIDs.length) { message.warning('请先勾选要执行的小说'); return; }
    void openAutomationPreflight({ bookIds: targetBookIDs, presetId: automationPresetID });
  }
  async function runAutomationAction(key, { immediate = false, bookIds = null, presetId = '', skipPreflight = false } = {}) {
    if (!batch?.id) { message.error('页面批次尚未加载，请刷新后重试'); return; }
    if (automationBusy) {
      // 请求悬挂时旧实现会静默吞掉所有点击。超过 30 秒视为僵死，复位放行。
      if (Date.now() - automationBusyAtRef.current > 30 * 1000) {
        automationBusyAtRef.current = 0;
        setAutomationBusy('');
        message.warning('上一次请求超时，已重置状态');
      } else {
        message.info('上一次操作仍在执行，请稍候');
        return;
      }
    }
    const queueBehindActive = key === 'start' && automationActive && Array.isArray(bookIds) && bookIds.length > 0;
    if (key === 'start' && automationActive && !queueBehindActive) {
      message.info(automationStartBlockedLabel);
      return;
    }
    const urlBatchId = typeof window !== 'undefined'
      ? new URLSearchParams(window.location.search).get('batchId') || ''
      : '';
    if (key === 'start' && urlBatchId && urlBatchId !== String(batch.id)) {
      message.error('页面批次与当前网址不一致，已阻止启动。请刷新页面后重试。');
      return;
    }
    const selectedPresetID = presetId || automationPresetID;
    const targetBookIDs = Array.isArray(bookIds)
      ? bookIds
      : automationBookScope === 'selected'
        ? selectedBookIds
        : [];
    if (key === 'start' && !selectedPresetID) { message.warning('请先选择自动化预设'); return; }
    if (key === 'start' && automationBookScope === 'selected' && !targetBookIDs.length) { message.warning('请先勾选要执行的小说'); return; }
    if (key === 'start' && immediate && !skipPreflight) {
      // 客户端能力是全局开关，先给一条明确提示；逐本配置问题交给服务端预检面板。
      if (!runCapability.available) { message.warning(`暂不能启动自动生产：文本/资产执行器：${runCapability.reason || '未配置'}`); return; }
      if (batchProductionMode !== 'comic' && automationRunMode !== 'storyboard_only' && !productionCapability.available) {
        message.warning(`暂不能启动自动生产：视频执行器：${productionCapability.reason || '未配置'}`);
        return;
      }
      // 立即执行先在服务端做一次只读预检，把阻断原因摆在点击当时；
      // 面板确认后带 skipPreflight=true 再进这里才真正提交。
      await openAutomationPreflight({
        bookIds: targetBookIDs.length ? targetBookIDs : books.map(book => book.id),
        presetId: selectedPresetID
      });
      return;
    }
    if (key === 'start') {
      // 定时执行没有点击面板，保留本地快检（不额外发请求）。
      const targetBooks = targetBookIDs.length ? books.filter(book => targetBookIDs.includes(book.id)) : books;
      const preflightReasons = automationPreflightReasons(targetBooks, automationRunMode, batchProductionMode);
      if (preflightReasons.length) {
        message.warning(`暂不能启动自动生产：${preflightReasons.join('；')}`);
        return;
      }
    }
    const scheduledAt = immediate ? '' : parseBeijingDatetimeLocal(automationScheduledAt);
    if (key === 'start' && !immediate && (!scheduledAt || new Date(scheduledAt).getTime() <= Date.now())) {
      message.warning('请选择晚于现在的北京时间自动启动时间');
      return;
    }
    setAutomationBusy(key);
    automationBusyAtRef.current = Date.now();
    const hideAutomationLoading = message.loading('正在提交自动化任务…', 0);
    try {
      if (key === 'start' && !(await saveContentSettings(targetBookIDs))) return;
      let result;
      if (key === 'start') result = await startBatchAutomation(batch.id, { presetId: selectedPresetID, runMode: automationRunMode, autoPublish: automationRunMode === 'full_submit', forceRegenerate: automationForceRegenerate, scheduleMode: immediate ? 'execute' : automationScheduleMode, scheduledAt: scheduledAt || undefined, concurrency: normalizeAutomationConcurrency(automationConcurrency), bookIds: targetBookIDs, productionMode: batchProductionMode, queueIfActive: queueBehindActive });
      else if (key === 'pause') result = await pauseBatchAutomation(batch.id);
      else if (key === 'resume') result = await resumeBatchAutomation(batch.id);
      else if (key === 'retry') result = await retryBatchAutomation(batch.id);
      else if (key === 'cancel') result = await cancelBatchAutomation(batch.id);
      const next = resultData(result, 'automation');
      setAutomationStatus(current => mergeAutomationLogSnapshot(current, next || current));
      if (key === 'start') { setAutomationStartOpen(false); message.success(queueBehindActive ? '已加入当前任务的后续队列，当前书完成后自动继续。' : (immediate ? '已立即启动自动化任务。' : (automationScheduleMode === 'submit' ? '已创建定时提交任务，生产完成后将在设定时间上传或发布。' : '已创建定时执行任务，将在设定时间开始生产。'))); }
      else if (key === 'pause') message.info('已暂停自动化；已提交给模型或合并器的在途任务不会被强制删除。');
      else if (key === 'resume') message.success('已继续自动化；已在途任务将继续回读，失败小说请使用“重试失败小说”。');
      else if (key === 'retry') message.success('已重置失败或阻塞小说，并从缺失阶段继续。');
      else if (key === 'cancel') message.info('已停止自动化编排；已在途的供应商任务仍会保留。');
      await Promise.allSettled([refreshBatch(), loadRuntimeStatus({ quiet: true })]);
    } catch (error) {
      message.error(error?.message || '批量工厂自动化操作失败');
    } finally { hideAutomationLoading(); automationBusyAtRef.current = 0; setAutomationBusy(''); }
  }

  function openMetadataEditor(book) {
    setMetadataBook(book);
    setMetadataValue({ ...(book?.sourceMetadata || {}) });
  }
  async function classifyBookMetadata(book, force = true) {
    if (!batch?.id || !book?.id || actionBusy) return;
    setActionBusy(`classify-${book.id}`);
    try {
      const result = await classifyBookPublishMetadata(batch.id, book.id, {
        textModelId: effectiveBookSettings(batch, book).textModelId,
        force
      });
      const nextBook = resultData(result, 'book');
      await refreshBatch();
      if (nextBook?.id === viewingBook?.id) setViewingBook(nextBook);
      message.success(resultData(result, 'reused') ? '已复用当前书的发布分类信息。' : '已识别当前书的男女频、风格和标签。');
    } catch (error) {
      // Classification failures are persisted by the service.  Reload the
      // current book so the modal keeps the actionable reason instead of
      // leaving the user with a disappearing toast only.
      try {
        const refreshed = resultData(await getBatch(batch.id), 'batch');
        const latest = (refreshed?.books || []).find(item => String(item?.id) === String(book.id));
        if (latest) setViewingBook(latest);
      } catch (_) { /* retain the original request error */ }
      message.error(error?.message || 'AI 判断发布信息失败');
    } finally { setActionBusy(''); }
  }

  // 只为没有分类状态的旧书补一次判断；已记录失败的书只允许用户手动重试。
  useEffect(() => {
    if (!batch?.id || !books.length) return undefined;
    const pending = books.filter(book => {
      const attemptKey = `${batch.id}:${book.id}`;
      return shouldAutoClassifyBook(book) && !autoClassificationInFlightRef.current.has(String(book.id)) && !autoClassificationAttemptedRef.current.has(attemptKey);
    });
    if (!pending.length) return undefined;
    let active = true;
    (async () => {
      for (const book of pending) {
        if (!active) break;
        const bookID = String(book.id);
        autoClassificationAttemptedRef.current.add(`${batch.id}:${bookID}`);
        autoClassificationInFlightRef.current.add(bookID);
        try {
          await classifyBookMetadata(book, false);
        } finally {
          autoClassificationInFlightRef.current.delete(bookID);
        }
      }
    })();
    return () => { active = false; };
    // 只在批次或书籍分类状态变化时重新判断，避免轮询反复触发 AI。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [batch?.id, books.map(book => `${book.id}:${book.revision}:${book.sourceMetadata?.classifyStatus || ''}:${book.sourceMetadata?.gender || ''}:${book.sourceMetadata?.style || ''}`).join('|')]);
	async function backfill121Metadata() {
		if (!batch?.id || actionBusy) return;
		setActionBusy('backfill-121-metadata');
		try {
			const result = await backfillBatch121Metadata(batch.id);
			const rows = Array.isArray(resultData(result, 'results')) ? resultData(result, 'results') : [];
			const backfilled = rows.filter(row => row?.status === 'backfilled').length;
			const aiFallback = rows.filter(row => row?.classification?.status === 'classified').length;
			const failed = rows.filter(row => row?.status === 'failed' || row?.classification?.status === 'failed').length;
			await refreshBatch();
			if (failed) message.warning(`121 已补全 ${backfilled} 本，AI 已兜底 ${aiFallback} 本；${failed} 本未完成，请在书籍资料中查看原因。`);
			else if (backfilled || aiFallback) message.success(`121 已补全 ${backfilled} 本，AI 已兜底 ${aiFallback} 本。`);
			else message.info('没有需要补全的 121 分类信息，也没有需要 AI 判断的书籍。');
		} catch (error) {
			message.error(error?.message || '121 分类信息回填失败');
		} finally { setActionBusy(''); }
	}
  async function saveBookMetadata() {
    if (!batch?.id || !metadataBook || metadataSaving) return;
    setMetadataSaving(true);
    try {
      await updateBookMetadata(batch.id, metadataBook.id, { metadata: metadataValue, expectedRevision: Number(metadataBook.revision || 0) });
      await refreshBatch();
      setViewingBook(null);
      setMetadataBook(null);
      message.success('小说列表元数据已保存。');
    } catch (error) { message.error(error?.message || '保存小说元数据失败'); } finally { setMetadataSaving(false); }
  }
  async function saveSettings(patch) {
    const normalized = normalizeAudioSourceSettings({ ...patch, audioDurationSeconds: 0 });
    try {
      await saveBatchSettings(batch.id, { patch: normalized, expectedRevision: Number(batch?.settingsState?.revision || 0) }, { suppressGlobalError: true });
    } catch (error) {
      if (Number(error?.status) !== 409) {
        message.error(error?.message || '应用统一引擎配置失败');
        return false;
      }
      try {
        const latestResult = await getBatch(batch.id);
        const latestBatch = resultData(latestResult, 'batch');
        if (!sameSettingsPatch(latestBatch?.settingsState?.patch, batch?.settingsState?.patch)) {
          await refreshBatch();
          message.warning('配置已更新，已刷新当前统一配置；请核对后重新保存。');
          return false;
        }
        await saveBatchSettings(batch.id, {
          patch: normalized,
          expectedRevision: Number(latestBatch?.settingsState?.revision || 0)
        }, { suppressGlobalError: true });
      } catch (retryError) {
        message.error(retryError?.message || '应用统一引擎配置失败');
        return false;
      }
    }
    try {
      try { await refreshBatch(); } catch (_) { /* 刷新失败不影响已落库的写透结果 */ }
      const savedResult = await getBatch(batch.id);
      const savedBatch = resultData(savedResult, 'batch');
      await syncUnifiedSettingsToBooks(savedBatch);
      const syncedResult = await getBatch(batch.id);
      const syncedBatch = resultData(syncedResult, 'batch');
      const syncedBooks = Array.isArray(syncedBatch?.books) ? syncedBatch.books : [];
      // 写透落库成功即关闭弹窗并给成功提示。配音逐行测量与提示词重编译耗时很长，
      // 转到后台队列，避免用户一直卡在“保存中”。
      message.success(`统一配置已覆盖本批次 ${syncedBooks.length} 本书的对应配置；后台正在按每本书自己的引擎配置更新配音与提示词。`);
      void runSettingsBackgroundSync(normalized, syncedBatch);
      return true;
    } catch (error) { message.error(error?.message || '统一配置已保存，但刷新工作台失败'); return false; }
  }

  // 保存后的后台收尾：逐本书测量配音 → 重编译最终 VIDEO Prompt。
  // 每本书都按 effectiveBookSettings（单书优先）解析引擎，批次统一配置只作兜底，
  // 因此不会出现“统一配置的引擎盖掉单书引擎”。进度/失败写入 backgroundSettingsTask。
  async function runSettingsBackgroundSync(patch, initialBatch) {
    const runId = backgroundSettingsRunRef.current + 1;
    backgroundSettingsRunRef.current = runId;
    const bookList = Array.isArray(initialBatch?.books) ? initialBatch.books : [];
    const patchState = next => setBackgroundSettingsTask(current => (
      current && current.runId === runId ? { ...current, ...next } : current
    ));
    setBackgroundSettingsTask({ runId, running: true, phase: 'audio', done: 0, total: bookList.length, currentTitle: '', failed: 0, errors: [], notice: '' });
    const aborted = () => backgroundSettingsRunRef.current !== runId;
    try {
      if (batchFactoryShouldPrepareAudio(patch)) {
        for (const book of bookList) {
          if (aborted()) return;
          try {
            await ensureBookAudioDuration(book, { force: true, quiet: true, batchSnapshot: initialBatch });
          } catch (error) {
            if (aborted()) return;
            setBackgroundSettingsTask(current => current && current.runId === runId ? {
              ...current,
              failed: current.failed + 1,
              errors: [...current.errors, `${book.title || '当前小说'}：${error?.message || '配音测量失败'}`]
            } : current);
          }
          if (aborted()) return;
          setBackgroundSettingsTask(current => current && current.runId === runId ? {
            ...current,
            done: current.done + 1,
            currentTitle: book.title || ''
          } : current);
        }
      }
      if (aborted()) return;
      if (batchFactoryShouldRecompileVideoPrompts(patch)) {
        patchState({ phase: 'recompile' });
        const compiledResult = await getBatch(initialBatch.id);
        const compiledBatch = resultData(compiledResult, 'batch');
        const recompileResult = await recompileUnifiedH3Prompts(compiledBatch);
        if (aborted()) return;
        setBackgroundSettingsTask(current => current && current.runId === runId ? {
          ...current,
          phase: 'done',
          failed: current.failed + recompileResult.failed.length,
          errors: [...current.errors, ...recompileResult.failed.map(item => `提示词重编译：${item}`)],
          notice: `${recompileResult.compiled} 本最终 VIDEO Prompt 已更新${recompileResult.manual ? `，${recompileResult.manual} 本保留手动提示词` : ''}${recompileResult.waitingForAudio ? `，${recompileResult.waitingForAudio} 本等待已有逐行配音后再编译` : ''}。`
        } : current);
      } else {
        patchState({ phase: 'done', notice: '配置已落库，本次无需重编译提示词。' });
      }
      try { await refreshBatch(); } catch (_) { /* 后台刷新失败不覆盖已有进度 */ }
      patchState({ running: false });
    } catch (error) {
      if (aborted()) return;
      setBackgroundSettingsTask(current => current && current.runId === runId ? {
        ...current, running: false, phase: 'failed',
        errors: [...current.errors, error?.message || '后台处理失败']
      } : current);
    }
  }
  function retrySettingsBackgroundSync() {
    void runSettingsBackgroundSync({ ...(batch?.settingsState?.patch || {}) }, batch);
  }
  async function previewChangeImpact(patch) {
    try { const impact = await getChangeImpact(batch.id, { patch, expectedRevision: Number(batch?.settingsState?.revision || 0) }); return { ok: true, impact: resultData(impact, 'impact') }; } catch (error) { return { ok: false, message: error?.message || '无法读取配置影响' }; }
  }

  function openComicBatchPublish(scope = 'all') {
    const targetIDs = scope === 'selected' ? selectedBookIds : books.map(book => book.id);
    if (!targetIDs.length) {
      message.info(scope === 'selected' ? '请先勾选要发布的小说。' : '当前批次还没有小说。');
      return;
    }
    setSelectedBookIds(targetIDs);
    setComicBatchPublishOpen(true);
  }
  const batchMenuItems = batchProductionMode === 'comic'
    ? [
        { key: 'selected-production', label: automationActive ? `排队执行已选（${selectedBookIds.length}）` : `立即执行已选（${selectedBookIds.length}）`, disabled: !selectedBookIds.length, title: automationActive ? '当前任务完成后自动继续，不会重复生成已完成小说' : !selectedBookIds.length ? '请先在小说列表中勾选要立即执行的小说' : undefined },
        { key: 'comic-pages', label: '批量分析漫画分页', disabled: !books.length },
        { key: 'comic-collection', label: '批量生成漫画并整理合集', disabled: !books.length }
      ]
    : [
        { key: 'selected-production', label: automationActive ? `排队执行已选（${selectedBookIds.length}）` : `立即执行已选（${selectedBookIds.length}）`, disabled: !selectedBookIds.length, title: automationActive ? '当前任务完成后自动继续，不会重复生成已完成小说' : !selectedBookIds.length ? '请先在小说列表中勾选要立即执行的小说' : undefined },
        { key: 'production', label: '批量生成视频', disabled: !productionCapability.available, title: productionCapability.reason },
        { key: 'merge', label: '合并当前主 VIDEO', disabled: !mergeCapability.available, title: mergeCapability.reason }
      ];
  const agentMenuItems = [
    { key: 'all', label: `全部小说（${books.length}）`, disabled: !books.length },
    { type: 'divider' },
    { key: 'selected', label: `优化选择（${selectedBookIds.length}）`, disabled: selectedBookIds.length !== 1, title: '先在小说列表中勾选一本小说' }
  ];
  const automationStartBlocked = !books.length;
  const automationRetryAction = batchFactoryAutomationRetryAction(automationStatus);
  const automationMenuItems = [
    { key: 'start', label: '启动自动生产（选择预设）', disabled: automationActive || automationStartBlocked },
    { key: 'pause', label: '暂停自动化', disabled: !automationActive },
    { key: 'resume', label: '继续已暂停自动化', disabled: !['paused', 'needs_attention', 'completed'].includes(automationState) },
    { key: 'retry', ...automationRetryAction },
    { key: 'cancel', label: '停止自动化编排', danger: true, disabled: ['idle', 'completed', 'cancelled'].includes(automationState) }
  ];
  const automationLabel = automationState === 'running'
    ? `自动生产 ${automationCounts.ready || 0}/${automationCounts.total || books.length}`
    : automationState === 'scheduled'
      ? '自动生产 · 已定时'
      : automationState === 'paused'
        ? '自动生产 · 已暂停'
        : automationState === 'needs_attention'
          ? '自动生产 · 待处理'
          : automationState === 'unavailable'
            ? '自动生产 · 状态暂不可读'
          : automationState === 'completed'
            ? '自动生产 · 已就绪'
            : '自动生产';
  const uploadMenuItems = [{ key: 'selected', label: `提交选中（${selectedBookIds.length}）`, disabled: !selectedBookIds.length }, { key: 'all', label: '提交全部' }];

  function moveColumnResize(event) {
    const resize = resizeStateRef.current;
    if (!resize || !Number.isFinite(event.clientX)) return;
    const nextWidths = resize.startingWidths.map((width, columnIndex) => columnIndex === resize.index ? Math.max(BATCH_FACTORY_TABLE_COLUMNS[resize.index].minWidth, Math.min(780, resize.startWidth + event.clientX - resize.startX)) : width);
    setColumnWidths(nextWidths);
  }

  function finishColumnResize() {
    if (!resizeActiveRef.current && !resizeStateRef.current) return;
    resizeActiveRef.current = false;
    resizeStateRef.current = null;
    document.body.classList.remove('batch-factory-column-resizing');
    document.removeEventListener('mousemove', moveColumnResize);
    document.removeEventListener('mouseup', finishColumnResize);
    document.removeEventListener('pointermove', moveColumnResize);
    document.removeEventListener('pointerup', finishColumnResize);
  }

  function startColumnResize(event, index, preventDefault = true) {
    if (resizeActiveRef.current) return;
    const header = event.currentTarget.closest('.shuihuo-workbench-head');
    const measured = header ? (getComputedStyle(header).gridTemplateColumns.match(/[\d.]+px/g)?.map(Number.parseFloat) || []) : [];
    const savedWidths = Array.isArray(columnWidths) && columnWidths.length === BATCH_FACTORY_TABLE_COLUMNS.length && columnWidths.every(Number.isFinite) ? columnWidths : null;
    const startingWidths = savedWidths || measured;
    if (startingWidths.length !== BATCH_FACTORY_TABLE_COLUMNS.length || !startingWidths.every(Number.isFinite)) return;
    if (preventDefault) {
      event.preventDefault();
      event.stopPropagation();
    }
    resizeActiveRef.current = true;
    resizeStateRef.current = { index, startX: event.clientX, startWidth: startingWidths[index], startingWidths };
    if (Number.isInteger(event.pointerId)) event.currentTarget.setPointerCapture?.(event.pointerId);
    document.body.classList.add('batch-factory-column-resizing');
    document.addEventListener('mousemove', moveColumnResize);
    document.addEventListener('mouseup', finishColumnResize, { once: true });
    document.addEventListener('pointermove', moveColumnResize);
    document.addEventListener('pointerup', finishColumnResize, { once: true });
  }

  const hasComicBooks = batchProductionMode === 'comic' || books.some(book => effectiveBookSettings(batch, book).productionMode === 'comic');
  const comicStats = Object.values(stageSummaries || {}).reduce((totals, summary) => ({
    imageReady: totals.imageReady + Number(summary?.comicImageReadyCount || 0),
    imageTotal: totals.imageTotal + Number(summary?.comicImageTotalCount || 0),
    collections: totals.collections + Number(summary?.comicCollectionCount || 0),
    collectionPages: totals.collectionPages + Number(summary?.comicCollectionPageCount || 0),
    published: totals.published + Number(summary?.comicPublishedCollectionCount || 0)
  }), { imageReady: 0, imageTotal: 0, collections: 0, collectionPages: 0, published: 0 });

  // 自动化状态提示：列表打开时作为头部状态行渲染（不再整宽压住列表内容），
  // 生产表视图下才作为表格上方的状态行。
  const automationStatusNode = automationState !== 'idle' ? (() => {
    const stateLabel = ({ running: '运行中', paused: '已暂停', scheduled: '定时待执行', needs_attention: '待处理', stopped: '已停止', completed: '已完成', failed: '失败' })[automationState] || automationState;
    const problemCount = Number(automationCounts.failed || 0) + Number(automationCounts.blocked || 0);
    const batchRetryCount = (Array.isArray(automationStatus?.books) ? automationStatus.books : []).filter(book => book?.batchRetryPending === true).length;
    // 漫画页进度条只统计漫画页主图，资产图阶段不会让它增长。这里把正在跑的
    // 书按阶段拆开显示，避免把“在生成资产图”误读成卡死。
    const runningBooks = (Array.isArray(automationStatus?.books) ? automationStatus.books : []).filter(book => book?.status === 'running');
    const assetStageCount = runningBooks.filter(book => String(book?.stage || '').includes('asset')).length;
    const pageStageCount = runningBooks.filter(book => book?.stage === 'comic_images').length;
    const lastError = String(automationStatus?.lastError || '').trim();
    return (
      <button
        type="button"
        data-testid="automation-status-line"
        className="batch-factory-automation-status-line"
        title="打开任务/日志查看自动化详情"
        onClick={() => setLogsOpen(true)}
      >
        自动化：{stateLabel}
        {problemCount > 0 ? `，失败/阻塞 ${problemCount} 本` : ''}
        {runningBooks.length > 0 ? `，出图中 ${runningBooks.length} 本（资产图 ${assetStageCount} 本 / 漫画页 ${pageStageCount} 本）` : ''}
        {batchRetryCount > 0 ? `，已跳过 ${batchRetryCount} 本等待批末回补` : ''}
        {lastError ? `，最近错误：${lastError}` : ''}
      </button>
    );
  })() : null;

  return <section className="shuihuo-workbench batch-factory-workbench">
    <header className="shuihuo-workbench-header">
      <div className="shuihuo-workbench-heading"><button className="shuihuo-back-link" type="button" title="返回批量作品列表" aria-label="返回批量作品列表" onClick={onBack}><ArrowLeftOutlined /></button><strong>批量工厂 V12 · {batch?.title || '未命名批量'}</strong><div className="shuihuo-workbench-progress" aria-label={`完成率 ${batchProgress.completionPercent}%`}><div className="shuihuo-workbench-progress-track batch-factory-completion-track">{progressSegments.map(segment => <i key={segment.key} className={`is-${segment.key}`} title={`${segment.label} ${segment.count} 本`} style={{ width: `${segment.percent}%` }} />)}</div><span>{batchProgress.completionPercent}%</span></div></div>
      <div className="shuihuo-workbench-toolbar" role="toolbar" aria-label="批量工厂工具栏">
        <Button type={novelListOpen ? 'primary' : 'text'} icon={<BarsOutlined />} onClick={() => setNovelListOpen(current => !current)}>{novelListOpen ? '返回生产表' : '小说列表'}</Button>
        <Button type="text" icon={<SettingOutlined />} onClick={() => setUnifiedSettingsOpen(true)}>统一配置</Button>
        <Dropdown menu={{ items: automationMenuItems, onClick: ({ key }) => key === 'start' ? openAutomationStart() : runAutomationAction(key) }}><Button type="text" loading={Boolean(automationBusy)} disabled={automationState === 'idle' && automationStartBlocked}>{automationLabel}</Button></Dropdown>
        <Dropdown menu={{ items: agentMenuItems, onClick: ({ key }) => key === 'all' ? openBatchInAgent() : openSelectedInAgent() }}><Button type="text" disabled={Boolean(actionBusy)} loading={actionBusy === 'agent-all' || actionBusy?.startsWith?.('agent-')}>Agent 优化</Button></Dropdown>
        <Dropdown menu={{ items: batchMenuItems, onClick: async ({ key }) => { if (key === 'selected-production') return runSelectedProduction(); if (key === 'production') return runProduction(); if (key === 'merge') return runMerge(); if (key === 'comic-pages') { setAutomationRunMode('storyboard_only'); return openAutomationStart(); } if (key === 'comic-collection') { setAutomationRunMode('video_no_submit'); return openAutomationStart(); } } }}><Button className="shuihuo-batch-button" type="text" icon={<PictureOutlined />} loading={actionBusy === 'production' || actionBusy === 'selected-production' || actionBusy === 'merge'}>批量操作</Button></Dropdown>
		{batchProductionMode === 'video' ? <Tooltip title={cancelCapability.available ? '取消可取消的本地执行器任务；其它供应商保持在途状态。' : cancelCapability.reason}><Button className="shuihuo-cancel-button" type="text" loading={actionBusy === 'cancel'} disabled={!cancelCapability.available || Boolean(actionBusy)} onClick={cancelProduction}>取消操作</Button></Tooltip> : null}
      </div>
      <div className="shuihuo-workbench-export"><Button type="text" icon={<BarsOutlined />} onClick={() => setLogsOpen(true)}>任务/日志</Button><Dropdown menu={{ items: uploadMenuItems, onClick: ({ key }) => batchProductionMode === 'comic' ? openComicBatchPublish(key) : setUploadMode(key) }}><Button icon={<UploadOutlined />}>{batchProductionMode === 'comic' ? '发布漫画' : '上传网络'}</Button></Dropdown></div>
      <div className="shuihuo-project-stats"><span><b>{books.length}</b> 本小说</span>{hasComicBooks ? <><span className="is-uploaded"><b>{comicStats.imageReady}/{comicStats.imageTotal}</b> 漫画图片</span><span><b>{comicStats.collections}</b> 漫画合集（{comicStats.collectionPages} 页）</span><span className={comicStats.published > 0 ? 'is-uploaded' : 'is-disabled'}><b>{comicStats.published}</b> 已发布合集</span></> : <><span className="is-uploaded"><b>{batchProgress.uploaded}</b> 上传成功</span>{batchProgress.failed ? <span className="is-failed"><b>{batchProgress.failed}</b> 异常</span> : null}{batchProgress.running ? <span className="is-running"><b>{batchProgress.running}</b> 执行中</span> : null}{batchProgress.awaitingUpload ? <span className="is-awaiting-upload"><b>{batchProgress.awaitingUpload}</b> 待上传</span> : null}{batchProgress.pending ? <span><b>{batchProgress.pending}</b> 待开始</span> : null}</>}{automationState !== 'idle' ? <span className={automationState === 'needs_attention' ? 'is-failed' : automationActive ? 'is-running' : 'is-awaiting-upload'}><b>{automationCounts.ready || 0}/{automationCounts.total || books.length}</b> 自动就绪</span> : null}{!hasComicBooks ? <><span><b>{mediaCounts.storyboards}</b> 分镜</span>{productionCapability.available ? <span className="is-running"><b>{mediaCounts.playable}</b> 可播放</span> : <span className="is-disabled">视频模块未启用</span>}{mergeCapability.available ? <span className="is-uploaded"><b>{mediaCounts.merges}</b> 成片</span> : <span className="is-disabled">合并模块未启用</span>}</> : null}</div>
    </header>
    {
      novelListOpen ? (
        <section
          className="batch-factory-novel-list-view"
          aria-label="小说列表视图"
        >
          <header className="batch-factory-novel-list-view-header">
            <div>
              <strong>小说列表</strong>
              <span>共 {books.length} 本，查看书目信息或添加新书</span>
              {automationStatusNode}
            </div>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => setNovelListOpen(false)}
              aria-label="返回生产表"
            >
              返回生产表
            </Button>
          </header>
          <NovelMetadata
            books={books}
            createdAt={batch?.createdAt}
            selectedBookIds={selectedBookIds}
            onSelectionChange={setSelectedBookIds}
            onViewBook={setViewingBook}
            onRunBook={runSingleBook}
            platformNames={platformNames}
            platformOptions={platformOptions}
            productionStatus={productionStatus}
            mergeStatus={mergeStatus}
            stageSummaries={stageSummaries}
            batchId={batch?.id}
            onContentReady={refreshBatch}
            onBackfill121Metadata={backfill121Metadata}
            backfilling121Metadata={actionBusy === "backfill-121-metadata"}
            onAddBooks={addBooksToBatch}
          />
        </section>
      ) : (
        <div
          className="shuihuo-workbench-table batch-factory-workbench-table"
          role="table"
          aria-label="批量工厂小说生产表"
        >
          <div
            className="shuihuo-workbench-head"
            role="row"
            style={
              columnWidths
                ? {
                    gridTemplateColumns: columnWidths
                      .map((width) => `${width}px`)
                      .join(" "),
                  }
                : undefined
            }
          >
            {BATCH_FACTORY_TABLE_COLUMNS.map((column, index) => {
              const columnLabel =
                batchProductionMode === "comic" && index === 4
                  ? "漫画页"
                  : column.label;
              return (
                <div role="columnheader" key={column.label}>
                  {index === 0 ? (
                    <>
                      <Checkbox
                        checked={
                          books.length > 0 &&
                          selectedBookIds.length === books.length
                        }
                        indeterminate={
                          selectedBookIds.length > 0 &&
                          selectedBookIds.length < books.length
                        }
                        onChange={(event) =>
                          setSelectedBookIds(
                            event.target.checked
                              ? books.map((book) => book.id)
                              : [],
                          )
                        }
                        aria-label="全选小说"
                      />
                      <span>序号</span>
                    </>
                  ) : (
                    columnLabel
                  )}
                  {index < BATCH_FACTORY_TABLE_COLUMNS.length - 1 ? (
                    <button
                      type="button"
                      draggable
                      className="batch-factory-column-resize-handle"
                      onPointerDown={(event) => startColumnResize(event, index)}
                      onMouseDown={(event) => startColumnResize(event, index)}
                      onDragStart={(event) =>
                        startColumnResize(event, index, false)
                      }
                      onDrag={moveColumnResize}
                      onDragEnd={finishColumnResize}
                      aria-label={`调整${columnLabel}列宽`}
                      title="拖动分隔线调整列宽"
                    >
                      <i aria-hidden="true" />
                    </button>
                  ) : null}
                </div>
              );
            })}
          </div>
          {backgroundSettingsTask ? (
            <div className={`batch-factory-background-task${backgroundSettingsTask.errors.length ? ' is-error' : ''}`} role="status" aria-live="polite">
              {backgroundSettingsTask.running ? (
                <span className="batch-factory-background-task-text">
                  {backgroundSettingsTask.phase === 'recompile'
                    ? '后台处理中 · 重编译提示词…'
                    : `后台处理中 · 配音测量 ${backgroundSettingsTask.done}/${backgroundSettingsTask.total}`}
                  {backgroundSettingsTask.currentTitle ? `（${backgroundSettingsTask.currentTitle}）` : ''}
                </span>
              ) : backgroundSettingsTask.errors.length ? (
                <>
                  <span className="batch-factory-background-task-text">后台处理有 {backgroundSettingsTask.errors.length} 项未完成：{backgroundSettingsTask.errors.join('；')}</span>
                  <Button size="small" onClick={retrySettingsBackgroundSync}>重试后台处理</Button>
                </>
              ) : (
                <span className="batch-factory-background-task-text">{backgroundSettingsTask.notice || '后台处理已完成。'}</span>
              )}
            </div>
          ) : null}
          {books.map((book, index) => {
            const stageSummary = stageSummaries[book.id];
            const productionMode =
              effectiveBookSettings(batch, book).productionMode === "comic"
                ? "comic"
                : "video";
            const automationBook = (() => {
              const scoped = automationBookForCard(automationStatus, batch.id, book.id);
              return scoped && !scoped.stale ? scoped : null;
            })();
            const state = batchFactoryBookState(book, {
              productionStatus,
              mergeStatus,
              stageSummary,
              productionMode,
              automationBook,
            });
            const rangeLines = contentRangeLinesForBook(book);
            const previewText = batchFactoryProductionText(
              book.sourceText,
              book.workingFrontContent,
              book,
            );
            const publishText = publishContentWithWorkingFront(
              book.sourceText,
              book.workingFrontContent,
              book,
            );
            const sourceVersionLabel = batchFactorySourceVersionLabel(book);
            const sourceVersionClass = batchFactorySourceVersionClass(book);
            const sourceSummary = batchFactorySourceSummary(
              book,
              productionMode,
            );
            const sourceStatus = batchFactorySourceStatus(book);
            const videos = book.videos || [];
            const lastFailedStage = stageSummary?.lastFailed;
            // 自动生产正在执行时给整行加流动光边：阶段不同颜色不同，
            // 用户在长列表里也能直接看到“这本文在跑什么”。
            const automationRunningStage = String(automationBook?.status || '').toLowerCase() === 'running'
              ? String(automationBook?.stage || '').trim()
              : '';
            return (
              <article
                className={`shuihuo-workbench-row batch-factory-book-row${automationRunningStage ? ' is-automation-running' : ''}`}
                data-automation-stage={automationRunningStage || undefined}
                key={book.id}
                role="row"
                style={
                  columnWidths
                    ? {
                        gridTemplateColumns: columnWidths
                          .map((width) => `${width}px`)
                          .join(" "),
                      }
                    : undefined
                }
              >
                <div className="shuihuo-workbench-cell shuihuo-order-cell">
                  <Checkbox
                    checked={selectedBookIds.includes(book.id)}
                    onChange={(event) =>
                      setSelectedBookIds((current) =>
                        event.target.checked
                          ? [...new Set([...current, book.id])]
                          : current.filter((id) => id !== book.id),
                      )
                    }
                    aria-label={`选择 ${book.title || `小说 ${index + 1}`}`}
                  />
                  <strong>{index + 1}</strong>
                </div>
                <div
                  className="shuihuo-workbench-cell batch-factory-book-content"
                  role="button"
                  tabIndex={0}
                  title="点击编辑当前小说所选行数的内容"
                  onClick={() => openContentEditor(book)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter" || event.key === " ") {
                      event.preventDefault();
                      openContentEditor(book);
                    }
                  }}
                >
                  <div className="batch-factory-book-title-row">
                    <strong>{book.title || `小说 ${index + 1}`}</strong>
                    <span
                      className={`batch-factory-production-source ${sourceSummary.ready ? "is-ready" : "is-missing"}`}
                      title={`${sourceVersionLabel} · ${sourceSummary.text}`}
                    >
                      <i
                        className={`batch-factory-content-mode is-${sourceVersionClass}`}
                      >
                        {sourceVersionLabel}
                      </i>
                    </span>
                    <span className={`batch-factory-source-kind is-${sourceStatus.tone}`} title={sourceStatus.detail}>{batchFactorySourceKind(book)}</span>
                  </div>
                  <span>
                    bookId：{book.bookId || "—"} · 书城：
                    {bookPlatformName(book, platformNames)} ·{" "}
                    {productionMode === "comic"
                      ? `正文总字数 ${batchFactoryCleanSourceText(book.sourceText).length} 字 · 漫画分析 ${comicAnalysisCharactersForBook(book)} 字`
                      : `生产前 ${rangeLines} 行`}
                  </span>
                  <p className={!previewText ? `batch-factory-source-empty is-${sourceStatus.tone}` : ""}>{previewText || sourceStatus.label}</p>
                  {!previewText ? (
                    <BatchFactoryGiantMaterialPendingProgress
                      book={book}
                      batchId={batch?.id}
                      visible
                      onContentReady={refreshBatch}
                    />
                  ) : null}
                  <Space wrap className="batch-factory-book-content-actions">
                    <Button
                      size="small"
                      className="batch-factory-book-content-action is-edit"
                      onClick={(event) => {
                        event.stopPropagation();
                        openContentEditor(book);
                      }}
                    >
                      编辑内容
                    </Button>
                    <Button
                      size="small"
                      className="batch-factory-book-content-action is-download"
                      disabled={!publishText}
                      onClick={(event) => {
                        event.stopPropagation();
                        downloadProductionContent({
                          title: book.title,
                          content: publishText,
                        });
                      }}
                    >
                      下载 TXT
                    </Button>
                    <Popconfirm
                      title="删除本书？"
                      description="将删除本书及本地生产记录，不影响已提交到 121 的内容。"
                      okText="删除"
                      cancelText="取消"
                      onConfirm={(event) => {
                        event?.stopPropagation?.();
                        return deleteBook(book);
                      }}
                    >
                      <Button
                        danger
                        size="small"
                        className="batch-factory-book-content-action is-delete"
                        icon={<DeleteOutlined />}
                        loading={actionBusy === `delete-${book.id}`}
                        onClick={(event) => event.stopPropagation()}
                      >
                        删除本书
                      </Button>
                    </Popconfirm>
                  </Space>
                </div>
                <div className="shuihuo-workbench-cell batch-factory-book-config-cell">
                  <div className="batch-factory-book-config-regions">
                    {batchFactoryWorkbenchConfigRegions(book).map((region) => {
                      const status =
                        region.combinedStatus ||
                        bookConfigRegionStatus(book, region.key, batch?.settingsState?.patch || {});
                      const detail =
                        region.key === "assets"
                          ? bookAssetSummary(book)
                          : status.label;
                      return (
                        <button
                          type="button"
                          key={region.key}
                          className={`batch-factory-book-config-region is-${status.tone}`}
                          onClick={() =>
                            region.key === "assets"
                              ? setAssetBook(book)
                              : setConfigTarget({ book, region: region.key })
                          }
                        >
                          <b>{region.label}</b>
                          <small>{detail}</small>
                        </button>
                      );
                    })}
                  </div>
                </div>
                <DeferredInlineBookPrompts
                  batch={batch}
                  book={book}
                  batchId={batch?.id}
                  settingsRevision={batch?.settingsState?.revision}
                  productionMode={productionMode}
                  selectedVideoId={
                    rowStoryboardSelection[book.id] || videos[0]?.id || ""
                  }
                  onSelectedVideoChange={(videoId) =>
                    setRowStoryboardSelection((current) => ({
                      ...current,
                      [book.id]: videoId,
                    }))
                  }
                  onManage={(videoId) => {
                    setPromptVideoId(videoId || "");
                    setPromptBook(book);
                  }}
                />
                <div className="shuihuo-workbench-cell shuihuo-library-cell batch-factory-book-library-cell">
                  {productionMode === "comic" ? (
                    <DeferredInlineComicLibrary
                      batchId={batch?.id}
                      book={book}
                      followUntil={comicFollowUntil}
                      refreshNonce={promptBook?.id === book.id || mediaBook?.id === book.id ? comicPageRefreshNonce : 0}
                      selectedVideoId={
                        rowStoryboardSelection[book.id] || videos[0]?.id || ""
                      }
                      onSelectedVideoChange={(videoId) =>
                        setRowStoryboardSelection((current) => ({
                          ...current,
                          [book.id]: videoId,
                        }))
                      }
                      onManage={(videoId) => {
                        setMediaVideoId(videoId || "");
                        setMediaStartTab("comic");
                        setMediaBook(book);
                      }}
                      onChanged={refreshBatch}
                    />
                  ) : (
                    <InlineMediaLibrary
                      book={book}
                      productionStatus={productionStatus}
                      mergeJob={latestBookMerge(mergeStatus, book.id)}
                      aspectRatio={
                        effectiveBookSettings(batch, book).aspectRatio
                      }
                      h3CompilationMessage={h3CompilationStatus(
                        effectiveBookSettings(batch, book),
                      )}
                      selectedVideoId={
                        rowStoryboardSelection[book.id] || videos[0]?.id || ""
                      }
                      onSelectedVideoChange={(videoId) =>
                        setRowStoryboardSelection((current) => ({
                          ...current,
                          [book.id]: videoId,
                        }))
                      }
                      onManage={(videoId) => {
                        setMediaVideoId(videoId || "");
                        setMediaStartTab("clips");
                        setMediaBook(book);
                      }}
                      onOpenMerge={() => {
                        setMediaVideoId("");
                        setMediaStartTab("merges");
                        setMediaBook(book);
                      }}
                    />
                  )}
                </div>
                <div className="shuihuo-workbench-cell batch-factory-actions">
                  <div className="batch-factory-action-group is-utility">
                    <span>资料与流程</span>
                    <div className="batch-factory-action-button-grid">
                      <Button size="small" onClick={() => setViewingBook(book)}>
                        查看资料
                      </Button>
                      {previewText ? (
                        <Button
                          size="small"
                          onClick={() => refreshBatch()}
                          disabled={Boolean(actionBusy)}
                        >
                          刷新状态
                        </Button>
                      ) : (
                        <Button
                          size="small"
                          type="primary"
                          onClick={() => fetchMissingBookSource(book)}
                          loading={actionBusy === `source-${book.id}`}
                          disabled={Boolean(actionBusy)}
                        >
                          获取正文
                        </Button>
                      )}
                    </div>
                  </div>
                  <div className="batch-factory-action-group is-production">
                    <span>
                      {productionMode === "comic" ? "漫画资产" : "资产获取"}
                    </span>
                    <div className="batch-factory-action-button-grid">
                      {productionMode === "comic" ? (
                        <Tooltip title="在资产设置中选择漫画资产提示词，并执行提取漫画资产。">
                          <Button
                            size="small"
                            onClick={() => setAssetBook(book)}
                            disabled={Boolean(actionBusy)}
                          >
                            资产设置
                          </Button>
                        </Tooltip>
                      ) : (
                        <Tooltip
                          title={
                            runCapability.available
                              ? "提取当前书的人物、场景、道具提示词；手动资产保留。"
                              : runCapability.reason
                          }
                        >
                          <Button
                            size="small"
                            onClick={() =>
                              runBookStageAction(book, "assets", "force")
                            }
                            loading={
                              actionBusy === stageActionKey("assets", book.id)
                            }
                            disabled={
                              !runCapability.available || Boolean(actionBusy)
                            }
                          >
                            提取资产
                          </Button>
                        </Tooltip>
                      )}
                      <Tooltip title="打开资产图选择与生成面板；选择资产和图片模型后生成。">
                        <Button
                          size="small"
                          onClick={() => setAssetBook(book)}
                          disabled={Boolean(actionBusy)}
                        >
                          生成图片
                        </Button>
                      </Tooltip>
                    </div>
                  </div>
                  {productionMode === "comic" ? (
                    <div className="batch-factory-action-group is-production">
                      <span>漫画获取</span>
                      <div className="batch-factory-action-button-grid">
                        <Tooltip
                          title={
                            runCapability.available
                              ? "根据漫画资产设置中已提取的资产，分析完整原文、确定漫画分页，并生成每一页漫画分镜提示词；已有完整分页和提示词时直接复用。"
                              : runCapability.reason
                          }
                        >
                          <Button
                            size="small"
                            onClick={() =>
                              runBookStageAction(book, "comic", "bootstrap")
                            }
                            loading={
                              actionBusy === stageActionKey("comic", book.id)
                            }
                            disabled={
                              !runCapability.available || Boolean(actionBusy)
                            }
                          >
                            提取
                          </Button>
                        </Tooltip>
                        <Tooltip title={`按每页 ${COMIC_PAGE_CONCURRENCY} 路并发生成当前小说全部漫画页（单本最多前 ${COMIC_PAGE_GENERATION_LIMIT} 页），并整理漫画合集；生成中可随时停止。`}>
                          <Button
                            size="small"
                            type="primary"
                            loading={isComicTaskBusy(`comic-book-${book.id}`)}
                            disabled={Boolean(actionBusy) || isComicTaskBusy(`comic-book-${book.id}`)}
                            onClick={() => generateBookComic(book)}
                          >
                            生成整本漫画
                          </Button>
                        </Tooltip>
                        {comicRuns[book.id]?.running ? (
                          <Tooltip title="停止本书剩余页的派发；已完成的页会保留，未完成页可重试。">
                            <Button size="small" danger onClick={() => stopComicRun(book.id)}>停止本批</Button>
                          </Tooltip>
                        ) : null}
                      </div>
                      {comicRuns[book.id] ? (
                        <div className="batch-factory-comic-run" style={{ marginTop: 8, width: "100%", fontSize: 12 }}>
                          <div>
                            {comicRuns[book.id].running
                              ? `正在生成整本漫画…（${comicRuns[book.id].completed || 0}/${comicRuns[book.id].total || 0} 页）`
                              : comicRuns[book.id].stopped
                                ? `已停止整本漫画出图（完成 ${comicRuns[book.id].completed || 0}/${comicRuns[book.id].total || 0} 页，未完成页可重试）`
                                : comicRuns[book.id].failedCount
                                  ? `整本漫画有 ${comicRuns[book.id].failedCount} 页未完成（完成 ${comicRuns[book.id].completed || 0}/${comicRuns[book.id].total || 0} 页）`
                                  : `已生成整本漫画（${comicRuns[book.id].total || 0} 页）`}
                          </div>
                          {comicRuns[book.id].truncated ? (
                            <div style={{ color: "var(--text-warning)" }}>
                              本次只处理前 {COMIC_PAGE_GENERATION_LIMIT} 页，其余 {comicRuns[book.id].skipped} 页未包含。
                            </div>
                          ) : null}
                          {comicRuns[book.id].failed?.length ? (
                            <Collapse
                              size="small"
                              ghost
                              items={[{
                                key: "failed",
                                label: `展开失败页清单（${comicRuns[book.id].failed.length} 页）`,
                                children: (
                                  <ul style={{ margin: 0, paddingLeft: 18 }}>
                                    {comicRuns[book.id].failed.map((item, index) => (
                                      <li key={`${item.label}-${index}`}>{item.label}：{item.reason}</li>
                                    ))}
                                  </ul>
                                )
                              }]}
                            />
                          ) : null}
                        </div>
                      ) : null}
                    </div>
                  ) : (
                    <div className="batch-factory-action-group is-production">
                      <span>视频获取</span>
                      <div className="batch-factory-action-button-grid">
                        <Tooltip
                          title={
                            runCapability.available
                              ? h3DirectorCards(book).length &&
                                usesH3VideoPrompt(
                                  effectiveBookSettings(batch, book),
                                )
                                ? "复用已有统一导演卡，按当前规则编译最终 VIDEO Prompt。"
                                : "按统一流程生成导演分镜：选中的视频提示词决定 SD 或 H3 的输出协议。"
                              : runCapability.reason
                          }
                        >
                          <Button
                            size="small"
                            onClick={() =>
                              runBookStageAction(book, "director", "compile")
                            }
                            loading={
                              actionBusy === stageActionKey("director", book.id)
                            }
                            disabled={
                              !runCapability.available || Boolean(actionBusy)
                            }
                          >
                            提取
                          </Button>
                        </Tooltip>
                        <Tooltip
                          title={
                            productionCapability.available
                              ? "按当前书最终编译视频提示词、可用参考图和画幅创建 VIDEO 任务。"
                              : productionCapability.reason
                          }
                        >
                          <Button
                            size="small"
                            className="batch-factory-action-video"
                            onClick={() => runBookStageAction(book, "video")}
                            loading={
                              actionBusy === stageActionKey("video", book.id)
                            }
                            disabled={
                              !productionCapability.available ||
                              Boolean(actionBusy)
                            }
                          >
                            生成视频
                          </Button>
                        </Tooltip>
                      </div>
                    </div>
                  )}
                  <div className="batch-factory-action-group is-recovery">
                    <span>异常处理</span>
                    <div className="batch-factory-action-button-grid">
                      <Button
                        size="small"
                        onClick={() => refreshBatch()}
                        disabled={Boolean(actionBusy)}
                      >
                        刷新
                      </Button>
                      <Button
                        size="small"
                        danger
                        onClick={() => retryLastFailedStage(book)}
                        loading={
                          actionBusy === stageActionKey("retry", book.id)
                        }
                        disabled={
                          Boolean(actionBusy) &&
                          actionBusy !== stageActionKey("retry", book.id)
                        }
                      >
                        重试失败步骤
                      </Button>
                    </div>
                  </div>
                </div>
                {backgroundSettingsTask?.running ? (
                  <div className="batch-factory-book-row-bg-task" role="status">
                    后台处理中 · 配音测量 {backgroundSettingsTask.done}/{backgroundSettingsTask.total}
                  </div>
                ) : null}
              </article>
            );
          })}
          {!books.length ? (
            <div className="shuihuo-workbench-empty">
              当前批量还没有小说。返回个人作品后，从“批量工厂”新建书单。
            </div>
          ) : null}
        </div>
      )}
    <Modal
      title={viewingBook?.title || "小说详情"}
      open={Boolean(viewingBook)}
      onCancel={() => setViewingBook(null)}
      footer={
        viewingBook ? (
          <Space>
            <Button
              loading={actionBusy === `classify-${viewingBook.id}`}
              onClick={() => classifyBookMetadata(viewingBook)}
            >
              AI 判断发布信息
            </Button>
            {isBookUploadedTo121(viewingBook) ? (
              <Button
                danger
                onClick={() => {
                  setSelectedBookIds([viewingBook.id]);
                  setUploadMode("reupload");
                }}
              >
                重新上传
              </Button>
            ) : null}
            <Button
              onClick={() => setRawSourceBook(viewingBook)}
              disabled={!String(viewingBook?.txtText || "").trim()}
            >
              查看原始正文
            </Button>
            <Button onClick={() => openSourceEditor(viewingBook)}>
              编辑完整正文
            </Button>
            <Button onClick={() => openMetadataEditor(viewingBook)}>
              编辑列表信息
            </Button>
          </Space>
        ) : null
      }
      width={860}
      className="batch-factory-book-detail-modal"
    >
      {viewingBook
        ? (() => {
            const summary = stageSummaries[viewingBook.id];
            const state = batchFactoryBookState(viewingBook, {
              productionStatus,
              mergeStatus,
              stageSummary: summary,
              automationBook: (Array.isArray(automationStatus?.books) ? automationStatus.books : []).find((item) => item?.bookId === viewingBook.id),
            });
            const timeline = batchFactoryBookTimeline(viewingBook, {
              productionStatus,
              mergeStatus,
              stageSummaries: stageSummaries?.[viewingBook.id],
            });
            const uploadProgress = uploadProgressForBook(viewingBook);
            const uploadHistory = Array.isArray(
              viewingBook?.sourceMetadata?.websiteSubmitHistory,
            )
              ? viewingBook.sourceMetadata.websiteSubmitHistory
              : [];
            const classifyBusy = actionBusy === `classify-${viewingBook.id}`;
            const sensitiveDetail = sensitiveProcessingDetail(viewingBook);
            const classifyFailed =
              ["failed", "pending_confirmation"].includes(
                String(
                  viewingBook?.sourceMetadata?.classifyStatus || "",
                ).trim(),
              ) && !hasClassificationValues(viewingBook?.sourceMetadata);
            return (
              <div className="batch-factory-book-detail">
                <section className="batch-factory-book-detail-status">
                  <h3>流程异常与生成状态</h3>
                  <div>
                    <Tag color={state.tone}>{state.label}</Tag>
                    <strong>{state.detail}</strong>
                  </div>
                  {classifyBusy ? (
                    <Alert
                      type="info"
                      showIcon
                      message="正在识别男女频、风格和标签…"
                      description="识别完成后自动继续，不会打断生产流程。"
                    />
                  ) : null}
                  {classifyFailed ? (
                    <Alert
                      type="warning"
                      showIcon
                      message="AI 判断未完成"
                      description={classificationDetail(
                        viewingBook.sourceMetadata,
                      )}
                    />
                  ) : null}
                  {uploadProgress ? (
                    <Alert
                      type={
                        uploadProgress.status === "failed" ? "error" : "info"
                      }
                      showIcon
                      message={`121 上传：${uploadProgress.message || uploadProgress.phase}`}
                      description={`阶段：${uploadProgress.phase || "—"} · ${uploadProgress.updatedAt || "—"}`}
                    />
                  ) : null}
                  {state.failedStage?.errorMessage ? (
                    <BatchFactoryErrorNotice
                      stageText={({
                        audio_planning: '配音与逐行时长',
                        assets: '资产处理',
                        director: '文本提示词',
                        video: '视频生成',
                        merge: '视频合成',
                        comic: '漫画分页与提示词',
                      })[String(state.failedStage.stage || '').trim()] || String(state.failedStage.stage || '').trim()}
                      error={state.failedStage.errorMessage}
                    />
                  ) : null}
                  {state.label === "异常" ? (
                    <Button
                      danger
                      size="small"
                      loading={
                        actionBusy === stageActionKey("retry", viewingBook.id)
                      }
                      onClick={() => retryLastFailedStage(viewingBook)}
                    >
                      重试失败步骤
                    </Button>
                  ) : null}
                  <div className="batch-factory-production-timeline is-detail">
                    {timeline.map((item) => (
                      <div className={`is-${item.status}`} key={item.key}>
                        <i aria-hidden="true" />
                        <span>{item.label}</span>
                        <small>{item.detail}</small>
                      </div>
                    ))}
                  </div>
                </section>
                <section className="batch-factory-book-detail-content">
                  <h3>
                    小说正文{" "}
                    <small>
                      {effectiveBookSettings(batch, viewingBook)
                        .productionMode === "comic"
                        ? `正文总字数 ${batchFactoryCleanSourceText(viewingBook.sourceText).length} 字 · 漫画实际分析 ${comicAnalysisCharactersForBook(viewingBook)} 字`
                        : `正文总字数 ${batchFactoryCleanSourceText(viewingBook.sourceText).length} 字 · 视频实际生产 ${batchFactoryPreviewText(viewingBook.sourceText, viewingBook).split(/\r?\n/).filter(Boolean).length} 行`}
                    </small>
                  </h3>
                  <pre>
                    {viewingBook.workingFrontContent ||
                      batchFactoryCleanSourceText(viewingBook.sourceText) ||
                      "尚未获取原文。"}
                  </pre>
                </section>
                <Descriptions bordered size="small" column={2}>
                  <Descriptions.Item label="Book ID">
                    {viewingBook.bookId || "—"}
                  </Descriptions.Item>
                  <Descriptions.Item label="书城">
                    {bookPlatformName(viewingBook, platformNames)}
                  </Descriptions.Item>
                  <Descriptions.Item label="风格">
                    {value(viewingBook.sourceMetadata, "style")}
                  </Descriptions.Item>
                  <Descriptions.Item label="男女频">
                    {value(viewingBook.sourceMetadata, "gender")}
                  </Descriptions.Item>
                  <Descriptions.Item label="标签">
                    {value(viewingBook.sourceMetadata, "tags")}
                  </Descriptions.Item>
                  <Descriptions.Item label="AI 判断">
                    {classificationLabel(viewingBook.sourceMetadata)}
                  </Descriptions.Item>
                  <Descriptions.Item label="来源">
                    {value(viewingBook.sourceMetadata, "sourceMode") ===
                    "manual_original"
                      ? "手动书单"
                      : "小说获取"}
                  </Descriptions.Item>
                  <Descriptions.Item label="敏感处理">
                    {sensitiveDetail.status === 'failed' ? '处理失败' : sensitiveDetail.status === 'review_required' ? '待人工确认' : sensitiveDetail.status === 'partial' ? '部分处理' : sensitiveDetail.status ? '已完成' : '未处理'}
                  </Descriptions.Item>
                  <Descriptions.Item label="处理字数">
                    {sensitiveDetail.originalChars > 0 ? `${sensitiveDetail.processedChars}/${sensitiveDetail.originalChars} 字` : (sensitiveDetail.processedChars ? `${sensitiveDetail.processedChars} 字` : '—')}
                  </Descriptions.Item>
                  <Descriptions.Item label="敏感命中 / 替换">
                    {`${sensitiveDetail.hitCount} / ${sensitiveDetail.replacedCount} 处`}
                  </Descriptions.Item>
                  <Descriptions.Item label="处理链路" span={2}>
                    {sensitiveDetail.stages.length ? sensitiveDetail.stages.map(stage => `${stage.title || stage.stage}${stage.changed ? '（已处理）' : ''}`).join(' → ') : '—'}
                  </Descriptions.Item>
                </Descriptions>
                {sensitiveDetail.status ? (
                  <Alert
                    type={['failed', 'review_required', 'partial'].includes(sensitiveDetail.status) ? "warning" : "success"}
                    showIcon
                    message={`敏感处理：命中 ${sensitiveDetail.hitCount} 处，已替换 ${sensitiveDetail.replacedCount} 处${sensitiveDetail.status === 'review_required' ? '，有未处理命中待人工确认' : ''}`}
                    description={sensitiveDetail.items.length ? <Button size="small" onClick={() => setSensitiveDetailBook(viewingBook)}>查看替换明细</Button> : '未发现需替换内容。'}
                  />
                ) : null}
                {!classifyBusy ? (
                  <Alert
                    type={classifyFailed ? "warning" : "info"}
                    showIcon
                    message={`AI 判断：${classificationLabel(viewingBook.sourceMetadata)}`}
                    description={classificationDetail(
                      viewingBook.sourceMetadata,
                    )}
                  />
                ) : null}
                {uploadHistory.length ? (
                  <section className="batch-factory-book-upload-history">
                    <h3>121 上传历史</h3>
                    {uploadHistory
                      .slice()
                      .reverse()
                      .map((item, index) => (
                        <p key={`${item.submittedAt || "history"}-${index}`}>
                          {item.submittedAt || "—"} · ${item.status || "—"} · $
                          {item.detail || "—"}
                        </p>
                      ))}
                  </section>
                ) : null}
              </div>
            );
          })()
        : null}
    </Modal>;
    <Modal title={sensitiveDetailBook ? `敏感词替换明细 · ${sensitiveDetailBook.title || sensitiveDetailBook.bookId || ''}` : '敏感词替换明细'} open={Boolean(sensitiveDetailBook)} onCancel={() => setSensitiveDetailBook(null)} footer={null} width={860} destroyOnClose>
      {sensitiveDetailBook ? (() => {
        const detail = sensitiveProcessingDetail(sensitiveDetailBook);
        return <Space direction="vertical" size={12} style={{ width: '100%' }}>
          <Alert type={detail.status === 'failed' ? "warning" : "info"} showIcon message={`命中 ${detail.hitCount} 处，已替换 ${detail.replacedCount} 处`} description="仅展示命中词和替换结果，不展示原文上下文。" />
          {detail.items.length ? <Descriptions bordered size="small" column={1} items={detail.items.map((item, index) => ({ key: `${item.hit_index || index}-${item.keyword || ''}`, label: `#${index + 1} · ${item.keyword || '敏感词'}`, children: <Space direction="vertical" size={2}><span>替换后：{item.after || item.replace || '—'}</span><small>方式：{item.source || '规则'} · 状态：{item.status || '已处理'}</small></Space> }))} /> : <Alert type="info" showIcon message="未记录到可查看的替换明细" description="较早获取的书籍只保存了汇总数据；重新获取后会记录新的替换明细。" />}
        </Space>;
      })() : null}
    </Modal>
    <Modal title={rawSourceBook ? `原始正文 · ${rawSourceBook.title || rawSourceBook.bookId || ''}` : '原始正文'} open={Boolean(rawSourceBook)} onCancel={() => setRawSourceBook(null)} footer={null} width={860} destroyOnClose>
      {rawSourceBook ? <><Alert type="info" showIcon message="上游返回的原始正文" description={`原始正文 ${String(rawSourceBook.txtText || '').length} 字；生产正文 ${String(rawSourceBook.sourceText || '').length} 字。原始正文仅供查看，不会直接替代生产正文。`} /><pre className="batch-factory-book-detail-raw-source">{rawSourceBook.txtText || '暂无原始正文。'}</pre></> : null}
    </Modal>
    <Modal title={metadataBook ? `编辑列表信息 · ${metadataBook.title}` : '编辑列表信息'} open={Boolean(metadataBook)} onCancel={() => setMetadataBook(null)} onOk={saveBookMetadata} confirmLoading={metadataSaving} okText="保存" width={620}>{metadataBook ? <Space direction="vertical" size={12} style={{ width: '100%' }}><Alert type="info" showIcon message="此处保存风格、男女频、标签、推荐理由和评级" description="不会改动小说正文或书城来源。" />{[['style', '风格'], ['gender', '男女频'], ['tags', '标签'], ['reason', '推荐理由'], ['rating', '评级']].map(([key, label]) => <label key={key} className="batch-factory-engine-field"><span><b>{label}</b></span><Input value={metadataValue[key] || ''} onChange={event => setMetadataValue(current => ({ ...current, [key]: event.target.value }))} /></label>)}</Space> : null}</Modal>
    <Modal className="batch-factory-production-content-modal" title={editingContentBook ? `编辑内容 · ${editingContentBook.title}` : '编辑内容'} open={Boolean(editingContentBook)} onCancel={() => setEditingContentBook(null)} onOk={saveWorkingContent} confirmLoading={contentSaving} okText="保存编辑内容" width={820} destroyOnClose><Space direction="vertical" size={14} style={{ width: '100%' }}><div style={{ display: 'flex', gap: 12, alignItems: 'flex-start' }}><label className="shuihuo-form-label" style={{ flex: 1 }}><span>书名<small>发布时按书名匹配漫画后台作品，可手工纠正</small></span><Input value={editingContentTitle} onChange={event => setEditingContentTitle(event.target.value)} placeholder="与漫画后台作品相同的真实书名" /></label><label className="shuihuo-form-label" style={{ flex: 1 }}><span>书号<small>外部书城的数字书号</small></span><Input value={editingContentSourceBookId} onChange={event => setEditingContentSourceBookId(event.target.value.replace(/\D/g, '').slice(0, 20))} placeholder="如 794582" inputMode="numeric" maxLength={20} /></label></div><div style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>书号填错会导致正文抓取失败，修改后保存并重试即可重新抓取。</div><Alert type="info" showIcon message={editingContentProductionMode === 'comic' ? '这里编辑当前书的完整漫画正文' : '这里只展示和编辑当前书所选行数的内容'} description={editingContentProductionMode === 'comic' ? '保存后，漫画资产、漫画分页和每页提示词都会使用这里的完整内容。' : '保存后，列表预览、生产和上传 TXT 都使用这里的编辑内容；未选择的后续原文会继续保留。'} />{editingContentBook ? <BatchFactoryGiantMaterialPendingProgress book={editingContentBook} batchId={batch?.id} onContentReady={refreshBatch} /> : null}{editingContentProductionMode === 'video' ? <label className="shuihuo-form-label"><span>内容范围<small>本次生产使用的正文行数</small></span><InputNumber min={1} max={500} value={editingContentRangeLines} addonAfter="行" style={{ width: 180 }} onChange={value => { const next = Math.min(500, Math.max(1, Number(value) || 1)); setEditingContentRangeLines(next); if (!editingContentBook) return; const source = editingContentMode === 'custom' ? String(editingContentValue || '').split(/\r?\n/).filter(Boolean).slice(0, next).join('\n') : batchFactoryPreviewText(editingContentBook.sourceText, { sourceMetadata: { ...(editingContentBook.sourceMetadata || {}), contentRangeLines: next } }); setEditingContentValue(source); }} /></label> : <label className="shuihuo-form-label"><span>漫画目标字数<small>按完整句子、对白和段落边界取整</small></span><InputNumber min={100} max={100000} step={100} value={editingContentCaptureCharacters} addonAfter="字" style={{ width: 180 }} onChange={value => setEditingContentCaptureCharacters(Math.min(100000, Math.max(100, Number(value) || 100)))} /></label>}<Input.TextArea rows={16} value={editingContentValue} onChange={event => { setEditingContentValue(event.target.value); setEditingContentMode('custom'); }} placeholder={editingContentProductionMode === 'comic' ? '输入当前小说的完整漫画正文' : '输入当前小说的编辑内容'} /><label className="shuihuo-form-label"><span>衍生开篇</span><Select value={derivedOpeningPresetId || undefined} onChange={setDerivedOpeningPresetId} options={derivedOpeningOptions} loading={!derivedOpeningOptions.length} placeholder="选择已发布的衍生开篇提示词" style={{ width: 320 }} /></label>{!workingFrontCapability.available ? <Alert type="warning" showIcon message="当前不能生成爆款候选" description={workingFrontCapability.reason || '请先完成当前书的可执行配置。'} /> : null}<Space wrap><Tooltip title={workingFrontCapability.available ? '按选中的衍生开篇提示词生成候选；不会立即改动完整正文。' : workingFrontCapability.reason}><Button type="primary" onClick={createViralCandidate} loading={rewritingFront} disabled={!workingFrontCapability.available || !String(editingContentValue || '').trim() || !derivedOpeningPresetId}>生成爆款候选</Button></Tooltip><Tooltip title={workingFrontCapability.available ? '按当前选择的同一提示词重新生成候选；不会立即改动完整正文。' : workingFrontCapability.reason}><Button onClick={createViralCandidate} loading={rewritingFront} disabled={!workingFrontCapability.available || !String(editingContentValue || '').trim() || !derivedOpeningPresetId}>重试生成爆款候选</Button></Tooltip></Space>{viralCandidate ? <Alert type="warning" showIcon message="爆款候选尚未录用" description={<Space direction="vertical" size={8} style={{ width: '100%' }}><pre className="batch-factory-viral-candidate">{viralCandidate}</pre><Space><Button type="primary" loading={sourceSaving} onClick={adoptViralCandidate}>录用并替换完整正文</Button><Button onClick={cancelViralCandidate}>取消候选</Button></Space></Space>} /> : null}</Space></Modal>
    <Modal title={sourceEditorBook ? `编辑完整正文 · ${sourceEditorBook.title}` : '编辑完整正文'} open={Boolean(sourceEditorBook)} onCancel={() => setSourceEditorBook(null)} onOk={() => saveSourceEditor()} confirmLoading={sourceSaving} okText="保存完整正文" width={860} destroyOnClose>
      <Space direction="vertical" size={14} style={{ width: '100%' }}>
        <Alert type="info" showIcon message="编辑的是完整获取正文" description="保存后会替换当前书原文（最多 4000 字）；工作区卡片和工作区编辑始终只显示你选择的前几行。" />
        <Button onClick={() => downloadProductionContent({ title: sourceEditorBook?.title, content: sourceEditorValue })} disabled={!String(sourceEditorValue || '').trim()}>下载完整正文 TXT</Button>
        <Input.TextArea rows={20} value={sourceEditorValue} onChange={event => setSourceEditorValue(event.target.value)} placeholder="完整获取正文" />
      </Space>
    </Modal>
    <BatchFactoryBookSettingsModal open={Boolean(configTarget)} batch={batch} book={configTarget?.book} activeRegion={configTarget?.region} onClose={() => setConfigTarget(null)} onSaved={refreshAfterBookSettingsSaved} onOpenBookAssets={book => setAssetBook(book)} />
    <BatchFactoryPreflightModal
      open={automationPreflightOpen}
      loading={automationPreflightLoading}
      result={automationPreflightResult}
      error={automationPreflightError}
      submitting={automationBusy === 'start'}
      onClose={() => setAutomationPreflightOpen(false)}
      onRun={confirmPreflightRun}
      onFix={handleAutomationPreflightFix}
    />
    <Modal title={assetBook ? `人物场景预设 · ${assetBook.title}` : '人物场景预设'} open={Boolean(assetBook)} onCancel={() => setAssetBook(null)} footer={null} width="min(1440px, calc(100vw - 48px))" className="batch-factory-assets-modal">{assetBook ? (() => { const settings = effectiveBookSettings(batch, assetBook); const productionMode = settings.productionMode === 'comic' ? 'comic' : 'video'; const comicRules = effectiveBookComicRules(batch, assetBook); return <AssetEditor book={assetBook} batchId={batch?.id} onSaved={refreshBatch} onGenerate={async textModelId => { await refreshAssetPresetSnapshot(assetBook); await runBookStageAction(assetBook, 'assets', 'missing', '', textModelId); }} onRegenerate={async textModelId => { await refreshAssetPresetSnapshot(assetBook); await runBookStageAction(assetBook, 'assets', 'force', '', textModelId); }} onRetry={textModelId => retryLastFailedStage(assetBook, '', textModelId)} productionMode={productionMode} assetRules={productionMode === 'comic' ? comicRules : effectiveBookAssetRules(batch, assetBook)} onAssetPromptChange={async selection => { const bookPromptConfig = assetBook?.settingsState?.patch?.aiPromptConfig || {}; const assetRules = effectiveBookAssetRules(batch, assetBook); await saveBookOverrideWithRetry(batch.id, assetBook.id, assetBook.revision, { patch: { aiPromptConfig: productionMode === 'comic' ? { ...bookPromptConfig, comic: { ...comicRules, assets: selection, enabled: true, scope: 'custom', bookIds: [assetBook.id] } } : { ...bookPromptConfig, assets: { ...assetRules, extraction: selection, scope: 'custom', bookIds: [assetBook.id] } } } }); await refreshBatch(); }} canGenerate={runCapability.available} generateReason={runCapability.reason} generating={actionBusy === stageActionKey('assets', assetBook.id)} engineSettings={settings} />; })() : null}</Modal>
    <Modal title={promptBook ? `分镜卡（${effectiveBookSettings(batch, promptBook).productionMode === 'comic' ? '漫画提示词' : '视频提示词'}） · ${promptBook.title}` : '分镜卡'} open={Boolean(promptBook)} onCancel={() => { setPromptBook(null); setPromptVideoId(''); }} footer={null} width={900} className="batch-factory-prompt-modal">{promptBook ? <PromptPanel key={`${promptBook.id}:${promptBook.revision || 0}:${promptBook.directorRevision?.id || ''}`} book={promptBook} batch={batch} batchId={batch?.id} settingsRevision={batch?.settingsState?.revision} productionMode={effectiveBookSettings(batch, promptBook).productionMode === 'comic' ? 'comic' : 'video'} initialVideoId={promptVideoId} isDeveloper={isDeveloper} onSaved={async () => { const latestBatch = await refreshBatch(); const latestBook = latestBatch?.books?.find(item => item.id === promptBook?.id); if (latestBook) setPromptBook(latestBook); return latestBook; }} onCompile={() => runBookStageAction(promptBook, 'director', 'compile')} onRegenerateVisual={() => runBookStageAction(promptBook, 'comic', 'force')} onRegenerateAllVisual={() => runBookStageAction(promptBook, 'comic', 'force')} onRetry={videoId => retryLastFailedStage(promptBook, videoId)} onGenerateVideo={videoId => runBookStageAction(promptBook, 'video', 'missing', videoId)} onGenerateVisual={(videoId, prompt) => generateBookComicPage(promptBook, videoId, prompt)} onDownloadAllComicZip={() => downloadBookComicZip(promptBook)} comicZipDownloading={actionBusy === `comic-zip-${promptBook.id}`} regenerating={actionBusy === stageActionKey('director', promptBook.id) || actionBusy === stageActionKey('video', promptBook.id) || actionBusy === stageActionKey('comic', promptBook.id) || actionBusy === stageActionKey('retry', promptBook.id) || String(actionBusy).startsWith(`comic-image-${promptBook.id}-`)} productionAvailable={productionCapability.available} productionReason={productionCapability.reason} onGenerateOpening={() => runBookStageAction(promptBook, 'opening', 'force')} generatingOpening={actionBusy === stageActionKey('opening', promptBook.id)} /> : null}</Modal>
	<Modal
      title={mediaBook && effectiveBookSettings(batch, mediaBook).productionMode === 'comic' ? null : (mediaBook ? `媒体库 · ${mediaBook.title}` : '媒体库')}
      open={Boolean(mediaBook)}
      onCancel={() => { setMediaBook(null); setMediaVideoId(''); setMediaStartTab('clips'); }}
      footer={null}
      width="min(1180px, calc(100vw - 40px))"
      closable={!mediaBook || effectiveBookSettings(batch, mediaBook).productionMode !== 'comic'}
      className={`batch-factory-media-library-modal${mediaBook && effectiveBookSettings(batch, mediaBook).productionMode === 'comic' ? ' is-comic' : ''}`}
      destroyOnClose
    >
	  {mediaBook ? (effectiveBookSettings(batch, mediaBook).productionMode === 'comic' ? <ComicMediaPanel batchId={batch?.id} book={mediaBook} refreshNonce={comicPageRefreshNonce} comicPublishSettings={effectiveBookSettings(batch, mediaBook).comicPublishSettings || {}} initialVideoId={mediaVideoId} onSaved={refreshBatch} onGeneratePage={(videoId, prompt) => generateBookComicPage(mediaBook, videoId, prompt)} generatingPage={Array.from(comicTaskKeys).some(key => key.startsWith(`comic-image-${mediaBook.id}-`))} onClose={() => { setMediaBook(null); setMediaVideoId(''); setMediaStartTab('clips'); }} /> : <MediaVersionPanel
        book={mediaBook}
        batchId={batch?.id}
        productionStatus={productionStatus}
        mergeJob={latestBookMerge(mergeStatus, mediaBook.id)}
        mergeJobs={bookMergeJobs(mergeStatus, mediaBook.id)}
        mergeSettings={effectiveBookSettings(batch, mediaBook)}
        mergeSpeed={mergeSpeed}
        onMergeSpeedChange={setMergeSpeed}
        onMerge={options => runBookMerge(mediaBook, options)}
        mergeAvailable={mergeCapability.available}
        mergeReason={mergeCapability.reason}
        merging={actionBusy === `merge-${mediaBook.id}`}
        onSaved={async () => { await refreshBatch(); await loadRuntimeStatus({ quiet: true }); }}
        onDeleted={async () => { await refreshBatch(); await loadRuntimeStatus({ quiet: true }); }}
        onUpload={() => { setSelectedBookIds([mediaBook.id]); setMediaBook(null); setMediaVideoId(''); setMediaStartTab('clips'); setUploadMode('selected'); }}
        onRegenerate={videoId => runBookStageAction(mediaBook, 'video', 'force', videoId)}
        onRetry={videoId => retryLastFailedStage(mediaBook, videoId)}
        regenerating={actionBusy === stageActionKey('video', mediaBook.id) || actionBusy === stageActionKey('retry', mediaBook.id)}
        productionAvailable={productionCapability.available}
        productionReason={productionCapability.reason}
        initialVideoId={mediaVideoId}
        initialTab={mediaStartTab}
        onOpenDirector={frameKey => { setPendingMediaPrompt({ bookId: mediaBook.id, frameKey: frameKey || '' }); setMediaBook(null); }}
      />) : null}
    </Modal>

    <BatchFactoryUnifiedSettingsModal open={unifiedSettingsOpen} batch={batch} onClose={() => setUnifiedSettingsOpen(false)} onSaved={saveSettings} onProductionModeChange={saveProductionMode} />
    <Modal title={batchProductionMode === 'comic' ? '漫画任务 / 日志' : '任务 / 日志'} open={logsOpen} onCancel={() => setLogsOpen(false)} footer={<Button onClick={() => loadRuntimeStatus()}>刷新日志</Button>} width={860}><BatchLogs books={books} automationStatus={automationStatus} productionStatus={productionStatus} mergeStatus={mergeStatus} error={logsError} productionMode={batchProductionMode} syncing={logsLoading} syncedAt={logsSyncedAt} onRetry={() => loadRuntimeStatus()} /></Modal>
    <Modal className="batch-factory-automation-modal" title={<div className="batch-factory-automation-title"><strong>自动生产</strong><span>批量执行设置</span></div>} open={automationStartOpen} onCancel={() => setAutomationStartOpen(false)} width={680} destroyOnClose footer={<div className="batch-factory-automation-footer"><Button onClick={() => setAutomationStartOpen(false)}>取消</Button><Button onClick={() => { setAutomationStartOpen(false); setUnifiedSettingsOpen(true); }}>统一配置</Button><span /><Button loading={contentSettingsSaving} onClick={saveContentSettings}>{batchProductionMode === 'comic' ? '保存正文截取设置' : '保存内容设置'}</Button><Tooltip title={automationActive ? automationStartBlockedLabel : ''}><span><Button loading={automationBusy === 'start'} disabled={automationActive} onClick={confirmImmediateAutomation}>{automationActive ? automationStartButtonLabel : '立即执行'}</Button></span></Tooltip><Button type="primary" loading={automationBusy === 'start'} disabled={automationActive} onClick={() => runAutomationAction('start')}>{automationActive ? '任务已存在' : '保存定时任务'}</Button></div>}>
      <div className="batch-factory-automation-form">
        {automationActive && <Alert type="warning" showIcon message={automationStartBlockedLabel} />}
        <section className="batch-factory-automation-section"><div className="batch-factory-automation-section-head"><span className="batch-factory-automation-section-mark">01</span><div><b>内容设置</b><small>{automationBookScope === 'selected' ? `应用到已选的 ${selectedBookIds.length} 本小说` : '应用到当前批量的全部小说'}</small></div></div><div className="batch-factory-automation-grid">{batchProductionMode === 'video' ? <label className="batch-factory-engine-field"><span><b>内容范围</b><small>生产使用的正文行数</small></span><InputNumber min={1} max={500} value={contentRangeLines} onChange={value => setContentRangeLines(value || 5)} addonAfter="行" style={{ width: '100%' }} /></label> : <label className="batch-factory-engine-field"><span><b>漫画分析</b><small>按完整句子和段落边界取整 · 预计约 {comicEstimatedPageRange(comicAnalysisCharacters).min}-{comicEstimatedPageRange(comicAnalysisCharacters).max} 张</small></span><InputNumber min={100} max={100000} step={100} value={comicAnalysisCharacters} onChange={value => setComicAnalysisCharacters(Math.min(100000, Math.max(100, Number(value) || 2000)))} addonAfter="字" style={{ width: '100%' }} /></label>}<label className="batch-factory-engine-field"><span><b>内容截取</b><small>送入本地处理链的正文上限</small></span><Select value={contentCaptureCharacters} onChange={setContentCaptureCharacters} style={{ width: '100%' }} options={[1000, 2000, 4000, 8000, 12000, 20000, 50000, 100000].map(value => ({ value, label: `${value} 字` }))} /></label></div></section>
        <section className="batch-factory-automation-section"><div className="batch-factory-automation-section-head"><span className="batch-factory-automation-section-mark">02</span><div><b>生产策略</b><small>选择预设和生成范围</small></div></div><div className="batch-factory-automation-grid"><label className="batch-factory-engine-field"><span><b>自动化预设</b></span><Select value={automationPresetID || undefined} onChange={setAutomationPresetID} placeholder="选择已保存预设" options={automationPresets.map(item => ({ value: item.id, label: `${item.name} · v${item.version}` }))} style={{ width: '100%' }} /></label><label className="batch-factory-engine-field"><span><b>执行模式</b></span><Select value={automationRunMode} onChange={value => { setAutomationRunMode(value); if (value !== 'full_submit') setAutomationScheduleMode('execute'); }} style={{ width: '100%' }} options={batchProductionMode === 'comic' ? [{ value: 'storyboard_only', label: '只生成漫画分页与提示词' }, { value: 'video_no_submit', label: '生成漫画合集不上传' }, { value: 'full_submit', label: '全自动生成并提交' }] : [{ value: 'storyboard_only', label: '只生成分镜' }, { value: 'video_no_submit', label: '生成视频不提交' }, { value: 'full_submit', label: '全自动生成并提交' }]} /></label><label className="batch-factory-engine-field"><span><b>执行书籍</b></span><Select value={automationBookScope} onChange={setAutomationBookScope} style={{ width: '100%' }} options={[{ value: 'all', label: `全部小说（${books.length} 本）` }, { value: 'selected', label: `已选小说（${selectedBookIds.length} 本）`, disabled: !selectedBookIds.length }]} /></label><label className="batch-factory-engine-field"><span><b>重复执行保护</b><small>默认复用已经生成的资产和视频</small></span><Switch checked={automationForceRegenerate} onChange={setAutomationForceRegenerate} checkedChildren="强制重做" unCheckedChildren="复用已有" /></label></div></section>
        <section className="batch-factory-automation-section"><div className="batch-factory-automation-section-head"><span className="batch-factory-automation-section-mark">03</span><div><b>启动/提交方式</b><small>立即开始，或按北京时间执行/提交</small></div></div><div className="batch-factory-automation-grid"><label className="batch-factory-engine-field"><span><b>自动启动时间</b></span><Input type="datetime-local" value={automationScheduledAt} onChange={event => setAutomationScheduledAt(event.target.value)} /></label><label className="batch-factory-engine-field"><span><b>定时方式</b></span><Select value={automationScheduleMode} onChange={setAutomationScheduleMode} style={{ width: '100%' }} options={[{ value: 'execute', label: '定时执行（到点才开始生产）' }, { value: 'submit', label: '定时提交（提前生产，到点上传/发布）', disabled: automationRunMode !== 'full_submit' }]} /></label><label className="batch-factory-engine-field"><span><b>同时处理书籍</b></span><Select value={automationConcurrency} onChange={value => setAutomationConcurrency(normalizeAutomationConcurrency(value))} style={{ width: '100%' }} options={[1, 2, 4].map(value => ({ value, label: `${value} 本并行` }))} /></label></div></section>
      </div>
    </Modal>
    <Modal title="批量发布漫画" open={comicBatchPublishOpen} onCancel={() => setComicBatchPublishOpen(false)} footer={null} width={720} destroyOnClose>
      <Space direction="vertical" style={{ width: '100%' }}>
        <Alert type="info" showIcon message="按书发布漫画合集" description="每本书只会发布已生成合集中的主图。进入后可检查制作者账号、音频分类和关联作品，再手动确认发布。" />
        {books.filter(book => selectedBookIds.includes(book.id)).map(book => <Card key={book.id} size="small"><Space style={{ width: '100%', justifyContent: 'space-between' }}><span>{book.title || book.bookId || book.id}</span><Button type="primary" onClick={() => { setMediaBook(book); setMediaStartTab('comic'); setComicBatchPublishOpen(false); }}>打开合集发布</Button></Space></Card>)}
      </Space>
    </Modal>
    <UploadNetwork batch={batch} books={books} selectedBookIds={selectedBookIds} productionStatus={productionStatus} mergeStatus={mergeStatus} mode={uploadMode} onClose={() => setUploadMode('')} onOpenPublish={() => { setUploadMode(''); setUnifiedSettingsOpen(true); }} onRefresh={refreshBatch} />
  </section>;
}
