import { Alert, Button, Input, InputNumber, Modal, Select, Space, Switch, Tabs, Tag, message } from 'antd';
import { useEffect, useMemo, useRef, useState } from 'react';
import { getWorkshopPlatforms } from '../../../shared/api/novelFetchWorkshop';
import { apiErrorMessage } from '../../../shared/apiErrors.js';
import { classifyBookPublishMetadata, createNovelFetchIntake, fetchBookOriginal, fetchDirectOriginals, getBatch, getBatchAutomationStatus, listAutomationPresets, listBatches, startBatchAutomation, updateBookMetadata } from '../../../shared/api/batchFactory';
import { batchFactoryPlatformOptions } from './batchFactoryPlatformOptions';
import { CUSTOM_COPY_OPTION, selectedCopyGroups } from './batchFactoryCustomCopy.js';
import { BatchFactoryCustomCopyPicker } from './BatchFactoryCustomCopyPicker.jsx';
import { getToken } from '../../../shared/api/client.js';
import { buildManualBatchSubmission, manualBookIDsFromInput, manualFetchedSourceText, removePlatformGroup, replacePlatformGroup, totalGroupBookCount, upsertPlatformGroup } from './batchFactoryManualFetch';
import { automationPresetSnapshot, formatBeijingDatetimeLocal, normalizeAutomationConcurrency, parseBeijingDatetimeLocal } from './batchFactoryAutomationSchedule';
import { resolveGiantMaterialForBatch } from '../giantMaterialExtractionClient.js';
import { duplicateGiantMaterialBookIndexes, parseGiantMaterialIds } from './batchFactoryGiantMaterialQueue.js';
import { createGiantMaterialJob, listGiantMaterialExecutors } from '../../../shared/api/giantMaterialExecutorPublic.js';
import { BatchFactoryGiantMaterialExecutorStatus } from './BatchFactoryGiantMaterialExecutorStatus.jsx';
import { comicEstimatedPageRange } from './batchFactoryContentRange.js';
import { BatchFactoryAiReasoningForm, BatchFactoryAutomationPresetManager, BatchFactoryComicPublishSettingsForm, BatchFactoryEngineSettingsForm, BatchFactoryProductionModeControl, BatchFactoryPublishSettingsForm } from './BatchFactoryUnifiedSettingsModal.jsx';
import {
  availableGiantMaterialBooks,
  buildGiantMaterialPlaceholderIntakes,
  findRegisteredGiantMaterialBook,
  giantMaterialBookKey,
  selectGiantMaterialBook
} from './batchFactoryGiantMaterialImport.js';

// 输入格式与列顺序固定为默认值（智能识别 + 完整元数据列），AI 会自动分析男女频/风格，无需用户手选
const DEFAULT_PARSE_MODE = 'smart';
const DEFAULT_COLUMN_PRESET_ID = 'full_metadata';
const DEFAULT_COLUMN_ORDER = '书籍ID,书名,男女频,风格,标签,推荐理由,评级';
const clone = value => JSON.parse(JSON.stringify(value || {}));

function normalizedError(error, fallback) {
  return apiErrorMessage(error, fallback || '请求失败');
}

const GIANT_MATERIAL_PLATFORM_OPTION = { value: 'giant_material', label: '巨量获取' };

function withGiantMaterialOption(options = []) {
  return [GIANT_MATERIAL_PLATFORM_OPTION, CUSTOM_COPY_OPTION, ...(Array.isArray(options) ? options : []).filter(option => ![GIANT_MATERIAL_PLATFORM_OPTION.value, CUSTOM_COPY_OPTION.value].includes(option?.value))];
}

function preferredPlatformId(options, current = '') {
  if (Array.isArray(options) && options.some(option => option?.value === current)) return current;
  return options.find(option => ![GIANT_MATERIAL_PLATFORM_OPTION.value, CUSTOM_COPY_OPTION.value].includes(option?.value))?.value || options[0]?.value || '';
}

function fallbackPlatformOptions() {
  return withGiantMaterialOption(batchFactoryPlatformOptions([], { fallback: true }));
}

export function BatchFactoryCreateModal({ open, onCancel, onCreated, onBatchUpdated }) {
  const [title, setTitle] = useState('');
  const [platformId, setPlatformId] = useState('');
  const [platformOptions, setPlatformOptions] = useState([]);
  const [platformState, setPlatformState] = useState('idle');
  const [platformError, setPlatformError] = useState('');
  const parseMode = DEFAULT_PARSE_MODE;
  const columnPresetId = DEFAULT_COLUMN_PRESET_ID;
  const columnOrder = DEFAULT_COLUMN_ORDER;
  const [inputText, setInputText] = useState('');
  const [scheduleDialogOpen, setScheduleDialogOpen] = useState(false);
  const [automationDialogMode, setAutomationDialogMode] = useState('scheduled');
  const [scheduleTasksOpen, setScheduleTasksOpen] = useState(false);
  const [scheduleTasks, setScheduleTasks] = useState([]);
  const [scheduleTasksLoading, setScheduleTasksLoading] = useState(false);
  const [scheduledAt, setScheduledAt] = useState('');
  const [automationScheduleMode, setAutomationScheduleMode] = useState('execute');
  const [automationPresets, setAutomationPresets] = useState([]);
  const [automationPresetID, setAutomationPresetID] = useState('');
  const [automationRunMode, setAutomationRunMode] = useState('full_submit');
  const [automationConcurrency, setAutomationConcurrency] = useState(4);
  const [contentRangeLines, setContentRangeLines] = useState(5);
  const [comicAnalysisCharacters, setComicAnalysisCharacters] = useState(2000);
  const [contentCaptureCharacters, setContentCaptureCharacters] = useState(4000);
  const [busy, setBusy] = useState(false);
  const [groups, setGroups] = useState([]);
  const [editingPlatformId, setEditingPlatformId] = useState(null);
  const [giantItems, setGiantItems] = useState([]);
  const [giantBusy, setGiantBusy] = useState(false);
  const [giantError, setGiantError] = useState('');
  const [executorHealth, setExecutorHealth] = useState(null);
  const [giantOriginalReadStrategy, setGiantOriginalReadStrategy] = useState('ocr_first');
  const [initialBatchSettings, setInitialBatchSettings] = useState({});
  const [initialSettingsOpen, setInitialSettingsOpen] = useState(false);
  const giantControllerRef = useRef(null);
  const customCopySessionRef = useRef('');

  async function loadPlatforms() {
    setPlatformState('loading');
    setPlatformError('');
    try {
      const result = await getWorkshopPlatforms();
      const options = withGiantMaterialOption(batchFactoryPlatformOptions(result?.platforms));
      if (!options.length) throw new Error('在线书城目录为空');
      setPlatformOptions(options);
      setPlatformId(current => preferredPlatformId(options, current));
      setPlatformState('ready');
    } catch (error) {
      const options = fallbackPlatformOptions();
      setPlatformOptions(options);
      setPlatformId(current => preferredPlatformId(options, current));
      setPlatformState('fallback');
      setPlatformError(normalizedError(error, '无法读取在线书城目录'));
    }
  }

  useEffect(() => {
    if (!open) return undefined;
    let active = true;
    setPlatformState('loading');
    setPlatformError('');
    getWorkshopPlatforms().then(result => {
      if (!active) return;
      const options = withGiantMaterialOption(batchFactoryPlatformOptions(result?.platforms));
      if (!options.length) throw new Error('在线书城目录为空');
      setPlatformOptions(options);
      setPlatformId(current => preferredPlatformId(options, current));
      setPlatformState('ready');
    }).catch(error => {
      if (!active) return;
      const options = fallbackPlatformOptions();
      setPlatformOptions(options);
      setPlatformId(current => preferredPlatformId(options, current));
      setPlatformState('fallback');
      setPlatformError(normalizedError(error, '无法读取在线书城目录'));
    });
    return () => { active = false; };
  }, [open]);

  const isGiantMaterial = platformId === GIANT_MATERIAL_PLATFORM_OPTION.value;
  const isCustomCopy = platformId === CUSTOM_COPY_OPTION.value;
  const isComicProduction = initialBatchSettings.productionMode === 'comic';
  const inputBookIds = useMemo(() => manualBookIDsFromInput(inputText), [inputText]);
  const giantDuplicateIndexes = useMemo(() => duplicateGiantMaterialBookIndexes(giantItems.map(item => (
    item.status === 'resolved' ? selectGiantMaterialBook(item.material, item.selectedBookKey) : null
  ))), [giantItems]);

  function reset() {
    customCopySessionRef.current = '';
    giantControllerRef.current?.abort();
    giantControllerRef.current = null;
    // 创建请求被网关中断后 busy 可能停在 true，表单关闭时不清理会让
    // “立即执行/创建”按钮一直转圈且点不动，只能刷新页面才能恢复。
    setBusy(false);
    setGiantItems([]);
    setGiantBusy(false);
    setGiantError('');
    setExecutorHealth(null);
    setGiantOriginalReadStrategy('ocr_first');
    setTitle('');
    setInputText('');
    setScheduleDialogOpen(false);
    setScheduleTasksOpen(false);
    setScheduleTasks([]);
    setScheduledAt('');
    setAutomationScheduleMode('execute');
    setAutomationDialogMode('scheduled');
    setAutomationPresetID('');
    setAutomationRunMode('full_submit');
    setAutomationConcurrency(2);
    setContentRangeLines(5);
    setComicAnalysisCharacters(2000);
    setContentCaptureCharacters(4000);
    setGroups([]);
    setEditingPlatformId(null);
    setInitialBatchSettings({});
    setInitialSettingsOpen(false);
  }

  function clearGiantMaterialDraft() {
    giantControllerRef.current?.abort();
    giantControllerRef.current = null;
    setGiantItems([]);
    setGiantBusy(false);
    setGiantError('');
  }

  function changePlatform(value) {
    if (editingPlatformId && value !== editingPlatformId) {
      message.warning('请先完成当前修改：点“添加书城”保存，或点当前标签的 × 放弃');
      return;
    }
    setPlatformId(value);
    if ([GIANT_MATERIAL_PLATFORM_OPTION.value, CUSTOM_COPY_OPTION.value].includes(value)) {
      setEditingPlatformId(null);
      setInputText('');
      clearGiantMaterialDraft();
    }
  }

  function handleAddPlatformGroup() {
    const id = String(platformId || '').trim();
    if (!id) return message.warning('请先选择书城');
    const ids = manualBookIDsFromInput(inputText);
    if (!ids.length) return message.warning('小说列表中没有有效的 Book ID（每行以书 ID 开头）');
    // 非编辑态向已存在书城追加时，按模拟合并后的预计数判断，避免同 ID 被重复计数
    const currentEditingOld = editingPlatformId
      ? manualBookIDsFromInput(groups.find(group => group.platformId === editingPlatformId)?.inputText).length
      : 0;
    let projected;
    if (editingPlatformId) {
      projected = totalGroupBookCount(groups) - currentEditingOld + ids.length;
    } else {
      const merged = upsertPlatformGroup(groups, { platformId, platformName: '', inputText });
      projected = totalGroupBookCount(merged);
    }
    if (projected > 50) {
      return message.warning('一个批量最多 50 本书');
    }
    if (editingPlatformId) {
      setGroups(current => replacePlatformGroup(current, editingPlatformId, inputText));
      setEditingPlatformId(null);
    } else {
      const platformName = platformOptions.find(option => option.value === platformId)?.label || platformId;
      setGroups(current => upsertPlatformGroup(current, { platformId, platformName, inputText }));
    }
    setInputText('');
  }

  function handleEditGroup(group) {
    // 编辑中点自己的标签：保持现状，不重载文本造成静默回滚
    if (editingPlatformId === group.platformId) return;
    if (editingPlatformId && editingPlatformId !== group.platformId) {
      return message.warning('请先完成当前修改：点“添加书城”保存，或点当前标签的 × 放弃');
    }
    setPlatformId(group.platformId);
    setInputText(group.inputText);
    setEditingPlatformId(group.platformId);
  }

  function handleRemoveGroup(removedPlatformId) {
    if (editingPlatformId === removedPlatformId) {
      // 编辑中点自己的 × = 放弃修改
      setEditingPlatformId(null);
      setInputText('');
      return;
    }
    if (editingPlatformId) {
      return message.warning('请先完成当前修改：点“添加书城”保存，或点当前标签的 × 放弃');
    }
    setGroups(current => removePlatformGroup(current, removedPlatformId));
  }

  const GIANT_ITEM_ERRORS = {
    INVALID_GIANT_MATERIAL_ID: 'ID 格式不对，应为 10—25 位数字',
    QINGYU_AUTH_NOT_CONFIGURED: '巨量素材解析服务未配置：请配置服务器插件地址与令牌',
    QINGYU_AUTH_FAILED: '巨量素材解析服务授权失效：请更新青语 N8 令牌',
    QINGYU_MATERIAL_RESPONSE_INVALID: '青语服务没有返回有效素材',
    QINGYU_BOOK_METADATA_INCOMPLETE: '素材没有完整的平台书名或 Book ID',
    QINGYU_VIDEO_DURATION_MISSING: '素材没有有效的视频时长'
  };

  function giantItemErrorText(code) {
    return GIANT_ITEM_ERRORS[code] || '解析失败，可重试';
  }

  const giantConfigBlockingCodes = new Set(['QINGYU_AUTH_FAILED', 'QINGYU_AUTH_NOT_CONFIGURED', 'GIANT_PLUGIN_NOT_CONFIGURED']);

  async function resolveAllGiantMaterials() {
    const { valid, invalid } = parseGiantMaterialIds(inputText);
    if (!valid.length && !invalid.length) return message.warning('请输入巨量素材 ID（每行一个）');
    giantControllerRef.current?.abort();
    const controller = new AbortController();
    giantControllerRef.current = controller;
    setGiantItems([
      ...valid.map(id => ({ id, status: 'pending', error: '', material: null, books: [], selectedBookKey: '' })),
      ...invalid.map(token => ({ id: token, status: 'error', error: 'INVALID_GIANT_MATERIAL_ID', material: null, books: [], selectedBookKey: '' }))
    ]);
    setGiantBusy(true);
    setGiantError('');
    try {
      for (const id of valid) {
        if (controller.signal.aborted) break;
        try {
          const material = await resolveGiantMaterialForBatch(id, { signal: controller.signal });
          const books = availableGiantMaterialBooks(material);
          if (!books.length) throw new Error('QINGYU_BOOK_METADATA_INCOMPLETE');
          setGiantItems(current => current.map(row => row.id === id
            ? { ...row, status: 'resolved', material, books, selectedBookKey: giantMaterialBookKey(books[0]) }
            : row));
        } catch (error) {
          const code = String(error?.message || error?.code || 'QINGYU_UPSTREAM_FAILED');
          if (giantConfigBlockingCodes.has(code)) {
            setGiantError(code);
            break;
          }
          setGiantItems(current => current.map(row => row.id === id
            ? { ...row, status: 'error', error: code }
            : row));
        }
      }
    } finally {
      setGiantBusy(false);
    }
  }

  async function queueGiantOcrFallback(batchId, entry, giantAutomationPlan, directReadError = '') {
    const executorResponse = await listGiantMaterialExecutors();
    const executors = Array.isArray(executorResponse?.executors)
      ? executorResponse.executors
      : Array.isArray(executorResponse?.data?.executors) ? executorResponse.data.executors : [];
    if (!executors.some(executor => executor?.online === true)) {
      throw new Error('GIANT_EXECUTOR_OFFLINE');
    }
    const durationSeconds = Number(entry.item.material?.durationSeconds || entry.item.material?.duration || 0);
    if (!(durationSeconds > 0)) throw new Error('QINGYU_VIDEO_DURATION_MISSING');
    const createdResponse = await createGiantMaterialJob({
      materialId: entry.item.id,
      platformBookId: entry.book.platformBookId || entry.book.bookId,
      title: entry.book.title,
      videoUrl: entry.item.material.videoUrl,
      durationSeconds,
      modelVersion: 'windows-paddleocr-v1',
      contentRangeLines
    });
    const createdJob = createdResponse?.job || createdResponse?.data?.job || createdResponse;
    if (!createdJob?.id) throw new Error('GIANT_EXECUTOR_FAILED');

    let latestError;
    for (let attempt = 0; attempt < 3; attempt += 1) {
      try {
        const latestResponse = await getBatch(batchId);
        const latestBatch = latestResponse?.batch || latestResponse?.data?.batch || latestResponse?.data || latestResponse;
        const latestBook = findRegisteredGiantMaterialBook(latestBatch?.books, entry.item.id);
        if (!latestBook?.id) throw new Error('GIANT_MATERIAL_BOOK_NOT_FOUND');
        await updateBookMetadata(batchId, latestBook.id, {
          metadata: {
            ...(latestBook.sourceMetadata || {}),
            executorJobId: createdJob.id,
            contentPending: true,
            originalReadStage: 'ocr',
            originalReadError: '',
            directOriginalReadError: directReadError,
            giantOcrState: String(createdJob.state || 'queued').toLowerCase(),
            giantAutomationPlan
          },
          expectedRevision: Number(latestBook.revision || 0)
        });
        return createdJob;
      } catch (bindError) {
        latestError = bindError;
        if (Number(bindError?.status) !== 409) break;
      }
    }
    throw latestError || new Error('GIANT_EXECUTOR_FAILED');
  }

  async function submitGiantMaterial({ scheduledRun = false, automationRun = false } = {}) {
    if (!title.trim()) return message.warning('请填写作品名称');
    const items = giantItems.filter(item => item.status === 'resolved');
    if (!items.length) return message.warning('请先点击“解析全部”并等待解析完成');
    if (scheduledRun && !scheduledAt) return message.warning('请选择北京时间自动启动时间');
    if ((scheduledRun || automationRun) && !automationPresetID) return message.warning('请选择自动化预设');
    const scheduledAtISO = scheduledRun ? parseBeijingDatetimeLocal(scheduledAt) : '';
    if (scheduledRun && (!scheduledAtISO || new Date(scheduledAtISO).getTime() <= Date.now())) return message.warning('北京时间自动启动时间需要晚于现在');
    if (giantOriginalReadStrategy !== 'direct_first' && executorHealth?.online !== true) return message.warning('巨量素材执行器未安装或未登录，请先在执行器配置页完成配对');
    const missingDuration = giantOriginalReadStrategy !== 'direct_first' && items.find(item => !(Number(item.material?.durationSeconds || item.material?.duration || 0) > 0));
    if (missingDuration) return message.warning(`素材 ${missingDuration.id} 没有有效的视频时长，暂时不能交给执行器处理`);
    const selected = items.map(item => ({ item, book: selectGiantMaterialBook(item.material, item.selectedBookKey) }));
    const missingSelection = selected.find(entry => !entry.book);
    if (missingSelection) return message.warning(`素材 ${missingSelection.item.id} 关联多条平台书籍记录，请在清单中选择一条`);

    setBusy(true);
    setGiantError('');
    try {
      const importedAt = new Date().toISOString();
      const intakePayload = buildGiantMaterialPlaceholderIntakes(
        selected.map(entry => ({ giantMaterialId: entry.item.id, material: entry.item.material, book: entry.book })),
        { contentRangeLines, importedAt, originalReadStrategy: giantOriginalReadStrategy }
      );
      const intakeResponse = await createNovelFetchIntake(intakePayload);
      const intake = intakeResponse?.intake || intakeResponse;
      if (!intake?.id) throw new Error('GIANT_MATERIAL_INTAKE_FAILED');
      const selectedPreset = (scheduledRun || automationRun) ? automationPresetSnapshot(automationPresets, automationPresetID) : null;
      if ((scheduledRun || automationRun) && !selectedPreset) throw new Error('所选自动化预设不存在或没有可用配置，请重新选择。');
      const giantAutomationPlan = (scheduledRun || automationRun) ? {
        scheduledAt: scheduledAtISO,
        scheduleMode: scheduledRun ? automationScheduleMode : 'execute',
        presetId: selectedPreset.id,
        runMode: automationRunMode,
        autoPublish: automationRun && automationRunMode === 'full_submit',
        concurrency: normalizeAutomationConcurrency(automationConcurrency)
      } : undefined;
      const batch = await onCreated?.({
        intakeId: intake.id,
        title: title.trim(),
        scheduledAt: scheduledAtISO,
        scheduleMode: scheduledRun ? automationScheduleMode : 'execute',
        // 巨量书在 OCR 回填前没有可生产正文。自动生产计划先随书保存，
        // 由正文回填成功的那一刻启动，避免被普通自动化误判为“原文缺失”。
        automationEnabled: false,
        autoPublishEnabled: false,
        presetId: automationPresetID,
        runMode: automationRunMode,
        automationConcurrency: normalizeAutomationConcurrency(automationConcurrency),
        initialBatchSettings: clone(initialBatchSettings),
        giantAutomationPlan: selectedPreset ? {
          presetId: selectedPreset.id,
          expectedPresetVersion: selectedPreset.version,
          runMode: automationRunMode,
          concurrency: normalizeAutomationConcurrency(automationConcurrency),
          scheduledAt: scheduledAtISO,
          scheduleMode: scheduledRun ? automationScheduleMode : 'execute'
        } : undefined
      });
      const batchId = batch?.id;
      if (!batchId) throw new Error('GIANT_MATERIAL_INTAKE_FAILED');
      let queued = 0;
      let hasDirectReadFailure = false;
      for (const entry of selected) {
        try {
          if (giantOriginalReadStrategy === 'direct_first') {
            const latestResponse = await getBatch(batchId);
            const latestBatch = latestResponse?.batch || latestResponse?.data?.batch || latestResponse?.data || latestResponse;
            const book = findRegisteredGiantMaterialBook(latestBatch?.books, entry.item.id);
            if (!book?.id) throw new Error('GIANT_MATERIAL_BOOK_NOT_FOUND');
            await updateBookMetadata(batchId, book.id, {
              metadata: { ...(book.sourceMetadata || {}), originalReadStage: 'direct', originalReadError: '', contentPending: true, giantAutomationPlan },
              expectedRevision: Number(book.revision || 0)
            });
            try {
              await fetchBookOriginal(batchId, book.id);
              // 书城直取成功后，结束巨量素材的占位读取状态；否则列表会一直
              // 把已经拿到正文的书当成“等待 OCR”，自动生产也无法正确判断就绪。
              const refreshedResponse = await getBatch(batchId);
              const refreshedBatch = refreshedResponse?.batch || refreshedResponse?.data?.batch || refreshedResponse?.data || refreshedResponse;
              const refreshedBook = findRegisteredGiantMaterialBook(refreshedBatch?.books, entry.item.id);
              if (refreshedBook?.id) {
                await updateBookMetadata(batchId, refreshedBook.id, {
                  metadata: {
                    ...(refreshedBook.sourceMetadata || {}),
                    contentPending: false,
                    originalReadStage: 'direct',
                    originalReadError: '',
                    giantAutomationPlan
                  },
                  expectedRevision: Number(refreshedBook.revision || 0)
                });
                try { await classifyBookPublishMetadata(batchId, refreshedBook.id, { force: false }); } catch (_) { /* 分类可在工作区单独重试 */ }
              }
            } catch (directFetchError) {
              hasDirectReadFailure = true;
              const directReadError = normalizedError(directFetchError, '书城获取原文失败');
              const refreshedResponse = await getBatch(batchId);
              const refreshedBatch = refreshedResponse?.batch || refreshedResponse?.data?.batch || refreshedResponse?.data || refreshedResponse;
              const refreshedBook = findRegisteredGiantMaterialBook(refreshedBatch?.books, entry.item.id);
              if (refreshedBook?.id) {
                await updateBookMetadata(batchId, refreshedBook.id, {
                  metadata: {
                    ...(refreshedBook.sourceMetadata || {}),
                    originalReadStage: 'failed',
                    originalReadError: directReadError,
                    contentPending: false,
                    giantAutomationPlan
                  },
                  expectedRevision: Number(refreshedBook.revision || 0)
                });
              }
              console.warn('巨量书城直取失败，等待用户手动重试或选择 OCR', entry.item.id, directFetchError);
              continue;
            }
            queued += 1;
            continue;
          }
          const durationSeconds = Number(entry.item.material.durationSeconds || entry.item.material.duration || 0);
          const createdResponse = await createGiantMaterialJob({
            materialId: entry.item.id,
            platformBookId: entry.book.platformBookId || entry.book.bookId,
            title: entry.book.title,
            videoUrl: entry.item.material.videoUrl,
            durationSeconds,
            modelVersion: 'windows-paddleocr-v1',
            contentRangeLines
          });
          const createdJob = createdResponse?.job || createdResponse?.data?.job || createdResponse;
          if (!createdJob?.id) throw new Error('GIANT_EXECUTOR_FAILED');
          // 创建批量会并行触发分类，不能使用 onCreated 回传的旧 revision；否则
          // metadata 的乐观锁冲突被吞掉，书卡就会显示“未绑定读取任务”。
          let latestError;
          for (let attempt = 0; attempt < 3; attempt += 1) {
            try {
              const latestResponse = await getBatch(batchId);
              const latestBatch = latestResponse?.batch || latestResponse?.data?.batch || latestResponse?.data || latestResponse;
              const book = findRegisteredGiantMaterialBook(latestBatch?.books, entry.item.id);
              if (!book?.id) throw new Error('GIANT_MATERIAL_BOOK_NOT_FOUND');
              await updateBookMetadata(batchId, book.id, {
                metadata: {
                  ...(book.sourceMetadata || {}),
                  executorJobId: createdJob.id,
                  contentPending: true,
                  originalReadStage: 'ocr',
                  originalReadError: '',
                  giantOcrState: String(createdJob.state || 'queued').toLowerCase(),
                  giantAutomationPlan
                },
                expectedRevision: Number(book.revision || 0)
              });
              latestError = null;
              break;
            } catch (bindError) {
              latestError = bindError;
              if (Number(bindError?.status) !== 409) break;
            }
          }
          if (latestError) throw latestError;
          queued += 1;
        } catch (error) {
          console.warn('巨量读取任务派发失败', entry.item.id, error);
        }
      }
      // 自动化只在整批素材的正文读取已全部派发后启动一次。引擎会把仍在直取或 OCR
      // 的书保持在“等待正文读取”，而不是把每条 ID 的成功回调都当成一条新生产任务。
      if (giantAutomationPlan?.presetId && !hasDirectReadFailure) {
        try {
          await startBatchAutomation(batchId, giantAutomationPlan);
        } catch (automationError) {
          // 正文读取任务已派发；自动生产失败不能把它们误标成“正文获取失败”。
          console.warn('巨量书城读取已派发，但自动制作未启动', batchId, automationError);
          message.warning(`正文读取已派发；自动生产请在工作区重试：${normalizedError(automationError, '启动失败')}`);
        }
      } else if (hasDirectReadFailure && giantAutomationPlan?.presetId) {
        message.warning('部分书城原文获取失败，自动生产未启动；请在工作区手动重试直取或选择 OCR。');
      }
      if (queued || giantOriginalReadStrategy === 'direct_first') {
        try { await onBatchUpdated?.(batchId); } catch (refreshError) { console.warn('巨量读取任务已派发，但工作区刷新失败', refreshError); }
      }
      message.success(queued
        ? giantOriginalReadStrategy === 'direct_first'
          ? `已创建批量并通过书城获取 ${queued}/${selected.length} 本正文；失败的书保留原因，可在工作区手动重试或使用滚屏 OCR。`
          : `已创建批量并排队读取 ${queued}/${selected.length} 本书；正文进度请在工作区书卡查看。`
        : giantOriginalReadStrategy === 'direct_first'
        ? '已创建批量，但书城获取原文失败；失败原因已保留，可在工作区手动重试或使用滚屏 OCR。'
          : '已创建批量，但读取任务派发失败；可在工作区用“原文获取”兜底。');
      reset();
    } catch (error) {
      const text = normalizedError(error, '巨量素材创建失败');
      setGiantError(text);
      message.error(text);
    } finally {
      setBusy(false);
    }
  }

  function validateDraft() {
    if (groups.some(group => Object.values(group.sourceMetadataByBookId || {}).some(meta => meta.sourceMode === 'custom_copy')) && customCopySessionRef.current !== getToken()) {
      message.warning('登录账号已变化，请关闭窗口后重新获取文案'); return false;
    }
    if (!title.trim()) return message.warning('请填写作品名称');
    if (isGiantMaterial) {
      if (!giantItems.some(item => item.status === 'resolved')) return message.warning('请先点击“解析全部”');
      return true;
    }
    if (!groups.length) return message.warning('请至少添加一个书城的书（粘贴后点“添加书城”）');
    if (editingPlatformId) return message.warning('请先完成当前书城的修改：点“保存书城修改”或 × 放弃');
    return true;
  }

  async function openAutomationDialog(mode = 'scheduled') {
    if (!validateDraft()) return;
    const scheduled = mode === 'scheduled';
    setAutomationDialogMode(scheduled ? 'scheduled' : 'immediate');
    if (!scheduled) setAutomationRunMode('full_submit');
    setScheduledAt(scheduled ? formatBeijingDatetimeLocal(new Date(Date.now() + 10 * 60 * 1000).toISOString()) : '');
    try {
      const result = await listAutomationPresets();
      setAutomationPresets(Array.isArray(result?.presets) ? result.presets : []);
    } catch (error) {
      message.error(normalizedError(error, '读取自动化预设失败'));
      return;
    }
    setScheduleDialogOpen(true);
  }

  async function openScheduleTasks() {
    setScheduleTasksOpen(true);
    setScheduleTasksLoading(true);
    try {
      const result = await listBatches();
      const batches = Array.isArray(result?.batches) ? result.batches : [];
      const tasks = await Promise.all(batches.map(async batch => {
        const status = await getBatchAutomationStatus(batch.id).catch(() => null);
        const automation = status?.automation || status || {};
        return { id: batch.id, title: batch.title || '未命名批量', state: automation.state || 'idle', scheduleMode: automation.scheduleMode || 'execute', scheduledAt: automation.scheduledAt || '' };
      }));
      setScheduleTasks(tasks.filter(task => task.state !== 'idle'));
    } catch (error) {
      message.error(normalizedError(error, '读取定时任务失败'));
    } finally {
      setScheduleTasksLoading(false);
    }
  }

  async function submit({ scheduledRun = false, automationRun = false } = {}) {
    if (isGiantMaterial) {
      await submitGiantMaterial({ scheduledRun, automationRun });
      return;
    }
    if (!validateDraft()) return;

    const automationEnabled = scheduledRun || automationRun;
    if (scheduledRun && !scheduledAt) return message.warning('请选择北京时间自动启动时间');
    if (automationEnabled && !automationPresetID) return message.warning('请选择自动化预设');
    const scheduledAtISO = scheduledRun ? parseBeijingDatetimeLocal(scheduledAt) : '';
    if (scheduledRun && (!scheduledAtISO || new Date(scheduledAtISO).getTime() <= Date.now())) return message.warning('北京时间自动启动时间需要晚于现在');

    setBusy(true);
    try {
      // 编辑中点了定时入口也会先校验；此时 groups 是唯一事实源。
      // 正文按组隔离：不同书城即便撞 Book ID 也各自抓各自的，互不覆盖。
      const sourcesByGroup = {};
      const sourceMetadataByGroup = {};
      const failedPlatforms = [];
      const failedReasons = [];
      let fetchedCount = 0;
      for (const group of groups) {
        const groupKey = String(group.platformId);
        const ids = manualBookIDsFromInput(group.inputText);
        const have = sourcesByGroup[groupKey] || group.sourceTextByBookId || {};
        sourcesByGroup[groupKey] = have;
        sourceMetadataByGroup[groupKey] = group.sourceMetadataByBookId || {};
        fetchedCount += ids.filter(bookId => String(have[bookId] || '').trim()).length;
        const pending = ids.filter(bookId => !String(have[bookId] || '').trim());
        if (!pending.length) continue;
        try {
          const result = await fetchDirectOriginals({
            platform: group.platformId,
            bookIds: pending,
            maxTxt: contentCaptureCharacters
          });
          const groupSources = {};
          const groupMetadata = {};
          for (const item of Array.isArray(result?.results) ? result.results : []) {
            const text = manualFetchedSourceText(item?.data);
            if (item?.status === 'ok' && text) {
              const bookId = String(item.bookId);
              groupSources[bookId] = text;
              if (item?.sourceMetadata && typeof item.sourceMetadata === 'object' && !Array.isArray(item.sourceMetadata)) {
                groupMetadata[bookId] = item.sourceMetadata;
              }
            }
          }
          sourcesByGroup[groupKey] = { ...have, ...groupSources };
          sourceMetadataByGroup[groupKey] = { ...groupMetadata, ...(sourceMetadataByGroup[groupKey] || {}) };
          // 只按本组请求的 pending id 计数，防止接口回传意外 ID 高估成功数
          for (const bookId of pending) {
            if (String(groupSources[bookId] || '').trim()) fetchedCount += 1;
          }
        } catch (error) {
          // 单组抓取失败不阻断：该组正文留空，书照样创建，列表中标红后可单独重试
          const platformLabel = group.platformName || group.platformId;
          failedPlatforms.push(platformLabel);
          failedReasons.push(`${platformLabel}：${normalizedError(error, '请求失败')}`);
          console.warn('分组抓取失败', group.platformId, error);
          sourcesByGroup[groupKey] = { ...have };
        }
      }
      // 设计约定：抓不到的书照样创建（正文为空，列表中标红），不阻断整个批量
      const totalCount = totalGroupBookCount(groups);

      if (!validateDraft()) return;

      await onCreated(buildManualBatchSubmission({
        title: title.trim(),
        groups: groups.map(group => ({
          platformId: group.platformId,
          platformName: group.platformName,
          inputText: group.inputText,
          sourceTextByBookId: sourcesByGroup[String(group.platformId)] || {},
          sourceMetadataByBookId: sourceMetadataByGroup[String(group.platformId)] || {}
        })),
        parseMode,
        columnPresetId,
        columnOrder,
        contentRangeLines,
        comicAnalysisCharacters,
        contentCaptureCharacters,
        scheduledAt: scheduledAtISO,
        scheduleMode: scheduledRun ? automationScheduleMode : 'execute',
        automationEnabled,
        autoPublishEnabled: automationEnabled && automationRunMode === 'full_submit',
        presetId: automationPresetID,
        runMode: automationRunMode,
        automationConcurrency: normalizeAutomationConcurrency(automationConcurrency),
        initialBatchSettings: clone(initialBatchSettings)
      }));
      reset();
      if (failedPlatforms.length > 0) {
        message.warning(`${fetchedCount}/${totalCount} 本已抓到正文；以下书城抓取失败：${failedPlatforms.join('、')}。${failedReasons.join('；')}失败的书已在列表中标红，可单独重试`);
      } else if (fetchedCount < totalCount) {
        message.warning(`${fetchedCount}/${totalCount} 本已抓到正文，其余书将在列表中标红，可单独重试`);
      }
    } catch (error) {
      message.error(normalizedError(error, '新建批量失败'));
    } finally {
      setBusy(false);
    }
  }

  const createDisabled = isGiantMaterial
    ? busy || giantBusy || !title.trim() || !giantItems.some(item => item.status === 'resolved')
    : busy || !title.trim() || editingPlatformId || !groups.length;

  return <><Modal
    className="shuihuo-create-project-modal"
    title="新建批量"
    open={open}
    onCancel={() => { onCancel(); reset(); }}
    width={760}
    footer={<Button onClick={() => { onCancel(); reset(); }}>取消</Button>}
  >
    {isGiantMaterial ? <BatchFactoryGiantMaterialExecutorStatus variant="dot" job={null} onHealthChange={setExecutorHealth} /> : null}

    <label className="shuihuo-form-label" htmlFor="batch-title">作品名称 <em>*</em></label>
    <Input id="batch-title" value={title} onChange={event => setTitle(event.target.value)} placeholder="请输入作品名称" maxLength={255} />

    <div style={{ marginTop: 16 }}>
      <label className="shuihuo-form-label">书城
        <Select
          value={platformId || undefined}
          onChange={changePlatform}
          options={platformOptions}
          loading={platformState === 'loading'}
          disabled={platformState === 'loading'}
          placeholder="请选择书城"
        />
      </label>
    </div>

    {platformState === 'fallback' ? <Alert
      type="warning"
      showIcon
      message="在线书城目录暂不可用，已启用内置书城列表"
      description={<Space direction="vertical" size={6}>
        <span>正文仍通过真实小说获取接口抓取，不会伪造内容。在线目录错误：{platformError}</span>
        <Button size="small" onClick={loadPlatforms} loading={platformState === 'loading'}>重试在线书城目录</Button>
      </Space>}
    /> : null}

    <label className="shuihuo-form-label" htmlFor="batch-input-text">{isCustomCopy ? '书籍 ID（每行一本）' : isGiantMaterial ? '巨量素材 ID' : '小说列表'} <em>*</em></label>
    <Input.TextArea
      id="batch-input-text"
      value={inputText}
      onChange={event => { setInputText(event.target.value); if (isGiantMaterial) clearGiantMaterialDraft(); }}
      rows={isGiantMaterial || isCustomCopy ? 3 : 9}
      placeholder={isCustomCopy ? '每行一个书籍 ID，例如：2100242549395006289' : isGiantMaterial ? '每行一个巨量素材 ID，数量不限\n例如：7689285355255727121' : '每行一本小说，可粘贴 ID、书名、男女频、风格、标签、推荐理由、评级。\n示例：2080989285751305136\t重生书\t女频\t现代爽文\t重生,逆袭\t女主逆袭\tS'}
    />
    {isCustomCopy ? <BatchFactoryCustomCopyPicker inputText={inputText} disabled={busy} onAdd={(results, selections) => {
      try {
        if (customCopySessionRef.current && customCopySessionRef.current !== getToken()) throw new Error('登录账号已变化，请关闭窗口后重新获取文案');
        const next = selectedCopyGroups(groups, results, selections, platformOptions);
        if (totalGroupBookCount(next) > 50) throw new Error('一个批量最多 50 本书');
        customCopySessionRef.current = getToken();
        setGroups(next); setInputText(''); message.success('所选文案已加入批量');
      } catch (error) { message.warning(error.message); }
    }} /> : null}
    {isGiantMaterial ? <div className="batch-factory-giant-read-strategy">
      <Switch checked={giantOriginalReadStrategy === 'direct_first'} onChange={checked => setGiantOriginalReadStrategy(checked ? 'direct_first' : 'ocr_first')} />
      <span>优先直接获取原文</span>
      <small>{giantOriginalReadStrategy === 'direct_first'
        ? '先通过书城和 Book ID 获取正文；失败后保留原因，可在工作区手动使用滚屏 OCR。'
        : '先读取视频滚屏；失败后自动通过书城获取正文。'}</small>
    </div> : null}

    {isGiantMaterial && giantError ? <Alert
      type="error"
      showIcon
      message="服务器令牌失效，需更新青语 N8 后台登录令牌后重新解析。"
      description="这是服务器级配置故障，其余素材不再逐个请求；更新令牌后点“重新解析”即可继续。"
      action={<Button size="small" onClick={resolveAllGiantMaterials}>重新解析</Button>}
      style={{ marginBottom: 12 }}
    /> : null}

    {isGiantMaterial && giantItems.length ? <div className="batch-factory-giant-items">
      {giantItems.map((item, index) => {
        const isDuplicate = giantDuplicateIndexes.has(index);
        const rowClassName = item.status === 'error'
          ? 'batch-factory-giant-item is-error'
          : `batch-factory-giant-item${isDuplicate ? ' is-duplicate' : ''}`;
        return <div key={`${item.id}-${index}`} className={rowClassName}>
          <span className="batch-factory-giant-item-index">{index + 1}</span>
          {item.status === 'error'
            ? <span className="batch-factory-giant-item-title is-error">{giantItemErrorText(item.error)}（{item.id}）</span>
            : item.status === 'resolved'
              ? <>
                <span className="batch-factory-giant-item-title">{item.material?.materialTitle || item.material?.title || item.id}</span>
                <span className="batch-factory-giant-item-meta">Book ID {(selectGiantMaterialBook(item.material, item.selectedBookKey)?.platformBookId) || '—'}</span>
                {isDuplicate ? <Tag className="batch-factory-giant-duplicate-tag">重复，将跳过</Tag> : null}
                {item.books.length > 1 ? <Select size="small" style={{ minWidth: 180 }} value={item.selectedBookKey || undefined} onChange={value => setGiantItems(current => current.map(row => row.id === item.id ? { ...row, selectedBookKey: value } : row))} options={item.books.map(book => ({ value: giantMaterialBookKey(book), label: `${book.title} · ${book.platformBookId}` }))} /> : null}
              </>
              : <span className="batch-factory-giant-item-title">{giantError ? `等待重新解析 ${item.id}` : `正在解析 ${item.id}…`}</span>}
        </div>;
      })}
    </div> : null}

    <div className="batch-factory-create-toolbar">
      {isGiantMaterial
        ? <><Button loading={giantBusy} disabled={giantBusy || busy || !inputText.trim()} onClick={resolveAllGiantMaterials}>解析全部</Button><Button className="batch-factory-execute-now" type="primary" disabled={createDisabled} loading={busy} onClick={() => openAutomationDialog('immediate')}>立即执行</Button></>
        : <>{!isCustomCopy ? <Button type="primary" onClick={handleAddPlatformGroup}>{editingPlatformId ? '保存书城修改' : '添加书城'}</Button> : null}<Button className="batch-factory-execute-now" type="primary" disabled={createDisabled} loading={busy} onClick={() => openAutomationDialog('immediate')}>立即执行</Button></>}
      <Button onClick={() => openAutomationDialog('scheduled')}>开始定时</Button>
      <Button onClick={openScheduleTasks}>定时任务</Button>
      <Button onClick={() => setInitialSettingsOpen(true)}>统一配置</Button>
    </div>
    {groups.some(group => Object.keys(group.sourceTextByBookId || {}).length > 0) ? <p style={{ fontSize: 12, opacity: .8 }}>“文案获取”使用所选文案全文；内容范围与截取长度仅用于其他书城。</p> : null}
    {!isGiantMaterial && groups.length ? (
      <div className="batch-factory-platform-groups" data-testid="platform-groups">
        {groups.map(group => (
          <Tag
            key={group.platformId}
            className={editingPlatformId === group.platformId ? 'platform-group-tag is-editing' : 'platform-group-tag'}
            closable
            onClose={event => { event.preventDefault(); handleRemoveGroup(group.platformId); }}
            onClick={() => handleEditGroup(group)}
            style={{ cursor: 'pointer', marginBottom: 4 }}
          >
            {group.platformName} ×{manualBookIDsFromInput(group.inputText).length}{Object.keys(group.sourceTextByBookId || {}).length ? ' · 已选文案' : ''}{editingPlatformId === group.platformId ? '（编辑中）' : ''}
          </Tag>
        ))}
      </div>
    ) : null}

    <div className="batch-factory-content-controls">
      <label className="shuihuo-form-label">生产模式<Select value={initialBatchSettings.productionMode || 'video'} onChange={productionMode => setInitialBatchSettings(current => ({ ...current, productionMode }))} options={[{ value: 'video', label: '视频生成' }, { value: 'comic', label: '漫画生成' }]} /></label>
      {isComicProduction
        ? <label className="shuihuo-form-label">漫画分析<InputNumber min={100} max={100000} step={100} value={comicAnalysisCharacters} onChange={value => setComicAnalysisCharacters(Math.min(100000, Math.max(100, Number(value) || 2000)))} addonAfter="字" /> <small>预计约 {comicEstimatedPageRange(comicAnalysisCharacters).min}-{comicEstimatedPageRange(comicAnalysisCharacters).max} 张 · 最多 35 张</small></label>
        : <label className="shuihuo-form-label">内容范围<InputNumber min={1} max={500} value={contentRangeLines} onChange={value => setContentRangeLines(value || 5)} addonAfter="行" /></label>}
      <label className="shuihuo-form-label">内容截取<Select value={contentCaptureCharacters} onChange={setContentCaptureCharacters} options={[1000, 2000, 4000, 8000, 12000, 20000, 50000, 100000].map(value => ({ value, label: `${value} 字` }))} /></label>
    </div>
  </Modal>
  <Modal
    title={<div className="batch-factory-unified-settings-title"><BatchFactoryAutomationPresetManager value={initialBatchSettings} onLoad={setInitialBatchSettings} /><span>新建批量 · 统一配置</span><BatchFactoryProductionModeControl value={initialBatchSettings.productionMode} onChange={productionMode => setInitialBatchSettings(current => ({ ...current, productionMode }))} /></div>}
    open={initialSettingsOpen}
    onCancel={() => setInitialSettingsOpen(false)}
    onOk={() => setInitialSettingsOpen(false)}
    okText="确认配置"
    cancelText="取消"
    width={980}
    destroyOnClose={false}
  >
    <Alert type="info" showIcon message="首次批量统一配置" description="配置会在创建批量时一次性保存为该批量默认值；后续仍可在工作台的“统一配置”继续修改，单书覆盖不会被清空。" />
    <Tabs items={[
      { key: 'models', label: '模型配置', children: <BatchFactoryEngineSettingsForm value={initialBatchSettings} onChange={setInitialBatchSettings} sections={['models', 'audio']} active={initialSettingsOpen} /> },
      { key: 'reasoning', label: 'AI 推理', children: <BatchFactoryAiReasoningForm productionMode={initialBatchSettings.productionMode} value={initialBatchSettings.aiPromptConfig} onChange={aiPromptConfig => setInitialBatchSettings(current => ({ ...current, aiPromptConfig }))} /> },
      { key: 'publish', label: '发布统一', children: isComicProduction ? <BatchFactoryComicPublishSettingsForm value={initialBatchSettings} onChange={setInitialBatchSettings} /> : <BatchFactoryPublishSettingsForm value={initialBatchSettings} onChange={setInitialBatchSettings} active={initialSettingsOpen} /> }
    ]} />
  </Modal>
  <Modal
    title={automationDialogMode === 'scheduled' ? '开始定时' : '立即执行自动生产'}
    open={scheduleDialogOpen}
    // 取消必须同时清除 busy：创建请求可能因网关 502/中断一直挂着，
    // 只关弹窗会让按钮永久停在加载态，重新打开后仍然点不动。
    onCancel={() => { setBusy(false); setScheduleDialogOpen(false); }}
    confirmLoading={busy}
    okText={automationDialogMode === 'scheduled' ? (automationScheduleMode === 'submit' ? '创建并定时提交' : '创建并定时执行') : '创建并立即执行'}
    cancelText="取消"
    onOk={() => submit({ scheduledRun: automationDialogMode === 'scheduled', automationRun: true })}
  >
    {automationDialogMode === 'scheduled' ? <><label className="shuihuo-form-label" htmlFor="batch-scheduled-at">自动启动时间（北京时间 UTC+8） <em>*</em></label>
    <Input id="batch-scheduled-at" type="datetime-local" value={scheduledAt} onChange={event => setScheduledAt(event.target.value)} />
    <small>{automationScheduleMode === 'submit' ? '提前完成生产和合成；到设定时间才上传或发布。' : '到设定时间才开始生产；生成、合成与上传按后续流程继续。'}</small></> : <Alert type="info" showIcon message="创建完成后立即启动自动生产" description="不会直接提交；只有选择全自动模式且成片完成后，才会上传视频管理系统。" />}
    <label className="shuihuo-form-label">自动化预设 <em>*</em></label>
    <Select value={automationPresetID || undefined} onChange={setAutomationPresetID} placeholder="选择已保存预设" options={automationPresets.map(item => ({ value: item.id, label: `${item.name} · v${item.version}` }))} />
    <label className="shuihuo-form-label">执行模式</label>
    <Select value={automationRunMode} onChange={value => { setAutomationRunMode(value); if (value !== 'full_submit') setAutomationScheduleMode('execute'); }} options={[{ value: 'storyboard_only', label: '只生成分镜' }, { value: 'video_no_submit', label: '生成视频不提交' }, { value: 'full_submit', label: '全自动生成并提交（成片完成后上传）' }]} />
    {automationDialogMode === 'scheduled' ? <><label className="shuihuo-form-label">定时方式</label><Select value={automationScheduleMode} onChange={setAutomationScheduleMode} options={[{ value: 'execute', label: '定时执行（到点才开始生产）' }, { value: 'submit', label: '定时提交（提前生产，到点上传/发布）', disabled: automationRunMode !== 'full_submit' }]} /></> : null}
    <label className="shuihuo-form-label">同时处理书籍</label>
    <Select value={automationConcurrency} onChange={value => setAutomationConcurrency(normalizeAutomationConcurrency(value))} options={[1, 2, 4].map(value => ({ value, label: `${value} 本并行` }))} />
  </Modal>
  <Modal title="定时任务" open={scheduleTasksOpen} onCancel={() => setScheduleTasksOpen(false)} footer={<Button onClick={() => setScheduleTasksOpen(false)}>关闭</Button>}>
    <Button loading={scheduleTasksLoading} onClick={openScheduleTasks}>刷新</Button>
    <div className="batch-factory-log-list">{scheduleTasks.length ? scheduleTasks.map(task => <section key={task.id}><strong>{task.title}</strong><span>{task.scheduledAt ? `北京时间 ${formatBeijingDatetimeLocal(task.scheduledAt)}` : '立即执行'} · {task.scheduleMode === 'submit' ? '定时提交' : '定时执行'} · {task.state}</span></section>) : <p>{scheduleTasksLoading ? '读取中…' : '暂无定时任务'}</p>}</div>
  </Modal></>;
}
