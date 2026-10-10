import { AppstoreOutlined, ArrowLeftOutlined, BarsOutlined, DownloadOutlined, ExportOutlined, FileSearchOutlined, FileTextOutlined, OrderedListOutlined, PictureOutlined, SettingOutlined, SoundOutlined, UploadOutlined, VideoCameraOutlined } from '@ant-design/icons';
import { Alert, Button, Checkbox, Dropdown, Input, InputNumber, Modal, Select, Tooltip, message } from 'antd';
import { useEffect, useMemo, useRef, useState } from 'react';
import { AiReasoningModal } from './AiReasoningModal';
import { BatchTaskModal } from './BatchTaskModal';
import { EngineSettingsModal } from './EngineSettingsModal';
import { MediaModal } from './MediaModal';
import { MediaPreview, StoryboardRow } from './StoryboardRow';
import { TaskDrawer } from './TaskDrawer';
import { taskReadiness, textReadiness } from './taskReadiness';
import { attachMedia, confirmSegmentation, createTask, deleteSegment, downloadGeneratedAssetImage, exportProject, fixedSegmentation, getProductionConfig, getProject, importSegmentation, insertStoryboard, listAssetImages, listModels, listSegmentVoiceSettings, replaceSegmentAssets, saveSegmentVoiceSettings, smartSegmentation, updateAsset, updateSegment, uploadMedia } from '../../../../shared/api/shuihuoProduction';
import { textToSpeech } from '../../../../shared/api/tts';
import { getConfig } from '../../../../shared/api/config';
import { audioBlobToDataUrl, narrationFilename, resolveNarrationSettings, resolveSpeakerVoiceAsset } from './directNarration';
import { defaultModelIdForKind } from './modelDefaults';
import { contiguousFollowingSceneSegments, sameAssetSet, sceneAssetIdsForSegment } from './sceneContinuity';

const blankInsert = { sourceText: '', subtitleText: '' };
const safeMedia = item => item?.media || item;
const defaultVoiceSettings = { voiceAssetId: null, speechRate: 1, pitch: 0 };
const segmentationPromptOptions = [{ value: 'default', label: '选择提示词' }];
const assetCategoryLabels = { character: '角色', scene: '场景', prop: '道具' };

function countMedia(items, kind) { return (items || []).map(safeMedia).filter(item => item?.kind === kind).length; }
function fileToDataURL(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => reject(new Error('读取音频文件失败'));
    reader.onload = () => resolve(reader.result);
    reader.readAsDataURL(file);
  });
}

function BindingAssetPreview({ asset, onPreview }) {
  const [url, setUrl] = useState('');

  useEffect(() => {
    let active = true;
    let objectUrl = '';
    setUrl('');
    listAssetImages(asset.id).then(result => {
      const image = (result.images || []).find(item => item.isPrimary) || result.images?.[0];
      if (!image) return null;
      return downloadGeneratedAssetImage(image.id);
    }).then(blob => {
      if (!blob || !active) return;
      objectUrl = URL.createObjectURL(blob);
      setUrl(objectUrl);
    }).catch(() => { if (active) setUrl(''); });
    return () => {
      active = false;
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [asset.id]);

  if (url) return <button type="button" className="shuihuo-binding-asset-preview" title="放大查看预设图" onClick={event => { event.preventDefault(); event.stopPropagation(); onPreview(asset); }}><img src={url} alt={`${asset.name}预设图`} /></button>;
  return <small>{asset.prompt || '未填写预设提示词'}</small>;
}

export function CommentaryWorkbench({ data, readiness, importNotice, onBackToProjects, onOpenAssets, onDataChange }) {
  const [segmenting, setSegmenting] = useState(false);
  const [segmentMode, setSegmentMode] = useState('fixed');
  const [segmentationPrompt, setSegmentationPrompt] = useState('default');
  const [candidateText, setCandidateText] = useState('');
  const [candidates, setCandidates] = useState([]);
  const [lines, setLines] = useState(3);
  const [textModels, setTextModels] = useState([]);
  const [textModelId, setTextModelId] = useState();
  const [editing, setEditing] = useState(null);
  const [binding, setBinding] = useState(null);
  const [selectedAssetIds, setSelectedAssetIds] = useState([]);
  const [insertAfter, setInsertAfter] = useState(null);
  const [batch, setBatch] = useState(null);
  const [aiReasoningOpen, setAiReasoningOpen] = useState(false);
  const [tasksOpen, setTasksOpen] = useState(false);
  const [taskFilter, setTaskFilter] = useState('all');
  const [engineOpen, setEngineOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const [exporting, setExporting] = useState(false);
  const [expandedField, setExpandedField] = useState(null);
  const [mediaUpload, setMediaUpload] = useState(null);
  const [mediaPreview, setMediaPreview] = useState(null);
  const [assetPreview, setAssetPreview] = useState(null);
  const [voiceSettingsBySegment, setVoiceSettingsBySegment] = useState({});
  const [voiceConfiguring, setVoiceConfiguring] = useState(null);
  const [narratingSegmentIds, setNarratingSegmentIds] = useState(() => new Set());
  const [submittingTaskSegmentIds, setSubmittingTaskSegmentIds] = useState(() => new Set());
  const segmentationFileRef = useRef(null);
  const narrationImportRef = useRef(null);
  const project = data.project;
  const media = data.media || [];
  const bindings = data.segmentAssetIDs || {};
  const imageAvailability = taskReadiness(readiness, 'image', { confirmed: project.segmentationStatus === 'confirmed' });
  const videoAvailability = taskReadiness(readiness, 'video', { confirmed: project.segmentationStatus === 'confirmed' });
  const textAvailability = textReadiness(readiness);
  const imageReady = imageAvailability.ready;
  const videoReady = videoAvailability.ready;
  // Narration uses the already-configured /api/tts service directly. It does
  // not enter the Shuihuo Redis task queue or require an audio model entry.
  const directNarrationReady = project.segmentationStatus === 'confirmed';
  const directNarrationReason = directNarrationReady ? '' : '请先确认分镜后再生成配音';
  const textReady = textAvailability.ready;
  const confirmed = project.segmentationStatus === 'confirmed';
  const voiceAssets = useMemo(() => (data.assets || []).filter(asset => asset.category === 'voice'), [data.assets]);
  const stats = useMemo(() => ({ storyboard: data.segments?.length || 0, image: countMedia(media, 'image'), video: countMedia(media, 'video'), audio: countMedia(media, 'audio') }), [data.segments, media]);
  const completedVideoSegmentIds = useMemo(() => new Set(media.map(safeMedia).filter(item => item?.kind === 'video').map(item => item?.segmentId).filter(Boolean)), [media]);
  const completionPercent = stats.storyboard ? Math.round((completedVideoSegmentIds.size / stats.storyboard) * 100) : 0;

  useEffect(() => {
    if (!segmenting) return undefined;
    let active = true;
    listModels().then(result => { if (active) setTextModels((result.models || []).filter(model => model.kind === 'text')); }).catch(() => { if (active) setTextModels([]); });
    return () => { active = false; };
  }, [segmenting]);

  useEffect(() => {
    if (!segmenting) return;
    setCandidates((data.segments || []).map(segment => ({ text: segment.subtitleText || segment.sourceText || '' })));
    setCandidateText(project.sourceText || (data.segments || []).map(segment => segment.sourceText || segment.subtitleText || '').join('\n'));
  }, [data.segments, project.sourceText, segmenting]);

  useEffect(() => {
    let active = true;
    setVoiceConfiguring(null);
    setVoiceSettingsBySegment({});
    listSegmentVoiceSettings(project.id).then(result => {
      if (active) setVoiceSettingsBySegment(result.settings || {});
    }).catch(error => { if (active) message.error(error.message || '读取配音设置失败'); });
    return () => { active = false; };
  }, [project.id]);

  useEffect(() => () => {
    if (assetPreview?.url) URL.revokeObjectURL(assetPreview.url);
  }, [assetPreview?.url]);

  function boundAssets(segment) { return (bindings[segment.id] || []).map(id => data.assets?.find(asset => asset.id === id)).filter(Boolean); }
  function settingsFor(segmentId) { return { ...defaultVoiceSettings, ...(voiceSettingsBySegment[segmentId] || {}) }; }
  function selectedVoiceAsset(segment, settings) {
    return resolveSpeakerVoiceAsset({ speaker: segment?.speaker, assets: data.assets || [], narratorVoiceAssetId: settings.voiceAssetId });
  }
  function openVoiceSettings(segment, focus = 'voice') {
    const settings = settingsFor(segment.id);
    const voice = selectedVoiceAsset(segment, settings);
    setVoiceConfiguring({ segment, focus, settings: { ...settings, voiceAssetId: voice?.id ?? settings.voiceAssetId } });
  }
  async function openAssetPreview(asset) {
    setAssetPreview({ asset, url: '' });
    try {
      const result = await listAssetImages(asset.id);
      const image = (result.images || []).find(item => item.isPrimary) || result.images?.[0];
      if (!image) throw new Error('该预设还没有生成图片');
      const blob = await downloadGeneratedAssetImage(image.id);
      const url = URL.createObjectURL(blob);
      setAssetPreview(current => {
        if (current?.url) URL.revokeObjectURL(current.url);
        return { asset, url };
      });
    } catch (error) {
      setAssetPreview(null);
      message.error(error.message || '读取预设图片失败');
    }
  }
  async function generateNarrationForSegment(segmentId, { refresh = true } = {}) {
    const segment = (data.segments || []).find(item => item.id === segmentId);
    const input = String(segment?.subtitleText || segment?.sourceText || '').trim();
    if (!segment || !input) throw new Error('该分镜没有可配音的字幕');

    setNarratingSegmentIds(current => new Set([...current, segmentId]));
    try {
      const config = await getConfig();
      const settings = resolveNarrationSettings({
        voiceAsset: selectedVoiceAsset(segment, settingsFor(segmentId)),
        settings: settingsFor(segmentId),
        defaults: config?.tts || {}
      });
      const blob = await textToSpeech({ input, ...settings });
      const dataUrl = await audioBlobToDataUrl(blob);
      await uploadMedia(project.id, { kind: 'audio', filename: narrationFilename(segmentId), dataUrl, segmentId });
      if (refresh) await refreshProject();
    } finally {
      setNarratingSegmentIds(current => {
        const next = new Set(current);
        next.delete(segmentId);
        return next;
      });
    }
  }
  async function generateNarrations(segmentIds) {
    const uniqueIds = [...new Set(segmentIds || [])];
    if (!uniqueIds.length) { message.warning('没有可生成配音的分镜'); return; }
    let completed = 0;
    const failures = [];
    for (const segmentId of uniqueIds) {
      try {
        await generateNarrationForSegment(segmentId, { refresh: false });
        completed += 1;
      } catch (error) {
        failures.push(error.message || `分镜 ${segmentId} 配音失败`);
      }
    }
    if (completed) await refreshProject();
    if (completed) message.success(`已生成并保存 ${completed} 条配音`);
    if (failures.length) message.error(`${failures.length} 条配音未生成：${failures[0]}`);
  }
  async function importNarrationFiles(event) {
    const files = [...(event.target.files || [])];
    event.target.value = '';
    if (!files.length || !data.segments?.length) return;
    setBusy(true);
    let imported = 0;
    const failures = [];
    try {
      for (const file of files) {
        const numberMatch = file.name.match(/(?:^|[^0-9])(\d{1,4})(?:[^0-9]|$)/);
        const numberIndex = numberMatch ? Number(numberMatch[1]) - 1 : -1;
        const target = numberIndex >= 0 && numberIndex < data.segments.length
          ? data.segments[numberIndex]
          : data.segments.find(segment => file.name.includes(segment.id));
        if (!target) {
          failures.push(`${file.name}（无法匹配分镜）`);
          continue;
        }
        try {
          await uploadMedia(project.id, { kind: 'audio', filename: file.name, dataUrl: await fileToDataURL(file), segmentId: target.id });
          imported += 1;
        } catch (error) {
          failures.push(`${file.name}（${error.message || '上传失败'}）`);
        }
      }
      if (imported) {
        await refreshProject();
        message.success(`已导入 ${imported} 条配音并绑定分镜`);
      }
      if (failures.length) message.warning(`有 ${failures.length} 个文件未导入：${failures[0]}`);
    } finally {
      setBusy(false);
    }
  }
  async function saveVoiceSettings() {
    if (!voiceConfiguring) return;
    const { segment, settings } = voiceConfiguring;
    const speaker = String(segment.speaker || '旁白').trim() || '旁白';
    setBusy(true);
    try {
      if (speaker !== '旁白') {
        const character = (data.assets || []).find(asset => asset.category === 'character' && asset.name === speaker);
        if (!character) throw new Error(`未找到“${speaker}”角色预设，无法绑定角色音色`);
        await updateAsset(character.id, { ...character, voiceAssetId: settings.voiceAssetId ?? null });
      }
      const saved = await saveSegmentVoiceSettings(segment.id, {
        voiceAssetId: speaker === '旁白' ? settings.voiceAssetId ?? null : null,
        speechRate: Number(settings.speechRate ?? 1),
        pitch: Number(settings.pitch ?? 0)
      });
      setVoiceSettingsBySegment(current => ({ ...current, [segment.id]: { ...defaultVoiceSettings, ...saved } }));
      setVoiceConfiguring(null);
      await refreshProject();
      message.success('配音设置已保存');
    } catch (error) {
      message.error(error.message || '保存配音设置失败');
    } finally {
      setBusy(false);
    }
  }
  async function refreshProject() { onDataChange(await getProject(project.id)); }
  function applyReadModel(next) { onDataChange(next); }
  async function previewSegmentation(nextMode = segmentMode) {
    setBusy(true);
    try {
      if (nextMode === 'smart' && !textModelId) { message.warning('请选择管理员启用的文本模型'); return; }
      const text = candidateText || project.sourceText;
      const result = nextMode === 'smart'
        ? await smartSegmentation(project.id, { text: candidateText, modelId: textModelId })
        : nextMode === 'import'
          ? await importSegmentation(project.id, { text })
          : await fixedSegmentation(project.id, { text, linesPerSegment: lines });
      setSegmentMode(nextMode);
      setCandidates(result.candidates || []);
    } catch (error) { message.error(error.message || '生成分镜候选失败'); } finally { setBusy(false); }
  }
  async function importSegmentationFile(event) {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file) return;
    setBusy(true);
    try {
      const text = await file.text();
      setCandidateText(text);
      const result = await importSegmentation(project.id, { text });
      setSegmentMode('import');
      setCandidates(result.candidates || []);
      message.success(`已导入 ${result.candidates?.length || 0} 条分镜`);
    } catch (error) {
      message.error(error.message || '导入分镜失败');
    } finally {
      setBusy(false);
    }
  }
  function updateCandidate(index, text) {
    setCandidates(items => items.map((item, itemIndex) => itemIndex === index ? { ...item, text } : item));
  }
  function exportSegmentationCandidates() {
    const rows = candidates.map((candidate, index) => `${index + 1}\t${candidate.text || ''}`).join('\n');
    const blob = new Blob([rows], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `${project.name || '漫剧解说'}-分镜.txt`;
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.setTimeout(() => URL.revokeObjectURL(url), 60000);
    message.success('分镜已导出');
  }
  async function confirmCandidates() {
    const nextCandidates = candidates.map(candidate => ({ text: candidate.text?.trim() || '', speaker: candidate.speaker?.trim() || '旁白' })).filter(candidate => candidate.text);
    if (!nextCandidates.length) { message.warning('请先生成或导入候选分镜'); return; }
    setBusy(true);
    try {
      await confirmSegmentation(project.id, nextCandidates);
      setSegmenting(false);
      setCandidates([]);
      await refreshProject();
      message.success('已确认分镜');
    } catch (error) { message.error(error.message || '确认分镜失败'); } finally { setBusy(false); }
  }
  async function saveEditing() {
    if (!editing) return;
    setBusy(true);
    try {
      await updateSegment(editing.id, {
        ...editing,
        subtitleText: editing.subtitleText || '',
		speaker: editing.speaker || '旁白',
        imagePrompt: editing.imagePrompt || '',
        videoPrompt: editing.videoPrompt || '',
        imagePromptLocked: Boolean(editing.imagePromptLocked),
        videoPromptLocked: Boolean(editing.videoPromptLocked)
      });
      setEditing(null); await refreshProject(); message.success('已保存分镜文案');
    } catch (error) { message.error(error.message || '保存分镜文案失败'); } finally { setBusy(false); }
  }
  async function saveExpandedField() {
    if (!expandedField) return;
    const { segment, field, value } = expandedField;
    setBusy(true);
    try {
      await updateSegment(segment.id, {
        ...segment,
        subtitleText: field === 'subtitleText' ? value : segment.subtitleText || '',
		speaker: segment.speaker || '旁白',
        imagePrompt: field === 'imagePrompt' ? value : segment.imagePrompt || '',
        videoPrompt: field === 'videoPrompt' ? value : segment.videoPrompt || '',
        imagePromptLocked: Boolean(segment.imagePromptLocked),
        videoPromptLocked: Boolean(segment.videoPromptLocked)
      });
      setExpandedField(null);
      await refreshProject();
      message.success('已保存编辑内容');
    } catch (error) {
      message.error(error.message || '保存编辑内容失败');
    } finally {
      setBusy(false);
    }
  }
  function openBinding(segment, category) {
    const assetById = new Map((data.assets || []).map(asset => [asset.id, asset]));
    setBinding({ segment, category });
    setSelectedAssetIds((bindings[segment.id] || []).filter(id => assetById.get(id)?.category === category));
  }
  function sceneRangeLabel(segments) {
    const orders = segments.map(segment => Number(segment.orderIndex)).filter(Number.isFinite);
    if (!orders.length) return '后续连续分镜';
    const first = orders[0];
    const last = orders[orders.length - 1];
    return first === last ? `第${first}段` : `第${first}-${last}段`;
  }
  function askSceneContinuity(segments) {
    const range = sceneRangeLabel(segments);
    return new Promise(resolve => {
      Modal.confirm({
        title: '场景已更新',
        content: `检测到后面还有连续分镜没有明确换场，是否将新场景沿用到第${range.replace(/^第/, '')}？`,
        okText: `沿用到第${range.replace(/^第/, '')}`,
        cancelText: '只更新当前段',
        onOk: () => resolve(true),
        onCancel: () => resolve(false)
      });
    });
  }
  async function saveBinding() {
    if (!binding) return;
    const assetById = new Map((data.assets || []).map(asset => [asset.id, asset]));
    const previousSceneIds = sceneAssetIdsForSegment(binding.segment.id, bindings, data.assets);
    const nextSceneIds = [...selectedAssetIds].map(Number).sort((left, right) => left - right);
    let continuitySegments = [];
    let propagate = false;
    if (binding.category === 'scene' && !sameAssetSet(previousSceneIds, nextSceneIds)) {
      continuitySegments = contiguousFollowingSceneSegments(data.segments || [], bindings, data.assets, binding.segment.id);
      propagate = continuitySegments.length > 0 && await askSceneContinuity(continuitySegments);
    }
    setBusy(true);
    try {
      const targets = [binding.segment, ...(propagate ? continuitySegments : [])];
      for (const target of targets) {
        const retained = (bindings[target.id] || []).filter(id => assetById.get(id)?.category !== binding.category);
        const assetIds = [...new Set([...retained, ...selectedAssetIds])];
        await replaceSegmentAssets(target.id, assetIds);
      }
      setBinding(null);
      await refreshProject();
      message.success(propagate ? `已更新场景，并沿用到${sceneRangeLabel(continuitySegments)}` : `已保存${assetCategoryLabels[binding.category]}预设绑定`);
    } catch (error) { message.error(error.message || '保存预设绑定失败'); } finally { setBusy(false); }
  }
  async function saveInsert() {
    if (!insertAfter?.sourceText?.trim()) { message.warning('请填写新增分镜的原文'); return; }
    setBusy(true);
    try { applyReadModel(await insertStoryboard(insertAfter.id, { sourceText: insertAfter.sourceText.trim(), subtitleText: insertAfter.subtitleText || '' })); setInsertAfter(null); message.success('已新增分镜'); } catch (error) { message.error(error.message || '新增分镜失败'); } finally { setBusy(false); }
  }
  async function saveSpeaker(segment, speaker) {
    setBusy(true);
    try {
      applyReadModel(await updateSegment(segment.id, {
        ...segment,
        speaker: speaker || '旁白',
        subtitleText: segment.subtitleText || '',
        imagePrompt: segment.imagePrompt || '',
        videoPrompt: segment.videoPrompt || '',
        imagePromptLocked: Boolean(segment.imagePromptLocked),
        videoPromptLocked: Boolean(segment.videoPromptLocked)
      }));
      message.success('发言者已更新');
    } catch (error) { message.error(error.message || '更新发言者失败'); } finally { setBusy(false); }
  }
  async function saveMediaUpload(payload) {
    setBusy(true);
    try {
      const result = await uploadMedia(project.id, payload);
      if (payload.segmentId) await attachMedia(result.media.id, payload.segmentId);
      setMediaUpload(null);
      await refreshProject();
      message.success('素材已上传并绑定分镜');
    } catch (error) {
      message.error(error.message || '上传素材失败');
    } finally {
      setBusy(false);
    }
  }
  async function removeStoryboard(segmentId) { await deleteSegment(segmentId); return getProject(project.id); }
  async function createConfiguredTask(kind, segmentId) {
    const label = kind === 'image' ? '图片' : kind === 'video' ? '视频' : '任务';
    const segmentOrder = (data.segments || []).find(segment => segment.id === segmentId)?.orderIndex || segmentId;
    setSubmittingTaskSegmentIds(current => new Set([...current, segmentId]));
    try {
      const [config, modelResult] = await Promise.all([getProductionConfig(), listModels()]);
      const models = (modelResult.models || []).filter(model => model?.kind === kind);
      const modelId = defaultModelIdForKind(models, kind, config?.[`${kind}ModelId`]);
      if (modelId === undefined || modelId === null) {
        throw new Error(`请先在引擎配置中选择${label}模型`);
      }
      const selectedModel = models.find(model => model.id === modelId);
      if (kind === 'video' && config?.videoGenerationMode === 'text_to_video' && selectedModel?.adapterKind === 'yd_video') {
        throw new Error('YD2.0 Mini 仅支持图生视频，请在引擎配置中切换图生视频或选择文生视频模型');
      }
      await createTask(project.id, { segmentId, kind, modelId });
      await refreshProject();
      message.success(`已提交第 ${segmentOrder} 个分镜${label}任务`);
    } catch (error) {
      message.error(error.message || `${label}任务提交失败`);
    } finally {
      setSubmittingTaskSegmentIds(current => {
        const next = new Set(current);
        next.delete(segmentId);
        return next;
      });
    }
  }
  async function downloadProjectExport() {
    setExporting(true);
    try {
      const blob = await exportProject(project.id);
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `${project.name || '漫剧解说'}-漫剧解说.zip`;
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.setTimeout(() => URL.revokeObjectURL(url), 60000);
      message.success('项目素材已开始下载');
    } catch (error) {
      message.error(error.message || '导出项目素材失败');
    } finally {
      setExporting(false);
    }
  }
  function openActiveTasks() {
    setTaskFilter('active');
    setTasksOpen(true);
  }
  function openAllTasks() {
    setTaskFilter('all');
    setTasksOpen(true);
  }
  function openAudioBatchFromSegmentation() {
    setSegmenting(false);
    generateNarrations(data.segments?.map(item => item.id) || []);
  }
  const batchItems = [{ key: 'image', icon: <PictureOutlined />, label: '批量生成图片', disabled: !imageReady }, { key: 'video', icon: <VideoCameraOutlined />, label: '批量生成视频', disabled: !videoReady }, { key: 'audio', icon: <SoundOutlined />, label: '批量生成配音', disabled: !directNarrationReady }];
  const bindingAssets = binding ? (data.assets || []).filter(asset => asset.category === binding.category) : [];

  return <section className="shuihuo-workbench">
    <header className="shuihuo-workbench-header"><div className="shuihuo-workbench-heading"><button className="shuihuo-back-link" type="button" title="返回项目库" aria-label="返回项目库" onClick={onBackToProjects}><ArrowLeftOutlined /></button><strong>{project.name}</strong><div className="shuihuo-workbench-progress" aria-label={`完成度 ${completionPercent}%`}><div className="shuihuo-workbench-progress-track"><i style={{ width: `${completionPercent}%` }} /></div><span>{completionPercent}%</span></div></div><div className="shuihuo-workbench-toolbar" role="toolbar" aria-label="工作台工具栏"><Button type="text" icon={<BarsOutlined />} onClick={() => { setSegmentMode('fixed'); setSegmenting(true); }}>分镜调整</Button><Button type="text" icon={<AppstoreOutlined />} onClick={onOpenAssets}>人物场景预设</Button><Button type="text" icon={<SettingOutlined />} onClick={() => setEngineOpen(true)}>引擎配置</Button><Tooltip title={textReady ? '按模型生成并应用提示词候选' : '需要可用的文本模型'}><Button type="text" icon={<FileSearchOutlined />} disabled={!confirmed || !textReady} onClick={() => setAiReasoningOpen(true)}>AI 推理</Button></Tooltip><Dropdown menu={{ items: batchItems, onClick: ({ key }) => key === 'audio' ? generateNarrations(data.segments?.map(item => item.id) || []) : setBatch({ kind: key, segmentIds: data.segments?.map(item => item.id) || [], initialScope: 'all' }) }}><Button className="shuihuo-batch-button" type="text" icon={<PictureOutlined />} disabled={!confirmed}>批量操作</Button></Dropdown><Tooltip title="查看排队或运行中的任务并逐项取消"><Button className="shuihuo-cancel-button" type="text" onClick={openActiveTasks}>取消操作</Button></Tooltip></div><div className="shuihuo-workbench-export"><Button type="text" icon={<BarsOutlined />} onClick={openAllTasks}>任务/日志</Button><Tooltip title={confirmed ? '导出已确认分镜、字幕和已保存素材' : '请先确认至少一个分镜'}><Button icon={<ExportOutlined />} disabled={!confirmed || exporting} loading={exporting} onClick={downloadProjectExport}>导出剪映</Button></Tooltip></div><div className="shuihuo-project-stats"><span><b>{stats.storyboard}</b> 分镜</span><span><b>{stats.image}</b> 图片</span><span><b>{stats.video}</b> 视频</span><span><b>{stats.audio}</b> 音频</span></div>{importNotice ? <span className="shuihuo-import-success">成功导入 {importNotice} 条文本</span> : null}</header>
    {!confirmed ? <Alert showIcon type="info" className="shuihuo-workbench-notice" message="分镜尚未确认" description="候选分镜可反复修改；确认前不允许提交图片、视频或配音任务。" /> : null}
    <div className="shuihuo-workbench-table" role="table" aria-label="漫剧解说分镜生产表"><div className="shuihuo-workbench-head" role="row">{['序号', '字幕', '配音', '预设', '提示词', '片段库', '操作'].map(item => <div role="columnheader" key={item}>{item}</div>)}</div>{(data.segments || []).map((segment, index) => { const settings = settingsFor(segment.id); const asset = selectedVoiceAsset(segment, settings); const speakers = [...new Set(['旁白', segment.speaker || '旁白', ...boundAssets(segment).filter(item => item.category === 'character').map(item => item.name)])]; return <StoryboardRow key={segment.id} index={index} segment={segment} assets={boundAssets(segment)} speakerOptions={speakers} media={media} sourceUnitIds={data.segmentSourceUnitIDs?.[segment.id] || []} voiceSettings={settings} voiceLabel={asset?.name} imageReady={imageReady} imageUnavailableReason={imageAvailability.reason} videoReady={videoReady} videoUnavailableReason={videoAvailability.reason} audioReady={directNarrationReady} audioUnavailableReason={directNarrationReason} narrationGenerating={narratingSegmentIds.has(segment.id)} taskSubmitting={submittingTaskSegmentIds.has(segment.id)} onDataChange={applyReadModel} onRefresh={refreshProject} onConfigureVoice={openVoiceSettings} onChangeSpeaker={saveSpeaker} onExpandField={({ segment: target, field, label }) => setExpandedField({ segment: target, field, label, value: target[field] || (field === 'subtitleText' ? target.sourceText || '' : '') })} onBindAssets={openBinding} onUploadMedia={(kind, target) => setMediaUpload({ kind, segmentId: target.id })} onPreviewMedia={setMediaPreview} onGenerateNarration={target => generateNarrations([target.id])} onTask={createConfiguredTask} onDelete={removeStoryboard} onInsertAfter={target => setInsertAfter({ id: target.id, ...blankInsert })} />; })}{!data.segments?.length ? <div className="shuihuo-workbench-empty">还没有分镜。使用顶部“分镜调整”创建候选并确认。</div> : null}</div>
    <Modal
      title={<span className="shuihuo-segmentation-title">分镜调整 <Tooltip title="导入、AI 分镜或按行拆分后，点击确认应用才会覆盖当前分镜。">?</Tooltip></span>}
      open={segmenting}
      onCancel={() => setSegmenting(false)}
      width={1080}
      className="shuihuo-segmentation-modal"
      footer={<><span className="shuihuo-segmentation-count">共 {candidates.length} 个分镜，{candidates.length} 条字幕</span><div><Button onClick={() => setSegmenting(false)}>取消</Button><Button type="primary" loading={busy} onClick={confirmCandidates}>确定应用</Button></div></>}
    >
      <input ref={segmentationFileRef} type="file" accept=".txt,.srt,text/plain,application/x-subrip" hidden onChange={importSegmentationFile} />
      <div className="shuihuo-segmentation-toolbar">
        <Button icon={<UploadOutlined />} onClick={() => segmentationFileRef.current?.click()}>导入分镜</Button>
        <Select value={textModelId} onChange={setTextModelId} placeholder={textModels.length ? '选择文本模型' : 'Gemini-3-Pro-Pr...'} options={textModels.map(model => ({ value: model.id, label: model.name }))} />
        <Select value={segmentationPrompt} onChange={setSegmentationPrompt} options={segmentationPromptOptions} />
        <Button className="shuihuo-segmentation-primary" icon={<FileSearchOutlined />} disabled={!textReady || !textModelId} loading={busy && segmentMode === 'smart'} onClick={() => previewSegmentation('smart')}>AI分镜</Button>
        <InputNumber min={1} max={20} value={lines} onChange={value => setLines(value || 1)} addonBefore="行" />
        <Button icon={<OrderedListOutlined />} loading={busy && segmentMode === 'fixed'} onClick={() => previewSegmentation('fixed')}>按行拆分</Button>
        <Tooltip title={directNarrationReady ? '按各分镜的音色、语速和音调直接生成并保存配音' : directNarrationReason}><Button icon={<SoundOutlined />} disabled={!directNarrationReady} onClick={openAudioBatchFromSegmentation}>智能配音</Button></Tooltip>
        <input ref={narrationImportRef} type="file" accept="audio/mpeg,audio/wav,audio/x-wav,audio/mp4" multiple hidden onChange={importNarrationFiles} />
        <Tooltip title="按文件名中的序号（如 01、02）自动匹配分镜，也支持直接使用分镜 ID"><Button icon={<FileTextOutlined />} disabled={!confirmed || busy} onClick={() => narrationImportRef.current?.click()}>导入配音</Button></Tooltip>
        <Button icon={<DownloadOutlined />} disabled={!candidates.length} onClick={exportSegmentationCandidates}>导出分镜</Button>
      </div>
      <div className="shuihuo-segmentation-table" role="table" aria-label="分镜调整候选表">
        <div className="shuihuo-segmentation-head" role="row"><strong role="columnheader">序号</strong><strong role="columnheader">字幕</strong></div>
        <div className="shuihuo-segmentation-body">
          {candidates.map((candidate, index) => <div className="shuihuo-segmentation-row" role="row" key={index}>
            <span role="cell">{index + 1}</span>
            <textarea aria-label={`分镜 ${index + 1} 字幕`} value={candidate.text || ''} onChange={event => updateCandidate(index, event.target.value)} />
          </div>)}
          {!candidates.length ? <div className="shuihuo-segmentation-empty">暂无分镜，导入文本或点击 AI分镜 / 按行拆分生成候选。</div> : null}
        </div>
      </div>
    </Modal>
    <Modal title="编辑字幕与提示词" open={Boolean(editing)} onCancel={() => setEditing(null)} onOk={saveEditing} okText="保存" confirmLoading={busy} width={760}><label className="shuihuo-form-label">字幕</label><Input.TextArea value={editing?.subtitleText || ''} onChange={event => setEditing(item => ({ ...item, subtitleText: event.target.value }))} rows={4} /><label className="shuihuo-form-label">发言者</label><Input value={editing?.speaker || '旁白'} onChange={event => setEditing(item => ({ ...item, speaker: event.target.value }))} placeholder="旁白或角色名" /><label className="shuihuo-form-label">图片提示词</label><Input.TextArea value={editing?.imagePrompt || ''} onChange={event => setEditing(item => ({ ...item, imagePrompt: event.target.value }))} rows={4} /><Checkbox checked={Boolean(editing?.imagePromptLocked)} onChange={event => setEditing(item => ({ ...item, imagePromptLocked: event.target.checked }))}>锁定图片提示词</Checkbox><label className="shuihuo-form-label">视频提示词</label><Input.TextArea value={editing?.videoPrompt || ''} onChange={event => setEditing(item => ({ ...item, videoPrompt: event.target.value }))} rows={4} /><Checkbox checked={Boolean(editing?.videoPromptLocked)} onChange={event => setEditing(item => ({ ...item, videoPromptLocked: event.target.checked }))}>锁定视频提示词</Checkbox></Modal>
    <Modal title="放大编辑" open={Boolean(expandedField)} onCancel={() => setExpandedField(null)} onOk={saveExpandedField} okText="保存" confirmLoading={busy} width={820}><label className="shuihuo-form-label">{expandedField?.label}</label><Input.TextArea value={expandedField?.value || ''} onChange={event => setExpandedField(item => ({ ...item, value: event.target.value }))} rows={16} autoFocus /></Modal>
    <Modal title="配音设置" open={Boolean(voiceConfiguring)} onCancel={() => setVoiceConfiguring(null)} onOk={saveVoiceSettings} confirmLoading={busy} okText="保存配音设置"><label className="shuihuo-form-label">{(voiceConfiguring?.segment?.speaker || '旁白') === '旁白' ? '旁白音色' : `角色“${voiceConfiguring?.segment?.speaker}”音色`}</label><Select allowClear autoFocus={voiceConfiguring?.focus === 'voice'} value={voiceConfiguring?.settings.voiceAssetId ?? undefined} placeholder={voiceAssets.length ? '选择项目音色库中的可调用音色' : '暂无音色，请先在人物场景预设中创建'} options={voiceAssets.map(asset => ({ value: asset.id, label: asset.name }))} onChange={voiceAssetId => setVoiceConfiguring(current => ({ ...current, settings: { ...current.settings, voiceAssetId: voiceAssetId ?? null } }))} />{(voiceConfiguring?.segment?.speaker || '旁白') !== '旁白' ? <p className="shuihuo-modal-note">角色音色绑定保存后会写入该角色预设，之后同名角色分镜会自动使用这个音色。</p> : null}<label className="shuihuo-form-label">配音语速</label><InputNumber autoFocus={voiceConfiguring?.focus === 'speechRate'} min={0.5} max={2} step={0.1} value={voiceConfiguring?.settings.speechRate ?? 1} onChange={speechRate => setVoiceConfiguring(current => ({ ...current, settings: { ...current.settings, speechRate: speechRate ?? 1 } }))} /><label className="shuihuo-form-label">音调</label><InputNumber autoFocus={voiceConfiguring?.focus === 'pitch'} min={-50} max={50} step={1} value={voiceConfiguring?.settings.pitch ?? 0} onChange={pitch => setVoiceConfiguring(current => ({ ...current, settings: { ...current.settings, pitch: pitch ?? 0 } }))} />{!voiceAssets.length ? <p className="shuihuo-modal-note">在“人物场景预设”创建分类为“音色”的项目预设后，即可选择旁白或绑定给角色。</p> : null}</Modal>
    <Modal title={`添加${assetCategoryLabels[binding?.category] || '预设'}`} open={Boolean(binding)} onCancel={() => setBinding(null)} onOk={saveBinding} okText="保存绑定" confirmLoading={busy}><div className="shuihuo-binding-list">{bindingAssets.map(asset => <Checkbox key={asset.id} checked={selectedAssetIds.includes(asset.id)} onChange={event => setSelectedAssetIds(ids => event.target.checked ? [...new Set([...ids, asset.id])] : ids.filter(id => id !== asset.id))}><span className="shuihuo-binding-asset"><b>{asset.name}</b><BindingAssetPreview asset={asset} onPreview={openAssetPreview} /></span></Checkbox>)}{!bindingAssets.length ? <span className="shuihuo-muted">还没有${assetCategoryLabels[binding?.category] || ''}预设，请先在人物场景预设中创建。</span> : null}</div></Modal>
    <Modal title="在后方新增分镜" open={Boolean(insertAfter)} onCancel={() => setInsertAfter(null)} onOk={saveInsert} okText="新增分镜" confirmLoading={busy}><label className="shuihuo-form-label">原文</label><Input.TextArea value={insertAfter?.sourceText || ''} onChange={event => setInsertAfter(item => ({ ...item, sourceText: event.target.value }))} rows={5} /><label className="shuihuo-form-label">字幕（可选）</label><Input.TextArea value={insertAfter?.subtitleText || ''} onChange={event => setInsertAfter(item => ({ ...item, subtitleText: event.target.value }))} rows={3} /></Modal>
    <EngineSettingsModal open={engineOpen} onClose={() => setEngineOpen(false)} />
    <AiReasoningModal open={aiReasoningOpen} projectId={project.id} segments={data.segments || []} onClose={() => setAiReasoningOpen(false)} onApplied={refreshProject} />
    <MediaModal open={Boolean(mediaUpload)} projectId={project.id} segments={data.segments || []} saving={busy} onCancel={() => setMediaUpload(null)} onUpload={saveMediaUpload} initialKind={mediaUpload?.kind || 'image'} initialSegmentId={mediaUpload?.segmentId} />
    <Modal title={mediaPreview?.kind === 'video' ? '视频预览' : '图片预览'} open={Boolean(mediaPreview)} footer={null} onCancel={() => setMediaPreview(null)} width={mediaPreview?.kind === 'video' ? 900 : 760} destroyOnClose><div className={`shuihuo-media-modal-preview is-${mediaPreview?.kind || 'image'}`}>{mediaPreview ? <MediaPreview media={safeMedia(mediaPreview)} /> : null}</div></Modal>
    <Modal title={assetPreview?.asset?.name ? `${assetPreview.asset.name}预设图` : '预设图预览'} open={Boolean(assetPreview)} footer={null} onCancel={() => setAssetPreview(null)} width={900} destroyOnClose><div className="shuihuo-media-modal-preview is-image">{assetPreview?.url ? <img src={assetPreview.url} alt={`${assetPreview.asset.name}预设图`} /> : <span className="shuihuo-muted">读取真实预设图片...</span>}</div></Modal>
    <BatchTaskModal open={Boolean(batch)} projectId={project.id} segments={data.segments || []} media={media} kind={batch?.kind || 'image'} availability={batch?.kind === 'video' ? videoAvailability : imageAvailability} initialSegmentIds={batch?.segmentIds || []} initialScope={batch?.initialScope || 'all'} onClose={() => setBatch(null)} onSubmitted={refreshProject} />
    <TaskDrawer open={tasksOpen} taskFilter={taskFilter} project={project} segments={data.segments || []} media={media} readiness={readiness} onClose={() => setTasksOpen(false)} onCompleted={refreshProject} />
  </section>;
}
