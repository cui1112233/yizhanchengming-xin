import { Alert, Button, Empty, Input, Modal, Space, Spin, Tag, Typography, message } from 'antd';
import { useCallback, useEffect, useMemo, useState } from 'react';
import * as batchFactoryV11 from '../../../shared/api/batchFactoryV11.js';
import { textToSpeech } from '../../../shared/api/tts.js';
import { BatchFactoryV11Workbench } from './BatchFactoryV11Workbench';
import { ProductionSettingsDrawer } from './BatchFactoryV11SettingsDrawers';
import { PublishSettingsDrawer } from './BatchFactoryV11PublishSettings';
import { BookSettingsModal, VideoSettingsDrawer } from './BatchFactoryV11ScopedSettings';
import { DirectorRefreshProvider } from './DirectorRefreshContext.jsx';
import { FinalPromptPreviewDrawer } from './FinalPromptPreviewDrawer.jsx';
import { ExternalPublishPanel } from './ExternalPublishPanel.jsx';
import { videoProviderForModel } from '../shuihuo/videoProviderBinding.js';
import { actionState, intakeCreateState } from './batchFactoryV11State.js';
import { createBf11UiAdapter } from './bf11UiAdapter.js';
import { createBf11Runtime, createdBatchIdFrom } from './bf11Runtime.js';
import './batch-factory-v11-scoped.css';
import './batch-factory-v11-detail.css';
import './batch-factory-v11-theme.css';

function requestParamsFromLocation(initialBatchId = '') {
  if (typeof window === 'undefined') return {};
  const search = new URLSearchParams(window.location.search || '');
  return {
    batchId: search.get('batch') || initialBatchId || '',
    intakeId: search.get('intake') || ''
  };
}

function emptySettingsState() {
  return { patch: {}, revision: 0, snapshot: null, compatibility: [] };
}

function batchForView(batch, books) {
  if (!batch) return null;
  return {
    ...batch,
    title: batch.title || `批次 ${batch.id}`,
    count: Number(batch.count ?? books.length)
  };
}

function blobToBase64(blob) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => reject(new Error('无法读取 TTS 音频数据'));
    reader.onload = () => resolve(String(reader.result || '').split(',').at(-1) || '');
    reader.readAsDataURL(blob);
  });
}

function ttsFingerprint(tts) {
  return `v11-tts-${JSON.stringify({
    voice: String(tts?.voice || ''),
    style: String(tts?.style || ''),
    speed: Number(tts?.speed || 0),
    pitch: Number(tts?.pitch || 0)
  })}`;
}

export function BatchFactoryV11UiPage({ initialBatchId = '' } = {}) {
  const adapter = useMemo(() => createBf11UiAdapter({
    ...batchFactoryV11,
    listPersonalConstraintPrompts: batchFactoryV11.listPersonalConstraintPrompts
  }), []);
  const runtime = useMemo(() => createBf11Runtime({ adapter }), [adapter]);
  const requestParams = useMemo(() => requestParamsFromLocation(initialBatchId), [initialBatchId]);
  const [runtimeState, setRuntimeState] = useState({ phase: 'loading' });
  const [creatingBatch, setCreatingBatch] = useState(false);
  const [productionSettingsOpen, setProductionSettingsOpen] = useState(false);
  const [bookSettingsTargetId, setBookSettingsTargetId] = useState('');
  const [videoSettingsTarget, setVideoSettingsTarget] = useState(null);
  const [directorAction, setDirectorAction] = useState({ type: '', bookId: '' });
  const [promptPreview, setPromptPreview] = useState({ open: false, loading: false, data: null, error: '' });
  const [productionBusy, setProductionBusy] = useState(false);
  const [mergeBusy, setMergeBusy] = useState(false);
  const [audioBusyBookId, setAudioBusyBookId] = useState('');
  const [publishSettingsOpen, setPublishSettingsOpen] = useState(false);
  const [externalPublishOpen, setExternalPublishOpen] = useState(false);
  const [batchManagerOpen, setBatchManagerOpen] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [newBatchTitle, setNewBatchTitle] = useState('');
  const [historyBatches, setHistoryBatches] = useState([]);
  const [scheduleOpen, setScheduleOpen] = useState(false);
  const [scheduleAt, setScheduleAt] = useState('');
  const [schedules, setSchedules] = useState([]);
  const [generatedShotImages, setGeneratedShotImages] = useState({});
  const [generatedAssetImages, setGeneratedAssetImages] = useState({ character: {}, scene: {}, prop: {} });

  const reload = useCallback(async ({ announce = false } = {}) => {
    setRuntimeState(current => ({ ...current, phase: 'loading' }));
    const next = await runtime.load(requestParams);
    setRuntimeState(next);
    if (announce && next.phase === 'ready') message.success('已刷新 V11 工作台');
    return next;
  }, [runtime, requestParams]);

  useEffect(() => {
    let cancelled = false;
    runtime.load(requestParams).then(next => {
      if (!cancelled) setRuntimeState(next);
    });
    return () => { cancelled = true; };
  }, [runtime, requestParams]);

  useEffect(() => {
    if (runtimeState.phase !== 'ready' || !runtimeState.batch?.id) return undefined;
    const activeProduction = (runtimeState.productionStatus?.jobs || []).some(job =>
      (job.tasks || []).some(task => task.status === 'queued' || task.status === 'running')
    );
    const activeMerge = (runtimeState.mergeStatus?.jobs || []).some(job =>
      job.status === 'queued' || job.status === 'running'
    );
    const active = activeProduction || activeMerge;
    if (!active) return undefined;
    let cancelled = false;
    const poll = async () => {
      const next = await runtime.load({ ...requestParams, batchId: runtimeState.batch.id });
      if (!cancelled && next.phase === 'ready') setRuntimeState(next);
    };
    const timer = window.setInterval(poll, 5000);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [runtime, requestParams, runtimeState.phase, runtimeState.batch?.id, runtimeState.productionStatus, runtimeState.mergeStatus]);

  if (runtimeState.phase === 'loading') {
    return <div data-bf-v11-ui="final" style={{ minHeight: 420, display: 'grid', placeItems: 'center' }}>
      <Spin size="large" tip="正在读取 Batch Factory V11…" />
    </div>;
  }

  if (runtimeState.phase === 'error') {
    return <div data-bf-v11-ui="final" style={{ minHeight: 420, display: 'grid', placeItems: 'center' }}>
      <Alert
        type="error"
        showIcon
        message="Batch Factory V11 暂时不可用"
        description={<div>
          <Typography.Paragraph>{runtimeState.message}</Typography.Paragraph>
          <Button onClick={() => reload()}>重新加载</Button>
        </div>}
      />
    </div>;
  }

  const batch = runtimeState.batch;
  const books = runtimeState.books || [];
  const capabilities = runtimeState.capabilities || {};
  const batchCreateAction = actionState(capabilities, 'batch.create');
  const intakeState = intakeCreateState(runtimeState.intake);
  const batchSettingsState = batch?.settingsState || emptySettingsState();
  const viewBatch = batchForView(batch, books);

  async function createBatchFromIntake() {
    if (!intakeState.intakeId || intakeState.consumed || batchCreateAction.disabled || creatingBatch) return false;
    setCreatingBatch(true);
    try {
      const result = await runtime.createBatchFromIntake({ intakeId: intakeState.intakeId, payload: {} });
      if (!result.ok) {
        message.error(result.message);
        return false;
      }
      const createdBatchId = createdBatchIdFrom(result.raw);
      const next = await runtime.load(createdBatchId
        ? { ...requestParams, batchId: createdBatchId }
        : requestParams);
      setRuntimeState(next);
      if (next.phase !== 'ready' || !next.batch) {
        message.warning('批次已创建，但重新读取工作台失败，请刷新后继续。');
        return true;
      }
      message.success('V11 批次已创建；Director 尚未启动。');
      return true;
    } finally {
      setCreatingBatch(false);
    }
  }

  if (!batch) {
    const intake = runtimeState.intake;
    const intakeStatus = intake?.status || (intakeState.consumed ? '已消费' : '待创建批次');
    const intakeCount = Number(intake?.bookCount ?? intake?.count ?? 0);
    const createDisabledReason = intakeState.consumed
      ? '该 Intake 已经被消费，不能重复创建批次。'
      : batchCreateAction.reason;

    return <div data-bf-v11-ui="final" style={{ minHeight: 460, display: 'grid', placeItems: 'center' }}>
      <Empty
        description={intake ? <Space direction="vertical" size={8}>
          <Typography.Text>已读取小说获取转入任务，等待你明确创建 V11 批次。</Typography.Text>
          <Space wrap>
            <Tag>Intake {intake.id || '—'}</Tag>
            <Tag color={intakeState.consumed ? 'default' : 'processing'}>{intakeStatus}</Tag>
            {intake.sourceTaskId ? <Tag>来源 {intake.sourceTaskId}</Tag> : null}
            {intakeCount > 0 ? <Tag>{intakeCount} 本</Tag> : null}
          </Space>
          <Typography.Text type="secondary">创建批次不会自动启动 Director。</Typography.Text>
        </Space> : '当前没有 Batch Factory V11 批次。'}
      >
        <Space wrap>
          {intake ? <Button
            type="primary"
            loading={creatingBatch}
            disabled={creatingBatch || intakeState.consumed || batchCreateAction.disabled || !intakeState.intakeId}
            title={createDisabledReason}
            onClick={createBatchFromIntake}
          >创建 V11 批次</Button> : null}
          <Button onClick={() => reload({ announce: true })}>刷新批次</Button>
        </Space>
      </Empty>
    </div>;
  }

  const activeBook = bookSettingsTargetId
    ? books.find(book => book.id === bookSettingsTargetId) || null
    : null;
  const activeVideoBook = videoSettingsTarget
    ? books.find(book => book.id === videoSettingsTarget.bookId) || null
    : null;
  const activeVideo = activeVideoBook && videoSettingsTarget
    ? (activeVideoBook.videos || []).find(video => video.id === videoSettingsTarget.videoId) || null
    : null;

  async function saveScope(input, successMessage) {
    const result = await runtime.save(input);
    if (!result.ok) {
      message.error(result.message);
      return false;
    }
    message.success(successMessage);
    const next = await runtime.load(requestParams);
    setRuntimeState(next);
    if (next.phase !== 'ready') {
      message.warning('设置已保存，但重新读取工作台失败，请稍后刷新。');
    }
    return true;
  }

  function previewBatchChangeImpact(patch) {
    return runtime.previewChangeImpact({
      batchId: batch.id,
      patch,
      revision: batchSettingsState.revision
    });
  }

  async function refreshDirectorRevision() {
    const next = await runtime.load({ ...requestParams, batchId: batch.id });
    setRuntimeState(next);
    if (next.phase !== 'ready') {
      message.error(next.message || '刷新 Director revision 失败');
      return false;
    }
    message.success('Director revision 已刷新');
    return true;
  }

  function saveBatchSettings(patch) {
    return saveScope({
      scope: 'batch',
      batchId: batch.id,
      patch,
      revision: batchSettingsState.revision
    }, '生产统一设置已保存');
  }

  function syncBatchConfigVersion({ versionConfigId } = {}) {
    if (!versionConfigId) return Promise.resolve(false);
    return saveScope({
      scope: 'batch',
      batchId: batch.id,
      patch: { versionConfigId },
      revision: batchSettingsState.revision
    }, '批量后台配置已同步');
  }

  function saveBookSettings(patch) {
    if (!activeBook) return Promise.resolve(false);
    return saveScope({
      scope: 'book',
      batchId: batch.id,
      bookId: activeBook.id,
      patch,
      revision: activeBook.settingsState?.revision || 0
    }, '当前小说设置已保存');
  }

  async function measureAudio(book) {
    if (!book?.id || audioBusyBookId) return false;
    const director = book.directorRevision;
    const document = director?.output?.h3_director || director?.output?.h3Director;
    const cards = document?.director_cards || [];
    if (!director?.id || !document || !cards.length) {
      message.warning('请先完成当前小说的 Director，再读取 TTS 配音时长。');
      return false;
    }
    setAudioBusyBookId(book.id);
    try {
      const tts = { voice: 'zh-CN-XiaoxiaoNeural', style: 'general', speed: 1.8, pitch: 10, ...(batchSettingsState.patch.tts || {}), ...(book.settingsState?.patch?.tts || {}) };
      const lines = [];
      for (const card of cards) {
        const sourceText = String(card.source_text || '').trim();
        if (!sourceText) throw new Error('Director 中存在空的配音原文');
        const audio = await textToSpeech({ input: sourceText, ...tts });
        lines.push({ source_key: card.source_key, source_text: sourceText, audio_base64: await blobToBase64(audio) });
      }
      const result = await batchFactoryV11.measureH3Audio(batch.id, book.id, {
        director_revision_id: director.id,
        tts_fingerprint: ttsFingerprint(tts),
        lines
      });
      const measurement = result?.audio_measurement?.measurement || result?.audioMeasurement?.measurement || {};
      const durationSeconds = Number(measurement.duration_ms || 0) / 1000;
      if (!(durationSeconds > 0)) throw new Error('TTS 已返回，但没有得到有效配音时长');
      const saveResult = await runtime.save({
        scope: 'book', batchId: batch.id, bookId: book.id,
        patch: { audioDurationSeconds: Number(durationSeconds.toFixed(2)), audioDurationFingerprint: ttsFingerprint(tts), audioDurationManual: false },
        revision: book.settingsState?.revision || 0
      });
      if (!saveResult.ok) throw new Error(saveResult.message || '保存配音时长失败');
      const next = await runtime.load({ ...requestParams, batchId: batch.id });
      setRuntimeState(next);
      message.success(`已读取 ${book.title || '当前小说'} 的 TTS 配音时长：${durationSeconds.toFixed(2)} 秒`);
      return true;
    } catch (error) {
      message.error(error?.message || '读取 TTS 配音时长失败');
      return false;
    } finally {
      setAudioBusyBookId('');
    }
  }

  function saveVideoSettings(patch) {
    if (!activeVideoBook || !activeVideo) return Promise.resolve(false);
    return saveScope({
      scope: 'video',
      batchId: batch.id,
      bookId: activeVideoBook.id,
      videoId: activeVideo.id,
      patch,
      revision: activeVideo.settingsState?.revision || 0
    }, '单 VIDEO 设置已保存');
  }

  async function finishDirectorAction(result, successMessage) {
    if (!result.ok) { message.error(result.message); return false; }
    const next = await runtime.load({ ...requestParams, batchId: batch.id });
    setRuntimeState(next);
    if (next.phase !== 'ready') { message.warning(`${successMessage}，但重新读取工作台失败，请刷新。`); return true; }
    message.success(successMessage);
    return true;
  }

  async function runHook(book) {
    if (!book?.id || directorAction.type) return false;
    setDirectorAction({ type: 'hook', bookId: book.id });
    try { return await finishDirectorAction(await runtime.runHook({ batchId: batch.id, bookId: book.id }), 'Hook 已生成，等待你审核批准'); }
    finally { setDirectorAction({ type: '', bookId: '' }); }
  }

  async function approveHook(book, hook) {
    if (!book?.id || !hook?.id || directorAction.type) return false;
    setDirectorAction({ type: 'approve', bookId: book.id });
    try { return await finishDirectorAction(await runtime.approveHook({ batchId: batch.id, bookId: book.id, hookId: hook.id }), 'Hook 已批准'); }
    finally { setDirectorAction({ type: '', bookId: '' }); }
  }

  async function runDirector(book) {
    if (!book?.id || directorAction.type) return false;
    setDirectorAction({ type: 'director', bookId: book.id });
    try { return await finishDirectorAction(await runtime.runDirector({ batchId: batch.id, bookId: book.id, textModelId: batchSettingsState.patch.textModelId || '' }), 'Director 已完成并生成新的 VIDEO identity'); }
    finally { setDirectorAction({ type: '', bookId: '' }); }
  }

  function refreshAssets(book) {
    if (!book?.id || directorAction.type) return false;
    return new Promise(resolve => {
      Modal.confirm({
        title: '重新获取资产',
        content: '这会重新执行 Director，并生成新的 VIDEO identity；当前旧 VIDEO 的单独设置会保留为历史覆盖。确定继续吗？',
        okText: '继续重新获取',
        cancelText: '取消',
        onOk: async () => { resolve(await runDirector(book)); },
        onCancel: () => resolve(false)
      });
    });
  }

  async function previewFinalPrompt(book, video, shot = null) {
    if (!book?.id || !video?.id) return false;
    setPromptPreview({ open: true, loading: true, data: null, error: '' });
    const result = await runtime.previewFinalPrompt({ batchId: batch.id, bookId: book.id, videoId: video.id, shotId: shot?.id || '' });
    if (!result.ok) {
      setPromptPreview({ open: true, loading: false, data: null, error: result.message });
      return false;
    }
    setPromptPreview({ open: true, loading: false, data: result.raw, error: '' });
    return true;
  }

  function newRequestId(prefix) {
    const random = globalThis.crypto?.randomUUID?.();
    return `${prefix}-${random || `${Date.now()}-${Math.random().toString(16).slice(2)}`}`;
  }

  async function runProduction(targetBatch, targetBook = null) {
    if (!targetBatch?.id || productionBusy) return false;
    const provider = videoProviderForModel(
      batchSettingsState.patch.videoModelId,
      batchSettingsState.patch.videoProvider || 'personal_api'
    );
    if (provider === 'doubao_local_executor' && !runtimeState.localExecutors.some(item => item.online)) {
      message.error('没有在线的豆包本地执行器；请先在生产统一设置生成配对码并让 Mac 执行器上线。');
      return false;
    }
    if (provider === 'autodl_comfyui' && runtimeState.videoProviders?.h3?.configured === false) {
      message.error('服务端尚未配置 AutoDL H3 API Key，请先配置后再提交。');
      return false;
    }
    setProductionBusy(true);
    try {
      const result = await runtime.runProduction({
        batchId: targetBatch.id,
        bookId: targetBook?.id || '',
        requestId: newRequestId('bf11-production'),
        provider,
        videoModelId: batchSettingsState.patch.videoModelId || ''
      });
      if (!result.ok) { message.error(result.message); return false; }
      const next = await runtime.load({ ...requestParams, batchId: targetBatch.id });
      setRuntimeState(next);
      if (next.phase !== 'ready') { message.warning('视频任务已提交，但刷新状态失败，请稍后重试。'); return true; }
      message.success(targetBook ? '当前小说待生成 VIDEO 已提交，状态会自动写回工作台。' : '待生成 VIDEO 已提交，状态会自动写回工作台。');
      return true;
    } finally {
      setProductionBusy(false);
    }
  }

  async function runMerge(targetBatch, options = {}) {
    if (!targetBatch?.id || mergeBusy) return false;
    setMergeBusy(true);
    try {
      const result = await runtime.runMerge({ batchId: targetBatch.id, requestId: newRequestId('bf11-merge'), ...options });
      if (!result.ok) { message.error(result.message); return false; }
      const next = await runtime.load({ ...requestParams, batchId: targetBatch.id });
      setRuntimeState(next);
      if (next.phase !== 'ready') { message.warning('合并任务已提交，但刷新状态失败，请稍后重试。'); return true; }
      message.success('待合并任务已提交，状态会写回工作台。');
      return true;
    } finally {
      setMergeBusy(false);
    }
  }

  async function runBatchDirector(targetBatch) {
    if (!targetBatch?.id || directorAction.type) return false;
    setDirectorAction({ type: 'batch-director', bookId: '' });
    try {
      const result = await runtime.runBatchDirector({ batchId: targetBatch.id });
      if (!result.ok) { message.error(result.message); return false; }
      const raw = result.raw && typeof result.raw === 'object' ? result.raw : {};
      const failures = Array.isArray(raw.failures) ? raw.failures : [];
      const revisions = Array.isArray(raw.directorRevisions) ? raw.directorRevisions : [];
      const bookNames = new Map(books.map(book => [book.id, book.title || book.id]));
      const failedNames = failures.map(item => bookNames.get(item?.bookId) || item?.bookId || '未知书籍');
      const listed = failedNames.slice(0, 3).join('、') + (failedNames.length > 3 ? ' 等' : '');
      const next = await runtime.load({ ...requestParams, batchId: targetBatch.id });
      setRuntimeState(next);
      if (!failures.length && raw.error) { message.error(raw.error); return false; }
      if (!revisions.length && failures.length) { message.error(`批量 Director 全部失败：${failures.length} 本未生成（${listed}）`); return false; }
      if (failures.length) {
        message.warning(`批量 Director 完成 ${revisions.length} 本；${failures.length} 本失败：${listed}`);
        return true;
      }
      message.success('批量 Director 已完成。');
      return true;
    } finally {
      setDirectorAction({ type: '', bookId: '' });
    }
  }

  async function generateAssetImage(book, type, item, prompt) {
    const modelId = batchSettingsState.patch.imageModelId || '';
    if (!modelId || !book?.id || !prompt) { message.error('请先选择图片模型并确认资产 Prompt'); return false; }
    const assetId = item?.id || item?.name || item?.label || prompt.slice(0, 40);
    try {
      const raw = await batchFactoryV11.generateConfiguredImage({ imageModelId: modelId, prompt });
      const imageUrl = raw?.imageUrl || raw?.url || '';
      if (!imageUrl) throw new Error('图片模型未返回图片地址');
      setGeneratedAssetImages(current => ({ ...current, [type]: { ...current[type], [assetId]: imageUrl } }));
      await runtime.saveDraft({ key: `asset:${type}:${assetId}:image`, kind: 'asset-image', scope: `${batch.id}:${book.id}`, content: imageUrl });
      message.success(`${type === 'character' ? '人物' : type === 'scene' ? '场景' : '道具'}图片已生成并保存`);
      return true;
    } catch (error) { message.error(error?.message || '资产图片生成失败'); return false; }
  }

  async function generateShotImage(book, video, shot) {
    if (!book?.id || !video?.id || !shot?.id || !shot.visualPrompt) return false;
    const modelId = batchSettingsState.patch.imageModelId || '';
    if (!modelId) { message.error('请先在生产统一设置中选择图片模型'); return false; }
    try {
      const raw = await batchFactoryV11.generateConfiguredImage({ imageModelId: modelId, prompt: shot.visualPrompt });
      const imageUrl = raw?.imageUrl || raw?.url || '';
      if (!imageUrl) throw new Error('图片模型未返回图片地址');
      await runtime.saveShotVisualImage({ batchId: batch.id, bookId: book.id, videoId: video.id, shotId: shot.id, imageUrl });
      setGeneratedShotImages(current => ({ ...current, [shot.id]: imageUrl }));
      await runtime.saveDraft({ key: `shot:${shot.id}:visual-image`, kind: 'visual-image', scope: `${batch.id}:${book.id}`, content: imageUrl });
      message.success('当前 Shot 画面图已生成并保存');
      return true;
    } catch (error) {
      message.error(error?.message || '图片生成失败');
      return false;
    }
  }

  async function saveVideoPrompt(book, video, visualPrompt, shot = null) {
    const result = shot
      ? await runtime.saveDraft({
        key: `shot:${shot.id}`,
        kind: 'visual-prompt',
        scope: `${batch.id}:${book.id}`,
        content: visualPrompt
      })
      : await runtime.saveVideoPrompt({ batchId: batch.id, bookId: book.id, videoId: video.id, visualPrompt, revision: video.settingsState?.revision || 0 });
    if (!result.ok) { message.error(result.message); return false; }
    await reload({ announce: false });
    message.success(shot ? '当前 Shot 画面提示词已保存' : '画面提示词已保存');
    return true;
  }

  async function saveBookSource(book, sourceText) {
    const result = await runtime.updateBookSource({
      batchId: batch.id,
      bookId: book.id,
      sourceText,
      revision: book.revision
    });
    if (!result.ok) { message.error(result.message); return false; }
    await reload({ announce: false });
    message.success('原文已保存；已有下游结果已标记为需更新');
    return true;
  }

  async function saveAssetPrompts(book, type, items, drafts) {
    if (!book?.id) return false;
    try {
      await Promise.all((items || []).map(item => {
      const name = typeof item === 'string' ? item : (item?.name || item?.label || item?.id || '未命名资产');
      const key = `${type}:${item?.id || name}:${items.indexOf(item)}`;
      const content = drafts[key] ?? (typeof item === 'string' ? '' : (item?.prompt || item?.visualPrompt || item?.description || ''));
      return batchFactoryV11.saveDraft({ key: `asset:${type}:${item?.id || name}`, kind: 'asset-prompt', scope: `${batch.id}:${book.id}`, content });
      }));
      message.success(`${type === 'character' ? '人物' : type === 'scene' ? '场景' : '道具'} Prompt 草稿已保存`);
      return true;
    } catch (error) {
      message.error(error?.message || '资产 Prompt 草稿保存失败');
      return false;
    }
  }

  async function saveConstraintDraft(payload) {
    const result = await runtime.saveDraft(payload);
    if (!result.ok) { message.error(result.message); return false; }
    message.success('当前草稿已保存');
    return true;
  }

  async function savePersonalPrompt(payload) {
    const kind = String(payload?.kind || '');
    const category = kind.startsWith('constraint:') ? kind.slice('constraint:'.length) : '';
    let result;
    if (category) {
      try {
        const raw = await batchFactoryV11.savePersonalConstraintPrompt({
          category,
          name: payload?.name || null,
          body: payload?.content || ''
        });
        await reload({ announce: false });
        result = { ok: true, raw };
      } catch (error) {
        result = { ok: false, message: error?.message || '个人提示词保存失败，请稍后重试。' };
      }
    } else {
      result = await runtime.createPrompt(payload);
    }
    if (!result.ok) { message.error(result.message); return false; }
    message.success('已保存为我的提示词');
    return true;
  }

  async function openHistory() {
    const result = await runtime.listBatches();
    if (!result.ok) { message.error(result.message); return; }
    setHistoryBatches(result.batches || []);
    setHistoryOpen(true);
  }

  async function createNewBatch() {
    const result = await runtime.createBatch({ title: newBatchTitle.trim() || '未命名批次', books: [] });
    if (!result.ok) { message.error(result.message); return; }
    const createdID = createdBatchIdFrom(result.raw);
    setBatchManagerOpen(false);
    setNewBatchTitle('');
    const next = await runtime.load({ ...requestParams, batchId: createdID });
    setRuntimeState(next);
    if (next.phase === 'ready') message.success('新批次已创建');
  }

  async function openSchedules() {
    try {
      const result = await batchFactoryV11.listSchedules();
      setSchedules(result?.schedules || []);
      setScheduleOpen(true);
    } catch (error) { message.error(error?.message || '读取定时任务失败'); }
  }

  async function createProductionSchedule() {
    if (!scheduleAt || !batch?.id) return;
    try {
      const runAt = new Date(scheduleAt);
      if (!Number.isFinite(runAt.getTime()) || runAt.getTime() <= Date.now()) {
        message.error('请选择未来的执行时间');
        return;
      }
      const result = await batchFactoryV11.createSchedule({
        batchId: batch.id,
        requestId: newRequestId('bf11-scheduled-production'),
        bookId: '',
        provider: batchSettingsState.patch.videoProvider || 'personal_api',
        runAt: runAt.toISOString(),
        inputSnapshot: { settings: batchSettingsState.patch, promptSelections: batchSettingsState.patch.promptSelections || {} }
      });
      setSchedules(current => [...current, result.schedule]);
      setScheduleAt('');
      message.success('定时生产任务已创建');
    } catch (error) { message.error(error?.message || '创建定时任务失败'); }
  }

  async function cancelSchedule(schedule) {
    try {
      await batchFactoryV11.deleteSchedule(schedule.id);
      setSchedules(current => current.filter(item => item.id !== schedule.id));
      message.success('定时任务已删除');
    } catch (error) { message.error(error?.message || '删除定时任务失败'); }
  }

  function getPublishCredential(provider) { return runtime.getPublishCredential(provider); }
  function savePublishCredential(provider, payload) { return runtime.savePublishCredential(provider, payload); }
  function createPublishIntent(provider, payload) { return runtime.createPublishIntent(provider, payload); }
  function confirmPublishIntent(provider, intentId) { return runtime.confirmPublishIntent(provider, intentId); }
  function submitPublishIntent(provider, intentId) { return runtime.submitPublishIntent(provider, intentId); }

  return <DirectorRefreshProvider onRefresh={refreshDirectorRevision}>
    <div data-bf-v11-ui="final">
      <Space direction="vertical" size={12} style={{ width: '100%' }}>
        <Space wrap>
          <Button onClick={openSchedules}>自动化 / 定时生产</Button>
          <Typography.Text type="secondary">一次性日期时间执行；提示词选择和 API 模型配置会随任务快照保存。</Typography.Text>
        </Space>
        <BatchFactoryV11Workbench
        batch={viewBatch}
        books={books}
        productionStatus={runtimeState.productionStatus}
        mergeStatus={runtimeState.mergeStatus}
        capabilities={capabilities}
        onOpenBatchManager={() => setBatchManagerOpen(true)}
        onOpenHistory={openHistory}
        onOpenBatchSettings={() => setProductionSettingsOpen(true)}
        onOpenPublishSettings={() => setPublishSettingsOpen(true)}
        onOpenBookSettings={book => setBookSettingsTargetId(book.id)}
        onOpenVideoSettings={(book, video) => setVideoSettingsTarget({ bookId: book.id, videoId: video.id })}
        onRunHook={runHook}
        onApproveHook={approveHook}
        onRunDirector={runDirector}
        onRunBatchDirector={runBatchDirector}
        onPreviewFinalPrompt={previewFinalPrompt}
        onRunProduction={runProduction}
        onRunMerge={runMerge}
        onMeasureAudio={measureAudio}
        audioBusyBookId={audioBusyBookId}
        onRunUpload={() => setExternalPublishOpen(true)}
        onSaveSource={saveBookSource}
        onSaveVideoPrompt={saveVideoPrompt}
        onRefreshAssets={refreshAssets}
        onSaveAssetPrompts={saveAssetPrompts}
        onGenerateShotImage={generateShotImage}
        generatedShotImages={generatedShotImages}
        onGenerateAssetImage={generateAssetImage}
        generatedAssetImages={generatedAssetImages}
      />
      </Space>

      <Modal title="自动化 / 定时生产" open={scheduleOpen} onCancel={() => setScheduleOpen(false)} footer={null}>
        <Space direction="vertical" style={{ width: '100%' }}>
          <Typography.Text strong>执行时间</Typography.Text>
          <Input type="datetime-local" value={scheduleAt} onChange={event => setScheduleAt(event.target.value)} />
          <Button type="primary" disabled={!scheduleAt} onClick={createProductionSchedule}>创建一次性定时任务</Button>
          <Typography.Text strong>定时任务</Typography.Text>
          {(schedules || []).map(schedule => <Space key={schedule.id} style={{ width: '100%', justifyContent: 'space-between' }}>
            <Typography.Text>{new Date(schedule.runAt).toLocaleString()} · {schedule.status}</Typography.Text>
            {schedule.status === 'scheduled' ? <Button danger size="small" onClick={() => cancelSchedule(schedule)}>删除</Button> : null}
          </Space>)}
          {!schedules.length ? <Typography.Text type="secondary">暂无定时任务</Typography.Text> : null}
        </Space>
      </Modal>

      <Modal title="新建批次" open={batchManagerOpen} onCancel={() => setBatchManagerOpen(false)} onOk={createNewBatch} okText="创建">
        <Input placeholder="批次名称" value={newBatchTitle} onChange={event => setNewBatchTitle(event.target.value)} />
        <Typography.Text type="secondary">创建后可从小说获取导入书目，或在当前批次继续配置。</Typography.Text>
      </Modal>

      <Modal title="历史批次" open={historyOpen} footer={null} onCancel={() => setHistoryOpen(false)}>
        <Space direction="vertical" style={{ width: '100%' }}>
          {(historyBatches || []).map(record => <Button key={record.id} block onClick={async () => { const next = await runtime.load({ ...requestParams, batchId: record.id }); setRuntimeState(next); setHistoryOpen(false); }}>{record.title || record.id} · {record.count || record.books?.length || 0} 本</Button>)}
          {!historyBatches.length ? <Empty description="暂无历史批次" /> : null}
        </Space>
      </Modal>

      <FinalPromptPreviewDrawer
        open={promptPreview.open}
        loading={promptPreview.loading}
        data={promptPreview.data}
        error={promptPreview.error}
        onClose={() => setPromptPreview(current => ({ ...current, open: false }))}
      />

      <ProductionSettingsDrawer
        open={productionSettingsOpen}
        batch={viewBatch}
        configVersions={runtimeState.configVersions || []}
        configVersionsError={runtimeState.configVersionsError || null}
        apiModels={runtimeState.apiModels || {}}
        promptCatalog={runtimeState.promptCatalog || []}
        personalPrompts={runtimeState.personalPrompts || {}}
        personalPromptsError={runtimeState.personalPromptsError || null}
        videoProviders={runtimeState.videoProviders || {}}
        localExecutors={runtimeState.localExecutors || []}
        onCreateLocalExecutorPairing={runtime.createLocalExecutorPairing}
        initialValue={batchSettingsState.patch}
        onClose={() => setProductionSettingsOpen(false)}
        onPreviewChangeImpact={previewBatchChangeImpact}
        onSyncConfigVersion={syncBatchConfigVersion}
        onSaveDraft={saveConstraintDraft}
        onSavePersonalPrompt={savePersonalPrompt}
        onSave={saveBatchSettings}
      />

      <PublishSettingsDrawer
        open={publishSettingsOpen}
        batch={viewBatch}
        initialValue={batchSettingsState.patch.publishSettings || undefined}
        onClose={() => setPublishSettingsOpen(false)}
        onSave={async value => {
          const ok = await saveScope({ scope: 'batch', batchId: batch.id, patch: { publishSettings: value }, revision: batchSettingsState.revision }, '发布统一设置已保存');
          return ok;
        }}
        onSync={async value => {
          await saveScope({ scope: 'batch', batchId: batch.id, patch: { publishSettings: value }, revision: batchSettingsState.revision }, '批量发布配置已同步');
        }}
      />

      <ExternalPublishPanel
        open={externalPublishOpen}
        batch={viewBatch}
        books={books}
        productionStatus={runtimeState.productionStatus}
        mergeStatus={runtimeState.mergeStatus}
        capabilities={capabilities}
        onClose={() => setExternalPublishOpen(false)}
        onGetCredential={getPublishCredential}
        onSaveCredential={savePublishCredential}
        onCreateIntent={createPublishIntent}
        onConfirmIntent={confirmPublishIntent}
        onSubmitIntent={submitPublishIntent}
      />

      <BookSettingsModal
        open={Boolean(activeBook)}
        book={activeBook}
        initialPatch={activeBook?.settingsState?.patch || {}}
        onClose={() => setBookSettingsTargetId('')}
        onSave={saveBookSettings}
      />

      <VideoSettingsDrawer
        open={Boolean(activeVideoBook && activeVideo)}
        book={activeVideoBook}
        video={activeVideo}
        modelMaxDuration={Number(batch.modelMaxDuration || 0)}
        initialPatch={activeVideo?.settingsState?.patch || {}}
        onClose={() => setVideoSettingsTarget(null)}
        onSave={saveVideoSettings}
      />
    </div>
  </DirectorRefreshProvider>;
}

export default BatchFactoryV11UiPage;
